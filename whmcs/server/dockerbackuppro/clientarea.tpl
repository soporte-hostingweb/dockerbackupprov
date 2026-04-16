<style>
    @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;700&display=swap');

    .dbp-v14-container {
        font-family: 'Outfit', sans-serif;
        background: #0b1120;
        border-radius: 16px;
        color: #f1f5f9;
        overflow: hidden;
        border: 1px solid rgba(255, 255, 255, 0.1);
        box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
        margin-bottom: 30px;
    }

    .dbp-header {
        background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
        padding: 24px 30px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .dbp-header h3 {
        margin: 0;
        font-weight: 700;
        font-size: 20px;
        letter-spacing: -0.025em;
        background: linear-gradient(to right, #60a5fa, #a78bfa);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
    }

    .dbp-plan-badge {
        background: rgba(96, 165, 250, 0.1);
        border: 1px solid rgba(96, 165, 250, 0.3);
        color: #60a5fa;
        padding: 4px 12px;
        border-radius: 9999px;
        font-size: 12px;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .dbp-body {
        padding: 30px;
    }

    .dbp-welcome-text {
        color: #94a3b8;
        font-size: 15px;
        line-height: 1.6;
        margin-bottom: 25px;
    }

    .dbp-activation-card {
        background: rgba(15, 23, 42, 0.6);
        backdrop-filter: blur(8px);
        border: 1px solid rgba(255, 255, 255, 0.05);
        border-radius: 12px;
        padding: 20px;
        margin-bottom: 25px;
        position: relative;
    }

    .dbp-step-label {
        font-size: 11px;
        text-transform: uppercase;
        color: #6366f1;
        font-weight: 700;
        margin-bottom: 8px;
        display: block;
        letter-spacing: 0.1em;
    }

    .dbp-command-box {
        background: #020617;
        padding: 16px;
        border-radius: 8px;
        font-family: 'Courier New', monospace;
        font-size: 14px;
        color: #cbd5e1;
        border: 1px solid rgba(255, 255, 255, 0.1);
        margin-top: 10px;
        word-break: break-all;
        position: relative;
        cursor: pointer;
        transition: all 0.2s;
    }

    .dbp-command-box:hover {
        border-color: rgba(99, 102, 241, 0.5);
        background: #000;
    }

    .dbp-copy-hint {
        position: absolute;
        right: 12px;
        top: 50%;
        transform: translateY(-50%);
        background: #1e293b;
        color: #94a3b8;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 10px;
        opacity: 0;
        transition: opacity 0.2s;
    }

    .dbp-command-box:hover .dbp-copy-hint {
        opacity: 1;
    }

    .dbp-stats-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 15px;
        margin-bottom: 25px;
    }

    .dbp-stat-item {
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid rgba(255, 255, 255, 0.05);
        padding: 15px;
        border-radius: 12px;
        text-align: center;
    }

    .dbp-stat-icon {
        font-size: 18px;
        margin-bottom: 8px;
        color: #a78bfa;
    }

    .dbp-stat-label {
        font-size: 11px;
        color: #64748b;
        text-transform: uppercase;
        display: block;
    }

    .dbp-stat-value {
        font-size: 14px;
        font-weight: 600;
        color: #f8fafc;
    }

    .dbp-footer-actions {
        display: flex;
        gap: 10px;
    }

    .dbp-btn {
        flex: 1;
        padding: 12px 20px;
        border-radius: 8px;
        font-weight: 600;
        font-size: 14px;
        text-align: center;
        text-decoration: none;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
    }

    .dbp-btn-primary {
        background: #6366f1;
        color: white;
    }

    .dbp-btn-primary:hover {
        background: #4f46e5;
        box-shadow: 0 0 20px rgba(99, 102, 241, 0.4);
        color: white;
    }

    .dbp-iframe-wrapper {
        margin-top: 30px;
        background: #000;
        border-radius: 16px;
        overflow: hidden;
        border: 1px solid rgba(255, 255, 255, 0.1);
        position: relative;
    }

    .dbp-iframe-header {
        background: #0f172a;
        padding: 12px 20px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    }
</style>

<div class="dbp-v14-container">
    <div class="dbp-header">
        <h3><i class="fas fa-shield-alt"></i> HW Cloud Recovery SaaS</h3>
        <span class="dbp-plan-badge">{$planType}</span>
    </div>
    
    <div class="dbp-body">
        <div class="dbp-welcome-text">
            Bienvenido a la protección de datos inteligente. Tu servidor ahora cuenta con conciencia de stack (WordPress/Bare-Metal) para garantizar respaldos consistentes y restauraciones en 1-clic.
        </div>

        <div class="dbp-stats-grid">
            <div class="dbp-stat-item">
                <div class="dbp-stat-icon"><i class="fas fa-microchip"></i></div>
                <span class="dbp-stat-label">Detección</span>
                <span class="dbp-stat-value">Smart V14</span>
            </div>
            <div class="dbp-stat-item">
                <div class="dbp-stat-icon"><i class="fas fa-database"></i></div>
                <span class="dbp-stat-label">DB Hook</span>
                <span class="dbp-stat-value">Activo</span>
            </div>
            <div class="dbp-stat-item">
                <div class="dbp-stat-icon"><i class="fas fa-bolt"></i></div>
                <span class="dbp-stat-label">Restore</span>
                <span class="dbp-stat-value">1-Click</span>
            </div>
        </div>

        <div class="dbp-activation-card">
            <span class="dbp-step-label">Paso 1: Instalación del Agente Inteligente</span>
            <p style="font-size: 13px; color: #94a3b8; margin-bottom: 12px;">Copia y pega este comando en la terminal de tu servidor (Root) para iniciar la protección:</p>
            
            <div class="dbp-command-box" id="installCmd" onclick="copyCommand()">
                {$installCommand}
                <span class="dbp-copy-hint" id="copyHint">Copiar <i class="far fa-copy"></i></span>
            </div>
        </div>

        <div style="background: rgba(245, 158, 11, 0.05); border: 1px solid rgba(245, 158, 11, 0.1); border-radius: 8px; padding: 15px; margin-bottom: 25px;">
            <p style="margin: 0; color: #f59e0b; font-size: 12px; line-height: 1.5;">
                <i class="fas fa-exclamation-triangle"></i> <strong>Tu Token es Privado:</strong> <code>{$dbpToken}</code>. <br>
                Este token vincula tu servidor con tu cuenta. No lo compartas ni lo incluyas en repositorios públicos.
            </p>
        </div>

        <div class="dbp-footer-actions">
            <a href="https://backup.hwperu.com/?sso={$dbpToken}{$debug}" target="_blank" class="dbp-btn dbp-btn-primary">
                <i class="fas fa-desktop"></i> Gestionar en Panel Externo
            </a>
            <a href="https://api.hwperu.com/kb/wordpress-recovery" target="_blank" class="dbp-btn" style="background: rgba(255,255,255,0.05); color: #fff; border: 1px solid rgba(255,255,255,0.1);">
                <i class="fas fa-graduation-cap"></i> Guía WordPress
            </a>
        </div>
    </div>
</div>

<!-- LIVE DASHBOARD IFRAME -->
<div class="dbp-iframe-wrapper">
    <div class="dbp-iframe-header">
        <span style="font-size: 13px; font-weight: 600; color: #94a3b8;"><i class="fas fa-satellite-dish" style="color: #10b981;"></i> Control en Tiempo Real</span>
        <span class="label label-success" style="background: #10b981;">V14.2 Optimized</span>
    </div>
    <iframe 
        src="https://backup.hwperu.com/?sso={$dbpToken}&embed=1{$debug}" 
        style="width: 100%; height: 850px; border: none; display: block;"
        scrolling="no"
    ></iframe>
</div>

<script>
    function copyCommand() {
        const cmdText = "{$installCommand}";
        navigator.clipboard.writeText(cmdText).then(() => {
            const hint = document.getElementById('copyHint');
            hint.innerHTML = '¡Copiado! <i class="fas fa-check"></i>';
            hint.style.color = '#10b981';
            setTimeout(() => {
                hint.innerHTML = 'Copiar <i class="far fa-copy"></i>';
                hint.style.color = '#94a3b8';
            }, 2000);
        });
    }
</script>
