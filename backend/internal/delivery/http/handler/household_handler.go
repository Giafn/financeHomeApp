package handler

import (
	"errors"

	"family-finance-api/internal/delivery/http/dto"
	"family-finance-api/internal/delivery/http/middleware"
	"family-finance-api/internal/pkg/apperror"
	"family-finance-api/internal/pkg/response"
	"family-finance-api/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type HouseholdHandler struct {
	householdUsecase *usecase.HouseholdUsecase
	validate         *validator.Validate
}

func NewHouseholdHandler(householdUsecase *usecase.HouseholdUsecase) *HouseholdHandler {
	return &HouseholdHandler{householdUsecase: householdUsecase, validate: validator.New()}
}

// Create godoc
// @Summary Buat rumah tangga baru
// @Description User membuat rumah tangga baru
// @Tags households
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateHouseholdRequest true "Create household request"
// @Success 201 {object} response.Envelope{data=dto.HouseholdResponse}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Failure 409 {object} response.Envelope
// @Router /households [post]
func (h *HouseholdHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateHouseholdRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	household, err := h.householdUsecase.CreateHousehold(c.Context(), userID, req.Name)
	if err != nil {
		if errors.Is(err, apperror.ErrAlreadyInHousehold) {
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal membuat rumah tangga")
	}

	return response.Success(c, fiber.StatusCreated, "rumah tangga berhasil dibuat", dto.HouseholdResponse{
		ID: household.ID.String(), Name: household.Name,
	})
}

// Join godoc
// @Summary Join rumah tangga dengan kode undangan
// @Description User bergabung ke rumah tangga menggunakan kode undangan
// @Tags households
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JoinHouseholdRequest true "Join household request"
// @Success 200 {object} response.Envelope{data=dto.HouseholdResponse}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /households/join [post]
func (h *HouseholdHandler) Join(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.JoinHouseholdRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	household, err := h.householdUsecase.JoinHousehold(c.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, apperror.ErrInvitationInvalid) || errors.Is(err, apperror.ErrAlreadyInHousehold) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal bergabung ke rumah tangga")
	}

	return response.Success(c, fiber.StatusOK, "berhasil bergabung ke rumah tangga", dto.HouseholdResponse{
		ID: household.ID.String(), Name: household.Name,
	})
}

// CreateInvitation godoc
// @Summary Buat kode undangan rumah tangga
// @Description Membuat kode undangan untuk mengajak user lain bergabung
// @Tags households
// @Produce json
// @Security BearerAuth
// @Success 201 {object} response.Envelope{data=dto.InvitationResponse}
// @Failure 401 {object} response.Envelope
// @Failure 403 {object} response.Envelope
// @Router /households/invitations [post]
func (h *HouseholdHandler) CreateInvitation(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	invitation, err := h.householdUsecase.CreateInvitation(c.Context(), userID)
	if err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal membuat kode undangan")
	}

	return response.Success(c, fiber.StatusCreated, "kode undangan berhasil dibuat", dto.InvitationResponse{
		Code:      invitation.Code,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetDetail godoc
// @Summary Detail rumah tangga
// @Tags households
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /households/me [get]
func (h *HouseholdHandler) GetDetail(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	detail, err := h.householdUsecase.GetHouseholdDetail(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "ok", detail)
}

// UpdateName godoc
// @Summary Update nama rumah tangga
// @Tags households
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Household ID"
// @Param request body dto.UpdateHouseholdRequest true "Update request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 403 {object} response.Envelope
// @Router /households/{id} [patch]
func (h *HouseholdHandler) UpdateName(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	householdIDStr := c.Params("id")
	householdID, err := uuid.Parse(householdIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID rumah tangga tidak valid")
	}

	var req dto.UpdateHouseholdRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.householdUsecase.UpdateHouseholdName(c.Context(), userID, householdID, req.Name); err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal update rumah tangga")
	}

	return response.Success(c, fiber.StatusOK, "nama rumah tangga berhasil diupdate", nil)
}

// GetMembers godoc
// @Summary List anggota rumah tangga
// @Tags households
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /households/members [get]
func (h *HouseholdHandler) GetMembers(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	members, err := h.householdUsecase.GetMembers(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "ok", members)
}

// RemoveMember godoc
// @Summary Keluarkan anggota dari rumah tangga
// @Tags households
// @Produce json
// @Security BearerAuth
// @Param targetUserID query string true "User ID to remove"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 403 {object} response.Envelope
// @Router /households/members [delete]
func (h *HouseholdHandler) RemoveMember(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	targetUserIDStr := c.Query("targetUserID")
	if targetUserIDStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "targetUserID diperlukan")
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "targetUserID tidak valid")
	}

	if err := h.householdUsecase.RemoveMember(c.Context(), userID, targetUserID); err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		if errors.Is(err, apperror.ErrCannotRemoveSoleOwner) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal keluarkan anggota")
	}

	return response.Success(c, fiber.StatusOK, "anggota berhasil dikeluarkan", nil)
}

// GetActiveInvitation godoc
// @Summary Dapatkan kode undangan aktif
// @Tags households
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /households/invitations/active [get]
func (h *HouseholdHandler) GetActiveInvitation(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	invitation, err := h.householdUsecase.GetActiveInvitation(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error())
	}

	if invitation == nil {
		return response.Success(c, fiber.StatusOK, "ok", nil)
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.InvitationResponse{
		Code:      invitation.Code,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
