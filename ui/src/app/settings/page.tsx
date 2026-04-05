'use client';
import { Settings, Shield } from "lucide-react";

export default function SettingsPage() {
  return (
    <div className="max-w-7xl mx-auto p-8 space-y-12">
      <div className="flex flex-col">
          <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
             <Settings className="text-emerald-500" size={32} />
             System Settings
          </h1>
          <p className="text-xs text-gray-500 uppercase tracking-widest mt-1 font-bold">Manage your DBP Cloud Preferences</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mt-12 opacity-50 cursor-not-allowed">
         <div className="bg-gray-950 border border-gray-800 p-8 rounded-2xl shadow-2xl">
            <Shield className="text-emerald-500 mb-4" size={24} />
            <h3 className="font-bold text-white uppercase tracking-tighter">Security & Access</h3>
            <p className="text-xs text-gray-500 mt-2 leading-relaxed">
               Configure your Master Admin Token and WHMCS SSO integrations here in Phase 2.
            </p>
         </div>

         <div className="bg-gray-950 border border-gray-800 p-8 rounded-2xl shadow-2xl">
            <h3 className="font-bold text-white uppercase tracking-tighter">API Endpoint</h3>
            <p className="text-xs text-gray-500 mt-2 leading-relaxed">
               Change the central control plane URL and port settings on the fly.
            </p>
         </div>
      </div>
    </div>
  );
}
