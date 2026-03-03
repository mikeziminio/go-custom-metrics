---
name: go-modern
description: Modern Go code style skill focusing on best practices for Go 1.26, including updated idioms, concurrency patterns, and standard library usage
---

# Modern Go Code Style Skill (Go 1.26)

## Overview
Modern Go code style skill focusing on best practices for Go 1.26, including updated idioms, concurrency patterns, and standard library usage.

## Description
This skill provides guidance on modern Go 1.26 coding practices, emphasizing clean, efficient, and idiomatic code that leverages the latest language features and standard library improvements.

## Roles
- senior go backend engineer
- senior architecture engineer

## Dependencies
- go 1.26+
- standard library packages

## Modern Patterns and Practices

### 1. Collections and Data Structures
```go
// Modern slice operations
users := []User{{Name: "Alice"}, {Name: "Bob"}}
filtered := slices.Filter(nil, users, func(u User) bool {
    return u.Name != ""
})

// Modern map usage with generics
cache := sync.Map[string, *User]
cache.Store("alice", &User{Name: "Alice"})
if value, ok := cache.Load("alice"); ok {
    // Use value
}

// Efficient slice operations
result := slices.Collect(slices.Where(users, func(u User) bool {
    return u.IsActive
}))
```

### 2. Concurrency Patterns
```go
// Modern WaitGroup usage with wg.Go() (Go 1.26+)
var wg sync.WaitGroup
for _, item := range items {
    wg.Go(func() {
        // Process item
        process(item)
    })
}
wg.Wait()

// Using errgroup for error handling
var eg errgroup.Group
for _, item := range items {
    item := item // capture loop variable
    eg.Go(func() error {
        return process(item)
    })
}
if err := eg.Wait(); err != nil {
    // handle error
}
```

### 3. Error Handling
```go
// Proper error wrapping
func processUser(id string) error {
    user, err := db.FindUser(id)
    if err != nil {
        return fmt.Errorf("failed to find user %s: %w", id, err)
    }
    // process user
    return nil
}

// Using errors.Is and errors.As for error checking
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

### 4. Context Usage
```go
// Proper context usage with timeout
const timeout = 5 * time.Second
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

// With cancellation
ctx, cancel := context.WithCancel(ctx)
defer cancel()
```

### 5. Testing with Modern Practices
```go
// Table-driven tests
func TestProcess(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     string
        wantErr  bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Process(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Process() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Go 1.26 Specific Features

### 1. Enhanced Slice Operations
```go
// New slice functions
slices.Sort(numbers)
slices.Reverse(numbers)
slices.Contains(numbers, 42)
slices.Index(numbers, 42)
```

### 2. Improved Errors Package
```go
// Better error handling with errors.Join
var errs []error
if err := processA(); err != nil {
    errs = append(errs, err)
}
if err := processB(); err != nil {
    errs = append(errs, err)
}
if len(errs) > 0 {
    return errors.Join(errs...)
}
```

### 3. New Time Package Features
```go
// Time comparisons
if now.After(expiryTime) {
    // handle expired
}

// Duration formatting
duration := 2 * time.Hour + 30 * time.Minute
fmt.Printf("%v", duration) // 2h30m0s
```

### 4. WaitGroup Go() Method (Go 1.26+)
```go
// Modern WaitGroup usage with Go() method
var wg sync.WaitGroup
for _, item := range items {
    wg.Go(func() {
        process(item)
    })
}
wg.Wait()
```

## Taskfile Integration

Add the following to your root Taskfile.yml:

```yaml
# Modern code style tasks
modern-style-check:
  desc: Check code style compliance
  cmds:
    - echo "Checking modern code style..."
    - go vet ./...
    - golangci-lint run
    - echo "Code style check completed"

modern-style-fix:
  desc: Apply modern code style fixes
  cmds:
    - echo "Applying modern code style fixes..."
    - gofmt -w .
    - goimports -w .
    - echo "Modern code style fixes applied"

lint-modern:
  desc: Run modern linting
  cmds:
    - echo "Running modern linting..."
    - golangci-lint run --config=.golangci.yml
    - echo "Modern linting completed"
```

## Best Practices Checklist

### ✅ Modern Go 1.26 Patterns
- Use slices and maps with built-in functions
- Prefer errgroup over manual wait groups when appropriate
- Use context with proper timeout/cancellation
- Implement proper error handling with %w formatting
- Leverage new slice and string operations
- Use generics appropriately where it improves clarity
- Follow idiomatic error wrapping patterns

### ❌ Avoid These Patterns
- Avoid manual wait group pattern with wg.Add() and wg.Done()
- Don't use old-style error handling without proper wrapping
- Avoid unused imports and variables
- Don't ignore errors in production code
- Don't use deprecated APIs
- Avoid long chains of if-else-if statements for discrete value matching

## Refactoring Long If-Else Chains to Switch Cases

### Problem with Long If-Else Chains
When handling multiple discrete values or conditions, long chains of if-else-if statements become hard to read, maintain, and debug. They can also be less performant than switch statements.

### Solution: Use Switch Statements or Map-Based Lookup
Instead of:
```go
if condition == "value1" {
    // handle value1
} else if condition == "value2" {
    // handle value2
} else if condition == "value3" {
    // handle value3
} // ... and so on
```

Use a switch statement or map-based approach for better readability and performance.

### Recommended Approaches

#### 1. Simple Switch Statement
```go
switch condition {
case "value1":
    // handle value1
case "value2":
    // handle value2
case "value3":
    // handle value3
default:
    // handle unknown case
}
```

#### 2. Map-Based Lookup (for complex mappings)
```go
// Define a mapping function
handlers := map[string]func() {
    "value1": handleValue1,
    "value2": handleValue2,
    "value3": handleValue3,
}

if handler, exists := handlers[condition]; exists {
    handler()
} else {
    // handle unknown case
}
```

#### 3. Example from Current Codebase
In the stasis events handler, parameters are extracted from args using this pattern:
```go
// Before: Long if-else chain
if strings.HasPrefix(arg, "scenario=") {
    scenario = strings.TrimPrefix(arg, "scenario=")
} else if strings.HasPrefix(arg, "call_id=") {
    call_id = strings.TrimPrefix(arg, "call_id=")
} else if strings.HasPrefix(arg, "unique_id=") {
    unique_id = strings.TrimPrefix(arg, "unique_id=")
}
// ... continued for many parameters

// After: Cleaner approach using map or switch with helper function
// (This is more of a pattern for better organization)
```

### Benefits
1. **Improved Readability**: Clear separation of cases
2. **Better Performance**: Switch statements are optimized by Go compiler
3. **Easier Maintenance**: Adding/removing cases is simpler
4. **Reduced Cognitive Load**: Easier to understand the logic flow
5. **Better Debugging**: Clearer breakpoints and traceability

### Best Practices
- Use switch statements for discrete value matching
- Use map-based lookup for complex transformations or when handling many cases
- Ensure all cases are covered (include default case when appropriate)
- Keep switch cases focused and atomic
- Consider using switch expressions (Go 1.26+) for returning values

## Example Modern Implementation

```go
// Modern approach using wg.Go() and errgroup
const timeout = 5 * time.Second

func ProcessItems(ctx context.Context, items []Item) error {
    var eg errgroup.Group
    results := make([]Result, len(items))
    
    for i, item := range items {
        i, item := i, item // capture loop variables
        
        eg.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                result, err := processItem(item)
                if err != nil {
                    return fmt.Errorf("processing item %d: %w", i, err)
                }
                results[i] = result
                return nil
            }
        })
    }
    
    if err := eg.Wait(); err != nil {
        return fmt.Errorf("failed to process items: %w", err)
    }
    
    // Use results
    return nil
}

// Alternative with WaitGroup Go() method (Go 1.26+)
func ProcessItemsWithWG(items []Item) {
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Go(func() {
            process(item)
        })
    }
    wg.Wait()
}
```

## Migration Guide

When migrating existing code to modern style:
1. Replace manual wait groups with errgroup
2. Update error handling to use %w formatting
3. Replace wg.Add() + wg.Done() with simpler patterns
4. Use slice operations instead of manual loops
5. Ensure proper context propagation
6. Apply modern error wrapping conventions
