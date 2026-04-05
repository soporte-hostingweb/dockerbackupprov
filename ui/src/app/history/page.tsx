'use client';
import { Clock, History } from "lucide-react";

export default function HistoryPage() {
  return (
    <div className="max-w-7xl mx-auto p-8 space-y-12">
      <div className="flex flex-col">
          <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
             <Clock className="text-emerald-500" size={32} />
             Backup History
          </h1>
          <p className="text-xs text-gray-500 uppercase tracking-widest mt-1 font-bold">Chronological log of your container snapshots</p>
      </div>

      <div className="bg-gray-900/50 border-2 border-dashed border-gray-800 rounded-2xl p-20 text-center flex flex-col items-center justify-center">
         <History className="text-gray-700 w-16 h-16 mb-4" />
         <h3 className="text-lg font-bold text-gray-300 uppercase tracking-tighter">Activity Log Coming Soon</h3>
         <p className="text-xs text-gray-500 max-w-sm leading-relaxed uppercase mt-2">
            The Phase 1 integration focuses on Restic Engine and File Explorer. Full historical reporting is planned for Phase 2.
         </p>
      </div>
    </div>
  );
}
