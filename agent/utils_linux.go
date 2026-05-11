//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

	return generateNewID(idFile)
}

func getMachineID() string {
	// Intentar leer desde /host_root/etc/machine-id (montaje recomendado en SaaS)
	data, err := os.ReadFile("/host_root/etc/machine-id")
	if err != nil {
		data, err = os.ReadFile("/etc/machine-id")
	}
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown_linux_machine"
}

func getDiskID() string {
	data, err := os.ReadFile("/host_root/sys/block/sda/device/serial")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	cmd := exec.Command("blkid", "-s", "UUID", "-o", "value", "/dev/sda1")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return "unknown_linux_disk"
}

func GetDiskCapacity() (string, string) {
	cmd := exec.Command("df", "-k", "/host_root")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown"
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 { return "unknown", "unknown" }
	fields := strings.Fields(lines[1])
	if len(fields) < 4 { return "unknown", "unknown" }
	totalK, _ := strconv.ParseFloat(fields[1], 64)
	freeK, _ := strconv.ParseFloat(fields[3], 64)
	totalGB := totalK / (1024 * 1024)
	freeGB := freeK / (1024 * 1024)
	return fmt.Sprintf("%.1fGB", freeGB), fmt.Sprintf("%.1fGB", totalGB)
}

func GetHostRoot() string {
	return "/host_root"
}
