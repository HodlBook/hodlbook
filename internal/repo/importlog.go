package repo

import (
	"hodlbook/internal/models"
)

func (r *Repository) CreateImportLog(log *models.ImportLog) error {
	return r.db.Create(log).Error
}

func (r *Repository) GetImportLogByID(id int64) (*models.ImportLog, error) {
	var log models.ImportLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *Repository) ListImportLogs() ([]models.ImportLog, error) {
	var logs []models.ImportLog
	if err := r.db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *Repository) UpdateImportLog(log *models.ImportLog) error {
	return r.db.Save(log).Error
}

func (r *Repository) DeleteImportLog(id int64) error {
	return r.db.Delete(&models.ImportLog{}, id).Error
}
