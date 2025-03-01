package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/farrell_ivander/test-conn/db"
	"github.com/farrell_ivander/test-conn/handlers"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize database connection
	dbConn := db.NewConnectionFromEnv()
	database, err := dbConn.GetDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	database.Close()

	// Initialize the handlers
	h := handlers.NewHandler()

	// Define routes
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/test-connection", h.TestConnectionHandler)
	http.HandleFunc("/articles", h.ListArticlesHandler)
	http.HandleFunc("/article", h.GetArticleHandler)
	http.HandleFunc("/image", h.GetImageHandler) // Add image serving handler

	// Article management routes
	http.HandleFunc("/article/new", h.NewArticleHandler)
	http.HandleFunc("/article/create", h.CreateArticleHandler)
	http.HandleFunc("/article/edit", h.EditArticleHandler)
	http.HandleFunc("/article/update", h.UpdateArticleHandler)
	http.HandleFunc("/article/delete", h.DeleteArticleHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	log.Printf("Server starting on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
