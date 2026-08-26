package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"
)

type ReportHandler struct {
	reportUsecase *usecase.ReportUsecase
}

func NewReportHandler(reportUsecase *usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: reportUsecase}
}

func mapReportErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrInvalidPeriodFormat), errors.Is(err, apperror.ErrInvalidPeriodType), errors.Is(err, apperror.ErrInvalidExportFormat):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat laporan")
	}
}

// GetTrend godoc
//
//	@Summary		Monthly income vs expense trend
//	@Tags			Reports
//	@Produce		json
//	@Param			months	query	int	false	"Number of months (default 6)"
//	@Success		200	{object}	response.Envelope
//	@Router			/reports/trend [get]
//	@Security		BearerAuth
func (h *ReportHandler) GetTrend(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	months := 0
	if v := c.Query("months", ""); v != "" {
		if m, err := strconv.Atoi(v); err == nil {
			months = m
		}
	}

	trend, err := h.reportUsecase.GetTrend(c.Context(), userID, months)
	if err != nil {
		return mapReportErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "ok", trend)
}

// GetCategoryBreakdown godoc
//
//	@Summary		Expense breakdown by category for a period
//	@Tags			Reports
//	@Produce		json
//	@Param			period		query	string	true	"YYYY-MM or YYYY"
//	@Param			period_type	query	string	true	"month or year"
//	@Param			type		query	string	false	"expense (default) or income"
//	@Success		200	{object}	response.Envelope{data=[]dto.CategoryBreakdownResponse}
//	@Failure		400	{object}	response.Envelope
//	@Router			/reports/category-breakdown [get]
//	@Security		BearerAuth
func (h *ReportHandler) GetCategoryBreakdown(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	period := c.Query("period", "")
	periodType := c.Query("period_type", "")
	txType := c.Query("type", "expense")

	items, err := h.reportUsecase.GetCategoryBreakdown(c.Context(), userID, period, periodType, txType)
	if err != nil {
		return mapReportErr(c, err)
	}

	responses := make([]dto.CategoryBreakdownResponse, len(items))
	for i, it := range items {
		responses[i] = dto.CategoryBreakdownResponse{
			CategoryID:   it.CategoryID.String(),
			CategoryName: it.CategoryName,
			Total:        it.Total,
			Percentage:   it.Percentage,
		}
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// GetMemberBreakdown godoc
//
//	@Summary		Income/expense breakdown by household member for a period
//	@Tags			Reports
//	@Produce		json
//	@Param			period		query	string	true	"YYYY-MM or YYYY"
//	@Param			period_type	query	string	true	"month or year"
//	@Success		200	{object}	response.Envelope{data=[]dto.MemberBreakdownResponse}
//	@Failure		400	{object}	response.Envelope
//	@Router			/reports/member-breakdown [get]
//	@Security		BearerAuth
func (h *ReportHandler) GetMemberBreakdown(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	period := c.Query("period", "")
	periodType := c.Query("period_type", "")

	items, err := h.reportUsecase.GetMemberBreakdown(c.Context(), userID, period, periodType)
	if err != nil {
		return mapReportErr(c, err)
	}

	responses := make([]dto.MemberBreakdownResponse, len(items))
	for i, it := range items {
		responses[i] = dto.MemberBreakdownResponse{
			UserID:       it.UserID.String(),
			Name:         it.Name,
			TotalExpense: it.TotalExpense,
			TotalIncome:  it.TotalIncome,
		}
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// GetComparison godoc
//
//	@Summary		Period-over-period expense comparison
//	@Tags			Reports
//	@Produce		json
//	@Param			period		query	string	true	"YYYY-MM or YYYY"
//	@Param			period_type	query	string	true	"month or year"
//	@Success		200	{object}	response.Envelope{data=dto.ComparisonResponse}
//	@Failure		400	{object}	response.Envelope
//	@Router			/reports/comparison [get]
//	@Security		BearerAuth
func (h *ReportHandler) GetComparison(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	period := c.Query("period", "")
	periodType := c.Query("period_type", "")

	result, err := h.reportUsecase.GetComparison(c.Context(), userID, period, periodType)
	if err != nil {
		return mapReportErr(c, err)
	}

	byCategory := make([]dto.ComparisonCategoryResponse, len(result.ByCategory))
	for i, c := range result.ByCategory {
		byCategory[i] = dto.ComparisonCategoryResponse{
			CategoryName:   c.CategoryName,
			Current:        c.Current,
			Previous:       c.Previous,
			DiffPercentage: c.DiffPercentage,
		}
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.ComparisonResponse{
		Current:        dto.ComparisonPeriodResponse{Period: result.Current.Period, TotalExpense: result.Current.TotalExpense},
		Previous:       dto.ComparisonPeriodResponse{Period: result.Previous.Period, TotalExpense: result.Previous.TotalExpense},
		DiffAmount:     result.DiffAmount,
		DiffPercentage: result.DiffPercentage,
		ByCategory:     byCategory,
	})
}

// ExportReport godoc
//
//	@Summary		Export report as PDF or Excel
//	@Tags			Reports
//	@Produce		application/octet-stream
//	@Param			format		query	string	true	"pdf or excel"
//	@Param			period		query	string	true	"YYYY-MM or YYYY"
//	@Param			period_type	query	string	true	"month or year"
//	@Success		200
//	@Failure		400	{object}	response.Envelope
//	@Router			/reports/export [get]
//	@Security		BearerAuth
func (h *ReportHandler) ExportReport(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	format := c.Query("format", "")
	period := c.Query("period", "")
	periodType := c.Query("period_type", "")

	data, filename, contentType, err := h.reportUsecase.GenerateExport(c.Context(), userID, format, period, periodType)
	if err != nil {
		return mapReportErr(c, err)
	}

	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(data)
}
