package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunDatabaseDump ejecuta un volcado de MySQL utilizando el cliente nativo del agente (V14.2.5 Hardening)
func RunDatabaseDump(config *AgentConfigV2) (string, error) {
	dumpPath := "/host_root/tmp/dbp_db_dump.sql"
	
	// 1. Caso A: Configuración Explícita
	if config != nil && config.DbEnabled {
		LogInfo("[DB-BACKUP] Using native client for SQL dump (Host: %s)...", config.DbHost)
		
		args := []string{"mysqldump"}
		
		// Conexión
		host := config.DbHost
		if host == "" || host == "localhost" {
			host = "172.17.0.1" // Apuntar al host de Docker por defecto si es local
		}
		
		args = append(args, "-h", host)
		if config.DbUser != "" { args = append(args, "-u", config.DbUser) }
		if config.DbPass != "" { args = append(args, fmt.Sprintf("-p%s", config.DbPass)) }
		
		// Opciones profesionales de consistencia
		args = append(args, "--single-transaction", "--quick", "--skip-lock-tables", "--no-tablespaces")
		
		// Bases de datos
		if len(config.DbNames) > 0 {
			args = append(args, "--databases")
			args = append(args, config.DbNames...)
		} else {
			args = append(args, "--all-databases")
		}
		
		args = append(args, "--result-file="+dumpPath)
		
		cmd := exec.Command(args[0], args[1:]...)
		err := cmd.Run()
		if err == nil {
			LogInfo("[DB-BACKUP] Native SQL dump SUCCESS.")
			return dumpPath, nil
		}
		
		// Fallback: Si el host falló pero tenemos Docker, intentamos el dump vía 'docker exec'
		LogInfo("[DB-BACKUP] Native dump failed, trying via Docker bridge discovery...")
		return TryDockerExecDump(config)
	}

	// 2. Caso B: Detección Automática (Legacy)
	LogInfo("[DB-BACKUP] Legacy detection: Trying Docker execution...")
	return TryDockerExecDump(config)
}

// TryDockerExecDump: Intenta encontrar un contenedor SQL y ejecutar el dump desde dentro
func TryDockerExecDump(config *AgentConfigV2) (string, error) {
	dumpPath := "/host_root/tmp/dbp_db_dump.sql"
	
	// Buscar contenedores que parezcan MySQL
	containers, _ := GetRunningContainers()
	var mysqlContainer string
	for _, c := range containers {
		low := strings.ToLower(c)
		if strings.Contains(low, "sql") || strings.Contains(low, "db") || strings.Contains(low, "maria") {
			mysqlContainer = c
			break
		}
	}

	if mysqlContainer != "" {
		LogInfo("[DB-BACKUP] Target found: %s. Executing internal dump...", mysqlContainer)
		
		user := "root"
		if config != nil && config.DbUser != "" { user = config.DbUser }
		
		passArg := ""
		if config != nil && config.DbPass != "" { passArg = "-p" + config.DbPass }

		// Ejecutar mysqldump dentro del contenedor y redirigir al host_root del agente
		// Comando: docker exec mysql_container mysqldump -u root -pPassword --all-databases
		cmdStr := fmt.Sprintf("docker exec %s mysqldump -u %s %s --all-databases --single-transaction --quick > %s", 
			mysqlContainer, user, passArg, dumpPath)
		
		cmd := exec.Command("/bin/sh", "-c", cmdStr)
		err := cmd.Run()
		if err == nil {
			LogInfo("[DB-BACKUP] Docker-exec dump SUCCESS.")
			return dumpPath, nil
		}
	}

	return "", fmt.Errorf("all database dump methods failed")
}

// CleanupDatabaseDump elimina el archivo temporal después del backup
func CleanupDatabaseDump() {
	dumpPath := "/host_root/tmp/dbp_db_dump.sql"
	if _, err := os.Stat(dumpPath); err == nil {
		os.Remove(dumpPath)
		LogInfo("[DB-BACKUP] Temporary dump file cleaned up.")
	}
}
