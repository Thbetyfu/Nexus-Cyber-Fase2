"use client"

import React, { useEffect, useState, useRef } from "react"
import { Shield, Activity, ShieldAlert, Crosshair, Server, Lock } from "lucide-react"
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as ChartTooltip, ResponsiveContainer
} from "recharts"
import { TelemetryLog, generateMockLogs } from "../lib/mock_telemetry"

// Switch to false when backend is ready
const USE_MOCK_DATA = true;

const SOCDashboard = () => {
  const [logs, setLogs] = useState<TelemetryLog[]>([])
  const [history, setHistory] = useState<any[]>([])

  // Dashboard Metrics
  const metrics = {
    allowed: logs.filter(l => l.status === "ALLOWED").length * 84, // mock multiplier
    blocked: logs.filter(l => l.status === "RATE_LIMITED").length * 27,
    honeypot: logs.filter(l => l.status === "HONEYPOT_REDIRECTED").length * 15,
    falsePositives: 0
  }

  const tableContainerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    // Initial Load
    setLogs(generateMockLogs(20));
    setHistory(Array.from({ length: 20 }).map((_, i) => ({
      time: `T-${20 - i}`,
      allowed: Math.floor(Math.random() * 50) + 10,
      honeypot: Math.floor(Math.random() * 5)
    })))

    // Polling Interval
    const fetchInterval = setInterval(() => {
      // Create new mock log
      const newLogs = generateMockLogs(1)

      setLogs((prev) => {
        const next = [newLogs[0], ...prev] // Put newest at top
        return next.slice(0, 100)
      })

      // Update Chart
      const now = new Date().toLocaleTimeString("id-ID", { hour12: false })
      setHistory((prev) => {
        const h = {
          time: now,
          allowed: Math.floor(Math.random() * 20),
          honeypot: newLogs[0].status === "HONEYPOT_REDIRECTED" ? 10 : 0
        }
        return [...prev, h].slice(-20)
      })

    }, 1500)

    return () => clearInterval(fetchInterval)
  }, [])

  return (
    <div className="min-h-screen bg-[#06080b] text-gray-200 font-sans selection:bg-blue-600/30 flex flex-col">
      {/* 🛡️ SOC HEADER */}
      <header className="flex items-center justify-between px-6 py-4 border-b border-gray-800/80 bg-[#090b0e] shadow-md z-10 sticky top-0">
        <div className="flex items-center gap-4">
          <div className="bg-emerald-500/10 p-2 rounded-lg border border-emerald-500/20 shadow-[0_0_15px_rgba(16,185,129,0.15)]">
            <Shield className="w-8 h-8 text-emerald-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white/90">NEXUS CYBER SOC</h1>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 bg-emerald-500"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              <p className="text-xs text-emerald-500/80 uppercase tracking-widest font-semibold">Active Monitoring • MTD Matrix v5.2</p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 border border-blue-900/40 bg-blue-950/20 rounded px-3 py-1.5 shadow-inner">
            <Server className="w-4 h-4 text-blue-500" />
            <span className="text-[11px] font-mono text-blue-300">SHUFFLER_PORT: :3001</span>
          </div>
        </div>
      </header>

      {/* 🚀 MAIN CONTENT */}
      <main className="flex-1 flex flex-col p-6 gap-6 overflow-hidden max-w-screen-2xl mx-auto w-full">

        {/* UPPER SECTION: Live Flow Metrics & Analytics */}
        <section className="grid grid-cols-1 lg:grid-cols-4 gap-6 shrink-0">

          {/* Metrics Column */}
          <div className="lg:col-span-1 flex flex-col gap-4">
            <div className="bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl p-5 shadow-lg group hover:border-blue-900/50 transition duration-300">
              <div className="flex justify-between items-start mb-2">
                <p className="text-xs text-gray-500 uppercase tracking-widest font-bold">Total Traffic (Safe)</p>
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

            <div className="bg-[#09110d]/80 backdrop-blur border border-emerald-900/30 rounded-xl p-5 shadow-lg group z-10 hover:border-emerald-900/50 transition">
              <div className="flex justify-between items-start mb-2">
                <p className="text-xs text-emerald-500/80 uppercase tracking-widest font-bold">False Positives</p>
                <Crosshair className="w-4 h-4 text-emerald-500/50" />
              </div>
              <p className="text-3xl font-bold text-emerald-400 font-mono tracking-tight">0%</p>
            </div>
          </div>

          {/* Chart Core Center */}
          <div className="lg:col-span-3 bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl p-5 flex flex-col shadow-lg">
            <h3 className="text-xs font-semibold uppercase tracking-widest text-gray-500 mb-4 flex items-center gap-2">
              <Activity className="w-4 h-4 text-blue-500" /> Real-time Traffic Vectors
            </h3>
            <div className="w-full flex-1 min-h-[180px]">
              <ResponsiveContainer width="100%" height="100%">
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
                  <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" vertical={false} />
                  <XAxis dataKey="time" stroke="#6b7280" fontSize={11} tickMargin={12} axisLine={false} tickLine={false} />
                  <YAxis stroke="#6b7280" fontSize={11} axisLine={false} tickLine={false} />
                  <ChartTooltip
                    contentStyle={{ backgroundColor: '#030712', borderColor: '#1f2937', color: '#f3f4f6', borderRadius: '8px', fontSize: '12px' }}
                    itemStyle={{ color: '#e5e7eb' }}
                  />
                  <Area type="monotone" dataKey="allowed" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#colorSafe)" isAnimationActive={false} />
                  <Area type="monotone" dataKey="honeypot" stroke="#ef4444" strokeWidth={2} fillOpacity={1} fill="url(#colorThreat)" isAnimationActive={false} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </section>

        {/* LOWER SECTION: Real-Time Access & Threat Log */}
        <section className="flex-1 flex flex-col bg-[#0c0f14]/80 backdrop-blur border border-gray-800/60 rounded-xl overflow-hidden shadow-lg min-h-[400px]">
          <div className="bg-[#090b0e] px-5 py-3 border-b border-gray-800/80 flex items-center justify-between sticky top-0 z-20 shadow-sm">
            <h3 className="text-sm font-semibold text-gray-200 uppercase tracking-wider flex items-center gap-2">
              <Lock className="w-4 h-4 text-emerald-500" />
              The Rogue Gallery: Access & Forensic Log
            </h3>
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
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[20%]">Geo-Location & ISP</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 w-[25%]">Target Endpoint / Info</th>
                  <th className="py-4 px-6 font-semibold border-b border-gray-800 text-center w-[15%]">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800/50">
                {logs.map((log, idx) => {
                  const isHacker = log.status !== "ALLOWED";
                  const rowClass = isHacker
                    ? "hover:bg-red-950/20 transition-colors bg-red-950/10"
                    : "hover:bg-blue-900/10 transition-colors";
                  return (
                    <tr key={idx} className={`group ${rowClass}`}>
                      {/* Column: Timestamp (Monospace) */}
                      <td className="py-4 px-6">
                        <span className="font-mono text-gray-400 text-xs var(--font-fira-code)">
                          {new Date(log.timestamp).toLocaleTimeString("id-ID", { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                        </span>
                      </td>

                      {/* Column: IP & ID (Monospace) */}
                      <td className="py-4 px-6">
                        <div className="flex flex-col">
                          <span className="font-mono text-gray-300 text-sm var(--font-fira-code)">{log.source_ip}</span>
                          {log.attacker_id && (
                            <span className="font-mono text-[10px] text-red-400/80 mt-1 uppercase var(--font-fira-code)">
                              {log.attacker_id}
                            </span>
                          )}
                        </div>
                      </td>

                      {/* Column: Location */}
                      <td className="py-4 px-6">
                        <div className="flex flex-col gap-0.5">
                          <span className="text-gray-300 text-sm">{log.geo_location || "Unknown"}</span>
                          <span className="text-gray-500 text-[10px] uppercase truncate">{log.isp || "Cloaked"}</span>
                        </div>
                      </td>

                      {/* Column: Endpoint */}
                      <td className="py-4 px-6">
                        <div className="flex flex-col gap-0.5 max-w-[250px]">
                          <span className="text-gray-200 text-sm truncate font-mono bg-black/50 px-2 py-0.5 border border-white/5 inline-block w-fit rounded">{log.endpoint}</span>
                          <span className="text-[10px] text-gray-500 mt-1 italic">
                            {log.device_fingerprint || "OS Profiling Hidden"}
                          </span>
                        </div>
                      </td>

                      {/* Column: Status */}
                      <td className="py-4 px-6 text-center">
                        {isHacker ? (
                          <div className="inline-flex flex-col items-center gap-1">
                            <span className="text-[10px] font-bold tracking-widest px-2.5 py-1 rounded-sm bg-red-500/20 text-red-500 border border-red-500/30">
                              HACKER
                            </span>
                            <span className="text-[9px] text-red-400/80 mt-0.5">
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

      </main>
    </div>
  )
}

export default SOCDashboard
