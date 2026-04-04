# Docker Backup Pro (DBP) - Master Deployment Guide

Este documento contiene las instrucciones fundamentales para ingenieros y personal DevOps sobre cómo, dónde y con qué requisitos desplegar la plataforma completa en producción de Docker Backup Pro.

---

## 🏗 Topología: ¿Dónde se instala cada componente?

El proyecto consta de 4 componentes que se distribuyen geográficamente. Es **crucial** entender que NO se instalan todos en el mismo servidor:

1. **`api/` (Control Plane):** Debe instalarse en un **Servidor Central Propio** dedicado a la empresa.
2. **`ui/` (Dashboard Web):** Puede instalarse en el mismo Servidor Central junto a la API o en un VPS frontend.
3. **`whmcs/` (Facturador):** Se aloja exclusivamente de manera interna tu instalación actual de WHMCS.
4. **`agent/` (Cliente recolector):** Se instalará **Múltiples veces**. Una ejecución dentro de contenedor por cada VPS de tu cliente que requiera protección de respaldos.

---

## 🛠 Requisitos, Dependencias y Configuración por Módulo

### 1. API Central (Backend en Go)
- **Infraestructura:** Ubuntu 22.04 LTS.
- **Ruta Recomendada:** `/var/www/dbp-api/`
- **Variables de Entorno (`api/.env`):**
  ```env
  DB_HOST=localhost
  DB_USER=dbpuser
  DB_PASS=SecretoSeguro123
  DB_NAME=dbp_data
  API_PORT=8080
  ```

### 2. UI Dashboard (Frontend en TypeScript/Next.js)
- **Infraestructura:** VPS con Linux (Mismo servidor que la API, o diferente).
- **Ruta Recomendada:** `/var/www/dbp-ui/`
- **Variables de Entorno (`ui/.env.local`):**
  ```env
  NEXT_PUBLIC_API_URL=https://api.hwperu.com/v1
  ```

### 3. Agente Cliente (Go / Restic SDK)
- **Infraestructura:** El VPS del cliente que contrató el servicio.
- **Variables de Entorno (Dentro del `docker-compose.yml` del cliente):**
  ```env
  DBP_API_TOKEN=EL_TOKEN_QUE_PROVEYÓ_TU_WHMCS
  DBP_API_ENDPOINT=https://api.hwperu.com/v1
  AWS_ACCESS_KEY_ID=tu_wasabi_key
  AWS_SECRET_ACCESS_KEY=tu_wasabi_secret
  RESTIC_REPOSITORY=s3:s3.wasabisys.com/tu-bucket-general
  ```

---

## 🚀 Guía de Despliegue Extendido (Vía Docker & Vercel)

A continuación explicaremos el despliegue moderno utilizando **Docker Compose** para orquestar la API y la Base de datos en un solo movimiento limpio, y enviando el Dashboard al CDN global de Vercel.

### Paso 1: Levantar la API Central (Con Docker)

En tu Servidor Ubuntu 22.04 limpio, asegúrate de tener instalado `docker` y `docker-compose`.

**1. Clonar el Repositorio Central y Configurar Variables:**
```bash
git clone https://github.com/soporte-hostingweb/dockerbackupprov.git /opt/docker-backup-pro
cd /opt/docker-backup-pro

# Clonar la plantilla de variables de entorno (NUNCA SUBIR EL .env REAL A GITHUB)
cp .env.example .env
nano .env 
# Personaliza el valor de DB_PASS y JWT_SECRET.
```

**2. Desplegar el Motor Backend (API + PostgreSQL):**
```bash
docker-compose up -d --build
```
*Este único comando hará todo: Descargar Postgres, crear su volumen interno, compilar tu código de Go de la API nativamente en un contenedor Alpine minimalista y conectar ambas existencias dentro de su propia red inviolable.*

**3. Revisar Estado Local de los Servicios:**
```bash
docker-compose ps
docker logs -f dbp-api
```
*(Asegúrate de que la API muestre `[BOOT] Server listening on port 8080...`)*

---

### Paso 2: Desplegar el Panel UI (En Vercel.com)

Esta es la forma más profesional y costo-efectiva. Tu panel correrá sin mantenimiento de servidor.

**1. En Vercel:**
- Conéctate a Vercel y enlaza tu repositorio Github (`dockerbackupprov`).
- Durante la configuración de importación, establece el **Framework Preset** como `Next.js` y asegúrate de que el **Root Directory** a compilar sea: `ui`.

**2. Variables de Entorno en Vercel:**
- Antes de dar click en Deploy, anda a *Environment Variables* y pega la que tienes en tu `.env.example`:
  - `NEXT_PUBLIC_API_URL` apuntando a tu dominio final (ej. `https://api.backup.hwperu.com/v1`).

**3. Despliegue Automático:**
- Haz clic en Deploy. Vercel se encargará de cualquier validación (no es necesario levantar el NodeJS localmente). A partir de ahora, cada "Push" actualizará tu panel comercial de forma totalmente desatendida.

---

### Paso 3: Aprovisionamiento Físico del Módulo en WHMCS

**1. Instalar la carpeta base (Vía cPanel / SFTP / SSH):**
- Debes conectarte al Cpanel/Servidor de la página donde tienes WHMCS (Usualmente distinta al Servidor Central Dockerbackup).
- Navega dentro de tus archivos de servidor Web a la ruta: `[Tu Carpeta WHMCS]/modules/servers/`.
- Crea una carpeta o arrastra la carpeta `dockerbackuppro` allí, cerciorándote de que la ruta final quede así:
  `[Tu Carpeta WHMCS]/modules/servers/dockerbackuppro/dockerbackuppro.php`

**2. Mapear el Producto en tu Sistema Operativo WHMCS:**
- Inicia sesión como Admnistrador a tu panel de Facturación WHMCS.
- Accede a la tuerca superior: **Ajustes > Productos y Servicios > Productos y Servicios**.
- Haz clic en **"Crear Nuevo Producto"**.
- Tipo de Producto: **Un Servicio (Other/Service)**.
- Nombre: Ej. *SaaS Docker Backup Pro VIP*.
- Grupo de Producto: A elección (P. ej: Añadidos a VPS).

**3. Activar el Motor Go/WHMCS Interno:**
- Abre el producto recién creado, y posicionate en la Tab Opciones: **Configuración del Módulo (Module Settings)**.
- En el desplegable *Nombre del Módulo* selecciona `"Docker Backup Pro"`.
- Automáticamente el PHP de tu código mostrará 2 nuevos campos rellenables:
  - *Storage Quota GB*: `50` (O lo que definas para ese paquete).
  - *API Endpoint*: Ingresa aquí la IP de tu VPS Central con su puerto (ej. `http://5.1.4.12:8080/v1`) o el subdominio si configuraste proxy (`https://api.backups.hwperu.com/v1`).
- Presiona la opción **"Configurar cuenta automáticamente tan pronto como se reciba el primer pago."**
- Ve y hazle un pedido _Checkout_ como un cliente interno falso en valor de $0USD, para comprobar como WHMCS le muestra el iFrame, Token de Acceso y Comando de instalación Linux en su respectiva *Área de Cliente*.
