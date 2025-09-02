package identifier

import "sync"

type IdentifierBuilder struct {
	criteria []FilterCriteria
	mutex    sync.RWMutex
}

func NewIdentifier() IIdentifier {
	return &IdentifierBuilder{
		criteria: make([]FilterCriteria, 0),
	}
}

func (ib *IdentifierBuilder) clone() *IdentifierBuilder {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	newCriteria := make([]FilterCriteria, len(ib.criteria))
	copy(newCriteria, ib.criteria)

	return &IdentifierBuilder{
		criteria: newCriteria,
	}
}

func (ib *IdentifierBuilder) addCriteria(criteria FilterCriteria) IIdentifier {
	newBuilder := ib.clone()
	newBuilder.criteria = append(newBuilder.criteria, criteria)
	return newBuilder
}

func (ib *IdentifierBuilder) Equal(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorEqual,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) NotEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorNotEqual,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) GreaterThan(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorGreaterThan,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) GreaterOrEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorGreaterEqual,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) LessThan(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLessThan,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) LessOrEqual(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLessEqual,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) Like(field string, pattern string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorLike,
		Value:    pattern,
	})
}

func (ib *IdentifierBuilder) In(field string, values []interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIn,
		Values:   values,
	})
}

func (ib *IdentifierBuilder) NotIn(field string, values []interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorNotIn,
		Values:   values,
	})
}

func (ib *IdentifierBuilder) IsNull(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIsNull,
	})
}

func (ib *IdentifierBuilder) IsNotNull(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorIsNotNull,
	})
}

func (ib *IdentifierBuilder) Between(field string, start, end interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorBetween,
		Values:   []interface{}{start, end},
	})
}

func (ib *IdentifierBuilder) Contains(field string, value interface{}) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorContains,
		Value:    value,
	})
}

func (ib *IdentifierBuilder) Has(field string) IIdentifier {
	return ib.addCriteria(FilterCriteria{
		Field:    field,
		Operator: FilterOperatorHas,
	})
}

func (ib *IdentifierBuilder) And(other IIdentifier) IIdentifier {
	if other == nil {
		return ib
	}

	newBuilder := ib.clone()
	otherCriteria := other.ToFilterCriteria()

	if len(otherCriteria) > 0 && len(newBuilder.criteria) > 0 {

		if len(newBuilder.criteria) > 0 {
			newBuilder.criteria[len(newBuilder.criteria)-1].LogicalOp = LogicalOperatorAnd
		}
	}

	newBuilder.criteria = append(newBuilder.criteria, otherCriteria...)
	return newBuilder
}

func (ib *IdentifierBuilder) Or(other IIdentifier) IIdentifier {
	if other == nil {
		return ib
	}

	newBuilder := ib.clone()
	otherCriteria := other.ToFilterCriteria()

	if len(otherCriteria) > 0 && len(newBuilder.criteria) > 0 {

		if len(newBuilder.criteria) > 0 {
			newBuilder.criteria[len(newBuilder.criteria)-1].LogicalOp = LogicalOperatorOr
		}
	}

	newBuilder.criteria = append(newBuilder.criteria, otherCriteria...)
	return newBuilder
}

func (ib *IdentifierBuilder) ToFilterCriteria() []FilterCriteria {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	if len(ib.criteria) == 0 {
		return nil
	}

	result := make([]FilterCriteria, len(ib.criteria))
	copy(result, ib.criteria)
	return result
}

func (ib *IdentifierBuilder) Reset() IIdentifier {
	return NewIdentifier()
}

var _ IIdentifier = (*IdentifierBuilder)(nil)
