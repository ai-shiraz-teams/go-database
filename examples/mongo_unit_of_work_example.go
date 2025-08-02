package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/unit_of_work"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExampleEntity struct {
	ID        int        `bson:"id"`
	Name      string     `bson:"name"`
	Slug      string     `bson:"slug"`
	Email     string     `bson:"email"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

func (e *ExampleEntity) GetID() int               { return e.ID }
func (e *ExampleEntity) GetSlug() string          { return e.Slug }
func (e *ExampleEntity) GetCreatedAt() time.Time  { return e.CreatedAt }
func (e *ExampleEntity) GetUpdatedAt() time.Time  { return e.UpdatedAt }
func (e *ExampleEntity) GetDeletedAt() *time.Time { return e.DeletedAt }

func MongoDBUnitOfWorkExample() {

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	uow := unit_of_work.NewMongoUnitOfWork[*ExampleEntity](client, "testdb", "examples")

	fmt.Println("=== Basic CRUD Operations ===")

	entity := &ExampleEntity{
		ID:    1,
		Name:  "John Doe",
		Slug:  "john-doe",
		Email: "john@example.com",
	}

	created, err := uow.Insert(ctx, entity)
	if err != nil {
		log.Printf("Insert failed: %v", err)
	} else {
		fmt.Printf("Created entity: %+v\n", created)
	}

	found, err := uow.FindOneBySlug(ctx, "john-doe")
	if err != nil {
		log.Printf("Find by slug failed: %v", err)
	} else {
		fmt.Printf("Found by slug: %+v\n", found)
	}

	foundByID, err := uow.FindOneById(ctx, 1)
	if err != nil {
		log.Printf("Find by ID failed: %v", err)
	} else {
		fmt.Printf("Found by ID: %+v\n", foundByID)
	}

	fmt.Println("\n=== Using Identifier Builder ===")

	ident := identifier.NewIdentifier().
		Equal("email", "john@example.com").
		And(identifier.NewIdentifier().GreaterThan("id", 0))

	foundByIdent, err := uow.FindOneByIdentifier(ctx, ident)
	if err != nil {
		log.Printf("Find by identifier failed: %v", err)
	} else {
		fmt.Printf("Found by identifier: %+v\n", foundByIdent)
	}

	fmt.Println("\n=== Pagination Example ===")

	for i := 2; i <= 10; i++ {
		entity := &ExampleEntity{
			ID:    i,
			Name:  fmt.Sprintf("User %d", i),
			Slug:  fmt.Sprintf("user-%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
		uow.Insert(ctx, entity)
	}

	queryParams := query.NewQueryParams[*ExampleEntity]()
	queryParams.Page = 1
	queryParams.PageSize = 5
	queryParams = queryParams.AddSortDesc("id")

	entities, total, err := uow.FindAllWithPagination(ctx, queryParams)
	if err != nil {
		log.Printf("Pagination failed: %v", err)
	} else {
		fmt.Printf("Paginated results: %d total, %d on this page\n", total, len(entities))
		for _, e := range entities {
			fmt.Printf("  - %s (ID: %d)\n", e.Name, e.ID)
		}
	}

	fmt.Println("\n=== Advanced Filtering ===")

	queryParams = query.NewQueryParams[*ExampleEntity]().
		WithFilters(identifier.NewIdentifier().Like("name", "User%"))
	queryParams.Limit = 3

	filtered, total, err := uow.FindAllWithPagination(ctx, queryParams)
	if err != nil {
		log.Printf("Filtering failed: %v", err)
	} else {
		fmt.Printf("Filtered results: %d total, showing %d\n", total, len(filtered))
		for _, e := range filtered {
			fmt.Printf("  - %s\n", e.Name)
		}
	}

	fmt.Println("\n=== Soft Delete Operations ===")

	deleteIdent := identifier.NewIdentifier().Equal("id", 5)
	deleted, err := uow.SoftDelete(ctx, deleteIdent)
	if err != nil {
		log.Printf("Soft delete failed: %v", err)
	} else {
		fmt.Printf("Soft deleted: %s\n", deleted.Name)
	}

	trashed, err := uow.GetTrashed(ctx)
	if err != nil {
		log.Printf("Get trashed failed: %v", err)
	} else {
		fmt.Printf("Trashed items count: %d\n", len(trashed))
	}

	restored, err := uow.Restore(ctx, deleteIdent)
	if err != nil {
		log.Printf("Restore failed: %v", err)
	} else {
		fmt.Printf("Restored: %s\n", restored.Name)
	}

	fmt.Println("\n=== Transaction Example ===")

	err = uow.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Begin transaction failed: %v", err)
		return
	}

	newEntity := &ExampleEntity{
		ID:    100,
		Name:  "Transaction User",
		Slug:  "transaction-user",
		Email: "transaction@example.com",
	}

	_, err = uow.Insert(ctx, newEntity)
	if err != nil {
		log.Printf("Transaction insert failed: %v", err)
		uow.RollbackTransaction(ctx)
		return
	}

	err = uow.CommitTransaction(ctx)
	if err != nil {
		log.Printf("Commit failed: %v", err)
		return
	}

	fmt.Println("Transaction completed successfully")

	fmt.Println("\n=== Bulk Operations ===")

	bulkEntities := []*ExampleEntity{
		{ID: 200, Name: "Bulk User 1", Slug: "bulk-user-1", Email: "bulk1@example.com"},
		{ID: 201, Name: "Bulk User 2", Slug: "bulk-user-2", Email: "bulk2@example.com"},
		{ID: 202, Name: "Bulk User 3", Slug: "bulk-user-3", Email: "bulk3@example.com"},
	}

	bulkCreated, err := uow.BulkInsert(ctx, bulkEntities)
	if err != nil {
		log.Printf("Bulk insert failed: %v", err)
	} else {
		fmt.Printf("Bulk inserted %d entities\n", len(bulkCreated))
	}

	bulkDeleteIdents := []identifier.IIdentifier{
		identifier.NewIdentifier().Equal("id", 200),
		identifier.NewIdentifier().Equal("id", 201),
	}

	err = uow.BulkSoftDelete(ctx, bulkDeleteIdents)
	if err != nil {
		log.Printf("Bulk soft delete failed: %v", err)
	} else {
		fmt.Println("Bulk soft delete completed")
	}

	fmt.Println("\n=== Count and Exists Operations ===")

	countParams := query.NewQueryParams[*ExampleEntity]()
	count, err := uow.Count(ctx, countParams)
	if err != nil {
		log.Printf("Count failed: %v", err)
	} else {
		fmt.Printf("Total entities: %d\n", count)
	}

	existsIdent := identifier.NewIdentifier().Equal("email", "john@example.com")
	exists, err := uow.Exists(ctx, existsIdent)
	if err != nil {
		log.Printf("Exists check failed: %v", err)
	} else {
		fmt.Printf("Entity exists: %t\n", exists)
	}

	fmt.Println("\n=== MongoDB Unit of Work Example Complete ===")
}

func MongoDBFactoryExample() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	factory := unit_of_work.NewMongoUnitOfWorkFactory(client, "testdb")

	fmt.Println("=== Factory Transaction Example ===")

	session, err := factory.NewTransaction(ctx)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return
	}

	userUOW := unit_of_work.NewMongoUnitOfWork[*ExampleEntity](client, "testdb", "users")
	orderUOW := unit_of_work.NewMongoUnitOfWork[*ExampleEntity](client, "testdb", "orders")

	userUOW.BeginTransaction(ctx)
	orderUOW.BeginTransaction(ctx)

	err = factory.CommitTransaction(ctx, session)
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		factory.RollbackTransaction(ctx, session)
		return
	}

	fmt.Println("Factory transaction completed successfully")
}
