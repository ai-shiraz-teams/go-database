package query

import (
	"reflect"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

type FilterOperator[T domain.IBaseModel, V comparable] interface {
	Equal(value V) FieldFilterBuilder[T]

	NotEqual(value V) FieldFilterBuilder[T]
}

type ComparableFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]

	GreaterThan(value V) FieldFilterBuilder[T]

	GreaterOrEqual(value V) FieldFilterBuilder[T]

	LessThan(value V) FieldFilterBuilder[T]

	LessOrEqual(value V) FieldFilterBuilder[T]

	Between(start, end V) FieldFilterBuilder[T]
}

type StringFilterOperator[T domain.IBaseModel] interface {
	ComparableFilterOperator[T, string]

	Like(pattern string) FieldFilterBuilder[T]

	Contains(substring string) FieldFilterBuilder[T]

	StartsWith(prefix string) FieldFilterBuilder[T]

	EndsWith(suffix string) FieldFilterBuilder[T]

	IContains(substring string) FieldFilterBuilder[T]

	ILike(pattern string) FieldFilterBuilder[T]
}

type SliceFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]

	In(values []V) FieldFilterBuilder[T]

	NotIn(values []V) FieldFilterBuilder[T]
}

type NullableFilterOperator[T domain.IBaseModel, V comparable] interface {
	FilterOperator[T, V]

	IsNull() FieldFilterBuilder[T]

	IsNotNull() FieldFilterBuilder[T]
}

type BooleanFilterOperator[T domain.IBaseModel] interface {
	FilterOperator[T, bool]

	IsTrue() FieldFilterBuilder[T]

	IsFalse() FieldFilterBuilder[T]
}

type FieldFilterOperator[T domain.IBaseModel] interface {
	GetFieldPath() string

	GetFieldType() TypeInfo

	AsString() (StringFilterOperator[T], bool)

	AsComparableInt() (ComparableFilterOperator[T, int], bool)

	AsComparableFloat() (ComparableFilterOperator[T, float64], bool)

	AsBoolean() (BooleanFilterOperator[T], bool)

	AsNullableString() (NullableFilterOperator[T, string], bool)
}

type FieldFilterBuilder[T domain.IBaseModel] interface {
	GetFieldPath() string

	GetFilterCriteria() []domain.FilterCriteria

	And() StronglyTypedFilterBuilder[T]

	Or() StronglyTypedFilterBuilder[T]
}

type StronglyTypedFilterBuilder[T domain.IBaseModel] interface {
	GetField(fieldName string) FieldFilterOperator[T]

	And(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T]
	Or(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T]

	ToFilterCriteria() []domain.FilterCriteria
	Build() StronglyTypedFilter[T]
}

type StronglyTypedFilter[T domain.IBaseModel] interface {
	ToFilterCriteria() []domain.FilterCriteria

	Clone() StronglyTypedFilter[T]

	IsEmpty() bool
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type SortBuilder[T domain.IBaseModel] interface {
	GetField(fieldName string, direction SortDirection) SortBuilder[T]

	Asc(fieldName string) SortBuilder[T]
	Desc(fieldName string) SortBuilder[T]

	ThenBy(fieldName string, direction SortDirection) SortBuilder[T]
	ThenAsc(fieldName string) SortBuilder[T]
	ThenDesc(fieldName string) SortBuilder[T]

	Build() []domain.SortField
}

type QueryParamsBuilder[T domain.IBaseModel] interface {
	Filter() StronglyTypedFilterBuilder[T]
	WithFilter(filter StronglyTypedFilter[T]) QueryParamsBuilder[T]

	Sort() SortBuilder[T]
	WithSort(sorts []domain.SortField) QueryParamsBuilder[T]

	WithLimit(limit int) QueryParamsBuilder[T]
	WithOffset(offset int) QueryParamsBuilder[T]
	WithPage(page, pageSize int) QueryParamsBuilder[T]

	WithPreloads(preloads []string) QueryParamsBuilder[T]
	WithPreload(preload string) QueryParamsBuilder[T]

	IncludeDeleted() QueryParamsBuilder[T]
	OnlyDeleted() QueryParamsBuilder[T]
	ExcludeDeleted() QueryParamsBuilder[T]

	Build() QueryParams[T]
}

type QueryParams[T domain.IBaseModel] struct {
	filter StronglyTypedFilter[T]

	sort []domain.SortField

	limit    int
	offset   int
	page     int
	pageSize int

	preloads []string

	includeDeleted bool
	onlyDeleted    bool

	search string
}

var _ QueryParamsType[*domain.BaseEntity] = (*QueryParams[*domain.BaseEntity])(nil)

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

	return qp
}

func (qp *QueryParams[T]) GetSortFields() []domain.SortField {
	return qp.sort
}

func (qp *QueryParams[T]) SetSortFields(sorts []domain.SortField) {
	qp.sort = sorts
}

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

type QueryParamsType[T domain.IBaseModel] interface {
	domain.IQueryParams[T]

	GetStronglyTypedFilter() StronglyTypedFilter[T]
	SetStronglyTypedFilter(filter StronglyTypedFilter[T])

	ToLiteral() QueryParamsLiteral[T]
	FromLiteral(literal QueryParamsLiteral[T]) error

	ValidateFieldPaths() error
	GetAvailableFields() map[string]*FieldInfo

	BuildFilter() StronglyTypedFilterBuilder[T]
	BuildSort() SortBuilder[T]
}

type FilterShape[T domain.IBaseModel] map[string]FieldFilterOperator[T]

type SortShape[T domain.IBaseModel] map[string]SortDirection

func (qp *QueryParams[T]) GetStronglyTypedFilter() StronglyTypedFilter[T] {
	return qp.filter
}

func (qp *QueryParams[T]) SetStronglyTypedFilter(filter StronglyTypedFilter[T]) {
	qp.filter = filter
}

func (qp *QueryParams[T]) ToLiteral() QueryParamsLiteral[T] {
	literal := QueryParamsLiteral[T]{
		Limit:          qp.limit,
		Offset:         qp.offset,
		Preloads:       qp.preloads,
		IncludeDeleted: qp.includeDeleted,
		OnlyDeleted:    qp.onlyDeleted,
	}

	if qp.filter != nil && !qp.filter.IsEmpty() {
		literal.Filter = make(FilterLiteral[T])

	}

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

func (qp *QueryParams[T]) ValidateFieldPaths() error {

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

func (qp *QueryParams[T]) BuildFilter() StronglyTypedFilterBuilder[T] {
	return NewStronglyTypedFilter[T](globalTypeRegistry)
}

func (qp *QueryParams[T]) BuildSort() SortBuilder[T] {

	return nil
}

type TypeInfo struct {
	Kind        reflect.Kind
	Type        reflect.Type
	IsPointer   bool
	IsSlice     bool
	ElementType reflect.Type
}

type FieldInfo struct {
	Name     string
	JSONName string
	GormName string
	TypeInfo TypeInfo
	Path     []string
}

type TypeRegistry interface {
	RegisterType(model domain.IBaseModel) error

	GetFieldInfo(modelType reflect.Type, fieldPath string) (*FieldInfo, error)

	GetFields(modelType reflect.Type) map[string]*FieldInfo

	ValidateFieldPath(modelType reflect.Type, fieldPath string) (*FieldInfo, error)
}
