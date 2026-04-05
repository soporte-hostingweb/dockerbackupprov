'use client';
import ServerList from "@/components/ServerList";

export default function ServersPage() {
  return (
    <div className="max-w-7xl mx-auto p-8 space-y-8">
      <div className="flex flex-col">
          <h1 className="text-3xl font-bold tracking-tight text-white">Cloud Servers</h1>
          <p className="text-xs text-gray-500 uppercase tracking-widest mt-1 font-bold">Manage your protected Docker environments</p>
      </div>
      <ServerList />
    </div>
  );
}
