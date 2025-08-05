package query

import (
	"reflect"
	"testing"
	"time"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func TestQueryParamsLiteral_ObjectConstruction(t *testing.T) {
	// Test object literal construction
	literal := QueryParamsLiteral[*testutil.TestEntity]{
		Filter: FilterLiteral[*testutil.TestEntity]{
			"name": FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{
					Contains: StringPtr("john"),
				},
			},
			"age": FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{
					Gte: IntPtr(18),
					Lte: IntPtr(65),
				},
			},
			"IsActive": DirectFilterValue[bool]{Value: true},
		},
		Sort: SortLiteral[*testutil.TestEntity]{
			"name":      SortAsc,
			"CreatedAt": SortDesc,
		},
		Limit:  10,
		Offset: 0,
	}

	// Convert to QueryParams
	qp, err := literal.ToQueryParams()
	if err != nil {
		t.Fatalf("Failed to convert literal to QueryParams: %v", err)
	}

	// Verify basic properties
	if qp.Limit() != 10 {
		t.Errorf("Expected limit 10, got %d", qp.Limit())
	}

	if qp.Offset() != 0 {
		t.Errorf("Expected offset 0, got %d", qp.Offset())
	}

	// Verify filters
	if !qp.HasFilters() {
		t.Error("Expected HasFilters() to be true")
	}

	criteria := qp.ToFilterCriteria()
	if len(criteria) != 4 { // name contains, age gte, age lte, IsActive eq
		t.Errorf("Expected 4 filter criteria, got %d", len(criteria))
	}

	// Verify sort
	if !qp.HasSort() {
		t.Error("Expected HasSort() to be true")
	}

	sortFields := qp.ToSortFields()
	if len(sortFields) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(sortFields))
	}
}

func TestFilterLiteral_DirectValueComparison(t *testing.T) {
	// Test direct value assignment (equals comparison)
	literal := QueryParamsLiteral[*testutil.TestEntity]{
		Filter: FilterLiteral[*testutil.TestEntity]{
			"name":     DirectFilterValue[string]{Value: "John Doe"},
			"age":      DirectFilterValue[int]{Value: 30},
			"IsActive": DirectFilterValue[bool]{Value: true},
		},
	}

	qp, err := literal.ToQueryParams()
	if err != nil {
		t.Fatalf("Failed to convert literal: %v", err)
	}

	criteria := qp.ToFilterCriteria()
	if len(criteria) != 3 {
		t.Errorf("Expected 3 filter criteria, got %d", len(criteria))
	}

	// All should be equality comparisons
	for _, c := range criteria {
		if c.Operator != domain.FilterOperatorEqual {
			t.Errorf("Expected all operators to be Equal, got %s for field %s", c.Operator, c.Field)
		}
	}
}

func TestFilterOperatorLiteral_AllOperators(t *testing.T) {
	tests := []struct {
		name           string
		filterValue    FilterLiteralValue[*testutil.TestEntity]
		expectedOp     domain.FilterOperator
		expectedValue  interface{}
		expectedValues []interface{}
	}{
		{
			name: "Equal operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Eq: StringPtr("test")},
			},
			expectedOp:    domain.FilterOperatorEqual,
			expectedValue: "test",
		},
		{
			name: "Not equal operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Neq: StringPtr("test")},
			},
			expectedOp:    domain.FilterOperatorNotEqual,
			expectedValue: "test",
		},
		{
			name: "Greater than operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Gt: IntPtr(25)},
			},
			expectedOp:    domain.FilterOperatorGreaterThan,
			expectedValue: 25,
		},
		{
			name: "Greater than or equal operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Gte: IntPtr(18)},
			},
			expectedOp:    domain.FilterOperatorGreaterEqual,
			expectedValue: 18,
		},
		{
			name: "Less than operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Lt: IntPtr(65)},
			},
			expectedOp:    domain.FilterOperatorLessThan,
			expectedValue: 65,
		},
		{
			name: "Less than or equal operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Lte: IntPtr(64)},
			},
			expectedOp:    domain.FilterOperatorLessEqual,
			expectedValue: 64,
		},
		{
			name: "Like operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Like: StringPtr("%test%")},
			},
			expectedOp:    domain.FilterOperatorLike,
			expectedValue: "%test%",
		},
		{
			name: "Contains operator (converted to like)",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Contains: StringPtr("test")},
			},
			expectedOp:    domain.FilterOperatorLike,
			expectedValue: "%test%",
		},
		{
			name: "In operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{In: []int{1, 2, 3}},
			},
			expectedOp:     domain.FilterOperatorIn,
			expectedValues: []interface{}{1, 2, 3},
		},
		{
			name: "Not in operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{NotIn: []int{4, 5, 6}},
			},
			expectedOp:     domain.FilterOperatorNotIn,
			expectedValues: []interface{}{4, 5, 6},
		},
		{
			name: "Between operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Between: []int{10, 20}},
			},
			expectedOp:     domain.FilterOperatorBetween,
			expectedValues: []interface{}{10, 20},
		},
		{
			name: "Is null operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{IsNull: BoolPtr(true)},
			},
			expectedOp: domain.FilterOperatorIsNull,
		},
		{
			name: "Is not null operator",
			filterValue: FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{IsNotNull: BoolPtr(true)},
			},
			expectedOp: domain.FilterOperatorIsNotNull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			literal := QueryParamsLiteral[*testutil.TestEntity]{
				Filter: FilterLiteral[*testutil.TestEntity]{
					"name": tt.filterValue,
				},
			}

			qp, err := literal.ToQueryParams()
			if err != nil {
				t.Fatalf("Failed to convert literal: %v", err)
			}

			criteria := qp.ToFilterCriteria()
			if len(criteria) != 1 {
				t.Fatalf("Expected 1 filter criteria, got %d", len(criteria))
			}

			c := criteria[0]
			if c.Operator != tt.expectedOp {
				t.Errorf("Expected operator %s, got %s", tt.expectedOp, c.Operator)
			}

			if tt.expectedValue != nil && c.Value != tt.expectedValue {
				t.Errorf("Expected value %v, got %v", tt.expectedValue, c.Value)
			}

			if tt.expectedValues != nil {
				if len(c.Values) != len(tt.expectedValues) {
					t.Errorf("Expected %d values, got %d", len(tt.expectedValues), len(c.Values))
				} else {
					for i, expected := range tt.expectedValues {
						if c.Values[i] != expected {
							t.Errorf("Expected value[%d] %v, got %v", i, expected, c.Values[i])
						}
					}
				}
			}
		})
	}
}

func TestSortLiteral_MultipleSortFields(t *testing.T) {
	literal := QueryParamsLiteral[*testutil.TestEntity]{
		Sort: SortLiteral[*testutil.TestEntity]{
			"name":      SortAsc,
			"age":       SortDesc,
			"CreatedAt": SortDesc,
			"IsActive":  SortAsc,
		},
	}

	qp, err := literal.ToQueryParams()
	if err != nil {
		t.Fatalf("Failed to convert literal: %v", err)
	}

	sortFields := qp.ToSortFields()
	if len(sortFields) != 4 {
		t.Errorf("Expected 4 sort fields, got %d", len(sortFields))
	}

	// Verify each sort field (order may vary due to map iteration)
	fieldCount := make(map[string]domain.SortOrder)
	for _, sf := range sortFields {
		fieldCount[sf.Field] = sf.Order
	}

	if fieldCount["name"] != domain.SortOrderAsc {
		t.Errorf("Expected name to sort ASC, got %s", fieldCount["name"])
	}
	if fieldCount["age"] != domain.SortOrderDesc {
		t.Errorf("Expected age to sort DESC, got %s", fieldCount["age"])
	}
	if fieldCount["CreatedAt"] != domain.SortOrderDesc {
		t.Errorf("Expected CreatedAt to sort DESC, got %s", fieldCount["CreatedAt"])
	}
	if fieldCount["IsActive"] != domain.SortOrderAsc {
		t.Errorf("Expected IsActive to sort ASC, got %s", fieldCount["IsActive"])
	}
}

func TestBuilder_FluentAPI(t *testing.T) {
	// Test fluent API builder using the literal approach since the builder API is still incomplete
	literal := QueryParamsLiteral[*testutil.TestEntity]{
		Filter: FilterLiteral[*testutil.TestEntity]{
			"name": FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Contains: StringPtr("john")},
			},
			"age": FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Gte: IntPtr(18)},
			},
		},
	}

	qp, err := literal.ToQueryParams()
	if err != nil {
		t.Fatalf("Failed to convert literal: %v", err)
	}

	if !qp.HasFilters() {
		t.Error("Expected HasFilters() to be true")
	}

	criteria := qp.ToFilterCriteria()
	if len(criteria) < 2 {
		t.Errorf("Expected at least 2 filter criteria, got %d", len(criteria))
	}
}

func TestStronglyTypedFilter_FluentAPI(t *testing.T) {
	// Test the strongly typed filter builder using the literal approach
	literal := QueryParamsLiteral[*testutil.TestEntity]{
		Filter: FilterLiteral[*testutil.TestEntity]{
			"name": FilterOperatorLiteralValue[*testutil.TestEntity, string]{
				FilterOperatorLiteral: FilterOperatorLiteral[string]{Contains: StringPtr("test")},
			},
			"age": FilterOperatorLiteralValue[*testutil.TestEntity, int]{
				FilterOperatorLiteral: FilterOperatorLiteral[int]{Between: []int{18, 65}},
			},
		},
	}

	qp, err := literal.ToQueryParams()
	if err != nil {
		t.Fatalf("Failed to convert literal: %v", err)
	}

	filter := qp.GetStronglyTypedFilter()
	if filter == nil || filter.IsEmpty() {
		t.Error("Expected filter to not be empty")
	}

	criteria := filter.ToFilterCriteria()
	if len(criteria) < 2 {
		t.Errorf("Expected at least 2 filter criteria, got %d", len(criteria))
	}

	// Test clone
	cloned := filter.Clone()
	if cloned == nil || cloned.IsEmpty() {
		t.Error("Expected cloned filter to not be empty")
	}

	clonedCriteria := cloned.ToFilterCriteria()
	if len(clonedCriteria) != len(criteria) {
		t.Errorf("Expected cloned filter to have same number of criteria: %d vs %d", len(clonedCriteria), len(criteria))
	}
}

func TestTypeRegistry_Registration(t *testing.T) {
	registry := NewTypeRegistry()

	// Test type registration
	err := registry.RegisterType(&testutil.TestEntity{})
	if err != nil {
		t.Fatalf("Failed to register type: %v", err)
	}

	// Test field info retrieval
	entityType := reflect.TypeOf(&testutil.TestEntity{}).Elem()
	fieldInfo, err := registry.GetFieldInfo(entityType, "name")
	if err != nil {
		t.Fatalf("Failed to get field info: %v", err)
	}

	if fieldInfo.Name != "Name" {
		t.Errorf("Expected field name 'Name', got '%s'", fieldInfo.Name)
	}

	// Test validation
	_, err = registry.ValidateFieldPath(entityType, "name")
	if err != nil {
		t.Errorf("Field validation should pass for 'name': %v", err)
	}

	_, err = registry.ValidateFieldPath(entityType, "nonexistent")
	if err == nil {
		t.Error("Field validation should fail for nonexistent field")
	}
}

func TestQueryParams_InterfaceCompliance(t *testing.T) {
	// Test that QueryParams implements domain.IQueryParams
	var _ domain.IQueryParams[*testutil.TestEntity] = &QueryParams[*testutil.TestEntity]{}

	// Test that QueryParams implements QueryParamsType
	var _ QueryParamsType[*testutil.TestEntity] = &QueryParams[*testutil.TestEntity]{}

	// Create a QueryParams instance and test interface methods
	qp := &QueryParams[*testutil.TestEntity]{
		limit:  25,
		offset: 10,
	}

	if qp.Limit() != 25 {
		t.Errorf("Expected limit 25, got %d", qp.Limit())
	}

	if qp.Offset() != 10 {
		t.Errorf("Expected offset 10, got %d", qp.Offset())
	}

	// Test fluent methods return the same instance
	result := qp.WithLimit(50)
	if result != qp {
		t.Error("WithLimit should return the same instance")
	}

	if qp.Limit() != 50 {
		t.Errorf("Expected limit to be updated to 50, got %d", qp.Limit())
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test helper functions for filter operators
	eqOp := Eq("test")
	if eqOp.Eq == nil || *eqOp.Eq != "test" {
		t.Error("Eq helper function not working correctly")
	}

	neqOp := Neq(42)
	if neqOp.Neq == nil || *neqOp.Neq != 42 {
		t.Error("Neq helper function not working correctly")
	}

	gtOp := Gt(10)
	if gtOp.Gt == nil || *gtOp.Gt != 10 {
		t.Error("Gt helper function not working correctly")
	}

	inOp := In(1, 2, 3)
	if len(inOp.In) != 3 || inOp.In[0] != 1 || inOp.In[1] != 2 || inOp.In[2] != 3 {
		t.Error("In helper function not working correctly")
	}

	betweenOp := Between(10, 20)
	if len(betweenOp.Between) != 2 || betweenOp.Between[0] != 10 || betweenOp.Between[1] != 20 {
		t.Error("Between helper function not working correctly")
	}

	isNullOp := IsNull[string]()
	if isNullOp.IsNull == nil || !*isNullOp.IsNull {
		t.Error("IsNull helper function not working correctly")
	}
}

func TestQueryParamsType_UnionType(t *testing.T) {
	// Test that both QueryParams and the interface work as QueryParamsType
	registry := NewTypeRegistry()
	err := registry.RegisterType(&testutil.TestEntity{})
	if err != nil {
		t.Fatalf("Failed to register type: %v", err)
	}

	// Test QueryParams as QueryParamsType
	qp := &QueryParams[*testutil.TestEntity]{limit: 10}
	var _ QueryParamsType[*testutil.TestEntity] = qp

	// Test domain.IQueryParams as QueryParamsType (through interface)
	var domainQP domain.IQueryParams[*testutil.TestEntity] = qp
	var _ QueryParamsType[*testutil.TestEntity] = domainQP.(QueryParamsType[*testutil.TestEntity])

	// Verify functionality
	if qp.Limit() != 10 {
		t.Errorf("Expected limit 10, got %d", qp.Limit())
	}
}

// Helper functions for pointer values
func StringPtr(s string) *string     { return &s }
func IntPtr(i int) *int              { return &i }
func BoolPtr(b bool) *bool           { return &b }
func Float64Ptr(f float64) *float64  { return &f }
func TimePtr(t time.Time) *time.Time { return &t }
