package main

import (
	"fmt"
	"os"
	"os/exec"
)

// RunDatabaseDump ejecuta un volcado de MySQL si se detecta en el sistema
func RunDatabaseDump() (string, error) {
	dumpPath := "/host_root/tmp/dbp_db_dump.sql"
	
	// 1. Detectar si MySQL está corriendo en el host ( bare-metal )
	// Intentamos ejecutar mysqldump vía chroot para usar el binario del host
	LogInfo("[DB-BACKUP] Attempting MySQL dump via chroot...")
	
	cmd := exec.Command("chroot", "/host_root", "mysqldump", "--all-databases", "--single-transaction", "--quick", "--result-file=/tmp/dbp_db_dump.sql")
	// Nota: El result-file dentro del chroot /tmp es /host_root/tmp/dbp_db_dump.sql fuera del chroot.
	
	output, err := cmd.CombinedOutput()
	if err == nil {
		LogInfo("[DB-BACKUP] Host MySQL dump SUCCESS.")
		return dumpPath, nil
	}

	LogInfo("[DB-BACKUP] Host mysqldump failed: %v. Output: %s", err, string(output))

	// 2. Fallback: Si falló el chroot, intentamos detectar si hay contenedores MySQL y dumpearlos
	// (En esta versión V14.2 simplificada, priorizamos el dump de host para WP bare-metal)
	
	return "", fmt.Errorf("no database dump possible on this host")
}

// CleanupDatabaseDump elimina el archivo temporal después del backup
func CleanupDatabaseDump() {
	dumpPath := "/host_root/tmp/dbp_db_dump.sql"
	if _, err := os.Stat(dumpPath); err == nil {
		os.Remove(dumpPath)
		LogInfo("[DB-BACKUP] Temporary dump file cleaned up.")
	}
}
