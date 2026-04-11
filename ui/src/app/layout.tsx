import { Suspense } from "react";
import "./globals.css";
import ClientLayoutWrapper from "@/components/ClientLayoutWrapper";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "HW Cloud Recovery | Orchestrator",
  description: "HW Cloud Recovery Control Plane",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="flex h-screen bg-black text-white p-0 m-0 overflow-hidden">
        <Suspense fallback={<div className="bg-black text-white p-8">Loading Dashboard...</div>}>
          <ClientLayoutWrapper>{children}</ClientLayoutWrapper>
        </Suspense>
      </body>
    </html>
  );
}



