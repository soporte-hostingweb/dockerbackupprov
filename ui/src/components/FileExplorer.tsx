'use client';

import { useState } from 'react';
import { ChevronRight, Folder, HardDrive, ShieldCheck, CheckSquare, Square, HelpCircle } from 'lucide-react';

interface FileExplorerProps {
  containerName: string;
  folders: string[];
}

export default function FileExplorer({ containerName, folders }: FileExplorerProps) {
  const [selectedFolders, setSelectedFolders] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);

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
        body: JSON.stringify({ paths: selectedFolders })
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

  // Helper para mostrar solo el nombre final de la carpeta en la UI
  const getBasename = (path: string) => path.split('/').pop() || path;

  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-lg overflow-hidden mt-4 animate-in fade-in slide-in-from-top-2 duration-300">
      <div className="bg-gray-800/50 px-4 py-2 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-gray-300">
          <HardDrive size={14} className="text-emerald-500" />
          <span className="font-mono text-[10px]">{containerName}</span>
          <ChevronRight size={12} className="text-gray-500" />
          <span className="text-[10px] text-gray-500 uppercase">Docker Volumes</span>
        </div>
        <div className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold">
          Selective Core
        </div>
      </div>

      <div className="p-2 max-h-60 overflow-y-auto custom-scrollbar">
        {folders && folders.length > 0 ? (
          folders.map((fullPath, idx) => (
            <div 
              key={idx}
              onClick={() => toggleFolder(fullPath)}
              className="flex items-center justify-between p-2 hover:bg-emerald-500/10 rounded group cursor-pointer transition-colors"
            >
              <div className="flex items-center gap-3">
                <Folder size={18} className="text-emerald-400 group-hover:text-emerald-300" />
                <div className="flex flex-col">
                   <span className="text-sm text-gray-200">{getBasename(fullPath)}</span>
                   <span className="text-[8px] text-gray-600 font-mono italic">Host: {fullPath.replace('/host_root', '')}</span>
                </div>
              </div>
              <div className="text-gray-500 group-hover:text-emerald-400">
                {selectedFolders.includes(fullPath) ? (
                  <CheckSquare size={18} className="text-emerald-500" />
                ) : (
                  <Square size={18} />
                )}
              </div>
            </div>
          ))
        ) : (
          <div className="p-8 text-center text-gray-500 text-sm">
            <HelpCircle className="mx-auto mb-2 text-gray-700" size={24} />
            No host volumes detected via /host_root bridge.
          </div>
        )}
      </div>

      {selectedFolders.length > 0 && (
        <div className="bg-emerald-950/30 p-3 flex items-center justify-between border-t border-emerald-900/50">
          <div className="flex items-center gap-2 text-xs text-emerald-400">
            <ShieldCheck size={14} />
            <span className="font-bold">{selectedFolders.length} targets ready</span>
          </div>
          <button 
            disabled={saving}
            onClick={handleSave}
            className="text-[10px] bg-emerald-600 hover:bg-emerald-500 text-white px-3 py-1.5 rounded transition-all uppercase font-black tracking-tighter shadow-lg shadow-emerald-900/40 border border-emerald-400/20"
          >
            {saving ? 'Syncing...' : 'Save Selection'}
          </button>
        </div>
      )}
    </div>
  );
}

