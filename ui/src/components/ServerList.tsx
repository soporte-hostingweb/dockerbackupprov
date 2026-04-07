'use client';

import { useEffect, useState } from "react";
import { Server, Activity, ShieldCheck, Settings, HardDrive, Database, ChevronDown, ChevronUp, RotateCcw } from "lucide-react";


import FileExplorer from "./FileExplorer";

interface AgentStatus {
  agent_id: string;
  status: string;
  last_sync: string;
  last_seen_unix?: number;
  containers: string[];
  explorer: Record<string, string[]>;
  token?: string;
  os: string;
  type: string;
  maintenance?: boolean;
  pending_force?: string;
  is_syncing?: boolean;
  active_pid?: number;
  snapshots?: any[];
}

interface ServerListProps {
  onRestore?: (agentId: string, snapshots: any[]) => void;
}

export default function ServerList({ onRestore }: ServerListProps) {
  const [agents, setAgents] = useState<Record<string, AgentStatus>>({});
  const [loading, setLoading] = useState(true);
  const [expandedAgent, setExpandedAgent] = useState<string | null>(null);
  const [schedules, setSchedules] = useState<Record<string, string>>({}); // V2.3: { agentId: "daily_2am" }

  useEffect(() => {
    async function fetchAgents() {
      const token = localStorage.getItem("dbp_sso_token");
      if (!token) return;

      try {
        const response = await fetch("https://api.hwperu.com/v1/agent/status", {
          headers: { "Authorization": token }
        });
        if (response.ok) {
          const data = await response.json();
          setAgents(data);
          
          // V5.0: Sincronizar el estado de schedules local con lo que viene del servidor
          const loadedSchedules: Record<string, string> = {};
          Object.entries(data).forEach(([id, agent]: [string, any]) => {
             loadedSchedules[id] = agent.schedule || "manual";
          });
          setSchedules(loadedSchedules);
        }
      } catch (error) {
        console.error("Error fetching agents:", error);
      } finally {
        setLoading(false);
      }
    }

    fetchAgents();
    const interval = setInterval(fetchAgents, 15000); // 15s refresh
    return () => clearInterval(interval);
  }, []);

  const handleSaveConfig = async (agentId: string) => {
    const token = localStorage.getItem("dbp_sso_token");
    if (!token) return;

    try {
       const schedule = schedules[agentId] || "manual";
       
       // V5.1.1: Calcular retención según el plan seleccionado
       let retention = 1;
       if (schedule === "daily_2am_basic" || schedule === "weekly_2am") retention = 2;
       if (schedule === "custom") retention = 7;

       // Obtenemos los paths actuales para no borrarlos al guardar solo el schedule
       const configResp = await fetch(`https://api.hwperu.com/v1/agent/config?agent_id=${agentId}`, {
         headers: { "Authorization": token }
       });
       let currentPaths = [];
       if (configResp.ok) {
          const configData = await configResp.json();
          currentPaths = configData.paths || [];
       }

       const response = await fetch(`https://api.hwperu.com/v1/agent/config/save`, {
         method: "POST",
         headers: { 
           "Authorization": token,
           "Content-Type": "application/json"
         },
         body: JSON.stringify({ 
            agent_id: agentId,
            schedule: schedule,
            retention: retention,
            paths: currentPaths
         })
       });

       if (response.ok) {
         alert(`✅ [PLAN ACTUALIZADO] Modo: ${schedule.toUpperCase()} | Retención: ${retention} copias`);
       }
    } catch (err) {
       console.error("Error saving config:", err);
       alert("Error al conectar con el Control Plane");
    }
  };

  const removeAgent = async (id: string) => {
    if (!confirm(`¿Eliminar servidor "${id}" del panel?`)) return;
    
    const token = localStorage.getItem("dbp_sso_token");
    try {
      const response = await fetch(`https://api.hwperu.com/v1/agent/status/${id}`, {
        method: "DELETE",
        headers: { "Authorization": token || "" }
      });
      if (response.ok) {
        setAgents(prev => {
          const next = { ...prev };
          delete next[id];
          return next;
        });
      }
    } catch (err) {
      console.error("Error removing agent:", err);
    }
  };

  const handleAction = async (agentId: string, action: string) => {
    const token = localStorage.getItem("dbp_sso_token");
    if (!token) return;

    // Confirmaciones especiales
    if (action === 'reset' && !confirm("¿Estás seguro de REINICIAR TODA LA CONFIGURACIÓN de este servidor? Se borrarán las rutas seleccionadas.")) return;
    if (action === 'kill_sync' && !confirm("¿Deseas TERMINAR el proceso de backup actual para reducir la carga?")) return;

    try {
      const response = await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
        method: "POST",
        headers: { 
          "Authorization": token,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ action })
      });

      if (response.ok) {
        // Refrescar localmente el estado si es necesario
        setAgents(prev => ({
          ...prev,
          [agentId]: {
            ...prev[agentId],
            status: action === 'reset' ? 'Resetting...' : prev[agentId].status,
            maintenance: action === 'maintenance_on' ? true : (action === 'maintenance_off' ? false : prev[agentId].maintenance)
          }
        }));
      }
    } catch (err) {
      console.error("Error sending action:", err);
    }
  };




  if (loading) return (
    <div className="flex items-center justify-center p-12 bg-gray-900/10 border border-gray-800/50 rounded-xl animate-pulse">
      <div className="flex flex-col items-center gap-4">
        <Activity className="h-8 w-8 text-emerald-500 animate-spin" />
        <span className="text-gray-500 text-[10px] font-black uppercase tracking-widest">Polling HWPeru Network...</span>
      </div>
    </div>
  );

  const agentEntries = Object.entries(agents);

  if (agentEntries.length === 0) {
    return (
      <div className="bg-gray-950/40 border border-dashed border-gray-800 rounded-xl p-16 text-center animate-in zoom-in duration-500">
        <div className="w-16 h-16 bg-gray-900 rounded-full flex items-center justify-center mx-auto mb-6 shadow-2xl">
           <Server className="h-8 w-8 text-gray-700" />
        </div>
        <h3 className="text-lg font-bold text-gray-200 uppercase tracking-tighter">No Protected VPS Found</h3>
        <p className="text-gray-500 max-w-sm mx-auto mt-2 text-xs leading-relaxed uppercase font-medium">
          The central API is active, but hasn't received any heartbeat with your session token yet.
        </p>
        <div className="mt-8 inline-flex items-center gap-2 px-4 py-1.5 bg-emerald-500/5 border border-emerald-500/10 rounded-full">
           <div className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse"></div>
           <span className="text-[10px] text-emerald-500 font-bold uppercase tracking-widest">Ready for first report</span>
        </div>
      </div>
    );
  }


  return (
    <div className="grid grid-cols-1 gap-6 pb-20">
      {agentEntries.map(([id, data]) => (
        <div key={id} className="bg-gray-950 border border-gray-800 rounded-xl overflow-hidden hover:border-emerald-900/40 transition-all duration-300 shadow-2xl">
          <div 
            className="p-6 flex items-center justify-between cursor-pointer group"
            onClick={() => setExpandedAgent(expandedAgent === id ? null : id)}
          >
            <div className="flex items-center gap-4">
              <div className="p-3 bg-emerald-500/10 rounded-lg group-hover:bg-emerald-500/20 transition-all">
                <Activity className="h-6 w-6 text-emerald-500" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-white flex items-center gap-2 uppercase tracking-wide">
                  {data.agent_id}
                  <span className="text-[10px] bg-gray-900 text-gray-500 px-2 py-0.5 rounded-full border border-gray-800">
                    {data.os || 'Linux'}
                  </span>
                  {localStorage.getItem("dbp_sso_token")?.startsWith("dbp_admin_") && data.token && (
                     <span className="text-[9px] bg-blue-950 text-blue-400 px-2 py-0.5 rounded border border-blue-800 font-bold uppercase tracking-widest">
                       Account: {data.token.substring(0, 15)}...
                     </span>
                  )}
                </h3>

                <p className="text-xs text-gray-500 mt-1 font-mono">Real-time Telemetry: {data.last_sync || 'Connected'}</p>
              </div>
            </div>

            <div className="flex items-center gap-8">
              {/* OFFLINE DETECTION LOGIC */}
              {(() => {
                const now = Math.floor(Date.now() / 1000);
                const isOffline = data.last_seen_unix ? (now - data.last_seen_unix > 65) : false;
                
                return (
                  <>
                    {isOffline && (
                      <button 
                        onClick={(e) => { e.stopPropagation(); removeAgent(id); }}
                        className="text-[9px] bg-red-950/40 text-red-500 hover:bg-red-500 hover:text-white px-2 py-1 rounded border border-red-900/30 font-black uppercase transition-all"
                      >
                        Remove Offline Agent
                      </button>
                    )}
                    
                    <div className="flex flex-col items-center">
                       <button
                         onClick={(e) => {
                           e.stopPropagation();
                           if (onRestore) onRestore(id, data.snapshots || []);
                         }}
                         className="flex flex-col items-center px-4 py-2 bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/20 rounded-xl group/btn transition-all animate-in zoom-in duration-300"
                       >
                         <RotateCcw className="h-5 w-5 text-blue-500 mb-1 group-hover/btn:rotate-[-45deg] transition-transform" />
                         <span className="text-[9px] font-black text-blue-500 uppercase tracking-widest">Restore Wizard</span>
                       </button>
                    </div>

                    <div className={`px-4 py-1.5 rounded-full text-[10px] font-black tracking-widest border uppercase ${
                      isOffline 
                      ? 'bg-red-950/20 text-red-500 border-red-900/50'
                      : (data.status === 'Healthy' || data.status === 'SUCCESS')
                        ? 'bg-emerald-400/10 text-emerald-400 border-emerald-400/20' 
                        : 'bg-red-400/10 text-red-400 border-red-400/20'
                    }`}>
                      {isOffline ? 'OFFLINE' : (data.status || 'Active')}
                    </div>

                    {data.pending_force === 'full' && (
                       <div className="absolute -top-2 -right-2 bg-blue-600 text-[8px] font-black text-white px-2 py-0.5 rounded-full border border-blue-400 shadow-lg animate-bounce uppercase">
                          Force Full
                       </div>
                    )}

                  </>
                );
              })()}
              
              {expandedAgent === id ? <ChevronUp className="h-5 w-5 text-emerald-500" /> : <ChevronDown className="h-5 w-5 text-gray-700" />}
            </div>
          </div>

          {/* EXPLORER VIEW SECTION */}
          {expandedAgent === id && (
            <div className="bg-black/60 border-t border-gray-900 p-6 animate-in slide-in-from-top-4 duration-300">
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-3">
                  <ShieldCheck className="text-emerald-500" size={20} />

                  <div>
                    <h4 className="text-sm font-bold uppercase tracking-wider text-gray-200">Container Data Explorer</h4>
                    <p className="text-[10px] text-gray-500 uppercase">Select folders below to include in current Wasabi S3 Plan</p>
                  </div>
                </div>
                  <div className="flex gap-2">
                    <select 
                      id={`schedule-${id}`}
                      className="bg-gray-800 text-[10px] text-emerald-400 px-3 py-1.5 rounded font-bold border border-gray-700 uppercase focus:outline-none focus:border-emerald-500"
                      value={schedules[id] || "manual"}
                      onChange={(e) => {
                         setSchedules(prev => ({ ...prev, [id]: e.target.value }));
                      }}
                    >

                       <option value="manual">Manual Only (Free - 1 Copy)</option>
                       <option value="daily_2am_basic">Daily Basic (Pro - 2 Copies)</option>
                       <option value="weekly_2am">Weekly (Standard - 2 Copies)</option>
                       <option value="custom">Custom (Enterprise - 7 Copies)</option>
                    </select>
                    
                    <button 
                      onClick={() => handleSaveConfig(id)}
                      className="bg-emerald-600 hover:bg-emerald-500 text-[10px] text-white px-4 py-1.5 rounded font-black transition-all border border-emerald-400/20 shadow-lg shadow-emerald-950/20 uppercase"
                    >
                      Save Configuration
                    </button>

                    <button 
                      onClick={() => handleAction(id, 'reset')}
                      className="bg-gray-800 hover:bg-gray-700 text-[10px] text-red-400 px-3 py-1.5 rounded font-bold transition-all border border-gray-700 uppercase"
                    >
                      Reset
                    </button>
                  </div>

              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                {data.containers && data.containers.map((container, idx) => (
                  <div key={idx} className="space-y-3">
                    <div className="flex items-center justify-between text-xs px-1">
                      <div className="flex items-center gap-2 text-gray-200">
                         {container.toLowerCase().includes('sql') || container.toLowerCase().includes('db') || container.toLowerCase().includes('maria')
                          ? <Database size={16} className="text-emerald-500 animate-pulse" />
                          : <HardDrive size={16} className="text-gray-500" />
                         }
                         <span className="font-black text-gray-300 uppercase tracking-tighter">{container}</span>
                      </div>
                      { (container.toLowerCase().includes('sql') || container.toLowerCase().includes('db')) && (
                        <span className="text-[9px] bg-emerald-950 text-emerald-400 px-2 py-0.5 rounded border border-emerald-800 font-bold">SQL DUMP ENABLED</span>
                      )}
                    </div>
                    
                    {/* El explorer data está mappeado por los primeros 10 caracteres del ID o nombre en el agente */}
                    <FileExplorer 
                      agentId={id}
                      containerName={container} 
                      folders={data.explorer ? data.explorer[container] || [] : []} 
                      schedule={schedules[id] || "manual"}
                    />



                  </div>
                ))}
              </div>
              
              <div className="mt-10 pt-8 border-t border-gray-800 flex flex-col md:flex-row justify-between items-center gap-4">
                <div className="flex items-center gap-4">
                   <div className="text-center">
                      <p className="text-[10px] text-gray-600 uppercase font-black">Plan Limit</p>
                      <p className="text-xs text-white font-bold">100 GB</p>
                   </div>
                   <div className="h-8 w-px bg-gray-800"></div>
                   <div className="text-center">
                      <p className="text-[10px] text-gray-600 uppercase font-black">Current Usage</p>
                      <p className="text-xs text-emerald-500 font-bold">12.4 GB</p>
                   </div>
                </div>
                
                <div className="flex flex-col gap-2 w-full md:w-auto">
                  {data.snapshots && data.snapshots.length > 0 ? (
                    data.snapshots.map((snap: any, idx: number) => (
                      <div key={idx} className="flex items-center justify-between p-3 rounded-lg bg-black/40 border border-gray-900 last:border-0 hover:bg-emerald-500/5 transition-colors group">
                        <div className="flex flex-col">
                          <span className="text-[10px] font-mono text-emerald-500 font-bold">{snap.short_id || snap.id}</span>
                          <span className="text-[10px] text-gray-500">{new Date(snap.time).toLocaleString()}</span>
                        </div>
                        <button 
                          onClick={() => onRestore && onRestore(id, data.snapshots || [])}
                          className="text-[9px] bg-gray-900 hover:bg-blue-600 text-blue-500 hover:text-white px-3 py-1 rounded border border-gray-800 transition-all font-black uppercase tracking-tighter"
                        >
                          Restore
                        </button>
                      </div>
                    ))
                  ) : (
                    <div className="p-8 text-center border-2 border-dashed border-gray-900 rounded-xl">
                      <p className="text-[9px] text-gray-600 font-bold uppercase italic">No snapshots available for recovery</p>
                    </div>
                  )}
                </div>

                <div className="flex gap-4">
                  <button 
                    onClick={() => {
                      if (data.is_syncing) {
                        handleAction(id, 'kill_sync').then(() => handleAction(id, 'maintenance_on'));
                      } else {
                        handleAction(id, data.maintenance ? 'maintenance_off' : 'maintenance_on');
                      }
                    }}
                    className={`${data.maintenance ? 'bg-orange-600 text-white border-orange-500' : 'bg-gray-900 text-gray-400 border-gray-800'} hover:bg-gray-800 text-[10px] px-6 py-2.5 rounded-lg font-bold border transition-all uppercase tracking-widest`}
                  >
                     {data.is_syncing ? 'Terminate & Pause' : (data.maintenance ? 'Resume Agent' : 'Maintenance Mode')}
                  </button>
                  
                  <div className="flex gap-0.5">
                    <button 
                      onClick={() => handleAction(id, 'force_selected')}
                      className="bg-emerald-600 hover:bg-emerald-500 text-white text-[10px] px-6 py-2.5 rounded-l-lg font-bold shadow-xl shadow-emerald-900/30 transition-all uppercase tracking-widest border border-emerald-400/20"
                    >
                      Force Selected
                    </button>
                    <button 
                      onClick={() => {
                        if (confirm("¿INICIAR BACKUP COMPLETO DEL SERVIDOR? Esto ignorará los filtros y respaldará la raíz del host.")) {
                          handleAction(id, 'force_full');
                        }
                      }}
                      className="bg-emerald-800 hover:bg-emerald-700 text-white text-[10px] px-4 py-2.5 rounded-r-lg font-bold shadow-xl transition-all uppercase tracking-widest border border-emerald-400/10"
                      title="Full Root Snapshot"
                    >
                      Full
                    </button>
                  </div>
                </div>

              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
