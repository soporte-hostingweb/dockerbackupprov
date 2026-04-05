package main

import (
	"fmt"
	"os"
	"os/exec"
)

// EnsureResticRepo verifica si el repositorio S3 ya está inicializado. Si no, lo inicializa.
func EnsureResticRepo() error {
	fmt.Println("\n[RESTIC] Validating S3 Wasabi Repository...")
	
	repo := os.Getenv("RESTIC_REPOSITORY")
	if repo == "" {
		return fmt.Errorf("RESTIC_REPOSITORY environment variable is missing")
	}

	// Ejecutar snapshot list para ver si el repo existe
	cmd := exec.Command("restic", "snapshots", "--json")
	cmd.Env = os.Environ()
	
	if err := cmd.Run(); err != nil {
		fmt.Println("[RESTIC] Repository not detected or uninitialized. Initializing now...")
		
		initCmd := exec.Command("restic", "init")
		initCmd.Env = os.Environ()
		output, initErr := initCmd.CombinedOutput()
		if initErr != nil {
			return fmt.Errorf("restic init failed: %v - Output: %s", initErr, string(output))
		}
		fmt.Println("[RESTIC] Repository successfully initialized on Wasabi S3.")
	} else {
		fmt.Println("[RESTIC] Wasabi S3 Repository is ready and accessible.")
	}
	return nil
}

// RunResticBackup ejecuta el respaldo real de las carpetas seleccionadas hacia S3
func RunResticBackup(paths []string) error {
	fmt.Println("\n[RESTIC] Starting Backup Engine...")
	if len(paths) == 0 {
		fmt.Println("[RESTIC] No directories selected for backup. Skipping cycle.")
		return nil
	}

	// Preparar argumentos: restic backup --json /path1 /path2 ...
	args := []string{"backup", "--json"}
	args = append(args, paths...)
	
	fmt.Printf("[RESTIC] Target paths: %v\n", paths)
	
	cmd := exec.Command("restic", args...)
	cmd.Env = os.Environ() // Heredar variables S3 (Repository, Keys, Password)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[RESTIC ERROR] Core failure: %v\n", err)
		return fmt.Errorf("Restic failed: %v - Output: %s", err, string(output))
	}
	
	fmt.Println("[RESTIC] Backup cycle completed successfully. Snapshot recorded.")
	return nil
}
