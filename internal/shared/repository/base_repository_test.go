package repository

import (
	"context"
	"testing"

	"github.com/ai-shiraz-teams/go-database/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database/internal/shared/query"
	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

// TestBaseRepository_NewBaseRepository validates repository creation
func TestBaseRepository_NewBaseRepository(t *testing.T) {
	// Arrange
	mockUow := &mockUnitOfWork{}

	// Act
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)

	// Assert
	if repo == nil {
		t.Error("Expected repository to be created")
	}
}

// TestBaseRepository_FindAll validates delegation to UnitOfWork
func TestBaseRepository_FindAll(t *testing.T) {
	// Arrange
	mockUow := &mockUnitOfWork{
		FindAllResult: testutil.CreateTestEntities(),
	}
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)

	// Act
	result, err := repo.FindAll(context.Background())

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !mockUow.FindAllCalled {
		t.Error("Expected FindAll to be called on UnitOfWork")
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 entities, got: %d", len(result))
	}
}

// TestBaseRepository_FindAllWithPagination validates pagination delegation
func TestBaseRepository_FindAllWithPagination(t *testing.T) {
	// Arrange
	mockUow := &mockUnitOfWork{
		FindAllWithPaginationResult: testutil.CreateTestEntities()[:1],
		FindAllWithPaginationCount:  1,
	}
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)
	params := query.NewQueryParams[*testutil.TestEntity]()

	// Act
	result, count, err := repo.FindAllWithPagination(context.Background(), params)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !mockUow.FindAllWithPaginationCalled {
		t.Error("Expected FindAllWithPagination to be called on UnitOfWork")
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 entity, got: %d", len(result))
	}
	if count != 1 {
		t.Errorf("Expected count of 1, got: %d", count)
	}
}

// TestBaseRepository_Insert validates entity insertion delegation
func TestBaseRepository_Insert(t *testing.T) {
	// Arrange
	entity := testutil.CreateTestEntities()[0]
	mockUow := &mockUnitOfWork{
		InsertResult: entity,
	}
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)

	// Act
	result, err := repo.Insert(context.Background(), entity)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !mockUow.InsertCalled {
		t.Error("Expected Insert to be called on UnitOfWork")
	}
	if result != entity {
		t.Error("Expected same entity to be returned")
	}
}

// TestBaseRepository_Update validates entity update delegation
func TestBaseRepository_Update(t *testing.T) {
	// Arrange
	entity := testutil.CreateTestEntities()[0]
	id := identifier.NewIdentifier().Equal("id", 1)
	mockUow := &mockUnitOfWork{
		UpdateResult: entity,
	}
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)

	// Act
	result, err := repo.Update(context.Background(), id, entity)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !mockUow.UpdateCalled {
		t.Error("Expected Update to be called on UnitOfWork")
	}
	if result != entity {
		t.Error("Expected same entity to be returned")
	}
}
