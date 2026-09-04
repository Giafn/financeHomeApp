package usecase

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	cycleperiod "homeapp/internal/pkg/period"
	"homeapp/internal/repository"
)

// BudgetPlanCategoryItem realisasi 1 kategori (leaf) dalam periode — reuse agregasi budget/spent
// yang sudah ada di BudgetRepository.ListByPeriod, tidak dihitung ulang di sini.
type BudgetPlanCategoryItem struct {
	CategoryID   uuid.UUID
	CategoryName string
	Amount       float64
	Spent        float64
	Percentage   float64
}

// BudgetPlanUnpaidBill tagihan periode ini yang belum lunas (upcoming/overdue).
type BudgetPlanUnpaidBill struct {
	BillPeriodID uuid.UUID
	BillName     string
	Amount       float64
	DueDate      string
	Status       string
}

// BudgetPlan gabungan pemasukan + budget + tagihan 1 periode. Selain itu, jawab pertanyaan
// INTI aplikasi: apakah saldo uang BERSAMA (household, bukan akun personal) yang ada SEKARANG
// cukup buat nutup sisa rencana budget + tagihan yang belum lunas — CurrentHouseholdBalance,
// TotalNeeded, Surplus, IsSufficient. Ini soal likuiditas saat ini, sengaja tidak pakai
// TotalIncome (uang yang sudah kepake sudah otomatis mengurangi saldo akun, dihitung ulang di
// sini bakal dobel).
type BudgetPlan struct {
	Period               string
	TotalIncome          float64
	TotalBudgeted        float64
	RemainingUnallocated float64
	Categories           []BudgetPlanCategoryItem
	UnpaidBills          []BudgetPlanUnpaidBill

	CurrentHouseholdBalance float64
	TotalNeeded             float64
	Surplus                 float64
	IsSufficient            bool
}

type BudgetPlanUsecase struct {
	budgetRepo      repository.BudgetRepository
	billPeriodRepo  repository.BillPeriodRepository
	transactionRepo repository.TransactionRepository
	householdRepo   repository.HouseholdRepository
	accountUsecase  *AccountUsecase
}

func NewBudgetPlanUsecase(
	budgetRepo repository.BudgetRepository,
	billPeriodRepo repository.BillPeriodRepository,
	transactionRepo repository.TransactionRepository,
	householdRepo repository.HouseholdRepository,
	accountUsecase *AccountUsecase,
) *BudgetPlanUsecase {
	return &BudgetPlanUsecase{
		budgetRepo:      budgetRepo,
		billPeriodRepo:  billPeriodRepo,
		transactionRepo: transactionRepo,
		householdRepo:   householdRepo,
		accountUsecase:  accountUsecase,
	}
}

// GetBudgetPlan murni baca — reuse BudgetRepository (spent per kategori sudah dihitung di sana,
// termasuk expense pembayaran bill karena bill payment tercatat sebagai transaksi berkategori
// sama, jadi tidak ada double count) + TotalByType untuk income + bill_periods belum lunas.
func (u *BudgetPlanUsecase) GetBudgetPlan(ctx context.Context, userID uuid.UUID, period string) (*BudgetPlan, error) {
	if !periodPattern.MatchString(period) {
		return nil, apperror.ErrInvalidPeriodFormat
	}

	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	household, err := u.householdRepo.FindByID(ctx, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	start, end, err := cycleperiod.CyclePeriodWindow(period, household.BudgetCycleStartDay)
	if err != nil {
		return nil, apperror.ErrInvalidPeriodFormat
	}

	totalIncome, err := u.transactionRepo.TotalByType(ctx, member.HouseholdID, start, end, "income")
	if err != nil {
		return nil, err
	}

	budgets, err := u.budgetRepo.ListByPeriod(ctx, member.HouseholdID, period, household.BudgetCycleStartDay)
	if err != nil {
		return nil, err
	}

	categories := make([]BudgetPlanCategoryItem, 0, len(budgets))
	var totalBudgeted, remainingBudgetNeeded float64
	for _, b := range budgets {
		percentage := 0.0
		if b.Amount > 0 {
			percentage = (b.Spent / b.Amount) * 100
		}
		categories = append(categories, BudgetPlanCategoryItem{
			CategoryID:   b.CategoryID,
			CategoryName: b.CategoryName,
			Amount:       b.Amount,
			Spent:        b.Spent,
			Percentage:   percentage,
		})
		totalBudgeted += b.Amount
		// Kalau kategori sudah overspend, kontribusinya ke "kebutuhan" 0 (bukan negatif) —
		// overspend itu udah tercermin di saldo akun yang berkurang, bukan kebutuhan tambahan.
		if headroom := b.Amount - b.Spent; headroom > 0 {
			remainingBudgetNeeded += headroom
		}
	}

	unpaid, err := u.billPeriodRepo.ListUnpaidByHouseholdAndPeriod(ctx, member.HouseholdID, period)
	if err != nil {
		return nil, err
	}

	unpaidBills := make([]BudgetPlanUnpaidBill, len(unpaid))
	var unpaidBillsTotal float64
	for i, p := range unpaid {
		unpaidBills[i] = BudgetPlanUnpaidBill{
			BillPeriodID: p.ID,
			BillName:     p.BillName,
			Amount:       p.Amount,
			DueDate:      p.DueDate.Format("2006-01-02"),
			Status:       string(p.Status),
		}
		unpaidBillsTotal += p.Amount
	}

	// Saldo uang BERSAMA sekarang — cuma akun owner_type=household, akun personal sengaja
	// tidak dihitung (itu bukan uang keluarga).
	accountsRaw, err := u.accountUsecase.ListAccounts(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	var currentHouseholdBalance float64
	for _, a := range accountsRaw {
		ownerType, _ := a["owner_type"].(entity.AccountOwnerType)
		balance, _ := a["current_balance"].(float64)
		if ownerType != entity.AccountOwnerPersonal {
			currentHouseholdBalance += balance
		}
	}

	totalNeeded := remainingBudgetNeeded + unpaidBillsTotal
	surplus := currentHouseholdBalance - totalNeeded

	return &BudgetPlan{
		Period:               period,
		TotalIncome:          totalIncome,
		TotalBudgeted:        totalBudgeted,
		RemainingUnallocated: totalIncome - totalBudgeted,
		Categories:           categories,
		UnpaidBills:          unpaidBills,

		CurrentHouseholdBalance: currentHouseholdBalance,
		TotalNeeded:             totalNeeded,
		Surplus:                 surplus,
		IsSufficient:            surplus >= 0,
	}, nil
}
