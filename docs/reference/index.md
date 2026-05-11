# Arquitectura y Seguridad

## Arquitectura de Seguridad (V14)

### 1. Fingerprint Híbrido Anti-Clonado
Cada agente genera una identidad única e irrepetible al instalarse:
```text
fingerprint = SHA256(machine-id + disk_serial_or_uuid + hostname)
```

Esta huella se envía en **cada latido** (`X-Agent-Fingerprint`). Si no coincide con la registrada al activar el token, el API rechaza la conexión, previniendo clonación de licencias.

### 2. Ciclo de Vida del Token SaaS
```text
WHMCS Provisión → Token generado (dbp_saas_XXXX)
      ↓
Estado: pending (válido 48h)
      ↓
Cliente instala → Token se vincula al hardware
Estado: activated
      ↓
Si pasan 48h sin instalar → Estado: expired
      ↓
Admin puede revocar → Estado: revoked
```

### 3. Capas de Seguridad

| Capa | Tecnología | Detalle |
|:---|:---|:---|
| **Data at Rest** | AES-256-GCM | Credenciales S3 cifradas en DB con llave maestra de 32 bytes |
| **Data in Transit** | TLS 1.3 | Toda comunicación Agente ↔ API |
| **Autenticación** | Triple Token | X-Agent-ID + X-Agent-Key + X-Agent-Fingerprint en cada request |
| **Imagen Docker** | GHCR Privado | Cliente nunca ve el código fuente |
| **PAT GHCR** | Controlado | Token de descarga nunca expuesto al cliente |

## Sandbox Test (Verificación Inmune)
Planes Enterprise ejecutan validaciones en redes aisladas para confirmar que un backup no solo descomprime bien, sino que las bases de datos levantan exitosamente.
