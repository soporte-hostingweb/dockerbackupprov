📋 Te genero preguntas estratégicas + casos reales
🧩 CASO 1: WordPress sin Docker (CRÍTICO)

Escenario:
Cliente compra PRO
Tiene:

Apache + PHP + MySQL directo en VPS
WooCommerce (transacciones activas)
❓ Preguntas clave
¿El sistema detecta automáticamente que NO hay Docker?
¿Qué muestra el File Explorer? (¿vacío?)
¿El usuario sabe que debe seleccionar /var/www/html?
¿Se está incluyendo la base de datos correctamente (/var/lib/mysql)?
¿Qué pasa si MySQL está escribiendo en ese momento? (modo LIVE)
¿El restore reconstruye correctamente WordPress funcional?
¿Se necesita dump de base de datos (mysqldump) o solo filesystem?
¿WooCommerce podría perder pedidos recientes?
¿El cliente entiende el riesgo de inconsistencia?
¿Deberías mostrar advertencia tipo:
👉 “Sistema sin Docker detectado — recomendamos backup consistente”?
🧩 CASO 2: WordPress + Alto tráfico (eCommerce activo)

Escenario:

50–100 pedidos por hora
Pagos en tiempo real
❓ Preguntas críticas
¿El modo LIVE es suficiente o peligroso?
¿Deberías forzar modo CONSISTENT en PRO?
¿Puedes pausar MySQL sin Docker?
¿El sistema debería ofrecer:
“Backup sin interrupción”
“Backup consistente (recomendado para ecommerce)”?
¿Qué pasa si el restore trae DB corrupta?
¿El Health Score detecta inconsistencias reales?
🧩 CASO 3: Cliente con Docker (ideal)

Escenario:

WordPress en Docker
DB en contenedor separado
❓ Preguntas
¿El sistema detecta automáticamente contenedores?
¿Funciona [ALL_TARGETS]:contenedor correctamente?
¿El modo CONSISTENT pausa contenedores correctamente?
¿El restore levanta servicios con Auto-Up?
¿Qué pasa si el contenedor tiene volúmenes externos?
🧩 CASO 4: VPS sin Docker + Node.js (PM2)

Escenario:

App en /home/app
Sin contenedores
❓ Preguntas
¿El panel queda vacío? (mala UX)
¿Existe opción tipo:
👉 “Modo servidor completo (recomendado)”?
¿El usuario sabe qué carpetas elegir?
¿Se respalda .env (crítico)?
¿Se respalda PM2 (~/.pm2)?
🧩 CASO 5: Cliente NO técnico (muy común)

Escenario:

No sabe qué es /etc, /var, etc.
❓ Preguntas UX
¿El sistema debería tener presets?

Ejemplo:

✅ “WordPress estándar”
✅ “Servidor completo”
✅ “Aplicación Node”
¿Puedes detectar automáticamente WordPress?
¿Puedes sugerir rutas automáticamente?
¿Puedes ocultar rutas técnicas?
🧩 CASO 6: Restore en servidor nuevo

Escenario:

Cliente pierde VPS
Instala en uno nuevo
❓ Preguntas
¿El restore incluye:
Archivos
Base de datos
Configuración web?
¿El sistema levanta servicios automáticamente SIN Docker?
¿El cliente tiene que:
reinstalar Apache?
reinstalar MySQL?
¿Tu producto realmente es “1-click recovery”?

👉 Aquí hay un gap fuerte.

🧩 CASO 7: Cliente mezcla Docker + NO Docker

Escenario híbrido:

WordPress tradicional
Redis en Docker
Otro servicio en contenedor
❓ Preguntas
¿El sistema soporta ambos mundos?
¿El backup combina:
/host_root
contenedores?
¿Hay duplicidad de datos?
¿El restore respeta dependencias?