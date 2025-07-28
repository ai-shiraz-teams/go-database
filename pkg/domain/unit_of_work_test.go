package domain

import (
	"context"
	"testing"
	"time"
)

// TestUser demonstrates a concrete implementation for testing
type TestUser struct {
	BaseEntity
	Email    string `json:"email"`
	Username string `json:"username"`
}

// TestQueryParamsCreation tests the creation and default values of QueryParams
func TestQueryParamsCreation(t *testing.T) {
	// Test default creation
	params := NewQueryParams[*TestUser]()

	if params.Page != 1 {
		t.Errorf("Expected default page to be 1, got %d", params.Page)
	}

	if params.PageSize != 50 {
		t.Errorf("Expected default page size to be 50, got %d", params.PageSize)
	}

	if params.IncludeDeleted != false {
		t.Errorf("Expected IncludeDeleted to be false by default")
	}

	if params.OnlyDeleted != false {
		t.Errorf("Expected OnlyDeleted to be false by default")
	}
}

// TestQueryParamsPrepareDefaults tests the PrepareDefaults method
func TestQueryParamsPrepareDefaults(t *testing.T) {
	params := &QueryParams[*TestUser]{
		Page:     0, // Invalid page
		PageSize: 0, // Invalid page size
	}

	params.PrepareDefaults(25, 100)

	if params.Page != 1 {
		t.Errorf("Expected page to be corrected to 1, got %d", params.Page)
	}

	if params.PageSize != 25 {
		t.Errorf("Expected page size to be set to default 25, got %d", params.PageSize)
	}

	if params.Offset != 0 {
		t.Errorf("Expected offset to be 0 for first page, got %d", params.Offset)
	}

	if params.Limit != 25 {
		t.Errorf("Expected limit to equal page size, got %d", params.Limit)
	}

	// Test page size capping
	params.PageSize = 150 // Above max limit
	params.PrepareDefaults(25, 100)

	if params.PageSize != 100 {
		t.Errorf("Expected page size to be capped at 100, got %d", params.PageSize)
	}
}

// TestQueryParamsChaining tests the fluent API methods
func TestQueryParamsChaining(t *testing.T) {
	params := NewQueryParams[*TestUser]().
		AddSortAsc("email").
		AddSortDesc("createdAt").
		WithSearch("test@example.com").
		IncludeDeletedRecords().
		AddPreload("Profile")

	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}

	if params.Sort[0].Field != "email" || params.Sort[0].Order != SortOrderAsc {
		t.Errorf("Expected first sort to be email ASC")
	}

	if params.Sort[1].Field != "createdAt" || params.Sort[1].Order != SortOrderDesc {
		t.Errorf("Expected second sort to be createdAt DESC")
	}

	if params.Search != "test@example.com" {
		t.Errorf("Expected search term to be set")
	}

	if !params.IncludeDeleted {
		t.Errorf("Expected IncludeDeleted to be true")
	}

	if len(params.Preloads) != 1 || params.Preloads[0] != "Profile" {
		t.Errorf("Expected preload to contain 'Profile'")
	}
}

// TestQueryParamsClone tests the deep copy functionality
func TestQueryParamsClone(t *testing.T) {
	original := NewQueryParams[*TestUser]().
		AddSortAsc("name").
		WithSearch("test")

	clone := original.Clone()

	// Modify original
	original.AddSortDesc("email")
	original.WithSearch("modified")

	// Clone should remain unchanged
	if len(clone.Sort) != 1 {
		t.Errorf("Clone should have 1 sort field, got %d", len(clone.Sort))
	}

	if clone.Search != "test" {
		t.Errorf("Clone search should remain 'test', got '%s'", clone.Search)
	}
}

// TestIdentifierBuilder tests the IIdentifier implementation
func TestIdentifierBuilder(t *testing.T) {
	// Test basic filters
	identifier := NewIdentifier().
		Equal("email", "test@example.com").
		GreaterThan("createdAt", time.Now().Add(-24*time.Hour))

	criteria := identifier.ToFilterCriteria()

	if len(criteria) != 2 {
		t.Errorf("Expected 2 filter criteria, got %d", len(criteria))
	}

	if criteria[0].Field != "email" || criteria[0].Operator != FilterOperatorEqual {
		t.Errorf("First criteria should be email equals")
	}

	if criteria[1].Field != "createdAt" || criteria[1].Operator != FilterOperatorGreaterThan {
		t.Errorf("Second criteria should be createdAt greater than")
	}
}

// TestIdentifierBuilderComplexFilters tests complex filter scenarios
func TestIdentifierBuilderComplexFilters(t *testing.T) {
	// Test IN operator
	identifier := NewIdentifier().
		In("status", []interface{}{"active", "pending", "inactive"})

	criteria := identifier.ToFilterCriteria()
	if len(criteria) != 1 {
		t.Errorf("Expected 1 criteria for IN operator")
	}

	if criteria[0].Operator != FilterOperatorIn {
		t.Errorf("Expected IN operator")
	}

	if len(criteria[0].Values) != 3 {
		t.Errorf("Expected 3 values for IN operator, got %d", len(criteria[0].Values))
	}

	// Test BETWEEN operator
	identifier = NewIdentifier().
		Between("price", 10.0, 100.0)

	criteria = identifier.ToFilterCriteria()
	if criteria[0].Operator != FilterOperatorBetween {
		t.Errorf("Expected BETWEEN operator")
	}

	if len(criteria[0].Values) != 2 {
		t.Errorf("Expected 2 values for BETWEEN operator")
	}
}

// TestIdentifierBuilderLogicalOperators tests AND/OR combinations
func TestIdentifierBuilderLogicalOperators(t *testing.T) {
	// Create two separate identifiers
	identifier1 := NewIdentifier().Equal("status", "active")
	identifier2 := NewIdentifier().GreaterThan("createdAt", time.Now())

	// Combine with AND
	combined := identifier1.And(identifier2)
	criteria := combined.ToFilterCriteria()

	if len(criteria) != 2 {
		t.Errorf("Expected 2 criteria after AND combination, got %d", len(criteria))
	}

	// The first criteria should have AND as logical operator
	if criteria[0].LogicalOp != LogicalOperatorAnd {
		t.Errorf("Expected first criteria to have AND logical operator")
	}

	// Test OR combination
	combinedOr := identifier1.Or(identifier2)
	criteriaOr := combinedOr.ToFilterCriteria()

	if len(criteriaOr) != 2 {
		t.Errorf("Expected 2 criteria after OR combination")
	}

	if criteriaOr[0].LogicalOp != LogicalOperatorOr {
		t.Errorf("Expected first criteria to have OR logical operator")
	}
}

// TestIdentifierBuilderImmutability tests that operations return new instances
func TestIdentifierBuilderImmutability(t *testing.T) {
	original := NewIdentifier().Equal("field1", "value1")
	modified := original.Equal("field2", "value2")

	originalCriteria := original.ToFilterCriteria()
	modifiedCriteria := modified.ToFilterCriteria()

	if len(originalCriteria) != 1 {
		t.Errorf("Original should have 1 criteria, got %d", len(originalCriteria))
	}

	if len(modifiedCriteria) != 2 {
		t.Errorf("Modified should have 2 criteria, got %d", len(modifiedCriteria))
	}
}

// TestQueryParamsWithFilters tests integration between QueryParams and IIdentifier
func TestQueryParamsWithFilters(t *testing.T) {
	identifier := NewIdentifier().
		Equal("status", "active").
		GreaterThan("createdAt", time.Now().Add(-24*time.Hour))

	params := NewQueryParams[*TestUser]().
		WithFilters(identifier)

	if len(params.Filters) != 2 {
		t.Errorf("Expected 2 filters from identifier, got %d", len(params.Filters))
	}

	// Test that filters were properly copied
	if params.Filters[0].Field != "status" {
		t.Errorf("Expected first filter field to be 'status'")
	}

	if params.Filters[1].Field != "createdAt" {
		t.Errorf("Expected second filter field to be 'createdAt'")
	}
}

// TestQueryParamsToListOptions tests backward compatibility conversion
func TestQueryParamsToListOptions(t *testing.T) {
	params := NewQueryParams[*TestUser]()
	params.Page = 2
	params.PageSize = 25
	params.AddSortDesc("email")
	params.IncludeDeletedRecords()
	params.PrepareDefaults(50, 100)

	listOpts := params.ToListOptions()

	if listOpts.Limit != 25 {
		t.Errorf("Expected limit to be 25, got %d", listOpts.Limit)
	}

	if listOpts.Offset != 25 {
		t.Errorf("Expected offset to be 25 (page 2), got %d", listOpts.Offset)
	}

	if listOpts.SortBy != "email" {
		t.Errorf("Expected sort by email, got %s", listOpts.SortBy)
	}

	if listOpts.SortOrder != "desc" {
		t.Errorf("Expected sort order desc, got %s", listOpts.SortOrder)
	}

	if !listOpts.IncludeDeleted {
		t.Errorf("Expected IncludeDeleted to be true")
	}
}

// MockUnitOfWork demonstrates how the IUnitOfWork interface would be implemented
type MockUnitOfWork[T IBaseModel] struct {
	entities []T
	inTx     bool
}

// Implement a few key methods to demonstrate interface compliance
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

func (m *MockUnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	return m.entities, nil
}

func (m *MockUnitOfWork[T]) FindAllWithPagination(ctx context.Context, query *QueryParams[T]) ([]T, uint, error) {
	return m.entities, uint(len(m.entities)), nil
}

func (m *MockUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	m.entities = append(m.entities, entity)
	return entity, nil
}

// Stub implementations for the remaining interface methods
func (m *MockUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) FindOneById(ctx context.Context, id uint) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) Update(ctx context.Context, identifier IIdentifier, entity T) (T, error) {
	return entity, nil
}

func (m *MockUnitOfWork[T]) Delete(ctx context.Context, identifier IIdentifier) error {
	return nil
}

func (m *MockUnitOfWork[T]) SoftDelete(ctx context.Context, identifier IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) HardDelete(ctx context.Context, identifier IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	return []T{}, nil
}

func (m *MockUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, query *QueryParams[T]) ([]T, uint, error) {
	return []T{}, 0, nil
}

func (m *MockUnitOfWork[T]) Restore(ctx context.Context, identifier IIdentifier) (T, error) {
	var zero T
	return zero, nil
}

func (m *MockUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	return nil
}

func (m *MockUnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	return entities, nil
}

func (m *MockUnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	return entities, nil
}

func (m *MockUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []IIdentifier) error {
	return nil
}

func (m *MockUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []IIdentifier) error {
	return nil
}

func (m *MockUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model IBaseModel, field string, value interface{}) (uint, error) {
	return 0, nil
}

func (m *MockUnitOfWork[T]) Count(ctx context.Context, query *QueryParams[T]) (int64, error) {
	return 0, nil
}

func (m *MockUnitOfWork[T]) Exists(ctx context.Context, identifier IIdentifier) (bool, error) {
	return false, nil
}

// Compile-time check to ensure MockUnitOfWork implements IUnitOfWork
var _ IUnitOfWork[*TestUser] = (*MockUnitOfWork[*TestUser])(nil)

// TestUnitOfWorkInterface tests that the interface can be implemented and used
func TestUnitOfWorkInterface(t *testing.T) {
	uow := &MockUnitOfWork[*TestUser]{}

	ctx := context.Background()

	// Test transaction management
	err := uow.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("BeginTransaction failed: %v", err)
	}

	if !uow.inTx {
		t.Errorf("Expected to be in transaction")
	}

	err = uow.CommitTransaction(ctx)
	if err != nil {
		t.Errorf("CommitTransaction failed: %v", err)
	}

	if uow.inTx {
		t.Errorf("Expected transaction to be committed")
	}

	// Test basic operations
	user := &TestUser{
		Email:    "test@example.com",
		Username: "testuser",
	}

	insertedUser, err := uow.Insert(ctx, user)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	if insertedUser.Email != user.Email {
		t.Errorf("Inserted user email mismatch")
	}

	// Test query with params
	params := NewQueryParams[*TestUser]()
	users, count, err := uow.FindAllWithPagination(ctx, params)
	if err != nil {
		t.Errorf("FindAllWithPagination failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}
