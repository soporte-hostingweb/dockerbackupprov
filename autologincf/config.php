<?php
/**
 * Configuración del módulo AccesoCF
 * Este archivo ayuda a WHMCS a detectar correctamente el módulo
 */

// Verificar que estamos en el contexto correcto de WHMCS
if (!defined("WHMCS")) {
    die("Este archivo no puede ser accedido directamente");
}

// Configuración del módulo
$module_config = array(
    'name' => 'AccesoCF',
    'version' => '1.13',
    'author' => 'Carlos Frias',
    'description' => 'Módulo para acceso automático a facturas vía enlaces seguros',
    'language' => 'spanish'
);

return $module_config;
?>
