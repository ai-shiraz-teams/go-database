package repository

import (
	"context"

	"go-database/pkg/domain"
	"go-database/pkg/infrastructure/unit_of_work"

	"gorm.io/gorm"
)

// Example: UserRepository demonstrates how feature-specific repositories
// can be built using the BaseRepository delegation pattern.

// IUserRepository defines business-specific operations for User entities.
// It embeds IBaseRepository to inherit all basic CRUD operations.
type IUserRepository interface {
	IBaseRepository[*domain.User]

	// Business-specific methods
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	SearchByEmailDomain(ctx context.Context, domain string) ([]*domain.User, error)
	GetActiveUsersCount(ctx context.Context) (int64, error)
}

// UserRepository provides User-specific repository operations by extending BaseRepository.
// It demonstrates the composition pattern where business logic delegates to BaseRepository,
// which in turn delegates to IUnitOfWork.
type UserRepository struct {
	IBaseRepository[*domain.User]
}

// NewUserRepository creates a new UserRepository instance.
// It accepts a UnitOfWork and creates a BaseRepository internally for delegation.
func NewUserRepository(uow domain.IUnitOfWork[*domain.User]) IUserRepository {
	return &UserRepository{
		IBaseRepository: NewBaseRepository(uow),
	}
}

// NewUserRepositoryFromDB creates a UserRepository with a PostgreSQL UnitOfWork.
// This is a convenience constructor for direct database access.
func NewUserRepositoryFromDB(db *gorm.DB) IUserRepository {
	postgresUOW := unit_of_work.NewPostgresUnitOfWork[*domain.User](db)
	return NewUserRepository(postgresUOW)
}

// Business-specific methods that add value beyond basic CRUD

// FindByEmail finds a user by their email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	identifier := domain.NewIdentifier().Equal("email", email)
	return r.FindOneByIdentifier(ctx, identifier)
}

// FindByUsername finds a user by their username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	identifier := domain.NewIdentifier().Equal("username", username)
	return r.FindOneByIdentifier(ctx, identifier)
}

// SearchByEmailDomain finds all users with emails in a specific domain
func (r *UserRepository) SearchByEmailDomain(ctx context.Context, emailDomain string) ([]*domain.User, error) {
	identifier := domain.NewIdentifier().Like("email", "%@"+emailDomain+"%")

	params := domain.NewQueryParams[*domain.User]()
	params.WithFilters(identifier)
	params.AddSortAsc("email")

	users, _, err := r.FindAllWithPagination(ctx, params)
	return users, err
}

// GetActiveUsersCount returns the count of users who are considered "active"
// (users with both email and username set)
func (r *UserRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	identifier := domain.NewIdentifier().
		IsNotNull("email").
		IsNotNull("username")

	params := domain.NewQueryParams[*domain.User]()
	params.WithFilters(identifier)

	return r.Count(ctx, params)
}

// Compile-time check to ensure UserRepository implements IUserRepository
var _ IUserRepository = (*UserRepository)(nil)

// Example usage patterns:

/*
// Pattern 1: Direct construction with UnitOfWork (recommended for dependency injection)
func NewUserService(uow domain.IUnitOfWork[*domain.User]) *UserService {
    return &UserService{
        userRepo: NewUserRepository(uow),
    }
}

// Pattern 2: Construction with database connection (convenience method)
func NewUserServiceFromDB(db *gorm.DB) *UserService {
    return &UserService{
        userRepo: NewUserRepositoryFromDB(db),
    }
}

// Pattern 3: Testing with mock UnitOfWork
func TestUserService(t *testing.T) {
    mockUOW := &MockUnitOfWork[*domain.User]{}
    userRepo := NewUserRepository(mockUOW)
    service := &UserService{userRepo: userRepo}

    // Test business logic without database dependencies
}
*/
