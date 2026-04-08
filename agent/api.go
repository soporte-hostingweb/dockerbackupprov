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
	req.Header.Set("Authorization", "Bearer " + os.Getenv("DBP_API_TOKEN"))
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
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DBP_API_TOKEN"))
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
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DBP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err == nil { resp.Body.Close() }
}

// ReportHeartbeat envía el estado del agente y recibe órdenes (V4.5.5)
func ReportHeartbeat(agentID string, containers []string, explorer map[string][]string, snapshots []interface{}, syncing bool, activePID int, lastBackupAt int64, freeSpace string, totalSpace string) (bool, string, bool, error) {
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
		LastBackupAt int64                  `json:"last_backup_unix"`
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
		LastBackupAt: lastBackupAt,
	}

	payload, _ := json.Marshal(payloadObj)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer " + os.Getenv("DBP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		LogInfo("[API ERROR] Network failure after %v: %v", latency, err)
		return false, "none", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		LogInfo("[API ERROR] Heartbeat REJECTED (Status: %s) after %v", resp.Status, latency)
		return false, "none", false, fmt.Errorf("heartbeat rejected: %s", resp.Status)
	}

	var result struct {
		Maintenance  bool   `json:"maintenance"`
		PendingForce string `json:"pending_force"`
		KillSync     bool   `json:"kill_sync"`
		CmdTask      string `json:"cmd_task"`
		CmdParam     string `json:"cmd_param"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "none", false, err
	}

	// Hot-Fix: Si hay una tarea de comando, la devolvemos al ciclo principal (Simplificado para V4.2.4)
	if result.CmdTask != "" && result.CmdTask != "none" {
		LogInfo("[TASK] Received remote command after %v: %s (%s)", latency, result.CmdTask, result.CmdParam)
		return result.Maintenance, result.CmdTask + ":" + result.CmdParam, result.KillSync, nil
	}

	LogInfo("[API] Heartbeat sent (RTT: %v). Maint: %v, Force: %s", latency, result.Maintenance, result.PendingForce)
	
	return result.Maintenance, result.PendingForce, result.KillSync, nil
}

// ReportTaskResult envía el resultado de un comando (ej: ls snapshot) al Control Plane (V4.2.4)
func ReportTaskResult(agentID string, task string, result string) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/task/result", apiEndpoint)
	
	payloadObj := struct {
		AgentID string `json:"agent_id"`
		Task    string `json:"task"`
		Result  string `json:"result"`
	}{
		AgentID: agentID,
		Task:    task,
		Result:  result,
	}

	payload, _ := json.Marshal(payloadObj)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer " + os.Getenv("DBP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second} // El 'ls' puede ser lento
	_, _ = client.Do(req)
}





// AgentConfigV2 contiene la respuesta extendida de la API (V5.1.1)
type AgentConfigV2 struct {
	Status          string   `json:"status"`
	Paths           []string `json:"paths"`
	Schedule        string   `json:"schedule"`
	Retention       int      `json:"retention"`
	TimeZone        string   `json:"timezone"`        // V7.1
	CustomSchedule  string   `json:"custom_schedule"` // V7.2
	FullRepoURL     string   `json:"full_repo_url"`
	ResticPassword  string   `json:"restic_password"`
	WasabiKey       string   `json:"wasabi_key"`
	WasabiSecret    string   `json:"wasabi_secret"`
}



// GetAgentConfig consulta a la API la selección de carpetas específica para este VPS
func GetAgentConfig(agentID string) (*AgentConfigV2, error) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/config?agent_id=%s", apiEndpoint, agentID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer " + os.Getenv("DBP_API_TOKEN"))

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
	token := os.Getenv("DBP_API_TOKEN")

	reqBody, _ := json.Marshal(map[string]interface{}{
		"activity_id": activityID,
		"agent_id":    agentID,
		"type":        taskType,
		"status":      status,
		"message":     message,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)

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

