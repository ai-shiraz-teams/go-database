package unit_of_work

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoFilterApplier handles the conversion of query parameters and identifiers to MongoDB filters
type MongoFilterApplier struct{}

// NewMongoFilterApplier creates a new MongoDB filter applier
func NewMongoFilterApplier() *MongoFilterApplier {
	return &MongoFilterApplier{}
}

// ApplyQueryParams converts QueryParams to MongoDB filter
func (mfa *MongoFilterApplier) ApplyQueryParams(baseFilter bson.M, queryParams interface{}) bson.M {
	if queryParams == nil {
		return baseFilter
	}

	v := reflect.ValueOf(queryParams)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if filtersField := v.FieldByName("Filters"); filtersField.IsValid() && filtersField.CanInterface() {
		if filters, ok := filtersField.Interface().([]identifier.FilterCriteria); ok {
			for _, filter := range filters {
				mfa.applyFilterCriteria(baseFilter, filter)
			}
		}
	}

	if filterField := v.FieldByName("Filter"); filterField.IsValid() && filterField.CanInterface() && !filterField.IsNil() {
		filterDoc, err := mfa.StructToBSONFilter(filterField.Interface())
		if err == nil {
			for key, value := range filterDoc {
				baseFilter[key] = value
			}
		}
	}

	return baseFilter
}

// BuildFilterFromIdentifier converts an IIdentifier to MongoDB filter
func (mfa *MongoFilterApplier) BuildFilterFromIdentifier(identifier identifier.IIdentifier) bson.M {
	filter := bson.M{}
	criteria := identifier.ToFilterCriteria()

	for _, criterion := range criteria {
		mfa.applyFilterCriteria(filter, criterion)
	}

	return filter
}

// applyFilterCriteria applies a single filter criterion to the MongoDB filter
func (mfa *MongoFilterApplier) applyFilterCriteria(filter bson.M, criterion identifier.FilterCriteria) {
	field := criterion.Field
	value := criterion.Value
	operator := criterion.Operator

	switch operator {
	case identifier.FilterOperatorEqual:
		filter[field] = value
	case identifier.FilterOperatorNotEqual:
		filter[field] = bson.M{"$ne": value}
	case identifier.FilterOperatorGreaterThan:
		filter[field] = bson.M{"$gt": value}
	case identifier.FilterOperatorGreaterEqual:
		filter[field] = bson.M{"$gte": value}
	case identifier.FilterOperatorLessThan:
		filter[field] = bson.M{"$lt": value}
	case identifier.FilterOperatorLessEqual:
		filter[field] = bson.M{"$lte": value}
	case identifier.FilterOperatorLike:
		pattern := mfa.convertLikeToRegex(value.(string))
		filter[field] = bson.M{"$regex": pattern, "$options": "i"}
	case identifier.FilterOperatorIn:
		filter[field] = bson.M{"$in": value}
	case identifier.FilterOperatorNotIn:
		filter[field] = bson.M{"$nin": value}
	case identifier.FilterOperatorBetween:
		if values, ok := value.([]interface{}); ok && len(values) == 2 {
			filter[field] = bson.M{"$gte": values[0], "$lte": values[1]}
		}
	case identifier.FilterOperatorIsNull:
		filter[field] = bson.M{"$exists": false}
	case identifier.FilterOperatorIsNotNull:
		filter[field] = bson.M{"$exists": true, "$ne": nil}
	case identifier.FilterOperatorContains:
		filter[field] = bson.M{"$elemMatch": bson.M{"$eq": value}}
	case identifier.FilterOperatorHas:
		filter[field] = bson.M{"$exists": true, "$ne": nil}
	}
}

// convertLikeToRegex converts SQL LIKE patterns to MongoDB regex
func (mfa *MongoFilterApplier) convertLikeToRegex(likePattern string) string {
	escaped := strings.ReplaceAll(likePattern, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, ".", "\\.")
	escaped = strings.ReplaceAll(escaped, "^", "\\^")
	escaped = strings.ReplaceAll(escaped, "$", "\\$")
	escaped = strings.ReplaceAll(escaped, "*", "\\*")
	escaped = strings.ReplaceAll(escaped, "+", "\\+")
	escaped = strings.ReplaceAll(escaped, "?", "\\?")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	escaped = strings.ReplaceAll(escaped, "[", "\\[")
	escaped = strings.ReplaceAll(escaped, "]", "\\]")
	escaped = strings.ReplaceAll(escaped, "{", "\\{")
	escaped = strings.ReplaceAll(escaped, "}", "\\}")
	escaped = strings.ReplaceAll(escaped, "|", "\\|")

	escaped = strings.ReplaceAll(escaped, "%", ".*")
	escaped = strings.ReplaceAll(escaped, "_", ".")

	return "^" + escaped + "$"
}

// BuildSortDocument converts sort parameters to MongoDB sort document
func (mfa *MongoFilterApplier) BuildSortDocument(sortParams []query.SortField) bson.D {
	sort := bson.D{}

	for _, param := range sortParams {
		direction := 1
		if string(param.Order) == "desc" {
			direction = -1
		}
		sort = append(sort, bson.E{Key: param.Field, Value: direction})
	}

	return sort
}

// StructToBSONFilter converts a struct to a MongoDB filter document
func (mfa *MongoFilterApplier) StructToBSONFilter(entity interface{}) (bson.M, error) {
	filter := bson.M{}

	if entity == nil {
		return filter, nil
	}

	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return filter, nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return filter, fmt.Errorf("entity must be a struct or pointer to struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldName := mfa.getFieldName(fieldType)

		if mfa.isZeroValue(field) {
			continue
		}

		filter[fieldName] = field.Interface()
	}

	return filter, nil
}

// StructToBSONUpdate converts a struct to a MongoDB update document
func (mfa *MongoFilterApplier) StructToBSONUpdate(entity interface{}) (bson.M, error) {
	update := bson.M{}

	if entity == nil {
		return update, nil
	}

	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return update, nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return update, fmt.Errorf("entity must be a struct or pointer to struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldName := mfa.getFieldName(fieldType)

		if fieldName == "_id" || fieldName == "id" {
			continue
		}

		update[fieldName] = field.Interface()
	}

	return update, nil
}

// getFieldName extracts the MongoDB field name from struct field tags
func (mfa *MongoFilterApplier) getFieldName(field reflect.StructField) string {

	if bsonTag := field.Tag.Get("bson"); bsonTag != "" {
		if bsonTag == "-" {
			return ""
		}

		if commaIdx := strings.Index(bsonTag, ","); commaIdx != -1 {
			return bsonTag[:commaIdx]
		}
		return bsonTag
	}

	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		if jsonTag == "-" {
			return ""
		}

		if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
			return jsonTag[:commaIdx]
		}
		return jsonTag
	}

	return mfa.toSnakeCase(field.Name)
}

// toSnakeCase converts PascalCase to snake_case
func (mfa *MongoFilterApplier) toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// isZeroValue checks if a value is the zero value for its type
func (mfa *MongoFilterApplier) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:

		if v.Type().String() == "time.Time" {
			return v.Interface().(interface{ IsZero() bool }).IsZero()
		}

		return false
	}
	return false
}

// ValidateFilterValue validates that a filter value is supported by MongoDB
func (mfa *MongoFilterApplier) ValidateFilterValue(fieldName string, value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return nil
	case primitive.ObjectID, primitive.DateTime:
		return nil
	case []interface{}:
		for _, elem := range v {
			if err := mfa.ValidateFilterValue(fieldName, elem); err != nil {
				return err
			}
		}
		return nil
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return nil
		}
		return fmt.Errorf("unsupported filter value type for field %s: %T", fieldName, value)
	}
}
