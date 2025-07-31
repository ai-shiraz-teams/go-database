package repository

import (
	"context"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"
)

// IBaseRepository defines the contract for repository layer that delegates to IUnitOfWork.
// This provides a clean abstraction for feature repositories and enables dependency injection,
// mocking, and decoupling from specific persistence implementations.
type IBaseRepository[T types.IBaseModel] interface {
	// Basic queries
	FindAll(ctx context.Context) ([]T, error)
	FindAllWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)
	FindOne(ctx context.Context, filter T) (T, error)
	FindOneById(ctx context.Context, id int) (T, error)
	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// Mutation operations
	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)
	Delete(ctx context.Context, identifier identifier.IIdentifier) error

	// Soft-delete lifecycle
	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// Bulk operations
	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error
	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	// Trash management
	GetTrashed(ctx context.Context) ([]T, error)
	GetTrashedWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)
	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	RestoreAll(ctx context.Context) error

	// Utility operations
	Count(ctx context.Context, query *query.QueryParams[T]) (int64, error)
	Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error)
}
