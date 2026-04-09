<?php
/**
 * Docker Backup Pro - WHMCS Addon Module
 * 
 * Este módulo permite a los administradores de HWPeru ver el estado global 
 * de todos los backups de todos los clientes de forma centralizada.
 */

if (!defined("WHMCS")) {
    die("This file cannot be accessed directly");
}

function dockerbackuppro_config() {
    return [
        "name" => "Docker Backup Pro (HWPeru Admin)",
        "description" => "Portal Administrativo central para monitoreo de Agentes y Backups en Wasabi S3.",
        "author" => "HWPeru / Docker Backup Pro",
        "language" => "spanish",
        "version" => "1.0",
        "fields" => [
            "master_token" => [
                "FriendlyName" => "Master Admin Token",
                "Type" => "text",
                "Size" => "50",
                "Default" => "dbp_admin_hwperu_master_2024_secure_v1",
                "Description" => "Token maestro configurado en la API Central (api.hwperu.com).",
            ],
            "api_endpoint" => [
                "FriendlyName" => "API Endpoint",
                "Type" => "text",
                "Default" => "https://api.hwperu.com",
            ],
            "portal_url" => [
                "FriendlyName" => "Portal Dashboard URL",
                "Type" => "text",
                "Default" => "https://backup.hwperu.com",
            ],
            "debug_mode" => [
                "FriendlyName" => "Modo Diagnóstico (Debug)",
                "Type" => "yesno",
                "Description" => "Activa herramientas técnicas en el dashboard para detectar fallos de red o errores de token.",
            ]
        ]
    ];
}


function dockerbackuppro_activate() {
    return ["status" => "success", "description" => "Portal Administrativo activado correctamente."];
}

function dockerbackuppro_output($vars) {
    // Generamos el enlace al portal con acceso maestro
    $masterToken = $vars['master_token'];
    $portalUrl = $vars['portal_url'];
    $apiUrl = $vars['api_endpoint'];
    $debug = ($vars['debug_mode'] == 'on') ? '&debug=1' : '';
    
    echo '<div style="margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center;">
            <div class="alert alert-info" style="margin: 0; flex-grow: 1; margin-right: 20px;">
                <i class="fas fa-shield-alt"></i> Actualmente estás visualizando el <b>Panel Maestro</b> de HWPeru.
            </div>
            <button onclick="testWasabi()" class="btn btn-primary">
                <i class="fas fa-network-wired"></i> Test Wasabi Link
            </button>
          </div>';

    // Script para el test de Wasabi
    echo '<script>
        function testWasabi() {
            const token = "' . $masterToken . '";
            const url = "' . $apiUrl . '/v1/admin/wasabi/ping";
            
            alert("Iniciando prueba de latencia hacia Wasabi S3...");
            
            fetch(url, {
                headers: { "Authorization": token }
            })
            .then(r => r.json())
            .then(data => {
                if(data.status === "Online") {
                    alert("✅ [WASABI OK] Conexión Exitosa.\nLatencia: " + data.latency_ms + "ms\nBucket: " + data.bucket);
                } else {
                    alert("❌ [ERROR] " + data.message);
                }
            })
            .catch(e => alert("❌ [API DOWN] No se pudo contactar con la API Central."));
        }
    </script>';

    // El iframe carga el dashboard en modo admin con el token maestro y debug si aplica
    echo '<iframe src="' . $portalUrl . '?admin=1&sso=' . $masterToken . $debug . '" 
            width="100%" 
            height="900" 
            style="border:0; border-radius: 8px; box-shadow: 0 10px 30px rgba(0,0,0,0.5);" 
            frameborder="0"></iframe>';
}
