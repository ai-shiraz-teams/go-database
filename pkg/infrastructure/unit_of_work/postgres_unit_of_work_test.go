package unit_of_work

import (
	"context"
	"testing"

	"go-database/pkg/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser is a concrete implementation for testing
type TestUser struct {
	domain.BaseEntity
	Email    string `gorm:"uniqueIndex" json:"email"`
	Username string `gorm:"uniqueIndex" json:"username"`
	Name     string `json:"name"`
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the test table
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate test table: %v", err)
	}

	return db
}

func TestNewPostgresUnitOfWork(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)

	if uow == nil {
		t.Fatal("Expected non-nil UnitOfWork instance")
	}
}

func TestPostgresUnitOfWork_BasicCRUD(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)
	ctx := context.Background()

	// Test Insert
	user := &TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}

	insertedUser, err := uow.Insert(ctx, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if insertedUser.GetID() == 0 {
		t.Error("Expected ID to be set after insert")
	}

	// Test FindOneById
	foundUser, err := uow.FindOneById(ctx, uint(insertedUser.GetID()))
	if err != nil {
		t.Fatalf("FindOneById failed: %v", err)
	}

	if foundUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
	}

	// Test Update with identifier
	identifier := domain.NewIdentifier().Equal("id", insertedUser.GetID())
	foundUser.Name = "Updated Name"

	updatedUser, err := uow.Update(ctx, identifier, foundUser)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updatedUser.Name != "Updated Name" {
		t.Errorf("Expected name to be updated to 'Updated Name', got %s", updatedUser.Name)
	}

	// Test SoftDelete
	deleteIdentifier := domain.NewIdentifier().Equal("id", insertedUser.GetID())
	deletedUser, err := uow.SoftDelete(ctx, deleteIdentifier)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	if deletedUser.GetID() != insertedUser.GetID() {
		t.Error("Deleted user ID should match original")
	}

	// Verify soft delete - should not be found in normal queries
	_, err = uow.FindOneById(ctx, uint(insertedUser.GetID()))
	if err == nil {
		t.Error("Expected error when finding soft-deleted entity")
	}

	// Test GetTrashed
	trashedUsers, err := uow.GetTrashed(ctx)
	if err != nil {
		t.Fatalf("GetTrashed failed: %v", err)
	}

	if len(trashedUsers) != 1 {
		t.Errorf("Expected 1 trashed user, got %d", len(trashedUsers))
	}

	// Test Restore
	restoreIdentifier := domain.NewIdentifier().Equal("id", insertedUser.GetID())
	restoredUser, err := uow.Restore(ctx, restoreIdentifier)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if restoredUser.GetID() != insertedUser.GetID() {
		t.Error("Restored user ID should match original")
	}

	// Verify restore - should now be found in normal queries
	_, err = uow.FindOneById(ctx, uint(insertedUser.GetID()))
	if err != nil {
		t.Errorf("Expected to find restored entity: %v", err)
	}
}

func TestPostgresUnitOfWork_Transactions(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)
	ctx := context.Background()

	// Test successful transaction
	err := uow.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}

	user := &TestUser{
		Email:    "tx@example.com",
		Username: "txuser",
		Name:     "Transaction User",
	}

	insertedUser, err := uow.Insert(ctx, user)
	if err != nil {
		t.Fatalf("Insert in transaction failed: %v", err)
	}

	err = uow.CommitTransaction(ctx)
	if err != nil {
		t.Fatalf("CommitTransaction failed: %v", err)
	}

	// Verify the user was committed
	_, err = uow.FindOneById(ctx, uint(insertedUser.GetID()))
	if err != nil {
		t.Errorf("Expected to find committed user: %v", err)
	}

	// Test rollback transaction
	err = uow.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}

	user2 := &TestUser{
		Email:    "rollback@example.com",
		Username: "rollbackuser",
		Name:     "Rollback User",
	}

	insertedUser2, err := uow.Insert(ctx, user2)
	if err != nil {
		t.Fatalf("Insert in transaction failed: %v", err)
	}

	uow.RollbackTransaction(ctx)

	// Verify the user was rolled back
	_, err = uow.FindOneById(ctx, uint(insertedUser2.GetID()))
	if err == nil {
		t.Error("Expected not to find rolled back user")
	}
}

func TestPostgresUnitOfWork_QueryParams(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)
	ctx := context.Background()

	// Insert test data
	users := []*TestUser{
		{BaseEntity: domain.BaseEntity{}, Email: "alice@example.com", Username: "alice", Name: "Alice"},
		{BaseEntity: domain.BaseEntity{}, Email: "bob@example.com", Username: "bob", Name: "Bob"},
		{BaseEntity: domain.BaseEntity{}, Email: "charlie@example.com", Username: "charlie", Name: "Charlie"},
	}

	for _, user := range users {
		_, err := uow.Insert(ctx, user)
		if err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}
	}

	// Test pagination
	params := domain.NewQueryParams[*TestUser]()
	params.Page = 1
	params.PageSize = 2
	params.PrepareDefaults(10, 100)

	results, total, err := uow.FindAllWithPagination(ctx, params)
	if err != nil {
		t.Fatalf("FindAllWithPagination failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if total != 3 {
		t.Errorf("Expected total count of 3, got %d", total)
	}

	// Test filtering with identifier
	identifier := domain.NewIdentifier().Like("email", "%alice%")
	params.WithFilters(identifier)

	filteredResults, filteredTotal, err := uow.FindAllWithPagination(ctx, params)
	if err != nil {
		t.Fatalf("Filtered pagination failed: %v", err)
	}

	if len(filteredResults) != 1 {
		t.Errorf("Expected 1 filtered result, got %d", len(filteredResults))
	}

	if filteredTotal != 1 {
		t.Errorf("Expected filtered total count of 1, got %d", filteredTotal)
	}

	if filteredResults[0].Email != "alice@example.com" {
		t.Errorf("Expected alice@example.com, got %s", filteredResults[0].Email)
	}
}

func TestPostgresUnitOfWork_BulkOperations(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)
	ctx := context.Background()

	// Test BulkInsert
	users := []*TestUser{
		{BaseEntity: domain.BaseEntity{}, Email: "bulk1@example.com", Username: "bulk1", Name: "Bulk User 1"},
		{BaseEntity: domain.BaseEntity{}, Email: "bulk2@example.com", Username: "bulk2", Name: "Bulk User 2"},
		{BaseEntity: domain.BaseEntity{}, Email: "bulk3@example.com", Username: "bulk3", Name: "Bulk User 3"},
	}

	insertedUsers, err := uow.BulkInsert(ctx, users)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	if len(insertedUsers) != 3 {
		t.Errorf("Expected 3 inserted users, got %d", len(insertedUsers))
	}

	for _, user := range insertedUsers {
		if user.GetID() == 0 {
			t.Error("Expected ID to be set after bulk insert")
		}
	}

	// Test Count
	params := domain.NewQueryParams[*TestUser]()
	count, err := uow.Count(ctx, params)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}

	// Test Exists
	identifier := domain.NewIdentifier().Equal("email", "bulk1@example.com")
	exists, err := uow.Exists(ctx, identifier)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("Expected entity to exist")
	}

	// Test BulkSoftDelete
	identifiers := []domain.IIdentifier{
		domain.NewIdentifier().Equal("email", "bulk1@example.com"),
		domain.NewIdentifier().Equal("email", "bulk2@example.com"),
	}

	err = uow.BulkSoftDelete(ctx, identifiers)
	if err != nil {
		t.Fatalf("BulkSoftDelete failed: %v", err)
	}

	// Verify soft deletion
	activeCount, err := uow.Count(ctx, params)
	if err != nil {
		t.Fatalf("Count after soft delete failed: %v", err)
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active user after soft delete, got %d", activeCount)
	}
}

func TestPostgresUnitOfWork_ComplexFiltering(t *testing.T) {
	db := setupTestDB(t)
	uow := NewPostgresUnitOfWork[*TestUser](db)
	ctx := context.Background()

	// Insert test data
	users := []*TestUser{
		{BaseEntity: domain.BaseEntity{}, Email: "admin@company.com", Username: "admin", Name: "Admin User"},
		{BaseEntity: domain.BaseEntity{}, Email: "user@company.com", Username: "user", Name: "Regular User"},
		{BaseEntity: domain.BaseEntity{}, Email: "guest@external.com", Username: "guest", Name: "Guest User"},
	}

	for _, user := range users {
		_, err := uow.Insert(ctx, user)
		if err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}
	}

	// Test complex filter: (email LIKE '%@company.com' AND name != 'Admin User')
	emailFilter := domain.NewIdentifier().Like("email", "%@company.com")
	nameFilter := domain.NewIdentifier().NotEqual("name", "Admin User")
	complexFilter := emailFilter.And(nameFilter)

	results, err := uow.FindOneByIdentifier(ctx, complexFilter)
	if err != nil {
		t.Fatalf("Complex filtering failed: %v", err)
	}

	if results.Email != "user@company.com" {
		t.Errorf("Expected user@company.com, got %s", results.Email)
	}

	// Test OR filter: (username = 'admin' OR username = 'guest')
	orFilter := domain.NewIdentifier().
		Equal("username", "admin").
		Or(domain.NewIdentifier().Equal("username", "guest"))

	params := domain.NewQueryParams[*TestUser]()
	params.WithFilters(orFilter)

	orResults, total, err := uow.FindAllWithPagination(ctx, params)
	if err != nil {
		t.Fatalf("OR filtering failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 results from OR filter, got %d", total)
	}

	if len(orResults) != 2 {
		t.Errorf("Expected 2 results, got %d", len(orResults))
	}
}
