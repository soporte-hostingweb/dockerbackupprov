import ServerList from "@/components/ServerList";

export default function DashboardPage() {
  return (
    <div className="max-w-6xl mx-auto space-y-8">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
        <button className="bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-md font-medium transition-colors">
          Add New VPS
        </button>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-gray-900 border border-gray-800 p-6 rounded-lg shadow-sm">
          <p className="text-sm text-gray-400 font-medium">Total Storage Used</p>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-3xl font-bold text-white">45.2</span>
            <span className="text-gray-400">GB / 100 GB</span>
          </div>
          <div className="w-full bg-gray-800 h-2 mt-4 rounded-full overflow-hidden">
            <div className="bg-emerald-500 h-full w-[45%]"></div>
          </div>
        </div>

        <div className="bg-gray-900 border border-gray-800 p-6 rounded-lg shadow-sm">
          <p className="text-sm text-gray-400 font-medium">Active Agents</p>
          <div className="mt-2 text-3xl font-bold text-emerald-400">3</div>
          <p className="text-xs text-gray-400 mt-2">All systems responding correctly to Heartbeats</p>
        </div>

        <div className="bg-gray-900 border border-gray-800 p-6 rounded-lg shadow-sm">
          <p className="text-sm text-gray-400 font-medium">Recent Snapshots</p>
          <div className="mt-2 text-3xl font-bold text-white">128</div>
          <p className="text-xs text-gray-400 mt-2">Safe snapshots across all VPS servers</p>
        </div>
      </div>

      {/* Server List view */}
      <div>
        <h2 className="text-xl font-semibold mb-4 text-white">Protected Servers</h2>
        <ServerList />
      </div>
    </div>
  );
}
