// hook_accesocf_tagmanager.php Esta dentro include/hooks/hook.php


<?php
require_once(dirname(__FILE__) . "/../../modules/addons/accesocf/models/acceso_cf.php");

add_hook('AdminAreaHeadOutput', 1, function($vars) {
    if ($vars['filename'] == 'invoices' && isset($_GET['action']) && $_GET['action'] == 'edit') {
        $invoiceId = (int) $_GET['id'];
        
        // Log de actividad para debugging
        logActivity("AccesoCF: Hook AdminAreaHeadOutput ejecutado para factura ID: $invoiceId");

        return <<<EOT
<script type="text/javascript">
    $(document).ready(function() {
        var btnHtml = `
            <div class="btn-group btn-group-sm">
                <button type="button" class="btn btn-default btn-sm" id="sendWhatsappBtn">
                    📲 Enviar WhatsApp
                </button>
            </div>
        `;
        $('.pull-right-md-larger').append(btnHtml);

        $('#sendWhatsappBtn').click(function() {
            console.log('AccesoCF: Botón WhatsApp clickeado para factura ID: {$invoiceId}');
            
            if (!confirm('¿Deseas generar un nuevo token (si está vencido) y enviar WhatsApp al cliente y sus contactos de facturación?')) {
                console.log('AccesoCF: Usuario canceló el envío de WhatsApp');
                return;
            }
            
            console.log('AccesoCF: Iniciando envío de WhatsApp...');
            
            $.ajax({
                url: 'addonmodules.php?module=accesocf&action=send_whatsapp',
                type: 'POST',
                dataType: 'json',
                data: { invoice_id: {$invoiceId} },
                beforeSend: function() {
                    console.log('AccesoCF: Enviando petición AJAX...');
                },
                success: function(resp){
                    console.log('AccesoCF: Respuesta recibida:', resp);
                    if(resp.success){
                        alert('✅ ' + resp.message);
                        console.log('AccesoCF: WhatsApp enviado exitosamente');
                    } else {
                        alert('❌ ' + (resp.message || 'No se pudo enviar el mensaje.'));
                        console.log('AccesoCF: Error al enviar WhatsApp:', resp.message);
                    }
                },
                error: function(xhr){
                    console.log('AccesoCF: Error de red:', xhr);
                    alert('❌ Error de red al intentar enviar WhatsApp.');
                }
            });
        });
    });
</script>
EOT;
    }
});

//add_hook('ClientAreaFooterOutput', 1, function ()
add_hook('ClientAreaHeadOutput', 1, function ()
{
    $jsCode = <<<EOT
<!-- Google Tag Manager -->
<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
+'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','GTM-NSS9K5GT');</script>
<!-- End Google Tag Manager -->
EOT;
		
		return $jsCode;
});

add_hook('ClientAreaHeadOutput', 1, function($vars)

{
	$SystemURL = 'https://cliente.hwperu.com';
	$url = "https://$_SERVER[HTTP_HOST]$_SERVER[REQUEST_URI]";
	$url = strtok($url, '?');
//	$Page = 'HWPeru';

	return <<<HTML
	<link rel="canonical" href="{$url}"/>
HTML;

});



function insert_body_code_hook($vars)
{
	return <<<HTML
	
<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id=GTM-NSS9K5GT"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->

HTML;
};
add_hook("ClientAreaHeaderOutput",1,"insert_body_code_hook");
