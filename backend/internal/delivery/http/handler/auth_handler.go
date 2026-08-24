package handler

import (
	"errors"

	"family-finance-api/internal/delivery/http/dto"
	"family-finance-api/internal/pkg/apperror"
	"family-finance-api/internal/pkg/response"
	"family-finance-api/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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
// @Description Membuat user baru dengan email dan password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Failure 409 {object} response.Envelope
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "payload tidak valid")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.authUsecase.Register(c.Context(), usecase.RegisterInput{
		Name: req.Name, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrEmailAlreadyUsed) {
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "gagal mendaftarkan user")
	}

	return response.Success(c, fiber.StatusCreated, "registrasi berhasil", dto.AuthResponse{
		Token: result.Token,
		User:  dto.UserResponse{ID: result.User.ID.String(), Name: result.User.Name, Email: result.User.Email},
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
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
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
		return response.Error(c, fiber.StatusInternalServerError, "gagal login")
	}

	return response.Success(c, fiber.StatusOK, "login berhasil", dto.AuthResponse{
		Token: result.Token,
		User:  dto.UserResponse{ID: result.User.ID.String(), Name: result.User.Name, Email: result.User.Email},
	})
}
