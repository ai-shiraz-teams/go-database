package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListService is a mock implementation of the ListService interface
type MockListService struct {
	mock.Mock
}

func (m *MockListService) ListWithQuery(params QueryParams) (interface{}, error) {
	args := m.Called(params)
	return args.Get(0), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestEntityListHandler_ValidQueryParams(t *testing.T) {
	// Arrange
	mockService := new(MockListService)

	// Mock successful response
	expectedData := gin.H{"items": []gin.H{{"id": 1, "name": "Test"}}}
	mockService.On("ListWithQuery", mock.AnythingOfType("QueryParams")).Return(expectedData, nil)

	router := setupTestRouter()
	router.GET("/entities", EntityListHandler(mockService))

	// Act - Test valid query parameters
	req, _ := http.NewRequest("GET", "/entities?limit=25&offset=2&sort=name:asc&filter=active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify service was called with correct parameters
	mockService.AssertExpectations(t)

	// Verify the call was made with proper parameters
	calls := mockService.Calls
	assert.Len(t, calls, 1) // ListWithQuery call

	// Check that ListWithQuery was called with proper parameters
	listCall := calls[0]
	assert.Equal(t, "ListWithQuery", listCall.Method)
	queryParams := listCall.Arguments[0].(QueryParams)
	assert.Equal(t, 25, queryParams.Limit)
	assert.Equal(t, 2, queryParams.Offset)
	assert.Equal(t, "name:asc", queryParams.Sort)
	assert.Equal(t, "active", queryParams.Filter)
}

func TestEntityListHandler_InvalidQueryParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid offset - zero",
			queryParams:    "offset=0&limit=10",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject offset=0 since offset must start from 1",
		},
		{
			name:           "Invalid offset - negative",
			queryParams:    "offset=-1&limit=10",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject negative offset",
		},
		{
			name:           "Invalid limit - zero",
			queryParams:    "offset=1&limit=0",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject limit=0",
		},
		{
			name:           "Invalid limit - negative",
			queryParams:    "offset=1&limit=-5",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject negative limit",
		},
		{
			name:           "Invalid limit - too large",
			queryParams:    "offset=1&limit=500",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject limit > 200",
		},
		{
			name:           "Invalid offset - non-numeric",
			queryParams:    "offset=abc&limit=10",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject non-numeric offset",
		},
		{
			name:           "Invalid limit - non-numeric",
			queryParams:    "offset=1&limit=xyz",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject non-numeric limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockListService)
			router := setupTestRouter()
			router.GET("/entities", EntityListHandler(mockService))

			// Act
			req, _ := http.NewRequest("GET", "/entities?"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// Verify service was NOT called for invalid requests
			mockService.AssertNotCalled(t, "ListWithQuery")
		})
	}
}

func TestEntityListHandler_QueryBindingWorks(t *testing.T) {
	// Arrange
	mockService := new(MockListService)
	expectedData := gin.H{"items": []gin.H{}}
	mockService.On("ListWithQuery", mock.AnythingOfType("QueryParams")).Return(expectedData, nil)

	router := setupTestRouter()
	router.GET("/entities", EntityListHandler(mockService))

	// Test complex query parameters
	queryParams := url.Values{}
	queryParams.Set("limit", "100")
	queryParams.Set("offset", "3")
	queryParams.Set("sort", "name:asc,-created_at,status:desc")
	queryParams.Set("filter", "status=active")
	queryParams.Set("include_deleted", "true")
	queryParams.Set("only_deleted", "false")
	queryParams.Set("preloads", "user,profile,comments")

	// Act
	req, _ := http.NewRequest("GET", "/entities?"+queryParams.Encode(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Get the query parameters that were passed to the service
	calls := mockService.Calls
	assert.Len(t, calls, 1)

	listCall := calls[0]
	queryParamsArg := listCall.Arguments[0].(QueryParams)

	// Verify all parameters were correctly bound
	assert.Equal(t, 100, queryParamsArg.Limit)
	assert.Equal(t, 3, queryParamsArg.Offset)
	assert.Equal(t, "name:asc,-created_at,status:desc", queryParamsArg.Sort)
	assert.Equal(t, "status=active", queryParamsArg.Filter)
	assert.True(t, queryParamsArg.IncludeDeleted)
	assert.False(t, queryParamsArg.OnlyDeleted)
	assert.Equal(t, "user,profile,comments", queryParamsArg.Preloads)

	mockService.AssertExpectations(t)
}

func TestEntityListHandler_DefaultValues(t *testing.T) {
	// Arrange
	mockService := new(MockListService)
	expectedData := gin.H{"items": []gin.H{}}
	mockService.On("ListWithQuery", mock.AnythingOfType("QueryParams")).Return(expectedData, nil)

	router := setupTestRouter()
	router.GET("/entities", EntityListHandler(mockService))

	// Act - Request with no query parameters
	req, _ := http.NewRequest("GET", "/entities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify default values were applied
	calls := mockService.Calls
	listCall := calls[0]
	queryParamsArg := listCall.Arguments[0].(QueryParams)

	// Check defaults: offset=1, limit=50 (after PrepareDefaults)
	assert.Equal(t, 1, queryParamsArg.Offset)
	assert.Equal(t, 50, queryParamsArg.Limit)
	assert.False(t, queryParamsArg.IncludeDeleted)
	assert.False(t, queryParamsArg.OnlyDeleted)
	assert.Equal(t, "", queryParamsArg.Sort)
	assert.Equal(t, "", queryParamsArg.Preloads)

	mockService.AssertExpectations(t)
}

func TestEntityListHandler_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockListService)
	mockService.On("ListWithQuery", mock.AnythingOfType("QueryParams")).Return(nil, assert.AnError)

	router := setupTestRouter()
	router.GET("/entities", EntityListHandler(mockService))

	// Act
	req, _ := http.NewRequest("GET", "/entities?limit=10&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestExampleListHandler_Integration(t *testing.T) {
	// Arrange
	router := setupTestRouter()
	router.GET("/examples", ExampleListHandler)

	// Act - Test with valid parameters
	req, _ := http.NewRequest("GET", "/examples?limit=20&offset=1&sort=name:desc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response contains expected structure
	assert.Contains(t, w.Body.String(), "\"items\":")
	assert.Contains(t, w.Body.String(), "\"status\":\"success\"")
	assert.Contains(t, w.Body.String(), "\"limit\":20")
	assert.Contains(t, w.Body.String(), "\"offset\":1")
}

func TestExampleListHandler_InvalidParams(t *testing.T) {
	// Arrange
	router := setupTestRouter()
	router.GET("/examples", ExampleListHandler)

	// Act - Test with invalid offset
	req, _ := http.NewRequest("GET", "/examples?offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// Test edge cases for query parameter parsing
func TestQueryParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]string
		expectValid bool
		description string
	}{
		{
			name: "Boundary limit exactly 200",
			params: map[string]string{
				"limit":  "200",
				"offset": "1",
			},
			expectValid: true,
			description: "Should accept limit=200 as it's the maximum allowed",
		},
		{
			name: "Boundary limit 201",
			params: map[string]string{
				"limit":  "201",
				"offset": "1",
			},
			expectValid: false,
			description: "Should reject limit=201 as it exceeds maximum",
		},
		{
			name: "Very large offset",
			params: map[string]string{
				"limit":  "50",
				"offset": "999999",
			},
			expectValid: true,
			description: "Should accept large but valid offset",
		},
		{
			name: "Complex sort string",
			params: map[string]string{
				"limit":  "10",
				"offset": "1",
				"sort":   "field1:asc,-field2,field3:desc,-field4,field5:asc",
			},
			expectValid: true,
			description: "Should handle complex sort strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockListService)

			if tt.expectValid {
				expectedData := gin.H{"items": []gin.H{}}
				mockService.On("ListWithQuery", mock.AnythingOfType("QueryParams")).Return(expectedData, nil)
			}

			router := setupTestRouter()
			router.GET("/entities", EntityListHandler(mockService))

			// Act
			req := createTestRequest(t, "/entities", tt.params)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if tt.expectValid {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
				mockService.AssertExpectations(t)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
				mockService.AssertNotCalled(t, "ListWithQuery")
			}
		})
	}
}

// Helper function to create test requests with query parameters
func createTestRequest(t *testing.T, path string, params map[string]string) *http.Request {
	t.Helper()

	queryParams := url.Values{}
	for key, value := range params {
		queryParams.Set(key, value)
	}

	var fullURL string
	if len(queryParams) > 0 {
		fullURL = path + "?" + queryParams.Encode()
	} else {
		fullURL = path
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	return req
}

// Benchmark tests to ensure performance
func BenchmarkQueryParams_PrepareDefaults(b *testing.B) {
	qp := QueryParams{
		Offset: 0,
		Limit:  0,
		Sort:   "name:asc,-created_at,status:desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset to initial state
		qp.Offset = 0
		qp.Limit = 0

		err := qp.PrepareDefaults()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParams_ToSimpleQueryParams(b *testing.B) {
	qp := QueryParams{
		Offset:         2,
		Limit:          100,
		Sort:           "name:asc,-created_at,status:desc",
		Filter:         "status=active",
		IncludeDeleted: true,
		Preloads:       "user,profile,comments",
	}

	err := qp.PrepareDefaults()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domainParams := qp.ToSimpleQueryParams()
		if domainParams == nil {
			b.Fatal("ToSimpleQueryParams returned nil")
		}
	}
}
