package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"
)

type BudgetPlanHandler struct {
	budgetPlanUsecase *usecase.BudgetPlanUsecase
}

func NewBudgetPlanHandler(budgetPlanUsecase *usecase.BudgetPlanUsecase) *BudgetPlanHandler {
	return &BudgetPlanHandler{budgetPlanUsecase: budgetPlanUsecase}
}

// GetBudgetPlan godoc
//
//	@Summary		Rencana anggaran gabungan (pemasukan, budget per kategori, tagihan belum lunas) untuk 1 periode
//	@Tags			BudgetPlan
//	@Produce		json
//	@Param			period	query	string	true	"Period YYYY-MM"
//	@Success		200	{object}	response.Envelope{data=dto.BudgetPlanResponse}
//	@Failure		400	{object}	response.Envelope
//	@Router			/budget-plan [get]
//	@Security		BearerAuth
func (h *BudgetPlanHandler) GetBudgetPlan(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	period := c.Query("period", "")

	plan, err := h.budgetPlanUsecase.GetBudgetPlan(c.Context(), userID, period)
	if err != nil {
		if errors.Is(err, apperror.ErrInvalidPeriodFormat) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat rencana anggaran")
	}

	categories := make([]dto.BudgetPlanCategoryResponse, len(plan.Categories))
	for i, cat := range plan.Categories {
		categories[i] = dto.BudgetPlanCategoryResponse{
			CategoryID:   cat.CategoryID.String(),
			CategoryName: cat.CategoryName,
			Amount:       cat.Amount,
			Spent:        cat.Spent,
			Percentage:   cat.Percentage,
		}
	}

	unpaidBills := make([]dto.BudgetPlanUnpaidBillResponse, len(plan.UnpaidBills))
	for i, b := range plan.UnpaidBills {
		unpaidBills[i] = dto.BudgetPlanUnpaidBillResponse{
			BillPeriodID: b.BillPeriodID.String(),
			BillName:     b.BillName,
			Amount:       b.Amount,
			DueDate:      b.DueDate,
			Status:       b.Status,
		}
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.BudgetPlanResponse{
		Period:               plan.Period,
		TotalIncome:          plan.TotalIncome,
		TotalBudgeted:        plan.TotalBudgeted,
		RemainingUnallocated: plan.RemainingUnallocated,
		Categories:           categories,
		UnpaidBills:          unpaidBills,

		CurrentHouseholdBalance: plan.CurrentHouseholdBalance,
		TotalNeeded:             plan.TotalNeeded,
		Surplus:                 plan.Surplus,
		IsSufficient:            plan.IsSufficient,
	})
}
