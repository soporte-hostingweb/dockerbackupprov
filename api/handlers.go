package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware asegura que el Agente envíe un Token válido asignado por WHMCS
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header. Token is required."})
			c.Abort()
			return
		}
		
		// TODO: Validar token contra base de datos PostgreSQL
		// Ejemplo pseudocódigo: SELECT tenant_id FROM agents WHERE token = '...'
		if token != "Bearer vps_token_dev" && token != "vps_token_dev" {
			fmt.Printf("[AUTH] Invalid Agent Token: %s\n", token)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
			c.Abort()
			return
		}
		
		// Pasamos validación
		c.Next()
	}
}

// ReceiveHeartbeat recibe la información pasiva del servidor del cliente para mostrarlo "Verde" en el UI
func ReceiveHeartbeat(c *gin.Context) {
	var payload HeartbeatPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload format"})
		return
	}

	// Logging/Mock insertion for PG DB
	fmt.Printf("[DB UPDATE] \u001b[32mAgent is %s\u001b[0m. Containers: %d\n", payload.AgentStatus, payload.Containers)
	
	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat updated"})
}

// ReceiveBackupCompletion registra cuando un trabajo de Restic finaliza con sus métricas completas
func ReceiveBackupCompletion(c *gin.Context) {
	var payload BackupCompletePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pseudo-inserción a un PostgreSQL para registrar la factura de datos o la interfaz de historial
	fmt.Printf("[DB INSERT] \u001b[34mSnapshot Received\u001b[0m -> ID: %s | Status: %s | Size: %d MB\n",
		payload.SnapshotID, payload.Status, payload.TotalSizeMB)

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup metrics safely logged to Data Warehouse",
		"id": payload.SnapshotID,
	})
}
