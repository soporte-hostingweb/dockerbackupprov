package main

// BackupCompletePayload mapea el JSON que nos envía el Agente luego de usar Restic
type BackupCompletePayload struct {
	AgentID      string `json:"agent_id"`
	Status       string `json:"status"`
	TotalSizeMB  int    `json:"total_size_mb"`
	DurationSecs int    `json:"duration_secs"`
	SnapshotID   string `json:"snapshot_id"`
	Timestamp    int64  `json:"timestamp"`
}

// HeartbeatPayload mapea el JSON del chequeo periódico de estado
type HeartbeatPayload struct {
	AgentID      string              `json:"agent_id"`
	OS           string              `json:"os"`
	Containers   []string            `json:"containers"`
	ExplorerData map[string][]string `json:"explorer_data"`
}


