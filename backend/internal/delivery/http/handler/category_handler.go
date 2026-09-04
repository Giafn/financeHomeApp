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
	"homeapp/internal/repository"
	"homeapp/internal/usecase"
)

type CategoryHandler struct {
	categoryUsecase      *usecase.CategoryUsecase
	householdRepository  repository.HouseholdRepository
	validator            *validator.Validate
}

func NewCategoryHandler(
	categoryUsecase *usecase.CategoryUsecase,
	householdRepository repository.HouseholdRepository,
	validator *validator.Validate,
) *CategoryHandler {
	return &CategoryHandler{
		categoryUsecase:      categoryUsecase,
		householdRepository:  householdRepository,
		validator:            validator,
	}
}

func mapCategoryErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "Kategori atau kategori induk tidak ditemukan")
	case errors.Is(err, apperror.ErrCategoryParentTypeMismatch),
		errors.Is(err, apperror.ErrCategoryNestingTooDeep),
		errors.Is(err, apperror.ErrCategoryHasChildren):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memproses kategori")
	}
}

// CreateCategory godoc
//
//	@Summary		Create a new category
//	@Description	Create a new transaction category for the household
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateCategoryRequest	true	"Category data"
//	@Success		201		{object}	response.SuccessResponse{data=dto.CategoryResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/categories [post]
//	@Security		BearerAuth
func (h *CategoryHandler) CreateCategory(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	member, err := h.householdRepository.FindMemberByUserID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Pengguna tidak tergabung dalam rumah tangga")
	}

	var req dto.CreateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "ID kategori induk tidak valid")
		}
		parentID = &id
	}

	category, err := h.categoryUsecase.CreateCategory(
		c.Context(),
		member.HouseholdID,
		userID,
		req.Name,
		req.Type,
		req.Icon,
		req.Color,
		parentID,
	)
	if err != nil {
		return mapCategoryErr(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Kategori berhasil dibuat", mapCategoryToResponse(category))
}

// ListCategories godoc
//
//	@Summary		List categories
//	@Description	Get list of categories for the household
//	@Tags			Categories
//	@Produce		json
//	@Param			type				query		string	false	"Filter by type (income/expense)"
//	@Param			include_archived	query		bool	false	"Include archived categories (default: false)"
//	@Success		200					{object}	response.Envelope
//	@Failure		401					{object}	response.Envelope
//	@Router			/categories [get]
//	@Security		BearerAuth
func (h *CategoryHandler) ListCategories(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)

	member, err := h.householdRepository.FindMemberByUserID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Pengguna tidak tergabung dalam rumah tangga")
	}

	categoryType := c.Query("type", "")
	includeArchived := c.Query("include_archived", "false") == "true"

	categories, err := h.categoryUsecase.ListCategories(c.Context(), member.HouseholdID, categoryType, includeArchived)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal memuat kategori")
	}

	responses := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = *mapCategoryToResponse(cat)
	}

	return response.Success(c, fiber.StatusOK, "ok", responses)
}

// UpdateCategory godoc
//
//	@Summary		Update a category
//	@Description	Update category details (name, icon, color, archive status)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Category ID"
//	@Param			body	body		dto.UpdateCategoryRequest	true	"Category data"
//	@Success		200		{object}	response.SuccessResponse{data=dto.CategoryResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/categories/{id} [patch]
//	@Security		BearerAuth
func (h *CategoryHandler) UpdateCategory(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	categoryIDStr := c.Params("id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID kategori tidak valid")
	}

	member, err := h.householdRepository.FindMemberByUserID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Pengguna tidak tergabung dalam rumah tangga")
	}

	var req dto.UpdateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "ID kategori induk tidak valid")
		}
		parentID = &id
	}

	category, err := h.categoryUsecase.UpdateCategory(
		c.Context(),
		categoryID,
		member.HouseholdID,
		req.Name,
		req.Icon,
		req.Color,
		req.IsArchived,
		parentID,
	)
	if err != nil {
		return mapCategoryErr(c, err)
	}

	return response.Success(c, fiber.StatusOK, "ok", mapCategoryToResponse(category))
}

// ArchiveCategory godoc
//
//	@Summary		Archive a category
//	@Description	Archive a category (soft delete)
//	@Tags			Categories
//	@Produce		json
//	@Param			id	path	string	true	"Category ID"
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/categories/{id} [delete]
//	@Security		BearerAuth
func (h *CategoryHandler) ArchiveCategory(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	categoryIDStr := c.Params("id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID kategori tidak valid")
	}

	member, err := h.householdRepository.FindMemberByUserID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Pengguna tidak tergabung dalam rumah tangga")
	}

	err = h.categoryUsecase.ArchiveCategory(c.Context(), categoryID, member.HouseholdID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, fiber.StatusNotFound, "Kategori tidak ditemukan")
		}
		return response.Error(c, fiber.StatusInternalServerError, "Gagal arsipkan kategori")
	}

	return response.Success(c, fiber.StatusOK, "Kategori berhasil diarsipkan", nil)
}

// UnarchiveCategory godoc
//
//	@Summary		Unarchive a category
//	@Description	Restore an archived category
//	@Tags			Categories
//	@Produce		json
//	@Param			id	path	string	true	"Category ID"
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/categories/{id}/unarchive [post]
//	@Security		BearerAuth
func (h *CategoryHandler) UnarchiveCategory(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(uuid.UUID)
	categoryIDStr := c.Params("id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ID kategori tidak valid")
	}

	member, err := h.householdRepository.FindMemberByUserID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Pengguna tidak tergabung dalam rumah tangga")
	}

	err = h.categoryUsecase.UnarchiveCategory(c.Context(), categoryID, member.HouseholdID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, fiber.StatusNotFound, "Kategori tidak ditemukan")
		}
		return response.Error(c, fiber.StatusInternalServerError, "Gagal aktifkan kembali kategori")
	}

	return response.Success(c, fiber.StatusOK, "Kategori berhasil diaktifkan kembali", nil)
}

func mapCategoryToResponse(category *entity.Category) *dto.CategoryResponse {
	resp := &dto.CategoryResponse{
		ID:          category.ID.String(),
		HouseholdID: category.HouseholdID.String(),
		Name:        category.Name,
		Type:        string(category.Type),
		Icon:        category.Icon,
		Color:       category.Color,
		IsArchived:  category.IsArchived,
		CreatedBy:   category.CreatedBy.String(),
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if category.ParentID != nil {
		id := category.ParentID.String()
		resp.ParentID = &id
	}
	return resp
}
