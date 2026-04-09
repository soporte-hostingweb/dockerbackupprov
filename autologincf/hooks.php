<?php
use Illuminate\Database\Capsule\Manager as Capsule;
use AccesoCf\AccesoCf;

function create_invoice_autologin_link($vars) {
  require_once(dirname(__FILE__) . "/models/acceso_cf.php");
  
  // Lista de plantillas de factura comunes o verificar por relid si es factura
  $invoiceEmailTemplates = [
    'Invoice Created', 
    'Invoice Payment Reminder', 
    'First Invoice Overdue Notice', 
    'Second Invoice Overdue Notice', 
    'Third Invoice Overdue Notice', 
    'Credit Card Payment Failed',
    'Invoice Payment Confirmation'
  ];

  $isInvoiceEmail = ($vars['type'] == 'invoice') || 
                    in_array($vars['messagename'], $invoiceEmailTemplates);

  if ($isInvoiceEmail && !empty($vars['relid'])) {
    $invoiceId = $vars['relid'];
    $invoice = Capsule::table('tblinvoices')->where('id', $invoiceId)->first();
    
    if ($invoice) {
      // Verificar si ya existe un token para esta factura y si todavía es válido
      $existing_token = \AccesoCf\AccesoCf::where('invoice_id', $invoice->id)->first();
      
      $is_valid = $existing_token && 
                  strtotime($existing_token->expiration) > time() && 
                  (!isset($existing_token->clicks) || $existing_token->clicks < 10);

      if ($is_valid) {
        $acceso_cf = $existing_token;
      } else {
        $acceso_cf = new AccesoCf;
        $acceso_cf->invoice_id = $invoice->id;
        $acceso_cf->user_id = $invoice->userid;
        $acceso_cf->generate_key();
        $acceso_cf->save();
      }

      global $CONFIG;
      $systemurl = ($CONFIG['SystemSSLURL']) ? $CONFIG['SystemSSLURL'].'/' : $CONFIG['SystemURL'].'/';
      $autolink = $systemurl ."index.php?m=accesocf&k=".$acceso_cf->key;
      $autolink_html = "<a href='".$autolink."' style='background-color: #007cba; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;'>Pagar Factura</a>";
      
      return array(
        "auto_login_link" => $autolink,
        "auto_login_link_html" => $autolink_html
      );
    }
  }
  
  return array();
}

function remove_autologin_links($vars) {
  // Set all invoice login links to expired since the invoice is now paid.
  require_once(dirname(__FILE__) . "/models/acceso_cf.php");
 
  \AccesoCf\AccesoCf::where('invoice_id', $vars['invoiceid'])->update(array("expiration" => date('Y-m-d',time())));
}

function disable_non_invoice_pages($vars) {
  // Ignorar si no hay sesión de autologin activa
  if (!isset($_SESSION['used_invoice_autologin'])) {
      return;
  }

  // EXCEPCIÓN CRÍTICA: No restringir si estamos en el propio módulo accesocf
  if (isset($_GET['m']) && $_GET['m'] == 'accesocf') {
      return;
  }

  $option = Capsule::table('tbladdonmodules')->where('module', 'accesocf')->where('setting', 'option1')->first();
  $isRestricted = $option && $option->value == "on";

  if ($isRestricted) {
      $allowedPages = ['invoice-payment', 'viewinvoice', 'clientarea'];
      $allowedFiles = ['viewinvoice', 'dologin', 'clientarea', 'index', 'login'];

      $templateFile = $vars['templatefile'] ?? '';
      $fileName = $vars['filename'] ?? '';

      // Si la página no está permitida, redirigir a login
      if (!in_array($templateFile, $allowedPages) && !in_array($fileName, $allowedFiles)) {
          logActivity("AccesoCF: Restricción activada. Bloqueado acceso a Template: $templateFile | File: $fileName. Redirigiendo a login.");
          
          unset($_SESSION['uid'], $_SESSION['upw'], $_SESSION['login_auth_tk'], $_SESSION['used_invoice_autologin']);
          header('Location: login.php');
          exit();
      }
  }
}


add_hook("EmailPreSend", 1, "create_invoice_autologin_link");
add_hook("InvoicePaid",1,"remove_autologin_links");
add_hook("InvoiceCancelled",1,"remove_autologin_links");
// add_hook("ClientAreaPage",1,"disable_non_invoice_pages");






function generateUniqueToken() {
    return bin2hex(random_bytes(20));
}

function obtenerUserIdPorInvoiceId($invoiceId) {
    $invoice = Capsule::table('tblinvoices')->where('id', $invoiceId)->first();
    return $invoice ? $invoice->userid : null;
}

function updateOrCreateToken($invoiceId, $userId) {
    $accesoCf = AccesoCf::where('invoice_id', $invoiceId)->first();

    if ($accesoCf) {
        $accesoCf->key = generateUniqueToken();
        $accesoCf->expiration = date('Y-m-d', strtotime('+30 days'));
        $accesoCf->save();
    } else {
        $accesoCf = new AccesoCf([
            'invoice_id' => $invoiceId,
            'user_id' => $userId,
            'key' => generateUniqueToken(),
            'expiration' => date('Y-m-d', strtotime('+30 days'))
        ]);
        $accesoCf->save();
        }   
    return $accesoCf->key;
}




add_hook('AdminClientServicesTabFields', 1, function($vars) {
    $serviceId = (int)$vars['id'];
    
    // Check if service exists
    if (!$serviceId) return [];

    // Cargar modelo
    require_once(dirname(__FILE__) . "/models/acceso_cf.php");

    // Button definition with updated JS
    $btnHtml = <<<HTML
    <button type="button" class="btn btn-default" id="sendCancelWhatsappBtn" style="color: #d9534f; border-color: #d9534f;"><i class="fas fa-whatsapp"></i> Enviar Cancelación</button>
    <script type="text/javascript">
    $(document).ready(function() {
        // Try to find the DiskNotify buttons container
        var diskNotifyBtn = $('#sendNotification');
        
        if (diskNotifyBtn.length > 0) {
            // Found DiskNotify! Move our button next to it.
            var container = diskNotifyBtn.parent(); // The flex div
            var myBtn = $('#sendCancelWhatsappBtn');
            
            // Append to the flex container
            container.append(myBtn);
            
            // Hide the row that WHMCS created for our field since it's now empty
            myBtn.closest('tr').hide();
        }

        $('#sendCancelWhatsappBtn').click(function(e) {
            e.preventDefault();
            if(!confirm('¿Estás seguro de enviar la notificación de cancelación por WhatsApp? Esto generará un nuevo token de 24h.')) {
                return;
            }
            
            var btn = $(this);
            var originalText = btn.html();
            btn.prop('disabled', true).text('Enviando...');
            
            $.ajax({
                url: 'addonmodules.php?module=accesocf&action=send_cancellation',
                type: 'POST',
                dataType: 'json',
                data: { service_id: {$serviceId} },
                success: function(resp){
                    if(resp.success){
                        alert('Mensaje enviado por WhatsApp.');
                    } else {
                        alert('Error: ' + (resp.message || 'Desconocido'));
                        if(resp.debug) console.log(resp.debug);
                    }
                },
                error: function(xhr){
                    alert('Error de red al enviar WhatsApp');
                },
                complete: function() {
                    btn.prop('disabled', false).html(originalText);
                }
            });
        });
    });
    </script>
HTML;

    return ['AccesoCF' => $btnHtml];
});


?>