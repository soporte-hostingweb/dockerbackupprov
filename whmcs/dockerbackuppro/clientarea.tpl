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
        <a href="https://portal.hwperu.com/login?sso={$dbpToken}" target="_blank" class="btn btn-default btn-block">
            <i class="fas fa-external-link-alt"></i> Open Dashboard in New Tab
        </a>
    </div>
</div>

<!-- DASHBOARD EMBEDDED -->
<div class="panel panel-default" style="margin-top: 20px; border: 1px solid #1f2937; background: #000;">
    <div class="panel-heading" style="background: #111827; color: #fff; border-bottom: 1px solid #1f2937;">
        <h3 class="panel-title"><i class="fas fa-desktop"></i> Live Backup Control Panel</h3>
    </div>
    <div class="panel-body" style="padding: 0; background: #000;">
        <iframe 
            src="https://portal.hwperu.com/login?sso={$dbpToken}&embed=1" 
            style="width: 100%; height: 750px; border: none; overflow: hidden;"
            scrolling="no"
        ></iframe>
    </div>
</div>
