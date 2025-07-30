package types

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestBaseEntity_GetID(t *testing.T) {
	// Arrange
	entity := BaseEntity{ID: 42}

	// Act
	result := entity.GetID()

	// Assert
	if result != 42 {
		t.Errorf("Expected ID 42, got %d", result)
	}
}

func TestBaseEntity_GetCreatedAt(t *testing.T) {
	// Arrange
	now := time.Now()
	entity := BaseEntity{CreatedAt: now}

	// Act
	result := entity.GetCreatedAt()

	// Assert
	if !result.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, result)
	}
}

func TestBaseEntity_GetUpdatedAt(t *testing.T) {
	// Arrange
	now := time.Now()
	entity := BaseEntity{UpdatedAt: now}

	// Act
	result := entity.GetUpdatedAt()

	// Assert
	if !result.Equal(now) {
		t.Errorf("Expected UpdatedAt %v, got %v", now, result)
	}
}

func TestBaseEntity_GetDeletedAt_WithValue(t *testing.T) {
	// Arrange
	now := time.Now()
	entity := BaseEntity{DeletedAt: gorm.DeletedAt{Time: now, Valid: true}}

	// Act
	result := entity.GetDeletedAt()

	// Assert
	if result == nil {
		t.Fatal("Expected non-nil DeletedAt")
	}
	if !result.Equal(now) {
		t.Errorf("Expected DeletedAt %v, got %v", now, *result)
	}
}

func TestBaseEntity_GetDeletedAt_Nil(t *testing.T) {
	// Arrange
	entity := BaseEntity{DeletedAt: gorm.DeletedAt{Valid: false}}

	// Act
	result := entity.GetDeletedAt()

	// Assert
	if result != nil {
		t.Errorf("Expected nil DeletedAt, got %v", *result)
	}
}

func TestBaseEntity_GetVersion(t *testing.T) {
	// Arrange
	entity := BaseEntity{Version: 5}

	// Act
	result := entity.GetVersion()

	// Assert
	if result != 5 {
		t.Errorf("Expected Version 5, got %d", result)
	}
}

func TestBaseEntity_SetVersion(t *testing.T) {
	// Arrange
	entity := BaseEntity{Version: 1}

	// Act
	entity.SetVersion(10)

	// Assert
	if entity.Version != 10 {
		t.Errorf("Expected Version 10, got %d", entity.Version)
	}
}

func TestBaseEntity_ImplementsIBaseModel(t *testing.T) {
	// Arrange
	entity := BaseEntity{}

	// Act & Assert - Compile-time check
	var _ IBaseModel = &entity

	// Additional runtime verification
	if entity.GetID() != 0 {
		t.Errorf("Expected default ID 0, got %d", entity.GetID())
	}
	if entity.GetVersion() != 0 {
		t.Errorf("Expected default Version 0, got %d", entity.GetVersion())
	}
}

func TestBaseEntity_DefaultValues(t *testing.T) {
	// Arrange
	entity := BaseEntity{}

	// Act & Assert
	if entity.GetID() != 0 {
		t.Errorf("Expected default ID 0, got %d", entity.GetID())
	}
	if entity.GetVersion() != 0 {
		t.Errorf("Expected default Version 0, got %d", entity.GetVersion())
	}
	if entity.GetDeletedAt() != nil {
		t.Errorf("Expected nil DeletedAt by default, got %v", entity.GetDeletedAt())
	}
	// CreatedAt and UpdatedAt will be zero time by default, which is expected
}

func TestBaseEntity_CompleteLifecycle(t *testing.T) {
	// Arrange
	entity := BaseEntity{}

	// Act - Simulate GORM auto-population
	now := time.Now()
	entity.ID = 1
	entity.CreatedAt = now
	entity.UpdatedAt = now
	entity.Version = 1

	// Assert initial state
	if entity.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", entity.GetID())
	}
	if entity.GetVersion() != 1 {
		t.Errorf("Expected Version 1, got %d", entity.GetVersion())
	}
	if entity.GetDeletedAt() != nil {
		t.Errorf("Expected nil DeletedAt, got %v", entity.GetDeletedAt())
	}

	// Act - Update scenario
	laterTime := now.Add(time.Hour)
	entity.UpdatedAt = laterTime
	entity.SetVersion(2)

	// Assert updated state
	if entity.GetVersion() != 2 {
		t.Errorf("Expected Version 2, got %d", entity.GetVersion())
	}
	if !entity.GetUpdatedAt().Equal(laterTime) {
		t.Errorf("Expected UpdatedAt %v, got %v", laterTime, entity.GetUpdatedAt())
	}

	// Act - Soft delete
	deleteTime := now.Add(2 * time.Hour)
	entity.DeletedAt = gorm.DeletedAt{Time: deleteTime, Valid: true}

	// Assert soft-deleted state
	deletedAt := entity.GetDeletedAt()
	if deletedAt == nil {
		t.Fatal("Expected non-nil DeletedAt after soft delete")
	}
	if !deletedAt.Equal(deleteTime) {
		t.Errorf("Expected DeletedAt %v, got %v", deleteTime, *deletedAt)
	}
}

func TestBaseEntity_VersionOptimisticLocking(t *testing.T) {
	tests := []struct {
		name           string
		initialVersion int
		newVersion     int
		expectedResult int
	}{
		{"Version increment", 1, 2, 2},
		{"Version reset", 5, 1, 1},
		{"Version zero", 0, 1, 1},
		{"Large version", 999, 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := BaseEntity{Version: tt.initialVersion}

			// Act
			entity.SetVersion(tt.newVersion)

			// Assert
			if entity.GetVersion() != tt.expectedResult {
				t.Errorf("Expected Version %d, got %d", tt.expectedResult, entity.GetVersion())
			}
		})
	}
}

func TestBaseEntity_ImmutabilityOfGetters(t *testing.T) {
	// Arrange
	now := time.Now()
	entity := BaseEntity{
		ID:        10,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   5,
	}

	// Act - Get values
	id1 := entity.GetID()
	version1 := entity.GetVersion()
	createdAt1 := entity.GetCreatedAt()
	updatedAt1 := entity.GetUpdatedAt()

	// Modify returned time values (should not affect entity) - use values to avoid staticcheck
	_ = createdAt1.Add(time.Hour)
	_ = updatedAt1.Add(time.Hour)

	// Act - Get values again
	id2 := entity.GetID()
	version2 := entity.GetVersion()
	createdAt2 := entity.GetCreatedAt()
	updatedAt2 := entity.GetUpdatedAt()

	// Assert - Original values unchanged
	if id1 != id2 {
		t.Errorf("ID changed: expected %d, got %d", id1, id2)
	}
	if version1 != version2 {
		t.Errorf("Version changed: expected %d, got %d", version1, version2)
	}
	if !createdAt2.Equal(now) {
		t.Errorf("CreatedAt changed: expected %v, got %v", now, createdAt2)
	}
	if !updatedAt2.Equal(now) {
		t.Errorf("UpdatedAt changed: expected %v, got %v", now, updatedAt2)
	}
}
