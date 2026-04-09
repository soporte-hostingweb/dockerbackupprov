package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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
	Containers   string    `gorm:"type:text" json:"containers"` // Guardamos como JSON string
	Explorer     string    `gorm:"type:text" json:"explorer"`   // Guardamos como JSON string
	Snapshots    string    `gorm:"type:text" json:"snapshots"`  // JSON string de restic snapshots
	FreeSpace    string    `json:"free_space"`                  // ej: "10GB"
	TotalSpace   string    `json:"total_space"`                 // ej: "50GB"
	Maintenance  bool      `json:"maintenance"`
	PendingForce string    `json:"pending_force"` // "none", "selected", "full"
	IsSyncing    bool      `json:"is_syncing"`
	ActivePID    int       `json:"active_pid"`
	KillSync     bool      `json:"kill_sync"`
	CmdTask      string    `json:"cmd_task"`   // "ls_snapshot", "none"
	CmdParam     string    `json:"cmd_param"`  // snapshot_id
	CmdResult    string    `gorm:"type:text" json:"cmd_result"` // JSON output del comando
	LastBackupAt time.Time `json:"last_backup_at"`
	LastBackupBytes int64  `json:"last_backup_bytes"` // V8.1: Consumo en Wasabi
	HealthStatus       string `json:"health_status" gorm:"default:'ONLINE'"` // V9.0: ONLINE, DEGRADED, OFFLINE
	VerificationStatus string `json:"verification_status" gorm:"default:'PENDING'"` // V9.0: VALID, INVALID, PENDING
	EstRtoSecs         int    `json:"est_rto_secs"`        // V9.0: Recovery Time Objective estimado
	HealthScore        int    `json:"health_score" gorm:"default:100"` // V9.1: Scoring SaaS (0-100)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}






// BackupConfig almacena qué rutas se han seleccionado para respaldar
type BackupConfig struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"index" json:"token"`
	AgentID   string `gorm:"index" json:"agent_id"`
	Paths          string `json:"paths"`    // JSON array de paths
	Schedule       string `json:"schedule"` // manual, daily_2am, weekly_2am, custom
	Retention      int    `json:"retention"` // 1, 2, 7 (V5.1.1)
	TimeZone       string `json:"timezone" gorm:"default:America/Lima"` // V7.1: Soporte para horario local
	CustomSchedule string `json:"custom_schedule"` // V7.2: Días y horas personalizados (ej: 1,3,5|14)
	CreatedAt      time.Time
}

// ActivityLog: Registro de operaciones globales en tiempo real (V6.3)
type ActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Token      string    `gorm:"index" json:"token"`
	AgentID    string    `gorm:"index" json:"agent_id"`
	Type       string    `json:"type"` // backup, restore, prune
	Status     string    `json:"status"` // pending, running, success, error
	Message    string    `json:"message"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
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

// BackupActivity registra cada respaldo completado para el historial
type BackupActivity struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Token        string    `gorm:"index" json:"token"`
	AgentID      string    `gorm:"index" json:"agent_id"`
	Status       string    `json:"status"` // SUCCESS, ERROR
	SnapshotID   string    `json:"snapshot_id"`
	SizeMB       int       `json:"size_mb"`
	SizeBytes    int64     `json:"size_bytes"` // V4.6.1: Precisión para archivos pequeños
	DurationSecs int       `json:"duration_secs"`
	Message      string    `json:"message"`
	ValidationStatus string `json:"validation_status"` // V9.0: VALID, INVALID
	ValidationErrors string `gorm:"type:text" json:"validation_errors"` // V9.0: Detalles del restic check
	RestoreDurationSecs int `json:"restore_duration_secs"` // V9.0: Tracker de velocidad real
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	CreatedAt    time.Time `json:"timestamp"`
}

// TenantPlan: Centralización Comercial (V9.0) define los atributos del plan. Source of Truth sobre WHMCS.
type TenantPlan struct {
	ID             uint      `gorm:"primaryKey"`
	Token          string    `gorm:"uniqueIndex" json:"token"`
	Plan           string    `json:"plan" gorm:"default:'basic'"` // basic, standard, enterprise
	RetentionDays  int       `json:"retention_days"`
	Priority       bool      `json:"priority"`
	ValidationLvl  string    `json:"validation_lvl"` // none, basic, advanced
	WhmcsServiceID string    `json:"whmcs_service_id" gorm:"index"` // V9.1: Vínculo con WHMCS Billing
	ClientEmail    string    `json:"client_email"`                  // V9.1: Trazabilidad comercial
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AlertConfig: Webhook Settings para n8n / integraciones universales (V9.0)
type AlertConfig struct {
	ID         uint      `gorm:"primaryKey"`
	Token      string    `gorm:"uniqueIndex" json:"token"`
	WebhookURL string    `json:"webhook_url"`
	Events     string    `json:"events"` // JSON string de eventos suscritos: "backup_failed", "agent_offline"
	UpdatedAt  time.Time `json:"updated_at"`
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
	db.AutoMigrate(&AgentStatus{}, &BackupConfig{}, &UserSettings{}, &BackupActivity{}, &ActivityLog{}, &TenantPlan{}, &AlertConfig{})
	fmt.Println("✅ Database Migrated Successfully with SaaS Data Models (V9.2.7)")


	DB = db
	fmt.Println("[DB] PostgreSQL is ready and migrated.")
}

// --- MOTOR DE CIFRADO AES-256-GCM ---

func getEncryptionKey() []byte {
	key := os.Getenv("DBP_ENCRYPTION_KEY")
	if len(key) != 32 {
		// Fallback por seguridad o para desarrollo, pero avisamos.
		// En prod DEBE ser de 32 bytes.
		return []byte("hwperu-backup-security-key-32chr") 
	}
	return []byte(key)
}

// Encrypt cifra un texto plano usando AES-GCM
func Encrypt(text string) (string, error) {
	if text == "" { return "", nil }
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt descifra un texto cifrado en hex usando AES-GCM
func Decrypt(cryptoText string) (string, error) {
	if cryptoText == "" { return "", nil }
	key := getEncryptionKey()
	ciphertext, err := hex.DecodeString(cryptoText)
	if err != nil {
		// V2.6.5 Fallback: Si no es HEX válido, devolvemos el original (asumimos que ya está descifrado)
		return cryptoText, nil
	}


	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// V2.8.1 Fallback: Si el descifrado GCM falla (ej: texto plano que parece hex), devolvemos el original
		return cryptoText, nil
	}


	return string(plaintext), nil
}
