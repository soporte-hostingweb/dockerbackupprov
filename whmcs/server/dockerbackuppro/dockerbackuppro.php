<?php
/**
 * Docker Backup Pro - WHMCS Server/Provisioning Module
 * Phase 5 Implementation
 */

if (!defined("WHMCS")) {
    die("This file cannot be accessed directly");
}

function dockerbackuppro_MetaData()
{
    return array(
        'DisplayName' => 'HW Cloud Recovery',
        'APIVersion' => '1.1',
        'RequiresServer' => false, 
    );
}

// Configurable options for the product in WHMCS Admin
function dockerbackuppro_ConfigOptions()
{
    return array(
        "Storage Quota GB" => array("Type" => "text", "Size" => "25", "Default" => "50"),
        "Debug Mode" => array("Type" => "yesno", "Description" => "Activa herramientas de diagnóstico en el área de cliente."),
    );
}

// Helper para obtener configuración global del Addon
function _dockerbackuppro_GetGlobalConfig() {
    $results = Illuminate\Database\Capsule\Manager::table('tbladdonmodules')
        ->where('module', 'dockerbackuppro')
        ->pluck('value', 'setting');
    return $results;
}

// Hook that triggers after successful payment / manual trigger
function dockerbackuppro_CreateAccount(array $params)
{
    return "success"; // El aprovisionamiento SaaS se delega al hook AfterAcceptOrder para mayor robustez
}

// ... (Suspend/Terminate functions stay the same) ...

// Render in Client Area Profile page
function dockerbackuppro_ClientArea(array $params)
{
    $config = _dockerbackuppro_GetGlobalConfig();
    $endpoint = rtrim($config['api_endpoint'] ?? 'https://api.hwperu.com', '/');
    
    // V11.3.0: Formato de Token Determinístico (SaaS Estable)
    $token = "dbp_saas_" . $params['serviceid']; 
    
    // Inyectamos las variables al TPL
    $debug = ($params['configoption2'] == 'on') ? '&debug=1' : ''; // configoption2 es ahora Debug Mode

    return array(
        'tabOverviewReplacementTemplate' => 'clientarea.tpl',
        'templateVariables' => array(
            'dbpToken' => $token,
            'apiEndpoint' => $endpoint,
            'debug' => $debug,
            'installCommand' => "curl -sSL {$endpoint}/install.sh | bash -s -- --token {$token}",
        ),
    );
}

