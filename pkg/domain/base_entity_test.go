package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBaseEntity_Interface(t *testing.T) {
	// Test that BaseEntity implements IBaseModel interface
	var base IBaseModel = &BaseEntity{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Test interface methods
	if base.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", base.GetID())
	}

	if base.GetVersion() != 1 {
		t.Errorf("Expected Version 1, got %d", base.GetVersion())
	}

	// Test version update
	base.SetVersion(2)
	if base.GetVersion() != 2 {
		t.Errorf("Expected Version 2 after update, got %d", base.GetVersion())
	}
}

func TestBaseEntity_JSONSerialization(t *testing.T) {
	now := time.Now()
	entity := &BaseEntity{
		ID:        42,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	// Test JSON marshaling
	data, err := json.Marshal(entity)
	if err != nil {
		t.Fatalf("Failed to marshal BaseEntity: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled BaseEntity
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal BaseEntity: %v", err)
	}

	if unmarshaled.ID != entity.ID {
		t.Errorf("Expected ID %d, got %d", entity.ID, unmarshaled.ID)
	}

	if unmarshaled.Version != entity.Version {
		t.Errorf("Expected Version %d, got %d", entity.Version, unmarshaled.Version)
	}
}

func TestBaseEntity_SoftDelete(t *testing.T) {
	entity := &BaseEntity{
		ID:      1,
		Version: 1,
	}

	// Initially not deleted
	if entity.GetDeletedAt() != nil {
		t.Error("Expected DeletedAt to be nil for new entity")
	}

	// Simulate soft delete
	deleteTime := time.Now()
	entity.DeletedAt = &deleteTime

	if entity.GetDeletedAt() == nil {
		t.Error("Expected DeletedAt to be set after soft delete")
	}

	if !entity.GetDeletedAt().Equal(deleteTime) {
		t.Error("DeletedAt timestamp mismatch")
	}
}

func TestEmbeddedEntity_User(t *testing.T) {
	user := &User{
		BaseEntity: BaseEntity{
			ID:      100,
			Version: 1,
		},
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Test that embedded BaseEntity works correctly
	var baseModel IBaseModel = user
	if baseModel.GetID() != 100 {
		t.Errorf("Expected ID 100, got %d", baseModel.GetID())
	}

	// Test JSON serialization of embedded entity
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal User: %v", err)
	}

	var unmarshaled User
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal User: %v", err)
	}

	if unmarshaled.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, unmarshaled.Email)
	}

	if unmarshaled.GetID() != user.GetID() {
		t.Errorf("Expected ID %d, got %d", user.GetID(), unmarshaled.GetID())
	}
}

func TestAuditableEntity(t *testing.T) {
	auditable := &AuditableEntity{
		BaseEntity: BaseEntity{
			ID:      1,
			Version: 1,
		},
		CreatedBy: 123,
		UpdatedBy: 456,
		AuditNote: "Initial creation",
	}

	if auditable.GetCreatedBy() != 123 {
		t.Errorf("Expected CreatedBy 123, got %d", auditable.GetCreatedBy())
	}

	if auditable.GetUpdatedBy() != 456 {
		t.Errorf("Expected UpdatedBy 456, got %d", auditable.GetUpdatedBy())
	}

	if auditable.GetAuditNote() != "Initial creation" {
		t.Errorf("Expected AuditNote 'Initial creation', got %s", auditable.GetAuditNote())
	}

	// Test setter methods
	auditable.SetUpdatedBy(789)
	auditable.SetAuditNote("Updated by admin")

	if auditable.GetUpdatedBy() != 789 {
		t.Errorf("Expected UpdatedBy 789 after update, got %d", auditable.GetUpdatedBy())
	}

	if auditable.GetAuditNote() != "Updated by admin" {
		t.Errorf("Expected AuditNote 'Updated by admin', got %s", auditable.GetAuditNote())
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkBaseEntity_GetID(b *testing.B) {
	entity := &BaseEntity{ID: 1}
	for i := 0; i < b.N; i++ {
		_ = entity.GetID()
	}
}

func BenchmarkBaseEntity_SetVersion(b *testing.B) {
	entity := &BaseEntity{Version: 1}
	for i := 0; i < b.N; i++ {
		entity.SetVersion(i)
	}
}
