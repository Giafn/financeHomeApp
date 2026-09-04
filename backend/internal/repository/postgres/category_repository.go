package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(category).Error
}

func (r *CategoryRepository) FindByID(ctx context.Context, id, householdID uuid.UUID) (*entity.Category, error) {
	var category entity.Category
	if err := r.db.WithContext(ctx).Where("id = ? AND household_id = ? AND deleted_at IS NULL", id, householdID).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) ListByHousehold(ctx context.Context, householdID uuid.UUID, categoryType string, includeArchived bool) ([]*entity.Category, error) {
	var categories []*entity.Category
	query := r.db.WithContext(ctx).Where("household_id = ? AND deleted_at IS NULL", householdID)

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	if !includeArchived {
		query = query.Where("is_archived = false")
	}

	if err := query.Order("CASE WHEN type = 'income' THEN 0 ELSE 1 END, name").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Model(category).
		Select("name", "icon", "color", "is_archived", "parent_id").
		Updates(category).Error
}

func (r *CategoryRepository) Archive(ctx context.Context, id, householdID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Category{}).
		Where("id = ? AND household_id = ?", id, householdID).
		Update("is_archived", true).Error
}

func (r *CategoryRepository) Unarchive(ctx context.Context, id, householdID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Category{}).
		Where("id = ? AND household_id = ?", id, householdID).
		Update("is_archived", false).Error
}

func (r *CategoryRepository) HasChildren(ctx context.Context, id, householdID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Category{}).
		Where("parent_id = ? AND household_id = ? AND deleted_at IS NULL", id, householdID).
		Count(&count).Error
	return count > 0, err
}
