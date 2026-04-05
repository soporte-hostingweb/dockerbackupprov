package main

import (
	"fmt"
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
	
	// Endpoint para el Dashboard (Público para el MVP, en prod requiere otro Auth)
	v1Agent.GET("/status", func(c *gin.Context) {
		c.JSON(200, agentStatusStore)
	})

	v1Agent.Use(AuthMiddleware()) // <-- Cada Request a esta URL debe tener Token
	{
		v1Agent.POST("/heartbeat", ReceiveHeartbeat)
		v1Agent.POST("/backup/complete", ReceiveBackupCompletion)
	}

	fmt.Println("[BOOT] Server listening on port 8080. Awaiting WHMCS & Agents...")
	r.Run(":8080")
}
