package common

// IBaseModel defines the minimal interface for database entities.
// This is a standalone interface to avoid circular dependencies.
type IBaseModel interface {
	GetID() int
	GetSlug() string
	IsDeleted() bool
}

// IQueryParams provides query parameter interface without domain dependencies
type IQueryParams[T IBaseModel] interface {
	Filter() interface{}
	Sort() interface{}
	Limit() int
	Offset() int
	Preloads() []string
	WithFilter(filter interface{}) IQueryParams[T]
	WithSort(sort interface{}) IQueryParams[T]
	WithLimit(limit int) IQueryParams[T]
	WithOffset(offset int) IQueryParams[T]
	WithPreloads(preloads []string) IQueryParams[T]
	PrepareDefaults() error
	HasFilters() bool
	HasSort() bool
	HasPreloads() bool
}
