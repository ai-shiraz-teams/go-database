package repository

import (
	"context"
	"testing"

	"go-database/pkg/domain"
)

// TestUser represents a concrete model for testing
type TestUser struct {
	domain.BaseEntity
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// MockUnitOfWork provides a mock implementation of IUnitOfWork for testing
type MockUnitOfWork[T domain.IBaseModel] struct {
	entities []T
	inTx     bool
}

// Transaction management
func (m *MockUnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	m.inTx = true
	return nil
}

func (m *MockUnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	m.inTx = false
	return nil
}

func (m *MockUnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	m.inTx = false
}

// Basic queries
func (m *MockUnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	return m.entities, nil
}

func (m *MockUnitOfWork[T]) FindAllWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	return m.entities, uint(len(m.entities)), nil
}

func (m *MockUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	if len(m.entities) > 0 {
		return m.entities[0], nil
	}
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) FindOneById(ctx context.Context, id uint) (T, error) {
	for _, entity := range m.entities {
		if uint(entity.GetID()) == id {
			return entity, nil
		}
	}
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	if len(m.entities) > 0 {
		return m.entities[0], nil
	}
	var zero T
	return zero, nil
}

// Mutation operations
func (m *MockUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	m.entities = append(m.entities, entity)
	return entity, nil
}

func (m *MockUnitOfWork[T]) Update(ctx context.Context, identifier domain.IIdentifier, entity T) (T, error) {
	return entity, nil
}

func (m *MockUnitOfWork[T]) Delete(ctx context.Context, identifier domain.IIdentifier) error {
	return nil
}

// Soft-delete lifecycle
func (m *MockUnitOfWork[T]) SoftDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) HardDelete(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	return []T{}, nil
}

func (m *MockUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, query *domain.QueryParams[T]) ([]T, uint, error) {
	return []T{}, 0, nil
}

func (m *MockUnitOfWork[T]) Restore(ctx context.Context, identifier domain.IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	return nil
}

// Bulk operations
func (m *MockUnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	m.entities = append(m.entities, entities...)
	return entities, nil
}

func (m *MockUnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	return entities, nil
}

func (m *MockUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	return nil
}

func (m *MockUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []domain.IIdentifier) error {
	return nil
}

// Utility operations
func (m *MockUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (uint, error) {
	return 0, nil
}

func (m *MockUnitOfWork[T]) Count(ctx context.Context, query *domain.QueryParams[T]) (int64, error) {
	return int64(len(m.entities)), nil
}

func (m *MockUnitOfWork[T]) Exists(ctx context.Context, identifier domain.IIdentifier) (bool, error) {
	return len(m.entities) > 0, nil
}

// Ensure MockUnitOfWork implements IUnitOfWork
var _ domain.IUnitOfWork[*TestUser] = (*MockUnitOfWork[*TestUser])(nil)

// Test cases

func TestNewBaseRepository(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{}
	repo := NewBaseRepository(mockUOW)

	if repo == nil {
		t.Fatal("Expected non-nil repository instance")
	}

	// Verify the repository implements the interface
	var _ IBaseRepository[*TestUser] = repo
}

func TestBaseRepository_Insert(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	user := &TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}

	insertedUser, err := repo.Insert(ctx, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if insertedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, insertedUser.Email)
	}

	// Verify it was added to the mock
	if len(mockUOW.entities) != 1 {
		t.Errorf("Expected 1 entity in mock, got %d", len(mockUOW.entities))
	}
}

func TestBaseRepository_FindAll(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{
		entities: []*TestUser{
			{Email: "user1@example.com", Username: "user1", Name: "User One"},
			{Email: "user2@example.com", Username: "user2", Name: "User Two"},
		},
	}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	users, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Email != "user1@example.com" {
		t.Errorf("Expected first user email to be user1@example.com, got %s", users[0].Email)
	}
}

func TestBaseRepository_FindAllWithPagination(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{
		entities: []*TestUser{
			{Email: "user1@example.com", Username: "user1", Name: "User One"},
			{Email: "user2@example.com", Username: "user2", Name: "User Two"},
		},
	}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	params := domain.NewQueryParams[*TestUser]()
	params.Page = 1
	params.PageSize = 10
	params.PrepareDefaults(50, 100)

	users, total, err := repo.FindAllWithPagination(ctx, params)
	if err != nil {
		t.Fatalf("FindAllWithPagination failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if total != 2 {
		t.Errorf("Expected total count of 2, got %d", total)
	}
}

func TestBaseRepository_BulkInsert(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	users := []*TestUser{
		{Email: "bulk1@example.com", Username: "bulk1", Name: "Bulk User 1"},
		{Email: "bulk2@example.com", Username: "bulk2", Name: "Bulk User 2"},
		{Email: "bulk3@example.com", Username: "bulk3", Name: "Bulk User 3"},
	}

	insertedUsers, err := repo.BulkInsert(ctx, users)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	if len(insertedUsers) != 3 {
		t.Errorf("Expected 3 inserted users, got %d", len(insertedUsers))
	}

	// Verify they were added to the mock
	if len(mockUOW.entities) != 3 {
		t.Errorf("Expected 3 entities in mock, got %d", len(mockUOW.entities))
	}
}

func TestBaseRepository_Count(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{
		entities: []*TestUser{
			{Email: "user1@example.com", Username: "user1", Name: "User One"},
			{Email: "user2@example.com", Username: "user2", Name: "User Two"},
		},
	}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	params := domain.NewQueryParams[*TestUser]()
	count, err := repo.Count(ctx, params)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count of 2, got %d", count)
	}
}

func TestBaseRepository_Exists(t *testing.T) {
	mockUOW := &MockUnitOfWork[*TestUser]{
		entities: []*TestUser{
			{Email: "user1@example.com", Username: "user1", Name: "User One"},
		},
	}
	repo := NewBaseRepository(mockUOW)
	ctx := context.Background()

	identifier := domain.NewIdentifier().Equal("email", "user1@example.com")
	exists, err := repo.Exists(ctx, identifier)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("Expected entity to exist")
	}

	// Test with empty mock
	emptyMockUOW := &MockUnitOfWork[*TestUser]{}
	emptyRepo := NewBaseRepository(emptyMockUOW)

	exists, err = emptyRepo.Exists(ctx, identifier)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if exists {
		t.Error("Expected entity not to exist")
	}
}
