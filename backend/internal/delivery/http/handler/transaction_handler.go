package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/response"
	"homeapp/internal/pkg/storage"
	"homeapp/internal/repository"
	"homeapp/internal/usecase"
)

type TransactionHandler struct {
	transactionUsecase *usecase.TransactionUsecase
	validator          *validator.Validate
	store              storage.Storage
}

func NewTransactionHandler(transactionUsecase *usecase.TransactionUsecase, validator *validator.Validate, store storage.Storage) *TransactionHandler {
	return &TransactionHandler{
		transactionUsecase: transactionUsecase,
		validator:          validator,
		store:              store,
	}
}

func mapTransactionErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "Data tidak ditemukan di rumah tangga ini")
	case errors.Is(err, apperror.ErrCategoryRequired),
		errors.Is(err, apperror.ErrCategoryTypeMismatch),
		errors.Is(err, apperror.ErrDestinationRequired),
		errors.Is(err, apperror.ErrTransferSameAccount),
		errors.Is(err, apperror.ErrAccountInactive):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrBillPeriodPaidConflict):
		return response.Error(c, fiber.StatusConflict, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memproses transaksi")
	}
}

// CreateTransaction godoc
//
//	@Summary		Create a new transaction
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateTransactionRequest	true	"Transaction data"
//	@Success		201		{object}	response.Envelope{data=dto.TransactionResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/transactions [post]
//	@Security		BearerAuth
func (h *TransactionHandler) CreateTransaction(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	var req dto.CreateTransactionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	input, err := parseTransactionInput(req.Type, req.AccountID, req.DestinationAccountID, req.CategoryID, req.Amount, req.AdminFee, req.Description, req.TransactionDate, req.AttachmentURL, req.GoalID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format ID atau tanggal tidak valid")
	}

	item, err := h.transactionUsecase.CreateTransaction(c.Context(), userID, *input)
	if err != nil {
		return mapTransactionErr(c, err)
	}

	trx := mapTransactionToResponse(item)
	trx.AttachmentURL = h.resolveAttachmentURL(c, trx.AttachmentURL)
	return response.Success(c, fiber.StatusCreated, "Transaksi berhasil dibuat", trx)
}

// ListTransactions godoc
//
//	@Summary		List transactions
//	@Tags			Transactions
//	@Produce		json
//	@Param			type		query	string	false	"Filter by type"
//	@Param			account_id	query	string	false	"Filter by account ID"
//	@Param			category_id	query	string	false	"Filter by category ID"
//	@Param			created_by	query	string	false	"Filter by creator user ID"
//	@Param			date_from	query	string	false	"Filter from date (YYYY-MM-DD)"
//	@Param			date_to		query	string	false	"Filter to date (YYYY-MM-DD)"
//	@Param			page		query	int		false	"Page number"
//	@Param			limit		query	int		false	"Items per page"
//	@Success		200			{object}	response.Envelope{data=dto.TransactionListResponse}
//	@Router			/transactions [get]
//	@Security		BearerAuth
func (h *TransactionHandler) ListTransactions(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	input := usecase.ListTransactionsInput{
		Type: c.Query("type", ""),
	}

	if v := c.Query("account_id", ""); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			input.AccountID = &id
		}
	}
	if v := c.Query("category_id", ""); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			input.CategoryID = &id
		}
	}
	if v := c.Query("created_by", ""); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			input.CreatedBy = &id
		}
	}
	if v := c.Query("date_from", ""); v != "" {
		if d, err := time.Parse(dateLayoutConst, v); err == nil {
			input.DateFrom = &d
		}
	}
	if v := c.Query("date_to", ""); v != "" {
		if d, err := time.Parse(dateLayoutConst, v); err == nil {
			input.DateTo = &d
		}
	}

	page := 1
	if v := c.Query("page", ""); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			page = p
		}
	}
	limit := 20
	if v := c.Query("limit", ""); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			limit = l
		}
	}
	input.Page = page
	input.Limit = limit

	items, total, err := h.transactionUsecase.ListTransactions(c.Context(), userID, input)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat transaksi")
	}

	responses := make([]dto.TransactionResponse, len(items))
	for i, item := range items {
		responses[i] = *mapTransactionToResponse(item)
		responses[i].AttachmentURL = h.resolveAttachmentURL(c, responses[i].AttachmentURL)
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.TransactionListResponse{
		Items: responses,
		Pagination: dto.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetTransaction godoc
//
//	@Summary		Get transaction detail
//	@Tags			Transactions
//	@Produce		json
//	@Param			id	path	string	true	"Transaction ID"
//	@Success		200	{object}	response.Envelope{data=dto.TransactionResponse}
//	@Failure		404	{object}	response.Envelope
//	@Router			/transactions/{id} [get]
//	@Security		BearerAuth
func (h *TransactionHandler) GetTransaction(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID transaksi tidak valid")
	}

	item, err := h.transactionUsecase.GetTransaction(c.Context(), userID, transactionID)
	if err != nil {
		return mapTransactionErr(c, err)
	}

	trx := mapTransactionToResponse(item)
	trx.AttachmentURL = h.resolveAttachmentURL(c, trx.AttachmentURL)

	return response.Success(c, fiber.StatusOK, "ok", trx)
}

// UpdateTransaction godoc
//
//	@Summary		Update a transaction
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Transaction ID"
//	@Param			body	body		dto.UpdateTransactionRequest	true	"Transaction data"
//	@Success		200		{object}	response.Envelope{data=dto.TransactionResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/transactions/{id} [patch]
//	@Security		BearerAuth
func (h *TransactionHandler) UpdateTransaction(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID transaksi tidak valid")
	}

	var req dto.UpdateTransactionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	existing, err := h.transactionUsecase.GetTransaction(c.Context(), userID, transactionID)
	if err != nil {
		return mapTransactionErr(c, err)
	}

	txType := string(existing.Type)
	if req.Type != nil {
		txType = *req.Type
	}
	accountID := existing.AccountID.String()
	if req.AccountID != nil {
		accountID = *req.AccountID
	}
	destAccountID := uuidPtrToStrPtr(existing.DestinationAccountID)
	if req.DestinationAccountID != nil {
		destAccountID = req.DestinationAccountID
	}
	categoryID := uuidPtrToStrPtr(existing.CategoryID)
	if req.CategoryID != nil {
		categoryID = req.CategoryID
	}
	amount := existing.Amount
	if req.Amount != nil {
		amount = *req.Amount
	}
	adminFee := existing.AdminFee
	if req.AdminFee != nil {
		adminFee = *req.AdminFee
	}
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	transactionDate := existing.TransactionDate.Format(dateLayoutConst)
	if req.TransactionDate != nil {
		transactionDate = *req.TransactionDate
	}
	attachmentURL := existing.AttachmentURL
	if req.AttachmentURL != nil {
		attachmentURL = req.AttachmentURL
	}
	goalID := uuidPtrToStrPtr(existing.GoalID)
	if req.GoalID != nil {
		goalID = req.GoalID
	}

	input, err := parseTransactionInput(txType, accountID, destAccountID, categoryID, amount, adminFee, description, transactionDate, attachmentURL, goalID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format ID atau tanggal tidak valid")
	}

	item, err := h.transactionUsecase.UpdateTransaction(c.Context(), userID, transactionID, *input)
	if err != nil {
		return mapTransactionErr(c, err)
	}

	trx := mapTransactionToResponse(item)
	trx.AttachmentURL = h.resolveAttachmentURL(c, trx.AttachmentURL)
	return response.Success(c, fiber.StatusOK, "ok", trx)
}

// DeleteTransaction godoc
//
//	@Summary		Delete a transaction
//	@Tags			Transactions
//	@Produce		json
//	@Param			id	path	string	true	"Transaction ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/transactions/{id} [delete]
//	@Security		BearerAuth
func (h *TransactionHandler) DeleteTransaction(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID transaksi tidak valid")
	}

	if err := h.transactionUsecase.DeleteTransaction(c.Context(), userID, transactionID); err != nil {
		return mapTransactionErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Transaksi berhasil dihapus", nil)
}

// GetQuickSelect godoc
//
//	@Summary		Get last-used account/category for quick-select
//	@Tags			Transactions
//	@Produce		json
//	@Success		200	{object}	response.Envelope{data=dto.QuickSelectResponse}
//	@Router			/transactions/quick-select [get]
//	@Security		BearerAuth
func (h *TransactionHandler) GetQuickSelect(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	accountID, categoryID, err := h.transactionUsecase.GetQuickSelect(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat quick-select")
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.QuickSelectResponse{
		LastAccountID:  uuidPtrToStrPtr(accountID),
		LastCategoryID: uuidPtrToStrPtr(categoryID),
	})
}

const dateLayoutConst = "2006-01-02"

func parseTransactionInput(txType, accountID string, destAccountID, categoryID *string, amount, adminFee float64, description *string, transactionDate string, attachmentURL *string, goalID *string) (*usecase.TransactionInput, error) {
	accID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, err
	}

	input := &usecase.TransactionInput{
		Type:            txType,
		AccountID:       accID,
		Amount:          amount,
		AdminFee:        adminFee,
		Description:     description,
		TransactionDate: transactionDate,
		AttachmentURL:   attachmentURL,
	}

	if destAccountID != nil {
		destID, err := uuid.Parse(*destAccountID)
		if err != nil {
			return nil, err
		}
		input.DestinationAccountID = &destID
	}

	if categoryID != nil {
		catID, err := uuid.Parse(*categoryID)
		if err != nil {
			return nil, err
		}
		input.CategoryID = &catID
	}

	if goalID != nil && *goalID != "" {
		gID, err := uuid.Parse(*goalID)
		if err != nil {
			return nil, err
		}
		input.GoalID = &gID
	}

	if _, err := time.Parse(dateLayoutConst, transactionDate); err != nil {
		return nil, err
	}

	return input, nil
}

func uuidPtrToStrPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// resolveAttachmentURL mengubah attachment_url ter-stored (yang berisi path S3
// tanpa autentikasi) menjadi URL yang bisa diakses client. Untuk S3 ini diganti
// dengan presigned GET URL; untuk local store dibiarkan apa adanya.
func (h *TransactionHandler) resolveAttachmentURL(c fiber.Ctx, url *string) *string {
	if url == nil || *url == "" || h.store == nil {
		return url
	}
	resolved, err := h.store.ReadURL(c.Context(), *url)
	if err != nil {
		return url
	}
	return &resolved
}

func mapTransactionToResponse(item *repository.TransactionListItem) *dto.TransactionResponse {
	return &dto.TransactionResponse{
		ID:                   item.ID.String(),
		HouseholdID:          item.HouseholdID.String(),
		Type:                 string(item.Type),
		AccountID:            item.AccountID.String(),
		AccountName:          item.AccountName,
		DestinationAccountID: uuidPtrToStrPtr(item.DestinationAccountID),
		CategoryID:           uuidPtrToStrPtr(item.CategoryID),
		CategoryName:         item.CategoryName,
		Amount:               item.Amount,
		AdminFee:             item.AdminFee,
		Description:          item.Description,
		TransactionDate:      item.TransactionDate.Format(dateLayoutConst),
		AttachmentURL:        item.AttachmentURL,
		GoalID:               uuidPtrToStrPtr(item.GoalID),
		BillPeriodID:         uuidPtrToStrPtr(item.BillPeriodID),
		CreatedBy:            item.CreatedBy.String(),
		CreatedByName:        item.CreatedByName,
		CreatedAt:            item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
