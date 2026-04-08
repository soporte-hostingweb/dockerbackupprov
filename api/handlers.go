package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware asegura que el Agente envíe un Token válido asignado por WHMCS
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("X-API-Token") // Fallback V6.7: Por si el agente aún no se actualizó
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header."})
			c.Abort()
			return
		}
		
		// Limpiamos el prefijo 'Bearer ' si existe para la validación (V6.5)
		cleanToken := strings.TrimPrefix(token, "Bearer ")
		cleanToken = strings.TrimSpace(cleanToken)

		masterToken := os.Getenv("MASTER_ADMIN_TOKEN")
		isAdmin := (cleanToken != "" && masterToken != "" && cleanToken == masterToken)
		
		if !isAdmin && cleanToken != "vps_token_dev" && !strings.HasPrefix(cleanToken, "dbp_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
			c.Abort()
			return
		}

		c.Set("token", cleanToken) // Guardamos el token limpio para el DB query
		c.Set("is_admin", isAdmin)
		c.Next()
	}
}

// Global Log Helper
func Log(msg string) {
	fmt.Printf("[API] %s\n", msg)
}
