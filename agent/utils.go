package main

import (
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

// ScanVolumeFolders escanea las subcarpetas de un volumen de Docker (V2.9)
func ScanVolumeFolders(path string) []string {
	var folders []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return folders
	}

	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			folders = append(folders, e.Name())
		}
	}
	return folders
}

