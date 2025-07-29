# Go Database SDK

A clean, pluggable Go SDK for database abstractions, repository patterns, unit of work, and dynamic filtering.

## Requirements

- Go 1.24+

## Install

```bash
go get github.com/ai-shiraz-teams/go-database-sdk
```

## Folder Structure

- `internal/shared/` — Core SDK interfaces and abstractions
  - `types/` — Base entities and model interfaces
  - `repository/` — Repository pattern interfaces and implementations
  - `unit_of_work/` — Unit of Work pattern interfaces
  - `identifier/` — Dynamic filtering system
  - `query/` — Query parameters and pagination
  - `errors/` — Domain-specific error types
- `pkg/infrastructure/` — Concrete implementations
  - `unit_of_work/` — PostgreSQL Unit of Work implementation
  - `repository/` — Base repository implementations

## Usage

### Basic Setup

```go
import (
    "github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
    "github.com/ai-shiraz-teams/go-database-sdk/internal/shared/repository"
    "github.com/ai-shiraz-teams/go-database-sdk/pkg/infrastructure/unit_of_work"
    "gorm.io/gorm"
)

// Define your entity
type User struct {
    types.BaseEntity
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Initialize Unit of Work
db := // your gorm.DB instance
uow := unit_of_work.NewPostgresUnitOfWork[User](db)

// Create Repository
repo := repository.NewBaseRepository[User](uow)
```

### Repository Operations

```go
// Create
user := User{Name: "John", Email: "john@example.com"}
createdUser, err := repo.Insert(ctx, user)

// Find by ID
user, err := repo.FindOneById(ctx, 1)

// Find with pagination
params := query.NewQueryParams[User]().
    WithSearch("john").
    AddSortAsc("name")
users, total, err := repo.FindAllWithPagination(ctx, params)
```

### Dynamic Filtering

```go
import "github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"

// Build complex filters
filter := identifier.NewIdentifier().
    Equal("status", "active").
    And(identifier.NewIdentifier().GreaterThan("age", 18))

users, err := repo.FindOneByIdentifier(ctx, filter)
```

### Transaction Management

```go
// Begin transaction
err := uow.BeginTransaction(ctx)
if err != nil {
    return err
}

// Perform operations
user, err := repo.Insert(ctx, newUser)
if err != nil {
    uow.RollbackTransaction(ctx)
    return err
}

// Commit
err = uow.CommitTransaction(ctx)
```