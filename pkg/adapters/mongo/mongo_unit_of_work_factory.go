package mongo

import (
	"context"
	"fmt"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUnitOfWorkFactory struct {
	client       *mongo.Client
	databaseName string
}

func NewMongoUnitOfWorkFactory(client *mongo.Client, databaseName string) domain.IUnitOfWorkFactory {
	return &MongoUnitOfWorkFactory{
		client:       client,
		databaseName: databaseName,
	}
}

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

func (f *MongoUnitOfWorkFactory) CommitTransaction(ctx context.Context, tx interface{}) error {
	session, ok := tx.(mongo.Session)
	if !ok {
		return fmt.Errorf("invalid transaction type: expected mongo.Session, got %T", tx)
	}

	err := session.CommitTransaction(ctx)
	session.EndSession(ctx)
	return err
}

func (f *MongoUnitOfWorkFactory) RollbackTransaction(ctx context.Context, tx interface{}) error {
	session, ok := tx.(mongo.Session)
	if !ok {
		return fmt.Errorf("invalid transaction type: expected mongo.Session, got %T", tx)
	}

	err := session.AbortTransaction(ctx)
	session.EndSession(ctx)
	return err
}
