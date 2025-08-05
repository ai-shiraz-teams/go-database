package query

import (
	"reflect"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

// FilterOperator defines type-safe comparison operations for different field types
type FilterOperator[T domain.IBaseModel, V comparable] interface {
	// Equal checks if field equals the provided value
	Equal(value V) FieldFilterBuilder[T]
	// NotEqual checks if field does not equal the provided value
	NotEqual(value V) FieldFilterBuilder[T]
}

// ComparableFilterOperator extends FilterOperator for comparable types (numbers, dates, strings)
type ComparableFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]
	// GreaterThan checks if field is greater than the provided value
	GreaterThan(value V) FieldFilterBuilder[T]
	// GreaterOrEqual checks if field is greater than or equal to the provided value
	GreaterOrEqual(value V) FieldFilterBuilder[T]
	// LessThan checks if field is less than the provided value
	LessThan(value V) FieldFilterBuilder[T]
	// LessOrEqual checks if field is less than or equal to the provided value
	LessOrEqual(value V) FieldFilterBuilder[T]
	// Between checks if field is between two values (inclusive)
	Between(start, end V) FieldFilterBuilder[T]
}

// StringFilterOperator extends ComparableFilterOperator for string types
type StringFilterOperator[T domain.IBaseModel] interface {
	ComparableFilterOperator[T, string]
	// Like performs pattern matching using SQL LIKE syntax
	Like(pattern string) FieldFilterBuilder[T]
	// Contains checks if string contains the provided substring
	Contains(substring string) FieldFilterBuilder[T]
	// StartsWith checks if string starts with the provided prefix
	StartsWith(prefix string) FieldFilterBuilder[T]
	// EndsWith checks if string ends with the provided suffix
	EndsWith(suffix string) FieldFilterBuilder[T]
	// IContains performs case-insensitive contains
	IContains(substring string) FieldFilterBuilder[T]
	// ILike performs case-insensitive pattern matching
	ILike(pattern string) FieldFilterBuilder[T]
}

// SliceFilterOperator provides operations for slice/array fields and IN operations
type SliceFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]
	// In checks if field value is in the provided slice
	In(values []V) FieldFilterBuilder[T]
	// NotIn checks if field value is not in the provided slice
	NotIn(values []V) FieldFilterBuilder[T]
}

// NullableFilterOperator provides null checking operations for pointer types
type NullableFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]
	// IsNull checks if field is NULL
	IsNull() FieldFilterBuilder[T]
	// IsNotNull checks if field is not NULL
	IsNotNull() FieldFilterBuilder[T]
}

// BooleanFilterOperator provides boolean-specific operations
type BooleanFilterOperator[T domain.IBaseModel] interface {
	FilterOperator[T, bool]
	// IsTrue checks if boolean field is true
	IsTrue() FieldFilterBuilder[T]
	// IsFalse checks if boolean field is false
	IsFalse() FieldFilterBuilder[T]
}

// FieldFilterOperator represents a type-safe field filter operator without using any
// This provides compile-time guarantees while allowing different operator types
type FieldFilterOperator[T domain.IBaseModel] interface {
	// GetFieldPath returns the field path
	GetFieldPath() string
	// GetFieldType returns the field type information
	GetFieldType() TypeInfo
	// AsString returns string operations if the field is a string
	AsString() (StringFilterOperator[T], bool)
	// AsComparableInt returns int operations if the field is an int
	AsComparableInt() (ComparableFilterOperator[T, int], bool)
	// AsComparableFloat returns float operations if the field is a float
	AsComparableFloat() (ComparableFilterOperator[T, float64], bool)
	// AsBoolean returns boolean operations if the field is a boolean
	AsBoolean() (BooleanFilterOperator[T], bool)
	// AsNullableString returns nullable string operations if the field is *string
	AsNullableString() (NullableFilterOperator[T, string], bool)
}

// FieldFilterBuilder provides the fluent API for building field-specific filters
type FieldFilterBuilder[T domain.IBaseModel] interface {
	// GetFieldPath returns the field path for this filter
	GetFieldPath() string
	// GetFilterCriteria returns the accumulated filter criteria
	GetFilterCriteria() []domain.FilterCriteria
	// And allows combining with other filters using AND logic
	And() StronglyTypedFilterBuilder[T]
	// Or allows combining with other filters using OR logic
	Or() StronglyTypedFilterBuilder[T]
}

// StronglyTypedFilterBuilder provides the main API for building strongly typed filters
type StronglyTypedFilterBuilder[T domain.IBaseModel] interface {
	// Field access methods - these will be generated based on T's structure
	// For each field in T, we'll have methods like:
	// - Name() StringFilterOperator[T] (for string fields)
	// - Age() ComparableFilterOperator[T, int] (for int fields)
	// - IsActive() BooleanFilterOperator[T] (for bool fields)
	// - Email() NullableFilterOperator[T, string] (for *string fields)

	// GetField provides dynamic field access by name with type safety
	GetField(fieldName string) FieldFilterOperator[T]

	// Logical operations
	And(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T]
	Or(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T]

	// Conversion methods
	ToFilterCriteria() []domain.FilterCriteria
	Build() StronglyTypedFilter[T]
}

// StronglyTypedFilter represents the final built filter
type StronglyTypedFilter[T domain.IBaseModel] interface {
	// Convert to domain filter criteria for database queries
	ToFilterCriteria() []domain.FilterCriteria
	// Clone creates a copy of this filter
	Clone() StronglyTypedFilter[T]
	// IsEmpty returns true if no filters are defined
	IsEmpty() bool
}

// SortDirection represents sort order direction
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// SortBuilder provides strongly typed sorting operations
type SortBuilder[T domain.IBaseModel] interface {
	// Field-specific sort methods will be generated based on T's structure
	// For each field in T, we'll have methods like:
	// - ByName(direction SortDirection) SortBuilder[T]
	// - ByCreatedAt(direction SortDirection) SortBuilder[T]

	// GetField provides dynamic field access for sorting
	GetField(fieldName string, direction SortDirection) SortBuilder[T]

	// Convenience methods
	Asc(fieldName string) SortBuilder[T]
	Desc(fieldName string) SortBuilder[T]

	// Multiple sorts
	ThenBy(fieldName string, direction SortDirection) SortBuilder[T]
	ThenAsc(fieldName string) SortBuilder[T]
	ThenDesc(fieldName string) SortBuilder[T]

	// Build result
	Build() []domain.SortField
}

// QueryParamsBuilder provides the main strongly typed query building API
type QueryParamsBuilder[T domain.IBaseModel] interface {
	// Filter operations
	Filter() StronglyTypedFilterBuilder[T]
	WithFilter(filter StronglyTypedFilter[T]) QueryParamsBuilder[T]

	// Sort operations
	Sort() SortBuilder[T]
	WithSort(sorts []domain.SortField) QueryParamsBuilder[T]

	// Pagination
	WithLimit(limit int) QueryParamsBuilder[T]
	WithOffset(offset int) QueryParamsBuilder[T]
	WithPage(page, pageSize int) QueryParamsBuilder[T]

	// Preloading
	WithPreloads(preloads []string) QueryParamsBuilder[T]
	WithPreload(preload string) QueryParamsBuilder[T]

	// Soft delete visibility
	IncludeDeleted() QueryParamsBuilder[T]
	OnlyDeleted() QueryParamsBuilder[T]
	ExcludeDeleted() QueryParamsBuilder[T]

	// Build final result
	Build() QueryParams[T]
}

// QueryParams represents the concrete implementation of QueryParamsType
// This is the main struct that users will work with for strongly typed queries
type QueryParams[T domain.IBaseModel] struct {
	// Core filtering using strongly typed filters
	filter StronglyTypedFilter[T]

	// Sorting configuration
	sort []domain.SortField

	// Pagination controls
	limit    int
	offset   int
	page     int
	pageSize int

	// Relationship preloading
	preloads []string

	// Soft-delete visibility
	includeDeleted bool
	onlyDeleted    bool

	// Free-text search
	search string
}

// Ensure QueryParams implements QueryParamsType
var _ QueryParamsType[*domain.BaseEntity] = (*QueryParams[*domain.BaseEntity])(nil)

// Implement domain.IQueryParams interface with strongly typed alternatives
func (qp *QueryParams[T]) Filter() interface{} {
	if qp.filter == nil {
		return nil
	}
	return qp.filter
}

func (qp *QueryParams[T]) Sort() interface{} {
	return qp.sort
}

func (qp *QueryParams[T]) Limit() int {
	return qp.limit
}

func (qp *QueryParams[T]) Offset() int {
	return qp.offset
}

func (qp *QueryParams[T]) Preloads() []string {
	return qp.preloads
}

func (qp *QueryParams[T]) WithFilter(filter interface{}) domain.IQueryParams[T] {
	if f, ok := filter.(StronglyTypedFilter[T]); ok {
		qp.filter = f
	}
	return qp
}

func (qp *QueryParams[T]) WithSort(sort interface{}) domain.IQueryParams[T] {
	if sorts, ok := sort.([]domain.SortField); ok {
		qp.sort = sorts
	}
	return qp
}

func (qp *QueryParams[T]) WithLimit(limit int) domain.IQueryParams[T] {
	qp.limit = limit
	return qp
}

func (qp *QueryParams[T]) WithOffset(offset int) domain.IQueryParams[T] {
	qp.offset = offset
	return qp
}

func (qp *QueryParams[T]) WithPreloads(preloads []string) domain.IQueryParams[T] {
	qp.preloads = preloads
	return qp
}

func (qp *QueryParams[T]) PrepareDefaults() error {
	if qp.limit <= 0 {
		qp.limit = 50
	}
	if qp.limit > 200 {
		qp.limit = 200
	}
	if qp.offset < 0 {
		qp.offset = 0
	}
	return nil
}

func (qp *QueryParams[T]) HasFilters() bool {
	return qp.filter != nil && !qp.filter.IsEmpty()
}

func (qp *QueryParams[T]) ToFilterCriteria() []domain.FilterCriteria {
	if qp.filter == nil {
		return []domain.FilterCriteria{}
	}
	return qp.filter.ToFilterCriteria()
}

func (qp *QueryParams[T]) HasSort() bool {
	return len(qp.sort) > 0
}

func (qp *QueryParams[T]) HasPreloads() bool {
	return len(qp.preloads) > 0
}

func (qp *QueryParams[T]) GetIncludeDeleted() bool {
	return qp.includeDeleted
}

func (qp *QueryParams[T]) GetOnlyDeleted() bool {
	return qp.onlyDeleted
}

func (qp *QueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) domain.IQueryParams[T] {
	qp.includeDeleted = includeDeleted
	qp.onlyDeleted = onlyDeleted
	return qp
}

func (qp *QueryParams[T]) ToSortFields() []domain.SortField {
	return qp.sort
}

func (qp *QueryParams[T]) WithFilters(identifier interface{}) domain.IQueryParams[T] {
	// This will be handled by the WithFilter method instead
	return qp
}

// Type-safe alternatives without duplicating existing methods
func (qp *QueryParams[T]) GetSortFields() []domain.SortField {
	return qp.sort
}

func (qp *QueryParams[T]) SetSortFields(sorts []domain.SortField) {
	qp.sort = sorts
}

// Type-safe filter builder methods
func (qp *QueryParams[T]) WithStrongFilter(filter StronglyTypedFilter[T]) *QueryParams[T] {
	qp.filter = filter
	return qp
}

func (qp *QueryParams[T]) WithSortFieldsChain(sorts []domain.SortField) *QueryParams[T] {
	qp.sort = sorts
	return qp
}

func (qp *QueryParams[T]) AddSortDesc(field string) domain.IQueryParams[T] {
	qp.sort = append(qp.sort, domain.SortField{Field: field, Order: domain.SortOrderDesc})
	return qp
}

func (qp *QueryParams[T]) AddSortAsc(field string) domain.IQueryParams[T] {
	qp.sort = append(qp.sort, domain.SortField{Field: field, Order: domain.SortOrderAsc})
	return qp
}

// QueryParamsType represents the central union type for strongly typed query parameters
// This is the main export that satisfies all requirements for type-safe, ergonomic query building.
// It can be used in multiple ways:
// 1. As QueryParams[T] for object literal construction with full IntelliSense
// 2. As domain.IQueryParams[T] interface for method-based construction
// 3. As QueryParamsLiteral[T] for direct JSON-like object construction
type QueryParamsType[T domain.IBaseModel] interface {
	domain.IQueryParams[T]

	// Strongly typed filter operations
	GetStronglyTypedFilter() StronglyTypedFilter[T]
	SetStronglyTypedFilter(filter StronglyTypedFilter[T])

	// Object literal support
	ToLiteral() QueryParamsLiteral[T]
	FromLiteral(literal QueryParamsLiteral[T]) error

	// Type safety validation
	ValidateFieldPaths() error
	GetAvailableFields() map[string]*FieldInfo

	// Enhanced query building
	BuildFilter() StronglyTypedFilterBuilder[T]
	BuildSort() SortBuilder[T]
}

// FilterShape represents the strongly typed filter structure for a given entity type T
// This allows for nested property access with full type safety and IntelliSense support
// The key is the field name, and the value is the filter operator for that field's type
type FilterShape[T domain.IBaseModel] map[string]FieldFilterOperator[T]

// SortShape represents the strongly typed sort structure for a given entity type T
type SortShape[T domain.IBaseModel] map[string]SortDirection

// GetStronglyTypedFilter returns the strongly typed filter if available
func (qp *QueryParams[T]) GetStronglyTypedFilter() StronglyTypedFilter[T] {
	return qp.filter
}

// SetStronglyTypedFilter sets the strongly typed filter
func (qp *QueryParams[T]) SetStronglyTypedFilter(filter StronglyTypedFilter[T]) {
	qp.filter = filter
}

// Object literal support
func (qp *QueryParams[T]) ToLiteral() QueryParamsLiteral[T] {
	literal := QueryParamsLiteral[T]{
		Limit:          qp.limit,
		Offset:         qp.offset,
		Preloads:       qp.preloads,
		IncludeDeleted: qp.includeDeleted,
		OnlyDeleted:    qp.onlyDeleted,
	}

	// Convert filter to literal format if present
	if qp.filter != nil && !qp.filter.IsEmpty() {
		literal.Filter = make(FilterLiteral[T])
		// This would need specific implementation based on filter structure
		// For now, we'll leave it as a placeholder
	}

	// Convert sort to literal format
	if len(qp.sort) > 0 {
		literal.Sort = make(SortLiteral[T])
		for _, sortField := range qp.sort {
			direction := SortAsc
			if sortField.Order == domain.SortOrderDesc {
				direction = SortDesc
			}
			literal.Sort[sortField.Field] = direction
		}
	}

	return literal
}

func (qp *QueryParams[T]) FromLiteral(literal QueryParamsLiteral[T]) error {
	converted, err := literal.ToQueryParams()
	if err != nil {
		return err
	}

	*qp = *converted
	return nil
}

// Type safety validation
func (qp *QueryParams[T]) ValidateFieldPaths() error {
	// Implementation would validate all field paths in filters and sorts
	// against the registered type information
	return nil
}

func (qp *QueryParams[T]) GetAvailableFields() map[string]*FieldInfo {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return globalTypeRegistry.GetFields(modelType)
}

// Enhanced query building
func (qp *QueryParams[T]) BuildFilter() StronglyTypedFilterBuilder[T] {
	return NewStronglyTypedFilter[T](globalTypeRegistry)
}

func (qp *QueryParams[T]) BuildSort() SortBuilder[T] {
	// For now, return nil as a placeholder - would need full SortBuilder implementation
	return nil
}

// TypeInfo provides runtime type information for field validation and suggestions
type TypeInfo struct {
	Kind        reflect.Kind
	Type        reflect.Type
	IsPointer   bool
	IsSlice     bool
	ElementType reflect.Type
}

// FieldInfo contains metadata about a field for type-safe operations
type FieldInfo struct {
	Name     string
	JSONName string
	GormName string
	TypeInfo TypeInfo
	Path     []string
}

// TypeRegistry manages type information for strongly typed operations
type TypeRegistry interface {
	// RegisterType registers a type for strongly typed operations using generics
	RegisterType(model domain.IBaseModel) error
	// GetFieldInfo returns field information for a given type and field path
	GetFieldInfo(modelType reflect.Type, fieldPath string) (*FieldInfo, error)
	// GetFields returns all available fields for a type
	GetFields(modelType reflect.Type) map[string]*FieldInfo
	// ValidateFieldPath validates that a field path exists and returns its type info
	ValidateFieldPath(modelType reflect.Type, fieldPath string) (*FieldInfo, error)
}
