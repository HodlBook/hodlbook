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
			return fmt.Errorf("failed to create data directory: %w", err)
		}

		// Open database
		conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
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
