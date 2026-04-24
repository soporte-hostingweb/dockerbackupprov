# 🎁 DBP SaaS - PLAN FREE (Community Edition)

El Plan Free está diseñado para micro-servicios, blogs personales y entornos de desarrollo que requieren protección básica sin costo mensual.

## 🚀 Características del Plan
- **Almacenamiento:** 1 GB en Wasabi S3 (Georeplicado).
- **Retención:** 1 punto de restauración (Manual).
- **Snapshot Mode:** Live (Zero-Downtime).
- **Soporte:** Comunitario.

---

## 📂 Casos de Uso y Ejemplos de Clientes

### Ejemplo 1: Cliente con Docker (Blog Ghost)
**Perfil:** Un desarrollador que corre su blog personal usando un contenedor de Ghost.
- **Configuración:** El contenedor guarda imágenes en `/var/lib/ghost/content`.
- **Estrategia:** Solo se respalda la carpeta de contenido para no exceder el 1 GB.

**Guía Rápida de Instalación:**
1. Obtener Token SaaS desde el panel (ej: `dbp_saas_free_xxxx`).
2. Ejecutar instalador:
   ```bash
   curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token dbp_saas_free_xxxx
   ```
3. En el Dashboard, expandir el servidor y seleccionar la ruta `/var/lib/docker/volumes/ghost_data/_data`.
4. Hacer clic en **"Force Selected"** para el primer backup manual.

### Ejemplo 2: Instalación Pura (Bare-Metal Nginx)
**Perfil:** Un sitio estático alojado directamente en el sistema de archivos de un VPS pequeño (512MB RAM).
- **Configuración:** Archivos HTML en `/var/www/personal-site`.
- **Estrategia:** Backup recursivo de la ruta web.

**Guía Rápida de Instalación:**
1. Instalar el agente usando el comando oficial con el Token.
2. Abrir el Dashboard -> Explorer.
3. Navegar hasta `/host_root/var/www/personal-site`.
4. Guardar configuración y ejecutar backup.

### Ejemplo 3: Panel cPanel (Cuenta Individual)
**Perfil:** Un usuario con una cuenta de cPanel que quiere un respaldo externo de su `public_html`.
- **Configuración:** Usuario final con acceso SSH limitado.
- **Estrategia:** Respaldo de archivos estáticos.

**Guía Rápida de Instalación:**
1. Solicitar acceso SSH al administrador o ejecutar vía terminal de cPanel.
2. Descargar e instalar el agente.
3. En el explorer del Dashboard, marcar `/host_root/home/usuario/public_html`.
4. Ejecutar backup manual semanalmente.

---

> [!NOTE]
> En el Plan Free, el usuario debe disparar la copia manualmente desde el panel para asegurar su persistencia.
