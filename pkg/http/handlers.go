package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListService defines the interface for services that can handle list queries
type ListService interface {
	ListWithQuery(params QueryParams) (interface{}, error)
}

// EntityListHandler creates an HTTP handler for listing entities with query parameters
func EntityListHandler(service ListService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query QueryParams

		// Validate explicitly provided parameters before binding
		if err := query.ValidateExplicitParams(c.Request.URL.Query()); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid query parameters",
				"details": err.Error(),
			})
			return
		}

		// Bind query parameters from request
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid query parameters",
				"details": err.Error(),
			})
			return
		}

		// Validate and prepare defaults
		if err := query.PrepareDefaults(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to prepare query parameters",
				"details": err.Error(),
			})
			return
		}

		// Call service with validated query parameters
		result, err := service.ListWithQuery(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to process query",
				"details": err.Error(),
			})
			return
		}

		// Return successful response
		c.JSON(http.StatusOK, gin.H{
			"data":   result,
			"status": "success",
		})
	}
}

// ExampleListHandler demonstrates a concrete implementation
func ExampleListHandler(c *gin.Context) {
	var query QueryParams

	// Validate explicitly provided parameters before binding
	if err := query.ValidateExplicitParams(c.Request.URL.Query()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Bind query parameters from request
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Validate and prepare defaults
	if err := query.PrepareDefaults(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to prepare query parameters",
			"details": err.Error(),
		})
		return
	}

	// Convert to domain query params
	domainParams := query.ToSimpleQueryParams()

	// Mock response data for demonstration
	mockData := gin.H{
		"items": []gin.H{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		},
		"pagination": gin.H{
			"limit":  query.Limit,
			"offset": query.Offset,
			"total":  42,
		},
		"filters": gin.H{
			"filter":         query.GetFilterValue(),
			"includeDeleted": query.IncludeDeleted,
			"onlyDeleted":    query.OnlyDeleted,
		},
		"sort":     query.Sort,
		"preloads": query.Preloads,
		"domainParams": gin.H{
			"limit":   domainParams.Limit(),
			"offset":  domainParams.Offset(),
			"hasSort": domainParams.HasSort(),
		},
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"data":   mockData,
		"status": "success",
	})
}
