package query

import (
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
)

// QueryParams provides a typed, reusable structure for paginated repository access.
// It supports pagination, filtering, sorting, search, preloading, and soft-delete visibility.
// This struct is designed to be compatible with Echo query binding and JSON serialization.
type QueryParams[T types.IBaseModel] struct {
	// Pagination fields
	Page     int `json:"page" query:"page"`         // Page number (1-based)
	PageSize int `json:"pageSize" query:"pageSize"` // Number of items per page
	Offset   int `json:"-"`                         // Calculated offset (auto-computed from Page and PageSize)
	Limit    int `json:"-"`                         // Calculated limit (auto-computed from PageSize)

	// Search functionality
	Search string `json:"search,omitempty" query:"search"` // Free-text search term

	// Sorting
	Sort []SortField `json:"sort,omitempty"` // Multiple sort fields with direction

	// Advanced filtering using IIdentifier system
	Filters []identifier.FilterCriteria `json:"filters,omitempty"`

	// Soft-delete visibility control
	IncludeDeleted bool `json:"includeDeleted,omitempty" query:"includeDeleted"` // Include soft-deleted records
	OnlyDeleted    bool `json:"onlyDeleted,omitempty" query:"onlyDeleted"`       // Show only soft-deleted records

	// Eager loading relationships
	Preloads []string `json:"preloads,omitempty" query:"preloads"` // List of relations to preload
}
