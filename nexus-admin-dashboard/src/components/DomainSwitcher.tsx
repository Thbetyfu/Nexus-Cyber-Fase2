"use client";

import React, { useState, useEffect } from 'react';
import { Globe, ChevronDown, Check, Plus, Trash2, AlertTriangle } from 'lucide-react';

interface DomainSwitcherProps {
    activeDomain: string;
    onDomainChange: (domain: string) => void;
    onAddClick: () => void;
    refreshTrigger?: number;
}

export default function DomainSwitcher({ activeDomain, onDomainChange, onAddClick, refreshTrigger = 0 }: DomainSwitcherProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [domains, setDomains] = useState<string[]>(['all']);

    const fetchDomains = async () => {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 3000);

        try {
            const res = await fetch('http://localhost:8080/api/domains', {
                signal: controller.signal,
                mode: 'cors'
            });
            clearTimeout(timeoutId);

            if (res.ok) {
                const data = await res.json();
                const fetchedDomains = data?.domains || (Array.isArray(data) ? data : []);
                if (fetchedDomains.length > 0) {
                    setDomains(['all', ...fetchedDomains.filter((d: string) => d !== 'all')]);
                }
            }
        } catch (err) {}
    };

    useEffect(() => {
        fetchDomains();
        const interval = setInterval(fetchDomains, 10000);
        return () => clearInterval(interval);
    }, [refreshTrigger]);

    return (
        <div className="relative">
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-3 bg-slate-900/50 border border-slate-700/50 hover:border-blue-500/50 rounded-lg px-4 py-2 transition-all group"
            >
                <div className="bg-blue-500/10 p-1.5 rounded-md border border-blue-500/20 group-hover:bg-blue-500/20">
                    <Globe className="w-4 h-4 text-blue-400" />
                </div>
                <div className="text-left">
                    <p className="text-[10px] text-slate-500 font-bold uppercase tracking-tighter leading-none mb-1">Active Workspace</p>
                    <p className="text-xs font-mono text-slate-200 truncate max-w-[150px]">
                        {activeDomain === 'all' ? 'GLOBAL_OVERWATCH' : activeDomain.toUpperCase()}
                    </p>
                </div>
                <ChevronDown className={`w-4 h-4 text-slate-500 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
            </button>

            {isOpen && (
                <>
                    <div
                        className="fixed inset-0 z-[100]"
                        onClick={() => {
                            setIsOpen(false);
                        }}
                    />
                    <div className="absolute bottom-full right-0 mb-2 w-72 bg-[#0a0c10]/95 border border-slate-700/50 rounded-xl shadow-2xl p-2 z-[101] backdrop-blur-xl animate-in fade-in zoom-in duration-150">
                        <div className="p-2 border-b border-slate-800/50 mb-1">
                            <button
                                onClick={() => {
                                    onAddClick();
                                    setIsOpen(false);
                                }}
                                className="flex items-center gap-2 w-full p-2.5 rounded-lg text-[10px] font-black uppercase tracking-widest text-blue-400 hover:bg-blue-500/10 hover:text-blue-300 transition-all group"
                            >
                                <Plus className="w-3.5 h-3.5 group-hover:scale-110 transition-transform" />
                                Add New Workspace
                            </button>
                        </div>
                        <div className="max-h-[300px] overflow-y-auto custom-scrollbar">
                            {domains.map((domain) => (
                                <div key={domain} className="group/item relative flex items-center gap-1">
                                    <button
                                        onClick={() => {
                                            onDomainChange(domain);
                                            setIsOpen(false);
                                        }}
                                        className={`flex-1 flex items-center justify-between p-3 rounded-lg text-[11px] transition-all ${activeDomain === domain
                                            ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                                            : 'text-slate-400 hover:bg-white/5 hover:text-white'
                                            }`}
                                    >
                                        <span className="font-mono truncate max-w-[200px]">
                                            {domain === 'all' ? 'All Workspaces (Global)' : domain}
                                        </span>
                                        {activeDomain === domain && <Check className="w-3 h-3" />}
                                    </button>
                                </div>
                            ))}
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}

