package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/usecase"
)

const defaultReminderDaysBefore = 5

type BillHandler struct {
	billUsecase *usecase.BillUsecase
	validator   *validator.Validate
}

func NewBillHandler(billUsecase *usecase.BillUsecase, validator *validator.Validate) *BillHandler {
	return &BillHandler{billUsecase: billUsecase, validator: validator}
}

func mapBillErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "Tagihan, periode, atau kategori tidak ditemukan")
	case errors.Is(err, apperror.ErrCategoryNotExpense),
		errors.Is(err, apperror.ErrInvalidPeriodFormat),
		errors.Is(err, apperror.ErrCategoryRequired),
		errors.Is(err, apperror.ErrCategoryTypeMismatch),
		errors.Is(err, apperror.ErrCategoryHasChildren),
		errors.Is(err, apperror.ErrAccountInactive):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrBillPeriodAlreadyPaid):
		return response.Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, apperror.ErrPersonalAccountForbidden):
		return response.Error(c, fiber.StatusForbidden, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memproses tagihan")
	}
}

// CreateBill godoc
//
//	@Summary		Create a recurring bill
//	@Tags			Bills
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateBillRequest	true	"Bill data"
//	@Success		201		{object}	response.Envelope{data=dto.CreateBillResponse}
//	@Failure		400		{object}	response.Envelope
//	@Router			/bills [post]
//	@Security		BearerAuth
func (h *BillHandler) CreateBill(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateBillRequest
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

	reminderDays := req.ReminderDaysBefore
	if reminderDays == 0 {
		reminderDays = defaultReminderDaysBefore
	}

	bill, periods, err := h.billUsecase.CreateBill(c.Context(), userID, req.Name, categoryID, req.Amount, req.DueDay, req.StartPeriod, req.EndPeriod, reminderDays)
	if err != nil {
		return mapBillErr(c, err)
	}

	periodResponses := make([]dto.BillPeriodResponse, len(periods))
	for i, p := range periods {
		periodResponses[i] = *mapBillPeriodToResponse(p)
	}

	return response.Success(c, fiber.StatusCreated, "Tagihan berhasil dibuat", dto.CreateBillResponse{
		BillResponse: *mapBillToResponse(&usecase.BillWithNextPeriod{Bill: bill}),
		BillPeriods:  periodResponses,
	})
}

// ListBills godoc
//
//	@Summary		List bills
//	@Tags			Bills
//	@Produce		json
//	@Param			is_active	query	bool	false	"Filter by active status"
//	@Success		200	{object}	response.Envelope{data=[]dto.BillResponse}
//	@Router			/bills [get]
//	@Security		BearerAuth
func (h *BillHandler) ListBills(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var isActive *bool
	if v := c.Query("is_active", ""); v != "" {
		b := v == "true"
		isActive = &b
	}

	bills, err := h.billUsecase.ListBills(c.Context(), userID, isActive)
	if err != nil {
		return mapBillErr(c, err)
	}

	responses := make([]dto.BillResponse, len(bills))
	for i, b := range bills {
		responses[i] = *mapBillToResponse(b)
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// GetBillPeriods godoc
//
//	@Summary		Get all periods for a bill
//	@Tags			Bills
//	@Produce		json
//	@Param			id	path	string	true	"Bill ID"
//	@Success		200	{object}	response.Envelope{data=[]dto.BillPeriodResponse}
//	@Failure		404	{object}	response.Envelope
//	@Router			/bills/{id}/periods [get]
//	@Security		BearerAuth
func (h *BillHandler) GetBillPeriods(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	billID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID tagihan tidak valid")
	}

	periods, err := h.billUsecase.GetBillPeriods(c.Context(), userID, billID)
	if err != nil {
		return mapBillErr(c, err)
	}

	responses := make([]dto.BillPeriodResponse, len(periods))
	for i, p := range periods {
		responses[i] = *mapBillPeriodToResponse(p)
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// UpdateBill godoc
//
//	@Summary		Update a bill (is_active, reminder_days_before, due_day only)
//	@Tags			Bills
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Bill ID"
//	@Param			body	body		dto.UpdateBillRequest	true	"Bill data"
//	@Success		200		{object}	response.Envelope{data=dto.BillResponse}
//	@Failure		404		{object}	response.Envelope
//	@Router			/bills/{id} [patch]
//	@Security		BearerAuth
func (h *BillHandler) UpdateBill(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	billID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID tagihan tidak valid")
	}

	var req dto.UpdateBillRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	bill, err := h.billUsecase.UpdateBill(c.Context(), userID, billID, usecase.UpdateBillInput{
		IsActive:           req.IsActive,
		ReminderDaysBefore: req.ReminderDaysBefore,
		DueDay:             req.DueDay,
		EndPeriod:          req.EndPeriod,
	})
	if err != nil {
		return mapBillErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "ok", mapBillToResponse(&usecase.BillWithNextPeriod{Bill: bill}))
}

// PayBillPeriod godoc
//
//	@Summary		Mark a bill period as paid (creates an expense transaction)
//	@Tags			Bills
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Bill Period ID"
//	@Param			body	body		dto.PayBillPeriodRequest	true	"Payment data"
//	@Success		200		{object}	response.Envelope{data=dto.BillPeriodResponse}
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/bill-periods/{id}/pay [post]
//	@Security		BearerAuth
func (h *BillHandler) PayBillPeriod(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	billPeriodID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID periode tagihan tidak valid")
	}

	var req dto.PayBillPeriodRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID akun tidak valid")
	}

	period, _, err := h.billUsecase.PayBillPeriod(c.Context(), userID, billPeriodID, accountID, req.Amount, req.TransactionDate)
	if err != nil {
		return mapBillErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Tagihan berhasil ditandai dibayar", mapBillPeriodToResponse(period))
}

func mapBillPeriodToResponse(p *entity.BillPeriod) *dto.BillPeriodResponse {
	resp := &dto.BillPeriodResponse{
		ID:      p.ID.String(),
		BillID:  p.BillID.String(),
		Period:  p.Period,
		DueDate: p.DueDate.Format(dateLayoutConst),
		Status:  string(p.Status),
	}
	if p.TransactionID != nil {
		s := p.TransactionID.String()
		resp.TransactionID = &s
	}
	if p.PaidAt != nil {
		s := p.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PaidAt = &s
	}
	return resp
}

func mapBillToResponse(b *usecase.BillWithNextPeriod) *dto.BillResponse {
	resp := &dto.BillResponse{
		ID:                 b.ID.String(),
		Name:               b.Name,
		CategoryID:         b.CategoryID.String(),
		Amount:             b.Amount,
		DueDay:             b.DueDay,
		StartPeriod:        b.StartPeriod,
		EndPeriod:          b.EndPeriod,
		ReminderDaysBefore: b.ReminderDaysBefore,
		IsActive:           b.IsActive,
	}
	if b.NextPeriod != nil {
		resp.NextPeriod = mapBillPeriodToResponse(b.NextPeriod)
	}
	return resp
}
