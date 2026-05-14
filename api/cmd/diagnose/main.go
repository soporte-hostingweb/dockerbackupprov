package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Estructuras autónomas para el diagnóstico (V15.5)
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
		fmt.Println("Uso: go run main.go <agent_id>")
		return
	}
	agentID := os.Args[1]

	dsn := "host=dbp-postgres user=dbpuser password=dbppass dbname=dbp-recovery port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a DB:", err)
	}

	fmt.Printf("\n=== DIAGNÓSTICO HW CLOUD RECOVERY: %s ===\n", agentID)

	var agent AgentStatus
	if err := db.First(&agent, "id = ?", agentID).Error; err != nil {
		fmt.Printf("❌ ERROR: Agente no encontrado.\n")
		return
	}
	fmt.Printf("ESTADO: %s | MANTENIMIENTO: %v | VISTO: %s\n", agent.HealthStatus, agent.Maintenance, agent.UpdatedAt.Format("15:04:05"))

	var plan TenantPlan
	db.First(&plan, "token = ?", agent.Token)
	fmt.Printf("PLAN ASIGNADO: %s\n", plan.Plan)

	var jobs []Job
	db.Where("agent_id = ? AND status = ?", agentID, "pending").Order("priority DESC").Find(&jobs)
	fmt.Printf("JOBS EN COLA: %d\n", len(jobs))

	if agent.CmdResult != "" {
		fmt.Printf("\nÚLTIMO RESULTADO (%d chars):\n", len(agent.CmdResult))
		preview := agent.CmdResult
		if len(preview) > 150 { preview = preview[:150] + "..." }
		fmt.Println(preview)
	}

	fmt.Println("==========================================")
}
