<?php
/**
 * DBP Lifecycle Management Hooks (v11.0)
 * Triggers API actions based on WHMCS Service Events
 */

function dbp_call_api($endpoint_path, $payload) {
    $addonSettings = Illuminate\Database\Capsule\Manager::table('tbladdonmodules')
        ->where('module', 'dockerbackuppro')
        ->pluck('value', 'setting');

    $endpoint = rtrim($addonSettings['api_endpoint'] ?? '', '/');
    $adminKey = $addonSettings['master_token'] ?? '';

    if (empty($endpoint) || empty($adminKey)) {
        logActivity("[DBP] Error: API Configuration missing in Addon Module.");
        return false;
    }

    $ch = curl_init($endpoint . $endpoint_path);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'X-Admin-Key: ' . $adminKey
    ]);

    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    return ['code' => $httpCode, 'response' => $response];
}

// 1. PROVISIÓN INICIAL
add_hook('AfterAcceptOrder', 1, function($vars) {
    $orderId = (int)$vars['orderid'];
    $services = Illuminate\Database\Capsule\Manager::table('tblhosting')->where('orderid', $orderId)->get();
    
    foreach ($services as $service) {
        $product = Illuminate\Database\Capsule\Manager::table('tblproducts')->where('id', $service->packageid)->first();
        if (!$product) continue;

        // Validar si es un producto DBP (por ServerType o por Nombre)
        $isDBP = ($product->servertype == 'dockerbackuppro' || 
                  stripos($product->name, 'Docker') !== false || 
                  stripos($product->name, 'DBP') !== false || 
                  stripos($product->name, 'Backup') !== false);
        
        if (!$isDBP) continue;

        $plan = 'basic';
        $retention = 2;
        $name = $product->name;
        if (stripos($name, 'Enterprise') !== false || stripos($name, 'Premium') !== false) { 
            $plan = 'enterprise'; 
            $retention = 30; 
        } elseif (stripos($name, 'Standard') !== false || stripos($name, 'Pro') !== false) { 
            $plan = 'standard'; 
            $retention = 7; 
        }

        $client = Illuminate\Database\Capsule\Manager::table('tblclients')->where('id', $service->userid)->first();

        $res = dbp_call_api('/v1/whmcs/provision', [
            'service_id'     => (string)$service->id,
            'client_email'   => $client->email,
            'plan'           => $plan,
            'retention_days' => (int)$retention
        ]);

        if ($res['code'] == 200) {
            $data = json_decode($res['response'], true);
            $token = $data['token'] ?? '';
            logActivity("[DBP] Provisión exitosa para Servicio #{$service->id}. Token: {$token}");
            
            // Guardar token en Custom Field (Asumimos ID 1 para 'Token')
            // Illuminate\Database\Capsule\Manager::table('tblcustomfieldsvalues')->updateOrInsert(...)
        }
    }
});

// 2. CAMBIO DE PLAN (Upgrade/Downgrade)
add_hook('AfterProductUpgrade', 1, function($vars) {
    $upgradeId = $vars['upgradeid'];
    $upgrade = Illuminate\Database\Capsule\Manager::table('tblupgrades')->where('id', $upgradeId)->first();
    $service = Illuminate\Database\Capsule\Manager::table('tblhosting')->where('id', $upgrade->relid)->first();
    $product = Illuminate\Database\Capsule\Manager::table('tblproducts')->where('id', $service->packageid)->first();

    // Validar si es DBP
    if (stripos($product->name, 'Docker') === false && stripos($product->name, 'DBP') === false && stripos($product->name, 'Backup') === false) return;

    $plan = 'basic';
    $retention = 2;
    $name = $product->name;
    if (stripos($name, 'Enterprise') !== false || stripos($name, 'Premium') !== false) { 
        $plan = 'enterprise'; 
        $retention = 30; 
    } elseif (stripos($name, 'Standard') !== false || stripos($name, 'Pro') !== false) { 
        $plan = 'standard'; 
        $retention = 7; 
    }

    dbp_call_api('/v1/whmcs/provision', [
        'service_id'     => (string)$service->id,
        'client_email'   => Illuminate\Database\Capsule\Manager::table('tblclients')->where('id', $service->userid)->value('email'),
        'plan'           => $plan,
        'retention_days' => (int)$retention
    ]);
    
    logActivity("[DBP] Plan actualizado por Upgrade/Sincronización Automática para Servicio #{$service->id} -> {$plan}");
});

// 3. SUSPENSIÓN
add_hook('ModuleSuspend', 1, function($vars) {
    $serviceId = $vars['params']['serviceid'];
    dbp_call_api('/v1/tenant/suspend', ['service_id' => (string)$serviceId]);
    logActivity("[DBP] Servicio #{$serviceId} suspendido (Mantenimiento Forzoso aplicado).");
});

// 4. REACTIVACIÓN
add_hook('ModuleUnsuspend', 1, function($vars) {
    $serviceId = $vars['params']['serviceid'];
    dbp_call_api('/v1/tenant/unsuspend', ['service_id' => (string)$serviceId]);
    logActivity("[DBP] Servicio #{$serviceId} reactivado.");
});

// 5. TERMINACIÓN
add_hook('ModuleTerminate', 1, function($vars) {
    $serviceId = $vars['params']['serviceid'];
    dbp_call_api('/v1/tenant/terminate', ['service_id' => (string)$serviceId]);
    logActivity("[DBP] Servicio #{$serviceId} terminado (Token invalidado).");
});
