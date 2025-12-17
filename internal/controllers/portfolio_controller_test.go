package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPortfolioController_Dashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repository := &repo.Repository{}
	ctrl, err := NewPortfolio(WithRepository(repository))
	require.NoError(t, err)
	r.GET("/", ctrl.Dashboard)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
}
