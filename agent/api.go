package main

import (
	"encoding/json"
	"fmt"
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

// ReportMetrics envía el estado final al API Central HTTP
func ReportMetrics(metrics BackupMetrics) {
	fmt.Println("\n[API] Compiling Telemetry Data...")

	payload, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		fmt.Printf("[API ERROR] Could not marshal metrics: %v\n", err)
		return
	}

	// Mocking request
	fmt.Printf("[API POST] https://api.dockerbackuppro.com/v1/agent/backup/complete\n%s\n", string(payload))

	// Simular envío
	time.Sleep(300 * time.Millisecond)

	fmt.Println("[API] Server responded: 200 OK (Job status recorded)")
	
	/* Ejecución Real
	req, _ := http.NewRequest("POST", "https://api.dockerbackuppro.com/v1/agent/backup/complete", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer " + os.Getenv("DBP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("API Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	*/
}
