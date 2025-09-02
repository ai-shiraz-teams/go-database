package identifier

type FilterCriteria struct {
	Field string `json:"field"`

	Operator FilterOperator `json:"operator"`

	Value interface{} `json:"value,omitempty"`

	Values []interface{} `json:"values,omitempty"`

	LogicalOp LogicalOperator `json:"logicalOp,omitempty"`

	Group []FilterCriteria `json:"group,omitempty"`
}
