package unit_of_work

import (
	"errors"
	"testing"

	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
)

// TestEntity for testing purposes
type TestEntity struct {
	types.BaseEntity
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TestTransactionOptions_Defaults(t *testing.T) {
	// Arrange
	options := TransactionOptions{}

	// Assert
	if options.IsolationLevel != "" {
		t.Errorf("Expected empty IsolationLevel, got %s", options.IsolationLevel)
	}
	if options.ReadOnly {
		t.Error("Expected ReadOnly to be false by default")
	}
	if options.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0, got %d", options.Timeout)
	}
}

func TestTransactionOptions_Values(t *testing.T) {
	// Arrange
	options := TransactionOptions{
		IsolationLevel: "READ_COMMITTED",
		ReadOnly:       true,
		Timeout:        30000,
	}

	// Assert
	if options.IsolationLevel != "READ_COMMITTED" {
		t.Errorf("Expected IsolationLevel 'READ_COMMITTED', got %s", options.IsolationLevel)
	}
	if !options.ReadOnly {
		t.Error("Expected ReadOnly to be true")
	}
	if options.Timeout != 30000 {
		t.Errorf("Expected Timeout 30000, got %d", options.Timeout)
	}
}

func TestBulkOperationResult_Defaults(t *testing.T) {
	// Arrange
	result := BulkOperationResult{}

	// Assert
	if result.SuccessCount != 0 {
		t.Errorf("Expected SuccessCount to be 0, got %d", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("Expected FailureCount to be 0, got %d", result.FailureCount)
	}
	if result.Errors != nil {
		t.Errorf("Expected Errors to be nil, got %v", result.Errors)
	}
	if result.ProcessedIDs != nil {
		t.Errorf("Expected ProcessedIDs to be nil, got %v", result.ProcessedIDs)
	}
}

func TestBulkOperationResult_WithData(t *testing.T) {
	// Arrange
	testErrors := []error{
		errors.New("test error"),
	}
	processedIDs := []int{1, 2, 3}

	result := BulkOperationResult{
		SuccessCount: 3,
		FailureCount: 1,
		Errors:       testErrors,
		ProcessedIDs: processedIDs,
	}

	// Assert
	if result.SuccessCount != 3 {
		t.Errorf("Expected SuccessCount 3, got %d", result.SuccessCount)
	}
	if result.FailureCount != 1 {
		t.Errorf("Expected FailureCount 1, got %d", result.FailureCount)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	if len(result.ProcessedIDs) != 3 {
		t.Errorf("Expected 3 processed IDs, got %d", len(result.ProcessedIDs))
	}
}

func TestBulkOperationResult_TotalOperations(t *testing.T) {
	// Arrange
	result := BulkOperationResult{
		SuccessCount: 5,
		FailureCount: 2,
	}

	// Act
	total := result.SuccessCount + result.FailureCount

	// Assert
	if total != 7 {
		t.Errorf("Expected total operations 7, got %d", total)
	}
}

func TestBulkOperationResult_SuccessRate(t *testing.T) {
	tests := []struct {
		name         string
		successCount int
		failureCount int
		expectedRate float64
	}{
		{"All success", 10, 0, 1.0},
		{"All failure", 0, 10, 0.0},
		{"Half success", 5, 5, 0.5},
		{"High success", 9, 1, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := BulkOperationResult{
				SuccessCount: tt.successCount,
				FailureCount: tt.failureCount,
			}

			// Act
			total := result.SuccessCount + result.FailureCount
			var rate float64
			if total > 0 {
				rate = float64(result.SuccessCount) / float64(total)
			}

			// Assert
			if rate != tt.expectedRate {
				t.Errorf("Expected success rate %.2f, got %.2f", tt.expectedRate, rate)
			}
		})
	}
}

func TestBulkOperationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name      string
		errors    []error
		hasErrors bool
	}{
		{"No errors", nil, false},
		{"Empty errors", []error{}, false},
		{"With errors", []error{errors.New("test error")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := BulkOperationResult{
				Errors: tt.errors,
			}

			// Act
			hasErrors := len(result.Errors) > 0

			// Assert
			if hasErrors != tt.hasErrors {
				t.Errorf("Expected hasErrors %v, got %v", tt.hasErrors, hasErrors)
			}
		})
	}
}

func TestBulkOperationResult_ErrorCount(t *testing.T) {
	// Arrange
	testErrors := make([]error, 3)
	for i := range testErrors {
		testErrors[i] = errors.New("test error")
	}

	result := BulkOperationResult{
		Errors: testErrors,
	}

	// Assert
	if len(result.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(result.Errors))
	}
}

func TestBulkOperationResult_ProcessedIDsCount(t *testing.T) {
	// Arrange
	processedIDs := []int{1, 2, 3, 4, 5}
	result := BulkOperationResult{
		ProcessedIDs: processedIDs,
	}

	// Assert
	if len(result.ProcessedIDs) != 5 {
		t.Errorf("Expected 5 processed IDs, got %d", len(result.ProcessedIDs))
	}
}

func TestBulkOperationResult_Consistency(t *testing.T) {
	// Test that SuccessCount matches ProcessedIDs length when no errors
	tests := []struct {
		name         string
		successCount int
		processedIDs []int
		shouldMatch  bool
	}{
		{"Matching counts", 3, []int{1, 2, 3}, true},
		{"Mismatched counts", 2, []int{1, 2, 3}, false},
		{"Empty both", 0, []int{}, true},
		{"Success but no IDs", 3, []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := BulkOperationResult{
				SuccessCount: tt.successCount,
				ProcessedIDs: tt.processedIDs,
			}

			// Act
			matches := result.SuccessCount == len(result.ProcessedIDs)

			// Assert
			if matches != tt.shouldMatch {
				t.Errorf("Expected consistency match %v, got %v", tt.shouldMatch, matches)
			}
		})
	}
}
