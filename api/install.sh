#!/bin/bash
# ==========================================================
# 🚀 HW Cloud Recovery - SaaS Agent Installer (V14.2.1)
# Distribución: ghcr.io/hwperu/dbp-agent:prod (Privado)
# ==========================================================

set -e

API_ENDPOINT="https://api.hwperu.com"
INSTALL_DIR="/opt/docker-backup-pro"
CONFIG_FILE="$INSTALL_DIR/agent.json"
GHCR_IMAGE="ghcr.io/hwperu/dbp-agent:prod"

echo "=========================================================="
echo "   🛡️  HW CLOUD RECOVERY - SISTEMA DE ACTIVACIÓN SaaS     "
echo "=========================================================="

# --- SOPORTE PARA TOKEN VIA PARÁMETRO ---
# Permite: curl -sSL .../install.sh | bash -s -- --token dbp_saas_XXX
SAAS_TOKEN=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --token) SAAS_TOKEN="$2"; shift 2;;
        *) shift;;
    esac
done

# 1. Verificar e instalar dependencias automáticamente
echo "[1/5] Verificando dependencias del sistema..."

# Auto-instalar jq si no está disponible
if ! command -v jq &> /dev/null; then
    echo "   ⚙️  Instalando jq automáticamente..."
    if command -v apt-get &> /dev/null; then
        apt-get update -qq && apt-get install -y -qq jq
    elif command -v yum &> /dev/null; then
        yum install -y -q jq
    elif command -v dnf &> /dev/null; then
        dnf install -y -q jq
    else
        echo "❌ No se pudo instalar jq automáticamente."
        echo "   Ejecuta manualmente: apt-get install jq (o yum install jq)"
        exit 1
    fi
fi

# Verificar e instalar Docker si no está disponible
if ! command -v docker &> /dev/null; then
    echo "   ⚙️  Docker no encontrado. Instalando automáticamente..."
    if ! command -v curl &> /dev/null; then
        apt-get install -y -qq curl 2>/dev/null || yum install -y curl 2>/dev/null
    fi
    curl -fsSL https://get.docker.com | sh
    if [ $? -ne 0 ]; then
        echo "❌ Error: No se pudo instalar Docker automáticamente."
        echo "   Instálalo manualmente: https://docs.docker.com/engine/install/"
        exit 1
    fi
    # Habilitar y arrancar Docker
    systemctl enable docker --now 2>/dev/null || service docker start 2>/dev/null
    echo "✅ Docker instalado correctamente"
fi

echo "✅ Dependencias OK"

# 2. Generación de Huella Híbrida (V14: Anti-Clonado)
echo "[2/5] Generando Huella de Hardware..."
MACHINE_ID=$(cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || echo "unknown")
DISK_ID=$(lsblk -d -o SERIAL /dev/sda 2>/dev/null | tail -n 1 | xargs 2>/dev/null || \
          blkid -s UUID -o value /dev/sda1 2>/dev/null || \
          cat /sys/class/dmi/id/product_uuid 2>/dev/null || echo "none")
HOSTNAME_VAL=$(hostname)

FINGERPRINT=$(echo -n "${MACHINE_ID}:${DISK_ID}:${HOSTNAME_VAL}" | sha256sum | awk '{print $1}')
echo "✅ Huella Generada: ${FINGERPRINT:0:12}..."

# 3. Solicitar token si no fue pasado por parámetro
echo ""
if [[ -z "$SAAS_TOKEN" ]]; then
    echo -n "🔑 Ingrese su Activation Token (dbp_saas_xxxx): "
    read -r SAAS_TOKEN
fi

if [[ ! $SAAS_TOKEN =~ ^dbp_saas_ ]]; then
    echo "❌ Error: Formato de token inválido. Debe empezar con 'dbp_saas_'."
    exit 1
fi

echo "[3/5] Validando licencia con HWPeru Cloud..."
RESPONSE=$(curl -s -X POST "$API_ENDPOINT/v1/activate" \
    -H "Content-Type: application/json" \
    -d "{
        \"token\": \"$SAAS_TOKEN\",
        \"fingerprint\": \"$FINGERPRINT\",
        \"hostname\": \"$HOSTNAME_VAL\",
        \"os\": \"linux\"
    }")

STATUS=$(echo "$RESPONSE" | jq -r '.status // "error"')

if [ "$STATUS" != "activated" ] && [ "$STATUS" != "re-activated" ]; then
    ERR=$(echo "$RESPONSE" | jq -r '.error // "Unknown error"')
    echo "❌ Error de Activación: $ERR"
    exit 1
fi

AGENT_ID=$(echo "$RESPONSE" | jq -r '.agent_id')
API_KEY=$(echo  "$RESPONSE" | jq -r '.api_key')
GHCR_PAT=$(echo "$RESPONSE" | jq -r '.ghcr_pat')

echo "✅ Activación Exitosa. AgentID: $AGENT_ID"

# 4. Autenticación GHCR y Descarga de Imagen Privada
echo "[4/5] Autenticando con GitHub Container Registry (GHCR)..."
mkdir -p "$INSTALL_DIR"

if [ -z "$GHCR_PAT" ] || [ "$GHCR_PAT" = "null" ]; then
    echo "❌ Error: No se recibió el token de acceso a GHCR. Contacte soporte."
    exit 1
fi

echo "$GHCR_PAT" | docker login ghcr.io -u hwperu --password-stdin
if [ $? -ne 0 ]; then
    echo "❌ Error: Falló la autenticación con GHCR."
    exit 1
fi

echo "   Descargando imagen: $GHCR_IMAGE ..."
docker pull "$GHCR_IMAGE"

# 5. Guardar credenciales y desplegar agente
echo "[5/5] Configurando y desplegando el agente..."

# Guardar credenciales persistentes
cat > "$CONFIG_FILE" <<EOF
{
  "agent_id": "$AGENT_ID",
  "api_key": "$API_KEY",
  "fingerprint": "$FINGERPRINT"
}
EOF
chmod 600 "$CONFIG_FILE"

# Crear docker-compose.yml (SaaS Ready, GHCR)
cat > "$INSTALL_DIR/docker-compose.yml" <<EOF
  # Watchtower: Sistema de Actualización Automática (V14.2+)
  # Monitorea cambios en GHCR y actualiza el agente sin intervención
  watchtower:
    image: containrrr/watchtower
    container_name: dbp-watchtower
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /root/.docker/config.json:/config.json:ro
    command: --interval 86400 --cleanup --include-stopped --include-restarting --revive-stopped dbp-client-agent
    depends_on:
      - agent
    logging:
      driver: "json-file"
      options:
        max-size: "5m"
        max-file: "2"
EOF

cd "$INSTALL_DIR" && docker compose up -d

echo ""
echo "=========================================================="
echo "✨  INSTALACIÓN V14 COMPLETADA  ✨"
echo ""
echo "  🛡️  Agente:     $AGENT_ID"
echo "  🖥️  Servidor:   $HOSTNAME_VAL"
echo "  🔒  Huella:     ${FINGERPRINT:0:16}..."
echo "  🐳  Imagen:     $GHCR_IMAGE"
echo ""
echo "  📊  Panel:      https://backup.hwperu.com"
echo "=========================================================="
