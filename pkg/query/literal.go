package query

import (
	"fmt"
	"reflect"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

// Global type registry instance
var globalTypeRegistry = NewTypeRegistry()

// QueryParamsLiteral represents a query parameter structure that can be constructed as an object literal
// This provides the strongly typed structure that users can construct directly
type QueryParamsLiteral[T domain.IBaseModel] struct {
	Filter         FilterLiteral[T] `json:"filter,omitempty"`
	Sort           SortLiteral[T]   `json:"sort,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
	Preloads       []string         `json:"preloads,omitempty"`
	IncludeDeleted bool             `json:"includeDeleted,omitempty"`
	OnlyDeleted    bool             `json:"onlyDeleted,omitempty"`
}

// FilterLiteral allows object literal construction of filters with type safety
// Uses strongly typed approach with union types instead of map[string]interface{}
type FilterLiteral[T domain.IBaseModel] map[string]FilterLiteralValue[T]

// FilterLiteralValue represents a strongly typed filter value that can be either:
// 1. A FilterOperatorLiteral for operator-based filtering
// 2. A slice of FilterLiteral for logical operations (AND/OR)
// 3. A direct value for equality comparison (strongly typed)
type FilterLiteralValue[T domain.IBaseModel] interface {
	isFilterLiteralValue()
}

// FilterOperatorLiteralValue represents an operator-based filter
type FilterOperatorLiteralValue[T domain.IBaseModel, V comparable] struct {
	FilterOperatorLiteral[V]
}

func (f FilterOperatorLiteralValue[T, V]) isFilterLiteralValue() {}

// FilterLogicalValue represents logical operations (AND/OR)
type FilterLogicalValue[T domain.IBaseModel] struct {
	Values []FilterLiteral[T]
}

func (f FilterLogicalValue[T]) isFilterLiteralValue() {}

// DirectFilterValue represents a direct value for equality comparison
type DirectFilterValue[V comparable] struct {
	Value V
}

func (f DirectFilterValue[V]) isFilterLiteralValue() {}

// SortLiteral allows object literal construction of sort parameters
type SortLiteral[T domain.IBaseModel] map[string]SortDirection

// FilterOperatorLiteral represents the available filter operations for a field
type FilterOperatorLiteral[V any] struct {
	Eq        *V    `json:"eq,omitempty"`          // Equal
	Neq       *V    `json:"neq,omitempty"`         // Not equal
	Gt        *V    `json:"gt,omitempty"`          // Greater than
	Gte       *V    `json:"gte,omitempty"`         // Greater than or equal
	Lt        *V    `json:"lt,omitempty"`          // Less than
	Lte       *V    `json:"lte,omitempty"`         // Less than or equal
	Like      *V    `json:"like,omitempty"`        // SQL LIKE pattern (strings only)
	Contains  *V    `json:"contains,omitempty"`    // Contains substring (strings only)
	In        []V   `json:"in,omitempty"`          // Value in list
	NotIn     []V   `json:"not_in,omitempty"`      // Value not in list
	Between   []V   `json:"between,omitempty"`     // Between two values (exactly 2 elements)
	IsNull    *bool `json:"is_null,omitempty"`     // Is NULL check
	IsNotNull *bool `json:"is_not_null,omitempty"` // Is NOT NULL check
}

// StringFilterOperatorLiteral extends FilterOperatorLiteral for string fields
type StringFilterOperatorLiteral = FilterOperatorLiteral[string]

// IntFilterOperatorLiteral extends FilterOperatorLiteral for int fields
type IntFilterOperatorLiteral = FilterOperatorLiteral[int]

// BoolFilterOperatorLiteral extends FilterOperatorLiteral for bool fields
type BoolFilterOperatorLiteral = FilterOperatorLiteral[bool]

// ComplexFilterLiteral allows for AND/OR logical operations
type ComplexFilterLiteral[T domain.IBaseModel] struct {
	And []FilterLiteral[T] `json:"and,omitempty"`
	Or  []FilterLiteral[T] `json:"or,omitempty"`
}

// ToQueryParams converts a QueryParamsLiteral to the internal QueryParams type
func (qpl *QueryParamsLiteral[T]) ToQueryParams() (*QueryParams[T], error) {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Register the type
	err := globalTypeRegistry.RegisterType(zero)
	if err != nil {
		return nil, err
	}

	qp := &QueryParams[T]{
		limit:          qpl.Limit,
		offset:         qpl.Offset,
		preloads:       qpl.Preloads,
		includeDeleted: qpl.IncludeDeleted,
		onlyDeleted:    qpl.OnlyDeleted,
	}

	// Convert filter literal to strongly typed filter
	if qpl.Filter != nil {
		filter, err := qpl.Filter.ToStronglyTypedFilter(globalTypeRegistry, modelType)
		if err != nil {
			return nil, err
		}
		qp.filter = filter
	}

	// Convert sort literal to sort fields
	if qpl.Sort != nil {
		sortFields, err := qpl.Sort.ToSortFields(globalTypeRegistry, modelType)
		if err != nil {
			return nil, err
		}
		qp.sort = sortFields
	}

	// Set defaults
	err = qp.PrepareDefaults()
	if err != nil {
		return nil, err
	}

	return qp, nil
}

// ToStronglyTypedFilter converts FilterLiteral to StronglyTypedFilter
func (fl FilterLiteral[T]) ToStronglyTypedFilter(registry TypeRegistry, modelType reflect.Type) (StronglyTypedFilter[T], error) {
	criteria := []domain.FilterCriteria{}

	for fieldPath, value := range fl {
		// Handle special cases for AND/OR
		if fieldPath == "AND" || fieldPath == "Or" {
			// Handle complex filters later
			continue
		}

		// Validate field path
		fieldInfo, err := registry.ValidateFieldPath(modelType, fieldPath)
		if err != nil {
			return nil, err
		}

		// Convert value to filter criteria
		fieldCriteria, err := fl.convertValueToFilterCriteria(fieldPath, value, fieldInfo)
		if err != nil {
			return nil, err
		}

		criteria = append(criteria, fieldCriteria...)
	}

	return &stronglyTypedFilter[T]{criteria: criteria}, nil
}

// convertValueToFilterCriteria converts a strongly typed field value to filter criteria
func (fl FilterLiteral[T]) convertValueToFilterCriteria(fieldPath string, value FilterLiteralValue[T], fieldInfo *FieldInfo) ([]domain.FilterCriteria, error) {
	criteria := []domain.FilterCriteria{}

	switch v := value.(type) {
	case FilterOperatorLiteralValue[T, string]:
		return fl.convertStringOperatorToCriteria(fieldPath, v.FilterOperatorLiteral), nil
	case FilterOperatorLiteralValue[T, int]:
		return fl.convertIntOperatorToCriteria(fieldPath, v.FilterOperatorLiteral), nil
	case FilterOperatorLiteralValue[T, bool]:
		return fl.convertBoolOperatorToCriteria(fieldPath, v.FilterOperatorLiteral), nil
	case FilterLogicalValue[T]:
		return fl.convertLogicalValueToCriteria(fieldPath, v)
	case DirectFilterValue[string]:
		criteria = append(criteria, domain.FilterCriteria{
			Field:    fieldPath,
			Operator: domain.FilterOperatorEqual,
			Value:    v.Value,
		})
	case DirectFilterValue[int]:
		criteria = append(criteria, domain.FilterCriteria{
			Field:    fieldPath,
			Operator: domain.FilterOperatorEqual,
			Value:    v.Value,
		})
	case DirectFilterValue[bool]:
		criteria = append(criteria, domain.FilterCriteria{
			Field:    fieldPath,
			Operator: domain.FilterOperatorEqual,
			Value:    v.Value,
		})
	default:
		return nil, fmt.Errorf("unsupported filter value type for field %s", fieldPath)
	}

	return criteria, nil
}

// convertStringOperatorToCriteria converts string operators to filter criteria
func (fl FilterLiteral[T]) convertStringOperatorToCriteria(fieldPath string, op FilterOperatorLiteral[string]) []domain.FilterCriteria {
	var criteria []domain.FilterCriteria

	if op.Eq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorEqual, Value: *op.Eq,
		})
	}
	if op.Neq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorNotEqual, Value: *op.Neq,
		})
	}
	if op.Like != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorLike, Value: *op.Like,
		})
	}
	if op.Contains != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorLike, Value: fmt.Sprintf("%%%s%%", *op.Contains),
		})
	}
	if len(op.In) > 0 {
		values := make([]interface{}, len(op.In))
		for i, v := range op.In {
			values[i] = v
		}
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorIn, Values: values,
		})
	}
	if len(op.NotIn) > 0 {
		values := make([]interface{}, len(op.NotIn))
		for i, v := range op.NotIn {
			values[i] = v
		}
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorNotIn, Values: values,
		})
	}
	if op.IsNull != nil && *op.IsNull {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorIsNull,
		})
	}
	if op.IsNotNull != nil && *op.IsNotNull {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorIsNotNull,
		})
	}

	return criteria
}

// convertIntOperatorToCriteria converts int operators to filter criteria
func (fl FilterLiteral[T]) convertIntOperatorToCriteria(fieldPath string, op FilterOperatorLiteral[int]) []domain.FilterCriteria {
	var criteria []domain.FilterCriteria

	if op.Eq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorEqual, Value: *op.Eq,
		})
	}
	if op.Neq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorNotEqual, Value: *op.Neq,
		})
	}
	if op.Gt != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorGreaterThan, Value: *op.Gt,
		})
	}
	if op.Gte != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorGreaterEqual, Value: *op.Gte,
		})
	}
	if op.Lt != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorLessThan, Value: *op.Lt,
		})
	}
	if op.Lte != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorLessEqual, Value: *op.Lte,
		})
	}
	if len(op.Between) == 2 {
		values := []interface{}{op.Between[0], op.Between[1]}
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorBetween, Values: values,
		})
	}
	if len(op.In) > 0 {
		values := make([]interface{}, len(op.In))
		for i, v := range op.In {
			values[i] = v
		}
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorIn, Values: values,
		})
	}

	return criteria
}

// convertBoolOperatorToCriteria converts bool operators to filter criteria
func (fl FilterLiteral[T]) convertBoolOperatorToCriteria(fieldPath string, op FilterOperatorLiteral[bool]) []domain.FilterCriteria {
	var criteria []domain.FilterCriteria

	if op.Eq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorEqual, Value: *op.Eq,
		})
	}
	if op.Neq != nil {
		criteria = append(criteria, domain.FilterCriteria{
			Field: fieldPath, Operator: domain.FilterOperatorNotEqual, Value: *op.Neq,
		})
	}

	return criteria
}

// convertLogicalValueToCriteria converts logical operations to filter criteria
func (fl FilterLiteral[T]) convertLogicalValueToCriteria(fieldPath string, logical FilterLogicalValue[T]) ([]domain.FilterCriteria, error) {
	// This would need more complex handling for nested logical operations
	// For now, return empty criteria as a placeholder
	return []domain.FilterCriteria{}, nil
}

// ToSortFields converts SortLiteral to domain.SortField slice
func (sl SortLiteral[T]) ToSortFields(registry TypeRegistry, modelType reflect.Type) ([]domain.SortField, error) {
	var sortFields []domain.SortField

	for fieldPath, direction := range sl {
		// Validate field path
		_, err := registry.ValidateFieldPath(modelType, fieldPath)
		if err != nil {
			return nil, err
		}

		sortOrder := domain.SortOrderAsc
		if direction == SortDesc {
			sortOrder = domain.SortOrderDesc
		}

		sortFields = append(sortFields, domain.SortField{
			Field: fieldPath,
			Order: sortOrder,
		})
	}

	return sortFields, nil
}

// NewQueryParams creates a new QueryParams from a literal structure
func NewQueryParams[T domain.IBaseModel](literal QueryParamsLiteral[T]) (QueryParamsType[T], error) {
	return literal.ToQueryParams()
}

// Builder provides a fluent API for building query parameters
func Builder[T domain.IBaseModel]() QueryParamsBuilder[T] {
	return NewQueryParamsBuilder[T](globalTypeRegistry)
}

// Filter provides a fluent API for building strongly typed filters
func Filter[T domain.IBaseModel]() StronglyTypedFilterBuilder[T] {
	return NewStronglyTypedFilter[T](globalTypeRegistry)
}

// RegisterType registers a type with the global type registry for type safety
func RegisterType[T domain.IBaseModel]() error {
	var zero T
	return globalTypeRegistry.RegisterType(zero)
}

// GetTypeRegistry returns the global type registry
func GetTypeRegistry() TypeRegistry {
	return globalTypeRegistry
}

// Helper functions for creating filter operators

// Eq creates an equals filter operator
func Eq[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Eq: &value}
}

// Neq creates a not equals filter operator
func Neq[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Neq: &value}
}

// Gt creates a greater than filter operator
func Gt[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Gt: &value}
}

// Gte creates a greater than or equal filter operator
func Gte[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Gte: &value}
}

// Lt creates a less than filter operator
func Lt[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Lt: &value}
}

// Lte creates a less than or equal filter operator
func Lte[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Lte: &value}
}

// Like creates a LIKE pattern filter operator (for strings)
func Like(pattern string) FilterOperatorLiteral[string] {
	return FilterOperatorLiteral[string]{Like: &pattern}
}

// Contains creates a contains substring filter operator (for strings)
func Contains(substring string) FilterOperatorLiteral[string] {
	return FilterOperatorLiteral[string]{Contains: &substring}
}

// In creates an IN filter operator
func In[V any](values ...V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{In: values}
}

// NotIn creates a NOT IN filter operator
func NotIn[V any](values ...V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{NotIn: values}
}

// Between creates a BETWEEN filter operator
func Between[V any](start, end V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Between: []V{start, end}}
}

// IsNull creates an IS NULL filter operator (can be applied to any field type)
func IsNull[V any]() FilterOperatorLiteral[V] {
	t := true
	return FilterOperatorLiteral[V]{IsNull: &t}
}

// IsNotNull creates an IS NOT NULL filter operator (can be applied to any field type)
func IsNotNull[V any]() FilterOperatorLiteral[V] {
	t := true
	return FilterOperatorLiteral[V]{IsNotNull: &t}
}
