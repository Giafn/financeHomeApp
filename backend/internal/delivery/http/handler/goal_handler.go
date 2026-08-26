package handler

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"
)

type GoalHandler struct {
	goalUsecase *usecase.GoalUsecase
	validator   *validator.Validate
}

func NewGoalHandler(goalUsecase *usecase.GoalUsecase, validator *validator.Validate) *GoalHandler {
	return &GoalHandler{goalUsecase: goalUsecase, validator: validator}
}

func mapGoalErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "Goal atau akun tidak ditemukan")
	case errors.Is(err, apperror.ErrAccountInactive):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrGoalHasContributions):
		return response.Error(c, fiber.StatusConflict, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memproses goal")
	}
}

// CreateGoal godoc
//
//	@Summary		Create a savings goal
//	@Tags			Goals
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateGoalRequest	true	"Goal data"
//	@Success		201		{object}	response.Envelope{data=dto.GoalResponse}
//	@Failure		400		{object}	response.Envelope
//	@Router			/goals [post]
//	@Security		BearerAuth
func (h *GoalHandler) CreateGoal(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateGoalRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	linkedAccountID, err := uuid.Parse(req.LinkedAccountID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID akun tidak valid")
	}

	var targetDate *time.Time
	if req.TargetDate != nil {
		d, err := time.Parse(dateLayoutConst, *req.TargetDate)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Format tanggal tidak valid")
		}
		targetDate = &d
	}

	goal, err := h.goalUsecase.CreateGoal(c.Context(), userID, req.Name, req.Icon, req.TargetAmount, linkedAccountID, targetDate)
	if err != nil {
		return mapGoalErr(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Goal berhasil dibuat", mapGoalToResponse(&usecase.GoalWithProgress{Goal: goal, CurrentAmount: 0, Percentage: 0}))
}

// ListGoals godoc
//
//	@Summary		List savings goals
//	@Tags			Goals
//	@Produce		json
//	@Param			status	query	string	false	"Filter by status (active/achieved/cancelled)"
//	@Success		200	{object}	response.Envelope{data=[]dto.GoalResponse}
//	@Router			/goals [get]
//	@Security		BearerAuth
func (h *GoalHandler) ListGoals(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	status := c.Query("status", "")

	goals, err := h.goalUsecase.ListGoals(c.Context(), userID, status)
	if err != nil {
		return mapGoalErr(c, err)
	}

	responses := make([]dto.GoalResponse, len(goals))
	for i, g := range goals {
		responses[i] = *mapGoalToResponse(g)
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// GetGoalDetail godoc
//
//	@Summary		Get goal detail with contribution history
//	@Tags			Goals
//	@Produce		json
//	@Param			id	path	string	true	"Goal ID"
//	@Success		200	{object}	response.Envelope{data=dto.GoalDetailResponse}
//	@Failure		404	{object}	response.Envelope
//	@Router			/goals/{id} [get]
//	@Security		BearerAuth
func (h *GoalHandler) GetGoalDetail(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID goal tidak valid")
	}

	detail, err := h.goalUsecase.GetGoalDetail(c.Context(), userID, goalID)
	if err != nil {
		return mapGoalErr(c, err)
	}

	contributions := make([]dto.TransactionResponse, len(detail.Contributions))
	for i, tx := range detail.Contributions {
		contributions[i] = *mapTransactionToResponse(tx)
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.GoalDetailResponse{
		GoalResponse:  *mapGoalToResponse(detail.GoalWithProgress),
		Contributions: contributions,
	})
}

// UpdateGoal godoc
//
//	@Summary		Update a goal
//	@Tags			Goals
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Goal ID"
//	@Param			body	body		dto.UpdateGoalRequest	true	"Goal data"
//	@Success		200		{object}	response.Envelope{data=dto.GoalResponse}
//	@Failure		404		{object}	response.Envelope
//	@Router			/goals/{id} [patch]
//	@Security		BearerAuth
func (h *GoalHandler) UpdateGoal(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID goal tidak valid")
	}

	var req dto.UpdateGoalRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	input := usecase.UpdateGoalInput{
		Name:         req.Name,
		Icon:         req.Icon,
		TargetAmount: req.TargetAmount,
		Status:       req.Status,
	}
	if req.TargetDate != nil {
		d, err := time.Parse(dateLayoutConst, *req.TargetDate)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Format tanggal tidak valid")
		}
		input.TargetDate = &d
	}

	goal, err := h.goalUsecase.UpdateGoal(c.Context(), userID, goalID, input)
	if err != nil {
		return mapGoalErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "ok", mapGoalToResponse(goal))
}

// DeleteGoal godoc
//
//	@Summary		Delete a goal
//	@Tags			Goals
//	@Produce		json
//	@Param			id	path	string	true	"Goal ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/goals/{id} [delete]
//	@Security		BearerAuth
func (h *GoalHandler) DeleteGoal(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID goal tidak valid")
	}

	if err := h.goalUsecase.DeleteGoal(c.Context(), userID, goalID); err != nil {
		return mapGoalErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Goal berhasil dihapus", nil)
}

func mapGoalToResponse(g *usecase.GoalWithProgress) *dto.GoalResponse {
	var targetDate *string
	if g.TargetDate != nil {
		s := g.TargetDate.Format(dateLayoutConst)
		targetDate = &s
	}

	return &dto.GoalResponse{
		ID:              g.ID.String(),
		Name:            g.Name,
		Icon:            g.Icon,
		TargetAmount:    g.TargetAmount,
		LinkedAccountID: g.LinkedAccountID.String(),
		TargetDate:      targetDate,
		Status:          string(g.Status),
		CurrentAmount:   g.CurrentAmount,
		Percentage:      g.Percentage,
	}
}
