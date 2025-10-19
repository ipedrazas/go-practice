# Exercise 9: Testing Fundamentals and TDD

## 🎯 Objective
Learn Go's testing framework by building a simple library with comprehensive tests using Test-Driven Development (TDD).

## 📋 Main Focus Areas
- **Go testing framework** (`testing` package)
- **Test-Driven Development (TDD) methodology**
- **Unit testing and test organization**
- **Mocking and test doubles**
- **Benchmark testing** (`testing.B`)
- **Table-driven tests**
- **Test coverage and reporting**

## 🔧 What You'll Build
A password validation library with:
- Comprehensive unit tests
- Table-driven test patterns
- Benchmark tests for performance
- Mock implementations for external dependencies
- High test coverage
- Examples demonstrating test organization

## 📝 Instructions

### Step 1: Set Up Testing Environment
Create a password validation library with:
1. Password strength validation functions
2. Tests written before implementation (TDD)
3. Test file naming conventions (`*_test.go`)
4. Proper test organization and structure

### Step 2: Basic Unit Tests
Implement tests for:
- Password length requirements
- Character complexity (uppercase, lowercase, numbers, symbols)
- Common password detection
- Password scoring algorithm

### Step 3: Table-Driven Tests
Refactor tests using table-driven patterns:
- Test multiple input/output combinations
- Organize test cases efficiently
- Handle edge cases and error conditions
- Provide clear test descriptions

### Step 4: Advanced Testing Concepts
Add these testing techniques:
- Benchmark tests for performance measurement
- Mock implementations for external services
- Subtests for better organization
- Test helpers and utilities
- Setup/teardown with `TestMain`

### Step 5: Test Coverage and Reporting
Implement:
- Test coverage analysis
- Coverage thresholds
- Integration with CI/CD
- Examples and documentation tests

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 09-testing-fundamentals

# Create your package directory
mkdir password
cd password

# Create the implementation file
touch validator.go

# Create the test file
touch validator_test.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Run tests with verbose output
go test -v ./...
```

## 💡 Implementation Tips

### Basic Test Structure
```go
func TestPasswordLength(t *testing.T) {
    tests := []struct {
        name     string
        password string
        want     bool
    }{
        {"valid password", "strongPass123!", true},
        {"too short", "short", false},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsValidLength(tt.password)
            if got != tt.want {
                t.Errorf("IsValidLength(%q) = %v, want %v", tt.password, got, tt.want)
            }
        })
    }
}
```

### Benchmark Testing
```go
func BenchmarkPasswordValidation(b *testing.B) {
    password := "TestPassword123!"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ValidatePassword(password)
    }
}
```

### Test with Mocks
```go
func TestPasswordWithBreachedCheck(t *testing.T) {
    mockService := &MockBreachedService{
        breached: true,
    }

    validator := NewValidator(mockService)
    result := validator.Validate("password123")

    if result.IsBreached != true {
        t.Error("Expected password to be marked as breached")
    }
}
```

### Example Tests
```go
func ExampleValidatePassword() {
    result := ValidatePassword("StrongPass123!")
    fmt.Printf("Score: %d, Valid: %v\n", result.Score, result.Valid)
    // Output: Score: 85, Valid: true
}
```

## 🧪 Test Cases to Implement

### Password Length Tests
- Empty string
- Too short passwords (<8 characters)
- Minimum valid length (8 characters)
- Long passwords (>128 characters)

### Character Complexity Tests
- No uppercase letters
- No lowercase letters
- No numbers
- No special characters
- All character types present

### Common Password Tests
- Passwords on common lists
- Dictionary words
- Repeated characters
- Sequential characters

### Performance Tests
- Validation speed for various password lengths
- Memory usage benchmarks
- Concurrent validation performance

## 📚 Go Concepts Covered
- `testing` package fundamentals
- TDD methodology (Red-Green-Refactor)
- Table-driven test patterns
- Benchmark testing
- Test coverage tools
- Mock and stub implementations
- Subtest organization
- Example documentation tests

## 🔗 Useful Resources
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests in Go](https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go)
- [Testing with Go](https://golang.org/doc/code.html#Testing)
- [Test Coverage](https://blog.golang.org/cover)

## ✅ Success Criteria
Your exercise should demonstrate:
- [ ] Proper TDD workflow (test first, then implement)
- [ ] Well-organized table-driven tests
- [ ] High test coverage (>90%)
- [ ] Benchmark tests for performance measurement
- [ ] Mock implementations for external dependencies
- [ ] Clear test naming and organization
- [ ] Proper use of subtests and test helpers

## 🎁 Bonus Features
If you want an extra challenge:
- Property-based testing with fuzzing
- Integration tests with external APIs
- Race condition detection
- Custom test assertions
- Test data generation utilities
- Mutation testing
- Performance profiling

## 🔄 TDD Workflow to Follow

1. **Red**: Write a failing test
2. **Green**: Write the minimum code to make it pass
3. **Refactor**: Improve the code while keeping tests green
4. **Repeat** for each feature

When you're ready, check out the [solution](./solution/) to see a complete implementation with comprehensive tests.