'use client';

import { useState } from 'react';
import { ChevronRight, Folder, HardDrive, ShieldCheck, CheckSquare, Square } from 'lucide-react';

interface FileExplorerProps {
  containerName: string;
  folders: string[];
}

export default function FileExplorer({ containerName, folders }: FileExplorerProps) {
  const [selectedFolders, setSelectedFolders] = useState<string[]>([]);

  const toggleFolder = (folder: string) => {
    setSelectedFolders(prev => 
      prev.includes(folder) ? prev.filter(f => f !== folder) : [...prev, folder]
    );
  };

  const handleSave = async () => {
    const token = localStorage.getItem('dbp_sso_token');
    alert(`⚡ [DBP CLOUD] Configurado: Respaldar ${selectedFolders.length} carpetas en Wasabi S3.`);
    
    // INTEGRACIÓN REAL:
    /*
    await fetch("https://api.hwperu.com/v1/agent/config", {
      method: "POST",
      headers: { "Authorization": token, "Content-Type": "application/json" },
      body: JSON.stringify({ selected: selectedFolders, container: containerName })
    });
    */
  };

  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-lg overflow-hidden mt-4 animate-in fade-in slide-in-from-top-2 duration-300">
      <div className="bg-gray-800/50 px-4 py-2 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-gray-300">
          <HardDrive size={14} className="text-emerald-500" />
          <span className="font-mono">{containerName}</span>
          <ChevronRight size={12} className="text-gray-500" />
          <span className="text-xs text-gray-500">volumes/</span>
        </div>
        <div className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold">
          Configurable Explorer
        </div>
      </div>

      <div className="p-2 max-h-60 overflow-y-auto custom-scrollbar">
        {folders && folders.length > 0 ? (
          folders.map((folder, idx) => (
            <div 
              key={idx}
              onClick={() => toggleFolder(folder)}
              className="flex items-center justify-between p-2 hover:bg-emerald-500/10 rounded group cursor-pointer transition-colors"
            >
              <div className="flex items-center gap-3">
                <Folder size={18} className="text-emerald-400 group-hover:text-emerald-300" />
                <span className="text-sm text-gray-200">{folder}</span>
              </div>
              <div className="text-gray-500 group-hover:text-emerald-400">
                {selectedFolders.includes(folder) ? (
                  <CheckSquare size={18} className="text-emerald-500" />
                ) : (
                  <Square size={18} />
                )}
              </div>
            </div>
          ))
        ) : (
          <div className="p-8 text-center text-gray-500 text-sm">
            No subfolders discovered in volumes or no mounts found.
          </div>
        )}
      </div>

      {selectedFolders.length > 0 && (
        <div className="bg-emerald-950/30 p-3 flex items-center justify-between border-t border-emerald-900/50">
          <div className="flex items-center gap-2 text-xs text-emerald-400">
            <ShieldCheck size={14} />
            <span>{selectedFolders.length} folders selected for S3 Sync</span>
          </div>
          <button 
            onClick={handleSave}
            className="text-[10px] bg-emerald-600 hover:bg-emerald-500 text-white px-2 py-1 rounded transition-all uppercase font-bold tracking-tighter shadow-lg shadow-emerald-900/20"
          >
            Save Changes
          </button>
        </div>
      )}
    </div>
  );
}
