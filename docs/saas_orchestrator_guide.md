# HW Cloud Recovery - SaaS Orchestrator Guide (V15)

## 1. Enlaces Permanentes de Gestión

### Panel de Administración (Master Admin)
Utilice este enlace para visualizar la infraestructura global de todos los clientes de HWPeru.
- **URL**: `https://backup.hwperu.com/?admin=1&token=TU_MASTER_KEY`
- **Funciones**: Visualización de todos los agentes, auditoría de integridad global, gestión de cuotas y test de conectividad S3 maestro.

### Panel de Usuario (Client Area Embed)
Este es el enlace que se integra dentro de WHMCS para cada cliente.
- **URL Ejemplo**: `https://backup.hwperu.com/?sso=dbp_saas_3070&embed=1`
- **Funciones**: Restore Wizard, Configuración de Backups, Monitor de Telemetría propio.

---

## 2. Matriz de Permisos por Plan (SaaS Policy Engine)

| Característica | Basic (Gratis) | Standard (Pro) | Enterprise (DRaaS) |
| :--- | :---: | :---: | :---: |
| **Retención Máxima** | 2 Snapshots | 7 Snapshots | 30+ Snapshots |
| **Backup Manual** | ✅ | ✅ | ✅ |
| **Backup Programado** | ❌ (Manual) | ✅ (Diario) | ✅ (Personalizado) |
| **Restore Wizard Pro** | ✅ (Básico) | ✅ (Full) | ✅ (Full + 1-Click WP) |
| **DB Consistency Hook** | ❌ | ✅ | ✅ (Avanzado) |
| **S3 Propio (BYOS)** | ❌ | ❌ | ✅ |
| **Clonación VPS (SSH)** | ❌ | ❌ | ✅ |

---

## 3. Configuración de Almacenamiento Propio (BYOS)

Para activar el almacenamiento propio, el cliente debe:
1. Ir a **Settings** en su panel.
2. Activar el Toggle **"Utilizar Almacenamiento S3 Propio"**.
3. Ingresar las credenciales de su proveedor (AWS, Wasabi, MinIO, etc).
4. El sistema re-configurará el agente automáticamente para dirigir los bloques de datos a su bucket privado.

---

## 4. Auditoría para el Administrador
El administrador puede identificar agentes con S3 propio buscando el icono de **"External Storage"** en el Dashboard Maestro. Esto permite gestionar el soporte técnico sabiendo que la data no reside en los servidores de HWPeru.
