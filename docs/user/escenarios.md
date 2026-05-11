# Escenarios y Casos de Uso

A continuación, se presentan los escenarios de instalación más comunes y cómo el Agente SaaS se adapta a cada entorno de forma transparente.

### ESCENARIO A: Docker Nativo (Ubuntu/Debian) 🐳
**Perfil**: Servidor moderno corriendo múltiples servicios web con Docker Compose.
**Resultado**: Ejecutas el comando de instalación e inmediatamente el Agente detecta el socket de Docker (`/var/run/docker.sock`).
**Beneficio**: En el dashboard verás la lista exacta de todos tus contenedores. Podrás hacer backup de volúmenes individuales o activar la "Protección Total" que agrupará todo automáticamente sin que detengas los servicios.

### ESCENARIO B: Ubuntu / AlmaLinux (Stack Tradicional sin Docker) 🐧
**Perfil**: VPS tradicional corriendo Apache/Nginx, PHP-FPM y MySQL de forma nativa.
**Comportamiento**: 
1. El script detecta que no hay Docker instalado.
2. Descarga e instala Docker Engine de forma invisible en segundo plano.
3. Levanta el agente en modo "Host-Mount".
**¿Qué respalda el agente?**
- `/var/www/html` o `/home/user/public_html` (Archivos web)
- `/etc/nginx` o `/etc/httpd` (Configuración del servidor)
- `/var/lib/mysql` (Bases de datos)
**Nota**: El agente no interrumpe MySQL ni Apache, copiando los archivos en vivo (Live Mode).

### ESCENARIO C: Sistema Personalizado (Node.js, PM2, Python) ⚙️
**Perfil**: Aplicación web desarrollada a medida, levantada con PM2 o systemd en un VPS.
**Configuración Recomendada**: 
- En tu panel web, no verás contenedores. Debes abrir el explorador de archivos y marcar `/home/app` (tu código fuente) y `/etc/nginx` (tu reverse proxy).
- Alternativamente, si tienes el plan Enterprise, el sistema seleccionará dinámicamente todo el disco principal (`/host_root`) excluyendo carpetas del sistema operativo.

### ESCENARIO D: Infraestructura Crítica Windows Server 🪟
**Perfil**: Servidor Windows con Active Directory, IIS, o SQL Server, con grandes volúmenes de datos (ej. 7TB).
**Instalación**: El cliente recibe un comando PowerShell para instalar el Agente Windows.
**Estrategia**:
- Se configura la "Protección Total" sobre la unidad `C:\` o volúmenes de datos `D:\`.
- **Deduplicación Masiva**: La copia inicial de 7TB tomará tiempo (dependiendo del ancho de banda), pero los respaldos diarios siguientes solo enviarán los megabytes modificados, demorando minutos.
- **Protección Anti-Ransomware**: Dado que los snapshots en Wasabi/MinIO son inmutables desde el servidor comprometido, si Windows es infectado por Ransomware, puedes usar el Modo Rescue V2 para volcar los 7TB en un nuevo servidor Windows completamente limpio.
