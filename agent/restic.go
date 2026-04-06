package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var GlobalExcludes = []string{
	"*/cache/*",
	"*/logs/*",
	"*/sessions/*",
	"*/tmp/*",
	"node_modules",
	".git",
}


// EnsureResticRepo verifica si el repositorio S3 ya está inicializado. Si no, lo inicializa.
func EnsureResticRepo(repoURL string, password string, s3Key string, s3Secret string) error {
	fmt.Println("\n[RESTIC] Validating S3 Wasabi Repository...")
	
	repo := repoURL
	if repo == "" {
		repo = os.Getenv("RESTIC_REPOSITORY")
	}
	if repo == "" {
		return fmt.Errorf("RESTIC_REPOSITORY environment variable is missing")
	}

	// Inyectar credenciales (V2.6.2)
	env := os.Environ()
	if password != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password))
	}
	if s3Key != "" {
		env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key))
	}
	if s3Secret != "" {
		env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret))
	}


	// Ejecutar snapshot list para ver si el repo existe
	cmd := exec.Command("restic", "-r", repo, "snapshots", "--json")
	cmd.Env = env

	
	if err := cmd.Run(); err != nil {
		fmt.Println("[RESTIC] Repository not detected or uninitialized. Initializing now...")
		
		initCmd := exec.Command("restic", "-r", repo, "init")
		initCmd.Env = env
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
func RunResticBackup(paths []string, repoURL string, password string, s3Key string, s3Secret string) error {
	fmt.Println("\n[RESTIC] Starting Incremental Backup Engine...")
	if len(paths) == 0 {
		fmt.Println("[RESTIC] No directories selected for backup. Skipping cycle.")
		return nil
	}

	repo := repoURL
	if repo == "" {
		repo = os.Getenv("RESTIC_REPOSITORY")
	}

	// Inyectar credenciales (V2.6.2)
	env := os.Environ()
	if password != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password))
	}
	if s3Key != "" {
		env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key))
	}
	if s3Secret != "" {
		env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret))
	}


	// Preparar argumentos: restic backup --json --exclude=... /path1 /path2 ...
	args := []string{"-r", repo, "backup", "--json"}
	for _, ex := range GlobalExcludes {
		args = append(args, "--exclude", ex)
	}
	args = append(args, paths...)

	
	fmt.Printf("[RESTIC] Target paths: %v\n", paths)
	
	cmd := exec.Command("restic", args...)
	cmd.Env = env // Heredar variables S3 (Repository, Keys, Password)
	
	// Capturamos salida para depuración (V2.9.2)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Iniciamos el proceso y guardamos el PID para el Control Plane
	if err := cmd.Start(); err != nil {
		fmt.Printf("[RESTIC ERROR] Launch failed: %v\n", err)
		return err
	}
	
	ActivePID = cmd.Process.Pid
	IsSyncing = true

	err := cmd.Wait()
	ActivePID = 0
	IsSyncing = false

	if err != nil {
		fmt.Printf("[RESTIC ERROR] Core failure (Wait): %v\n", err)
		// Mostrar las últimas líneas de error para diagnóstico (V2.9.2)
		outStr := output.String()
		lines := strings.Split(outStr, "\n")
		lastLines := lines
		if len(lines) > 5 {
			lastLines = lines[len(lines)-6:]
		}
		fmt.Printf("[RESTIC OUTPUT] \n---\n%s\n---\n", strings.Join(lastLines, "\n"))
		return fmt.Errorf("Restic failed: %v", err)
	}


	
	fmt.Println("[RESTIC] Backup cycle completed successfully. Snapshot recorded.")
	
	// Tras el backup, aplicamos la política de retención automática
	_ = ApplyRetentionPolicy(repo, password, s3Key, s3Secret)


	return nil
}



// ApplyRetentionPolicy purga snapshots antiguos siguiendo la regla (7d, 4w, 2m)
func ApplyRetentionPolicy(repoURL string, password string, s3Key string, s3Secret string) error {
	fmt.Println("[RESTIC] Applying Retention Policy: 7 daily, 4 weekly, 2 monthly...")
	
	repo := repoURL
	if repo == "" {
		repo = os.Getenv("RESTIC_REPOSITORY")
	}

	args := []string{
		"-r", repo,
		"forget", 
		"--keep-daily", "7", 
		"--keep-weekly", "4", 
		"--keep-monthly", "2", 
		"--prune",
	}

	cmd := exec.Command("restic", args...)
	
	// Inyectar credenciales (V2.6.2)
	env := os.Environ()
	if password != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password))
	}
	if s3Key != "" {
		env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key))
	}
	if s3Secret != "" {
		env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret))
	}
	cmd.Env = env


	
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[RESTIC ERROR] Retention failed: %v\n", err)
		return err
	}
	
	fmt.Println("[RESTIC] Retention successful. Storage optimized.")
	fmt.Printf("[DEBUG] Restic output: %s\n", string(output))
	return nil
}


// GetSnapshotsJSON devuelve la lista de snapshots en formato JSON crudo
func GetSnapshotsJSON(repoURL string, password string, s3Key string, s3Secret string) []byte {
	repo := repoURL
	if repo == "" {
		repo = os.Getenv("RESTIC_REPOSITORY")
	}
	if repo == "" {
		return []byte("[]")
	}

	cmd := exec.Command("restic", "-r", repo, "snapshots", "--json")
	
	// Inyectar credenciales (V2.6.2)
	env := os.Environ()
	if password != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password))
	}
	if s3Key != "" {
		env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key))
	}
	if s3Secret != "" {
		env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("[RESTIC ERROR] Failed to list snapshots: %v\n", err)
		return []byte("[]")
	}
	return output
}



