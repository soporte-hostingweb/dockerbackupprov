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

  // Si no ha montado, mostramos una estructura base neutra para evitar el Error #419
  // Pero mantenemos las etiquetas body y main consistentes
  return (
    <body className={`flex h-screen bg-black text-white ${!mounted ? 'opacity-0' : 'opacity-100 transition-opacity duration-300'}`}>
      {/* Solo ocultamos el Sidebar si estamos seguros de que es modo embed */}
      {mounted && !isEmbed && <Sidebar />}
      
      {/* En el servidor (mounted=false) renderizamos con el Sidebar por defecto si es necesario, 
          o simplemente dejamos el espacio. Intentemos ser lo más estables posible. */}
      {!mounted && <Sidebar />}

      <main className={`flex-1 overflow-y-auto bg-gray-950 ${isEmbed ? 'p-0' : 'p-8'}`}>
        {children}
      </main>
    </body>
  );
}


