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
  const [selectedFolder, setSelectedFolder] = useState<string | null>(null);
  const [selectedSnapshot, setSelectedSnapshot] = useState<any>(null);
  const [selectedPaths, setSelectedPaths] = useState<string[]>([]);
  const [restorePath, setRestorePath] = useState("/restore_data");
  const [explorerContent, setExplorerContent] = useState<any[]>([]);
  const [isLodingContent, setIsLoadingContent] = useState(false);
  const [agentData, setAgentData] = useState<any>(null);

  // Cargar datos del agente (espacio en disco) al abrir
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

  // EXTRAER CARPETAS ÚNICAS (Paso 1)
  const uniquePaths = Array.from(new Set(snapshots.flatMap(s => s.paths || []))).sort();

  // FILTRAR SNAPSHOTS PARA LA CARPETA ELEGIDA (Paso 2)
  const availableSnapshots = selectedFolder 
    ? snapshots.filter(s => s.paths?.includes(selectedFolder))
               .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
    : [];

  const togglePath = (path: string) => {
    if (selectedPaths.includes(path)) {
        setSelectedPaths(selectedPaths.filter(p => p !== path));
    } else {
        setSelectedPaths([...selectedPaths, path]);
    }
  };

  // Solicitar listado de archivos
  const fetchSnapshotContent = async (snapId: string) => {
    setIsLoadingContent(true);
    try {
        const folder = selectedFolder || "";
        const cleanPath = folder.replace('📂 ', '').replace('📄 ', '');
        await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
            method: "POST",
            headers: { "Authorization": token, "Content-Type": "application/json" },
            body: JSON.stringify({ action: "ls_snapshot", snapshot_id: snapId, path: cleanPath })
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
                    // V4.6.2: Si es un objeto único (un solo archivo), lo convertimos en array
                    if (!Array.isArray(parsed)) {
                        parsed = [parsed];
                    }
                    setExplorerContent(parsed);
                    setIsLoadingContent(false);
                    setStep(3); 
                } catch (e) {
                    console.error("Error parsing content:", e);
                    // Si falla el parseo, quizá sea un error crudo del agente
                    alert("Error al procesar el contenido del servidor. Intenta de nuevo.");
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
                    <span className="text-[9px] text-blue-500 font-black uppercase tracking-widest bg-blue-500/5 px-2 py-0.5 rounded-full border border-blue-500/10">V4.6.2 PRECISE</span>
                    <span className="text-[9px] text-gray-600 font-bold uppercase tracking-widest">{agentId}</span>
                </div>
             </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-full text-gray-500 hover:text-white transition-all">
            <X size={20} />
          </button>
        </div>

        {/* Wizard Steps Progress */}
        <div className="px-8 pt-6 flex items-center justify-between pointer-events-none">
            {[1,2,3,4].map((s) => (
                <div key={s} className="flex items-center flex-1 last:flex-none">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black border transition-all ${step >= s ? 'bg-blue-600 border-blue-400 text-white shadow-lg' : 'bg-gray-900 border-gray-800 text-gray-600'}`}>
                        {s < step ? '✓' : s}
                    </div>
                    {s < 4 && <div className={`h-[1px] flex-1 mx-2 ${step > s ? 'bg-blue-600' : 'bg-gray-800'}`}></div>}
                </div>
            ))}
        </div>

        <div className="p-8 overflow-y-auto custom-scrollbar flex-1">
          
          {/* STEP 1: SELECT FOLDER */}
          {step === 1 && (
            <div className="space-y-6 animate-in slide-in-from-bottom-4 duration-500">
                <div className="space-y-2">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2">
                        <Folder size={14} className="text-blue-500" />
                        ¿Qué recurso deseas recuperar?
                    </h4>
                    <p className="text-[10px] text-gray-500 uppercase font-bold tracking-tighter">Listado de volúmenes detectados en el storage</p>
                </div>

                <div className="space-y-3">
                    {uniquePaths.length > 0 ? uniquePaths.map((path) => (
                        <button 
                            key={path}
                            onClick={() => { setSelectedFolder(path); setStep(2); }}
                            className="w-full p-5 bg-gray-900/40 border border-gray-800 hover:border-blue-500/40 hover:bg-blue-500/[0.03] text-left rounded-3xl transition-all group flex items-center justify-between"
                        >
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-gray-800 rounded-2xl text-blue-500/50 group-hover:text-blue-400 transition-colors">
                                    <HardDrive size={18} />
                                </div>
                                <div>
                                    <span className="text-xs font-black text-white uppercase tracking-tight">{path}</span>
                                    <p className="text-[9px] text-gray-600 font-bold uppercase mt-0.5">Detectado en {snapshots.filter(s => s.paths?.includes(path)).length} copias</p>
                                </div>
                            </div>
                            <ChevronRight size={16} className="text-gray-700 group-hover:text-blue-500 group-hover:translate-x-1 transition-all" />
                        </button>
                    )) : (
                        <div className="p-20 text-center border-2 border-dashed border-gray-900 rounded-3xl">
                            <Database className="mx-auto text-gray-800 mb-4" size={40} />
                            <p className="text-xs text-gray-600 font-black uppercase italic">No se detectaron recursos en la nube</p>
                        </div>
                    )}
                </div>
            </div>
          )}

          {/* STEP 2: SELECT VERSION */}
          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <button onClick={() => setStep(1)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white flex items-center gap-1 transition-colors">
                    ← Volver a Selección de Carpeta
                </button>
                <div className="space-y-2">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2">
                        <Calendar size={14} className="text-blue-500" />
                        Puntos de Restauración: {selectedFolder}
                    </h4>
                </div>

                <div className="space-y-3">
                    {availableSnapshots.map((s: any) => (
                        <button 
                            key={s.id}
                            onClick={() => { setSelectedSnapshot(s); fetchSnapshotContent(s.id); }}
                            disabled={isLodingContent}
                            className="w-full flex items-center justify-between p-5 bg-gray-900/40 border border-gray-800 rounded-3xl hover:border-emerald-500/40 hover:bg-emerald-500/[0.02] transition-all group text-left"
                        >
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-gray-800 rounded-2xl text-gray-500 group-hover:text-emerald-500 transition-colors">
                                    <Clock size={18} />
                                </div>
                                <div>
                                    <span className="text-base font-black text-white italic group-hover:text-emerald-400 transition-colors">
                                        {new Date(s.time).toLocaleDateString()} - {new Date(s.time).toLocaleTimeString()}
                                    </span>
                                    <div className="flex items-center gap-3 mt-0.5">
                                        <span className="text-[9px] font-mono text-gray-500 tracking-tighter">{s.short_id || s.id}</span>
                                    </div>
                                </div>
                            </div>
                            {isLodingContent && selectedSnapshot?.id === s.id ? (
                                <Activity size={18} className="text-blue-500 animate-spin" />
                            ) : (
                                <ChevronRight size={20} className="text-gray-700 group-hover:text-emerald-500 transition-all" />
                            )}
                        </button>
                    ))}
                </div>
            </div>
          )}

          {/* STEP 3: SELECT FILES & DESTINATION */}
          {step === 3 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-500">
                <button onClick={() => setStep(2)} className="text-[10px] text-gray-500 font-black uppercase tracking-widest hover:text-white flex items-center gap-1 transition-colors">
                    ← Volver a Línea de Tiempo
                </button>
                
                {/* Alerta de Capacidad */}
                <div className={`p-4 rounded-2xl border flex items-center gap-4 ${agentData?.free_space?.includes('GB') ? 'bg-blue-500/5 border-blue-500/20' : 'bg-red-500/5 border-red-500/20'}`}>
                    <div className={agentData?.free_space?.includes('GB') ? 'text-blue-500' : 'text-red-500'}>
                        <HardDrive size={24} />
                    </div>
                    <div>
                        <p className="text-[10px] text-gray-500 font-black uppercase">Capacidad del Servidor Destino</p>
                        <p className="text-xs font-black text-white uppercase italic">
                            {agentData?.free_space || 'Desconocido'} libres de {agentData?.total_space || '---'}
                        </p>
                        <p className="text-[9px] text-gray-600 font-bold mt-1 uppercase">
                            Snapshot completo detectado: ~{(selectedSnapshot?.size || 0) / (1024*1024)} MB
                        </p>
                    </div>
                </div>

                <div className="space-y-2">
                    <h4 className="text-xs font-black text-white uppercase tracking-widest flex items-center gap-2">
                        <CheckSquare size={14} className="text-blue-500" />
                        Selección Granular de Archivos
                    </h4>
                    <p className="text-[9px] text-gray-500 uppercase font-black italic">Si no seleccionas nada, se restaurará el volumen completo.</p>
                </div>

                <div className="bg-black/30 border border-gray-900 rounded-3xl overflow-hidden max-h-[30vh] overflow-y-auto custom-scrollbar">
                    {explorerContent.map((item: string, idx: number) => (
                        <div key={idx} className="flex items-center justify-between p-4 border-b border-gray-900 group hover:bg-white/[0.02]">
                            <div className="flex items-center gap-3">
                                {item.startsWith("📂") ? <Folder size={14} className="text-blue-400/50" /> : <FileText size={14} className="text-gray-500" />}
                                <span className="text-[11px] font-mono text-gray-300">{item}</span>
                            </div>
                            <button onClick={() => togglePath(item)} className="p-1 hover:text-white transition-colors">
                                {selectedPaths.includes(item) ? <CheckSquare size={18} className="text-emerald-500" /> : <Square size={18} className="text-gray-800" />}
                            </button>
                        </div>
                    ))}
                </div>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-[10px] text-gray-500 font-black uppercase tracking-widest ml-1 italic group flex items-center gap-2">
                        <FolderInput size={12} className="text-blue-500" />
                        Ruta de Destino en el Servidor
                    </label>
                    <input 
                      type="text" 
                      value={restorePath} 
                      onChange={(e) => setRestorePath(e.target.value)}
                      className="w-full bg-black/40 border border-gray-800 rounded-2xl px-6 py-4 text-xs text-white focus:border-blue-500 outline-none font-mono tracking-widest"
                      placeholder="/restore_data"
                    />
                  </div>

                  <div className="p-5 bg-orange-500/5 border border-orange-500/10 rounded-2xl flex gap-4 items-start">
                    <AlertCircle className="text-orange-500 shrink-0" size={18} />
                    <p className="text-[9px] text-orange-400/70 font-black uppercase leading-relaxed tracking-wider">
                       ADVERTENCIA: Los archivos existentes en la ruta de destino serán SOBRESCRITOS. Se recomienda usar un directorio vacío.
                    </p>
                  </div>

                  <button 
                    onClick={handleRestore}
                    disabled={loading}
                    className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-xs py-5 rounded-3xl transition-all shadow-xl shadow-emerald-950/40 flex items-center justify-center gap-3 active:scale-[0.98]"
                  >
                    {loading ? <Activity className="animate-spin" size={18} /> : <ShieldCheck size={18} />}
                    {loading ? 'MODULANDO RECONSTRUCCIÓN...' : 'INICIAR RESTAURACIÓN'}
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
                     La tarea ha sido agendada en el agente. Los archivos aparecerán en <span className="text-emerald-500 font-mono italic">{restorePath}</span> en breves momentos.
                  </p>
               </div>
               <div className="bg-gray-900/30 p-6 rounded-3xl border border-gray-900 inline-block">
                  <div className="flex items-center gap-4 text-left">
                     <div className="text-gray-600"><Clock size={20} /></div>
                     <div>
                        <p className="text-[9px] text-gray-500 font-black uppercase">Estado del Proceso</p>
                        <p className="text-xs text-white font-bold uppercase italic">Vigilando PID del Agente...</p>
                     </div>
                  </div>
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
