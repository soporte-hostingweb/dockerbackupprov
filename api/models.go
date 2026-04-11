package main

// BackupCompletePayload mapea el JSON que nos envía el Agente después de usar Restic
type BackupCompletePayload struct {
	AgentID      string `json:"agent_id"`
	Status       string `json:"status"`
	TotalSizeMB  int    `json:"total_size_mb"`
	DurationSecs int    `json:"duration_secs"`
	SnapshotID   string `json:"snapshot_id"`
	Timestamp    int64  `json:"timestamp"` // finished_at
	StartedAt    int64  `json:"started_at"`
}


// HeartbeatPayload mapea el JSON del chequeo periódico de estado
type HeartbeatPayload struct {
	AgentID      string              `json:"agent_id"`
	OS           string              `json:"os"`
	Containers   []string            `json:"containers"`
	ExplorerData map[string][]string `json:"explorer_data"`
	FreeSpace    string              `json:"free_space"`
	TotalSpace   string              `json:"total_space"`
}

// UserSettingsPayload mapea el JSON de configuración de Wasabi (V9.0 SaaS Pro)
type UserSettingsPayload struct {
	WasabiKey      string `json:"wasabi_key"`
	WasabiSecret   string `json:"wasabi_secret"`
	WasabiBucket   string `json:"wasabi_bucket"`
	WasabiRegion   string `json:"wasabi_region"`
	S3Endpoint     string `json:"s3_endpoint"`
	ResticPass     string `json:"restic_password"`
	WebhookURL     string `json:"webhook_url"`
	WebhookEvents  string `json:"webhook_events"` // Formato: "backup_failed,agent_offline"
}
