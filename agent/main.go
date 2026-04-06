package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

var (
	IsSyncing bool
	ActivePID int
)


func main() {
	// CONFIGURACIÓN HORARIA: Operamos en UTC por defecto para evitar discrepancias SaaS
	time.Local = time.UTC
	fmt.Println("[INFO] Local time synchronized to UTC.")
	ActivePID := 0
	IsSyncing := false
	var lastBackupUnix int64 = 0

	godotenv.Load()
	fmt.Println("[INFO] DBP Agent Booting...")

	// Inicializa el cliente Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize Docker Client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	// Ciclo Infinito de Escaneo y Reporte (Cada 1 Minuto)
	fmt.Println("[INFO] Agent entering long-running heartbeat mode...")
	for {
		fmt.Printf("\n--- [CYCLE START: %s] ---\n", time.Now().Format(time.RFC1123))
		
		// Obtener lista de contenedores
		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			fmt.Printf("[ERROR] Failed to list containers: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Printf("[INFO] Discovered %d containers.\n", len(containers))
		var backupPaths []string
		var containerNames []string
		explorerData := make(map[string][]string)

		for _, c := range containers {
			if len(c.Names) == 0 {
				continue
			}
			name := c.Names[0][1:]

			// EXCLUIMOS EL PROPIO AGENTE
			if name == "dbp-client-agent" {
				continue
			}

			containerNames = append(containerNames, name)

			inspect, err := cli.ContainerInspect(ctx, c.ID)
			if err != nil {
				continue
			}

			// Lógica de descubrimiento de Bases de Datos
			imageName := inspect.Config.Image
			if containsString(imageName, "mysql") || containsString(imageName, "mariadb") || containsString(imageName, "postgres") {
				dumpPath := "/tmp/dbp_dump_" + c.ID[:10] + ".sql"
				backupPaths = append(backupPaths, dumpPath)
			}

			for _, mount := range inspect.Mounts {
				if mount.Type == "bind" || mount.Type == "volume" {
					hostPath := "/host_root" + mount.Source
					
					info, err := os.Stat(hostPath)
					if err != nil || !info.IsDir() {
						continue 
					}

					backupPaths = append(backupPaths, hostPath)
					
					subfolders := ScanVolumeFolders(hostPath)
					explorerData[name] = append(explorerData[name], subfolders...)
				}
			}
		}

		// Determinar ID del Agente de forma robústa
		agentID := os.Getenv("DBP_AGENT_ID")
		if agentID == "" {
			hostname, err := os.Hostname()
			if err != nil || hostname == "" {
				// Fallback definitivo si el sistema no tiene hostname (raro en Linux/Docker)
				agentID = "vps-agent-" + time.Now().Format("20060102150405")
			} else {
				agentID = hostname
			}
		}


		// 1. Reportar Heartbeat (Lista de contenedores + Explorer Data + Snapshots)
		// Usamos el repo base para el primer reporte de snapshots
		baseRepo := os.Getenv("RESTIC_REPOSITORY") 
		
		fmt.Printf("[HEARTBEAT] Reporting status for agent %s (%s) to Control Plane...\n", agentID, runtime.GOOS)
		snapRaw := GetSnapshotsJSON(baseRepo)
		var snapshots []interface{}
		json.Unmarshal(snapRaw, &snapshots)

		maint, force, kill, err := ReportHeartbeat(agentID, containerNames, explorerData, snapshots, IsSyncing, ActivePID, lastBackupUnix)



		
		if kill && IsSyncing && ActivePID > 0 {
			fmt.Printf("[KILL] Remote terminate requested for PID %d...\n", ActivePID)
			proc, err := os.FindProcess(ActivePID)
			if err == nil {
				proc.Signal(os.Interrupt) // Intento de cierre limpio
				time.Sleep(2 * time.Second)
				proc.Kill() // Si no cerró, matamos
			}
			IsSyncing = false
			ActivePID = 0
		}

		if maint {
			fmt.Println("[INFO] Maintenance Mode Active. Pausing backup cycles.")
			time.Sleep(60 * time.Second)
			continue
		}


		// 2. Obtener la SELECCIÓN del cliente desde la API
		config, err := GetAgentConfig(agentID)
		if err != nil {
			fmt.Printf("[API ERROR] Could not fetch backup config: %v\n", err)
			time.Sleep(60 * time.Second)
			continue
		}

		// isolation logic (V2.3)
		var isolatedBucket string
		if config.FullRepoURL != "" {
			isolatedBucket = config.FullRepoURL
			fmt.Printf("[ISOLATION] Targeting unique repository: %s\n", isolatedBucket)
		} else {
			isolatedBucket = os.Getenv("RESTIC_REPOSITORY")
		}
		
		// 3. Asegurar que el Repositorio esté inicializado antes de seguir (V2.3)
		if isolatedBucket != "" {
			if err := EnsureResticRepo(isolatedBucket); err != nil {
				fmt.Printf("[WARNING] Restic Repo ensure failed: %v\n", err)
			}
		}



		selectedPaths := config.Paths

		
		// LÓGICA DE AUTO-INTEGRACIÓN (V2.2 FALLBACK)
		if len(selectedPaths) == 0 && config.Status == "no_config" {
			if len(backupPaths) > 0 {
				fmt.Printf("[PROACTIVE] No manual config found. Using AUTO-DISCOVERY mode (SQL Dumps: %d target paths)\n", len(backupPaths))
				selectedPaths = backupPaths
			}
		} 



		// LÓGICA DE PROGRAMACIÓN (SCHEDULER V2.3)
		shouldRun := false
		now := time.Now()

		if force != "none" {
			shouldRun = true
		} else if config.Schedule == "every_1h" {
			if time.Now().Unix() - lastBackupUnix > 3600 {
				shouldRun = true
			}
		} else if config.Schedule == "daily_2am" {
			// Ventana de mantenimiento entre las 2 y las 4 AM UTC
			if now.Hour() >= 2 && now.Hour() <= 4 {
				lastDate := time.Unix(lastBackupUnix, 0).Format("2006-01-02")
				currentDate := now.Format("2006-01-02")
				if lastDate != currentDate {
					shouldRun = true
				}
			}
		}

		if shouldRun {
			fmt.Printf("[SCHEDULER] Triggering backup cycle (Schedule: %s, Force: %s)...\n", config.Schedule, force)
			currentPaths := selectedPaths
			if force == "full" {
				currentPaths = []string{"/host_root"}
			}
			
			IsSyncing = true
			// Pasamos el bucket aislado específicamente para este comando
			err = RunResticBackup(currentPaths, isolatedBucket)
			IsSyncing = false

			
			if err == nil {
				lastBackupUnix = time.Now().Unix()
			}
		} else {
			if force == "none" {
				fmt.Printf("[IDLE] Waiting for schedule (%s)... Last backup: %s\n", 
					config.Schedule, time.Unix(lastBackupUnix, 0).Format("15:04:05"))
			}
		}



		finalStatus := "SUCCESS"
		if err != nil {
			finalStatus = "FAILED"
		}

		// 4. Enviar Telemetría de Respaldo
		metrics := BackupMetrics{
			AgentID:      agentID,
			Status:       finalStatus,
			TotalSizeMB:  0, // TODO: Calcular desde restic output
			DurationSecs: 0,
			SnapshotID:   "auto",
			Timestamp:    time.Now().Unix(),
		}
		ReportMetrics(metrics)

		fmt.Println("[INFO] Cycle completed. Sleeping for 60 seconds...")
		time.Sleep(60 * time.Second)
	}
}

func containsString(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func ScanVolumeFolders(path string) []string {
	var items []string
	files, err := os.ReadDir(path)
	if err != nil {
		return items
	}
	for _, f := range files {
		fullPath := path
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		
		prefix := "📄 "
		if f.IsDir() {
			prefix = "📂 "
		}
		items = append(items, prefix+fullPath+f.Name())
	}
	return items
}
