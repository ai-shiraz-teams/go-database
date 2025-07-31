package repository

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"

	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

// mockUnitOfWork provides a local mock implementation for repository testing.
// This avoids import cycles with the testutil package.
type mockUnitOfWork struct {
	// Mock call tracking fields
	FindAllCalled                  bool
	FindAllWithPaginationCalled    bool
	FindOneCalled                  bool
	FindOneByIdCalled              bool
	FindOneBySlugCalled            bool
	FindOneByIdentifierCalled      bool
	InsertCalled                   bool
	UpdateCalled                   bool
	DeleteCalled                   bool
	SoftDeleteCalled               bool
	HardDeleteCalled               bool
	BulkInsertCalled               bool
	BulkUpdateCalled               bool
	BulkSoftDeleteCalled           bool
	BulkHardDeleteCalled           bool
	GetTrashedCalled               bool
	GetTrashedWithPaginationCalled bool
	RestoreCalled                  bool
	RestoreAllCalled               bool
	CountCalled                    bool
	ExistsCalled                   bool
	BeginTransactionCalled         bool
	CommitTransactionCalled        bool
	RollbackTransactionCalled      bool
	ResolveIDByUniqueFieldCalled   bool

	// Mock return values
	FindAllResult                  []*testutil.TestEntity
	FindAllWithPaginationResult    []*testutil.TestEntity
	FindAllWithPaginationCount     int64
	FindOneResult                  *testutil.TestEntity
	FindOneByIdResult              *testutil.TestEntity
	FindOneBySlugResult            *testutil.TestEntity
	FindOneByIdentifierResult      *testutil.TestEntity
	InsertResult                   *testutil.TestEntity
	UpdateResult                   *testutil.TestEntity
	SoftDeleteResult               *testutil.TestEntity
	HardDeleteResult               *testutil.TestEntity
	BulkInsertResult               []*testutil.TestEntity
	BulkUpdateResult               []*testutil.TestEntity
	GetTrashedResult               []*testutil.TestEntity
	GetTrashedWithPaginationResult []*testutil.TestEntity
	GetTrashedWithPaginationCount  int64
	RestoreResult                  *testutil.TestEntity
	CountResult                    int64
	ExistsResult                   bool
	ResolveIDByUniqueFieldResult   int

	// Mock error values
	FindAllError                  error
	FindAllWithPaginationError    error
	FindOneError                  error
	FindOneByIdError              error
	FindOneBySlugError            error
	FindOneByIdentifierError      error
	InsertError                   error
	UpdateError                   error
	DeleteError                   error
	SoftDeleteError               error
	HardDeleteError               error
	BulkInsertError               error
	BulkUpdateError               error
	BulkSoftDeleteError           error
	BulkHardDeleteError           error
	GetTrashedError               error
	GetTrashedWithPaginationError error
	RestoreError                  error
	RestoreAllError               error
	CountError                    error
	ExistsError                   error
	BeginTransactionError         error
	CommitTransactionError        error
	ResolveIDByUniqueFieldError   error
}

// Mock method implementations
func (m *mockUnitOfWork) FindAll(ctx context.Context) ([]*testutil.TestEntity, error) {
	m.FindAllCalled = true
	return m.FindAllResult, m.FindAllError
}

func (m *mockUnitOfWork) FindAllWithPagination(ctx context.Context, params *query.QueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, int64, error) {
	m.FindAllWithPaginationCalled = true
	return m.FindAllWithPaginationResult, m.FindAllWithPaginationCount, m.FindAllWithPaginationError
}

func (m *mockUnitOfWork) FindOne(ctx context.Context, filter *testutil.TestEntity) (*testutil.TestEntity, error) {
	m.FindOneCalled = true
	return m.FindOneResult, m.FindOneError
}

func (m *mockUnitOfWork) FindOneById(ctx context.Context, id int) (*testutil.TestEntity, error) {
	m.FindOneByIdCalled = true
	return m.FindOneByIdResult, m.FindOneByIdError
}

func (m *mockUnitOfWork) FindOneBySlug(ctx context.Context, slug string) (*testutil.TestEntity, error) {
	m.FindOneBySlugCalled = true
	return m.FindOneBySlugResult, m.FindOneBySlugError
}

func (m *mockUnitOfWork) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.FindOneByIdentifierCalled = true
	return m.FindOneByIdentifierResult, m.FindOneByIdentifierError
}

func (m *mockUnitOfWork) Insert(ctx context.Context, entity *testutil.TestEntity) (*testutil.TestEntity, error) {
	m.InsertCalled = true
	return m.InsertResult, m.InsertError
}

func (m *mockUnitOfWork) Update(ctx context.Context, identifier identifier.IIdentifier, entity *testutil.TestEntity) (*testutil.TestEntity, error) {
	m.UpdateCalled = true
	return m.UpdateResult, m.UpdateError
}

func (m *mockUnitOfWork) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	m.DeleteCalled = true
	return m.DeleteError
}

func (m *mockUnitOfWork) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.SoftDeleteCalled = true
	return m.SoftDeleteResult, m.SoftDeleteError
}

func (m *mockUnitOfWork) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.HardDeleteCalled = true
	return m.HardDeleteResult, m.HardDeleteError
}

func (m *mockUnitOfWork) BulkInsert(ctx context.Context, entities []*testutil.TestEntity) ([]*testutil.TestEntity, error) {
	m.BulkInsertCalled = true
	return m.BulkInsertResult, m.BulkInsertError
}

func (m *mockUnitOfWork) BulkUpdate(ctx context.Context, entities []*testutil.TestEntity) ([]*testutil.TestEntity, error) {
	m.BulkUpdateCalled = true
	return m.BulkUpdateResult, m.BulkUpdateError
}

func (m *mockUnitOfWork) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.BulkSoftDeleteCalled = true
	return m.BulkSoftDeleteError
}

func (m *mockUnitOfWork) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.BulkHardDeleteCalled = true
	return m.BulkHardDeleteError
}

func (m *mockUnitOfWork) GetTrashed(ctx context.Context) ([]*testutil.TestEntity, error) {
	m.GetTrashedCalled = true
	return m.GetTrashedResult, m.GetTrashedError
}

func (m *mockUnitOfWork) GetTrashedWithPagination(ctx context.Context, params *query.QueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, int64, error) {
	m.GetTrashedWithPaginationCalled = true
	return m.GetTrashedWithPaginationResult, m.GetTrashedWithPaginationCount, m.GetTrashedWithPaginationError
}

func (m *mockUnitOfWork) Restore(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.RestoreCalled = true
	return m.RestoreResult, m.RestoreError
}

func (m *mockUnitOfWork) RestoreAll(ctx context.Context) error {
	m.RestoreAllCalled = true
	return m.RestoreAllError
}

func (m *mockUnitOfWork) Count(ctx context.Context, params *query.QueryParams[*testutil.TestEntity]) (int64, error) {
	m.CountCalled = true
	return m.CountResult, m.CountError
}

func (m *mockUnitOfWork) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	m.ExistsCalled = true
	return m.ExistsResult, m.ExistsError
}

func (m *mockUnitOfWork) BeginTransaction(ctx context.Context) error {
	m.BeginTransactionCalled = true
	return m.BeginTransactionError
}

func (m *mockUnitOfWork) CommitTransaction(ctx context.Context) error {
	m.CommitTransactionCalled = true
	return m.CommitTransactionError
}

func (m *mockUnitOfWork) RollbackTransaction(ctx context.Context) {
	m.RollbackTransactionCalled = true
}

func (m *mockUnitOfWork) ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error) {
	m.ResolveIDByUniqueFieldCalled = true
	return m.ResolveIDByUniqueFieldResult, m.ResolveIDByUniqueFieldError
}
