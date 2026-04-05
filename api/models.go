package main

// BackupCompletePayload mapea el JSON que nos envía el Agente después de usar Restic
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

// UserSettingsPayload mapea el JSON de configuración de Wasabi
type UserSettingsPayload struct {
	WasabiKey    string `json:"wasabi_key"`
	WasabiSecret string `json:"wasabi_secret"`
	WasabiBucket string `json:"wasabi_bucket"`
	WasabiRegion string `json:"wasabi_region"`
	ResticPass   string `json:"restic_password"`
}
