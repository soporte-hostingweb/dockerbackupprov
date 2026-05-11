//go:build windows

package main

import (
	"fmt"
)

// GlobalExcludes: Exclusiones OBLIGATORIAS para Windows
var GlobalExcludes = []string{
	"C:\\System Volume Information",
	"C:\\$Recycle.Bin",
	"C:\\Windows\\Temp",
	"C:\\Windows\\Prefetch",
	"*/AppData/Local/Temp/*",
	"*/cache/*",
	"*/__pycache__/*",
	"*/node_modules/.cache/*",
}

func CheckDockerEnvironment() bool {
	// En Windows, asumimos que no hay Docker gestionable vía CLI de la misma forma que en Linux
	// para el propósito de orquestación de restauraciones (por ahora).
	return false
}

func InstallDockerOnHost() error {
	return fmt.Errorf("automated docker installation not supported on Windows yet")
}

func ValidateRestoredServices(targetPath string, autoUp bool) string {
	return "Service validation not supported on Windows yet"
}
