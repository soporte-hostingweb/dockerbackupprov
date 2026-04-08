# Docker Backup Pro - Rescue V2 (Bare-Metal Clone)

Este documento detalla la arquitectura para la restauración y clonación de servidores enteros hacia nuevos VPS vírgenes sin afectar al nodo de producción original.

## 🎯 Concepto Central (Zero-Impact Clone)
El objetivo de la V2 es evitar que el servidor de producción sufra penalizaciones de ancho de banda o disco mientras un cliente clona un entorno. El API actuará como **Orquestador SSH** inyectando un agente efímero ("Rescue Agent") directamente en el nuevo servidor, el cual tirará de los datos del bucket de Wasabi correspondiente al inquilino.

## 🚦 Flujo de Trabajo (Workflow)

### 1. Validación Previa (UI & API)
- El cliente abre el `Restore Wizard` en el Dashboard.
- El sistema detecta que es un backup `force_full` (raíz del sistema).
- Aparece el botón premium: **Restore to New VPS**.
- El cliente ingresa credenciales SSH del nuevo equipo (IP, root, pass).
- El API pausa temporalmente (modo `maintenance_on`) los respaldos cíclicos del servidor origen para evitar inconsistencias en el "tenant path".

### 2. Reconocimiento (SSH Handshake)
- El API usa Go nativo (`golang.org/x/crypto/ssh`) para conectar al servidor destino en menos de 1 segundo.
- Comando: `df -k / | awk 'NR==2 {print $4}'`.
- **Cálculo de Viabilidad:**
  - `Espacio Disponible > (Tamaño del Backup * 2) + Ponderado de Error`.
  - Si no pasa el check, se aborta instantáneamente notificando a la UI.

### 3. Inyección del Rescue Agent
- El API envía un payload bash comprimido al nuevo VPS.
- Este script no instala Docker ni el API, simplemente:
  1. Descarga el binario de Restic.
  2. Expone en memoria RAM las variables AWS del Token del Cliente.
  3. Ejecuta `restic restore [SNAP_ID] --target /`.

### 4. Telemetría Asíncrona 
- Mientras Restic procesa, un micro-script envía un CURL cada 10 segundos al endpoint `/v1/agent/activity/report` del API Central informando el porcentaje.
- La UI consulta las actividades y muestra la barra de progreso sin que el usuario deba mantener la pantalla encendida (Background Job).

### 5. Finalización
- Una vez restaurado, el Rescue Agent borra su rastro y lanza un `reboot` automático.
- El API quita la pausa (`maintenance_off`) del agente origen.
- El Dashboard marca el clon como `SUCCESS`.

## ⚙️ Análisis de Rendimiento y Tuning (Para 100+ Clientes)
Al delegar la descarga al VPS Destino, el API se convierte solo en un semáforo.
- **Ancho de banda del API:** Prácticamente 0 Mbps ("solo texto de control").
- **CPU del API:** Menos del 2% mientras mantiene conexiones SSH pasivas.
- **RAM del API:** Cada conexión activa usa unos ~50 KB. 100 restauraciones pesarán 5 MB.
- **Recursos API Recomendados:** 2 vCPU, 4GB RAM (Soportará sin problemas).

## 🌍 Sistema i18n
La V2 incluye también Internacionalización estática inyectada:
- Se procesará mediante un `Manager` en el backend que priorizará el archivo `custom_lang.json` si un administrador lo crea. Si no, usará diccionarios alojados en memoria, protegiendo así las I/O del disco y garantizando respuestas a latencia casi cero para la UI.
