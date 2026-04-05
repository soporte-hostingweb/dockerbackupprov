'use client';

import { useSearchParams } from "next/navigation";
import Sidebar from "@/components/Sidebar";

export default function ClientLayoutWrapper({
  children,
}: {
  children: React.ReactNode;
}) {
  const searchParams = useSearchParams();
  const isEmbed = searchParams.get("embed") === "1";

  return (
    <body className="flex h-screen bg-black text-white">
      {!isEmbed && <Sidebar />}
      <main className={`flex-1 overflow-y-auto bg-gray-950 ${isEmbed ? 'p-0' : 'p-8'}`}>
        {children}
      </main>
    </body>
  );
}
