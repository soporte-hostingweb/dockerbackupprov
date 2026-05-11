# Análisis y Funciones Obsoletas

Este documento mantiene un registro de las funciones deprecadas o planificadas que pueden figurar en documentación muy antigua, y su estado actual en la **V14.1.0**.

## 1. Identidad del Agente
- **Obsoleto:** El uso de un archivo estático `/host_root/etc/dbp_agent_id` introducido en V2.4.
- **Actual:** `Fingerprint` (Huella digital híbrida calculada en tiempo de ejecución) desde V13.

## 2. Restic Check (Verificación de Integridad)
- **Obsoleto:** Lanzar `restic check` manualmente o solo al arranque del agente.
- **Actual:** El motor de Go lo ejecuta automáticamente al terminar un backup (`RunResticVerify()`) y reporta "VALID" o "INVALID" inmediatamente.

## 3. Modo Rescue V2 (Inyección SSH)
- **Concepto:** Utilizaba un script para inyectar comandos vía SSH en un servidor bare-metal nuevo y restaurar directamente.
- **Estado:** Pospuesto/Interno. Interfiere con el modelo de responsabilidades Zero-Config. El flujo actual preferido es instalar el Agente V14 usando el comando `curl` estándar, y luego usar el Restore Wizard.

## 4. Work in Progress (Cuellos de Botella)
- **Asynchronous Job Worker:** Aún se depende de concurrencia nativa (Goroutines) para guardar Heartbeats. Planeado migrar a colas de mensajería (RabbitMQ/Redis Streams) para soportar más de 1000 agentes simultáneos.
- **Post-Restore Validation:** Falta la validación empírica en el servidor destino (ej. ping al puerto del contenedor) tras un `restic restore`. Actualmente solo se valida que el binario de Restic no falló.
