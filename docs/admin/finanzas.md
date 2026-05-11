# Estrategia Financiera y de Almacenamiento

El sistema HW Cloud Recovery soporta tres modelos de despliegue de almacenamiento, lo que te permite jugar con los márgenes de ganancia.

## 1. Modelo Self-Hosted S3 (MinIO) - Recomendado
En este modelo, configuras un servidor propio (VPS) de gran capacidad (ej. 10TB) e instalas MinIO. En tu archivo `.env` del API, configuras el `MINIO_ENDPOINT` como el storage por defecto.
*   **Costo:** Un VPS de 10TB suele costar ~$50 USD / mes (aprox. $0.005 USD por GB).
*   **Margen de Ganancia:** Este es el modelo más rentable. Controlas el costo fijo y vendes el "Plan Completo" a un precio premium.
*   **Precio de Venta Sugerido (Pro - 100GB):** S/. 50.00 / mes.
*   **Ganancia:** ~98% de margen.

## 2. Modelo Cloud Premium (Wasabi S3)
Wasabi ofrece alta disponibilidad y es ideal si no quieres administrar tu propio VPS de almacenamiento.
*   **Costo Fijo:** $6.99 USD por 1 TB (1024 GB) al mes ($0.0068 USD por GB).
*   **Costo Real en Soles:** S/. 0.026 por GB.
*   **Precio de Venta Sugerido (Enterprise - 500GB):** S/. 120.00 / mes.
*   **Costo Wasabi:** S/. 14.00.
*   **Ganancia:** 88% de margen.

## 3. Modelo BYOS (Bring Your Own Storage)
Para clientes corporativos que ya tienen cuentas en AWS, Cloudflare R2 o su propio Wasabi.
*   **El Cliente:** Activa el "check" en su panel e introduce sus credenciales S3.
*   **El Proveedor (Tú):** Vendes únicamente la **Licencia del Software SaaS** (Orquestación, Agente Inteligente, Restore Wizard).
*   **Costo para ti:** S/. 0.00 (El almacenamiento lo paga el cliente a su proveedor).
*   **Precio de Venta Sugerido (Solo Software):** S/. 29.00 / mes por servidor.
*   **Ganancia:** 100% de margen operativo.

---

## Optimización de Costos Internos

1.  **Auto-Pruning:** Restic deduplica los datos. Un backup de 100GB retenido por 7 días ocupará ~106GB, no 700GB. `restic forget --keep-daily 7 --prune` liberará el espacio viejo automáticamente.
2.  **Mínimo de Facturación Wasabi:** Si usas Wasabi, cobran un mínimo de 1 TB. Si usas MinIO, este problema no existe.
3.  **Cross-Region (Plan Enterprise):** Puedes ofrecer a tus clientes Enterprise que su backup primario vaya a MinIO (Local) y una réplica vaya a Wasabi S3 (Cloud), multiplicando el valor del servicio.

## Integración WHMCS

Crea los productos en WHMCS con opciones configurables (Configurable Options):
- **Opciones de Storage:** "Usar Almacenamiento Incluido (Premium)" vs "Usar Mi Propio S3 (BYOS - Descuento)".
- **Planes SaaS:** `start_10gb`, `pro_100gb`, `enterprise_500gb`, o `license_only` para BYOS.
