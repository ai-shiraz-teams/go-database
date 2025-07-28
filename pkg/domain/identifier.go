package domain

import "sync"

// IdentifierBuilder provides a concrete implementation of IIdentifier interface.
// It builds filter criteria in a fluent, chainable manner while maintaining immutability.
// Each operation returns a new instance to ensure thread safety and prevent side effects.
type IdentifierBuilder struct {
	criteria []FilterCriteria
	mutex    sync.RWMutex // Ensures thread safety for read operations
}

// NewIdentifier creates a new empty IdentifierBuilder instance
func NewIdentifier() IIdentifier {
	return &IdentifierBuilder{
		criteria: make([]FilterCriteria, 0),
	}
}

// clone creates a deep copy of the current builder state to maintain immutability
func (ib *IdentifierBuilder) clone() *IdentifierBuilder {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	newBuilder := &IdentifierBuilder{
		criteria: make([]FilterCriteria, len(ib.criteria)),
	}
	copy(newBuilder.criteria, ib.criteria)
	return newBuilder
}

// addCriteria adds a new filter criteria and returns a new builder instance
func (ib *IdentifierBuilder) addCriteria(criteria FilterCriteria) IIdentifier {
	newBuilder := ib.clone()
	newBuilder.criteria = append(newBuilder.criteria, criteria)
	return newBuilder
}

// Equal adds an equality filter condition
func (ib *IdentifierBuilder) Equal(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorEqual,
		Value:    value,
	})
}

// NotEqual adds a non-equality filter condition
func (ib *IdentifierBuilder) NotEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorNotEqual,
		Value:    value,
	})
}

// GreaterThan adds a greater-than filter condition
func (ib *IdentifierBuilder) GreaterThan(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorGreaterThan,
		Value:    value,
	})
}

// GreaterOrEqual adds a greater-than-or-equal filter condition
func (ib *IdentifierBuilder) GreaterOrEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorGreaterEqual,
		Value:    value,
	})
}

// LessThan adds a less-than filter condition
func (ib *IdentifierBuilder) LessThan(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLessThan,
		Value:    value,
	})
}

// LessOrEqual adds a less-than-or-equal filter condition
func (ib *IdentifierBuilder) LessOrEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLessEqual,
		Value:    value,
	})
}

// Like adds a pattern matching filter condition (SQL LIKE operator)
func (ib *IdentifierBuilder) Like(field string, value string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLike,
		Value:    value,
	})
}

// In adds a filter condition that checks if field value is in the provided list
func (ib *IdentifierBuilder) In(field string, values []interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIn,
		Values:   values,
	})
}

// NotIn adds a filter condition that checks if field value is not in the provided list
func (ib *IdentifierBuilder) NotIn(field string, values []interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorNotIn,
		Values:   values,
	})
}

// IsNull adds a filter condition that checks if field value is NULL
func (ib *IdentifierBuilder) IsNull(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIsNull,
	})
}

// IsNotNull adds a filter condition that checks if field value is not NULL
func (ib *IdentifierBuilder) IsNotNull(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIsNotNull,
	})
}

// Between adds a range filter condition that checks if field value is between start and end
func (ib *IdentifierBuilder) Between(field string, start, end interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorBetween,
		Values:   []interface{}{start, end},
	})
}

// Contains adds a filter condition for JSON/array field containment
func (ib *IdentifierBuilder) Contains(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorContains,
		Value:    value,
	})
}

// Has adds a filter condition that checks for field existence (useful for JSON fields)
func (ib *IdentifierBuilder) Has(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorHas,
	})
}

// And combines the current builder with another identifier using AND logic
func (ib *IdentifierBuilder) And(identifier IIdentifier) IIdentifier {
	if identifier == nil {
		return ib
	}

	otherCriteria := identifier.ToFilterCriteria()
	if len(otherCriteria) == 0 {
		return ib
	}

	newBuilder := ib.clone()

	// If we have existing criteria, mark the last one as AND
	if len(newBuilder.criteria) > 0 {
		lastIndex := len(newBuilder.criteria) - 1
		newBuilder.criteria[lastIndex].LogicalOp = LogicalOperatorAnd
	}

	// Add the other criteria
	newBuilder.criteria = append(newBuilder.criteria, otherCriteria...)

	return newBuilder
}

// Or combines the current builder with another identifier using OR logic
func (ib *IdentifierBuilder) Or(identifier IIdentifier) IIdentifier {
	if identifier == nil {
		return ib
	}

	otherCriteria := identifier.ToFilterCriteria()
	if len(otherCriteria) == 0 {
		return ib
	}

	newBuilder := ib.clone()

	// If we have existing criteria, mark the last one as OR
	if len(newBuilder.criteria) > 0 {
		lastIndex := len(newBuilder.criteria) - 1
		newBuilder.criteria[lastIndex].LogicalOp = LogicalOperatorOr
	}

	// Add the other criteria
	newBuilder.criteria = append(newBuilder.criteria, otherCriteria...)

	return newBuilder
}

// ToFilterCriteria returns the accumulated filter criteria as a slice
func (ib *IdentifierBuilder) ToFilterCriteria() []FilterCriteria {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	if len(ib.criteria) == 0 {
		return nil
	}

	// Create a deep copy to prevent external modification
	result := make([]FilterCriteria, len(ib.criteria))
	copy(result, ib.criteria)
	return result
}

// Reset clears all filter criteria and returns a fresh builder
func (ib *IdentifierBuilder) Reset() IIdentifier {
	return &IdentifierBuilder{
		criteria: make([]FilterCriteria, 0),
	}
}

// Compile-time check to ensure IdentifierBuilder implements IIdentifier
var _ IIdentifier = (*IdentifierBuilder)(nil)
