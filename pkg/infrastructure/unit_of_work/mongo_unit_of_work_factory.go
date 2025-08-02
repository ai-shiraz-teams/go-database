package unit_of_work

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoUnitOfWorkFactory provides MongoDB-based implementation of IUnitOfWorkFactory
type MongoUnitOfWorkFactory struct {
	client       *mongo.Client
	databaseName string
}

// NewMongoUnitOfWorkFactory creates a new MongoDB Unit of Work factory
func NewMongoUnitOfWorkFactory(client *mongo.Client, databaseName string) IUnitOfWorkFactory {
	return &MongoUnitOfWorkFactory{
		client:       client,
		databaseName: databaseName,
	}
}

// NewTransaction starts a new MongoDB session for transaction support
func (f *MongoUnitOfWorkFactory) NewTransaction(ctx context.Context) (interface{}, error) {
	session, err := f.client.StartSession()
	if err != nil {
		return nil, err
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return session, nil
}

// CommitTransaction commits the provided MongoDB session
func (f *MongoUnitOfWorkFactory) CommitTransaction(ctx context.Context, tx interface{}) error {
	session, ok := tx.(mongo.Session)
	if !ok {
		return fmt.Errorf("invalid transaction type: expected mongo.Session, got %T", tx)
	}

	err := session.CommitTransaction(ctx)
	session.EndSession(ctx)
	return err
}

// RollbackTransaction aborts the provided MongoDB session
func (f *MongoUnitOfWorkFactory) RollbackTransaction(ctx context.Context, tx interface{}) error {
	session, ok := tx.(mongo.Session)
	if !ok {
		return fmt.Errorf("invalid transaction type: expected mongo.Session, got %T", tx)
	}

	err := session.AbortTransaction(ctx)
	session.EndSession(ctx)
	return err
}
