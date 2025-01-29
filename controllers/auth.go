package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const AUTH_TOKEN = "kVvKz9TEvWAgQs3MtSJqTXSjtLfMMcyD"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func validateToken(token string) bool {
	return token == fmt.Sprintf("Bearer %s", AUTH_TOKEN)
}
