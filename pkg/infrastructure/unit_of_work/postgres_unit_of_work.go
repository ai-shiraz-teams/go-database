package unit_of_work

import (
	"context"
	"errors"
	"fmt"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresUnitOfWork[T types.IBaseModel] struct {
	db            *gorm.DB
	filterApplier *FilterApplier
	tx            *gorm.DB
}

func NewPostgresUnitOfWork[T types.IBaseModel](db *gorm.DB) IUnitOfWork[T] {
	return &PostgresUnitOfWork[T]{
		db:            db,
		filterApplier: NewFilterApplier(),
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

func (uow *PostgresUnitOfWork[T]) FindAllWithPagination(ctx context.Context, query *query.QueryParams[T]) ([]T, int64, error) {
	db := uow.getDB()

	query.PrepareDefaults()

	baseQuery := db.Model(new(T))

	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)

	offset := query.ComputedOffset
	limit := query.ComputedLimit

	var total int64
	countQuery := filteredQuery.Session(&gorm.Session{NewDB: true})
	if err := countQuery.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entities []T
	if err := filteredQuery.WithContext(ctx).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (uow *PostgresUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	var entity T
	db := uow.getDB()
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

func (uow *PostgresUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, params *query.QueryParams[T]) ([]T, int64, error) {

	if params == nil {
		params = query.NewQueryParams[T]()
	}
	params.OnlyDeleted = true
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

func (uow *PostgresUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error) {
	var entity T
	db := uow.getDB()

	if err := db.WithContext(ctx).Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value).First(&entity).Error; err != nil {
		return 0, err
	}

	return entity.GetID(), nil
}

func (uow *PostgresUnitOfWork[T]) Count(ctx context.Context, query *query.QueryParams[T]) (int64, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, query)

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

func (uow *PostgresUnitOfWork[T]) Connect(ctx context.Context, parentEntity T, relationField string, childEntity types.IBaseModel) error {
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

	var childEntity types.IBaseModel
	childQuery := uow.filterApplier.ApplyIdentifier(db.Model(&childEntity), childIdentifier)
	if err := childQuery.WithContext(ctx).First(&childEntity).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Model(&parentEntity).Association(relationField).Append(childEntity)
}

func (uow *PostgresUnitOfWork[T]) CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity types.IBaseModel) (types.IBaseModel, error) {
	db := uow.getDB()

	if err := db.WithContext(ctx).Create(childEntity).Error; err != nil {
		return nil, err
	}

	slugIdentifier := identifier.NewIdentifier().Equal("slug", childEntity.GetSlug())
	return childEntity, uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, slugIdentifier)
}

func (uow *PostgresUnitOfWork[T]) ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity types.IBaseModel) (types.IBaseModel, error) {
	db := uow.getDB()

	var existing types.IBaseModel
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

func (uow *PostgresUnitOfWork[T]) FindAllByQuery(ctx context.Context, queryParams *query.QueryParams[T]) ([]T, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var entities []T
	if err := filteredQuery.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (uow *PostgresUnitOfWork[T]) FindFirst(ctx context.Context, queryParams *query.QueryParams[T]) (T, error) {
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

func (uow *PostgresUnitOfWork[T]) FindAllWithProjection(ctx context.Context, queryParams *query.QueryParams[T], fields []string) ([]map[string]interface{}, error) {
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

func (uow *PostgresUnitOfWork[T]) CountDistinct(ctx context.Context, field string, queryParams *query.QueryParams[T]) (int64, error) {
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

func (uow *PostgresUnitOfWork[T]) BulkUpdatePartial(ctx context.Context, updates []BulkUpdateOperation) (BulkOperationResult, error) {
	result := BulkOperationResult{}
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

func (uow *PostgresUnitOfWork[T]) DeleteAll(ctx context.Context, queryParams *query.QueryParams[T], hardDelete bool) (BulkOperationResult, error) {
	result := BulkOperationResult{}
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

func (uow *PostgresUnitOfWork[T]) BulkRestore(ctx context.Context, identifiers []identifier.IIdentifier) (BulkOperationResult, error) {
	result := BulkOperationResult{}
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

func (uow *PostgresUnitOfWork[T]) GetTrashedByQuery(ctx context.Context, queryParams *query.QueryParams[T]) ([]T, error) {
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

func (uow *PostgresUnitOfWork[T]) ExistsWithQuery(ctx context.Context, queryParams *query.QueryParams[T]) (bool, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var count int64
	if err := filteredQuery.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (uow *PostgresUnitOfWork[T]) GetDistinctValues(ctx context.Context, field string, queryParams *query.QueryParams[T]) ([]interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var values []interface{}
	if err := filteredQuery.WithContext(ctx).Distinct(field).Pluck(field, &values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (uow *PostgresUnitOfWork[T]) Aggregate(ctx context.Context, operation AggregateOperation, field string, queryParams *query.QueryParams[T]) (interface{}, error) {
	db := uow.getDB()
	baseQuery := db.Model(new(T))
	filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

	var result interface{}
	var selectClause string

	switch operation {
	case AggregateSum:
		selectClause = fmt.Sprintf("SUM(%s)", field)
	case AggregateAvg:
		selectClause = fmt.Sprintf("AVG(%s)", field)
	case AggregateMin:
		selectClause = fmt.Sprintf("MIN(%s)", field)
	case AggregateMax:
		selectClause = fmt.Sprintf("MAX(%s)", field)
	case AggregateCount:
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

func (uow *PostgresUnitOfWork[T]) FindAllStream(ctx context.Context, queryParams *query.QueryParams[T], batchSize int) (<-chan StreamResult[T], error) {
	resultChan := make(chan StreamResult[T], batchSize)

	go func() {
		defer close(resultChan)

		offset := 0
		for {
			var entities []T
			db := uow.getDB()
			baseQuery := db.Model(new(T))
			filteredQuery := uow.filterApplier.ApplyQueryParams(baseQuery, queryParams)

			err := filteredQuery.WithContext(ctx).Offset(offset).Limit(batchSize).Find(&entities).Error
			if err != nil {
				resultChan <- StreamResult[T]{Error: err}
				return
			}

			if len(entities) == 0 {
				return
			}

			for _, entity := range entities {
				select {
				case resultChan <- StreamResult[T]{Data: entity}:
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

func (uow *PostgresUnitOfWork[T]) ExecuteInBatches(ctx context.Context, queryParams *query.QueryParams[T], batchSize int, processor func([]T) error) error {
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
