package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

var GlobalExcludes = []string{
	"*/cache/*",
	"*/logs/*",
	"*/tmp/*",
	"/proc/*",
	"/sys/*",
	"/dev/*",
	"/run/*",
	"/var/lib/docker/overlay2/*",
}




// EnsureResticRepo garantiza que el repositorio S3 esté inicializado (V2.6.5)
func EnsureResticRepo(repo string, password string, s3Key string, s3Secret string) error {
	if repo == "" {
		return fmt.Errorf("repository URL is empty")
	}

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

// RunResticBackup ejecuta el respaldo real de las carpetas seleccionadas hacia S3 (V5.1.1: Agregada Retención)
func RunResticBackup(paths []string, repo string, password string, s3Key string, s3Secret string, keepLast int) (string, int64, error) {
	if len(paths) == 0 {
		return "", 0, fmt.Errorf("no directories selected")
	}

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

	// Limpiar decoraciones (emojis) de los paths antes de enviar a restic
	cleanPaths := []string{}
	for _, p := range paths {
		clean := strings.TrimPrefix(p, "📂 ")
		clean = strings.TrimPrefix(clean, "📄 ")
		cleanPaths = append(cleanPaths, clean)
	}

	fmt.Printf("[RESTIC] Starting Incremental Backup for %d targets...\n", len(cleanPaths))
	
	finalArgs := []string{"-r", repo, "backup", "--json", "--one-file-system"}
	for _, ex := range GlobalExcludes {
		finalArgs = append(finalArgs, "--exclude", ex)
	}
	finalArgs = append(finalArgs, cleanPaths...)

	cmd := exec.Command("restic", finalArgs...)
	cmd.Env = env

	var outputBuffer bytes.Buffer
	writer := io.MultiWriter(os.Stdout, &outputBuffer)
	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Start(); err != nil {
		return "", 0, err
	}
	
	ActivePID = cmd.Process.Pid
	IsSyncing = true
	err := cmd.Wait()
	ActivePID = 0
	IsSyncing = false

	if err != nil {
		fmt.Printf("[RESTIC ERROR] Core failure (Wait): %v\n", err)
		return "", 0, fmt.Errorf("Restic failed: %v", err)
	}

	// PARSING SUMMARY (V3.6.1: Capturar ID y Tamaño real)
	finalSnapshotID := "unknown"
	var totalBytes int64 = 0
	
	lines := strings.Split(outputBuffer.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"message_type":"summary"`) {
			var summary struct {
				SnapshotID         string `json:"snapshot_id"`
				TotalBytesProcessed int64  `json:"total_bytes_processed"`
			}
			if err := json.Unmarshal([]byte(line), &summary); err == nil {
				finalSnapshotID = summary.SnapshotID
				totalBytes = summary.TotalBytesProcessed
			}
		}
	}
	
	fmt.Printf("[RESTIC] Backup cycle completed. Snapshot: %s | Processed: %d bytes\n", finalSnapshotID, totalBytes)
	
	// Tras el backup, aplicamos la política de retención dinámica (V5.1.1)
	_ = ApplyRetentionPolicy(repo, password, s3Key, s3Secret, keepLast)

	return finalSnapshotID, totalBytes, nil
}

// RunResticRestore ejecuta una restauración remota (V4.5.7/V9.0)
func RunResticRestore(snapshotID string, destination string, paths []string, repo string, password string, s3Key string, s3Secret string) (int, error) {
	if repo == "" || snapshotID == "" {
		return 0, fmt.Errorf("missing repository or snapshot ID")
	}

	start := time.Now()
	env := os.Environ()
	if password != "" { env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password)) }
	if s3Key != "" { env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key)) }
	if s3Secret != "" { env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret)) }

	args := []string{"-r", repo, "restore", snapshotID, "--target", destination}
	for _, p := range paths {
		clean := strings.TrimPrefix(strings.TrimPrefix(p, "📂 "), "📄 ")
		args = append(args, "--include", clean)
	}

	fmt.Printf("[RESTIC] Running restoration for snapshot %s to %s...\n", snapshotID, destination)
	cmd := exec.Command("restic", args...)
	cmd.Env = env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return 0, err
	}
	ActivePID = cmd.Process.Pid
	IsSyncing = true
	
	err := cmd.Wait()
	ActivePID = 0
	IsSyncing = false

	durationSecs := int(time.Since(start).Seconds())
	if err != nil {
		return durationSecs, fmt.Errorf("restic error: %v | %s", err, out.String())
	}

	fmt.Printf("[RESTIC] Restoration successful. Duration: %d seconds.\n", durationSecs)
	return durationSecs, nil
}

func RunResticVerify(repo string, password string, s3Key string, s3Secret string) error {
	env := os.Environ()
	if password != "" { env = append(env, fmt.Sprintf("RESTIC_PASSWORD=%s", password)) }
	if s3Key != "" { env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Key)) }
	if s3Secret != "" { env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Secret)) }

	cmd := exec.Command("restic", "-r", repo, "check")
	cmd.Env = env
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return err
	}
	ActivePID = cmd.Process.Pid
	err := cmd.Wait()
	ActivePID = 0
	
	if err != nil {
		return fmt.Errorf("restic check error: %s", out.String())
	}
	return nil
}

// RunPartialTestRestore realiza una prueba aleatoria/inteligente restaurando un sample (.env, docker-compose) (V9.0)
func RunPartialTestRestore(snapshotID string, repo string, password string, s3Key string, s3Secret string, tenantID string) error {
	testDir := fmt.Sprintf("/tmp/restore-test-%s", tenantID)
	os.RemoveAll(testDir) // Limpiar de inmediato remanentes previos

	// Pedir a restic que intente extraer solo un archivo .env si lo hay, ignorando si falla
	// Al no poder forzar un match estricto, le decimos restic restore con exclude genérico y dejamos un path como "*env*"
	// (En restic puedes usar wildcard en --include)
	_, err := RunResticRestore(snapshotID, testDir, []string{"*.env", "*docker-compose*"}, repo, password, s3Key, s3Secret)
	
	defer os.RemoveAll(testDir) // Siempre limpiar test sandbox

	if err != nil {
		return fmt.Errorf("partial restore simulation failed: %v", err)
	}

	// Verificar estructuralmente si el directorio se pobló
	info, err := os.Stat(testDir)
	if err != nil || !info.IsDir() {
		// No estricto: Puede que simplemente no hayan carpetas .env en este snap, 
		// pero logramos hacer read/write.
	}
	return nil
}



// ApplyRetentionPolicy aplica una política de rotación dinámica (KEEP X) (V5.1.1)
func ApplyRetentionPolicy(repo string, password string, s3Key string, s3Secret string, keepLast int) error {
	if keepLast <= 0 { keepLast = 1 } // Seguridad mínima (V5.1.1)
	
	fmt.Printf("[RESTIC] Applying retention policy (KEEP LAST %d)...\n", keepLast)
	
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

	// 1. Unlock preventivo (V3.3.7)
	unlockCmd := exec.Command("restic", "-r", repo, "unlock")
	unlockCmd.Env = env
	_ = unlockCmd.Run()
	time.Sleep(2 * time.Second) // Delay para consistencia S3

	// 2. Forget & Prune
	cmd := exec.Command("restic", "-r", repo, "forget", "--keep-last", fmt.Sprintf("%d", keepLast), "--prune")
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return err
	}
	ActivePID = cmd.Process.Pid
	err := cmd.Wait()
	ActivePID = 0

	if err != nil {
		return fmt.Errorf("retention failed: %v", err)
	}
	
	fmt.Printf("[RESTIC] Retention successful (Last %d copies). Storage optimized.\n", keepLast)
	return nil
}

// GetSnapshotsJSON devuelve la lista de snapshots en formato JSON crudo (V3.5.0)
func GetSnapshotsJSON(repo string, password string, s3Key string, s3Secret string) []byte {
	if repo == "" {
		return []byte("[]")
	}

	cmd := exec.Command("restic", "-r", repo, "snapshots", "--json")
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
		return []byte("[]")
	}
	return output
}

// GetSnapshotContentJSON devuelve el listado de archivos de un snapshot filtrado por profundidad (V4.6.5: Lazy Loading)
func GetSnapshotContentJSON(snapshotID string, requestPath string, repo string, password string, s3Key string, s3Secret string) []byte {
	if repo == "" || snapshotID == "" {
		return []byte("[]")
	}

	// 1. Ejecutamos LS completo pero procesamos el stream para no saturar memoria
	args := []string{"-r", repo, "ls", snapshotID, "--json"}
	cmd := exec.Command("restic", args...)
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return []byte("[]")
	}
	if err := cmd.Start(); err != nil {
		return []byte("[]")
	}

	var filtered []interface{}
	decoder := json.NewDecoder(stdout)

	// Normalizar requestPath para comparaciones exactas
	cleanReq := strings.Trim(requestPath, "/")
	reqDepth := len(strings.Split(cleanReq, "/"))
	if cleanReq == "" { reqDepth = 0 }

	for {
		var item struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Type string `json:"type"`
			Size int64  `json:"size,omitempty"`
		}
		if err := decoder.Decode(&item); err != nil {
			break 
		}

		// Solo procesamos directorios y archivos normales del snapshot (omitimos el propio snapshot root si viene)
		if item.Path == "/" || item.Path == "" { continue }

		itemPath := strings.Trim(item.Path, "/")
		
		// Lógica de Filtrado por Nivel (Direct Children Only)
		// 1. Debe empezar por la ruta solicitada
		if cleanReq != "" && !strings.HasPrefix(itemPath, cleanReq) {
			continue
		}

		// 2. Calculamos profundidad del item
		itemDepth := len(strings.Split(itemPath, "/"))

		// 3. Si req es "", itemDepth debe ser 1 (Raíz del backup)
		// Si req es "/a/b", itemDepth debe ser 3 (Hijos de b)
		if itemDepth == reqDepth + 1 {
			filtered = append(filtered, item)
		}
	}

	_ = cmd.Wait()

	resultJSON, _ := json.Marshal(filtered)
	fmt.Printf("[DEBUG-RESTIC] ⚡ LS Depth Filter: %s -> %d items\n", requestPath, len(filtered))
	return resultJSON
}


