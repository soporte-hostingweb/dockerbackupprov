package main

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// --- MODELOS DE BASE DE DATOS ---

// AgentStatus almacena el último latido y estado de cada VPS
type AgentStatus struct {
	ID           string    `gorm:"primaryKey" json:"agent_id"`
	Token        string    `json:"token"`
	Status       string    `json:"status"`
	LastSeen     time.Time `json:"last_sync"`
	LastSeenUnix int64     `json:"last_seen_unix"`
	OS           string    `json:"os"`
	Containers   string    `json:"containers"` // Guardamos como JSON string
	Explorer     string    `json:"explorer"`   // Guardamos como JSON string
	Snapshots    string    `json:"snapshots"`  // JSON string de restic snapshots
	CreatedAt    time.Time

	UpdatedAt    time.Time
}

// BackupConfig almacena qué rutas se han seleccionado para respaldar
type BackupConfig struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"index" json:"token"`
	AgentID   string `gorm:"index" json:"agent_id"`
	Paths     string `json:"paths"` // JSON array de paths
	CreatedAt time.Time
}

// UserSettings almacena las credenciales de Wasabi por cliente
type UserSettings struct {
	ID           uint   `gorm:"primaryKey"`
	Token        string `gorm:"uniqueIndex" json:"token"`
	WasabiKey    string `json:"wasabi_key"`
	WasabiSecret string `json:"wasabi_secret"`
	WasabiBucket string `json:"wasabi_bucket"`
	WasabiRegion string `json:"wasabi_region"`
	ResticPass   string `json:"restic_password"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InitDB inicializa la conexión a PostgreSQL y realiza las migraciones
func InitDB() {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, pass, name, port)

	fmt.Printf("[DB] Connecting to PostgreSQL at %s:%s...\n", host, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("[DB ERROR] Failed to connect: %v\n", err)
		os.Exit(1)
	}

	// Auto-Migración de esquemas
	fmt.Println("[DB] Running automatic migrations...")
	db.AutoMigrate(&AgentStatus{}, &BackupConfig{}, &UserSettings{})

	DB = db
	fmt.Println("[DB] PostgreSQL is ready and migrated.")
}
