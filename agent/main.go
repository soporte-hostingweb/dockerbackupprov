package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)



var IsSyncing bool = false
var ActivePID int = 0

func main() {

	fmt.Println("🚀 Docker Backup Pro Agent Starting...")
	
	// Cargar persistencia si existe (V2.4)
	agentID := GetPersistentID()
	var lastBackupUnix int64 = 0

	// Loop principal (V2.0)
	for {
		fmt.Println("\n--- [CYCLE START: " + time.Now().UTC().Format(time.RFC822) + "] ---")

		// Descubrimiento de contenedores (V1.0)
		containerNames, _ := GetRunningContainers()
		fmt.Printf("[INFO] Discovered %d containers.\n", len(containerNames))

		// Preparar explorer data para reporte (V3.8.4: Path Resolution)
		explorerData := make(map[string][]string)
		pathToRealMap := make(map[string]string) // Mapea "📂 Carpeta" -> "/host_root/..."
		var backupPaths []string

		for _, name := range containerNames {
			if name == "dbp-client-agent" { continue }
			
			// 1. Obtener los montajes REALES del host usando docker inspect (V3.8.1)
			hostMounts := GetContainerMounts(name)
			
			for _, hostPath := range hostMounts {
				bridgePath := "/host_root" + hostPath
				
				if info, err := os.Stat(bridgePath); err == nil && info.IsDir() {
					backupPaths = append(backupPaths, bridgePath)
					
					// V3.8.3: Discovery Total (Archivos y Carpetas)
					subItems := ScanVolumeFolders(bridgePath)
					for _, item := range subItems {
						// Para evitar colisiones entre contenedores, prefijamos la entrada en el mapa si es necesario, 
						// pero el UI espera solo el nombre para mostrar. 
						// Guardamos la resolución absoluta para el backup.
						explorerData[name] = append(explorerData[name], item)
						
						// Resolvemos: "📄 dump.rdb" -> "/host_root/var/lib/docker/volumes/redis/_data/dump.rdb"
						itemName := strings.TrimPrefix(strings.TrimPrefix(item, "📂 "), "📄 ")
						realPath := bridgePath + "/" + itemName
						pathToRealMap[item] = realPath
					}
					
					// Opción para respaldar el volumen entero
					fullVolEntry := "📂 [Full Volume] " + name
					explorerData[name] = append(explorerData[name], fullVolEntry)
					pathToRealMap[fullVolEntry] = bridgePath
				}
			}
		}




		// 1. Obtener Configuración Primaria (V3.4.2)
		fmt.Printf("[CONFIG] Fetching policy for agent %s...\n", agentID)
		config, errConfig := GetAgentConfig(agentID)
		if errConfig != nil {
			fmt.Printf("[ERROR] Could not fetch config: %v\n", errConfig)
			config = &AgentConfigV2{Status: "no_config"}
		}

		// 2. Resolver Repositorio y Credenciales (V3.4.2)
		var repo, pass, key, secret string
		if config != nil && config.FullRepoURL != "" {
			repo = config.FullRepoURL
			pass = config.ResticPassword
			key = config.WasabiKey
			secret = config.WasabiSecret
		} else {
			repo = os.Getenv("RESTIC_REPOSITORY")
			pass = os.Getenv("RESTIC_PASSWORD")
			key = os.Getenv("AWS_ACCESS_KEY_ID")
			secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
		}

		// 3. Obtener Snapshots para el reporte (V3.5.1)
		var snapshots []interface{}
		if repo != "" && pass != "" {
			snapshotsRaw := GetSnapshotsJSON(repo, pass, key, secret)
			if errU := json.Unmarshal(snapshotsRaw, &snapshots); errU != nil {
				fmt.Printf("[CRITICAL-JSON] Failed to unmarshal snapshots: %v\n", errU)
				snapshots = []interface{}{}
			} else {
				fmt.Printf("[SNAPSHOTS] Successfully detected %d snapshots in repository.\n", len(snapshots))
			}
		}

		// 4. Reportar Heartbeat
		maint, force, kill, errHeart := ReportHeartbeat(agentID, containerNames, explorerData, snapshots, IsSyncing, ActivePID, lastBackupUnix)
		if errHeart != nil {
			fmt.Printf("[WARNING] Heartbeat failed: %v\n", errHeart)
		}

		if kill && IsSyncing && ActivePID > 0 {
			fmt.Printf("[KILL] Remote terminate requested for PID %d...\n", ActivePID)
			if proc, errP := os.FindProcess(ActivePID); errP == nil {
				proc.Signal(os.Interrupt)
				time.Sleep(2 * time.Second)
				proc.Kill()
			}
			IsSyncing = false
			ActivePID = 0
		}

		if maint {
			fmt.Println("[INFO] Maintenance Mode Active. Pausing backup cycles.")
			time.Sleep(60 * time.Second)
			continue
		}

		// 5. Validar Repositorio
		if repo != "" {
			fmt.Printf("[RESTIC] Validating S3 Wasabi Repository...\n")
			_ = EnsureResticRepo(repo, pass, key, secret)
		}

		// 6. Lógica de Programación / Scheduler
		shouldRun := false
		now := time.Now()

		if force != "none" && force != "" {
			shouldRun = true
		} else if config.Schedule == "every_1h" {
			if time.Now().Unix()-lastBackupUnix > 3600 {
				shouldRun = true
			}
		} else if config.Schedule == "daily_2am" {
			if now.Hour() >= 2 && now.Hour() <= 4 {
				lastDate := time.Unix(lastBackupUnix, 0).Format("2006-01-02")
				if lastDate != now.Format("2006-01-02") {
					shouldRun = true
				}
			}
		}

		if shouldRun && !IsSyncing {
			fmt.Printf("[SCHEDULER] Triggering ASYNC backup cycle (Schedule: %s, Force: %s)...\n", config.Schedule, force)
			
			currentPaths := []string{}
			for _, sel := range config.Paths {
				if realPath, ok := pathToRealMap[sel]; ok {
					currentPaths = append(currentPaths, realPath)
				} else {
					// Si no está en el mapa (quizás es una ruta manual antigua), dejarla como está
					currentPaths = append(currentPaths, sel)
				}
			}

			if len(currentPaths) == 0 && config.Status == "no_config" {
				currentPaths = backupPaths
			}
			if force == "full" {
				currentPaths = []string{"/host_root"}
			}

			fmt.Printf("[BACKUP] Targets Resolved: %v\n", currentPaths)

			// ASYNC BACKUP (V3.6.1)
			go func(paths []string, r, p, k, s string) {
				IsSyncing = true
				// Reportar inicio inmediato
				ReportHeartbeat(agentID, containerNames, explorerData, snapshots, true, os.Getpid(), lastBackupUnix)

				snapID, bytesProcessed, errB := RunResticBackup(paths, r, p, k, s)
				IsSyncing = false
				
				status := "SUCCESS"
				if errB != nil {
					status = "ERROR"
					fmt.Printf("[ASYNC ERROR] Backup failed: %v\n", errB)
				} else {
					lastBackupUnix = time.Now().Unix()
					fmt.Printf("[ASYNC] Backup finished successfully. ID: %s | Size: %d bytes\n", snapID, bytesProcessed)
				}

				// Reportar métricas reales al Control Plane
				ReportMetrics(BackupMetrics{
					AgentID:      agentID,
					Status:       status,
					Timestamp:    time.Now().Unix(),
					SnapshotID:   snapID,
					TotalSizeMB:  int(bytesProcessed / (1024 * 1024)),
					DurationSecs: 0,
				})
			}(currentPaths, repo, pass, key, secret)
		} else {
			if force == "none" || force == "" {
				fmt.Printf("[IDLE] Waiting for schedule (%s)... Last backup: %s\n", 
					config.Schedule, time.Unix(lastBackupUnix, 0).Format("15:04:05"))
			}
		}

		time.Sleep(60 * time.Second)
	}
}
