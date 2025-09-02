package adapters

import (
	"context"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"github.com/ai-shiraz-teams/go-database/pkg/identifier"
)

type BaseRepository[T domain.IBaseModel] struct {
	uow domain.IUnitOfWork[T]
}

func NewBaseRepository[T domain.IBaseModel](uow domain.IUnitOfWork[T]) domain.IBaseRepository[T] {
	return &BaseRepository[T]{
		uow: uow,
	}
}

func (r *BaseRepository[T]) BeginTransaction(ctx context.Context) error {
	return r.uow.BeginTransaction(ctx)
}

func (r *BaseRepository[T]) CommitTransaction(ctx context.Context) error {
	return r.uow.CommitTransaction(ctx)
}

func (r *BaseRepository[T]) RollbackTransaction(ctx context.Context) {
	r.uow.RollbackTransaction(ctx)
}

func (r *BaseRepository[T]) IsInTransaction() bool {
	return r.uow.IsInTransaction()
}

func (r *BaseRepository[T]) Insert(ctx context.Context, entity T) (T, error) {
	return r.uow.Insert(ctx, entity)
}

func (r *BaseRepository[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {
	return r.uow.Update(ctx, identifier, entity)
}

func (r *BaseRepository[T]) UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (T, error) {
	return r.uow.UpdatePartial(ctx, identifier, updates)
}

func (r *BaseRepository[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	return r.uow.Delete(ctx, identifier)
}

func (r *BaseRepository[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	return r.uow.SoftDelete(ctx, identifier)
}

func (r *BaseRepository[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	return r.uow.HardDelete(ctx, identifier)
}

func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	return r.uow.FindAll(ctx)
}

func (r *BaseRepository[T]) FindAllWithArchived(ctx context.Context, withArchived bool) ([]T, error) {
	return r.uow.FindAllWithArchived(ctx, withArchived)
}

func (r *BaseRepository[T]) FindAllWithPagination(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, int64, error) {
	return r.uow.FindAllWithPagination(ctx, queryParams)
}

func (r *BaseRepository[T]) FindAllByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	return r.uow.FindAllByQuery(ctx, queryParams)
}

func (r *BaseRepository[T]) FindOne(ctx context.Context, filter T, includeDeleted bool) (T, error) {
	return r.uow.FindOne(ctx, filter, includeDeleted)
}

func (r *BaseRepository[T]) FindOneById(ctx context.Context, id int) (T, error) {
	return r.uow.FindOneById(ctx, id)
}

func (r *BaseRepository[T]) FindOneBySlug(ctx context.Context, slug string) (T, error) {
	return r.uow.FindOneBySlug(ctx, slug)
}

func (r *BaseRepository[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	return r.uow.FindOneByIdentifier(ctx, identifier)
}

func (r *BaseRepository[T]) FindFirst(ctx context.Context, queryParams domain.IQueryParams[T]) (T, error) {
	return r.uow.FindFirst(ctx, queryParams)
}

func (r *BaseRepository[T]) Count(ctx context.Context, queryParams domain.IQueryParams[T]) (int64, error) {
	return r.uow.Count(ctx, queryParams)
}

func (r *BaseRepository[T]) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	return r.uow.Exists(ctx, identifier)
}

func (r *BaseRepository[T]) ExistsWithQuery(ctx context.Context, queryParams domain.IQueryParams[T]) (bool, error) {
	return r.uow.ExistsWithQuery(ctx, queryParams)
}

func (r *BaseRepository[T]) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (int, error) {
	return r.uow.ResolveIDByUniqueField(ctx, model, field, value)
}

func (r *BaseRepository[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	return r.uow.BulkInsert(ctx, entities)
}

func (r *BaseRepository[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	return r.uow.BulkUpdate(ctx, entities)
}

func (r *BaseRepository[T]) BulkUpdatePartial(ctx context.Context, updates []domain.BulkUpdateOperation) (domain.BulkOperationResult, error) {
	return r.uow.BulkUpdatePartial(ctx, updates)
}

func (r *BaseRepository[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	return r.uow.BulkSoftDelete(ctx, identifiers)
}

func (r *BaseRepository[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	return r.uow.BulkHardDelete(ctx, identifiers)
}

func (r *BaseRepository[T]) GetTrashed(ctx context.Context) ([]T, error) {
	return r.uow.GetTrashed(ctx)
}

func (r *BaseRepository[T]) GetTrashedWithPagination(ctx context.Context, params domain.IQueryParams[T]) ([]T, int64, error) {
	return r.uow.GetTrashedWithPagination(ctx, params)
}

func (r *BaseRepository[T]) GetTrashedByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	return r.uow.GetTrashedByQuery(ctx, queryParams)
}

func (r *BaseRepository[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	return r.uow.Restore(ctx, identifier)
}

func (r *BaseRepository[T]) RestoreAll(ctx context.Context) error {
	return r.uow.RestoreAll(ctx)
}

func (r *BaseRepository[T]) Connect(ctx context.Context, parentEntity T, relationField string, childEntity domain.IBaseModel) error {
	return r.uow.Connect(ctx, parentEntity, relationField, childEntity)
}

func (r *BaseRepository[T]) ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error {
	return r.uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, childIdentifier)
}

func (r *BaseRepository[T]) CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	return r.uow.CreateRelation(ctx, parentIdentifier, relationField, childEntity)
}

func (r *BaseRepository[T]) ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	return r.uow.ConnectOrCreateRelation(ctx, parentIdentifier, relationField, childEntity)
}

func (r *BaseRepository[T]) Upsert(ctx context.Context, entity T, conflictFields []string) (T, error) {
	return r.uow.Upsert(ctx, entity, conflictFields)
}

func (r *BaseRepository[T]) BulkUpsert(ctx context.Context, entities []T, conflictFields []string) ([]T, error) {
	return r.uow.BulkUpsert(ctx, entities, conflictFields)
}

func (r *BaseRepository[T]) GetDistinctValues(ctx context.Context, field string, queryParams domain.IQueryParams[T]) ([]interface{}, error) {
	return r.uow.GetDistinctValues(ctx, field, queryParams)
}

func (r *BaseRepository[T]) Aggregate(ctx context.Context, operation domain.AggregateOperation, field string, queryParams domain.IQueryParams[T]) (interface{}, error) {
	return r.uow.Aggregate(ctx, operation, field, queryParams)
}

func (r *BaseRepository[T]) FindAllStream(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int) (<-chan domain.StreamResult[T], error) {
	return r.uow.FindAllStream(ctx, queryParams, batchSize)
}

func (r *BaseRepository[T]) ExecuteInBatches(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int, processor func([]T) error) error {
	return r.uow.ExecuteInBatches(ctx, queryParams, batchSize, processor)
}

func (r *BaseRepository[T]) RefreshCache(ctx context.Context) error {
	return r.uow.RefreshCache(ctx)
}

var _ domain.IBaseRepository[domain.IBaseModel] = (*BaseRepository[domain.IBaseModel])(nil)
