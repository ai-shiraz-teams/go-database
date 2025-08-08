package query

import (
	"fmt"
	"reflect"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

var globalTypeRegistry = NewTypeRegistry()

type QueryParamsLiteral[T domain.IBaseModel] struct {
	Filter         FilterLiteral[T] `json:"filter,omitempty"`
	Sort           SortLiteral[T]   `json:"sort,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
	Preloads       []string         `json:"preloads,omitempty"`
	IncludeDeleted bool             `json:"includeDeleted,omitempty"`
	OnlyDeleted    bool             `json:"onlyDeleted,omitempty"`
}

type FilterLiteral[T domain.IBaseModel] map[string]FilterLiteralValue[T]

type FilterLiteralValue[T domain.IBaseModel] interface {
	isFilterLiteralValue()
}

type FilterOperatorLiteralValue[T domain.IBaseModel, V comparable] struct {
	FilterOperatorLiteral[V]
}

func (f FilterOperatorLiteralValue[T, V]) isFilterLiteralValue() {}

type FilterLogicalValue[T domain.IBaseModel] struct {
	Values []FilterLiteral[T]
}

func (f FilterLogicalValue[T]) isFilterLiteralValue() {}

type DirectFilterValue[V comparable] struct {
	Value V
}

func (f DirectFilterValue[V]) isFilterLiteralValue() {}

type SortLiteral[T domain.IBaseModel] map[string]SortDirection

type FilterOperatorLiteral[V any] struct {
	Eq        *V    `json:"eq,omitempty"`
	Neq       *V    `json:"neq,omitempty"`
	Gt        *V    `json:"gt,omitempty"`
	Gte       *V    `json:"gte,omitempty"`
	Lt        *V    `json:"lt,omitempty"`
	Lte       *V    `json:"lte,omitempty"`
	Like      *V    `json:"like,omitempty"`
	Contains  *V    `json:"contains,omitempty"`
	In        []V   `json:"in,omitempty"`
	NotIn     []V   `json:"not_in,omitempty"`
	Between   []V   `json:"between,omitempty"`
	IsNull    *bool `json:"is_null,omitempty"`
	IsNotNull *bool `json:"is_not_null,omitempty"`
}

type StringFilterOperatorLiteral = FilterOperatorLiteral[string]

type IntFilterOperatorLiteral = FilterOperatorLiteral[int]

type BoolFilterOperatorLiteral = FilterOperatorLiteral[bool]

type ComplexFilterLiteral[T domain.IBaseModel] struct {
	And []FilterLiteral[T] `json:"and,omitempty"`
	Or  []FilterLiteral[T] `json:"or,omitempty"`
}

func (qpl *QueryParamsLiteral[T]) ToQueryParams() (*QueryParams[T], error) {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

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

	if qpl.Filter != nil {
		filter, err := qpl.Filter.ToStronglyTypedFilter(globalTypeRegistry, modelType)
		if err != nil {
			return nil, err
		}
		qp.filter = filter
	}

	if qpl.Sort != nil {
		sortFields, err := qpl.Sort.ToSortFields(globalTypeRegistry, modelType)
		if err != nil {
			return nil, err
		}
		qp.sort = sortFields
	}

	err = qp.PrepareDefaults()
	if err != nil {
		return nil, err
	}

	return qp, nil
}

func (fl FilterLiteral[T]) ToStronglyTypedFilter(registry TypeRegistry, modelType reflect.Type) (StronglyTypedFilter[T], error) {
	criteria := []domain.FilterCriteria{}

	for fieldPath, value := range fl {

		if fieldPath == "AND" || fieldPath == "Or" {

			continue
		}

		fieldInfo, err := registry.ValidateFieldPath(modelType, fieldPath)
		if err != nil {
			return nil, err
		}

		fieldCriteria, err := fl.convertValueToFilterCriteria(fieldPath, value, fieldInfo)
		if err != nil {
			return nil, err
		}

		criteria = append(criteria, fieldCriteria...)
	}

	return &stronglyTypedFilter[T]{criteria: criteria}, nil
}

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

func (fl FilterLiteral[T]) convertLogicalValueToCriteria(fieldPath string, logical FilterLogicalValue[T]) ([]domain.FilterCriteria, error) {

	return []domain.FilterCriteria{}, nil
}

func (sl SortLiteral[T]) ToSortFields(registry TypeRegistry, modelType reflect.Type) ([]domain.SortField, error) {
	var sortFields []domain.SortField

	for fieldPath, direction := range sl {

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

func NewQueryParams[T domain.IBaseModel](literal QueryParamsLiteral[T]) (QueryParamsType[T], error) {
	return literal.ToQueryParams()
}

func Builder[T domain.IBaseModel]() QueryParamsBuilder[T] {
	return NewQueryParamsBuilder[T](globalTypeRegistry)
}

func Filter[T domain.IBaseModel]() StronglyTypedFilterBuilder[T] {
	return NewStronglyTypedFilter[T](globalTypeRegistry)
}

func RegisterType[T domain.IBaseModel]() error {
	var zero T
	return globalTypeRegistry.RegisterType(zero)
}

func GetTypeRegistry() TypeRegistry {
	return globalTypeRegistry
}

func Eq[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Eq: &value}
}

func Neq[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Neq: &value}
}

func Gt[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Gt: &value}
}

func Gte[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Gte: &value}
}

func Lt[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Lt: &value}
}

func Lte[V any](value V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Lte: &value}
}

func Like(pattern string) FilterOperatorLiteral[string] {
	return FilterOperatorLiteral[string]{Like: &pattern}
}

func Contains(substring string) FilterOperatorLiteral[string] {
	return FilterOperatorLiteral[string]{Contains: &substring}
}

func In[V any](values ...V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{In: values}
}

func NotIn[V any](values ...V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{NotIn: values}
}

func Between[V any](start, end V) FilterOperatorLiteral[V] {
	return FilterOperatorLiteral[V]{Between: []V{start, end}}
}

func IsNull[V any]() FilterOperatorLiteral[V] {
	t := true
	return FilterOperatorLiteral[V]{IsNull: &t}
}

func IsNotNull[V any]() FilterOperatorLiteral[V] {
	t := true
	return FilterOperatorLiteral[V]{IsNotNull: &t}
}
