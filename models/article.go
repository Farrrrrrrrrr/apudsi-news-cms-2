package models

import (
	"database/sql"
	"time"
)

// Article represents a news article
type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ImageBlob   []byte    `json:"-"` // Binary image data
	ImageType   string    `json:"-"` // MIME type of the image
	HasImage    bool      `json:"has_image"`
}

// GetArticles fetches articles from the database with optional limit
func GetArticles(db *sql.DB, limit int) ([]Article, error) {
	query := `
		SELECT id, title, description, image_url, author, created_at, updated_at,
		       CASE WHEN image_data IS NOT NULL THEN true ELSE false END as has_image
		FROM articles ORDER BY created_at DESC
	`
	if limit > 0 {
		query += " LIMIT ?"
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = db.Query(query, limit)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Description,
			&article.ImageURL,
			&article.Author,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.HasImage,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// GetArticleByID fetches a single article by ID
func GetArticleByID(db *sql.DB, id int) (*Article, error) {
	query := `
		SELECT id, title, description, image_url, author, created_at, updated_at, 
		       image_type, CASE WHEN image_data IS NOT NULL THEN true ELSE false END as has_image
		FROM articles WHERE id = ?
	`

	var article Article
	err := db.QueryRow(query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Description,
		&article.ImageURL,
		&article.Author,
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.ImageType,
		&article.HasImage,
	)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

// SearchArticles searches for articles matching the given term
func SearchArticles(db *sql.DB, term string) ([]Article, error) {
	query := `
		SELECT id, title, description, image_url, author, created_at, updated_at,
		       CASE WHEN image_data IS NOT NULL THEN true ELSE false END as has_image
		FROM articles 
		WHERE title LIKE ? OR description LIKE ? OR author LIKE ?
		ORDER BY created_at DESC
	`

	searchTerm := "%" + term + "%"
	rows, err := db.Query(query, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Description,
			&article.ImageURL,
			&article.Author,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.HasImage,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// CreateArticle inserts a new article into the database
func CreateArticle(db *sql.DB, article *Article) (int, error) {
	query := `
		INSERT INTO articles (title, description, image_url, author, image_data, image_type)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query,
		article.Title,
		article.Description,
		article.ImageURL,
		article.Author,
		article.ImageBlob,
		article.ImageType,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateArticle updates an existing article in the database
func UpdateArticle(db *sql.DB, article *Article) error {
	var query string
	var args []interface{}

	// If new image data is provided, update the image as well
	if article.ImageBlob != nil {
		query = `
			UPDATE articles 
			SET title = ?, description = ?, image_url = ?, author = ?, image_data = ?, image_type = ?
			WHERE id = ?
		`
		args = []interface{}{
			article.Title,
			article.Description,
			article.ImageURL,
			article.Author,
			article.ImageBlob,
			article.ImageType,
			article.ID,
		}
	} else {
		// Otherwise only update the text fields and image URL
		query = `
			UPDATE articles 
			SET title = ?, description = ?, image_url = ?, author = ?
			WHERE id = ?
		`
		args = []interface{}{
			article.Title,
			article.Description,
			article.ImageURL,
			article.Author,
			article.ID,
		}
	}

	_, err := db.Exec(query, args...)
	return err
}

// DeleteArticle removes an article from the database by ID
func DeleteArticle(db *sql.DB, id int) error {
	query := "DELETE FROM articles WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}

// GetImageByArticleID retrieves the image data for a specific article
func GetImageByArticleID(db *sql.DB, id int) ([]byte, string, error) {
	query := "SELECT image_data, image_type FROM articles WHERE id = ? AND image_data IS NOT NULL"

	var imageData []byte
	var imageType string

	err := db.QueryRow(query, id).Scan(&imageData, &imageType)
	if err != nil {
		return nil, "", err
	}

	return imageData, imageType, nil
}
