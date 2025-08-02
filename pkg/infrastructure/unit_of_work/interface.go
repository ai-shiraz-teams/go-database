package unit_of_work

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"
)

// BulkUpdateOperation represents a single update operation in a bulk update
type BulkUpdateOperation struct {
	Identifier identifier.IIdentifier `json:"identifier"`
	Updates    map[string]interface{} `json:"updates"`
}

// AggregateOperation defines the type of aggregation to perform
type AggregateOperation string

const (
	AggregateSum      AggregateOperation = "sum"
	AggregateAvg      AggregateOperation = "avg"
	AggregateMin      AggregateOperation = "min"
	AggregateMax      AggregateOperation = "max"
	AggregateCount    AggregateOperation = "count"
	AggregateStdDev   AggregateOperation = "stddev"
	AggregateVariance AggregateOperation = "variance"
)

// StreamResult represents a single result in a stream
type StreamResult[T types.IBaseModel] struct {
	Data  T     `json:"data"`
	Error error `json:"error,omitempty"`
}

// ProjectionResult represents a result with only selected fields
type ProjectionResult map[string]interface{}

// RelationshipType defines the type of relationship between entities
type RelationshipType string

const (
	RelationshipOneToOne   RelationshipType = "one_to_one"
	RelationshipOneToMany  RelationshipType = "one_to_many"
	RelationshipManyToOne  RelationshipType = "many_to_one"
	RelationshipManyToMany RelationshipType = "many_to_many"
)

// CacheHint provides hints for query optimization and caching
type CacheHint struct {
	ReadOnly     bool   `json:"readOnly"`
	TTL          int64  `json:"ttl"`
	CacheKey     string `json:"cacheKey,omitempty"`
	SkipCache    bool   `json:"skipCache"`
	RefreshCache bool   `json:"refreshCache"`
}

type IUnitOfWork[T types.IBaseModel] interface {
	BeginTransaction(ctx context.Context) error

	CommitTransaction(ctx context.Context) error

	RollbackTransaction(ctx context.Context)

	IsInTransaction() bool

	Connect(ctx context.Context, parentEntity T, relationField string, childEntity types.IBaseModel) error

	ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error

	CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity types.IBaseModel) (types.IBaseModel, error)

	ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity types.IBaseModel) (types.IBaseModel, error)

	FindAll(ctx context.Context) ([]T, error)

	FindAllWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)

	FindAllByQuery(ctx context.Context, query *query.QueryParams[T]) ([]T, error)

	FindFirst(ctx context.Context, query *query.QueryParams[T]) (T, error)

	FindManyRaw(ctx context.Context, rawQuery string, args ...interface{}) ([]T, error)

	FindOneWithProjection(ctx context.Context, identifier identifier.IIdentifier, fields []string) (map[string]interface{}, error)

	FindAllWithProjection(ctx context.Context, query *query.QueryParams[T], fields []string) ([]map[string]interface{}, error)

	Count(ctx context.Context, query *query.QueryParams[T]) (int64, error)

	CountDistinct(ctx context.Context, field string, query *query.QueryParams[T]) (int64, error)

	FindOne(ctx context.Context, filter T) (T, error)

	FindOneById(ctx context.Context, id int) (T, error)

	FindOneBySlug(ctx context.Context, slug string) (T, error)

	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	Insert(ctx context.Context, entity T) (T, error)

	BulkInsert(ctx context.Context, entities []T) ([]T, error)

	Upsert(ctx context.Context, entity T, conflictFields []string) (T, error)

	BulkUpsert(ctx context.Context, entities []T, conflictFields []string) ([]T, error)

	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)

	UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (T, error)

	BulkUpdate(ctx context.Context, entities []T) ([]T, error)

	BulkUpdatePartial(ctx context.Context, updates []BulkUpdateOperation) (BulkOperationResult, error)

	Delete(ctx context.Context, identifier identifier.IIdentifier) error

	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	DeleteAll(ctx context.Context, query *query.QueryParams[T], hardDelete bool) (BulkOperationResult, error)

	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	RestoreAll(ctx context.Context) error

	BulkRestore(ctx context.Context, identifiers []identifier.IIdentifier) (BulkOperationResult, error)

	GetTrashed(ctx context.Context) ([]T, error)

	GetTrashedWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)

	GetTrashedByQuery(ctx context.Context, query *query.QueryParams[T]) ([]T, error)

	ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error)

	Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error)

	ExistsWithQuery(ctx context.Context, query *query.QueryParams[T]) (bool, error)

	GetDistinctValues(ctx context.Context, field string, query *query.QueryParams[T]) ([]interface{}, error)

	Aggregate(ctx context.Context, operation AggregateOperation, field string, query *query.QueryParams[T]) (interface{}, error)

	FindAllStream(ctx context.Context, query *query.QueryParams[T], batchSize int) (<-chan StreamResult[T], error)

	ExecuteInBatches(ctx context.Context, query *query.QueryParams[T], batchSize int, processor func([]T) error) error

	RefreshCache(ctx context.Context) error
}

type IUnitOfWorkFactory interface {
	NewTransaction(ctx context.Context) (interface{}, error)

	CommitTransaction(ctx context.Context, tx interface{}) error

	RollbackTransaction(ctx context.Context, tx interface{}) error
}

type TransactionOptions struct {
	IsolationLevel string

	ReadOnly bool

	Timeout int64
}

type BulkOperationResult struct {
	SuccessCount int

	FailureCount int

	Errors []error

	ProcessedIDs []int
}
