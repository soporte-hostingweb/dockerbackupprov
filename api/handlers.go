package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthMiddleware asegura la autenticación del Agente o Dashboard (V14: Hardening)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Intentar Autenticación por AgentKey (V13/V14: SaaS Mode)
		agentID := c.GetHeader("X-Agent-ID")
		agentKey := c.GetHeader("X-Agent-Key")
		agentFingerprint := c.GetHeader("X-Agent-Fingerprint")

		if agentID != "" && agentKey != "" {
			var agent AgentStatus
			if err := DB.First(&agent, "id = ?", agentID).Error; err == nil {
				// Validar ApiKey (Hashed)
				errK := bcrypt.CompareHashAndPassword([]byte(agent.ApiKey), []byte(agentKey))
				if errK == nil {
					// VALIDACIÓN CRÍTICA: La huella debe coincidir con la registrada (Anti-Clonado)
					if agent.Fingerprint != "" && agent.Fingerprint != agentFingerprint {
						fmt.Printf("[SECURITY] Fingerprint mismatch for agent %s. Blocked.\n", agentID)
						c.JSON(http.StatusForbidden, gin.H{"error": "Licensing error: Hardware signature mismatch."})
						c.Abort()
						return
					}
					
					c.Set("token", agent.Token) // Heredar token del tenant para consultas internas
					c.Set("agent_metadata", agent)
					c.Set("is_admin", false)
					c.Next()
					return
				}
			}
		}

		// 2. Fallback: Autenticación por Token de Tenant / Admin (Dashboard)
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("X-API-Token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing security credentials."})
			c.Abort()
			return
		}
		
		cleanToken := strings.TrimPrefix(token, "Bearer ")
		cleanToken = strings.TrimSpace(cleanToken)

		masterToken := os.Getenv("MASTER_ADMIN_TOKEN")
		isAdmin := (cleanToken != "" && masterToken != "" && cleanToken == masterToken)
		
		if !isAdmin && cleanToken != "vps_token_dev" && !strings.HasPrefix(cleanToken, "dbp_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid security token"})
			c.Abort()
			return
		}

		c.Set("token", cleanToken) 
		c.Set("is_admin", isAdmin)
		c.Next()
	}
}

// Global Log Helper
func Log(msg string) {
	fmt.Printf("[API] %s\n", msg)
}
