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

type UserHandler struct {
	userUsecase *usecase.UserUsecase
	validate    *validator.Validate
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, validate: validator.New()}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Ambil profil user dan household yang diikuti
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	profile, err := h.userUsecase.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil profil")
	}

	return response.Success(c, fiber.StatusOK, "", dto.UserProfileResponse{
		ID:        profile.ID.String(),
		Name:      profile.Name,
		Email:     profile.Email,
		AvatarURL: profile.AvatarURL,
		Household: func() *dto.UserHousehold {
			if profile.Household == nil {
				return nil
			}
			return &dto.UserHousehold{
				ID:   profile.Household.ID.String(),
				Name: profile.Household.Name,
				Role: profile.Household.Role,
			}
		}(),
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update nama dan/atau avatar user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /users/me [patch]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	profile, err := h.userUsecase.UpdateProfile(c.Context(), userID, usecase.UpdateProfileInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal update profil")
	}

	return response.Success(c, fiber.StatusOK, "profil berhasil diupdate", dto.UserProfileResponse{
		ID:        profile.ID.String(),
		Name:      profile.Name,
		Email:     profile.Email,
		AvatarURL: profile.AvatarURL,
		Household: func() *dto.UserHousehold {
			if profile.Household == nil {
				return nil
			}
			return &dto.UserHousehold{
				ID:   profile.Household.ID.String(),
				Name: profile.Household.Name,
				Role: profile.Household.Role,
			}
		}(),
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Ubah password user dengan validasi password lama
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ChangePasswordRequest true "Change password request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /users/me/change-password [post]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.userUsecase.ChangePassword(c.Context(), userID, usecase.ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrInvalidCredential) {
			return response.Error(c, fiber.StatusBadRequest, "password lama tidak sesuai")
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengubah password")
	}

	return response.Success(c, fiber.StatusOK, "password berhasil diubah", nil)
}
