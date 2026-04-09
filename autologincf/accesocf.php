<?php
/**
 * AccesoCF Module for WHMCS
 * Permite a los clientes acceder a sus facturas mediante enlaces seguros
 * 
 * @author Carlos Frias
 * @version 1.13
 */

// Verificar que estamos en el contexto correcto de WHMCS
if (!defined("WHMCS")) {
    die("Este archivo no puede ser accedido directamente");
}

use Illuminate\Database\Capsule\Manager as Capsule;
	
function accesocf_config() {
    $configarray = array(
    "name" => "Invoice Auto Login",
    "description" => "Allows your clients to login via a link sent in invoice emails. This link will expire in 20 days or when the invoice is paid or cancelled.",
    "version" => "1.13",
    "author" => "Carlos Frias",
    "language" => "english",
    "fields" => array(
    "option1" => array ("FriendlyName" => "Only allow access to the invoice and credit card payment screens upon auto login", "Type" => "yesno", "Size" => "10", "Description" => "", "Default" => ""),
    "expiration_days" => array ("FriendlyName" => "Expire the link after how many days.", "Type" => "text", "Size" => "10", "Description" => "", "Default" => "20"),
    "whatsappApiUrl" => array ("FriendlyName" => "URL API WhatsApp", "Type" => "text", "Size" => "50", "Description" => "URL base del API (ej: https://api.hwperu.com) - Se configurará automáticamente el endpoint /message/sendText/Central", "Default" => ""),
    "whatsappApiKey" => array ("FriendlyName" => "API Key WhatsApp", "Type" => "text", "Size" => "60", "Description" => "Token o clave para autenticación del API de WhatsApp", "Default" => ""),
    "whatsappTemplate" => array ("FriendlyName" => "Plantilla de WhatsApp (Cobranza)", "Type" => "textarea", "Rows" => "5", "Cols" => "50", "Description" => "Plantilla para mensajes de cobranza. Variables: {auto_login_link}, {invoice_id}, {client_name}, {invoice_total}, {due_date}", "Default" => "Hola {client_name}, tienes una factura pendiente por {invoice_total} con vencimiento {due_date}. Puedes pagarla aquí: {auto_login_link}"),
    "cancellationTemplate" => array ("FriendlyName" => "Plantilla de Cancelación", "Type" => "textarea", "Rows" => "5", "Cols" => "50", "Description" => "Plantilla para mensajes de cancelación. Variables: {auto_login_link}, {service_name}, {client_name}", "Default" => "Hola {client_name}, tu servicio {service_name} ha sido suspendido/cancelado. Para reactivar, por favor paga aquí: {auto_login_link}"),
    ),

    );
    return $configarray;
}

function accesocf_write_autoauthkey($whmcsRoot) {
    $configFile = $whmcsRoot . '/configuration.php';
    if (!file_exists($configFile) || !is_writable($configFile)) {
        return null;
    }
    $configContent = file_get_contents($configFile);
    if (strpos($configContent, '$autoauthkey') !== false) {
        preg_match('/\$autoauthkey\s*=\s*[\'"]([^\'"]+)[\'"]\s*;/', $configContent, $matches);
        return $matches[1] ?? null;
    }
    $randomKey = bin2hex(random_bytes(20));
    $appendLine = "\n// AccesoCF AutoAuth key (added automatically by AccesoCF module)\n\$autoauthkey = '{$randomKey}';\n";
    file_put_contents($configFile, $appendLine, FILE_APPEND | LOCK_EX);
    return $randomKey;
}

function accesocf_enable_autoauth($whmcsRoot) {
    $key = accesocf_write_autoauthkey($whmcsRoot);
    $existing = Capsule::table('tblconfiguration')->where('setting', 'AllowAutoAuth')->first();
    if ($existing) {
        Capsule::table('tblconfiguration')->where('setting', 'AllowAutoAuth')->update(['value' => 'on']);
    } else {
        Capsule::table('tblconfiguration')->insert(['setting' => 'AllowAutoAuth', 'value' => 'on']);
    }
    return $key;
}

function accesocf_activate() {
  try {
    if (!Capsule::schema()->hasTable('accesocf')) {
      Capsule::schema()->create('accesocf', function ($table) {
          $table->string('key')->primary();
          $table->integer('user_id');
          $table->integer('invoice_id');
          $table->integer('clicks')->default(0);
          $table->dateTime('expiration');
      });
    }
  } catch (\Exception $e) {
    return array('status' => 'error', 'description' => 'Error creando tabla: ' . $e->getMessage());
  }

  $existingKey = Capsule::table('tbladdonmodules')->where('module', 'accesocf')->where('setting', 'key')->first();
  if (!$existingKey) {
      Capsule::table('tbladdonmodules')->insert(['module'  => 'accesocf', 'setting' => 'key', 'value'   => uniqid()]);
  }

  $whmcsRoot = dirname(dirname(dirname(__FILE__))); 
  if (!file_exists($whmcsRoot . '/configuration.php')) {
      $whmcsRoot = dirname(dirname(dirname(dirname(__FILE__)))); 
  }
  accesocf_enable_autoauth($whmcsRoot);

  return array('status' => 'success', 'description' => 'Module Activated. AutoAuth configurado automáticamente.');
}

function accesocf_deactivate() {
  return array('status'=>'success','description'=>'Module Deactivated.');
}

function accesocf_upgrade($vars) {
   try {
     if (!Capsule::schema()->hasColumn('accesocf', 'clicks')) {
       Capsule::schema()->table('accesocf', function($table) { $table->integer('clicks')->default(0); });
     }
     Capsule::connection()->statement("ALTER TABLE `accesocf` MODIFY COLUMN `expiration` DATETIME NULL");
   } catch (\Exception $e) {
     error_log("AccesoCF Upgrade Error: " . $e->getMessage());
   }
}

function accesocf_clientarea($vars) {
  ob_start();
  if (isset($_GET['k'])) {
    $possiblePaths = [dirname(__FILE__) . "/models/acceso_cf.php", dirname(__FILE__) . "/model/acceso_cf.php", dirname(__FILE__) . "/acceso_cf.php"];
    foreach ($possiblePaths as $path) { if (file_exists($path)) { require_once($path); break; } }
    
    $autologin = \AccesoCf\AccesoCf::where('key', $_GET['k'])->first();
    if ($autologin != false && strtotime($autologin->expiration) > time()) {
      if (isset($autologin->clicks) && $autologin->clicks >= 10) {
          logActivity("AccesoCF: Token limit exceeded for Key: " . $_GET['k']);
          ob_end_clean();
          header('Location: dologin.php?login_error=Time+Expired');
          exit();
      }

      $client = $autologin->client;
      if (!$client) {
          ob_end_clean();
          header('Location: dologin.php?login_error=Client+Not+Found');
          exit();
      }

  	  $_SESSION['used_invoice_autologin'] = true;
      $autologin->clicks = ($autologin->clicks ?? 0) + 1;
      $autologin->save();

      $results = localAPI("CreateSsoToken", ["client_id" => $autologin->user_id, "destination" => "sso:custom_redirect", "sso_redirect_path" => "viewinvoice.php?id=" . $autologin->invoice_id]);
   
      if ($results['result'] == 'success' && !empty($results['redirect_url'])) {
          ob_end_clean();
          header('Location: ' . $results['redirect_url']);
          exit();
      }

      // SSO FALLBACK
      logActivity("AccesoCF: SSO bloqueado para User ID " . $autologin->user_id . ". Intentando inyectar sesión.");
      $clientRow = Capsule::table('tblclients')->where('id', $autologin->user_id)->first();
      $userLink  = Capsule::table('tblusers_clients')->where('client_id', $autologin->user_id)->first();

      if ($clientRow && $userLink) {
          $_SESSION['uid'] = $clientRow->id;
          $_SESSION['upw'] = $clientRow->password;
          $_SESSION['auth_user_id'] = $userLink->auth_user_id; 
          $_SESSION['used_invoice_autologin'] = true;
          session_write_close();
          ob_end_clean();
          header('Location: viewinvoice.php?id=' . $autologin->invoice_id);
          exit();
      } else {
          ob_end_clean();
          header('Location: dologin.php?login_error=Auth+Failed');
          exit();
      }
    } else {
      ob_end_clean();
      header('Location: dologin.php?login_error=Time+Expired');
      exit();
    }
  }
  ob_end_clean();
  header( 'Location: dologin.php');
  exit();  
}

function processWhatsAppTemplate($template, $variables) {
  foreach ($variables as $key => $value) { $template = str_replace('{' . $key . '}', $value, $template); }
  return $template;
}

function accesocf_output($vars) {
  $action = isset($_GET['action']) ? $_GET['action'] : '';
  
  if ($action === 'test_api') {
    header('Content-Type: application/json');
    $apiUrl = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiUrl')->value('value');
    $apiKey = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiKey')->value('value');
    if (empty($apiUrl) || empty($apiKey)) { echo json_encode(['success' => false, 'message' => 'Configuración incompleta']); exit; }
    if (!strpos($apiUrl, '/', 8)) { $apiUrl = rtrim($apiUrl, '/') . '/message/sendText/Central'; }
    
    $ch = curl_init();
    curl_setopt_array($ch, [
      CURLOPT_URL => $apiUrl,
      CURLOPT_RETURNTRANSFER => true,
      CURLOPT_TIMEOUT => 10,
      CURLOPT_CUSTOMREQUEST => "POST",
      CURLOPT_POSTFIELDS => json_encode(['number' => '+51999999999', 'text' => 'Prueba']),
      CURLOPT_HTTPHEADER => ["Content-Type: application/json", "apikey: " . $apiKey],
    ]);
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    echo json_encode(['success' => $httpCode < 400, 'message' => "Código HTTP: $httpCode", 'response' => $response]);
    exit;
  }
  
  if ($action === 'send_whatsapp') {
    header('Content-Type: application/json');
    try {
      $invoiceId = isset($_POST['invoice_id']) ? (int)$_POST['invoice_id'] : 0;
      if ($invoiceId <= 0) { echo json_encode(['success' => false, 'message' => 'ID de factura inválido']); exit; }

      $apiUrl = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiUrl')->value('value');
      $apiKey = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiKey')->value('value');
      if (empty($apiUrl) || empty($apiKey)) { echo json_encode(['success' => false, 'message' => 'Configurar API']); exit; }
      if (!strpos($apiUrl, '/', 8)) { $apiUrl = rtrim($apiUrl, '/') . '/message/sendText/Central'; }

      $invoice = Capsule::table('tblinvoices')->where('id', $invoiceId)->first();
      $client = Capsule::table('tblclients')->where('id', $invoice->userid)->first();
      
      $possiblePaths = [dirname(__FILE__) . "/models/acceso_cf.php", dirname(__FILE__) . "/model/acceso_cf.php", dirname(__FILE__) . "/acceso_cf.php"];
      foreach ($possiblePaths as $path) { if (file_exists($path)) { require_once($path); break; } }
      
      \AccesoCf\AccesoCf::where('invoice_id', $invoiceId)->delete();
      $acceso_cf = new \AccesoCf\AccesoCf;
      $acceso_cf->invoice_id = $invoiceId;
      $acceso_cf->user_id = $invoice->userid;
      $acceso_cf->clicks = 0;
      $acceso_cf->generate_key();
      $acceso_cf->save();
      
      global $CONFIG;
      $systemurl = ($CONFIG['SystemSSLURL']) ? $CONFIG['SystemSSLURL'].'/' : $CONFIG['SystemURL'].'/';
      $tokenLink = $systemurl . "index.php?m=accesocf&k=" . $acceso_cf->key;
      $template = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappTemplate')->value('value');
      $message = (!empty($whatsappTemplate)) ? processWhatsAppTemplate($whatsappTemplate, ['auto_login_link' => $tokenLink]) : "Factura pendiente: " . $tokenLink;

      $ch = curl_init();
      curl_setopt_array($ch, [
        CURLOPT_URL => $apiUrl,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_POSTFIELDS => json_encode(['number' => $client->phonenumber, 'text' => $message]),
        CURLOPT_HTTPHEADER => ["Content-Type: application/json", "apikey: " . $apiKey],
        CURLOPT_TIMEOUT => 30,
      ]);
      $response = curl_exec($ch);
      $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
      curl_close($ch);
      echo json_encode(['success' => $httpCode < 400, 'message' => 'WhatsApp enviado']);
    } catch (\Exception $e) { echo json_encode(['success' => false, 'message' => $e->getMessage()]); }
    exit;
  }
  
  if ($action === 'send_cancellation') {
    header('Content-Type: application/json');
    try {
        $serviceId = isset($_POST['service_id']) ? (int)$_POST['service_id'] : 0;
        $service = Capsule::table('tblhosting')->where('id', $serviceId)->first();
        $client = Capsule::table('tblclients')->where('id', $service->userid)->first();
        $invoice = Capsule::table('tblinvoices')->where('userid', $service->userid)->orderBy('id', 'desc')->first();
        
        $possiblePaths = [dirname(__FILE__) . "/models/acceso_cf.php", dirname(__FILE__) . "/model/acceso_cf.php", dirname(__FILE__) . "/acceso_cf.php"];
        foreach ($possiblePaths as $path) { if (file_exists($path)) { require_once($path); break; } }

        $acceso_cf = new \AccesoCf\AccesoCf;
        $acceso_cf->invoice_id = $invoice->id;
        $acceso_cf->user_id = $service->userid;
        $acceso_cf->generate_key();
        $acceso_cf->expiration = date('Y-m-d H:i:s', strtotime('+24 hours'));
        $acceso_cf->save();
        
        global $CONFIG;
        $systemurl = ($CONFIG['SystemSSLURL']) ? $CONFIG['SystemSSLURL'].'/' : $CONFIG['SystemURL'].'/';
        $tokenLink = $systemurl . "index.php?m=accesocf&k=" . $acceso_cf->key;
        $message = "Servicio suspendido. Paga aquí: " . $tokenLink;

        $apiUrl = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiUrl')->value('value');
        $apiKey = Capsule::table('tbladdonmodules')->where('module','accesocf')->where('setting','whatsappApiKey')->value('value');
        if (!empty($apiUrl) && !strpos($apiUrl, '/', 8)) { $apiUrl = rtrim($apiUrl, '/') . '/message/sendText/Central'; }

        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $apiUrl,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_POSTFIELDS => json_encode(['number' => $client->phonenumber, 'text' => $message]),
            CURLOPT_HTTPHEADER => ["Content-Type: application/json", "apikey: " . $apiKey],
            CURLOPT_TIMEOUT => 30
        ]);
        curl_exec($ch);
        curl_close($ch);
        echo json_encode(['success' => true, 'message' => 'Cancelación enviada']);
    } catch (\Exception $e) { echo json_encode(['success' => false, 'message' => $e->getMessage()]); }
    exit;
  }

  // --- PANEL DE ADMINISTRACIÓN ---
  $possiblePaths = [dirname(__FILE__) . "/models/acceso_cf.php", dirname(__FILE__) . "/model/acceso_cf.php", dirname(__FILE__) . "/acceso_cf.php"];
  foreach ($possiblePaths as $path) { if (file_exists($path)) { require_once($path); break; } }

  $logs = Capsule::table('accesocf')
    ->leftJoin('tblclients', 'accesocf.user_id', '=', 'tblclients.id')
    ->leftJoin('tblinvoices', 'accesocf.invoice_id', '=', 'tblinvoices.id')
    ->select('accesocf.*', 'tblclients.firstname', 'tblclients.lastname', 'tblclients.email', 'tblinvoices.total')
    ->orderBy('accesocf.expiration', 'desc')->limit(20)->get();

  $total = Capsule::table('accesocf')->count();
  $active = Capsule::table('accesocf')->where('clicks', '>', 0)->count();

  echo '<style>
    .acf-box { font-family: "Inter", sans-serif; background: #f8fafc; padding: 25px; border-radius: 12px; color: #1e293b; max-width: 1000px; margin: 20px 0; border: 1px solid #e2e8f0; }
    .acf-stat-row { display: flex; gap: 15px; margin-bottom: 25px; }
    .acf-stat { background: white; padding: 15px 25px; border-radius: 12px; flex: 1; box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #f1f5f9; }
    .acf-stat b { font-size: 24px; color: #0f172a; display: block; }
    .acf-stat small { color: #64748b; text-transform: uppercase; font-size: 10px; font-weight: 600; letter-spacing: 0.05em; }
    .acf-table { width: 100%; border-collapse: collapse; background: white; border-radius: 12px; overflow: hidden; border: 1px solid #f1f5f9; }
    .acf-table th { background: #f8fafc; padding: 12px 15px; text-align: left; font-size: 11px; color: #64748b; text-transform: uppercase; }
    .acf-table td { padding: 12px 15px; border-bottom: 1px solid #f1f5f9; font-size: 13px; }
    .badge { padding: 4px 10px; border-radius: 9999px; font-size: 10px; font-weight: 600; }
    .badge-success { background: #dcfce7; color: #166534; }
    .badge-pending { background: #fef9c3; color: #854d0e; }
  </style>
  <div class="acf-box">
    <div style="display:flex; align-items:center; gap:12px; margin-bottom:25px;">
      <div style="background:linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); color:white; width:40px; height:40px; border-radius:10px; display:flex; align-items:center; justify-content:center; font-weight:bold; font-size:18px; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.4);">CF</div>
      <div>
        <h2 style="margin:0; font-size:18px; font-weight:700; color:#0f172a;">Acceso Directo CF</h2>
        <p style="margin:0; font-size:12px; color:#64748b;">Monitor de Conversión y Logs de Acceso</p>
      </div>
    </div>
    <div class="acf-stat-row">
      <div class="acf-stat"><small>Tokens Generados</small><b>'.$total.'</b></div>
      <div class="acf-stat"><small>Accesos Reales</small><b>'.$active.'</b></div>
      <div class="acf-stat"><small>Efectividad</small><b>'.($total > 0 ? round(($active / $total) * 100, 1) : 0).'%</b></div>
    </div>
    <table class="acf-table">
      <thead><tr><th>Cliente</th><th>Factura</th><th>Clics</th><th>Estado</th></tr></thead>
      <tbody>';
      if ($logs->isEmpty()) {
          echo '<tr><td colspan="4" style="text-align:center; padding:30px; color:#94a3b8;">No se han generado tokens todavía.</td></tr>';
      } else {
          foreach ($logs as $log) {
            $status = ($log->clicks > 0) ? '<span class="badge badge-success">ACCEDIÓ</span>' : '<span class="badge badge-pending">PENDIENTE</span>';
            $client = trim($log->firstname . " " . $log->lastname) ?: "Cliente eliminado";
            echo "<tr><td><b>$client</b><br><small style='color:#64748b'>$log->email</small></td><td><a href='invoices.php?action=edit&id=$log->invoice_id' target='_blank'>#$log->invoice_id</a></td><td><b>$log->clicks</b></td><td>$status</td></tr>";
          }
      }
      echo '</tbody></table></div>';
}