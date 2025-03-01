package db

import (
	"database/sql"
	"log"
)

// RunMigrations executes all necessary database migrations
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	if err := createArticlesTableIfNotExists(db); err != nil {
		return err
	}

	if err := updateArticlesTableSchema(db); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

// createArticlesTableIfNotExists creates the articles table if it doesn't exist
func createArticlesTableIfNotExists(db *sql.DB) error {
	log.Println("Checking for articles table...")

	// Check if table exists
	var tableName string
	err := db.QueryRow("SHOW TABLES LIKE 'articles'").Scan(&tableName)

	// If no error, table exists
	if err == nil {
		log.Println("Articles table already exists")
		return nil
	}

	// If error is not "no rows", something else is wrong
	if err != sql.ErrNoRows {
		return err
	}

	// Table doesn't exist, create it
	log.Println("Creating articles table...")

	_, err = db.Exec(`
		CREATE TABLE articles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			image_url VARCHAR(255),
			author VARCHAR(100) NOT NULL,
			image_data MEDIUMBLOB,
			image_type VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`)

	if err != nil {
		return err
	}

	log.Println("Articles table created successfully")
	return nil
}

// updateArticlesTableSchema updates the articles table with any missing columns
func updateArticlesTableSchema(db *sql.DB) error {
	log.Println("Checking for missing columns in articles table...")

	// Check if the image_data column exists
	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'articles' 
		AND COLUMN_NAME = 'image_data'
	`).Scan(&columnExists)

	if err != nil {
		return err
	}

	// If image_data column doesn't exist, add it along with image_type
	if !columnExists {
		log.Println("Adding image_data and image_type columns to articles table...")

		_, err = db.Exec(`
			ALTER TABLE articles
			ADD COLUMN image_data MEDIUMBLOB AFTER author,
			ADD COLUMN image_type VARCHAR(100) AFTER image_data
		`)

		if err != nil {
			return err
		}

		log.Println("Image columns added successfully")
	} else {
		log.Println("Image columns already exist")
	}

	return nil
}
