package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ImportExportTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	ctrl   *Controller
	repo   *repo.Repository
}

func (s *ImportExportTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NoError(db.AutoMigrate(&models.Asset{}, &models.Exchange{}, &models.AssetHistoricValue{}, &models.ImportLog{}))
	s.db = db

	repository, err := repo.New(db)
	s.Require().NoError(err)
	s.repo = repository

	ctrl, err := New(WithRepository(repository))
	s.Require().NoError(err)
	s.ctrl = ctrl

	s.router = gin.New()
	api := s.router.Group("/api")

	assets := api.Group("/assets")
	assets.GET("/export", ctrl.ExportAssets)
	assets.POST("/import", ctrl.ImportAssets)

	exchanges := api.Group("/exchanges")
	exchanges.GET("/export", ctrl.ExportExchanges)

	imports := api.Group("/imports")
	imports.GET("", ctrl.ListImportLogs)
	imports.GET("/:id", ctrl.GetImportLog)
	imports.POST("/:id/retry", ctrl.RetryImport)
	imports.DELETE("/:id", ctrl.DeleteImportLog)
}

func (s *ImportExportTestSuite) SetupTest() {
	s.db.Exec("DELETE FROM assets")
	s.db.Exec("DELETE FROM exchanges")
	s.db.Exec("DELETE FROM import_logs")
}

func (s *ImportExportTestSuite) seedAssets() {
	assets := []models.Asset{
		{Symbol: "BTC", Name: "Bitcoin", Amount: 1.5, TransactionType: "deposit", Timestamp: time.Now()},
		{Symbol: "ETH", Name: "Ethereum", Amount: 10.0, TransactionType: "deposit", Timestamp: time.Now()},
		{Symbol: "BTC", Name: "Bitcoin", Amount: 0.5, TransactionType: "withdrawal", Timestamp: time.Now()},
	}
	for _, a := range assets {
		s.db.Create(&a)
	}
}

func (s *ImportExportTestSuite) seedExchanges() {
	exchanges := []models.Exchange{
		{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 0.1, ToAmount: 1.5, Timestamp: time.Now()},
		{FromSymbol: "ETH", ToSymbol: "USDT", FromAmount: 5.0, ToAmount: 5000.0, Timestamp: time.Now()},
	}
	for _, e := range exchanges {
		s.db.Create(&e)
	}
}

// Export Tests

func (s *ImportExportTestSuite) TestExportAssets_CSV() {
	s.seedAssets()

	req := httptest.NewRequest(http.MethodGet, "/api/assets/export?format=csv", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Header().Get("Content-Type"), "text/csv")
	s.Contains(w.Header().Get("Content-Disposition"), "attachment")

	body := w.Body.String()
	s.Contains(body, "symbol;name;amount;transaction_type;timestamp;notes")
	s.Contains(body, "BTC;Bitcoin")
	s.Contains(body, "ETH;Ethereum")
}

func (s *ImportExportTestSuite) TestExportAssets_JSON() {
	s.seedAssets()

	req := httptest.NewRequest(http.MethodGet, "/api/assets/export?format=json", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Header().Get("Content-Type"), "application/json")

	var assets []models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &assets))
	s.Len(assets, 3)
}

func (s *ImportExportTestSuite) TestExportAssets_InvalidFormat() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets/export?format=xml", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ImportExportTestSuite) TestExportExchanges_CSV() {
	s.seedExchanges()

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/export?format=csv", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Header().Get("Content-Type"), "text/csv")

	body := w.Body.String()
	s.Contains(body, "from_symbol;to_symbol;from_amount;to_amount;fee;fee_currency;timestamp;notes")
	s.Contains(body, "BTC;ETH")
}

func (s *ImportExportTestSuite) TestExportExchanges_JSON() {
	s.seedExchanges()

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/export?format=json", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var exchanges []models.Exchange
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &exchanges))
	s.Len(exchanges, 2)
}

// Import Tests

func (s *ImportExportTestSuite) createMultipartRequest(url, fieldname, filename, content string, format string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(fieldname, filename)
	part.Write([]byte(content))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, url+"?format="+format, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func (s *ImportExportTestSuite) TestImportAssets_JSON_AllValid() {
	jsonData := `[
		{"symbol": "BTC", "name": "Bitcoin", "amount": 1.0, "transaction_type": "deposit"},
		{"symbol": "ETH", "name": "Ethereum", "amount": 5.0, "transaction_type": "deposit"}
	]`

	req := s.createMultipartRequest("/api/assets/import", "file", "test.json", jsonData, "json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(2, result.Imported)
	s.Equal(0, result.Failed)
	s.Equal("completed", result.Status)

	var count int64
	s.db.Model(&models.Asset{}).Count(&count)
	s.Equal(int64(2), count)
}

func (s *ImportExportTestSuite) TestImportAssets_JSON_PartialValid() {
	jsonData := `[
		{"symbol": "BTC", "name": "Bitcoin", "amount": 1.0, "transaction_type": "deposit"},
		{"symbol": "", "name": "Missing Symbol", "amount": 5.0, "transaction_type": "deposit"},
		{"symbol": "ETH", "name": "Ethereum", "amount": -1.0, "transaction_type": "deposit"}
	]`

	req := s.createMultipartRequest("/api/assets/import", "file", "test.json", jsonData, "json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(1, result.Imported)
	s.Equal(2, result.Failed)
	s.Equal("partial", result.Status)
	s.Len(result.Errors, 2)
}

func (s *ImportExportTestSuite) TestImportAssets_JSON_AllInvalid() {
	jsonData := `[
		{"symbol": "", "amount": 1.0, "transaction_type": "deposit"},
		{"symbol": "BTC", "amount": -1.0, "transaction_type": "deposit"}
	]`

	req := s.createMultipartRequest("/api/assets/import", "file", "test.json", jsonData, "json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(0, result.Imported)
	s.Equal(2, result.Failed)
	s.Equal("failed", result.Status)
}

func (s *ImportExportTestSuite) TestImportAssets_CSV_Valid() {
	csvData := `symbol;name;amount;transaction_type;timestamp;notes
BTC;Bitcoin;1.5;deposit;2024-01-15T10:30:00Z;Initial
ETH;Ethereum;10.0;deposit;2024-01-16T11:00:00Z;Second`

	req := s.createMultipartRequest("/api/assets/import", "file", "test.csv", csvData, "csv")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(2, result.Imported)
	s.Equal(0, result.Failed)
	s.Equal("completed", result.Status)
}

func (s *ImportExportTestSuite) TestImportAssets_CSV_InvalidTransaction() {
	csvData := `symbol;name;amount;transaction_type
BTC;Bitcoin;1.5;invalid_type`

	req := s.createMultipartRequest("/api/assets/import", "file", "test.csv", csvData, "csv")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(0, result.Imported)
	s.Equal(1, result.Failed)
}

func (s *ImportExportTestSuite) TestImportAssets_NoFile() {
	req := httptest.NewRequest(http.MethodPost, "/api/assets/import?format=json", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ImportExportTestSuite) TestImportAssets_InvalidFormat() {
	req := s.createMultipartRequest("/api/assets/import", "file", "test.xml", "<data/>", "xml")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// Import Log Tests

func (s *ImportExportTestSuite) TestListImportLogs_Empty() {
	req := httptest.NewRequest(http.MethodGet, "/api/imports", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var logs []models.ImportLog
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &logs))
	s.Empty(logs)
}

func (s *ImportExportTestSuite) TestListImportLogs_AfterImport() {
	jsonData := `[{"symbol": "BTC", "amount": 1.0, "transaction_type": "deposit"}]`
	req := s.createMultipartRequest("/api/assets/import", "file", "test.json", jsonData, "json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodGet, "/api/imports", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var logs []models.ImportLog
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &logs))
	s.Len(logs, 1)
	s.Equal("test.json", logs[0].Filename)
	s.Equal("json", logs[0].Format)
	s.Equal("asset", logs[0].EntityType)
}

func (s *ImportExportTestSuite) TestGetImportLog() {
	log := &models.ImportLog{Filename: "test.csv", Format: "csv", EntityType: "asset", Status: "completed"}
	s.db.Create(log)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/imports/%d", log.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result models.ImportLog
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal("test.csv", result.Filename)
}

func (s *ImportExportTestSuite) TestGetImportLog_NotFound() {
	req := httptest.NewRequest(http.MethodGet, "/api/imports/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ImportExportTestSuite) TestDeleteImportLog() {
	log := &models.ImportLog{Filename: "test.csv", Format: "csv", EntityType: "asset", Status: "completed"}
	s.db.Create(log)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/imports/%d", log.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)

	var count int64
	s.db.Model(&models.ImportLog{}).Count(&count)
	s.Equal(int64(0), count)
}

func (s *ImportExportTestSuite) TestRetryImport() {
	failedData := `[{"row":1,"data":{"symbol":"","amount":1},"message":"symbol is required"}]`
	log := &models.ImportLog{
		Filename:     "test.json",
		Format:       "json",
		EntityType:   "asset",
		TotalRows:    1,
		ImportedRows: 0,
		FailedRows:   1,
		Status:       "failed",
		FailedData:   failedData,
	}
	s.db.Create(log)

	correctedAssets := []models.Asset{
		{Symbol: "BTC", Name: "Bitcoin", Amount: 1.0, TransactionType: "deposit"},
	}
	body, _ := json.Marshal(correctedAssets)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/imports/%d/retry", log.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result ImportResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Equal(1, result.Imported)
	s.Equal(0, result.Failed)

	var updatedLog models.ImportLog
	s.db.First(&updatedLog, log.ID)
	s.Equal("completed", updatedLog.Status)
	s.Equal(1, updatedLog.ImportedRows)
}

func (s *ImportExportTestSuite) TestRetryImport_NotFound() {
	correctedAssets := []models.Asset{{Symbol: "BTC", Amount: 1.0, TransactionType: "deposit"}}
	body, _ := json.Marshal(correctedAssets)

	req := httptest.NewRequest(http.MethodPost, "/api/imports/999/retry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// Helper function tests

func (s *ImportExportTestSuite) TestParseAssetsFromCSV() {
	csvData := `symbol;name;amount;transaction_type;timestamp;notes
BTC;Bitcoin;1.5;deposit;2024-01-15T10:30:00Z;Test note
ETH;Ethereum;10.0;withdrawal;;No timestamp`

	assets, errors := parseAssetsFromCSV([]byte(csvData))
	s.Len(assets, 2)
	s.Empty(errors)

	s.Equal("BTC", assets[0].Symbol)
	s.Equal("Bitcoin", assets[0].Name)
	s.Equal(1.5, assets[0].Amount)
	s.Equal("deposit", assets[0].TransactionType)
	s.Equal("Test note", assets[0].Notes)

	s.Equal("ETH", assets[1].Symbol)
	s.Equal("withdrawal", assets[1].TransactionType)
}

func (s *ImportExportTestSuite) TestParseAssetsFromCSV_InvalidFormat() {
	csvData := `not a valid csv with proper structure`

	assets, errors := parseAssetsFromCSV([]byte(csvData))
	s.Empty(assets)
	s.Len(errors, 1)
	s.Contains(errors[0].Message, "at least one data row")
}

func (s *ImportExportTestSuite) TestParseAssetsFromJSON() {
	jsonData := `[
		{"symbol": "BTC", "name": "Bitcoin", "amount": 1.5, "transaction_type": "deposit"},
		{"symbol": "ETH", "amount": 10.0, "transaction_type": "withdrawal"}
	]`

	assets, errors := parseAssetsFromJSON([]byte(jsonData))
	s.Len(assets, 2)
	s.Empty(errors)

	s.Equal("BTC", assets[0].Symbol)
	s.Equal(1.5, assets[0].Amount)
}

func (s *ImportExportTestSuite) TestParseAssetsFromJSON_InvalidJSON() {
	jsonData := `not valid json`

	assets, errors := parseAssetsFromJSON([]byte(jsonData))
	s.Empty(assets)
	s.Len(errors, 1)
	s.Contains(errors[0].Message, "invalid JSON")
}

func (s *ImportExportTestSuite) TestValidateAssetFields() {
	validAsset := &models.Asset{Symbol: "BTC", Amount: 1.0, TransactionType: "deposit"}
	s.NoError(validateAssetFields(validAsset))

	missingSymbol := &models.Asset{Amount: 1.0, TransactionType: "deposit"}
	s.Error(validateAssetFields(missingSymbol))
	s.Contains(validateAssetFields(missingSymbol).Error(), "symbol")

	negativeAmount := &models.Asset{Symbol: "BTC", Amount: -1.0, TransactionType: "deposit"}
	s.Error(validateAssetFields(negativeAmount))
	s.Contains(validateAssetFields(negativeAmount).Error(), "amount")

	invalidType := &models.Asset{Symbol: "BTC", Amount: 1.0, TransactionType: "invalid"}
	s.Error(validateAssetFields(invalidType))
	s.Contains(validateAssetFields(invalidType).Error(), "transaction_type")

	withdrawAlias := &models.Asset{Symbol: "BTC", Amount: 1.0, TransactionType: "withdraw"}
	s.NoError(validateAssetFields(withdrawAlias))
	s.Equal("withdrawal", withdrawAlias.TransactionType)
}

func (s *ImportExportTestSuite) TestAssetsToCSV() {
	assets := []models.Asset{
		{Symbol: "BTC", Name: "Bitcoin", Amount: 1.5, TransactionType: "deposit", Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Notes: "Test"},
	}

	csv := string(assetsToCSV(assets))
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	s.Len(lines, 2)
	s.Equal("symbol;name;amount;transaction_type;timestamp;notes", lines[0])
	s.Contains(lines[1], "BTC;Bitcoin;1.5;deposit;")
}

func (s *ImportExportTestSuite) TestExchangesToCSV() {
	exchanges := []models.Exchange{
		{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 1.0, ToAmount: 15.0, Fee: 0.001, FeeCurrency: "BTC", Timestamp: time.Now()},
	}

	csv := string(exchangesToCSV(exchanges))
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	s.Len(lines, 2)
	s.Equal("from_symbol;to_symbol;from_amount;to_amount;fee;fee_currency;timestamp;notes", lines[0])
	s.Contains(lines[1], "BTC;ETH;1;15;0.001;BTC;")
}

func TestImportExport(t *testing.T) {
	suite.Run(t, new(ImportExportTestSuite))
}
