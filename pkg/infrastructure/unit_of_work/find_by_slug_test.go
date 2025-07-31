package unit_of_work

import (
	"context"
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

// TestPostgresUnitOfWork_FindOneBySlug validates slug-based entity retrieval
func TestPostgresUnitOfWork_FindOneBySlug(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entity := &testutil.TestEntity{
		Name:        "Test Entity",
		Description: "Test Description",
		Status:      "active",
	}
	entity.Slug = "test-entity-slug"

	inserted, err := uow.Insert(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	// Act
	result, err := uow.FindOneBySlug(ctx, "test-entity-slug")

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
	if result.Slug != "test-entity-slug" {
		t.Errorf("Expected Slug 'test-entity-slug', got '%s'", result.Slug)
	}
	if result.Name != "Test Entity" {
		t.Errorf("Expected Name 'Test Entity', got '%s'", result.Name)
	}
}

// TestPostgresUnitOfWork_FindOneBySlug_NotFound validates handling of non-existent slugs
func TestPostgresUnitOfWork_FindOneBySlug_NotFound(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	// Act
	_, err := uow.FindOneBySlug(ctx, "non-existent-slug")

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent slug")
	}
}
