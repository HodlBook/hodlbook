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

	createdAsset    *models.Asset
	createdAsset2   *models.Asset
	createdExchange *models.Exchange
}

func (s *ControllerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.Require().NoError(db.AutoMigrate(&models.Asset{}, &models.Exchange{}, &models.AssetHistoricValue{}))
	s.db = db

	repository, err := repo.New(db)
	s.Require().NoError(err)

	ctrl, err := New(WithRepository(repository))
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

	exchanges := api.Group("/exchanges")
	exchanges.GET("", ctrl.ListExchanges)
	exchanges.POST("", ctrl.CreateExchange)
	exchanges.GET("/:id", ctrl.GetExchange)
	exchanges.PUT("/:id", ctrl.UpdateExchange)
	exchanges.DELETE("/:id", ctrl.DeleteExchange)

	portfolio := api.Group("/portfolio")
	portfolio.GET("/summary", ctrl.PortfolioSummary)
	portfolio.GET("/allocation", ctrl.PortfolioAllocation)
	portfolio.GET("/performance", ctrl.PortfolioPerformance)
	portfolio.GET("/history", ctrl.PortfolioHistory)
}

// Asset Tests

func (s *ControllerTestSuite) Test01_Asset_ListEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/api/assets", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.AssetListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Empty(result.Assets)
}

func (s *ControllerTestSuite) Test02_Asset_Create() {
	asset := models.Asset{
		Symbol:          "BTC",
		Name:            "Bitcoin",
		Amount:          1.5,
		TransactionType: "deposit",
		Timestamp:       time.Now(),
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
	s.Equal(1.5, created.Amount)
	s.Equal("deposit", created.TransactionType)

	s.createdAsset = &created
}

func (s *ControllerTestSuite) Test03_Asset_CreateSecond() {
	asset := models.Asset{
		Symbol:          "ETH",
		Name:            "Ethereum",
		Amount:          10.0,
		TransactionType: "deposit",
		Timestamp:       time.Now(),
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
		Symbol:          "BTC",
		Name:            "Bitcoin Updated",
		Amount:          2.0,
		TransactionType: "deposit",
		Timestamp:       time.Now(),
	}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/assets/%d", s.createdAsset.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var asset models.Asset
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &asset))
	s.Equal("Bitcoin Updated", asset.Name)
	s.Equal(2.0, asset.Amount)
}

func (s *ControllerTestSuite) Test08_Asset_UpdateNotFound() {
	updated := models.Asset{Symbol: "XRP", Name: "Ripple", Amount: 1.0, TransactionType: "deposit"}
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

	var result repo.AssetListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Len(result.Assets, 2)
}

func (s *ControllerTestSuite) Test10_Asset_CreateInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/api/assets", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
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
	exchange := models.Exchange{
		FromSymbol:  "BTC",
		ToSymbol:    "ETH",
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

func (s *ControllerTestSuite) Test42_Exchange_CreateSameSymbol() {
	exchange := models.Exchange{
		FromSymbol: "BTC",
		ToSymbol:   "BTC",
		FromAmount: 1.0,
		ToAmount:   1.0,
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
		FromSymbol: "BTC",
		ToSymbol:   "ETH",
		FromAmount: 2.0,
		ToAmount:   30.0,
		Notes:      "Updated exchange",
		Timestamp:  time.Now(),
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
	updated := models.Exchange{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 1.0, ToAmount: 1.0}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ControllerTestSuite) Test48_Exchange_ListWithFilters() {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges?from_symbol=BTC", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result repo.ExchangeListResult
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))
	s.Len(result.Exchanges, 1)
}

func (s *ControllerTestSuite) Test49_Exchange_CreateInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// Portfolio Tests

func (s *ControllerTestSuite) Test60_Portfolio_Summary() {
	req := httptest.NewRequest(http.MethodGet, "/api/portfolio/summary", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Contains(result, "total_value")
	s.Contains(result, "currency")
	s.Contains(result, "holdings")
	s.Equal("USD", result["currency"])
}

func (s *ControllerTestSuite) Test61_Portfolio_Allocation() {
	req := httptest.NewRequest(http.MethodGet, "/api/portfolio/allocation", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Contains(result, "total_value")
	s.Contains(result, "allocations")
}

func (s *ControllerTestSuite) Test62_Portfolio_Performance() {
	req := httptest.NewRequest(http.MethodGet, "/api/portfolio/performance", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Contains(result, "total_cost_basis")
	s.Contains(result, "total_current_value")
	s.Contains(result, "total_profit_loss")
}

func (s *ControllerTestSuite) Test63_Portfolio_History() {
	req := httptest.NewRequest(http.MethodGet, "/api/portfolio/history", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var result map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Contains(result, "days")
	s.Contains(result, "history")
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

	s.Equal(http.StatusNoContent, w.Code)
}

func (s *ControllerTestSuite) Test92_Asset_Delete() {
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

func (s *ControllerTestSuite) Test93_Asset_DeleteNotFound() {
	req := httptest.NewRequest(http.MethodDelete, "/api/assets/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)
}

func TestControllers(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
