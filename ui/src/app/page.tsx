'use client';

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { AlertCircle, Terminal, HelpCircle, ShieldCheck } from "lucide-react";
import ServerList from "@/components/ServerList";


export default function DashboardPage() {
  const searchParams = useSearchParams();
  const isEmbed = searchParams.get("embed") === "1";
  const [agentCount, setAgentCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Capturamos el token de la URL si viene de WHMCS (SSO)
    const token = searchParams.get("sso");
    if (token) {
      console.log("[SSO] New token detected in URL, updating storage...");
      localStorage.setItem("dbp_sso_token", token);
      // Forzamos un refresco de estado si el token ha cambiado
      window.dispatchEvent(new Event("storage"));
    }
  }, [searchParams]);


  useEffect(() => {
    async function fetchStats() {
      const token = localStorage.getItem("dbp_sso_token");
      if (!token) return;

      try {
        const response = await fetch("https://api.hwperu.com/v1/agent/status", {
          headers: { "Authorization": token }
        });
        if (response.ok) {
          const data = await response.json();
          setAgentCount(Object.keys(data).length);
        } else if (response.status === 401) {
          console.warn("[AUTH] Token rejected (401). Purging invalid storage...");
          localStorage.removeItem("dbp_sso_token");
          setAgentCount(0);
        }
      } catch (error) {
        console.error("Error fetching stats:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchStats();
  }, [searchParams]);


  const isDebug = searchParams.get("debug") === "1";
  const currentToken = typeof window !== 'undefined' ? localStorage.getItem("dbp_sso_token") : null;



  return (
    <div className={`mx-auto space-y-8 ${isEmbed ? 'max-w-full p-2' : 'max-w-7xl p-8'}`}>
      <div className="flex justify-between items-center bg-gray-950/50 p-6 rounded-2xl border border-gray-900 shadow-2xl">
        <div className="flex flex-col">
          <h1 className={`${isEmbed ? 'text-2xl' : 'text-3xl'} font-bold tracking-tight text-white`}>
            {searchParams.get("admin") === "1" ? "Master Control Panel" : "System Overview"}
          </h1>
          <p className="text-xs text-gray-500 uppercase tracking-widest mt-1 font-bold">
            {searchParams.get("admin") === "1" ? "HWPERU GLOBAL INFRASTRUCTURE" : "PROTECTED CLIENT VPS"}
          </p>
        </div>

        <div className="flex gap-2">
          <span className="px-3 py-1 bg-emerald-500/10 text-emerald-500 text-[10px] font-black rounded-full border border-emerald-500/20 uppercase tracking-widest">
            CONTROL PLANE LIVE
          </span>
        </div>
      </div>


      {/* DIAGNOSTIC PANEL (Only if debug=1) */}
      {isDebug && (
        <div className="bg-amber-950/20 border border-amber-900/50 rounded-xl p-6 animate-in slide-in-from-top-4 duration-500">
          <div className="flex items-center gap-3 mb-4">
             <Terminal className="text-amber-500" size={20} />
             <h3 className="text-sm font-black uppercase tracking-tighter text-amber-500">Diagnostic & Debugging Console</h3>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-4">
               <div>
                 <p className="text-[10px] text-gray-500 uppercase font-bold">API endpoint detected</p>
                 <p className="text-xs font-mono text-gray-300">https://api.hwperu.com/v1/agent/status</p>
               </div>
               <div>
                 <p className="text-[10px] text-gray-500 uppercase font-bold">Active Session Token</p>
                 <p className="text-xs font-mono text-amber-200 break-all bg-black/40 p-2 rounded border border-amber-900/30">
                    {currentToken || "MISSING / NOT DETECTED"}
                 </p>
               </div>
            </div>

            <div className="bg-black/40 p-4 rounded-lg border border-amber-900/20">
               <div className="flex gap-2 items-start">
                  <AlertCircle className="text-amber-500 shrink-0" size={16} />
                  <div className="text-[11px] text-gray-400 leading-relaxed">
                     <p className="font-bold text-amber-500 mb-1 uppercase">¿Ves 0 Agentes?</p>
                     <p>1. Verifica que el token del instalador en tu VPS coincida con el de arriba.</p>
                     <p>2. El agente necesita hasta 60s para reportar el primer latido.</p>
                     <p>3. Asegúrate de que el firewall de tu VPS permita salida al puerto 80/443.</p>
                  </div>
               </div>
            </div>
          </div>
        </div>
      )}

      {/* Metrics Row */}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl shadow-sm">
          <p className="text-sm text-gray-400 font-medium font-mono uppercase tracking-widest">Total Storage</p>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-3xl font-bold text-white">0.0</span>
            <span className="text-gray-400">GB / 50 GB</span>
          </div>
          <div className="w-full bg-gray-800 h-1.5 mt-4 rounded-full overflow-hidden">
            <div className="bg-emerald-500 h-full w-[2%]"></div>
          </div>
        </div>

        <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl shadow-sm border-l-4 border-l-emerald-500">
          <p className="text-sm text-gray-400 font-medium font-mono uppercase tracking-widest">Active Agents</p>
          <div className="mt-2 text-3xl font-bold text-white">
            {loading ? "..." : agentCount}
          </div>
          <p className="text-xs text-gray-500 mt-2">Connecting from your protected VPS servers</p>
        </div>

        <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl shadow-sm">
          <p className="text-sm text-gray-400 font-medium font-mono uppercase tracking-widest">Global Status</p>
          <div className="mt-2 text-3xl font-bold text-white">HEALTHY</div>
          <p className="text-xs text-emerald-500 mt-2">All API systems operational</p>
        </div>
      </div>

      {/* Server List view */}
      <div className="pt-4">
        <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
          <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></span>
          Protected Environments
        </h2>
        <ServerList />
      </div>
    </div>
  );
}

