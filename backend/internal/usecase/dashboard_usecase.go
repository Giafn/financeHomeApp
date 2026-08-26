package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/repository"
)

// DashboardUsecase menyusun 1 response agregat dari usecase modul lain — sengaja TIDAK
// menulis ulang rumus kalkulasi saldo/budget/goal di sini (rawan drift dari endpoint
// aslinya), cukup panggil AccountUsecase/BudgetUsecase/GoalUsecase/TransactionUsecase
// yang sudah ada, dan menambah query agregasi baru hanya untuk bagian yang belum
// punya usecase (upcoming bills 7 hari, tren bulanan, breakdown per anggota).
type DashboardUsecase struct {
	accountUsecase     *AccountUsecase
	budgetUsecase      *BudgetUsecase
	goalUsecase        *GoalUsecase
	transactionUsecase *TransactionUsecase
	householdRepo      repository.HouseholdRepository
	transactionRepo    repository.TransactionRepository
	billPeriodRepo     repository.BillPeriodRepository
}

func NewDashboardUsecase(
	accountUsecase *AccountUsecase,
	budgetUsecase *BudgetUsecase,
	goalUsecase *GoalUsecase,
	transactionUsecase *TransactionUsecase,
	householdRepo repository.HouseholdRepository,
	transactionRepo repository.TransactionRepository,
	billPeriodRepo repository.BillPeriodRepository,
) *DashboardUsecase {
	return &DashboardUsecase{
		accountUsecase:     accountUsecase,
		budgetUsecase:      budgetUsecase,
		goalUsecase:        goalUsecase,
		transactionUsecase: transactionUsecase,
		householdRepo:      householdRepo,
		transactionRepo:    transactionRepo,
		billPeriodRepo:     billPeriodRepo,
	}
}

const upcomingBillWindowDays = 7
const trendMonths = 6

type DashboardAccountSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CurrentBalance float64   `json:"current_balance"`
}

type DashboardBudgetSummary struct {
	TotalBudget float64 `json:"total_budget"`
	TotalSpent  float64 `json:"total_spent"`
	Percentage  float64 `json:"percentage"`
}

type DashboardUpcomingBill struct {
	BillPeriodID string  `json:"bill_period_id"`
	BillName     string  `json:"bill_name"`
	Amount       float64 `json:"amount"`
	DueDate      string  `json:"due_date"`
}

type DashboardGoal struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
}

type DashboardSummary struct {
	TotalBalance        float64                        `json:"total_balance"`
	Accounts            []DashboardAccountSummary       `json:"accounts"`
	BudgetSummary       DashboardBudgetSummary          `json:"budget_summary"`
	UpcomingBills       []DashboardUpcomingBill         `json:"upcoming_bills"`
	ActiveGoals         []DashboardGoal                 `json:"active_goals"`
	MonthlyTrend        []repository.MonthlyTrendItem   `json:"monthly_trend"`
	MemberBreakdown     []repository.MemberBreakdownItem `json:"member_breakdown"`
	RecentTransactions  []*repository.TransactionListItem `json:"recent_transactions"`
}

func (u *DashboardUsecase) GetSummary(ctx context.Context, userID uuid.UUID) (*DashboardSummary, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Saldo akun — reuse AccountUsecase.ListAccounts (formula current_balance sama persis
	// dengan yang dipakai halaman /accounts).
	accountsRaw, err := u.accountUsecase.ListAccounts(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	accounts := make([]DashboardAccountSummary, 0, len(accountsRaw))
	var totalBalance float64
	for _, a := range accountsRaw {
		balance, _ := a["current_balance"].(float64)
		id, _ := a["id"].(uuid.UUID)
		name, _ := a["name"].(string)
		accounts = append(accounts, DashboardAccountSummary{ID: id, Name: name, CurrentBalance: balance})
		totalBalance += balance
	}

	// 2. Budget bulan ini — reuse BudgetUsecase.ListBudgets (per kategori), jumlahkan jadi 1 agregat.
	currentPeriod := time.Now().Format("2006-01")
	budgets, err := u.budgetUsecase.ListBudgets(ctx, userID, currentPeriod)
	if err != nil {
		return nil, err
	}
	var budgetSummary DashboardBudgetSummary
	for _, b := range budgets {
		budgetSummary.TotalBudget += b.Amount
		budgetSummary.TotalSpent += b.Spent
	}
	if budgetSummary.TotalBudget > 0 {
		budgetSummary.Percentage = (budgetSummary.TotalSpent / budgetSummary.TotalBudget) * 100
	}

	// 3. Tagihan mendatang 7 hari — household-scoped, belum ada usecase khusus jadi query repo langsung.
	upcomingPeriods, err := u.billPeriodRepo.ListUpcomingForHousehold(ctx, member.HouseholdID, upcomingBillWindowDays)
	if err != nil {
		return nil, err
	}
	upcomingBills := make([]DashboardUpcomingBill, len(upcomingPeriods))
	for i, p := range upcomingPeriods {
		upcomingBills[i] = DashboardUpcomingBill{
			BillPeriodID: p.ID.String(),
			BillName:     p.BillName,
			Amount:       p.Amount,
			DueDate:      p.DueDate.Format("2006-01-02"),
		}
	}

	// 4. Goals aktif — reuse GoalUsecase.ListGoals (formula current_amount/percentage sama
	// dengan halaman /goals).
	goals, err := u.goalUsecase.ListGoals(ctx, userID, "active")
	if err != nil {
		return nil, err
	}
	activeGoals := make([]DashboardGoal, len(goals))
	for i, g := range goals {
		activeGoals[i] = DashboardGoal{ID: g.ID.String(), Name: g.Name, Percentage: g.Percentage}
	}

	// 5. Tren 6 bulan — household-scoped agregasi baru.
	trend, err := u.transactionRepo.MonthlyTrend(ctx, member.HouseholdID, trendMonths)
	if err != nil {
		return nil, err
	}

	// 6. Breakdown per anggota bulan ini — household-scoped agregasi baru.
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	breakdown, err := u.transactionRepo.MemberBreakdown(ctx, member.HouseholdID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	// 7. Transaksi terbaru — reuse TransactionUsecase.ListTransactions.
	recent, _, err := u.transactionUsecase.ListTransactions(ctx, userID, ListTransactionsInput{Page: 1, Limit: 10})
	if err != nil {
		return nil, err
	}

	return &DashboardSummary{
		TotalBalance:       totalBalance,
		Accounts:           accounts,
		BudgetSummary:      budgetSummary,
		UpcomingBills:      upcomingBills,
		ActiveGoals:        activeGoals,
		MonthlyTrend:       trend,
		MemberBreakdown:    breakdown,
		RecentTransactions: recent,
	}, nil
}
