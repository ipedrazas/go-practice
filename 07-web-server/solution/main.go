package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PageData struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Posts       []BlogPost `json:"posts,omitempty"`
	CurrentTime time.Time  `json:"current_time"`
	RequestID   string     `json:"request_id"`
}

type BlogPost struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
}

type Server struct {
	port      int
	templates *template.Template
	posts     []BlogPost
}

func main() {
	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	server := &Server{
		port:  port,
		posts: loadSamplePosts(),
	}

	if err := server.loadTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	server.setupRoutes()
	server.start()
}

func (s *Server) loadTemplates() error {
	templateDir := "templates"
	s.templates = template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))
	return nil
}

func (s *Server) setupRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Page routes
	http.HandleFunc("/", s.withLogging(s.homeHandler))
	http.HandleFunc("/about", s.withLogging(s.aboutHandler))
	http.HandleFunc("/blog", s.withLogging(s.blogHandler))
	http.HandleFunc("/contact", s.withLogging(s.contactHandler))

	// Form handlers
	http.HandleFunc("/contact/submit", s.withLogging(s.contactSubmitHandler))

	// API routes
	http.HandleFunc("/api/posts", s.withLogging(s.apiPostsHandler))
	http.HandleFunc("/api/posts/", s.withLogging(s.apiPostHandler))

	// Health check
	http.HandleFunc("/health", s.withLogging(s.healthHandler))

	// Catch-all for 404
	http.HandleFunc("/", s.withLogging(s.notFoundHandler))
}

func (s *Server) start() {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Visit http://localhost%d to see the website", s.port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.notFoundHandler(w, r)
		return
	}

	data := PageData{
		Title:       "Welcome to Go Web Server",
		Content:     "This is a simple web server built with Go",
		Posts:       s.posts[:3], // Show latest 3 posts
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "index.html", data)
}

func (s *Server) aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "About Us",
		Content:     "Learn more about our Go web server project",
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "about.html", data)
}

func (s *Server) blogHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Blog",
		Content:     "Latest blog posts",
		Posts:       s.posts,
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "blog.html", data)
}

func (s *Server) contactHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Contact",
		Content:     "Get in touch with us",
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "contact.html", data)
}

func (s *Server) contactSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	// In a real application, you would save this to a database
	log.Printf("Contact form submission: Name=%s, Email=%s, Message=%s", name, email, message)

	// Show thank you page
	data := PageData{
		Title:       "Thank You!",
		Content:     fmt.Sprintf("Thank you for your message, %s! We'll get back to you at %s.", name, email),
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "thankyou.html", data)
}

func (s *Server) apiPostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.posts)
	case http.MethodPost:
		var post BlogPost
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		post.ID = len(s.posts) + 1
		post.Date = time.Now()
		s.posts = append(s.posts, post)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(post)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) apiPostHandler(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Find post
	var post *BlogPost
	for i := range s.posts {
		if s.posts[i].ID == id {
			post = &s.posts[i]
			break
		}
	}

	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	case http.MethodPut:
		var updatedPost BlogPost
		if err := json.NewDecoder(r.Body).Decode(&updatedPost); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		post.Title = updatedPost.Title
		post.Content = updatedPost.Content
		post.Author = updatedPost.Author

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	case http.MethodDelete:
		// Remove post from slice
		for i, p := range s.posts {
			if p.ID == id {
				s.posts = append(s.posts[:i], s.posts[i+1:]...)
				break
			}
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"posts":     len(s.posts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	data := PageData{
		Title:       "Page Not Found",
		Content:     "The page you're looking for doesn't exist.",
		CurrentTime: time.Now(),
		RequestID:   generateRequestID(),
	}

	s.renderTemplate(w, "404.html", data)
}

func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if s.templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}

	tmpl := s.templates.Lookup(templateName)
	if tmpl == nil {
		http.Error(w, "Template not found: "+templateName, http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func loadSamplePosts() []BlogPost {
	return []BlogPost{
		{
			ID:      1,
			Title:   "Welcome to Our Blog",
			Content: "This is our first blog post built with Go!",
			Author:  "Go Developer",
			Date:    time.Now().AddDate(0, 0, -7),
		},
		{
			ID:      2,
			Title:   "Building Web Servers with Go",
			Content: "Go provides excellent standard library packages for building web servers.",
			Author:  "Go Expert",
			Date:    time.Now().AddDate(0, 0, -5),
		},
		{
			ID:      3,
			Title:   "HTTP Templates in Go",
			Content: "The html/template package provides secure template rendering.",
			Author:  "Template Master",
			Date:    time.Now().AddDate(0, 0, -2),
		},
	}
}
