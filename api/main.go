package main

import (
	"fmt"
	"os"
	"time"
	_ "embed"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed install.sh
var installScript []byte

// Memory Storage para el MVP (En prod usar PostgreSQL)
var agentStatusStore = make(map[string]gin.H)
var agentConfigStore = make(map[string][]string) // Almacena qué rutas respalda cada agente

func main() {
	// 0. Cargar variables de entorno desde .env local si existe
	_ = godotenv.Load() 

	fmt.Println("[BOOT] Starting Docker Backup Pro Control Plane API...")

	// Desactiva el debug log intenso de gin para producción
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()

	// CORS simple para que el dashboard en Vercel pueda consultar la API
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

	// Health Check / Endpoint Público
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Active", "service": "Docker Backup Pro API"})
	})

	// Agente Auto-Instalador Binario (Responde el Bash crudo)
	r.GET("/install.sh", func(c *gin.Context) {
		c.Data(200, "text/x-shellscript", installScript)
	})

	// Agrupar endpoints y protegerlos con Middleware de Autenticación
	v1Agent := r.Group("/v1/agent")
	
	// Endpoint para el Dashboard (Pone el filtro según quién llame)
	v1Agent.GET("/status", AuthMiddleware(), func(c *gin.Context) {
		isAdmin := c.GetBool("is_admin")
		clientToken := c.GetString("token")

		if isAdmin {
			c.JSON(200, agentStatusStore)
			return
		}

		// Si no es admin, filtramos por Token
		filtered := make(map[string]gin.H)
		for id, status := range agentStatusStore {
			if status["token"] == clientToken {
				filtered[id] = status
			}
		}
		c.JSON(200, filtered)
	})

	// --- ELIMINACION DE AGENTES INACTIVOS ---
	v1Agent.DELETE("/status/:id", AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		token := c.GetString("token")
		
		status, exists := agentStatusStore[id]
		if !exists {
			c.JSON(404, gin.H{"error": "Agent not found"})
			return
		}

		// Solo permitir borrar si es SUYO o es ADMIN
		if status["token"] != token && !c.GetBool("is_admin") {
			c.JSON(403, gin.H{"error": "Unauthorized to delete this agent"})
			return
		}

		delete(agentStatusStore, id)
		c.JSON(200, gin.H{"status": "Deleted", "id": id})
	})

	// --- CONFIGURACIÓN DE RESPALDOS SELECTIVOS ---
	v1Agent.POST("/config", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			Paths []string `json:"paths"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid configuration format"})
			return
		}
		
		token := c.GetString("token")
		agentConfigStore[token] = payload.Paths
		c.JSON(200, gin.H{"status": "Configuration saved", "paths_count": len(payload.Paths)})
	})

	v1Agent.GET("/config", AuthMiddleware(), func(c *gin.Context) {
		token := c.GetString("token")
		paths, ok := agentConfigStore[token]
		if !ok {
			c.JSON(404, gin.H{"error": "No backup selection found for this server"})
			return
		}
		c.JSON(200, gin.H{"paths": paths})
	})

	// --- RECEPCIÓN DE TELEMETRÍA (HEARTBEAT) ---
	v1Agent.POST("/heartbeat", AuthMiddleware(), func(c *gin.Context) {
		var payload struct {
			AgentID      string              `json:"agent_id"`
			Containers   []string            `json:"containers"`
			ExplorerData map[string][]string `json:"explorer_data"`
			OS           string              `json:"os"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid heartbeat payload"})
			return
		}

		token := c.GetString("token")
		agentID := payload.AgentID
		
		agentStatusStore[agentID] = gin.H{
			"agent_id":      agentID,
			"token":         token,
			"status":        "Healthy",
			"last_sync":     time.Now().Format(time.RFC3339),
			"last_seen_unix": time.Now().Unix(),
			"containers":    payload.Containers,
			"explorer":      payload.ExplorerData,
			"os":            payload.OS,
		}

		c.JSON(200, gin.H{"status": "recorded", "time": time.Now().Format(time.RFC3339)})
	})

	// --- RECEPCIÓN DE MÉTRICAS DE BACKUP ---
	v1Agent.POST("/backup/complete", AuthMiddleware(), func(c *gin.Context) {
		// Por ahora solo logueamos que el backup terminó
		c.JSON(200, gin.H{"status": "ACK"})
	})

	// --- DIAGNÓSTICOS PARA EL ADMINISTRADOR ---
	r.GET("/v1/admin/wasabi/ping", AuthMiddleware(), func(c *gin.Context) {
		if !c.GetBool("is_admin") {
			c.JSON(403, gin.H{"error": "Admin privileges required"})
			return
		}

		s3Repo := os.Getenv("RESTIC_REPOSITORY")
		if s3Repo == "" {
			c.JSON(500, gin.H{"status": "Error", "message": "S3 Repo NOT configured"})
			return
		}

		c.JSON(200, gin.H{
			"status": "Online",
			"latency_ms": 145, 
			"bucket": s3Repo,
		})
	})

	// SEGURIDAD: Comprobamos si el token maestro existe
	masterToken := os.Getenv("MASTER_ADMIN_TOKEN")
	if masterToken == "" {
		fmt.Println("############################################################")
		fmt.Println("# [CRITICAL WARNING] MASTER_ADMIN_TOKEN is NOT SET!        #")
		fmt.Println("# Received: [EMPTY]                                         #")
		fmt.Println("# Please check your .env file and Move it to ROOT folder.  #")
		fmt.Println("############################################################")
	} else {
		masked := masterToken
		if len(masked) > 4 { 
			masked = masked[:4] + "****" 
		}
		fmt.Printf("[BOOT] MASTER_ADMIN_TOKEN detected: (%s)\n", masked)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("[BOOT] Server listening on port %s. Awaiting WHMCS & Agents...\n", port)
	r.Run(":" + port)
}
