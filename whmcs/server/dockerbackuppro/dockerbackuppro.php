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
        'DisplayName' => 'HW Cloud Recovery - Smart SaaS',
        'APIVersion' => '14.2',
        'RequiresServer' => false, 
    );
}

// Configurable options for the product in WHMCS Admin
function dockerbackuppro_ConfigOptions()
{
    return array(
        "Storage Quota GB" => array("Type" => "text", "Size" => "25", "Default" => "50"),
        "Debug Mode" => array("Type" => "yesno", "Description" => "Activa herramientas de diagnóstico en el área de cliente."),
        "Plan Type" => array("Type" => "dropdown", "Options" => "Standard,PRO WordPress,Enterprise"),
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
    return "success"; 
}

// Render in Client Area Profile page
function dockerbackuppro_ClientArea(array $params)
{
    $config = _dockerbackuppro_GetGlobalConfig();
    $endpoint = rtrim($config['api_endpoint'] ?? 'https://api.hwperu.com', '/');
    
    // V14.2.5: Recuperar Token Opaco y Seguro desde los Custom Fields
    $token = Illuminate\Database\Capsule\Manager::table('tblcustomfieldsvalues')
        ->join('tblcustomfields', 'tblcustomfields.id', '=', 'tblcustomfieldsvalues.fieldid')
        ->where('tblcustomfieldsvalues.relid', $params['serviceid'])
        ->where('tblcustomfields.fieldname', 'like', 'Token%')
        ->value('value');

    // Fallback: Si no hay token guardado (servicio viejo), mantenemos el formato anterior para no romper compatibilidad
    if (!$token) {
        $token = "dbp_saas_" . $params['serviceid'];
    }
    
    $debug = ($params['configoption2'] == 'on') ? '&debug=1' : '';
    $planType = $params['configoption3'] ?? 'Standard';

    return array(
        'tabOverviewReplacementTemplate' => 'clientarea.tpl',
        'templateVariables' => array(
            'dbpToken' => $token,
            'apiEndpoint' => $endpoint,
            'debug' => $debug,
            'planType' => $planType,
            'installCommand' => "curl -sSL {$endpoint}/install.sh | bash -s -- --token {$token}",
        ),
    );
}

