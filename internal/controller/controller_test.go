package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ControllerTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	ctrl   *Controller

	createdAsset       *models.Asset
	createdAsset2      *models.Asset
	createdTransaction *models.Transaction
	createdExchange    *models.Exchange
}

func (s *ControllerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NoError(db.AutoMigrate(&models.Asset{}, &models.Transaction{}, &models.Exchange{}))
	s.db = db

	repository, err := repo.New(db)
	s.Require().NoError(err)

	ctrl, err := New(repository)
	s.Require().NoError(err)
	s.ctrl = ctrl

	s.router = gin.New()
	api := s.router.Group("/api")

	assets := api.Group("/assets")
	assets.GET("", ctrl.ListAssets)
	assets.POST("", ctrl.CreateAsset)
	assets.GET("/:id", ctrl.GetAsset)
	assets.PUT("/:id", ctrl.UpdateAsset)
	assets.DELETE("/:id", ctrl.DeleteAsset)

	transactions := api.Group("/transactions")
	transactions.GET("", ctrl.ListTransactions)
	transactions.POST("", ctrl.CreateTransaction)
	transactions.GET("/:id", ctrl.GetTransaction)
	transactions.PUT("/:id", ctrl.UpdateTransaction)
	transactions.DELETE("/:id", ctrl.DeleteTransaction)

	exchanges := api.Group("/exchanges")
	exchanges.GET("", ctrl.ListExchanges)
	exchanges.POST("", ctrl.CreateExchange)
	exchanges.GET("/:id", ctrl.GetExchange)
	exchanges.PUT("/:id", ctrl.UpdateExchange)
	exchanges.DELETE("/:id", ctrl.DeleteExchange)
}

// Asset Tests

func (s *ControllerTestSuite) Test01_Asset_ListEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var assets []models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &assets))
	s.Empty(assets)
}

func (s *ControllerTestSuite) Test02_Asset_Create() {
	asset := models.Asset{
		Symbol: "BTC",
		Name:   "Bitcoin",
	}
	body, _ := json.Marshal(asset)

	req := httptest.NewRequest(http.MethodPost, "/api/assets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var created models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &created))
	s.NotZero(created.ID)
	s.Equal("BTC", created.Symbol)
	s.Equal("Bitcoin", created.Name)

	s.createdAsset = &created
}

func (s *ControllerTestSuite) Test03_Asset_CreateSecond() {
	asset := models.Asset{
		Symbol: "ETH",
		Name:   "Ethereum",
	}
	body, _ := json.Marshal(asset)

	req := httptest.NewRequest(http.MethodPost, "/api/assets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var created models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &created))
	s.createdAsset2 = &created
}

func (s *ControllerTestSuite) Test04_Asset_Get() {
	s.Require().NotNil(s.createdAsset)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/assets/%d", s.createdAsset.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var asset models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &asset))
	s.Equal(s.createdAsset.ID, asset.ID)
	s.Equal("BTC", asset.Symbol)
}

func (s *ControllerTestSuite) Test05_Asset_GetNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test06_Asset_GetInvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets/invalid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) Test07_Asset_Update() {
	s.Require().NotNil(s.createdAsset)

	updated := models.Asset{
		Symbol: "BTC",
		Name:   "Bitcoin (Updated)",
	}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/assets/%d", s.createdAsset.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var asset models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &asset))
	s.Equal("Bitcoin (Updated)", asset.Name)
}

func (s *ControllerTestSuite) Test08_Asset_UpdateNotFound() {
	updated := models.Asset{Symbol: "XRP", Name: "Ripple"}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/assets/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test09_Asset_List() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var assets []models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &assets))
	s.Len(assets, 2)
}

func (s *ControllerTestSuite) Test10_Asset_CreateInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/api/assets", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// Transaction Tests

func (s *ControllerTestSuite) Test20_Transaction_ListEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.TransactionListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Empty(result.Transactions)
	s.Equal(int64(0), result.Total)
}

func (s *ControllerTestSuite) Test21_Transaction_Create() {
	s.Require().NotNil(s.createdAsset)

	tx := models.Transaction{
		Type:      "deposit",
		AssetID:   s.createdAsset.ID,
		Amount:    1.5,
		Notes:     "Initial deposit",
		Timestamp: time.Now(),
	}
	body, _ := json.Marshal(tx)

	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var created models.Transaction
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &created))
	s.NotZero(created.ID)
	s.Equal("deposit", created.Type)
	s.Equal(1.5, created.Amount)

	s.createdTransaction = &created
}

func (s *ControllerTestSuite) Test22_Transaction_Get() {
	s.Require().NotNil(s.createdTransaction)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transactions/%d", s.createdTransaction.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var tx models.Transaction
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &tx))
	s.Equal(s.createdTransaction.ID, tx.ID)
}

func (s *ControllerTestSuite) Test23_Transaction_GetNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/api/transactions/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test24_Transaction_GetInvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/api/transactions/invalid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) Test25_Transaction_Update() {
	s.Require().NotNil(s.createdTransaction)

	updated := models.Transaction{
		Type:      "deposit",
		AssetID:   s.createdAsset.ID,
		Amount:    2.5,
		Notes:     "Updated deposit",
		Timestamp: time.Now(),
	}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/transactions/%d", s.createdTransaction.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var tx models.Transaction
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &tx))
	s.Equal(2.5, tx.Amount)
	s.Equal("Updated deposit", tx.Notes)
}

func (s *ControllerTestSuite) Test26_Transaction_UpdateNotFound() {
	updated := models.Transaction{Type: "deposit", AssetID: 1, Amount: 1.0}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/transactions/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test27_Transaction_ListWithFilters() {
	s.Require().NotNil(s.createdAsset)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transactions?asset_id=%d&type=deposit", s.createdAsset.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.TransactionListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Len(result.Transactions, 1)
}

// Exchange Tests

func (s *ControllerTestSuite) Test40_Exchange_ListEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.ExchangeListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Empty(result.Exchanges)
}

func (s *ControllerTestSuite) Test41_Exchange_Create() {
	s.Require().NotNil(s.createdAsset)
	s.Require().NotNil(s.createdAsset2)

	exchange := models.Exchange{
		FromAssetID: s.createdAsset.ID,
		ToAssetID:   s.createdAsset2.ID,
		FromAmount:  1.0,
		ToAmount:    15.0,
		Fee:         0.001,
		FeeCurrency: "BTC",
		Notes:       "BTC to ETH swap",
		Timestamp:   time.Now(),
	}
	body, _ := json.Marshal(exchange)

	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var created models.Exchange
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &created))
	s.NotZero(created.ID)
	s.Equal(1.0, created.FromAmount)
	s.Equal(15.0, created.ToAmount)

	s.createdExchange = &created
}

func (s *ControllerTestSuite) Test42_Exchange_CreateSameAsset() {
	s.Require().NotNil(s.createdAsset)

	exchange := models.Exchange{
		FromAssetID: s.createdAsset.ID,
		ToAssetID:   s.createdAsset.ID,
		FromAmount:  1.0,
		ToAmount:    1.0,
	}
	body, _ := json.Marshal(exchange)

	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) Test43_Exchange_Get() {
	s.Require().NotNil(s.createdExchange)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/exchanges/%d", s.createdExchange.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var exchange models.Exchange
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &exchange))
	s.Equal(s.createdExchange.ID, exchange.ID)
}

func (s *ControllerTestSuite) Test44_Exchange_GetNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test45_Exchange_GetInvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/invalid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) Test46_Exchange_Update() {
	s.Require().NotNil(s.createdExchange)

	updated := models.Exchange{
		FromAssetID: s.createdAsset.ID,
		ToAssetID:   s.createdAsset2.ID,
		FromAmount:  2.0,
		ToAmount:    30.0,
		Notes:       "Updated exchange",
		Timestamp:   time.Now(),
	}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/exchanges/%d", s.createdExchange.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var exchange models.Exchange
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &exchange))
	s.Equal(2.0, exchange.FromAmount)
	s.Equal(30.0, exchange.ToAmount)
}

func (s *ControllerTestSuite) Test47_Exchange_UpdateNotFound() {
	updated := models.Exchange{FromAssetID: 1, ToAssetID: 2, FromAmount: 1.0, ToAmount: 1.0}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test48_Exchange_ListWithFilters() {
	s.Require().NotNil(s.createdAsset)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/exchanges?from_asset_id=%d", s.createdAsset.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.ExchangeListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Len(result.Exchanges, 1)
}

// Delete Tests

func (s *ControllerTestSuite) Test90_Exchange_Delete() {
	s.Require().NotNil(s.createdExchange)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/exchanges/%d", s.createdExchange.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/exchanges/%d", s.createdExchange.ID), nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test91_Exchange_DeleteNotFound() {
	req := httptest.NewRequest(http.MethodDelete, "/api/exchanges/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test92_Transaction_Delete() {
	s.Require().NotNil(s.createdTransaction)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/transactions/%d", s.createdTransaction.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transactions/%d", s.createdTransaction.ID), nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test93_Transaction_DeleteNotFound() {
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test94_Asset_Delete() {
	s.Require().NotNil(s.createdAsset)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/assets/%d", s.createdAsset.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/assets/%d", s.createdAsset.ID), nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test95_Asset_DeleteNotFound() {
	req := httptest.NewRequest(http.MethodDelete, "/api/assets/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func TestControllers(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
