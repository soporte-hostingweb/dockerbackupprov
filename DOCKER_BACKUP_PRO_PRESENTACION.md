# 🚀 Informe de Arquitectura y Negocio: Docker Backup Pro

**Destinatario:** Arquitecto de TI / Dirección Técnica
**Documento:** Presentación de Funcionalidad y Factibilidad Comercial
**Versión Actual del Sistema:** V8.1 (Enterprise Ready)

## 1. Resumen Ejecutivo
**Docker Backup Pro (DBP)** es una solución SaaS de orquestación de respaldos diseñada específicamente para entornos contenedorizados (Docker). Su arquitectura elimina la necesidad de intervenciones manuales en los servidores Linux/Ubuntu de los clientes, automatizando la extracción, cifrado y subida de datos directamente a un almacenamiento S3 inmutable (Wasabi). El sistema está diseñado para integrarse nativamente con WHMCS, permitiendo automatizar la venta, el alta de clientes (Aislamiento Multi-Tenant) y la facturación, sin sacrificar la seguridad de los datos.

---

## 2. Flujo de Trabajo del Sistema (Workflow)

El ecosistema se divide en 3 componentes críticos que interactúan bajo un modelo asíncrono y de cero impacto en producción:

1. **El Agente (DBP Agent):** Un contenedor ultraligero que se despliega en el VPS del cliente. Tiene acceso al socket de Docker y a los volúmenes en puente (`/host_root`). No expone puertos. Evalúa cronogramas e inyecta los datos cifrados directamente a Wasabi de forma unidireccional.
2. **El Control Plane (API Orquestadora):** Servidor central desarrollado en Go. Monitorea latidos (Heartbeats) cada 10 segundos, gestiona las políticas de permisos, inyecta credenciales KMS y consolida las métricas de uso de Wasabi de todos los *Tenants*.
3. **Frontend (Dashboard UI):** Aplicación React (Next.js) que consume el Control Plane. Ofrece una vista visual para explorar rutas, restaurar archivos (File Explorer interactivo) e iniciar clonaciones Bare-Metal a nuevos servidores.

### ¿Por qué usar este sistema frente a Restic puro o scripts Bash?
* **Telemetría Central: ** Si un script bash falla en un servidor remoto, usted no se entera. DBP alerta en el dashboard si un agente está *OFFLINE* o falla.
* **Seguridad (Criptografía):** Las llaves KMS de Wasabi no están en el VPS del cliente. El API las pasa mediante tokens condicionales. 
* **Gestión Unificada:** El cliente gestiona sus copias sin saber usar la consola SSH ni comprender Docker.

---

## 3. Funcionalidad y Ejemplos de Uso (Snapshots)

Docker Backup Pro es inteligente al diferenciar cómo resguarda la información basándose en la configuración del cliente.

### A. Modalidad "Select All Dinámico" (Respaldo Continuo)
**Caso de uso:** Un cliente instala WordPress con Docker, pero constantemente añade nuevos subdirectorios, plugins (volúmenes de Docker) u otras configuraciones.
**Funcionamiento:** 
El usuario da clic en `Select All (Dynamic)` en la UI. El sistema intercepta el **puente de volumen (Host Bridge)** del contenedor. 
* **Beneficio:** Incluso si el cliente crea nuevas carpetas por SSH dentro de ese contenedor mañana, el Agente las rastreará automáticamente sin necesidad de que el cliente vuelva al panel a seleccionarlas. Garantiza que el hilo de backups nunca se rompa.

### B. Modalidad "Custom Select" (Archivos Específicos)
**Caso de uso:** Un servidor con bases de datos MySQL gigantes donde solo importan los volcados diarios y la carpeta `/var/lib/mysql`, ignorando cachés y logs.
**Funcionamiento:** 
El usuario navega vía *File Explorer* y marca una a una las carpetas estáticas a resguardar. Ahorra gigabytes de almacenamiento y procesado mensual.

### C. Snapshot Completo + Bare-Metal Restore (V8.0)
**Caso de uso:** Desastre total. El VPS del cliente fue hackeado, borrado u ocurió una falla catastrófica de hardware en su data center.
**Funcionamiento:**
1. El Administrador enlista un nuevo VPS vacío y despliega una instancia de Ubuntu pura.
2. En el Dashboard, selecciona **Clone to New VPS (Bare-Metal)**.
3. Se inyecta la IP, puerto (22) y el password Root.
4. El Control Plane se conecta silenciosamente vía SSH al nuevo VPS, inyecta un agente de rescate ciego que conecta con Wasabi y vuelca todo el *"Full Snapshot"*. 
5. El sistema se reinicia automáticamente recuperando absolutamente toda la vida del sistema origin. 
*(Zero Impact: Se descarga directamente desde Wasabi -> Nuevo Servidor, sin consumir recursos del servidor origen si este todavía existiera).*

---

## 4. Evolución Tecnológica (Versiones Hito)

* **V3.0:** Consolidación de Telemetría UDP/HTTP y descubrimiento de contenedores en tiempo real.
* **V4.0:** Consolidación de Restic y Deduplicación. Almacenaje de múltiples estados de tiempo consumiendo poco espacio.
* **V5.0/V6.0:** Arquitectura Serverless UI y Monitor de Actividades con Auditoría Global en tiempo real.
* **V7.0/V7.5:** Aislamiento Multi-Tenant (Segmentación criptográfica por cliente en S3) y Scheduler Adaptativo Multinacional (Respaldo por Zona Horaria).
* **V8.0/V8.1:** Integración de "Restore V2 Bare-Metal Orchestrator" e Internacionalización interna (i18n). Sistema que traduce en GB el tamaño exacto consumido desde el S3.

---

## 5. Planes de Negocio y Finanzas (Pricing Strategy)

Dado que Wasabi cobra aproximadamente **$6.99 USD por TB/mes** (sin cobros por ingreso/egreso de data "Egress Fees"), esto deja un margen bruto increíble para los microplanes en integraciones tipo WHMCS.

### Estrategia de Retorno Inmediato
A continuación, sugerencia de empaquetados para la venta final al usuario. 

| Plan WHMCS | Espacio Wasabi (GB) | Costo Real AWS/Wasabi | Precio de Venta (Cliente) | Margen ROI | Target de Mercado |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Startup Docker**| 100 GB | ~$0.70 USD / mes | **$4.99 USD / mes** | > 600% | VPS Pequeños, Devs |
| **Business Plus** | 500 GB | ~$3.50 USD / mes | **$14.99 USD / mes** | > 300% | E-Commerce, Pymes |
| **Enterprise Ops**| 2 TB (2000 GB) | ~$14.00 USD / mes | **$49.99 USD / mes**| > 250% | Agencias y SaaS |

*Nota: Puesto que Restic hace deduplicación (solo guarda lo que cambia), un cliente que sube 100 GB y retiene 7 copias de 7 días, puede que ocupe de hecho solo 105 GB (no 700 GB), abaratando enormemente el costo real de proveedor, sumando ganancias exponenciales invisibles.*

### Estrategia de Venta Opcional
1. **Ad-on Silencioso:** Vender un VPS e incluir automáticamente en el carrito "Protección Docker Backup 15GB" por $2.00 adicionales. La tasa de adopción de seguros ante pérdidas es del 40%.
2. **Upsell en Restauraciones Bare-Metal:** La opción de recuperar a un "Nuevo VPS" de inmediato atrae clientes que no pueden permitirse tiempo fuera de línea, justificando planes "Enterprise" o cobros extra tipo *Asistencia Crítica*.

---
**Conclusión para Ingeniería/Arquitectura:**  
El sistema es inherentemente resiliente frente a vulnerabilidades directas porque los VPS de los clientes NO conocen la arquitectura back-end real ni poseen las llaves maestras de borrado de Wasabi. Docker Backup Pro actúa en un rol pasivo/agresivo de recolección garantizando inmunidad a Ransomware.
