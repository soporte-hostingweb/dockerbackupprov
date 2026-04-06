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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}






// BackupConfig almacena qué rutas se han seleccionado para respaldar
type BackupConfig struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"index" json:"token"`
	AgentID   string `gorm:"index" json:"agent_id"`
	Paths     string `json:"paths"`    // JSON array de paths
	Schedule  string `json:"schedule"` // manual, daily_2am, every_1h, etc.
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
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	CreatedAt    time.Time `json:"timestamp"`
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
	db.AutoMigrate(&AgentStatus{}, &BackupConfig{}, &UserSettings{}, &BackupActivity{})


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
