import "./globals.css";
import Sidebar from "@/components/Sidebar";

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
      <body className="flex h-screen bg-black text-white">
        <Sidebar />
        <main className="flex-1 overflow-y-auto bg-gray-950 p-8">
          {children}
        </main>
      </body>
    </html>
  );
}
