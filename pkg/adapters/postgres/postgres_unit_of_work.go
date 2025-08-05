package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"github.com/ai-shiraz-teams/go-database/pkg/identifier"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FilterApplier interface for applying query parameters to GORM queries
type FilterApplier[T domain.IBaseModel] interface {
	ApplyQueryParams(db *gorm.DB, queryParams domain.IQueryParams[T]) *gorm.DB
	ApplyIdentifier(db *gorm.DB, identifier identifier.IIdentifier) *gorm.DB
}

// GormFilterApplier implements FilterApplier for GORM
type GormFilterApplier[T domain.IBaseModel] struct{}

func NewGormFilterApplier[T domain.IBaseModel]() FilterApplier[T] {
	return &GormFilterApplier[T]{}
}

func (fa *GormFilterApplier[T]) ApplyQueryParams(db *gorm.DB, queryParams domain.IQueryParams[T]) *gorm.DB {
	// Apply filters
	if queryParams.HasFilters() {
		criteria := queryParams.ToFilterCriteria()
		for _, criterion := range criteria {
			db = fa.applyCriterion(db, criterion)
		}
	}

	// Apply preloads
	if queryParams.HasPreloads() {
		for _, preload := range queryParams.Preloads() {
			db = db.Preload(preload)
		}
	}

	// Apply sorting
	if queryParams.HasSort() {
		if sortMap, ok := queryParams.Sort().(domain.SortMap[T]); ok {
			for field, order := range sortMap {
				switch order {
				case domain.SortOrderAsc:
					db = db.Order(field + " ASC")
				case domain.SortOrderDesc:
					db = db.Order(field + " DESC")
				}
			}
		}
	}

	return db
}

func (fa *GormFilterApplier[T]) applyCriterion(db *gorm.DB, criterion domain.FilterCriteria) *gorm.DB {
	switch criterion.Operator {
	case domain.FilterOperatorEqual:
		return db.Where(criterion.Field+" = ?", criterion.Value)
	case domain.FilterOperatorNotEqual:
		return db.Where(criterion.Field+" != ?", criterion.Value)
	case domain.FilterOperatorGreaterThan:
		return db.Where(criterion.Field+" > ?", criterion.Value)
	case domain.FilterOperatorGreaterEqual:
		return db.Where(criterion.Field+" >= ?", criterion.Value)
	case domain.FilterOperatorLessThan:
		return db.Where(criterion.Field+" < ?", criterion.Value)
	case domain.FilterOperatorLessEqual:
		return db.Where(criterion.Field+" <= ?", criterion.Value)
	case domain.FilterOperatorLike:
		return db.Where(criterion.Field+" LIKE ?", criterion.Value)
	case domain.FilterOperatorIn:
		return db.Where(criterion.Field+" IN ?", criterion.Values)
	case domain.FilterOperatorNotIn:
		return db.Where(criterion.Field+" NOT IN ?", criterion.Values)
	case domain.FilterOperatorIsNull:
		return db.Where(criterion.Field + " IS NULL")
	case domain.FilterOperatorIsNotNull:
		return db.Where(criterion.Field + " IS NOT NULL")
	case domain.FilterOperatorBetween:
		if len(criterion.Values) >= 2 {
			return db.Where(criterion.Field+" BETWEEN ? AND ?", criterion.Values[0], criterion.Values[1])
		}
	}
	return db
}

// ApplyIdentifier applies identifier filters to GORM query
func (fa *GormFilterApplier[T]) ApplyIdentifier(db *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	if identifier == nil {
		return db
	}

	criteria := identifier.ToFilterCriteria()
	for _, criterion := range criteria {
		// Convert identifier.FilterCriteria to domain.FilterCriteria
		domainCriterion := domain.FilterCriteria{
			Field:     criterion.Field,
			Operator:  domain.FilterOperator(criterion.Operator),
			Value:     criterion.Value,
			Values:    criterion.Values,
			LogicalOp: domain.LogicalOperator(criterion.LogicalOp),
		}

		// Convert Group if present
		if len(criterion.Group) > 0 {
			domainCriterion.Group = make([]domain.FilterCriteria, len(criterion.Group))
			for i, gc := range criterion.Group {
				domainCriterion.Group[i] = domain.FilterCriteria{
					Field:     gc.Field,
					Operator:  domain.FilterOperator(gc.Operator),
					Value:     gc.Value,
					Values:    gc.Values,
					LogicalOp: domain.LogicalOperator(gc.LogicalOp),
				}
			}
		}

		db = fa.applyCriterion(db, domainCriterion)
	}

	return db
}

// BuildQueryFromIdentifier builds a GORM query from an identifier
func BuildQueryFromIdentifier[T domain.IBaseModel](db *gorm.DB, identifier identifier.IIdentifier) *gorm.DB {
	applier := &GormFilterApplier[T]{}
	query := db.Model(new(T))
	return applier.ApplyIdentifier(query, identifier)
}

type PostgresUnitOfWork[T domain.IBaseModel] struct {
	db            *gorm.DB
	tx            *gorm.DB
	filterApplier FilterApplier[T]
}

func NewPostgresUnitOfWork[T domain.IBaseModel](db *gorm.DB) domain.IUnitOfWork[T] {
	return &PostgresUnitOfWork[T]{
		db:            db,
		filterApplier: NewGormFilterApplier[T](),
	}
}

func (uow *PostgresUnitOfWork[T]) getDB() *gorm.DB {
	if uow.tx != nil {
		return uow.tx
	}
	return uow.db
}

func (uow *PostgresUnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already in progress")
	}

	tx := uow.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	uow.tx = tx
	return nil
}

func (uow *PostgresUnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	err := uow.tx.Commit().Error
	uow.tx = nil
	return err
}

func (uow *PostgresUnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	if uow.tx != nil {
		uow.tx.Rollback()
		uow.tx = nil
	}
}

func (uow *PostgresUnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	db := uow.getDB()
	if err := db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) FindAllWithArchived(ctx context.Context, withArchived bool) ([]T, error) {
	var entities []T
	db := uow.getDB()

	if withArchived {
		// Include archived/soft-deleted entities by using Unscoped
		db = db.Unscoped()
	}
	// Default behavior (withArchived=false) uses GORM's built-in soft-delete filtering

	if err := db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) FindAllWithPagination(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, int64, error) {
	db := uow.getDB()

	queryParams.PrepareDefaults()

	baseQuery := db.Model(new(T))

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	// Convert 1-based Offset to 0-based for SQL: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
	sqlOffset := 0
	if queryParams.Offset() > 1 {
		sqlOffset = (queryParams.Offset() - 1) * queryParams.Limit()
	}
	limit := queryParams.Limit()

	var total int64
	countQuery := filteredQuery.Session(&gorm.Session{NewDB: true})
	if err := countQuery.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entities []T
	if err := filteredQuery.WithContext(ctx).Offset(sqlOffset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (uow *PostgresUnitOfWork[T]) FindAllWithPaginationAndArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) ([]T, int64, error) {
	db := uow.getDB()

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	queryParams.PrepareDefaults()

	baseQuery := db.Model(new(T))

	// Apply withArchived logic by temporarily overriding the query parameter
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, false)
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	// Restore original value
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())
	}

	// Convert 1-based Offset to 0-based for SQL: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
	sqlOffset := 0
	if queryParams.Offset() > 1 {
		sqlOffset = (queryParams.Offset() - 1) * queryParams.Limit()
	}
	limit := queryParams.Limit()

	var total int64
	countQuery := filteredQuery.Session(&gorm.Session{NewDB: true})
	if err := countQuery.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entities []T
	if err := filteredQuery.WithContext(ctx).Offset(sqlOffset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (uow *PostgresUnitOfWork[T]) FindOne(ctx context.Context, filter T, includeDeleted bool) (T, error) {
	var entity T
	db := uow.getDB()

	if includeDeleted {
		db = db.Unscoped()
	}

	if err := db.WithContext(ctx).Where(filter).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneById(ctx context.Context, id int) (T, error) {
	var entity T
	db := uow.getDB()
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneBySlug(ctx context.Context, slug string) (T, error) {
	var entity T
	db := uow.getDB()
	if err := db.WithContext(ctx).Where("slug = ?", slug).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// WithArchived methods implementation
func (uow *PostgresUnitOfWork[T]) FindOneWithArchived(ctx context.Context, filter T, withArchived bool) (T, error) {
	var entity T
	db := uow.getDB()

	if withArchived {
		db = db.Unscoped()
	}

	if err := db.WithContext(ctx).Where(filter).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneByIdWithArchived(ctx context.Context, id int, withArchived bool) (T, error) {
	var entity T
	db := uow.getDB()

	if withArchived {
		db = db.Unscoped()
	}

	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneBySlugWithArchived(ctx context.Context, slug string, withArchived bool) (T, error) {
	var entity T
	db := uow.getDB()

	if withArchived {
		db = db.Unscoped()
	}

	if err := db.WithContext(ctx).Where("slug = ?", slug).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneByIdentifierWithArchived(ctx context.Context, identifier identifier.IIdentifier, withArchived bool) (T, error) {
	var entity T
	db := uow.getDB()

	if withArchived {
		db = db.Unscoped()
	}

	query := BuildQueryFromIdentifier[T](db, identifier)
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	db := uow.getDB()
	if err := db.WithContext(ctx).Create(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {

	_, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	db := uow.getDB()
	if err := db.WithContext(ctx).Save(entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	return query.WithContext(ctx).Delete(new(T)).Error
}

func (uow *PostgresUnitOfWork[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {

	entity, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)
	if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {

	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()
	var entity T
	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}

	if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	db := uow.getDB()
	var entities []T
	if err := db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, params domain.IQueryParams[T]) ([]T, int64, error) {

	if params == nil {
		params = domain.NewQueryParams[T]()
	}
	params = params.WithDeletedVisibility(false, true)
	return uow.FindAllWithPagination(ctx, params)
}

func (uow *PostgresUnitOfWork[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()

	var entity T
	if err := query.WithContext(ctx).Where("deleted_at IS NOT NULL").First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}

	if err := query.WithContext(ctx).Update("deleted_at", nil).Error; err != nil {
		var zero T
		return zero, err
	}

	var restoredEntity T
	if err := db.WithContext(ctx).First(&restoredEntity, uint(entity.GetID())).Error; err != nil {
		var zero T
		return zero, err
	}

	return restoredEntity, nil
}

func (uow *PostgresUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	db := uow.getDB()
	return db.WithContext(ctx).Model(new(T)).Unscoped().Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error
}

func (uow *PostgresUnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	db := uow.getDB()
	if err := db.WithContext(ctx).Create(&entities).Error; err != nil {
		return nil, err
	}

	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	db := uow.getDB()

	for i, entity := range entities {
		if err := db.WithContext(ctx).Save(&entity).Error; err != nil {
			return nil, err
		}
		entities[i] = entity
	}

	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	db := uow.getDB()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](db, identifier)
		if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
			return err
		}
	}

	return nil
}

func (uow *PostgresUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	db := uow.getDB()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()
		if err := query.WithContext(ctx).Delete(new(T)).Error; err != nil {
			return err
		}
	}

	return nil
}

func (uow *PostgresUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (int, error) {
	var entity T
	db := uow.getDB()

	if err := db.WithContext(ctx).Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value).First(&entity).Error; err != nil {
		return 0, err
	}

	return entity.GetID(), nil
}

func (uow *PostgresUnitOfWork[T]) Count(ctx context.Context, queryParams domain.IQueryParams[T]) (int64, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))

	// Create a simple query params adapter
	simpleParams := domain.NewQueryParams[T]()
	if queryParams != nil {
		if queryParams.HasFilters() {
			simpleParams = simpleParams.WithFilter(queryParams.Filter())
		}
		if queryParams.HasSort() {
			simpleParams = simpleParams.WithSort(queryParams.Sort())
		}
		simpleParams = simpleParams.WithLimit(queryParams.Limit()).WithOffset(queryParams.Offset())
		if queryParams.HasPreloads() {
			simpleParams = simpleParams.WithPreloads(queryParams.Preloads())
		}
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, simpleParams)

	var count int64
	if err := filteredQuery.WithContext(ctx).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (uow *PostgresUnitOfWork[T]) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)

	var count int64
	if err := query.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ========================
// New Extended Methods
// ========================

// IsInTransaction returns true if currently within a transaction
func (uow *PostgresUnitOfWork[T]) IsInTransaction() bool {
	return uow.tx != nil
}

// ========================
// 🧩 Entity Linking & Relations
// ========================

func (uow *PostgresUnitOfWork[T]) Connect(ctx context.Context, parentEntity T, relationField string, childEntity domain.IBaseModel) error {
	db := uow.getDB()
	return db.WithContext(ctx).Model(&parentEntity).Association(relationField).Append(childEntity)
}

func (uow *PostgresUnitOfWork[T]) ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error {
	db := uow.getDB()

	var parentEntity T
	parentQuery := BuildQueryFromIdentifier[T](db, parentIdentifier)
	if err := parentQuery.WithContext(ctx).First(&parentEntity).Error; err != nil {
		return err
	}

	var childEntity domain.IBaseModel
	childQuery := uow.filterApplier.ApplyIdentifier(db.Model(&childEntity), childIdentifier)
	if err := childQuery.WithContext(ctx).First(&childEntity).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Model(&parentEntity).Association(relationField).Append(childEntity)
}

func (uow *PostgresUnitOfWork[T]) CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	db := uow.getDB()

	if err := db.WithContext(ctx).Create(childEntity).Error; err != nil {
		return nil, err
	}

	slugIdentifier := identifier.NewIdentifier().Equal("slug", childEntity.GetSlug())
	return childEntity, uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, slugIdentifier)
}

func (uow *PostgresUnitOfWork[T]) ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	db := uow.getDB()

	var existing domain.IBaseModel
	if err := db.WithContext(ctx).Where("slug = ?", childEntity.GetSlug()).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uow.CreateRelation(ctx, parentIdentifier, relationField, childEntity)
		}
		return nil, err
	}

	slugIdentifier := identifier.NewIdentifier().Equal("slug", existing.GetSlug())
	return existing, uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, slugIdentifier)
}

// ========================
// 📄 Enhanced Retrieval
// ========================

func (uow *PostgresUnitOfWork[T]) FindAllByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var entities []T
	if err := filteredQuery.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) FindFirst(ctx context.Context, queryParams domain.IQueryParams[T]) (T, error) {
	var entity T
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	if err := filteredQuery.WithContext(ctx).First(&entity).Error; err != nil {
		return entity, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindManyRaw(ctx context.Context, rawQuery string, args ...interface{}) ([]T, error) {
	var entities []T
	db := uow.getDB()

	if err := db.WithContext(ctx).Raw(rawQuery, args...).Scan(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// ========================
// 🔍 Projections & Filters
// ========================

func (uow *PostgresUnitOfWork[T]) FindOneWithProjection(ctx context.Context, identifier identifier.IIdentifier, fields []string) (map[string]interface{}, error) {
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)

	if len(fields) > 0 {
		query = query.Select(fields)
	}

	var result map[string]interface{}
	if err := query.WithContext(ctx).Take(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (uow *PostgresUnitOfWork[T]) FindAllWithProjection(ctx context.Context, queryParams domain.IQueryParams[T], fields []string) ([]map[string]interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	if len(fields) > 0 {
		filteredQuery = filteredQuery.Select(fields)
	}

	var results []map[string]interface{}
	if err := filteredQuery.WithContext(ctx).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// WithArchived query methods implementation
func (uow *PostgresUnitOfWork[T]) FindAllByQueryWithArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) ([]T, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))

	// Apply withArchived logic by temporarily overriding the query parameter
	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	var entities []T
	if err := filteredQuery.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) FindFirstWithArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) (T, error) {
	var entity T
	db := uow.getDB()
	baseQuery := db.Model(new(T))

	// Apply withArchived logic by temporarily overriding the query parameter
	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	if err := filteredQuery.WithContext(ctx).First(&entity).Error; err != nil {
		return entity, err
	}
	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) FindOneWithProjectionAndArchived(ctx context.Context, identifier identifier.IIdentifier, fields []string, withArchived bool) (map[string]interface{}, error) {
	db := uow.getDB()

	if withArchived {
		db = db.Unscoped()
	}

	query := BuildQueryFromIdentifier[T](db, identifier)

	if len(fields) > 0 {
		query = query.Select(fields)
	}

	var result map[string]interface{}
	if err := query.WithContext(ctx).Take(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (uow *PostgresUnitOfWork[T]) FindAllWithProjectionAndArchived(ctx context.Context, queryParams domain.IQueryParams[T], fields []string, withArchived bool) ([]map[string]interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))

	// Apply withArchived logic by temporarily overriding the query parameter
	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	if len(fields) > 0 {
		filteredQuery = filteredQuery.Select(fields)
	}

	var results []map[string]interface{}
	if err := filteredQuery.WithContext(ctx).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (uow *PostgresUnitOfWork[T]) CountDistinct(ctx context.Context, field string, queryParams domain.IQueryParams[T]) (int64, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var count int64
	if err := filteredQuery.WithContext(ctx).Distinct(field).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========================
// 📥 Enhanced Create/Insert
// ========================

func (uow *PostgresUnitOfWork[T]) Upsert(ctx context.Context, entity T, conflictFields []string) (T, error) {
	db := uow.getDB()

	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: conflictFields[0]}},
		DoUpdates: clause.AssignmentColumns(conflictFields),
	}).Create(&entity)

	return entity, result.Error
}

func (uow *PostgresUnitOfWork[T]) BulkUpsert(ctx context.Context, entities []T, conflictFields []string) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	db := uow.getDB()

	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: conflictFields[0]}},
		DoUpdates: clause.AssignmentColumns(conflictFields),
	}).Create(&entities)

	return entities, result.Error
}

// ========================
// 🛠 Enhanced Update
// ========================

func (uow *PostgresUnitOfWork[T]) UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (T, error) {
	var entity T
	db := uow.getDB()
	query := BuildQueryFromIdentifier[T](db, identifier)

	if err := query.WithContext(ctx).Updates(updates).Error; err != nil {
		return entity, err
	}

	if err := query.WithContext(ctx).First(&entity).Error; err != nil {
		return entity, err
	}

	return entity, nil
}

func (uow *PostgresUnitOfWork[T]) BulkUpdatePartial(ctx context.Context, updates []domain.BulkUpdateOperation) (domain.BulkOperationResult, error) {
	result := domain.BulkOperationResult{}
	db := uow.getDB()

	for _, update := range updates {
		query := BuildQueryFromIdentifier[T](db, update.Identifier)
		if err := query.WithContext(ctx).Updates(update.Updates).Error; err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, err)
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// ========================
// 🧹 Enhanced Deletion
// ========================

func (uow *PostgresUnitOfWork[T]) DeleteAll(ctx context.Context, queryParams domain.IQueryParams[T], hardDelete bool) (domain.BulkOperationResult, error) {
	result := domain.BulkOperationResult{}
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	if hardDelete {
		filteredQuery = filteredQuery.Unscoped()
	}

	if err := filteredQuery.WithContext(ctx).Delete(new(T)).Error; err != nil {
		result.FailureCount++
		result.Errors = append(result.Errors, err)
		return result, err
	}

	result.SuccessCount = int(filteredQuery.RowsAffected)
	return result, nil
}

// ========================
// ♻️ Enhanced Restore
// ========================

func (uow *PostgresUnitOfWork[T]) BulkRestore(ctx context.Context, identifiers []identifier.IIdentifier) (domain.BulkOperationResult, error) {
	result := domain.BulkOperationResult{}
	db := uow.getDB()

	for _, identifier := range identifiers {
		query := BuildQueryFromIdentifier[T](db, identifier).Unscoped()
		if err := query.WithContext(ctx).Update("deleted_at", nil).Error; err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, err)
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// ========================
// 🗑 Enhanced Trash Views
// ========================

func (uow *PostgresUnitOfWork[T]) GetTrashedByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	var entities []T
	db := uow.getDB()
	baseQuery := db.Model(new(T)).Unscoped().Where("deleted_at IS NOT NULL")
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	if err := filteredQuery.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// ========================
// 🔧 Enhanced Utility Operations
// ========================

func (uow *PostgresUnitOfWork[T]) ExistsWithQuery(ctx context.Context, queryParams domain.IQueryParams[T]) (bool, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var count int64
	if err := filteredQuery.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (uow *PostgresUnitOfWork[T]) GetDistinctValues(ctx context.Context, field string, queryParams domain.IQueryParams[T]) ([]interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var values []interface{}
	if err := filteredQuery.WithContext(ctx).Distinct(field).Pluck(field, &values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (uow *PostgresUnitOfWork[T]) Aggregate(ctx context.Context, operation domain.AggregateOperation, field string, queryParams domain.IQueryParams[T]) (interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))

	// Create a simple query params adapter since we only need basic functionality
	simpleParams := domain.NewQueryParams[T]()
	if queryParams != nil {
		if queryParams.HasFilters() {
			// Copy filters from domain params to query params
			simpleParams = simpleParams.WithFilter(queryParams.Filter())
		}
		if queryParams.HasSort() {
			simpleParams = simpleParams.WithSort(queryParams.Sort())
		}
		simpleParams = simpleParams.WithLimit(queryParams.Limit()).WithOffset(queryParams.Offset())
		if queryParams.HasPreloads() {
			simpleParams = simpleParams.WithPreloads(queryParams.Preloads())
		}
	}

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, simpleParams)

	var result interface{}
	var selectClause string

	switch operation {
	case domain.AggregateSum:
		selectClause = fmt.Sprintf("SUM(%s)", field)
	case domain.AggregateAvg:
		selectClause = fmt.Sprintf("AVG(%s)", field)
	case domain.AggregateMin:
		selectClause = fmt.Sprintf("MIN(%s)", field)
	case domain.AggregateMax:
		selectClause = fmt.Sprintf("MAX(%s)", field)
	case domain.AggregateCount:
		selectClause = fmt.Sprintf("COUNT(%s)", field)
	default:
		return nil, fmt.Errorf("unsupported aggregate operation: %s", operation)
	}

	if err := filteredQuery.WithContext(ctx).Select(selectClause).Row().Scan(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ========================
// 📊 Performance & Streaming
// ========================

func (uow *PostgresUnitOfWork[T]) FindAllStream(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int) (<-chan domain.StreamResult[T], error) {
	resultChan := make(chan domain.StreamResult[T], batchSize)

	// Create a simple query params adapter
	simpleParams := domain.NewQueryParams[T]()
	if queryParams != nil {
		if queryParams.HasFilters() {
			simpleParams = simpleParams.WithFilter(queryParams.Filter())
		}
		if queryParams.HasSort() {
			simpleParams = simpleParams.WithSort(queryParams.Sort())
		}
		simpleParams = simpleParams.WithLimit(queryParams.Limit()).WithOffset(queryParams.Offset())
		if queryParams.HasPreloads() {
			simpleParams = simpleParams.WithPreloads(queryParams.Preloads())
		}
	}

	go func() {
		defer close(resultChan)

		offset := 0
		for {
			var entities []T
			db := uow.getDB()
			baseQuery := db.Model(new(T))
			filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, simpleParams)

			err := filteredQuery.WithContext(ctx).Offset(offset).Limit(batchSize).Find(&entities).Error
			if err != nil {
				resultChan <- domain.StreamResult[T]{Error: err}
				return
			}

			if len(entities) == 0 {
				return
			}

			for _, entity := range entities {
				select {
				case resultChan <- domain.StreamResult[T]{Data: entity}:
				case <-ctx.Done():
					return
				}
			}

			if len(entities) < batchSize {
				return
			}

			offset += batchSize
		}
	}()

	return resultChan, nil
}

func (uow *PostgresUnitOfWork[T]) ExecuteInBatches(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int, processor func([]T) error) error {
	offset := 0
	for {
		var entities []T
		db := uow.getDB()
		baseQuery := db.Model(new(T))
		filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

		if err := filteredQuery.WithContext(ctx).Offset(offset).Limit(batchSize).Find(&entities).Error; err != nil {
			return err
		}

		if len(entities) == 0 {
			break
		}

		if err := processor(entities); err != nil {
			return err
		}

		if len(entities) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}

func (uow *PostgresUnitOfWork[T]) RefreshCache(ctx context.Context) error {
	// PostgreSQL implementation doesn't have caching by default
	// This method is a no-op but satisfies the interface
	return nil
}
