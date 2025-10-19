# Exercise 7: Simple Web Server with Templates

## 🎯 Objective
Create a web server that serves dynamic content using HTML templates and static files.

## 📋 Main Focus Areas
- **HTTP server programming** (`net/http` package)
- **HTML templates** (`html/template` package)
- **Static file serving**
- **HTTP routing and handlers**
- **Request handling and form processing**

## 🔧 What You'll Build
A web server that:
- Serves static files (CSS, JS, images)
- Renders HTML templates with dynamic data
- Handles different URL routes
- Processes form submissions
- Provides a simple REST API
- Includes basic middleware

## 📝 Instructions

### Step 1: Basic HTTP Server
Create a server that:
1. Listens on a configurable port
2. Serves a simple homepage
3. Handles basic routing
4. Returns proper HTTP status codes

### Step 2: Template System
Add template support:
- Create HTML templates with dynamic content
- Pass data structures to templates
- Handle template inheritance
- Include partials and components

### Step 3: Static Files
Implement static file serving:
- CSS and JavaScript files
- Images and other assets
- Proper MIME types
- Cache control headers

### Step 4: Advanced Features
Add these capabilities:
- Form handling and validation
- JSON API endpoints
- Basic authentication middleware
- Logging middleware
- Error handling pages
- Graceful shutdown

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 07-web-server

# Create directory structure
mkdir -p templates static/css static/js

# Create your main.go file
touch main.go

# Create templates
echo '<html><body><h1>{{.Title}}</h1><p>{{.Content}}</p></body></html>' > templates/index.html

# Run your server
go run main.go

# Visit http://localhost:8080
```

## 💡 Implementation Tips

### Basic HTTP Server
```go
func main() {
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/about", aboutHandler)

    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the homepage!")
}
```

### Template Rendering
```go
func renderTemplate(w http.ResponseWriter, template string, data interface{}) {
    t, err := template.ParseFiles("templates/" + template)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    t.Execute(w, data)
}
```

### Static File Serving
```go
fs := http.FileServer(http.Dir("static/"))
http.Handle("/static/", http.StripPrefix("/static/", fs))
```

## 🧪 Test Cases
Test different aspects:
- Page rendering
- Static file serving
- Form submissions
- Error handling
- API endpoints
- Concurrent requests

## 📚 Go Concepts Covered
- `net/http` package for web servers
- `html/template` for secure template rendering
- HTTP routing and handlers
- Middleware patterns
- Static file serving
- Request/response processing

## ✅ Success Criteria
Your server should be able to:
- [ ] Serve HTML pages with dynamic content
- [ ] Handle static files correctly
- [ ] Process form submissions
- [ ] Provide JSON API endpoints
- [ ] Handle errors gracefully
- [ ] Be concurrent-safe

## 🎁 Bonus Features
If you want an extra challenge:
- Database integration
- User authentication
- Session management
- WebSocket support
- File upload handling
- Rate limiting
- Metrics collection

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.