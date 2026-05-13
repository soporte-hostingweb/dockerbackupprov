'use client';
import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { 
  AlertCircle, Terminal, ShieldCheck, Activity, Network, 
  Database, Server, History, Settings, Cloud, Lock, RotateCcw
} from "lucide-react";

import ServerList from "@/components/ServerList";
import RestoreModal from "@/components/RestoreModal";
import GlobalActivity from "@/components/GlobalActivity"; // V6.3: Telemetría Global
import Onboarding from "@/components/Onboarding"; // V14.2: Onboarding Inteligente

type TabType = 'servers' | 'history' | 'settings' | 'admin';

export default function DashboardPage() {
  const searchParams = useSearchParams();
  const isEmbed = searchParams.get("embed") === "1";
  const isAdminParam = searchParams.get("admin") === "1";
  
  const [activeTab, setActiveTab] = useState<TabType>('servers');
  const [agents, setAgents] = useState<any>({});
  const [agentCount, setAgentCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [activities, setActivities] = useState<any[]>([]);
  const [onboardingAgent, setOnboardingAgent] = useState<any>(null); // V14.2
  
  // Restore Modal State
  const [isRestoreOpen, setIsRestoreOpen] = useState(false);
  const [restoreAgentId, setRestoreAgentId] = useState("");
  const [restoreSnapshots, setRestoreSnapshots] = useState<any[]>([]);

  const [wasabiStatus, setWasabiStatus] = useState<{ status: string; latency_ms: number; bucket: string } | null>(null);
  const [testingWasabi, setTestingWasabi] = useState(false);
  const [testResult, setTestResult] = useState<{message: string, success: boolean} | null>(null);

  const [settings, setSettings] = useState({
    wasabi_key: '',
    wasabi_secret: '',
    wasabi_bucket: '',
    wasabi_region: 'us-east-1',
    s3_endpoint: '',
    s3_force_path_style: true,
    s3_insecure: false,
    restic_password: '',
    webhook_url: '',
    webhook_events: 'backup_failed,agent_offline,restore_completed,verification_failed'
  });
  const [savingSettings, setSavingSettings] = useState(false);

  const token = searchParams.get("sso") || localStorage.getItem("dbp_token") || "";

  useEffect(() => {
    if (searchParams.get("sso")) {
      localStorage.setItem("dbp_token", searchParams.get("sso")!);
    }
  }, [searchParams]);

  const fetchData = async () => {
    if (!token) return;
    try {
      const respStatus = await fetch("https://api.hwperu.com/v1/agent/status", {
        headers: { "Authorization": token }
      });
      if (respStatus.ok) {
        const data = await respStatus.json();
        setAgents(data);
        setAgentCount(Object.keys(data).length);

        // V14.2: Auto-lanzar Onboarding si detectamos un agente nuevo sin config
        if (!onboardingAgent) {
          const freshAgentId = Object.keys(data).find(id => {
            const a = data[id];
            // Verificar si ya fue configurado (guardado en localStorage)
            const alreadyOnboarded = localStorage.getItem(`onboarded_${id}`) === 'true';
            if (alreadyOnboarded) return false;
            // Si es default "Basic" y no tiene snapshots ni contenedores reportados (o es WP detectado)
            return a.protection_level === "Basic" && (!a.paths || a.paths.length === 0);
          });
          if (freshAgentId) setOnboardingAgent(data[freshAgentId]);
        }
      }

      if (activeTab === 'settings') {
          const respSettings = await fetch("https://api.hwperu.com/v1/user/settings", {
            headers: { "Authorization": token }
          });
          if (respSettings.ok) {
            const sData = await respSettings.json();
            setSettings({
              wasabi_key: sData.wasabi_key || '',
              wasabi_secret: sData.wasabi_secret || '',
              wasabi_bucket: sData.wasabi_bucket || '',
              wasabi_region: sData.wasabi_region || 'us-east-1',
              s3_endpoint: sData.s3_endpoint || '',
              s3_force_path_style: sData.s3_force_path_style ?? true,
              s3_insecure: sData.s3_insecure ?? false,
              restic_password: sData.restic_password || '',
              webhook_url: sData.webhook_url || '',
              webhook_events: sData.webhook_events || 'backup_failed,agent_offline,restore_completed,verification_failed'
            });
          }
      }

      if (activeTab === 'history') {
          // V6.6: Apuntamos al nuevo endpoint de Telemetría Pro
          const respHist = await fetch("https://api.hwperu.com/v1/activities", {
            headers: { "Authorization": token }
          });
          if (respHist.ok) {
            const hData = await respHist.json();
            setActivities(hData);
          }
      }
    } catch (error) {
      console.error("Error fetching dashboard data:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // Polling cada 30s
    return () => clearInterval(interval);
  }, [activeTab, token]);

  const testS3Connection = async () => {
    setTestingWasabi(true);
    try {
      const resp = await fetch("https://api.hwperu.com/v1/user/test-wasabi", {
        method: "POST",
        headers: { 
          "Authorization": token,
          "Content-Type": "application/json"
        },
        body: JSON.stringify(settings)
      });
      const data = await resp.json();
      
      if (data.success) {
        setTestResult({
          message: `✅ CONEXIÓN EXITOSA (${data.latency_ms}ms): ${data.message}`,
          success: true
        });
      } else {
        setTestResult({
          message: `❌ ERROR: ${data.error || data.message || 'Falló la conexión'}`,
          success: false
        });
      }
    } catch (err) {
      setTestResult({ message: "❌ Error crítico al contactar con el API", success: false });
    } finally {
      setTestingWasabi(false);
      // Auto-ocultar después de 5 segundos
      setTimeout(() => setTestResult(null), 5000);
    }
  };

  const saveSettings = async (e: any, isGlobal = false) => {
    e.preventDefault();
    setSavingSettings(true);
    try {
      const url = `https://api.hwperu.com/v1/user/settings${isGlobal ? '?is_global=true' : ''}`;
      const resp = await fetch(url, {
        method: "POST",
        headers: { "Authorization": token, "Content-Type": "application/json" },
        body: JSON.stringify(settings)
      });
      if (resp.ok) alert("✅ Configuración guardada correctamente.");
    } catch (err) {
      alert("❌ Error al guardar.");
    } finally {
      setSavingSettings(false);
    }
  };

  const testWasabi = async () => {
    setTestingWasabi(true);
    try {
      const resp = await fetch("https://api.hwperu.com/v1/admin/wasabi/ping", {
        headers: { "Authorization": token }
      });
      if (resp.ok) setWasabiStatus(await resp.json());
    } catch (err) {
      console.error("Test failed:", err);
    } finally {
      setTestingWasabi(false);
    }
  };

  const openRestore = (agentId: string, snapshots: any[]) => {
      setRestoreAgentId(agentId);
      setRestoreSnapshots(snapshots);
      setIsRestoreOpen(true);
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'servers':
        return (
          <div className="space-y-8 animate-in fade-in duration-500">
             <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">Cluster Storage Load</p>
                   <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-black text-white italic">PRO</span>
                    <span className="text-xs text-gray-500 font-bold uppercase tracking-tighter">Edition Active</span>
                  </div>
                </div>
                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl border-l-4 border-l-emerald-500">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">Connected Agents</p>
                  <div className="text-3xl font-black text-white italic">{loading ? "..." : agentCount}</div>
                </div>
                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">API Health Status</p>
                  <div className="flex items-center gap-2">
                     <Activity className="text-emerald-500 animate-pulse" size={20} />
                     <span className="text-xl font-black text-white uppercase italic">Active</span>
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
                  <div className="lg:col-span-3 space-y-8">
                      <ServerList onRestore={openRestore} />
                  </div>
                  <div className="lg:col-span-1">
                      <GlobalActivity token={token} />
                  </div>
              </div>
          </div>
        );

      case 'history':
        return (
          <div className="animate-in slide-in-from-bottom-4 duration-500 space-y-6">
             <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                   <History className="text-emerald-500" size={24} />
                   <h2 className="text-2xl font-black text-white italic uppercase tracking-tighter">Control Absolute Audit Log</h2>
                </div>
                <span className="text-[10px] text-gray-500 font-bold uppercase tracking-widest">{activities.length} entries registered</span>
             </div>

             <div className="bg-gray-950/50 border border-gray-900 rounded-3xl overflow-hidden shadow-2xl">
                <table className="w-full text-left border-collapse">
                   <thead>
                      <tr className="bg-black/40 border-b border-gray-900">
                         <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Event</th>
                         <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Source Agent</th>
                         <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Status</th>
                         <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Timestamp</th>
                         <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Detailed Log Message</th>
                      </tr>
                   </thead>
                   <tbody className="divide-y divide-gray-900/50">
                      {activities.map((act: any) => (
                         <tr key={act.id} className="hover:bg-gray-900/30 transition-colors group">
                            <td className="p-5">
                               <div className="flex items-center gap-2">
                                  {act.type === 'backup' && <Database size={14} className="text-emerald-500" />}
                                  {act.type === 'restore' && <RotateCcw size={14} className="text-sky-500" />}
                                  {act.type === 'prune' && <ShieldCheck size={14} className="text-amber-500" />}
                                  {act.type === 'OFFLINE' && <AlertCircle size={14} className="text-red-500 animate-pulse" />}
                                  {act.type === 'DELETED' && <Lock size={14} className="text-gray-500" />}
                                  <span className="text-xs font-black text-white uppercase italic tracking-tighter">{act.type}</span>
                               </div>
                            </td>
                            <td className="p-5">
                               <span className="text-[11px] font-mono text-gray-400 group-hover:text-emerald-400 transition-colors">{act.agent_id}</span>
                            </td>
                            <td className="p-5">
                                <span className={`px-3 py-1 rounded-full text-[9px] font-black uppercase border ${
                                   act.status === 'success' ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' :
                                   act.status === 'error' ? 'bg-red-500/10 text-red-500 border-red-500/20' :
                                   'bg-amber-500/10 text-amber-500 border-amber-500/20'
                                }`}>
                                   {act.status}
                                </span>
                            </td>
                            <td className="p-5 text-[10px] text-gray-500 font-bold whitespace-nowrap">
                               {new Date(act.started_at).toLocaleString()}
                            </td>
                            <td className="p-5">
                               <p className="text-[11px] text-gray-300 font-medium max-w-md line-clamp-1 group-hover:line-clamp-none transition-all">
                                  {act.message}
                               </p>
                            </td>
                         </tr>
                      ))}
                   </tbody>
                </table>
             </div>

             {activities.length === 0 && (
                <div className="bg-gray-950/20 border-2 border-dashed border-gray-900 rounded-3xl p-20 text-center">
                   <History className="mx-auto text-gray-800 mb-4 animate-pulse" size={48} />
                   <p className="text-sm text-gray-600 font-bold uppercase italic tracking-widest">No system events recorded yet.</p>
                </div>
             )}
          </div>
        );

      case 'settings':
        return (
          <div className="max-w-2xl mx-auto animate-in zoom-in-95 duration-500">
              <form 
                onSubmit={(e) => { e.preventDefault(); saveSettings(e); }}
                className="bg-gray-950/50 border border-gray-900 rounded-3xl p-8 shadow-2xl space-y-6"
              >
                <div className="flex items-center gap-3 mb-2">
                    <div className="p-3 bg-emerald-500/10 rounded-xl border border-emerald-500/20 text-emerald-500 shadow-[0_0_15px_rgba(16,185,129,0.1)]"><Cloud size={24} /></div>
                    <h3 className="text-lg font-black text-white uppercase italic tracking-widest">S3 Tenant Configuration</h3>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Access Key</label>
                      <input type="text" value={settings.wasabi_key} onChange={(e) => setSettings({...settings, wasabi_key: e.target.value})} className="premium-input w-full font-mono text-emerald-400" placeholder="AKIA..." />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Secret Key</label>
                      <input type="password" value={settings.wasabi_secret} onChange={(e) => setSettings({...settings, wasabi_secret: e.target.value})} className="premium-input w-full font-mono text-emerald-400" placeholder="••••••••" />
                   </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Bucket Name</label>
                      <input type="text" value={settings.wasabi_bucket} onChange={(e) => setSettings({...settings, wasabi_bucket: e.target.value})} className="premium-input w-full font-mono" placeholder="my-backups" />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Restic Password</label>
                      <input type="password" value={settings.restic_password} onChange={(e) => setSettings({...settings, restic_password: e.target.value})} className="premium-input w-full font-mono text-emerald-300 border-emerald-500/20" placeholder="••••••••" />
                   </div>
                </div>

                <div className="space-y-2">
                   <div className="flex items-center justify-between">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">S3 Custom Endpoint (Optional)</label>
                      <span className="text-[9px] text-emerald-500/60 font-bold uppercase italic bg-emerald-500/5 px-2 py-0.5 rounded-full border border-emerald-500/10">No insertar s3:https://</span>
                   </div>
                   <input 
                     type="text" 
                     placeholder="s3.ca-central-1.wasabisys.com" 
                     value={settings.s3_endpoint} 
                     onChange={(e) => setSettings({...settings, s3_endpoint: e.target.value})} 
                     className="premium-input w-full font-mono text-blue-400" 
                   />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="flex items-center justify-between p-4 bg-white/5 border border-white/5 rounded-2xl hover:bg-white/10 transition-all group">
                       <div className="flex flex-col">
                          <span className="text-[10px] text-white font-black uppercase tracking-widest">Force Path Style</span>
                          <span className="text-[9px] text-gray-500 uppercase font-bold">Required for Minio / Private S3</span>
                       </div>
                       <label className="switch">
                          <input 
                            type="checkbox" 
                            checked={settings.s3_force_path_style}
                            onChange={(e) => setSettings({...settings, s3_force_path_style: e.target.checked})}
                          />
                          <span className="slider"></span>
                       </label>
                    </div>

                    <div className="flex items-center justify-between p-4 bg-white/5 border border-white/5 rounded-2xl hover:bg-white/10 transition-all group">
                       <div className="flex flex-col">
                          <span className="text-[10px] text-white font-black uppercase tracking-widest">Skip SSL Verify</span>
                          <span className="text-[9px] text-red-500/60 uppercase font-bold italic">Insecure Handshake</span>
                       </div>
                       <label className="switch">
                          <input 
                            type="checkbox" 
                            checked={settings.s3_insecure}
                            onChange={(e) => setSettings({...settings, s3_insecure: e.target.checked})}
                          />
                          <span className="slider"></span>
                       </label>
                    </div>
                </div>

                 <div className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2">
                    <button 
                      type="button" 
                      onClick={testS3Connection} 
                      disabled={testingWasabi || savingSettings} 
                      className="md:col-span-1 bg-white/5 hover:bg-white/10 text-emerald-400 font-black uppercase text-[10px] py-4 rounded-2xl transition-all border border-emerald-500/20 glow-emerald disabled:opacity-50"
                    >
                       {testingWasabi ? '🔄 Testing...' : 'Test Connection'}
                    </button>
                    <button 
                      type="button"
                      onClick={(e) => saveSettings(e)}
                      disabled={savingSettings}
                      className="md:col-span-3 bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-[11px] py-4 rounded-2xl transition-all shadow-[0_0_20px_rgba(16,185,129,0.2)] hover:shadow-[0_0_30px_rgba(16,185,129,0.4)] disabled:opacity-50"
                    >
                       {savingSettings ? '💾 Saving...' : 'Save Configuration'}
                    </button>
                 </div>

                 {testResult && (
                    <div className={`text-center p-3 rounded-xl border animate-in fade-in slide-in-from-top-2 duration-300 ${testResult.success ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : 'bg-red-500/10 border-red-500/20 text-red-400'}`}>
                        <p className="text-[10px] font-black uppercase tracking-widest">{testResult.message}</p>
                    </div>
                 )}

                <div className="pt-6 border-t border-gray-900 space-y-4">
                    <div className="flex items-center gap-3">
                        <div className="p-3 bg-amber-500/10 rounded-xl border border-amber-500/20 text-amber-500"><AlertCircle size={24} /></div>
                        <h3 className="text-lg font-black text-white uppercase italic tracking-widest">Universal Webhook (n8n / API)</h3>
                    </div>
                    <p className="text-[10px] text-gray-500 font-bold uppercase tracking-wider">Desacopla tus notificaciones. Enviamos un POST con el evento estructurado hacia tu URL configurada en n8n.</p>
                    
                    <div className="space-y-2">
                        <label className="text-[10px] text-gray-500 font-black uppercase">Webhook URL (Target)</label>
                        <input 
                          type="text" 
                          placeholder="https://tu-instancia-n8n.com/webhook/..."
                          value={settings.webhook_url} 
                          onChange={(e) => setSettings({...settings, webhook_url: e.target.value})} 
                          className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-blue-400 focus:border-blue-500 outline-none font-mono" 
                        />
                    </div>

                    <div className="space-y-3">
                        <label className="text-[10px] text-gray-500 font-black uppercase">Subscribed Events</label>
                        <div className="grid grid-cols-2 gap-3">
                            {['backup_failed', 'agent_offline', 'restore_completed', 'verification_failed'].map(ev => (
                                <label key={ev} className="flex items-center gap-2 cursor-pointer group">
                                    <input 
                                      type="checkbox" 
                                      checked={settings.webhook_events.includes(ev)}
                                      onChange={(e) => {
                                          const current = settings.webhook_events.split(',').filter(x => x);
                                          const next = e.target.checked ? [...current, ev] : current.filter(x => x !== ev);
                                          setSettings({...settings, webhook_events: next.join(',')});
                                      }}
                                      className="sr-only"
                                    />
                                    <div className={`w-4 h-4 rounded border flex items-center justify-center transition-all ${settings.webhook_events.includes(ev) ? 'bg-blue-500 border-blue-500' : 'border-gray-800 bg-black/40 group-hover:border-gray-600'}`}>
                                        {settings.webhook_events.includes(ev) && <div className="w-1.5 h-1.5 bg-white rounded-full"></div>}
                                    </div>
                                    <span className={`text-[10px] font-bold uppercase transition-colors ${settings.webhook_events.includes(ev) ? 'text-white' : 'text-gray-600'}`}>{ev.replace('_', ' ')}</span>
                                </label>
                            ))}
                        </div>
                    </div>
                </div>
              </form>
          </div>
        );

      case 'admin':
        return (
          <div className="space-y-8 animate-in fade-in duration-500">
             <div className="bg-emerald-950/10 border border-emerald-900/30 rounded-2xl p-8 flex items-center justify-between">
                <div className="flex items-center gap-4">
                   <Network className="text-emerald-500" size={32} />
                   <div>
                      <h3 className="text-xl font-black text-white italic uppercase tracking-tighter">Global Cloud Health</h3>
                      <p className="text-[10px] text-gray-500 font-bold uppercase tracking-widest mt-1">Multi-Region S3 Connectivity Pulse</p>
                   </div>
                </div>
                <button onClick={testWasabi} disabled={testingWasabi} className="bg-emerald-600 hover:bg-emerald-500 text-white text-[10px] font-black px-6 py-3 rounded-xl border border-emerald-400/20 uppercase transition-all shadow-xl shadow-emerald-950/50">
                  {testingWasabi ? 'PROBING...' : 'VERIFY CLOUD LINK'}
                </button>
             </div>
          </div>
        );
    }
  };

  return (
    <div className={`mx-auto space-y-12 ${isEmbed ? 'max-w-full p-2' : 'max-w-7xl p-8'}`}>
      
      <div className="flex flex-col md:flex-row justify-between items-center bg-gray-950/50 p-8 rounded-[2.5rem] border border-gray-900 shadow-2xl gap-8 shadow-emerald-950/5">
        <div className="flex flex-col">
          <h1 className="text-4xl md:text-6xl font-black tracking-tighter text-white uppercase italic">
            HW CLOUD <span className="text-blue-500">RECOVERY</span>
          </h1>
          <p className="text-[10px] text-gray-500 uppercase tracking-[0.4em] mt-1 font-black leading-none">Enterprise Disaster Recovery as a Service (DRaaS)</p>
        </div>

        <div className="flex bg-black/40 p-2 rounded-2xl border border-gray-800">
           {(['servers', 'history', 'settings'] as TabType[]).map((tab) => (
             <button key={tab} onClick={() => setActiveTab(tab)} className={`flex items-center gap-2 px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all ${activeTab === tab ? 'bg-emerald-600 text-white shadow-lg' : 'text-gray-500 hover:text-gray-300'}`}>
                {tab === 'servers' && <Server size={14} />}
                {tab === 'history' && <History size={14} />}
                {tab === 'settings' && <Settings size={14} />}
                {tab.charAt(0).toUpperCase() + tab.slice(1)}
             </button>
           ))}
           {isAdminParam && !isEmbed && (
             <button onClick={() => setActiveTab('admin')} className={`flex items-center gap-2 px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all ${activeTab === 'admin' ? 'bg-red-600 text-white shadow-lg' : 'text-red-500/60 hover:text-red-400'}`}>
                <Network size={14} /> Admin
             </button>
           )}
        </div>
      </div>

      <div className="min-h-[60vh]">{renderTabContent()}</div>

      <RestoreModal 
        isOpen={isRestoreOpen} 
        onClose={() => { setIsRestoreOpen(false); fetchData(); }} 
        agentId={restoreAgentId} 
        snapshots={restoreSnapshots}
        token={token}
      />

      {onboardingAgent && (
        <Onboarding 
          agentId={onboardingAgent.agent_id}
          detectedStack={onboardingAgent.detected_stack}
          onCancel={() => {
            if (onboardingAgent?.agent_id) {
              localStorage.setItem(`onboarded_${onboardingAgent.agent_id}`, 'true');
            }
            setOnboardingAgent(null);
          }}
          onComplete={async (config) => {
            try {
              // Guardar configuración del preset
              let paths = ["/host_root"]; // Default full
              if (config.mode === 'wordpress') {
                paths = ["[ALL_SYSTEM_ROOT]"]; // Alias para WP Preset en el backend
              }
              
              await fetch(`https://api.hwperu.com/v1/agent/config/save`, {
                method: "POST",
                headers: { "Authorization": token, "Content-Type": "application/json" },
                body: JSON.stringify({
                  agent_id: onboardingAgent.agent_id,
                  protection_level: config.protection_level,
                  paths: paths,
                  schedule: "daily_2am_basic",
                  is_auto_managed: true
                })
              });
              localStorage.setItem(`onboarded_${onboardingAgent.agent_id}`, 'true');
              setOnboardingAgent(null);
              fetchData();
            } catch (err) {
              alert("Error al guardar configuración inicial.");
            }
          }}
        />
      )}
    </div>
  );
}
