package repo

import (
	"hodlbook/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportLogRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	log := &models.ImportLog{
		Filename:     "test.csv",
		Format:       "csv",
		EntityType:   "asset",
		TotalRows:    10,
		ImportedRows: 8,
		FailedRows:   2,
		Status:       "partial",
		FailedData:   `[{"row":1,"message":"error"}]`,
	}

	err = repository.CreateImportLog(log)
	require.NoError(t, err)
	require.NotZero(t, log.ID)

	got, err := repository.GetImportLogByID(log.ID)
	require.NoError(t, err)
	require.Equal(t, log.Filename, got.Filename)
	require.Equal(t, log.Format, got.Format)
	require.Equal(t, log.EntityType, got.EntityType)
	require.Equal(t, log.TotalRows, got.TotalRows)
	require.Equal(t, log.ImportedRows, got.ImportedRows)
	require.Equal(t, log.FailedRows, got.FailedRows)
	require.Equal(t, log.Status, got.Status)
	require.Equal(t, log.FailedData, got.FailedData)

	logs, err := repository.ListImportLogs()
	require.NoError(t, err)
	require.Len(t, logs, 1)

	log.Status = "completed"
	log.FailedRows = 0
	log.ImportedRows = 10
	log.FailedData = "[]"
	err = repository.UpdateImportLog(log)
	require.NoError(t, err)

	got, err = repository.GetImportLogByID(log.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", got.Status)
	require.Equal(t, 0, got.FailedRows)
	require.Equal(t, 10, got.ImportedRows)

	err = repository.DeleteImportLog(log.ID)
	require.NoError(t, err)

	_, err = repository.GetImportLogByID(log.ID)
	require.Error(t, err)
}

func TestImportLogRepository_ListOrdering(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	log1 := &models.ImportLog{Filename: "first.csv", Format: "csv", EntityType: "asset", Status: "completed"}
	log2 := &models.ImportLog{Filename: "second.csv", Format: "json", EntityType: "asset", Status: "partial"}

	require.NoError(t, repository.CreateImportLog(log1))
	require.NoError(t, repository.CreateImportLog(log2))

	logs, err := repository.ListImportLogs()
	require.NoError(t, err)
	require.Len(t, logs, 2)
	require.Equal(t, "second.csv", logs[0].Filename)
	require.Equal(t, "first.csv", logs[1].Filename)
}

func TestImportLogRepository_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	_, err = repository.GetImportLogByID(999)
	require.Error(t, err)
}
