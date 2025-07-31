package query

import (
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"
)

func NewQueryParams[T types.IBaseModel]() *QueryParams[T] {
	return &QueryParams[T]{
		Page:           1,
		PageSize:       50,
		Search:         "",
		Sort:           make([]SortField, 0),
		Filters:        make([]identifier.FilterCriteria, 0),
		IncludeDeleted: false,
		OnlyDeleted:    false,
		Preloads:       make([]string, 0),
	}
}

func (qp *QueryParams[T]) PrepareDefaults() *QueryParams[T] {

	normalizedPage, normalizedPageSize := NormalizePagination(qp.Offset, qp.Limit, qp.Page, qp.PageSize)

	validatedPage, validatedPageSize := ValidatePaginationBounds(normalizedPage, normalizedPageSize, 200)

	qp.Page = validatedPage
	qp.PageSize = validatedPageSize

	qp.ComputedOffset, qp.ComputedLimit = CalculateOffsetLimit(qp.Page, qp.PageSize)

	if qp.Sort == nil {
		qp.Sort = make([]SortField, 0)
	}
	if qp.Filters == nil {
		qp.Filters = make([]identifier.FilterCriteria, 0)
	}
	if qp.Preloads == nil {
		qp.Preloads = make([]string, 0)
	}

	return qp
}

func (qp *QueryParams[T]) WithFilters(identifier identifier.IIdentifier) *QueryParams[T] {
	if identifier != nil {
		qp.Filters = identifier.ToFilterCriteria()
	}
	return qp
}

func (qp *QueryParams[T]) AddSort(field string, order SortOrder) *QueryParams[T] {
	qp.Sort = append(qp.Sort, SortField{
		Field: field,
		Order: order,
	})
	return qp
}

func (qp *QueryParams[T]) AddSortAsc(field string) *QueryParams[T] {
	return qp.AddSort(field, SortOrderAsc)
}

func (qp *QueryParams[T]) AddSortDesc(field string) *QueryParams[T] {
	return qp.AddSort(field, SortOrderDesc)
}

func (qp *QueryParams[T]) ClearSort() *QueryParams[T] {
	qp.Sort = make([]SortField, 0)
	return qp
}

func (qp *QueryParams[T]) WithSearch(searchTerm string) *QueryParams[T] {
	qp.Search = searchTerm
	return qp
}

func (qp *QueryParams[T]) WithPreloads(preloads []string) *QueryParams[T] {
	qp.Preloads = preloads
	return qp
}

func (qp *QueryParams[T]) AddPreload(preload string) *QueryParams[T] {
	qp.Preloads = append(qp.Preloads, preload)
	return qp
}

func (qp *QueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) *QueryParams[T] {
	qp.IncludeDeleted = includeDeleted
	qp.OnlyDeleted = onlyDeleted
	return qp
}

func (qp *QueryParams[T]) IncludeDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = true
	qp.OnlyDeleted = false
	return qp
}

func (qp *QueryParams[T]) OnlyDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = false
	qp.OnlyDeleted = true
	return qp
}

func (qp *QueryParams[T]) ExcludeDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = false
	qp.OnlyDeleted = false
	return qp
}

func (qp *QueryParams[T]) HasSearch() bool {
	return qp.Search != ""
}

func (qp *QueryParams[T]) HasFilters() bool {
	return len(qp.Filters) > 0
}

func (qp *QueryParams[T]) HasSort() bool {
	return len(qp.Sort) > 0
}

func (qp *QueryParams[T]) HasPreloads() bool {
	return len(qp.Preloads) > 0
}

func (qp *QueryParams[T]) Clone() *QueryParams[T] {
	newParams := &QueryParams[T]{
		Page:           qp.Page,
		PageSize:       qp.PageSize,
		Offset:         qp.Offset,
		Limit:          qp.Limit,
		ComputedOffset: qp.ComputedOffset,
		ComputedLimit:  qp.ComputedLimit,
		Search:         qp.Search,
		IncludeDeleted: qp.IncludeDeleted,
		OnlyDeleted:    qp.OnlyDeleted,
	}

	if qp.Sort != nil {
		newParams.Sort = make([]SortField, len(qp.Sort))
		copy(newParams.Sort, qp.Sort)
	}

	if qp.Filters != nil {
		newParams.Filters = make([]identifier.FilterCriteria, len(qp.Filters))
		copy(newParams.Filters, qp.Filters)
	}

	if qp.Preloads != nil {
		newParams.Preloads = make([]string, len(qp.Preloads))
		copy(newParams.Preloads, qp.Preloads)
	}

	return newParams
}
