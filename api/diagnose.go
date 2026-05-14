package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Reutilizamos las estructuras mínimas necesarias
type AgentStatus struct {
	ID              string    `gorm:"primaryKey"`
	Token           string
	HealthStatus    string
	Maintenance     bool
	PendingForce    string
	CmdTask         string
	CmdResult       string `gorm:"type:text"`
	UpdatedAt       time.Time
}

type TenantPlan struct {
	Token string `gorm:"primaryKey"`
	Plan  string
}

type Job struct {
	ID        uint `gorm:"primaryKey"`
	AgentID   string
	Type      string
	Status    string
	Param     string
	Priority  int
	CreatedAt time.Time
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run diagnose.go <agent_id>")
		return
	}
	agentID := os.Args[1]

	dsn := "host=dbp-postgres user=dbpuser password=dbppass dbname=dbp-recovery port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a DB:", err)
	}

	fmt.Printf("\n=== DIAGNÓSTICO HW CLOUD RECOVERY: %s ===\n", agentID)

	// 1. Verificar Agente
	var agent AgentStatus
	if err := db.First(&agent, "id = ?", agentID).Error; err != nil {
		fmt.Printf("❌ ERROR: Agente no encontrado en la base de datos.\n")
		return
	}
	fmt.Printf("ESTADO: %s | MANTENIMIENTO: %v | VISTO: %s\n", agent.HealthStatus, agent.Maintenance, agent.UpdatedAt.Format("15:04:05"))

	// 2. Verificar Plan
	var plan TenantPlan
	db.First(&plan, "token = ?", agent.Token)
	fmt.Printf("PLAN ASIGNADO (WHMCS): %s\n", plan.Plan)

	// 3. Verificar Señales Legacy
	fmt.Printf("SEÑAL FORCE (Legacy): %s | TAREA ACTUAL: %s\n", agent.PendingForce, agent.CmdTask)

	// 4. Verificar Jobs Pendientes (V15)
	var jobs []Job
	db.Where("agent_id = ? AND status = ?", agentID, "pending").Order("priority DESC").Find(&jobs)
	fmt.Printf("JOBS PENDIENTES EN COLA: %d\n", len(jobs))
	for _, j := range jobs {
		fmt.Printf("  - [%d] TIPO: %s | PRIORIDAD: %d | CREADO: %s\n", j.ID, j.Type, j.Priority, j.CreatedAt.Format("15:04:05"))
	}

	// 5. Analizar Último Resultado de Comando
	if agent.CmdResult != "" {
		fmt.Printf("\nÚLTIMO RESULTADO (CMD_RESULT):\n")
		fmt.Printf("LONGITUD: %d caracteres\n", len(agent.CmdResult))
		preview := agent.CmdResult
		if len(preview) > 200 {
			preview = preview[:200] + "... [TRUNCADO]"
		}
		fmt.Printf("VISTA PREVIA:\n%s\n", preview)
		
		// Intento de parseo JSON
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(agent.CmdResult), &js); err == nil {
			fmt.Println("FORMATO DETECTADO: JSON ✅")
		} else {
			fmt.Println("FORMATO DETECTADO: TEXTO PLANO/RESTIC ✅")
		}
	} else {
		fmt.Printf("\nÚLTIMO RESULTADO: [VACÍO] ⚠️\n")
	}

	fmt.Println("==========================================\n")
}
