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

  useEffect(() => {
    const sso = searchParams.get('sso');
    if (sso) {
      setToken(sso);
      localStorage.setItem('dbp_token', sso);
    } else {
      const stored = localStorage.getItem('dbp_token');
      if (stored) setToken(stored);
    }
  }, [searchParams]);

  return (
    <div className="max-w-7xl mx-auto p-8 space-y-8">
      <div className="flex flex-col">
          <h1 className="text-3xl font-bold tracking-tight text-white italic uppercase">Cloud Servers</h1>
          <p className="text-[10px] text-gray-500 uppercase tracking-[0.3em] mt-1 font-black">Manage your protected Docker environments</p>
      </div>
      
      <ServerList 
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

