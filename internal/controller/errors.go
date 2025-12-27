package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrNilRepository = errors.New("repository cannot be nil")
)

type APIError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func errorResponse(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, APIError{Error: message})
}

func errorWithDetails(ctx *gin.Context, status int, message string, details string) {
	ctx.JSON(status, APIError{Error: message, Details: details})
}

func badRequest(ctx *gin.Context, message string) {
	errorResponse(ctx, http.StatusBadRequest, message)
}

func badRequestWithDetails(ctx *gin.Context, message string, details string) {
	errorWithDetails(ctx, http.StatusBadRequest, message, details)
}

func notFound(ctx *gin.Context, message string) {
	errorResponse(ctx, http.StatusNotFound, message)
}

func internalError(ctx *gin.Context, message string) {
	errorResponse(ctx, http.StatusInternalServerError, message)
}

func serviceUnavailable(ctx *gin.Context, message string) {
	errorResponse(ctx, http.StatusServiceUnavailable, message)
}
