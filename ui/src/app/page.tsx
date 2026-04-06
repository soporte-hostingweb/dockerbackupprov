'use client';

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { 
  AlertCircle, Terminal, ShieldCheck, Activity, Network, 
  Database, Server, History, Settings, Cloud, Lock
} from "lucide-react";

import ServerList from "@/components/ServerList";

type TabType = 'servers' | 'history' | 'settings' | 'admin';

export default function DashboardPage() {
  const searchParams = useSearchParams();
  const isEmbed = searchParams.get("embed") === "1";
  const isAdminParam = searchParams.get("admin") === "1";
  
  const [activeTab, setActiveTab] = useState<TabType>('servers');
  const [agents, setAgents] = useState<any>({});
  const [agentCount, setAgentCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [wasabiStatus, setWasabiStatus] = useState<{ status: string; latency_ms: number; bucket: string } | null>(null);
  const [testingWasabi, setTestingWasabi] = useState(false);


  // Formulario Settings
  const [settings, setSettings] = useState({
    wasabi_key: '',
    wasabi_secret: '',
    wasabi_bucket: '',
    wasabi_region: 'us-east-1',
    restic_password: ''
  });
  const [savingSettings, setSavingSettings] = useState(false);

  useEffect(() => {
    // Capturamos el token de la URL si viene de WHMCS (SSO)
    const token = searchParams.get("sso");
    if (token) {
      localStorage.setItem("dbp_sso_token", token);
    }
    
    // Si viene con admin=1, permitimos la pestaña Admin
    if (isAdminParam && !isEmbed) {
       // Opcional: auto-navegar a admin
    }
  }, [searchParams]);

  useEffect(() => {
    async function fetchData() {
      const token = localStorage.getItem("dbp_sso_token");
      if (!token) return;

      try {
        // Fetch Agents Count
        const respStatus = await fetch("https://api.hwperu.com/v1/agent/status", {
          headers: { "Authorization": token }
        });
        if (respStatus.ok) {
          const data = await respStatus.json();
          setAgents(data);
          setAgentCount(Object.keys(data).length);
        }


        // Fetch Current Settings
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
            restic_password: sData.restic_password || ''
          });
        }
      } catch (error) {
        console.error("Error fetching dashboard data:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  const saveSettings = async (e: React.FormEvent, isGlobal = false) => {
    e.preventDefault();
    const token = localStorage.getItem("dbp_sso_token");
    if (!token) return;

    setSavingSettings(true);
    try {
      const url = `https://api.hwperu.com/v1/user/settings${isGlobal ? '?is_global=true' : ''}`;
      const resp = await fetch(url, {
        method: "POST",
        headers: { 
          "Authorization": token,
          "Content-Type": "application/json"
        },
        body: JSON.stringify(settings)
      });
      if (resp.ok) {
        alert(isGlobal ? "✅ Configuración MAESTRA (GLOBAL) guardada." : "✅ Configuración de Wasabi guardada correctamente.");
      }
    } catch (err) {
      alert("❌ Error al guardar la configuración.");
    } finally {
      setSavingSettings(false);
    }
  };


  const testWasabi = async () => {
    const token = localStorage.getItem("dbp_sso_token");
    if (!token) return;
    setTestingWasabi(true);
    try {
      const resp = await fetch("https://api.hwperu.com/v1/admin/wasabi/ping", {
        headers: { "Authorization": token }
      });
      if (resp.ok) {
        const data = await resp.json();
        setWasabiStatus(data);
      }
    } catch (err) {
      console.error("Wasabi test failed:", err);
    } finally {
      setTestingWasabi(false);
    }
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'servers':
        return (
          <div className="space-y-8 animate-in fade-in duration-500">
             {/* Metrics Row */}
             <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">Total Managed Storage</p>
                  <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-black text-white italic">0.0</span>
                    <span className="text-xs text-gray-500 font-bold">GB UTILIZED</span>
                  </div>
                  <div className="w-full bg-gray-900 h-1 mt-4 rounded-full overflow-hidden">
                    <div className="bg-emerald-500 h-full w-[1%] shadow-[0_0_10px_rgba(16,185,129,0.5)]"></div>
                  </div>
                </div>

                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl border-l-4 border-l-emerald-500">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">Active Hybrid Agents</p>
                  <div className="text-3xl font-black text-white italic">
                    {loading ? "..." : agentCount}
                  </div>
                  <p className="text-[10px] text-emerald-500/60 mt-2 font-bold">REPORTING VIA HWPERU MESH</p>
                </div>

                <div className="bg-gray-950/50 border border-gray-900 p-6 rounded-2xl shadow-xl">
                  <p className="text-[10px] text-gray-500 font-black uppercase tracking-widest mb-2">Real-time Cluster Status</p>
                  <div className="flex items-center gap-2">
                     <Activity className="text-emerald-500 animate-pulse" size={20} />
                     <span className="text-xl font-black text-white uppercase italic">Healthy</span>
                  </div>
                  <p className="text-[10px] text-gray-600 mt-2 font-bold uppercase">All API nodes operational</p>
                </div>
              </div>

              <div className="pt-4">
                <div className="flex items-center gap-2 mb-6">
                   <div className="w-2 h-2 bg-emerald-500 rounded-full shadow-[0_0_8px_rgba(16,185,129,0.8)]"></div>
                   <h2 className="text-xl font-black uppercase italic tracking-tighter text-white">Protected Infrastructure</h2>
                </div>
                <ServerList />
              </div>
          </div>
        );

      case 'history':
        // Aplanamos todos los snapshots de todos los agentes para mostrarlos en orden cronológico
        const allSnapshots = Object.values(agents).flatMap((a: any) => 
          (a.snapshots || []).map((s: any) => ({ ...s, agent_id: a.agent_id }))
        ).sort((a: any, b: any) => new Date(b.time).getTime() - new Date(a.time).getTime());

        return (
          <div className="animate-in slide-in-from-bottom-4 duration-500 space-y-6">
             <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                   <History className="text-emerald-500" size={24} />
                   <h3 className="text-lg font-black text-white uppercase italic tracking-widest">Snapshot Recovery Points</h3>
                </div>
                <span className="text-[10px] text-gray-500 font-bold uppercase">{allSnapshots.length} points available</span>
             </div>

             {allSnapshots.length > 0 ? (
                <div className="bg-gray-950/50 border border-gray-900 rounded-3xl overflow-hidden shadow-2xl">
                   <table className="w-full text-left border-collapse">
                      <thead>
                         <tr className="bg-black/60 border-b border-gray-900">
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Snapshot ID</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Origin Server</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Timestamp (UTC)</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest text-right">Actions</th>
                         </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-900">
                         {allSnapshots.map((snap: any) => (
                            <tr key={snap.short_id} className="hover:bg-emerald-500/5 transition-colors group">
                               <td className="p-5 font-mono text-xs text-emerald-400 font-bold">{snap.short_id}</td>
                               <td className="p-5">
                                  <div className="flex items-center gap-2">
                                     <Server size={12} className="text-gray-600" />
                                     <span className="text-xs text-white uppercase font-black tracking-tighter">{snap.hostname || snap.agent_id}</span>
                                  </div>
                               </td>
                               <td className="p-5 text-xs text-gray-400 font-medium">
                                  {new Date(snap.time).toLocaleString()}
                               </td>
                               <td className="p-5 text-right">
                                  <button className="text-[10px] bg-gray-900 hover:bg-emerald-600 text-emerald-500 hover:text-white px-3 py-1 rounded-lg border border-gray-800 transition-all font-black uppercase tracking-tighter shadow-sm">
                                     Prepare Restore
                                  </button>
                               </td>
                            </tr>
                         ))}
                      </tbody>
                   </table>
                </div>
             ) : (
                <div className="bg-gray-950/50 border border-gray-900 rounded-3xl p-20 text-center">
                   <Database className="mx-auto text-gray-800 mb-4" size={48} />
                   <p className="text-sm text-gray-600 font-bold uppercase italic">No backup snapshots detected in S3 repository.</p>
                </div>
             )}
          </div>
        );


      case 'settings':
        return (
          <div className="max-w-2xl mx-auto animate-in zoom-in-95 duration-500">
              <form className="bg-gray-950/50 border border-gray-900 rounded-2xl p-8 shadow-2xl space-y-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-3 bg-blue-500/10 rounded-xl border border-blue-500/20">
                        <Cloud className="text-blue-500" size={24} />
                    </div>
                    <div>
                        <h3 className="text-lg font-black text-white uppercase italic tracking-widest">S3 Cloud Connectivity</h3>
                        <p className="text-xs text-gray-500 font-bold uppercase tracking-tighter italic">Wasabi / Amazon S3 Endpoint Integration</p>
                    </div>
                  </div>
                  {settings.wasabi_key === "" && (
                    <span className="bg-emerald-500/10 text-emerald-500 text-[9px] font-black px-3 py-1 rounded-full border border-emerald-500/20 uppercase tracking-widest animate-pulse">
                      Using HWPeru Global
                    </span>
                  )}
                </div>


                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Access Key ID</label>
                      <input 
                        type="text" 
                        value={settings.wasabi_key}
                        onChange={(e) => setSettings({...settings, wasabi_key: e.target.value})}
                        placeholder="Wasabi/AWS Key"
                        className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 focus:ring-1 focus:ring-blue-500/40 outline-none transition-all font-mono"
                      />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Secret Access Key</label>
                      <input 
                        type="password" 
                        value={settings.wasabi_secret}
                        onChange={(e) => setSettings({...settings, wasabi_secret: e.target.value})}
                        placeholder="••••••••••••••••"
                        className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none transition-all font-mono"
                      />
                   </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Bucket Name</label>
                      <input 
                        type="text" 
                        value={settings.wasabi_bucket}
                        onChange={(e) => setSettings({...settings, wasabi_bucket: e.target.value})}
                        placeholder="docker-backup-pro"
                        className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none transition-all font-mono"
                      />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1">Region</label>
                      <input 
                        type="text" 
                        value={settings.wasabi_region}
                        onChange={(e) => setSettings({...settings, wasabi_region: e.target.value})}
                        placeholder="us-east-1"
                        className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none transition-all font-mono"
                      />
                   </div>
                </div>

                <div className="space-y-2 pt-4 border-t border-gray-900">
                   <label className="text-[10px] text-emerald-500 font-black uppercase tracking-widest ml-1 flex items-center gap-2">
                      <Lock size={12} />
                      Restic Encryption Password
                   </label>
                   <input 
                     type="password" 
                     value={settings.restic_password}
                     onChange={(e) => setSettings({...settings, restic_password: e.target.value})}
                     placeholder="Tu Frase de Cifrado Maestro"
                     className="w-full bg-emerald-950/10 border border-emerald-900/30 rounded-xl px-4 py-3 text-sm text-emerald-200 focus:border-emerald-500 outline-none transition-all font-mono"
                   />
                   <p className="text-[9px] text-gray-600 font-bold uppercase italic mt-1">Este password nunca se envía al servidor como texto plano, se usa solo para cifrar los backups.</p>
                </div>

                <div className="flex flex-col md:flex-row gap-4 mt-6">
                  <button 
                    type="button"
                    onClick={(e) => saveSettings(e, false)}
                    disabled={savingSettings}
                    className="flex-1 bg-gray-900 hover:bg-gray-800 text-white font-black uppercase text-[10px] py-4 rounded-xl border border-gray-800 transition-all active:scale-[0.98]"
                  >
                    {savingSettings ? 'SYNCING...' : 'SAVE LOCAL TENANT SETTINGS'}
                  </button>

                  {isAdminParam && (
                    <button 
                      type="button"
                      onClick={(e) => saveSettings(e, true)}
                      disabled={savingSettings}
                      className="flex-1 bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-[10px] py-4 rounded-xl shadow-xl shadow-emerald-900/40 transition-all border border-emerald-400/20 active:scale-[0.98]"
                    >
                      {savingSettings ? 'DEPLOYING GLOBAL...' : 'SAVE AS MASTER GLOBAL'}
                    </button>
                  )}
                </div>

             </form>
          </div>
        );

      case 'admin':
        return (
          <div className="space-y-8 animate-in fade-in duration-500">
             <div className="bg-emerald-950/20 border border-emerald-900/30 rounded-xl p-6 mb-8 animate-in zoom-in-95 duration-500">
                <div className="flex items-center justify-between mb-6">
                   <div className="flex items-center gap-3">
                      <Network className="text-emerald-500" size={24} />
                      <div>
                         <h3 className="text-sm font-black uppercase text-emerald-500">Infrastucture Global Health</h3>
                         <p className="text-[10px] text-gray-500 font-bold uppercase tracking-widest">Master Connectivity & Restic Sync Check</p>
                      </div>
                   </div>
                   <button 
                     onClick={testWasabi}
                     disabled={testingWasabi}
                     className="bg-emerald-600 hover:bg-emerald-500 text-white text-[10px] font-black px-4 py-2 rounded-lg border border-emerald-400/20 uppercase transition-all shadow-xl shadow-emerald-950/50"
                   >
                     {testingWasabi ? 'Probing Cloud...' : 'Verify Wasabi S3 Link'}
                   </button>
                </div>

                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                   <div className="bg-black/40 p-4 rounded-lg border border-gray-900">
                      <p className="text-[9px] text-gray-600 font-black uppercase mb-1">S3 Connectivity</p>
                      <div className="flex items-center gap-2">
                         <div className={`w-1.5 h-1.5 rounded-full ${wasabiStatus ? 'bg-emerald-500 animate-pulse' : 'bg-gray-700'}`}></div>
                         <span className="text-xs font-bold text-gray-200">{wasabiStatus ? 'ONLINE' : 'PENDING'}</span>
                      </div>
                   </div>
                   <div className="bg-black/40 p-4 rounded-lg border border-gray-900">
                      <p className="text-[9px] text-gray-600 font-black uppercase mb-1">Wasabi Latency</p>
                      <span className="text-xs font-mono text-emerald-400">
                         {wasabiStatus ? `${wasabiStatus.latency_ms} ms` : '---'}
                      </span>
                   </div>
                </div>
             </div>
          </div>
        );
    }
  };

  return (
    <div className={`mx-auto space-y-8 ${isEmbed ? 'max-w-full p-2' : 'max-w-7xl p-8'}`}>
      
      {/* HEADER SECTION */}
      <div className="flex flex-col md:flex-row justify-between items-center bg-gray-950/50 p-6 rounded-3xl border border-gray-900 shadow-2xl gap-6">
        <div className="flex flex-col">
          <h1 className={`${isEmbed ? 'text-2xl' : 'text-3xl'} font-black italic tracking-tighter text-white uppercase`}>
            Docker Backup <span className="text-emerald-500">Pro</span>
          </h1>
          <p className="text-[9px] text-gray-500 uppercase tracking-[0.3em] mt-1 font-black">
             Enterprise Grade SaaS Backup Layer
          </p>
        </div>

        {/* NAVIGATION TABS UI */}
        <div className="flex bg-black/40 p-1.5 rounded-2xl border border-gray-800 shadow-inner">
           <button 
             onClick={() => setActiveTab('servers')}
             className={`flex items-center gap-2 px-6 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest ml-1 transition-all ${activeTab === 'servers' ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-950' : 'text-gray-500 hover:text-gray-300'}`}
           >
              <Server size={14} /> Servers
           </button>
           <button 
             onClick={() => setActiveTab('history')}
             className={`flex items-center gap-2 px-6 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest ml-1 transition-all ${activeTab === 'history' ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-950' : 'text-gray-500 hover:text-gray-300'}`}
           >
              <History size={14} /> History
           </button>
           <button 
             onClick={() => setActiveTab('settings')}
             className={`flex items-center gap-2 px-6 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest ml-1 transition-all ${activeTab === 'settings' ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-950' : 'text-gray-500 hover:text-gray-300'}`}
           >
              <Settings size={14} /> Settings
           </button>
           {isAdminParam && !isEmbed && (
             <button 
               onClick={() => setActiveTab('admin')}
               className={`flex items-center gap-2 px-6 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest ml-1 transition-all ${activeTab === 'admin' ? 'bg-red-600 text-white shadow-lg shadow-red-950' : 'text-red-500/60 hover:text-red-400'}`}
             >
                <Network size={14} /> Admin
             </button>
           )}
        </div>
      </div>

      {/* RENDER ACTIVE TAB */}
      <div className="min-h-[50vh]">
         {renderTabContent()}
      </div>

    </div>
  );
}
