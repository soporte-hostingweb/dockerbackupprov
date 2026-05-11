# Manual de Operación: Restore Wizard

El nuevo **Asistente de Restauración Avanzada** permite recuperaciones granulares con precisión quirúrgica, sin necesidad de restaurar el servidor completo si solo quieres un archivo.

## Jerarquía de Recuperación
El sistema organiza los puntos de restauración en tres niveles de profundidad:
1.  **Día (Calendario):** Selección de la fecha en la que existía el dato deseado.
2.  **Hora (Snapshot):** Selección del momento exacto del día (útil si se realizaron múltiples respaldos manuales o automáticos).
3.  **Contenido (Granular):** Exploración de las carpetas internas del respaldo para recuperar solo lo necesario.

## Flujo de Trabajo del Wizard

### 1. Selección de Punto de Control
- Al abrir el wizard desde el dashboard, el sistema agrupa los IDs de snapshot por fecha.
- El usuario selecciona el día y luego la hora específica de la lista desplegable.

### 2. Exploración del Snapshot
- El agente ejecuta `restic ls <id>` en segundo plano.
- El usuario selecciona qué carpetas o archivos desea restaurar (ej: Solo la base de datos `/database` o solo el código `/public_html`).

### 3. Configuración del Destino
- Se define un **Path de Recuperación** (por defecto `/restore_data`).
- El sistema descargará únicamente los archivos seleccionados, ahorrando tiempo y ancho de banda.
- **Auto-Up:** Si está activado, el agente intentará levantar los servicios Docker restaurados.

## Consideraciones de Rendimiento
- **Carga de Índices:** Explorar snapshots muy grandes (+500GB) puede tardar unos segundos mientras el agente descarga la estructura.
- **Seguridad:** Los archivos restaurados mantienen sus permisos originales de Linux.
- **Espacio en Disco:** Asegúrate de tener suficiente espacio en el host para el volcado de datos.
