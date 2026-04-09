<?php
/**
 * Script de verificación para el módulo AccesoCF
 * Ejecutar desde el directorio del módulo para verificar la instalación
 */

echo "=== VERIFICACIÓN DEL MÓDULO ACCESOCF ===\n\n";

// Verificar archivos principales
$required_files = [
    'accesocf.php',
    'hook.php', 
    'config.php',
    'model/acceso_cf.php'
];

echo "1. Verificando archivos requeridos:\n";
foreach ($required_files as $file) {
    if (file_exists($file)) {
        echo "   ✓ $file - OK\n";
    } else {
        echo "   ✗ $file - FALTANTE\n";
    }
}

// Verificar funciones del módulo
echo "\n2. Verificando funciones del módulo:\n";
if (file_exists('accesocf.php')) {
    $content = file_get_contents('accesocf.php');
    $required_functions = [
        'accesocf_config',
        'accesocf_activate', 
        'accesocf_deactivate',
        'accesocf_upgrade',
        'accesocf_clientarea',
        'accesocf_output'
    ];
    
    foreach ($required_functions as $func) {
        if (strpos($content, "function $func") !== false) {
            echo "   ✓ $func() - OK\n";
        } else {
            echo "   ✗ $func() - FALTANTE\n";
        }
    }
}

// Verificar permisos
echo "\n3. Verificando permisos de archivos:\n";
foreach ($required_files as $file) {
    if (file_exists($file)) {
        $perms = fileperms($file);
        $readable = is_readable($file) ? 'SÍ' : 'NO';
        echo "   $file: " . substr(sprintf('%o', $perms), -4) . " (legible: $readable)\n";
    }
}

// Verificar estructura de directorio
echo "\n4. Verificando estructura de directorio:\n";
$current_dir = basename(getcwd());
if ($current_dir === 'accesocf') {
    echo "   ✓ Directorio correcto: $current_dir\n";
} else {
    echo "   ⚠ Directorio actual: $current_dir (debería ser 'accesocf')\n";
}

// Verificar ubicación
echo "\n5. Verificando ubicación:\n";
$path = getcwd();
if (strpos($path, 'modules/addons/accesocf') !== false) {
    echo "   ✓ Ubicación correcta en modules/addons/\n";
} else {
    echo "   ⚠ Ubicación: $path\n";
    echo "   Debería estar en: .../modules/addons/accesocf/\n";
}

echo "\n=== FIN DE VERIFICACIÓN ===\n";
echo "\nSi todos los elementos están marcados con ✓, el módulo debería funcionar correctamente.\n";
echo "Si hay elementos marcados con ✗ o ⚠, corregir antes de activar en WHMCS.\n";
?>
