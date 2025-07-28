package query

// SortField represents a single field to sort by with its direction
type SortField struct {
	// Field is the name of the field to sort by
	Field string `json:"field"`
	// Order is the direction to sort (asc/desc)
	Order SortOrder `json:"order"`
}
