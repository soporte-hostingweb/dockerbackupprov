package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	fmt.Println("[INFO] DBP Agent Booting...")

	// Inicializa el cliente Docker leyendo del entorno (ej. unix:///var/run/docker.sock en Linux o named pipe en Windows)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize Docker Client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	// Obtener lista de contenedores
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Fatalf("[ERROR] Failed to list containers: %v\n Asegúrate de tener Docker corriendo y permisos sobre el socket.", err)
	}

	fmt.Printf("[INFO] Discovered %d containers.\n", len(containers))

	var backupPaths []string

	for _, c := range containers {
		containerName := c.Names[0]
		fmt.Printf("\n[INFO] Inspecting Container: %s (ID: %s)\n", containerName, c.ID[:10])

		// Inspeccionar el contenedor a fondo para leer sus volúmenes y variables
		inspect, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			fmt.Printf("[WARN] Failed to inspect container %s: %v\n", containerName, err)
			continue
		}

		// Detectar Bases de Datos por el nombre de la imagen principal
		imageName := inspect.Config.Image
		isDB := false
		dbType := ""

		if containsString(imageName, "mysql") || containsString(imageName, "mariadb") {
			isDB = true
			dbType = "MySQL/MariaDB"
		} else if containsString(imageName, "postgres") {
			isDB = true
			dbType = "PostgreSQL"
		}

		if isDB {
			fmt.Printf("  -> [DETECTED] Database Instance: %s\n", dbType)
			fmt.Printf("  -> [ACTION] Queued for safe SQL Dump (docker exec)...\n")
			// En la versión funcional, aquí generamos el dump y agregamos el path del dump a backupPaths
			backupPaths = append(backupPaths, "/tmp/dbp_mysql_dump_"+c.ID[:10]+".sql")
		}

		// Listar los montajes (Bind mounts o Volumes) para respaldar archivos persistentes
		for _, mount := range inspect.Mounts {
			if mount.Type == "bind" || mount.Type == "volume" {
				fmt.Printf("  -> [MOUNT] Type: %s | Host Source: %s | Container Dest: %s\n", mount.Type, mount.Source, mount.Destination)
				backupPaths = append(backupPaths, mount.Source)
			}
		}
	}

	fmt.Println("\n[INFO] End of Discovery Phase.")

	// Ejecutar proceso de respaldo unificado de Restic
	err = RunResticBackup(backupPaths)
	finalStatus := "SUCCESS"
	if err != nil {
		finalStatus = "FAILED"
		fmt.Printf("[ERROR] Backup Engine Failed: %v\n", err)
	}

	// Enviar Telemetría y estado a la API Central
	metrics := BackupMetrics{
		AgentID:      "vps_token_dev",
		Status:       finalStatus,
		TotalSizeMB:  2450, // Ejemplo: 2.4GB obtenidos parseando el Output real de restic
		DurationSecs: 180,  // Ejemplo de tiempo
		SnapshotID:   "8f9b2a1a",
		Timestamp:    time.Now().Unix(),
	}
	ReportMetrics(metrics)

	fmt.Println("\n[INFO] DBP Agent Run Cycle Completed.")
}

// Función auxiliar simple para buscar subcadenas (similar a strings.Contains)
func containsString(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
