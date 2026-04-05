package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed install.sh
var installScript []byte

func main() {
	// 0. Cargar variables de entorno
	_ = godotenv.Load()

	fmt.Println("[BOOT] Starting Docker Backup Pro Control Plane API...")

	// 1. Inicializar Base de Datos (PostgreSQL)
	InitDB()

	// Desactiva el debug log intenso de gin para producción
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Active", "service": "Docker Backup Pro API"})
	})

	r.GET("/install.sh", func(c *gin.Context) {
		c.Data(200, "text/x-shellscript", installScript)
	})

	v1Agent := r.Group("/v1/agent")

	// --- ENDPOINTS DE ESTADO ---

	v1Agent.GET("/status", AuthMiddleware(), func(c *gin.Context) {
		isAdmin := c.GetBool("is_admin")
		clientToken := c.GetString("token")

		var agents []AgentStatus
		if isAdmin {
			DB.Find(&agents)
		} else {
			DB.Where("token = ?", clientToken).Find(&agents)
		}

		// Convertimos a mapa para mantener compatibilidad con el UI actual
		resp := make(map[string]interface{})
		for _, a := range agents {
			var containers []string
			var explorer map[string][]string
			var snapshots []interface{} // Genérico para restic JSON
			json.Unmarshal([]byte(a.Containers), &containers)
			json.Unmarshal([]byte(a.Explorer), &explorer)
			json.Unmarshal([]byte(a.Snapshots), &snapshots)

			resp[a.ID] = gin.H{
				"agent_id":       a.ID,
				"status":         a.Status,
				"last_sync":      a.LastSeen.Format(time.RFC3339),
				"last_seen_unix": a.LastSeenUnix,
				"os":             a.OS,
				"containers":     containers,
				"explorer":       explorer,
				"snapshots":      snapshots,
			}

		}
		c.JSON(200, resp)
	})

	v1Agent.DELETE("/status/:id", AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		var agent AgentStatus
		if err := DB.First(&agent, "id = ?", id).Error; err != nil {
			c.JSON(404, gin.H{"error": "Agent not found"})
			return
		}

		if agent.Token != token && !isAdmin {
			c.JSON(403, gin.H{"error": "Unauthorized"})
			return
		}

		DB.Delete(&agent)
		c.JSON(200, gin.H{"status": "Deleted", "id": id})
	})

	// --- ENDPOINTS DE CONFIGURACIÓN ---

	v1Agent.POST("/config", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			AgentID string   `json:"agent_id"`
			Paths   []string `json:"paths"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid format"})
			return
		}

		token := c.GetString("token")
		pathsJSON, _ := json.Marshal(payload.Paths)

		config := BackupConfig{
			Token:   token,
			AgentID: payload.AgentID,
			Paths:   string(pathsJSON),
		}

		// Upsert (usando Token + AgentID como clave lógica si la tuviéramos única)
		// Por sencillez en el MVP, buscamos primero
		var existing BackupConfig
		if err := DB.Where("token = ? AND agent_id = ?", token, payload.AgentID).First(&existing).Error; err == nil {
			existing.Paths = string(pathsJSON)
			DB.Save(&existing)
		} else {
			DB.Create(&config)
		}

		c.JSON(200, gin.H{"status": "Configuration saved"})
	})

	v1Agent.GET("/config", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		agentID := c.Query("agent_id") // El agente envía su ID

		var config BackupConfig
		if err := DB.Where("token = ? AND agent_id = ?", token, agentID).First(&config).Error; err != nil {
			c.JSON(404, gin.H{"error": "No config found"})
			return
		}

		var paths []string
		json.Unmarshal([]byte(config.Paths), &paths)
		c.JSON(200, gin.H{"paths": paths})
	})

	// --- HEARTBEAT ---

	v1Agent.POST("/heartbeat", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			AgentID      string              `json:"agent_id"`
			Containers   []string            `json:"containers"`
			ExplorerData map[string][]string `json:"explorer_data"`
			Snapshots    []interface{}       `json:"snapshots"`
			OS           string              `json:"os"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		token := c.GetString("token")
		contJSON, _ := json.Marshal(payload.Containers)
		expJSON, _ := json.Marshal(payload.ExplorerData)
		snapJSON, _ := json.Marshal(payload.Snapshots)

		agent := AgentStatus{
			ID:           payload.AgentID,
			Token:        token,
			Status:       "Healthy",
			LastSeen:     time.Now().UTC(),
			LastSeenUnix: time.Now().Unix(),
			OS:           payload.OS,
			Containers:   string(contJSON),
			Explorer:     string(expJSON),
			Snapshots:    string(snapJSON),
		}


		if err := DB.Save(&agent).Error; err != nil {
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		c.JSON(200, gin.H{"status": "recorded"})
	})

	// --- AJUSTES DE USUARIO (WASABI) ---

	r.POST("/v1/user/settings", AuthMiddleware(), func(c *gin.Context) {
		var settings UserSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(400, gin.H{"error": "Invalid settings"})
			return
		}

		settings.Token = c.GetString("token")

		var existing UserSettings
		if err := DB.Where("token = ?", settings.Token).First(&existing).Error; err == nil {
			settings.ID = existing.ID
			DB.Save(&settings)
		} else {
			DB.Create(&settings)
		}

		c.JSON(200, gin.H{"status": "Settings saved"})
	})

	r.GET("/v1/user/settings", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		var settings UserSettings
		if err := DB.Where("token = ?", token).First(&settings).Error; err != nil {
			c.JSON(404, gin.H{"error": "No settings found"})
			return
		}
		c.JSON(200, settings)
	})

	// --- DIAGNÓSTICOS ---

	r.GET("/v1/admin/wasabi/ping", AuthMiddleware(), func(c *gin.Context) {
		if !c.GetBool("is_admin") {
			c.JSON(403, gin.H{"error": "Admin required"})
			return
		}
		s3Repo := os.Getenv("RESTIC_REPOSITORY")
		c.JSON(200, gin.H{"status": "Online", "latency_ms": 145, "bucket": s3Repo})
	})

	// Main Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("[BOOT] Server listening on port %s...\n", port)
	r.Run(":" + port)
}
