'use client';
import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { 
  AlertCircle, Terminal, ShieldCheck, Activity, Network, 
  Database, Server, History, Settings, Cloud, Lock, RotateCcw
} from "lucide-react";

import ServerList from "@/components/ServerList";
import RestoreModal from "@/components/RestoreModal";

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
  
  // Restore Modal State
  const [isRestoreOpen, setIsRestoreOpen] = useState(false);
  const [restoreAgentId, setRestoreAgentId] = useState("");
  const [restoreSnapshots, setRestoreSnapshots] = useState<any[]>([]);

  const [wasabiStatus, setWasabiStatus] = useState<{ status: string; latency_ms: number; bucket: string } | null>(null);
  const [testingWasabi, setTestingWasabi] = useState(false);

  const [settings, setSettings] = useState({
    wasabi_key: '',
    wasabi_secret: '',
    wasabi_bucket: '',
    wasabi_region: 'us-east-1',
    restic_password: ''
  });
  const [savingSettings, setSavingSettings] = useState(false);

  const token = searchParams.get("sso") || localStorage.getItem("dbp_sso_token") || "";

  useEffect(() => {
    if (searchParams.get("sso")) {
      localStorage.setItem("dbp_sso_token", searchParams.get("sso")!);
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
              restic_password: sData.restic_password || ''
            });
          }
      }

      if (activeTab === 'history') {
          const respHist = await fetch("https://api.hwperu.com/v1/history", {
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

  const saveSettings = async (e: React.FormEvent, isGlobal = false) => {
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
              <ServerList onRestore={openRestore} />
          </div>
        );

      case 'history':
        return (
          <div className="animate-in slide-in-from-bottom-4 duration-500 space-y-6">
             <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                   <History className="text-emerald-500" size={24} />
                   <h3 className="text-lg font-black text-white uppercase italic tracking-widest">Global Activity Log</h3>
                </div>
                <span className="text-[10px] text-gray-500 font-bold uppercase">{activities.length} total entries</span>
             </div>
             {activities.length > 0 ? (
                <div className="bg-gray-950/50 border border-gray-900 rounded-3xl overflow-hidden shadow-2xl">
                   <table className="w-full text-left border-collapse">
                      <thead>
                         <tr className="bg-black/60 border-b border-gray-900">
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Timestamp</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Agent ID</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Snapshot ID</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Size</th>
                            <th className="p-5 text-[10px] text-gray-500 font-black uppercase tracking-widest">Status</th>
                         </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-900">
                         {activities.map((act: any) => (
                            <tr key={act.id} className="hover:bg-emerald-500/5 transition-colors">
                               <td className="p-5 text-xs text-gray-400 font-mono">{new Date(act.timestamp).toLocaleString()}</td>
                               <td className="p-5 text-xs text-white uppercase font-black italic">{act.agent_id}</td>
                               <td className="p-5 text-xs text-emerald-400 font-mono">{act.snapshot_id || '---'}</td>
                               <td className="p-5 text-xs text-gray-300 font-bold whitespace-nowrap">
                                  {act.size_bytes > 1024 * 1024 
                                    ? `${(act.size_bytes / (1024 * 1024)).toFixed(2)} MB` 
                                    : act.size_bytes > 1024 
                                      ? `${(act.size_bytes / 1024).toFixed(1)} KB`
                                      : `${act.size_bytes || 0} B`}
                               </td>
                               <td className="p-5">
                                  <span className={`px-3 py-1 rounded-full text-[9px] font-black uppercase ${act.status === 'SUCCESS' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/25' : 'bg-red-500/10 text-red-500 border border-red-500/25'}`}>
                                     {act.status}
                                  </span>
                               </td>
                            </tr>
                         ))}
                      </tbody>
                   </table>
                </div>
             ) : (
                <div className="bg-gray-950/20 border-2 border-dashed border-gray-900 rounded-3xl p-20 text-center">
                   <History className="mx-auto text-gray-800 mb-4 animate-pulse" size={48} />
                   <p className="text-sm text-gray-600 font-bold uppercase italic tracking-widest">No history recorded yet in HWPeru Cloud.</p>
                </div>
             )}
          </div>
        );

      case 'settings':
        return (
          <div className="max-w-2xl mx-auto animate-in zoom-in-95 duration-500">
              <form className="bg-gray-950/50 border border-gray-900 rounded-3xl p-8 shadow-2xl space-y-6">
                <div className="flex items-center gap-3 mb-4">
                    <div className="p-3 bg-blue-500/10 rounded-xl border border-blue-500/20 text-blue-500"><Cloud size={24} /></div>
                    <h3 className="text-lg font-black text-white uppercase italic tracking-widest">S3 Tenant Configuration</h3>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase">Access Key</label>
                      <input type="text" value={settings.wasabi_key} onChange={(e) => setSettings({...settings, wasabi_key: e.target.value})} className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none font-mono" />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase">Secret Key</label>
                      <input type="password" value={settings.wasabi_secret} onChange={(e) => setSettings({...settings, wasabi_secret: e.target.value})} className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none font-mono" />
                   </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase">Bucket</label>
                      <input type="text" value={settings.wasabi_bucket} onChange={(e) => setSettings({...settings, wasabi_bucket: e.target.value})} className="w-full bg-black/40 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none font-mono" />
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] text-gray-500 font-black uppercase">Restic Password</label>
                      <input type="password" value={settings.restic_password} onChange={(e) => setSettings({...settings, restic_password: e.target.value})} className="w-full bg-emerald-950/10 border border-emerald-900/30 rounded-xl px-4 py-3 text-sm text-emerald-200 outline-none font-mono" />
                   </div>
                </div>
                <button type="button" onClick={(e) => saveSettings(e, false)} disabled={savingSettings} className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-xs py-4 rounded-2xl transition-all shadow-xl shadow-emerald-950/40">
                    {savingSettings ? 'SYNCING...' : 'SAVE CONFIGURATION'}
                </button>
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
          <h1 className="text-3xl font-black italic tracking-tighter text-white uppercase">Docker Backup <span className="text-emerald-500">Pro</span></h1>
          <p className="text-[10px] text-gray-500 uppercase tracking-[0.4em] mt-1 font-black leading-none">Enterprise Infrastructure Protection Layer</p>
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
    </div>
  );
}
