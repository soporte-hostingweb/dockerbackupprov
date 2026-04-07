'use client';
import { useState, useEffect } from "react";
import { X, RotateCcw, FolderInput, ShieldCheck, Database, HardDrive, AlertCircle, ChevronRight, Calendar, Clock, Search, FileText, Folder, Activity, CheckSquare, Square } from "lucide-react";

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
  const [explorerContent, setExplorerContent] = useState<any[]>([]);
  const [isLodingContent, setIsLoadingContent] = useState(false);
  const [agentData, setAgentData] = useState<any>(null);
  const [currentPath, setCurrentPath] = useState(""); // V4.6.5: Ruta actual dentro del snapshot

  // Cargar datos del agente al abrir
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

  // ORDENAR SNAPSHOTS CRONOLÓGICAMENTE (Paso 1)
  const sortedSnapshots = [...snapshots].sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());

  const togglePath = (path: string) => {
    if (selectedPaths.includes(path)) {
        setSelectedPaths(selectedPaths.filter(p => p !== path));
    } else {
        setSelectedPaths([...selectedPaths, path]);
    }
  };

  const selectAll = () => {
    const allPathsInView = explorerContent.map(item => {
        const name = typeof item === 'string' ? item : item.path;
        return name;
    });
    // Unimos los de la vista actual a los ya seleccionados sin duplicar
    const newSelection = Array.from(new Set([...selectedPaths, ...allPathsInView]));
    setSelectedPaths(newSelection);
  };

  const clearSelection = () => setSelectedPaths([]);

  // Solicitar listado de archivos del snapshot seleccionado (soporta niveles)
  const fetchSnapshotContent = async (snapId: string, path: string) => {
    setIsLoadingContent(true);
    setCurrentPath(path);
    try {
        await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ action: "ls_snapshot", snapshot_id: snapId, path: path })
        });

        // Poll for result (V4.7.2 Super Turbo: Polling cada 1s cuando hay tarea pendiente)
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
                     } catch (e) {
                        console.error("Parse error:", e);
                     }
                 }
                 setIsLoadingContent(false);
                 if (step === 1) setStep(2);
             }

             if (attempts > 30) {
                 clearInterval(poll);
                 alert("Timeout al esperar respuesta del agente.");
                 setIsLoadingContent(false);
             }
        }, 1000);

    } catch (err) {
        console.error(err);
        setIsLoadingContent(false);
    }
  };

  const handleRestore = async () => {
    setLoading(true);
    try {
      const resp = await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
        method: "POST",
        headers: { "Authorization": token, "Content-Type": "application/json" },
        body: JSON.stringify({ 
            action: "restore", 
            snapshot_id: selectedSnapshot.id || selectedSnapshot.short_id,
            destination: restorePath,
            paths: selectedPaths 
        })
      });
      if (resp.ok) setStep(4);
    } catch (err) {
      alert("Error de red");
    } finally {
      setLoading(false);
    }
  };

  // Helper para volver atrás en breadcrumbs
  const navigateBack = () => {
    const parts = currentPath.split('/').filter(p => p !== "");
    parts.pop();
    const newPath = parts.length > 0 ? "/" + parts.join('/') : "";
    fetchSnapshotContent(selectedSnapshot.id, newPath);
  };

  // Helper para limpiar nombres técnicos (V4.6.7)
  const formatDisplayName = (path: string, name: string) => {
    let clean = name || path.split('/').pop() || "";
    // Eliminar prefijos técnicos comunes que confunden al usuario
    clean = clean.replace(/^\/host_root\//i, "");
    clean = clean.replace(/^host_root\//i, "");
    clean = clean.replace(/^\/HOST_ROOT\//i, "");
    clean = clean.replace(/^HOST_ROOT\//i, "");
    return clean.toUpperCase();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/90 backdrop-blur-md animate-in fade-in duration-500">
      <div className="bg-gray-950 border border-gray-900 w-full max-w-2xl rounded-[2.5rem] overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
        
        {/* Header */}
        <div className="bg-gray-900/50 p-6 border-b border-gray-900 flex justify-between items-center tracking-tighter">
          <div className="flex items-center gap-4">
             <div className="p-3 bg-blue-500/10 rounded-2xl border border-blue-500/20 text-blue-400">
                <RotateCcw size={24} />
             </div>
             <div>
                <h3 className="text-lg font-black text-white uppercase italic">Restore Wizard Pro</h3>
                <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[9px] text-emerald-500 font-black uppercase tracking-widest bg-emerald-500/5 px-2 py-0.5 rounded-full border border-emerald-500/10">V4.7.2 SUPER TURBO</span>
                    <span className="text-[9px] text-gray-600 font-bold uppercase tracking-widest">{agentId}</span>
                </div>
             </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-full text-gray-500 hover:text-white transition-all">
            <X size={20} />
          </button>
        </div>

        {/* Wizard Steps Progress */}
        <div className="px-12 pt-6 flex items-center justify-between pointer-events-none">
            {[1,2,3].map((s) => (
                <div key={s} className="flex items-center flex-1 last:flex-none">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black border transition-all ${step >= s ? 'bg-blue-600 border-blue-400 text-white shadow-lg' : 'bg-gray-900 border-gray-800 text-gray-600'}`}>
                        {s < step ? '✓' : s}
                    </div>
                    {s < 3 && <div className={`h-[1px] flex-1 mx-2 ${step > s ? 'bg-blue-600' : 'bg-gray-800'}`}></div>}
                </div>
            ))}
        </div>

        <div className="p-8 overflow-y-auto custom-scrollbar flex-1">
          
          {/* STEP 1: SELECT DATE/TIME */}
          {step === 1 && (
            <div className="space-y-6 animate-in slide-in-from-bottom-4 duration-500">
                <div className="space-y-2">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2">
                        <Calendar size={14} className="text-blue-500" />
                        Puntos de Restauración Disponibles
                    </h4>
                    <p className="text-[10px] text-gray-500 uppercase font-bold tracking-tighter text-blue-400/50 italic">Selecciona una fecha para explorar su contenido</p>
                </div>

                <div className="grid grid-cols-1 gap-3">
                    {sortedSnapshots.map((s: any) => (
                        <button 
                            key={s.id}
                            onClick={() => { setSelectedSnapshot(s); fetchSnapshotContent(s.id, ""); }}
                            disabled={isLodingContent}
                            className="w-full flex items-center justify-between p-5 bg-gray-900/40 border border-gray-800 rounded-3xl hover:border-blue-500/40 hover:bg-blue-500/[0.02] transition-all group text-left"
                        >
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-gray-800 rounded-2xl text-gray-600 group-hover:text-blue-500 transition-colors">
                                    <Clock size={18} />
                                </div>
                                <div className="flex flex-col">
                                    <span className="text-base font-black text-white italic group-hover:text-blue-400 transition-colors uppercase tracking-tighter">
                                        {new Date(s.time).toLocaleDateString()} — {new Date(s.time).toLocaleTimeString()}
                                    </span>
                                    <span className="text-[9px] font-mono text-gray-600 tracking-tighter mt-0.5">{s.short_id || s.id}</span>
                                </div>
                            </div>
                            {isLodingContent && selectedSnapshot?.id === s.id ? (
                                <Activity size={18} className="text-blue-500 animate-spin" />
                            ) : (
                                <ChevronRight size={20} className="text-gray-700 group-hover:text-blue-500 transition-all" />
                            )}
                        </button>
                    ))}
                </div>
            </div>
          )}

          {/* STEP 2: DYNAMIC EXPLORER (V4.6.7 Estetica Screenshot) */}
          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <div className="flex justify-between items-center bg-gray-900/30 p-4 rounded-3xl border border-gray-900">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-blue-500/10 rounded-xl text-blue-500 border border-blue-500/20">
                            <Database size={16} />
                        </div>
                        <div className="flex items-center gap-2">
                            <button onClick={() => setStep(1)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white transition-colors">
                                {agentId.toUpperCase()}
                            </button>
                            <span className="text-gray-700">/</span>
                            <span className="text-[10px] text-blue-400 font-black uppercase tracking-widest italic">{currentPath || "ROOT"}</span>
                        </div>
                    </div>
                    <div className="flex gap-2">
                        <button onClick={selectAll} className="text-[9px] px-3 py-1 bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-lg hover:bg-emerald-500/20 transition-all font-black uppercase">SELECT ALL</button>
                        <button onClick={clearSelection} className="text-[9px] px-3 py-1 bg-gray-800 text-gray-400 border border-gray-700 rounded-lg hover:bg-white/5 hover:text-white transition-all font-black uppercase">CLEAR</button>
                    </div>
                </div>

                {currentPath && (
                    <button onClick={navigateBack} className="flex items-center gap-2 text-[10px] text-gray-500 font-black uppercase hover:text-blue-400 transition-colors ml-2">
                        ← .. VOLVER ATRÁS
                    </button>
                )}
                       {/* Lista de Archivos estilo Imagen 2 (V4.7.1 Grid 2 Columnas) */}
                <div className="grid grid-cols-2 gap-3 max-h-[45vh] overflow-y-auto custom-scrollbar pr-2 relative p-1">
                    {isLodingContent && (
                        <div className="absolute inset-0 bg-black/40 backdrop-blur-sm z-20 flex items-center justify-center rounded-3xl">
                            <Activity size={32} className="text-blue-500 animate-spin" />
                        </div>
                    )}
                    
                    {explorerContent.length === 0 && !isLodingContent && (
                         <div className="col-span-2 p-12 text-center border-2 border-dashed border-gray-900 rounded-3xl">
                            <p className="text-[10px] text-gray-600 font-black uppercase italic">Carpeta Vacía</p>
                         </div>
                    )}

                    {explorerContent.map((item: any, idx: number) => {
                        const path = item.path;
                        const isDir = item.type === "dir";
                        const isSelected = selectedPaths.includes(path);
                        const displayName = formatDisplayName(path, item.name);
                        
                        return (
                            <div 
                                key={idx} 
                                className={`flex items-center justify-between p-3 bg-gray-900/40 border rounded-2xl transition-all group cursor-pointer ${isSelected ? 'border-emerald-500/40 bg-emerald-500/[0.02]' : 'border-gray-900 hover:border-gray-800'}`}
                                onClick={() => isDir ? fetchSnapshotContent(selectedSnapshot.id, path) : togglePath(path)}
                            >
                                <div className="flex items-center gap-3 flex-1 min-w-0">
                                    <div className={`p-2 rounded-xl shrink-0 ${isDir ? 'bg-emerald-500/10 text-emerald-500' : 'bg-gray-800 text-gray-500'} group-hover:scale-105 transition-transform`}>
                                        {isDir ? <Folder size={16} /> : <FileText size={16} />}
                                    </div>
                                    <div className="flex flex-col min-w-0">
                                        <div className="flex items-center gap-1.5 truncate">
                                            {isDir && item.path.includes("Volume") && <span className="text-[8px] text-emerald-500 font-black italic shrink-0">VOL</span>}
                                            <span className={`text-[12px] font-black uppercase italic tracking-tighter truncate ${isSelected ? 'text-white' : 'text-gray-300'}`}>
                                                {displayName}
                                            </span>
                                        </div>
                                        <span className="text-[8px] text-gray-600 font-mono italic truncate opacity-40 group-hover:opacity-100 transition-opacity whitespace-nowrap overflow-hidden text-ellipsis">
                                            {path}
                                        </span>
                                    </div>
                                </div>
                                
                                <div className="flex items-center gap-2 shrink-0">
                                    {isSelected && <div className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>}
                                    {isDir && <ChevronRight size={14} className="text-gray-700 group-hover:text-blue-500" />}
                                </div>
                            </div>
                        );
                    })}
                </div>

                {/* Footer del Paso 2 idéntico a Imagen 2 */}
                <div className="flex items-center justify-between pt-4 border-t border-gray-900 mt-2">
                    <div className="flex items-center gap-2 ml-2">
                        <div className="w-5 h-5 bg-emerald-500/10 rounded-lg flex items-center justify-center text-emerald-500">
                             <ShieldCheck size={12} />
                        </div>
                        <span className="text-[10px] text-emerald-500 font-black uppercase italic tracking-widest">
                            {selectedPaths.length} TARGETS READY
                        </span>
                    </div>

                    <button 
                       onClick={() => setStep(3)} 
                       disabled={selectedPaths.length === 0}
                       className="bg-[#059669] hover:bg-[#10b981] disabled:bg-gray-900 disabled:text-gray-700 text-white font-black uppercase text-xs px-10 py-5 rounded-2xl transition-all shadow-xl active:scale-[0.98] flex items-center gap-3"
                    >
                        LOCK SELECTION
                    </button>
                </div>
            </div>
          )}

          {/* STEP 3: DESTINATION & CONFIRM */}
          {step === 3 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <button onClick={() => setStep(2)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white flex items-center gap-1 transition-colors">
                    ← VOLVER AL EXPLORADOR
                </button>

                <div className="space-y-4">
                   <div className="p-6 bg-emerald-500/[0.03] border border-emerald-500/10 rounded-[2rem]">
                        <h4 className="text-[10px] text-emerald-500 font-black uppercase italic mb-4">Resumen de Reconstrucción</h4>
                        <div className="space-y-2 max-h-[15vh] overflow-y-auto custom-scrollbar">
                            {selectedPaths.map(p => (
                                <div key={p} className="flex items-center gap-2 text-[10px] font-mono text-gray-400 truncate bg-black/20 p-2 rounded-xl border border-white/5">
                                    <ShieldCheck size={12} className="text-emerald-500" /> {p}
                                </div>
                            ))}
                        </div>
                   </div>

                   <div className="space-y-2">
                        <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1 italic flex items-center gap-2">
                            <FolderInput size={12} className="text-emerald-500" />
                            Ruta de Destino en Servidor
                        </label>
                        <input 
                          type="text" 
                          value={restorePath} 
                          onChange={(e) => setRestorePath(e.target.value)}
                          className="w-full bg-black/40 border border-gray-800 rounded-2xl px-6 py-5 text-xs text-white focus:border-blue-500 outline-none font-mono tracking-widest"
                          placeholder="/restore_data"
                        />
                   </div>

                   <button 
                    onClick={handleRestore}
                    disabled={loading}
                    className="w-full bg-blue-600 hover:bg-blue-500 text-white font-black uppercase text-xs py-6 rounded-3xl transition-all shadow-xl shadow-blue-950/40 flex items-center justify-center gap-3 active:scale-[0.98]"
                   >
                    {loading ? <Activity className="animate-spin" size={18} /> : <ShieldCheck size={18} />}
                    {loading ? 'Sincronizando...' : 'INICIAR RESTAURACIÓN'}
                  </button>
                </div>
            </div>
          )}

          {/* STEP 4: SUCCESS */}
          {step === 4 && (
            <div className="text-center space-y-8 animate-in zoom-in-95 duration-700 py-12">
               <div className="w-24 h-24 bg-emerald-500/10 border border-emerald-500/20 rounded-[2.5rem] flex items-center justify-center mx-auto text-emerald-500 shadow-2xl shadow-emerald-500/10 animate-bounce">
                  <ShieldCheck size={48} />
               </div>
               <div className="space-y-3">
                  <h3 className="text-2xl font-black text-white italic uppercase tracking-tighter">RECONSTRUCTION COMMENCED</h3>
                  <p className="text-[10px] text-gray-500 font-bold uppercase tracking-widest max-w-sm mx-auto leading-relaxed">
                     La tarea ha sido agendada. Los archivos aparecerán en <span className="text-emerald-500 font-mono italic">{restorePath}</span> pronto.
                  </p>
               </div>
               <button onClick={onClose} className="block w-full text-[10px] text-gray-600 font-black uppercase tracking-widest hover:text-white transition-colors">
                  Cerrar Asistente
               </button>
            </div>
          )}

        </div>
      </div>
    </div>
  );
}
