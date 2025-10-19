# Exercise 4: JSON Configuration Validator

## 🎯 Objective
Create a CLI tool that validates JSON configuration files against predefined schemas and rules.

## 📋 Main Focus Areas
- **JSON parsing and manipulation** (`encoding/json` package)
- **Struct definitions and tags** for JSON mapping
- **Validation and error handling**
- **File I/O operations**
- **Schema validation patterns**

## 🔧 What You'll Build
A JSON validator that:
- Validates JSON syntax and structure
- Checks required fields and data types
- Validates value ranges and patterns
- Supports custom validation rules
- Generates detailed error reports
- Can fix common JSON issues

## 📝 Instructions

### Step 1: Basic JSON Validation
Create a tool that:
1. Reads JSON configuration files
2. Validates JSON syntax
3. Checks for required fields
4. Validates data types against expected types

### Step 2: Command-line Interface
Add support for these flags:
- `-c, --config`: Configuration file to validate
- `-s, --schema`: Schema file for validation rules
- `-f, --fix`: Attempt to fix common issues
- `-q, --quiet`: Suppress success messages
- `-o, --output`: Output format (text, json)

### Step 3: Schema Definition
Implement a schema system that supports:
- Required fields
- Data type validation (string, number, boolean, array, object)
- Value constraints (min/max length, ranges)
- Pattern matching (regex)
- Nested object validation
- Array validation (item types, length constraints)

### Step 4: Advanced Features
Add these capabilities:
- Multiple file validation
- Custom validation functions
- Environment variable substitution
- Default value injection
- Configuration merging
- Pretty-printing and formatting

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 04-json-validator

# Create a sample config file
cat > config.json << EOF
{
  "server": {
    "host": "localhost",
    "port": 8080,
    "ssl": true
  },
  "database": {
    "driver": "mysql",
    "connection": "user:pass@localhost/db"
  },
  "log_level": "info"
}
EOF

# Create your main.go file
touch main.go

# Run your solution
go run main.go -c config.json
```

## 💡 Implementation Tips

### JSON Struct Definitions
```go
type Config struct {
    Server   ServerConfig `json:"server" validate:"required"`
    Database DBConfig     `json:"database" validate:"required"`
    LogLevel string       `json:"log_level" validate:"required,oneof=debug info warn error"`
}

type ServerConfig struct {
    Host string `json:"host" validate:"required,hostname"`
    Port int    `json:"port" validate:"required,min=1,max=65535"`
    SSL  bool   `json:"ssl"`
}
```

### Custom Validation
```go
func validateConfig(config *Config) []error {
    var errors []error

    if config.Server.Port < 1 || config.Server.Port > 65535 {
        errors = append(errors, fmt.Errorf("invalid port: %d", config.Server.Port))
    }

    return errors
}
```

### JSON Parsing with Error Handling
```go
func parseConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("JSON parse error: %w", err)
    }

    return &config, nil
}
```

## 🧪 Test Cases
Create validation scenarios:
- Valid configurations
- Missing required fields
- Invalid data types
- Out-of-range values
- Malformed JSON
- Nested validation errors

## 📚 Go Concepts Covered
- `encoding/json` package for JSON operations
- Struct tags for JSON mapping
- Custom validation logic
- Error handling and reporting
- File I/O operations
- Interface-based design patterns
- Reflection for dynamic validation

## 🔗 Useful Resources
- [Go encoding/json package](https://pkg.go.dev/encoding/json)
- [JSON Schema specification](https://json-schema.org/)
- [Configuration management best practices](https://12factor.net/config)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Parse and validate JSON files correctly
- [ ] Detect and report validation errors clearly
- [ ] Support custom validation rules
- [ ] Handle nested JSON structures
- [ ] Provide helpful error messages
- [ ] Attempt to fix common issues when requested

## 🎁 Bonus Features
If you want an extra challenge:
- Support for YAML configuration files
- JSON Schema Draft 7 compliance
- Environment variable expansion
- Configuration file encryption
- Live configuration reloading
- Web API for validation

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.