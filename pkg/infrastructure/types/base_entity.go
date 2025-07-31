package types

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BaseEntity struct {
	ID   int    `gorm:"primaryKey" json:"-"`
	Slug string `gorm:"uniqueIndex" json:"slug"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

func (b *BaseEntity) GetID() int {
	return b.ID
}

func (b *BaseEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b *BaseEntity) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

func (b *BaseEntity) GetDeletedAt() *time.Time {
	if b.DeletedAt.Valid {
		return &b.DeletedAt.Time
	}
	return nil
}

// BeforeCreate GORM hook to generate slug if empty
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.Slug == "" {
		// Generate a unique slug using timestamp and random suffix
		b.Slug = fmt.Sprintf("entity-%d", time.Now().UnixNano())
	}
	return nil
}
