package identifier

// FilterCriteria represents a single filter condition that can be applied to a query.
// It's designed to be ORM-agnostic and can be converted to various query formats.
type FilterCriteria struct {
	// Field is the name of the database column or entity field to filter on
	Field string `json:"field"`

	// Operator defines the type of comparison to perform
	Operator FilterOperator `json:"operator"`

	// Value is the value to compare against (can be nil for null checks)
	Value interface{} `json:"value,omitempty"`

	// Values is used for operators that require multiple values (IN, NOT_IN, BETWEEN)
	Values []interface{} `json:"values,omitempty"`

	// LogicalOp defines how this criteria combines with the next one (AND/OR)
	// This is used when multiple criteria are present in a list
	LogicalOp LogicalOperator `json:"logicalOp,omitempty"`

	// Group allows nesting of filter criteria for complex conditions
	// When Group is not empty, Field/Operator/Value are ignored
	Group []FilterCriteria `json:"group,omitempty"`
}
