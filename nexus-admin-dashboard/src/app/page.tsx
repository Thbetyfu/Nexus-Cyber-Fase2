"use client"

/* 
   NEXUS_UX_STABILITY_COVENANT [LOCKED-BY-ANTIGRAVITY]
   - Peraturan 1: Dilarang keras menambahkan scrollIntoView() atau focus() di file ini.
   - Peraturan 2: Dashboard harus tetap berada di tingkat tinggi viewport (100vh).
*/
import React, { useEffect, useState, useRef } from "react"
import {
  Shield, Zap, Activity, ShieldAlert, Cpu, Ghost, Globe, Terminal,
  ChevronRight, Server, Search, MessageSquare, Send, Bot,
  AlertTriangle, ShieldCheck, Lock, RotateCcw, Crosshair, Plus, X, Loader2,
  WifiOff
} from 'lucide-react';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as ChartTooltip, ResponsiveContainer
} from "recharts"
import NechatWidget from '@/components/NechatWidget';
import DomainSwitcher from '@/components/DomainSwitcher';
import AddRouteModal from '@/components/AddRouteModal';
import AiTerminalWidget from '@/components/AiTerminalWidget';
import ThreatMapWidget from '@/components/ThreatMapWidget';

// Type definition for Live Telemetry Log
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

// Type definition for Cognitive AI Events
export interface AIEventLog {
  timestamp: string;
  layer: string;
  status: string;
  detail_action: string;
}

// Custom Hook for Real-Time SOC Polling (Phase 7: Multi-Tenant aware)
function useTelemetry(url: string, intervalMs: number = 2000) {
  const [logs, setLogs] = useState<TelemetryLog[]>([])
  const [metrics, setMetrics] = useState({ allowed: 0, blocked: 0, honeypot: 0, panics: 0 })
  const [shufflerData, setShufflerData] = useState({ port: 3001, status: "OFFLINE" })
  const [isLive, setIsLive] = useState(true)

  // 1. MATRIX CLOCK SYNC: Pre-populate timeline with real-time markers
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

        // 2. MATRIX SYNC: Initialize or Reset markers on gateway restart
        if (isFirstFetch.current || currentTrafficTotal < prevTrafficRef.current) {
          prevTrafficRef.current = currentTrafficTotal
          prevHoneypotRef.current = currentHoneypotTotal
          isFirstFetch.current = false

          // [NEW: HISTORICAL RECONSTRUCTION]
          // Scan the last 30 logs to pre-populate the chart with correct metrics
          const historicalLogs = (data.recent_logs || []).slice(0, 30).reverse();
          const reconstructedTimeline = historicalLogs.map((l: any, i: number) => ({
            time: new Date(l.timestamp).toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
            allowed: l.status === "ALLOWED" ? 1 : 0,
            honeypot: (l.status.includes("HONEYPOT") || l.status.includes("PATCH") || l.status.includes("DIVERTED")) ? 1 : 0
          }));

          if (reconstructedTimeline.length > 0) {
            timeline.current = [...initialTimeline.slice(0, 40 - reconstructedTimeline.length), ...reconstructedTimeline];
            setHistory([...timeline.current])
          }
          return
        }

        // 3. CALCULATION: Delta Per Second (Accurate RPS)
        const trafficRPS = Math.max(0, currentTrafficTotal - prevTrafficRef.current)
        const honeypotRPS = Math.max(0, currentHoneypotTotal - prevHoneypotRef.current)

        prevTrafficRef.current = currentTrafficTotal
        prevHoneypotRef.current = currentHoneypotTotal

        const timeLabel = new Date().toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
        const point = { time: timeLabel, allowed: trafficRPS, honeypot: honeypotRPS }

        timeline.current = [...timeline.current, point].slice(-40)
        setHistory([...timeline.current])

      } catch (error) {
        if (pollingActive) {
          setIsLive(false)
        }
      }
    };

    fetchTelemetry()
    const interval = setInterval(fetchTelemetry, intervalMs)
    return () => { pollingActive = false; clearInterval(interval); }

  }, [url, intervalMs])

  return { logs, metrics, history, shufflerData, isLive }
}

// Generic Hook for catching the fastest AI cognitive streams
function useAIEvents(url: string, intervalMs: number = 1000) {
  const [events, setEvents] = useState<AIEventLog[]>([]);
  useEffect(() => {
    let pollingActive = true;
    const fetchAPI = async () => {
      try {
        const res = await fetch(url, { cache: 'no-store' });
        if (!res.ok) return;
        const data = await res.json();
        if (pollingActive) {
          console.log('[NEXUS-DEBUG] AI Events Data:', data);
          setEvents(data || []);
        }
      } catch (err) { }
    };
    fetchAPI();
    const timer = setInterval(fetchAPI, intervalMs);
    return () => { pollingActive = false; clearInterval(timer); };
  }, [url, intervalMs]);
  return events;
}

const SOCDashboard = () => {
  const [activeDomain, setActiveDomain] = useState<string>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [logLimit, setLogLimit] = useState<number>(10); // Default Matrix View Focus

  const { logs, metrics, history, shufflerData, isLive } = useTelemetry(`http://localhost:8081/api/telemetry?domain=${activeDomain}`, 2000)
  const aiEvents = useAIEvents('http://localhost:8081/api/ai-events', 1000)

  const tableContainerRef = useRef<HTMLDivElement>(null)
  const thoughtContainerRef = useRef<HTMLDivElement>(null)

  const handlePanic = async () => {
    try {
      await fetch("http://localhost:8081/api/panic", { method: "POST" })
    } catch (err) {
      console.error("Rescue Protocol Trigger Failed", err)
    }
  }

  const handleReset = async () => {
    console.log("🚀 NEXUS SYSTEM: Triggering total cognitive purge...");
    if (window.confirm("🚨 PURGE ALL SYSTEM DATA? This will reset metrics, clear AI memory (antibodies), and wipe all forensic logs. Are you sure?")) {
      try {
        const res = await fetch("http://localhost:8081/api/system/reset", {
          method: "POST",
          mode: 'cors'
        })
        console.log("📡 Gateway Response Status:", res.status);
        if (res.ok) {
          console.log("✅ RESET SUCCESSFUL. Refreshing matrix...");
          alert("✅ NEXUS PURGE SUCCESSFUL: System memory cleared.");
          window.location.reload();
        }
      } catch (err) {
        console.error("System Purge Failed", err)
        alert("⚠️ SYSTEM PURGE FAILED: Could not reach Gateway.");
      }
    }
  }

  // 1. DYNAMIC LOG SLICER: Filter entry based on user choice
  const displayLogs = logs.slice(0, logLimit);

  useEffect(() => {
    // Scroll removed to prevent jumping
  }, [logs])

  useEffect(() => {
    // Scroll removed to prevent jumping
  }, [aiEvents])

  return (
    <div className="relative min-h-screen bg-[#06080b] text-gray-200 font-sans selection:bg-blue-600/30 flex flex-col">
      {!isLive && (
        <div className="absolute inset-0 z-[100] backdrop-blur-sm bg-black/40 flex items-center justify-center p-6">
          <div className="bg-[#0c0f14] border border-red-900/40 rounded-2xl p-8 shadow-2xl max-w-md w-full text-center flex flex-col items-center gap-6">
            <div className="relative h-20 w-20">
              <span className="absolute inset-0 rounded-full bg-red-500/20 animate-ping opacity-50"></span>
              <div className="relative h-20 w-20 flex items-center justify-center bg-red-500/10 rounded-full border border-red-500/30">
                <WifiOff className="w-10 h-10 text-red-500" />
              </div>
            </div>
            <div className="space-y-2">
              <h2 className="text-2xl font-bold text-white tracking-tight">Sync Connection Lost</h2>
              <p className="text-gray-400 text-sm leading-relaxed">
                The Dashboard has lost contact with the Nexus Core Gateway.
                Retrying connection in seconds...
              </p>
            </div>
            <button
              onClick={() => window.location.reload()}
              className="px-6 py-2.5 bg-red-500/10 hover:bg-red-500/20 text-red-500 border border-red-500/30 rounded-lg text-sm font-semibold transition flex items-center gap-2 group"
            >
              <RotateCcw className="w-4 h-4 group-hover:rotate-180 transition-transform duration-500" />
              Force Matrix Reload
            </button>
          </div>
        </div>
      )}
      {/* 🛡️ SOC HEADER - Lifted to Absolute Top Layer */}
      <header className="flex items-center justify-between px-6 py-4 border-b border-gray-800/80 bg-[#090b0e] shadow-md z-[1000] sticky top-0 shrink-0">
        <div className="flex items-center gap-4">
          <div className="bg-emerald-500/10 p-2 rounded-lg border border-emerald-500/20 shadow-[0_0_15px_rgba(16,185,129,0.15)]">
            <Shield className="w-8 h-8 text-emerald-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white/90">NEXUS CYBER SOC</h1>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="relative flex h-2 w-2">
                <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${isLive ? 'bg-emerald-500' : 'bg-red-500'}`}></span>
                <span className={`relative inline-flex rounded-full h-2 w-2 ${isLive ? 'bg-emerald-500' : 'bg-red-500'}`}></span>
              </span>
              <p className={`text-xs ${isLive ? 'text-emerald-500/80' : 'text-red-500/80'} uppercase tracking-widest font-semibold`}>
                {isLive ? 'Active Monitoring • MTD Matrix v7.0' : 'Sync Interrupted • Reconnecting...'}
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-6">
          <DomainSwitcher
            activeDomain={activeDomain}
            onDomainChange={setActiveDomain}
            onAddClick={() => setIsModalOpen(true)}
          />

          {/* [NEW: EXECUTIVE REPORTING] Professional PDF Intelligence Report Engine */}
          <button
            onClick={async () => {
              const btn = document.getElementById('btn-report') as HTMLButtonElement;
              if (btn) {
                // [PRE-IGNITION] Open the window IMMEDIATELY to bypass Pop-up Blocker
                const printWindow = window.open('', '_blank');
                if (!printWindow) {
                  alert("⚠️ POP-UP BLOCKED! Harap izinkan pop-up untuk dashboard ini agar laporan PDF dapat diunduh.");
                  return;
                }

                // Initial UI for the loading state
                printWindow.document.write('<html><body style="background:#0c0e12;color:#3b82f6;display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif;"><div><h1>NEXUS CORTEX: Synthesizing Intelligence...</h1><p style="color:#64748b;text-align:center;">Please wait while we aggregate security metrics...</p></div></body></html>');

                btn.disabled = true;
                btn.innerText = "📄 Synthesizing PDF intelligence...";

                try {
                  const res = await fetch(`http://localhost:8081/api/report/generate?domain=${activeDomain}`);
                  const data = await res.json();

                  if (data.status === 'success') {
                    const reportHtml = `
                        <html>
                          <head>
                            <title>Nexus_Incident_Report_${activeDomain}</title>
                            <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;700;800&family=JetBrains+Mono&display=swap" rel="stylesheet">
                            <style>
                              body { font-family: 'Inter', sans-serif; padding: 50px; color: #1e293b; line-height: 1.6; background: white; }
                              .header { border-bottom: 3px solid #3b82f6; padding-bottom: 20px; margin-bottom: 30px; display: flex; justify-content: space-between; align-items: flex-end; }
                              .logo { font-weight: 800; font-size: 24px; color: #1d4ed8; letter-spacing: -1px; }
                              .meta { text-align: right; color: #64748b; font-size: 12px; font-family: 'JetBrains Mono', monospace; }
                              h1, h2, h3 { color: #0f172a; margin-top: 1.5em; border-left: 4px solid #3b82f6; padding-left: 15px; }
                              table { width: 100%; border-collapse: collapse; margin: 20px 0; font-size: 14px; }
                              th { background: #f8fafc; text-align: left; padding: 12px; border: 1px solid #e2e8f0; font-weight: 700; }
                              td { padding: 12px; border: 1px solid #e2e8f0; }
                              .footer { margin-top: 50px; font-size: 10px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 10px; text-align: center; }
                              @media print { .no-print { display: none; } }
                            </style>
                          </head>
                          <body>
                            <div class="header">
                              <div class="logo">NEXUS CYBER COMMAND CENTER</div>
                              <div class="meta">
                                DOC_ID: ${activeDomain.toUpperCase()}_SOC_${new Date().getTime()}<br>
                                TIMESTAMP: ${new Date().toLocaleString()}
                              </div>
                            </div>
                            <div id="content"></div>
                            <div class="footer">CONFIDENTIAL - FOR AUTHORIZED EYES ONLY - NEXUS CYBER INTELLIGENCE UNIT</div>
                            <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
                            <script>
                              document.getElementById('content').innerHTML = marked.parse(\`${data.report_content.replace(/`/g, '\\`').replace(/\$/g, '\\$')}\`);
                              window.onload = () => { setTimeout(() => { window.print(); }, 800); };
                            </script>
                          </body>
                        </html>
                    `;
                    printWindow.document.open();
                    printWindow.document.write(reportHtml);
                    printWindow.document.close();
                  } else {
                    printWindow.close();
                  }
                } catch (e) {
                  console.error("PDF Synthesis Failed", e);
                  printWindow.close();
                  alert("Gagal melakukan sintesis laporan. Pastikan Gateway (8080) menyala.");
                } finally {
                  btn.disabled = false;
                  btn.innerText = "📄 Generate AI Report";
                }
              }
            }}
            id="btn-report"
            className="flex items-center gap-2 bg-blue-600/10 border border-blue-500/30 hover:bg-blue-600/20 text-blue-400 px-3 py-1.5 rounded text-[11px] font-bold transition-all"
          >
            📄 Generate AI Report
          </button>

          <button
            onClick={handleReset}
            className="flex items-center gap-2 bg-yellow-500/10 border border-yellow-500/30 hover:bg-yellow-500/20 text-yellow-500 px-3 py-1.5 rounded text-[11px] font-bold transition-all group"
          >
            <RotateCcw className="w-3.5 h-3.5 group-hover:rotate-180 transition-transform duration-500" />
            SYSTEM RESET
          </button>

          <div className="relative group">
            <button
              onClick={handlePanic}
              title="🚨 KILL-SWITCH DARURAT: Tekan jika ada serangan APT / Zero-Day menembus masuk!"
              className="flex items-center gap-2 border border-red-500/40 bg-red-950/20 hover:bg-red-500/40 rounded px-3 py-1.5 shadow-inner transition-all group"
            >
              <ShieldAlert className="w-4 h-4 text-red-500 group-hover:scale-110 transition-transform" />
              <span className="text-red-500 font-bold text-[11px] tracking-wider uppercase">Emergency Rescue</span>
            </button>
            <div className="absolute top-full right-0 mt-3 w-80 bg-[#0c0e12] border border-red-500/50 rounded-xl p-4 text-[12px] text-red-100 hidden group-hover:block pointer-events-none z-[9999] backdrop-blur-3xl shadow-[0_0_50px_rgba(239,68,68,0.2)] border-t-red-500 border-t-2 animate-in fade-in zoom-in duration-200">
              <div className="flex items-start gap-4">
                <div className="bg-red-500/20 p-2 rounded-lg border border-red-500/30">
                  <Zap className="w-5 h-5 text-red-500" />
                </div>
                <div>
                  <p className="font-black mb-1.5 text-red-400 uppercase tracking-tighter text-sm">Protokol Penyelamatan Darurat</p>
                  <p className="leading-relaxed opacity-90 text-[11px] text-slate-300">
                    Gunakan <span className="text-red-400 font-bold italic underline">Kill-Switch</span> ini untuk memutus semua jalur serangan aktif.
                    Direkomendasikan hanya untuk menghadapi ancaman level <span className="text-white font-bold">APT (Advanced Persistent Threat)</span> atau <span className="text-white font-bold">Zero-Day</span>.
                  </p>
                </div>
              </div>
              <div className="mt-4 pt-3 border-t border-red-500/10 flex justify-between items-center">
                <span className="text-[10px] text-red-500/50 font-mono">STATUS: READY</span>
                <span className="text-[10px] text-red-400/80 italic font-medium px-2 py-0.5 bg-red-500/10 rounded">Nexus Matrix Enforced</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3 border border-blue-900/40 bg-blue-950/20 rounded-lg px-3 py-1.5 shadow-inner">
            <Server className="w-4 h-4 text-blue-500" />
            <div className="flex flex-col">
              <span className="text-[10px] font-bold text-blue-400/70 tracking-tighter uppercase leading-none mb-1">OJK_BACKEND_TARGET</span>
              <div className="flex items-center gap-2">
                <span className="text-[11px] font-mono text-blue-100">PORT:{shufflerData.port}</span>
                <div className={`w-1.5 h-1.5 rounded-full ${shufflerData.status === 'CONNECTED' ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'bg-red-500 animate-pulse shadow-[0_0_8px_rgba(239,68,68,0.5)]'}`}></div>
                <span className={`text-[9px] font-bold tracking-widest ${shufflerData.status === 'CONNECTED' ? 'text-emerald-500' : 'text-red-500'}`}>
                  {shufflerData.status}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* 🚀 MAIN CONTENT */}
      <main className="flex-1 flex flex-col p-6 gap-6 max-w-screen-2xl mx-auto w-full">
        <section className="grid grid-cols-1 lg:grid-cols-4 gap-6 shrink-0">
          {/* Metrics Column */}
          <div className="lg:col-span-1 flex flex-col gap-4">
            <div className="bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl p-5 shadow-lg group hover:border-blue-900/50 transition duration-300">
              <div className="flex justify-between items-start mb-2">
                <p className="text-xs text-gray-500 uppercase tracking-widest font-bold">Traffic Source ({activeDomain === 'all' ? 'GLOBAL' : activeDomain})</p>
                <Activity className="w-4 h-4 text-blue-500/50 group-hover:text-blue-400" />
              </div>
              <p className="text-3xl font-bold text-gray-200 font-mono tracking-tight">{metrics.allowed.toLocaleString()}</p>
            </div>

            <div className="bg-[#120808]/80 backdrop-blur border border-red-900/30 rounded-xl p-5 shadow-lg relative overflow-hidden group hover:border-red-900/50 transition">
              <div className="flex justify-between items-start mb-2 relative z-10">
                <p className="text-xs text-red-500/80 uppercase tracking-widest font-bold">Honeypot Trapped</p>
                <ShieldAlert className="w-4 h-4 text-red-500/50 group-hover:text-red-400" />
              </div>
              <p className="text-3xl font-bold text-red-500 font-mono tracking-tight relative z-10">{metrics.honeypot.toLocaleString()}</p>
            </div>

            <div className="bg-[#0b1218]/80 backdrop-blur border border-blue-900/30 rounded-xl p-5 shadow-lg group z-10 hover:border-blue-900/50 transition">
              <div className="flex justify-between items-start mb-2">
                <p className="text-xs text-blue-400/80 uppercase tracking-widest font-bold">Rescue Protocols</p>
                <RotateCcw className="w-4 h-4 text-blue-500/50 group-hover:rotate-180 transition-transform duration-500" />
              </div>
              <p className="text-3xl font-bold text-blue-400 font-mono tracking-tight">{metrics.panics.toLocaleString()}</p>
            </div>
          </div>

          {/* 🛡️ LIVE THREAT MAP (Visual Situation Room) */}
          <div className="lg:col-span-3 h-[400px]">
            <ThreatMapWidget />
          </div>
        </section>

        {/* Real-time Traffic Vectors (Moved below for hierarchy) */}
        <section className="bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl p-5 flex flex-col shadow-lg h-[250px] shrink-0">
          <h3 className="text-xs font-semibold uppercase tracking-widest text-gray-500 mb-4 flex items-center gap-2">
            <Activity className="w-4 h-4 text-blue-500" /> Real-time Traffic Vectors
          </h3>
          <div className="w-full flex-1 min-h-[180px]">
            <ResponsiveContainer width="100%" height={180}>
              <AreaChart data={history} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorSafe" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorThreat" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.5} />
                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" vertical={true} opacity={0.3} />
                <XAxis dataKey="time" stroke="#9ca3af" fontSize={11} tickMargin={12} axisLine={false} tickLine={false} />
                <YAxis
                  stroke="#9ca3af"
                  fontSize={11}
                  axisLine={false}
                  tickLine={false}
                  width={30}
                  domain={[0, (dataMax: number) => Math.max(10, dataMax + 5)]}
                  allowDecimals={false}
                />
                <ChartTooltip
                  contentStyle={{ backgroundColor: '#030712', borderColor: '#374151', color: '#f3f4f6', borderRadius: '8px', fontSize: '12px' }}
                  itemStyle={{ color: '#e5e7eb' }}
                />
                <Area type="monotone" dataKey="allowed" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorSafe)" isAnimationActive={false} dot={{ r: 1, fill: '#3b82f6', fillOpacity: 1, strokeWidth: 0 }} />
                <Area type="monotone" dataKey="honeypot" stroke="#ef4444" strokeWidth={3} fillOpacity={1} fill="url(#colorThreat)" isAnimationActive={false} dot={{ r: 1, fill: '#ef4444', fillOpacity: 1, strokeWidth: 0 }} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        {/* MIDDLE SECTION: AI Cognitive Core & Self-Repair Tracker */}
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6 shrink-0 h-72">
          {/* AI Terminal Stream (SSE Linked) */}
          <AiTerminalWidget />

          {/* Autonomous Operations Timeline */}
          <div className="bg-[#0c0f14]/80 backdrop-blur border border-emerald-900/30 rounded-xl flex flex-col shadow-lg overflow-hidden h-full">
            <div className="bg-[#090b0e] px-4 py-2 border-b border-emerald-900/30 flex items-center justify-between sticky top-0 z-10 shrink-0">
              <h3 className="text-xs font-semibold text-emerald-500 uppercase tracking-widest flex items-center gap-2">
                <Cpu className="w-4 h-4" /> Autonomous Operations
              </h3>
              <span className="text-[10px] text-emerald-500/70 font-mono bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">SELF_REPAIR_LOG</span>
            </div>
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
              {(() => {
                const activeInterventions = aiEvents.filter(e =>
                  e.layer === 'Self-Repair' ||
                  e.layer === 'Virtual-Patch' ||
                  e.status === 'MITIGATING' ||
                  e.status === 'INSTANT_DROP_PATCH' ||
                  e.status === 'IMMUNE' ||
                  e.status === 'VIRTUAL_PATCH'
                );

                if (activeInterventions.length === 0) {
                  return <div className="text-slate-600 font-mono text-xs flex items-center justify-center h-full">No interventions generated. System stable.</div>;
                }

                return (
                  <div className="flex flex-col gap-3 relative before:absolute before:inset-0 before:ml-2 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-emerald-500/20 before:to-transparent">
                    {activeInterventions.reverse().map((ev, i) => (
                      <div key={i} className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
                        <div className="bg-emerald-500/20 border border-emerald-500/50 w-4 h-4 rounded-full shadow-[0_0_10px_rgba(16,185,129,0.3)] shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 flex items-center justify-center relative z-10 ml-0 md:ml-auto">
                          <div className="w-1.5 h-1.5 rounded-full bg-emerald-400"></div>
                        </div>
                        <div className="w-[calc(100%-2.5rem)] md:w-[calc(50%-1.5rem)] bg-[#05080c] border border-slate-800/80 p-3 rounded-lg shadow group-hover:border-emerald-500/40 transition-colors">
                          <div className="flex items-center justify-between mb-1">
                            <span className="font-bold text-emerald-400 text-[10px] uppercase tracking-wider">{ev.status}</span>
                            <span className="font-mono text-slate-500 text-[9px]">{new Date(ev.timestamp).toLocaleTimeString('en-US', { hour12: false })}</span>
                          </div>
                          <p className="text-slate-300 text-[11px] leading-tight break-words">{ev.detail_action}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                );
              })()}
            </div>
          </div>
        </section>

        {/* LOWER SECTION: Real-Time Access & Threat Log */}
        <section className="flex-1 flex flex-col bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl overflow-hidden shadow-lg min-h-[400px]">
          <div className="bg-[#090b0e] px-5 py-3 border-b border-gray-800/80 flex items-center justify-between sticky top-0 z-20 shadow-sm">
            <div className="flex items-center gap-4">
              <h3 className="text-sm font-semibold text-gray-200 uppercase tracking-wider flex items-center gap-2">
                <Lock className="w-4 h-4 text-emerald-500" />
                The Rogue Gallery: Access & Forensic Log
              </h3>
              {/* DROPDOWN FILTER ENTRY */}
              <div className="flex items-center gap-2 bg-black/40 border border-gray-800 rounded px-2 py-1">
                <span className="text-[9px] text-gray-500 uppercase font-black font-mono">Entries:</span>
                <select
                  value={logLimit}
                  onChange={(e) => setLogLimit(Number(e.target.value))}
                  className="bg-transparent text-[10px] text-emerald-500 font-bold font-mono outline-none cursor-pointer"
                >
                  <option value={5}>5 UNIT</option>
                  <option value={10}>10 UNIT</option>
                  <option value={50}>50 UNIT</option>
                  <option value={100}>100 UNIT</option>
                </select>
              </div>
            </div>
            <span className="text-[10px] bg-red-500/10 text-red-400 border border-red-500/20 px-2 py-1 rounded font-mono shadow-inner tracking-widest">
              🔴 LIVE STREAM
            </span>
          </div>

          <div className="flex-1 overflow-auto bg-[#07090c]" ref={tableContainerRef}>
            <table className="w-full text-left border-collapse text-sm">
              <thead className="bg-[#0a0d11] text-xs uppercase text-gray-500 tracking-wider sticky top-0 z-10 shadow-md">
                <tr>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[15%]">Timestamp</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[15%]">IP / Actor ID</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[20%]">Geo-Location & Domain</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[25%]">Target Endpoint / Info</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 text-center w-[15%]">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800/50">
                {displayLogs.map((log, idx) => {
                  const isHacker = log.status !== "ALLOWED";
                  const rowClass = isHacker
                    ? "hover:bg-red-950/20 transition-colors bg-red-950/10"
                    : "hover:bg-blue-900/10 transition-colors";
                  return (
                    <tr key={idx} className={`group ${rowClass}`}>
                      <td className="py-4 px-6">
                        <span className="font-mono text-gray-400 text-xs uppercase tracking-tighter">
                          {new Date(log.timestamp).toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                        </span>
                      </td>

                      <td className="py-4 px-6">
                        <div className="flex flex-col">
                          <span className="font-mono text-gray-300 text-sm">{log.source_ip}</span>
                          {log.attacker_id && (
                            <span className="font-mono text-[10px] text-red-400/80 mt-1 uppercase">
                              {log.attacker_id}
                            </span>
                          )}
                        </div>
                      </td>

                      <td className="py-4 px-6">
                        <div className="flex flex-col gap-0.5">
                          <span className="text-gray-300 text-sm">{log.geo_location || "Unknown"}</span>
                          <span className="text-blue-500/80 font-mono text-[10px] uppercase font-bold">{log.target_domain || "ALL_WORKSPACES"}</span>
                        </div>
                      </td>

                      <td className="py-4 px-6">
                        <div className="flex flex-col gap-0.5 max-w-[250px]">
                          <span className="text-gray-200 text-sm truncate font-mono bg-black/50 px-2 py-0.5 border border-white/5 inline-block w-fit rounded">{log.endpoint}</span>
                          <span className="text-[10px] text-gray-500 mt-1 italic">
                            {log.device_fingerprint || "OS Profiling Hidden"}
                          </span>
                        </div>
                      </td>

                      <td className="py-4 px-6 text-center">
                        {isHacker ? (
                          <div className="inline-flex flex-col items-center gap-1">
                            <span className="text-[10px] font-bold tracking-widest px-2.5 py-1 rounded-sm bg-red-500/20 text-red-500 border border-red-500/30">
                              HACKER
                            </span>
                            <span className="text-[9px] text-red-400/80 mt-0.5 italic">
                              {log.status === "HONEYPOT_REDIRECTED" ? `TARPIT (${log.latency_ms}ms)` : 'BLOCKED'}
                            </span>
                          </div>
                        ) : (
                          <span className="text-[10px] font-bold tracking-widest px-2.5 py-1 rounded-sm bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                            USER
                          </span>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </section>
      </main >

      {/* 🤖 NECHAT WIDGET */}
      < NechatWidget activeDomain={activeDomain} />

      {/* ➕ ADD ROUTE MODAL */}
      < AddRouteModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={() => {
          // Success handler
        }}
      />
    </div >
  )
}

export default SOCDashboard
