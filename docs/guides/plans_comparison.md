# HW Cloud Recovery - Especificación Técnica de Planes (V14.3.6)

Este documento detalla las funcionalidades comprobadas y validadas en la infraestructura SaaS actual tras el endurecimiento del sistema (Hardening Phase).

## 🚀 Plan STANDARD (Validado)
El motor base de recuperación para entornos Docker estándar.

### Funcionalidades Comprobadas:
*   **Backup Asíncrono de Contenedores**: Capacidad de realizar copias de seguridad sin detener servicios mediante el motor Restic.
*   **Gestión de Puntos de Restauración**: Retención automática de las últimas 7 copias (Policy: KEEP LAST 7).
*   **Explorer de Datos Básico**: Visualización de archivos dentro de contenedores en tiempo real.
*   **Conexión a S3/Wasabi Estándar**: Soporte para buckets en regiones oficiales de Wasabi y AWS.
*   **Wizard de Restauración V5.0**: Interfaz paso a paso para recuperación de archivos específicos.

---

## 💎 Plan PRO (Validado - Hardened)
Diseñado para infraestructuras complejas, nubes privadas y entornos de alta seguridad.

### Funcionalidades Exclusivas Comprobadas:
*   **Universal S3 Compatibility (BYOS)**: Soporte completo para **Minio**, nubes privadas y proveedores S3 locales.
*   **Intelligent Protocol Detection**: El sistema conmuta automáticamente entre **HTTP y HTTPS** basado en el puerto y la configuración de seguridad, evitando errores de handshake.
*   **S3 Insecure Mode (Skip SSL)**: Permite conectar con servidores de almacenamiento que usan certificados autofirmados o no tienen SSL (crucial para entornos de red interna).
*   **Force Path Style Addressing**: Compatibilidad con arquitecturas S3 antiguas o privadas donde el bucket forma parte de la ruta y no del subdominio.
*   **Real-time Connection Intelligence**: Notificaciones dinámicas en el panel sobre latencia y estado de conectividad sin recarga de página.
*   **Multi-Agent Tier 2 Recovery**: Motor de auto-reinicio de agentes locales si se detecta pérdida de latencia masiva.

---

## 🛡️ Funciones de Seguridad Global (Comprobadas)
*   **PostgreSQL Persistence**: Persistencia de estados booleanos (`true`/`false`) garantizada incluso en migraciones de esquema complejas.
*   **Async Heartbeat**: Monitoreo de salud del agente con RTT (Round Trip Time) visible en el dashboard.
*   **Atomic Selection Lock**: Sistema de bloqueo de selección de archivos para garantizar integridad en el envío de tareas al agente.
