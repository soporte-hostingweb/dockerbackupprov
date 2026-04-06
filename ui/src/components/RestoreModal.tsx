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
    // No limpiamos explorerContent inmediatamente para evitar parpadeo si es rápido
    try {
        await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ action: "ls_snapshot", snapshot_id: snapId, path: path })
        });

        let attempts = 0;
        const poll = setInterval(async () => {
            attempts++;
            const resp = await fetch(`https://api.hwperu.com/v1/agent/status`, {
                headers: { "Authorization": token }
            });
            const data = await resp.json();
            const currentAgent = data[agentId];

            if (currentAgent.cmd_result && currentAgent.cmd_result !== "loading" && currentAgent.cmd_result !== "none") {
                clearInterval(poll);
                try {
                    let parsed = JSON.parse(currentAgent.cmd_result);
                    if (!Array.isArray(parsed)) parsed = [parsed];
                    setExplorerContent(parsed);
                    setIsLoadingContent(false);
                    if (step === 1) setStep(2); 
                } catch (e) {
                    console.error("Error parsing content:", e);
                    setIsLoadingContent(false);
                }
            }

            if (attempts > 30) { 
                clearInterval(poll);
                alert("Timeout al esperar respuesta del agente.");
                setIsLoadingContent(false);
            }
        }, 2000);

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
                    <span className="text-[9px] text-blue-500 font-black uppercase tracking-widest bg-blue-500/5 px-2 py-0.5 rounded-full border border-blue-500/10">V4.6.5 OPTIMIZED</span>
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

          {/* STEP 2: DYNAMIC EXPLORER (V4.6.5 Lazy Loading) */}
          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <button onClick={() => setStep(1)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white flex items-center gap-1 transition-colors">
                            ← Inicio
                        </button>
                        {currentPath && (
                            <>
                                <span className="text-gray-800">/</span>
                                <button onClick={navigateBack} className="text-[10px] text-blue-500 font-black uppercase tracking-widest hover:text-white transition-colors">
                                    .. VOLVER
                                </button>
                            </>
                        )}
                    </div>
                    <div className="flex gap-2">
                        <button onClick={selectAll} className="text-[9px] px-3 py-1 bg-gray-900 text-gray-400 border border-gray-800 rounded-full hover:bg-blue-500/10 hover:text-blue-500 transition-all font-black uppercase">Select All</button>
                        <button onClick={clearSelection} className="text-[9px] px-3 py-1 bg-gray-900 text-gray-400 border border-gray-800 rounded-full hover:bg-red-500/10 hover:text-red-500 transition-all font-black uppercase">Clear</button>
                    </div>
                </div>

                <div className="space-y-2">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2 italic">
                        <Folder size={14} className="text-emerald-500" />
                        {currentPath || "/"}
                    </h4>
                </div>

                {/* File List */}
                <div className="bg-black/30 border border-gray-900 rounded-[2rem] overflow-hidden max-h-[35vh] flex flex-col relative">
                    {isLodingContent && (
                        <div className="absolute inset-0 bg-black/60 backdrop-blur-sm z-10 flex items-center justify-center">
                            <Activity size={24} className="text-blue-500 animate-spin" />
                        </div>
                    )}
                    <div className="overflow-y-auto custom-scrollbar flex-1">
                        {explorerContent.map((item: any, idx: number) => {
                            const path = item.path;
                            const isDir = item.type === "dir";
                            const cleanName = item.name;
                            const isSelected = selectedPaths.includes(path);
                            
                            return (
                                <div key={idx} className="flex items-center justify-between p-4 px-6 border-b border-gray-900/50 group hover:bg-emerald-500/[0.03] transition-all">
                                    <div 
                                        className="flex items-center gap-4 flex-1 cursor-pointer"
                                        onClick={() => isDir ? fetchSnapshotContent(selectedSnapshot.id, path) : togglePath(path)}
                                    >
                                        <div className={`p-2 rounded-xl ${isDir ? 'bg-emerald-500/10 text-emerald-500/60' : 'bg-gray-800 text-gray-500'} group-hover:scale-110 transition-transform`}>
                                            {isDir ? <Folder size={16} /> : <FileText size={16} />}
                                        </div>
                                        <div className="flex flex-col">
                                            <span className={`text-xs font-bold uppercase tracking-tighter ${isDir ? 'text-blue-400' : 'text-gray-200'}`}>
                                                {cleanName} {isDir && "→"}
                                            </span>
                                            {item.size > 0 && <span className="text-[9px] text-gray-600 font-mono">{(item.size / 1024).toFixed(1)} KB</span>}
                                        </div>
                                    </div>
                                    <button 
                                        onClick={() => togglePath(path)}
                                        className={`w-5 h-5 rounded-md border flex items-center justify-center transition-all ${isSelected ? 'bg-emerald-600 border-emerald-400' : 'bg-gray-900 border-gray-800'}`}
                                    >
                                        {isSelected && <ShieldCheck size={14} className="text-white" />}
                                    </button>
                                </div>
                            );
                        })}
                    </div>
                </div>

                <button 
                   onClick={() => setStep(3)} 
                   disabled={selectedPaths.length === 0}
                   className="w-full bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-900 disabled:text-gray-700 text-white font-black uppercase text-xs py-5 rounded-3xl transition-all shadow-xl shadow-emerald-950/40 flex items-center justify-center gap-3 active:scale-[0.98]"
                >
                    <FolderInput size={18} />
                    LOCK SELECTION ({selectedPaths.length})
                </button>
            </div>
          )}

          {/* STEP 3: DESTINATION & CONFIRM */}
          {step === 3 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <button onClick={() => setStep(2)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white flex items-center gap-1 transition-colors">
                    ← Volver al Explorador
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
                            Ruta de Destino
                        </label>
                        <input 
                          type="text" 
                          value={restorePath} 
                          onChange={(e) => setRestorePath(e.target.value)}
                          className="w-full bg-black/40 border border-gray-800 rounded-2xl px-6 py-4 text-xs text-white focus:border-emerald-500 outline-none font-mono tracking-widest"
                          placeholder="/restore_data"
                        />
                   </div>

                   <button 
                    onClick={handleRestore}
                    disabled={loading}
                    className="w-full bg-blue-600 hover:bg-blue-500 text-white font-black uppercase text-xs py-5 rounded-3xl transition-all shadow-xl shadow-blue-950/40 flex items-center justify-center gap-3 active:scale-[0.98]"
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
