package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

// fieldFilterOperator implements FieldFilterOperator interface
type fieldFilterOperator[T domain.IBaseModel] struct {
	fieldPath string
	fieldType TypeInfo
}

func (f *fieldFilterOperator[T]) GetFieldPath() string {
	return f.fieldPath
}

func (f *fieldFilterOperator[T]) GetFieldType() TypeInfo {
	return f.fieldType
}

func (f *fieldFilterOperator[T]) AsString() (StringFilterOperator[T], bool) {
	if f.fieldType.Kind == reflect.String {
		return nil, true // TODO: implement actual string operator
	}
	return nil, false
}

func (f *fieldFilterOperator[T]) AsComparableInt() (ComparableFilterOperator[T, int], bool) {
	switch f.fieldType.Kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil, true // TODO: implement actual int operator
	}
	return nil, false
}

func (f *fieldFilterOperator[T]) AsComparableFloat() (ComparableFilterOperator[T, float64], bool) {
	switch f.fieldType.Kind {
	case reflect.Float32, reflect.Float64:
		return nil, true // TODO: implement actual float operator
	}
	return nil, false
}

func (f *fieldFilterOperator[T]) AsBoolean() (BooleanFilterOperator[T], bool) {
	if f.fieldType.Kind == reflect.Bool {
		return nil, true // TODO: implement actual boolean operator
	}
	return nil, false
}

func (f *fieldFilterOperator[T]) AsNullableString() (NullableFilterOperator[T, string], bool) {
	if f.fieldType.IsPointer && f.fieldType.ElementType != nil && f.fieldType.ElementType.Kind() == reflect.String {
		return nil, true // TODO: implement actual nullable string operator
	}
	return nil, false
}

type fieldFilterBuilder[T domain.IBaseModel] struct {
	registry  TypeRegistry
	fieldName string
	fieldType reflect.Type
	fieldPath string
	criteria  []domain.FilterCriteria
	builder   *stronglyTypedFilterBuilder[T]
}

func (f *fieldFilterBuilder[T]) GetFieldPath() string {
	return f.fieldPath
}

func (f *fieldFilterBuilder[T]) GetFilterCriteria() []domain.FilterCriteria {
	return f.criteria
}

func (f *fieldFilterBuilder[T]) And() StronglyTypedFilterBuilder[T] {
	if len(f.criteria) > 0 {
		lastIdx := len(f.criteria) - 1
		f.criteria[lastIdx].LogicalOp = domain.LogicalOperatorAnd
	}
	f.builder.criteria = append(f.builder.criteria, f.criteria...)
	return f.builder
}

func (f *fieldFilterBuilder[T]) Or() StronglyTypedFilterBuilder[T] {
	if len(f.criteria) > 0 {
		lastIdx := len(f.criteria) - 1
		f.criteria[lastIdx].LogicalOp = domain.LogicalOperatorOr
	}
	f.builder.criteria = append(f.builder.criteria, f.criteria...)
	return f.builder
}

// addCriteria adds a filter criteria to this field filter
func (f *fieldFilterBuilder[T]) addCriteria(operator domain.FilterOperator, value interface{}, values []interface{}) *fieldFilterBuilder[T] {
	criteria := domain.FilterCriteria{
		Field:    f.fieldPath,
		Operator: operator,
		Value:    value,
		Values:   values,
	}
	f.criteria = append(f.criteria, criteria)
	return f
}

// stringFilterOperator provides string-specific filter operations
type stringFilterOperator[T domain.IBaseModel] struct {
	*fieldFilterBuilder[T]
}

func (s *stringFilterOperator[T]) Equal(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorEqual, value, nil)
}

func (s *stringFilterOperator[T]) NotEqual(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorNotEqual, value, nil)
}

func (s *stringFilterOperator[T]) GreaterThan(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorGreaterThan, value, nil)
}

func (s *stringFilterOperator[T]) GreaterOrEqual(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorGreaterEqual, value, nil)
}

func (s *stringFilterOperator[T]) LessThan(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLessThan, value, nil)
}

func (s *stringFilterOperator[T]) LessOrEqual(value string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLessEqual, value, nil)
}

func (s *stringFilterOperator[T]) Between(start, end string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorBetween, nil, []interface{}{start, end})
}

func (s *stringFilterOperator[T]) Like(pattern string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLike, pattern, nil)
}

func (s *stringFilterOperator[T]) Contains(substring string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLike, "%"+substring+"%", nil)
}

func (s *stringFilterOperator[T]) StartsWith(prefix string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLike, prefix+"%", nil)
}

func (s *stringFilterOperator[T]) EndsWith(suffix string) FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorLike, "%"+suffix, nil)
}

func (s *stringFilterOperator[T]) In(values []string) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return s.addCriteria(domain.FilterOperatorIn, nil, interfaces)
}

func (s *stringFilterOperator[T]) NotIn(values []string) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return s.addCriteria(domain.FilterOperatorNotIn, nil, interfaces)
}

func (s *stringFilterOperator[T]) IsNull() FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorIsNull, nil, nil)
}

func (s *stringFilterOperator[T]) IsNotNull() FieldFilterBuilder[T] {
	return s.addCriteria(domain.FilterOperatorIsNotNull, nil, nil)
}

// intFilterOperator implements ComparableFilterOperator for int
type intFilterOperator[T domain.IBaseModel] struct {
	*fieldFilterBuilder[T]
}

func (i *intFilterOperator[T]) Equal(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorEqual, value, nil)
}

func (i *intFilterOperator[T]) NotEqual(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorNotEqual, value, nil)
}

func (i *intFilterOperator[T]) GreaterThan(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorGreaterThan, value, nil)
}

func (i *intFilterOperator[T]) GreaterOrEqual(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorGreaterEqual, value, nil)
}

func (i *intFilterOperator[T]) LessThan(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorLessThan, value, nil)
}

func (i *intFilterOperator[T]) LessOrEqual(value int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorLessEqual, value, nil)
}

func (i *intFilterOperator[T]) Between(start, end int) FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorBetween, nil, []interface{}{start, end})
}

func (i *intFilterOperator[T]) In(values []int) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for idx, v := range values {
		interfaces[idx] = v
	}
	return i.addCriteria(domain.FilterOperatorIn, nil, interfaces)
}

func (i *intFilterOperator[T]) NotIn(values []int) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for idx, v := range values {
		interfaces[idx] = v
	}
	return i.addCriteria(domain.FilterOperatorNotIn, nil, interfaces)
}

func (i *intFilterOperator[T]) IsNull() FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorIsNull, nil, nil)
}

func (i *intFilterOperator[T]) IsNotNull() FieldFilterBuilder[T] {
	return i.addCriteria(domain.FilterOperatorIsNotNull, nil, nil)
}

// booleanFilterOperator implements BooleanFilterOperator
type booleanFilterOperator[T domain.IBaseModel] struct {
	*fieldFilterBuilder[T]
}

func (b *booleanFilterOperator[T]) Equal(value bool) FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorEqual, value, nil)
}

func (b *booleanFilterOperator[T]) NotEqual(value bool) FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorNotEqual, value, nil)
}

func (b *booleanFilterOperator[T]) IsTrue() FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorEqual, true, nil)
}

func (b *booleanFilterOperator[T]) IsFalse() FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorEqual, false, nil)
}

func (b *booleanFilterOperator[T]) In(values []bool) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return b.addCriteria(domain.FilterOperatorIn, nil, interfaces)
}

func (b *booleanFilterOperator[T]) NotIn(values []bool) FieldFilterBuilder[T] {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return b.addCriteria(domain.FilterOperatorNotIn, nil, interfaces)
}

func (b *booleanFilterOperator[T]) IsNull() FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorIsNull, nil, nil)
}

func (b *booleanFilterOperator[T]) IsNotNull() FieldFilterBuilder[T] {
	return b.addCriteria(domain.FilterOperatorIsNotNull, nil, nil)
}

// stronglyTypedFilterBuilder implements StronglyTypedFilterBuilder interface
type stronglyTypedFilterBuilder[T domain.IBaseModel] struct {
	criteria     []domain.FilterCriteria
	typeInfo     reflect.Type
	fieldCache   map[string]*FieldInfo
	typeRegistry TypeRegistry
}

func (s *stronglyTypedFilterBuilder[T]) GetField(fieldName string) FieldFilterOperator[T] {
	fieldInfo, err := s.typeRegistry.ValidateFieldPath(s.typeInfo, fieldName)
	if err != nil {
		// Return a generic filter that will fail validation
		return &fieldFilterOperator[T]{
			fieldPath: fieldName,
			fieldType: TypeInfo{},
		}
	}

	return &fieldFilterOperator[T]{
		fieldPath: fieldName,
		fieldType: fieldInfo.TypeInfo,
	}
}

func (s *stronglyTypedFilterBuilder[T]) And(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T] {
	if otherBuilder, ok := other.(*stronglyTypedFilterBuilder[T]); ok {
		// Set AND operator on the last criteria of current builder
		if len(s.criteria) > 0 {
			lastIdx := len(s.criteria) - 1
			s.criteria[lastIdx].LogicalOp = domain.LogicalOperatorAnd
		}
		s.criteria = append(s.criteria, otherBuilder.criteria...)
	}
	return s
}

func (s *stronglyTypedFilterBuilder[T]) Or(other StronglyTypedFilterBuilder[T]) StronglyTypedFilterBuilder[T] {
	if otherBuilder, ok := other.(*stronglyTypedFilterBuilder[T]); ok {
		// Set OR operator on the last criteria of current builder
		if len(s.criteria) > 0 {
			lastIdx := len(s.criteria) - 1
			s.criteria[lastIdx].LogicalOp = domain.LogicalOperatorOr
		}
		s.criteria = append(s.criteria, otherBuilder.criteria...)
	}
	return s
}

func (s *stronglyTypedFilterBuilder[T]) ToFilterCriteria() []domain.FilterCriteria {
	return s.criteria
}

func (s *stronglyTypedFilterBuilder[T]) Build() StronglyTypedFilter[T] {
	return &stronglyTypedFilter[T]{
		criteria: append([]domain.FilterCriteria{}, s.criteria...),
	}
}

// stronglyTypedFilter implements StronglyTypedFilter interface
type stronglyTypedFilter[T domain.IBaseModel] struct {
	criteria []domain.FilterCriteria
}

func (s *stronglyTypedFilter[T]) ToFilterCriteria() []domain.FilterCriteria {
	return s.criteria
}

func (s *stronglyTypedFilter[T]) Clone() StronglyTypedFilter[T] {
	cloned := make([]domain.FilterCriteria, len(s.criteria))
	copy(cloned, s.criteria)
	return &stronglyTypedFilter[T]{criteria: cloned}
}

func (s *stronglyTypedFilter[T]) IsEmpty() bool {
	return len(s.criteria) == 0
}

// sortBuilder implements SortBuilder interface
type sortBuilder[T domain.IBaseModel] struct {
	sorts        []domain.SortField
	typeInfo     reflect.Type
	typeRegistry TypeRegistry
}

func (s *sortBuilder[T]) GetField(fieldName string, direction SortDirection) SortBuilder[T] {
	_, err := s.typeRegistry.ValidateFieldPath(s.typeInfo, fieldName)
	if err != nil {
		// Skip invalid fields
		return s
	}

	sortOrder := domain.SortOrderAsc
	if direction == SortDesc {
		sortOrder = domain.SortOrderDesc
	}

	s.sorts = append(s.sorts, domain.SortField{
		Field: fieldName,
		Order: sortOrder,
	})
	return s
}

func (s *sortBuilder[T]) Asc(fieldName string) SortBuilder[T] {
	return s.GetField(fieldName, SortAsc)
}

func (s *sortBuilder[T]) Desc(fieldName string) SortBuilder[T] {
	return s.GetField(fieldName, SortDesc)
}

func (s *sortBuilder[T]) ThenBy(fieldName string, direction SortDirection) SortBuilder[T] {
	return s.GetField(fieldName, direction)
}

func (s *sortBuilder[T]) ThenAsc(fieldName string) SortBuilder[T] {
	return s.GetField(fieldName, SortAsc)
}

func (s *sortBuilder[T]) ThenDesc(fieldName string) SortBuilder[T] {
	return s.GetField(fieldName, SortDesc)
}

func (s *sortBuilder[T]) Build() []domain.SortField {
	return s.sorts
}

// queryParamsBuilder implements QueryParamsBuilder interface
type queryParamsBuilder[T domain.IBaseModel] struct {
	queryParams  *QueryParams[T]
	typeInfo     reflect.Type
	typeRegistry TypeRegistry
}

func (q *queryParamsBuilder[T]) Filter() StronglyTypedFilterBuilder[T] {
	return &stronglyTypedFilterBuilder[T]{
		criteria:     []domain.FilterCriteria{},
		typeInfo:     q.typeInfo,
		typeRegistry: q.typeRegistry,
	}
}

func (q *queryParamsBuilder[T]) WithFilter(filter StronglyTypedFilter[T]) QueryParamsBuilder[T] {
	q.queryParams.filter = filter
	return q
}

func (q *queryParamsBuilder[T]) Sort() SortBuilder[T] {
	return &sortBuilder[T]{
		sorts:        []domain.SortField{},
		typeInfo:     q.typeInfo,
		typeRegistry: q.typeRegistry,
	}
}

func (q *queryParamsBuilder[T]) WithSort(sorts []domain.SortField) QueryParamsBuilder[T] {
	q.queryParams.sort = sorts
	return q
}

func (q *queryParamsBuilder[T]) WithLimit(limit int) QueryParamsBuilder[T] {
	q.queryParams.limit = limit
	return q
}

func (q *queryParamsBuilder[T]) WithOffset(offset int) QueryParamsBuilder[T] {
	q.queryParams.offset = offset
	return q
}

func (q *queryParamsBuilder[T]) WithPage(page, pageSize int) QueryParamsBuilder[T] {
	q.queryParams.limit = pageSize
	q.queryParams.offset = (page - 1) * pageSize
	return q
}

func (q *queryParamsBuilder[T]) WithPreloads(preloads []string) QueryParamsBuilder[T] {
	q.queryParams.preloads = preloads
	return q
}

func (q *queryParamsBuilder[T]) WithPreload(preload string) QueryParamsBuilder[T] {
	q.queryParams.preloads = append(q.queryParams.preloads, preload)
	return q
}

func (q *queryParamsBuilder[T]) IncludeDeleted() QueryParamsBuilder[T] {
	q.queryParams.includeDeleted = true
	q.queryParams.onlyDeleted = false
	return q
}

func (q *queryParamsBuilder[T]) OnlyDeleted() QueryParamsBuilder[T] {
	q.queryParams.includeDeleted = false
	q.queryParams.onlyDeleted = true
	return q
}

func (q *queryParamsBuilder[T]) ExcludeDeleted() QueryParamsBuilder[T] {
	q.queryParams.includeDeleted = false
	q.queryParams.onlyDeleted = false
	return q
}

func (q *queryParamsBuilder[T]) Build() QueryParams[T] {
	return *q.queryParams
}

// typeRegistry implements TypeRegistry interface
type typeRegistry struct {
	typeCache map[reflect.Type]map[string]*FieldInfo
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		typeCache: make(map[reflect.Type]map[string]*FieldInfo),
	}
}

func (tr *typeRegistry) RegisterType(model domain.IBaseModel) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if _, exists := tr.typeCache[modelType]; exists {
		return nil // Already registered
	}

	fields := make(map[string]*FieldInfo)
	err := tr.extractFields(modelType, "", []string{}, fields)
	if err != nil {
		return err
	}

	tr.typeCache[modelType] = fields
	return nil
}

func (tr *typeRegistry) extractFields(typ reflect.Type, prefix string, path []string, fields map[string]*FieldInfo) error {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		jsonName := field.Tag.Get("json")
		gormName := field.Tag.Get("gorm")

		// Parse JSON tag
		if jsonName != "" {
			if parts := strings.Split(jsonName, ","); len(parts) > 0 && parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fullPath := fieldName
		if prefix != "" {
			fullPath = prefix + "." + fieldName
		}

		currentPath := append(path, fieldName)

		fieldInfo := &FieldInfo{
			Name:     field.Name,
			JSONName: fieldName,
			GormName: tr.extractGormColumnName(gormName),
			TypeInfo: TypeInfo{
				Kind:      field.Type.Kind(),
				Type:      field.Type,
				IsPointer: field.Type.Kind() == reflect.Ptr,
				IsSlice:   field.Type.Kind() == reflect.Slice,
			},
			Path: currentPath,
		}

		if field.Type.Kind() == reflect.Ptr {
			fieldInfo.TypeInfo.ElementType = field.Type.Elem()
		} else if field.Type.Kind() == reflect.Slice {
			fieldInfo.TypeInfo.ElementType = field.Type.Elem()
		}

		fields[fullPath] = fieldInfo

		// Handle nested structs
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct && fieldType != typ { // Avoid infinite recursion
			// Recursively extract nested fields
			err := tr.extractFields(fieldType, fullPath, currentPath, fields)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (tr *typeRegistry) extractGormColumnName(gormTag string) string {
	if gormTag == "" {
		return ""
	}

	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func (tr *typeRegistry) GetFieldInfo(modelType reflect.Type, fieldPath string) (*FieldInfo, error) {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	fields, exists := tr.typeCache[modelType]
	if !exists {
		return nil, fmt.Errorf("type %s not registered", modelType.Name())
	}

	fieldInfo, exists := fields[fieldPath]
	if !exists {
		return nil, fmt.Errorf("field %s not found in type %s", fieldPath, modelType.Name())
	}

	return fieldInfo, nil
}

func (tr *typeRegistry) GetFields(modelType reflect.Type) map[string]*FieldInfo {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return tr.typeCache[modelType]
}

func (tr *typeRegistry) ValidateFieldPath(modelType reflect.Type, fieldPath string) (*FieldInfo, error) {
	return tr.GetFieldInfo(modelType, fieldPath)
}

// Factory functions
func NewQueryParamsBuilder[T domain.IBaseModel](registry TypeRegistry) QueryParamsBuilder[T] {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Register the type if not already registered
	registry.RegisterType(zero)

	return &queryParamsBuilder[T]{
		queryParams: &QueryParams[T]{
			limit:  50,
			offset: 0,
		},
		typeInfo:     modelType,
		typeRegistry: registry,
	}
}

func NewStronglyTypedFilter[T domain.IBaseModel](registry TypeRegistry) StronglyTypedFilterBuilder[T] {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Register the type if not already registered
	registry.RegisterType(zero)

	return &stronglyTypedFilterBuilder[T]{
		criteria:     []domain.FilterCriteria{},
		typeInfo:     modelType,
		typeRegistry: registry,
	}
}
