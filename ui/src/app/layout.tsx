import { Suspense } from "react";
import "./globals.css";
import ClientLayoutWrapper from "@/components/ClientLayoutWrapper";

export const metadata = {
  title: "DBP | Dashboard",
  description: "Docker Backup Pro Control Plane",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <Suspense fallback={<body className="bg-black text-white">Loading...</body>}>
        <ClientLayoutWrapper>{children}</ClientLayoutWrapper>
      </Suspense>
    </html>
  );
}


