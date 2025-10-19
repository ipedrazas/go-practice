# Exercise 8: Exercise Index Generator Tool

## 🎯 Objective
Create a tool that automatically generates the main README.md index by scanning all exercise directories and extracting their metadata.

## 📋 Main Focus Areas
- **Directory traversal and analysis**
- **Markdown generation**
- **File parsing and content extraction**
- **Template-based content generation**
- **Metadata extraction from structured content**

## 🔧 What You'll Build
An index generator that:
- Scans directories for exercises
- Extracts exercise metadata from README files
- Generates a formatted index with links
- Updates the main README.md automatically
- Validates exercise structure
- Maintains consistent formatting

## 📝 Instructions

### Step 1: Directory Scanner
Create a tool that:
1. Scans the current directory for exercise folders
2. Identifies exercise directories (numbered folders)
3. Reads README.md files from each exercise
4. Extracts metadata (title, description, focus areas)

### Step 2: Content Parser
Parse README files to extract:
- Exercise title and objective
- Main focus areas
- Description snippets
- Prerequisites if any
- Difficulty or estimated time

### Step 3: Markdown Generator
Generate the main README.md with:
- Project introduction
- Exercise syllabus table
- Links to individual exercises
- Getting started guide
- Progress tracking section

### Step 4: Advanced Features
Add these capabilities:
- Exercise validation (check required files)
- Category grouping of exercises
- Progress tracking (completed/incomplete)
- Dependency detection between exercises
- Automatic changelog generation

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 08-index-generator

# Create your main.go file
touch main.go

# Run the tool to regenerate the main index
go run main.go

# Check the generated README.md
cat ../README.md
```

## 💡 Implementation Tips

### Directory Scanning
```go
func findExerciseDirectories(root string) ([]string, error) {
    var exercises []string

    entries, err := os.ReadDir(root)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        if entry.IsDir() && isExerciseDirectory(entry.Name()) {
            exercises = append(exercises, entry.Name())
        }
    }

    sort.Strings(exercises)
    return exercises, nil
}

func isExerciseDirectory(name string) bool {
    // Match pattern like "01-exercise-name"
    matched, _ := regexp.MatchString(`^\d{2}-.+`, name)
    return matched
}
```

### README Parsing
```go
type ExerciseMetadata struct {
    Number      int
    Name        string
    Title       string
    Description string
    Focus       []string
    Path        string
}

func parseExerciseReadme(path string) (*ExerciseMetadata, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    metadata := &ExerciseMetadata{Path: path}

    // Extract title from first h1
    titleRegex := regexp.MustCompile(`^# (.+)$`)
    // Extract focus areas
    focusRegex := regexp.MustCompile(`- \*\*(.+)\*\*`)

    // Parse content...

    return metadata, nil
}
```

### Template Generation
```go
const indexTemplate = `# Go Practice: Real-World CLI Exercises

{{.Introduction}}

## 📚 Exercise Syllabus

{{range .Exercises}}
| [{{.Number}}. {{.Title}}](./{{.Path}}/) | {{.Focus}} | {{.Description}} |
{{end}}

## 🚀 Getting Started

{{.GettingStarted}}
`
```

## 🧪 Test Cases
Test with different scenarios:
- New exercises added
- Missing README files
- Malformed README content
- Different exercise numbering
- Special characters in names

## 📚 Go Concepts Covered
- Regular expressions for parsing
- Template engines for generation
- File system operations
- String manipulation and processing
- Data structure organization
- Error handling for malformed input

## ✅ Success Criteria
Your tool should be able to:
- [ ] Automatically discover all exercise directories
- [ ] Extract metadata from README files
- [ ] Generate a well-formatted index
- [ ] Handle missing or malformed files gracefully
- [ ] Maintain consistent formatting
- [ ] Update the main README.md automatically

## 🎁 Bonus Features
If you want an extra challenge:
- Support for multiple output formats
- Exercise dependency graph generation
- Integration with git to track changes
- Web-based exercise browser
- Exercise statistics and analytics
- Automatic validation of exercise completeness

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.

## 🔄 Meta Note
This exercise is special - it's the tool that helps maintain this entire course! After completing it, you can run it to keep the main README.md synchronized with any new exercises you add.