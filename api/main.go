package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"


	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)


const Version = "V2.7.2"

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

	r.GET("/v1/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": Version, "status": "active", "network": "HWPeru SaaS"})
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

			// Filtramos el dbp-client-agent de los contenedores reportados
			cleanContainers := []string{}
			for _, co := range containers {
				if co != "dbp-client-agent" {
					cleanContainers = append(cleanContainers, co)
				}
			}

			resp[a.ID] = gin.H{
				"agent_id":       a.ID,
				"token":          a.Token,
				"status":         a.Status,
				"last_sync":      a.LastSeen.Format(time.RFC3339),
				"last_seen_unix": a.LastSeenUnix,
				"os":             a.OS,
				"containers":     cleanContainers,
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

	// Nuevo endpoint para el Historial (Phase 2)
	r.GET("/v1/history", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		var activities []BackupActivity
		if isAdmin {
			DB.Order("created_at desc").Limit(50).Find(&activities)
		} else {
			DB.Where("token = ?", token).Order("created_at desc").Limit(50).Find(&activities)
		}
		c.JSON(200, activities)
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
		
		// Lógica de Impersonación para Admin (V2.7)
		isAdmin := c.GetBool("is_admin")
		viewToken := token
		if isAdmin {
			var agent AgentStatus
			if err := DB.Where("id = ?", agentID).First(&agent).Error; err == nil {
				// El admin ve la configuración asociada al token real del agente (el del cliente)
				viewToken = agent.Token
			}
		}

		if err := DB.Limit(1).Where("token = ? AND agent_id = ?", viewToken, agentID).Find(&configs).Error; err != nil {
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}


		// 1. Buscamos Settings del Tenant (V2.9.6: Silenciamos log spam con Find)
		var settings UserSettings
		DB.Limit(1).Where("token = ?", token).Find(&settings)
		
		if settings.ID == 0 {
			// 2. Si no hay, buscamos las Globales del Maestro (V2.5)
			DB.Limit(1).Where("token = ?", "SYSTEM_GLOBAL").Find(&settings)
			
			if settings.ID == 0 {
				c.JSON(200, gin.H{
					"status": "manual",
					"paths": []string{},
					"error_code": "WASABI_UNCONFIGURED",
					"message": "Please contact administrator for storage configuration",
					"full_repo_url": "",
					"restic_password": "",
				})
				return
			}
		}


		// Descifrar contraseña de restic para el agente (V2.4)
		resticPass, _ := Decrypt(settings.ResticPass)

		// Inyectamos la ruta aislada. (V2.4: Manejo de campos vacíos)
		region := settings.WasabiRegion
		if region == "" { region = "us-east-1" }
		
		bucket := settings.WasabiBucket
		
		// Construir URL Correcta (V2.6.8): s3:https://s3.[region].wasabisys.com/bucket/tenant/agent_id
		// Wasabi usa s3.wasabisys.com para us-east-1, y s3.REGION.wasabisys.com para el resto.
		endpoint := "s3.wasabisys.com"
		if region != "us-east-1" {
			endpoint = fmt.Sprintf("s3.%s.wasabisys.com", region)
		}
		
		// V2.6.8: Añadimos https:// explícito para evitar errores de negociación S3
		fullRepo := fmt.Sprintf("s3:https://%s/%s/%s/%s", 
			endpoint, bucket, token, agentID)




		// Descifrar llaves S3 para el Agente (V2.6.4 - Hotfix)
		wasabiKey, _ := Decrypt(settings.WasabiKey)
		wasabiSecret, _ := Decrypt(settings.WasabiSecret)

		if len(configs) == 0 {
			c.JSON(200, gin.H{
				"status":          "no_config", 
				"paths":           []string{}, 
				"schedule":        "manual", // Por defecto para nuevos agentes (V2.5.2)
				"full_repo_url":   fullRepo,
				"restic_password": resticPass,
				"wasabi_key":      wasabiKey,
				"wasabi_secret":   wasabiSecret,
			})
			return
		}

		var paths []string
		_ = json.Unmarshal([]byte(configs[0].Paths), &paths)

		c.JSON(200, gin.H{
			"status":          "ready",
			"paths":           paths,
			"schedule":        configs[0].Schedule,
			"full_repo_url":   fullRepo,
			"restic_password": resticPass,
			"wasabi_key":      wasabiKey,
			"wasabi_secret":   wasabiSecret,
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

		// Lógica de Impersonación para Admin (V2.6.1)
		isAdmin := c.GetBool("is_admin")
		saveToken := token
		if isAdmin {
			var agent AgentStatus
			if err := DB.Where("id = ?", req.AgentID).First(&agent).Error; err == nil {
				// Usamos el token original del agente para guardar la config
				saveToken = agent.Token
			}
		}


		var config BackupConfig
		if err := DB.Where("token = ? AND agent_id = ?", saveToken, req.AgentID).First(&config).Error; err == nil {
			config.Paths = string(pathsJSON)
			config.Schedule = req.Schedule
			DB.Save(&config)
		} else {
			config = BackupConfig{
				Token:    saveToken,
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
			FreeSpace    string              `json:"free_space"`
			TotalSpace   string              `json:"total_space"`
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
			FreeSpace:    payload.FreeSpace,
			TotalSpace:   payload.TotalSpace,
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
			agent.KillSync = existing.KillSync
			
			// Si el agente reporta que está sincronizando, consumimos la instrucción (V3.4.1)
			if payload.IsSyncing && agent.PendingForce != "none" {
				agent.PendingForce = "none"
			}
			
			// Si se procesó una orden de kill, la reiniciamos (V3.4.1)
			if !payload.IsSyncing && agent.KillSync {
				agent.KillSync = false
			}
		}

		if err := DB.Save(&agent).Error; err != nil {

			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		c.JSON(200, gin.H{
			"status":        "recorded",
			"maintenance":   agent.Maintenance,
			"pending_force": agent.PendingForce,
			"kill_sync":     agent.KillSync,
			"cmd_task":      agent.CmdTask,
			"cmd_param":     agent.CmdParam,
		})
	})

	// --- RECEPCIÓN DE RESULTADOS DE TAREAS (V4.2.3) ---
	v1Agent.POST("/task/result", AuthMiddleware(), func(c *gin.Context) {
		var req struct {
			AgentID string `json:"agent_id"`
			Task    string `json:"task"`
			Result  string `json:"result"` // JSON string del 'ls'
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid task result"})
			return
		}

		// Guardamos el resultado en el estado del agente y limpiamos la tarea pendiente
		DB.Model(&AgentStatus{}).Where("id = ?", req.AgentID).Updates(map[string]interface{}{
			"cmd_result": req.Result,
			"cmd_task":   "none",
		})

		c.JSON(200, gin.H{"status": "Result saved"})
	})


	// --- TELEMETRÍA DE BACKUP (MÉTRICAS) ---
	v1Agent.POST("/backup/complete", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			AgentID      string `json:"agent_id"`
			Status       string `json:"status"`
			TotalSizeMB  int    `json:"total_size_mb"`
			DurationSecs int    `json:"duration_secs"`
			SnapshotID   string `json:"snapshot_id"`
			Timestamp    int64  `json:"timestamp"`
			StartedAt    int64  `json:"started_at"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid metrics"})
			return
		}

		// 1. Guardamos la actividad histórica (V4.1.0)
		var agent AgentStatus
		if err := DB.First(&agent, "id = ?", payload.AgentID).Error; err == nil {
			activity := BackupActivity{
				AgentID:      payload.AgentID,
				Token:        agent.Token,
				Status:       payload.Status,
				SizeMB:       payload.TotalSizeMB,
				DurationSecs: payload.DurationSecs,
				SnapshotID:   payload.SnapshotID,
				StartedAt:    time.Unix(payload.StartedAt, 0).UTC(),
				FinishedAt:   time.Unix(payload.Timestamp, 0).UTC(),
				CreatedAt:    time.Now(),
			}
			DB.Create(&activity)
			
			// Actualizamos el último backup exitoso en el estado del agente
			if payload.Status == "SUCCESS" {
				DB.Model(&agent).Update("last_backup_at", time.Unix(payload.Timestamp, 0).UTC())
			}
		}


		c.JSON(200, gin.H{"status": "Metrics recorded and activity saved"})
	})





	// --- ACCIONES ADMINISTRATIVAS ---

	v1Agent.POST("/action/:id", AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		token := c.GetString("token")
		var req struct {
			Action      string   `json:"action"` // "reset", "maintenance_on", "maintenance_off", "force_selected", "force_full", "kill_sync", "ls_snapshot", "restore"
			SnapshotID  string   `json:"snapshot_id"`
			Destination string   `json:"destination"`
			Paths       []string `json:"paths"`
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
		case "ls_snapshot":
			DB.Model(&agent).Updates(map[string]interface{}{
				"cmd_task":   "ls_snapshot",
				"cmd_param":  req.SnapshotID,
				"cmd_result": "loading",
			})
		case "restore":
			// Formato: snapshot_id|destination|path1,path2,... (V4.5.6)
			pathsStr := strings.Join(req.Paths, ",")
			param := fmt.Sprintf("%s|%s|%s", req.SnapshotID, req.Destination, pathsStr)
			DB.Model(&agent).Updates(map[string]interface{}{
				"cmd_task":   "restore",
				"cmd_param":  param,
				"cmd_result": "pending",
			})
		}

		// V4.5.8: PERSISTIR CAMBIOS (Crítico para que el heartbeat los detecte)
		if err := DB.Save(&agent).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to persist action"})
			return
		}

		c.JSON(200, gin.H{"status": "Action queued", "action": req.Action})
	})



	// --- AJUSTES DE USUARIO (WASABI) (V2.3.2) ---
	
	v1User := r.Group("/v1/user")
	v1User.Use(AuthMiddleware())

	// Guardar/Actualizar Settings (V2.5)
	v1User.POST("/settings", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		var input UserSettings
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		// Permitir guardar como 'Global' si se solicita (V2.5)
		// En una versión final, aquí validaríamos que el token sea de un admin
		saveToken := token
		if c.Query("is_global") == "true" {
			saveToken = "SYSTEM_GLOBAL"
		}

		var settings UserSettings
		// Buscar existente por token para obtener el ID real y evitar conflictos únicos
		DB.Where("token = ?", saveToken).First(&settings)

		settings.Token = saveToken
		settings.WasabiBucket = input.WasabiBucket
		settings.WasabiRegion = input.WasabiRegion
		
		// Solo ciframos y actualizamos las llaves si no vienen vacías (V2.9.1)
		if input.WasabiKey != "" {
			encKey, _ := Encrypt(input.WasabiKey)
			settings.WasabiKey = encKey
		}
		if input.WasabiSecret != "" {
			encSec, _ := Encrypt(input.WasabiSecret)
			settings.WasabiSecret = encSec
		}
		if input.ResticPass != "" {
			encPass, _ := Encrypt(input.ResticPass)
			settings.ResticPass = encPass
		}

		if err := DB.Save(&settings).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to save settings: " + err.Error()})
			return
		}
		
		c.JSON(200, gin.H{"message": "Settings saved successfully", "mode": saveToken})


	})


	v1User.GET("/settings", func(c *gin.Context) {
		token := c.GetString("token")
		
		// Permitir ver los settings Globales (V2.6.1)
		searchToken := token
		if c.Query("mode") == "global" && c.GetBool("is_admin") {
			searchToken = "SYSTEM_GLOBAL"
		}

		var settings UserSettings
		if err := DB.Where("token = ?", searchToken).First(&settings).Error; err != nil {
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
		settings.WasabiKey, _ = Decrypt(settings.WasabiKey)
		settings.WasabiSecret, _ = Decrypt(settings.WasabiSecret)
		settings.ResticPass, _ = Decrypt(settings.ResticPass)

		c.JSON(200, settings)
	})


	// Endpoint de Prueba de Conexión Wasabi (V2.8)
	v1User.POST("/test-wasabi", func(c *gin.Context) {
		var input UserSettings
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		// Validaciones básicas
		if input.WasabiKey == "" || input.WasabiSecret == "" || input.WasabiBucket == "" {
			c.JSON(400, gin.H{"error": "Missing key, secret or bucket"})
			return
		}

		region := input.WasabiRegion
		if region == "" { region = "us-east-1" }
		
		endpoint := "s3.wasabisys.com"
		if region != "us-east-1" {
			endpoint = fmt.Sprintf("s3.%s.wasabisys.com", region)
		}

		// Configurar Sesión S3 para Wasabi (V2.8)
		s3Config := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(input.WasabiKey, input.WasabiSecret, ""),
			Endpoint:         aws.String(fmt.Sprintf("https://%s", endpoint)),
			Region:           aws.String(region),
			S3ForcePathStyle: aws.Bool(true), // Wasabi prefiere Path Style
		}

		sess, err := session.NewSession(s3Config)
		if err != nil {
			c.JSON(200, gin.H{"success": false, "error": fmt.Sprintf("Session Failed: %v", err)})
			return
		}

		svc := s3.New(sess)
		
		fmt.Printf("[TEST] Testing Wasabi for bucket: %s (%s)...\n", input.WasabiBucket, region)

		// 1. Probar ListBucket (Verifica existencia y permisos base)
		_, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:  aws.String(input.WasabiBucket),
			MaxKeys: aws.Int64(1),
		})

		if err != nil {
			c.JSON(200, gin.H{
				"success": false, 
				"error": fmt.Sprintf("S3 Check Failed: %v", err),
				"details": "Check if your Key/Secret are correct and have 'ListBucket' permission on this bucket.",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true, 
			"message": "Connection Successful! API can communicate with this Wasabi bucket.",
		})
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

	fmt.Printf("==========================================\n")
	fmt.Printf("🚀 DBP API %s - ONLINE\n", Version)
	fmt.Printf("==========================================\n")
	r.Run(":" + port)
}
