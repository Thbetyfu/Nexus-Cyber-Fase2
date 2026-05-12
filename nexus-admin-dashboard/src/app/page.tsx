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
        const data = await res.json()
        if (!pollingActive) return

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

  return { logs, metrics, history, shufflerData, isLive }
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
  
  const { logs, metrics, history, shufflerData, isLive } = useTelemetry(`http://localhost:8080/api/telemetry?domain=${activeDomain}`, 2000)
  const aiEvents = useAIEvents('http://localhost:8080/api/ai-events', 1000)

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
    "system-status": 14
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
        {isBooting && <BootSequence onComplete={() => setIsBooting(false)} />}
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
        </div>

        {/* Floating Windows Layer */}
        <AnimatePresence>
          {/* Metrics Window */}
          {openWindows.includes("metrics") && (
            <WindowFrame
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
    </div>
  )
}

export default NCCDashboard
