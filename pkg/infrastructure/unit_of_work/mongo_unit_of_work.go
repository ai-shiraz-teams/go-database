package unit_of_work

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUnitOfWork[T types.IBaseModel] struct {
	client         *mongo.Client
	database       *mongo.Database
	collection     *mongo.Collection
	session        mongo.Session
	filterApplier  *MongoFilterApplier
	collectionName string
}

func NewMongoUnitOfWork[T types.IBaseModel](client *mongo.Client, databaseName string, collectionName string) IUnitOfWork[T] {
	database := client.Database(databaseName)
	collection := database.Collection(collectionName)

	return &MongoUnitOfWork[T]{
		client:         client,
		database:       database,
		collection:     collection,
		filterApplier:  NewMongoFilterApplier(),
		collectionName: collectionName,
	}
}

func (uow *MongoUnitOfWork[T]) getCollection() *mongo.Collection {
	return uow.collection
}

func (uow *MongoUnitOfWork[T]) getSessionContext(ctx context.Context) context.Context {
	if uow.session != nil {
		return mongo.NewSessionContext(ctx, uow.session)
	}
	return ctx
}

func (uow *MongoUnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	if uow.session != nil {
		return fmt.Errorf("transaction already in progress")
	}

	session, err := uow.client.StartSession()
	if err != nil {
		return err
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return err
	}

	uow.session = session
	return nil
}

func (uow *MongoUnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	if uow.session == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	err := uow.session.CommitTransaction(ctx)
	uow.session.EndSession(ctx)
	uow.session = nil
	return err
}

func (uow *MongoUnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	if uow.session != nil {
		uow.session.AbortTransaction(ctx)
		uow.session.EndSession(ctx)
		uow.session = nil
	}
}

func (uow *MongoUnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	cursor, err := collection.Find(sessionCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(sessionCtx)

	var entities []T
	for cursor.Next(sessionCtx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}

func (uow *MongoUnitOfWork[T]) FindAllWithPagination(ctx context.Context, queryParams *query.QueryParams[T]) ([]T, int64, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	queryParams.PrepareDefaults()

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	total, err := collection.CountDocuments(sessionCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64(queryParams.ComputedOffset))
	findOptions.SetLimit(int64(queryParams.ComputedLimit))

	if len(queryParams.Sort) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.Sort)
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(sessionCtx)

	var entities []T
	for cursor.Next(sessionCtx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, 0, err
		}
		entities = append(entities, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (uow *MongoUnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filterDoc, err := uow.entityToFilter(filter)
	if err != nil {
		var zero T
		return zero, err
	}

	filterDoc["deleted_at"] = bson.M{"$exists": false}

	var entity T
	err = collection.FindOne(sessionCtx, filterDoc).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneById(ctx context.Context, id int) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{
		"id":         id,
		"deleted_at": bson.M{"$exists": false},
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneBySlug(ctx context.Context, slug string) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{
		"slug":       slug,
		"deleted_at": bson.M{"$exists": false},
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	uow.setTimestamps(entity, true, false)
	uow.setObjectID(entity)

	_, err := collection.InsertOne(sessionCtx, entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {
	_, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	uow.setTimestamps(entity, false, true)

	updateDoc := bson.M{"$set": entity}
	_, err = collection.UpdateOne(sessionCtx, filter, updateDoc)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	_, err := uow.SoftDelete(ctx, identifier)
	return err
}

func (uow *MongoUnitOfWork[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	entity, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	now := time.Now()
	updateDoc := bson.M{"$set": bson.M{"deleted_at": now}}

	_, err = collection.UpdateOne(sessionCtx, filter, updateDoc)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	_, err = collection.DeleteOne(sessionCtx, filter)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"deleted_at": bson.M{"$exists": true}}
	cursor, err := collection.Find(sessionCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(sessionCtx)

	var entities []T
	for cursor.Next(sessionCtx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}

func (uow *MongoUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, params *query.QueryParams[T]) ([]T, int64, error) {
	if params == nil {
		params = query.NewQueryParams[T]()
	}
	params.OnlyDeleted = true
	return uow.FindAllWithPagination(ctx, params)
}

func (uow *MongoUnitOfWork[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": true}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	updateDoc := bson.M{"$unset": bson.M{"deleted_at": ""}}
	_, err = collection.UpdateOne(sessionCtx, filter, updateDoc)
	if err != nil {
		var zero T
		return zero, err
	}

	restoredEntity, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	return restoredEntity, nil
}

func (uow *MongoUnitOfWork[T]) RestoreAll(ctx context.Context) error {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"deleted_at": bson.M{"$exists": true}}
	updateDoc := bson.M{"$unset": bson.M{"deleted_at": ""}}

	_, err := collection.UpdateMany(sessionCtx, filter, updateDoc)
	return err
}

func (uow *MongoUnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	documents := make([]interface{}, len(entities))
	for i, entity := range entities {
		uow.setTimestamps(entity, true, false)
		uow.setObjectID(entity)
		documents[i] = entity
	}

	_, err := collection.InsertMany(sessionCtx, documents)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (uow *MongoUnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	for i, entity := range entities {
		uow.setTimestamps(entity, false, true)

		filter := bson.M{"id": entity.GetID()}
		updateDoc := bson.M{"$set": entity}

		_, err := collection.UpdateOne(sessionCtx, filter, updateDoc)
		if err != nil {
			return nil, err
		}
		entities[i] = entity
	}

	return entities, nil
}

func (uow *MongoUnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	for _, identifier := range identifiers {
		filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
		filter["deleted_at"] = bson.M{"$exists": false}

		now := time.Now()
		updateDoc := bson.M{"$set": bson.M{"deleted_at": now}}

		_, err := collection.UpdateOne(sessionCtx, filter, updateDoc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uow *MongoUnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	for _, identifier := range identifiers {
		filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
		_, err := collection.DeleteOne(sessionCtx, filter)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uow *MongoUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model types.IBaseModel, field string, value interface{}) (int, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{
		field:        value,
		"deleted_at": bson.M{"$exists": false},
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		return 0, err
	}

	return entity.GetID(), nil
}

func (uow *MongoUnitOfWork[T]) Count(ctx context.Context, queryParams *query.QueryParams[T]) (int64, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.ApplyQueryParams(bson.M{"deleted_at": bson.M{"$exists": true}}, queryParams)
	return collection.CountDocuments(sessionCtx, filter)
}

func (uow *MongoUnitOfWork[T]) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	count, err := collection.CountDocuments(sessionCtx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (uow *MongoUnitOfWork[T]) entityToFilter(entity T) (bson.M, error) {
	doc := bson.M{}

	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.IsZero() {
			tag := fieldType.Tag.Get("bson")
			if tag == "" {
				tag = fieldType.Tag.Get("json")
			}
			if tag == "" {
				tag = fieldType.Name
			}

			if tag != "-" && tag != "" {
				doc[tag] = field.Interface()
			}
		}
	}

	return doc, nil
}

func (uow *MongoUnitOfWork[T]) setTimestamps(entity T, isCreate bool, isUpdate bool) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.CanSet() {
		return
	}

	now := time.Now()

	if isCreate {
		if createdAtField := v.FieldByName("CreatedAt"); createdAtField.IsValid() && createdAtField.CanSet() {
			createdAtField.Set(reflect.ValueOf(now))
		}
	}

	if isUpdate {
		if updatedAtField := v.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
			updatedAtField.Set(reflect.ValueOf(now))
		}
	}
}

func (uow *MongoUnitOfWork[T]) setObjectID(entity T) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.CanSet() {
		return
	}

	if objectIDField := v.FieldByName("ObjectID"); objectIDField.IsValid() && objectIDField.CanSet() && objectIDField.IsZero() {
		objectIDField.Set(reflect.ValueOf(primitive.NewObjectID()))
	}
}
