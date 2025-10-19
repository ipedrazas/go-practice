package password

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"testing"
)

// MockBreachedService implements BreachedService for testing
type MockBreachedService struct {
	breached map[string]bool
}

func (m *MockBreachedService) IsBreached(password string) bool {
	return m.breached[password]
}

func NewMockBreachedService(breachedPasswords []string) *MockBreachedService {
	breached := make(map[string]bool)
	for _, pwd := range breachedPasswords {
		breached[pwd] = true
	}
	return &MockBreachedService{breached: breached}
}

func TestPasswordValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string
		password string
		breached []string
		want     ValidationResult
	}{
		{
			name:     "strong password",
			password: "StrongP@ssw0rd123!",
			breached: []string{},
			want: ValidationResult{
				Valid:   true,
				Score:   100,
				Errors:  []string{},
				IsBreached: false,
			},
		},
		{
			name:     "too short password",
			password: "Short",
			breached: []string{},
			want: ValidationResult{
				Valid:  false,
				Score:  0,
				Errors: []string{"Password must be at least 8 characters"},
				IsBreached: false,
			},
		},
		{
			name:     "breached password",
			password: "password123",
			breached: []string{"password123"},
			want: ValidationResult{
				Valid:     false,
				Score:     0,
				Errors:    []string{"Password has been found in data breaches"},
				IsBreached: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := NewMockBreachedService(tt.breached)
			validator := NewPasswordValidator(mockSvc)
			got := validator.Validate(tt.password)

			// Check validation result
			if got.Valid != tt.want.Valid {
				t.Errorf("Validate() Valid = %v, want %v", got.Valid, tt.want.Valid)
			}

			// Check if breached status matches
			if got.IsBreached != tt.want.IsBreached {
				t.Errorf("Validate() IsBreached = %v, want %v", got.IsBreached, tt.want.IsBreached)
			}

			// Check if we have errors when we expect them
			if len(got.Errors) == 0 && len(tt.want.Errors) > 0 {
				t.Errorf("Validate() expected errors but got none")
			}

			// Check score is reasonable
			if tt.want.Valid && got.Score < 60 {
				t.Errorf("Validate() valid password should have score >= 60, got %d", got.Score)
			}

			// Check for appropriate suggestions
			if tt.password == "Short" && len(got.Suggestions) == 0 {
				t.Errorf("Validate() should provide suggestions for weak passwords")
			}
		})
	}
}

func TestIsValidLength(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"empty string", "", false},
		{"too short", "abc", false},
		{"minimum length", "abcdefgh", true},
		{"good length", "StrongPassword123!", true},
		{"too long", strings.Repeat("a", 130), false},
		{"max length", strings.Repeat("a", 128), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidLength(tt.password)
			if got != tt.want {
				t.Errorf("isValidLength(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestCheckComplexity(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		minScore int
	}{
		{"no uppercase", "lowercase123!", 30},
		{"no lowercase", "UPPERCASE123!", 30},
		{"no numbers", "NoNumbers!", 35},
		{"no special chars", "NoSpecial123", 30},
		{"all character types", "Mixed123!@#", 55},
		{"only lowercase", "lowercase", 10},
		{"only numbers", "12345678", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{}
			got := validator.checkComplexity(tt.password, result)
			if got < tt.minScore {
				t.Errorf("checkComplexity(%q) = %v, want >= %v", tt.password, got, tt.minScore)
			}
		})
	}
}

func TestHasCommonPatterns(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"sequential numbers", "password1234", true},
		{"sequential letters", "passwordabcd", true},
		{"repeated characters", "passwordaaa", true},
		{"keyboard pattern", "passwordqwerty", true},
		{"keyboard pattern reverse", "passwordytrewq", true},
		{"no patterns", "Pa$$w0rd!xYz", false},
		{"partial sequential", "pass12word", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.hasCommonPatterns(strings.ToLower(tt.password))
			if got != tt.want {
				t.Errorf("hasCommonPatterns(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestIsSequential(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"sequential ascending", "abc", true},
		{"sequential numbers", "123", true},
		{"sequential descending", "cba", true},
		{"sequential numbers descending", "321", true},
		{"non sequential", "axc", false},
		{"short string", "ab", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isSequential(tt.password)
			if got != tt.want {
				t.Errorf("isSequential(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestIsRepeated(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"three same chars", "aaab", true},
		{"three same numbers", "1112", true},
		{"no repeats", "abc123", false},
		{"only two repeats", "aab", false},
		{"empty string", "", false},
		{"single char", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isRepeated(tt.password)
			if got != tt.want {
				t.Errorf("isRepeated(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestContainsCommonWords(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"contains password", "mypassword123", true},
		{"contains admin", "adminaccess", true},
		{"contains welcome", "welcome2023", true},
		{"no common words", "xzyq!@#$123", false},
		{"short words", "the cat", false}, // words < 3 chars are ignored
		{"mixed case", "MyPassword123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.containsCommonWords(strings.ToLower(tt.password))
			if got != tt.want {
				t.Errorf("containsCommonWords(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name            string
		password        string
		complexityScore int
		wantMin         int
		wantMax         int
	}{
		{"weak password", "weak", 10, 20, 40},
		{"moderate password", "Password123", 40, 60, 80},
		{"strong password", "StrongP@ssw0rd!", 55, 80, 100},
		{"very strong password", "VeryStr0ng!@#Passw0rd", 65, 90, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.calculateScore(tt.password, tt.complexityScore)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateScore(%q, %d) = %v, want between %d and %d",
					tt.password, tt.complexityScore, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateEntropy(t *testing.T) {
	validator := NewPasswordValidator(nil)

	tests := []struct {
		name     string
		password string
		wantMin  float64
	}{
		{"only lowercase", "abcdefgh", 37.0}, // 8 * log2(26)
		{"lowercase + uppercase", "abcdEFGH", 45.0}, // 8 * log2(52)
		{"alphanumeric", "abcd1234", 47.0}, // 8 * log2(62)
		{"all characters", "abC1!dE2", 51.0}, // 8 * log2(94)
		{"short password", "abc", 14.0}, // 3 * log2(26)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.calculateEntropy(tt.password)
			if got < tt.wantMin {
				t.Errorf("calculateEntropy(%q) = %v, want >= %v", tt.password, got, tt.wantMin)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name           string
		length         int
		includeUpper   bool
		includeLower   bool
		includeNumbers bool
		includeSymbols bool
		expectError    bool
	}{
		{"valid all types", 12, true, true, true, true, false},
		{"valid no symbols", 10, true, true, true, false, false},
		{"valid only lowercase", 8, false, true, false, false, false},
		{"too short", 3, true, true, true, true, true},
		{"no character types", 10, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePassword(tt.length, tt.includeUpper, tt.includeLower,
				tt.includeNumbers, tt.includeSymbols)

			if tt.expectError {
				if err == nil {
					t.Errorf("GeneratePassword() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GeneratePassword() unexpected error: %v", err)
				return
			}

			if len(got) != tt.length {
				t.Errorf("GeneratePassword() length = %v, want %v", len(got), tt.length)
			}

			// Verify character types
			if tt.includeUpper {
				if !regexp.MustCompile(`[A-Z]`).MatchString(got) {
					t.Errorf("GeneratePassword() should include uppercase letters")
				}
			}
			if tt.includeLower {
				if !regexp.MustCompile(`[a-z]`).MatchString(got) {
					t.Errorf("GeneratePassword() should include lowercase letters")
				}
			}
			if tt.includeNumbers {
				if !regexp.MustCompile(`[0-9]`).MatchString(got) {
					t.Errorf("GeneratePassword() should include numbers")
				}
			}
			if tt.includeSymbols {
				if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`).MatchString(got) {
					t.Errorf("GeneratePassword() should include symbols")
				}
			}
		})
	}
}

func TestEstimateStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"very weak", "abc", "Very Weak"},
		{"weak", "abcdef", "Weak"},
		{"moderate", "Password1", "Moderate"},
		{"strong", "Password123!", "Strong"},
		{"very strong", "VeryStr0ng!@#Passw0rd", "Very Strong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateStrength(tt.password)
			if got != tt.want {
				t.Errorf("EstimateStrength(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPasswordValidation(b *testing.B) {
	validator := NewPasswordValidator(nil)
	password := "StrongP@ssw0rd123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(password)
	}
}

func BenchmarkPasswordValidationWeak(b *testing.B) {
	validator := NewPasswordValidator(nil)
	password := "weak"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(password)
	}
}

func BenchmarkGeneratePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeneratePassword(16, true, true, true, true)
	}
}

// Example tests
func ExamplePasswordValidator_Validate() {
	validator := NewPasswordValidator(nil)
	result := validator.Validate("StrongP@ssw0rd123!")
	fmt.Printf("Valid: %v, Score: %d\n", result.Valid, result.Score)
	// Output: Valid: true, Score: 100
}

func ExampleGeneratePassword() {
	password, err := GeneratePassword(12, true, true, true, true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Generated password length: %d\n", len(password))
	// Output: Generated password length: 12
}

func ExampleEstimateStrength() {
	strength := EstimateStrength("StrongP@ssw0rd123!")
	fmt.Println("Password strength:", strength)
	// Output: Password strength: Very Strong
}

// Subtests for better organization
func TestPasswordValidationSubtests(t *testing.T) {
	validator := NewPasswordValidator(nil)

	t.Run("length validation", func(t *testing.T) {
		t.Run("empty password", func(t *testing.T) {
			result := validator.Validate("")
			if result.Valid {
				t.Error("Empty password should be invalid")
			}
		})

		t.Run("minimum length", func(t *testing.T) {
			result := validator.Validate("abcdefgh")
			if !result.Valid {
				t.Error("8 character password should be valid")
			}
		})
	})

	t.Run("complexity validation", func(t *testing.T) {
		t.Run("missing uppercase", func(t *testing.T) {
			result := validator.Validate("lowercase123!")
			if len(result.Suggestions) == 0 {
				t.Error("Should suggest adding uppercase letters")
			}
		})

		t.Run("all character types", func(t *testing.T) {
			result := validator.Validate("Mixed123!@#")
			if result.Score < 50 {
				t.Errorf("Complex password should have high score, got %d", result.Score)
			}
		})
	})
}

// Test helper functions
func TestLog2(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{1, 0},
		{2, 1},
		{4, 2},
		{8, 3},
		{16, 4},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("log2(%f)", tt.input), func(t *testing.T) {
			got := log2(tt.input)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("log2(%f) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Integration test with realistic scenarios
func TestRealWorldScenarios(t *testing.T) {
	validator := NewPasswordValidator(NewMockBreachedService([]string{
		"password123", "123456", "qwerty", "admin",
	}))

	scenarios := []struct {
		name        string
		password    string
		expectValid bool
		minScore    int
	}{
		{
			name:        "corporate password policy",
			password:    "MyC0mpany!2024",
			expectValid: true,
			minScore:    70,
		},
		{
			name:        "common but modified",
			password:    "Password123!",
			expectValid: false,
			minScore:    0,
		},
		{
			name:        "passphrase style",
			password:    "Correct-Horse-Battery-Staple",
			expectValid: true,
			minScore:    60,
		},
		{
			name:        "random looking",
			password:    "xK9@mQ7$pL2#nR5",
			expectValid: true,
			minScore:    90,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := validator.Validate(scenario.password)

			if result.Valid != scenario.expectValid {
				t.Errorf("Expected valid=%v, got %v for password %s",
					scenario.expectValid, result.Valid, scenario.password)
			}

			if result.Score < scenario.minScore {
				t.Errorf("Expected score >= %d, got %d for password %s",
					scenario.minScore, result.Score, scenario.password)
			}
		})
	}
}