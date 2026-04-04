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
  NEXT_PUBLIC_API_URL=https://api.tuempresa.com/v1
  ```

### 3. Agente Cliente (Go / Restic SDK)
- **Infraestructura:** El VPS del cliente que contrató el servicio.
- **Variables de Entorno (Dentro del `docker-compose.yml` del cliente):**
  ```env
  DBP_API_TOKEN=EL_TOKEN_QUE_PROVEYÓ_TU_WHMCS
  DBP_API_ENDPOINT=https://api.tuempresa.com/v1
  AWS_ACCESS_KEY_ID=tu_wasabi_key
  AWS_SECRET_ACCESS_KEY=tu_wasabi_secret
  RESTIC_REPOSITORY=s3:s3.wasabisys.com/tu-bucket-general
  ```

---

## 🚀 Guía de Despliegue Extendido

A continuación, los comandos exactos a ejecutar como usuario `root` en un servidor Ubuntu 22.04 limpio para levantar la plataforma real.

### Paso 1: Levantar la API Central (Comandos Ubuntu 22.04)

**1. Instalar dependencias esenciales (Go y PostgreSQL):**
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y postgresql postgresql-contrib golang-go git nano curl
```

**2. Aprovisionar y Configurar la Base de Datos Relacional:**
```bash
sudo -u postgres psql -c "CREATE DATABASE dbp_data;"
sudo -u postgres psql -c "CREATE USER dbpuser WITH ENCRYPTED PASSWORD 'SecretoSeguro123';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE dbp_data TO dbpuser;"
```

**3. Descargar y Compilar la API de Go:**
```bash
mkdir -p /var/www/dbp/
cd /var/www/dbp/

# Clonar del repositorio (Requerirás ingresar credenciales si es privado):
git clone https://github.com/soporte-hostingweb/dockerbackupprov.git .

# Ingresar al componente API y Construir
cd api/
go mod tidy
go build -o dbp-api .
```

**4. Crear un Demonio (SystemD) para ejecución 24/7 en segundo plano:**
```bash
sudo nano /etc/systemd/system/dbp-api.service
```
*Pega el siguiente contenido en el editor nano, cuidando de que las rutas relativas sean exactas:*
```ini
[Unit]
Description=Docker Backup Pro API Central
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/dbp/api
ExecStart=/var/www/dbp/api/dbp-api
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
*Guarda los cambios (Ctrl+O, Enter) y cierra (Ctrl+X).*

**5. Encender y Activar Servicio API en Automático:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable dbp-api
sudo systemctl start dbp-api
sudo systemctl status dbp-api 
# Validar en los logs que dice "[BOOT] Server listening on port 8080..."
```

---

### Paso 2: Desplegar el Panel UI (Next.js con Node.js & PM2)

Ejecutaremos el panel web en el mismo Servidor VPS.

**1. Enganchar repositorio NodeSource (Node.js 20 Lts) e instalar PM2:**
```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
sudo npm install -g pm2
```

**2. Compilar de Desarrollo a Entorno Producción Optimizado:**
```bash
cd /var/www/dbp/ui/

# Instalar los paquetes Node
npm install

# Declarar e inyectar el entorno apuntando a nuestra API interna en el puerto 8080
echo 'NEXT_PUBLIC_API_URL=http://localhost:8080/v1' > .env.local

# Compilar frontend
npm run build
```

**3. Encender el Dashboard permanente usando PM2:**
```bash
# pm2 start tomará el mando de next.js
pm2 start npm --name "dbp-dashboard" -- run start

# Guardar estado actual de PM2 para resistir reinicios de servidor
pm2 startup
pm2 save
```
*Con este comando se lanzará el frontend internamente en el puerto `:3000` de forma persistente.*

*(Nota Administrativa: Para que tu interfaz web Next.js y el API estén expuestos de una forma amigable como `https://panel.hwperu.com`, debes colocar **NGINX** (proxyPass inverso) al puerto `3000` y `8080` de ese servidor acompañado de SSL con Certbot).*

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
