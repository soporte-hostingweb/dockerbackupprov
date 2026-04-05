import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, Server, Settings, DatabaseBackup, Clock } from 'lucide-react';


export default function Sidebar() {
  return (
    <aside className="w-64 bg-black border-r border-gray-800 flex flex-col">
      <div className="h-16 flex items-center px-6 border-b border-gray-800 font-bold tracking-wider text-lg mt-4 mb-4">
        <DatabaseBackup className="w-6 h-6 text-emerald-500 mr-2" />
        DBP <span className="font-light text-gray-400 ml-1">Control</span>
      </div>
      
      <nav className="flex-1 py-6 px-3 space-y-1">
        <Link href="/" className="flex items-center px-3 py-2 text-sm font-medium rounded-md bg-gray-900 text-white group">
          <Home className="w-5 h-5 mr-3 text-emerald-500" />
          Dashboard
        </Link>
        <Link href="/servers" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-300 hover:bg-gray-800 hover:text-white transition-colors group">
          <Server className="w-5 h-5 mr-3 text-gray-400 group-hover:text-emerald-500" />
          Servers
        </Link>
        <Link href="/history" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-300 hover:bg-gray-800 hover:text-white transition-colors group">
          <Clock className="w-5 h-5 mr-3 text-gray-400 group-hover:text-emerald-500" />
          History
        </Link>
        <Link href="/settings" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-300 hover:bg-gray-800 hover:text-white transition-colors group">
          <Settings className="w-5 h-5 mr-3 text-gray-400 group-hover:text-emerald-500" />
          Settings
        </Link>
      </nav>


      <div className="p-4 border-t border-gray-800">
        <button className="w-full bg-gray-800 text-gray-300 hover:text-white text-sm font-medium py-2 rounded-md hover:bg-gray-700 transition">
          Return to WHMCS
        </button>
      </div>
    </aside>
  );
}
