package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseEntity struct {
	ID int `gorm:"primaryKey" json:"id"`

	Slug string `gorm:"uniqueIndex" json:"slug"`

	CreatedAt time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Version int `gorm:"default:1" json:"version"`
}

func (b *BaseEntity) GetID() int {
	return b.ID
}

func (b *BaseEntity) GetSlug() string {
	return b.Slug
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

func (b *BaseEntity) GetVersion() int {
	return b.Version
}

func (b *BaseEntity) SetVersion(version int) {
	b.Version = version
}

func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.Slug == "" {
		b.Slug = uuid.New().String()
	}
	return nil
}
