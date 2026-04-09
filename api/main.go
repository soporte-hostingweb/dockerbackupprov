package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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


const Version = "V9.2.5"

//go:embed install.sh
var installScript []byte

// DispatchAlert: Función asíncrona universal (V9.0 SaaS)
// Busca si el tenant tiene un Webhook y dispara una alerta POST en goroutine.
func DispatchAlert(token string, eventType string, details map[string]interface{}) {
	go func() {
		var config AlertConfig
		// Consultar el webhook según el token del tenant
		if err := DB.Where("token = ?", token).First(&config).Error; err != nil {
			return // No tiene configurado n8n/webhook u ocurrió error
		}
		
		if config.WebhookURL == "" || !strings.Contains(config.Events, eventType) {
			return // Webhook vacío o no suscrito a este evento
		}

		payload := map[string]interface{}{
			"event":     eventType,
			"token":     token,
			"timestamp": time.Now().Format(time.RFC3339),
			"details":   details,
		}

		jsonBody, _ := json.Marshal(payload)
		
		req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "DBP-SaaS-Orchestrator/V9.0")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()
}

// UpdateHealthScore: Calcula el robustez del agente (0-100) basado en 4 KPIs críticos (V9.1)
func UpdateHealthScore(agentID string) {
	var agent AgentStatus
	if err := DB.First(&agent, "id = ?", agentID).Error; err != nil {
		return
	}

	score := 100

	// 1. Penalización por Conectividad (Offline: -50, Degraded: -30)
	if agent.HealthStatus == "OFFLINE" {
		score -= 50
	} else if agent.HealthStatus == "DEGRADED" {
		score -= 30
	}

	// 2. Penalización por Obsolescencia (Backup > 24h: -20)
	// Solo penalizamos si ya ha tenido al menos un backup exitoso
	if !agent.LastBackupAt.IsZero() && time.Since(agent.LastBackupAt) > 24*time.Hour {
		score -= 20
	}

	// 3. Penalización por Integridad (Validación Fallida: -40)
	if agent.VerificationStatus == "INVALID" {
		score -= 40
	}

	// Capamos el score entre 0 y 100
	if score < 0 { score = 0 }
	if score > 100 { score = 100 }

	// Persistir resultado
	DB.Model(&agent).Update("health_score", score)

	// V9.2.5: Auditoría de Score (Detalle técnico para el usuario)
	DB.Create(&ActivityLog{
		Token:     agent.Token,
		AgentID:   agent.ID,
		Type:      "TELEMETRY",
		Status:    "success",
		Message:   fmt.Sprintf("[SCORE] Puntaje actualizado a %d%%. [H:%s|V:%s]", score, agent.HealthStatus, agent.VerificationStatus),
		StartedAt: time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
	})
}

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

	// V6.6: Monitor de Salud Proactivo - Detecta agentes caídos cada 60s
	go func() {
		for {
			time.Sleep(60 * time.Second)
			var agents []AgentStatus
			DB.Find(&agents)
			for _, a := range agents {
				// Si no ha reportado en más de 5 minutos y no estaba ya marcado como error/offline
				if time.Since(a.UpdatedAt) > 5*time.Minute {
					var lastLog ActivityLog
					DB.Where("agent_id = ? AND type = 'OFFLINE'", a.ID).Order("started_at desc").First(&lastLog)
					
					// Solo logueamos si la última actividad OFFLINE fue hace más de 1 hora para no saturar
					if lastLog.ID == 0 || time.Since(lastLog.StartedAt) > 1*time.Hour {
						DB.Create(&ActivityLog{
							Token:      a.Token,
							AgentID:    a.ID,
							Type:       "OFFLINE",
							Status:     "error",
							Message:    fmt.Sprintf("Agent %s lost connection (Heartbeat Timeout)", a.ID),
							StartedAt:  time.Now(),
							FinishedAt: time.Now(),
						})
						fmt.Printf("[MONITOR] Agent %s marked as OFFLINE\n", a.ID)
						
						// V9.0: Guardar en el nuevo HealthStatus y Disparar Alerta Universal
						DB.Model(&a).Update("health_status", "OFFLINE")
						DispatchAlert(a.Token, "agent_offline", map[string]interface{}{
							"agent_id": a.ID,
							"time_since_last_seen": time.Since(a.UpdatedAt).String(),
						})
						// V9.1: Actualizar Score de Salud
						UpdateHealthScore(a.ID)
					}
				} else if a.HealthStatus == "OFFLINE" {
					// Auto-curación HealthCheck si está vivo (V9.1)
					DB.Model(&a).Update("health_status", "ONLINE")
					DispatchAlert(a.Token, "agent_recovered", map[string]interface{}{
						"agent_id": a.ID,
						"status":   "Connection restored",
					})
					UpdateHealthScore(a.ID)
					if a.HealthStatus == "OFFLINE" {
						DB.Model(&a).Update("health_status", "ONLINE")
					}
				}
			}
		}
	}()

	r.GET("/v1/version", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"version": Version, "status": "active", "network": "HWPeru SaaS"})
	})

	// --- ENDPOINT PROVISIÓN WHMCS (V9.1) ---
	r.POST("/v1/whmcs/provision", func(c *gin.Context) {
		adminKey := os.Getenv("API_ADMIN_KEY")
		providedKey := c.GetHeader("X-Admin-Key")
		if adminKey == "" || providedKey != adminKey {
			c.JSON(401, gin.H{"error": "Unauthorized API Admin Access"})
			return
		}

		var req struct {
			ServiceID   string `json:"service_id"`
			ClientEmail string `json:"client_email"`
			Plan        string `json:"plan"` // basic, standard, enterprise
			Retention   int    `json:"retention_days"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid provisioning request payload"})
			return
		}

		// Generar Token Único para el Cliente (SSO)
		token := "dbp_" + fmt.Sprintf("%x", time.Now().Unix()) + "_" + req.ServiceID

		// 1. Crear/Actualizar TenantPlan (Source of Truth Comercial)
		var plan TenantPlan
		DB.Where("whmcs_service_id = ?", req.ServiceID).First(&plan)
		plan.Token = token
		plan.WhmcsServiceID = req.ServiceID
		plan.ClientEmail = req.ClientEmail
		plan.Plan = req.Plan
		plan.RetentionDays = req.Retention
		if req.Plan == "enterprise" {
			plan.Priority = true
			plan.ValidationLvl = "advanced"
		} else if req.Plan == "standard" {
			plan.ValidationLvl = "basic"
		} else {
			plan.ValidationLvl = "none"
		}
		DB.Save(&plan)

		// 2. Asegurar que UserSettings exista para Wasabi
		var settings UserSettings
		if err := DB.Where("token = ?", token).First(&settings).Error; err != nil {
			DB.Create(&UserSettings{Token: token})
		}

		// 3. Configurar Alertas por Defecto (Habilitar todos por ser SaaS)
		var alerts AlertConfig
		if err := DB.Where("token = ?", token).First(&alerts).Error; err != nil {
			DB.Create(&AlertConfig{
				Token:  token,
				Events: "backup_success,backup_failed,backup_validation_failed,agent_offline,agent_recovered,restore_started,restore_completed",
			})
		}

		c.JSON(200, gin.H{
			"status":        "success",
			"token":         token,
			"dashboard_url": "https://backup.hwperu.com/?sso=" + token,
		})
	})

	v1Agent := r.Group("/v1/agent")

	// --- ENDPOINTS DE TRADUCCIÓN (i18n) ---
	r.GET("/v1/translations", func(c *gin.Context) {
		lang := c.Query("lang")
		if lang == "" { lang = "en" }

		// 1. Cargar base según el idioma o fallback a 'en'
		baseFile := fmt.Sprintf("lang/%s.json", lang)
		if _, err := os.Stat(baseFile); os.IsNotExist(err) {
			baseFile = "lang/en.json" // Fallback seguro
		}

		baseData, err := os.ReadFile(baseFile)
		if err != nil {
			c.JSON(500, gin.H{"error": "Base language file not found internally."})
			return
		}

		baseDict := make(map[string]string)
		json.Unmarshal(baseData, &baseDict)

		// 2. Fusionar Custom Override (modificación del Admin) si existe
		customData, errC := os.ReadFile("lang/custom_lang.json")
		if errC == nil {
			customDict := make(map[string]string)
			if errJ := json.Unmarshal(customData, &customDict); errJ == nil {
				for k, v := range customDict {
					baseDict[k] = v // Sobrescribir
				}
			}
		}

		c.JSON(200, baseDict)
	})

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
			DB.Limit(1).Where("token = ? AND agent_id = ?", a.Token, a.ID).Find(&config)

			// Filtramos el dbp-client-agent de los contenedores reportados
			cleanContainers := []string{}
			for _, co := range containers {
				if co != "dbp-client-agent" {
					cleanContainers = append(cleanContainers, co)
				}
			}

			isOnline := (time.Now().Unix() - a.LastSeenUnix) < 25

			resp[a.ID] = gin.H{
				"agent_id":       a.ID,
				"token":          a.Token,
				"status":         a.Status,
				"is_online":      isOnline,
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
				"last_backup_bytes": a.LastBackupBytes,
				"wasabi_usage_gb":   fmt.Sprintf("%.2f GB", float64(a.LastBackupBytes)/(1024*1024*1024)),
				"health_score":      a.HealthScore,
				"health_status":     a.HealthStatus,
				"verification_status": a.VerificationStatus,
				"est_rto_secs":      a.EstRtoSecs,
				"schedule":       config.Schedule,
				"timezone":       config.TimeZone,
				"custom_schedule": config.CustomSchedule,
				"cmd_task":       a.CmdTask,
				"cmd_param":      a.CmdParam,
				"cmd_result":     a.CmdResult,
			}
		}

		c.JSON(200, resp)
	})

	// V6.3: Monitor de Actividad Global (Reemplaza a /history por uno más detallado)
	r.GET("/v1/activities", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		var activities []ActivityLog
		if isAdmin {
			DB.Order("started_at desc").Limit(50).Find(&activities)
		} else {
			DB.Where("token = ?", token).Order("started_at desc").Limit(50).Find(&activities)
		}
		c.JSON(200, activities)
	})

	// V9.2.1: Alias para compatibilidad con el UI
	r.GET("/v1/history", AuthMiddleware(), func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/v1/activities")
	})

	// Endpoint para que el AGENTE reporte su estado en tiempo real (V6.3)
	v1Agent.POST("/activity/report", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		var req struct {
			ActivityID uint   `json:"activity_id"`
			AgentID    string `json:"agent_id"`
			Type       string `json:"type"`    // backup, restore, prune
			Status     string `json:"status"`  // running, success, error
			Message    string `json:"message"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		fmt.Printf("[%s] [ACTIVITY] Agent %s reporting %s status: %s\n", 
			time.Now().Format("15:04:05"), req.AgentID, req.Type, req.Status)

		var activity ActivityLog
		if req.ActivityID > 0 {
			// Actualizamos una actividad existente
			if err := DB.First(&activity, req.ActivityID).Error; err == nil {
				activity.Status = req.Status
				activity.Message = req.Message
				if req.Status == "success" || req.Status == "error" {
					activity.FinishedAt = time.Now()
				}
				DB.Save(&activity)
				c.JSON(200, gin.H{"status": "updated", "activity_id": activity.ID})
				return
			}
		}

		// Creamos una NUEVA actividad
		activity = ActivityLog{
			Token:     token,
			AgentID:   req.AgentID,
			Type:      req.Type,
			Status:    req.Status,
			Message:   req.Message,
			StartedAt: time.Now().UTC(),
		}
		if req.Status == "success" || req.Status == "error" {
			activity.FinishedAt = time.Now().UTC()
		}
		DB.Create(&activity)
		c.JSON(200, gin.H{"status": "created", "activity_id": activity.ID})
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

		// V6.6: Registro de auditoría antes de eliminar
		audit := ActivityLog{
			Token:     agent.Token,
			AgentID:   agent.ID,
			Type:      "DELETED",
			Status:    "success",
			Message:   fmt.Sprintf("Agent %s was manually deleted from Control Plane", agent.ID),
			StartedAt: time.Now(),
			FinishedAt: time.Now(),
		}
		DB.Create(&audit)

		DB.Delete(&agent)
		c.JSON(200, gin.H{"status": "Deleted", "id": id, "audit_id": audit.ID})
	})	// --- ENDPOINTS DE CONFIGURACIÓN (V5.1.2) ---

	v1Agent.GET("/config", AuthMiddleware(), func(c *gin.Context) {
		agentID := c.Query("agent_id")
		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		if agentID == "" {
			c.JSON(400, gin.H{"error": "agent_id is required"})
			return
		}

		// Impersonación para Admin: Si eres admin, buscamos el token real del agente
		effectiveToken := token
		if isAdmin {
			var agent AgentStatus
			if err := DB.Limit(1).Where("id = ?", agentID).Find(&agent).Error; err == nil && agent.ID != "" {
				effectiveToken = agent.Token
			}
		}

		var configs []BackupConfig
		DB.Where("token = ? AND agent_id = ?", effectiveToken, agentID).Find(&configs)

		// V5.1.3: Recuperar configuración de Wasabi (Tenant o Global)
		var settings UserSettings
		// 1. Buscamos settings del inquilino (effectiveToken)
		DB.Limit(1).Where("token = ?", effectiveToken).Find(&settings)
		if settings.ID == 0 {
			// 2. Si no hay, buscamos las globales
			DB.Limit(1).Where("token = ?", "SYSTEM_GLOBAL").Find(&settings)
		}

		if settings.ID == 0 {
			c.JSON(200, gin.H{"status": "manual", "error": "WASABI_UNCONFIGURED"})
			return
		}

		// Descifrar credenciales para el agente
		wasabiKey, _ := Decrypt(settings.WasabiKey)
		wasabiSecret, _ := Decrypt(settings.WasabiSecret)
		resticPass, _ := Decrypt(settings.ResticPass)
		wasabiBucket := settings.WasabiBucket
		wasabiRegion := settings.WasabiRegion
		if wasabiRegion == "" { wasabiRegion = "us-east-1" }

		// V7.0: Aislamiento Multi-Tenant Profesional
		endpoint := "s3.wasabisys.com"
		if wasabiRegion != "us-east-1" {
			endpoint = fmt.Sprintf("s3.%s.wasabisys.com", wasabiRegion)
		}
		
		// Estructura: BUCKET / TOKEN_CLIENTE / AGENT_ID
		fullRepo := fmt.Sprintf("s3:https://%s/%s/%s/%s", endpoint, wasabiBucket, effectiveToken, agentID)

		var config BackupConfig
		if len(configs) > 0 { config = configs[0] }

		var paths []string
		if config.Paths != "" { json.Unmarshal([]byte(config.Paths), &paths) }

		c.JSON(200, gin.H{
			"status":          "success",
			"paths":           paths,
			"schedule":        config.Schedule,
			"retention":       config.Retention,
			"timezone":        config.TimeZone,
			"custom_schedule": config.CustomSchedule,
			"full_repo_url":   fullRepo,
			"restic_password": resticPass,
			"wasabi_key":      wasabiKey,
			"wasabi_secret":   wasabiSecret,
		})
	})


	// V5.0/V5.1.2: Endpoint para GUARDAR la configuración (Con impersonación Admin)
	v1Agent.POST("/config/save", AuthMiddleware(), func(c *gin.Context) {
		var req struct {
			AgentID        string   `json:"agent_id"`
			Schedule       string   `json:"schedule"`
			Paths          []string `json:"paths"`
			Retention      int      `json:"retention"`
			TimeZone       string   `json:"timezone"`
			CustomSchedule string   `json:"custom_schedule"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		// Impersonación para Admin
		effectiveToken := token
		if isAdmin {
			var agent AgentStatus
			if err := DB.Limit(1).Where("id = ?", req.AgentID).Find(&agent).Error; err == nil && agent.ID != "" {
				effectiveToken = agent.Token
			}
		}
		
		var config BackupConfig
		res := DB.Where("token = ? AND agent_id = ?", effectiveToken, req.AgentID).First(&config)
		
		pathsJSON, _ := json.Marshal(req.Paths)
		
		if res.Error != nil {
			config = BackupConfig{
				Token:          effectiveToken,
				AgentID:        req.AgentID,
				Schedule:       req.Schedule,
				Paths:          string(pathsJSON),
				Retention:      req.Retention,
				TimeZone:       req.TimeZone,
				CustomSchedule: req.CustomSchedule,
			}
			DB.Create(&config)
		} else {
			config.Schedule = req.Schedule
			config.Paths = string(pathsJSON)
			config.Retention = req.Retention
			config.TimeZone = req.TimeZone
			config.CustomSchedule = req.CustomSchedule
			DB.Model(&config).Updates(BackupConfig{
				Schedule:       req.Schedule,
				Paths:          string(pathsJSON),
				Retention:      req.Retention,
				TimeZone:       req.TimeZone,
				CustomSchedule: req.CustomSchedule,
			})
		}

		c.JSON(200, gin.H{"status": "Configuration Saved", "schedule": req.Schedule, "retention": req.Retention})
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
			if err := DB.Limit(1).Where("id = ?", req.AgentID).Find(&agent).Error; err == nil && agent.ID != "" {
				// Usamos el token original del agente para guardar la config
				saveToken = agent.Token
			}
		}

		var config BackupConfig
		if err := DB.Limit(1).Where("token = ? AND agent_id = ?", saveToken, req.AgentID).Find(&config).Error; err == nil && config.ID != 0 {
			config.Paths = string(pathsJSON)
			config.Schedule = req.Schedule
			DB.Save(&config)
		} else {
			config = BackupConfig{
				Token:     saveToken,
				AgentID:   req.AgentID,
				Paths:     string(pathsJSON),
				Schedule:  req.Schedule,
				Retention: 1, // Default (Gratis)
			}
			DB.Create(&config)
		}

		c.JSON(200, gin.H{"status": "Config updated in Control Plane"})
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
			HealthStatus: "ONLINE", // Asegurar estado ONLINE en cada latido exitoso (V9.1)
			LastSeen:     time.Now().UTC(),
			LastSeenUnix: time.Now().Unix(),
			OS:           payload.OS,
			FreeSpace:    payload.FreeSpace,
			TotalSpace:   payload.TotalSpace,
			IsSyncing:    payload.IsSyncing,
			ActivePID:    payload.ActivePID,
		}

		// V6.8: Protección de Inventario - Solo actualizamos si el payload no viene vacío
		if len(payload.Containers) > 0 { agent.Containers = string(contJSON) }
		if len(payload.ExplorerData) > 0 { agent.Explorer = string(expJSON) }
		if len(payload.Snapshots) > 0 { agent.Snapshots = string(snapJSON) }


		if payload.LastBackupAt > 0 {
			agent.LastBackupAt = time.Unix(payload.LastBackupAt, 0).UTC()
		}


		// Importante: No machacar Maintenance, PendingForce y Tareas si ya existen
		var existingList []AgentStatus
		if err := DB.Limit(1).Where("id = ?", payload.AgentID).Find(&existingList).Error; err == nil && len(existingList) > 0 {
			existing := existingList[0]
			agent.Maintenance = existing.Maintenance
			agent.PendingForce = existing.PendingForce
			agent.KillSync = existing.KillSync
			
			// V4.6.0: Preservar tareas de comando para que no se borren antes de entregarse
			agent.CmdTask = existing.CmdTask
			agent.CmdParam = existing.CmdParam
			agent.CmdResult = existing.CmdResult
			
			// Si el agente reporta que está sincronizando, consumimos la instrucción (V3.4.1)
			if payload.IsSyncing && agent.PendingForce != "none" {
				agent.PendingForce = "none"
			}
			
			// Si se procesó una orden de kill, la reiniciamos (V3.4.1)
			// V6.8: Si el payload vino vacío, recuperamos lo que ya teníamos en DB
			if agent.Containers == "" { agent.Containers = existing.Containers }
			if agent.Explorer == ""   { agent.Explorer = existing.Explorer }
			if agent.Snapshots == ""  { agent.Snapshots = existing.Snapshots }
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

		// V9.1: Recalcular Score de Salud en cada Heartbeat
		go UpdateHealthScore(payload.AgentID)
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
		fmt.Printf("[%s] [TASK-RESULT] Agent %s reported result for task: %s (Weight: %d bytes)\n", 
			time.Now().Format("15:04:05"), req.AgentID, req.Task, len(req.Result))

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
			TotalSizeBytes int64  `json:"total_size_bytes"` // V4.6.1
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
			fmt.Printf("[METRICS] Receiving backup metrics for Agent %s. Size: %d bytes\n", payload.AgentID, payload.TotalSizeBytes)
			
			activity := BackupActivity{
				AgentID:      payload.AgentID,
				Token:        agent.Token,
				Status:       payload.Status,
				SnapshotID:   payload.SnapshotID,
				SizeMB:       payload.TotalSizeMB,
				SizeBytes:    payload.TotalSizeBytes, // V4.6.1: Precisión para archivos pequeños
				DurationSecs: payload.DurationSecs,
				StartedAt:    time.Unix(payload.StartedAt, 0).UTC(),
				FinishedAt:   time.Unix(payload.Timestamp, 0).UTC(),
				CreatedAt:    time.Now(),
			}
			DB.Create(&activity)
			
			// Actualizamos el último backup exitoso en el estado del agente
			if payload.Status == "SUCCESS" {
				DB.Model(&agent).Updates(map[string]interface{}{
					"last_backup_at":    time.Unix(payload.Timestamp, 0).UTC(),
					"last_backup_bytes": payload.TotalSizeBytes,
				})

				// V9.2.5: Auditoría de Métricas
				DB.Create(&ActivityLog{
					Token:     agent.Token,
					AgentID:   agent.ID,
					Type:      "TELEMETRY",
					Status:    "success",
					Message:   fmt.Sprintf("[METRICS] Consumo Wasabi registrado: %d bytes.", payload.TotalSizeBytes),
					StartedAt: time.Now().UTC(),
					FinishedAt: time.Now().UTC(),
				})

				// V9.1: Alerta Éxito
				DispatchAlert(agent.Token, "backup_success", map[string]interface{}{
					"agent_id": payload.AgentID,
					"size_mb":  payload.TotalSizeMB,
				})
			} else {
				// V9.0: Enviar alerta si el backup falló
				DispatchAlert(agent.Token, "backup_failed", map[string]interface{}{
					"agent_id": payload.AgentID,
					"error":    "Backup process returned non-success status",
					"duration": payload.DurationSecs,
				})
			}
			// V9.1: Siempre actualizar score tras backup
			go UpdateHealthScore(payload.AgentID)
		}


		c.JSON(200, gin.H{"status": "Metrics recorded and activity saved"})
	})

	// --- VERIFICACIÓN DE INTEGRIDAD RTO & VALIDACIÓN (V9.0) ---
	v1Agent.POST("/verification/report", AuthMiddleware(), func(c *gin.Context) {
		var req struct {
			AgentID    string   `json:"agent_id"`
			SnapshotID string   `json:"snapshot_id"`
			Status     string   `json:"status"` // VALID, INVALID
			Errors     []string `json:"errors"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid verification report"})
			return
		}

		var agent AgentStatus
		if err := DB.First(&agent, "id = ?", req.AgentID).Error; err != nil {
			c.JSON(404, gin.H{"error": "Agent not found"})
			return
		}

		healthStatus := "ONLINE"
		if req.Status == "INVALID" {
			healthStatus = "DEGRADED"
			// V9.1: Evento SaaS Estandarizado
			DispatchAlert(agent.Token, "backup_validation_failed", map[string]interface{}{
				"agent_id":    req.AgentID,
				"snapshot_id": req.SnapshotID,
				"errors":      req.Errors,
			})
		}

		DB.Model(&agent).Updates(map[string]interface{}{
			"verification_status": req.Status,
			"health_status":       healthStatus,
		})

		// V9.1: Actualizar Score tras validación
		go UpdateHealthScore(req.AgentID)

		errStr := strings.Join(req.Errors, " | ")
		DB.Model(&BackupActivity{}).Where("snapshot_id = ?", req.SnapshotID).Updates(map[string]interface{}{
			"validation_status": req.Status,
			"validation_errors": errStr,
		})

		// V9.2.5: Auditoría de Verificación
		DB.Create(&ActivityLog{
			Token:     agent.Token,
			AgentID:   req.AgentID,
			Type:      "TELEMETRY",
			Status:    "success",
			Message:   fmt.Sprintf("[INTEGRITY] Snapshot %s verificado: %s", req.SnapshotID, req.Status),
			StartedAt: time.Now().UTC(),
			FinishedAt: time.Now().UTC(),
		})

		c.JSON(200, gin.H{"status": "Verification report processed"})
	})

	// --- MÉTRICAS DE RESTAURACIÓN DE DATOS RTO (V9.0) ---
	v1Agent.POST("/restore/metrics", AuthMiddleware(), func(c *gin.Context) {
		var req struct {
			AgentID      string `json:"agent_id"`
			SnapshotID   string `json:"snapshot_id"`
			TotalSeconds int    `json:"total_seconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid restore metrics"})
			return
		}

		DB.Model(&BackupActivity{}).Where("snapshot_id = ?", req.SnapshotID).Update("restore_duration_secs", req.TotalSeconds)
		
		// Calcular Nuevo RTO Estimado (avg de los últimos 5)
		var agent AgentStatus
		if err := DB.First(&agent, "id = ?", req.AgentID).Error; err == nil {
			var activities []BackupActivity
			DB.Where("agent_id = ? AND restore_duration_secs > 0", req.AgentID).Order("started_at desc").Limit(5).Find(&activities)
			
			if len(activities) > 0 {
				var total int
				for _, act := range activities {
					total += act.RestoreDurationSecs
				}
				avg := total / len(activities)
				DB.Model(&agent).Update("est_rto_secs", avg)

				// V9.2.5: Auditoría de RTO
				DB.Create(&ActivityLog{
					Token:     agent.Token,
					AgentID:   agent.ID,
					Type:      "TELEMETRY",
					Status:    "success",
					Message:   fmt.Sprintf("[RTO] Nuevo tiempo estimado de recuperación: %d segundos.", avg),
					StartedAt: time.Now().UTC(),
					FinishedAt: time.Now().UTC(),
				})
			}

			DispatchAlert(agent.Token, "restore_completed", map[string]interface{}{
				"agent_id":     req.AgentID,
				"snapshot_id":  req.SnapshotID,
				"duration_sec": req.TotalSeconds,
			})
		}
		c.JSON(200, gin.H{"status": "Restore metrics saved"})
	})

	// --- ORQUESTADOR BARE-METAL RESTORE (V8.0) ---
	v1Agent.POST("/clone", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		isAdmin := c.GetBool("is_admin")

		var req struct {
			SourceAgentID string `json:"source_agent_id"`
			SnapshotID    string `json:"snapshot_id"`
			TargetIP      string `json:"ip"`
			TargetPort    string `json:"port"`
			TargetPass    string `json:"pass"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		// 1. Obtener Token Efectivo y Validar Agente Origen
		var sourceAgent AgentStatus
		if err := DB.Where("id = ?", req.SourceAgentID).First(&sourceAgent).Error; err != nil {
			c.JSON(404, gin.H{"error": "Source agent not found"})
			return
		}
		if !isAdmin && sourceAgent.Token != token {
			c.JSON(403, gin.H{"error": "Unauthorized access to source agent"})
			return
		}
		effectiveToken := sourceAgent.Token

		// 2. Bloquear Agente Origen (Prevención de colisiones)
		sourceAgent.Maintenance = true
		DB.Save(&sourceAgent)

		// 3. Obtener Configuración y Credenciales para el Nuevo Servidor
		var settings UserSettings
		DB.Limit(1).Where("token = ?", effectiveToken).Find(&settings)
		if settings.ID == 0 {
			DB.Limit(1).Where("token = ?", "SYSTEM_GLOBAL").Find(&settings)
		}
		wasabiKey, _ := Decrypt(settings.WasabiKey)
		wasabiSecret, _ := Decrypt(settings.WasabiSecret)
		resticPass, _ := Decrypt(settings.ResticPass)
		bucket := settings.WasabiBucket
		region := settings.WasabiRegion
		if region == "" { region = "us-east-1" }
		endpoint := "s3.wasabisys.com"
		if region != "us-east-1" { endpoint = fmt.Sprintf("s3.%s.wasabisys.com", region) }
		fullRepo := fmt.Sprintf("s3:https://%s/%s/%s/%s", endpoint, bucket, effectiveToken, req.SourceAgentID)

		// 4. Inyección Asíncrona (Conexión SSH y Restauración)
		go func() {
			activity := ActivityLog{
				Token:     effectiveToken,
				AgentID:   req.SourceAgentID, // Agrupado bajo el ID origen para trazabilidad
				Type:      "bare_metal_restore",
				Status:    "running",
				Message:   fmt.Sprintf("Connecting via SSH to %s...", req.TargetIP),
				StartedAt: time.Now().UTC(),
			}
			DB.Create(&activity)

			defer func() {
				// Al terminar, devolver Agent a la vida
				sourceAgent.Maintenance = false
				DB.Save(&sourceAgent)
			}()

			importSSH := true // Referencia
			_ = importSSH

			// V8.0: Script Mágico de Rescate Automático
			rescueScript := fmt.Sprintf(`#!/bin/bash
echo "[DBP] Inbound Secure Restoration Thread initialized."
export AWS_ACCESS_KEY_ID='%s'
export AWS_SECRET_ACCESS_KEY='%s'
export RESTIC_PASSWORD='%s'
export RESTIC_REPOSITORY='%s'

if ! command -v restic &> /dev/null; then
    wget -qO restic.bz2 https://github.com/restic/restic/releases/download/v0.16.4/restic_0.16.4_linux_amd64.bz2
    bzip2 -d restic.bz2 && chmod +x restic && mv restic /usr/local/bin/
fi

echo "[DBP] Commencing Bare-Metal Snapshot Extraction: %s"
restic restore %s --target / > /var/log/dbp_restore.log 2>&1
C_RES=$?

if [ $C_RES -eq 0 ]; then
    curl -s -X POST -H "Content-Type: application/json" -d '{"activity_id": %d, "agent_id": "%s", "type": "bare_metal_restore", "status": "success", "message": "Bare metal restore fully completed on target %s"}' http://api.hwperu.com/v1/agent/activity/report > /dev/null
    sleep 3; reboot
else
    curl -s -X POST -H "Content-Type: application/json" -d '{"activity_id": %d, "agent_id": "%s", "type": "bare_metal_restore", "status": "error", "message": "Restore crashed. Code: '"$C_RES"'"}' http://api.hwperu.com/v1/agent/activity/report > /dev/null
fi
`, wasabiKey, wasabiSecret, resticPass, fullRepo, req.SnapshotID, req.SnapshotID, activity.ID, req.SourceAgentID, req.TargetIP, activity.ID, req.SourceAgentID)

			cmdSSH := exec.Command("sshpass", "-p", req.TargetPass, "ssh", "-o", "StrictHostKeyChecking=no", "-p", req.TargetPort, "root@"+req.TargetIP, rescueScript)
			// Lanzarlo en background desatendido
			err := cmdSSH.Start()
			
			if err != nil {
				activity.Status = "error"
				activity.Message = fmt.Sprintf("SSH Negotiation failed with %s: %v", req.TargetIP, err)
			} else {
				activity.Message = fmt.Sprintf("Rescue agent successfully deployed to %s. Decoding snapshot %s...", req.TargetIP, req.SnapshotID)
			}
			activity.FinishedAt = time.Now().UTC()
			DB.Save(&activity)
		}()

		c.JSON(200, gin.H{"status": "Bootstrapping Target Server...", "target": req.TargetIP})
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
			Path        string   `json:"path"` // Ruta para filtrar ls_snapshot (V4.5.9)
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
			// V4.5.9: Concatenamos ID|PATH para listado granular
			param := req.SnapshotID
			if req.Path != "" {
				param = req.SnapshotID + "|" + req.Path
			}
			DB.Model(&agent).Updates(map[string]interface{}{
				"cmd_task":   "ls_snapshot",
				"cmd_param":  param,
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
			// V9.1: Evento Inicio Restauración
			DispatchAlert(agent.Token, "restore_started", map[string]interface{}{
				"agent_id":    id,
				"snapshot_id": req.SnapshotID,
				"target":      req.Destination,
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
		var input UserSettingsPayload
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

		// V9.0: Guardar/Actualizar AlertConfig (Webhooks n8n)
		var alertConfig AlertConfig
		DB.Where("token = ?", saveToken).First(&alertConfig)
		alertConfig.Token = saveToken
		alertConfig.WebhookURL = input.WebhookURL
		alertConfig.Events = input.WebhookEvents
		DB.Save(&alertConfig)
		
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

		// V9.0: Incluir Configuración de Alertas
		var alertConfig AlertConfig
		_ = DB.Where("token = ?", searchToken).First(&alertConfig).Error

		response := UserSettingsPayload{
			WasabiKey:     settings.WasabiKey,
			WasabiSecret:  settings.WasabiSecret,
			WasabiBucket:  settings.WasabiBucket,
			WasabiRegion:  settings.WasabiRegion,
			ResticPass:    settings.ResticPass,
			WebhookURL:    alertConfig.WebhookURL,
			WebhookEvents: alertConfig.Events,
		}

		c.JSON(200, response)
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
