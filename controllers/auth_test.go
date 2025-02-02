package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performRequest(method, path, token string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware())

	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	w := performRequest("GET", "/protected", "Bearer "+AUTH_TOKEN)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "access granted"}`, w.Body.String())
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	w := performRequest("GET", "/protected", "Bearer InvalidToken")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"message": "Unauthorized"}`, w.Body.String())
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	w := performRequest("GET", "/protected", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"message": "Unauthorized"}`, w.Body.String())
}
