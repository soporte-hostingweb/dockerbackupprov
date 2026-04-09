<?php
define("WHMCS", true);
require_once(__DIR__ . "/../../init.php");
use Illuminate\Database\Capsule\Manager as Capsule;

$key = '39309da0f527ea185ea0c4429e1155be8e7c5d94';
$token = Capsule::table('accesocf')->where('key', $key)->first();

echo "--- TOKEN STATUS ---\n";
if ($token) {
    echo "Key: " . $token->key . "\n";
    echo "User ID: " . $token->user_id . "\n";
    echo "Invoice ID: " . $token->invoice_id . "\n";
    echo "Clicks: " . $token->clicks . "\n";
    echo "Expiration: " . $token->expiration . "\n";
    
    $expires = strtotime($token->expiration);
    $now = time();
    echo "Expires at: " . date("Y-m-d H:i:s", $expires) . " (Timestamp: $expires)\n";
    echo "Current time: " . date("Y-m-d H:i:s", $now) . " (Timestamp: $now)\n";
    
    if ($expires < $now) {
        echo "RESULT: TOKEN EXPIRED\n";
    } else if ($token->clicks >= 10) {
        echo "RESULT: CLICK LIMIT REACHED\n";
    } else {
        echo "RESULT: TOKEN VALID\n";
    }
} else {
    echo "RESULT: TOKEN NOT FOUND\n";
    
    // Veamos los últimos tokens generados para comparar
    echo "\n--- RECENT TOKENS ---\n";
    $recent = Capsule::table('accesocf')->orderBy('expiration', 'desc')->limit(5)->get();
    foreach ($recent as $r) {
        echo "Key: " . substr($r->key, 0, 10) . "... | Inv: " . $r->invoice_id . " | Exp: " . $r->expiration . "\n";
    }
}
echo "Timezone: " . date_default_timezone_get() . "\n";
