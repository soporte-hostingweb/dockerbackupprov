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
        <a href="https://backup.hwperu.com/?sso={$dbpToken}{$debug}" target="_blank" class="btn btn-default btn-block">
            <i class="fas fa-external-link-alt"></i> Open Dashboard in New Tab
        </a>
    </div>
</div>

<!-- DASHBOARD EMBEDDED -->
<div style="margin-top: 20px; background: #000; border-radius: 8px; overflow: hidden; border: 1px solid #1f2937;">
    <div style="background: #111827; padding: 15px 20px; border-bottom: 1px solid #1f2937; display: flex; align-items: center; justify-content: space-between;">
        <h3 style="margin: 0; color: #fff; font-size: 16px;"><i class="fas fa-desktop"></i> Live Backup Control Panel</h3>
        <span class="label label-success">Connected Live</span>
    </div>
    <div style="padding: 0; background: #000;">
        <iframe 
            src="https://backup.hwperu.com/?sso={$dbpToken}&embed=1{$debug}" 
            style="width: 100%; height: 800px; border: none; display: block;"
            scrolling="no"
        ></iframe>
    </div>
</div>



