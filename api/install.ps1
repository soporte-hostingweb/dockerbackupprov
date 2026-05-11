# ==========================================================
# 🚀 HW Cloud Recovery - Windows Agent Installer (V15.0.0)
# ==========================================================

$API_ENDPOINT = "https://api.hwperu.com"
$INSTALL_DIR = "C:\Program Files\DockerBackupPro"
$CONFIG_FILE = "C:\ProgramData\dbp\agent.json"

echo "=========================================================="
echo "   🛡️  HW CLOUD RECOVERY - SISTEMA DE ACTIVACIÓN SaaS     "
echo "=========================================================="

# 1. Verificar Privilegios de Administrador
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "❌ Error: Debe ejecutar este script como Administrador."
    exit
}

# 2. Obtener Token
$SAAS_TOKEN = ""
if ($args.Count -gt 0) {
    for ($i=0; $i -lt $args.Count; $i++) {
        if ($args[$i] -eq "--token") {
            $SAAS_TOKEN = $args[$i+1]
        }
    }
}

if (-not $SAAS_TOKEN) {
    $SAAS_TOKEN = Read-Host "🔑 Ingrese su Activation Token (hw_pk_xxxx)"
}

if (-not ($SAAS_TOKEN -match "^hw_pk_" -or $SAAS_TOKEN -match "^dbp_saas_")) {
    Write-Error "❌ Error: Formato de token inválido."
    exit
}

# 3. Generar Huella de Hardware
echo "[1/4] Generando Huella de Hardware..."
$machineId = (Get-CimInstance Win32_ComputerSystemProduct).UUID
$diskId = (Get-CimInstance Win32_DiskDrive | Select-Object -First 1).SerialNumber
$hostname = $env:COMPUTERNAME

$rawFingerprint = "$machineId:$diskId:$hostname"
$sha256 = [System.Security.Cryptography.SHA256]::Create()
$fingerprintBytes = $sha256.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($rawFingerprint))
$FINGERPRINT = [System.BitConverter]::ToString($fingerprintBytes).Replace("-", "").ToLower()

echo "✅ Huella Generada: $($FINGERPRINT.Substring(0,12))..."

# 4. Activar Licencia
echo "[2/4] Validando licencia con HWPeru Cloud..."
$body = @{
    token = $SAAS_TOKEN
    fingerprint = $FINGERPRINT
    hostname = $hostname
    os = "windows"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_ENDPOINT/v1/activate" -Method Post -Body $body -ContentType "application/json"
} catch {
    Write-Error "❌ Error de conexión con el API: $_"
    exit
}

if ($response.status -ne "activated" -and $response.status -ne "re-activated") {
    Write-Error "❌ Error de Activación: $($response.error)"
    exit
}

$AGENT_ID = $response.agent_id
$API_KEY = $response.api_key

echo "✅ Activación Exitosa. AgentID: $AGENT_ID"

# 5. Descargar Binarios
echo "[3/4] Descargando componentes..."
if (-not (Test-Path $INSTALL_DIR)) { New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null }
if (-not (Test-Path "C:\ProgramData\dbp")) { New-Item -ItemType Directory -Path "C:\ProgramData\dbp" | Out-Null }

Invoke-WebRequest -Uri "$API_ENDPOINT/bin/restic.exe" -OutFile "$INSTALL_DIR\restic.exe"
Invoke-WebRequest -Uri "$API_ENDPOINT/bin/dbp-agent.exe" -OutFile "$INSTALL_DIR\dbp-agent.exe"

# 6. Guardar Configuración
$config = @{
    agent_id = $AGENT_ID
    api_key = $API_KEY
    fingerprint = $FINGERPRINT
} | ConvertTo-Json

$config | Out-File -FilePath $CONFIG_FILE -Encoding utf8

# 7. Registrar Servicio (Usando PowerShell nativo)
echo "[4/4] Configurando servicio del sistema..."
$serviceName = "DockerBackupProAgent"

if (Get-Service $serviceName -ErrorAction SilentlyContinue) {
    Stop-Service $serviceName
    Remove-Service $serviceName
    sleep 2
}

New-Service -Name $serviceName -BinaryPathName "`"$INSTALL_DIR\dbp-agent.exe`"" -DisplayName "Docker Backup Pro Agent" -Description "Servicio de respaldo HW Cloud Recovery" -StartupType Automatic
Start-Service $serviceName

echo ""
echo "=========================================================="
echo "✨  INSTALACIÓN V15.0.0 COMPLETADA  ✨"
echo ""
echo "  🛡️  Agente:     $AGENT_ID"
echo "  🖥️  Servidor:   $hostname"
echo "  📊  Panel:      https://backup.hwperu.com"
echo "=========================================================="
