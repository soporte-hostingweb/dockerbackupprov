package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)




var IsSyncing bool = false
var ActivePID int = 0
var LastKnownSnapshots []interface{} // V6.6: Caché persistente para evitar parpadeos
var LastKnownContainers []string      // V6.7: Caché de contenedores para evitar parpadeos
var LastKnownExplorer map[string][]string // V6.8: Estado persistente del explorador

func main() {
	LogInfo("🚀 Docker Backup Pro Agent Starting...")
	
	// Cargar persistencia si existe (V2.4)
	agentID := GetPersistentID()
	var lastBackupUnix int64 = 0

	// Loop principal (V4.7.0: Zero-Latency Interactive)
	var lastRepoCheck time.Time
	var lastWizardActivity time.Time

	for {
		LogInfo("--- [CYCLE START] ---")

		// 1. Obtener Configuración Primaria (V3.4.2)
		config, errConfig := GetAgentConfig(agentID)
		if errConfig != nil {
			LogInfo("[ERROR] Could not fetch config: %v", errConfig)
			config = &AgentConfigV2{Status: "no_config"}
		}

		// 2. Resolver Repositorio y Credenciales
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

		// 2.5 Validar/Inicializar repositorio (Optimizado V4.7.0: Cada 5 min)
		if repo != "" && time.Since(lastRepoCheck) > 5*time.Minute {
			LogInfo("[RESTIC] Validating S3 Wasabi Repository...")
			if err := EnsureResticRepo(repo, pass, key, secret); err == nil {
				lastRepoCheck = time.Now()
			}
		}

		// 3. Reportar Heartbeat de INMEDIATO para recibir tareas (V4.6.8)
		maint, taskInfo, kill, errHeart := ReportHeartbeat(agentID, nil, nil, nil, IsSyncing, ActivePID, lastBackupUnix, "", "")
		if errHeart != nil {
			LogInfo("[WARNING] Heartbeat failed: %v", errHeart)
		}

		// 4. PRIORIDAD: Procesar Tareas Remotas (Wizard / Restore)
		if strings.HasPrefix(taskInfo, "ls_snapshot:") || strings.HasPrefix(taskInfo, "restore:") {
			lastWizardActivity = time.Now() // V4.7.1: Marcar actividad para entrar en Modo Turbo
			
			if strings.HasPrefix(taskInfo, "ls_snapshot:") {
				params := strings.TrimPrefix(taskInfo, "ls_snapshot:")
				parts := strings.Split(params, "|")
				snapID, path := parts[0], ""
				if len(parts) > 1 { path = parts[1] }
				LogInfo("[TASK] Executing remote directory listing for snapshot %s (Path: %s)", snapID, path)
				lsResult := GetSnapshotContentJSON(snapID, path, repo, pass, key, secret)
				ReportTaskResult(agentID, "ls_snapshot", string(lsResult))
				LogInfo("[TASK] Snapshot listing completed and reported.")
			}
			if strings.HasPrefix(taskInfo, "restore:") {
				params := strings.TrimPrefix(taskInfo, "restore:")
				parts := strings.Split(params, "|")
				if len(parts) >= 2 {
					snapID, dest := parts[0], parts[1]
					var paths []string
					if len(parts) > 2 && parts[2] != "" { paths = strings.Split(parts[2], ",") }
					LogInfo("[TASK] Executing remote restoration of %s to %s", snapID, dest)
					errR := RunResticRestore(snapID, dest, paths, repo, pass, key, secret)
					resMsg := "Success"
					if errR != nil { resMsg = "Error: " + errR.Error() }
					ReportTaskResult(agentID, "restore", resMsg)
					lastWizardActivity = time.Time{} // Resetear turbo tras restore
				}
			}
			continue
		}

		// 5. Descubrimiento y Mantenimiento si no hay tareas críticas
		if maint {
			LogInfo("[INFO] Maintenance Mode Active. Pausing cycle.")
			time.Sleep(10 * time.Second)
			continue
		}

		// V4.7.1: MODO TURBO - Si hay actividad reciente del Wizard, omitimos el pesado proceso de discovery
		if time.Since(lastWizardActivity) < 30*time.Second {
			LogInfo("[TURBO] Skipping heavy discovery (Wizard Active). Using persistent state.")
			free, total := GetDiskCapacity()
			// V6.8: Enviamos el último estado conocido en lugar de 'nil' para proteger la vista
			_, taskInfo, _, _ = ReportHeartbeat(agentID, LastKnownContainers, LastKnownExplorer, LastKnownSnapshots, IsSyncing, ActivePID, lastBackupUnix, free, total)
			if taskInfo != "" && taskInfo != "none" { continue }
			
			// V4.7.2: Super Turbo (Reducción de 5s a 2s de espera interactiva)
			time.Sleep(2 * time.Second) 
			continue
		}

		// Descubrimiento de contenedores (Normal) con Estado Permanente V6.8
		containerNames, errC := GetRunningContainers()
		if errC == nil && len(containerNames) > 0 {
			LastKnownContainers = containerNames
			LogInfo("[INFO] Discovered %d containers.", len(containerNames))
		}

		// Preparar explorer data de forma atómica
		currentExplorer := make(map[string][]string)
		pathToRealMap := make(map[string]string) 
		var backupPaths []string

		for _, name := range LastKnownContainers {
			if name == "dbp-client-agent" { continue }
			hostMounts := GetContainerMounts(name)
			for _, hostPath := range hostMounts {
				bridgePath := "/host_root" + hostPath
				if info, err := os.Stat(bridgePath); err == nil && info.IsDir() {
					backupPaths = append(backupPaths, bridgePath)
					subItems := ScanVolumeFolders(bridgePath)
					for _, item := range subItems {
						currentExplorer[name] = append(currentExplorer[name], item)
						itemName := strings.TrimPrefix(strings.TrimPrefix(item, "📂 "), "📄 ")
						pathToRealMap[item] = bridgePath + "/" + itemName
					}
					fullVolEntry := "📂 [Full Volume] " + name
					currentExplorer[name] = append(currentExplorer[name], fullVolEntry)
					pathToRealMap[fullVolEntry] = bridgePath
				}
			}
		}
		if len(currentExplorer) > 0 { LastKnownExplorer = currentExplorer }

		// Reportar Heartbeat Completo con Estado Persistente
		if repo != "" && pass != "" {
			snapshotsRaw := GetSnapshotsJSON(repo, pass, key, secret)
			var currentSnapshots []interface{}
			if errS := json.Unmarshal(snapshotsRaw, &currentSnapshots); errS == nil && len(currentSnapshots) > 0 {
				LastKnownSnapshots = currentSnapshots
				LogInfo("[SNAPSHOTS] Detected %d snapshots.", len(LastKnownSnapshots))
			}
		}

		free, total := GetDiskCapacity()
		_, taskInfo, _, _ = ReportHeartbeat(agentID, LastKnownContainers, LastKnownExplorer, LastKnownSnapshots, IsSyncing, ActivePID, lastBackupUnix, free, total)

		// 6. Scheduler (Backup Logic)
		force := "none"
		if !strings.Contains(taskInfo, ":") { force = taskInfo }

		if kill && IsSyncing && ActivePID > 0 {
			LogInfo("[KILL] Terminating PID %d...", ActivePID)
			if proc, errP := os.FindProcess(ActivePID); errP == nil {
				proc.Signal(os.Interrupt)
				time.Sleep(2 * time.Second)
				proc.Kill()
			}
			IsSyncing = false
		}

		shouldRun := false
		if force != "none" && force != "" {
			shouldRun = true
		} else if config.Schedule != "manual" && config.Schedule != "" {
			// V7.1: Soporte para Zonas Horarias Regionales
			tz := config.TimeZone
			if tz == "" { tz = "America/Lima" }
			loc, errLoc := time.LoadLocation(tz)
			if errLoc != nil { loc = time.FixedZone("UTC-5", -5*60*60) } // Fallback a Perú si falla la carga

			now := time.Now().In(loc)
			h := now.Hour()
			d := int(now.Weekday())
			if d == 0 { d = 7 } // Normalizar Domingo a 7

			// V7.2: Lógica de Programación por Planes
			isScheduledTime := false
			if config.Schedule == "daily_2am" || config.Schedule == "daily_2am_basic" {
				if h == 2 && now.Minute() < 10 { isScheduledTime = true }
			} else if config.Schedule == "weekly_2am" {
				if d == 7 && h == 2 && now.Minute() < 10 { isScheduledTime = true }
			} else if config.Schedule == "custom" && config.CustomSchedule != "" {
				// Formato: "1,3,5|14" (Dias|Hora)
				parts := strings.Split(config.CustomSchedule, "|")
				if len(parts) == 2 {
					confDays := strings.Split(parts[0], ",")
					confHour := 0
					fmt.Sscanf(parts[1], "%d", &confHour)
					if confHour == 24 { confHour = 0 } // Normalizar 24h a 0h

					dayMatch := false
					for _, cd := range confDays {
						if cd == fmt.Sprintf("%d", d) { dayMatch = true; break }
					}

					if dayMatch && h == confHour && now.Minute() < 10 {
						isScheduledTime = true
					}
				}
			}

			if isScheduledTime {
				if time.Unix(lastBackupUnix, 0).Format("2006-01-02") != now.Format("2006-01-02") {
					shouldRun = true
					LogInfo("[SCHEDULER] Auto-trigger detected (TZ: %s, Plan: %s)", tz, config.Schedule)
				}
			}
		}

		if shouldRun && !IsSyncing {
			LogInfo("[SCHEDULER] Triggering backup...")
			currentPaths := []string{}
			for _, sel := range config.Paths {
				if realPath, ok := pathToRealMap[sel]; ok {
					currentPaths = append(currentPaths, realPath)
				} else {
					currentPaths = append(currentPaths, sel)
				}
			}
			if len(currentPaths) == 0 && config.Status == "no_config" { currentPaths = backupPaths }
			if force == "full" { currentPaths = []string{"/host_root"} }

			go func(paths []string, r, p, k, s string) {
				startedAt := time.Now().Unix()
				IsSyncing = true
				
				// V6.3: Reportar Inicio al Monitor Global
				activityID := ReportActivity(0, agentID, "backup", "running", "Scheduled backup started")
				
				f, t := GetDiskCapacity()
				ReportHeartbeat(agentID, LastKnownContainers, LastKnownExplorer, LastKnownSnapshots, true, ActivePID, lastBackupUnix, f, t)
				
				snapID, bytesProcessed, errB := RunResticBackup(paths, r, p, k, s, config.Retention)
				
				finishedAt := time.Now().Unix()
				duration := int(finishedAt - startedAt)
				status := "success"
				msg := fmt.Sprintf("Backup completed. Size: %d bytes", bytesProcessed)
				
				if errB != nil {
					status = "error"
					msg = fmt.Sprintf("Backup failed: %v", errB)
					LogInfo("[ASYNC ERROR] %v", errB)
				} else {
					lastBackupUnix = finishedAt
					LogInfo("[ASYNC] Finished. ID: %s | Dur: %ds", snapID, duration)
				}

				// V6.3: Reportar Fin al Monitor Global
				ReportActivity(activityID, agentID, "backup", status, msg)

				ReportMetrics(BackupMetrics{
					AgentID: agentID, Status: strings.ToUpper(status), StartedAt: startedAt, Timestamp: finishedAt,
					SnapshotID: snapID, TotalSizeMB: int(bytesProcessed / (1024 * 1024)),
					TotalSizeBytes: bytesProcessed, DurationSecs: duration,
				})

				// Actualizar snapshots tras el backup
				rawSnaps := GetSnapshotsJSON(r, p, k, s)
				var updatedSnapshots []interface{}
				json.Unmarshal(rawSnaps, &updatedSnapshots)
				if len(updatedSnapshots) > 0 { LastKnownSnapshots = updatedSnapshots }
				ReportHeartbeat(agentID, LastKnownContainers, LastKnownExplorer, LastKnownSnapshots, false, ActivePID, lastBackupUnix, f, t)

				IsSyncing = false
			}(currentPaths, repo, pass, key, secret)
		} else {
			if force == "none" || force == "" {
				LogInfo("[IDLE] Waiting schedule (%s). Last: %s", config.Schedule, time.Unix(lastBackupUnix, 0).Format("15:04:05"))
			}
		}

		// PROCESAMIENTO DE TAREAS REMOTAS (Restore, etc.)
		if taskInfo != "" && taskInfo != "none" {
			parts := strings.Split(taskInfo, "|")
			taskType := parts[0]

			if taskType == "restore" && len(parts) >= 3 {
				snapID := parts[1]
				target := parts[2]
				
				// V6.3: Reportar Inicio de Restauración
				LogInfo("[AUDIT] Reporting restoration start...")
				activityID := ReportActivity(0, agentID, "restore", "running", fmt.Sprintf("Restoring snapshot %s to %s", snapID, target))
				
				var restorePaths []string
				if len(parts) > 3 {
					restorePaths = strings.Split(parts[3], ",")
				}

				go func() {
					errR := RunResticRestore(snapID, target, restorePaths, repo, pass, key, secret)
					
					status := "success"
					msg := "Restoration completed successfully"
					if errR != nil {
						status = "error"
						msg = fmt.Sprintf("Restoration failed: %v", errR)
					}
					
					// V6.3: Reportar Fin de Restauración
					LogInfo("[AUDIT] Reporting restoration final status: %s", status)
					ReportActivity(activityID, agentID, "restore", status, msg)
					
					ReportTaskResult(agentID, taskType, msg)
				}()
			}
			continue
		}

		// REDUCCIÓN CRÍTICA DE LATENCIA (V4.6.8: de 60s a 10s)
		time.Sleep(10 * time.Second)
	}
}

// LogInfo helper para añadir timestamps (V4.6.8)
func LogInfo(format string, a ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("[%s] %s\n", timestamp, msg)
}

// GetDiskCapacity obtiene el espacio libre y total del host (/host_root) (V4.5.5)
func GetDiskCapacity() (string, string) {
	cmd := exec.Command("df", "-k", "/host_root")
	output, err := cmd.Output()
	if err != nil {
		LogInfo("[DISK ERROR] %v", err)
		return "unknown", "unknown"
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 { return "unknown", "unknown" }
	fields := strings.Fields(lines[1])
	if len(fields) < 4 { return "unknown", "unknown" }
	totalK, _ := strconv.ParseFloat(fields[1], 64)
	freeK, _ := strconv.ParseFloat(fields[3], 64)
	totalGB := totalK / (1024 * 1024)
	freeGB := freeK / (1024 * 1024)
	return fmt.Sprintf("%.1fGB", freeGB), fmt.Sprintf("%.1fGB", totalGB)
}

