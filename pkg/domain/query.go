package domain

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

type SortMap[T IBaseModel] map[string]SortOrder

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
	ToFilterCriteria() []FilterCriteria
	HasSort() bool
	HasPreloads() bool

	GetIncludeDeleted() bool
	GetOnlyDeleted() bool
	WithDeletedVisibility(includeDeleted, onlyDeleted bool) IQueryParams[T]

	ToSortFields() []SortField

	WithFilters(identifier interface{}) IQueryParams[T]
	AddSortDesc(field string) IQueryParams[T]
	AddSortAsc(field string) IQueryParams[T]
}

type SimpleQueryParams[T IBaseModel] struct {
	filter         interface{}
	sort           interface{}
	limit          int
	offset         int
	preloads       []string
	includeDeleted bool
	onlyDeleted    bool
	filters        []FilterCriteria
}

func NewQueryParams[T IBaseModel]() IQueryParams[T] {
	return &SimpleQueryParams[T]{
		limit:    50,
		offset:   0,
		preloads: make([]string, 0),
		filters:  make([]FilterCriteria, 0),
	}
}

func NewSimpleQueryParams[T IBaseModel]() IQueryParams[T] {
	return NewQueryParams[T]()
}

func (q *SimpleQueryParams[T]) Filter() interface{} { return q.filter }
func (q *SimpleQueryParams[T]) Sort() interface{}   { return q.sort }
func (q *SimpleQueryParams[T]) Limit() int          { return q.limit }
func (q *SimpleQueryParams[T]) Offset() int         { return q.offset }
func (q *SimpleQueryParams[T]) Preloads() []string  { return q.preloads }
func (q *SimpleQueryParams[T]) WithFilter(filter interface{}) IQueryParams[T] {
	q.filter = filter
	return q
}
func (q *SimpleQueryParams[T]) WithSort(sort interface{}) IQueryParams[T] { q.sort = sort; return q }
func (q *SimpleQueryParams[T]) WithLimit(limit int) IQueryParams[T]       { q.limit = limit; return q }
func (q *SimpleQueryParams[T]) WithOffset(offset int) IQueryParams[T]     { q.offset = offset; return q }
func (q *SimpleQueryParams[T]) WithPreloads(preloads []string) IQueryParams[T] {
	q.preloads = preloads
	return q
}
func (q *SimpleQueryParams[T]) PrepareDefaults() error             { return nil }
func (q *SimpleQueryParams[T]) HasFilters() bool                   { return len(q.filters) > 0 }
func (q *SimpleQueryParams[T]) ToFilterCriteria() []FilterCriteria { return q.filters }
func (q *SimpleQueryParams[T]) HasSort() bool                      { return q.sort != nil }
func (q *SimpleQueryParams[T]) HasPreloads() bool                  { return len(q.preloads) > 0 }
func (q *SimpleQueryParams[T]) GetIncludeDeleted() bool            { return q.includeDeleted }
func (q *SimpleQueryParams[T]) GetOnlyDeleted() bool               { return q.onlyDeleted }
func (q *SimpleQueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) IQueryParams[T] {
	q.includeDeleted = includeDeleted
	q.onlyDeleted = onlyDeleted
	return q
}
func (q *SimpleQueryParams[T]) ToSortFields() []SortField {

	if q.sort == nil {
		return []SortField{}
	}

	if sortMap, ok := q.sort.(SortMap[T]); ok {
		var fields []SortField
		for field, order := range sortMap {
			fields = append(fields, SortField{Field: field, Order: order})
		}
		return fields
	}

	if sortFields, ok := q.sort.([]SortField); ok {
		return sortFields
	}

	return []SortField{}
}

func (q *SimpleQueryParams[T]) WithFilters(identifier interface{}) IQueryParams[T] {

	if id, ok := identifier.(interface{ ToFilterCriteria() []FilterCriteria }); ok {
		q.filters = id.ToFilterCriteria()
	}
	return q
}

func (q *SimpleQueryParams[T]) AddSortDesc(field string) IQueryParams[T] {
	return q.AddSort(field, SortOrderDesc)
}

func (q *SimpleQueryParams[T]) AddSortAsc(field string) IQueryParams[T] {
	return q.AddSort(field, SortOrderAsc)
}

func (q *SimpleQueryParams[T]) AddSort(field string, order SortOrder) IQueryParams[T] {

	currentSorts := q.ToSortFields()
	currentSorts = append(currentSorts, SortField{Field: field, Order: order})
	q.sort = currentSorts
	return q
}

type FilterCriteria struct {
	Field     string           `json:"field"`
	Operator  FilterOperator   `json:"operator"`
	Value     interface{}      `json:"value,omitempty"`
	Values    []interface{}    `json:"values,omitempty"`
	Group     []FilterCriteria `json:"group,omitempty"`
	LogicalOp LogicalOperator  `json:"logical_op,omitempty"`
}

type FilterOperator string

const (
	FilterOperatorEqual        FilterOperator = "="
	FilterOperatorNotEqual     FilterOperator = "!="
	FilterOperatorGreaterThan  FilterOperator = ">"
	FilterOperatorGreaterEqual FilterOperator = ">="
	FilterOperatorLessThan     FilterOperator = "<"
	FilterOperatorLessEqual    FilterOperator = "<="
	FilterOperatorLike         FilterOperator = "LIKE"
	FilterOperatorIn           FilterOperator = "IN"
	FilterOperatorNotIn        FilterOperator = "NOT IN"
	FilterOperatorIsNull       FilterOperator = "IS NULL"
	FilterOperatorIsNotNull    FilterOperator = "IS NOT NULL"
	FilterOperatorBetween      FilterOperator = "BETWEEN"
	FilterOperatorContains     FilterOperator = "CONTAINS"
	FilterOperatorHas          FilterOperator = "HAS"
)

type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "AND"
	LogicalOperatorOr  LogicalOperator = "OR"
)

type RecursiveFilter struct {
	Field     string            `json:"field,omitempty"`
	Operator  FilterOperator    `json:"operator,omitempty"`
	Value     interface{}       `json:"value,omitempty"`
	Values    []interface{}     `json:"values,omitempty"`
	Children  []RecursiveFilter `json:"children,omitempty"`
	LogicalOp LogicalOperator   `json:"logical_op,omitempty"`
}

type RecursiveFilterParser struct{}

func NewRecursiveFilterParser() *RecursiveFilterParser {
	return &RecursiveFilterParser{}
}

func (rfp *RecursiveFilterParser) Parse(filter RecursiveFilter) []FilterCriteria {
	var criteria []FilterCriteria

	if filter.Field != "" {
		criterion := FilterCriteria{
			Field:     filter.Field,
			Operator:  filter.Operator,
			Value:     filter.Value,
			Values:    filter.Values,
			LogicalOp: filter.LogicalOp,
		}
		criteria = append(criteria, criterion)
	}

	for _, child := range filter.Children {
		childCriteria := rfp.Parse(child)
		criteria = append(criteria, childCriteria...)
	}

	return criteria
}
