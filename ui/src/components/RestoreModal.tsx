'use client';
import { useState, useEffect } from "react";
import { Server, X, RotateCcw, FolderInput, ShieldCheck, Database, HardDrive, AlertCircle, ChevronRight, Calendar, Clock, Search, FileText, Folder, Activity, CheckSquare, Square, Zap, Globe } from "lucide-react";

interface RestoreModalProps {
  isOpen: boolean;
  onClose: () => void;
  agentId: string;
  snapshots: any[];
  token: string;
}

export default function RestoreModal({ isOpen, onClose, agentId, snapshots, token }: RestoreModalProps) {
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [selectedSnapshot, setSelectedSnapshot] = useState<any>(null);
  const [selectedPaths, setSelectedPaths] = useState<string[]>([]);
  const [restorePath, setRestorePath] = useState("/restore_data");
  const [isOverwriteMode, setIsOverwriteMode] = useState(false);
  const [explorerContent, setExplorerContent] = useState<any[]>([]);
  const [isLodingContent, setIsLoadingContent] = useState(false);
  const [agentData, setAgentData] = useState<any>(null);
  const [currentPath, setCurrentPath] = useState(""); 
  
  // V8.0: Bare-Metal Clone State
  const [isCloneMode, setIsCloneMode] = useState(false);
  const [targetIP, setTargetIP] = useState("");
  const [targetPort, setTargetPort] = useState("22");
  const [targetPass, setTargetPass] = useState("");
  const [authCode, setAuthCode] = useState("");
  const [phone, setPhone] = useState("");
  const [requestingCode, setRequestingCode] = useState(false);

  useEffect(() => {
    if (isOpen && agentId) {
        fetch(`https://api.hwperu.com/v1/agent/status`, {
            headers: { "Authorization": token }
        })
        .then(res => res.json())
        .then(data => setAgentData(data[agentId]))
        .catch(err => console.error("Error loading agent data:", err));
    }
  }, [isOpen, agentId, token]);

  if (!isOpen) return null;

  const sortedSnapshots = [...snapshots].sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());

  // FIX: Toggle sin duplicados
  const togglePath = (path: string) => {
    setSelectedPaths(prev => {
        if (prev.includes(path)) {
            return prev.filter(p => p !== path);
        }
        return [...prev, path];
    });
  };

  // FIX: Select All sin duplicados
  const selectAll = () => {
    const allPathsInView = explorerContent.map(item => typeof item === 'string' ? item : item.path);
    setSelectedPaths(prev => Array.from(new Set([...prev, ...allPathsInView])));
  };

  const clearSelection = () => setSelectedPaths([]);

  const fetchSnapshotContent = async (snapId: string, path: string) => {
    setIsLoadingContent(true);
    setCurrentPath(path);
    try {
        await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ action: "ls_snapshot", snapshot_id: snapId, path: path })
        });

        let attempts = 0;
        const poll = setInterval(async () => {
             attempts++;
             const statusResp = await fetch(`https://api.hwperu.com/v1/agent/status?agent_id=${agentId}`, {
                 headers: { "Authorization": token }
             });
             const statusData = await statusResp.json();
             const agent = statusData[agentId];
             
             if (agent && agent.cmd_task === "none") {
                 clearInterval(poll);
                 if (agent.cmd_result) {
                     try {
                        const parsed = JSON.parse(agent.cmd_result);
                        setExplorerContent(Array.isArray(parsed) ? parsed : [parsed]);
                     } catch (e) { console.error(e); }
                 }
                 setIsLoadingContent(false);
                 if (step === 1) setStep(2);
             }
             if (attempts > 30) { clearInterval(poll); setIsLoadingContent(false); }
        }, 1000);
    } catch (err) { setIsLoadingContent(false); }
  };

  const handleRestore = async () => {
    setLoading(true);
    try {
      if (isCloneMode) {
          const resp = await fetch(`https://api.hwperu.com/v1/agent/clone`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ 
                source_agent_id: agentId,
                snapshot_id: selectedSnapshot.id || selectedSnapshot.short_id,
                ip: targetIP,
                port: targetPort,
                pass: targetPass,
                auth_code: authCode
            })
          });
          if (resp.ok) setStep(4);
          else alert("Error orchestrating Target SSH Connection.");
      } else {
          const resp = await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ 
                action: "restore", 
                snapshot_id: selectedSnapshot.id || selectedSnapshot.short_id,
                destination: isOverwriteMode ? "/" : restorePath,
                paths: selectedPaths 
            })
          });
          if (resp.ok) setStep(4);
      }
    } catch (err) { alert("Error de red"); } finally { setLoading(false); }
  };

  const formatDisplayName = (path: string, name: string) => {
    let clean = name || path.split('/').pop() || "";
    clean = clean.replace(/^(\/)?host_root\//i, "");
    return clean.toUpperCase();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/90 backdrop-blur-md animate-in fade-in duration-500">
      <div className="bg-gray-950 border border-gray-900 w-full max-w-2xl rounded-[2.5rem] overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
        
        {/* Header */}
        <div className="p-6 border-b border-gray-900 flex justify-between items-center">
          <div className="flex items-center gap-4">
             <div className="p-3 bg-blue-500/10 rounded-2xl border border-blue-500/20 text-blue-400">
                <RotateCcw size={24} />
             </div>
             <div>
                <h3 className="text-lg font-black text-white italic uppercase">Restore Wizard Pro</h3>
                <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[9px] text-emerald-500 font-black uppercase tracking-widest bg-emerald-500/5 px-2 py-0.5 rounded-full border border-emerald-500/10">V5.0 SAFE RESTORE</span>
                    <span className="text-[9px] text-gray-600 font-bold uppercase">{agentId}</span>
                </div>
             </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-full text-gray-500 hover:text-white transition-all"><X size={20} /></button>
        </div>

        {/* Wizard Steps */}
        <div className="px-12 pt-6 flex items-center justify-between">
            {[1,2,3].map((s) => (
                <div key={s} className="flex items-center flex-1 last:flex-none">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black border transition-all ${step >= s ? 'bg-blue-600 border-blue-400 text-white' : 'bg-gray-900 border-gray-800 text-gray-600'}`}>{s < step ? '✓' : s}</div>
                    {s < 3 && <div className={`h-[1px] flex-1 mx-2 ${step > s ? 'bg-blue-600' : 'bg-gray-800'}`}></div>}
                </div>
            ))}
        </div>

        <div className="p-8 overflow-y-auto custom-scrollbar flex-1">
          {step === 1 && (
            <div className="space-y-6 animate-in slide-in-from-bottom-4">
                <div className="flex items-center justify-between">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2"><Calendar size={14} className="text-blue-500" /> Puntos de Restauración</h4>
                    {agentData?.detected_stack?.wordpress && (
                        <div className="flex items-center gap-2 px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-full">
                            <Zap size={10} className="text-emerald-500" />
                            <span className="text-[8px] text-emerald-500 font-black uppercase">WordPress Detectado</span>
                        </div>
                    )}
                </div>
                <div className="grid grid-cols-1 gap-3">
                    {sortedSnapshots.map((s: any) => (
                        <button key={s.id} onClick={() => { setSelectedSnapshot(s); fetchSnapshotContent(s.id, ""); }} className="w-full flex items-center justify-between p-5 bg-gray-900/40 border border-gray-800 rounded-3xl hover:border-blue-500 transition-all group">
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-gray-800 rounded-2xl text-gray-600 group-hover:text-blue-500"><Clock size={18} /></div>
                                <div className="flex flex-col">
                                    <span className="text-base font-black text-white italic transition-colors uppercase tracking-tighter">{new Date(s.time).toLocaleString()}</span>
                                    <span className="text-[9px] font-mono text-gray-600 tracking-tighter mt-0.5">{s.short_id || s.id}</span>
                                </div>
                            </div>
                            <ChevronRight size={20} className="text-gray-700 group-hover:text-blue-500" />
                        </button>
                    ))}
                </div>

                {agentData?.detected_stack?.wordpress && (
                    <div className="mt-8 p-6 bg-blue-500/5 border border-blue-500/20 rounded-[2.5rem] space-y-4 animate-in fade-in duration-700">
                        <div className="flex items-center gap-3">
                            <div className="p-3 bg-blue-500/10 rounded-2xl text-blue-400"><Globe size={20} /></div>
                            <div>
                                <h5 className="text-sm font-black text-white italic uppercase">Recuperación Express (1-Click)</h5>
                                <p className="text-[10px] text-gray-500 font-medium uppercase tracking-widest">Reconstrucción total de WordPress + DB + SSL</p>
                            </div>
                        </div>
                        <button 
                            onClick={() => {
                                const latest = sortedSnapshots[0];
                                if (latest) {
                                    setSelectedSnapshot(latest);
                                    setSelectedPaths(["[ALL_SYSTEM_ROOT]"]);
                                    setIsOverwriteMode(true);
                                    setStep(3);
                                }
                            }}
                            className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-black py-4 rounded-2xl text-xs uppercase tracking-widest shadow-xl shadow-emerald-950/20 transition-all flex items-center justify-center gap-2"
                        >
                            <Zap size={16} /> RECONSTRUIR ÚLTIMA COPIA (WP)
                        </button>
                    </div>
                )}
            </div>
          )}

          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4">
                <div className="flex justify-between items-center bg-gray-900/30 p-4 rounded-3xl border border-gray-900">
                    <div className="flex items-center gap-3 overflow-hidden">
                        <div className="p-2 bg-blue-500/10 rounded-xl text-blue-500"><Database size={16} /></div>
                        <span className="text-[10px] text-blue-400 font-black uppercase italic truncate tracking-tighter">{agentId} / {currentPath || "ROOT"}</span>
                    </div>
                    <div className="flex gap-2 shrink-0">
                        <button onClick={selectAll} className="text-[9px] px-3 py-1 bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-lg font-black uppercase">SELECT ALL</button>
                        <button onClick={clearSelection} className="text-[9px] px-3 py-1 bg-gray-800 text-gray-400 rounded-lg font-black uppercase">CLEAR</button>
                    </div>
                </div>

                {currentPath && <button onClick={() => { const p = currentPath.split('/'); p.pop(); fetchSnapshotContent(selectedSnapshot.id, p.join('/')); }} className="text-[10px] text-gray-500 font-black uppercase italic hover:text-blue-400">← VOLVER ATRÁS</button>}

                <div className="grid grid-cols-1 gap-3 max-h-[45vh] overflow-y-auto pr-2 relative">
                    {explorerContent.map((item: any, idx: number) => {
                        const path = item.path;
                        const isDir = item.type === "dir";
                        const isSelected = selectedPaths.includes(path);
                        return (
                            <div key={idx} onClick={() => isDir ? fetchSnapshotContent(selectedSnapshot.id, path) : togglePath(path)} className={`flex items-center justify-between p-4 bg-gray-900/40 border rounded-2xl cursor-pointer transition-all ${isSelected ? 'border-emerald-500/40 bg-emerald-500/[0.05]' : 'border-gray-900 hover:border-gray-800'}`}>
                                <div className="flex items-center gap-3 flex-1 min-w-0">
                                    <div className={`p-2 rounded-xl ${isDir ? 'bg-emerald-500/10 text-emerald-500' : 'bg-gray-800 text-gray-500'}`}>{isDir ? <Folder size={16} /> : <FileText size={16} />}</div>
                                    <div className="flex flex-col min-w-0">
                                        <span className={`text-[12px] font-black uppercase italic tracking-tighter truncate ${isSelected ? 'text-white' : 'text-gray-300'}`}>{formatDisplayName(path, item.name)}</span>
                                        <span className="text-[8px] text-gray-600 font-mono italic truncate opacity-60">{path}</span>
                                    </div>
                                </div>
                                {isDir ? <ChevronRight size={14} className="text-gray-700" /> : <div className={`w-4 h-4 border-2 rounded flex items-center justify-center transition-all ${isSelected ? 'bg-emerald-500 border-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'border-gray-800'}`}>{isSelected && <div className="w-1.5 h-1.5 bg-white rounded-full"></div>}</div>}
                            </div>
                        );
                    })}
                </div>

                <div className="flex items-center justify-between pt-4 border-t border-gray-900 mt-2">
                    <span className="text-[10px] text-emerald-500 font-black uppercase italic tracking-widest">{selectedPaths.length} TARGETS SELECTED</span>
                    <button onClick={() => setStep(3)} disabled={selectedPaths.length === 0} className="bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-900 disabled:text-gray-700 text-white font-black uppercase text-xs px-10 py-5 rounded-2xl transition-all shadow-xl">LOCK SELECTION</button>
                </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-6 animate-in slide-in-from-right-4">
                <button onClick={() => setStep(2)} className="text-[10px] text-gray-500 font-black uppercase hover:text-white transition-colors">← EDIT SELECTION</button>

                <div className="p-6 bg-emerald-500/[0.03] border border-emerald-500/10 rounded-[2rem] space-y-3">
                    <h4 className="text-[10px] text-emerald-500 font-black uppercase italic">Deployment Summary</h4>
                    <div className="space-y-2 max-h-[15vh] overflow-y-auto pr-2 custom-scrollbar">
                        {selectedPaths.map(p => (
                            <div key={p} className="flex flex-col p-3 bg-black/20 rounded-xl border border-white/5">
                                <span className="text-[11px] font-black text-white uppercase italic">{p.split('/').pop()?.toUpperCase()}</span>
                                <span className="text-[8px] font-mono text-gray-600 italic truncate">{p}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div className={`p-6 rounded-[2rem] border transition-all ${snapshots.length < 2 && !isCloneMode ? 'bg-red-500/5 border-red-500/20' : 'bg-orange-500/5 border-orange-500/20'}`}>
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <div className={`p-3 rounded-2xl ${snapshots.length < 2 && !isCloneMode ? 'bg-red-500/10 text-red-500' : 'bg-orange-500/10 text-orange-500'}`}><AlertCircle size={20} /></div>
                            <div>
                                <h4 className="text-[11px] font-black text-white uppercase italic">In-Place Restore (Direct Host)</h4>
                                <p className="text-[9px] text-gray-500 font-bold uppercase tracking-widest">{snapshots.length < 2 && !isCloneMode ? '❌ BLOCK: Min 2 snapshots required' : '⚠️ WARNING: Overwrites current host files'}</p>
                            </div>
                        </div>
                        <button disabled={snapshots.length < 2 || isCloneMode} onClick={() => setIsOverwriteMode(!isOverwriteMode)} className={`w-12 h-6 rounded-full transition-all flex items-center px-1 ${isOverwriteMode ? 'bg-orange-500' : 'bg-gray-800'} ${snapshots.length < 2 || isCloneMode ? 'opacity-20 cursor-not-allowed' : ''}`}>
                            <div className={`w-4 h-4 bg-white rounded-full transition-all ${isOverwriteMode ? 'translate-x-6' : 'translate-x-0'}`}></div>
                        </button>
                    </div>
                </div>

                {/* V8.0 Bare Metal Clone Toggle */}
                <div className="p-6 rounded-[2rem] border bg-emerald-500/5 border-emerald-500/20 transition-all">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <div className="p-3 rounded-2xl bg-emerald-500/10 text-emerald-500"><Server size={20} /></div>
                            <div>
                                <h4 className="text-[11px] font-black text-white uppercase italic">Clone to New VPS (Bare-Metal)</h4>
                                <p className="text-[9px] text-emerald-600 font-bold uppercase tracking-widest">Reconstruct this snapshot in an empty instance via SSH</p>
                            </div>
                        </div>
                        <button onClick={() => { setIsCloneMode(!isCloneMode); setIsOverwriteMode(false); }} className={`w-12 h-6 rounded-full transition-all flex items-center px-1 ${isCloneMode ? 'bg-emerald-500' : 'bg-gray-800'}`}>
                            <div className={`w-4 h-4 bg-white rounded-full transition-all ${isCloneMode ? 'translate-x-6' : 'translate-x-0'}`}></div>
                        </button>
                    </div>

                    {isCloneMode && (
                        <div className="mt-6 space-y-3 animate-in fade-in slide-in-from-top-2 p-4 bg-black/40 border border-emerald-900/50 rounded-2xl">
                            <div className="grid grid-cols-2 gap-3">
                                <div className="space-y-1">
                                    <label className="text-[9px] text-emerald-600 font-black uppercase ml-1 tracking-widest">Target IP</label>
                                    <input type="text" value={targetIP} onChange={(e) => setTargetIP(e.target.value)} className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-xs text-white focus:border-emerald-500 outline-none font-mono" placeholder="192.168.1.10" />
                                </div>
                                <div className="space-y-1">
                                    <label className="text-[9px] text-emerald-600 font-black uppercase ml-1 tracking-widest">SSH Port</label>
                                    <input type="text" value={targetPort} onChange={(e) => setTargetPort(e.target.value)} className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-xs text-white focus:border-emerald-500 outline-none font-mono" placeholder="22" />
                                </div>
                            </div>
                            <div className="space-y-1">
                                <label className="text-[9px] text-emerald-600 font-black uppercase ml-1 tracking-widest">Root Password</label>
                                <input type="password" value={targetPass} onChange={(e) => setTargetPass(e.target.value)} className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-xs text-white focus:border-emerald-500 outline-none font-mono" placeholder="••••••••••••" />
                            </div>

                            <div className="pt-4 border-t border-emerald-900/30 space-y-4">
                                <div className="flex items-center gap-3">
                                    <ShieldCheck className="w-4 h-4 text-emerald-500" />
                                    <h5 className="text-[10px] text-white font-black uppercase italic">Doble Factor de Seguridad (WhatsApp)</h5>
                                </div>
                                <div className="flex gap-2">
                                    <input type="text" value={phone} onChange={(e) => setPhone(e.target.value)} className="flex-1 bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-xs text-white focus:border-emerald-500 outline-none font-mono" placeholder="51987654321" />
                                    <button 
                                        type="button"
                                        onClick={async () => {
                                            setRequestingCode(true);
                                            try {
                                                const r = await fetch('https://api.hwperu.com/v1/auth/request-code', {
                                                    method: 'POST',
                                                    headers: { "Authorization": token, "Content-Type": "application/json" },
                                                    body: JSON.stringify({ action: "clone_authorize", phone: phone })
                                                });
                                                if(r.ok) alert("Código enviado a WhatsApp");
                                                else alert("Error al solicitar código");
                                            } catch(e) { alert("Error de conexión"); }
                                            finally { setRequestingCode(false); }
                                        }}
                                        disabled={!phone || requestingCode}
                                        className="bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-500 text-[10px] font-black uppercase px-4 rounded-xl border border-emerald-500/20 disabled:opacity-30"
                                    >
                                        {requestingCode ? '...' : 'Solicitar Código'}
                                    </button>
                                </div>
                                <div className="space-y-1">
                                    <label className="text-[9px] text-emerald-600 font-black uppercase ml-1 tracking-widest">Código de Verificación (6 dígitos)</label>
                                    <input type="text" value={authCode} onChange={(e) => setAuthCode(e.target.value)} className="w-full bg-emerald-500/5 border border-emerald-500/30 rounded-xl px-4 py-4 text-sm text-center text-white tracking-[0.5em] font-black focus:border-emerald-500 outline-none" placeholder="000000" maxLength={6} />
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {!isOverwriteMode && !isCloneMode && (
                    <div className="space-y-2 animate-in fade-in duration-300">
                        <label className="text-[10px] text-gray-500 font-black uppercase italic ml-1">Safe Sandbox Destination</label>
                        <input type="text" value={restorePath} onChange={(e) => setRestorePath(e.target.value)} className="w-full bg-black/40 border border-gray-800 rounded-2xl px-6 py-5 text-xs text-white focus:border-blue-500 outline-none font-mono tracking-widest" placeholder="/restore_data" />
                    </div>
                )}

                <button onClick={handleRestore} disabled={loading || (isCloneMode && (!targetIP || !targetPass || !authCode))} className={`w-full font-black uppercase text-xs py-6 rounded-3xl transition-all shadow-xl flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed ${isCloneMode ? 'bg-emerald-600 hover:bg-emerald-500' : (isOverwriteMode ? 'bg-orange-600 hover:bg-orange-500' : 'bg-blue-600 hover:bg-blue-500')}`}>
                    {loading ? <Activity className="animate-spin" size={18} /> : (isCloneMode ? <Server size={18} /> : (isOverwriteMode ? <Database size={18} /> : <ShieldCheck size={18} />))}
                    {loading ? 'SYNCING...' : (isCloneMode ? 'ORCHESTRATE BARE-METAL CLONE' : (isOverwriteMode ? 'CONFIRM HOST OVERWRITE' : 'START SAFE RESTORATION'))}
                </button>
            </div>
          )}

          {step === 4 && (
            <div className="text-center space-y-6 py-12 animate-in zoom-in-95 duration-700">
               <div className="w-24 h-24 bg-emerald-500/10 border border-emerald-500/20 rounded-[2.5rem] flex items-center justify-center mx-auto text-emerald-500 shadow-2xl animate-bounce"><ShieldCheck size={48} /></div>
               <div className="space-y-2">
                  <h3 className="text-2xl font-black text-white italic uppercase tracking-tighter">TASK COMMENCED</h3>
                  <p className="text-[10px] text-gray-500 font-bold uppercase tracking-widest max-w-sm mx-auto leading-relaxed">
                     Restoration is running in the background. Monitoring via <span className="text-blue-500">Global Activity</span> panel.
                  </p>
               </div>
               <button onClick={onClose} className="text-[10px] text-gray-600 font-black uppercase tracking-widest hover:text-white">Close Wizard</button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
