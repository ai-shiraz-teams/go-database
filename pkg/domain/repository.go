package domain

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/pkg/identifier"
)

type IBaseRepository[T IBaseModel] interface {
	FindAll(ctx context.Context) ([]T, error)
	FindAllWithPagination(ctx context.Context, query IQueryParams[T]) ([]T, int64, error)
	FindOne(ctx context.Context, filter T, includeDeleted bool) (T, error)
	FindOneBySlug(ctx context.Context, slug string) (T, error)
	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)
	Delete(ctx context.Context, identifier identifier.IIdentifier) error

	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error
	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	GetTrashed(ctx context.Context) ([]T, error)
	GetTrashedWithPagination(ctx context.Context, query IQueryParams[T]) ([]T, int64, error)
	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	RestoreAll(ctx context.Context) error

	Count(ctx context.Context, query IQueryParams[T]) (int64, error)
	Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error)
}
