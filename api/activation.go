package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ActivationRequest: Payload del handshake inicial (V13)
type ActivationRequest struct {
	Token       string `json:"token"` // dbp_saas_xxxx
	Fingerprint string `json:"fingerprint"`
	OS          string `json:"os"`
	Hostname    string `json:"hostname"`
}

// RegisterActivationHandlers monta los endpoints de licenciamiento (V13)
func RegisterActivationHandlers(r *gin.Engine) {
	r.POST("/v1/activate", func(c *gin.Context) {
		var req ActivationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activation payload"})
			return
		}

		var saasToken ActivationToken
		if err := DB.Where("token = ?", req.Token).First(&saasToken).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Activation token not found"})
			return
		}

		// 1. Validar Ciclo de Vida (48h)
		if saasToken.Status == "expired" || time.Now().After(saasToken.ExpiresAt) {
			if saasToken.Status == "pending" {
				DB.Model(&saasToken).Update("status", "expired")
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Token expired. Please request a new one."})
			return
		}

		if saasToken.Status == "revoked" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token revoked by administrator"})
			return
		}

		// 2. Validar Huella Digital (Anti-Clonado)
		if saasToken.Status == "activated" {
			if saasToken.Fingerprint != req.Fingerprint {
				fmt.Printf("[SECURITY] Blocking activation attempt for token %s. Fingerprint mismatch.\n", req.Token)
				c.JSON(http.StatusForbidden, gin.H{"error": "Licensing error: Token already bound to another server."})
				return
			}
			// Permitir reinstalación en el mismo hardware: Recuperar credenciales originales
			var agent AgentStatus
			if err := DB.First(&agent, "id = ?", saasToken.AgentID).Error; err == nil {
				c.JSON(http.StatusOK, gin.H{
					"status":   "re-activated",
					"agent_id": agent.ID,
					"api_key":  "REDACTED_USE_ORIGINAL", // En una implementación real, aquí se gestionaría la rotación o recuperación
					"ghcr_pat": os.Getenv("GHCR_READ_PAT"),
				})
				return
			}
		}

		// 3. Activación de Nuevo Nodo
		fmt.Printf("[SAAS] Activating new node: %s (%s)\n", req.Hostname, req.Fingerprint)
		
		agentID := fmt.Sprintf("agt_%s", hex.EncodeToString(generateRandomBytes(6)))
		apiKeyRaw := hex.EncodeToString(generateRandomBytes(32))
		hashedKey, _ := bcrypt.GenerateFromPassword([]byte(apiKeyRaw), bcrypt.DefaultCost)

		// Guardar Estado del Agente
		newAgent := AgentStatus{
			ID:          agentID,
			Token:       saasToken.TenantToken,
			Fingerprint: req.Fingerprint,
			ApiKey:      string(hashedKey),
			OS:          req.OS,
			Status:      "Healthy",
			HealthStatus: "ONLINE",
			LastSeen:    time.Now().UTC(),
		}

		if err := DB.Create(&newAgent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent identity"})
			return
		}

		// Actualizar Token a ACTIVADO
		DB.Model(&saasToken).Updates(map[string]interface{}{
			"status":      "activated",
			"fingerprint": req.Fingerprint,
			"agent_id":    agentID,
		})

		c.JSON(http.StatusOK, gin.H{
			"status":   "activated",
			"agent_id": agentID,
			"api_key":  apiKeyRaw,
			"ghcr_pat": os.Getenv("GHCR_READ_PAT"),
		})
	})

	// --- ENDPOINTS ADMINISTRATIVOS (Requieren MASTER_ADMIN_TOKEN) ---

	r.POST("/v1/admin/license/generate", AuthMiddleware(), func(c *gin.Context) {
		if !c.GetBool("is_admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			return
		}

		var req struct {
			TenantToken string `json:"tenant_token" binding:"required"`
			CustomToken string `json:"custom_token"` // Opcional: Para el flujo de prueba del usuario
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: tenant_token is required"})
			return
		}

		tokenStr := req.CustomToken
		if tokenStr == "" {
			tokenStr = fmt.Sprintf("dbp_saas_%s", hex.EncodeToString(generateRandomBytes(8)))
		}

		newToken := ActivationToken{
			Token:       tokenStr,
			TenantToken: req.TenantToken,
			Status:      "pending",
			ExpiresAt:   time.Now().Add(48 * time.Hour),
		}

		if err := DB.Create(&newToken).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "generated",
			"token":        newToken.Token,
			"expires_at":   newToken.ExpiresAt,
			"tenant":       newToken.TenantToken,
			"instructions": "Use this token once within 48h.",
		})
	})
}

// generateRandomBytes helper para tokens
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// RunTokenExpirationWorker: Invalida tokens de 48h (V13)
func RunTokenExpirationWorker() {
	fmt.Println("[WORKER] Token Expiration Service (V13) started.")
	for {
		time.Sleep(1 * time.Hour)
		now := time.Now().UTC()
		
		result := DB.Model(&ActivationToken{}).
			Where("status = ? AND expires_at < ?", "pending", now).
			Update("status", "expired")
		
		if result.RowsAffected > 0 {
			fmt.Printf("[WORKER] Marked %d activation tokens as EXPIRED.\n", result.RowsAffected)
		}
	}
}
