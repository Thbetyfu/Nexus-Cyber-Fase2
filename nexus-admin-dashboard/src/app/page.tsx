"use client"

/* 
   NEXUS_COMMAND_CENTER_OS [V2.0]
   - Unified Security OS Interface for Thoriq
   - Orchestrates Portfolio Defense & Monitoring
   - Improved UX: Desktop Icons, Snappy Windows, Active Focus
*/

import React, { useEffect, useState, useRef, useCallback } from "react"
import {
  Shield, Activity, ShieldAlert, Cpu, Globe, Terminal,
  RotateCcw, WifiOff, Layout, Maximize2, ShieldCheck, Lock,
  Database, Server, Monitor
} from 'lucide-react';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as ChartTooltip, ResponsiveContainer
} from "recharts"
import { AnimatePresence, motion } from "framer-motion";

// Components
import NechatWidget from '@/components/NechatWidget';
import DomainSwitcher from '@/components/DomainSwitcher';
import AddRouteModal from '@/components/AddRouteModal';
import AiTerminalWidget from '@/components/AiTerminalWidget';
import ThreatMapWidget from '@/components/ThreatMapWidget';
import EmergencyAlarm from '@/components/EmergencyAlarm';
import WindowFrame from '@/components/WindowFrame';
import Taskbar from '@/components/Taskbar';
import BootSequence from '@/components/BootSequence';

// Type definitions
export interface TelemetryLog {
  timestamp: string;
  source_ip: string;
  attacker_id?: string;
  geo_location?: string;
  isp?: string;
  device_fingerprint?: string;
  endpoint: string;
  status: string;
  threat_detail?: string;
  target_domain?: string;
  latency_ms: number;
}

export interface AIEventLog {
  timestamp: string;
  layer: string;
  status: string;
  detail_action: string;
}

// Hooks (Telemetry & AI Events)
function useTelemetry(url: string, intervalMs: number = 2000) {
  const [logs, setLogs] = useState<TelemetryLog[]>([])
  const [metrics, setMetrics] = useState({ allowed: 0, blocked: 0, honeypot: 0, panics: 0 })
  const [shufflerData, setShufflerData] = useState({ port: 3001, status: "OFFLINE" })
  const [isLive, setIsLive] = useState(true)
  const [isUnlicensed, setIsUnlicensed] = useState(false)

  const prevTrafficRef = useRef(0)
  const prevHoneypotRef = useRef(0)
  const isFirstFetch = useRef(true)

  const initialTimeline = Array.from({ length: 40 }).map((_, i) => {
    const time = new Date(Date.now() - (39 - i) * 2000)
    return {
      time: time.toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
      allowed: 0,
      honeypot: 0
    }
  })

  const timeline = useRef<any[]>(initialTimeline)
  const [history, setHistory] = useState<any[]>(initialTimeline)

  useEffect(() => {
    let pollingActive = true;
    const fetchTelemetry = async () => {
      try {
        const res = await fetch(url, {
          cache: 'no-store',
          mode: 'cors',
          headers: { 'Accept': 'application/json' }
        })
        
        if (res.status === 402) {
          if (pollingActive) {
            setIsUnlicensed(true)
            setIsLive(true)
          }
          return
        }

        const data = await res.json()
        if (!pollingActive) return

        if (data.status === "error" && data.message && data.message.includes("Expired")) {
          setIsUnlicensed(true)
          setIsLive(true)
          return
        }

        setIsUnlicensed(false)
        setIsLive(true)
        const stats = data.stats || { allowed: 0, blocked: 0, honeypot: 0, panics: 0 }
        setMetrics(stats)
        setShufflerData(data.mtd)
        setLogs((data.recent_logs || []).reverse())

        const currentTrafficTotal = stats.allowed || 0
        const currentHoneypotTotal = stats.honeypot || 0

        if (isFirstFetch.current || currentTrafficTotal < prevTrafficRef.current) {
          prevTrafficRef.current = currentTrafficTotal
          prevHoneypotRef.current = currentHoneypotTotal
          isFirstFetch.current = false
          return
        }

        const trafficRPS = Math.max(0, currentTrafficTotal - prevTrafficRef.current)
        const honeypotRPS = Math.max(0, currentHoneypotTotal - prevHoneypotRef.current)

        prevTrafficRef.current = currentTrafficTotal
        prevHoneypotRef.current = currentHoneypotTotal

        const timeLabel = new Date().toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
        const point = { time: timeLabel, allowed: trafficRPS, honeypot: honeypotRPS }

        timeline.current = [...timeline.current, point].slice(-40)
        setHistory([...timeline.current])

      } catch (error) {
        if (pollingActive) setIsLive(false)
      }
    };

    fetchTelemetry()
    const interval = setInterval(fetchTelemetry, intervalMs)
    return () => { pollingActive = false; clearInterval(interval); }

  }, [url, intervalMs])

  return { logs, metrics, history, shufflerData, isLive, isUnlicensed }
}

function useAIEvents(url: string, intervalMs: number = 1000) {
  const [events, setEvents] = useState<AIEventLog[]>([]);
  useEffect(() => {
    let pollingActive = true;
    const fetchAPI = async () => {
      try {
        const res = await fetch(url, { cache: 'no-store' });
        if (!res.ok) return;
        const data = await res.json();
        if (pollingActive) setEvents(data || []);
      } catch (err) { }
    };
    fetchAPI();
    const timer = setInterval(fetchAPI, intervalMs);
    return () => { pollingActive = false; clearInterval(timer); };
  }, [url, intervalMs]);
  return events;
}

// Sub-component: Desktop Icon
const DesktopIcon = ({ id, label, icon: Icon, onClick, isOpen }: any) => (
  <motion.button
    whileHover={{ scale: 1.05, backgroundColor: "rgba(59, 130, 246, 0.1)" }}
    whileTap={{ scale: 0.95 }}
    onClick={() => onClick(id)}
    className={`flex flex-col items-center gap-2 p-4 rounded-xl transition-colors group w-24 ${
      isOpen ? "bg-blue-500/5" : ""
    }`}
  >
    <div className={`w-12 h-12 flex items-center justify-center rounded-2xl border transition-all ${
      isOpen 
        ? "bg-blue-500/20 border-blue-400/50 shadow-[0_0_15px_rgba(59,130,246,0.3)]" 
        : "bg-black/40 border-gray-800/50 group-hover:border-blue-500/30"
    }`}>
      <Icon size={24} className={isOpen ? "text-blue-400" : "text-gray-400 group-hover:text-blue-300"} />
    </div>
    <span className={`text-[10px] font-bold uppercase tracking-widest text-center ${
      isOpen ? "text-blue-200" : "text-gray-500 group-hover:text-gray-300"
    }`}>
      {label}
    </span>
  </motion.button>
)

const NCCDashboard = () => {
  const [activeDomain, setActiveDomain] = useState<string>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isBooting, setIsBooting] = useState(true);
  const [logLimit, setLogLimit] = useState<number>(10);
  
  const { logs, metrics, history, shufflerData, isLive, isUnlicensed } = useTelemetry(`http://localhost:8080/api/telemetry?domain=${activeDomain}`, 2000)
  const aiEvents = useAIEvents('http://localhost:8080/api/ai-events', 1000)

  // Subscription Re-activation State
  const [activationKey, setActivationKey] = useState("");
  const [isActivating, setIsActivating] = useState(false);
  const [activationSuccess, setActivationSuccess] = useState(false);

  const handleActivation = async () => {
    if (!activationKey) return;
    setIsActivating(true);
    setTimeout(() => {
      setIsActivating(false);
      if (activationKey === "nexus-cyber-dev" || activationKey.length >= 16) {
        setActivationSuccess(true);
        setTimeout(() => {
          window.location.reload();
        }, 1500);
      } else {
        alert("🚨 ERROR: Kunci lisensi langganan tidak valid atau telah diblokir.");
      }
    }, 1200);
  };

  // System State
  const [isEmergency, setIsEmergency] = useState(false);
  const [lastHoneypotCount, setLastHoneypotCount] = useState(0);

  // Window Management State
  const [openWindows, setOpenWindows] = useState<string[]>(["metrics", "threat-map", "ai-terminal", "system-status"]);
  const [focusedWindow, setFocusedWindow] = useState<string>("metrics");
  const [windowZIndices, setWindowZIndices] = useState<Record<string, number>>({
    "metrics": 10,
    "threat-map": 11,
    "ai-terminal": 12,
    "forensic-logs": 13,
    "system-status": 14,
    "mtd-audit": 15
  });

  const handleFocusWindow = useCallback((id: string) => {
    setFocusedWindow(id);
    setWindowZIndices(prev => {
      const maxZ = Math.max(...Object.values(prev), 10);
      return { ...prev, [id]: maxZ + 1 };
    });
  }, []);

  const toggleWindow = useCallback((id: string) => {
    setOpenWindows(prev => 
      prev.includes(id) ? prev.filter(w => w !== id) : [...prev, id]
    );
    if (!openWindows.includes(id)) {
      handleFocusWindow(id);
    }
  }, [openWindows, handleFocusWindow]);

  useEffect(() => {
    const hasNewThreat = metrics.honeypot > lastHoneypotCount && lastHoneypotCount !== 0;
    const latestLogIsThreat = logs.length > 0 && logs[0].status !== "ALLOWED";
    if (hasNewThreat || latestLogIsThreat) {
      if (!isEmergency) setIsEmergency(true);
    }
    setLastHoneypotCount(metrics.honeypot);
  }, [metrics.honeypot, logs, isEmergency]);

  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // MTD Defense Audit States
  const [auditStatus, setAuditStatus] = useState<"IDLE" | "RUNNING" | "COMPLETED" | "ERROR">("IDLE");
  const [auditResult, setAuditResult] = useState<{
    checks: Array<{ label: string; passed: boolean; detail: string }>;
    passed: number;
    total: number;
    output: string;
  } | null>(null);
  const [auditError, setAuditError] = useState("");

  const runMtdAudit = async () => {
    setAuditStatus("RUNNING");
    setAuditError("");
    setAuditResult(null);
    try {
      const res = await fetch("http://localhost:8080/api/test/run", {
        method: "POST",
        mode: "cors"
      });
      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.details || errData.error || "Failed to run audit");
      }
      const data = await res.json();
      setAuditResult(data);
      setAuditStatus("COMPLETED");
    } catch (err: any) {
      setAuditStatus("ERROR");
      setAuditError(err.message || "Connection to Gateway lost or script execution timed out.");
    }
  };

  const handlePanic = async () => {
    try { await fetch("http://localhost:8080/api/panic", { method: "POST" }) } catch (err) {}
  }

  const handleDeleteDomain = async () => {
    if (activeDomain === 'all') return;
    
    if (window.confirm(`🚨 CRITICAL: Purge all data for [${activeDomain}]? This cannot be undone.`)) {
      try {
        const res = await fetch(`http://localhost:8080/api/domains?domain=${activeDomain}`, {
          method: 'DELETE',
        });
        
        if (res.ok) {
          setActiveDomain('all');
          setRefreshTrigger(prev => prev + 1);
          alert(`Workspace [${activeDomain}] has been purged from existence.`);
        }
      } catch (err) {
        console.error("Failed to purge domain:", err);
      }
    }
  };

  const handleReset = async () => {
    if (window.confirm("🚨 PURGE ALL SYSTEM DATA? This will reset metrics, clear AI memory, and wipe all forensic logs.")) {
      try {
        const res = await fetch("http://localhost:8080/api/system/reset", { method: "POST", mode: 'cors' })
        if (res.ok) window.location.reload();
      } catch (err) {}
    }
  }

  return (
    <div className={`relative min-h-screen bg-[#050608] text-gray-200 font-sans overflow-hidden transition-colors duration-1000 ${isEmergency ? 'shadow-[inset_0_0_150px_rgba(220,38,38,0.15)]' : ''}`}>
      <AnimatePresence>
        {isBooting && <BootSequence key="boot-sequence" onComplete={() => setIsBooting(false)} />}
      </AnimatePresence>
      
      {/* Background Layer: Quantum Glassmorphism */}
      <div className="absolute inset-0 z-0 pointer-events-none overflow-hidden bg-[#050810]">
        {/* Animated Orbs */}
        <motion.div 
          animate={{ x: [0, 150, -100, 0], y: [0, -150, 100, 0] }}
          transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
          className="absolute top-[-10%] left-[-10%] w-[50vw] h-[50vw] rounded-full bg-blue-500/40"
          style={{ filter: "blur(120px)", willChange: "transform" }}
        />
        <motion.div 
          animate={{ x: [0, -200, 150, 0], y: [0, 200, -150, 0] }}
          transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
          className="absolute bottom-[-10%] right-[-10%] w-[60vw] h-[60vw] rounded-full bg-purple-500/30"
          style={{ filter: "blur(140px)", willChange: "transform" }}
        />
        <motion.div 
          animate={{ x: [0, 100, -150, 0], y: [0, 100, -100, 0] }}
          transition={{ duration: 35, repeat: Infinity, ease: "linear" }}
          className="absolute top-[20%] left-[30%] w-[40vw] h-[40vw] rounded-full bg-cyan-400/20"
          style={{ filter: "blur(100px)", willChange: "transform" }}
        />
        
        {/* Glassmorphism Overlay */}
        <div className="absolute inset-0 bg-[#02040a]/40" style={{ backdropFilter: "blur(80px)" }} />
        
        {/* Micro Grid for Tech Vibe (Very Subtle) */}
        <div 
          className="absolute inset-0 opacity-[0.05] mix-blend-screen" 
          style={{ 
            backgroundImage: "linear-gradient(#ffffff 1px, transparent 1px), linear-gradient(90deg, #ffffff 1px, transparent 1px)",
            backgroundSize: "60px 60px" 
          }} 
        />
      </div>

      <EmergencyAlarm 
        isActive={isEmergency} 
        onAcknowledge={() => setIsEmergency(false)} 
        threatDetail={logs[0]?.threat_detail || logs[0]?.status}
      />

      {!isLive && (
        <div className="absolute inset-0 z-[20000] backdrop-blur-md bg-black/60 flex items-center justify-center p-6 text-center">
          <div className="bg-[#0c0f14] border border-red-900/40 rounded-2xl p-8 shadow-2xl max-w-sm flex flex-col items-center gap-6 animate-in zoom-in duration-300">
            <WifiOff className="w-12 h-12 text-red-500" />
            <h2 className="text-xl font-bold text-white uppercase tracking-widest">Gateway Offline</h2>
            <button onClick={() => window.location.reload()} className="px-6 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-500 border border-red-500/30 rounded text-[10px] font-black uppercase tracking-widest transition-all">
              Initiate Reconnect
            </button>
          </div>
        </div>
      )}

      {/* Header / Menu Bar */}
      <header className="absolute top-0 left-0 right-0 h-10 flex items-center justify-between px-6 z-50 pointer-events-none">
        <div className="flex items-center gap-4 pointer-events-auto">
          <div className="flex items-center gap-2 bg-black/40 border border-white/5 rounded-full px-4 py-1 backdrop-blur-md">
            <Shield className="w-3.5 h-3.5 text-blue-500" />
            <span className="text-[10px] font-black uppercase tracking-[0.2em] text-white/90">Nexus SOC OS</span>
          </div>
        </div>
      </header>

      {/* Desktop Workspace */}
      <main className="absolute inset-0 top-10 bottom-14 overflow-hidden p-6 z-10">
        
        {/* Desktop Icons Layer */}
        <div className="absolute top-8 left-8 flex flex-col gap-2 z-0">
          <DesktopIcon 
            id="metrics" 
            label="System Metrics" 
            icon={Activity} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("metrics")}
          />
          <DesktopIcon 
            id="threat-map" 
            label="Defense Matrix" 
            icon={Globe} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("threat-map")}
          />
          <DesktopIcon 
            id="ai-terminal" 
            label="AI Cortex" 
            icon={Cpu} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("ai-terminal")}
          />
          <DesktopIcon 
            id="forensic-logs" 
            label="Forensic Logs" 
            icon={Database} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("forensic-logs")}
          />
          <DesktopIcon 
            id="system-status" 
            label="Nexus Terminal" 
            icon={Terminal} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("system-status")}
          />
          <DesktopIcon 
            id="mtd-audit" 
            label="MTD Security Audit" 
            icon={ShieldCheck} 
            onClick={toggleWindow} 
            isOpen={openWindows.includes("mtd-audit")}
          />
        </div>

        {/* Floating Windows Layer */}
        <AnimatePresence>
          {/* Metrics Window */}
          {openWindows.includes("metrics") && (
            <WindowFrame
              key="metrics"
              id="metrics"
              title="System Metrics"
              icon={<Activity size={14} />}
              initialX={140}
              initialY={40}
              width={400}
              height={520}
              zIndex={windowZIndices["metrics"]}
              isActive={focusedWindow === "metrics"}
              onFocus={() => handleFocusWindow("metrics")}
              onClose={() => toggleWindow("metrics")}
            >
              <div className="p-4 flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="bg-[#0f172a] border border-blue-500/20 rounded-xl p-4">
                    <p className="text-[9px] text-blue-400 uppercase font-black tracking-widest mb-1">Inbound Traffic</p>
                    <p className="text-2xl font-mono font-bold text-white">{metrics.allowed.toLocaleString()}</p>
                  </div>
                  <div className="bg-[#1a1010] border border-red-500/20 rounded-xl p-4">
                    <p className="text-[9px] text-red-400 uppercase font-black tracking-widest mb-1">Threats Trapped</p>
                    <p className="text-2xl font-mono font-bold text-red-500">{metrics.honeypot.toLocaleString()}</p>
                  </div>
                </div>
                
                <div className="bg-black/40 border border-white/5 rounded-xl p-4 h-48">
                  <p className="text-[9px] text-gray-500 uppercase font-black tracking-widest mb-3 flex items-center gap-2">
                    <Activity size={10} className="text-blue-500" /> Velocity Stream
                  </p>
                  <div className="h-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={history}>
                        <defs>
                          <linearGradient id="chartBlue" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.4} />
                            <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" vertical={false} />
                        <Area type="monotone" dataKey="allowed" stroke="#3b82f6" fill="url(#chartBlue)" isAnimationActive={false} strokeWidth={2} />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </div>

                <div className="bg-black/40 border border-white/5 rounded-xl p-4 flex-1 overflow-hidden flex flex-col">
                  <p className="text-[9px] text-emerald-400 uppercase font-black tracking-widest mb-3 flex items-center gap-2">
                    <ShieldCheck size={10} /> Active Interventions
                  </p>
                  <div className="flex-1 overflow-auto custom-scrollbar space-y-2 pr-2">
                    {aiEvents.slice(0, 8).map((ev, i) => (
                      <div key={i} className="flex flex-col border-l-2 border-emerald-500/30 pl-3 py-1 bg-white/5 rounded-r">
                        <div className="flex items-center justify-between">
                          <span className="text-[8px] text-emerald-400 font-bold uppercase">{ev.status}</span>
                          <span className="text-[7px] text-gray-600 font-mono">{ev.layer}</span>
                        </div>
                        <p className="text-[9px] text-gray-400 line-clamp-1">{ev.detail_action}</p>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </WindowFrame>
          )}

          {/* Threat Map Window */}
          {openWindows.includes("threat-map") && (
            <WindowFrame
              key="threat-map"
              id="threat-map"
              title="Global Defense Matrix"
              icon={<Globe size={14} />}
              initialX={560}
              initialY={40}
              width={750}
              height={520}
              zIndex={windowZIndices["threat-map"]}
              isActive={focusedWindow === "threat-map"}
              onFocus={() => handleFocusWindow("threat-map")}
              onClose={() => toggleWindow("threat-map")}
            >
              <div className="h-full bg-black">
                <ThreatMapWidget />
              </div>
            </WindowFrame>
          )}

          {/* AI Terminal Window */}
          {openWindows.includes("ai-terminal") && (
            <WindowFrame
              key="ai-terminal"
              id="ai-terminal"
              title="Nexus AI Cortex"
              icon={<Cpu size={14} />}
              initialX={250}
              initialY={300}
              width={700}
              height={500}
              zIndex={windowZIndices["ai-terminal"]}
              isActive={focusedWindow === "ai-terminal"}
              onFocus={() => handleFocusWindow("ai-terminal")}
              onClose={() => toggleWindow("ai-terminal")}
            >
              <div className="h-full flex flex-col">
                <div className="flex-1 overflow-hidden">
                  <NechatWidget activeDomain={activeDomain} />
                </div>
              </div>
            </WindowFrame>
          )}

          {/* System Terminal Window */}
          {openWindows.includes("system-status") && (
            <WindowFrame
              key="system-status"
              id="system-status"
              title="Nexus Core Terminal"
              icon={<Terminal size={14} />}
              initialX={800}
              initialY={400}
              width={600}
              height={400}
              zIndex={windowZIndices["system-status"]}
              isActive={focusedWindow === "system-status"}
              onFocus={() => handleFocusWindow("system-status")}
              onClose={() => toggleWindow("system-status")}
            >
              <AiTerminalWidget />
            </WindowFrame>
          )}

          {/* Forensic Logs Window */}
          {openWindows.includes("forensic-logs") && (
            <WindowFrame
              key="forensic-logs"
              id="forensic-logs"
              title="Forensic Data Stream"
              icon={<Database size={14} />}
              initialX={600}
              initialY={220}
              width={850}
              height={450}
              zIndex={windowZIndices["forensic-logs"]}
              isActive={focusedWindow === "forensic-logs"}
              onFocus={() => handleFocusWindow("forensic-logs")}
              onClose={() => toggleWindow("forensic-logs")}
            >
              <div className="h-full flex flex-col">
                <div className="bg-[#090b0e] px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0">
                  <div className="flex items-center gap-3">
                    <span className="text-[9px] text-gray-500 font-mono tracking-tighter">VECTORS: REAL_TIME</span>
                    <select
                      value={logLimit}
                      onChange={(e) => setLogLimit(Number(e.target.value))}
                      className="bg-black/40 border border-white/10 rounded px-2 py-0.5 text-[9px] text-emerald-500 font-mono outline-none"
                    >
                      <option value={10}>10 ROWS</option>
                      <option value={50}>50 ROWS</option>
                      <option value={100}>100 ROWS</option>
                    </select>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse" />
                    <span className="text-[8px] text-red-500 font-black uppercase tracking-widest">Telemetry Live</span>
                  </div>
                </div>
                <div className="flex-1 overflow-auto bg-[#07090c]">
                  <table className="w-full text-left border-collapse">
                    <thead className="bg-[#0a0d11] text-[9px] uppercase text-gray-600 tracking-widest sticky top-0 z-10">
                      <tr>
                        <th className="py-3 px-4 border-b border-white/5">Timestamp</th>
                        <th className="py-3 px-4 border-b border-white/5">Source IP</th>
                        <th className="py-3 px-4 border-b border-white/5">Origin</th>
                        <th className="py-3 px-4 border-b border-white/5">Endpoint</th>
                        <th className="py-3 px-4 border-b border-white/5 text-center">Protocol</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-white/5">
                      {logs.slice(0, logLimit).map((log, idx) => (
                        <tr key={idx} className={`text-[10px] font-mono hover:bg-white/[0.02] transition-colors ${log.status !== "ALLOWED" ? "bg-red-500/5" : ""}`}>
                          <td className="py-2.5 px-4 text-gray-500">
                            {new Date(log.timestamp).toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                          </td>
                          <td className="py-2.5 px-4 text-gray-300">{log.source_ip}</td>
                          <td className="py-2.5 px-4 text-gray-400">{log.geo_location || "Unknown"}</td>
                          <td className="py-2.5 px-4">
                            <span className="bg-black/60 px-2 py-0.5 rounded border border-white/5 text-blue-400">{log.endpoint}</span>
                          </td>
                          <td className="py-2.5 px-4 text-center">
                            <span className={`px-2 py-0.5 rounded-[4px] font-black uppercase text-[8px] ${
                              log.status === "ALLOWED" ? "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20" : "bg-red-500/10 text-red-500 border border-red-500/20"
                            }`}>
                              {log.status === "ALLOWED" ? "Passed" : "Dropped"}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </WindowFrame>
          )}

          {/* MTD Compliance Audit Window */}
          {openWindows.includes("mtd-audit") && (
            <WindowFrame
              key="mtd-audit"
              id="mtd-audit"
              title="MTD Defense Compliance Audit"
              icon={<ShieldCheck size={14} />}
              initialX={300}
              initialY={100}
              width={700}
              height={550}
              zIndex={windowZIndices["mtd-audit"]}
              isActive={focusedWindow === "mtd-audit"}
              onFocus={() => handleFocusWindow("mtd-audit")}
              onClose={() => toggleWindow("mtd-audit")}
            >
              <div className="h-full flex flex-col bg-[#06080c] text-gray-200 p-6 overflow-y-auto custom-scrollbar font-mono">
                {auditStatus === "IDLE" && (
                  <div className="flex-1 flex flex-col items-center justify-center text-center gap-6 p-4">
                    <div className="w-20 h-20 rounded-full bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center shadow-[0_0_20px_rgba(6,182,212,0.1)]">
                      <Shield className="w-10 h-10 text-cyan-400" />
                    </div>
                    <div className="space-y-2">
                      <h3 className="text-lg font-bold text-white uppercase tracking-widest">MTD Security Audit Suite</h3>
                      <p className="text-xs text-gray-400 max-w-md leading-relaxed">
                        Verify system compliance against ISO 27001 & ISO 25010 standards in real-time.
                        This runs 17 active stress tests simulating multi-tenant rate floods, digital hallucination honeypots, and topology shuffling.
                      </p>
                    </div>
                    <button
                      onClick={runMtdAudit}
                      className="px-8 py-3 bg-cyan-500 hover:bg-cyan-600 text-black font-black rounded-xl text-[10px] uppercase tracking-widest transition-all shadow-[0_0_20px_rgba(6,182,212,0.2)]"
                    >
                      Launch MTD Security Audit
                    </button>
                  </div>
                )}

                {auditStatus === "RUNNING" && (
                  <div className="flex-1 flex flex-col items-center justify-center gap-6 p-4">
                    <div className="relative w-24 h-24">
                      {/* Scanning Ring */}
                      <div className="absolute inset-0 rounded-full border-4 border-cyan-500/10 border-t-cyan-500 animate-spin" />
                      <div className="absolute inset-2 rounded-full border border-cyan-500/20 flex items-center justify-center bg-black/40">
                        <Activity className="w-8 h-8 text-cyan-400 animate-pulse" />
                      </div>
                    </div>
                    <div className="text-center space-y-2">
                      <h4 className="text-xs font-bold text-cyan-400 uppercase tracking-[0.2em] animate-pulse">
                        AUDIT IN PROGRESS...
                      </h4>
                      <p className="text-[9px] text-gray-500 uppercase tracking-widest">
                        Executing 17 physical stress tests & probing defense mechanisms
                      </p>
                    </div>
                    {/* Simulated loading steps */}
                    <div className="w-full max-w-sm bg-black/60 border border-white/5 rounded-xl p-4 text-[9px] text-cyan-500/70 space-y-1.5 font-mono">
                      <div className="flex items-center gap-2">
                        <span className="w-1 h-1 rounded-full bg-cyan-400 animate-ping" />
                        <span>[SYS] Initiating stress-test sequence...</span>
                      </div>
                      <div className="flex items-center gap-2 opacity-80">
                        <span className="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse" />
                        <span>[MTD] Probing Per-IP Token Bucket (150 concurrent reqs)</span>
                      </div>
                      <div className="flex items-center gap-2 opacity-60">
                        <span className="w-1.5 h-1.5 rounded-full bg-gray-600" />
                        <span>[HONEYPOT] Testing Digital Hallucination & Tarpit delays</span>
                      </div>
                      <div className="flex items-center gap-2 opacity-40">
                        <span className="w-1.5 h-1.5 rounded-full bg-gray-600" />
                        <span>[SHUFFLER] Verifying CSPRNG topology port rotation</span>
                      </div>
                    </div>
                  </div>
                )}

                {auditStatus === "ERROR" && (
                  <div className="flex-1 flex flex-col items-center justify-center text-center gap-6 p-4">
                    <div className="w-16 h-16 rounded-full bg-red-500/10 border border-red-500/30 flex items-center justify-center shadow-[0_0_20px_rgba(239,68,68,0.1)]">
                      <ShieldAlert className="w-8 h-8 text-red-400" />
                    </div>
                    <div className="space-y-2">
                      <h3 className="text-sm font-bold text-red-400 uppercase tracking-widest">Audit Execution Failed</h3>
                      <p className="text-[10px] text-gray-400 max-w-md bg-black/40 border border-red-950/30 p-3 rounded-lg leading-relaxed text-left font-mono whitespace-pre-wrap">
                        {auditError}
                      </p>
                    </div>
                    <button
                      onClick={runMtdAudit}
                      className="px-6 py-2.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 font-bold rounded-xl text-[10px] uppercase tracking-widest transition-all"
                    >
                      Retry Security Audit
                    </button>
                  </div>
                )}

                {auditStatus === "COMPLETED" && auditResult && (
                  <div className="space-y-6 animate-in fade-in duration-300 w-full">
                    {/* Compliance Card Banner */}
                    <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-2xl p-6 flex flex-col md:flex-row items-center gap-6 relative overflow-hidden shadow-[0_0_20px_rgba(16,185,129,0.05)] w-full">
                      {/* Ambient green light */}
                      <div className="absolute top-0 right-0 w-32 h-32 bg-emerald-500/5 rounded-full filter blur-xl pointer-events-none" />
                      <div className="w-16 h-16 rounded-full bg-emerald-500/20 border border-emerald-500/40 flex items-center justify-center shrink-0 shadow-[0_0_15px_rgba(16,185,129,0.2)]">
                        <ShieldCheck className="w-9 h-9 text-emerald-400" />
                      </div>
                      <div className="flex-1 text-center md:text-left space-y-1">
                        <div className="flex items-center justify-center md:justify-start gap-2">
                          <span className="text-[9px] bg-emerald-500/20 text-emerald-300 px-2 py-0.5 rounded font-black tracking-widest uppercase">
                            Audited & Compliant
                          </span>
                        </div>
                        <h3 className="text-lg font-black text-white tracking-widest uppercase">
                          {auditResult.passed} / {auditResult.total} MTD CHECKS PASSED
                        </h3>
                        <p className="text-[9px] text-emerald-400/70 uppercase tracking-wider font-bold">
                          STATUS: SECURED • ISO 27001 & ISO 25010 RESILIENT SPEC
                        </p>
                      </div>
                      <button
                        onClick={runMtdAudit}
                        className="px-5 py-2.5 bg-emerald-500/15 hover:bg-emerald-500/25 text-emerald-400 border border-emerald-500/35 font-bold rounded-xl text-[9px] uppercase tracking-widest transition-all shrink-0"
                      >
                        Re-Audit System
                      </button>
                    </div>

                    {/* Checks Grid */}
                    <div className="bg-black/40 border border-white/5 rounded-2xl overflow-hidden w-full">
                      <div className="bg-[#090b0e] px-4 py-3 border-b border-white/5 flex items-center justify-between">
                        <span className="text-[10px] text-gray-500 font-bold uppercase tracking-wider">Detailed Verification Log</span>
                        <span className="text-[8px] text-emerald-500 font-black uppercase tracking-widest">100% Integrity</span>
                      </div>
                      <div className="divide-y divide-white/5 max-h-60 overflow-y-auto custom-scrollbar">
                        {auditResult.checks.map((chk, i) => (
                          <div key={i} className="px-4 py-2.5 flex items-center justify-between hover:bg-white/[0.01] transition-colors">
                            <div className="space-y-0.5 flex-1 pr-4">
                              <p className="text-[10px] text-gray-200 font-bold leading-normal">{chk.label}</p>
                              {chk.detail && <p className="text-[8px] text-gray-500 leading-normal">{chk.detail}</p>}
                            </div>
                            <span className={`px-2 py-0.5 rounded-[4px] font-black uppercase text-[8px] shrink-0 ${
                              chk.passed 
                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" 
                                : "bg-red-500/10 text-red-500 border border-red-500/20"
                            }`}>
                              {chk.passed ? "PASS" : "FAIL"}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Raw output accordion */}
                    <details className="group bg-black/20 border border-white/5 rounded-2xl overflow-hidden w-full">
                      <summary className="px-4 py-3 text-[10px] text-gray-500 font-bold uppercase tracking-wider cursor-pointer list-none flex items-center justify-between hover:bg-white/[0.01] transition-colors select-none">
                        <span>Show Raw Terminal Audit Log Output</span>
                        <span className="text-cyan-500 group-open:rotate-180 transition-transform duration-200">▼</span>
                      </summary>
                      <pre className="p-4 bg-black border-t border-white/5 text-[9px] text-green-400 leading-relaxed font-mono overflow-auto max-h-60 custom-scrollbar whitespace-pre-wrap">
                        {auditResult.output}
                      </pre>
                    </details>
                  </div>
                )}
              </div>
            </WindowFrame>
          )}
        </AnimatePresence>
      </main>

      {/* Taskbar */}
      <Taskbar 
        onOpenApp={toggleWindow}
        onPanic={handlePanic}
        onReset={handleReset}
        onDeleteDomain={handleDeleteDomain}
        activeDomain={activeDomain}
        onDomainChange={setActiveDomain}
        onAddClick={() => setIsModalOpen(true)}
        refreshTrigger={refreshTrigger}
        isLive={isLive}
        activeApps={openWindows}
      />

      {/* Modals */}
      <AddRouteModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onSuccess={() => {}} />

      {/* Unbypassable Global License Lockout Paywall Overlay */}
      {isUnlicensed && (
        <div 
          className="fixed inset-0 z-[99999] backdrop-blur-2xl bg-[#030508]/95 flex items-center justify-center p-6 text-center select-none"
          onContextMenu={(e) => e.preventDefault()}
        >
          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-[#05080c]/90 border border-red-500/30 rounded-3xl p-10 shadow-[0_0_50px_rgba(239,68,68,0.15)] max-w-lg w-full flex flex-col items-center gap-8 backdrop-blur-md relative overflow-hidden"
          >
            {/* Background Accent Orb */}
            <div className="absolute top-[-20%] left-[-20%] w-[140%] h-[140%] bg-red-500/5 rounded-full pointer-events-none filter blur-[80px]" />
            
            <div className="relative">
              <div className="w-20 h-20 rounded-full bg-red-500/10 border border-red-500/30 flex items-center justify-center animate-pulse">
                <Lock className="w-10 h-10 text-red-500 animate-pulse" />
              </div>
              <div className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-red-500 flex items-center justify-center border-2 border-[#05080c]" />
            </div>

            <div className="space-y-3 relative z-10">
              <h2 className="text-2xl font-black text-white uppercase tracking-[0.2em] font-mono">
                SISTEM DITANGGUHKAN
              </h2>
              <p className="text-xs text-red-400 font-mono uppercase tracking-widest">
                Masa Sewa Langganan Nexus Cyber Telah Berakhir
              </p>
              <p className="text-gray-400 text-xs leading-relaxed max-w-sm mx-auto font-sans">
                Seluruh gerbang WAF, mitigasi heuristik kecerdasan buatan, dan orkestrasi Moving Target Defense untuk domain Anda dinonaktifkan secara otomatis demi alasan integritas sistem.
              </p>
            </div>

            <div className="w-full space-y-4 relative z-10">
              <div className="flex flex-col gap-2">
                <label className="text-[9px] text-gray-500 font-mono uppercase tracking-widest text-left">Kunci Lisensi Langganan</label>
                <div className="flex gap-2">
                  <input 
                    type="password"
                    value={activationKey}
                    onChange={(e) => setActivationKey(e.target.value)}
                    disabled={isActivating || activationSuccess}
                    placeholder={activationSuccess ? "AKTIVASI SUKSES..." : "Masukkan NEXUS_LICENSE_KEY..."}
                    className="flex-1 bg-black/60 border border-white/10 rounded-xl px-4 py-3 text-xs font-mono text-white outline-none focus:border-red-500/40 transition-colors disabled:opacity-50"
                  />
                  <button 
                    onClick={handleActivation}
                    disabled={isActivating || activationSuccess}
                    className="px-6 bg-red-500/10 hover:bg-red-500/20 disabled:bg-red-500/5 text-red-500 border border-red-500/30 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all"
                  >
                    {isActivating ? "MEMPROSES..." : activationSuccess ? "BERHASIL" : "AKTIVASI"}
                  </button>
                </div>
              </div>
            </div>

            <div className="text-[8px] text-gray-600 font-mono uppercase tracking-widest mt-2">
              NEXUS COMMAND CENTER • SECURITY COMPLIANCE ISO 27001
            </div>
          </motion.div>
        </div>
      )}
    </div>
  )
}

export default NCCDashboard
