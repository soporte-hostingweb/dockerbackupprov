package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"runtime"
)

// AgentCredentials almacena la identidad persistente (V13)
type AgentCredentials struct {
	AgentID     string `json:"agent_id"`
	ApiKey      string `json:"api_key"`
	Fingerprint string `json:"fingerprint"`
}

var CurrentCreds AgentCredentials


// BackupMetrics representa el payload de telemetría a enviar al Control Plane
type BackupMetrics struct {
	AgentID          string `json:"agent_id"`
	Status           string `json:"status"`
	TotalSizeMB      int    `json:"total_size_mb"`
	TotalSizeBytes   int64  `json:"total_size_bytes"` // V4.6.1: Precisión para archivos pequeños
	FileCount        int    `json:"file_count"`
	UploadSpeedKbps  int    `json:"upload_speed_kbps"`
	DurationSecs     int    `json:"duration_secs"`
	SnapshotID       string `json:"snapshot_id"`
	Timestamp        int64  `json:"timestamp"` // finished_at
	StartedAt        int64  `json:"started_at"`
}



// HeartbeatPayload para enviar la lista de contenedores y estado del sistema
type HeartbeatPayload struct {
	AgentID        string              `json:"agent_id"`
	Containers     []string            `json:"containers"`
	ExplorerData   map[string][]string `json:"explorer_data"`
	Snapshots      []interface{}       `json:"snapshots"`
	OS             string              `json:"os"`
	IsSyncing      bool                `json:"is_syncing"`
	ActivePID      int                 `json:"active_pid"`
	LastBackupUnix int64               `json:"last_backup_unix"`
}





// ReportMetrics envía el estado final al API Central HTTP
func ReportMetrics(metrics BackupMetrics) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/backup/complete", apiEndpoint)
	
	payload, err := json.Marshal(metrics)

	if err != nil {
		fmt.Printf("[API ERROR] Could not marshal metrics: %v\n", err)
		return
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[API ERROR] Failed to send metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("[API] Metrics sent to %s. Status: %s\n", url, resp.Status)
}

// ReportRestoreMetrics envía el RTO calculado a la nube (V9.0)
func ReportRestoreMetrics(agentID, snapshotID string, durationSecs int) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" { apiEndpoint = "https://api.hwperu.com" }
	
	url := fmt.Sprintf("%s/v1/agent/restore/metrics", apiEndpoint)
	payload, _ := json.Marshal(map[string]interface{}{
		"agent_id":      agentID,
		"snapshot_id":   snapshotID,
		"total_seconds": durationSecs,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil { resp.Body.Close() }
}

// ReportVerification envía el resultado destructivo o check del repo (V9.0)
func ReportVerification(agentID, snapshotID, status string, errors []string) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" { apiEndpoint = "https://api.hwperu.com" }
	
	url := fmt.Sprintf("%s/v1/agent/verification/report", apiEndpoint)
	payload, _ := json.Marshal(map[string]interface{}{
		"agent_id":    agentID,
		"snapshot_id": snapshotID,
		"status":      status,
		"errors":      errors,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err == nil { resp.Body.Close() }
}

// ReportHeartbeat envía el estado del agente y recibe órdenes (V4.5.5)
func ReportHeartbeat(agentID string, containers []string, explorer map[string][]string, snapshots []interface{}, syncing bool, activePID int, lastBackupAt int64, freeSpace string, totalSpace string) (bool, string, uint, bool, error) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}

	url := fmt.Sprintf("%s/v1/agent/heartbeat", apiEndpoint)
	
	payloadObj := struct {
		AgentID      string                 `json:"agent_id"`
		Containers   []string               `json:"containers"`
		ExplorerData map[string][]string    `json:"explorer_data"`
		Snapshots    []interface{}          `json:"snapshots"`
		FreeSpace    string                 `json:"free_space"`
		TotalSpace   string                 `json:"total_space"`
		OS           string                 `json:"os"`
		IsSyncing    bool                   `json:"is_syncing"`
		ActivePID    int                    `json:"active_pid"`
		Version      string                 `json:"version"`          // V14
		LastBackupAt int64                  `json:"last_backup_unix"`
		HasDocker    bool                   `json:"has_docker"`
		DetectedStack StackInfo             `json:"detected_stack"`
	}{
		AgentID:      agentID,
		Containers:   containers,
		ExplorerData: explorer,
		Snapshots:    snapshots,
		FreeSpace:    freeSpace,
		TotalSpace:   totalSpace,
		OS:           runtime.GOOS,
		IsSyncing:    syncing,
		ActivePID:    activePID,
		Version:      Version,
		LastBackupAt: lastBackupAt,
		HasDocker:    DetectStack().HasDocker,
		DetectedStack: DetectStack(),
	}

	payload, _ := json.Marshal(payloadObj)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		LogInfo("[API ERROR] Network failure after %v: %v", latency, err)
		return false, "none", 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		LogInfo("[API ERROR] Heartbeat REJECTED (Status: %s) after %v", resp.Status, latency)
		return false, "none", 0, false, fmt.Errorf("heartbeat rejected: %s", resp.Status)
	}

	var result struct {
		Maintenance  bool   `json:"maintenance"`
		PendingForce string `json:"pending_force"`
		KillSync     bool   `json:"kill_sync"`
		CmdTask      string `json:"cmd_task"`
		CmdParam     string `json:"cmd_param"`
		JobID        uint   `json:"cmd_job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "none", 0, false, err
	}

	// Hot-Fix: Si hay una tarea de comando, la devolvemos al ciclo principal (Simplificado para V4.2.4)
	if result.CmdTask != "" && result.CmdTask != "none" {
		LogInfo("[TASK] Received remote command after %v: %s (%s) [JobID: %d]", latency, result.CmdTask, result.CmdParam, result.JobID)
		return result.Maintenance, result.CmdTask + ":" + result.CmdParam, result.JobID, result.KillSync, nil
	}

	LogInfo("[API] Heartbeat sent (RTT: %v). Maint: %v, Force: %s", latency, result.Maintenance, result.PendingForce)
	
	// V14: Soporte para actualización de versión forzada (Opcional en esta fase)
	
	return result.Maintenance, result.PendingForce, result.JobID, result.KillSync, nil
}

// ReportTaskResult envía el resultado de un comando al Control Plane (V10.1: Incluye JobID)
func ReportTaskResult(agentID string, task string, result string, jobID uint) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/task/result", apiEndpoint)
	
	payloadObj := struct {
		AgentID string `json:"agent_id"`
		Task    string `json:"task"`
		Result  string `json:"result"`
		JobID   uint   `json:"job_id"`
	}{
		AgentID: agentID,
		Task:    task,
		Result:  result,
		JobID:   jobID,
	}

	payload, _ := json.Marshal(payloadObj)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	_, _ = client.Do(req)
}





// AgentConfigV2 contiene la respuesta extendida de la API (V14.1: Control SaaS Completo)
type AgentConfigV2 struct {
	Status          string   `json:"status"`
	Paths           []string `json:"paths"`
	Schedule        string   `json:"schedule"`
	Retention       int      `json:"retention"`
	TimeZone        string   `json:"timezone"`
	CustomSchedule  string   `json:"custom_schedule"`
	FullRepoURL     string   `json:"full_repo_url"`
	ResticPassword  string   `json:"restic_password"`
	WasabiKey       string   `json:"wasabi_key"`
	WasabiSecret    string   `json:"wasabi_secret"`
	// V14.1: Campos de Control SaaS - El Agente Solo Ejecuta, No Decide
	ProtectionLevel string   `json:"protection_level"` // Basic, Advanced, Total
	SnapshotMode    string   `json:"snapshot_mode"`    // live | consistent
	IsAutoManaged   bool     `json:"is_auto_managed"`  // true = API manda, false = usuario configura
	IsDynamic       bool     `json:"is_dynamic"`       // true = incluir todos los contenedores dinámicamente
}

// ResolveBackupPaths convierte la configuración del API a paths reales de restic.
// El agente NUNCA decide las rutas por sí solo. Siempre lee del backend.
// Si encuentra el sentinel [ALL_SYSTEM_ROOT], expande a la raíz del host con exclusiones de producción.
func ResolveBackupPaths(config *AgentConfigV2) []string {
	if config == nil {
		return []string{}
	}

	// Buscar el sentinel de Protección Total
	for _, p := range config.Paths {
		if p == "[ALL_SYSTEM_ROOT]" {
			// Snapshot completo: raíz del host montada en /host_root
			// Las exclusiones se agregan automáticamente en restic.go (GlobalExcludes)
			// Esto produce: restic backup /host_root --exclude /host_root/proc --exclude ...
			LogInfo("[CONFIG] Protection Level: TOTAL - Full system snapshot via /host_root")
			return []string{"/host_root"}
		}
	}

	// Si dynamic y sin paths, incluir todos los contenedores detectados como [ALL_TARGETS]
	if config.IsDynamic && len(config.Paths) == 0 {
		LogInfo("[CONFIG] Protection Level: ADVANCED - Dynamic tracking, all containers")
		return []string{"/host_root"}
	}

	// Paths manuales (Basic o Advanced con selección específica)
	LogInfo("[CONFIG] Protection Level: %s - %d specific paths", config.ProtectionLevel, len(config.Paths))
	return config.Paths
}



// GetAgentConfig consulta a la API la selección de carpetas específica para este VPS
func GetAgentConfig(agentID string) (*AgentConfigV2, error) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/config?agent_id=%s", apiEndpoint, agentID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("config fetch failed: %s", resp.Status)
	}

	var config AgentConfigV2
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ReportActivity: Notifica el estado de una tarea al Control Plane (V6.3)
func ReportActivity(activityID uint, agentID string, taskType string, status string, message string) uint {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" { apiEndpoint = "https://api.hwperu.com" }
	url := fmt.Sprintf("%s/v1/agent/activity/report", apiEndpoint)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"activity_id": activityID,
		"agent_id":    agentID,
		"type":        taskType,
		"status":      status,
		"message":     message,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-ID", CurrentCreds.AgentID)
	req.Header.Set("X-Agent-Key", CurrentCreds.ApiKey)
	req.Header.Set("X-Agent-Fingerprint", CurrentCreds.Fingerprint)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Could not report activity: %v\n", err)
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		ActivityID uint `json:"activity_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.ActivityID
}

