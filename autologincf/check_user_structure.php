<?php
define("WHMCS", true);
// Usar la ruta absoluta al root de WHMCS (3 niveles arriba desde addons/accesocf)
require_once(dirname(__FILE__) . "/../../init.php");
use Illuminate\Database\Capsule\Manager as Capsule;

$clientId = 1228;
echo "Investigando sesión para Cliente: $clientId\n";

$userLink = Capsule::table('tblusers_clients')->where('client_id', $clientId)->first();
if ($userLink) {
    $user = Capsule::table('tblusers')->where('id', $userLink->auth_user_id)->first();
    if ($user) {
        echo "User ID: " . $user->id . "\n";
        echo "User Email: " . $user->email . "\n";
        echo "Has Password in tblusers: " . (!empty($user->password) ? 'Yes' : 'No') . "\n";
        
        $client = Capsule::table('tblclients')->where('id', $clientId)->first();
        if ($client) {
            echo "Has Password in tblclients: " . (!empty($client->password) ? 'Yes' : 'No') . "\n";
            echo "Passwords Match: " . ($user->password === $client->password ? 'Yes' : 'No') . "\n";
        }
    } else {
        echo "ERROR: User ID {$userLink->auth_user_id} referenced but not found in tblusers.\n";
    }
} else {
    echo "ERROR: No link found for client $clientId in tblusers_clients.\n";
}
