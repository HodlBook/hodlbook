package controllers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAssetController_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Correct the absolute path for loading templates
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(filepath.Dir(b)))
	r.LoadHTMLGlob(filepath.Join(basePath, "templates/**/*"))

	ctrl, err := NewAsset()
	require.NoError(t, err)

	r.GET("/assets", ctrl.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
