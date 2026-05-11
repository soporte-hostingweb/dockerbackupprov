# Comparativa de Planes

El servicio ofrece tres niveles principales de protección.

| Característica | FREE (Basic) | PRO (Standard) | ENTERPRISE |
|:---|:---:|:---:|:---:|
| **Retención** | 2 días | 7 días | 30 - 365 días |
| **Prioridad de Cola** | Baja (1) | Estándar (2) | VIP (3) |
| **Programación** | Solo manual | Diaria automática | Personalizada (días/hora) |
| **Verificación Integridad** | Ninguna | Semanal (Light) | Diaria Automatizada |
| **Restore Automático** | No | Sí (Agent) | Sí + Sandbox |
| **Nivel de Protección** | Manual | Avanzado | Total (auto-gestionado) |
| **Snapshot Mode** | live | live | live + consistent |
| **Auto-Config** | No | No | Sí (días + rutas + horario) |

## Opciones de Almacenamiento (S3)
Todos los planes pueden correr sobre:
1. **Nube Premium Provista (MinIO/Wasabi):** No te preocupas de nada, el almacenamiento está incluido en tu suscripción SaaS.
2. **BYOS (Bring Your Own Storage):** Si ya tienes tu propia cuenta de AWS S3, Wasabi o Cloudflare R2, puedes habilitar la opción en tu panel para conectar tu bucket. En este modo solo pagarás la licencia del software de orquestación (Agent + Panel).

## 🟢 FREE (Basic) — Selección Manual
- **Sin interrupción de servicios**
- Solo carpetas que el usuario marca en el panel.
- Sin verificación de integridad.

## 🔵 PRO (Standard) — Copia Automatizada + Dynamic Tracking
- Los tags especiales `[ALL_TARGETS]:contenedor` se resuelven dinámicamente.
- **Sin interrupción de servicios** (modo live).
- Programación automática.
- Verificación automática tras el backup.

## 👑 ENTERPRISE — Protección Total Automática
- Configuración automática que respalda TODO el servidor (`/host_root` completo excluyendo el SO) cada noche.
- Sandbox test restore para confirmar estatus DR-Ready.
