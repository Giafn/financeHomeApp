package handler

import (
	"errors"

	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountUsecase *usecase.AccountUsecase
	validate       *validator.Validate
}

func NewAccountHandler(accountUsecase *usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{accountUsecase: accountUsecase, validate: validator.New()}
}

// Create godoc
// @Summary Buat akun baru
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAccountRequest true "Create account request"
// @Success 201 {object} response.Envelope{data=dto.AccountResponse}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /accounts [post]
func (h *AccountHandler) Create(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	account, err := h.accountUsecase.CreateAccount(c.Context(), userID, req.Name, req.Type, req.InitialBalance)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal membuat akun")
	}

	return response.Success(c, fiber.StatusCreated, "akun berhasil dibuat", dto.AccountResponse{
		ID:             account.ID.String(),
		Name:           account.Name,
		Type:           string(account.Type),
		InitialBalance: account.InitialBalance,
		CurrentBalance: account.InitialBalance,
		IsActive:       account.IsActive,
	})
}

// List godoc
// @Summary List akun household
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param include_inactive query bool false "Include inactive accounts"
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /accounts [get]
func (h *AccountHandler) List(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	includeInactive := c.Query("include_inactive", "false") == "true"

	accounts, err := h.accountUsecase.ListAccounts(c.Context(), userID, includeInactive)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal memuat akun")
	}

	return response.Success(c, fiber.StatusOK, "ok", accounts)
}

// GetDetail godoc
// @Summary Detail akun
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.Envelope{data=dto.AccountResponse}
// @Failure 404 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetDetail(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID akun tidak valid")
	}

	account, err := h.accountUsecase.GetAccount(c.Context(), userID, accountID)
	if err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusNotFound, "akun tidak ditemukan")
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.AccountResponse{
		ID:             account.ID.String(),
		Name:           account.Name,
		Type:           string(account.Type),
		InitialBalance: account.InitialBalance,
		CurrentBalance: account.InitialBalance,
		IsActive:       account.IsActive,
	})
}

// Update godoc
// @Summary Update akun
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body dto.UpdateAccountRequest true "Update account request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 404 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /accounts/{id} [patch]
func (h *AccountHandler) Update(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID akun tidak valid")
	}

	var req dto.UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	// Build updates map dari non-nil fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.accountUsecase.UpdateAccount(c.Context(), userID, accountID, updates); err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusNotFound, "akun tidak ditemukan")
	}

	return response.Success(c, fiber.StatusOK, "akun berhasil diupdate", nil)
}
