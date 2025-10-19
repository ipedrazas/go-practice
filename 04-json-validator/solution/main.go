package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Config represents the main configuration structure
type Config struct {
	Server   ServerConfig   `json:"server" validate:"required"`
	Database DatabaseConfig `json:"database" validate:"required"`
	LogLevel string         `json:"log_level" validate:"required,oneof=debug info warn error"`
	Features []string       `json:"features,omitempty"`
}

type ServerConfig struct {
	Host            string `json:"host" validate:"required,hostname"`
	Port            int    `json:"port" validate:"required,min=1,max=65535"`
	SSL             bool   `json:"ssl"`
	Timeout         int    `json:"timeout,omitempty" validate:"min=1,max=300"`
	MaxConnections  int    `json:"max_connections,omitempty" validate:"min=1"`
}

type DatabaseConfig struct {
	Driver      string `json:"driver" validate:"required,oneof=mysql postgres sqlite"`
	Host        string `json:"host,omitempty" validate:"hostname"`
	Port        int    `json:"port,omitempty" validate:"min=1,max=65535"`
	Database    string `json:"database" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Connection  string `json:"connection,omitempty"`
	MaxOpenConn int    `json:"max_open_conns,omitempty" validate:"min=1"`
	MaxIdleConn int    `json:"max_idle_conns,omitempty" validate:"min=1"`
}

// Schema represents validation rules
type Schema struct {
	Required []string                 `json:"required"`
	Properties map[string]PropertySchema `json:"properties"`
}

type PropertySchema struct {
	Type        string             `json:"type"`
	Required    bool               `json:"required"`
	Min         *float64           `json:"min,omitempty"`
	Max         *float64           `json:"max,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	Enum        []string           `json:"enum,omitempty"`
	Properties  map[string]PropertySchema `json:"properties,omitempty"`
	Items       *PropertySchema    `json:"items,omitempty"`
}

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
	Warnings []string `json:"warnings,omitempty"`
}

func main() {
	var (
		configFile = flag.String("c", "", "Configuration file to validate (required)")
		schemaFile = flag.String("s", "", "Schema file for validation rules")
		fix        = flag.Bool("f", false, "Attempt to fix common issues")
		quiet      = flag.Bool("q", false, "Suppress success messages")
		output     = flag.String("o", "text", "Output format (text, json)")
		help       = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Validate JSON configuration files against schemas.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -c config.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c config.json -s schema.json -f\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c config.json -o json\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Configuration file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	validator := &JSONValidator{
		ConfigFile: *configFile,
		SchemaFile: *schemaFile,
		Fix:        *fix,
		Quiet:      *quiet,
		Output:     *output,
	}

	if err := validator.Validate(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
}

type JSONValidator struct {
	ConfigFile string
	SchemaFile string
	Fix        bool
	Quiet      bool
	Output     string
}

func (jv *JSONValidator) Validate() error {
	// Read and parse configuration
	config, err := jv.parseConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Load schema if provided
	var schema *Schema
	if jv.SchemaFile != "" {
		schema, err = jv.loadSchema()
		if err != nil {
			return fmt.Errorf("failed to load schema: %w", err)
		}
	}

	// Validate configuration
	result := jv.validateConfig(config, schema)

	// Attempt to fix issues if requested
	if jv.Fix && !result.Valid {
		if err := jv.fixConfig(config, schema); err != nil {
			return fmt.Errorf("failed to fix config: %w", err)
		}
		// Re-validate after fixing
		result = jv.validateConfig(config, schema)
	}

	// Output results
	return jv.outputResults(result)
}

func (jv *JSONValidator) parseConfig() (*Config, error) {
	data, err := os.ReadFile(jv.ConfigFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return &config, nil
}

func (jv *JSONValidator) loadSchema() (*Schema, error) {
	data, err := os.ReadFile(jv.SchemaFile)
	if err != nil {
		return nil, err
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("schema parse error: %w", err)
	}

	return &schema, nil
}

func (jv *JSONValidator) validateConfig(config *Config, schema *Schema) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Basic struct validation using reflection
	jv.validateStruct(reflect.ValueOf(config).Elem(), "", schema, &result)

	// Custom validation rules
	jv.customValidation(config, &result)

	result.Valid = len(result.Errors) == 0
	return result
}

func (jv *JSONValidator) validateStruct(v reflect.Value, prefix string, schema *Schema, result *ValidationResult) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get JSON tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag (handle omitempty)
		jsonName := strings.Split(jsonTag, ",")[0]
		fieldName := prefix
		if fieldName != "" {
			fieldName += "."
		}
		fieldName += jsonName

		// Check if field is required
		validateTag := fieldType.Tag.Get("validate")
		if strings.Contains(validateTag, "required") {
			if jv.isZeroValue(field) {
				result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", fieldName))
				continue
			}
		}

		// Skip zero values for optional fields
		if jv.isZeroValue(field) {
			continue
		}

		// Validate field based on type
		jv.validateField(field, fieldName, validateTag, schema, result)
	}
}

func (jv *JSONValidator) validateField(field reflect.Value, fieldName, validateTag string, schema *Schema, result *ValidationResult) {
	switch field.Kind() {
	case reflect.String:
		jv.validateStringField(field.String(), fieldName, validateTag, result)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		jv.validateIntField(field.Int(), fieldName, validateTag, result)
	case reflect.Bool:
		// Boolean values don't need additional validation
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			jv.validateStringSlice(field, fieldName, validateTag, result)
		}
	case reflect.Struct:
		jv.validateStruct(field, fieldName, schema, result)
	}
}

func (jv *JSONValidator) validateStringField(value, fieldName, validateTag string, result *ValidationResult) {
	rules := strings.Split(validateTag, ",")

	for _, rule := range rules {
		if strings.HasPrefix(rule, "min=") {
			minLength, _ := strconv.Atoi(strings.TrimPrefix(rule, "min="))
			if len(value) < minLength {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: minimum length is %d, got %d", fieldName, minLength, len(value)))
			}
		}

		if strings.HasPrefix(rule, "max=") {
			maxLength, _ := strconv.Atoi(strings.TrimPrefix(rule, "max="))
			if len(value) > maxLength {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: maximum length is %d, got %d", fieldName, maxLength, len(value)))
			}
		}

		if strings.HasPrefix(rule, "oneof=") {
			options := strings.Split(strings.TrimPrefix(rule, "oneof="), " ")
			valid := false
			for _, option := range options {
				if value == option {
					valid = true
					break
				}
			}
			if !valid {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: must be one of %v, got %s", fieldName, options, value))
			}
		}

		if rule == "hostname" {
			if !jv.isValidHostname(value) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: invalid hostname format", fieldName))
			}
		}
	}
}

func (jv *JSONValidator) validateIntField(value int64, fieldName, validateTag string, result *ValidationResult) {
	rules := strings.Split(validateTag, ",")

	for _, rule := range rules {
		if strings.HasPrefix(rule, "min=") {
			minValue, _ := strconv.ParseInt(strings.TrimPrefix(rule, "min="), 10, 64)
			if value < minValue {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: minimum value is %d, got %d", fieldName, minValue, value))
			}
		}

		if strings.HasPrefix(rule, "max=") {
			maxValue, _ := strconv.ParseInt(strings.TrimPrefix(rule, "max="), 10, 64)
			if value > maxValue {
				result.Errors = append(result.Errors,
					fmt.Sprintf("field %s: maximum value is %d, got %d", fieldName, maxValue, value))
			}
		}
	}
}

func (jv *JSONValidator) validateStringSlice(field reflect.Value, fieldName, validateTag string, result *ValidationResult) {
	// Validate string array elements
	for i := 0; i < field.Len(); i++ {
		element := field.Index(i).String()
		elementName := fmt.Sprintf("%s[%d]", fieldName, i)
		jv.validateStringField(element, elementName, validateTag, result)
	}
}

func (jv *JSONValidator) customValidation(config *Config, result *ValidationResult) {
	// Validate database configuration
	if config.Database.Driver == "sqlite" && config.Database.Host != "" {
		result.Warnings = append(result.Warnings,
			"SQLite driver doesn't use host field")
	}

	// Validate server configuration
	if config.Server.SSL && config.Server.Port == 80 {
		result.Warnings = append(result.Warnings,
			"SSL enabled but using port 80 (typically for HTTP)")
	}

	if !config.Server.SSL && config.Server.Port == 443 {
		result.Warnings = append(result.Warnings,
			"SSL disabled but using port 443 (typically for HTTPS)")
	}

	// Validate database connection vs individual settings
	if config.Database.Connection != "" {
		if config.Database.Host != "" || config.Database.Username != "" {
			result.Warnings = append(result.Warnings,
				"Both connection string and individual database settings provided")
		}
	}
}

func (jv *JSONValidator) fixConfig(config *Config, schema *Schema) error {
	// Apply default values
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30
	}

	if config.Server.MaxConnections == 0 {
		config.Server.MaxConnections = 100
	}

	if config.Database.MaxOpenConn == 0 {
		config.Database.MaxOpenConn = 25
	}

	if config.Database.MaxIdleConn == 0 {
		config.Database.MaxIdleConn = 5
	}

	// Set default database port based on driver
	if config.Database.Port == 0 {
		switch config.Database.Driver {
		case "mysql":
			config.Database.Port = 3306
		case "postgres":
			config.Database.Port = 5432
		case "sqlite":
			config.Database.Port = 0 // SQLite doesn't use ports
		}
	}

	// Set default database host if not provided
	if config.Database.Host == "" && config.Database.Driver != "sqlite" {
		config.Database.Host = "localhost"
	}

	// Save fixed configuration
	return jv.saveConfig(config)
}

func (jv *JSONValidator) saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(jv.ConfigFile, data, 0644)
}

func (jv *JSONValidator) outputResults(result ValidationResult) error {
	switch jv.Output {
	case "json":
		return jv.outputJSON(result)
	default:
		return jv.outputText(result)
	}
}

func (jv *JSONValidator) outputText(result ValidationResult) error {
	fmt.Printf("Configuration Validation Results\n")
	fmt.Printf("===============================\n\n")

	if result.Valid {
		fmt.Printf("✅ Configuration is valid!\n")
		if len(result.Warnings) > 0 {
			fmt.Printf("\n⚠️  Warnings:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("  %s\n", warning)
			}
		}
	} else {
		fmt.Printf("❌ Configuration has errors:\n\n")
		for _, error := range result.Errors {
			fmt.Printf("  • %s\n", error)
		}
		fmt.Printf("\n%d error(s) found\n", len(result.Errors))
	}

	if jv.Fix && !result.Valid {
		fmt.Printf("\n🔧 Fix mode attempted - some issues may have been resolved\n")
	}

	return nil
}

func (jv *JSONValidator) outputJSON(result ValidationResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (jv *JSONValidator) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func (jv *JSONValidator) isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Simple hostname validation regex
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(hostname)
}