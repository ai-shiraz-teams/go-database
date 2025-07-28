package repository

import (
	"context"

	"go-database/pkg/domain"

	"gorm.io/gorm"
)

// IBaseRepository defines the interface for basic database operations using GORM.
// This is used internally by the UnitOfWork implementation and provides type-safe
// database operations for any entity implementing IBaseModel.
type IBaseRepository[T domain.IBaseModel] interface {
	// Create inserts a new entity
	Create(ctx context.Context, entity T) (T, error)

	// FindByID retrieves an entity by its ID
	FindByID(ctx context.Context, id uint) (T, error)

	// FindOne retrieves a single entity matching the query
	FindOne(ctx context.Context, query *gorm.DB) (T, error)

	// FindAll retrieves all entities matching the query
	FindAll(ctx context.Context, query *gorm.DB) ([]T, error)

	// FindWithPagination retrieves entities with pagination
	FindWithPagination(ctx context.Context, query *gorm.DB, offset, limit int) ([]T, int64, error)

	// Update modifies an existing entity
	Update(ctx context.Context, entity T) (T, error)

	// Delete performs a soft delete
	Delete(ctx context.Context, query *gorm.DB) error

	// HardDelete permanently removes entities
	HardDelete(ctx context.Context, query *gorm.DB) error

	// Restore recovers soft-deleted entities
	Restore(ctx context.Context, query *gorm.DB) error

	// Count returns the number of entities matching the query
	Count(ctx context.Context, query *gorm.DB) (int64, error)

	// Exists checks if any entity matches the query
	Exists(ctx context.Context, query *gorm.DB) (bool, error)

	// GetDB returns a clone of the database connection
	GetDB() *gorm.DB

	// WithTransaction returns a repository instance using the provided transaction
	WithTransaction(tx *gorm.DB) IBaseRepository[T]
}

// BaseRepository provides GORM-based implementation of IBaseRepository.
// It handles all basic database operations for entities implementing IBaseModel.
type BaseRepository[T domain.IBaseModel] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new BaseRepository instance
func NewBaseRepository[T domain.IBaseModel](db *gorm.DB) IBaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

// Create inserts a new entity and returns it with populated fields
func (r *BaseRepository[T]) Create(ctx context.Context, entity T) (T, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindByID retrieves an entity by its ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uint) (T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindOne retrieves a single entity matching the query
func (r *BaseRepository[T]) FindOne(ctx context.Context, query *gorm.DB) (T, error) {
	var entity T
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindAll retrieves all entities matching the query
func (r *BaseRepository[T]) FindAll(ctx context.Context, query *gorm.DB) ([]T, error) {
	var entities []T
	if err := query.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// FindWithPagination retrieves entities with pagination and returns total count
func (r *BaseRepository[T]) FindWithPagination(ctx context.Context, query *gorm.DB, offset, limit int) ([]T, int64, error) {
	var entities []T
	var total int64

	// Count total records first
	countQuery := query.Session(&gorm.Session{NewDB: true})
	if err := countQuery.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.WithContext(ctx).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Update modifies an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity T) (T, error) {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// Delete performs a soft delete by setting deleted_at
func (r *BaseRepository[T]) Delete(ctx context.Context, query *gorm.DB) error {
	return query.WithContext(ctx).Delete(new(T)).Error
}

// HardDelete permanently removes entities from the database
func (r *BaseRepository[T]) HardDelete(ctx context.Context, query *gorm.DB) error {
	return query.WithContext(ctx).Unscoped().Delete(new(T)).Error
}

// Restore recovers soft-deleted entities by clearing deleted_at
func (r *BaseRepository[T]) Restore(ctx context.Context, query *gorm.DB) error {
	return query.WithContext(ctx).Unscoped().Model(new(T)).Update("deleted_at", nil).Error
}

// Count returns the number of entities matching the query
func (r *BaseRepository[T]) Count(ctx context.Context, query *gorm.DB) (int64, error) {
	var count int64
	err := query.WithContext(ctx).Model(new(T)).Count(&count).Error
	return count, err
}

// Exists checks if any entity matches the query
func (r *BaseRepository[T]) Exists(ctx context.Context, query *gorm.DB) (bool, error) {
	var count int64
	err := query.WithContext(ctx).Model(new(T)).Limit(1).Count(&count).Error
	return count > 0, err
}

// GetDB returns a clone of the database connection
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db.Session(&gorm.Session{NewDB: true})
}

// WithTransaction returns a repository instance using the provided transaction
func (r *BaseRepository[T]) WithTransaction(tx *gorm.DB) IBaseRepository[T] {
	return &BaseRepository[T]{
		db: tx,
	}
}

// Compile-time check to ensure BaseRepository implements IBaseRepository
var _ IBaseRepository[*domain.User] = (*BaseRepository[*domain.User])(nil)
