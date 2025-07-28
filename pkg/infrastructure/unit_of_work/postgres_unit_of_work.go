package unit_of_work

import (
	"context"
	"fmt"

	"go-database/pkg/domain"
	"go-database/pkg/infrastructure/repository"

	"gorm.io/gorm"
)

// PostgresUnitOfWork provides a GORM-based implementation of IUnitOfWork for PostgreSQL.
// It uses composition with BaseRepository to handle database operations and maintains
// transaction safety across all operations.
type PostgresUnitOfWork[T domain.IBaseModel] struct {
	db            *gorm.DB
	baseRepo      repository.IBaseRepository[T]
	filterApplier *FilterApplier
	tx            *gorm.DB // Current transaction, nil if not in transaction
}

// NewPostgresUnitOfWork creates a new PostgreSQL UnitOfWork instance
func NewPostgresUnitOfWork[T domain.IBaseModel](db *gorm.DB) domain.IUnitOfWork[T] {
	return &PostgresUnitOfWork[T]{
		db:            db,
		baseRepo:      repository.NewBaseRepository[T](db),
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

// getRepo returns a repository instance using the current database connection
func (uow *PostgresUnitOfWork[T]) getRepo() repository.IBaseRepository[T] {
	currentDB := uow.getDB()
	if currentDB != uow.db {
		return uow.baseRepo.WithTransaction(currentDB)
	}
	return uow.baseRepo
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
	repo := uow.getRepo()
	query := repo.GetDB().Model(new(T))
	return repo.FindAll(ctx, query)
}

// FindAllWithPagination retrieves entities with pagination support and returns total count
func (uow *PostgresUnitOfWork[T]) FindAllWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	repo := uow.getRepo()

	// Start with base query
	baseQuery := repo.GetDB().Model(new(T))

	// Apply QueryParams filters, sorting, etc.
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)

	// Get pagination values
	offset := query.Offset
	limit := query.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	// Execute paginated query
	entities, total, err := repo.FindWithPagination(ctx, filteredQuery, offset, limit)
	return entities, uint(total), err
}

// FindOne retrieves a single entity matching the provided filter
func (uow *PostgresUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	repo := uow.getRepo()
	query := repo.GetDB().Model(new(T)).Where(filter)
	return repo.FindOne(ctx, query)
}

// FindOneById retrieves a single entity by its ID
func (uow *PostgresUnitOfWork[T]) FindOneById(ctx context.Context, id uint) (T, error) {
	repo := uow.getRepo()
	return repo.FindByID(ctx, id)
}

// FindOneByIdentifier retrieves a single entity using the IIdentifier filter system
func (uow *PostgresUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
	return repo.FindOne(ctx, query)
}

// Mutation operations

// Insert creates a new entity and returns the created entity with populated fields
func (uow *PostgresUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	repo := uow.getRepo()
	return repo.Create(ctx, entity)
}

// Update modifies entities matching the identifier with the provided entity data
func (uow *PostgresUnitOfWork[T]) Update(ctx context.Context, identifier domain.IIdentifier, entity T) (T, error) {
	repo := uow.getRepo()

	// First verify the entity exists
	_, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	// Update the entity (this preserves the ID and other fields)
	return repo.Update(ctx, entity)
}

// Delete performs a logical operation (soft-delete by default)
func (uow *PostgresUnitOfWork[T]) Delete(ctx context.Context, identifier domain.IIdentifier) error {
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
	return repo.Delete(ctx, query)
}

// Soft-delete lifecycle management

// SoftDelete performs soft deletion by setting DeletedAt timestamp
func (uow *PostgresUnitOfWork[T]) SoftDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	// First find the entity
	entity, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	// Perform soft delete
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
	if err := repo.Delete(ctx, query); err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// HardDelete permanently removes entities from the database
func (uow *PostgresUnitOfWork[T]) HardDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	// First find the entity (including soft-deleted ones)
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier).Unscoped()
	entity, err := repo.FindOne(ctx, query)
	if err != nil {
		var zero T
		return zero, err
	}

	// Perform hard delete
	if err := repo.HardDelete(ctx, query); err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// GetTrashed retrieves all soft-deleted entities
func (uow *PostgresUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	repo := uow.getRepo()
	query := repo.GetDB().Model(new(T)).Unscoped().Where("deleted_at IS NOT NULL")
	return repo.FindAll(ctx, query)
}

// GetTrashedWithPagination retrieves soft-deleted entities with pagination
func (uow *PostgresUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	// Force only deleted records
	if query == nil {
		query = domain.NewQueryParams[T]()
	}
	query.OnlyDeleted = true
	return uow.FindAllWithPagination(ctx, query)
}

// Restore recovers soft-deleted entities by clearing their DeletedAt timestamp
func (uow *PostgresUnitOfWork[T]) Restore(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier).Unscoped()

	// First find the soft-deleted entity
	entity, err := repo.FindOne(ctx, query.Where("deleted_at IS NOT NULL"))
	if err != nil {
		var zero T
		return zero, err
	}

	// Restore the entity
	if err := repo.Restore(ctx, query); err != nil {
		var zero T
		return zero, err
	}

	// Return the restored entity
	return repo.FindByID(ctx, uint(entity.GetID()))
}

// RestoreAll recovers all soft-deleted entities of type T
func (uow *PostgresUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	repo := uow.getRepo()
	query := repo.GetDB().Model(new(T)).Unscoped().Where("deleted_at IS NOT NULL")
	return repo.Restore(ctx, query)
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
func (uow *PostgresUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	repo := uow.getRepo()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
		if err := repo.Delete(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// BulkHardDelete permanently removes multiple entities identified by the provided identifiers
func (uow *PostgresUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	repo := uow.getRepo()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
		if err := repo.HardDelete(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// Utility operations

// ResolveIDByUniqueField finds the ID of an entity by searching a unique field
func (uow *PostgresUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (uint, error) {
	var entity T
	db := uow.getDB()

	if err := db.WithContext(ctx).Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value).First(&entity).Error; err != nil {
		return 0, err
	}

	return uint(entity.GetID()), nil
}

// Count returns the total number of entities matching the query parameters
func (uow *PostgresUnitOfWork[T]) Count(ctx context.Context, query *domain.QueryParams[T]) (int64, error) {
	repo := uow.getRepo()
	baseQuery := repo.GetDB().Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)
	return repo.Count(ctx, filteredQuery)
}

// Exists checks if any entity matches the provided identifier
func (uow *PostgresUnitOfWork[T]) Exists(ctx context.Context, identifier domain.IIdentifier) (bool, error) {
	repo := uow.getRepo()
	query := BuildQueryFromIdentifier[T](repo.GetDB(), identifier)
	return repo.Exists(ctx, query)
}

// Compile-time check to ensure PostgresUnitOfWork implements IUnitOfWork
var _ domain.IUnitOfWork[*domain.User] = (*PostgresUnitOfWork[*domain.User])(nil)
