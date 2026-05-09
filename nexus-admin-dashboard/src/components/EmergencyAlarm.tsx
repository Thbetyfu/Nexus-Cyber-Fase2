import React, { useEffect, useState } from 'react';
import { ShieldAlert, Zap, X } from 'lucide-react';

interface EmergencyAlarmProps {
  isActive: boolean;
  onAcknowledge: () => void;
  threatDetail?: string;
}

const EmergencyAlarm: React.FC<EmergencyAlarmProps> = ({ isActive, onAcknowledge, threatDetail }) => {
  const [show, setShow] = useState(false);

  useEffect(() => {
    if (isActive) {
      setShow(true);
    }
  }, [isActive]);

  if (!isActive && !show) return null;

  return (
    <div className={`fixed inset-0 z-[9999] pointer-events-none transition-all duration-500 ${isActive ? 'opacity-100' : 'opacity-0'}`}>
      {/* Flashing Red Border */}
      <div className="absolute inset-0 border-[10px] border-red-600/30 animate-pulse pointer-events-none" />
      
      {/* Background Glow */}
      <div className="absolute inset-0 bg-red-950/10 backdrop-blur-[1px] pointer-events-none" />

      {/* Top Banner */}
      <div className="absolute top-0 left-0 right-0 h-20 bg-red-600 flex items-center justify-between px-8 shadow-[0_4px_50px_rgba(220,38,38,0.6)] pointer-events-auto animate-in slide-in-from-top duration-500 z-[10001]">
        <div className="flex items-center gap-4">
          <div className="bg-white/20 p-2 rounded animate-pulse">
            <ShieldAlert className="w-10 h-10 text-white" />
          </div>
          <div>
            <h2 className="text-white font-black text-2xl tracking-tighter uppercase leading-none">Emergency: Active Siege Detected</h2>
            <p className="text-red-100 text-xs font-bold tracking-widest uppercase mt-1">
              {threatDetail || "AI-Core has triggered Digital Hallucination Protocol."}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-6">
          <div className="flex flex-col items-end border-r border-white/20 pr-6">
            <span className="text-[10px] font-bold text-white/60 uppercase tracking-widest">Protocol Status</span>
            <span className="text-white font-mono font-black text-lg">HONEYPOT_ACTIVE</span>
          </div>
          <button 
            onClick={() => {
              setShow(false);
              onAcknowledge();
            }}
            className="bg-white text-red-600 px-6 py-3 rounded-lg font-black text-sm hover:bg-gray-100 transition shadow-[0_10px_20px_rgba(0,0,0,0.2)] flex items-center gap-2 group"
          >
            <X className="w-5 h-5 group-hover:rotate-90 transition-transform" />
            ACKNOWLEDGE & PURGE
          </button>
        </div>
      </div>

      {/* 🛡️ HONEYPOT WATERMARK (For Demos) */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none overflow-hidden">
        <div className="rotate-[-25deg] flex flex-col gap-8 opacity-[0.07]">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex gap-20 whitespace-nowrap">
              {Array.from({ length: 4 }).map((_, j) => (
                <span key={j} className="text-8xl font-black text-red-500 tracking-[1em] uppercase">
                  Digital Hallucination Active
                </span>
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Center Alert (Optional, can be distracting) */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 pointer-events-none">
        <div className="relative">
          <div className="absolute inset-0 bg-red-600/20 blur-3xl animate-ping rounded-full" />
          <Zap className="w-32 h-32 text-red-600/40 animate-pulse" />
        </div>
      </div>
    </div>
  );
};

export default EmergencyAlarm;
