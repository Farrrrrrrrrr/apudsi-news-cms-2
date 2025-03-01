package db

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// DBConnection represents a database connection
type DBConnection struct {
	Host          string
	Port          string
	Username      string
	Password      string
	DBName        string
	SSLMode       string
	CACertPath    string
	CACertContent string // Add this new field
	SkipVerify    bool
}

// NewConnectionFromEnv creates a new DBConnection from environment variables
func NewConnectionFromEnv() *DBConnection {
	return &DBConnection{
		Host:          os.Getenv("DB_HOST"),
		Port:          os.Getenv("DB_PORT"),
		Username:      os.Getenv("DB_USERNAME"),
		Password:      os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		SSLMode:       os.Getenv("DB_SSL_MODE"),
		CACertPath:    os.Getenv("DB_CA_CERT_PATH"),
		CACertContent: os.Getenv("DB_CA_CERT"), // Read cert content from env var
		SkipVerify:    os.Getenv("DB_SKIP_VERIFY") == "true",
	}
}

// DSN returns the Data Source Name for database connection
func (c *DBConnection) DSN() string {
	// MySQL format: username:password@tcp(host:port)/dbname?tls=true
	tlsConfig := "false"

	if c.SSLMode == "require" || c.SSLMode == "true" {
		if c.CACertPath != "" {
			// Use custom TLS config with CA cert
			tlsConfig = "custom"

			// Register TLS config
			rootCertPool := x509.NewCertPool()
			pem, err := ioutil.ReadFile(c.CACertPath)
			if err == nil {
				if ok := rootCertPool.AppendCertsFromPEM(pem); ok {
					mysql.RegisterTLSConfig("custom", &tls.Config{
						RootCAs:            rootCertPool,
						InsecureSkipVerify: c.SkipVerify,
					})
				}
			}
		} else if c.SkipVerify {
			// Skip verification mode
			tlsConfig = "skip-verify"
			mysql.RegisterTLSConfig("skip-verify", &tls.Config{
				InsecureSkipVerify: true,
			})
		} else {
			// Standard TLS
			tlsConfig = "true"
		}
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=%s&parseTime=true",
		c.Username, c.Password, c.Host, c.Port, c.DBName, tlsConfig)
}

// TestConnection attempts to connect to the database and returns error if unsuccessful
func (c *DBConnection) TestConnection() error {
	db, err := sql.Open("mysql", c.DSN())
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)

	// Test the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	return nil
}

// GetDB returns a database connection
func (c *DBConnection) GetDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}
