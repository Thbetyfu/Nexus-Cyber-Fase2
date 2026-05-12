"use client";

import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Shield, Cpu, Lock, Zap, ShieldCheck } from "lucide-react";

interface BootStep {
  id: number;
  label: string;
  status: "pending" | "loading" | "complete";
  icon: any;
}

export default function BootSequence({ onComplete }: { onComplete: () => void }) {
  const [steps, setSteps] = useState<BootStep[]>([
    { id: 1, label: "INITIATING PQC CORE (ML-KEM-768)", status: "pending", icon: Lock },
    { id: 2, label: "SYNCHRONIZING NEXUS-SOC-BRAIN AI", status: "pending", icon: Cpu },
    { id: 3, label: "CALIBRATING MTD TOPOLOGY MATRIX", status: "pending", icon: Zap },
    { id: 4, label: "ESTABLISHING QUANTUM SHIELD LAYER", status: "pending", icon: Shield },
    { id: 5, label: "AUTHENTICATING COMMANDER ACCESS", status: "pending", icon: ShieldCheck },
  ]);

  const [currentStep, setCurrentStep] = useState(0);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    if (currentStep >= steps.length) {
      const timer = setTimeout(() => onComplete(), 1000);
      return () => clearTimeout(timer);
    }

    // Update status of current step
    setSteps(prev => prev.map((s, i) => i === currentStep ? { ...s, status: "loading" } : s));

    const duration = 1200 + Math.random() * 800; // Randomize speed a bit
    const timer = setTimeout(() => {
      setSteps(prev => prev.map((s, i) => i === currentStep ? { ...s, status: "complete" } : s));
      setCurrentStep(prev => prev + 1);
      setProgress(((currentStep + 1) / steps.length) * 100);
    }, duration);

    return () => clearTimeout(timer);
  }, [currentStep, onComplete, steps.length]);

  return (
    <div className="fixed inset-0 z-[100000] bg-[#050608] flex items-center justify-center overflow-hidden font-mono">
      {/* Background Tech Elements */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_center,rgba(59,130,246,0.1)_0%,transparent_70%)]" />
        <div className="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.02)_1px,transparent_1px)] bg-[size:40px_40px]" />
      </div>

      <motion.div 
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="relative z-10 w-full max-w-lg p-8"
      >
        {/* Header */}
        <div className="flex items-center gap-4 mb-12">
          <div className="w-16 h-16 rounded-2xl bg-blue-500/10 border border-blue-500/30 flex items-center justify-center shadow-[0_0_30px_rgba(59,130,246,0.2)]">
            <Shield className="w-8 h-8 text-blue-500 animate-pulse" />
          </div>
          <div>
            <h1 className="text-2xl font-black text-white tracking-[0.3em] uppercase mb-1">Nexus OS</h1>
            <p className="text-xs text-blue-400 font-bold tracking-[0.1em] opacity-80 uppercase">Autonomous Defense System v2.9.4</p>
          </div>
        </div>

        {/* Steps */}
        <div className="space-y-4 mb-12">
          {steps.map((step, idx) => (
            <motion.div 
              key={step.id}
              initial={{ x: -20, opacity: 0 }}
              animate={{ x: 0, opacity: idx <= currentStep ? 1 : 0.3 }}
              className={`flex items-center justify-between p-4 rounded-lg border transition-all duration-500 ${
                step.status === "complete" 
                  ? "bg-blue-500/5 border-blue-500/30 text-blue-100" 
                  : step.status === "loading"
                    ? "bg-white/5 border-white/20 text-white"
                    : "bg-transparent border-transparent text-gray-600"
              }`}
            >
              <div className="flex items-center gap-4">
                <step.icon className={`w-4 h-4 ${step.status === "complete" ? "text-blue-400" : step.status === "loading" ? "text-white animate-spin" : "text-gray-600"}`} />
                <span className="text-[10px] font-black tracking-widest uppercase">{step.label}</span>
              </div>
              <div className="text-[10px] font-black uppercase tracking-widest">
                {step.status === "complete" && <span className="text-blue-400">SUCCESS</span>}
                {step.status === "loading" && <span className="animate-pulse">LOADING</span>}
                {step.status === "pending" && <span>PENDING</span>}
              </div>
            </motion.div>
          ))}
        </div>

        {/* Progress Bar */}
        <div className="relative h-1.5 w-full bg-white/5 rounded-full overflow-hidden mb-4">
          <motion.div 
            className="absolute top-0 left-0 h-full bg-blue-500 shadow-[0_0_15px_rgba(59,130,246,0.8)]"
            initial={{ width: "0%" }}
            animate={{ width: `${progress}%` }}
            transition={{ ease: "easeInOut", duration: 0.5 }}
          />
        </div>
        <div className="flex justify-between items-center text-[10px] text-gray-500 font-bold uppercase tracking-widest">
          <span>Bootstrapping System Modules</span>
          <span className="text-blue-400">{Math.round(progress)}%</span>
        </div>
      </motion.div>

      {/* Glitch Overlay (Subtle) */}
      <div className="absolute inset-0 pointer-events-none bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-[0.03] mix-blend-overlay" />
    </div>
  );
}
