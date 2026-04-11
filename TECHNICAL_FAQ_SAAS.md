# 🛡️ Hard-Talk Technical FAQ: HW Cloud Recovery (Enterprise Roadmap)

Este documento reemplaza a los manuales generales. Aquí respondemos a las "preguntas duras" sobre el estado real de la tecnología, diferenciando entre lo **implementado**, lo que es una **limitación** y el **plan de solución**.

---

## 1. Motor de Backups — Validación REAL

### ¿restic check está automatizado?
- **Estado Actual:** No totalmente. Se ejecuta bajo demanda o en el arranque del agente.
- **Limitación:** El API no sabe si la data en S3 se degradó (bit rot) hasta que el administrador lo lanza manualmente.
- **[PLAN V11.8]:** Implementar `restic check --read-data-subset=1%` en cada ciclo de Heartbeat. Si falla, el agente reporta "DATA_CORRUPT" y se bloquea el restore para evitar inyectar basura en el servidor del cliente.

---

## 2. Backups Interrumpidos (Control de Negocio)

### ¿Cómo detectamos que un backup quedó incompleto?
- **Estado Actual:** Si Restic completa un resume exitoso, el API registra `Status: SUCCESS`.
- **Limitación:** El cliente no sabe si ese respaldo es solo un diferencial de 5 minutos o si falló antes de subir la base de datos crítica.
- **[PLAN V11.8]:** Parsear el output JSON de Restic. Diferenciaremos en el Dashboard: 
  - `FULL_SUCCESS` (Todo procesado) 
  - `PARTIAL_RESUME` (Continuación de una falla previa).

---

## 3. S3 Multi-Cloud — Inteligencia y Fallover

### ¿Qué pasa si Wasabi está lento pero no caído?
- **Estado Actual:** El agente espera hasta el timeout. No hay detección de degradación.
- **[PLAN V11.9]:** Implementar **Circuit Breaker** en el Agente. Si la latencia de subida cae por debajo de 1MB/s sistemáticamente, el agente marcará "PROVIDER_DEGRADED" y el administrador podrá gatillar un **Failover Multi-Cloud** (mover el repo a R2/AWS).

---

## 4. Agente — Observabilidad Real y Zombies

### ¿Cómo detectamos un agente “zombie”? 
- **Estado Actual:** El Heartbeat solo dice "estoy vivo". No garantiza que el hilo de backup no esté bloqueado.
- **[PLAN V11.9]:** El Heartbeat reportará el `ActionState` (Idle, Task_Running, Locked) y el `PID` del subproceso Restic. Si el estado es `Task_Running` por más de 12h sin progreso de red, el API lo marcará como **NODE_ZOMBIE**.

---

## 5. Restore — Evidencia vs Teoría

### ¿Quién valida que el restore fue exitoso?
- **Estado Actual:** Solo validamos el código de salida del binario Restic.
- **[PLAN V12.0 - CRÍTICO]:** Implementar **Post-Restore Validation Hooks**. El sistema ejecutará un comando (ej. `docker ps` o una consulta SQL) tras restaurar. Si el servicio no responde, el dashboard marcará: `RESTORE_FINISHED_WITH_SERVICE_ERROR`.

---

## 6. Docker Awareness — El riesgo del Registry

### ¿Qué pasa si la imagen original ya no existe en Docker Hub?
- **Riesgo Real:** El restore fallará al no poder hacer `docker pull`.
- **[PLAN V12.0]:** Opción **"Air-Gapped Backup"**. Guardar el `.tar` de las imágenes Docker críticas junto con los volúmenes en el S3. Permite reconstruir el entorno sin internet externo.

---

## 7. Health Score — ¿Humo o Realidad?

### ¿Es determinístico?
- **Estado Actual:** Heurístico (Basado en edad y conexión). Puede dar falsos positivos.
- **[PLAN V11.8]:** El Score será **Evidencial**. Solo llegará a 100 si el último `check` de integridad fue exitoso y el nodo reporta permisos de escritura/lectura correctos en su disco local.

---

## 8. Alertas — El riesgo del "En Memoria"

### ¿Podemos perder alertas críticas?
- **Fallo Real:** Si el API se reinicia mientras hay 50 alertas en cola de `DispatchAlert` (goroutines), **esas alertas se pierden.**
- **[SOLUCIÓN URGENTE]:** Migrar el motor de alertas a **Redis Streams**. Las alertas se encolarán de forma persistente y el API reintentará el envío hasta confirmar el 200 OK del webhook destino.

---

## 9. Seguridad — ¿Aislamiento Físico?

### ¿Qué evita que un bug acceda a otro prefijo?
- **Estado Actual:** Aislamiento lógico basado en tokens. 
- **[PLAN ENTERPRISE]:** Para clientes corporativos, el sistema creará **IAM Policies dinámicas** o buckets separados. El agente solo recibirá credenciales limitadas exclusivamente a su prefijo S3 mediante *AWS Security Token Service (STS)*.

---

## 10. Escalabilidad — El Cuello de Botella

### ¿Dónde está el límite real hoy?
- **Realidad:** El cuello de botella hoy es la **Concurrencia de la DB**. Al no tener una cola de trabajos (Job Queue), 500 heartbeats simultáneos pueden saturar el pool de conexiones de Postgres.
- **[PLAN V12.0]:** Implementar **Asynchronous Job Worker** (RabbitMQ/Redis). El API solo recibe la petición y la encola; los workers procesan la lógica pesada por separado.

---

## 11. Testing — El Hito de Autenticidad

### ¿Cuántos restores completos hemos probado?
- **Respuesta Brutal:** Menos del 2% de los escenarios posibles. 
- **[ROADMAP V12 - CERTIFICACIÓN]:** Automatizar el ciclo:
  `Provisioning -> Dummy Data -> Backup -> Destrucción de VM -> Restore Bare-Metal -> Validación SQL -> Destrucción`. 
  Solo cuando este ciclo corra en cada commit, podremos garantizar un 99.9% de RTO.

---
© 2026 HWPeru - Technical Risk Assessment Team.
