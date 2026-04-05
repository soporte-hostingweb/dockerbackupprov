'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

function LoginHandler() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const ssoToken = searchParams.get('sso');
  const embed = searchParams.get('embed');

  useEffect(() => {
    if (ssoToken) {
      // Almacenamos el token para futuras peticiones a la API
      localStorage.setItem('dbp_token', ssoToken);
      
      // Construimos la URL de redirección manteniendo el modo embed si existe
      let nextUrl = '/';
      if (embed === '1') {
        nextUrl += '?embed=1';
      }
      
      router.push(nextUrl);
    } else {
      // Si no hay token, simplemente vamos al dashboard principal
      router.push('/');
    }
  }, [ssoToken, embed, router]);

  return (
    <div className="flex h-screen items-center justify-center bg-black text-white">
      <div className="text-center">
        <div className="h-12 w-12 animate-spin rounded-full border-4 border-emerald-500 border-t-transparent mx-auto mb-4"></div>
        <h1 className="text-xl font-bold">Authenticating with DBP Cloud...</h1>
        <p className="text-gray-400 mt-2">Connecting your WHMCS secure session</p>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <LoginHandler />
    </Suspense>
  );
}
