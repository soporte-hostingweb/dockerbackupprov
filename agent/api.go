package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// BackupMetrics representa el payload de telemetría a enviar al Control Plane
type BackupMetrics struct {
	AgentID      string `json:"agent_id"`
	Status       string `json:"status"`
	TotalSizeMB  int    `json:"total_size_mb"`
	DurationSecs int    `json:"duration_secs"`
	SnapshotID   string `json:"snapshot_id"`
	Timestamp    int64  `json:"timestamp"`
}

// HeartbeatPayload para enviar la lista de contenedores y estado del sistema
type HeartbeatPayload struct {
	AgentID      string              `json:"agent_id"`
	Containers   []string            `json:"containers"`
	ExplorerData map[string][]string `json:"explorer_data"`
	OS           string              `json:"os"`
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
	req.Header.Set("Authorization", os.Getenv("DBP_API_TOKEN"))
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

// ReportHeartbeat envía el estado pasivo del servidor (lista de contenedores y sus carpetas)
func ReportHeartbeat(agentID string, containers []string, explorerData map[string][]string) {
	apiEndpoint := os.Getenv("DBP_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.hwperu.com"
	}
	url := fmt.Sprintf("%s/v1/agent/heartbeat", apiEndpoint)

	payloadObj := HeartbeatPayload{
		AgentID:      agentID,
		Containers:   containers,
		ExplorerData: explorerData,
		OS:           "Linux (Docker)",
	}

	payload, _ := json.Marshal(payloadObj)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", os.Getenv("DBP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[API ERROR] Network failure sending heartbeat: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[API ERROR] Heartbeat REJECTED by server. Status: %s. Check your DBP_API_TOKEN!\n", resp.Status)
		return
	}

	fmt.Printf("[API] Heartbeat (Containers: %d) sent. Status: OK\n", len(containers))
}


