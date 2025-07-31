package identifier

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
