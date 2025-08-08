package identifier

// IIdentifier defines the contract for building dynamic filter queries.
type IIdentifier interface {
	Equal(field string, value interface{}) IIdentifier
	NotEqual(field string, value interface{}) IIdentifier
	GreaterThan(field string, value interface{}) IIdentifier
	GreaterOrEqual(field string, value interface{}) IIdentifier
	LessThan(field string, value interface{}) IIdentifier
	LessOrEqual(field string, value interface{}) IIdentifier

	Like(field string, pattern string) IIdentifier
	In(field string, values []interface{}) IIdentifier
	NotIn(field string, values []interface{}) IIdentifier
	Between(field string, start, end interface{}) IIdentifier

	IsNull(field string) IIdentifier
	IsNotNull(field string) IIdentifier

	Contains(field string, value interface{}) IIdentifier
	Has(field string) IIdentifier

	And(other IIdentifier) IIdentifier
	Or(other IIdentifier) IIdentifier

	ToFilterCriteria() []FilterCriteria
	Reset() IIdentifier
}
