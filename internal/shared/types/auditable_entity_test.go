package types

import (
	"testing"
	"time"
)

// TestAuditableEntity_EmbeddedFields validates BaseEntity embedding
func TestAuditableEntity_EmbeddedFields(t *testing.T) {
	// Arrange & Act
	entity := AuditableEntity{}

	// Assert - BaseEntity fields should be accessible
	if entity.ID != 0 {
		t.Errorf("Expected default ID 0, got %d", entity.ID)
	}

	if entity.Version != 0 {
		t.Errorf("Expected default Version 0, got %d", entity.Version)
	}

	// Verify it implements IBaseModel through BaseEntity
	var _ IBaseModel = &entity
}

// TestAuditableEntity_GetCreatedBy validates CreatedBy getter
func TestAuditableEntity_GetCreatedBy(t *testing.T) {
	tests := []struct {
		name      string
		createdBy int
	}{
		{
			name:      "Zero value",
			createdBy: 0,
		},
		{
			name:      "Positive value",
			createdBy: 123,
		},
		{
			name:      "Large value",
			createdBy: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{CreatedBy: tt.createdBy}

			// Act
			result := entity.GetCreatedBy()

			// Assert
			if result != tt.createdBy {
				t.Errorf("Expected CreatedBy %d, got %d", tt.createdBy, result)
			}
		})
	}
}

// TestAuditableEntity_GetUpdatedBy validates UpdatedBy getter
func TestAuditableEntity_GetUpdatedBy(t *testing.T) {
	tests := []struct {
		name      string
		updatedBy int
	}{
		{
			name:      "Zero value",
			updatedBy: 0,
		},
		{
			name:      "Positive value",
			updatedBy: 456,
		},
		{
			name:      "Large value",
			updatedBy: 888888,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{UpdatedBy: tt.updatedBy}

			// Act
			result := entity.GetUpdatedBy()

			// Assert
			if result != tt.updatedBy {
				t.Errorf("Expected UpdatedBy %d, got %d", tt.updatedBy, result)
			}
		})
	}
}

// TestAuditableEntity_GetAuditNote validates AuditNote getter
func TestAuditableEntity_GetAuditNote(t *testing.T) {
	tests := []struct {
		name      string
		auditNote string
	}{
		{
			name:      "Empty string",
			auditNote: "",
		},
		{
			name:      "Simple note",
			auditNote: "Created by admin",
		},
		{
			name:      "Multi-word note",
			auditNote: "Updated user profile information",
		},
		{
			name:      "Note with special characters",
			auditNote: "Update: User's email changed from old@example.com to new@example.com",
		},
		{
			name:      "Long note",
			auditNote: "This is a very long audit note that contains detailed information about the changes made to the entity, including the reason for the change and the context in which it was made.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{AuditNote: tt.auditNote}

			// Act
			result := entity.GetAuditNote()

			// Assert
			if result != tt.auditNote {
				t.Errorf("Expected AuditNote %q, got %q", tt.auditNote, result)
			}
		})
	}
}

// TestAuditableEntity_SetCreatedBy validates CreatedBy setter
func TestAuditableEntity_SetCreatedBy(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		expected int
	}{
		{
			name:     "Set to zero",
			userID:   0,
			expected: 0,
		},
		{
			name:     "Set to positive value",
			userID:   789,
			expected: 789,
		},
		{
			name:     "Set to large value",
			userID:   1234567,
			expected: 1234567,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{}

			// Act
			entity.SetCreatedBy(tt.userID)

			// Assert
			if entity.CreatedBy != tt.expected {
				t.Errorf("Expected CreatedBy %d, got %d", tt.expected, entity.CreatedBy)
			}

			if entity.GetCreatedBy() != tt.expected {
				t.Errorf("Expected GetCreatedBy() %d, got %d", tt.expected, entity.GetCreatedBy())
			}
		})
	}
}

// TestAuditableEntity_SetUpdatedBy validates UpdatedBy setter
func TestAuditableEntity_SetUpdatedBy(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		expected int
	}{
		{
			name:     "Set to zero",
			userID:   0,
			expected: 0,
		},
		{
			name:     "Set to positive value",
			userID:   321,
			expected: 321,
		},
		{
			name:     "Set to large value",
			userID:   7654321,
			expected: 7654321,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{}

			// Act
			entity.SetUpdatedBy(tt.userID)

			// Assert
			if entity.UpdatedBy != tt.expected {
				t.Errorf("Expected UpdatedBy %d, got %d", tt.expected, entity.UpdatedBy)
			}

			if entity.GetUpdatedBy() != tt.expected {
				t.Errorf("Expected GetUpdatedBy() %d, got %d", tt.expected, entity.GetUpdatedBy())
			}
		})
	}
}

// TestAuditableEntity_SetAuditNote validates AuditNote setter
func TestAuditableEntity_SetAuditNote(t *testing.T) {
	tests := []struct {
		name     string
		note     string
		expected string
	}{
		{
			name:     "Set empty note",
			note:     "",
			expected: "",
		},
		{
			name:     "Set simple note",
			note:     "Entity created",
			expected: "Entity created",
		},
		{
			name:     "Set detailed note",
			note:     "User profile updated by administrator due to data quality issues",
			expected: "User profile updated by administrator due to data quality issues",
		},
		{
			name:     "Set note with special characters",
			note:     "Email changed: test@example.com → new@example.com (verified)",
			expected: "Email changed: test@example.com → new@example.com (verified)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			entity := AuditableEntity{}

			// Act
			entity.SetAuditNote(tt.note)

			// Assert
			if entity.AuditNote != tt.expected {
				t.Errorf("Expected AuditNote %q, got %q", tt.expected, entity.AuditNote)
			}

			if entity.GetAuditNote() != tt.expected {
				t.Errorf("Expected GetAuditNote() %q, got %q", tt.expected, entity.GetAuditNote())
			}
		})
	}
}

// TestAuditableEntity_CompleteAuditCycle validates complete audit lifecycle
func TestAuditableEntity_CompleteAuditCycle(t *testing.T) {
	// Arrange
	entity := AuditableEntity{}
	creatorID := 100
	updaterID := 200
	createNote := "Entity created by system"
	updateNote := "Entity updated by user"

	// Act - Initial creation
	entity.SetCreatedBy(creatorID)
	entity.SetAuditNote(createNote)

	// Assert - After creation
	if entity.GetCreatedBy() != creatorID {
		t.Errorf("Expected CreatedBy %d, got %d", creatorID, entity.GetCreatedBy())
	}

	if entity.GetUpdatedBy() != 0 {
		t.Errorf("Expected UpdatedBy 0 (not set), got %d", entity.GetUpdatedBy())
	}

	if entity.GetAuditNote() != createNote {
		t.Errorf("Expected AuditNote %q, got %q", createNote, entity.GetAuditNote())
	}

	// Act - Update
	entity.SetUpdatedBy(updaterID)
	entity.SetAuditNote(updateNote)

	// Assert - After update
	if entity.GetCreatedBy() != creatorID {
		t.Errorf("Expected CreatedBy %d (unchanged), got %d", creatorID, entity.GetCreatedBy())
	}

	if entity.GetUpdatedBy() != updaterID {
		t.Errorf("Expected UpdatedBy %d, got %d", updaterID, entity.GetUpdatedBy())
	}

	if entity.GetAuditNote() != updateNote {
		t.Errorf("Expected AuditNote %q, got %q", updateNote, entity.GetAuditNote())
	}
}

// TestAuditableEntity_BaseEntityFunctionality validates BaseEntity methods work
func TestAuditableEntity_BaseEntityFunctionality(t *testing.T) {
	// Arrange
	entity := AuditableEntity{}
	entity.ID = 42
	entity.Version = 5

	// Act & Assert - BaseEntity methods should work
	if entity.GetID() != 42 {
		t.Errorf("Expected ID 42, got %d", entity.GetID())
	}

	if entity.GetVersion() != 5 {
		t.Errorf("Expected Version 5, got %d", entity.GetVersion())
	}

	// Test SetVersion from BaseEntity
	entity.SetVersion(10)
	if entity.GetVersion() != 10 {
		t.Errorf("Expected Version 10 after SetVersion, got %d", entity.GetVersion())
	}

	// Test timestamps are accessible
	now := time.Now()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	if entity.GetCreatedAt() != now {
		t.Errorf("Expected CreatedAt %v, got %v", now, entity.GetCreatedAt())
	}

	if entity.GetUpdatedAt() != now {
		t.Errorf("Expected UpdatedAt %v, got %v", now, entity.GetUpdatedAt())
	}
}

// TestAuditableEntity_DefaultValues validates default field values
func TestAuditableEntity_DefaultValues(t *testing.T) {
	// Arrange & Act
	entity := AuditableEntity{}

	// Assert - Default audit field values
	if entity.CreatedBy != 0 {
		t.Errorf("Expected default CreatedBy 0, got %d", entity.CreatedBy)
	}

	if entity.UpdatedBy != 0 {
		t.Errorf("Expected default UpdatedBy 0, got %d", entity.UpdatedBy)
	}

	if entity.AuditNote != "" {
		t.Errorf("Expected default AuditNote empty, got %q", entity.AuditNote)
	}

	// Assert - BaseEntity default values
	if entity.ID != 0 {
		t.Errorf("Expected default ID 0, got %d", entity.ID)
	}

	if entity.Version != 0 {
		t.Errorf("Expected default Version 0, got %d", entity.Version)
	}

	if !entity.CreatedAt.IsZero() {
		t.Errorf("Expected default CreatedAt to be zero time, got %v", entity.CreatedAt)
	}

	if !entity.UpdatedAt.IsZero() {
		t.Errorf("Expected default UpdatedAt to be zero time, got %v", entity.UpdatedAt)
	}
}

// TestAuditableEntity_Immutability validates getter immutability
func TestAuditableEntity_Immutability(t *testing.T) {
	// Arrange
	entity := AuditableEntity{
		CreatedBy: 100,
		UpdatedBy: 200,
		AuditNote: "Original note",
	}

	// Act - Get values (should not modify original)
	_ = entity.GetCreatedBy()
	_ = entity.GetUpdatedBy()
	_ = entity.GetAuditNote()

	// Assert - Original values should be unchanged (getters return copies)
	if entity.GetCreatedBy() != 100 {
		t.Errorf("Expected original CreatedBy 100, got %d", entity.GetCreatedBy())
	}

	if entity.GetUpdatedBy() != 200 {
		t.Errorf("Expected original UpdatedBy 200, got %d", entity.GetUpdatedBy())
	}

	if entity.GetAuditNote() != "Original note" {
		t.Errorf("Expected original AuditNote 'Original note', got %q", entity.GetAuditNote())
	}
}
