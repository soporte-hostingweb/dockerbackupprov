'use client';

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import "./globals.css";
import Sidebar from "@/components/Sidebar";

function LayoutContent({ children }: { children: React.ReactNode }) {
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

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <Suspense fallback={<body className="bg-black text-white">Loading...</body>}>
        <LayoutContent>{children}</LayoutContent>
      </Suspense>
    </html>
  );
}

