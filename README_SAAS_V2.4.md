# 🛡️ Docker Backup Pro - SaaS Stabilization (V2.4)

Este documento detalla la arquitectura de estabilización implementada para transformar **Docker Backup Pro** en una plataforma SaaS escalable y fiable, eliminando la volatilidad de identidades y facilitando el despliegue "Zero-Config".

---

## 🚀 Mejoras Clave

### 1. 🆔 Identidad Persistente (Host-Bound ID)
Anteriormente, el Agente usaba el ID del contenedor Docker como identificador único. Esto causaba que, al actualizar la imagen o reiniciar el servicio, el VPS perdiera su configuración de carpetas bloqueadas.

**Solución V2.4:**
*   El Agente ahora busca un archivo de identidad en `/host_root/etc/dbp_agent_id`.
*   Si no existe, lo genera una sola vez (basado en el hardware del host) y lo persiste.
*   **Beneficio:** Puedes borrar, actualizar o mover el contenedor y el Dashboard siempre reconocerá al mismo servidor con sus backups intactos.

### 2. 🔑 Handover de Credenciales Seguro
Eliminamos la necesidad de configurar manualmente el Agente con variables de entorno de Wasabi o contraseñas de Restic.

**Lógica de Entrega:**
1.  El Administrador configura el Bucket y el Password en el Dashboard (Tab: Settings).
2.  La API guarda estos datos **cifrados con AES-256-GCM**.
3.  Durante el Heartbeat (HTTPS), el Agente recibe su `full_repo_url` y su `restic_password` en un payload seguro.
4.  El Agente inyecta estas credenciales dinámicamente solo durante la ejecución del proceso restic.

### 3. 🌐 Aislamiento Dinámico de Repositorios
Cada cliente y cada VPS tiene un repositorio físico aislado en Wasabi:
`s3:region.wasabisys.com/bucket/tenant_token/persistent_agent_id`

---

## 🛠️ Guía de Actualización

### Para el Administrador (API):
1. Hacer `git pull origin master`.
2. Ejecutar `docker compose up -d --build dbp-api`.
3. **IMPORTANTE:** Ir a la pestaña "Settings" en el Dashboard y completar los datos de Wasabi por primera vez.

### Para el Cliente (Agente):
1. El script de instalación automática (`install.sh`) ya apunta a la última versión.
2. Para VPS existentes: `docker compose pull agent && docker compose up -d agent`.

---

## 📈 Roadmap de Seguridad
* [v2.4] Cifrado AES-GCM en Base de Datos.
* [v2.4] Identidad basada en Machine-ID.
* [v2.5] Soporte para cuotas de disco por Tenant.
* [v2.6] Notificaciones vía WhatsApp/Email sobre fallos de backup.

---
© 2026 HWPeru - SaaS Backup Infrastructure.
