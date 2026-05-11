# Guía de Instalación para Usuarios

## Instalación desde WHMCS

Entra a **Mi Portal → Mis Productos y Servicios → Detalles del Servicio**. Verás un panel con el siguiente comando:

```bash
curl -sSL https://api.hwperu.com/install.sh | bash -s -- --token dbp_saas_3070
```

### Flujo de Instalación

Ejecuta el comando en tu VPS como `root`.

1. **Verificar dependencias:** El instalador revisará si tienes `jq` y Docker instalados. Si faltan, se autoinstalarán.
2. **Identidad del Hardware:** Se genera una huella única del hardware para seguridad.
3. **Autenticación:** El agente se comunica con el API y descarga la imagen privada de despliegue.
4. **Despliegue:** Se levanta el contenedor del agente y en aproximadamente 30 segundos aparecerá conectado en el panel web.

## Entendiendo el Panel

| Elemento | Función |
|:---|:---|
| **HEALTH: 100%** | Puntaje de salud del agente. Baja si hay un problema o está offline. |
| **OPERATIVO / OFFLINE** | Muestra en tiempo real si el servidor está conectado. |
| **RTO** | Tiempo estimado de recuperación. |
| **RPO** | Antigüedad del último backup. |

## Cómo usar el Health Score
El Health Score se calcula en base a:
- Conexión actual (-50 puntos si está offline).
- Si el último backup ocurrió hace más de 24 horas (-20 puntos).
- Si la verificación de integridad mensual o semanal falla (-40 puntos).
