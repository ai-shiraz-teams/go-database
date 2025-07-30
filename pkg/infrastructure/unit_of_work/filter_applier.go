package unit_of_work

import (
	"fmt"
	"reflect"

	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
	queryparams "github.com/ai-shiraz-teams/go-database-sdk/internal/shared/query"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"

	"gorm.io/gorm"
)

// FilterApplier provides utilities to convert IIdentifier filters to GORM queries.
// This maintains separation between domain logic and ORM implementation.
type FilterApplier struct{}

// NewFilterApplier creates a new FilterApplier instance
func NewFilterApplier() *FilterApplier {
	return &FilterApplier{}
}

// ApplyFilters converts FilterCriteria from IIdentifier to GORM query conditions
func (fa *FilterApplier) ApplyFilters(query *gorm.DB, filters []identifier.FilterCriteria) *gorm.DB {
	if len(filters) == 0 {
		return query
	}

	for i, filter := range filters {
		// For the first filter, always use WHERE
		// For subsequent filters, check the logical operator of the PREVIOUS filter
		isFirst := i == 0
		useOr := false

		if !isFirst && i > 0 {
			// The logical operator is stored on the previous filter
			prevFilter := filters[i-1]
			useOr = prevFilter.LogicalOp == identifier.LogicalOperatorOr
		}

		query = fa.applyFilter(query, filter, isFirst, useOr)
	}

	return query
}

// applyFilter applies a single FilterCriteria to the GORM query
func (fa *FilterApplier) applyFilter(query *gorm.DB, filter identifier.FilterCriteria, isFirst bool, useOr bool) *gorm.DB {
	// Handle grouped filters (nested conditions)
	if len(filter.Group) > 0 {
		return fa.applyGroupFilter(query, filter, isFirst, useOr)
	}

	// Handle individual filter conditions
	return fa.applySingleFilter(query, filter, isFirst, useOr)
}

// applyGroupFilter handles nested filter groups with AND/OR logic
func (fa *FilterApplier) applyGroupFilter(query *gorm.DB, filter identifier.FilterCriteria, isFirst bool, useOr bool) *gorm.DB {
	groupQuery := fa.ApplyFilters(query.Session(&gorm.Session{NewDB: true}), filter.Group)

	if isFirst {
		return query.Where(groupQuery)
	} else if useOr {
		return query.Or(groupQuery)
	} else {
		return query.Where(groupQuery)
	}
}

// applySingleFilter applies individual filter conditions based on operator
func (fa *FilterApplier) applySingleFilter(query *gorm.DB, filter identifier.FilterCriteria, isFirst bool, useOr bool) *gorm.DB {
	field := filter.Field
	operator := filter.Operator
	value := filter.Value
	values := filter.Values

	var condition string
	var args []interface{}

	switch operator {
	case identifier.FilterOperatorEqual:
		condition = fmt.Sprintf("%s = ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorNotEqual:
		condition = fmt.Sprintf("%s != ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorGreaterThan:
		condition = fmt.Sprintf("%s > ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorGreaterEqual:
		condition = fmt.Sprintf("%s >= ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorLessThan:
		condition = fmt.Sprintf("%s < ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorLessEqual:
		condition = fmt.Sprintf("%s <= ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorLike:
		condition = fmt.Sprintf("%s LIKE ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorIn:
		if len(values) > 0 {
			condition = fmt.Sprintf("%s IN ?", field)
			args = []interface{}{values}
		} else {
			// Handle empty IN clause - return no results
			condition = "1 = 0"
		}

	case identifier.FilterOperatorNotIn:
		if len(values) > 0 {
			condition = fmt.Sprintf("%s NOT IN ?", field)
			args = []interface{}{values}
		} else {
			// Handle empty NOT IN clause - return all results
			condition = "1 = 1"
		}

	case identifier.FilterOperatorIsNull:
		condition = fmt.Sprintf("%s IS NULL", field)

	case identifier.FilterOperatorIsNotNull:
		condition = fmt.Sprintf("%s IS NOT NULL", field)

	case identifier.FilterOperatorBetween:
		if len(values) >= 2 {
			condition = fmt.Sprintf("%s BETWEEN ? AND ?", field)
			args = []interface{}{values[0], values[1]}
		}

	case identifier.FilterOperatorContains:
		// For JSON fields - PostgreSQL specific
		condition = fmt.Sprintf("%s @> ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorHas:
		// For JSON fields - PostgreSQL specific
		condition = fmt.Sprintf("%s ?", field)
		args = []interface{}{value}

	default:
		// Unknown operator, skip this filter
		return query
	}

	// Apply the condition with proper logical operator
	if isFirst {
		return query.Where(condition, args...)
	} else if useOr {
		return query.Or(condition, args...)
	} else {
		return query.Where(condition, args...)
	}
}

// ApplyQueryParams converts QueryParams to GORM query with filters, sorting, and soft-delete handling
func (fa *FilterApplier) ApplyQueryParams(query *gorm.DB, params interface{}) *gorm.DB {
	if params == nil {
		return query
	}

	// Use reflection to access QueryParams fields since we can't use generics in methods
	// This is a safe workaround for the generic method limitation
	val := reflect.ValueOf(params)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Extract filters
	if filtersField := val.FieldByName("Filters"); filtersField.IsValid() {
		if filters, ok := filtersField.Interface().([]identifier.FilterCriteria); ok && len(filters) > 0 {
			query = fa.ApplyFilters(query, filters)
		}
	}

	// Extract search
	if searchField := val.FieldByName("Search"); searchField.IsValid() {
		if search, ok := searchField.Interface().(string); ok && search != "" {
			// Basic search implementation - should be overridden in specific repositories
			query = query.Where("CAST(id AS TEXT) LIKE ?", "%"+search+"%")
		}
	}

	// Extract soft-delete visibility
	var onlyDeleted, includeDeleted bool
	if onlyDeletedField := val.FieldByName("OnlyDeleted"); onlyDeletedField.IsValid() {
		onlyDeleted, _ = onlyDeletedField.Interface().(bool)
	}
	if includeDeletedField := val.FieldByName("IncludeDeleted"); includeDeletedField.IsValid() {
		includeDeleted, _ = includeDeletedField.Interface().(bool)
	}

	if onlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	} else {
		query = query.Unscoped()
	}

	// Extract sorting
	if sortField := val.FieldByName("Sort"); sortField.IsValid() {
		if sorts, ok := sortField.Interface().([]queryparams.SortField); ok && len(sorts) > 0 {
			for _, sort := range sorts {
				query = query.Order(fmt.Sprintf("%s %s", sort.Field, sort.Order))
			}
		} else {
			query = query.Order("id ASC")
		}
	}

	// Extract preloads
	if preloadsField := val.FieldByName("Preloads"); preloadsField.IsValid() {
		if preloads, ok := preloadsField.Interface().([]string); ok {
			for _, preload := range preloads {
				query = query.Preload(preload)
			}
		}
	}

	return query
}

// ApplyIdentifier converts IIdentifier to GORM query conditions
func (fa *FilterApplier) ApplyIdentifier(query *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	if identifier == nil {
		return query
	}

	filters := identifier.ToFilterCriteria()
	return fa.ApplyFilters(query, filters)
}

// BuildQueryFromIdentifier creates a complete query from an IIdentifier
func BuildQueryFromIdentifier[T types.IBaseModel](db *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	query := db.Model(new(T))
	fa := NewFilterApplier()
	return fa.ApplyIdentifier(query, identifier)
}

// ValidateFilterValue checks if a filter value is compatible with the field type
func (fa *FilterApplier) ValidateFilterValue(fieldName string, value interface{}) error {
	if value == nil {
		return nil
	}

	// Basic type validation - could be extended for more sophisticated validation
	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return nil
	case []interface{}:
		// Validate each element in slice
		for _, elem := range v {
			if err := fa.ValidateFilterValue(fieldName, elem); err != nil {
				return err
			}
		}
		return nil
	default:
		// Check if it's a slice or array using reflection
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return nil
		}
		return fmt.Errorf("unsupported filter value type for field %s: %T", fieldName, value)
	}
}
