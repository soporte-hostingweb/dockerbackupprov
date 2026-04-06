'use client';
import { useState } from "react";
import { X, RotateCcw, FolderInput, ShieldCheck, Database, HardDrive, AlertCircle } from "lucide-react";

interface RestoreModalProps {
  isOpen: boolean;
  onClose: () => void;
  agentId: string;
  snapshots: any[];
  token: string;
}

export default function RestoreModal({ isOpen, onClose, agentId, snapshots, token }: RestoreModalProps) {
  const [selectedSnapshot, setSelectedSnapshot] = useState("");
  const [restorePath, setRestorePath] = useState("/restore_data");
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(1);

  if (!isOpen) return null;

  const handleRestore = async () => {
    if (!selectedSnapshot) return alert("Selecciona un snapshot primero");
    
    setLoading(true);
    try {
      const resp = await fetch(`https://api.hwperu.com/v1/agent/action/${agentId}`, {
        method: "POST",
        headers: { 
            "Authorization": token,
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ 
            action: "restore",
            snapshot_id: selectedSnapshot,
            destination: restorePath
        })
      });

      if (resp.ok) {
        setStep(3); // Éxito
      } else {
        alert("Error al iniciar restauración");
      }
    } catch (err) {
      alert("Error de red");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-gray-950 border border-gray-900 w-full max-w-lg rounded-3xl overflow-hidden shadow-2xl shadow-blue-900/10">
        
        {/* Header */}
        <div className="bg-gray-900/50 p-6 border-b border-gray-900 flex justify-between items-center">
          <div className="flex items-center gap-3">
             <div className="p-2 bg-blue-500/10 rounded-lg border border-blue-500/20 text-blue-500">
                <RotateCcw size={20} />
             </div>
             <div>
                <h3 className="text-sm font-black text-white uppercase italic tracking-widest">Restore Wizard</h3>
                <p className="text-[10px] text-gray-500 font-bold uppercase tracking-tighter">Infrastructure Recovery Engine</p>
             </div>
          </div>
          <button onClick={onClose} className="text-gray-500 hover:text-white transition-colors">
            <X size={20} />
          </button>
        </div>

        <div className="p-8 space-y-6">
          {step === 1 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
                <div className="space-y-4">
                   <div className="flex items-center gap-2 text-[10px] text-gray-500 font-black uppercase tracking-widest">
                      <Database size={12} />
                      1. Select Recovery Point
                   </div>
                   <div className="max-h-48 overflow-y-auto custom-scrollbar border border-gray-900 rounded-xl bg-black/40">
                      {snapshots && snapshots.length > 0 ? (
                        snapshots.map((s: any) => (
                           <label key={s.id || s.short_id} className={`flex items-center justify-between p-4 cursor-pointer hover:bg-white/5 transition-colors border-b border-gray-900 last:border-0 ${selectedSnapshot === (s.id || s.short_id) ? 'bg-blue-500/10' : ''}`}>
                              <div className="flex flex-col">
                                 <span className="text-xs font-bold text-white font-mono">{s.short_id || s.id}</span>
                                 <span className="text-[10px] text-gray-500">{new Date(s.time).toLocaleString()}</span>
                              </div>
                              <input 
                                type="radio" 
                                name="snapshot" 
                                className="w-4 h-4 accent-blue-500"
                                onChange={() => setSelectedSnapshot(s.id || s.short_id)}
                                checked={selectedSnapshot === (s.id || s.short_id)}
                              />
                           </label>
                        ))
                      ) : (
                        <div className="p-8 text-center text-xs text-gray-600 font-bold uppercase italic">No snapshots found for this agent</div>
                      )}
                   </div>
                </div>
                <button 
                  onClick={() => setStep(2)}
                  disabled={!selectedSnapshot}
                  className="w-full bg-blue-600 disabled:opacity-50 hover:bg-blue-500 text-white font-black uppercase text-xs py-4 rounded-2xl shadow-xl shadow-blue-900/20 transition-all active:scale-[0.98]"
                >
                  Continue to Destination
                </button>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
               <div className="space-y-4">
                  <div className="flex items-center gap-2 text-[10px] text-gray-500 font-black uppercase tracking-widest">
                      <FolderInput size={12} />
                      2. Destination Path
                  </div>
                  <input 
                    type="text" 
                    value={restorePath}
                    onChange={(e) => setRestorePath(e.target.value)}
                    placeholder="/restore/here"
                    className="w-full bg-black/60 border border-gray-800 rounded-xl px-4 py-3 text-sm text-white focus:border-blue-500 outline-none transition-all font-mono"
                  />
                  <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl flex items-start gap-3">
                     <AlertCircle className="text-amber-500 flex-shrink-0" size={16} />
                     <p className="text-[9px] text-amber-500/80 font-bold uppercase leading-relaxed">
                        WARNNIG: Files already existing in the destination path will be overwritten by Restic. We recommend using a new directory.
                     </p>
                  </div>
               </div>

               <div className="flex gap-4">
                  <button onClick={() => setStep(1)} className="flex-1 text-gray-500 hover:text-white font-black uppercase text-xs py-4">Back</button>
                  <button 
                    onClick={handleRestore}
                    disabled={loading}
                    className="flex-[2] bg-emerald-600 hover:bg-emerald-500 text-white font-black uppercase text-xs py-4 rounded-2xl shadow-xl shadow-emerald-900/20 transition-all active:scale-[0.98]"
                  >
                    {loading ? 'INITIATING...' : 'Start Restoration'}
                  </button>
               </div>
            </div>
          )}

          {step === 3 && (
            <div className="p-8 text-center space-y-6 animate-in zoom-in-95 duration-500">
               <div className="w-16 h-16 bg-emerald-500/20 rounded-full flex items-center justify-center mx-auto border border-emerald-500/30 text-emerald-500">
                  <ShieldCheck size={32} />
               </div>
               <div>
                  <h4 className="text-lg font-black text-white uppercase italic">Restore Instruction Sent</h4>
                  <p className="text-[10px] text-gray-500 font-bold uppercase tracking-tighter mt-2 max-w-xs mx-auto">
                    The agent has received your request. You can monitor the progress in the agent logs or wait for the system to notify completion.
                  </p>
               </div>
               <button 
                 onClick={onClose}
                 className="w-full bg-gray-900 hover:bg-gray-800 text-white font-black uppercase text-xs py-4 rounded-2xl"
               >
                 Close Wizard
               </button>
            </div>
          )}
        </div>

      </div>
    </div>
  );
}
