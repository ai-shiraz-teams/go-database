package unit_of_work

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database/internal/shared/query"
	"github.com/ai-shiraz-teams/go-database/internal/shared/types"
)

// IUnitOfWork defines the contract for transactional repository access across all modules.
// It provides a fully generic and reusable abstraction that wraps all database layer access
// and enforces clean layering with CQRS, soft delete logic, bulk operations, and typed querying.
//
// This interface serves as the primary abstraction for repository access and ensures:
// - Transaction safety for all operations
// - Type safety through generics
// - Separation of soft-delete and hard-delete logic
// - Support for complex filtering via IIdentifier
// - Pagination and filtering via QueryParams[T]
// - Bulk operations with explicit semantics
type IUnitOfWork[T types.IBaseModel] interface {
	// Transaction management
	// BeginTransaction starts a new database transaction
	BeginTransaction(ctx context.Context) error

	// CommitTransaction commits the current transaction
	CommitTransaction(ctx context.Context) error

	// RollbackTransaction rolls back the current transaction
	RollbackTransaction(ctx context.Context)

	// Basic queries
	// FindAll retrieves all entities of type T (excluding soft-deleted by default)
	FindAll(ctx context.Context) ([]T, error)

	// FindAllWithPagination retrieves entities with pagination support and returns total count
	FindAllWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)

	// FindOne retrieves a single entity matching the provided filter
	FindOne(ctx context.Context, filter T) (T, error)

	// FindOneById retrieves a single entity by its ID
	FindOneById(ctx context.Context, id int) (T, error)

	// FindOneByIdentifier retrieves a single entity using the IIdentifier filter system
	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// Mutation operations
	// Insert creates a new entity and returns the created entity with populated fields
	Insert(ctx context.Context, entity T) (T, error)

	// Update modifies entities matching the identifier with the provided entity data
	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)

	// Delete performs a logical operation (soft-delete by default, hard-delete if configured)
	Delete(ctx context.Context, identifier identifier.IIdentifier) error

	// Soft-delete lifecycle management
	// SoftDelete performs soft deletion by setting DeletedAt timestamp
	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// HardDelete permanently removes entities from the database
	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// GetTrashed retrieves all soft-deleted entities
	GetTrashed(ctx context.Context) ([]T, error)

	// GetTrashedWithPagination retrieves soft-deleted entities with pagination
	GetTrashedWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error)

	// Restore recovers soft-deleted entities by clearing their DeletedAt timestamp
	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// RestoreAll recovers all soft-deleted entities of type T
	RestoreAll(ctx context.Context) error

	// Bulk operations - explicit and efficient for large datasets
	// BulkInsert creates multiple entities in a single operation
	BulkInsert(ctx context.Context, entities []T) ([]T, error)

	// BulkUpdate modifies multiple entities in a single operation
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)

	// BulkSoftDelete soft-deletes multiple entities identified by the provided identifiers
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	// BulkHardDelete permanently removes multiple entities identified by the provided identifiers
	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	// Utility operations
	// ResolveIDByUniqueField finds the ID of an entity by searching a unique field
	ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error)

	// Count returns the total number of entities matching the query parameters
	Count(ctx context.Context, query *query.QueryParams[T]) (int64, error)

	// Exists checks if any entity matches the provided identifier
	Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error)
}

// IUnitOfWorkFactory defines the contract for creating unit of work instances.
// This allows for dependency injection and testing with different implementations.
//
// Note: Due to Go's limitation with generic interface methods, factory methods
// should be implemented as standalone functions rather than interface methods.
type IUnitOfWorkFactory interface {
	// NewTransaction starts a new database transaction that can be used across multiple unit of work instances
	NewTransaction(ctx context.Context) (interface{}, error)

	// CommitTransaction commits the provided transaction
	CommitTransaction(ctx context.Context, tx interface{}) error

	// RollbackTransaction rolls back the provided transaction
	RollbackTransaction(ctx context.Context, tx interface{}) error
}

// TransactionOptions defines configuration for transaction behavior
type TransactionOptions struct {
	// IsolationLevel specifies the transaction isolation level
	IsolationLevel string

	// ReadOnly indicates if the transaction should be read-only
	ReadOnly bool

	// Timeout specifies the maximum duration for the transaction
	Timeout int64
}

// BulkOperationResult provides information about the outcome of bulk operations
type BulkOperationResult struct {
	// SuccessCount is the number of entities successfully processed
	SuccessCount int

	// FailureCount is the number of entities that failed to process
	FailureCount int

	// Errors contains any errors that occurred during processing
	Errors []error

	// ProcessedIDs contains the IDs of entities that were successfully processed
	ProcessedIDs []int
}
