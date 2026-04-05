package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	fmt.Println("[INFO] DBP Agent Booting...")

	// 0. Asegurar Repositorio S3 (Wasabi)
	if err := EnsureResticRepo(); err != nil {
		log.Printf("[WARNING] Restic Repo check failed: %v. Backups might fail until S3 is ready.", err)
	}

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
			if len(c.Names) > 0 {
				name := strings.TrimPrefix(c.Names[0], "/")
				containerNames = append(containerNames, name)
			}

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
					backupPaths = append(backupPaths, mount.Source)
					// Escaneamos carpetas para el Explorer del UI
					explorerData[c.ID[:10]] = append(explorerData[c.ID[:10]], ScanVolumeFolders(mount.Source)...)
				}
			}
		}

		// Determinar ID del Agente
		agentID := os.Getenv("DBP_AGENT_ID")
		if agentID == "" {
			hostname, _ := os.Hostname()
			agentID = hostname
		}

		// 1. Reportar Heartbeat (Lista de contenedores en vivo + Explorer Data) - ¡PRIMERO!
		ReportHeartbeat(agentID, containerNames, explorerData)

		// 2. Asegurar Repo e Inicializar Respaldo
		if err := EnsureResticRepo(); err != nil {
			fmt.Printf("[ERROR] S3 Repo check failed: %v. Skipping backup this cycle.\n", err)
			time.Sleep(60 * time.Second)
			continue
		}

		err = RunResticBackup(backupPaths)
		finalStatus := "SUCCESS"
		if err != nil {
			finalStatus = "FAILED"
		}

		// 3. Enviar Telemetría de Respaldo
		metrics := BackupMetrics{
			AgentID:      agentID,
			Status:       finalStatus,
			TotalSizeMB:  2450, 
			DurationSecs: 180,  
			SnapshotID:   "8f9b2a1a",
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
	var folders []string
	files, err := os.ReadDir(path)
	if err != nil {
		return folders
	}
	for _, f := range files {
		if f.IsDir() {
			folders = append(folders, f.Name())
		}
	}
	return folders
}
