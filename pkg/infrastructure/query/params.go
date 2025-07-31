package query

import (
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"
)

type QueryParams[T types.IBaseModel] struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"pageSize" query:"pageSize"`
	Offset   int `json:"offset" query:"offset"`
	Limit    int `json:"limit" query:"limit"`

	ComputedOffset int `json:"-"`
	ComputedLimit  int `json:"-"`

	Search string `json:"search,omitempty" query:"search"`

	Sort []SortField `json:"sort,omitempty"`

	Filters []identifier.FilterCriteria `json:"filters,omitempty"`

	IncludeDeleted bool `json:"includeDeleted,omitempty" query:"includeDeleted"`
	OnlyDeleted    bool `json:"onlyDeleted,omitempty" query:"onlyDeleted"`

	Preloads []string `json:"preloads,omitempty" query:"preloads"`
}
