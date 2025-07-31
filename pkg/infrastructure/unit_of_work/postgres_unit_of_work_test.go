package unit_of_work

import (
	"context"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func TestNewPostgresUnitOfWork(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)

	// Act
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)

	// Assert
	if uow == nil {
		t.Fatal("Expected non-nil unit of work")
	}

	// Type assertion to verify correct implementation
	if _, ok := uow.(*PostgresUnitOfWork[*testutil.TestEntity]); !ok {
		t.Fatal("Expected PostgresUnitOfWork implementation")
	}
}

func TestPostgresUnitOfWork_Transaction_Management(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T, IUnitOfWork[*testutil.TestEntity]) error
		expectError bool
	}{
		{
			name: "Begin transaction successfully",
			setupFunc: func(t *testing.T, uow IUnitOfWork[*testutil.TestEntity]) error {
				return uow.BeginTransaction(context.Background())
			},
			expectError: false,
		},
		{
			name: "Fail to begin transaction when already in transaction",
			setupFunc: func(t *testing.T, uow IUnitOfWork[*testutil.TestEntity]) error {
				// Start first transaction
				if err := uow.BeginTransaction(context.Background()); err != nil {
					return err
				}
				// Try to start second transaction
				return uow.BeginTransaction(context.Background())
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)

			// Act
			err := tt.setupFunc(t, uow)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestPostgresUnitOfWork_CommitTransaction(t *testing.T) {
	tests := []struct {
		name           string
		hasTransaction bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "Commit active transaction successfully",
			hasTransaction: true,
			expectError:    false,
		},
		{
			name:           "Fail to commit when no active transaction",
			hasTransaction: false,
			expectError:    true,
			expectedErrMsg: "no active transaction to commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)

			if tt.hasTransaction {
				err := uow.BeginTransaction(context.Background())
				if err != nil {
					t.Fatalf("Failed to begin transaction: %v", err)
				}
			}

			// Act
			err := uow.CommitTransaction(context.Background())

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestPostgresUnitOfWork_RollbackTransaction(t *testing.T) {
	tests := []struct {
		name           string
		hasTransaction bool
	}{
		{"Rollback active transaction", true},
		{"Rollback when no transaction (should not panic)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)

			if tt.hasTransaction {
				err := uow.BeginTransaction(context.Background())
				if err != nil {
					t.Fatalf("Failed to begin transaction: %v", err)
				}
			}

			// Act (should not panic)
			uow.RollbackTransaction(context.Background())

			// Assert - test passes if no panic occurs
		})
	}
}

func TestPostgresUnitOfWork_Insert(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	// Act
	result, err := uow.Insert(ctx, entity)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.GetID() == 0 {
		t.Error("Expected ID to be set")
	}
	if result.Name != "Test Entity" {
		t.Errorf("Expected Name 'Test Entity', got '%s'", result.Name)
	}
}

func TestPostgresUnitOfWork_FindAll(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Insert test data
	entities := []*testutil.TestEntity{
		{Name: "Entity 1", Status: "active"},
		{Name: "Entity 2", Status: "active"},
		{Name: "Entity 3", Status: "inactive"},
	}

	for _, entity := range entities {
		_, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert test entity: %v", err)
		}
	}

	// Act
	results, err := uow.FindAll(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 entities, got %d", len(results))
	}
}

func TestPostgresUnitOfWork_FindOneById(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	// Act
	result, err := uow.FindOneById(ctx, inserted.GetID())

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.GetID() != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), result.GetID())
	}
	if result.Name != "Test Entity" {
		t.Errorf("Expected Name 'Test Entity', got '%s'", result.Name)
	}
}

func TestPostgresUnitOfWork_FindOneByIdentifier(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	identifierBuilder := identifier.NewIdentifier().Equal("name", "Test Entity")

	// Act
	result, err := uow.FindOneByIdentifier(ctx, identifierBuilder)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.GetID() != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), result.GetID())
	}
}

func TestPostgresUnitOfWork_Update(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Original Name",
		Description: "Original Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

	// Modify the entity
	inserted.Name = "Updated Name"
	inserted.Description = "Updated Description"

	// Act
	result, err := uow.Update(ctx, identifierBuilder, inserted)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got '%s'", result.Name)
	}
	if result.Description != "Updated Description" {
		t.Errorf("Expected Description 'Updated Description', got '%s'", result.Description)
	}
}

func TestPostgresUnitOfWork_SoftDelete(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

	// Act
	result, err := uow.SoftDelete(ctx, identifierBuilder)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify entity is soft-deleted by trying to find it
	_, findErr := uow.FindOneById(ctx, inserted.GetID())
	if findErr == nil {
		t.Error("Expected entity to be soft-deleted and not findable")
	}
}

func TestPostgresUnitOfWork_HardDelete(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

	// Act
	result, err := uow.HardDelete(ctx, identifierBuilder)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.GetID() != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), result.GetID())
	}

	// Verify entity is hard-deleted
	_, findErr := uow.FindOneById(ctx, inserted.GetID())
	if findErr == nil {
		t.Error("Expected entity to be hard-deleted and not findable")
	}
}

func TestPostgresUnitOfWork_GetTrashed(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Insert and soft-delete some entities
	entities := []*testutil.TestEntity{
		{Name: "Entity 1", Status: "active"},
		{Name: "Entity 2", Status: "active"},
	}

	for _, entity := range entities {
		inserted, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert entity: %v", err)
		}

		identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

		_, err = uow.SoftDelete(ctx, identifierBuilder)
		if err != nil {
			t.Fatalf("Failed to soft delete entity: %v", err)
		}
	}

	// Act
	trashedEntities, err := uow.GetTrashed(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(trashedEntities) != 2 {
		t.Errorf("Expected 2 trashed entities, got %d", len(trashedEntities))
	}
}

func TestPostgresUnitOfWork_Restore(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

	// Soft delete first
	_, err = uow.SoftDelete(ctx, identifierBuilder)
	if err != nil {
		t.Fatalf("Failed to soft delete entity: %v", err)
	}

	// Act
	result, err := uow.Restore(ctx, identifierBuilder)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify entity is restored and findable again
	found, findErr := uow.FindOneById(ctx, inserted.GetID())
	if findErr != nil {
		t.Fatalf("Expected to find restored entity, got error: %v", findErr)
	}
	if found.GetID() != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), found.GetID())
	}
}

func TestPostgresUnitOfWork_BulkInsert(t *testing.T) {
	tests := []struct {
		name          string
		entities      []*testutil.TestEntity
		expectError   bool
		expectedCount int
	}{
		{
			name: "Insert multiple entities successfully",
			entities: []*testutil.TestEntity{
				{Name: "Entity 1", Status: "active"},
				{Name: "Entity 2", Status: "active"},
				{Name: "Entity 3", Status: "inactive"},
			},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:          "Insert empty slice",
			entities:      []*testutil.TestEntity{},
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
			ctx := context.Background()

			// Act
			result, err := uow.BulkInsert(ctx, tt.entities)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d entities, got %d", tt.expectedCount, len(result))
			}

			// Verify all entities have IDs if successful
			if !tt.expectError && len(result) > 0 {
				for i, entity := range result {
					if entity.GetID() == 0 {
						t.Errorf("Entity %d should have ID set", i)
					}
				}
			}
		})
	}
}

func TestPostgresUnitOfWork_Count(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Insert test data
	entities := []*testutil.TestEntity{
		{Name: "Entity 1", Status: "active"},
		{Name: "Entity 2", Status: "active"},
		{Name: "Entity 3", Status: "inactive"},
	}

	for _, entity := range entities {
		_, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert test entity: %v", err)
		}
	}

	queryParams := query.NewQueryParams[*testutil.TestEntity]()

	// Act
	count, err := uow.Count(ctx, queryParams)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestPostgresUnitOfWork_Exists(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	tests := []struct {
		name       string
		identifier identifier.IIdentifier
		expected   bool
	}{
		{
			name:       "Entity exists",
			identifier: identifier.NewIdentifier().Equal("id", inserted.GetID()),
			expected:   true,
		},
		{
			name:       "Entity does not exist",
			identifier: identifier.NewIdentifier().Equal("id", 99999),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			exists, err := uow.Exists(ctx, tt.identifier)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("Expected exists %v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestPostgresUnitOfWork_FindAllWithPagination(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Insert test data
	for i := 1; i <= 10; i++ {
		entity := &testutil.TestEntity{
			Name:   "Entity " + string(rune(i+'0')),
			Status: "active",
		}
		_, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert test entity: %v", err)
		}
	}

	queryParams := query.NewQueryParams[*testutil.TestEntity]()
	queryParams.Limit = 5
	queryParams.Offset = 0

	// Act
	results, total, err := uow.FindAllWithPagination(ctx, queryParams)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if total != 10 {
		t.Errorf("Expected total 10, got %d", total)
	}
}

func TestPostgresUnitOfWork_Transaction_Lifecycle(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Act & Assert - Complete transaction lifecycle
	err := uow.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Insert entity within transaction
	entity := &testutil.TestEntity{
		Name:        "Transaction Entity",
		Description: "Test in transaction",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert entity in transaction: %v", err)
	}

	// Commit transaction
	err = uow.CommitTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify entity persisted after commit
	found, err := uow.FindOneById(ctx, inserted.GetID())
	if err != nil {
		t.Fatalf("Failed to find entity after commit: %v", err)
	}
	if found.GetID() != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), found.GetID())
	}
}

func TestPostgresUnitOfWork_Transaction_Rollback_Lifecycle(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Act & Assert - Transaction rollback lifecycle
	err := uow.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Insert entity within transaction
	entity := &testutil.TestEntity{
		Name:        "Rollback Entity",
		Description: "Test rollback",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert entity in transaction: %v", err)
	}

	// Rollback transaction
	uow.RollbackTransaction(ctx)

	// Verify entity was not persisted after rollback
	_, err = uow.FindOneById(ctx, inserted.GetID())
	if err == nil {
		t.Error("Expected entity to not exist after rollback, but it was found")
	}
}

func TestPostgresUnitOfWork_ResolveIDByUniqueField(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Unique Name",
		Description: "Test Description",
		Status:      "active",
	}

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	// Act
	resolvedID, err := uow.ResolveIDByUniqueField(ctx, inserted, "name", "Unique Name")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resolvedID != inserted.GetID() {
		t.Errorf("Expected ID %d, got %d", inserted.GetID(), resolvedID)
	}
}

func TestPostgresUnitOfWork_RestoreAll(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Insert and soft-delete multiple entities
	entities := []*testutil.TestEntity{
		{Name: "Entity 1", Status: "active"},
		{Name: "Entity 2", Status: "active"},
		{Name: "Entity 3", Status: "active"},
	}

	var insertedIDs []int
	for _, entity := range entities {
		inserted, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert entity: %v", err)
		}
		insertedIDs = append(insertedIDs, inserted.GetID())

		identifierBuilder := identifier.NewIdentifier().Equal("id", inserted.GetID())

		_, err = uow.SoftDelete(ctx, identifierBuilder)
		if err != nil {
			t.Fatalf("Failed to soft delete entity: %v", err)
		}
	}

	// Act
	err := uow.RestoreAll(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all entities are restored
	for _, id := range insertedIDs {
		found, findErr := uow.FindOneById(ctx, id)
		if findErr != nil {
			t.Errorf("Expected to find restored entity with ID %d, got error: %v", id, findErr)
		}
		if found.GetID() != id {
			t.Errorf("Expected ID %d, got %d", id, found.GetID())
		}
	}
}

func TestPostgresUnitOfWork_Error_Cases(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(*testing.T, IUnitOfWork[*testutil.TestEntity])
		expectError bool
	}{
		{
			name: "FindOneById with non-existent ID",
			testFunc: func(t *testing.T, uow IUnitOfWork[*testutil.TestEntity]) {
				_, err := uow.FindOneById(context.Background(), 99999)
				if err == nil {
					t.Error("Expected error for non-existent ID")
				}
			},
			expectError: true,
		},
		{
			name: "Update non-existent entity",
			testFunc: func(t *testing.T, uow IUnitOfWork[*testutil.TestEntity]) {
				identifierBuilder := identifier.NewIdentifier().Equal("id", 99999)
				entity := &testutil.TestEntity{Name: "Non-existent"}
				_, err := uow.Update(context.Background(), identifierBuilder, entity)
				if err == nil {
					t.Error("Expected error for updating non-existent entity")
				}
			},
			expectError: true,
		},
		{
			name: "SoftDelete non-existent entity",
			testFunc: func(t *testing.T, uow IUnitOfWork[*testutil.TestEntity]) {
				identifierBuilder := identifier.NewIdentifier().Equal("id", 99999)
				_, err := uow.SoftDelete(context.Background(), identifierBuilder)
				if err == nil {
					t.Error("Expected error for soft deleting non-existent entity")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)

			// Act & Assert
			tt.testFunc(t, uow)
		})
	}
}
