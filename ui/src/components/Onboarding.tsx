'use client';

import React from 'react';
import { 
  ShieldCheck, 
  Database, 
  Globe, 
  Server, 
  CheckCircle2, 
  AlertCircle,
  ArrowRight,
  Zap
} from "lucide-react";

interface DetectedStack {
  wordpress: boolean;
  mysql: boolean;
  nginx: boolean;
  apache: boolean;
  node: boolean;
  pm2: boolean;
  has_docker: boolean;
}

interface OnboardingProps {
  agentId: string;
  detectedStack: DetectedStack;
  onComplete: (config: any) => void;
  onCancel: () => void;
}

export default function Onboarding({ agentId, detectedStack, onComplete, onCancel }: OnboardingProps) {
  const [step, setStep] = React.useState(1);
  const [selectedPlan, setSelectedPlan] = React.useState<'wordpress' | 'full' | 'app' | 'advanced'>('wordpress');

  React.useEffect(() => {
    if (detectedStack.wordpress) setSelectedPlan('wordpress');
    else if (detectedStack.has_docker) setSelectedPlan('app');
    else setSelectedPlan('full');
  }, [detectedStack]);

  const handleFinish = () => {
    onComplete({
        mode: selectedPlan,
        protection_level: selectedPlan === 'wordpress' ? 'Advanced' : (selectedPlan === 'full' ? 'Total' : 'Basic')
    });
  };

  return (
    <div className="fixed inset-0 z-[100] bg-black/90 backdrop-blur-xl flex items-center justify-center p-4">
      <div className="bg-gray-950 border border-gray-900 w-full max-w-2xl rounded-[3rem] overflow-hidden shadow-[0_0_100px_rgba(16,185,129,0.1)]">
        
        {/* Progress Bar */}
        <div className="h-1.5 w-full bg-gray-900 flex">
            <div className={`h-full bg-emerald-500 transition-all duration-700 ${step === 1 ? 'w-1/3' : (step === 2 ? 'w-2/3' : 'w-full')}`}></div>
        </div>

        <div className="p-10 md:p-16">
            
            {step === 1 && (
                <div className="space-y-8 animate-in fade-in slide-in-from-bottom-5 duration-500">
                    <div className="space-y-2">
                        <div className="inline-flex items-center gap-2 px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-full mb-4">
                            <Zap size={14} className="text-emerald-500" />
                            <span className="text-[10px] text-emerald-500 font-extrabold uppercase tracking-widest">Detección Inteligente</span>
                        </div>
                        <h2 className="text-4xl font-black text-white italic tracking-tighter">
                            BIENVENIDO A <span className="text-blue-500">HW CLOUD</span>
                        </h2>
                        <p className="text-gray-500 font-medium uppercase text-xs tracking-widest">Analizamos tu servidor <span className="text-white font-mono">{agentId}</span></p>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className={`p-4 rounded-2xl border transition-all ${detectedStack.wordpress ? 'bg-emerald-500/10 border-emerald-500/40' : 'bg-gray-950 border-gray-900 opacity-40'}`}>
                            <div className="flex items-center gap-3">
                                < Globe className={detectedStack.wordpress ? 'text-emerald-500' : 'text-gray-600'} />
                                <span className={`text-sm font-black uppercase italic ${detectedStack.wordpress ? 'text-white' : 'text-gray-600'}`}>WordPress</span>
                            </div>
                        </div>
                        <div className={`p-4 rounded-2xl border transition-all ${detectedStack.mysql ? 'bg-emerald-500/10 border-emerald-500/40' : 'bg-gray-950 border-gray-900 opacity-40'}`}>
                            <div className="flex items-center gap-3">
                                < Database className={detectedStack.mysql ? 'text-emerald-500' : 'text-gray-600'} />
                                <span className={`text-sm font-black uppercase italic ${detectedStack.mysql ? 'text-white' : 'text-gray-600'}`}>MySQL DB</span>
                            </div>
                        </div>
                        <div className={`p-4 rounded-2xl border transition-all ${detectedStack.has_docker ? 'bg-blue-500/10 border-blue-500/40' : 'bg-gray-950 border-gray-900 opacity-40'}`}>
                            <div className="flex items-center gap-3">
                                < ShieldCheck className={detectedStack.has_docker ? 'text-blue-500' : 'text-gray-600'} />
                                <span className={`text-sm font-black uppercase italic ${detectedStack.has_docker ? 'text-white' : 'text-gray-600'}`}>Docker Stack</span>
                            </div>
                        </div>
                        <div className={`p-4 rounded-2xl border transition-all ${!detectedStack.has_docker ? 'bg-amber-500/10 border-amber-500/40' : 'bg-gray-950 border-gray-900 opacity-40'}`}>
                            <div className="flex items-center gap-3">
                                < Server className={!detectedStack.has_docker ? 'text-amber-500' : 'text-gray-600'} />
                                <span className={`text-sm font-black uppercase italic ${!detectedStack.has_docker ? 'text-white' : 'text-gray-600'}`}>Bare-Metal</span>
                            </div>
                        </div>
                    </div>

                    <button 
                        onClick={() => setStep(2)}
                        className="w-full bg-white text-black font-black py-4 rounded-2xl flex items-center justify-center gap-3 hover:bg-emerald-500 hover:text-white transition-all shadow-xl shadow-emerald-950/20"
                    >
                        CONTINUAR <ArrowRight size={20} />
                    </button>
                </div>
            )}

            {step === 2 && (
                <div className="space-y-8 animate-in fade-in slide-in-from-right-10 duration-500">
                    <div className="space-y-1">
                        <h3 className="text-2xl font-black text-white italic tracking-tighter uppercase">¿Cómo deseas proteger tu servidor?</h3>
                        <p className="text-gray-500 text-xs font-medium uppercase tracking-widest">Selecciona el modo que mejor se adapte a tu necesidad.</p>
                    </div>

                    <div className="space-y-3">
                        <div 
                            onClick={() => setSelectedPlan('wordpress')}
                            className={`p-6 rounded-3xl border cursor-pointer transition-all ${selectedPlan === 'wordpress' ? 'bg-emerald-500/10 border-emerald-500 shadow-lg shadow-emerald-500/5' : 'bg-black/40 border-gray-900 hover:border-gray-800'}`}
                        >
                            <div className="flex items-start justify-between">
                                <div className="flex gap-4">
                                    <div className={`p-3 rounded-2xl ${selectedPlan === 'wordpress' ? 'bg-emerald-500 text-black' : 'bg-gray-900 text-gray-500'}`}><Globe size={24} /></div>
                                    <div>
                                        <p className="text-sm font-black text-white uppercase italic">Modo WordPress (Recomendado)</p>
                                        <p className="text-[10px] text-gray-500 mt-1 font-medium">Auto-protección de archivos, DB y configuración SSL.</p>
                                    </div>
                                </div>
                                {selectedPlan === 'wordpress' && <CheckCircle2 size={24} className="text-emerald-500" />}
                            </div>
                        </div>

                        <div 
                            onClick={() => setSelectedPlan('full')}
                            className={`p-6 rounded-3xl border cursor-pointer transition-all ${selectedPlan === 'full' ? 'bg-blue-500/10 border-blue-500 shadow-lg shadow-blue-500/5' : 'bg-black/40 border-gray-900 hover:border-gray-800'}`}
                        >
                            <div className="flex items-start justify-between">
                                <div className="flex gap-4">
                                    <div className={`p-3 rounded-2xl ${selectedPlan === 'full' ? 'bg-blue-500 text-white' : 'bg-gray-900 text-gray-500'}`}><Server size={24} /></div>
                                    <div>
                                        <p className="text-sm font-black text-white uppercase italic">Servidor Completo</p>
                                        <p className="text-[10px] text-gray-500 mt-1 font-medium">Snapshot total del sistema de archivos con recuperación inteligente.</p>
                                    </div>
                                </div>
                                {selectedPlan === 'full' && <CheckCircle2 size={24} className="text-blue-500" />}
                            </div>
                        </div>

                        <div 
                            onClick={() => setSelectedPlan('advanced')}
                            className={`p-6 rounded-3xl border cursor-pointer transition-all ${selectedPlan === 'advanced' ? 'bg-amber-500/10 border-amber-500 shadow-lg shadow-amber-500/5' : 'bg-black/40 border-gray-900 hover:border-gray-800'}`}
                        >
                            <div className="flex items-start justify-between">
                                <div className="flex gap-4">
                                    <div className={`p-3 rounded-2xl ${selectedPlan === 'advanced' ? 'bg-amber-500 text-black' : 'bg-gray-900 text-gray-500'}`}><Settings size={24} /></div>
                                    <div>
                                        <p className="text-sm font-black text-white uppercase italic">Configuración Avanzada</p>
                                        <p className="text-[10px] text-gray-500 mt-1 font-medium">Control total sobre rutas y exclusiones manuales.</p>
                                    </div>
                                </div>
                                {selectedPlan === 'advanced' && <CheckCircle2 size={24} className="text-amber-500" />}
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-4 mt-8">
                        <button onClick={onCancel} className="flex-1 bg-gray-900 text-gray-500 font-bold py-4 rounded-2xl hover:bg-gray-800 transition-all uppercase text-xs tracking-widest">CANCELAR</button>
                        <button onClick={handleFinish} className="flex-[2] bg-emerald-600 text-white font-black py-4 rounded-2xl hover:bg-emerald-500 transition-all shadow-xl shadow-emerald-950/20 uppercase text-xs tracking-widest">ACTIVAR PROTECCIÓN</button>
                    </div>
                </div>
            )}
        </div>
      </div>
    </div>
  );
}
