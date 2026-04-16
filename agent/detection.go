package main

import (
	"os"
	"path/filepath"
	"strings"
)

// StackInfo contiene la detección de tecnologías en el servidor
type StackInfo struct {
	HasDocker bool `json:"has_docker"`
	WordPress bool `json:"wordpress"`
	MySQL     bool `json:"mysql"`
	Nginx     bool `json:"nginx"`
	Apache    bool `json:"apache"`
	Node      bool `json:"node"`
	PM2       bool `json:"pm2"`
}

// DetectStack escanea el servidor buscando firmas de tecnologías comunes
func DetectStack() StackInfo {
	info := StackInfo{}

	// 1. Detección de Docker (Socket)
	if _, err := os.Stat("/host_root/var/run/docker.sock"); err == nil {
		info.HasDocker = true
	} else if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		info.HasDocker = true
	}

	// 2. Detección de WordPress
	// Escaneamos rutas comunes en el host
	wpPaths := []string{
		"/host_root/var/www/html",
		"/host_root/var/www",
		"/host_root/home",
	}

	for _, p := range wpPaths {
		found := false
		filepath.Walk(p, func(path string, osInfo os.FileInfo, err error) error {
			if err != nil { return err }
			if !osInfo.IsDir() && osInfo.Name() == "wp-config.php" {
				info.WordPress = true
				found = true
				return filepath.SkipDir
			}
			// Limitar profundidad para no matar el performance
			if osInfo.IsDir() && strings.Count(path, string(os.PathSeparator)) > 5 {
				return filepath.SkipDir
			}
			return nil
		})
		if found { break }
	}

	// 3. Detección de MySQL / MariaDB
	if _, err := os.Stat("/host_root/var/lib/mysql"); err == nil {
		info.MySQL = true
	}
	// También checar si existe el binario o socket
	if _, err := os.Stat("/host_root/var/run/mysqld/mysqld.sock"); err == nil {
		info.MySQL = true
	}

	// 4. Servidores Web
	if _, err := os.Stat("/host_root/etc/nginx"); err == nil {
		info.Nginx = true
	}
	if _, err := os.Stat("/host_root/etc/apache2"); err == nil || os.Stat("/host_root/etc/httpd") == nil {
		info.Apache = true
	}

	// 5. Node.js / PM2
	if _, err := os.Stat("/host_root/root/.pm2"); err == nil {
		info.PM2 = true
		info.Node = true
	}
	
	return info
}
