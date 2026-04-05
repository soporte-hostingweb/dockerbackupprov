import React, { useEffect, useState } from "react";
import { HardDrive, Server, ShieldCheck, ShieldAlert, RefreshCcw } from "lucide-react";

export default function ServerList() {
  const [servers, setServers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch("https://api.hwperu.com/v1/agent/status");
        const data = await res.json();
        
        // Convertimos el mapa de la API a un array para el componente
        const serverList = Object.values(data).map((s: any) => ({
          id: s.agent_id,
          name: s.agent_id.split('_')[0], // Usamos el prefijo como nombre
          ip: "10.0.0.1", // IP simulada
          os: "Linux VPS",
          status: s.status === "SUCCESS" || s.status === "Healthy" ? "ok" : "error",
          lastBackup: s.last_sync || "Just now",
          size: s.total_size_mb ? `${(s.total_size_mb / 1024).toFixed(1)} GB` : "0 GB"
        }));
        
        setServers(serverList);
      } catch (error) {
        console.error("Failed to fetch DBP status:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 30000); // Auto-refresh cada 30seg
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 border border-dashed border-gray-800 rounded-lg">
         <RefreshCcw className="h-8 w-8 animate-spin text-emerald-500 mb-4" />
         <p className="text-gray-400">Consulting HWPeru Cloud Control...</p>
      </div>
    );
  }

  if (servers.length === 0) {
    return (
      <div className="text-center p-12 border border-dashed border-gray-800 rounded-lg bg-gray-950">
        <Server className="h-12 w-12 text-gray-700 mx-auto mb-4" />
        <h3 className="text-white font-semibold">No active agents found</h3>
        <p className="text-gray-500 text-sm mt-2 max-w-xs mx-auto">
          Please run the installation command in your VPS to start monitoring.
        </p>
      </div>
    );
  }

  return (
    <div className="border border-gray-800 rounded-lg overflow-hidden bg-gray-900">
      <table className="min-w-full divide-y divide-gray-800">
        <thead className="bg-black/50">
          <tr>
            <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Server</th>
            <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Health Status</th>
            <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Last Sync</th>
            <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Size</th>
            <th scope="col" className="relative px-6 py-4"><span className="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody className="bg-gray-900 divide-y divide-gray-800">
          {servers.map((server) => (
            <tr key={server.id} className="hover:bg-gray-800/60 transition duration-150">
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="flex items-center">
                  <div className="flex-shrink-0 h-10 w-10 bg-gray-800 rounded flex items-center justify-center border border-gray-700">
                    <Server className="h-5 w-5 text-gray-400" />
                  </div>
                  <div className="ml-4">
                    <div className="text-sm font-medium text-white">{server.name}</div>
                    <div className="text-sm text-gray-500">{server.ip} &middot; {server.os}</div>
                  </div>
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                {server.status === "ok" ? (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-400/10 text-emerald-400 border border-emerald-400/20">
                    <ShieldCheck className="w-3.5 h-3.5 mr-1" /> Healthy
                  </span>
                ) : (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-red-400/10 text-red-400 border border-red-400/20">
                    <ShieldAlert className="w-3.5 h-3.5 mr-1" /> Action Required
                  </span>
                )}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                {server.lastBackup}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                <div className="flex items-center">
                  <HardDrive className="w-4 h-4 mr-2 text-gray-500" />
                  {server.size}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <a href="#" className="inline-block text-emerald-500 hover:text-emerald-400 mr-5 transition">Manage Options</a>
                <a href="#" className="inline-block text-white bg-gray-800 hover:bg-gray-700 px-3 py-1.5 rounded transition">Restore Wizard</a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
