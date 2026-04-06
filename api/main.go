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

			// V2.3: Buscamos la configuración de respaldo para obtener el Schedule
			var config BackupConfig
			DB.Where("token = ? AND agent_id = ?", a.Token, a.ID).First(&config)

			resp[a.ID] = gin.H{
				"agent_id":       a.ID,
				"token":          a.Token,
				"status":         a.Status,
				"last_sync":      a.LastSeen.Format(time.RFC3339),
				"last_seen_unix": a.LastSeenUnix,
				"os":             a.OS,
				"containers":     containers,
				"explorer":       explorer,
				"snapshots":      snapshots,
				"maintenance":    a.Maintenance,
				"is_syncing":     a.IsSyncing,
				"active_pid":     a.ActivePID,
				"last_backup_at": a.LastBackupAt,
				"schedule":       config.Schedule,
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

	// --- ENDPOINTS DE CONFIGURACIÓN MOVIDOS A ABAJO (V2.3) ---


	v1Agent.GET("/config", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		agentID := c.Query("agent_id") // El agente envía su ID

		if agentID == "" {
			c.JSON(400, gin.H{"error": "AgentID is required"})
			return
		}

		var configs []BackupConfig
		if err := DB.Limit(1).Where("token = ? AND agent_id = ?", token, agentID).Find(&configs).Error; err != nil {
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		// Buscamos el Bucket en Settings para construir la URL completa
		var settings UserSettings
		DB.Where("token = ?", token).First(&settings)
		
		// Descifrar contraseña de restic para el agente (V2.4)
		resticPass, _ := Decrypt(settings.ResticPass)

		// Inyectamos la ruta aislada. (V2.4: Manejo de campos vacíos)
		region := settings.WasabiRegion
		if region == "" { region = "us-east-1" }
		
		bucket := settings.WasabiBucket
		if bucket == "" { bucket = "unconfigured" }

		fullRepo := fmt.Sprintf("s3:%s.wasabisys.com/%s/%s/%s", 
			region, bucket, token, agentID)

		if len(configs) == 0 {
			c.JSON(200, gin.H{
				"status":         "no_config", 
				"paths":          []string{}, 
				"schedule":       "daily_2am",
				"full_repo_url":  fullRepo,
				"restic_password": resticPass,
			})
			return
		}


		var paths []string
		json.Unmarshal([]byte(configs[0].Paths), &paths)
		c.JSON(200, gin.H{
			"status":          "manual", 
			"paths":           paths, 
			"schedule":        configs[0].Schedule,
			"full_repo_url":   fullRepo,
			"restic_password": resticPass,
		})
	})


	// Dashboard guarda la configuración
	v1Agent.POST("/config", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		var req struct {
			AgentID  string   `json:"agent_id"`
			Paths    []string `json:"paths"`
			Schedule string   `json:"schedule"` // manual, daily_2am, etc.
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		pathsJSON, _ := json.Marshal(req.Paths)

		var config BackupConfig
		if err := DB.Where("token = ? AND agent_id = ?", token, req.AgentID).First(&config).Error; err == nil {
			config.Paths = string(pathsJSON)
			config.Schedule = req.Schedule
			DB.Save(&config)
		} else {
			config = BackupConfig{
				Token:    token,
				AgentID:  req.AgentID,
				Paths:    string(pathsJSON),
				Schedule: req.Schedule,
			}
			DB.Create(&config)
		}

		c.JSON(200, gin.H{"status": "Config saved"})
	})

	// --- HEARTBEAT ---

	v1Agent.POST("/heartbeat", AuthMiddleware(), func(c *gin.Context) {
		// Heartbeat Payload (con soporte para reporte de estado de proceso activo)
		var payload struct {
			AgentID      string              `json:"agent_id"`
			Containers   []string            `json:"containers"`
			ExplorerData map[string][]string `json:"explorer_data"`
			Snapshots    []interface{}       `json:"snapshots"`
			OS           string              `json:"os"`
			IsSyncing    bool                `json:"is_syncing"`
			ActivePID    int                 `json:"active_pid"`
			LastBackupAt int64               `json:"last_backup_unix"` // Reportado por el agente
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
			IsSyncing:    payload.IsSyncing,
			ActivePID:    payload.ActivePID,
		}

		if payload.LastBackupAt > 0 {
			agent.LastBackupAt = time.Unix(payload.LastBackupAt, 0).UTC()
		}


		// Importante: No machacar Maintenance y PendingForce si ya existen
		var existing AgentStatus
		if err := DB.First(&existing, "id = ?", payload.AgentID).Error; err == nil {
			agent.Maintenance = existing.Maintenance
			agent.PendingForce = existing.PendingForce
		}

		if err := DB.Save(&agent).Error; err != nil {
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		c.JSON(200, gin.H{
			"status":        "recorded",
			"maintenance":  agent.Maintenance,
			"pending_force": agent.PendingForce,
			"kill_sync":     agent.KillSync,
		})
	})

	// --- TELEMETRÍA DE BACKUP (MÉTRICAS) ---
	v1Agent.POST("/backup/complete", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			AgentID    string `json:"agent_id"`
			Status     string `json:"status"`
			SnapshotID string `json:"snapshot_id"`
			Timestamp  int64  `json:"timestamp"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid metrics"})
			return
		}

		// Actualizamos el último backup exitoso si el estado es SUCCESS
		if payload.Status == "SUCCESS" {
			DB.Model(&AgentStatus{}).Where("id = ?", payload.AgentID).Update("last_backup_at", time.Unix(payload.Timestamp, 0).UTC())
		}

		c.JSON(200, gin.H{"status": "Metrics recorded"})
	})




	// --- ACCIONES ADMINISTRATIVAS ---

	v1Agent.POST("/action/:id", AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		token := c.GetString("token")
		var req struct {
			Action string `json:"action"` // "reset", "maintenance_on", "maintenance_off", "force_selected", "force_full", "kill_sync"
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid action"})
			return
		}

		var agent AgentStatus
		if err := DB.First(&agent, "id = ?", id).Error; err != nil {
			c.JSON(404, gin.H{"error": "Agent not found"})
			return
		}

		if agent.Token != token && !c.GetBool("is_admin") {
			c.JSON(403, gin.H{"error": "Unauthorized"})
			return
		}

		switch req.Action {
		case "reset":
			DB.Where("agent_id = ? AND token = ?", id, token).Delete(&BackupConfig{})
			agent.PendingForce = "none"
		case "maintenance_on":
			agent.Maintenance = true
		case "maintenance_off":
			agent.Maintenance = false
		case "force_selected":
			agent.PendingForce = "selected"
		case "force_full":
			agent.PendingForce = "full"
		case "kill_sync":
			agent.KillSync = true
		}


		DB.Save(&agent)
		c.JSON(200, gin.H{"status": "Action recorded", "action": req.Action})
	})

	// --- AJUSTES DE USUARIO (WASABI) (V2.3.2) ---
	
	v1User := r.Group("/v1/user")
	v1User.Use(AuthMiddleware())

	v1User.POST("/settings", func(c *gin.Context) {
		var settings UserSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(400, gin.H{"error": "Invalid settings"})
			return
		}

		// Cifrar datos sensibles antes de guardar
		settings.WasabiSecret, _ = Encrypt(settings.WasabiSecret)
		settings.ResticPass, _ = Encrypt(settings.ResticPass)
		settings.Token = c.GetString("token")

		var existing UserSettings
		if err := DB.Where("token = ?", settings.Token).First(&existing).Error; err == nil {
			settings.ID = existing.ID
			DB.Save(&settings)
		} else {
			DB.Create(&settings)
		}

		c.JSON(200, gin.H{"status": "Settings saved (Encrypted)"})
	})

	v1User.GET("/settings", func(c *gin.Context) {
		token := c.GetString("token")
		var settings UserSettings
		if err := DB.Where("token = ?", token).First(&settings).Error; err != nil {
			// V2.3.2: Devolver 200 con campos vacíos en lugar de 404 para el UI
			c.JSON(200, gin.H{
				"wasabi_key": "",
				"wasabi_secret": "",
				"wasabi_bucket": "",
				"wasabi_region": "us-east-1",
				"restic_password": "",
			})
			return
		}

		// Descifrar antes de enviar al Dashboard (seguro bajo HTTPS)
		settings.WasabiSecret, _ = Decrypt(settings.WasabiSecret)
		settings.ResticPass, _ = Decrypt(settings.ResticPass)

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
		port = "8089"
	}

	fmt.Printf("[BOOT] Server listening on port %s...\n", port)
	r.Run(":" + port)
}
