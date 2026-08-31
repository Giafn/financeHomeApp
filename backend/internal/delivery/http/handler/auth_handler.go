package handler

import (
	"errors"

	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	validate    *validator.Validate
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, validate: validator.New()}
}

// Register godoc
// @Summary Register user baru
// @Description Membuat user baru dan mengirim email verifikasi
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 409 {object} response.Envelope
// @Router /auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.authUsecase.Register(c.Context(), usecase.RegisterInput{
		Name: req.Name, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrEmailAlreadyUsed) {
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal mendaftarkan user")
	}

	return response.Success(c, fiber.StatusCreated, "registrasi berhasil, silakan verifikasi email kamu", dto.AuthResponse{
		User: dto.UserResponse{ID: user.ID.String(), Name: user.Name, Email: user.Email},
	})
}

// Login godoc
// @Summary Login user
// @Description Login dengan email dan password, return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.authUsecase.Login(c.Context(), usecase.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrInvalidCredential) {
			return response.Error(c, fiber.StatusUnauthorized, err.Error())
		}
		if errors.Is(err, apperror.ErrEmailNotVerified) {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal login")
	}

	return response.Success(c, fiber.StatusOK, "login berhasil", dto.AuthResponse{
		Token: result.Token,
		User:  dto.UserResponse{ID: result.User.ID.String(), Name: result.User.Name, Email: result.User.Email},
	})
}

// VerifyEmail godoc
// @Summary Verifikasi email user
// @Description Validasi token verifikasi email setelah registrasi
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verify email request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 410 {object} response.Envelope
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.authUsecase.VerifyEmail(c.Context(), req.Token)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidVerificationToken):
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, apperror.ErrVerificationTokenExpired):
			return response.Error(c, fiber.StatusGone, err.Error())
		default:
			return response.Error(c, fiber.StatusInternalServerError, "gagal verifikasi email")
		}
	}

	return response.Success(c, fiber.StatusOK, "email berhasil diverifikasi", nil)
}

// ResendVerification godoc
// @Summary Kirim ulang email verifikasi
// @Description Mengirim ulang email verifikasi jika tidak diterima
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "Resend verification request"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.authUsecase.ResendVerification(c.Context(), usecase.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidCredential):
			return response.Error(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperror.ErrEmailAlreadyVerified):
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		default:
			return response.Error(c, fiber.StatusInternalServerError, "gagal mengirim email verifikasi")
		}
	}

	return response.Success(c, fiber.StatusOK, "email verifikasi telah dikirim ulang", nil)
}
