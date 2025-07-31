package unit_of_work

import (
	"context"
	"fmt"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"

	"gorm.io/gorm"
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
