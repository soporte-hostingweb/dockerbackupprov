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
		isAdmin := (token == os.Getenv("MASTER_ADMIN_TOKEN"))
		
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


// ReceiveHeartbeat recibe la información pasiva del servidor del cliente para mostrarlo "Verde" en el UI
func ReceiveHeartbeat(c *gin.Context) {
	var payload HeartbeatPayload
	token := c.GetString("token")
	
	if err := c.ShouldBindJSON(&payload); err != nil {
		fmt.Printf("[API ERROR] Malformed heartbeat from %s: %v\n", token, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload format"})
		return
	}

	fmt.Printf("[HEARTBEAT] Received from Agent: %s (Token: %s...)\n", payload.AgentID, token[:10])

	agentStatusStore[payload.AgentID] = gin.H{
		"token":       token,

		"agent_id":    payload.AgentID,
		"status":      "Healthy",
		"last_sync":   "Just now",
		"containers":  payload.Containers,
		"explorer":    payload.ExplorerData,
		"os":          payload.OS,
		"type":        "Heartbeat",
	}



	
	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat updated"})
}

// ReceiveBackupCompletion registra cuando un trabajo de Restic finaliza con sus métricas completas
func ReceiveBackupCompletion(c *gin.Context) {
	var payload BackupCompletePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Almacenamos en memoria para el Dashboard
	agentStatusStore[payload.AgentID] = gin.H{
		"token":         c.GetString("token"),
		"agent_id":      payload.AgentID,
		"status":        payload.Status,
		"total_size_mb": payload.TotalSizeMB,
		"snapshot_id":   payload.SnapshotID,
		"last_sync":     "1 min ago",
		"health":        "Healthy",
	}


	fmt.Printf("[DB INSERT] Agent %s reported snapshot %s\n", payload.AgentID, payload.SnapshotID)
	c.JSON(http.StatusOK, gin.H{"message": "Metrics logged", "id": payload.SnapshotID})
}
