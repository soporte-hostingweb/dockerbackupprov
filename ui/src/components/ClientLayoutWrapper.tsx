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
  const [mounted, setMounted] = useState(false);
  const [isEmbed, setIsEmbed] = useState(false);

  useEffect(() => {
    setMounted(true);
    setIsEmbed(searchParams.get("embed") === "1");
  }, [searchParams]);

  return (
    <>
      <div className={isEmbed ? "no-sidebar" : ""}>
        <Sidebar />
      </div>
      <main className={`flex-1 overflow-y-auto bg-gray-950 ${isEmbed ? 'p-0' : 'p-8'}`}>
        {children}
      </main>
    </>
  );
}



