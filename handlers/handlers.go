package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/farrell_ivander/test-conn/db"
	"github.com/farrell_ivander/test-conn/models"
)

// Handler holds handler dependencies
type Handler struct {
	templates *template.Template
}

// NewHandler initializes and returns a new Handler
func NewHandler() *Handler {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	return &Handler{
		templates: templates,
	}
}

// HomeHandler handles the home page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	h.templates.ExecuteTemplate(w, "index.html", nil)
}

// TestConnectionHandler handles database connection tests
func (h *Handler) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form values for custom connection
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	var conn *db.DBConnection

	// Check if the request wants to use environment variables
	if r.FormValue("use_env") == "true" {
		conn = db.NewConnectionFromEnv()
	} else {
		// Use form-provided connection details
		conn = &db.DBConnection{
			Host:     r.FormValue("host"),
			Port:     r.FormValue("port"),
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
			DBName:   r.FormValue("dbname"),
			SSLMode:  r.FormValue("sslmode"),
		}
	}

	// Test the connection
	err := conn.TestConnection()

	// Prepare the response
	result := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["message"] = "Connection successful!"
	}

	// Return the result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListArticlesHandler handles listing and searching articles
func (h *Handler) ListArticlesHandler(w http.ResponseWriter, r *http.Request) {
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		h.templates.ExecuteTemplate(w, "articles.html", map[string]interface{}{
			"Error": "Failed to connect to database: " + err.Error(),
		})
		return
	}
	defer database.Close()

	searchTerm := r.URL.Query().Get("search")
	var articles []models.Article

	if searchTerm != "" {
		// Search for articles
		articles, err = models.SearchArticles(database, searchTerm)
	} else {
		// Get all articles (limit to 50 for performance)
		articles, err = models.GetArticles(database, 50)
	}

	if err != nil {
		h.templates.ExecuteTemplate(w, "articles.html", map[string]interface{}{
			"Error": "Failed to fetch articles: " + err.Error(),
		})
		return
	}

	h.templates.ExecuteTemplate(w, "articles.html", map[string]interface{}{
		"Articles":   articles,
		"SearchTerm": searchTerm,
	})
}

// GetArticleHandler handles displaying a single article
func (h *Handler) GetArticleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Redirect(w, r, "/articles", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		h.templates.ExecuteTemplate(w, "article.html", map[string]interface{}{
			"Error": "Failed to connect to database: " + err.Error(),
		})
		return
	}
	defer database.Close()

	article, err := models.GetArticleByID(database, id)
	if err != nil {
		h.templates.ExecuteTemplate(w, "article.html", map[string]interface{}{
			"Error": "Failed to fetch article: " + err.Error(),
		})
		return
	}

	h.templates.ExecuteTemplate(w, "article.html", map[string]interface{}{
		"Article": article,
	})
}

// GetImageHandler serves images stored in the database
func (h *Handler) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Article ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Connect to database
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get image from database
	imageBlob, contentType, err := models.GetImageByArticleID(database, id)
	if err != nil {
		http.Error(w, "Failed to retrieve image", http.StatusNotFound)
		return
	}

	// If content type is not provided, try to guess it
	if contentType == "" {
		contentType = http.DetectContentType(imageBlob)
	}

	// Set content type header and write image data
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))
	w.Write(imageBlob)
}

// NewArticleHandler displays the form for creating a new article
func (h *Handler) NewArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Display the form
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Create New Article",
			"FormURL": "/article/create",
			"Article": &models.Article{},
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// CreateArticleHandler handles creating a new article
func (h *Handler) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form data with 10MB max memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create article from form data
	article := &models.Article{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		ImageURL:    r.FormValue("image_url"),
		Author:      r.FormValue("author"),
	}

	// Check if a file was uploaded
	file, header, err := r.FormFile("image")
	if err == nil {
		// File was uploaded
		defer file.Close()

		// Read the file content
		imageBlob, err := ioutil.ReadAll(file)
		if err != nil {
			h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
				"Title":   "Create New Article",
				"FormURL": "/article/create",
				"Article": article,
				"Error":   "Failed to read uploaded image: " + err.Error(),
			})
			return
		}

		// Set the image blob and content type
		article.ImageBlob = imageBlob
		article.ImageType = getContentType(header.Filename, header.Header.Get("Content-Type"))
	}

	// Validate required fields
	if article.Title == "" || article.Description == "" || article.Author == "" {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Create New Article",
			"FormURL": "/article/create",
			"Article": article,
			"Error":   "Title, Description and Author are required",
		})
		return
	}

	// Connect to database
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Create New Article",
			"FormURL": "/article/create",
			"Article": article,
			"Error":   "Failed to connect to database: " + err.Error(),
		})
		return
	}
	defer database.Close()

	// Create article in database
	id, err := models.CreateArticle(database, article)
	if err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Create New Article",
			"FormURL": "/article/create",
			"Article": article,
			"Error":   "Failed to create article: " + err.Error(),
		})
		return
	}

	// Redirect to the new article
	http.Redirect(w, r, "/article?id="+strconv.Itoa(id), http.StatusSeeOther)
}

// EditArticleHandler displays the form for editing an article
func (h *Handler) EditArticleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Redirect(w, r, "/articles", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Connect to database
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title": "Edit Article",
			"Error": "Failed to connect to database: " + err.Error(),
		})
		return
	}
	defer database.Close()

	// Get article to edit
	article, err := models.GetArticleByID(database, id)
	if err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title": "Edit Article",
			"Error": "Failed to fetch article: " + err.Error(),
		})
		return
	}

	h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
		"Title":   "Edit Article",
		"FormURL": "/article/update?id=" + idStr,
		"Article": article,
	})
}

// UpdateArticleHandler handles updating an existing article
func (h *Handler) UpdateArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Article ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Parse the multipart form data with 10MB max memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create article from form data
	article := &models.Article{
		ID:          id,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		ImageURL:    r.FormValue("image_url"),
		Author:      r.FormValue("author"),
	}

	// Check if a file was uploaded
	file, header, err := r.FormFile("image")
	if err == nil {
		// File was uploaded
		defer file.Close()

		// Read the file content
		imageBlob, err := ioutil.ReadAll(file)
		if err != nil {
			h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
				"Title":   "Edit Article",
				"FormURL": "/article/update?id=" + idStr,
				"Article": article,
				"Error":   "Failed to read uploaded image: " + err.Error(),
			})
			return
		}

		// Set the image blob and content type
		article.ImageBlob = imageBlob
		article.ImageType = getContentType(header.Filename, header.Header.Get("Content-Type"))
	}

	// Validate required fields
	if article.Title == "" || article.Description == "" || article.Author == "" {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Edit Article",
			"FormURL": "/article/update?id=" + idStr,
			"Article": article,
			"Error":   "Title, Description and Author are required",
		})
		return
	}

	// Connect to database
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Edit Article",
			"FormURL": "/article/update?id=" + idStr,
			"Article": article,
			"Error":   "Failed to connect to database: " + err.Error(),
		})
		return
	}
	defer database.Close()

	// Update article in database
	if err := models.UpdateArticle(database, article); err != nil {
		h.templates.ExecuteTemplate(w, "article_form.html", map[string]interface{}{
			"Title":   "Edit Article",
			"FormURL": "/article/update?id=" + idStr,
			"Article": article,
			"Error":   "Failed to update article: " + err.Error(),
		})
		return
	}

	// Redirect to the updated article
	http.Redirect(w, r, "/article?id="+idStr, http.StatusSeeOther)
}

// DeleteArticleHandler handles deleting an article
func (h *Handler) DeleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "Article ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Connect to database
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Delete article from database
	if err := models.DeleteArticle(database, id); err != nil {
		http.Error(w, "Failed to delete article: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response for AJAX requests
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Article deleted successfully",
	})
}

// getContentType determines the content type from filename and provided content type
func getContentType(filename, providedType string) string {
	if providedType != "" && providedType != "application/octet-stream" {
		return providedType
	}

	// Try to detect from extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		// Try to guess from extension using MIME package
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType
		}
	}

	// Default to generic binary
	return "application/octet-stream"
}
