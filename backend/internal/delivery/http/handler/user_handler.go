package handler

import (
	"errors"

	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/pkg/storage"
	"homeapp/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
	validate    *validator.Validate
	store       storage.Storage
}

func NewUserHandler(userUsecase *usecase.UserUsecase, store storage.Storage) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, validate: validator.New(), store: store}
}

// resolveAvatar mengubah avatar_url ter-stored menjadi URL yang bisa diakses client
// (presigned GET untuk S3; untuk local store dibiarkan apa adanya).
func (h *UserHandler) resolveAvatar(c fiber.Ctx, url *string) *string {
	if url == nil || *url == "" || h.store == nil {
		return url
	}
	resolved, err := h.store.ReadURL(c.Context(), *url)
	if err != nil {
		return url
	}
	return &resolved
}

func (h *UserHandler) toResponse(c fiber.Ctx, profile *usecase.UserProfile) dto.UserProfileResponse {
	var household *dto.UserHousehold
	if profile.Household != nil {
		household = &dto.UserHousehold{
			ID:   profile.Household.ID.String(),
			Name: profile.Household.Name,
			Role: profile.Household.Role,
		}
	}
	return dto.UserProfileResponse{
		ID:        profile.ID.String(),
		Name:      profile.Name,
		Email:     profile.Email,
		AvatarURL: h.resolveAvatar(c, profile.AvatarURL),
		Household: household,
	}
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
func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	profile, err := h.userUsecase.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil profil")
	}

	return response.Success(c, fiber.StatusOK, "", h.toResponse(c, profile))
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
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
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

	return response.Success(c, fiber.StatusOK, "profil berhasil diupdate", h.toResponse(c, profile))
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
func (h *UserHandler) ChangePassword(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
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
