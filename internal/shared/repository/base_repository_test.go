package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/query"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/unit_of_work"
)

// MockEntity implements IBaseModel for testing
type MockEntity struct {
	types.BaseEntity
}

// MockUnitOfWork implements IUnitOfWork for testing delegation
type MockUnitOfWork struct {
	FindAllCalled                  bool
	FindAllWithPaginationCalled    bool
	FindOneCalled                  bool
	FindOneByIdCalled              bool
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

	// Return values for methods
	FindAllResult                  []*MockEntity
	FindAllError                   error
	FindAllWithPaginationResult    []*MockEntity
	FindAllWithPaginationCount     int64
	FindAllWithPaginationError     error
	FindOneResult                  *MockEntity
	FindOneError                   error
	FindOneByIdResult              *MockEntity
	FindOneByIdError               error
	FindOneByIdentifierResult      *MockEntity
	FindOneByIdentifierError       error
	InsertResult                   *MockEntity
	InsertError                    error
	UpdateResult                   *MockEntity
	UpdateError                    error
	DeleteError                    error
	SoftDeleteResult               *MockEntity
	SoftDeleteError                error
	HardDeleteResult               *MockEntity
	HardDeleteError                error
	BulkInsertResult               []*MockEntity
	BulkInsertError                error
	BulkUpdateResult               []*MockEntity
	BulkUpdateError                error
	BulkSoftDeleteError            error
	BulkHardDeleteError            error
	GetTrashedResult               []*MockEntity
	GetTrashedError                error
	GetTrashedWithPaginationResult []*MockEntity
	GetTrashedWithPaginationCount  int64
	GetTrashedWithPaginationError  error
	RestoreResult                  *MockEntity
	RestoreError                   error
	RestoreAllError                error
	CountResult                    int64
	CountError                     error
	ExistsResult                   bool
	ExistsError                    error
	BeginTransactionError          error
	CommitTransactionError         error
	ResolveIDByUniqueFieldResult   int
	ResolveIDByUniqueFieldError    error
}

// Mock implementation of IUnitOfWork interface
func (m *MockUnitOfWork) FindAll(ctx context.Context) ([]*MockEntity, error) {
	m.FindAllCalled = true
	return m.FindAllResult, m.FindAllError
}

func (m *MockUnitOfWork) FindAllWithPagination(ctx context.Context, params *query.QueryParams[*MockEntity]) ([]*MockEntity, int64, error) {
	m.FindAllWithPaginationCalled = true
	return m.FindAllWithPaginationResult, m.FindAllWithPaginationCount, m.FindAllWithPaginationError
}

func (m *MockUnitOfWork) FindOne(ctx context.Context, filter *MockEntity) (*MockEntity, error) {
	m.FindOneCalled = true
	return m.FindOneResult, m.FindOneError
}

func (m *MockUnitOfWork) FindOneById(ctx context.Context, id int) (*MockEntity, error) {
	m.FindOneByIdCalled = true
	return m.FindOneByIdResult, m.FindOneByIdError
}

func (m *MockUnitOfWork) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (*MockEntity, error) {
	m.FindOneByIdentifierCalled = true
	return m.FindOneByIdentifierResult, m.FindOneByIdentifierError
}

func (m *MockUnitOfWork) Insert(ctx context.Context, entity *MockEntity) (*MockEntity, error) {
	m.InsertCalled = true
	return m.InsertResult, m.InsertError
}

func (m *MockUnitOfWork) Update(ctx context.Context, identifier identifier.IIdentifier, entity *MockEntity) (*MockEntity, error) {
	m.UpdateCalled = true
	return m.UpdateResult, m.UpdateError
}

func (m *MockUnitOfWork) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	m.DeleteCalled = true
	return m.DeleteError
}

func (m *MockUnitOfWork) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (*MockEntity, error) {
	m.SoftDeleteCalled = true
	return m.SoftDeleteResult, m.SoftDeleteError
}

func (m *MockUnitOfWork) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (*MockEntity, error) {
	m.HardDeleteCalled = true
	return m.HardDeleteResult, m.HardDeleteError
}

func (m *MockUnitOfWork) BulkInsert(ctx context.Context, entities []*MockEntity) ([]*MockEntity, error) {
	m.BulkInsertCalled = true
	return m.BulkInsertResult, m.BulkInsertError
}

func (m *MockUnitOfWork) BulkUpdate(ctx context.Context, entities []*MockEntity) ([]*MockEntity, error) {
	m.BulkUpdateCalled = true
	return m.BulkUpdateResult, m.BulkUpdateError
}

func (m *MockUnitOfWork) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.BulkSoftDeleteCalled = true
	return m.BulkSoftDeleteError
}

func (m *MockUnitOfWork) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.BulkHardDeleteCalled = true
	return m.BulkHardDeleteError
}

func (m *MockUnitOfWork) GetTrashed(ctx context.Context) ([]*MockEntity, error) {
	m.GetTrashedCalled = true
	return m.GetTrashedResult, m.GetTrashedError
}

func (m *MockUnitOfWork) GetTrashedWithPagination(ctx context.Context, params *query.QueryParams[*MockEntity]) ([]*MockEntity, int64, error) {
	m.GetTrashedWithPaginationCalled = true
	return m.GetTrashedWithPaginationResult, m.GetTrashedWithPaginationCount, m.GetTrashedWithPaginationError
}

func (m *MockUnitOfWork) Restore(ctx context.Context, identifier identifier.IIdentifier) (*MockEntity, error) {
	m.RestoreCalled = true
	return m.RestoreResult, m.RestoreError
}

func (m *MockUnitOfWork) RestoreAll(ctx context.Context) error {
	m.RestoreAllCalled = true
	return m.RestoreAllError
}

func (m *MockUnitOfWork) Count(ctx context.Context, params *query.QueryParams[*MockEntity]) (int64, error) {
	m.CountCalled = true
	return m.CountResult, m.CountError
}

func (m *MockUnitOfWork) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	m.ExistsCalled = true
	return m.ExistsResult, m.ExistsError
}

func (m *MockUnitOfWork) BeginTransaction(ctx context.Context) error {
	m.BeginTransactionCalled = true
	return m.BeginTransactionError
}

func (m *MockUnitOfWork) CommitTransaction(ctx context.Context) error {
	m.CommitTransactionCalled = true
	return m.CommitTransactionError
}

func (m *MockUnitOfWork) RollbackTransaction(ctx context.Context) {
	m.RollbackTransactionCalled = true
}

func (m *MockUnitOfWork) ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error) {
	m.ResolveIDByUniqueFieldCalled = true
	return m.ResolveIDByUniqueFieldResult, m.ResolveIDByUniqueFieldError
}

// Compile-time check
var _ unit_of_work.IUnitOfWork[*MockEntity] = (*MockUnitOfWork)(nil)

// TestNewBaseRepository validates repository creation
func TestNewBaseRepository(t *testing.T) {
	// Arrange
	mockUow := &MockUnitOfWork{}

	// Act
	repo := NewBaseRepository[*MockEntity](mockUow)

	// Assert
	if repo == nil {
		t.Fatal("NewBaseRepository returned nil")
	}

	// Verify it implements the interface
	var _ IBaseRepository[*MockEntity] = repo
}

// TestBaseRepository_FindAll validates FindAll delegation
func TestBaseRepository_FindAll(t *testing.T) {
	// Arrange
	mockUow := &MockUnitOfWork{
		FindAllResult: []*MockEntity{{BaseEntity: types.BaseEntity{ID: 1}}},
		FindAllError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()

	// Act
	result, err := repo.FindAll(ctx)

	// Assert
	if !mockUow.FindAllCalled {
		t.Error("Expected FindAll to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	if result[0].ID != 1 {
		t.Errorf("Expected ID 1, got %d", result[0].ID)
	}
}

// TestBaseRepository_FindAll_Error validates error propagation
func TestBaseRepository_FindAll_Error(t *testing.T) {
	// Arrange
	expectedError := errors.New("database error")
	mockUow := &MockUnitOfWork{
		FindAllResult: nil,
		FindAllError:  expectedError,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()

	// Act
	result, err := repo.FindAll(ctx)

	// Assert
	if !mockUow.FindAllCalled {
		t.Error("Expected FindAll to be called on UnitOfWork")
	}

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

// TestBaseRepository_FindAllWithPagination validates pagination delegation
func TestBaseRepository_FindAllWithPagination(t *testing.T) {
	// Arrange
	mockUow := &MockUnitOfWork{
		FindAllWithPaginationResult: []*MockEntity{{BaseEntity: types.BaseEntity{ID: 1}}},
		FindAllWithPaginationCount:  1,
		FindAllWithPaginationError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	params := query.NewQueryParams[*MockEntity]()

	// Act
	result, count, err := repo.FindAllWithPagination(ctx, params)

	// Assert
	if !mockUow.FindAllWithPaginationCalled {
		t.Error("Expected FindAllWithPagination to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

// TestBaseRepository_FindOneById validates FindOneById delegation
func TestBaseRepository_FindOneById(t *testing.T) {
	// Arrange
	expectedEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 123}}
	mockUow := &MockUnitOfWork{
		FindOneByIdResult: expectedEntity,
		FindOneByIdError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()

	// Act
	result, err := repo.FindOneById(ctx, 123)

	// Assert
	if !mockUow.FindOneByIdCalled {
		t.Error("Expected FindOneById to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.ID != 123 {
		t.Errorf("Expected ID 123, got %d", result.ID)
	}
}

// TestBaseRepository_Insert validates Insert delegation
func TestBaseRepository_Insert(t *testing.T) {
	// Arrange
	inputEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 0}}
	expectedEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 1}}
	mockUow := &MockUnitOfWork{
		InsertResult: expectedEntity,
		InsertError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()

	// Act
	result, err := repo.Insert(ctx, inputEntity)

	// Assert
	if !mockUow.InsertCalled {
		t.Error("Expected Insert to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

// TestBaseRepository_Update validates Update delegation
func TestBaseRepository_Update(t *testing.T) {
	// Arrange
	updateEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 1}}
	expectedEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 1}}
	mockUow := &MockUnitOfWork{
		UpdateResult: expectedEntity,
		UpdateError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	mockIdentifier := identifier.NewIdentifier()

	// Act
	result, err := repo.Update(ctx, mockIdentifier, updateEntity)

	// Assert
	if !mockUow.UpdateCalled {
		t.Error("Expected Update to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

// TestBaseRepository_Delete validates Delete delegation
func TestBaseRepository_Delete(t *testing.T) {
	// Arrange
	mockUow := &MockUnitOfWork{
		DeleteError: nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	mockIdentifier := identifier.NewIdentifier()

	// Act
	err := repo.Delete(ctx, mockIdentifier)

	// Assert
	if !mockUow.DeleteCalled {
		t.Error("Expected Delete to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestBaseRepository_SoftDelete validates SoftDelete delegation
func TestBaseRepository_SoftDelete(t *testing.T) {
	// Arrange
	expectedEntity := &MockEntity{BaseEntity: types.BaseEntity{ID: 1}}
	mockUow := &MockUnitOfWork{
		SoftDeleteResult: expectedEntity,
		SoftDeleteError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	mockIdentifier := identifier.NewIdentifier()

	// Act
	result, err := repo.SoftDelete(ctx, mockIdentifier)

	// Assert
	if !mockUow.SoftDeleteCalled {
		t.Error("Expected SoftDelete to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

// TestBaseRepository_BulkInsert validates BulkInsert delegation
func TestBaseRepository_BulkInsert(t *testing.T) {
	// Arrange
	inputEntities := []*MockEntity{
		{BaseEntity: types.BaseEntity{ID: 0}},
		{BaseEntity: types.BaseEntity{ID: 0}},
	}
	expectedEntities := []*MockEntity{
		{BaseEntity: types.BaseEntity{ID: 1}},
		{BaseEntity: types.BaseEntity{ID: 2}},
	}
	mockUow := &MockUnitOfWork{
		BulkInsertResult: expectedEntities,
		BulkInsertError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()

	// Act
	result, err := repo.BulkInsert(ctx, inputEntities)

	// Assert
	if !mockUow.BulkInsertCalled {
		t.Error("Expected BulkInsert to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	if result[0].ID != 1 {
		t.Errorf("Expected first ID 1, got %d", result[0].ID)
	}

	if result[1].ID != 2 {
		t.Errorf("Expected second ID 2, got %d", result[1].ID)
	}
}

// TestBaseRepository_Count validates Count delegation
func TestBaseRepository_Count(t *testing.T) {
	// Arrange
	expectedCount := int64(42)
	mockUow := &MockUnitOfWork{
		CountResult: expectedCount,
		CountError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	params := query.NewQueryParams[*MockEntity]()

	// Act
	result, err := repo.Count(ctx, params)

	// Assert
	if !mockUow.CountCalled {
		t.Error("Expected Count to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, result)
	}
}

// TestBaseRepository_Exists validates Exists delegation
func TestBaseRepository_Exists(t *testing.T) {
	// Arrange
	mockUow := &MockUnitOfWork{
		ExistsResult: true,
		ExistsError:  nil,
	}
	repo := NewBaseRepository[*MockEntity](mockUow)
	ctx := context.Background()
	mockIdentifier := identifier.NewIdentifier()

	// Act
	result, err := repo.Exists(ctx, mockIdentifier)

	// Assert
	if !mockUow.ExistsCalled {
		t.Error("Expected Exists to be called on UnitOfWork")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !result {
		t.Error("Expected true result, got false")
	}
}
