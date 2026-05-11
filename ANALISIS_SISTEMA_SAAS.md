# 📊 Análisis del Sistema Backup SaaS (HW Cloud Recovery)

**Fecha de Análisis:** Mayo 2026  
**Versión Actual en Código:** V14.1.0 / V14.2.5  

Este documento resume el estado actual del servicio de respaldo SaaS de Docker Backup Pro, destacando funciones olvidadas o en conflicto, los componentes actuales, lo que falta por implementar y áreas de mejora en la Experiencia de Usuario (UX).

---

## 1. 🔍 Funciones Olvidadas y Conflictos en Documentación vs Código

Durante el análisis del código fuente y los múltiples archivos `.md`, se detectaron las siguientes discrepancias y "funciones zombie":

### A. Identidad del Agente (El conflicto del Fingerprint)
*   **Lo que decía la documentación antigua (V2.4):** El agente dependía de un archivo estático en `/host_root/etc/dbp_agent_id` para mantener la identidad del VPS.
*   **Lo que dice el código actual (V13/V14):** El código (`agent/utils.go` y `api/db.go`) abandonó ese método. Ahora utiliza un `Fingerprint` híbrido generado al vuelo mediante SHA256 combinando `machine-id`, `disk_serial` y `hostname`.
*   **Conclusión:** Toda referencia a la creación manual de identidades (`dbp_agent_id`) debe ser considerada obsoleta y fue eliminada de la documentación final para evitar confusiones.

### B. Verificación de Integridad (`restic check`)
*   **Lo que decía el FAQ Técnico:** Indicaba explícitamente que la verificación no estaba automatizada y dependía del lanzamiento manual por el administrador.
*   **Lo que hace el código actual:** En `agent/main.go` (Línea 341+), el agente ejecuta automáticamente la función `RunResticVerify()` y reporta el estado (`VALID` o `INVALID`) al finalizar un ciclo de backup para todos los planes superiores al básico (Standard y Enterprise).
*   **Conclusión:** La función ya está automatizada. El FAQ subestimaba las capacidades reales del sistema actual.

### C. Modo Rescue V2 (Bare-Metal Clone Pull-Based)
*   **Diseño Antiguo:** La documentación mencionaba una inyección de un agente vía SSH (`golang.org/x/crypto/ssh`), lo cual es un riesgo grave de seguridad y compromete el modelo "Zero-Cognitive Load".
*   **Nuevo Diseño Implementado (Pull-Based):** Ahora la web genera un **"Rescue Token"** temporal de un solo uso. El cliente recibe un comando `curl` que debe pegar en el nuevo servidor. Este script (Rescue Agent) descarga Restic, se autentica con el API, solicita en memoria las credenciales de S3 limitadas al snapshot y realiza la clonación bare-metal (`restic restore [SNAP] --target /`). Cero contraseñas SSH involucradas, 100% aislado y seguro.

---

## 2. 🏗️ Lo que se tiene actualmente

La arquitectura es sumamente sólida y cuenta con:
*   **API / Control Plane:** Desarrollado en Go 1.25, gestiona el licenciamiento (tokens que expiran en 48h), orquestación y Heartbeats (latidos) cada 10 segundos.
*   **Agent SaaS:** Un binario hiperligero basado en Docker (GHCR privado) que detecta automáticamente si el servidor usa Docker, Node.js, o WordPress tradicional, autoinstalando dependencias necesarias (como `jq` o Docker engine).
*   **Integración Comercial (WHMCS):** Hooks PHP que automatizan la creación del Tenant y entregan el comando `curl` de instalación al cliente al procesarse el pago.
*   **Seguridad:** Comunicaciones bajo TLS 1.3, credenciales Wasabi S3 cifradas en BD con AES-256-GCM, y deduplicación en repositorios segregados por cliente.
*   **Smart Awareness (V14.2.5):** Capacidad de detección automática de las rutas críticas del servidor sin requerir que el usuario sepa de Linux.

---

## 3. 🚧 Lo que falta (Brechas de Implementación)

*   **Asynchronous Job Worker (Mencionado en V12):** Actualmente, si 500 agentes reportan su Heartbeat o finalizan un backup simultáneamente, PostgreSQL y el servidor Go pueden saturarse por el pool de conexiones. Falta implementar RabbitMQ o Redis Streams para encolar la carga pesada (como reportes de telemetría y disparos de Webhooks).
*   **Post-Restore Validation:** El sistema valida que Restic finalizó la descompresión sin errores, pero no valida si el servicio web del cliente realmente "levantó" (ej. hacer un ping HTTP local al contenedor recién restaurado).
*   **Failover Multi-Cloud:** Todo depende de Wasabi S3. Si Wasabi experimenta latencia global, los backups fallan por timeout. Falta un Circuit Breaker real que cambie dinámicamente a R2 o AWS S3 como fallback en caso de emergencia.

---

## 4. 🎨 Mejoras de Experiencia de Usuario (UX)

Para que la plataforma pase de verse "técnica" a sentirse **Premium y Empresarial**, se sugieren las siguientes mejoras en el Dashboard de Next.js:

1.  **Lenguaje de Negocio, no de DevOps:** 
    *   *Actual:* "Restic repository check invalid".
    *   *Mejora UX:* "Detectamos una inconsistencia en el origen de los datos. Nuestro equipo ya fue notificado y está validando la copia de seguridad."
2.  **Micro-Animaciones en el Restore Wizard:** Al iniciar una restauración, evitar una barra de progreso estática. Mostrar una animación del flujo de datos (Nube -> Servidor) con estimaciones dinámicas de RTO (Recovery Time Objective) que tranquilizan al usuario en un momento de crisis.
3.  **Visualización del Health Score:** En lugar de un porcentaje crudo, implementar un velocímetro semántico o anillos de colores (Verde = Seguro, Amarillo = En Riesgo, Rojo = Desastre), acompañado de un resumen de "Último latido hace X segundos".
4.  **Botón de Pánico "1-Click Full Restore":** Oculto detrás de una confirmación modal robusta, que permita restaurar absolutamente todo el servidor a la noche anterior sin navegar por el explorador de archivos.
5.  **Avisos de Mantenimiento Transparentes:** Cuando el API pausa un servidor por operaciones internas (Modo Mantenimiento), la UI debe mostrarlo como "Sistema realizando optimizaciones de almacenamiento" y no como "Agente Desconectado/Error".
