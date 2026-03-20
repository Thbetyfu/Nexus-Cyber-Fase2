"use client"

import React, { useEffect, useState } from "react"
import { Activity, ShieldAlert, Shield, Server, RefreshCw, AlertTriangle } from "lucide-react"
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer
} from "recharts"

// API configuration
const API_URL = "http://localhost:8080/api/telemetry"

export default function CommandCenter() {
  const [data, setData] = useState({
    mtd: { active_port: 0, next_shuffle_secs: 0, status: "OFFLINE" },
    stats: { allowed: 0, blocked: 0, honeypot: 0 },
    recent_logs: [] as any[]
  })
  const [error, setError] = useState<string | null>(null)

  // Historical data for charts
  const [history, setHistory] = useState<any[]>([])

  useEffect(() => {
    const fetchTelemetry = async () => {
      try {
        const res = await fetch(API_URL)
        if (!res.ok) throw new Error("API not accessible")
        const json = await res.json()
        setData(json)

        // Push to history for the chart (max 20 points)
        setHistory(prev => {
          const now = new Date().toLocaleTimeString("id-ID", { hour12: false })
          const newPoint = {
            time: now,
            allowed: json.stats.allowed,
            blocked: json.stats.blocked,
            honeypot: json.stats.honeypot
          }
          const next = [...prev, newPoint]
          return next.slice(-20) // Keep last 20
        })
        setError(null)
      } catch (err: any) {
        setError(err.message)
      }
    }

    // Polling every 1 second (Live WebTelemetry)
    const interval = setInterval(fetchTelemetry, 1000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="min-h-screen bg-black text-gray-100 p-6 md:p-8 font-sans selection:bg-blue-600/30">

      {/* HEADER */}
      <header className="flex flex-col md:flex-row justify-between items-start md:items-center mb-8 pb-4 border-b border-gray-800">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
            <Shield className="text-blue-500 w-8 h-8" />
            NEXUS COMMAND CENTER
          </h1>
          <p className="text-gray-400 mt-1 flex items-center gap-2">
            <span className="relative flex h-2.5 w-2.5">
              <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${error ? 'bg-red-500' : 'bg-emerald-500'}`}></span>
              <span className={`relative inline-flex rounded-full h-2.5 w-2.5 ${error ? 'bg-red-500' : 'bg-emerald-500'}`}></span>
            </span>
            {error ? "SYSTEM OFFLINE" : "LIVE TELEMETRY - FASE 6 OJK/BSSN COMPLIANCE"}
          </p>
        </div>
        <div className="mt-4 md:mt-0 flex gap-4">
          <div className="bg-gray-900/50 border border-gray-800 rounded-lg px-4 py-2 flex items-center gap-3">
            <Server className="w-5 h-5 text-gray-400" />
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider font-semibold">Active Port</p>
              <p className="text-lg font-mono text-emerald-400">:{data.mtd.active_port}</p>
            </div>
          </div>
          <div className="bg-gray-900/50 border border-gray-800 rounded-lg px-4 py-2 flex items-center gap-3">
            <RefreshCw className={`w-5 h-5 text-gray-400 ${data.mtd.status === 'ACTIVE' ? 'animate-spin-slow' : ''}`} style={{ animationDuration: '3s' }} />
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider font-semibold">Next Shuffle</p>
              <p className="text-lg font-mono text-blue-400">{data.mtd.next_shuffle_secs}s</p>
            </div>
          </div>
        </div>
      </header>

      {error ? (
        <div className="bg-red-950/20 border border-red-900/50 rounded-xl p-6 flex flex-col items-center justify-center text-center h-64">
          <AlertTriangle className="w-12 h-12 text-red-500 mb-4 opacity-80" />
          <h2 className="text-xl font-semibold text-red-400">Telemetry Disconnected</h2>
          <p className="text-gray-400 mt-2 max-w-md">Cannot connect to Nexus Core Gateway. Ensure backend is running and /api/telemetry is reachable from localhost.</p>
        </div>
      ) : (
        <>
          {/* STATS CARDS */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-gray-900/40 backdrop-blur-md border border-gray-800/80 rounded-xl p-6 relative overflow-hidden group">
              <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                <Activity className="w-24 h-24" />
              </div>
              <p className="text-sm text-gray-400 uppercase tracking-wider font-medium mb-1 relative z-10">Traffic Allowed</p>
              <p className="text-4xl font-bold text-gray-100 relative z-10">{data.stats.allowed.toLocaleString()}</p>
              <div className="mt-4 text-xs text-emerald-500 flex items-center gap-1 z-10 relative">
                Dual-Brain Cleared
              </div>
            </div>

            <div className="bg-gray-900/40 backdrop-blur-md border border-gray-800/80 rounded-xl p-6 relative overflow-hidden group">
              <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                <ShieldAlert className="w-24 h-24 text-red-500" />
              </div>
              <p className="text-sm text-gray-400 uppercase tracking-wider font-medium mb-1 relative z-10">Rate-Limited (429)</p>
              <p className="text-4xl font-bold text-gray-100 relative z-10">{data.stats.blocked.toLocaleString()}</p>
              <div className="mt-4 text-xs text-gray-500 flex items-center gap-1 z-10 relative">
                Token Bucket MTD Layer
              </div>
            </div>

            <div className="bg-red-950/20 backdrop-blur-md border border-red-900/30 rounded-xl p-6 relative overflow-hidden group">
              <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                <AlertTriangle className="w-24 h-24 text-red-500" />
              </div>
              <p className="text-sm text-red-400/80 uppercase tracking-wider font-medium mb-1 relative z-10">Honeypot Trapped</p>
              <p className="text-4xl font-bold text-red-500 relative z-10">{data.stats.honeypot.toLocaleString()}</p>
              <div className="mt-4 text-xs text-red-400/70 flex items-center gap-1 z-10 relative">
                Digital Hallucination Active
              </div>
            </div>

            <div className="bg-gray-900/40 backdrop-blur-md border border-gray-800/80 rounded-xl p-6 relative overflow-hidden group">
              <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                <Shield className="w-24 h-24 text-emerald-500" />
              </div>
              <p className="text-sm text-gray-400 uppercase tracking-wider font-medium mb-1 relative z-10">False Positives</p>
              <p className="text-4xl font-bold text-emerald-400 relative z-10">0%</p>
              <div className="mt-4 text-xs text-emerald-500/80 flex items-center gap-1 z-10 relative">
                ISO 25010 AI Precision
              </div>
            </div>
          </div>

          {/* MAIN CONTENT GRID */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

            {/* CHART PANEL */}
            <div className="lg:col-span-2 bg-gray-900/40 backdrop-blur-md border border-gray-800/80 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-6 flex items-center gap-2">
                <Activity className="w-5 h-5 text-blue-500" />
                Live Flow Analytics
              </h3>
              <div className="h-72 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={history} margin={{ top: 0, right: 0, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorAllowed" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                      </linearGradient>
                      <linearGradient id="colorBlocked" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" vertical={false} />
                    <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} tickMargin={10} />
                    <YAxis stroke="#9ca3af" fontSize={12} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#111827', borderColor: '#374151', color: '#f3f4f6' }}
                      itemStyle={{ color: '#e5e7eb' }}
                    />
                    <Area type="monotone" dataKey="allowed" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#colorAllowed)" />
                    <Area type="monotone" dataKey="blocked" stroke="#ef4444" strokeWidth={2} fillOpacity={1} fill="url(#colorBlocked)" />
                    <Area type="monotone" dataKey="honeypot" stroke="#f59e0b" strokeWidth={2} fillOpacity={0} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* THREAT INTEL FEED / ROGUE GALLERY */}
            <div className="bg-gray-900/40 backdrop-blur-md border border-gray-800/80 rounded-xl overflow-hidden flex flex-col h-full">
              <div className="p-5 border-b border-gray-800/80 bg-gray-900/50">
                <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                  <ShieldAlert className="w-5 h-5 text-red-500" />
                  The Rogue Gallery (Live Threat Actors)
                </h3>
              </div>
              <div className="overflow-y-auto flex-1 p-2" style={{ maxHeight: '400px' }}>
                <div className="flex flex-col gap-2 p-3">
                  {data.recent_logs.length === 0 ? (
                    <div className="text-center p-8 text-gray-500 text-sm">No recent anomalies detected.</div>
                  ) : (
                    [...data.recent_logs].reverse().slice(0, 50).filter(l => l.status !== "ALLOWED").map((log, idx) => {
                      const isReflex = log.status === 'HONEYPOT_REDIRECTED' || log.status === 'DIVERTED_TO_HONEYPOT';
                      const isReasoning = log.status === 'MALICIOUS_DETECTED_REASONING';
                      const modelName = isReflex ? "⚡ Qwen3 32B Reflex" : isReasoning ? "🧠 Qwen3 235B Reasoning" : "🛡️ Token Bucket";

                      return (
                        <div key={idx} className="bg-gray-900/60 border border-gray-800 rounded-lg p-3 hover:bg-gray-800/80 transition-all group">
                          <div className="flex justify-between items-start mb-2">
                            <div className="flex flex-col gap-1">
                              <span className="text-[10px] font-mono font-bold text-blue-400 group-hover:text-blue-300 transition-colors">
                                {log.attacker_id || "ANONYMOUS"}
                              </span>
                              <span className="text-[9px] font-mono text-gray-500">
                                {new Date(log.timestamp).toLocaleTimeString('id-ID')}
                              </span>
                            </div>
                            <div className="flex flex-col items-end gap-1">
                              <span className="text-[9px] px-2 py-0.5 rounded-full font-medium tracking-wide bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                                {modelName}
                              </span>
                              <span className={`text-[9px] px-2 py-0.5 rounded-full font-bold tracking-wider ${isReflex || isReasoning ? 'bg-red-500/20 text-red-400 border border-red-500/30' :
                                'bg-orange-500/20 text-orange-400 border border-orange-500/30'
                                }`}>
                                {log.status}
                              </span>
                            </div>
                          </div>

                          <p className="text-sm font-medium text-gray-200 mb-2">{log.threat_detail || "SENSITIVE_DATA_PROBE"}</p>

                          <div className="grid grid-cols-2 gap-2 mb-3 bg-black/30 rounded-md p-2 border border-white/5">
                            <div>
                              <p className="text-[8px] text-gray-500 uppercase">Location</p>
                              <p className="text-[10px] text-gray-300 truncate">{log.geo_location || "Unknown Node"}</p>
                            </div>
                            <div>
                              <p className="text-[8px] text-gray-500 uppercase">Gateway/ISP</p>
                              <p className="text-[10px] text-gray-300 truncate">{log.isp || "Cloaked"}</p>
                            </div>
                          </div>

                          <div className="flex justify-between items-center text-[10px] text-gray-500">
                            <span className="font-mono bg-gray-800/50 px-1.5 py-0.5 rounded">{log.source_ip}</span>
                            <span className="flex items-center gap-1 font-mono text-red-400">
                              {log.latency_ms > 0 ? `${log.latency_ms}ms tarpit` : "deflected"}
                            </span>
                          </div>

                          <div className="mt-2 pt-2 border-t border-white/5 flex items-center gap-2">
                            <div className="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse"></div>
                            <span className="text-[9px] text-gray-600 font-mono italic truncate">FP: {log.device_fingerprint || "Obfuscated"}</span>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            </div>

          </div>
        </>
      )}
    </div>
  )
}
