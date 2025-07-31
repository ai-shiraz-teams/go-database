package types

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity provides common fields that should be embedded in all domain entities.
type BaseEntity struct {
	// ID is the primary key for all entities
	ID int `gorm:"primaryKey" json:"-"`

	Slug string `gorm:"slug;uniqueIndex" json:"slug"`

	// CreatedAt timestamp when the entity was first created
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt timestamp when the entity was last modified
	UpdatedAt time.Time `json:"updatedAt"`

	// DeletedAt enables soft delete functionality (GORM will automatically handle this)
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	// Version field for optimistic locking support
	Version int `gorm:"default:1" json:"version"`
}

// GetID returns the entity's unique identifier
func (b *BaseEntity) GetID() int {
	return b.ID
}

// GetCreatedAt returns when the entity was created
func (b *BaseEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns when the entity was last updated
func (b *BaseEntity) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// GetDeletedAt returns the soft deletion timestamp, nil if not deleted
func (b *BaseEntity) GetDeletedAt() *time.Time {
	if b.DeletedAt.Valid {
		return &b.DeletedAt.Time
	}
	return nil
}

// GetVersion returns the current version for optimistic locking
func (b *BaseEntity) GetVersion() int {
	return b.Version
}

// SetVersion updates the version field (used by repositories for optimistic locking)
func (b *BaseEntity) SetVersion(version int) {
	b.Version = version
}
