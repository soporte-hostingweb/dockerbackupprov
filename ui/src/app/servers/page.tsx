'use client';
import { useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";
import ServerList from "@/components/ServerList";
import RestoreModal from "@/components/RestoreModal";

export default function ServersPage() {
  const searchParams = useSearchParams();
  const [token, setToken] = useState("");
  const [isRestoreOpen, setIsRestoreOpen] = useState(false);
  const [restoreAgentId, setRestoreAgentId] = useState("");
  const [restoreSnapshots, setRestoreSnapshots] = useState<any[]>([]);

  const [agents, setAgents] = useState({});
  const [plan, setPlan] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = async (ssoToken: string) => {
    try {
      const [statusResp, planResp] = await Promise.all([
        fetch('https://api.hwperu.com/v1/activities', { headers: { 'Authorization': ssoToken } }),
        fetch('https://api.hwperu.com/v1/tenant/plan', { headers: { 'Authorization': ssoToken } })
      ]);

      if (statusResp.ok) {
        const data = await statusResp.json();
        setAgents(data.agents || {});
      }
      if (planResp.ok) {
        const p = await planResp.json();
        setPlan(p);
      }
    } catch (err) {
      console.error("Fetch error:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const sso = searchParams.get('sso');
    const currentToken = sso || localStorage.getItem('dbp_token');
    
    if (sso) {
      setToken(sso);
      localStorage.setItem('dbp_token', sso);
    } else if (currentToken) {
      setToken(currentToken);
    }

    if (currentToken) {
      fetchData(currentToken);
      const interval = setInterval(() => fetchData(currentToken), 5000);
      return () => clearInterval(interval);
    }
  }, [searchParams]);

  return (
    <div className="max-w-7xl mx-auto p-8 space-y-8">
      <div className="flex flex-col">
          <h1 className="text-3xl font-bold tracking-tight text-white italic uppercase">Cloud Servers</h1>
          <p className="text-[10px] text-gray-500 uppercase tracking-[0.3em] mt-1 font-black">Manage your protected Docker environments</p>
      </div>
      
      <ServerList 
        agents={agents}
        plan={plan}
        onRestore={(id, snaps) => {
          setRestoreAgentId(id);
          setRestoreSnapshots(snaps);
          setIsRestoreOpen(true);
        }}
      />

      <RestoreModal 
        isOpen={isRestoreOpen} 
        onClose={() => setIsRestoreOpen(false)} 
        agentId={restoreAgentId} 
        snapshots={restoreSnapshots}
        token={token}
      />
    </div>
  );
}

