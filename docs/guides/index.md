# Guías de Backup Total y Parcial

El panel de configuración permite personalizar exactamente qué carpetas e información quieres enviar a tu almacenamiento (Wasabi S3).

## Backup Total (Protección Total Inteligente)

Para usuarios que no quieren complicaciones (o plan Enterprise autoconfigurado):
En el panel verás el tag `[ALL_SYSTEM_ROOT]`. El sistema seleccionará dinámicamente todo el disco principal del servidor, excluyendo automáticamente archivos temporales (`/proc`, `/sys`, `/tmp`).
- Para ejecutarlo de inmediato: Haz click en el botón **Full**.

## Backup Parcial (Carpetas Específicas)

Si necesitas hacer backup solo de partes concretas:
1. Abre el **FileExplorer** en tu panel.
2. Expande el árbol de directorios de tu servidor o contenedores.
3. Selecciona las rutas necesarias (por ejemplo, `/var/www/html` o `/var/lib/mysql`).
4. Haz clic en **Save Configuration** o **Lock Selection**.
5. Si deseas ejecutarlo manualmente ahora, presiona **Force Selected**.

## Snapshot Modes

Existen dos estrategias de backup para garantizar consistencia:
- **LIVE (Default):** El backup corre sin pausar ningún servicio. Útil para cero downtime, aunque podría generar una ligera inconsistencia en bases de datos muy activas si no hay volcado SQL programado.
- **CONSISTENT (Enterprise):** Realiza una micropausa controlada (por ejemplo `docker pause`) durante la toma del snapshot de la base de datos, asegurando una copia 100% exacta y consistente, a cambio de unos segundos de downtime.
