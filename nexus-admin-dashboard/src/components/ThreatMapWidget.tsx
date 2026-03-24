"use client";

import React, { useEffect, useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Activity, Globe, Crosshair, AlertTriangle, RefreshCw } from 'lucide-react';

interface Threat {
    id: string;
    attacker_ip: string;
    source_lat: number;
    source_lng: number;
    target_lat: number;
    target_lng: number;
    type: string;
    status?: string; // ALLOWED, BLOCKED, BREACHED
}

interface SentinelNode {
    id: string;
    name: string;
    lat: number;
    lng: number;
    label: string;
    status: 'online' | 'under_attack' | 'repairing' | 'breached';
    lastAttackTime: number;
}

const MAP_WIDTH = 800;
const MAP_HEIGHT = 450;

// 🔵 INITIAL SENTINEL HUB (Fixed Locations)
const INITIAL_SENTINELS: SentinelNode[] = [
    { id: "ojk", name: "JW_JAKARTA", lat: -6.20, lng: 106.81, label: "ojk.go.id (Database)", status: 'online', lastAttackTime: 0 },
    { id: "pns", name: "JW_SINGAPORE", lat: 1.35, lng: 103.81, label: "portal-nexus", status: 'online', lastAttackTime: 0 },
    { id: "audit", name: "JW_SYDNEY", lat: -33.86, lng: 151.20, label: "audit-hub", status: 'online', lastAttackTime: 0 },
    { id: "cloud", name: "JW_FRANKFURT", lat: 50.11, lng: 8.68, label: "cloud-storage", status: 'online', lastAttackTime: 0 },
];

export default function ThreatMapWidget() {
    const [threats, setThreats] = useState<Threat[]>([]);
    const [sentinels, setSentinels] = useState<SentinelNode[]>(INITIAL_SENTINELS);
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
        const eventSource = new EventSource('http://localhost:8081/api/stream/threats');

        eventSource.onmessage = (event) => {
            try {
                const newThreat = JSON.parse(event.data);
                if (!newThreat.attacker_ip) return;

                setThreats(prev => [newThreat, ...prev].slice(0, 15));

                // 🎯 Update Sentinel Status based on attack target
                setSentinels(prev => prev.map(node => {
                    // Match coordinates to identify which sentinel is attacked
                    const isTarget = Math.abs(node.lat - newThreat.target_lat) < 0.1 &&
                        Math.abs(node.lng - newThreat.target_lng) < 0.1;

                    if (isTarget) {
                        const newStatus = (newThreat.type.includes("CRITICAL") || newThreat.status === "BREACHED") ? 'breached' : 'under_attack';
                        return { ...node, status: newStatus, lastAttackTime: Date.now() };
                    }
                    return node;
                }));
            } catch (e) { }
        };

        return () => eventSource.close();
    }, []);

    // ♻️ AUTO-REPAIR LOGIC (Reset node status after clear time)
    useEffect(() => {
        const repairInterval = setInterval(() => {
            const now = Date.now();
            setSentinels(prev => prev.map(node => {
                if (node.status === 'under_attack' && now - node.lastAttackTime > 5000) {
                    return { ...node, status: 'repairing' };
                }
                if (node.status === 'repairing' && now - node.lastAttackTime > 8000) {
                    return { ...node, status: 'online' };
                }
                // Breached nodes stay grey longer to show impact
                if (node.status === 'breached' && now - node.lastAttackTime > 15000) {
                    return { ...node, status: 'repairing' };
                }
                return node;
            }));
        }, 1000);
        return () => clearInterval(repairInterval);
    }, []);

    const project = (lat: number, lng: number) => {
        const x = (lng + 180) * (MAP_WIDTH / 360);
        const y = (90 - lat) * (MAP_HEIGHT / 180);
        return { x, y };
    };

    if (!mounted) return null;

    return (
        <div className="relative w-full h-full bg-[#030507] rounded-xl border border-blue-900/30 overflow-hidden shadow-2xl flex font-sans">
            {/* 🛠️ TACTICAL RADAR LAYER */}
            <div className="flex-1 relative bg-[radial-gradient(circle_at_center,_#0a192f_0%,_#030507_100%)] overflow-hidden">
                <svg viewBox={`0 0 ${MAP_WIDTH} ${MAP_HEIGHT}`} className="w-full h-full p-4">
                    {/* RADAR GRIDS */}
                    {[1, 2, 3].map(i => (
                        <circle key={`radar-${i}`} cx={MAP_WIDTH / 2} cy={MAP_HEIGHT / 2} r={100 * i} fill="none" stroke="rgba(6,182,212,0.05)" strokeWidth="0.5" />
                    ))}
                    <line x1={MAP_WIDTH / 2} y1="0" x2={MAP_WIDTH / 2} y2={MAP_HEIGHT} stroke="rgba(6,182,212,0.05)" strokeWidth="0.5" />
                    <line x1="0" y1={MAP_HEIGHT / 2} x2={MAP_WIDTH} y2={MAP_HEIGHT / 2} stroke="rgba(6,182,212,0.05)" strokeWidth="0.5" />

                    {/* 🛡️ SENTINEL NODES (Fixed) */}
                    {sentinels.map((node, i) => {
                        const { x, y } = project(node.lat, node.lng);
                        const isBreached = node.status === 'breached';
                        const isAttacked = node.status === 'under_attack';
                        const isRepairing = node.status === 'repairing';

                        return (
                            <g key={`node-${node.id}`}>
                                {/* Status Ring */}
                                <motion.circle
                                    cx={x} cy={y} r="20"
                                    fill="none"
                                    stroke={isBreached ? "#4b5563" : (isAttacked ? "#ef4444" : "#3b82f6")}
                                    strokeWidth="1"
                                    strokeDasharray="4 4"
                                    animate={{ rotate: 360, opacity: [0.2, 0.5, 0.2] }}
                                    transition={{ duration: 5, repeat: Infinity, ease: "linear" }}
                                />

                                {/* Core Node */}
                                <circle
                                    cx={x} cy={y} r="6"
                                    fill={isBreached ? "#4b5563" : (isAttacked ? "#ef4444" : (isRepairing ? "#fbbf24" : "#3b82f6"))}
                                    className={isAttacked ? "animate-ping" : ""}
                                />

                                {isRepairing && (
                                    <motion.g animate={{ rotate: 360 }} transition={{ duration: 2, repeat: Infinity, ease: "linear" }} style={{ originX: x, originY: y }}>
                                        <RefreshCw cx={x - 15} cy={y - 15} size={10} className="text-yellow-500 opacity-50" />
                                    </motion.g>
                                )}

                                <text x={x + 12} y={y + 4} fill={isBreached ? "#9ca3af" : "#60a5fa"} fontSize="9" className="font-mono font-bold tracking-tighter uppercase">
                                    {node.label}
                                </text>
                                {isBreached && <text x={x + 12} y={y + 14} fill="#ef4444" fontSize="7" className="font-black animate-pulse">! BREACH DETECTED</text>}
                            </g>
                        );
                    })}

                    {/* ⚔️ LIVE ATTACK VECTORS */}
                    <AnimatePresence>
                        {threats.map((threat, idx) => {
                            const start = project(threat.source_lat, threat.source_lng);
                            const end = project(threat.target_lat, threat.target_lng);
                            const midX = (start.x + end.x) / 2;
                            const midY = Math.min(start.y, end.y) - 50;

                            return (
                                <g key={threat.id || `trt-${idx}`}>
                                    {/* 🔴 Attack Source (Red Dot) */}
                                    <motion.g initial={{ scale: 0, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0, opacity: 0 }}>
                                        <circle cx={start.x} cy={start.y} r="4" fill="#ef4444" />
                                        <circle cx={start.x} cy={start.y} r="10" fill="none" stroke="#ef4444" strokeWidth="0.5" className="animate-ping" />
                                        <text x={start.x + 8} y={start.y - 5} fill="#ef4444" fontSize="8" className="font-mono font-bold">{threat.attacker_ip}</text>
                                    </motion.g>

                                    {/* ⚡ Attack Vector Line */}
                                    <motion.path
                                        d={`M ${start.x} ${start.y} Q ${midX} ${midY} ${end.x} ${end.y}`}
                                        fill="none"
                                        stroke={threat.id?.includes("CRITICAL") ? "#ef4444" : "rgba(239, 68, 68, 0.4)"}
                                        strokeWidth="1.5"
                                        initial={{ pathLength: 0 }}
                                        animate={{ pathLength: 1 }}
                                        exit={{ opacity: 0 }}
                                        transition={{ duration: 1.2, ease: "easeOut" }}
                                    />

                                    {/* Impact Pulse */}
                                    <motion.circle
                                        cx={end.x} cy={end.y} r="15"
                                        fill="rgba(239, 68, 68, 0.15)"
                                        initial={{ scale: 0 }}
                                        animate={{ scale: [1, 2.5, 1], opacity: [0.6, 0, 0.6] }}
                                        transition={{ duration: 1.5, repeat: Infinity }}
                                    />
                                </g>
                            );
                        })}
                    </AnimatePresence>
                </svg>

                {/* 🏮 WORLD OVERLAY LEGEND */}
                <div className="absolute top-4 left-4 flex flex-col gap-2">
                    <div className="flex items-center gap-2 bg-black/60 border border-blue-900/30 px-3 py-1.5 rounded-lg backdrop-blur-md">
                        <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                        <span className="text-[10px] font-bold text-blue-400 uppercase tracking-widest">Sentinel Online</span>
                    </div>
                    <div className="flex items-center gap-2 bg-black/60 border border-red-900/30 px-3 py-1.5 rounded-lg backdrop-blur-md">
                        <div className="w-2 h-2 rounded-full bg-red-500 animate-ping"></div>
                        <span className="text-[10px] font-bold text-red-500 uppercase tracking-widest">Active Siege</span>
                    </div>
                    <div className="flex items-center gap-2 bg-black/60 border border-gray-600/30 px-3 py-1.5 rounded-lg backdrop-blur-md">
                        <div className="w-2 h-2 rounded-full bg-gray-500"></div>
                        <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Breach Fallout</span>
                    </div>
                </div>
            </div>

            {/* 📋 VECTORS_LIVE SIDEBAR */}
            <div className="w-64 border-l border-blue-900/20 bg-[#05070a]/80 backdrop-blur-md p-4 overflow-hidden flex flex-col">
                <div className="flex items-center justify-between mb-4 pb-2 border-b border-blue-900/20">
                    <div className="flex items-center gap-2">
                        <Activity className="w-4 h-4 text-blue-400" />
                        <span className="text-[10px] font-black tracking-widest text-blue-500 uppercase">Vectors_Live</span>
                    </div>
                    <span className="px-1.5 py-0.5 bg-red-500/10 text-red-500 text-[8px] font-bold rounded border border-red-500/20 animate-pulse">WARGAME</span>
                </div>

                <div className="flex-1 overflow-y-auto custom-scrollbar flex flex-col gap-3">
                    <AnimatePresence mode="popLayout">
                        {threats.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-full opacity-20 gap-3">
                                <Globe className="w-10 h-10 animate-pulse" />
                                <span className="text-[8px] font-mono tracking-widest">Monitoring world coordinates...</span>
                            </div>
                        ) : (
                            threats.map((t, i) => (
                                <motion.div
                                    key={t.id || i}
                                    initial={{ x: 20, opacity: 0 }}
                                    animate={{ x: 0, opacity: 1 }}
                                    exit={{ x: -20, opacity: 0 }}
                                    className="bg-blue-500/5 border-l-2 border-blue-500/30 p-2.5 rounded-r-md"
                                >
                                    <div className="flex justify-between items-start mb-1">
                                        <span className="text-[8px] font-black text-blue-400 font-mono tracking-tighter truncate w-3/4">{t.attacker_ip}</span>
                                        <span className="text-[7px] text-slate-500 font-mono">0.4ms</span>
                                    </div>
                                    <div className="flex items-center gap-1.5">
                                        <Crosshair className="w-2.5 h-2.5 text-red-500" />
                                        <span className="text-[8px] font-bold text-slate-400 uppercase tracking-tighter">{t.type}</span>
                                    </div>
                                    <div className="mt-1 flex items-center justify-between">
                                        <span className="text-[7px] text-blue-500/60 font-mono">COORD: {t.source_lat.toFixed(2)}, {t.source_lng.toFixed(2)}</span>
                                        {t.status === "BREACHED" && <span className="text-[7px] text-red-500 font-black tracking-tighter">BREACHED</span>}
                                    </div>
                                </motion.div>
                            ))
                        )}
                    </AnimatePresence>
                </div>

                <div className="mt-4 pt-2 border-t border-blue-900/20 flex items-center justify-between">
                    <span className="text-[8px] font-bold text-slate-500 uppercase tracking-widest">Sentinels: {sentinels.filter(s => s.status === 'online').length} Online</span>
                    <div className="flex items-center gap-1">
                        <div className="w-1 h-1 rounded-full bg-emerald-500 animate-pulse"></div>
                        <span className="text-[8px] font-bold text-emerald-500 uppercase">Secure</span>
                    </div>
                </div>
            </div>
        </div>
    );
}
