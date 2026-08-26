package usecase

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

var periodPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

type BudgetUsecase struct {
	budgetRepo    repository.BudgetRepository
	categoryRepo  repository.CategoryRepository
	householdRepo repository.HouseholdRepository
}

func NewBudgetUsecase(budgetRepo repository.BudgetRepository, categoryRepo repository.CategoryRepository, householdRepo repository.HouseholdRepository) *BudgetUsecase {
	return &BudgetUsecase{
		budgetRepo:    budgetRepo,
		categoryRepo:  categoryRepo,
		householdRepo: householdRepo,
	}
}

func (u *BudgetUsecase) validateCategory(ctx context.Context, householdID, categoryID uuid.UUID) error {
	category, err := u.categoryRepo.FindByID(ctx, categoryID, householdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	if category.Type != entity.CategoryExpense {
		return apperror.ErrCategoryNotExpense
	}
	return nil
}

func (u *BudgetUsecase) CreateBudget(ctx context.Context, userID, categoryID uuid.UUID, period string, amount float64) (*entity.Budget, error) {
	if !periodPattern.MatchString(period) {
		return nil, apperror.ErrInvalidPeriodFormat
	}

	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := u.validateCategory(ctx, member.HouseholdID, categoryID); err != nil {
		return nil, err
	}

	if _, err := u.budgetRepo.FindByCategoryPeriod(ctx, member.HouseholdID, categoryID, period); err == nil {
		return nil, apperror.ErrBudgetAlreadyExists
	} else if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	budget := &entity.Budget{
		HouseholdID: member.HouseholdID,
		CategoryID:  categoryID,
		Period:      period,
		Amount:      amount,
		CreatedBy:   userID,
	}

	if err := u.budgetRepo.Create(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

func (u *BudgetUsecase) ListBudgets(ctx context.Context, userID uuid.UUID, period string) ([]*repository.BudgetWithSpent, error) {
	if !periodPattern.MatchString(period) {
		return nil, apperror.ErrInvalidPeriodFormat
	}

	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := u.budgetRepo.ListByPeriod(ctx, member.HouseholdID, period)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []*repository.BudgetWithSpent{}
	}
	return items, nil
}

func (u *BudgetUsecase) UpdateBudget(ctx context.Context, userID, budgetID uuid.UUID, amount float64) (*entity.Budget, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	budget, err := u.budgetRepo.FindByID(ctx, budgetID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	budget.Amount = amount
	if err := u.budgetRepo.Update(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

func (u *BudgetUsecase) DeleteBudget(ctx context.Context, userID, budgetID uuid.UUID) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := u.budgetRepo.FindByID(ctx, budgetID, member.HouseholdID); err != nil {
		return err
	}

	return u.budgetRepo.Delete(ctx, budgetID, member.HouseholdID)
}
