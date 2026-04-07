'use client';
import { useState, useEffect } from "react";
import { Activity, CheckCircle2, XCircle, Clock, Server } from "lucide-react";

interface ActivityLog {
  id: number;
  agent_id: string;
  type: string;
  status: string;
  message: string;
  started_at: string;
  finished_at: string;
}

export default function GlobalActivity({ token }: { token: string }) {
  const [activities, setActivities] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchActivities = async () => {
    try {
      const resp = await fetch(`https://api.hwperu.com/v1/activities`, {
        headers: { "Authorization": token }
      });
      const data = await resp.json();
      setActivities(data || []);
    } catch (err) {
      console.error("Error fetching activities:", err);
    } finally {
      setLoading(false);
    }
  };

  // Polling cada 5 segundos para tiempo real
  useEffect(() => {
    fetchActivities();
    const interval = setInterval(fetchActivities, 5000);
    return () => clearInterval(interval);
  }, [token]);

  if (loading && activities.length === 0) return null;

  return (
    <div className="bg-gray-950/50 border border-gray-900 rounded-[2rem] overflow-hidden flex flex-col h-full shadow-2xl">
      <div className="p-6 border-b border-gray-900 bg-gray-900/20 flex items-center justify-between">
        <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-500/10 rounded-xl text-emerald-500">
                <Activity size={18} className="animate-pulse" />
            </div>
            <h3 className="text-xs font-black text-white uppercase italic tracking-widest">Global Activity Monitor</h3>
        </div>
        <span className="text-[9px] text-gray-600 font-bold uppercase tracking-widest bg-black/40 px-2 py-1 rounded-lg border border-white/5">Real-Time</span>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-3">
        {activities.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 opacity-20">
            <Clock size={32} className="mb-2" />
            <p className="text-[10px] font-black uppercase italic">No recent activity</p>
          </div>
        ) : (
          activities.map((log) => (
            <div key={log.id} className="p-4 bg-gray-900/40 border border-gray-900 rounded-2xl hover:border-gray-800 transition-all group">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2">
                    <div className={`p-1.5 rounded-lg ${log.status === 'running' ? 'bg-blue-500/10 text-blue-400' : log.status === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                        {log.status === 'running' ? <Activity size={14} className="animate-spin" /> : log.status === 'success' ? <CheckCircle2 size={14} /> : <XCircle size={14} />}
                    </div>
                    <span className={`text-[10px] font-black uppercase italic tracking-tighter ${log.status === 'running' ? 'text-blue-400' : 'text-gray-400'}`}>
                        {log.type} — {log.agent_id}
                    </span>
                </div>
                <span className="text-[8px] text-gray-700 font-mono italic">
                    {new Date(log.started_at).toLocaleTimeString()}
                </span>
              </div>
              
              <p className="text-[10px] text-gray-500 font-bold uppercase tracking-widest mt-2 leading-relaxed line-clamp-1 group-hover:line-clamp-none transition-all">
                {log.message}
              </p>

              {log.status === 'running' && (
                <div className="mt-3 h-1 w-full bg-gray-800 rounded-full overflow-hidden">
                    <div className="h-full bg-blue-500 animate-progress-ind"></div>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
