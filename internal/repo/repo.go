package repo

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNilDatabase = errors.New("database instance cannot be nil")
)

// Repository contains all repository interfaces
type Repository struct {
	Transaction *TransactionRepository
	Price       *PriceRepository
}

// Option is the functional options pattern for Repository
type RepositoryOption func(*Repository) error

// New creates a new Repository with options
func New(opts ...RepositoryOption) (*Repository, error) {
	r := &Repository{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// WithTransactionRepository sets the transaction repository
func WithTransactionRepository(txRepo *TransactionRepository) RepositoryOption {
	return func(r *Repository) error {
		if txRepo == nil {
			return errors.New("transaction repository cannot be nil")
		}
		r.Transaction = txRepo
		return nil
	}
}

// WithPriceRepository sets the price repository
func WithPriceRepository(priceRepo *PriceRepository) RepositoryOption {
	return func(r *Repository) error {
		if priceRepo == nil {
			return errors.New("price repository cannot be nil")
		}
		r.Price = priceRepo
		return nil
	}
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Transaction{}, &Price{})
}
