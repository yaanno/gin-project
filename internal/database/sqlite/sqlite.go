package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConfig struct {
	Path     string
	InMemory bool
}

var (
	sqliteDB *sql.DB
)

// InitSQLite initializes a SQLite database connection
func InitSQLite(config SQLiteConfig) (*sql.DB, error) {
	var (
		connStr string
		err     error
	)

	// Determine connection string based on configuration
	if config.InMemory {
		// In-memory database
		connStr = "file::memory:?cache=shared"
	} else {
		// Persistent database
		if config.Path == "" {
			// Default to a data directory if no path specified
			dataDir := filepath.Join(".", "data")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create data directory: %v", err)
			}
			config.Path = filepath.Join(dataDir, "users.db")
		}

		connStr = fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL", config.Path)
	}

	// Open database connection
	sqliteDB, err = sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %v", err)
	}

	// Configure connection pool
	sqliteDB.SetMaxOpenConns(1)
	sqliteDB.SetMaxIdleConns(1)

	// Test the connection
	if err = sqliteDB.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}

	return sqliteDB, nil
}

// RunSQLiteMigrations sets up the necessary tables
func RunSQLiteMigrations() error {
	// Create users table
	createUserTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Create unique indexes for faster lookups
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := sqliteDB.Exec(createUserTableQuery)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	return nil
}

// CreateInMemoryTestDB creates an in-memory test database
func CreateInMemoryTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("error creating in-memory test database: %v", err)
	}

	// Run migrations
	createUserTableQuery := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err = db.Exec(createUserTableQuery)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating users table in test database: %v", err)
	}

	return db, nil
}

// CloseSQLiteDB closes the database connection
func CloseSQLiteDB() {
	if sqliteDB != nil {
		sqliteDB.Close()
	}
}
