package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"
)

type DashboardHandler struct {
	dashboardUsecase *usecase.DashboardUsecase
}

func NewDashboardHandler(dashboardUsecase *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{dashboardUsecase: dashboardUsecase}
}

// GetDashboard godoc
//
//	@Summary		Get dashboard summary (single aggregate fetch)
//	@Tags			Dashboard
//	@Produce		json
//	@Success		200	{object}	response.Envelope
//	@Router			/dashboard [get]
//	@Security		BearerAuth
func (h *DashboardHandler) GetDashboard(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	summary, err := h.dashboardUsecase.GetSummary(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat dashboard")
	}

	return response.Success(c, fiber.StatusOK, "ok", summary)
}
