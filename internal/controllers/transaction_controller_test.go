package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTransactionTestController(t *testing.T) *TransactionController {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&repo.Transaction{}))
	txRepo, err := repo.NewTransaction(repo.WithDB(db))
	require.NoError(t, err)
	ctrl, err := NewTransaction(WithTransactionRepo(txRepo))
	require.NoError(t, err)
	return ctrl
}

func TestTransactionController_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := setupTransactionTestController(t)
	r.GET("/transactions", ctrl.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/transactions", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
}
