package unit_of_work

import (
	"fmt"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	query2 "github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"
	"reflect"

	"gorm.io/gorm"
)

type FilterApplier struct{}

func NewFilterApplier() *FilterApplier {
	return &FilterApplier{}
}

func (fa *FilterApplier) ApplyFilters(query *gorm.DB, filters []identifier.FilterCriteria) *gorm.DB {
	if len(filters) == 0 {
		return query
	}

	for i, filter := range filters {
		isFirst := i == 0
		useOr := false

		if !isFirst && i > 0 {
			prevFilter := filters[i-1]
			useOr = prevFilter.LogicalOp == identifier.LogicalOperatorOr
		}

		query = fa.applyFilter(query, filter, isFirst, useOr)
	}

	return query
}

func (fa *FilterApplier) applyFilter(query *gorm.DB, filter identifier.FilterCriteria, isFirst bool, useOr bool) *gorm.DB {
	if len(filter.Group) > 0 {
		return fa.applyGroupFilter(query, filter, isFirst, useOr)
	}

	return fa.applySingleFilter(query, filter, isFirst, useOr)
}

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
			condition = "1 = 0"
		}

	case identifier.FilterOperatorNotIn:
		if len(values) > 0 {
			condition = fmt.Sprintf("%s NOT IN ?", field)
			args = []interface{}{values}
		} else {
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
		condition = fmt.Sprintf("%s @> ?", field)
		args = []interface{}{value}

	case identifier.FilterOperatorHas:
		condition = fmt.Sprintf("%s ?", field)
		args = []interface{}{value}

	default:
		return query
	}

	if isFirst {
		return query.Where(condition, args...)
	} else if useOr {
		return query.Or(condition, args...)
	} else {
		return query.Where(condition, args...)
	}
}

func (fa *FilterApplier) ApplyQueryParams(query *gorm.DB, params interface{}) *gorm.DB {
	if params == nil {
		return query
	}

	val := reflect.ValueOf(params)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if filtersField := val.FieldByName("Filters"); filtersField.IsValid() {
		if filters, ok := filtersField.Interface().([]identifier.FilterCriteria); ok && len(filters) > 0 {
			query = fa.ApplyFilters(query, filters)
		}
	}

	if searchField := val.FieldByName("Search"); searchField.IsValid() {
		if search, ok := searchField.Interface().(string); ok && search != "" {
			query = query.Where("CAST(id AS TEXT) LIKE ?", "%"+search+"%")
		}
	}

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

	if sortField := val.FieldByName("Sort"); sortField.IsValid() {
		if sorts, ok := sortField.Interface().([]query2.SortField); ok && len(sorts) > 0 {
			for _, sort := range sorts {
				query = query.Order(fmt.Sprintf("%s %s", sort.Field, sort.Order))
			}
		} else {
			query = query.Order("id ASC")
		}
	}

	if preloadsField := val.FieldByName("Preloads"); preloadsField.IsValid() {
		if preloads, ok := preloadsField.Interface().([]string); ok {
			for _, preload := range preloads {
				query = query.Preload(preload)
			}
		}
	}

	return query
}

func (fa *FilterApplier) ApplyIdentifier(query *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	if identifier == nil {
		return query
	}

	filters := identifier.ToFilterCriteria()
	return fa.ApplyFilters(query, filters)
}

func BuildQueryFromIdentifier[T types.IBaseModel](db *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	query := db.Model(new(T))
	fa := NewFilterApplier()
	return fa.ApplyIdentifier(query, identifier)
}

func (fa *FilterApplier) ValidateFilterValue(fieldName string, value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return nil
	case []interface{}:
		for _, elem := range v {
			if err := fa.ValidateFilterValue(fieldName, elem); err != nil {
				return err
			}
		}
		return nil
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return nil
		}
		return fmt.Errorf("unsupported filter value type for field %s: %T", fieldName, value)
	}
}
