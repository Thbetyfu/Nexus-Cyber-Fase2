"use client"

import React, { useEffect, useState, useRef } from 'react';
import { Terminal, ServerOff, Loader2 } from 'lucide-react';

interface AIEventLog {
    timestamp: string;
    layer: string;
    status: string;
    detail_action: string;
    error?: string;
}

export default function AiTerminalWidget() {
    const [aiStatus, setAiStatus] = useState({ state: "INITIALIZING", latency: 0, model: "QWEN3-CORE" });
    const [stream, setStream] = useState<AIEventLog[]>([]);
    const terminalEndRef = useRef<HTMLDivElement>(null);

    // Auto-scroll logic
    useEffect(() => {
        if (terminalEndRef.current) {
            terminalEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [stream]);

    // Status ping logic
    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const res = await fetch('http://localhost:8080/api/ai/status');
                const data = await res.json();
                setAiStatus({ state: data.status, latency: data.latency_ms, model: data.model });
            } catch (err) {
                setAiStatus({ state: "DISCONNECTED", latency: 0, model: "NEXUS-CORE" });
            }
        };

        fetchStatus();
        const timer = setInterval(fetchStatus, 5000);
        return () => clearInterval(timer);
    }, []);

    // SSE Stream Logic
    useEffect(() => {
        let eventSource: EventSource | null = null;
        let reconnectTimeout: NodeJS.Timeout;

        const connectSSE = () => {
            eventSource = new EventSource('http://localhost:8080/api/ai/stream');

            eventSource.onmessage = (e) => {
                if (e.data === ': heartbeat') return; // Silent discard of backend heartbeats

                try {
                    const data = JSON.parse(e.data);
                    if (data.error) {
                        setStream(prev => [...prev, { timestamp: new Date().toISOString(), layer: "SYSTEM", status: "ERROR", detail_action: `> [ERROR] ${data.error}. Retrying...` }].slice(-50));
                        return
                    }

                    let prefix = "";
                    if (data.layer === "Reflex") prefix = "> [REFLEX_CORE] ";
                    else if (data.layer === "Reasoning") prefix = "> [INTENT_ANALYSIS] ";
                    else if (data.layer === "Self-Repair") prefix = "> [REPAIR_MODULE] ";
                    else prefix = "> [SYS] ";

                    setStream(prev => {
                        const newMsg = { ...data, detail_action: `${prefix}${data.detail_action}` };
                        return [...prev.slice(-50), newMsg];
                    });
                } catch (err) {
                    // console.debug("SSE Heartbeat/Comment bypass", e.data);
                }
            };

            eventSource.onerror = (e) => {
                // eventSource?.close(); // Let browser handle retry unless it's a hard failure
                if (eventSource?.readyState === EventSource.CLOSED) {
                    setStream(prev => [...prev.slice(-50), { timestamp: new Date().toISOString(), layer: "SYSTEM", status: "ERROR", detail_action: "> [ERROR] Telemetry connection lost. Retrying in 5s..." }]);
                    reconnectTimeout = setTimeout(connectSSE, 5000);
                }
            };
        };

        connectSSE();

        return () => {
            if (eventSource) eventSource.close();
            clearTimeout(reconnectTimeout);
        };
    }, []);

    const [inputValue, setInputValue] = useState("");

    const handleCommandSubmit = async (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter' && inputValue.trim() !== '') {
            const cmd = inputValue.trim();
            setInputValue("");

            // Optimistic UI Append
            setStream(prev => [...prev.slice(-50), {
                timestamp: new Date().toISOString(),
                layer: "Admin",
                status: "EXEC",
                detail_action: `nexus_admin@soc:~$ ${cmd}`
            }]);

            try {
                const res = await fetch('http://localhost:8080/api/cli/execute', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ command: cmd })
                });

                if (res.ok) {
                    const data = await res.json();
                    setStream(prev => [...prev.slice(-50), {
                        timestamp: new Date().toISOString(),
                        layer: "System",
                        status: "OK",
                        detail_action: data.response
                    }]);
                } else {
                    setStream(prev => [...prev.slice(-50), {
                        timestamp: new Date().toISOString(),
                        layer: "System",
                        status: "ERROR",
                        detail_action: "[ERROR] Command routing failed."
                    }]);
                }
            } catch (err) {
                setStream(prev => [...prev.slice(-50), {
                    timestamp: new Date().toISOString(),
                    layer: "System",
                    status: "ERROR",
                    detail_action: "[ERROR] Execution offline."
                }]);
            }
        }
    };

    return (
        <div className="bg-[#030507] border border-cyan-900/30 rounded-xl flex flex-col shadow-[0_0_15px_rgba(6,182,212,0.05)] overflow-hidden h-full relative">
            <div className="bg-[#05080c] px-4 py-2 border-b border-cyan-900/30 flex items-center justify-between sticky top-0 z-10 shrink-0">
                <h3 className="text-xs font-semibold text-cyan-500 uppercase tracking-widest flex items-center gap-2">
                    <Terminal className="w-4 h-4" /> Nexus Core Terminal
                </h3>
                <div className="flex items-center gap-2">
                    {aiStatus.state === 'ONLINE' ? (
                        <>
                            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                            <span className="text-[10px] text-emerald-400 font-mono font-bold tracking-tighter">
                                {aiStatus.model}: ONLINE ({aiStatus.latency}ms)
                            </span>
                        </>
                    ) : aiStatus.state === 'INITIALIZING' ? (
                        <>
                            <Loader2 className="w-3 h-3 text-cyan-500 animate-spin" />
                            <span className="text-[10px] text-cyan-400 font-mono tracking-tighter">INITIALIZING...</span>
                        </>
                    ) : (
                        <>
                            <ServerOff className="w-3 h-3 text-red-500" />
                            <span className="text-[10px] text-red-500 font-mono font-bold tracking-tighter">
                                AI CORE: DISCONNECTED
                            </span>
                        </>
                    )}
                </div>
            </div>
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar bg-black font-mono">
                <div className="flex flex-col gap-1 w-full text-[11px] leading-relaxed">
                    <div className="text-cyan-600 mb-2 whitespace-pre font-black leading-none tracking-tighter">
                        {`N EX US   C O R E   O S   v7.0`}
                    </div>
                    <div className="text-cyan-500/50 mb-4">{`> Loading secure cognitive streams...`}</div>

                    {stream.map((log, index) => (
                        <div key={index} className="flex">
                            <span className={`w-full whitespace-pre-wrap break-words ${log.status === 'ERROR' ? 'text-red-500' :
                                log.layer === 'Self-Repair' ? 'text-emerald-400 font-bold' :
                                    log.layer === 'Reasoning' ? 'text-fuchsia-400' :
                                        log.layer === 'Admin' ? 'text-blue-400 font-bold' :
                                            'text-green-400'
                                }`}>
                                {log.detail_action}
                            </span>
                        </div>
                    ))}

                    <div className="flex items-center mt-2 group">
                        <span className="text-emerald-500 shrink-0 mr-2 font-bold tracking-tighter">nexus_admin@soc:~$</span>
                        <input
                            type="text"
                            className="bg-transparent border-none outline-none text-green-400 w-full font-mono text-[11px] focus:ring-0 p-0"
                            placeholder="Type /help or @nexus [query]..."
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            onKeyDown={handleCommandSubmit}
                            autoComplete="off"
                            spellCheck="false"
                        />
                    </div>

                    <div ref={terminalEndRef} />
                </div>
            </div>
        </div>
    );
}
