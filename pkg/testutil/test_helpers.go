package testutil

import (
	"context"
	"testing"

	"github.com/ai-shiraz-teams/go-database/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database/internal/shared/types"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestEntity is a unified test entity for all SDK tests.
// This replaces all duplicate MockEntity, TestEntity, FilterTestEntity across the codebase.
type TestEntity struct {
	types.BaseEntity
	Name        string `gorm:"column:name" json:"name"`
	Email       string `gorm:"column:email" json:"email"`
	Age         int    `gorm:"column:age" json:"age"`
	IsActive    bool   `gorm:"column:is_active" json:"is_active"`
	Description string `gorm:"column:description" json:"description"`
	Status      string `gorm:"column:status" json:"status"`
}

// TableName returns the table name for GORM
func (te *TestEntity) TableName() string {
	return "test_entities"
}

// SetupTestDB creates a standardized in-memory SQLite database for testing.
// This replaces all duplicate setupTestDB, setupFilterTestDB functions across the codebase.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate the unified test entity
	if err := db.AutoMigrate(&TestEntity{}); err != nil {
		t.Fatalf("Failed to migrate test entity: %v", err)
	}

	return db
}

// CreateTestEntities creates sample test entities for testing purposes
func CreateTestEntities() []*TestEntity {
	return []*TestEntity{
		{
			BaseEntity:  types.BaseEntity{ID: 1},
			Name:        "John Doe",
			Email:       "john@example.com",
			Age:         30,
			IsActive:    true,
			Description: "Test user 1",
			Status:      "active",
		},
		{
			BaseEntity:  types.BaseEntity{ID: 2},
			Name:        "Jane Smith",
			Email:       "jane@example.com",
			Age:         25,
			IsActive:    false,
			Description: "Test user 2",
			Status:      "inactive",
		},
		{
			BaseEntity:  types.BaseEntity{ID: 3},
			Name:        "Bob Johnson",
			Email:       "bob@example.com",
			Age:         35,
			IsActive:    true,
			Description: "Test user 3",
			Status:      "active",
		},
	}
}

// MockUnitOfWork provides a unified mock implementation for IUnitOfWork testing.
// This replaces all duplicate MockUnitOfWork implementations across the codebase.
type MockUnitOfWork struct {
	// Mock call tracking fields
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

	// Mock return values
	FindAllResult                  []*TestEntity
	FindAllWithPaginationResult    []*TestEntity
	FindAllWithPaginationCount     int64
	FindOneResult                  *TestEntity
	FindOneByIdResult              *TestEntity
	FindOneByIdentifierResult      *TestEntity
	InsertResult                   *TestEntity
	UpdateResult                   *TestEntity
	SoftDeleteResult               *TestEntity
	HardDeleteResult               *TestEntity
	BulkInsertResult               []*TestEntity
	BulkUpdateResult               []*TestEntity
	GetTrashedResult               []*TestEntity
	GetTrashedWithPaginationResult []*TestEntity
	GetTrashedWithPaginationCount  int64
	RestoreResult                  *TestEntity
	CountResult                    int64
	ExistsResult                   bool
	ResolveIDByUniqueFieldResult   int

	// Mock error values
	FindAllError                  error
	FindAllWithPaginationError    error
	FindOneError                  error
	FindOneByIdError              error
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

// MockUnitOfWork method implementations
func (m *MockUnitOfWork) FindAll(ctx context.Context) ([]*TestEntity, error) {
	m.FindAllCalled = true
	return m.FindAllResult, m.FindAllError
}

func (m *MockUnitOfWork) FindAllWithPagination(ctx context.Context, params interface{}) ([]*TestEntity, int64, error) {
	m.FindAllWithPaginationCalled = true
	return m.FindAllWithPaginationResult, m.FindAllWithPaginationCount, m.FindAllWithPaginationError
}

func (m *MockUnitOfWork) FindOne(ctx context.Context, filter *TestEntity) (*TestEntity, error) {
	m.FindOneCalled = true
	return m.FindOneResult, m.FindOneError
}

func (m *MockUnitOfWork) FindOneById(ctx context.Context, id int) (*TestEntity, error) {
	m.FindOneByIdCalled = true
	return m.FindOneByIdResult, m.FindOneByIdError
}

func (m *MockUnitOfWork) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (*TestEntity, error) {
	m.FindOneByIdentifierCalled = true
	return m.FindOneByIdentifierResult, m.FindOneByIdentifierError
}

func (m *MockUnitOfWork) Insert(ctx context.Context, entity *TestEntity) (*TestEntity, error) {
	m.InsertCalled = true
	return m.InsertResult, m.InsertError
}

func (m *MockUnitOfWork) Update(ctx context.Context, identifier identifier.IIdentifier, entity *TestEntity) (*TestEntity, error) {
	m.UpdateCalled = true
	return m.UpdateResult, m.UpdateError
}

func (m *MockUnitOfWork) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	m.DeleteCalled = true
	return m.DeleteError
}

func (m *MockUnitOfWork) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (*TestEntity, error) {
	m.SoftDeleteCalled = true
	return m.SoftDeleteResult, m.SoftDeleteError
}

func (m *MockUnitOfWork) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (*TestEntity, error) {
	m.HardDeleteCalled = true
	return m.HardDeleteResult, m.HardDeleteError
}

func (m *MockUnitOfWork) BulkInsert(ctx context.Context, entities []*TestEntity) ([]*TestEntity, error) {
	m.BulkInsertCalled = true
	return m.BulkInsertResult, m.BulkInsertError
}

func (m *MockUnitOfWork) BulkUpdate(ctx context.Context, entities []*TestEntity) ([]*TestEntity, error) {
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

func (m *MockUnitOfWork) GetTrashed(ctx context.Context) ([]*TestEntity, error) {
	m.GetTrashedCalled = true
	return m.GetTrashedResult, m.GetTrashedError
}

func (m *MockUnitOfWork) GetTrashedWithPagination(ctx context.Context, params interface{}) ([]*TestEntity, int64, error) {
	m.GetTrashedWithPaginationCalled = true
	return m.GetTrashedWithPaginationResult, m.GetTrashedWithPaginationCount, m.GetTrashedWithPaginationError
}

func (m *MockUnitOfWork) Restore(ctx context.Context, identifier identifier.IIdentifier) (*TestEntity, error) {
	m.RestoreCalled = true
	return m.RestoreResult, m.RestoreError
}

func (m *MockUnitOfWork) RestoreAll(ctx context.Context) error {
	m.RestoreAllCalled = true
	return m.RestoreAllError
}

func (m *MockUnitOfWork) Count(ctx context.Context, params interface{}) (int64, error) {
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
