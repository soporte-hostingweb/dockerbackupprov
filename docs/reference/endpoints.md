# Endpoints del API

La comunicación entre los agentes y el panel de control se realiza mediante una API REST en Go.

## Endpoints Públicos

| Método | Endpoint | Función |
|:---|:---|:---|
| `GET` | `/ping` | Health check del API |
| `GET` | `/install.sh` | Descarga el script instalador del agente |
| `POST` | `/v1/activate` | Activación de token + vinculación de hardware |
| `POST` | `/v1/auth/login` | Login al dashboard (retorna JWT) |

## Agente (Requieren Autenticación)
*Requieren Headers: X-Agent-ID + X-Agent-Key + X-Agent-Fingerprint*

| Método | Endpoint | Función |
|:---|:---|:---|
| `POST` | `/v1/agent/heartbeat` | Latido del agente (cada 10s), recibe instrucciones |
| `GET` | `/v1/agent/config` | Obtiene configuración de backup del agente |
| `POST` | `/v1/agent/config` | Guarda rutas seleccionadas |
| `POST` | `/v1/agent/config/save` | Guarda configuración completa (schedule + protection_level) |
| `GET` | `/v1/agent/status` | Estado de todos los agentes del tenant |
| `POST` | `/v1/agent/action/:id` | Envía acciones (force_selected, force_full, maintenance_on...) |
| `POST` | `/v1/agent/backup/complete` | Reporta resultado del backup |
| `DELETE` | `/v1/agent/status/:id` | Elimina agente del panel |

## Admin (Requieren X-Admin-Key o MASTER_ADMIN_TOKEN)

| Método | Endpoint | Función |
|:---|:---|:---|
| `POST` | `/v1/whmcs/provision` | Provisiona un nuevo tenant desde WHMCS |
| `POST` | `/v1/admin/license/generate` | Genera/regenera token SaaS para un servicio |
| `POST` | `/v1/tenant/suspend` | Suspende operaciones del tenant |
| `POST` | `/v1/tenant/unsuspend` | Reactiva el tenant |
| `POST` | `/v1/tenant/upgrade` | Cambia el plan del tenant |
