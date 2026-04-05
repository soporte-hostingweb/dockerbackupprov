'use client';

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import Sidebar from "@/components/Sidebar";

export default function ClientLayoutWrapper({
  children,
}: {
  children: React.ReactNode;
}) {
  const searchParams = useSearchParams();
  const [isEmbed, setIsEmbed] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setIsEmbed(searchParams.get("embed") === "1");
    setMounted(true);
  }, [searchParams]);

  if (!mounted) {
    return (
      <body className="bg-black text-white p-0 m-0">
        <main className="flex-1 bg-gray-950 p-0 m-0">
          {children}
        </main>
      </body>
    );
  }

  return (
    <body className="flex h-screen bg-black text-white">
      {!isEmbed && <Sidebar />}
      <main className={`flex-1 overflow-y-auto bg-gray-950 ${isEmbed ? 'p-0' : 'p-8'}`}>
        {children}
      </main>
    </body>
  );
}

