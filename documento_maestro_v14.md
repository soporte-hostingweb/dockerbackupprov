# 🛡️ HW Cloud Recovery — Documento Maestro V14.1
### Enterprise Disaster Recovery as a Service (DRaaS)

**Versión del Sistema:** V14.1.0  
**Estado:** Producción SaaS  
**Infraestructura:** `api.hwperu.com` · `backup.hwperu.com` · `ghcr.io/hwperu/dbp-agent:prod`

---

## 1. FUNCIÓN PRINCIPAL

HW Cloud Recovery es una **plataforma SaaS de respaldo y recuperación ante desastres** diseñada para proteger servidores VPS de manera automática, sin interrumpir los servicios activos.

### ¿Qué hace exactamente?

| Función | Descripción |
|:---|:---|
| **Backup Automatizado** | Respaldo incremental de archivos, carpetas y contenedores Docker usando Restic + Wasabi S3 |
| **Recuperación 1-Click** | Restauración completa o parcial de datos desde cualquier snapshot histórico |
| **Monitoreo 24/7** | Health Score en tiempo real, alertas de fallos vía Webhook (n8n, WhatsApp) |
| **Licenciamiento SaaS** | Control de acceso por token hardware-bound, anti-clonado, expiración automática |
| **Integración WHMCS** | Provisión automática al vender un servicio, panel embebido en el área de cliente |
| **Multi-Plan** | Tres niveles con políticas técnicas distintas aplicadas automáticamente |
| **Zero-Downtime** | Backup en caliente sin interrupción de servicios (JetBackup style) |

---

## 📄 DOCUMENTACIÓN DE PLANES Y ESCENARIOS

Para guías detalladas de instalación y casos de éxito por nivel de servicio, consulte:

- 🎁 [**PLAN FREE**](file:///e:/WHMCS/dockerkup/PLANS_FREE.md): Micro-servicios y blogs personales.
- 🛡️ [**PLAN STANDARD**](file:///e:/WHMCS/dockerkup/PLANS_STANDARD.md): E-commerce y WP en producción diaria.
- 💎 [**PLAN ENTERPRISE**](file:///e:/WHMCS/dockerkup/PLANS_ENTERPRISE.md): Misión crítica, DR proactivo y alta disponibilidad.

---

---

## 2. PLATAFORMAS Y TECNOLOGÍAS

### Stack Completo

```
┌─────────────────────────────────────────────┐
│              CONTROL PLANE                   │
│  api.hwperu.com (Go + Gin + PostgreSQL)      │
│  - Motor de Licencias SaaS                  │
│  - Policy Engine por Plan                   │
│  - Dispatcher de Alertas (Webhook)          │
│  - Endpoints REST para Agente y UI          │
└────────────────┬────────────────────────────┘
                 │ HTTPS / TLS 1.3
┌────────────────▼────────────────────────────┐
│              DATA PLANE (Cliente)            │
│  ghcr.io/hwperu/dbp-agent:prod (Alpine Go)  │
│  - Heartbeat cada 10 segundos               │
│  - Ejecuta backups según config del API     │
│  - Genera Fingerprint de Hardware (SHA256)  │
└────────────────┬────────────────────────────┘
                 │ S3 Protocol
┌────────────────▼────────────────────────────┐
│              STORAGE                         │
│  Wasabi S3 (o cualquier S3-compatible)      │
│  - Datos cifrados con Restic (AES-256)      │
│  - Retención configurable por plan          │
└─────────────────────────────────────────────┘
```

### Componentes

| Componente | Tecnología | Propósito |
|:---|:---|:---|
| **API Central** | Go 1.25 + Gin Framework | Control Plane, endpoints REST, motor de reglas |
| **Base de Datos** | PostgreSQL | Persistencia de agentes, configs, licencias, logs |
| **Agente** | Go 1.25 compilado en Alpine | Se ejecuta en el VPS del cliente como contenedor Docker |
| **Storage** | Restic + Wasabi S3 | Backup incremental cifrado con deduplicación |
| **Dashboard UI** | Next.js + TypeScript | Panel de administración y cliente |
| **WHMCS** | PHP Module + TPL | Integración de billing y provisión automática |
| **Distribución** | GHCR (GitHub Container Registry) | Imagen Docker privada del agente |
| **Alertas** | Webhook universal (n8n/WhatsApp) | Notificaciones de eventos críticos |

---

## 3. ARQUITECTURA DE SEGURIDAD (V14)

### 3.1 Fingerprint Híbrido Anti-Clonado
Cada agente genera una identidad única e irrepetible al instalarse:
```
fingerprint = SHA256(machine-id + disk_serial_or_uuid + hostname)
```
- `machine-id`: ID único del sistema operativo (`/etc/machine-id`)
- `disk_serial`: Serie física del disco (`/sys/block/sda/device/serial`)
- `hostname`: Nombre del servidor

Esta huella se envía en **cada latido** (`X-Agent-Fingerprint`). Si no coincide con la registrada al activar el token, el API rechaza la conexión.

### 3.2 Ciclo de Vida del Token SaaS
```
WHMCS Provisión → Token generado (dbp_saas_XXXX)
      ↓
Estado: pending (válido 48h)
      ↓
Cliente instala → Token se vincula al hardware
Estado: activated
      ↓
Si pasan 48h sin instalar → Estado: expired
      ↓
Admin puede revocar → Estado: revoked
```

### 3.3 Capas de Seguridad

| Capa | Tecnología | Detalle |
|:---|:---|:---|
| **Data at Rest** | AES-256-GCM | Credenciales S3 cifradas en DB con llave maestra de 32 bytes |
| **Data in Transit** | TLS 1.3 | Toda comunicación Agente ↔ API |
| **Autenticación** | X-Agent-ID + X-Agent-Key + X-Agent-Fingerprint | Triple validación en cada request |
| **Imagen Docker** | GHCR Privado | Cliente nunca ve el código fuente |
| **PAT GHCR** | Solo en backend | Token de descarga nunca expuesto al cliente |
| **Base de Datos** | Red interna Docker | PostgreSQL sin exposición pública |

---

## 4. CÓMO INSTALAR (ADMINISTRADOR)

### 4.1 Requisitos del Servidor API

- Servidor Linux con Docker + Docker Compose
- Puerto 443 (HTTPS) abierto
- Dominio apuntando al servidor (`api.hwperu.com`)
- Cuenta GitHub con PAT (permisos `write:packages`, `read:packages`)

### 4.2 Variables de Entorno (`.env` en el servidor API)

```bash
# Conexión a Base de Datos
DB_HOST=postgres
DB_USER=dbp_user
DB_PASS=contraseña_segura
DB_NAME=dbp_production
DB_PORT=5432

# Seguridad
MASTER_ADMIN_TOKEN=token_maestro_para_whmcs  # Para generar licencias
API_ADMIN_KEY=llave_admin_para_endpoints      # Para suspender/activar tenants
DBP_ENCRYPTION_KEY=llave_aes_de_32_caracteres # Cifrado de credenciales S3

# Distribución Privada (GHCR)
GHCR_READ_PAT=ghp_token_de_lectura_de_github  # Se entrega al agente al activar

# Wasabi S3 Global (Fallback para clientes sin S3 propio)
WASABI_ACCESS_KEY=...
WASABI_SECRET_KEY=...
WASABI_BUCKET=hwperu-backups
WASABI_REGION=us-east-1

# Alertas (Webhook Maestro)
WEBHOOK_URL=https://tu-n8n.com/webhook/xyz
```

### 4.3 Despliegue de la API

```bash
# 1. Clonar repositorio privado
git clone https://github.com/soporte-hostingweb/dockerbackupprov.git
cd dockerbackupprov

# 2. Configurar variables de entorno
cp .env.example .env
nano .env  # Editar con los valores reales

# 3. Levantar servicios
docker compose up -d

# 4. Verificar que está activo
curl https://api.hwperu.com/ping
# Respuesta esperada: {"status":"ok","version":"V14.1.0"}

# 5. Subir imagen del agente a GHCR
echo "TU_GITHUB_PAT" | docker login ghcr.io -u hwperu --password-stdin
docker build -t ghcr.io/hwperu/dbp-agent:prod ./agent/
docker push ghcr.io/hwperu/dbp-agent:prod
```

### 4.4 Configuración del Módulo WHMCS

**Ubicación de archivos:**
```
whmcs/modules/servers/dockerbackuppro/
├── dockerbackuppro.php   ← Módulo principal
└── clientarea.tpl        ← Plantilla del área de cliente
```

**En WHMCS Admin → Configuración → Módulos de Servidor:**
1. Activar el módulo `HW Cloud Recovery`
2. Configurar el campo `api_endpoint`: `https://api.hwperu.com`
3. Configurar el campo `master_token`: Tu `MASTER_ADMIN_TOKEN`

**Cómo funciona el módulo PHP:**
```php
// Al vender un producto, WHMCS llama a dockerbackuppro_ClientArea()
// El módulo genera el token determinístico:
$token = "dbp_saas_" . $params['serviceid'];

// Y construye el comando de instalación:
$installCommand = "curl -sSL {$endpoint}/install.sh | bash -s -- --token {$token}";

// Este comando se muestra al cliente en su área de servicio
```

### 4.5 Auto-Provisión al Vender en WHMCS

Cuando WHMCS acepta un pedido, el hook llama a:
```
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
- Crea el `TenantPlan` con las políticas del plan
- Para plan Enterprise: pre-configura "Protección Total" automáticamente
- Crea `UserSettings` con Wasabi global como fallback
- Activa alertas heredadas del Webhook Maestro
- Deja el token `dbp_saas_3070` listo para que el cliente instale

---

## 5. CÓMO INSTALAR (CLIENTE)

### 5.1 Desde WHMCS (Área de Cliente)

El cliente entra a **Mi Portal → Mis Productos y Servicios → Detalles del Servicio** y ve:

```
┌─────────────────────────────────────────────┐
│ HW Cloud Recovery - Continuidad Operativa   │
│                                             │
│ Copia este comando en tu servidor:          │
│ ┌─────────────────────────────────────────┐ │
│ │ curl -sSL https://api.hwperu.com/       │ │
│ │ install.sh | bash -s -- --token         │ │
│ │ dbp_saas_3070                           │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ Your Authorization Token: dbp_saas_3070    │
│ [Open Dashboard in New Tab]                 │
│                                             │
│ ┌─── Live Backup Control Panel ────────────┐│
│ │  [Panel embebido en iframe]              ││
│ └──────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
```

### 5.2 Flujo Completo de Instalación (Lo que ocurre al ejecutar el comando)

```bash
# El cliente ejecuta en su VPS como root:
curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token dbp_saas_3070
```

**Paso a paso interno:**

```
[1/5] Verificar dependencias
      ├── jq no encontrado → apt-get install jq (automático)
      └── Docker no encontrado → curl -fsSL https://get.docker.com | sh (automático)

[2/5] Generar Huella de Hardware
      fingerprint = SHA256(machine-id + disk_serial + hostname)
      ej: "a3f9b2c1d4e5..."

[3/5] Handshake con API
      POST https://api.hwperu.com/v1/activate
      {token, fingerprint, hostname, os}
      ↓
      API responde: {agent_id, api_key, ghcr_pat}

[4/5] Autenticación GHCR y descarga de imagen
      docker login ghcr.io -u hwperu (PAT temporal)
      docker pull ghcr.io/hwperu/dbp-agent:prod

[5/5] Despliegue del agente
      Guarda credenciales en /opt/docker-backup-pro/agent.json
      Crea docker-compose.yml con volúmenes correctos
      docker compose up -d

✅ El agente aparece en backup.hwperu.com en ~30 segundos
```

### 5.3 Arquitectura del Agente en el VPS del Cliente

```
VPS del Cliente
├── /opt/docker-backup-pro/
│   ├── agent.json          ← Credenciales cifradas (agent_id, api_key, fingerprint)
│   └── docker-compose.yml  ← Configuración del contenedor
│
└── Contenedor: dbp-client-agent
    ├── Montajes:
    │   ├── /var/run/docker.sock:ro  (Lee contenedores del host)
    │   ├── /:/host_root:ro          (Acceso a archivos del host)
    │   └── /opt/docker-backup-pro:/app/data  (Credenciales)
    │
    └── Ciclo cada 10 segundos:
        ├── Lee config del API (qué respaldar, cuándo, en qué modo)
        ├── Envía Heartbeat (estado, contenedores, snapshots)
        └── Ejecuta backup si hay tarea o es hora programada
```

---

## 6. GUÍA DE CADA BOTÓN Y ACCIÓN DEL PANEL

### 6.1 Panel Principal — Tarjeta del Servidor

| Elemento | Función |
|:---|:---|
| **HEALTH: 100%** | Puntaje de salud calculado automáticamente (0-100) |
| **OPERATIVO / OFFLINE** | Estado del agente basado en el último latido (<65 segundos = online) |
| **RTO: X min** | Recovery Time Objective estimado (cuánto tardaría la recuperación) |
| **RPO: Xh** | Recovery Point Objective (antigüedad del último backup disponible) |
| **DR READY** | Badge azul: indica que un sandbox test fue exitoso (Enterprise) |
| **Test DR Now** | Dispara una prueba de verificación en el nodo verificador |
| **Restore Wizard** | Abre el asistente de restauración con lista de snapshots |
| **Remove Offline Agent** | Elimina del panel un agente que lleva más de 65 segundos sin latido |
| **OFFLINE / OPERATIVO** | Badge de estado en tiempo real |

### 6.2 Panel Expandido — Configuración de Backup

| Elemento | Función técnica |
|:---|:---|
| **Selector de Zona Horaria** | Configura la zona horaria para el horario de backups |
| **Selector de Plan** | `Básico/Manual`, `Estándar (Diario)`, `Pro (Semanal)`, `Enterprise (Personalizado)` |
| **Save Configuration** | Llama a `POST /v1/agent/config/save` con schedule, paths y nivel de protección |
| **Reset** | Envía `action: reset` → borra rutas seleccionadas en la DB |
| **Selector de Días** | (Solo plan custom) Define qué días de la semana ejecutar el backup |
| **Slider de Hora** | (Solo plan custom) Define a qué hora del día ejecutar el backup |

### 6.3 FileExplorer — Selección de Archivos

| Elemento | Función técnica |
|:---|:---|
| **🔄 Protección automática inteligente** | Activa `[ALL_TARGETS]:contenedor` → el agente incluirá todos los archivos del contenedor dinámicamente |
| **💾 Copia completa del sistema** | Selecciona `[Full Volume] nombre_contenedor` → respaldo del volumen completo |
| **Checkboxes de carpetas** | Selección manual de rutas específicas dentro de cada contenedor |
| **Lock Selection** | Guarda la selección actual llamando a `POST /v1/agent/config` |

### 6.4 Botones de Acción Inferiores

| Botón | Acción | Comportamiento interno |
|:---|:---|:---|
| **Force Selected** | Backup inmediato de rutas seleccionadas | Envía `action: force_selected` → agente ejecuta backup en el siguiente ciclo |
| **Full** | Backup completo del servidor | Envía `action: force_full` → agente hace `restic backup /host_root` con exclusiones |
| **Maintenance Mode** | Pausa todas las operaciones | Pone `maintenance: true` en DB → el agente no ejecuta nada |
| **Resume Agent** | Reactiva tras mantenimiento | Pone `maintenance: false` |
| **Terminate & Pause** | Mata el proceso activo y entra en mantenimiento | Envía `action: kill_sync` + `maintenance_on` |

### 6.5 Sección de Snapshots

| Elemento | Función |
|:---|:---|
| **Lista de snapshots** | Muestra ID corto + fecha de cada backup disponible |
| **Restore** (en cada snapshot) | Abre el Restore Wizard prefiltrado a ese snapshot |
| **Plan Health** | Estado actual de la verificación de integridad |
| **Estimated RTO** | Tiempo estimado de recuperación en minutos |
| **Plan Limit** | Espacio máximo asignado en Wasabi |
| **SaaS Security Score** | Health Score del agente (misma métrica que en la cabecera) |
| **Current Usage (Wasabi)** | GB consumidos en el repositorio S3 |

### 6.6 Restore Wizard

| Paso | Función |
|:---|:---|
| **1. Selección de Snapshot** | Lista todos los snapshots disponibles con fecha y hora |
| **2. Explorador de Archivos** | Navegar dentro del snapshot para restaurar solo lo necesario |
| **3. Destino de Restauración** | Ruta donde se depositarán los archivos recuperados |
| **4. Auto-Up** | Si está activado, el agente intentará levantar los servicios Docker restaurados |
| **Confirmar Restore** | Envía tarea `restore:snapID|destino|paths|autoUp` al agente via heartbeat |

---

## 7. TIPOS DE PLANES Y COPIAS

### 7.1 Policy Engine (Código Real)

```go
var PolicyEngine = map[string]PlanPolicy{
    "basic": {
        MaxRetentionDays: 2,
        ValidationLvl:    "none",
        Priority:         1,
        AllowRestoreAuto: false,
    },
    "standard": {
        MaxRetentionDays: 7,
        ValidationLvl:    "basic",
        IntegrityLvl:     "light",
        Priority:         2,
        AllowRestoreAuto: true,
    },
    "enterprise": {
        MaxRetentionDays: 30,
        ValidationLvl:    "advanced",
        IntegrityLvl:     "medium",
        Priority:         3,
        AllowRestoreAuto: true,
    },
}
```

### 7.2 Comparativa de Planes

| Característica | FREE (Basic) | PRO (Standard) | ENTERPRISE |
|:---|:---:|:---:|:---:|
| **Retención** | 2 días | 7 días | 30 días |
| **Prioridad de Cola** | Baja (1) | Estándar (2) | VIP (3) |
| **Programación** | Solo manual | Diaria automática | Personalizada (días/hora) |
| **Verificación Integridad** | Ninguna | Semanal (light) | Diaria (medium) |
| **Restore Automático** | No | Sí (Agent) | Sí + Sandbox |
| **Nivel de Protección** | Manual | Avanzado | Total (auto-gestionado) |
| **Snapshot Mode** | live | live | live + opción consistent |
| **Auto-Config al instalar** | No | No | Sí (días + rutas + horario) |

### 7.3 Tipos de Copia por Plan

#### 🟢 FREE (Basic) — Selección Manual
```bash
# El usuario selecciona rutas específicas manualmente
# El agente ejecuta:
restic backup /host_root/var/www/html \
  --exclude /host_root/proc \
  --exclude /host_root/sys \
  --exclude /host_root/dev \
  --exclude /host_root/tmp \
  --exclude /host_root/var/lib/docker/overlay2
```
- **Sin interrupción de servicios**
- **Solo las carpetas que el usuario marcó en el panel**
- No hay verificación de integridad posterior
- Retención: máximo 2 copias (2 días)

#### 🔵 PRO (Standard) — Copia Automatizada + Dynamic Tracking
```bash
# El usuario activa "Protección automática inteligente" por contenedor
# Los [ALL_TARGETS]:contenedor se resuelven dinámicamente
# El agente ejecuta:
restic backup /host_root/app /host_root/data /host_root/config \
  --exclude /host_root/proc \
  [... exclusiones de producción ...]
```
- **Sin interrupción de servicios** (modo live)
- Programación automática diaria o semanal
- Verificación ligera semanal (`restic check`)
- Retención: 7 días
- Si el check falla, Health Score baja y se dispara alerta

#### 👑 ENTERPRISE — Protección Total Automática
```bash
# Al provisionar vía WHMCS, el sistema pre-configura automáticamente:
# ProtectionLevel=Total, paths=[ALL_SYSTEM_ROOT], schedule=daily 02:00

# El agente resuelve [ALL_SYSTEM_ROOT] como raíz del host:
restic backup /host_root \
  --exclude /host_root/proc \
  --exclude /host_root/sys \
  --exclude /host_root/dev \
  --exclude /host_root/run \
  --exclude /host_root/tmp \
  --exclude /host_root/var/tmp \
  --exclude /host_root/var/lib/docker/overlay2 \
  --exclude /host_root/var/lib/docker/aufs
```
- **Todo el servidor respaldado cada noche sin tocar nada**
- Modo LIVE por defecto (cero downtime)
- Modo CONSISTENT opcional (pausa controlada de contenedores)
- Verificación diaria de integridad
- Retención: 30 días
- Sandbox restore test para certificar "DR-Ready"

### 7.4 Snapshot Mode — Detalle Técnico

```go
// LIVE (default para TODOS los planes):
// restic backup corre sin pausar ningún servicio
// Posible inconsistencia mínima en archivos abiertos (ej: DB sin dump)
// ✅ Seguro para producción, cero downtime

// CONSISTENT (solo Enterprise, activado por el cliente):
// Antes del backup:
exec.Command("chroot /host_root docker ps -q | xargs -r docker pause").Run()
// ... restic backup ...
// Después del backup (defer):
exec.Command("chroot /host_root docker ps -q | xargs -r docker unpause").Run()
// ⚠️ Causa downtime temporal, datos 100% consistentes
```

---

## 8. HEALTH SCORE — CÓMO SE CALCULA

```go
func UpdateHealthScore(agentID string) {
    score := 100

    // -50 si está OFFLINE
    // -30 si está DEGRADED
    if agent.HealthStatus == "OFFLINE"  { score -= 50 }
    if agent.HealthStatus == "DEGRADED" { score -= 30 }

    // -20 si el último backup fue hace más de 24 horas
    if time.Since(agent.LastBackupAt) > 24*time.Hour { score -= 20 }

    // -15 si está en recuperación activa (reinicio local)
    // -40 si está en estado de desastre confirmado
    if agent.RecoveryTier == 2 { score -= 15 }
    if agent.RecoveryTier >= 3 { score -= 40 }

    // -40 si la verificación de integridad falló (planes que la requieren)
    if agent.VerificationStatus == "INVALID" { score -= 40 }

    // Clamped entre 0 y 100
}
```

---

## 9. ESCENARIOS DE CLIENTES

### ESCENARIO A: VPS Limpio con Docker ✅
**Perfil**: Ubuntu 22.04, Docker activos, hosting dedicado.

```bash
curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token dbp_saas_3070
# Duración: ~45 segundos
# Sin intervención adicional
```
**Resultado**: Agente operativo inmediatamente.

---

### ESCENARIO B: WordPress Instalado Manualmente (sin Docker) 🔧
**Perfil**: Cliente migrado de hosting compartido, Apache/Nginx + PHP + MySQL directo en el OS.

**¿Qué backed el agente?**
- `/var/www/html` — Archivos PHP y temas de WordPress
- `/etc/nginx` o `/etc/apache2` — Configuración del servidor web
- `/etc/letsencrypt` — Certificados SSL activos
- `/var/lib/mysql` — Bases de datos MySQL (si tiene permisos)
- `/etc/php` — Configuración de PHP

**¿Qué pasa al instalar?**
```
⚙️  Docker no encontrado. Instalando automáticamente...
[Descarga e instala Docker usando get.docker.com (~2 min)]
✅ Docker instalado correctamente
[Continúa con la activación y despliegue del agente]
```

> ⚠️ **Nota para soporte**: El backup NO pausa Apache ni MySQL. WordPress permanece funcional. En modo LIVE puede haber una mínima inconsistencia en archivos de sesión, pero los datos críticos (archivos PHP + DB) se respaldan correctamente.

**Configuración recomendada para este cliente (Plan PRO):**
1. Entrar al panel → Expandir el servidor
2. Seleccionar `Estándar (Backup Diario)` como plan
3. En el explorador: marcar `/var/www/html`, `/etc/nginx`, `/etc/letsencrypt`
4. `Save Configuration` → backup diario automático a las 2 AM

---

### ESCENARIO C: Sistema Personalizado sin Docker (Node.js, Python, etc.) 🔧
**Perfil**: Aplicación Node.js levantada con PM2, sin Docker.

**Mismo flujo que WordPress**: Docker se instala automáticamente.

**¿Qué respaldar?**
- `/home/app` o `/var/www/app` — Código fuente
- `/etc/nginx` — Configuración proxy
- `~/.pm2` — Configuración de PM2

**Limitación**: El explorador de contenedores en el panel estará vacío (sin Docker activo en el cliente). Para este caso, usar `Select All (Dynamic)` que tomará `/host_root` completo.

---

### ESCENARIO D: Token Expirado o No Usado 🔴
**Síntoma**:
```
❌ Error de Activación: Token expired or not found
```
**Causa**: Han pasado más de 48 horas desde que WHMCS generó el token y el cliente no instaló.

**Solución (Admin)**:
```bash
curl -X POST https://api.hwperu.com/v1/admin/license/generate \
  -H "X-Master-Token: TU_MASTER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plan": "enterprise", "service_id": "3070"}'
```

---

### ESCENARIO E: Cliente Pierde su Servidor (Disaster Recovery Real) 🆘
**Situación**: El VPS del cliente fue eliminado por accidente o hubo un fallo catastrófico del proveedor.

**Proceso de recuperación:**
1. El cliente contrata un nuevo VPS
2. Ejecuta el mismo comando de instalación (el token sigue vinculado a su cuenta WHMCS)
3. El sistema detecta huella diferente (hardware nuevo) → responde `re-activated`
4. El agente se conecta al API
5. El cliente abre el **Restore Wizard** → selecciona el snapshot más reciente → restaura todo
6. El agente descarga los datos desde Wasabi S3 y reconstruye la estructura de archivos

**RTO esperado**: Depende del tamaño de los datos (estimado en panel).

---

### ESCENARIO F: `jq` no Instalado (CentOS / Debian Mínimo) ⚠️
**Síntoma anterior** (resuelto desde V14.2.1):
```
❌ Error: Se requiere 'jq'. Instálalo e intenta de nuevo.
```
**Comportamiento actual** (auto-instalación):
```
⚙️  Instalando jq automáticamente...
```

---

### ESCENARIO G: Agente OFFLINE en el Panel
**Diagnóstico desde el VPS del cliente:**
```bash
# Ver estado del contenedor
docker ps | grep dbp-client-agent

# Ver logs recientes
docker logs dbp-client-agent --tail 50

# Reiniciar
cd /opt/docker-backup-pro && docker compose restart
```

**Causas comunes:**

| Causa | Solución |
|:---|:---|
| Contenedor caído | `docker compose up -d` |
| API no alcanzable | Verificar que el cliente puede llegar a `api.hwperu.com:443` |
| Token revocado | Generar nuevo token desde admin |
| Imagen desactualizada | `docker pull ghcr.io/hwperu/dbp-agent:prod` + restart |

---

## 10. ENDPOINTS DEL API (REFERENCIA TÉCNICA)

### Públicos
| Método | Endpoint | Función |
|:---|:---|:---|
| `GET` | `/ping` | Health check del API (`{"status":"ok","version":"V14.1.0"}`) |
| `GET` | `/install.sh` | Descarga el script instalador del agente |
| `POST` | `/v1/activate` | Activación de token + vinculación de hardware |
| `POST` | `/v1/auth/login` | Login al dashboard (retorna JWT) |

### Agente (Requieren X-Agent-ID + X-Agent-Key + X-Agent-Fingerprint)
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

### Admin (Requieren X-Admin-Key o MASTER_ADMIN_TOKEN)
| Método | Endpoint | Función |
|:---|:---|:---|
| `POST` | `/v1/whmcs/provision` | Provisiona un nuevo tenant desde WHMCS |
| `POST` | `/v1/admin/license/generate` | Genera/regenera token SaaS para un servicio |
| `POST` | `/v1/tenant/suspend` | Suspende todas las operaciones del tenant |
| `POST` | `/v1/tenant/unsuspend` | Reactiva el tenant |
| `POST` | `/v1/tenant/upgrade` | Cambia el plan del tenant |

---

## 11. FLUJO TÉCNICO COMPLETO (De Principio a Fin)

```
[1] ADMIN configura módulo WHMCS con API endpoint y tokens
              ↓
[2] CLIENTE compra el servicio en WHMCS
              ↓
[3] WHMCS llama a /v1/whmcs/provision
    API crea TenantPlan + UserSettings + AlertConfig
    Token: dbp_saas_{service_id}
    (Para Enterprise: pre-configura Protección Total automáticamente)
              ↓
[4] CLIENTE ve en área de cliente:
    curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token dbp_saas_XXX
              ↓
[5] CLIENTE ejecuta el comando en su VPS
    - Auto-instala jq si falta
    - Auto-instala Docker si falta
    - Genera fingerprint de hardware
    - Llama a /v1/activate con token + fingerprint
    - Recibe agent_id + api_key + ghcr_pat
    - Hace login a GHCR (imagen privada)
    - Levanta dbp-client-agent con docker compose
              ↓
[6] AGENTE arranca y cada 10 segundos:
    - Lee config del API (paths, schedule, protection_level, snapshot_mode)
    - Envía heartbeat con estado, contenedores, snapshots, versión
    - Ejecuta backup si toca según el horario o hay tarea pendiente
              ↓
[7] En hora programada:
    - ResolveBackupPaths() determina qué respaldar según protection_level
    - RunResticBackup() ejecuta con GlobalExcludes (sin proc, sys, dev, tmp)
    - ApplyRetentionPolicy() limpia snapshots viejos según plan
    - ReportMetrics() notifica resultado al Control Plane
    - UpdateHealthScore() recalcula salud del agente
    - DispatchAlert() envía webhook si hay fallo
              ↓
[8] CLIENTE visualiza en backup.hwperu.com:
    - Snapshots disponibles
    - Health Score
    - Historial de backups
    - Restore Wizard para recuperar datos
```

---

## 12. ESTADO ACTUAL DEL PROYECTO (V14.1.0)

| Módulo | Estado | Notas |
|:---|:---|:---|
| **Control Plane API** | ✅ Producción | Go V14.1.0 desplegado en `api.hwperu.com` |
| **Agente Docker** | ✅ Build local | Imagen lista para push a GHCR |
| **Dashboard UI** | ✅ Producción | Next.js en `backup.hwperu.com` (Vercel) |
| **Módulo WHMCS** | ✅ Producción | Integrado y probado |
| **Licenciamiento SaaS** | ✅ Producción | Tokens, fingerprint, 48h expiry |
| **Auto-install Docker** | ✅ Producción | `install.sh` V14.2.1 activo |
| **Snapshot Mode** | ✅ Código listo | `live` por defecto, `consistent` para Enterprise |
| **ResolveBackupPaths** | ✅ Código listo | Agente executor puro, no decide rutas |
| **GlobalExcludes prod-safe** | ✅ Código listo | `/proc`, `/sys`, `/dev`, `/tmp` excluidos |
| **Push a GHCR** | ⏳ Pendiente | Requiere login manual con GitHub PAT |
| **Redespliegue API** | ⏳ Pendiente | Para activar install.sh V14.2.1 en producción |

---

> [!IMPORTANT]
> **Próximo paso crítico**: Redesplegar la API en el servidor (`git pull` + `docker compose build` + `docker compose up -d`) y hacer push a GHCR (`docker push ghcr.io/hwperu/dbp-agent:prod`) para que todos los cambios de V14.1 estén activos en producción.

> [!TIP]
> **El diferencial comercial de HW Cloud Recovery**: El cliente no necesita saber nada de Docker, backups ni restic. Ejecuta un comando, y desde ese momento su servidor está protegido automáticamente. El panel le muestra en lenguaje de negocio (no técnico) si está "Protegido" o en "Riesgo".
