---
name: go-test-modify
description: Comprehensive skill for creating and modifying unit and integration tests in Go
---

# Go Testing Skill

## Overview
Comprehensive skill for creating and modifying unit and integration tests in Go

## Description
This skill provides testing functionality for Go projects, including unit testing, integration testing, and test automation. It follows Go testing best practices and integrates with the project's testing infrastructure.

## Constraints

- **NEVER** modify any files with the exception of `*_test.go` files.

## Dependencies
- go
- gotest
- testify (for assertions)

## Integration Points

### Test Structure
Tests should follow Go conventions:
- Files ending with `_test.go`
- Functions starting with `Test` prefix
- Use `testing.T` for test cases
- Use `testify` for assertions when needed

## Example Test Patterns

### Basic Unit Test
```go
func TestCalculateTotal(t *testing.T) {
    // Arrange
    input := []int{1, 2, 3, 4, 5}
    expected := 15
    
    // Act
    result := CalculateTotal(input)
    
    // Assert
    assert.Equal(t, expected, result)
}
```

### Mock Testing
```go
func TestProcessOrderWithMock(t *testing.T) {
    // Create mock dependencies
    mockDB := &MockDatabase{}
    mockDB.On("SaveOrder", mock.Anything).Return(nil)
    
    // Create service with mocks
    service := NewOrderService(mockDB)
    
    // Test
    err := service.ProcessOrder(order)
    
    // Assert
    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
}
```

### Integration Test
```go
func TestDatabaseConnection(t *testing.T) {
    // Integration test connecting to real database
    db, err := sql.Open("postgres", "connection-string")
    assert.NoError(t, err)
    defer db.Close()
    
    err = db.Ping()
    assert.NoError(t, err)
}
```

## Sample Tests from go-loyalty-system

### Integration Test with Test Containers (from internal/db/db_test.go):
```go
func TestFetchUserByLogin(t *testing.T) {
    ctx := t.Context()
    fillUsers(t)
    defer deleteUsers(t)

    var testCases = []struct {
        name    string
        login   string
        expUser *model.User
        expErr  error
    }{
        {
            name:  "success",
            login: "mike",
            expUser: &model.User{
                ID:        1,
                Login:     "mike",
                Password:  "mikepass",
                Token:     "miketoken",
                Balance:   0,
                Withdrawn: 0,
            },
        },
        {
            name:  "success with balance",
            login: "eugene",
            expUser: &model.User{
                ID:        2,
                Login:     "eugene",
                Password:  "eugenepass",
                Token:     "eugenetoken",
                Balance:   10.80,
                Withdrawn: 20,
            },
        },
        {
            name:   "user not found",
            login:  "nonexistentlogin",
            expErr: model.ErrUserNotFound,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            user, err := db.fetchUserByLogin(ctx, tc.login)
            assert.Equal(t, tc.expUser, user)
            assert.Equal(t, tc.expErr, err)
        })
    }
}
```

### Test with Test Containers and Mock Setup:
```go
func TestFetchUserByToken(t *testing.T) {
    ctx := t.Context()
    fillUsers(t)
    defer deleteUsers(t)

    var testCases = []struct {
        name    string
        token   string
        expUser *model.User
        expErr  error
    }{
        {
            name:  "success",
            token: "miketoken",
            expUser: &model.User{
                ID:        1,
                Login:     "mike",
                Password:  "mikepass",
                Token:     "miketoken",
                Balance:   0,
                Withdrawn: 0,
            },
        },
        {
            name:  "success with balance",
            token: "eugenetoken",
            expUser: &model.User{
                ID:        2,
                Login:     "eugene",
                Password:  "eugenepass",
                Token:     "eugenetoken",
                Balance:   10.80,
                Withdrawn: 20,
            },
        },
        {
            name:   "token not found",
            token:  "nonexistenttoken",
            expErr: model.ErrTokenNotFound,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            user, err := db.fetchUserByToken(ctx, tc.token)
            assert.Equal(t, tc.expUser, user)
            assert.Equal(t, tc.expErr, err)
        })
    }
}
```

## Best Practices for Test Organization

### ALWAYS Use Test Cases Slice Instead of Multiple t.Run Calls

In Go testing, when you have multiple related test scenarios, always organize them using a test cases slice structure instead of multiple individual t.Run calls. This approach provides several advantages:

1. **Better Maintainability**: All test scenarios are defined in one place
2. **Easier Extensions**: Adding new test cases is simple and consistent
3. **Improved Readability**: Clear separation of test data and expected outcomes
4. **Consistent Structure**: Follows the same pattern across all test files

#### Bad Practice (Multiple t.Run calls):
```go
func TestProcessDocumentCheck(t *testing.T) {
    t.Run("Deal ID of 0 should return Failed", func(t *testing.T) {
        // ... test implementation
    })
    
    t.Run("Normal case where all documents are uploaded should return All", func(t *testing.T) {
        // ... test implementation
    })
    
    // ... more t.Run calls
}
```

#### Good Practice (Test Cases Slice):
```go
func TestProcessDocumentCheck(t *testing.T) {
    testCases := []struct {
        name           string
        dealID         uint32
        surveys        []model.Survey
        documents      []model.Document
        expectedResult string
        expectedError  string
    }{
        {
            name:           "Deal ID of 0 should return Failed",
            dealID:         0,
            expectedResult: Failed,
        },
        {
            name: "Normal case where all documents are uploaded should return All",
            dealID: 123,
            surveys: []model.Survey{
                // ... survey data
            },
            documents: []model.Document{
                // ... document data
            },
            expectedResult: All,
        },
        // ... more test cases
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ... test implementation using tc values
        })
    }
}
```

### NEVER Use Manual Mocks, Always Use Mockery Generated Mocks

In Go testing, you should always use generated mocks from Mockery instead of manual mock implementations. Generated mocks provide:

1. **Consistency**: Standardized mock implementations across the codebase
2. **Type Safety**: Proper type checking and compile-time verification
3. **Maintainability**: Easier updates when interfaces change
4. **Functionality**: Full support for expectations, stubs, and verification
5. **Integration**: Proper integration with testify/mock framework

#### Bad Practice (Manual Mocks):
```go
// Manual mock implementation - NOT RECOMMENDED
type mockDocumentsAvailabilityFetcher struct {
	GetDocumentsAvailabilityFunc func(ctx context.Context, dealID int) ([]model.Document, error)
}

func (m *mockDocumentsAvailabilityFetcher) GetDocumentsAvailability(ctx context.Context, dealID int) ([]model.Document, error) {
	if m.GetDocumentsAvailabilityFunc != nil {
		return m.GetDocumentsAvailabilityFunc(ctx, dealID)
	}
	return nil, nil
}
```

#### Good Practice (Generated Mocks):
```go
// Using generated mockery mocks - RECOMMENDED
func TestProcessDocumentCheck(t *testing.T) {
    mockSurveysFetcher := NewDealSurveysFetcherMock(t)
    mockDocumentsFetcher := NewDocumentsAvailabilityFetcherMock(t)
    
    // Set up expectations
    mockDocumentsFetcher.EXPECT().GetDocumentsAvailability(ctx, 123).Return([]model.Document{{TypeID: 1, Uploaded: true}}, nil)
    
    // ... test implementation
}
```

### Test Setup with Test Containers (from TestMain):
```go
func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error
    log, err = zap.NewDevelopment()
    if err != nil {
        stdlog.Fatalf("failed to init logger: %v", err)
    }
    pgContainer = getContainer(ctx)
    connURL, err := pgContainer.ConnectionString(ctx)
    if err != nil {
        log.Fatal("failed to fetch connection URL", zap.Error(err))
    }
    db = getDB(ctx, connURL)

    code := m.Run()

    db.Close()
    pgContainer.Terminate(ctx)

    os.Exit(code)
}
```

## Database testing

For testing database queries, use "github.com/testcontainers/testcontainers-go"
It's important to test all DB queries with a real PostgreSQL database.
