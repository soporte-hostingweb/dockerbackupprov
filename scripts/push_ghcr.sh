#!/bin/bash
# ==========================================================
# Script para push de imagen a GHCR (GitHub Container Registry)
# Ejecutar manualmente en la terminal del equipo de desarrollo
# ==========================================================

# PASO 1: Login con tu PAT de GitHub (necesita permisos: write:packages)
# Reemplaza TU_GITHUB_PAT con tu Personal Access Token
echo "TU_GITHUB_PAT" | docker login ghcr.io -u hwperu --password-stdin

# PASO 2: Push de imagen de producción
docker push ghcr.io/hwperu/dbp-agent:prod

# PASO 3: Push del tag de versión
docker push ghcr.io/hwperu/dbp-agent:v14.1.0

echo "✅ Push completado. Verifica en: https://github.com/orgs/hwperu/packages"
