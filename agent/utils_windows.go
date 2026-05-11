//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

// GetPersistentID recupera el ID único del agente en Windows
func GetPersistentID() string {
	idDir := os.Getenv("ProgramData") + "\\dbp"
	idFile := idDir + "\\agent_id"
	
	_ = os.MkdirAll(idDir, 0755)

	data, err := os.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	return generateNewID(idFile)
}

func getMachineID() string {
	// UUID de la BIOS/Motherboard vía WMIC
	cmd := exec.Command("wmic", "csproduct", "get", "uuid")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 1 {
			return strings.TrimSpace(lines[1])
		}
	}
	return "unknown_windows_machine"
}

func getDiskID() string {
	// Número de serie del disco físico vía WMIC
	cmd := exec.Command("wmic", "diskdrive", "get", "serialnumber")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 1 {
			return strings.TrimSpace(lines[1])
		}
	}
	return "unknown_windows_disk"
}

func GetDiskCapacity() (string, string) {
	// En Windows usamos PowerShell para obtener el espacio en C:
	cmd := exec.Command("powershell", "-Command", "Get-PSDrive C | Select-Object Free, Used | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown"
	}
	
	// Simplificando para este caso, extraemos los valores manualmente o vía regex si no queremos importar un parser de JSON
	// Pero mejor usamos un comando directo que devuelva solo los números
	cmd = exec.Command("powershell", "-Command", "(Get-PSDrive C).Free; (Get-PSDrive C).Used + (Get-PSDrive C).Free")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) >= 2 {
			freeBytes, _ := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
			totalBytes, _ := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
			
			freeGB := freeBytes / (1024 * 1024 * 1024)
			totalGB := totalBytes / (1024 * 1024 * 1024)
			
			return fmt.Sprintf("%.1fGB", freeGB), fmt.Sprintf("%.1fGB", totalGB)
		}
	}
	
	return "unknown", "unknown"
}

func GetHostRoot() string {
	return "" // En Windows no usamos prefijo /host_root
}
