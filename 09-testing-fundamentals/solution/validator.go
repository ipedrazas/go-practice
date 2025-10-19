package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// ValidationResult represents the result of password validation
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	Score       int      `json:"score"`
	Errors      []string `json:"errors"`
	Suggestions []string `json:"suggestions"`
	IsBreached  bool     `json:"is_breached"`
}

// BreachedService interface for checking breached passwords
type BreachedService interface {
	IsBreached(password string) bool
}

// PasswordValidator handles password validation
type PasswordValidator struct {
	minLength    int
	maxLength    int
	breachedSvc  BreachedService
	commonWords  map[string]bool
}

// NewPasswordValidator creates a new password validator
func NewPasswordValidator(breachedSvc BreachedService) *PasswordValidator {
	return &PasswordValidator{
		minLength:   8,
		maxLength:   128,
		breachedSvc: breachedSvc,
		commonWords: loadCommonWords(),
	}
}

// Validate performs comprehensive password validation
func (pv *PasswordValidator) Validate(password string) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []string{},
		Suggestions: []string{},
		IsBreached:  false,
	}

	// Check length
	if !pv.isValidLength(password) {
		result.Valid = false
		if len(password) < pv.minLength {
			result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters", pv.minLength))
			result.Suggestions = append(result.Suggestions, "Add more characters")
		}
		if len(password) > pv.maxLength {
			result.Errors = append(result.Errors, fmt.Sprintf("Password must be no more than %d characters", pv.maxLength))
			result.Suggestions = append(result.Suggestions, "Use a shorter password")
		}
	}

	// Check character complexity
	complexityScore := pv.checkComplexity(password, result)

	// Check for common patterns
	if pv.hasCommonPatterns(password) {
		result.Valid = false
		result.Errors = append(result.Errors, "Password contains common patterns")
		result.Suggestions = append(result.Suggestions, "Avoid common patterns and sequences")
	}

	// Check against common words
	if pv.containsCommonWords(password) {
		result.Valid = false
		result.Errors = append(result.Errors, "Password contains common words")
		result.Suggestions = append(result.Suggestions, "Avoid dictionary words")
	}

	// Check if breached
	if pv.breachedSvc != nil && pv.breachedSvc.IsBreached(password) {
		result.IsBreached = true
		result.Valid = false
		result.Errors = append(result.Errors, "Password has been found in data breaches")
		result.Suggestions = append(result.Suggestions, "Choose a unique password")
	}

	// Calculate overall score
	result.Score = pv.calculateScore(password, complexityScore)

	// Final validation based on score
	if result.Score < 60 {
		result.Valid = false
		if len(result.Errors) == 0 {
			result.Errors = append(result.Errors, "Password is too weak")
		}
	}

	return result
}

// isValidLength checks if password length is within acceptable range
func (pv *PasswordValidator) isValidLength(password string) bool {
	length := len(password)
	return length >= pv.minLength && length <= pv.maxLength
}

// checkComplexity evaluates password character complexity
func (pv *PasswordValidator) checkComplexity(password string, result *ValidationResult) int {
	score := 0
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
			score += 10
		case unicode.IsLower(char):
			hasLower = true
			score += 10
		case unicode.IsNumber(char):
			hasNumber = true
			score += 10
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
			score += 15
		}
	}

	// Add complexity bonus
	if hasUpper && hasLower && hasNumber && hasSpecial {
		score += 20
	}

	// Add specific suggestions
	if !hasUpper {
		result.Suggestions = append(result.Suggestions, "Add uppercase letters")
	}
	if !hasLower {
		result.Suggestions = append(result.Suggestions, "Add lowercase letters")
	}
	if !hasNumber {
		result.Suggestions = append(result.Suggestions, "Add numbers")
	}
	if !hasSpecial {
		result.Suggestions = append(result.Suggestions, "Add special characters")
	}

	return score
}

// hasCommonPatterns checks for common password patterns
func (pv *PasswordValidator) hasCommonPatterns(password string) bool {
	lowerPassword := strings.ToLower(password)

	// Check for sequential characters
	if pv.isSequential(lowerPassword) {
		return true
	}

	// Check for repeated characters
	if pv.isRepeated(lowerPassword) {
		return true
	}

	// Check for keyboard patterns
	if pv.isKeyboardPattern(lowerPassword) {
		return true
	}

	return false
}

// isSequential checks for sequential characters
func (pv *PasswordValidator) isSequential(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
	}
	return false
}

// isRepeated checks for repeated characters
func (pv *PasswordValidator) isRepeated(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

// isKeyboardPattern checks for keyboard patterns
func (pv *PasswordValidator) isKeyboardPattern(password string) bool {
	keyboardPatterns := []string{
		"qwerty", "asdf", "zxcv", "1234", "qwertyuiop",
		"asdfghjkl", "zxcvbnm", "qwertyuiopasdfghjklzxcvbnm",
	}

	for _, pattern := range keyboardPatterns {
		if strings.Contains(password, pattern) {
			return true
		}
	}
	return false
}

// containsCommonWords checks if password contains common words
func (pv *PasswordValidator) containsCommonWords(password string) bool {
	lowerPassword := strings.ToLower(password)

	// Split password into possible words
	words := pv.extractWords(lowerPassword)

	for _, word := range words {
		if pv.commonWords[word] {
			return true
		}
	}

	return false
}

// extractWords extracts possible words from password
func (pv *PasswordValidator) extractWords(password string) []string {
	var words []string
	currentWord := ""

	for _, char := range password {
		if unicode.IsLetter(char) {
			currentWord += string(char)
		} else {
			if len(currentWord) >= 3 {
				words = append(words, currentWord)
			}
			currentWord = ""
		}
	}

	if len(currentWord) >= 3 {
		words = append(words, currentWord)
	}

	return words
}

// calculateScore calculates overall password strength score
func (pv *PasswordValidator) calculateScore(password string, complexityScore int) int {
	score := 0

	// Length contribution
	length := len(password)
	if length >= 8 {
		score += 20
	}
	if length >= 12 {
		score += 20
	}
	if length >= 16 {
		score += 10
	}

	// Complexity contribution
	score += complexityScore

	// Entropy contribution
	entropy := pv.calculateEntropy(password)
	score += int(entropy)

	// Cap the score at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateEntropy calculates password entropy
func (pv *PasswordValidator) calculateEntropy(password string) float64 {
	charSets := 0
	hasLower := false
	hasUpper := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasLower {
		charSets += 26
	}
	if hasUpper {
		charSets += 26
	}
	if hasNumber {
		charSets += 10
	}
	if hasSpecial {
		charSets += 32
	}

	if charSets == 0 {
		return 0
	}

	entropy := float64(len(password)) * log2(float64(charSets))
	return entropy
}

// log2 calculates base-2 logarithm
func log2(x float64) float64 {
	return 0.693147180559945309 * log(x) // ln(2) * ln(x)
}

// log calculates natural logarithm (approximation)
func log(x float64) float64 {
	// Simple Taylor series approximation for ln(1+x)
	if x <= 0 {
		return 0
	}
	x -= 1
	result := 0.0
	term := x
	for i := 1; i <= 20; i++ {
		if i%2 == 1 {
			result += term / float64(i)
		} else {
			result -= term / float64(i)
		}
		term *= x
	}
	return result
}

// loadCommonWords loads a set of common words
func loadCommonWords() map[string]bool {
	words := map[string]bool{
		"password": true, "123456": true, "123456789": true, "12345678": true,
		"12345": true, "1234567": true, "1234567890": true, "qwerty": true,
		"abc123": true, "password123": true, "admin": true, "letmein": true,
		"welcome": true, "monkey": true, "login": true, "dragon": true,
		"master": true, "hello": true, "freedom": true, "whatever": true,
		"qazwsx": true, "trustno1": true, "123qwe": true, "1q2w3e4r": true,
		"zxcvbnm": true, "123abc": true, "password1": true, "iloveyou": true,
	}

	return words
}

// GeneratePassword generates a random password with specified criteria
func GeneratePassword(length int, includeUpper, includeLower, includeNumbers, includeSymbols bool) (string, error) {
	if length < 4 {
		return "", fmt.Errorf("password length must be at least 4 characters")
	}

	var charset string
	if includeLower {
		charset += "abcdefghijklmnopqrstuvwxyz"
	}
	if includeUpper {
		charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if includeNumbers {
		charset += "0123456789"
	}
	if includeSymbols {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	if charset == "" {
		return "", fmt.Errorf("at least one character type must be selected")
	}

	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

// EstimateStrength provides a quick strength estimate
func EstimateStrength(password string) string {
	validator := NewPasswordValidator(nil)
	result := validator.Validate(password)

	switch {
	case result.Score >= 80:
		return "Very Strong"
	case result.Score >= 60:
		return "Strong"
	case result.Score >= 40:
		return "Moderate"
	case result.Score >= 20:
		return "Weak"
	default:
		return "Very Weak"
	}
}