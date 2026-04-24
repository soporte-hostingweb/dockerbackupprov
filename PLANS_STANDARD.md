# 🛡️ DBP SaaS - PLAN STANDARD (Business Choice)

El Plan Standard es la solución ideal para pequeñas y medianas empresas (PyMEs) que operan sitios de comercio electrónico o aplicaciones web que requieren automatización diaria.

## 🚀 Características del Plan
- **Almacenamiento:** 100 GB en Wasabi S3 (Protección Triple).
- **Retención:** 7 días (Rotación automática).
- **Snapshot Mode:** Live (Zero-Downtime) + Smart SQL Backup.
- **Frecuencia:** Copia Diaria automática (02:00 AM).
- **Seguridad:** Cifrado AES-256 en tránsito y reposo.

---

## 📂 Casos de Uso y Ejemplos de Clientes

### Ejemplo 1: Cliente con Docker (Tienda WooCommerce)
**Perfil:** Una tienda online con WordPress y MySQL corriendo en contenedores separados.
- **Configuración:** `wp-content` en un volumen y la DB en otro.
- **Estrategia:** Respaldo de archivos y volcado SQL automático vía socket.

**Guía Rápida de Instalación:**
1. Instalar el agente con el Token Standard.
2. En el Dashboard -> **Configuración de Base de Datos**:
   - Host: `localhost` (El agente detectará el contenedor).
   - User: `db_user`.
   - Databases: `["woo_production"]`.
3. En el Explorer, marcar la carpeta de medios de WordPress.
4. Programar en **"Estándar (Backup Diario)"**.

### Ejemplo 2: Instalación Pura (WordPress Bare-Metal)
**Perfil:** Un sitio WordPress instalado tradicionalmente sobre Ubuntu + Apache + MySQL (sin Docker).
- **Configuración:** `/var/www/html` y MySQL 8.0 local.
- **Estrategia:** Hardening activo (el agente usa su `mysqldump` nativo).

**Guía Rápida de Instalación:**
1. Ejecutar el comando de instalación en el servidor.
2. Configurar la sección SQL del Dashboard con el usuario `root` y la contraseña local.
3. Marcar `/host_root/var/www/html` en el explorador.
4. Guardar y activar el cron automático.

### Ejemplo 3: Panel WHM (Múltiples cuentas)
**Perfil:** Un revendedor de hosting con 10 cuentas de cPanel en un servidor.
- **Configuración:** `/home` particionado por usuarios.
- **Estrategia:** Respaldo selectivo de los directorios de usuario más críticos.

**Guía Rápida de Instalación:**
1. Instalar a nivel de root en el servidor WHM.
2. Usar el explorer para navegar por `/host_root/home`.
3. Marcar las carpetas de los clientes Premium.
4. Configurar el backup para que ocurra a las 3 AM (fuera de pico de tráfico).

---

> [!TIP]
> El Plan Standard permite la recuperación de archivos individuales a través del **Restore Wizard**, ideal para accidentes comunes como el borrado de una imagen.
