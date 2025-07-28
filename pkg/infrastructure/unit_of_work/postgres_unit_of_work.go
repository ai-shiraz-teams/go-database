package unit_of_work

import (
	"context"
	"fmt"

	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/query"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/unit_of_work"

	"gorm.io/gorm"
)

// PostgresUnitOfWork provides a GORM-based implementation of IUnitOfWork for PostgreSQL.
// It operates directly on GORM database connections and maintains transaction safety
// across all operations without any repository dependencies.
type PostgresUnitOfWork[T types.IBaseModel] struct {
	db            *gorm.DB
	filterApplier *FilterApplier
	tx            *gorm.DB // Current transaction, nil if not in transaction
}

// NewPostgresUnitOfWork creates a new PostgreSQL UnitOfWork instance
func NewPostgresUnitOfWork[T types.IBaseModel](db *gorm.DB) unit_of_work.IUnitOfWork[T] {
	return &PostgresUnitOfWork[T]{
		db:            db,
		filterApplier: NewFilterApplier(),
	}
}

// getDB returns the current database connection (transaction if active, otherwise main db)
func (uow *PostgresUnitOfWork[T]) getDB() *gorm.DB {
	if uow.tx != nil {
		return uow.tx
	}
	return uow.db
}

// Transaction management

// BeginTransaction starts a new database transaction
func (uow *PostgresUnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already in progress")
	}

	tx := uow.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	uow.tx = tx
	return nil
}

// CommitTransaction commits the current transaction
func (uow *PostgresUnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	err := uow.tx.Commit().Error
	uow.tx = nil
	return err
}

// RollbackTransaction rolls back the current transaction
func (uow *PostgresUnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	if uow.tx != nil {
		uow.tx.Rollback()
		uow.tx = nil
	}
}

// Basic queries

// FindAll retrieves all entities (excluding soft-deleted by default)
func (uow *PostgresUnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	db := uow.getDB()
	if err := db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// FindAllWithPagination retrieves entities with pagination support and returns total count
func (uow *PostgresUnitOfWork[T]) FindAllWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, uint, error) {
	db := uow.getDB()

	// Start with base query
	baseQuery := db.Model(new(T))

	// Apply QueryParams filters, sorting, etc.
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)

	// Get pagination values
	offset := query.Offset
	limit := query.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	// Count total records first
	var total int64
	countQuery := filteredQuery.Session(&gorm.Session{NewDB: true})
	if err := countQuery.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var entities []T
	if err := filteredQuery.WithContext(ctx).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, uint(total), nil
}

// FindOne retrieves a single entity matching the provided filter
func (uow *PostgresUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	var entity T
	db := uow.getDB()
	if err := db.WithContext(ctx).Where(filter).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindOneById retrieves a single entity by its ID
func (uow *PostgresUnitOfWork[T]) FindOneById(ctx context.Context, id uint) (T, error) {
	var entity T
	db := uow.getDB()
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindOneByIdentifier retrieves a single entity using the IIdentifier filter system
func (uow *PostgresUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// Mutation operations

// Insert creates a new entity and returns the created entity with populated fields
func (uow *PostgresUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	db := uow.getDB()
	if err := db.WithContext(ctx).Create(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// Update modifies entities matching the identifier with the provided entity data
func (uow *PostgresUnitOfWork[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {
	// First verify the entity exists
	_, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	// Update the entity (this preserves the ID and other fields)
	db := uow.getDB()
	if err := db.WithContext(ctx).Save(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// Delete performs a logical operation (soft-delete by default)
func (uow *PostgresUnitOfWork[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	return query.WithContext(ctx).Delete(new(T)).Error
}

// Soft-delete lifecycle management

// SoftDelete performs soft deletion by setting DeletedAt timestamp
func (uow *PostgresUnitOfWork[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	// First find the entity
	entity, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	// Perform soft delete
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// HardDelete permanently removes entities from the database
func (uow *PostgresUnitOfWork[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	// First find the entity (including soft-deleted ones)
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()
	var entity T
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}

	// Perform hard delete
	if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// GetTrashed retrieves all soft-deleted entities
func (uow *PostgresUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	db := uow.getDB()
	var entities []T
	if err := db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// GetTrashedWithPagination retrieves soft-deleted entities with pagination
func (uow *PostgresUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, params *query.QueryParams[T]) ([]T, uint, error) {
	// Force only deleted records
	if params == nil {
		params = query.NewQueryParams[T]()
	}
	params.OnlyDeleted = true
	return uow.FindAllWithPagination(ctx, params)
}

// Restore recovers soft-deleted entities by clearing their DeletedAt timestamp
func (uow *PostgresUnitOfWork[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()

	// First find the soft-deleted entity
	var entity T
	if err := query.WithContext(ctx).Where("deleted_at IS NOT NULL").First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}

	// Restore the entity by setting deleted_at to NULL
	if err := query.WithContext(ctx).Update("deleted_at", nil).Error; err != nil {
		var zero T
		return zero, err
	}

	// Return the restored entity by finding it again
	var restoredEntity T
	if err := db.WithContext(ctx).First(&restoredEntity, uint(entity.GetID())).Error; err != nil {
		var zero T
		return zero, err
	}

	return restoredEntity, nil
}

// RestoreAll recovers all soft-deleted entities of type T
func (uow *PostgresUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	db := uow.getDB()
	return db.WithContext(ctx).Model(new(T)).Unscoped().Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error
}

// Bulk operations

// BulkInsert creates multiple entities in a single operation
func (uow *PostgresUnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	db := uow.getDB()
	if err := db.WithContext(ctx).Create(&entities).Error; err != nil {
		return nil, err
	}

	return entities, nil
}

// BulkUpdate modifies multiple entities in a single operation
func (uow *PostgresUnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	db := uow.getDB()

	// GORM doesn't have a direct bulk update, so we update each entity
	// In a transaction, this is still efficient
	for i, entity := range entities {
		if err := db.WithContext(ctx).Save(&entity).Error; err != nil {
			return nil, err
		}
		entities[i] = entity
	}

	return entities, nil
}

// BulkSoftDelete soft-deletes multiple entities identified by the provided identifiers
func (uow *PostgresUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	db := uow.getDB()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](db, identifier)
		if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
			return err
		}
	}

	return nil
}

// BulkHardDelete permanently removes multiple entities identified by the provided identifiers
func (uow *PostgresUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	db := uow.getDB()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()
		if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
			return err
		}
	}

	return nil
}

// Utility operations

// ResolveIDByUniqueField finds the ID of an entity by searching a unique field
func (uow *PostgresUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (uint, error) {
	var entity T
	db := uow.getDB()

	if err := db.WithContext(ctx).Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value).First(&entity).Error; err != nil {
		return 0, err
	}

	return uint(entity.GetID()), nil
}

// Count returns the total number of entities matching the query parameters
func (uow *PostgresUnitOfWork[T]) Count(ctx context.Context, query *query.QueryParams[T]) (int64, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)

	var count int64
	if err := filteredQuery.WithContext(ctx).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists checks if any entity matches the provided identifier
func (uow *PostgresUnitOfWork[T]) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)

	var count int64
	if err := query.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
