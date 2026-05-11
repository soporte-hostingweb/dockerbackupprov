# Configuración del Módulo WHMCS

**Ubicación de archivos en WHMCS:**
```text
whmcs/modules/servers/dockerbackuppro/
├── dockerbackuppro.php   ← Módulo principal
└── clientarea.tpl        ← Plantilla del área de cliente
```

## Configuración en WHMCS Admin
Navega a **Configuración → Módulos de Servidor**:
1. Activar el módulo `HW Cloud Recovery`
2. Configurar el campo `api_endpoint`: `https://api.hwperu.com`
3. Configurar el campo `master_token`: Tu `MASTER_ADMIN_TOKEN`

## Cómo funciona el módulo PHP

Al vender un producto, WHMCS llama a la función `dockerbackuppro_ClientArea()`. El módulo genera un token determinístico:
```php
$token = "dbp_saas_" . $params['serviceid'];
$installCommand = "curl -sSL {$endpoint}/install.sh | bash -s -- --token {$token}";
```

## Auto-Provisión al Vender

Cuando WHMCS acepta un pedido, el hook llama al API:
```http
POST https://api.hwperu.com/v1/whmcs/provision
Header: X-Admin-Key: TU_API_ADMIN_KEY
Body:
{
  "service_id": "3070",
  "client_email": "cliente@correo.com",
  "plan": "enterprise",
  "retention_days": 30
}
```

**Lo que el API hace automáticamente:**
- Crea el `TenantPlan` con las políticas del plan.
- Para plan Enterprise: pre-configura "Protección Total" automáticamente.
- Crea `UserSettings` con Wasabi global como fallback.
- Activa alertas heredadas del Webhook Maestro.
- Deja el token `dbp_saas_3070` listo para que el cliente instale.
