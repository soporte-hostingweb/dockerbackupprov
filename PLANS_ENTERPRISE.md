# 💎 DBP SaaS - PLAN ENTERPRISE (Mission Critical)

El Plan Enterprise es nuestra solución de nivel más alto, diseñada para corporaciones que no pueden permitirse ni un minuto de pérdida de datos. Incluye el protocolo de **Recuperación Proactiva (DR)**.

## 🚀 Características del Plan
- **Almacenamiento:** 1 TB+ (Expandible).
- **Retención:** 30 días + Archivado mensual.
- **Snapshot Mode:** Consistent (Smart Pause) + Pro-Validation.
- **Frecuencia:** Personalizada (Ej: Cada 2 horas).
- **DR Ready:** Recuperación en VM HW Cloud con un solo clic.

---

## 📂 Casos de Uso y Ejemplos de Clientes

### Ejemplo 1: Cliente con Docker (SaaS de Gestión)
**Perfil:** Una aplicación SaaS dinámica con Redis, PostgreSQL y múltiples microservicios.
- **Configuración:** Estructura compleja de volúmenes compartidos.
- **Estrategia:** Modo Consistente con "Smart Pause" para asegurar que la DB esté quieta durante el copiado.

**Guía Rápida de Instalación:**
1. Instalar el agente Enterprise.
2. En el Dashboard, seleccionar **"Enterprise / Premium (Personalizado)"**.
3. Definir ejecución cada 4 horas.
4. Habilitar la protección SQL directa para PostgreSQL.
5. Verificación Proactiva: El sistema restaurará el backup en un nodo de prueba cada 24h para asegurar que el RTO se cumpla.

### Ejemplo 2: Instalación Pura (Portal de Noticias)
**Perfil:** Un sitio de alto tráfico con miles de imágenes y una DB de 20GB.
- **Configuración:** RAID 10 Local, sin paneles de control.
- **Estrategia:** Snapshot de raíz total `/ [ALL_SYSTEM_ROOT]`.

**Guía Rápida de Instalación:**
1. Desplegar el agente.
2. Activar el **"Nivel de Protección: TOTAL"**. Esto respaldará todo el host automáticamente.
3. El agente usará deduplicación agresiva para subir los 20GB de DB en minutos tras el primer backup.
4. Configurar alertas por WhatsApp ante el menor error.

### Ejemplo 3: Panel cPanel / WHM (Servidor Completo)
**Perfil:** Una empresa que gestiona su propia infraestructura de hosting con 100+ clientes.
- **Configuración:** Servidor de producción crítico.
- **Estrategia:** Backup de todo el servidor y DR activo.

**Guía Rápida de Instalación:**
1. Instalación a nivel de Kernel en el servidor principal.
2. Selección de todos los puntos de montaje de usuario.
3. Configuración de un **VIRTUALIZOR_TEMPLATE** en el panel administrativo.
4. En caso de fallo crítico de hardware, el cliente pulsa **"Recuperar Servidor"** y en 5 minutos tiene un nuevo VPS idéntico con toda la data restaurada.

---

> [!IMPORTANT]
> El Plan Enterprise garantiza un RTO (Recovery Time Objective) de menos de 15 minutos mediante el aprovisionamiento automatizado en nuestra red HW Cloud.
