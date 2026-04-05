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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header."})
			c.Abort()
			return
		}
		
		// MVP: Permitimos tokens que empiecen con dbp_tenant_ o el de dev
		masterToken := os.Getenv("MASTER_ADMIN_TOKEN")
		isAdmin := (token != "" && masterToken != "" && token == masterToken)
		
		if !isAdmin && token != "Bearer vps_token_dev" && token != "vps_token_dev" && !strings.HasPrefix(token, "dbp_tenant_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
			c.Abort()
			return
		}

		c.Set("token", token)
		c.Set("is_admin", isAdmin)
		c.Next()
	}
}

// Global Log Helper
func Log(msg string) {
	fmt.Printf("[API] %s\n", msg)
}
