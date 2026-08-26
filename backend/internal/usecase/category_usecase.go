package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUsecase(categoryRepo repository.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
	}
}

var defaultCategories = []struct {
	name string
	typ  entity.CategoryType
	icon string
	color string
}{
	// Income
	{"Gaji", entity.CategoryIncome, "briefcase", "#4CAF50"},
	{"Bonus", entity.CategoryIncome, "gift", "#8BC34A"},
	{"Hadiah", entity.CategoryIncome, "heart", "#FF9800"},
	{"Lainnya", entity.CategoryIncome, "plus-circle", "#9C27B0"},
	// Expense
	{"Makan & Minum", entity.CategoryExpense, "coffee", "#FF5722"},
	{"Transportasi", entity.CategoryExpense, "car", "#2196F3"},
	{"Belanja", entity.CategoryExpense, "shopping-cart", "#E91E63"},
	{"Tagihan & Utilitas", entity.CategoryExpense, "zap", "#FF6F00"},
	{"Kesehatan", entity.CategoryExpense, "heart", "#FF0000"},
	{"Hiburan", entity.CategoryExpense, "play-circle", "#673AB7"},
	{"Pendidikan", entity.CategoryExpense, "book", "#3F51B5"},
	{"Rumah Tangga", entity.CategoryExpense, "home", "#607D8B"},
	{"Lainnya", entity.CategoryExpense, "more-horizontal", "#795548"},
}

func (u *CategoryUsecase) SeedDefaultCategories(ctx context.Context, householdID, userID uuid.UUID) error {
	for _, cat := range defaultCategories {
		category := &entity.Category{
			BaseModel: entity.BaseModel{
				ID: uuid.New(),
			},
			HouseholdID: householdID,
			Name:        cat.name,
			Type:        cat.typ,
			Icon:        &cat.icon,
			Color:       &cat.color,
			IsArchived:  false,
			CreatedBy:   userID,
		}

		if err := u.categoryRepo.Create(ctx, category); err != nil {
			return err
		}
	}

	return nil
}

func (u *CategoryUsecase) CreateCategory(ctx context.Context, householdID, userID uuid.UUID, name, categoryType string, icon, color *string) (*entity.Category, error) {
	category := &entity.Category{
		BaseModel: entity.BaseModel{
			ID: uuid.New(),
		},
		HouseholdID: householdID,
		Name:        name,
		Type:        entity.CategoryType(categoryType),
		Icon:        icon,
		Color:       color,
		IsArchived:  false,
		CreatedBy:   userID,
	}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (u *CategoryUsecase) GetCategory(ctx context.Context, categoryID, householdID uuid.UUID) (*entity.Category, error) {
	category, err := u.categoryRepo.FindByID(ctx, categoryID, householdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

func (u *CategoryUsecase) ListCategories(ctx context.Context, householdID uuid.UUID, categoryType string, includeArchived bool) ([]*entity.Category, error) {
	categories, err := u.categoryRepo.ListByHousehold(ctx, householdID, categoryType, includeArchived)
	if err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []*entity.Category{}
	}

	return categories, nil
}

// UpdateCategory tidak menerima parameter type — spec melarang ubah type via PATCH umum
// (perlu cek category_id dipakai transaksi atau tidak, baru tersedia di Phase 06).
func (u *CategoryUsecase) UpdateCategory(ctx context.Context, categoryID, householdID uuid.UUID, name *string, icon, color *string, isArchived *bool) (*entity.Category, error) {
	category, err := u.categoryRepo.FindByID(ctx, categoryID, householdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	if name != nil {
		category.Name = *name
	}
	if icon != nil {
		category.Icon = icon
	}
	if color != nil {
		category.Color = color
	}
	if isArchived != nil {
		category.IsArchived = *isArchived
	}

	if err := u.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (u *CategoryUsecase) ArchiveCategory(ctx context.Context, categoryID, householdID uuid.UUID) error {
	_, err := u.categoryRepo.FindByID(ctx, categoryID, householdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	return u.categoryRepo.Archive(ctx, categoryID, householdID)
}

func (u *CategoryUsecase) UnarchiveCategory(ctx context.Context, categoryID, householdID uuid.UUID) error {
	_, err := u.categoryRepo.FindByID(ctx, categoryID, householdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	return u.categoryRepo.Unarchive(ctx, categoryID, householdID)
}
