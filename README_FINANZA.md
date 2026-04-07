# 📈 Análisis Financiero: Docker Backup Pro (DBP)

Este documento detalla la estructura de costos y la rentabilidad esperada al comercializar DBP Pro mediante el almacenamiento en Wasabi S3.

## 💰 Estructura de Costos (Wasabi S3)

Wasabi tiene un costo fijo simple de **$6.99 USD por 1 TB (1024 GB) al mes**.
*   **Costo por GB:** $0.0068 USD.
*   **Tipo de Cambio Sugerido:** S/. 3.80 por $1 USD.
*   **Costo Real por GB en Soles:** S/. 0.026.

### Ejemplo de Costo por Retención (7 Días)
Gracias a la **Deduplicación de Restic**, retener 7 días de backup de una fuente de 100GB no consume 700GB.
*   **Día 1 (Full):** 100 GB.
*   **Días 2-7 (Deltas del 1%):** 1 GB x 6 = 6 GB.
*   **Total Ocupado:** ~106 GB.
*   **Costo en Wasabi:** $0.74 USD (**S/. 2.80 mensuales**).

---

## 📉 Proyección de Planes y Ganancias

Basado en un precio de venta sugerido y el costo operativo de Wasabi + VPS Central.

### 🥉 Plan Start (Hostings pequeños)
*   **Capacidad:** 10 GB.
*   **Precio de Venta:** S/. 15.00 / mes.
*   **Costo Wasabi:** S/. 0.30.
*   **Ganancia Bruta:** S/. 14.70 (**98% de Margen**).

### 🥈 Plan Pro (Recomendado para VPS)
*   **Capacidad:** 100 GB.
*   **Precio de Venta:** S/. 50.00 / mes.
*   **Costo Wasabi:** S/. 2.80.
*   **Ganancia Bruta:** S/. 47.20 (**94% de Margen**).

### 🥇 Plan Enterprise
*   **Capacidad:** 500 GB.
*   **Precio de Venta:** S/. 120.00 / mes.
*   **Costo Wasabi:** S/. 14.00.
*   **Ganancia Bruta:** S/. 106.00 (**88% de Margen**).

---

## 🚀 Optimización de Costos (Wasabi Best Practices)

1.  **Auto-Pruning:** Se implementará el comando `restic forget --keep-daily 7 --prune`. Esto asegura que los datos de más de 7 días se borren físicamente, liberando espacio y bajando tu factura.
2.  **Mínimo de Facturación:** Wasabi cobra un mínimo de 1 TB si el total de tus buckets es menor a eso. 
    *   *Estrategia:* Debes tener al menos 10 clientes del Plan Pro (100GB cada uno) para empezar a pagar exactamente lo que consumes.
3.  **Sin cargo por Egress:** Wasabi no cobra por descargar datos (restauraciones), lo cual hace que tus costos sean **predecibles**.

---

## 🛒 Integración WHMCS

Se recomienda crear un **Producto Adicional (Addon)** o un **Configurable Option** en WHMCS llamado "Backup Pro Plans" con los siguientes valores para automatizar la facturación:
- `manual_only` -> Gratis.
- `daily_2am` -> S/. 50.
- `custom_pro` -> S/. 120.
