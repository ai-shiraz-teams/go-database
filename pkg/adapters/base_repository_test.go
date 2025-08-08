package adapters

import (
	"context"
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
	"github.com/ai-shiraz-teams/go-database/pkg/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func newTestIdentifierByID(id int) identifier.IIdentifier {
	return identifier.NewIdentifier().Equal("id", id)
}

type MockUnitOfWork struct {
	MethodsCalled []string
}

func (m *MockUnitOfWork) trackCall(method string) {
	m.MethodsCalled = append(m.MethodsCalled, method)
}

func (m *MockUnitOfWork) BeginTransaction(ctx context.Context) error {
	m.trackCall("BeginTransaction")
	return nil
}

func (m *MockUnitOfWork) CommitTransaction(ctx context.Context) error {
	m.trackCall("CommitTransaction")
	return nil
}

func (m *MockUnitOfWork) RollbackTransaction(ctx context.Context) {
	m.trackCall("RollbackTransaction")
}

func (m *MockUnitOfWork) IsInTransaction() bool {
	m.trackCall("IsInTransaction")
	return false
}

func (m *MockUnitOfWork) Insert(ctx context.Context, entity *testutil.TestEntity) (*testutil.TestEntity, error) {
	m.trackCall("Insert")
	return entity, nil
}

func (m *MockUnitOfWork) Update(ctx context.Context, identifier identifier.IIdentifier, entity *testutil.TestEntity) (*testutil.TestEntity, error) {
	m.trackCall("Update")
	return entity, nil
}

func (m *MockUnitOfWork) UpdatePartial(ctx context.Context, identifier identifier.IIdentifier, updates map[string]interface{}) (*testutil.TestEntity, error) {
	m.trackCall("UpdatePartial")
	return &testutil.TestEntity{Name: "Updated"}, nil
}

func (m *MockUnitOfWork) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	m.trackCall("Delete")
	return nil
}

func (m *MockUnitOfWork) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.trackCall("SoftDelete")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.trackCall("HardDelete")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindAll(ctx context.Context) ([]*testutil.TestEntity, error) {
	m.trackCall("FindAll")
	return []*testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindAllWithArchived(ctx context.Context, withArchived bool) ([]*testutil.TestEntity, error) {
	m.trackCall("FindAllWithArchived")
	return []*testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindAllWithPagination(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, int64, error) {
	m.trackCall("FindAllWithPagination")
	return []*testutil.TestEntity{}, int64(0), nil
}

func (m *MockUnitOfWork) FindAllByQuery(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, error) {
	m.trackCall("FindAllByQuery")
	return []*testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindOne(ctx context.Context, filter *testutil.TestEntity, includeDeleted bool) (*testutil.TestEntity, error) {
	m.trackCall("FindOne")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindOneById(ctx context.Context, id int) (*testutil.TestEntity, error) {
	m.trackCall("FindOneById")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindOneBySlug(ctx context.Context, slug string) (*testutil.TestEntity, error) {
	m.trackCall("FindOneBySlug")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.trackCall("FindOneByIdentifier")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) FindFirst(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) (*testutil.TestEntity, error) {
	m.trackCall("FindFirst")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) Count(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) (int64, error) {
	m.trackCall("Count")
	return int64(0), nil
}

func (m *MockUnitOfWork) Exists(ctx context.Context, identifier identifier.IIdentifier) (bool, error) {
	m.trackCall("Exists")
	return false, nil
}

func (m *MockUnitOfWork) ExistsWithQuery(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) (bool, error) {
	m.trackCall("ExistsWithQuery")
	return false, nil
}

func (m *MockUnitOfWork) ResolveIDByUniqueField(ctx context.Context, model domain.IBaseModel, field string, value interface{}) (int, error) {
	m.trackCall("ResolveIDByUniqueField")
	return 0, nil
}

func (m *MockUnitOfWork) BulkInsert(ctx context.Context, entities []*testutil.TestEntity) ([]*testutil.TestEntity, error) {
	m.trackCall("BulkInsert")
	return entities, nil
}

func (m *MockUnitOfWork) BulkUpdate(ctx context.Context, entities []*testutil.TestEntity) ([]*testutil.TestEntity, error) {
	m.trackCall("BulkUpdate")
	return entities, nil
}

func (m *MockUnitOfWork) BulkUpdatePartial(ctx context.Context, updates []domain.BulkUpdateOperation) (domain.BulkOperationResult, error) {
	m.trackCall("BulkUpdatePartial")
	return domain.BulkOperationResult{SuccessCount: 1}, nil
}

func (m *MockUnitOfWork) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.trackCall("BulkSoftDelete")
	return nil
}

func (m *MockUnitOfWork) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	m.trackCall("BulkHardDelete")
	return nil
}

func (m *MockUnitOfWork) GetTrashed(ctx context.Context) ([]*testutil.TestEntity, error) {
	m.trackCall("GetTrashed")
	return []*testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) GetTrashedWithPagination(ctx context.Context, params domain.IQueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, int64, error) {
	m.trackCall("GetTrashedWithPagination")
	return []*testutil.TestEntity{}, int64(0), nil
}

func (m *MockUnitOfWork) GetTrashedByQuery(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity]) ([]*testutil.TestEntity, error) {
	m.trackCall("GetTrashedByQuery")
	return []*testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) Restore(ctx context.Context, identifier identifier.IIdentifier) (*testutil.TestEntity, error) {
	m.trackCall("Restore")
	return &testutil.TestEntity{}, nil
}

func (m *MockUnitOfWork) RestoreAll(ctx context.Context) error {
	m.trackCall("RestoreAll")
	return nil
}

func (m *MockUnitOfWork) Connect(ctx context.Context, parentEntity *testutil.TestEntity, relationField string, childEntity domain.IBaseModel) error {
	m.trackCall("Connect")
	return nil
}

func (m *MockUnitOfWork) ConnectByIdentifier(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childIdentifier identifier.IIdentifier) error {
	m.trackCall("ConnectByIdentifier")
	return nil
}

func (m *MockUnitOfWork) CreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	m.trackCall("CreateRelation")
	return childEntity, nil
}

func (m *MockUnitOfWork) ConnectOrCreateRelation(ctx context.Context, parentIdentifier identifier.IIdentifier, relationField string, childEntity domain.IBaseModel) (domain.IBaseModel, error) {
	m.trackCall("ConnectOrCreateRelation")
	return childEntity, nil
}

func (m *MockUnitOfWork) Upsert(ctx context.Context, entity *testutil.TestEntity, conflictFields []string) (*testutil.TestEntity, error) {
	m.trackCall("Upsert")
	return entity, nil
}

func (m *MockUnitOfWork) BulkUpsert(ctx context.Context, entities []*testutil.TestEntity, conflictFields []string) ([]*testutil.TestEntity, error) {
	m.trackCall("BulkUpsert")
	return entities, nil
}

func (m *MockUnitOfWork) GetDistinctValues(ctx context.Context, field string, queryParams domain.IQueryParams[*testutil.TestEntity]) ([]interface{}, error) {
	m.trackCall("GetDistinctValues")
	return []interface{}{}, nil
}

func (m *MockUnitOfWork) Aggregate(ctx context.Context, operation domain.AggregateOperation, field string, queryParams domain.IQueryParams[*testutil.TestEntity]) (interface{}, error) {
	m.trackCall("Aggregate")
	return 0, nil
}

func (m *MockUnitOfWork) FindAllStream(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity], batchSize int) (<-chan domain.StreamResult[*testutil.TestEntity], error) {
	m.trackCall("FindAllStream")
	ch := make(chan domain.StreamResult[*testutil.TestEntity])
	close(ch)
	return ch, nil
}

func (m *MockUnitOfWork) ExecuteInBatches(ctx context.Context, queryParams domain.IQueryParams[*testutil.TestEntity], batchSize int, processor func([]*testutil.TestEntity) error) error {
	m.trackCall("ExecuteInBatches")
	return nil
}

func (m *MockUnitOfWork) RefreshCache(ctx context.Context) error {
	m.trackCall("RefreshCache")
	return nil
}

func TestBaseRepository_DelegationToUnitOfWork(t *testing.T) {
	mockUow := new(MockUnitOfWork)
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)
	ctx := context.Background()

	testCases := []struct {
		name     string
		callFunc func()
		expected string
	}{
		{
			name: "BeginTransaction",
			callFunc: func() {
				repo.BeginTransaction(ctx)
			},
			expected: "BeginTransaction",
		},
		{
			name: "CommitTransaction",
			callFunc: func() {
				repo.CommitTransaction(ctx)
			},
			expected: "CommitTransaction",
		},
		{
			name: "RollbackTransaction",
			callFunc: func() {
				repo.RollbackTransaction(ctx)
			},
			expected: "RollbackTransaction",
		},
		{
			name: "IsInTransaction",
			callFunc: func() {
				repo.IsInTransaction()
			},
			expected: "IsInTransaction",
		},
		{
			name: "Insert",
			callFunc: func() {
				repo.Insert(ctx, &testutil.TestEntity{})
			},
			expected: "Insert",
		},
		{
			name: "Update",
			callFunc: func() {
				repo.Update(ctx, newTestIdentifierByID(1), &testutil.TestEntity{})
			},
			expected: "Update",
		},
		{
			name: "UpdatePartial",
			callFunc: func() {
				repo.UpdatePartial(ctx, newTestIdentifierByID(1), map[string]interface{}{"name": "test"})
			},
			expected: "UpdatePartial",
		},
		{
			name: "Delete",
			callFunc: func() {
				repo.Delete(ctx, newTestIdentifierByID(1))
			},
			expected: "Delete",
		},
		{
			name: "SoftDelete",
			callFunc: func() {
				repo.SoftDelete(ctx, newTestIdentifierByID(1))
			},
			expected: "SoftDelete",
		},
		{
			name: "HardDelete",
			callFunc: func() {
				repo.HardDelete(ctx, newTestIdentifierByID(1))
			},
			expected: "HardDelete",
		},
		{
			name: "FindAll",
			callFunc: func() {
				repo.FindAll(ctx)
			},
			expected: "FindAll",
		},
		{
			name: "FindAllWithArchived",
			callFunc: func() {
				repo.FindAllWithArchived(ctx, true)
			},
			expected: "FindAllWithArchived",
		},
		{
			name: "Count",
			callFunc: func() {

				queryParams := domain.NewQueryParams[*testutil.TestEntity]()
				repo.Count(ctx, queryParams)
			},
			expected: "Count",
		},
		{
			name: "Exists",
			callFunc: func() {
				repo.Exists(ctx, newTestIdentifierByID(1))
			},
			expected: "Exists",
		},
		{
			name: "BulkInsert",
			callFunc: func() {
				repo.BulkInsert(ctx, []*testutil.TestEntity{})
			},
			expected: "BulkInsert",
		},
		{
			name: "GetTrashed",
			callFunc: func() {
				repo.GetTrashed(ctx)
			},
			expected: "GetTrashed",
		},
		{
			name: "RestoreAll",
			callFunc: func() {
				repo.RestoreAll(ctx)
			},
			expected: "RestoreAll",
		},
		{
			name: "RefreshCache",
			callFunc: func() {
				repo.RefreshCache(ctx)
			},
			expected: "RefreshCache",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockUow.MethodsCalled = []string{}

			tc.callFunc()

			if len(mockUow.MethodsCalled) == 0 {
				t.Errorf("Expected %s to be called on UnitOfWork, but no methods were called", tc.expected)
			} else if mockUow.MethodsCalled[0] != tc.expected {
				t.Errorf("Expected %s to be called on UnitOfWork, but %s was called", tc.expected, mockUow.MethodsCalled[0])
			}
		})
	}
}

func TestBaseRepository_InterfaceCompliance(t *testing.T) {
	mockUow := new(MockUnitOfWork)
	var repo domain.IBaseRepository[*testutil.TestEntity] = NewBaseRepository[*testutil.TestEntity](mockUow)
	if repo == nil {
		t.Error("BaseRepository should implement IBaseRepository interface")
	}
}

func TestBaseRepository_Constructor(t *testing.T) {
	mockUow := new(MockUnitOfWork)
	repo := NewBaseRepository[*testutil.TestEntity](mockUow)

	if repo == nil {
		t.Error("NewBaseRepository should return a non-nil repository")
	}

	ctx := context.Background()
	mockUow.MethodsCalled = []string{}
	repo.FindAll(ctx)

	if len(mockUow.MethodsCalled) == 0 || mockUow.MethodsCalled[0] != "FindAll" {
		t.Error("Repository should delegate to UnitOfWork")
	}
}
