package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConfig struct {
	Path     string
	InMemory bool
}

type SQLiteDatabase struct {
	db *sql.DB
}

// InitSQLite initializes a SQLite database connection
func NewSQLiteDatabase(config SQLiteConfig) (*SQLiteDatabase, error) {
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
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(25)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(3 * time.Minute) // Maximum idle time before closing

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}

	return &SQLiteDatabase{db: db}, nil
}

// RunSQLiteMigrations sets up the necessary tables
func (s *SQLiteDatabase) RunSQLiteMigrations() error {
	// Create users table
	createUserTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL
		);

		-- Create unique indexes for faster lookups
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := s.db.Exec(createUserTableQuery)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	return nil
}

func (s *SQLiteDatabase) Conn(ctx context.Context) (*sql.Conn, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}
	return conn, nil
}

func (s *SQLiteDatabase) ExecuteQuery(query string, args ...interface{}) (sql.Result, error) {
	conn, err := s.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}
	defer conn.Close()

	res, err := conn.ExecContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	return res, nil
}

func (s *SQLiteDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	conn, err := s.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}
	defer conn.Close()

	rows, err := conn.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	return rows, nil
}

func (s *SQLiteDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	conn, err := s.Conn(context.Background())
	if err != nil {
		return nil
	}
	defer conn.Close()

	row := conn.QueryRowContext(context.Background(), query, args...)
	return row
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL
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
func (s *SQLiteDatabase) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
