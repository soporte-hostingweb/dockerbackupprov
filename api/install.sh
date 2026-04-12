#!/bin/bash
# ==========================================================
# 🚀 HW Cloud Recovery - SaaS Agent Installer (V14.2.0)
# Protegido por HWPeru Architecture (ghcr.io/hwperu/dbp-agent)
# ==========================================================

API_ENDPOINT="https://api.hwperu.com"
INSTALL_DIR="/opt/docker-backup-pro"
CONFIG_FILE="$INSTALL_DIR/agent.json"

echo "=========================================================="
echo "      HW CLOUD RECOVERY - SISTEMA DE ACTIVACIÓN SaaS      "
echo "=========================================================="

# 1. Verificar dependencias
for cmd in docker curl jq sha256sum; do
    if ! command -v $cmd &> /dev/null; then
        echo "❌ Error: Se requiere $cmd para continuar."
        exit 1
    fi
done

# 2. Generación de Huella Híbrida (V14: Anti-Clonado)
echo "[1/4] Generando Huella de Hardware..."
MACHINE_ID=$(cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || echo "unknown")
DISK_ID=$(lsblk -d -o SERIAL /dev/sda 2>/dev/null | tail -n 1 | xargs || blkid -s UUID -o value /dev/sda1 2>/dev/null || echo "none")
HOSTNAME=$(hostname)

FINGERPRINT=$(echo -n "$MACHINE_ID:$DISK_ID:$HOSTNAME" | sha256sum | awk '{print $1}')
echo "✅ Huella Generada: ${FINGERPRINT:0:12}..."

# 3. Handshake de Activación
echo ""
echo -n "🔑 Ingrese su Activation Token (dbp_saas_xxxx): "
read -r SAAS_TOKEN

if [[ ! $SAAS_TOKEN =~ ^dbp_saas_ ]]; then
    echo "❌ Error: Formato de token inválido. Debe empezar con dbp_saas_."
    exit 1
fi

echo "[2/4] Validando licencia con HWPeru Cloud..."
RESPONSE=$(curl -s -X POST "$API_ENDPOINT/v1/activate" \
    -H "Content-Type: application/json" \
    -d "{
        \"token\": \"$SAAS_TOKEN\",
        \"fingerprint\": \"$FINGERPRINT\",
        \"hostname\": \"$HOSTNAME\",
        \"os\": \"linux\"
    }")

STATUS=$(echo "$RESPONSE" | jq -r '.status')

if [ "$STATUS" != "activated" ] && [ "$STATUS" != "re-activated" ]; then
    ERR=$(echo "$RESPONSE" | jq -r '.error')
    echo "❌ Error de Activación: $ERR"
    exit 1
fi

AGENT_ID=$(echo "$RESPONSE" | jq -r '.agent_id')
API_KEY=$(echo "$RESPONSE" | jq -r '.api_key')
GHCR_PAT=$(echo "$RESPONSE" | jq -r '.ghcr_pat')

echo "✅ Activación Exitosa. AgentID: $AGENT_ID"

# 4. Preparación de Entorno y Pull Seguro (GHCR)
echo "[3/4] Autenticando con GitHub Container Registry..."
mkdir -p "$INSTALL_DIR"
echo "$GHCR_PAT" | docker login ghcr.io -u hwperu --password-stdin &> /dev/null

if [ $? -ne 0 ]; then
    echo "❌ Error: Falló la autenticación con GHCR. Contacte soporte."
    exit 1
fi

echo "[4/4] Descargando imagen privada (ghcr.io/hwperu/dbp-agent:prod)..."
docker pull ghcr.io/hwperu/dbp-agent:prod &> /dev/null

# Guardar credenciales persistentes
cat <<EOF > "$CONFIG_FILE"
{
  "agent_id": "$AGENT_ID",
  "api_key": "$API_KEY",
  "fingerprint": "$FINGERPRINT"
}
EOF

# Crear docker-compose.yml SaaS Ready
cat <<EOF > "$INSTALL_DIR/docker-compose.yml"
version: '3.8'
services:
  agent:
    image: ghcr.io/hwperu/dbp-agent:prod
    container_name: dbp-client-agent
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /:/host_root:rw
      - dbp-data:/app/data
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  dbp-data:
    driver: local
EOF

# Inyectar el agent.json en el volumen (Para que el agente lo lea al iniciar)
# Como docker v13 no tiene copy a volumenes facilmente, usamos un mount temporal o simplemente lo dejamos en el host y lo montamos
# Mejoramos el compose para montar el config directamente
sed -i 's|- dbp-data:/app/data|- '"$INSTALL_DIR"':/app/data|g' "$INSTALL_DIR/docker-compose.yml"

echo "🚀 Iniciando HW Cloud Recovery Agent..."
cd "$INSTALL_DIR" && docker compose up -d

echo "=========================================================="
echo "✨ INSTALACIÓN V14 COMPLETADA ✨"
echo "Estado: PROTEGIDO POR HWPERU SaaS"
echo "Huella vinculada correctamente."
echo "=========================================================="
