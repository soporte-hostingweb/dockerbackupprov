'use client';
import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Clock, History, Activity, Box, Database, HardDrive, CheckCircle2, XCircle } from "lucide-react";

interface BackupActivity {
  id: number;
  agent_id: string;
  status: string;
  snapshot_id: string;
  size_mb: number;
  duration_secs: number;
  timestamp: string;
}

export default function HistoryPage() {
  const searchParams = useSearchParams();
  const sso = searchParams.get("sso");
  const [activities, setActivities] = useState<BackupActivity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchHistory() {
      try {
        const response = await fetch("https://api.hwperu.com/v1/history", {
          headers: { "Authorization": sso || localStorage.getItem("dbp_sso_token") || "" }
        });
        if (response.ok) {
          const data = await response.json();
          setActivities(data);
        }
      } catch (err) {
        console.error("Failed to fetch history:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchHistory();
  }, [sso]);

  return (
    <div className="max-w-7xl mx-auto p-8 space-y-8 animate-in fade-in duration-500">
      <div className="flex flex-col">
          <h1 className="text-3xl font-black tracking-tight text-white italic flex items-center gap-3 uppercase">
             <History className="text-emerald-500" size={32} />
             Backup History
          </h1>
          <p className="text-[10px] text-gray-500 uppercase tracking-[0.4em] mt-1 font-black">Chronological ledger of infrastructure snapshots</p>
      </div>

      {loading ? (
        <div className="p-20 text-center animate-pulse">
           <Activity className="h-10 w-10 text-emerald-500 mx-auto animate-spin mb-4" />
           <span className="text-xs text-gray-600 font-black uppercase tracking-widest">Scanning HWPeru History...</span>
        </div>
      ) : activities.length === 0 ? (
        <div className="bg-gray-950/20 border-2 border-dashed border-gray-900 rounded-3xl p-20 text-center flex flex-col items-center justify-center">
           <Box className="text-gray-800 w-16 h-16 mb-4" />
           <h3 className="text-lg font-black text-gray-700 uppercase tracking-tighter italic">No activities recorded yet</h3>
           <p className="text-[10px] text-gray-600 max-w-sm leading-relaxed uppercase mt-2 font-bold italic">
              When your agents complete a backup cycle, the metadata will appear here.
           </p>
        </div>
      ) : (
        <div className="bg-gray-950 border border-gray-900 rounded-3xl overflow-hidden shadow-2xl">
           <table className="w-full text-left border-collapse">
              <thead>
                 <tr className="bg-black/40 border-b border-gray-900">
                    <th className="px-6 py-5 text-[10px] font-black text-gray-500 uppercase tracking-widest">Timestamp</th>
                    <th className="px-6 py-5 text-[10px] font-black text-gray-500 uppercase tracking-widest">Source Agent</th>
                    <th className="px-6 py-5 text-[10px] font-black text-gray-500 uppercase tracking-widest">Snapshot ID</th>
                    <th className="px-6 py-5 text-[10px] font-black text-gray-500 uppercase tracking-widest text-center">Size</th>
                    <th className="px-6 py-5 text-[10px] font-black text-gray-500 uppercase tracking-widest text-center">Status</th>
                 </tr>
              </thead>
              <tbody className="divide-y divide-gray-900/50">
                 {activities.map((act) => (
                    <tr key={act.id} className="hover:bg-emerald-500/[0.02] transition-colors group">
                       <td className="px-6 py-4">
                          <div className="flex flex-col">
                             <span className="text-xs text-gray-300 font-bold">{new Date(act.timestamp).toLocaleDateString()}</span>
                             <span className="text-[10px] text-gray-600 font-mono italic">{new Date(act.timestamp).toLocaleTimeString()}</span>
                          </div>
                       </td>
                       <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                             <div className="w-8 h-8 rounded-lg bg-gray-900 flex items-center justify-center border border-gray-800">
                                <Database size={14} className="text-emerald-500/50" />
                             </div>
                             <span className="text-xs font-black text-white italic uppercase tracking-tighter">{act.agent_id}</span>
                          </div>
                       </td>
                       <td className="px-6 py-4">
                          <span className="text-[10px] font-mono bg-black/50 px-3 py-1 rounded-md border border-gray-800 text-gray-400 group-hover:border-emerald-500/30 transition-colors">
                             {act.snapshot_id || 'N/A'}
                          </span>
                       </td>
                       <td className="px-6 py-4 text-center">
                          <div className="flex flex-col items-center">
                            <span className="text-xs font-bold text-gray-300 italic">{act.size_mb > 0 ? `${act.size_mb} MB` : 'Incremental'}</span>
                            <span className="text-[9px] text-gray-600 uppercase font-bold tracking-tighter">{act.duration_secs}s elapsed</span>
                          </div>
                       </td>
                       <td className="px-6 py-4">
                          <div className="flex justify-center">
                             {act.status === 'SUCCESS' ? (
                                <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-500 text-[10px] font-black uppercase tracking-widest">
                                   <CheckCircle2 size={12} /> Success
                                </div>
                             ) : (
                                <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-red-500/10 border border-red-500/20 text-red-500 text-[10px] font-black uppercase tracking-widest">
                                   <XCircle size={12} /> Failure
                                </div>
                             )}
                          </div>
                       </td>
                    </tr>
                 ))}
              </tbody>
           </table>
        </div>
      )}
    </div>
  );
}

