package mongo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"github.com/ai-shiraz-teams/go-database/pkg/identifier"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUnitOfWork[T domain.IBaseModel] struct {
	client         *mongo.Client
	database       *mongo.Database
	collection     *mongo.Collection
	session        mongo.Session
	filterApplier  *MongoFilterApplier
	collectionName string
}

func NewMongoUnitOfWork[T domain.IBaseModel](client *mongo.Client, databaseName string, collectionName string) domain.IUnitOfWork[T] {
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
		_ = uow.session.AbortTransaction(ctx) // Ignore error as this is a cleanup operation
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

func (uow *MongoUnitOfWork[T]) FindAllWithArchived(ctx context.Context, withArchived bool) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	var filter bson.M
	if withArchived {
		// Include all entities (both archived and non-archived)
		filter = bson.M{}
	} else {
		// Default behavior: only non-archived entities
		filter = bson.M{"deleted_at": bson.M{"$exists": false}}
	}

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

func (uow *MongoUnitOfWork[T]) FindAllWithPagination(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, int64, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	queryParams.PrepareDefaults()

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	total, err := collection.CountDocuments(sessionCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	// Convert 1-based Offset to 0-based for MongoDB: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
	mongoSkip := int64(0)
	if queryParams.Offset() > 1 {
		mongoSkip = int64((queryParams.Offset() - 1) * queryParams.Limit())
	}
	findOptions.SetSkip(mongoSkip)
	findOptions.SetLimit(int64(queryParams.Limit()))

	if queryParams.HasSort() && queryParams.ToSortFields() != nil {
		sortFields := queryParams.ToSortFields()
		sortDoc := uow.filterApplier.BuildSortDocument(sortFields)
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

func (uow *MongoUnitOfWork[T]) FindAllWithPaginationAndArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) ([]T, int64, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	queryParams.PrepareDefaults()

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}

	// Apply withArchived logic by temporarily overriding the query parameter
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	// Count total documents matching the filter
	total, err := collection.CountDocuments(sessionCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	// Convert 1-based Offset to 0-based for MongoDB: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
	mongoSkip := int64(0)
	if queryParams.Offset() > 1 {
		mongoSkip = int64((queryParams.Offset() - 1) * queryParams.Limit())
	}
	findOptions.SetSkip(mongoSkip)
	findOptions.SetLimit(int64(queryParams.Limit()))

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
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

func (uow *MongoUnitOfWork[T]) FindOne(ctx context.Context, filter T, includeDeleted bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filterDoc, err := uow.entityToFilter(filter)
	if err != nil {
		var zero T
		return zero, err
	}

	// Apply soft-delete filtering based on includeDeleted flag
	if !includeDeleted {
		filterDoc["deleted_at"] = bson.M{"$exists": false}
	}

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

// WithArchived methods implementation for MongoDB
func (uow *MongoUnitOfWork[T]) FindOneWithArchived(ctx context.Context, filter T, withArchived bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filterDoc, err := uow.entityToFilter(filter)
	if err != nil {
		var zero T
		return zero, err
	}

	if !withArchived {
		// Only include non-archived entities
		filterDoc["deleted_at"] = bson.M{"$exists": false}
	}
	// If withArchived is true, don't add any deleted_at filter (include all)

	var entity T
	err = collection.FindOne(sessionCtx, filterDoc).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneByIdWithArchived(ctx context.Context, id int, withArchived bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"id": id}
	if !withArchived {
		filter["deleted_at"] = bson.M{"$exists": false}
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneBySlugWithArchived(ctx context.Context, slug string, withArchived bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"slug": slug}
	if !withArchived {
		filter["deleted_at"] = bson.M{"$exists": false}
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneByIdentifierWithArchived(ctx context.Context, identifier identifier.IIdentifier, withArchived bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	if !withArchived {
		filter["deleted_at"] = bson.M{"$exists": false}
	}

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

func (uow *MongoUnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, params domain.IQueryParams[T]) ([]T, int64, error) {
	if params == nil {
		params = domain.NewQueryParams[T]()
	}
	params = params.WithDeletedVisibility(false, true)
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

func (uow *MongoUnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (int, error) {
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

func (uow *MongoUnitOfWork[T]) Count(ctx context.Context, queryParams domain.IQueryParams[T]) (int64, error) {
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

// ========================
// 🔁 Transaction Management (Additional)
// ========================

func (uow *MongoUnitOfWork[T]) IsInTransaction() bool {
	return uow.session != nil
}

// ========================
// 🧩 Entity Linking & Relations
// ========================

func (uow *MongoUnitOfWork[T]) Connect(ctx context.Context, parentEntity T, relationField string, childEntity domain.IBaseModel) error {
	// In MongoDB, relationships are typically embedded or referenced
	// This implementation uses embedded approach for simplicity
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	// Find parent document
	filter := bson.M{"id": parentEntity.GetID(), "deleted_at": bson.M{"$exists": false}}

	// Create update document to embed child entity in relation field
	update := bson.M{
		"$set": bson.M{
			relationField: childEntity,
			"updated_at":  time.Now(),
		},
	}

	_, err := collection.UpdateOne(sessionCtx, filter, update)
	return err
}

func (uow *MongoUnitOfWork[T]) ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error {
	// First find the child entity
	childCollection := uow.database.Collection(uow.collectionName) // Assuming same collection for simplicity
	childFilter := uow.filterApplier.BuildFilterFromIdentifier(childIdentifier)
	childFilter["deleted_at"] = bson.M{"$exists": false}

	var childEntity interface{}
	sessionCtx := uow.getSessionContext(ctx)
	err := childCollection.FindOne(sessionCtx, childFilter).Decode(&childEntity)
	if err != nil {
		return err
	}

	// Now update parent with child reference
	collection := uow.getCollection()
	parentFilter := uow.filterApplier.BuildFilterFromIdentifier(parentIdentifier)
	parentFilter["deleted_at"] = bson.M{"$exists": false}

	update := bson.M{
		"$set": bson.M{
			relationField: childEntity,
			"updated_at":  time.Now(),
		},
	}

	_, err = collection.UpdateOne(sessionCtx, parentFilter, update)
	return err
}

func (uow *MongoUnitOfWork[T]) CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	// First create the child entity
	childCollection := uow.database.Collection(uow.collectionName) // Assuming same collection for simplicity
	sessionCtx := uow.getSessionContext(ctx)

	// Set timestamps and ObjectID for child
	uow.setTimestampsForInterface(childEntity, true, false)
	uow.setObjectIDForInterface(childEntity)

	_, err := childCollection.InsertOne(sessionCtx, childEntity)
	if err != nil {
		return nil, err
	}

	// Now connect it to parent
	childIdentifier := identifier.NewIdentifier().Equal("id", childEntity.GetID())
	err = uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, childIdentifier)
	if err != nil {
		return nil, err
	}

	return childEntity, nil
}

func (uow *MongoUnitOfWork[T]) ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	// Try to find existing child entity
	childCollection := uow.database.Collection(uow.collectionName)
	sessionCtx := uow.getSessionContext(ctx)

	filter := bson.M{"id": childEntity.GetID(), "deleted_at": bson.M{"$exists": false}}
	var existingChild domain.IBaseModel
	err := childCollection.FindOne(sessionCtx, filter).Decode(&existingChild)

	if err == mongo.ErrNoDocuments {
		// Child doesn't exist, create it
		return uow.CreateRelation(ctx, parentIdentifier, relationField, childEntity)
	} else if err != nil {
		return nil, err
	}

	// Child exists, just connect it
	childIdentifier := identifier.NewIdentifier().Equal("id", existingChild.GetID())
	err = uow.ConnectByIdentifier(ctx, parentIdentifier, relationField, childIdentifier)
	if err != nil {
		return nil, err
	}

	return existingChild, nil
}

// ========================
// 📄 Retrieval (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) FindAllByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	queryParams.PrepareDefaults()

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	findOptions := options.Find()
	if queryParams.Limit() > 0 {
		// Convert 1-based Offset to 0-based for MongoDB: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
		mongoSkip := int64(0)
		if queryParams.Offset() > 1 {
			mongoSkip = int64((queryParams.Offset() - 1) * queryParams.Limit())
		}
		findOptions.SetSkip(mongoSkip)
		findOptions.SetLimit(int64(queryParams.Limit()))
	}

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
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

	return entities, cursor.Err()
}

func (uow *MongoUnitOfWork[T]) FindFirst(ctx context.Context, queryParams domain.IQueryParams[T]) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	findOptions := options.FindOne()
	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter, findOptions).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindManyRaw(ctx context.Context, rawQuery string, args ...interface{}) ([]T, error) {
	// For MongoDB, we need to parse the raw query as BSON
	// This is a simplified implementation - in practice, you might want more sophisticated parsing
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	// Attempt to parse rawQuery as BSON
	var filter bson.M
	if err := bson.UnmarshalExtJSON([]byte(rawQuery), true, &filter); err != nil {
		// If parsing fails, treat as MongoDB shell-style query
		filter = bson.M{"$where": rawQuery}
	}

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

	return entities, cursor.Err()
}

// ========================
// 🔍 Projections & Filters
// ========================

func (uow *MongoUnitOfWork[T]) FindOneWithProjection(ctx context.Context, identifier identifier.IIdentifier, fields []string) (map[string]interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	// Build projection document
	projection := bson.M{}
	for _, field := range fields {
		projection[field] = 1
	}

	findOptions := options.FindOne().SetProjection(projection)

	var result map[string]interface{}
	err := collection.FindOne(sessionCtx, filter, findOptions).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (uow *MongoUnitOfWork[T]) FindAllWithProjection(ctx context.Context, queryParams domain.IQueryParams[T], fields []string) ([]map[string]interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	queryParams.PrepareDefaults()

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Build projection document
	projection := bson.M{}
	for _, field := range fields {
		projection[field] = 1
	}

	findOptions := options.Find().SetProjection(projection)
	if queryParams.Limit() > 0 {
		// Convert 1-based Offset to 0-based for MongoDB: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
		mongoSkip := int64(0)
		if queryParams.Offset() > 1 {
			mongoSkip = int64((queryParams.Offset() - 1) * queryParams.Limit())
		}
		findOptions.SetSkip(mongoSkip)
		findOptions.SetLimit(int64(queryParams.Limit()))
	}

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(sessionCtx)

	var results []map[string]interface{}
	for cursor.Next(sessionCtx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, cursor.Err()
}

// Additional withArchived query methods for MongoDB
func (uow *MongoUnitOfWork[T]) FindAllByQueryWithArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}

	// Apply withArchived logic by temporarily overriding the query parameter
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	findOptions := options.Find()

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
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

	return entities, cursor.Err()
}

func (uow *MongoUnitOfWork[T]) FindFirstWithArchived(ctx context.Context, queryParams domain.IQueryParams[T], withArchived bool) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}

	// Apply withArchived logic by temporarily overriding the query parameter
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	findOptions := options.FindOne()

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	var entity T
	err := collection.FindOne(sessionCtx, filter, findOptions).Decode(&entity)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) FindOneWithProjectionAndArchived(ctx context.Context, identifier identifier.IIdentifier, fields []string, withArchived bool) (map[string]interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	if !withArchived {
		filter["deleted_at"] = bson.M{"$exists": false}
	}

	findOptions := options.FindOne()
	if len(fields) > 0 {
		projection := bson.M{}
		for _, field := range fields {
			projection[field] = 1
		}
		findOptions.SetProjection(projection)
	}

	var result map[string]interface{}
	err := collection.FindOne(sessionCtx, filter, findOptions).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (uow *MongoUnitOfWork[T]) FindAllWithProjectionAndArchived(ctx context.Context, queryParams domain.IQueryParams[T], fields []string, withArchived bool) ([]map[string]interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}

	// Apply withArchived logic by temporarily overriding the query parameter
	originalIncludeDeleted := queryParams.GetIncludeDeleted()
	if withArchived {
		queryParams = queryParams.WithDeletedVisibility(true, queryParams.GetOnlyDeleted())
	}

	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Restore original value
	queryParams = queryParams.WithDeletedVisibility(originalIncludeDeleted, queryParams.GetOnlyDeleted())

	findOptions := options.Find()

	if len(fields) > 0 {
		projection := bson.M{}
		for _, field := range fields {
			projection[field] = 1
		}
		findOptions.SetProjection(projection)
	}

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(sessionCtx)

	var results []map[string]interface{}
	for cursor.Next(sessionCtx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, cursor.Err()
}

func (uow *MongoUnitOfWork[T]) CountDistinct(ctx context.Context, field string, queryParams domain.IQueryParams[T]) (int64, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	// Use aggregation pipeline for distinct count
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id": "$" + field,
		}},
		{"$count": "distinctCount"},
	}

	cursor, err := collection.Aggregate(sessionCtx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(sessionCtx)

	var result []bson.M
	if err := cursor.All(sessionCtx, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	if count, ok := result[0]["distinctCount"].(int32); ok {
		return int64(count), nil
	}

	return 0, fmt.Errorf("unexpected result format from distinct count aggregation")
}

// ========================
// 📥 Create / Insert (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) Upsert(ctx context.Context, entity T, conflictFields []string) (T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	// Build filter based on conflict fields
	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for _, fieldName := range conflictFields {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			tag := fieldType.Tag.Get("bson")
			if tag == "" {
				tag = fieldType.Tag.Get("json")
			}
			if tag == "" {
				tag = fieldType.Name
			}

			if tag == fieldName || fieldType.Name == fieldName {
				filter[tag] = field.Interface()
				break
			}
		}
	}

	uow.setTimestamps(entity, true, true) // Set both created and updated timestamps

	upsertOptions := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(sessionCtx, filter, entity, upsertOptions)
	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

func (uow *MongoUnitOfWork[T]) BulkUpsert(ctx context.Context, entities []T, conflictFields []string) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	var result []T
	for _, entity := range entities {
		upsertedEntity, err := uow.Upsert(ctx, entity, conflictFields)
		if err != nil {
			return nil, err
		}
		result = append(result, upsertedEntity)
	}

	return result, nil
}

// ========================
// 🛠 Update (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (T, error) {
	// First find the entity to ensure it exists
	_, err := uow.FindOneByIdentifier(ctx, identifier)
	if err != nil {
		var zero T
		return zero, err
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
	filter["deleted_at"] = bson.M{"$exists": false}

	// Add updated timestamp
	updates["updated_at"] = time.Now()

	updateDoc := bson.M{"$set": updates}
	_, err = collection.UpdateOne(sessionCtx, filter, updateDoc)
	if err != nil {
		var zero T
		return zero, err
	}

	// Return updated entity
	return uow.FindOneByIdentifier(ctx, identifier)
}

func (uow *MongoUnitOfWork[T]) BulkUpdatePartial(ctx context.Context, updates []domain.BulkUpdateOperation) (domain.BulkOperationResult, error) {
	result := domain.BulkOperationResult{
		ProcessedIDs: []int{},
		Errors:       []error{},
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	for _, update := range updates {
		filter := uow.filterApplier.BuildFilterFromIdentifier(update.Identifier)
		filter["deleted_at"] = bson.M{"$exists": false}

		// Add updated timestamp
		update.Updates["updated_at"] = time.Now()

		updateDoc := bson.M{"$set": update.Updates}
		updateResult, err := collection.UpdateOne(sessionCtx, filter, updateDoc)

		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FailureCount++
		} else if updateResult.ModifiedCount > 0 {
			result.SuccessCount++
			// Note: In a real implementation, you might want to extract ID from the identifier
			// but this requires knowing the specific identifier implementation
		}
	}

	return result, nil
}

// ========================
// 🧹 Deletion (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) DeleteAll(ctx context.Context, queryParams domain.IQueryParams[T], hardDelete bool) (domain.BulkOperationResult, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	result := domain.BulkOperationResult{
		ProcessedIDs: []int{},
		Errors:       []error{},
	}

	if hardDelete {
		// Hard delete - permanently remove documents
		deleteResult, err := collection.DeleteMany(sessionCtx, filter)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FailureCount = int(deleteResult.DeletedCount)
		} else {
			result.SuccessCount = int(deleteResult.DeletedCount)
		}
	} else {
		// Soft delete - set deleted_at timestamp
		updateDoc := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
		updateResult, err := collection.UpdateMany(sessionCtx, filter, updateDoc)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FailureCount = int(updateResult.ModifiedCount)
		} else {
			result.SuccessCount = int(updateResult.ModifiedCount)
		}
	}

	return result, nil
}

func (uow *MongoUnitOfWork[T]) BulkRestore(ctx context.Context, identifiers []identifier.IIdentifier) (domain.BulkOperationResult, error) {
	result := domain.BulkOperationResult{
		ProcessedIDs: []int{},
		Errors:       []error{},
	}

	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	for _, identifier := range identifiers {
		filter := uow.filterApplier.BuildFilterFromIdentifier(identifier)
		filter["deleted_at"] = bson.M{"$exists": true}

		updateDoc := bson.M{"$unset": bson.M{"deleted_at": ""}}
		updateResult, err := collection.UpdateOne(sessionCtx, filter, updateDoc)

		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FailureCount++
		} else if updateResult.ModifiedCount > 0 {
			result.SuccessCount++
			// Note: In a real implementation, you might want to extract ID from the identifier
			// but this requires knowing the specific identifier implementation
		}
	}

	return result, nil
}

// ========================
// 🗑 Trash Views (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) GetTrashedByQuery(ctx context.Context, queryParams domain.IQueryParams[T]) ([]T, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}
	queryParams.PrepareDefaults()

	filter := bson.M{"deleted_at": bson.M{"$exists": true}}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	findOptions := options.Find()
	if queryParams.Limit() > 0 {
		// Convert 1-based Offset to 0-based for MongoDB: Offset=1 means page 1 (start at 0), Offset=2 means page 2 (start at Limit)
		mongoSkip := int64(0)
		if queryParams.Offset() > 1 {
			mongoSkip = int64((queryParams.Offset() - 1) * queryParams.Limit())
		}
		findOptions.SetSkip(mongoSkip)
		findOptions.SetLimit(int64(queryParams.Limit()))
	}

	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
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

	return entities, cursor.Err()
}

// ========================
// 🔧 Utility Operations (Extended)
// ========================

func (uow *MongoUnitOfWork[T]) ExistsWithQuery(ctx context.Context, queryParams domain.IQueryParams[T]) (bool, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	count, err := collection.CountDocuments(sessionCtx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (uow *MongoUnitOfWork[T]) GetDistinctValues(ctx context.Context, field string, queryParams domain.IQueryParams[T]) ([]interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	distinctValues, err := collection.Distinct(sessionCtx, field, filter)
	if err != nil {
		return nil, err
	}

	return distinctValues, nil
}

func (uow *MongoUnitOfWork[T]) Aggregate(ctx context.Context, operation domain.AggregateOperation, field string, queryParams domain.IQueryParams[T]) (interface{}, error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	var groupStage bson.M
	switch operation {
	case domain.AggregateSum:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$sum": "$" + field}}
	case domain.AggregateAvg:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$avg": "$" + field}}
	case domain.AggregateMin:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$min": "$" + field}}
	case domain.AggregateMax:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$max": "$" + field}}
	case domain.AggregateCount:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$sum": 1}}
	case domain.AggregateStdDev:
		groupStage = bson.M{"_id": nil, "result": bson.M{"$stdDevPop": "$" + field}}
	case domain.AggregateVariance:
		// MongoDB doesn't have direct variance, but we can calculate it using stdDev
		groupStage = bson.M{"_id": nil, "stddev": bson.M{"$stdDevPop": "$" + field}}
	default:
		return nil, fmt.Errorf("unsupported aggregate operation: %s", operation)
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": groupStage},
	}

	cursor, err := collection.Aggregate(sessionCtx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(sessionCtx)

	var results []bson.M
	if err := cursor.All(sessionCtx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]

	if operation == domain.AggregateVariance {
		// Calculate variance from standard deviation (variance = stddev^2)
		if stddev, ok := result["stddev"].(float64); ok {
			return stddev * stddev, nil
		}
		return nil, fmt.Errorf("unexpected result format for variance calculation")
	}

	return result["result"], nil
}

// ========================
// 📊 Performance & Streaming
// ========================

func (uow *MongoUnitOfWork[T]) FindAllStream(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int) (<-chan domain.StreamResult[T], error) {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	// Start with empty base filter - let ApplyQueryParams handle soft-delete logic
	filter := bson.M{}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	findOptions := options.Find().SetBatchSize(int32(batchSize))
	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan domain.StreamResult[T], batchSize)

	go func() {
		defer close(resultChan)
		defer cursor.Close(sessionCtx)

		for cursor.Next(sessionCtx) {
			var entity T
			if err := cursor.Decode(&entity); err != nil {
				resultChan <- domain.StreamResult[T]{Error: err}
				return
			}
			resultChan <- domain.StreamResult[T]{Data: entity}
		}

		if err := cursor.Err(); err != nil {
			resultChan <- domain.StreamResult[T]{Error: err}
		}
	}()

	return resultChan, nil
}

func (uow *MongoUnitOfWork[T]) ExecuteInBatches(ctx context.Context, queryParams domain.IQueryParams[T], batchSize int, processor func([]T) error) error {
	collection := uow.getCollection()
	sessionCtx := uow.getSessionContext(ctx)

	if queryParams == nil {
		queryParams = domain.NewQueryParams[T]()
	}

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	filter = uow.filterApplier.ApplyQueryParams(filter, queryParams)

	findOptions := options.Find().SetBatchSize(int32(batchSize))
	if len(queryParams.ToSortFields()) > 0 {
		sortDoc := uow.filterApplier.BuildSortDocument(queryParams.ToSortFields())
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(sessionCtx, filter, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(sessionCtx)

	var batch []T
	for cursor.Next(sessionCtx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return err
		}

		batch = append(batch, entity)

		if len(batch) >= batchSize {
			if err := processor(batch); err != nil {
				return err
			}
			batch = batch[:0] // Reset slice but keep capacity
		}
	}

	// Process remaining items in final batch
	if len(batch) > 0 {
		if err := processor(batch); err != nil {
			return err
		}
	}

	return cursor.Err()
}

func (uow *MongoUnitOfWork[T]) RefreshCache(ctx context.Context) error {
	// MongoDB doesn't have built-in caching like some ORMs
	// This implementation is a no-op, but could be extended to work with external caching layers
	return nil
}

// ========================
// 🔧 Helper Methods for Interface Support
// ========================

func (uow *MongoUnitOfWork[T]) setTimestampsForInterface(entity domain.IBaseModel, isCreate bool, isUpdate bool) {
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

func (uow *MongoUnitOfWork[T]) setObjectIDForInterface(entity domain.IBaseModel) {
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
