import React from "react";
import { HardDrive, Server, ShieldCheck, ShieldAlert } from "lucide-react";

// Mock para simular los datos que vendrían de la API Central de Go
const servers = [
  { id: "vps-01", name: "Production Web", ip: "192.168.1.50", os: "Ubuntu 22.04", status: "ok", lastBackup: "10 mins ago", size: "15.0 GB" },
  { id: "vps-02", name: "Database Master", ip: "192.168.1.51", os: "Debian 12", status: "ok", lastBackup: "1 hr ago", size: "28.5 GB" },
  { id: "vps-03", name: "Staging Test Area", ip: "192.168.1.52", os: "Ubuntu 22.04", status: "error", lastBackup: "Failing (3 days)", size: "1.7 GB" },
];

export default function ServerList() {
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
