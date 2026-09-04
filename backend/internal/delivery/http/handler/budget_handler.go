package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/repository"
	"homeapp/internal/usecase"
)

type BudgetHandler struct {
	budgetUsecase *usecase.BudgetUsecase
	validator     *validator.Validate
}

func NewBudgetHandler(budgetUsecase *usecase.BudgetUsecase, validator *validator.Validate) *BudgetHandler {
	return &BudgetHandler{budgetUsecase: budgetUsecase, validator: validator}
}

func mapBudgetErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "Budget atau kategori tidak ditemukan")
	case errors.Is(err, apperror.ErrCategoryNotExpense),
		errors.Is(err, apperror.ErrCategoryHasChildren):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrInvalidPeriodFormat):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrBudgetAlreadyExists):
		return response.Error(c, fiber.StatusConflict, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memproses budget")
	}
}

// CreateBudget godoc
//
//	@Summary		Create a budget for a category+period
//	@Tags			Budgets
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateBudgetRequest	true	"Budget data"
//	@Success		201		{object}	response.Envelope{data=dto.BudgetResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/budgets [post]
//	@Security		BearerAuth
func (h *BudgetHandler) CreateBudget(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateBudgetRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID kategori tidak valid")
	}

	budget, err := h.budgetUsecase.CreateBudget(c.Context(), userID, categoryID, req.Period, req.Amount)
	if err != nil {
		return mapBudgetErr(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Budget berhasil dibuat", dto.BudgetResponse{
		ID:         budget.ID.String(),
		CategoryID: budget.CategoryID.String(),
		Period:     budget.Period,
		Amount:     budget.Amount,
	})
}

// ListBudgets godoc
//
//	@Summary		List budgets for a period with spent/percentage
//	@Tags			Budgets
//	@Produce		json
//	@Param			period	query	string	true	"Period YYYY-MM"
//	@Success		200	{object}	response.Envelope{data=[]dto.BudgetResponse}
//	@Router			/budgets [get]
//	@Security		BearerAuth
func (h *BudgetHandler) ListBudgets(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	period := c.Query("period", "")

	items, err := h.budgetUsecase.ListBudgets(c.Context(), userID, period)
	if err != nil {
		return mapBudgetErr(c, err)
	}

	responses := make([]dto.BudgetResponse, len(items))
	for i, item := range items {
		responses[i] = *mapBudgetToResponse(item)
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// UpdateBudget godoc
//
//	@Summary		Update a budget's amount
//	@Tags			Budgets
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Budget ID"
//	@Param			body	body		dto.UpdateBudgetRequest	true	"Budget data"
//	@Success		200		{object}	response.Envelope{data=dto.BudgetResponse}
//	@Failure		404		{object}	response.Envelope
//	@Router			/budgets/{id} [patch]
//	@Security		BearerAuth
func (h *BudgetHandler) UpdateBudget(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	budgetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID budget tidak valid")
	}

	var req dto.UpdateBudgetRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	budget, err := h.budgetUsecase.UpdateBudget(c.Context(), userID, budgetID, req.Amount)
	if err != nil {
		return mapBudgetErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.BudgetResponse{
		ID:         budget.ID.String(),
		CategoryID: budget.CategoryID.String(),
		Period:     budget.Period,
		Amount:     budget.Amount,
	})
}

// DeleteBudget godoc
//
//	@Summary		Delete a budget
//	@Tags			Budgets
//	@Produce		json
//	@Param			id	path	string	true	"Budget ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/budgets/{id} [delete]
//	@Security		BearerAuth
func (h *BudgetHandler) DeleteBudget(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	budgetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID budget tidak valid")
	}

	if err := h.budgetUsecase.DeleteBudget(c.Context(), userID, budgetID); err != nil {
		return mapBudgetErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Budget berhasil dihapus", nil)
}

func mapBudgetToResponse(item *repository.BudgetWithSpent) *dto.BudgetResponse {
	percentage := 0.0
	if item.Amount > 0 {
		percentage = (item.Spent / item.Amount) * 100
	}
	return &dto.BudgetResponse{
		ID:           item.ID.String(),
		CategoryID:   item.CategoryID.String(),
		CategoryName: item.CategoryName,
		Period:       item.Period,
		Amount:       item.Amount,
		Spent:        item.Spent,
		Percentage:   percentage,
	}
}
