package domain

import (
	"context"
	"time"
)

// This file contains comprehensive examples demonstrating how to use the QueryParams[T] struct
// and IUnitOfWork interface for typed, reusable, and paginated repository access.

// ExampleUserService demonstrates how a service layer would use QueryParams and IUnitOfWork
type ExampleUserService struct {
	userUOW IUnitOfWork[*User]
}

// NewExampleUserService creates a new user service with the provided unit of work
func NewExampleUserService(userUOW IUnitOfWork[*User]) *ExampleUserService {
	return &ExampleUserService{
		userUOW: userUOW,
	}
}

// GetActiveUsers demonstrates basic querying with QueryParams
func (s *ExampleUserService) GetActiveUsers(ctx context.Context, page, pageSize int) ([]User, uint, error) {
	// Create query parameters with pagination
	params := NewQueryParams[*User]()
	params.Page = page
	params.PageSize = pageSize
	params.PrepareDefaults(50, 200) // Default: 50, Max: 200

	// Add filtering for users with email (non-null emails as proxy for "active")
	identifier := NewIdentifier().
		IsNotNull("email").
		IsNotNull("username")

	params.WithFilters(identifier)

	// Add sorting by creation date (newest first)
	params.AddSortDesc("createdAt")

	// Execute query
	users, totalCount, err := s.userUOW.FindAllWithPagination(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Convert pointers to values for return
	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, totalCount, nil
}

// SearchUsers demonstrates text search with complex filtering
func (s *ExampleUserService) SearchUsers(ctx context.Context, searchTerm string, includeInactive bool) ([]User, error) {
	params := NewQueryParams[*User]().
		WithSearch(searchTerm).
		AddSortAsc("username")

	// Build dynamic filters based on parameters
	identifier := NewIdentifier()

	if !includeInactive {
		// Filter for users with non-empty username as proxy for "active"
		identifier = identifier.IsNotNull("username")
	}

	// Add date range filter (users created in last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	identifier = identifier.GreaterOrEqual("createdAt", thirtyDaysAgo)

	params.WithFilters(identifier)

	users, _, err := s.userUOW.FindAllWithPagination(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, nil
}

// GetUsersByEmailDomain demonstrates complex filtering with IN operator
func (s *ExampleUserService) GetUsersByEmailDomain(ctx context.Context, domains []string) ([]User, error) {
	// Create LIKE filters for each domain
	identifier := NewIdentifier()

	// Build OR conditions for each domain
	for i, domain := range domains {
		domainFilter := NewIdentifier().Like("email", "%@"+domain)
		if i == 0 {
			identifier = domainFilter
		} else {
			identifier = identifier.Or(domainFilter)
		}
	}

	params := NewQueryParams[*User]().
		WithFilters(identifier).
		AddSortAsc("email")

	users, _, err := s.userUOW.FindAllWithPagination(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, nil
}

// CreateUserWithTransaction demonstrates transactional operations
func (s *ExampleUserService) CreateUserWithTransaction(ctx context.Context, user *User) (*User, error) {
	// Begin transaction
	err := s.userUOW.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			s.userUOW.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Check if user with email already exists
	identifier := NewIdentifier().Equal("email", user.Email)
	exists, err := s.userUOW.Exists(ctx, identifier)
	if err != nil {
		s.userUOW.RollbackTransaction(ctx)
		return nil, err
	}

	if exists {
		s.userUOW.RollbackTransaction(ctx)
		return nil, &UserAlreadyExistsError{Email: user.Email}
	}

	// Insert the user
	createdUser, err := s.userUOW.Insert(ctx, user)
	if err != nil {
		s.userUOW.RollbackTransaction(ctx)
		return nil, err
	}

	// Commit transaction
	err = s.userUOW.CommitTransaction(ctx)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// UpdateUserEmail demonstrates update operations with optimistic locking
func (s *ExampleUserService) UpdateUserEmail(ctx context.Context, userID uint, newEmail string) (*User, error) {
	// Find user by ID
	user, err := s.userUOW.FindOneById(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update email
	user.Email = newEmail

	// Create identifier for the specific user with version check (optimistic locking)
	identifier := NewIdentifier().
		Equal("id", userID).
		Equal("version", user.GetVersion())

	// Update using identifier
	updatedUser, err := s.userUOW.Update(ctx, identifier, user)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// SoftDeleteInactiveUsers demonstrates bulk soft delete operations
func (s *ExampleUserService) SoftDeleteInactiveUsers(ctx context.Context, inactiveSince time.Time) error {
	// Create identifier for users with missing data (as proxy for inactive)
	identifier := NewIdentifier().
		IsNull("firstName").
		LessThan("createdAt", inactiveSince)

	// Use bulk soft delete for efficiency
	identifiers := []IIdentifier{identifier}
	return s.userUOW.BulkSoftDelete(ctx, identifiers)
}

// GetTrashedUsersWithPagination demonstrates working with soft-deleted records
func (s *ExampleUserService) GetTrashedUsersWithPagination(ctx context.Context, page, pageSize int) ([]User, uint, error) {
	params := NewQueryParams[*User]()
	params.Page = page
	params.PageSize = pageSize
	params.OnlyDeletedRecords() // Show only soft-deleted records
	params.AddSortDesc("deletedAt")

	params.PrepareDefaults(50, 100)

	users, totalCount, err := s.userUOW.GetTrashedWithPagination(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, totalCount, nil
}

// RestoreUserByEmail demonstrates restore operations
func (s *ExampleUserService) RestoreUserByEmail(ctx context.Context, email string) (*User, error) {
	identifier := NewIdentifier().Equal("email", email)

	return s.userUOW.Restore(ctx, identifier)
}

// GetUserStatistics demonstrates count operations
func (s *ExampleUserService) GetUserStatistics(ctx context.Context) (*UserStatistics, error) {
	// Count users with complete profiles (firstName and lastName not null)
	completeParams := NewQueryParams[*User]().
		WithFilters(NewIdentifier().
			IsNotNull("firstName").
			IsNotNull("lastName"))

	completeCount, err := s.userUOW.Count(ctx, completeParams)
	if err != nil {
		return nil, err
	}

	// Count users with incomplete profiles
	incompleteParams := NewQueryParams[*User]().
		WithFilters(NewIdentifier().
			IsNull("firstName").
			Or(NewIdentifier().IsNull("lastName")))

	incompleteCount, err := s.userUOW.Count(ctx, incompleteParams)
	if err != nil {
		return nil, err
	}

	// Count soft-deleted users
	deletedParams := NewQueryParams[*User]()
	deletedParams.OnlyDeletedRecords()

	deletedCount, err := s.userUOW.Count(ctx, deletedParams)
	if err != nil {
		return nil, err
	}

	return &UserStatistics{
		ActiveUsers:   completeCount,
		InactiveUsers: incompleteCount,
		DeletedUsers:  deletedCount,
		TotalUsers:    completeCount + incompleteCount,
	}, nil
}

// ExampleAdvancedFiltering demonstrates complex filter combinations
func (s *ExampleUserService) ExampleAdvancedFiltering(ctx context.Context) ([]User, error) {
	// Create complex filter:
	// (firstName IS NOT NULL OR lastName IS NOT NULL) AND
	// createdAt > lastMonth AND
	// (email LIKE '%@company.com' OR email LIKE '%@enterprise.com')

	lastMonth := time.Now().Add(-30 * 24 * time.Hour)

	// Name filters (firstName OR lastName not null)
	nameFilter := NewIdentifier().
		IsNotNull("firstName").
		Or(NewIdentifier().IsNotNull("lastName"))

	// Date filter
	dateFilter := NewIdentifier().GreaterThan("createdAt", lastMonth)

	// Email domain filters
	emailFilter := NewIdentifier().
		Like("email", "%@company.com").
		Or(NewIdentifier().Like("email", "%@enterprise.com"))

	// Combine all filters with AND
	combinedFilter := nameFilter.
		And(dateFilter).
		And(emailFilter)

	params := NewQueryParams[*User]().
		WithFilters(combinedFilter).
		AddSortDesc("createdAt").
		AddSortAsc("email") // Secondary sort

	users, _, err := s.userUOW.FindAllWithPagination(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, nil
}

// ExampleBulkOperations demonstrates bulk insert and update operations
func (s *ExampleUserService) ExampleBulkOperations(ctx context.Context, newUsers []*User) error {
	// Begin transaction for bulk operations
	err := s.userUOW.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			s.userUOW.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Bulk insert new users
	insertedUsers, err := s.userUOW.BulkInsert(ctx, newUsers)
	if err != nil {
		s.userUOW.RollbackTransaction(ctx)
		return err
	}

	// Update usernames of inserted users (example modification)
	for i, user := range insertedUsers {
		user.Username = user.Username + "_verified"
		insertedUsers[i] = user
	}

	_, err = s.userUOW.BulkUpdate(ctx, insertedUsers)
	if err != nil {
		s.userUOW.RollbackTransaction(ctx)
		return err
	}

	return s.userUOW.CommitTransaction(ctx)
}

// Supporting types for examples

// UserAlreadyExistsError represents an error when trying to create a user that already exists
type UserAlreadyExistsError struct {
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	return "user with email " + e.Email + " already exists"
}

// UserStatistics holds statistical information about users
type UserStatistics struct {
	ActiveUsers   int64 `json:"activeUsers"`
	InactiveUsers int64 `json:"inactiveUsers"`
	DeletedUsers  int64 `json:"deletedUsers"`
	TotalUsers    int64 `json:"totalUsers"`
}
