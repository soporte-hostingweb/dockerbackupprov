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
        'DisplayName' => 'Docker Backup Pro',
        'APIVersion' => '1.1',
        'RequiresServer' => false, 
    );
}

// Configurable options for the product in WHMCS Admin
function dockerbackuppro_ConfigOptions()
{
    return array(
        "Storage Quota GB" => array("Type" => "text", "Size" => "25", "Default" => "50"),
        "API Endpoint" => array("Type" => "text", "Size" => "25", "Default" => "https://api.dockerbackuppro.com/v1"),
    );
}

// Hook that triggers after successful payment / manual trigger
function dockerbackuppro_CreateAccount(array $params)
{
    try {
        $quota = $params['configoption1'];
        $endpoint = $params['configoption2'];
        $serviceId = $params['serviceid'];
        $userId = $params['userid'];
        $clientEmail = $params['clientsdetails']['email'];

        // [MOCK] Petición a nuestra API Go (Fase 3) para registrar el sub-arrendatario de S3
        /*
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $endpoint . "/admin/provision");
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode(['email' => $clientEmail, 'quota' => $quota]));
        ...
        */

        // Simulamos que la API nos devuelve el token seguro. Lo guardamos en campo interno.
        $mockGeneratedToken = "dbp_tenant_" . md5($userId . $serviceId);
        
        return "success";
    } catch (Exception $e) {
        return $e->getMessage();
    }
}

// Triggers when invoice is overdue
function dockerbackuppro_SuspendAccount(array $params)
{
    try {
        // [MOCK] $params['configoption2'] . "/admin/suspend"
        return "success";
    } catch (Exception $e) {
        return $e->getMessage();
    }
}

function dockerbackuppro_UnsuspendAccount(array $params)
{
    try {
        // [MOCK] $params['configoption2'] . "/admin/unsuspend"
        return "success";
    } catch (Exception $e) {
        return $e->getMessage();
    }
}

function dockerbackuppro_TerminateAccount(array $params)
{
    try {
        // [MOCK] Notify API to delete tenant metrics and wipe S3 bucket folder
        return "success";
    } catch (Exception $e) {
        return $e->getMessage();
    }
}

// Render in Client Area Profile page
function dockerbackuppro_ClientArea(array $params)
{
    // Recuperamos el token almacenado (Simulation)
    $token = "dbp_tenant_" . md5($params['userid'] . $params['serviceid']); 
    
    // Inyectamos las variables al TPL
    return array(
        'tabOverviewReplacementTemplate' => 'clientarea.tpl',
        'templateVariables' => array(
            'dbpToken' => $token,
            'apiEndpoint' => $params['configoption2'],
            'installCommand' => "curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token {$token}",
        ),
    );
}
