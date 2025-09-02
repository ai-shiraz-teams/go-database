package domain

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/pkg/identifier"
)

type IBaseRepository[T IBaseModel] interface {
	// Transaction Management
	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context)
	IsInTransaction() bool

	// Basic Operations
	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)
	UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (T, error)
	Delete(ctx context.Context, identifier identifier.IIdentifier) error
	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// Query Operations
	FindAll(ctx context.Context) ([]T, error)
	FindAllWithArchived(ctx context.Context, withArchived bool) ([]T, error)
	FindAllWithPagination(ctx context.Context, queryParams IQueryParams[T]) ([]T, int64, error)
	FindAllByQuery(ctx context.Context, queryParams IQueryParams[T]) ([]T, error)
	FindOne(ctx context.Context, filter T, includeDeleted bool) (T, error)
	FindOneById(ctx context.Context, id int) (T, error)
	FindOneBySlug(ctx context.Context, slug string) (T, error)
	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	FindFirst(ctx context.Context, queryParams IQueryParams[T]) (T, error)

	// Utility Operations
	Count(ctx context.Context, queryParams IQueryParams[T]) (int64, error)
	Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error)
	ExistsWithQuery(ctx context.Context, queryParams IQueryParams[T]) (bool, error)
	ResolveIDByUniqueField(ctx context.Context, model IBaseModel, field string, value interface{}) (int, error)

	// Bulk Operations
	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkUpdatePartial(ctx context.Context, updates []BulkUpdateOperation) (BulkOperationResult, error)
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error
	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	// Trash Management
	GetTrashed(ctx context.Context) ([]T, error)
	GetTrashedWithPagination(ctx context.Context, params IQueryParams[T]) ([]T, int64, error)
	GetTrashedByQuery(ctx context.Context, queryParams IQueryParams[T]) ([]T, error)
	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	RestoreAll(ctx context.Context) error

	// Relational Operations
	Connect(ctx context.Context, parentEntity T, relationField string, childEntity IBaseModel) error
	ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error
	CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity IBaseModel) (IBaseModel, error)
	ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity IBaseModel) (IBaseModel, error)

	// Advanced Operations
	Upsert(ctx context.Context, entity T, conflictFields []string) (T, error)
	BulkUpsert(ctx context.Context, entities []T, conflictFields []string) ([]T, error)
	GetDistinctValues(ctx context.Context, field string, queryParams IQueryParams[T]) ([]interface{}, error)
	Aggregate(ctx context.Context, operation AggregateOperation, field string, queryParams IQueryParams[T]) (interface{}, error)

	// Streaming and Performance
	FindAllStream(ctx context.Context, queryParams IQueryParams[T], batchSize int) (<-chan StreamResult[T], error)
	ExecuteInBatches(ctx context.Context, queryParams IQueryParams[T], batchSize int, processor func([]T) error) error
	RefreshCache(ctx context.Context) error
}
