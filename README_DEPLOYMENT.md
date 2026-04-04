# Docker Backup Pro (DBP) - Master Deployment Guide

Este documento contiene las instrucciones fundamentales para ingenieros y personal DevOps sobre cómo, dónde y con qué requisitos desplegar la plataforma completa en producción de Docker Backup Pro.

---

## 🏗 Topología: ¿Dónde se instala cada componente?

El proyecto consta de 4 componentes que se distribuyen geográficamente. Es **crucial** entender que NO se instalan todos en el mismo servidor:

1. **`api/` (Control Plane):** Debe instalarse en un **Servidor Central Propio** dedicado a la empresa (recomendado un VPS Linux limpio con al menos 2GB de RAM).
2. **`ui/` (Dashboard Web):** Puede instalarse en el mismo Servidor Central junto a la API, o de manera más eficiente desplegarse en infraestructuras orientadas a frontend como **Vercel** o **Render**.
3. **`whmcs/` (Facturador):** Se aloja exclusivamente de manera interna tu instalación actual de WHMCS.
4. **`agent/` (Cliente recolector):** No lo hospedas tú. Se instalará **Múltiples veces**. Una ejecución dentro de contenedor por cada VPS de tu cliente que requiera protección de respaldos.

---

## 🛠 Requisitos, Dependencias y Configuración por Módulo

### 1. API Central (Backend en Go)
- **Infraestructura:** Ubuntu 22.04 LTS (o similar).
- **Dependencias base:** Servidor PostgreSQL 14+, Lenguaje Go v1.21+. Nginx o Caddy como Proxy Inverso.
- **Ruta Recomendada:** `/var/www/dbp-api/`
- **Campos a rellenar (Variables de Entorno `api/.env`):**
  ```env
  # Credenciales de tu BD relacional
  DB_HOST=localhost
  DB_USER=postgres
  DB_PASS=SecretoSeguro123
  DB_NAME=dbp_data
  API_PORT=8080
  JWT_SECRET=tu_llave_maestra_jwt_para_firmar_tokens
  ```

### 2. UI Dashboard (Frontend en TypeScript/Next.js)
- **Infraestructura:** VPS con Linux o Vercel Platform.
- **Dependencias base:** Node.js v18 (ó 20+), Git.
- **Campos a rellenar (Variables de Entorno `ui/.env.local`):**
  ```env
  NEXT_PUBLIC_API_URL=https://api.tuempresa.com/v1
  ```

### 3. Módulo WHMCS (PHP)
- **Infraestructura:** Hosting cPanel o VPS con WHMCS v8+.
- **Ruta de instalación estricta:** `[TUDIRECTORIO_WHMCS]/modules/servers/dockerbackuppro/`
- **Campos a rellenar:** Dentro de la pestaña de producto en WHMCS, deberás definir el "Endpoint" que generaste en el backend, por ejemplo `https://api.tuempresa.com/v1` y las cuotas de subida en Gigabytes (ej. `50`).

### 4. Agente Cliente (Go / Restic SDK)
- **Infraestructura:** El VPS del cliente que contrató el servicio.
- **Dependencias base:** Solamente Docker Desktop o Docker Engine (19.03+). _No requiere instalar base de datos locales ya que usa recursos de docker alpine_.
- **Campos a rellenar por cliente (Dentro del `docker-compose.yml` que recibe el cliente):**
  ```env
  DBP_API_TOKEN=EL_TOKEN_QUE_PROVEYÓ_TU_WHMCS
  DBP_API_ENDPOINT=https://api.tuempresa.com/v1
  
  # Credenciales Maestras de Wasabi S3
  AWS_ACCESS_KEY_ID=tu_wasabi_key
  AWS_SECRET_ACCESS_KEY=tu_wasabi_secret
  RESTIC_REPOSITORY=s3:s3.wasabisys.com/tu-bucket-general
  ```

---

## 🚀 Pasos Secuenciales para Puesta en Marcha (Instalación)

### Paso 1: Compilar y Levantar la API
1. Conéctate a tu Servidor Central por SSH.
2. Descarga la carpeta `/api` construida previamente.
3. Instala PostgreSQL: `apt install postgresql` e inicializa el usuario y la BD `dbp_data`.
4. Compila y ejecuta la API usando Go:
   ```bash
   cd api/
   go mod tidy
   go build -o dbp-api .
   ./dbp-api
   ```
   *(Recomendación avanzada: Configurar un archivo de servicio SystemD para este binario, de forma que se ejecute siempre en segundo plano e inicie automáticamente en reinicios).*

### Paso 2: Desplegar el Panel UI (Web)
1. Toma la carpeta `/ui` y súbela a un repositorio en Github Privado.
2. Accede a **Vercel.com**, y presiona "Import from Github". Vercel auto-configurará los comandos (`npm install` y `npm run build`), por lo que no requerirás mantener un servidor frontend extra.
3. Agrégale solamente la variable de entorno `NEXT_PUBLIC_API_URL` apuntando a la IP pública o Dominio validado de tu API del Paso 1.

### Paso 3: Configurar el Facturador Oculto en WHMCS
1. Vía SFTP, copia todo el contenido del código entregado en `/whmcs` y suéltalo dentro de la carpeta `modules/servers/` de tu sitio de WHMCS.
2. Inicia sesión como Staff en tu WHMCS y dirígete a `Ajustes Reales > Productos y Servicios`.
3. Crea un ítem de Venta nuevo (Servicio) con modelo de Cobro a elegir. En "Configuración del Módulo", selecciona **Docker Backup Pro**.
4. Haz una orden desde tú mismo a modo de prueba para confirmar la visualización del Template ClientArea (`clientarea.tpl`).

### Paso 4: Realizar Prueba Extremo a Extremo en un VPS Cliente Real
Una vez activo el facturador y el endpoint API, dirígete a un servidor de entorno de pruebas con Docker.
Crea la carpeta y el archivo Compose para evaluar al Agente y lánzalo:
   ```bash
   mkdir -p /opt/docker-backup-pro && cd /opt/docker-backup-pro
   nano docker-compose.yml 
   # Copia la plantilla que dimos en la Guía de Instalaciones con un Token de prueba válido 
   
   docker-compose up -d
   ```
Verás como los logs del Agente inician el escaneo, ubican las carpetas y hacen un POST exitoso rebotado por el Middleware de Autenticación de tu API Central que completamos en la fase 3.
