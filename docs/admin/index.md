# Instalación y Setup (Administrador)

## Arquitectura

```
┌─────────────────────────────────────────────┐
│              CONTROL PLANE                  │
│  api.hwperu.com (Go + Gin + PostgreSQL)     │
│  - Motor de Licencias SaaS                  │
│  - Policy Engine por Plan                   │
│  - Dispatcher de Alertas (Webhook)          │
│  - Endpoints REST para Agente y UI          │
└────────────────┬────────────────────────────┘
                 │ HTTPS / TLS 1.3
┌────────────────▼────────────────────────────┐
│              DATA PLANE (Cliente)           │
│  ghcr.io/hwperu/dbp-agent:prod (Alpine Go)  │
│  - Heartbeat cada 10 segundos               │
│  - Ejecuta backups según config del API     │
│  - Genera Fingerprint de Hardware (SHA256)  │
└────────────────┬────────────────────────────┘
                 │ S3 Protocol
┌────────────────▼────────────────────────────┐
│              STORAGE                        │
│  Wasabi S3 (o cualquier S3-compatible)      │
│  - Datos cifrados con Restic (AES-256)      │
│  - Retención configurable por plan          │
└─────────────────────────────────────────────┘
```

## Requisitos del Servidor API

- Servidor Linux con Docker + Docker Compose
- Puerto 443 (HTTPS) abierto
- Dominio apuntando al servidor (`api.hwperu.com`)
- Cuenta GitHub con PAT (permisos `write:packages`, `read:packages`)

## Variables de Entorno (`.env` en el servidor API)

```bash
# Conexión a Base de Datos
DB_HOST=postgres
DB_USER=dbp_user
DB_PASS=contraseña_segura
DB_NAME=dbp_production
DB_PORT=5432

# Seguridad
MASTER_ADMIN_TOKEN=token_maestro_para_whmcs
API_ADMIN_KEY=llave_admin_para_endpoints
DBP_ENCRYPTION_KEY=llave_aes_de_32_caracteres

# Distribución Privada (GHCR)
GHCR_READ_PAT=ghp_token_de_lectura_de_github

# Wasabi S3 Global (Fallback)
WASABI_ACCESS_KEY=...
WASABI_SECRET_KEY=...
WASABI_BUCKET=hwperu-backups
WASABI_REGION=us-east-1

# Alertas (Webhook Maestro)
WEBHOOK_URL=https://tu-n8n.com/webhook/xyz
```

## Despliegue de la API

```bash
# 1. Clonar repositorio privado
git clone https://github.com/soporte-hostingweb/dockerbackupprov.git
cd dockerbackupprov

# 2. Configurar variables de entorno
cp .env.example .env
nano .env

# 3. Levantar servicios
docker compose up -d

# 4. Verificar que está activo
curl https://api.hwperu.com/ping

# 5. Subir imagen del agente a GHCR
echo "TU_GITHUB_PAT" | docker login ghcr.io -u hwperu --password-stdin
docker build -t ghcr.io/hwperu/dbp-agent:prod ./agent/
docker push ghcr.io/hwperu/dbp-agent:prod
```
