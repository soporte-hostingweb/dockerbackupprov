<?php
namespace AccesoCf;
use Illuminate\Database\Capsule\Manager as Capsule;

class AccesoCf extends \Illuminate\Database\Eloquent\Model {	
  
  protected $fillable = ['key', 'user_id','invoice_id','expiration','clicks'];
  public $timestamps = false;
  protected $table = 'accesocf';
  protected $primaryKey = 'key';
  public $incrementing = false;
 

  public function client() {
     return $this->belongsTo('\WHMCS\User\Client','user_id');
  }

  function generate_key() {
    $addonSettings = Capsule::table('tbladdonmodules')->where('module','accesocf')->get();
    
    $key2 = $addonSettings->where('setting','key')->first()->value ?? uniqid();
    $expiration_days = $addonSettings->where('setting','expiration_days')->first()->value ?? 20;

    if (empty($expiration_days)) {
        $expiration_days = 20;
    }
    $key_created = false;   
    while (!$key_created) { 
      $key = uniqid();
      $timestamp = time();
      $key = sha1($key.$timestamp.$key2);
      \AccesoCf\AccesoCf::where('key', $key)->count();
      if (\AccesoCf\AccesoCf::where('key', $key)->count() == 0) {
	    $key_created = true;
      }
    }
    $this->key = $key;
    $this->expiration = date('Y-m-d H:i:s', time() + ($expiration_days * 86400));
    
    return true;
  }
  

}
?>
