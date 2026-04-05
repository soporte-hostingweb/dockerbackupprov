'use client';

import { useState, useEffect } from 'react';
import { ChevronRight, Folder, HardDrive, ShieldCheck, CheckSquare, Square, HelpCircle } from 'lucide-react';

interface FileExplorerProps {
  agentId: string;
  containerName: string;
  folders: string[];
}

export default function FileExplorer({ agentId, containerName, folders }: FileExplorerProps) {
  const [selectedFolders, setSelectedFolders] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);
  const [loadingConfig, setLoadingConfig] = useState(true);

  // EFECTO: Cargar configuración existente desde la DB al montar
  useEffect(() => {
    async function fetchSavedConfig() {
      const token = localStorage.getItem('dbp_sso_token');
      if (!token) return;

      try {
        const response = await fetch(`https://api.hwperu.com/v1/agent/config?agent_id=${agentId}`, {
          headers: { "Authorization": token }
        });
        if (response.ok) {
          const data = await response.json();
          if (data.paths) {
            setSelectedFolders(data.paths);
            console.log(`[CONFIG] Loaded ${data.paths.length} paths for agent ${agentId}`);
          }
        }
      } catch (err) {
        console.error("Failed to load saved config:", err);
      } finally {
        setLoadingConfig(false);
      }
    }
    fetchSavedConfig();
  }, [agentId]);

  const toggleFolder = (folderPath: string) => {
    setSelectedFolders(prev => 
      prev.includes(folderPath) ? prev.filter(f => f !== folderPath) : [...prev, folderPath]
    );
  };

  const handleSave = async () => {
    const token = localStorage.getItem('dbp_sso_token');
    if (!token) return;

    setSaving(true);
    try {
      const response = await fetch("https://api.hwperu.com/v1/agent/config", {
        method: "POST",
        headers: { "Authorization": token, "Content-Type": "application/json" },
        body: JSON.stringify({ 
          agent_id: agentId,
          paths: selectedFolders 
        })
      });

      if (response.ok) {
        alert("✅ [CONFIG] Selección de respaldo sincronizada con el Agente.");
      }
    } catch (err) {
      console.error("Failed to save selection:", err);
    } finally {
      setSaving(false);
    }
  };

  const getBasename = (path: string) => path.split('/').pop() || path;

  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-lg overflow-hidden mt-4 animate-in fade-in slide-in-from-top-2 duration-300">
      <div className="bg-gray-800/50 px-4 py-2 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-gray-300">
          <HardDrive size={14} className="text-emerald-500" />
          <span className="font-mono text-[10px]">{containerName}</span>
          <ChevronRight size={12} className="text-gray-500" />
          <span className="text-[10px] text-gray-500 uppercase">Snapshot Targets</span>
        </div>
        {loadingConfig && <span className="text-[10px] text-emerald-500 animate-pulse font-bold">LOADING PERSISTENCE...</span>}
      </div>

      <div className="p-2 max-h-60 overflow-y-auto custom-scrollbar">
        {folders && folders.length > 0 ? (
          folders.map((fullPath, idx) => {
            const isSelected = selectedFolders.some(f => f === fullPath);
            return (
              <div 
                key={idx}
                onClick={() => toggleFolder(fullPath)}
                className={`flex items-center justify-between p-2 rounded group cursor-pointer transition-colors ${isSelected ? 'bg-emerald-500/10' : 'hover:bg-emerald-500/5'}`}
              >
                <div className="flex items-center gap-3">
                  <Folder size={18} className={`${isSelected ? 'text-emerald-400' : 'text-gray-600'} group-hover:text-emerald-300`} />
                  <div className="flex flex-col">
                     <span className={`text-sm ${isSelected ? 'text-white font-bold' : 'text-gray-400'}`}>{getBasename(fullPath)}</span>
                     <span className="text-[8px] text-gray-600 font-mono italic">Host: {fullPath.replace('/host_root', '')}</span>
                  </div>
                </div>
                <div className={`${isSelected ? 'text-emerald-400' : 'text-gray-700'} group-hover:text-emerald-400`}>
                  {isSelected ? <CheckSquare size={18} /> : <Square size={18} />}
                </div>
              </div>
            );
          })
        ) : (
          <div className="p-8 text-center text-gray-500 text-sm">
            <HelpCircle className="mx-auto mb-2 text-gray-700" size={24} />
            No host volumes detected via /host_root bridge.
          </div>
        )}
      </div>

      {(selectedFolders.length > 0 || !loadingConfig) && (
        <div className="bg-emerald-950/30 p-3 flex items-center justify-between border-t border-emerald-900/50">
          <div className="flex items-center gap-2 text-xs text-emerald-400">
            <ShieldCheck size={14} />
            <span className="font-bold">{selectedFolders.length} targets ready</span>
          </div>
          <button 
            disabled={saving}
            onClick={handleSave}
            className="text-[10px] bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-all uppercase font-black tracking-widest shadow-lg shadow-emerald-950 border border-emerald-400/20"
          >
            {saving ? 'Syncing...' : 'Lock Selection'}
          </button>
        </div>
      )}
    </div>
  );
}
