package domain

// FilterOperator defines the type of comparison operation for filtering
type FilterOperator string

const (
	// Comparison operators
	FilterOperatorEqual        FilterOperator = "eq"
	FilterOperatorNotEqual     FilterOperator = "neq"
	FilterOperatorGreaterThan  FilterOperator = "gt"
	FilterOperatorGreaterEqual FilterOperator = "gte"
	FilterOperatorLessThan     FilterOperator = "lt"
	FilterOperatorLessEqual    FilterOperator = "lte"
	FilterOperatorLike         FilterOperator = "like"
	FilterOperatorIn           FilterOperator = "in"
	FilterOperatorNotIn        FilterOperator = "not_in"
	FilterOperatorIsNull       FilterOperator = "is_null"
	FilterOperatorIsNotNull    FilterOperator = "is_not_null"
	FilterOperatorBetween      FilterOperator = "between"
	FilterOperatorContains     FilterOperator = "contains"
	FilterOperatorHas          FilterOperator = "has"
)

// LogicalOperator defines how multiple filter criteria are combined
type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "and"
	LogicalOperatorOr  LogicalOperator = "or"
)

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

// IIdentifier defines the contract for building dynamic filter queries.
// It provides a fluent API for constructing complex filter conditions
// that can be converted to FilterCriteria for use with repositories.
type IIdentifier interface {
	// Basic comparison operators
	Equal(field string, value interface{}) IIdentifier
	NotEqual(field string, value interface{}) IIdentifier
	GreaterThan(field string, value interface{}) IIdentifier
	GreaterOrEqual(field string, value interface{}) IIdentifier
	LessThan(field string, value interface{}) IIdentifier
	LessOrEqual(field string, value interface{}) IIdentifier

	// String and pattern matching
	Like(field string, value string) IIdentifier

	// Collection operators
	In(field string, values []interface{}) IIdentifier
	NotIn(field string, values []interface{}) IIdentifier

	// Null checks
	IsNull(field string) IIdentifier
	IsNotNull(field string) IIdentifier

	// Range operators
	Between(field string, start, end interface{}) IIdentifier

	// JSON and array operators (for advanced databases)
	Contains(field string, value interface{}) IIdentifier
	Has(field string) IIdentifier

	// Logical operators for combining conditions
	And(identifier IIdentifier) IIdentifier
	Or(identifier IIdentifier) IIdentifier

	// Output conversion
	ToFilterCriteria() []FilterCriteria

	// Reset builder state
	Reset() IIdentifier
}
