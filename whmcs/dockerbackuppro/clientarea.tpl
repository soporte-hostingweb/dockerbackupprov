<div class="panel panel-default">
    <div class="panel-heading">
        <h3 class="panel-title">Docker Backup Pro - Agent Installation</h3>
    </div>
    <div class="panel-body">
        <p>To protect your VPS environment, please login as <b>root</b> into your server and run the following automated installation command:</p>
        
        <div class="alert alert-info" style="font-family: monospace; user-select: all; padding: 15px; margin: 15px 0;">
            {$installCommand}
        </div>
        
        <hr>
        
        <p><strong>Your Authorization Token:</strong> <code style="user-select: all;">{$dbpToken}</code></p>
        <p class="text-muted"><small>Keep this token secret. Do not share it.</small></p>
        
        <br>
        <a href="https://panel.dockerbackuppro.com/login?sso={$dbpToken}" target="_blank" class="btn btn-success btn-block" style="background-color: #10b981; border-color: #059669;">
            <i class="fas fa-lock"></i> Open Cloud Dashboard
        </a>
    </div>
</div>
