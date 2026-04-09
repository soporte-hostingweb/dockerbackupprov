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
        "version" => "11.2",
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
            ],
            "n8n_webhook" => [
                "FriendlyName" => "N8N Webhook URL",
                "Type" => "text",
                "Size" => "100",
                "Description" => "URL global para recibir alertas (n8n). V11.2+",
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
            <div style="display: flex; gap: 10px;">
                <button onclick="syncSaaSPlan()" class="btn btn-warning">
                    <i class="fas fa-sync"></i> Forzar Sincronización SaaS
                </button>
                <button onclick="testWasabi()" class="btn btn-primary">
                    <i class="fas fa-network-wired"></i> Test Wasabi Link
                </button>
            </div>
          </div>';

    // Script para la sincronización y el test
    echo '<script>
        function syncSaaSPlan() {
            if(!confirm("¿Estás seguro de forzar la sincronización de todos los servicios activos con el API? Esto reparará los planes Manual Only.")) return;
            
            const btn = event.target.closest("button");
            btn.disabled = true;
            btn.innerHTML = "<i class=\"fas fa-spinner fa-spin\"></i> Sincronizando...";

            window.location.href = window.location.href + "&action=sync_saas";
        }
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

    // Lógica de Sincronización Forzada (V11.2)
    if (isset($_GET['action']) && $_GET['action'] == 'sync_saas') {
        $hooksPath = ROOTDIR . DIRECTORY_SEPARATOR . 'includes' . DIRECTORY_SEPARATOR . 'hooks' . DIRECTORY_SEPARATOR . 'dbp_provisioning.php';
        if (file_exists($hooksPath)) {
            require_once($hooksPath);
            
            $services = Illuminate\Database\Capsule\Manager::table('tblhosting')
                ->join('tblproducts', 'tblhosting.packageid', '=', 'tblproducts.id')
                ->where(function($query) {
                    $query->where('tblproducts.servertype', 'dockerbackuppro')
                          ->orWhere('tblproducts.name', 'like', '%Docker%Backup%')
                          ->orWhere('tblproducts.name', 'like', '%DBP%');
                })
                ->where('tblhosting.domainstatus', 'Active')
                ->select('tblhosting.id', 'tblhosting.userid', 'tblproducts.name')
                ->get();

            $count = 0;
            foreach ($services as $service) {
                $plan = 'basic';
                $retention = 2;
                $name = $service->name;
                if (stripos($name, 'Enterprise') !== false || stripos($name, 'Premium') !== false) { 
                    $plan = 'enterprise'; 
                    $retention = 30; 
                } elseif (stripos($name, 'Standard') !== false || stripos($name, 'Pro') !== false) { 
                    $plan = 'standard'; 
                    $retention = 7; 
                }

                $client = Illuminate\Database\Capsule\Manager::table('tblclients')->where('id', $service->userid)->first();

                // Llama al hook con la nueva función (definida en el archivo de hooks)
                if (function_exists('dbp_call_api')) {
                    dbp_call_api('/v1/whmcs/provision', [
                        'service_id'     => (string)$service->id,
                        'client_email'   => $client->email,
                        'plan'           => $plan,
                        'retention_days' => (int)$retention
                    ]);
                    $count++;
                }
            }

            // Sincronizar el Webhook Global
            $webhookUrl = $vars['n8n_webhook'];
            if (function_exists('dbp_call_api') && !empty($webhookUrl)) {
                dbp_call_api('/v1/admin/webhook', [
                    'webhook_url' => $webhookUrl
                ]);
            }

            echo "<div class='alert alert-success'>Sincronización completada: {$count} servicios procesados en el API y Webhook Global actualizado.</div>";
        } else {
            echo "<div class='alert alert-danger'>Error crítico: No se encuentra el motor de API en: {$hooksPath}</div>";
        }
    }

    // El iframe carga el dashboard en modo admin con el token maestro y debug si aplica
    echo '<iframe src="' . $portalUrl . '?admin=1&sso=' . $masterToken . $debug . '" 
            width="100%" 
            height="900" 
            style="border:0; border-radius: 8px; box-shadow: 0 10px 30px rgba(0,0,0,0.5);" 
            frameborder="0"></iframe>';
}
