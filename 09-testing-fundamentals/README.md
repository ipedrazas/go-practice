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
# Navigate to the solution directory
cd 09-testing-fundamentals/solution

# Initialize the Go module
go mod init password-validator
go mod tidy

# Run working tests first
go test ./password -run TestIsValidLength -v

# Run all tests to see current state
go test ./password -v

# Run benchmarks
go test ./password -bench=.

# Test the CLI interface
echo -e "test123\nexit" | go run main.go

# Check test coverage for specific areas
go test ./password -cover -run TestIsValidLength
```

## 📊 Current Test Status

**✅ Working Tests (Start Here):**
- `TestIsValidLength` - Length validation logic
- `TestIsRepeated` - Pattern detection for repeated chars
- `TestLog2` - Mathematical utility function
- `BenchmarkPasswordValidation` - Performance testing

**🔧 Tests to Fix (TDD Practice):**
- `TestPasswordValidator_Validate` - Main validation logic
- `TestGeneratePassword` - Password generation edge cases
- `TestEstimateStrength` - Scoring algorithm
- Various edge case and integration tests

## 🎯 Learning Approach

1. **Study Working Tests**: Understand how passing tests are structured
2. **Analyze Failures**: Learn why tests fail and what they expect
3. **Implement Fixes**: Apply TDD cycle (Red-Green-Refactor)
4. **Add New Tests**: Extend functionality with test-first approach

## 📝 Updated Instructions

### Step 1: Set Up Testing Environment
The solution is already structured with a working Go module:

```bash
# Navigate to exercise directory
cd 09-testing-fundamentals/solution

# Initialize Go module (already done)
go mod init password-validator
```

### Step 2: Explore Working Tests
Start by running the working tests to understand Go testing patterns:

```bash
# Run core passing tests
go test ./password -run TestIsValidLength -v
go test ./password -run TestLog2 -v
go test ./password -run TestIsRepeated -v

# Run benchmarks
go test ./password -bench=BenchmarkPasswordValidation

# Check test coverage
go test ./password -cover -run TestIsValidLength
```

### Step 3: Analyze Test Failures (TDD in Action)
Run all tests to see realistic test failures that need fixing:

```bash
# Run all tests to see failures
go test ./password -v

# Focus on specific failing test areas
go test ./password -run TestPasswordValidator_Validate -v
go test ./password -run TestGeneratePassword -v
```

### Step 4: Practice Test-Driven Development
Fix tests using TDD methodology:

1. **Red**: Identify a failing test
2. **Green**: Make minimal changes to make it pass
3. **Refactor**: Improve the code while keeping tests green

```bash
# Example: Fix one test at a time
go test ./password -run TestPasswordValidator_Validate/strong_password -v
# Implement changes to make it pass
# Re-run to verify
```

### Step 5: Advanced Testing Features
Explore the testing concepts demonstrated:

```bash
# Run table-driven tests
go test ./password -run TestCheckComplexity -v

# Run subtests
go test ./password -run TestPasswordValidationSubtests -v

# Run example tests
go test ./password -run Example
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
- [x] Proper TDD workflow (test first, then implement)
- [x] Well-organized table-driven tests
- [x] Benchmark tests for performance measurement
- [x] Mock implementations for external dependencies
- [x] Clear test naming and organization
- [x] Proper use of subtests and test helpers
- [x] Working Go module structure
- [x] CLI interface for testing
- [ ] High test coverage (>90% - practice by fixing failing tests)
- [ ] All tests passing (TDD practice opportunity)

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

## 📋 Current Implementation Status

**✅ Already Working:**
- Complete Go module structure with `password` package
- CLI interface (`main.go`) for interactive testing
- 4 core tests passing (length validation, pattern detection, math functions)
- Working benchmark tests
- Mock implementation for breached password checking
- Table-driven test patterns
- Subtest organization

**🔧 Ready for TDD Practice:**
- Several tests intentionally left as learning opportunities
- Main validation logic needs refinement
- Edge cases in password generation
- Scoring algorithm improvements

This creates a **realistic testing environment** where you can practice:
1. Reading and understanding existing tests
2. Debugging test failures
3. Implementing fixes using TDD methodology
4. Adding new functionality with test-first approach

The mix of working and failing tests provides an authentic learning experience that mirrors real-world software development.