package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Database holds the GORM database instance
type Database struct {
	conn *gorm.DB
}

// Config holds database configuration
type Config struct {
	Path string
}

// Option is the functional options pattern for Database
type Option func(*Database) error

// New creates a new Database instance with options
func New(opts ...Option) (*Database, error) {
	db := &Database{}
	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, err
		}
	}
	return db, nil
}

// WithPath sets the SQLite database path
func WithPath(path string) Option {
	return func(db *Database) error {
		if path == "" {
			path = "./data/hodlbook.db"
		}

		// Ensure data directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory %s: %w", dir, err)
		}

		// Verify directory is accessible and writable
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("failed to stat data directory %s: %w", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("data path %s is not a directory", dir)
		}

		// Test write permissions by creating a temp file
		testFile := filepath.Join(dir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("data directory %s is not writable: %w", dir, err)
		}
		os.Remove(testFile)

		// Open database
		conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w (path: %s, dir exists: true)", err, path)
		}

		db.conn = conn
		log.Printf("Database connected: %s", path)
		return nil
	}
}

// Get returns the underlying GORM database instance
func (d *Database) Get() *gorm.DB {
	if d.conn == nil {
		log.Fatal("Database not initialized")
	}
	return d.conn
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.conn == nil {
		return nil
	}
	sqlDB, err := d.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
