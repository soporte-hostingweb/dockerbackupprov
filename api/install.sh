#!/bin/bash
# ==========================================
# Docker Backup Pro - Auto Install Agent Script
# Property of HWPeru (api.hwperu.com)
# ==========================================

TOKEN=""
API_ENDPOINT="https://api.hwperu.com"


# Extraer los argumentos del CLI
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --token) TOKEN="$2"; shift ;;
        *) echo "Parametro desconocido: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$TOKEN" ]; then
    echo "ERROR CRITICO: El token de instalación es obligatorio."
    echo "Ejemplo: curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token XYZ123"
    exit 1
fi

echo "=========================================="
echo "Instalando Docker Backup Agent - HWPeru..."
echo "=========================================="

mkdir -p /opt/docker-backup-pro
cd /opt/docker-backup-pro

# Creamos el Docker-compose al vuelo asimilando su Token único de WHMCS
cat <<EOF > docker-compose.yml
version: '3.3'
services:
  agent:
    # La próxima vez que subas esto a la nube pública usa josebenites21/dbp-agent
    image: josebenites21/dbp-agent:latest
    container_name: dbp-client-agent
    restart: always
    environment:
      - DBP_API_TOKEN=${TOKEN}
      - DBP_API_ENDPOINT=${API_ENDPOINT}
      # Aqui el cliente debería modificar AWS_ACCESS_KEY_ID u otra configuración
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /:/host_root:rw # V6.9: Cambiado a RW para permitir restauraciones in-place
      # Persistencia de identidad (V2.4.3)
      - dbp-agent-data:/app/data

volumes:
  dbp-agent-data:
EOF

echo "[+] Levantando contenedor en $HOSTNAME..."

if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

echo "=========================================="
echo "🚀 ¡Instalación Exitosa! 🚀"
echo "El Agente ya se conectó a api.hwperu.com"
echo "Puedes revisar el estado en https://portal.hwperu.com"
echo "=========================================="
