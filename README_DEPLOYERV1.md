# 📝 FICHA TÉCNICA: DOCKER BACKUP PRO - DEPLOYER V1

Este documento detalla la arquitectura de implementación del sistema de backups SaaS para entornos Docker, comparando sus capacidades con estándares de la industria como JetBackup 5.

## 1. Arquitectura de Sincronización
### Motor de Incremental "CDC" (Content Defined Chunking)
A diferencia de sistemas basados en archivos (JetBackup), DBP utiliza **Restic**, que opera a nivel de bloques de datos.
- **Sincronización en Caliente:** Permite respaldar archivos abiertos (Bases de datos, Logs) subiendo únicamente los bloques que han cambiado.
- **Deduplicación Global:** Si dos contenedores tienen el mismo archivo, solo se almacena una vez en Wasabi S3.
- **Integridad:** Hash SHA-256 nativo para cada bloque movido.

## 2. Gestión de Procesos y Latencia (Dashboard)
El sistema implementa un monitoreo activo de los hilos de ejecución en el Agente.

### 🕹️ Mantenimiento y Terminación
- **Botón Maintenance:** Posee lógica dual. 
  - Si el Agente reporta un backup activo (`IsSyncing: true`), el botón permite **Terminar el Proceso Lento**.
  - **Acción:** Envía una señal `SIGINT` al PID rastreado del Agente para un cierre seguro.
  - **Log:** Registra `TERMINATED_BY_USER` en el historial de snapshots.

### ⚡ Force Global Snapshot
Permite disparar un ciclo de backup fuera de cron con dos alcances:
1. **Targeted (SaaS):** Solo las rutas seleccionadas por el cliente en el UI.
2. **Exhaustive (Full):** Respaldo de la raíz del host configurada en `/host_root` para casos de desastre.

## 3. Capa de Persistencia (Capa SaaS)
- **Base de Datos:** PostgreSQL 14 (Dockerized).
- **Esquema `AgentStatus`:**
  - `Maintenance`: Flag booleano para pausar cron reactivamente.
  - `ActivePID`: Control de procesos en el VPS del cliente.
  - `IsSyncing`: Estado visual de progreso en el Dashboard.
  - `Snapshots`: Caché de puntos de restauración para carga instantánea.

## 4. Seguridad de Datos
- **Cifrado:** AES-256 (Client-side). La API Central no posee la frase de cifrado del repositorio restic.
- **Comunicación:** JWT/SSO Token validado por el Control Plane.

---
**Versión:** 1.0.0-PROD
**Autor:** Antigravity AI Engine
**Soporte:** HWPeru - Soporte Hosting Web
