package query

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

// CategoryEntity represents a sample entity for testing
type CategoryEntity struct {
	domain.BaseEntity
	Name        string   `json:"name" gorm:"column:name"`
	Description *string  `json:"description,omitempty" gorm:"column:description"`
	IsActive    bool     `json:"isActive" gorm:"column:is_active"`
	Budget      float64  `json:"budget" gorm:"column:budget"`
	Tags        []string `json:"tags" gorm:"type:json"`
	ParentID    *int     `json:"parentId,omitempty" gorm:"column:parent_id"`
}

func TestQueryParamsType_SimpleTest(t *testing.T) {
	// Simple test to check if the basic structure compiles
	literal := QueryParamsLiteral[*CategoryEntity]{
		Filter: FilterLiteral[*CategoryEntity]{
			"name": DirectFilterValue[string]{Value: "test"},
		},
		Limit:  10,
		Offset: 0,
	}

	// Convert to QueryParams
	_, err := literal.ToQueryParams()
	if err != nil {
		t.Logf("Conversion failed: %v", err)
	}
}
