# Manual de Operación: Restore Wizard V4.0.0 (JetBackup Style) 🔐

Este documento describe la arquitectura y el flujo de trabajo del nuevo **Asistente de Restauración Avanzada** de Docker Backup Pro, diseñado para ofrecer una experiencia similar a JetBackup 5, permitiendo recuperaciones granulares con precisión quirúrgica.

## 🕒 Jerarquía de Recuperación
El sistema organiza los puntos de restauración en tres niveles de profundidad:
1.  **Día (Calendario):** Selección de la fecha en la que existía el dato deseado.
2.  **Hora (Snapshot):** Selección del momento exacto del día (útil si se realizaron múltiples respaldos manuales o automáticos).
3.  **Contenido (Granular):** Exploración de las carpetas internas del respaldo para recuperar solo lo necesario.

## 🛠️ Flujo de Trabajo del Wizard

### 1. Selección de Punto de Control
- Al abrir el wizard desde el dashboard, el sistema consulta los metadatos de Wasabi S3.
- Agrupa los IDs de snapshot por fecha.
- El usuario selecciona el día y luego la hora específica de la lista desplegable.

### 2. Exploración del Snapshot
- El agente ejecuta `restic ls <id> --json` en segundo plano.
- El usuario selecciona qué carpetas o archivos desea restaurar (ej: Solo la base de datos `/database` o solo el código `/public_html`).

### 3. Configuración del Destino
- Se define un **Path de Recuperación** (por defecto `/restore_data`).
- El sistema utiliza el flag `--include` de Restic para volcar únicamente los archivos seleccionados, ahorrando tiempo y ancho de banda.

## ⚠️ Consideraciones de Rendimiento
- **Carga de Índices:** Explorar snapshots muy grandes (ej: +500GB) puede tardar unos segundos mientras el agente descarga la estructura desde S3.
- **Seguridad:** Los archivos restaurados mantienen sus permisos originales de Linux.
- **Espacio en Disco:** Asegúrate de tener suficiente espacio en el host para el volcado de datos antes de iniciar.

---
**Nota Técnica:** Esta versión utiliza el motor Restic v0.16+ para optimizar las restauraciones parciales.
