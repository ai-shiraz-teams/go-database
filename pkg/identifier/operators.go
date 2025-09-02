package identifier

type FilterOperator string

const (
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

type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "and"
	LogicalOperatorOr  LogicalOperator = "or"
)
