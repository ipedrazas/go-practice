// main.go provides a CLI interface for the password validation library
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"password-validator"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Password Validation Tool ===")
	fmt.Println("This tool demonstrates the password validation library.")
	fmt.Println("Type 'exit' to quit, 'help' for commands.")
	fmt.Println()

	for {
		fmt.Print("Enter a password to validate (or command): ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return
		case "help":
			showHelp()
		case "generate":
			generatePasswordInteractive(scanner)
		case "demo":
			runDemo()
		default:
			validatePassword(input)
		}
		fmt.Println()
	}
}

func validatePassword(pwd string) {
	validator := password.NewPasswordValidator(nil)
	result := validator.Validate(pwd)

	fmt.Printf("Password: %q\n", pwd)
	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Score: %d/100\n", result.Score)
	fmt.Printf("Strength: %s\n", password.EstimateStrength(pwd))

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("Suggestions:")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}

	if result.IsBreached {
		fmt.Println("⚠️  This password has been found in data breaches!")
	}
}

func generatePasswordInteractive(scanner *bufio.Scanner) {
	fmt.Println("Password Generation")
	fmt.Print("Length (default 16): ")
	scanner.Scan()
	lengthStr := strings.TrimSpace(scanner.Text())

	length := 16
	if lengthStr != "" {
		fmt.Sscanf(lengthStr, "%d", &length)
	}

	fmt.Print("Include uppercase? (y/n, default y): ")
	scanner.Scan()
	upper := strings.ToLower(strings.TrimSpace(scanner.Text())) != "n"

	fmt.Print("Include lowercase? (y/n, default y): ")
	scanner.Scan()
	lower := strings.ToLower(strings.TrimSpace(scanner.Text())) != "n"

	fmt.Print("Include numbers? (y/n, default y): ")
	scanner.Scan()
	numbers := strings.ToLower(strings.TrimSpace(scanner.Text())) != "n"

	fmt.Print("Include symbols? (y/n, default y): ")
	scanner.Scan()
	symbols := strings.ToLower(strings.TrimSpace(scanner.Text())) != "n"

	pwd, err := password.GeneratePassword(length, upper, lower, numbers, symbols)
	if err != nil {
		fmt.Printf("Error generating password: %v\n", err)
		return
	}

	fmt.Printf("Generated password: %s\n", pwd)
	validatePassword(pwd)
}

func runDemo() {
	fmt.Println("=== Password Validation Demo ===")

	testPasswords := []string{
		"weak",
		"password123",
		"StrongP@ssw0rd123!",
		"Correct-Horse-Battery-Staple",
		"xK9@mQ7$pL2#nR5",
	}

	for _, pwd := range testPasswords {
		fmt.Printf("\n--- Testing: %q ---\n", pwd)
		validatePassword(pwd)
	}

	fmt.Println("\n=== Password Generation Demo ===")

	generated, _ := password.GeneratePassword(16, true, true, true, true)
	fmt.Printf("Generated password: %s\n", generated)
	validatePassword(generated)
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  <password>  - Validate a password")
	fmt.Println("  generate    - Generate a random password")
	fmt.Println("  demo        - Run demonstration")
	fmt.Println("  help        - Show this help")
	fmt.Println("  exit/quit   - Exit the program")
}