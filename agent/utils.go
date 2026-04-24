package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"strings"
	"github.com/google/uuid"
)




// GetPersistentID recupera el ID único del agente o genera uno nuevo (V3.7.1: Persistencia mejorada)
func GetPersistentID() string {
	idDir := "/etc/dbp"
	idFile := idDir + "/agent_id"
	
	// Asegurar que el directorio existe
	_ = os.MkdirAll(idDir, 0755)

	data, err := os.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	// Fallback para migración (si existía en la raíz)
	oldData, oldErr := os.ReadFile("/.agent_id")
	if oldErr == nil && len(oldData) > 0 {
		id := strings.TrimSpace(string(oldData))
		_ = os.WriteFile(idFile, []byte(id), 0644)
		return id
	}

	// Si no existe, generar uno nuevo
	newID := uuid.New().String()[:12]
	_ = os.WriteFile(idFile, []byte(newID), 0644)
	return newID
}


// GetRunningContainers obtiene la lista de nombres de contenedores activos
func GetRunningContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var names []string
	for _, l := range lines {
		if l != "" {
			names = append(names, l)
		}
	}
	return names, nil
}

// ScanVolumeFolders escanea las subcarpetas de un volumen de Docker (V3.8.3: Discovery Total)
func ScanVolumeFolders(path string) []string {
	var results []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return results
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") { continue }
		
		if e.IsDir() {
			results = append(results, "📂 "+name)
		} else {
			results = append(results, "📄 "+name)
		}
	}
	return results
}


// GetContainerMounts obtiene las rutas reales del host para los volúmenes de un contenedor (V3.8.1)
func GetContainerMounts(containerName string) []string {
	// Query: .Mounts contains all volume and bind info
	cmd := exec.Command("docker", "inspect", "--format", "{{range .Mounts}}{{.Source}} {{end}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	rawPaths := strings.Split(strings.TrimSpace(string(output)), " ")
	var hostPaths []string
	for _, p := range rawPaths {
		path := strings.TrimSpace(p)
		if path == "" { continue }
		
		// Filtrar rutas de sistema obvias
		if strings.Contains(path, "/docker.sock") || 
		   strings.Contains(path, "/etc/resolv.conf") ||
		   strings.Contains(path, "/etc/hostname") ||
		   strings.Contains(path, "/etc/hosts") {
			continue
		}
		
		hostPaths = append(hostPaths, path)
	}
	return hostPaths
}

// GetContainersForPaths: Identifica qué contenedores están usando las rutas que vamos a respaldar (V14.2.5)
func GetContainersForPaths(paths []string) []string {
	var targets []string
	
	// 1. Obtener todos los contenedores activos
	allContainers, err := GetRunningContainers()
	if err != nil {
		return targets
	}

	for _, container := range allContainers {
		mounts := GetContainerMounts(container)
		isRelevant := false
		
		for _, mount := range mounts {
			for _, p := range paths {
				// Normalizar paths: eliminar decoraciones y /host_root
				cleanP := strings.TrimPrefix(p, "📂 ")
				cleanP = strings.TrimPrefix(cleanP, "📄 ")
				cleanP = strings.TrimPrefix(cleanP, "/host_root")
				
				if cleanP == "" || cleanP == "/" { continue }

				// Si el path de backup está dentro de un mount, o el mount está dentro del path de backup
				if strings.HasPrefix(cleanP, mount) || strings.HasPrefix(mount, cleanP) {
					isRelevant = true
					break
				}
			}
			if isRelevant { break }
		}
		
		if isRelevant {
			targets = append(targets, container)
		}
	}
	
	return targets
}


// GenerateFingerprint: Crea la huella digital híbrida SHA256 (V13)
func GenerateFingerprint() string {
	machineID := getMachineID()
	diskID := getDiskID()
	hostname, _ := os.Hostname()

	// Fórmula: SHA256(machine-id + disk-id + hostname)
	raw := machineID + ":" + diskID + ":" + hostname
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func getMachineID() string {
	// Intentar leer desde /host_root/etc/machine-id (montaje recomendado en SaaS)
	data, err := os.ReadFile("/host_root/etc/machine-id")
	if err != nil {
		// Fallback local
		data, err = os.ReadFile("/etc/machine-id")
	}
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown_machine"
}

func getDiskID() string {
	// Prioridad: Serie física del disco sda (vía /sys)
	data, err := os.ReadFile("/host_root/sys/block/sda/device/serial")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	// Fallback: UUID de la partición raíz
	cmd := exec.Command("blkid", "-s", "UUID", "-o", "value", "/dev/sda1")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return "unknown_disk"
}
