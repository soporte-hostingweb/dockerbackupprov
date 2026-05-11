//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GlobalExcludes: Exclusiones OBLIGATORIAS para Linux
var GlobalExcludes = []string{
	"/host_root/proc",
	"/host_root/sys",
	"/host_root/dev",
	"/host_root/run",
	"/host_root/tmp",
	"/host_root/var/tmp",
	"/host_root/var/lib/docker/overlay2",
	"/host_root/var/lib/docker/aufs",
	"*/cache/*",
	"*/__pycache__/*",
	"*/node_modules/.cache/*",
}

func CheckDockerEnvironment() bool {
	fmt.Println("[PREP] Checking Docker availability on Host...")
	cmd := exec.Command("chroot", "/host_root", "docker", "--version")
	if err := cmd.Run(); err != nil {
		fmt.Println("[PREP] Docker NOT found or not responding on Host.")
		return false
	}
	fmt.Println("[PREP] Docker is available and ready on Host.")
	return true
}

func InstallDockerOnHost() error {
	fmt.Println("[PREP] Starting automated Docker installation on Host...")
	installCmd := "curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh"
	cmd := exec.Command("chroot", "/host_root", "sh", "-c", installCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker installation failed: %v | Output: %s", err, string(output))
	}
	_ = exec.Command("chroot", "/host_root", "systemctl", "enable", "--now", "docker").Run()
	fmt.Println("[PREP] Docker successfully installed on Host.")
	return nil
}

func ValidateRestoredServices(targetPath string, autoUp bool) string {
	if !autoUp { return "Service validation skipped (autoUp=false)" }
	fmt.Printf("[SANDBOX] Starting isolated validation in %s...\n", targetPath)
	exec.Command("chroot", "/host_root", "docker", "network", "create", "--internal", "dbp_sandbox_net").Run()
	hostPath := strings.TrimPrefix(targetPath, "/host_root")
	if hostPath == "" { hostPath = "/" }
	findCmd := exec.Command("chroot", "/host_root", "find", hostPath, "-name", "docker-compose.yml")
	output, _ := findCmd.Output()
	composeFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(composeFiles) == 0 || composeFiles[0] == "" { return "No docker-compose.yml found in restored path." }
	results := []string{}
	for _, cf := range composeFiles {
		dir := strings.TrimSuffix(cf, "/docker-compose.yml")
		sandboxCmd := fmt.Sprintf("cd %s && docker-compose -p dbp_sandbox up -d", dir)
		upCmd := exec.Command("chroot", "/host_root", "sh", "-c", sandboxCmd)
		if err := upCmd.Run(); err != nil {
			sandboxCmd = fmt.Sprintf("cd %s && docker compose -p dbp_sandbox up -d", dir)
			upCmd = exec.Command("chroot", "/host_root", "sh", "-c", sandboxCmd)
			if err2 := upCmd.Run(); err2 != nil {
				results = append(results, fmt.Sprintf("Sandbox Failed: %v", err2))
				continue
			}
		}
		checkCmd := exec.Command("chroot", "/host_root", "docker", "ps", "--filter", "name=dbp_sandbox", "--format", "{{.Status}}")
		status, _ := checkCmd.Output()
		results = append(results, fmt.Sprintf("Sandbox UP [%s]: %s", strings.TrimSpace(string(status)), cf))
	}
	return strings.Join(results, " | ")
}
