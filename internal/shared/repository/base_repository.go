package repository

import (
	"context"

	"go-database/pkg/domain"
)

// IBaseRepository defines the contract for repository layer that delegates to IUnitOfWork.
// This provides a clean abstraction for feature repositories and enables dependency injection,
// mocking, and decoupling from specific persistence implementations.
type IBaseRepository[T domain.IBaseModel] interface {
	// Basic queries
	FindAll(ctx context.Context) ([]T, error)
	FindAllWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error)
	FindOne(ctx context.Context, filter T) (T, error)
	FindOneById(ctx context.Context, id uint) (T, error)
	FindOneByIdentifier(ctx context.Context, identifier domain.IIdentifier) (T, error)

	// Mutation operations
	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, identifier domain.IIdentifier, entity T) (T, error)
	Delete(ctx context.Context, identifier domain.IIdentifier) error

	// Soft-delete lifecycle
	SoftDelete(ctx context.Context, identifier domain.IIdentifier) (T, error)
	HardDelete(ctx context.Context, identifier domain.IIdentifier) (T, error)

	// Bulk operations
	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkSoftDelete(ctx context.Context, identifiers []domain.IIdentifier) error
	BulkHardDelete(ctx context.Context, identifiers []domain.IIdentifier) error

	// Trash management
	GetTrashed(ctx context.Context) ([]T, error)
	GetTrashedWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error)
	Restore(ctx context.Context, identifier domain.IIdentifier) (T, error)
	RestoreAll(ctx context.Context) error

	// Utility operations
	Count(ctx context.Context, query *domain.QueryParams[T]) (int64, error)
	Exists(ctx context.Context, identifier domain.IIdentifier) (bool, error)
}

// BaseRepository provides a generic repository implementation that delegates all operations
// to an IUnitOfWork instance. This follows the composition over inheritance principle and
// enables clean separation between business logic and persistence layer.
type BaseRepository[T domain.IBaseModel] struct {
	uow domain.IUnitOfWork[T]
}

// NewBaseRepository creates a new BaseRepository instance that delegates to the provided UnitOfWork
func NewBaseRepository[T domain.IBaseModel](uow domain.IUnitOfWork[T]) IBaseRepository[T] {
	return &BaseRepository[T]{
		uow: uow,
	}
}

// Basic queries

// FindAll retrieves all entities (excluding soft-deleted by default)
func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	return r.uow.FindAll(ctx)
}

// FindAllWithPagination retrieves entities with pagination support and returns total count
func (r *BaseRepository[T]) FindAllWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	return r.uow.FindAllWithPagination(ctx, query)
}

// FindOne retrieves a single entity matching the provided filter
func (r *BaseRepository[T]) FindOne(ctx context.Context, filter T) (T, error) {
	return r.uow.FindOne(ctx, filter)
}

// FindOneById retrieves a single entity by its ID
func (r *BaseRepository[T]) FindOneById(ctx context.Context, id uint) (T, error) {
	return r.uow.FindOneById(ctx, id)
}

// FindOneByIdentifier retrieves a single entity using the IIdentifier filter system
func (r *BaseRepository[T]) FindOneByIdentifier(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	return r.uow.FindOneByIdentifier(ctx, identifier)
}

// Mutation operations

// Insert creates a new entity and returns the created entity with populated fields
func (r *BaseRepository[T]) Insert(ctx context.Context, entity T) (T, error) {
	return r.uow.Insert(ctx, entity)
}

// Update modifies entities matching the identifier with the provided entity data
func (r *BaseRepository[T]) Update(ctx context.Context, identifier domain.IIdentifier, entity T) (T, error) {
	return r.uow.Update(ctx, identifier, entity)
}

// Delete performs a logical operation (soft-delete by default, hard-delete if configured)
func (r *BaseRepository[T]) Delete(ctx context.Context, identifier domain.IIdentifier) error {
	return r.uow.Delete(ctx, identifier)
}

// Soft-delete lifecycle management

// SoftDelete performs soft deletion by setting DeletedAt timestamp
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	return r.uow.SoftDelete(ctx, identifier)
}

// HardDelete permanently removes entities from the database
func (r *BaseRepository[T]) HardDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	return r.uow.HardDelete(ctx, identifier)
}

// Bulk operations

// BulkInsert creates multiple entities in a single operation
func (r *BaseRepository[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	return r.uow.BulkInsert(ctx, entities)
}

// BulkUpdate modifies multiple entities in a single operation
func (r *BaseRepository[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	return r.uow.BulkUpdate(ctx, entities)
}

// BulkSoftDelete soft-deletes multiple entities identified by the provided identifiers
func (r *BaseRepository[T]) BulkSoftDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	return r.uow.BulkSoftDelete(ctx, identifiers)
}

// BulkHardDelete permanently removes multiple entities identified by the provided identifiers
func (r *BaseRepository[T]) BulkHardDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	return r.uow.BulkHardDelete(ctx, identifiers)
}

// Trash management

// GetTrashed retrieves all soft-deleted entities
func (r *BaseRepository[T]) GetTrashed(ctx context.Context) ([]T, error) {
	return r.uow.GetTrashed(ctx)
}

// GetTrashedWithPagination retrieves soft-deleted entities with pagination
func (r *BaseRepository[T]) GetTrashedWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	return r.uow.GetTrashedWithPagination(ctx, query)
}

// Restore recovers soft-deleted entities by clearing their DeletedAt timestamp
func (r *BaseRepository[T]) Restore(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	return r.uow.Restore(ctx, identifier)
}

// RestoreAll recovers all soft-deleted entities of type T
func (r *BaseRepository[T]) RestoreAll(ctx context.Context) error {
	return r.uow.RestoreAll(ctx)
}

// Utility operations

// Count returns the total number of entities matching the query parameters
func (r *BaseRepository[T]) Count(ctx context.Context, query *domain.QueryParams[T]) (int64, error) {
	return r.uow.Count(ctx, query)
}

// Exists checks if any entity matches the provided identifier
func (r *BaseRepository[T]) Exists(ctx context.Context, identifier domain.IIdentifier) (bool, error) {
	return r.uow.Exists(ctx, identifier)
}

// Compile-time check to ensure BaseRepository implements IBaseRepository
var _ IBaseRepository[*domain.User] = (*BaseRepository[*domain.User])(nil)
