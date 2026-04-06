package main

import (
	"os/exec"
	"strings"
	"io/ioutil"
	"github.com/google/uuid"
)



// GetPersistentID recupera el ID único del agente o genera uno nuevo (V2.4)
func GetPersistentID() string {
	idFile := "/.agent_id"
	data, err := ioutil.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	// Si no existe, generar uno nuevo
	newID := uuid.New().String()[:12] // Usamos 12 caracteres para brevedad
	_ = ioutil.WriteFile(idFile, []byte(newID), 0644)
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
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return folders
	}

	for _, f := range files {
		if f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			folders = append(folders, f.Name())
		}
	}
	return folders
}
