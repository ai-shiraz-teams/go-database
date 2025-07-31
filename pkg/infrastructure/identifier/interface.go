package identifier

// IIdentifier defines the contract for building dynamic filter queries.
// It provides a fluent API for constructing complex filter conditions
// that can be converted to FilterCriteria for use with repositories.
type IIdentifier interface {
	// Basic comparison operations
	Equal(field string, value interface{}) IIdentifier
	NotEqual(field string, value interface{}) IIdentifier
	GreaterThan(field string, value interface{}) IIdentifier
	GreaterOrEqual(field string, value interface{}) IIdentifier
	LessThan(field string, value interface{}) IIdentifier
	LessOrEqual(field string, value interface{}) IIdentifier

	// Pattern matching and collection operations
	Like(field string, pattern string) IIdentifier
	In(field string, values []interface{}) IIdentifier
	NotIn(field string, values []interface{}) IIdentifier
	Between(field string, start, end interface{}) IIdentifier

	// Null checks
	IsNull(field string) IIdentifier
	IsNotNull(field string) IIdentifier

	// JSON and advanced operations
	Contains(field string, value interface{}) IIdentifier
	Has(field string) IIdentifier

	// Logical operations for combining identifiers
	And(other IIdentifier) IIdentifier
	Or(other IIdentifier) IIdentifier

	// Conversion and utility methods
	ToFilterCriteria() []FilterCriteria
	Reset() IIdentifier
}
