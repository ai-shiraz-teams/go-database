package domain

import "time"

type IBaseModel interface {
	GetID() int

	GetSlug() string

	GetCreatedAt() time.Time

	GetUpdatedAt() time.Time

	GetDeletedAt() *time.Time

	GetVersion() int

	SetVersion(version int)
}
