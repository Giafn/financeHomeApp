package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

type GoalUsecase struct {
	goalRepo        repository.GoalRepository
	accountRepo     repository.AccountRepository
	householdRepo   repository.HouseholdRepository
	transactionRepo repository.TransactionRepository
}

func NewGoalUsecase(
	goalRepo repository.GoalRepository,
	accountRepo repository.AccountRepository,
	householdRepo repository.HouseholdRepository,
	transactionRepo repository.TransactionRepository,
) *GoalUsecase {
	return &GoalUsecase{
		goalRepo:        goalRepo,
		accountRepo:     accountRepo,
		householdRepo:   householdRepo,
		transactionRepo: transactionRepo,
	}
}

// GoalWithProgress goal + current_amount/percentage terhitung — bukan kolom DB,
// selalu dihitung ulang dari transaksi (sama seperti pola current_balance akun & spent budget).
type GoalWithProgress struct {
	*entity.Goal
	CurrentAmount float64
	Percentage    float64
}

func (u *GoalUsecase) currentAmount(ctx context.Context, goal *entity.Goal) (float64, error) {
	incoming, err := u.goalRepo.SumTransferAmount(ctx, goal.ID, goal.LinkedAccountID, true)
	if err != nil {
		return 0, err
	}
	outgoing, err := u.goalRepo.SumTransferAmount(ctx, goal.ID, goal.LinkedAccountID, false)
	if err != nil {
		return 0, err
	}
	return incoming - outgoing, nil
}

func (u *GoalUsecase) withProgress(ctx context.Context, goal *entity.Goal) (*GoalWithProgress, error) {
	amount, err := u.currentAmount(ctx, goal)
	if err != nil {
		return nil, err
	}
	pct := 0.0
	if goal.TargetAmount > 0 {
		pct = (amount / goal.TargetAmount) * 100
	}
	return &GoalWithProgress{Goal: goal, CurrentAmount: amount, Percentage: pct}, nil
}

func (u *GoalUsecase) CreateGoal(ctx context.Context, userID uuid.UUID, name string, icon *string, targetAmount float64, linkedAccountID uuid.UUID, targetDate *time.Time) (*entity.Goal, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	account, err := u.accountRepo.FindByID(ctx, linkedAccountID)
	if err != nil {
		return nil, err
	}
	if account.HouseholdID != member.HouseholdID {
		return nil, apperror.ErrNotFound
	}
	if !account.IsActive {
		return nil, apperror.ErrAccountInactive
	}

	goal := &entity.Goal{
		HouseholdID:     member.HouseholdID,
		Name:            name,
		Icon:            icon,
		TargetAmount:    targetAmount,
		LinkedAccountID: linkedAccountID,
		TargetDate:      targetDate,
		Status:          entity.GoalActive,
		CreatedBy:       userID,
	}

	if err := u.goalRepo.Create(ctx, goal); err != nil {
		return nil, err
	}

	return goal, nil
}

func (u *GoalUsecase) ListGoals(ctx context.Context, userID uuid.UUID, status string) ([]*GoalWithProgress, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	goals, err := u.goalRepo.ListByHouseholdAndStatus(ctx, member.HouseholdID, status)
	if err != nil {
		return nil, err
	}

	result := make([]*GoalWithProgress, 0, len(goals))
	for _, g := range goals {
		gp, err := u.withProgress(ctx, g)
		if err != nil {
			return nil, err
		}
		result = append(result, gp)
	}
	return result, nil
}

type GoalDetail struct {
	*GoalWithProgress
	Contributions []*repository.TransactionListItem
}

func (u *GoalUsecase) GetGoalDetail(ctx context.Context, userID, goalID uuid.UUID) (*GoalDetail, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	goal, err := u.goalRepo.FindByIDAndHousehold(ctx, goalID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	gp, err := u.withProgress(ctx, goal)
	if err != nil {
		return nil, err
	}

	contributions, _, err := u.transactionRepo.List(ctx, member.HouseholdID, repository.TransactionFilter{
		GoalID: &goalID,
		Page:   1,
		Limit:  1000,
	})
	if err != nil {
		return nil, err
	}

	return &GoalDetail{GoalWithProgress: gp, Contributions: contributions}, nil
}

type UpdateGoalInput struct {
	Name         *string
	Icon         *string
	TargetAmount *float64
	TargetDate   *time.Time
	Status       *string
}

func (u *GoalUsecase) UpdateGoal(ctx context.Context, userID, goalID uuid.UUID, input UpdateGoalInput) (*GoalWithProgress, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	goal, err := u.goalRepo.FindByIDAndHousehold(ctx, goalID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		goal.Name = *input.Name
	}
	if input.Icon != nil {
		goal.Icon = input.Icon
	}
	if input.TargetAmount != nil {
		goal.TargetAmount = *input.TargetAmount
	}
	if input.TargetDate != nil {
		goal.TargetDate = input.TargetDate
	}
	if input.Status != nil {
		goal.Status = entity.GoalStatus(*input.Status)
	}

	if err := u.goalRepo.Update(ctx, goal); err != nil {
		return nil, err
	}

	return u.withProgress(ctx, goal)
}

func (u *GoalUsecase) DeleteGoal(ctx context.Context, userID, goalID uuid.UUID) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := u.goalRepo.FindByIDAndHousehold(ctx, goalID, member.HouseholdID); err != nil {
		return err
	}

	hasContributions, err := u.goalRepo.HasContributions(ctx, goalID)
	if err != nil {
		return err
	}
	if hasContributions {
		return apperror.ErrGoalHasContributions
	}

	return u.goalRepo.Delete(ctx, goalID, member.HouseholdID)
}

// RecalculateStatus mengimplementasikan usecase.GoalRecalculator — dipanggil TransactionUsecase
// setiap kali transaksi ber-goal_id berubah. active->achieved kalau current_amount >= target,
// achieved->active kalau turun lagi di bawah target (penarikan dana). cancelled tidak disentuh.
func (u *GoalUsecase) RecalculateStatus(ctx context.Context, goalID uuid.UUID) error {
	goal, err := u.goalRepo.FindByID(ctx, goalID)
	if err != nil {
		return err
	}
	if goal.Status == entity.GoalCancelled {
		return nil
	}

	amount, err := u.currentAmount(ctx, goal)
	if err != nil {
		return err
	}

	achieved := amount >= goal.TargetAmount
	newStatus := goal.Status
	if achieved && goal.Status == entity.GoalActive {
		newStatus = entity.GoalAchieved
	} else if !achieved && goal.Status == entity.GoalAchieved {
		newStatus = entity.GoalActive
	}

	if newStatus == goal.Status {
		return nil
	}
	return u.goalRepo.UpdateStatus(ctx, goalID, newStatus)
}
