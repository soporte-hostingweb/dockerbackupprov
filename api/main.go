package main

import (
	"fmt"
	"os"
	_ "embed"
	"github.com/gin-gonic/gin"
)



//go:embed install.sh
var installScript []byte

// Memory Storage para el MVP (En prod usar PostgreSQL)
var agentStatusStore = make(map[string]gin.H)

func main() {
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

	// --- CONFIGURACIÓN DE RESPALDOS (SELECTIVE BACKUP) ---
	var agentConfigStore = make(map[string]gin.H)

	v1Agent.POST("/config", func(c *gin.Context) {
		var payload gin.H
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid config"})
			return
		}
		agentID := c.GetString("token") // Usamos el token como ID de config para aislamiento
		agentConfigStore[agentID] = payload
		c.JSON(200, gin.H{"status": "Configuration saved", "agent": agentID})
	})

	v1Agent.GET("/config", func(c *gin.Context) {
		agentID := c.GetString("token")
		config, ok := agentConfigStore[agentID]
		if !ok {
			c.JSON(404, gin.H{"error": "No config found for this agent"})
			return
		}
		c.JSON(200, config)
	})

	v1Agent.Use(AuthMiddleware(), func(c *gin.Context) {
		// Middleware adicional si se necesita
		c.Next()
	}) 

	{
		v1Agent.POST("/heartbeat", ReceiveHeartbeat)
		v1Agent.POST("/backup/complete", ReceiveBackupCompletion)
	}

	// SEGURIDAD: Comprobamos si el token maestro existe
	masterToken := os.Getenv("MASTER_ADMIN_TOKEN")
	if masterToken == "" {
		fmt.Println("############################################################")
		fmt.Println("# [CRITICAL WARNING] MASTER_ADMIN_TOKEN is NOT SET!        #")
		fmt.Println("# Admin access via WHMCS and SSO will FAIL (401).         #")
		fmt.Println("# Please check your .env file and Restart the container.  #")
		fmt.Println("############################################################")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("[BOOT] Server listening on port %s. Awaiting WHMCS & Agents...\n", port)
	r.Run(":" + port)
}

