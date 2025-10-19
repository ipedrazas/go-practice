package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

type ExerciseMetadata struct {
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Focus       string `json:"focus"`
	Path        string `json:"path"`
	HasSolution bool   `json:"has_solution"`
}

type IndexData struct {
	Exercises      []ExerciseMetadata
	Introduction   string
	GettingStarted string
	TotalCount     int
	LastUpdated    string
}

func main() {
	var (
		rootDir = flag.String("d", "..", "Root directory containing exercises")
		output  = flag.String("o", "../README.md", "Output file for generated index")
		preview = flag.Bool("p", false, "Preview output without writing to file")
		verbose = flag.Bool("v", false, "Verbose output")
		help    = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate README.md index from exercise directories.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                    # Generate index for parent directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d . -o README.md  # Generate in current directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -p                 # Preview without writing\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	generator := &IndexGenerator{
		RootDir: *rootDir,
		Output:  *output,
		Preview: *preview,
		Verbose: *verbose,
	}

	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate index: %v", err)
	}
}

type IndexGenerator struct {
	RootDir string
	Output  string
	Preview bool
	Verbose bool
}

func (ig *IndexGenerator) Generate() error {
	if ig.Verbose {
		fmt.Printf("Scanning for exercises in: %s\n", ig.RootDir)
	}

	// Find exercise directories
	exerciseDirs, err := ig.findExerciseDirectories()
	if err != nil {
		return fmt.Errorf("failed to find exercise directories: %w", err)
	}

	if ig.Verbose {
		fmt.Printf("Found %d exercise directories\n", len(exerciseDirs))
	}

	// Parse exercise metadata
	var exercises []ExerciseMetadata
	for _, dir := range exerciseDirs {
		metadata, err := ig.parseExercise(dir)
		if err != nil {
			if ig.Verbose {
				log.Printf("Warning: Failed to parse exercise %s: %v", dir, err)
			}
			continue
		}
		exercises = append(exercises, metadata)
	}

	// Sort exercises by number
	sort.Slice(exercises, func(i, j int) bool {
		return exercises[i].Number < exercises[j].Number
	})

	if ig.Verbose {
		fmt.Printf("Successfully parsed %d exercises\n", len(exercises))
	}

	// Create index data
	data := IndexData{
		Exercises:      exercises,
		Introduction:   ig.getIntroduction(),
		GettingStarted: ig.getGettingStarted(),
		TotalCount:     len(exercises),
		LastUpdated:    ig.getCurrentTime(),
	}

	// Generate content
	content, err := ig.generateIndex(data)
	if err != nil {
		return fmt.Errorf("failed to generate index content: %w", err)
	}

	// Write or preview output
	if ig.Preview {
		fmt.Println("\n=== Generated Index Preview ===")
		fmt.Println(content)
		fmt.Println("=== End Preview ===")
	} else {
		if err := ig.writeOutput(content); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("Successfully generated index with %d exercises\n", len(exercises))
		fmt.Printf("Output written to: %s\n", ig.Output)
	}

	return nil
}

func (ig *IndexGenerator) findExerciseDirectories() ([]string, error) {
	var exercises []string

	entries, err := os.ReadDir(ig.RootDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if isExerciseDirectory(entry.Name()) {
			exercises = append(exercises, entry.Name())
		}
	}

	return exercises, nil
}

func isExerciseDirectory(name string) bool {
	// Match pattern like "01-exercise-name"
	matched, _ := regexp.MatchString(`^\d{2}-.+`, name)
	return matched
}

func (ig *IndexGenerator) parseExercise(dirName string) (ExerciseMetadata, error) {
	var metadata ExerciseMetadata

	// Extract number and name from directory name
	parts := strings.SplitN(dirName, "-", 2)
	if len(parts) < 2 {
		return metadata, fmt.Errorf("invalid directory name format: %s", dirName)
	}

	number, err := fmt.Sscanf(parts[0], "%d", &metadata.Number)
	if number != 1 || err != nil {
		return metadata, fmt.Errorf("invalid exercise number in directory: %s", dirName)
	}

	metadata.Name = strings.ReplaceAll(parts[1], "-", " ")
	metadata.Path = dirName

	// Check if solution exists
	solutionPath := filepath.Join(ig.RootDir, dirName, "solution", "main.go")
	if _, err := os.Stat(solutionPath); err == nil {
		metadata.HasSolution = true
	}

	// Parse README.md
	readmePath := filepath.Join(ig.RootDir, dirName, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return metadata, fmt.Errorf("failed to read README.md: %w", err)
	}

	return ig.parseReadmeContent(string(content), metadata)
}

func (ig *IndexGenerator) parseReadmeContent(content string, metadata ExerciseMetadata) (ExerciseMetadata, error) {
	lines := strings.Split(content, "\n")

	var inMainFocus bool
	var focusParts []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract title (first h1)
		if strings.HasPrefix(line, "# ") && metadata.Title == "" {
			tokens := strings.Split(line, ":")
			metadata.Title = strings.TrimSpace(tokens[1])
		}

		// Extract objective (first paragraph after title)
		if metadata.Title != "" && metadata.Description == "" && line != "" && !strings.HasPrefix(line, "#") {
			metadata.Description = line
		}

		// Extract main focus areas
		if strings.Contains(line, "## 📋 Main Focus Areas") {
			inMainFocus = true
			continue
		}

		if inMainFocus && strings.HasPrefix(line, "##") {
			inMainFocus = false
			continue
		}

		if inMainFocus && strings.HasPrefix(line, "- **") {
			// Extract focus area
			focusRegex := regexp.MustCompile(`- \*\*([^*]+)\*\*`)
			matches := focusRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				focusParts = append(focusParts, matches[1])
			}
		}
	}

	metadata.Focus = strings.Join(focusParts, ", ")

	// Fallback title if not found in README
	if metadata.Title == "" {
		metadata.Title = fmt.Sprintf("%s", metadata.Name)
	}

	return metadata, nil
}

func (ig *IndexGenerator) generateIndex(data IndexData) (string, error) {
	tmpl := `# Go Practice: Real-World CLI Exercises

{{.Introduction}}

## 🎯 Learning Philosophy

Instead of abstract examples and toy functions, these exercises build complete, useful CLI applications. You'll learn Go by solving real problems and creating tools that have genuine utility.

## 📚 Exercise Syllabus

| Exercise | Focus | Description |
|----------|-------|-------------|
{{range .Exercises -}}
| [{{.Number}}. {{.Title}}](./{{.Path}}/) | {{.Focus}} | {{.Description}} |
{{end -}}

## 🚀 Getting Started

{{.GettingStarted}}

## 📖 How to Use These Exercises

1. **Follow the Order**: Exercises build on previous concepts
2. **Read the Instructions First**: Understand what you're building
3. **Try It Yourself**: Write the code before looking at solutions
4. **Experiment**: Modify the tools to add your own features
5. **Build on Them**: Combine tools or extend them for new use cases

## 🛠 Prerequisites

- Go 1.19 or later installed
- Basic understanding of programming concepts
- Text editor or IDE
- Terminal/command line

## 📝 Tips for Success

- **Read Error Messages**: Go's error messages are helpful
- **Use the Standard Library**: Avoid external packages unless specified
- **Test as You Go**: Run your code frequently to catch issues early
- **Read the Docs**: When stuck, check the Go documentation for the relevant package

## 🤝 Contributing

Found a bug? Want to add an exercise? Contributions are welcome!

---

Ready to start? Jump into [Exercise 1: URL Downloader](./01-url-downloader/) and begin your Go practice journey!

---
*Index generated on {{.LastUpdated}} • {{.TotalCount}} exercises*`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (ig *IndexGenerator) getIntroduction() string {
	return `Welcome to Go Practice! This collection of hands-on exercises focuses on building practical CLI tools using Go's standard library. Each exercise is designed to teach you real-world skills while creating useful tools you might actually use.`
}

func (ig *IndexGenerator) getGettingStarted() string {
	return `Each exercise follows this structure:
- **Folder**: Contains the complete exercise in its own directory
- **Instructions**: Step-by-step guide to build the tool
- **Focus Areas**: Key Go concepts and standard library packages you'll learn
- **Solution**: Complete implementation for reference`
}

func (ig *IndexGenerator) getCurrentTime() string {
	return fmt.Sprintf("%d-%02d-%02d",
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day())
}

func (ig *IndexGenerator) writeOutput(content string) error {
	// Create backup of existing README if it exists
	if _, err := os.Stat(ig.Output); err == nil {
		backupPath := ig.Output + ".backup"
		if err := os.Rename(ig.Output, backupPath); err != nil {
			if ig.Verbose {
				log.Printf("Warning: Could not create backup: %v", err)
			}
		} else if ig.Verbose {
			fmt.Printf("Created backup: %s\n", backupPath)
		}
	}

	// Write new content
	return os.WriteFile(ig.Output, []byte(content), 0644)
}
