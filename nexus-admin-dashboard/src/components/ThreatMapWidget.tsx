"use client";

import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Activity, Globe, Crosshair } from 'lucide-react';

interface Threat {
    id: string;
    attacker_ip: string;
    source_lat: number;
    source_lng: number;
    target_lat: number;
    target_lng: number;
    type: string;
}

// Global Map Constants
const MAP_WIDTH = 800;
const MAP_HEIGHT = 400;

// 🔵 PROTECTED SENTINEL NODES (Where your websites live)
const PROTECTED_NODES = [
    { name: "JW_JAKARTA", lat: -6.20, lng: 106.81, label: "ojk.go.id" },
    { name: "JW_SINGAPORE", lat: 1.35, lng: 103.81, label: "portal-pns" },
    { name: "JW_SYDNEY", lat: -33.86, lng: 151.20, label: "audit-hub" },
    { name: "JW_FRANKFURT", lat: 50.11, lng: 8.68, label: "cloud-storage" },
];

export default function ThreatMapWidget() {
    const [threats, setThreats] = useState<Threat[]>([]);
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
        const eventSource = new EventSource('http://localhost:8080/api/stream/threats');
        eventSource.onmessage = (event) => {
            try {
                const newThreat = JSON.parse(event.data);
                if (newThreat.attacker_ip) {
                    setThreats(prev => [newThreat, ...prev].slice(0, 10));
                }
            } catch (e) { }
        };
        return () => eventSource.close();
    }, []);

    const project = (lat: number, lng: number) => {
        const x = (lng + 180) * (MAP_WIDTH / 360);
        const y = (90 - lat) * (MAP_HEIGHT / 180);
        return { x, y };
    };

    if (!mounted) return null;

    return (
        <div className="relative w-full h-full bg-[#030507] rounded-xl border border-blue-900/30 overflow-hidden shadow-2xl flex font-sans">
            {/* 🗺️ WORLD SHIELD DISPLAY */}
            <div className="flex-1 relative bg-[radial-gradient(circle_at_center,_#0a192f_0%,_#030507_100%)] overflow-hidden">
                <svg viewBox={`0 0 ${MAP_WIDTH} ${MAP_HEIGHT}`} className="w-full h-full p-6">
                    {/* WORLD GRID */}
                    {[...Array(12)].map((_, i) => (
                        <line key={`v-${i}`} x1={(i + 1) * (MAP_WIDTH / 13)} y1="0" x2={(i + 1) * (MAP_WIDTH / 13)} y2={MAP_HEIGHT} stroke="rgba(6,182,212,0.05)" strokeWidth="1" />
                    ))}
                    {[...Array(8)].map((_, i) => (
                        <line key={`h-${i}`} x1="0" y1={(i + 1) * (MAP_HEIGHT / 9)} x2={MAP_WIDTH} y2={(i + 1) * (MAP_HEIGHT / 9)} stroke="rgba(6,182,212,0.05)" strokeWidth="1" />
                    ))}

                    {/* 🔵 BLUE SENTINELS (Protected Sites) */}
                    {PROTECTED_NODES.map((node, i) => {
                        const { x, y } = project(node.lat, node.lng);
                        return (
                            <g key={`sentinel-${i}`}>
                                <circle cx={x} cy={y} r="5" fill="#3b82f6" className="animate-pulse shadow-[0_0_15px_#3b82f6]" />
                                <circle cx={x} cy={y} r="12" fill="none" stroke="#3b82f6" strokeWidth="0.5" strokeDasharray="2 2">
                                    <animateTransform attributeName="transform" type="rotate" from={`0 ${x} ${y}`} to={`360 ${x} ${y}`} dur="10s" repeatCount="indefinite" />
                                </circle>
                                <text x={x + 10} y={y + 4} fill="#60a5fa" fontSize="8" className="font-mono font-bold uppercase tracking-tighter opacity-80">{node.label}</text>
                            </g>
                        );
                    })}

                    {/* 🔴 ATTACK ARCS */}
                    <AnimatePresence>
                        {threats.map((threat, idx) => {
                            const start = project(threat.source_lat, threat.source_lng);
                            const end = project(threat.target_lat, threat.target_lng);
                            const midX = (start.x + end.x) / 2;
                            const midY = Math.min(start.y, end.y) - 60;

                            return (
                                <g key={threat.id + idx}>
                                    <motion.circle initial={{ scale: 0 }} animate={{ scale: 1 }} cx={start.x} cy={start.y} r="3" fill="#ef4444" />
                                    <motion.path
                                        d={`M ${start.x} ${start.y} Q ${midX} ${midY} ${end.x} ${end.y}`}
                                        fill="none" stroke="rgba(239, 68, 68, 0.6)" strokeWidth="1.5"
                                        initial={{ pathLength: 0 }} animate={{ pathLength: 1 }} transition={{ duration: 1.5 }}
                                    />
                                    <motion.circle cx={end.x} cy={end.y} r="15" fill="rgba(239, 68, 68, 0.2)"
                                        animate={{ scale: [1, 2, 1], opacity: [0.5, 0, 0.5] }} transition={{ duration: 1.5, repeat: Infinity }}
                                    />
                                    <motion.circle r="3" fill="#ef4444">
                                        <animateMotion path={`M ${start.x} ${start.y} Q ${midX} ${midY} ${end.x} ${end.y}`} dur="1.5s" repeatCount="indefinite" />
                                    </motion.circle>
                                </g>
                            );
                        })}
                    </AnimatePresence>
                </svg>

                {/* UI OVERLAY */}
                <div className="absolute top-6 left-6">
                    <div className="flex items-center gap-3 bg-blue-900/10 border border-blue-500/20 px-4 py-2 rounded-lg backdrop-blur-xl">
                        <Globe className="w-5 h-5 text-cyan-400 animate-spin-slow" />
                        <div>
                            <h4 className="text-[10px] font-black tracking-[0.2em] text-cyan-400">GLOBAL_SHIELD_V12</h4>
                            <p className="text-[8px] text-slate-500 font-mono italic">Nexus Autonomous Sovereignty Active</p>
                        </div>
                    </div>
                </div>
            </div>

            {/* 📜 SIDEBAR LOG [V12] */}
            <div className="w-72 border-l border-blue-900/20 flex flex-col bg-[#07090c]/90 backdrop-blur-2xl">
                <div className="p-4 border-b border-blue-900/20 bg-blue-950/20 flex items-center justify-between">
                    <h3 className="text-[10px] font-bold text-slate-400 tracking-widest flex items-center gap-2">
                        <Activity className="w-3 h-3 text-cyan-500" /> VECTORS_LIVE
                    </h3>
                    <span className="text-[8px] bg-red-500/20 text-red-500 px-1.5 py-0.5 rounded font-black animate-pulse">WARGAME</span>
                </div>
                <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar">
                    {threats.map((t, i) => (
                        <motion.div initial={{ x: 20, opacity: 0 }} animate={{ x: 0, opacity: 1 }} key={i} className="p-3 bg-red-950/5 border-l-2 border-red-500/40 rounded-r shadow-sm flex flex-col gap-1.5">
                            <div className="flex justify-between items-center">
                                <span className="text-[10px] font-mono font-bold text-red-400 tracking-tighter">{t.attacker_ip}</span>
                                <span className="text-[8px] px-1 bg-red-500/10 text-red-400/80 rounded uppercase font-bold">{t.type.split('_').pop()}</span>
                            </div>
                            <div className="flex items-center gap-2 text-[9px] text-slate-400">
                                <Crosshair className="w-3 h-3 text-cyan-500" />
                                <span>TARGET: {PROTECTED_NODES.find(n => n.lat === t.target_lat)?.label || "nexus-core"}</span>
                            </div>
                            <div className="text-[9px] font-black text-slate-500 tracking-wider">
                                {t.type.split('_')[0]} ATTEMPT • BLOCKING_...
                            </div>
                        </motion.div>
                    ))}
                    {threats.length === 0 && <div className="text-[9px] text-slate-600 font-mono text-center h-full flex items-center justify-center italic opacity-50">Monitoring world coordinates...</div>}
                </div>
                <div className="p-3 bg-black/40 border-t border-blue-900/10">
                    <div className="flex items-center justify-between text-[8px] font-bold text-cyan-500/40">
                        <span>SENTINELS: 04 ONLINE</span>
                        <span className="flex items-center gap-1"><span className="w-1 h-1 bg-emerald-500 rounded-full"></span>SECURE</span>
                    </div>
                </div>
            </div>

            <style jsx>{`
                @keyframes spin-slow { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
                .animate-spin-slow { animation: spin-slow 20s linear infinite; }
            `}</style>
        </div>
    );
}
