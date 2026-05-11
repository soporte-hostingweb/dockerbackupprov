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
	hostRoot := GetHostRoot()

	// 1. Detección de Docker (Socket)
	dockerSocket := "/var/run/docker.sock"
	if hostRoot != "" {
		if _, err := os.Stat(hostRoot + dockerSocket); err == nil {
			info.HasDocker = true
		}
	}
	if !info.HasDocker {
		if _, err := os.Stat(dockerSocket); err == nil {
			info.HasDocker = true
		}
	}

	// 2. Detección de WordPress
	wpPaths := []string{
		hostRoot + "/var/www/html",
		hostRoot + "/var/www",
		hostRoot + "/home",
	}
	if hostRoot == "" { // Windows fallback
		wpPaths = []string{"C:\\inetpub\\wwwroot", "C:\\inetpub"}
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
			if osInfo.IsDir() && strings.Count(path, string(os.PathSeparator)) > 5 {
				return filepath.SkipDir
			}
			return nil
		})
		if found { break }
	}

	// 3. Detección de MySQL / MariaDB
	if _, err := os.Stat(hostRoot + "/var/lib/mysql"); err == nil {
		info.MySQL = true
	}
	if _, err := os.Stat(hostRoot + "/var/run/mysqld/mysqld.sock"); err == nil {
		info.MySQL = true
	}

	// 4. Servidores Web
	if _, err := os.Stat(hostRoot + "/etc/nginx"); err == nil {
		info.Nginx = true
	}
	if _, err := os.Stat(hostRoot + "/etc/apache2"); err == nil {
		info.Apache = true
	} else if _, err := os.Stat(hostRoot + "/etc/httpd"); err == nil {
		info.Apache = true
	}

	// 5. Node.js / PM2
	if _, err := os.Stat(hostRoot + "/root/.pm2"); err == nil {
		info.PM2 = true
		info.Node = true
	}
	
	return info
}
