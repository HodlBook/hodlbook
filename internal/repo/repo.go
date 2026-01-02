package repo

import (
	"errors"

	"hodlbook/internal/models"

	"gorm.io/gorm"
)

var ErrNilDatabase = errors.New("database cannot be nil")

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Migrate() error {
	if err := r.db.AutoMigrate(
		&models.Asset{},
		&models.Exchange{},
		&models.Price{},
		&models.AssetHistoricValue{},
		&models.ImportLog{},
	); err != nil {
		return err
	}

	r.db.Exec("DROP INDEX IF EXISTS uni_assets_symbol")
	r.db.Exec("DROP INDEX IF EXISTS idx_assets_symbol_unique")
	r.db.Exec("DROP INDEX IF EXISTS idx_assets_symbol")

	return nil
}
