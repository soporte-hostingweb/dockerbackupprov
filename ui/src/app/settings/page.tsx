'use client';
import { useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Settings, Shield, Cloud, Lock, Save, Globe } from "lucide-react";

export default function SettingsPage() {
  const searchParams = useSearchParams();
  const isAdmin = searchParams.get("admin") === "1";
  const sso = searchParams.get("sso");
  
  const [settings, setSettings] = useState({
    wasabi_key: '',
    wasabi_secret: '',
    wasabi_bucket: '',
    wasabi_region: 'us-east-1',
    restic_password: ''
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{success: boolean, message?: string, error?: string, details?: string} | null>(null);


  useEffect(() => {
    if (!isAdmin) return;
    
    async function fetchGlobalSettings() {
      try {
        // Obtenemos los settings específicamente Globales del sistema
        const resp = await fetch("https://api.hwperu.com/v1/user/settings?mode=global", {
          headers: { "Authorization": sso || "" }
        });
        if (resp.ok) {
          const data = await resp.json();
          setSettings({
            wasabi_key: data.wasabi_key || '',
            wasabi_secret: data.wasabi_secret || '',
            wasabi_bucket: data.wasabi_bucket || '',
            wasabi_region: data.wasabi_region || 'us-east-1',
            restic_password: data.restic_password || ''
          });
        }
      } catch (e) {
        console.error("Global settings fetch failed", e);
      } finally {
        setLoading(false);
      }
    }
    fetchGlobalSettings();
  }, [isAdmin, sso]);

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const resp = await fetch("https://api.hwperu.com/v1/user/test-wasabi", {
        method: "POST",
        headers: { 
          "Authorization": sso || "",
          "Content-Type": "application/json"
        },
        body: JSON.stringify(settings)
      });
      const data = await resp.json();
      setTestResult(data);
    } catch (err) {
      setTestResult({ success: false, error: "Network or API failure during test." });
    } finally {
      setTesting(false);
    }
  };

  const saveGlobal = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const resp = await fetch("https://api.hwperu.com/v1/user/settings?is_global=true", {
        method: "POST",
        headers: { 
          "Authorization": sso || "",
          "Content-Type": "application/json"
        },
        body: JSON.stringify(settings)
      });
      if (resp.ok) {
        alert("✅ CONFIGURACIÓN MAESTRA GLOBAL ACTUALIZADA.\nTodos los clientes sin configuración propia usarán estos datos.");
      }
    } catch (err) {
      alert("❌ Error al guardar.");
    } finally {
      setSaving(false);
    }
  };


  if (!isAdmin) {
    return <div className="p-20 text-center text-red-500 font-black uppercase italic">Access Denied: Admin privileges required.</div>;
  }

  return (
    <div className="max-w-4xl mx-auto p-8 space-y-8 animate-in fade-in duration-700">
      <div className="flex flex-col md:flex-row justify-between items-end gap-4 border-b border-gray-900 pb-8">
          <div>
            <h1 className="text-4xl font-black tracking-tighter text-white italic flex items-center gap-3 uppercase">
                <Globe className="text-emerald-500" size={36} />
                Global Infrastructure
            </h1>
            <p className="text-[10px] text-gray-500 uppercase tracking-[0.4em] mt-2 font-black">
                Master SaaS Orchestration & Default Storage
            </p>
          </div>
          <div className="bg-emerald-500/10 border border-emerald-500/20 px-4 py-2 rounded-xl">
             <span className="text-[10px] font-black text-emerald-500 uppercase tracking-widest flex items-center gap-2">
                <Shield size={12} />
                Master Admin Mode
             </span>
          </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-1 space-y-4 text-gray-400">
             <h3 className="text-white font-black uppercase text-xs tracking-widest italic">¿Qué es esto?</h3>
             <p className="text-[11px] leading-relaxed">
                Aquí defines el almacenamiento **Wasabi S3 Central**. 
                Si un cliente no configura su propio bucket, el sistema usará estos datos para proteger sus servidores automáticamente.
             </p>
             <div className="p-4 bg-gray-950 rounded-xl border border-gray-900 text-[10px] font-bold text-gray-600 italic">
                Nota: Los datos se guardan con cifrado AES-256-GCM y nunca se exponen al cliente.
             </div>
          </div>

          <form onSubmit={saveGlobal} className="md:col-span-2 bg-gray-950 border border-gray-900 p-8 rounded-3xl shadow-2xl space-y-6">
             <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-2">
                   <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest">Master Wasabi Key</label>
                   <input 
                     type="text" value={settings.wasabi_key} onChange={(e)=>setSettings({...settings, wasabi_key: e.target.value})}
                     className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-emerald-500 outline-none transition-all font-mono"
                   />
                </div>
                <div className="space-y-2">
                   <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest">Master Wasabi Secret</label>
                   <input 
                     type="password" value={settings.wasabi_secret} onChange={(e)=>setSettings({...settings, wasabi_secret: e.target.value})}
                     className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-emerald-500 outline-none transition-all font-mono"
                   />
                </div>
             </div>

             <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-2">
                   <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest">Master Bucket Name</label>
                   <input 
                     type="text" value={settings.wasabi_bucket} onChange={(e)=>setSettings({...settings, wasabi_bucket: e.target.value})}
                     className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-emerald-500 outline-none transition-all font-mono"
                   />
                </div>
                <div className="space-y-2">
                   <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest">Master Region</label>
                   <input 
                     type="text" value={settings.wasabi_region} onChange={(e)=>setSettings({...settings, wasabi_region: e.target.value})}
                     className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-emerald-500 outline-none transition-all font-mono"
                   />
                </div>
             </div>

             <div className="space-y-2 pt-4 border-t border-gray-900">
                <label className="text-[10px] text-emerald-500 font-black uppercase tracking-widest flex items-center gap-2 italic">
                   <Lock size={12} /> Master Encryption Phrasename
                </label>
                <input 
                  type="password" value={settings.restic_password} onChange={(e)=>setSettings({...settings, restic_password: e.target.value})}
                  className="w-full bg-emerald-950/10 border border-emerald-900/30 rounded-xl px-4 py-3 text-sm text-emerald-200 focus:border-emerald-500 outline-none transition-all font-mono"
                />
             </div>

              {testResult && (
                <div className={`p-4 rounded-xl border text-[11px] leading-relaxed animate-in zoom-in duration-300 ${testResult.success ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-red-500/10 border-red-500/30 text-red-400'}`}>
                   <p className="font-black uppercase mb-1">{testResult.success ? 'Success' : 'Connection Error'}</p>
                   <p>{testResult.message || testResult.error}</p>
                   {testResult.details && <p className="mt-2 text-gray-500 italic opacity-80">{testResult.details}</p>}
                </div>
              )}

              <div className="flex gap-4">
                <button 
                  type="button" disabled={testing}
                  onClick={handleTest}
                  className="flex-1 bg-gray-900 hover:bg-gray-800 text-gray-400 font-black uppercase text-xs py-4 rounded-2xl border border-gray-800 transition-all flex items-center justify-center gap-2"
                >
                    <Cloud size={16} className={testing ? 'animate-pulse' : ''} />
                    {testing ? 'PINGING WASABI...' : 'TEST CONNECTION'}
                </button>

                <button 
                  type="submit" disabled={saving}
                  className="flex-[2] bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-xs py-4 rounded-2xl shadow-xl shadow-emerald-900/40 transition-all active:scale-[0.98] flex items-center justify-center gap-2"
                >
                    <Save size={16} />
                    {saving ? 'UPDATING GLOBAL FABRIC...' : 'SAVE MASTER CONFIGURATION'}
                </button>
              </div>

          </form>
      </div>
    </div>
  );
}
