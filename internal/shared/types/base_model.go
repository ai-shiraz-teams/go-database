package types

import "time"

// IBaseModel defines the contract that all entities with BaseEntity must satisfy.
// This interface enables generic repository patterns and type-safe operations.
type IBaseModel interface {
	// GetID returns the entity's unique identifier
	GetID() int

	// GetCreatedAt returns when the entity was created
	GetCreatedAt() time.Time

	// GetUpdatedAt returns when the entity was last updated
	GetUpdatedAt() time.Time

	// GetDeletedAt returns the soft deletion timestamp, nil if not deleted
	GetDeletedAt() *time.Time

	// GetVersion returns the current version for optimistic locking
	GetVersion() int

	// SetVersion updates the version field (used by repositories for optimistic locking)
	SetVersion(version int)
}
