package main

import (
	"fmt"
	_ "embed"
	"github.com/gin-gonic/gin"
)

//go:embed install.sh
var installScript []byte

func main() {
	fmt.Println("[BOOT] Starting Docker Backup Pro Control Plane API...")

	// Desactiva el debug log intenso de gin para producción
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()

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
	v1Agent.Use(AuthMiddleware()) // <-- Cada Request a esta URL debe tener Token
	{
		v1Agent.POST("/heartbeat", ReceiveHeartbeat)
		v1Agent.POST("/backup/complete", ReceiveBackupCompletion)
	}

	fmt.Println("[BOOT] Server listening on port 8080. Awaiting WHMCS & Agents...")
	r.Run(":8080")
}
