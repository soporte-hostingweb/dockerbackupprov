<?php
/**
 * DBP Provisioning Hook: Link WHMCS Service to SaaS Control Plane
 * Centralized SaaS Configuration V9.1
 */

use Illuminate\Database\Capsule\Manager as Capsule;

add_hook('AfterAcceptOrder', 1, function($vars) {
    $orderId = (int)$vars['orderid'];
    
    // 1. Obtener Configuración Global desde el Addon Module
    $addonSettings = Capsule::table('tbladdonmodules')
        ->where('module', 'dockerbackuppro')
        ->pluck('value', 'setting');

    $endpoint = rtrim($addonSettings['api_endpoint'] ?? '', '/');
    $adminKey = $addonSettings['master_token'] ?? '';

    if (empty($endpoint) || empty($adminKey)) {
        logActivity("[DBP] Error: API Endpoint o Master Token no configurados en el Addon Module.");
        return;
    }

    // 2. Obtener los servicios asociados a esta orden
    $services = Capsule::table('tblhosting')->where('orderid', $orderId)->get();
    
    foreach ($services as $service) {
        // Validar que el producto use nuestro módulo provisioning
        $product = Capsule::table('tblproducts')
            ->where('id', $service->packageid)
            ->where('servertype', 'dockerbackuppro')
            ->first();

        if (!$product) continue;

        // 3. Mapeo de Planes internos SaaS basados en nombre del producto
        $plan = 'basic';
        if (stripos($product->name, 'Enterprise') !== false) $plan = 'enterprise';
        elseif (stripos($product->name, 'Standard') !== false) $plan = 'standard';

        // 4. Obtener Email del cliente
        $client = Capsule::table('tblclients')->where('id', $service->userid)->first();

        $payload = [
            'service_id'     => (string)$service->id,
            'client_email'   => $client->email,
            'plan'           => $plan,
            'retention_days' => ($plan == 'enterprise') ? 30 : 7
        ];

        // 5. Ejecutar Provisión vía API Central
        $ch = curl_init($endpoint . '/v1/whmcs/provision');
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

        if ($httpCode != 200) {
            logActivity("[DBP] Fallo en provisión API para servicio {$service->id}. HTTP Code: {$httpCode}");
        }
    }
});
