import React from 'react';
import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';
import { Home, Server, Settings, DatabaseBackup, Clock, Shield } from 'lucide-react';


export default function Sidebar() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const isAdmin = searchParams.get("admin") === "1";
  const sso = searchParams.get("sso");

  // Mantener los parámetros en los links
  const query = `?sso=${sso || ''}${isAdmin ? '&admin=1' : ''}`;

  return (
    <aside className="w-64 bg-black border-r border-gray-800 flex flex-col">
      <div className="h-16 flex items-center px-6 border-b border-gray-800 font-bold tracking-wider text-xl mt-4 mb-4 uppercase italic">
        <DatabaseBackup className="w-6 h-6 text-emerald-500 mr-2" />
        DBP <span className="font-light text-gray-500 ml-1 tracking-tighter">Control</span>
      </div>
      
      <nav className="flex-1 py-6 px-3 space-y-1">
        <Link href={`/${query}`} className={`flex items-center px-4 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all ${pathname === '/' ? 'bg-gray-900 text-white shadow-lg border border-gray-800' : 'text-gray-500 hover:bg-gray-900 hover:text-white'}`}>
          <Home className="w-4 h-4 mr-3 text-emerald-500" />
          Dashboard
        </Link>
        <Link href={`/servers${query}`} className={`flex items-center px-4 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all ${pathname === '/servers' ? 'bg-gray-900 text-white shadow-lg border border-gray-800' : 'text-gray-500 hover:bg-gray-900 hover:text-white'}`}>
          <Server className="w-4 h-4 mr-3 text-gray-400 group-hover:text-emerald-500" />
          Servers
        </Link>
        <Link href={`/history${query}`} className={`flex items-center px-4 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all ${pathname === '/history' ? 'bg-gray-900 text-white shadow-lg border border-gray-800' : 'text-gray-500 hover:bg-gray-900 hover:text-white'}`}>
          <Clock className="w-4 h-4 mr-3 text-gray-400 group-hover:text-emerald-500" />
          History
        </Link>
        
        {isAdmin && (
          <Link href={`/settings${query}`} className={`flex items-center px-4 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all ${pathname === '/settings' ? 'bg-gray-900 text-white shadow-lg border border-gray-800' : 'text-gray-500 hover:bg-gray-900 hover:text-white'}`}>
            <Settings className="w-4 h-4 mr-3 text-gray-400 group-hover:text-emerald-500" />
            System & Master
          </Link>
        )}
      </nav>



      <div className="p-4 border-t border-gray-800">
        <button className="w-full bg-gray-800 text-gray-300 hover:text-white text-sm font-medium py-2 rounded-md hover:bg-gray-700 transition">
          Return to WHMCS
        </button>
      </div>
    </aside>
  );
}
