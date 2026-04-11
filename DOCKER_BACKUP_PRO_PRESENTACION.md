# 🚀 HW Cloud Recovery: Enterprise Disaster Recovery as a Service (V11.7.0)

**Destinatario:** Senior Management / DevOps Operations / CTO  
**Estado del Proyecto:** V11.7.0 (SaaS Observability Milestone)  
**Visión:** Garantizar la continuidad absoluta del negocio mediante una infraestructura de respaldo agnóstica, distribuida y auto-gestionada.

---

## 1. Resumen Ejecutivo (Agnosticismo y Control)
**HW Cloud Recovery** ha evolucionado de ser una herramienta vinculada a Wasabi a convertirse en un **Orquestador Universal de DRaaS**. La plataforma ahora permite la recuperación de servidores completos (Bare-Metal) en minutos, utilizando cualquier backend compatible con S3, eliminando el riesgo de dependencia de proveedor (*Vendor Lock-in*).

---

## 2. Pilares Tecnológicos de la Versión 11.7.0

### 2.1 S3 Universal & Multi-Region
El backend ha sido desacoplado de endpoints estáticos. 
- **Compatibilidad Extendida:** Soporte nativo para AWS S3, Cloudflare R2, MinIO, Wasabi y DigitalOcean Spaces.
- **Endpoints Dinámicos:** El sistema detecta y formatea automáticamente las regiones (ej. `ca-central-1`) para asegurar la firma de seguridad AWS V4 en cualquier parte del mundo.

### 2.2 Dual Observability Dashboard (Master + Tenant)
Hemos resuelto la brecha de visibilidad en modelos SaaS:
- **Fallback Maestro:** Si un cliente no configura alertas, el administrador recibe todos los eventos por defecto.
- **Doble Notificación:** Capacidad de enviar alertas críticas simultáneamente al webhook del cliente (n8n/WhatsApp) y al panel central de la empresa.
- **Auto-Provisioning:** Los nuevos inquilinos nacen con monitoreo activo desde la primera orden en WHMCS.

### 2.3 Bare-Metal Restore & SSH Resonance
El proceso de recuperación ya no depende de que el servidor origen esté vivo:
- **Rescue Script:** Un binario efímero se inyecta vía SSH en el nuevo servidor, descarga la data y reconstruye la orquestación Docker automáticamente.
- **2FA Control:** Las clonaciones requieren un código de autorización dinámico generado en el momento, protegiendo contra secuestros de datos.

---

## 3. Seguridad y Cifrado (Grado Militar)

| Capa | Tecnología | Propósito |
| :--- | :--- | :--- |
| **Data-at-Rest** | AES-256-GCM | Todas las credenciales S3 y Restic se almacenan cifradas con una llave maestra de 32 caracteres. |
| **Data-in-Transit** | TLS 1.3 / HTTPS | Comunicación Agente <-> API blindada contra ataques Man-in-the-Middle. |
| **Redes** | Docker Internal Net | La base de datos PostgreSQL no tiene exposición externa; es invisible para el internet público. |
| **Acceso** | Auth Tokens | Identidad persistente basada en Machine-ID que sobrevive a reinicios y formateos de contenedores. |

---

## 4. Guía Operativa (Roles)

### 👨‍💻 Para el Administrador (Master Control)
1.  **Configuración Global:** Define en la pestaña "Admin" el Webhook Maestro donde llegarán todas las fallas del cluster.
2.  **Monitoreo de Salud:** Visualización del *Health Score* global. Un puntaje bajo indica latencia en backups o repositorios mal configurados.
3.  **Provisión WHMCS:** El módulo PHP se encarga de crear el tenant. El API le asignará automáticamente el webhook maestro como respaldo.

### 👤 Para el Cliente (Inquilino)
1.  **Configuración S3:** Puede usar el almacenamiento de la empresa (default) o añadir sus propias llaves en la pestaña "Settings".
2.  **Gestión de Backups:** Selección granular de carpetas. El agente detecta automáticamente si Docker está instalado.
3.  **Restauración 1-Click:** Capacidad de explorar instantáneas y devolver archivos específicos o el servidor completo al estado anterior.

---

## 5. Planes Adaptados (Matriz SaaS)

| Plan | Almacenamiento | Observabilidad | Funcionalidad DR |
| :--- | :--- | :--- | :--- |
| **Basic** | Wasabi Shared | Solo Errores | Backup Diario, Restore manual. |
| **Standard** | Wasabi / S3 Custom | Alertas + Health Score | Backup c/4h, Verificación RTO. |
| **Enterprise** | **Full Agnostic S3** | Dual Dispatch + Webhooks | Clonado Bare-Metal, Prioridad Gold. |

---

## 6. Reglas de Uso y Buenas Prácticas
- **Formato de Endpoints:** Al configurar S3 manual, no insertar el prefijo `s3:https://`; el API lo añadirá automáticamente.
- **Política de Retención:** El sistema purga automáticamente snapshots antiguos según el plan configurado para ahorrar costos de almacenamiento.
- **Hardening del Agente:** Se recomienda ejecutar el agente con acceso de solo lectura al root del sistema para máxima seguridad.

---

**Conclusión:**  
HW Cloud Recovery V11.7.0 no es solo un software de backups; es una póliza de seguro automatizada para infraestructuras Docker. Estamos listos para escalar a nivel global con total independencia de proveedores de nube.
