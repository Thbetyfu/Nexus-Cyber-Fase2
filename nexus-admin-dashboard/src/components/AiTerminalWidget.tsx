"use client"

/* 
   NEXUS_UX_STABILITY_COVENANT [LOCKED-BY-ANTIGRAVITY]
   - Peraturan 1: terminalRef.scrollTop dilarang diubah menjadi scrollIntoView.
   - Peraturan 2: Jangan pernah menambahkan autoFocus pada input terminal ini.
*/
import React, { useEffect, useState, useRef } from 'react';
import { Terminal, ServerOff, Loader2 } from 'lucide-react';

interface AIEventLog {
    timestamp: string;
    layer: string;
    status: string;
    detail_action: string;
    error?: string;
}

// Snappy super high-tech text typist effect
const simulateStreamingText = (fullText: string, onUpdate: (currentText: string) => void, onComplete: () => void) => {
    let index = 0;
    let current = "";
    const interval = setInterval(() => {
        if (index < fullText.length) {
            current += fullText[index];
            onUpdate(current);
            index++;
        } else {
            clearInterval(interval);
            onComplete();
        }
    }, 12); // Snappy 12ms per char typing speed
};

export default function AiTerminalWidget() {
    const [aiStatus, setAiStatus] = useState({ state: "INITIALIZING", latency: 0, model: "QWEN3-CORE" });
    const [stream, setStream] = useState<AIEventLog[]>([]);
    const terminalRef = useRef<HTMLDivElement>(null);

    // Dynamic autocomplete suggestion state
    const [inputValue, setInputValue] = useState("");
    const [commandHistory, setCommandHistory] = useState<string[]>([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const [suggestions, setSuggestions] = useState<string[]>([]);

    const allCommands = [
        "/status",
        "/stats",
        "/shuffle",
        "/ban ",
        "/unban ",
        "/honeystats",
        "/patches",
        "/simulate-attack",
        "@nexus ",
        "clear"
    ];

    // Filter autocomplete suggestions based on input
    useEffect(() => {
        const val = inputValue.trim();
        if (val.startsWith("/") || val.startsWith("@") || val.length > 0) {
            const filtered = allCommands.filter(c => 
                c.toLowerCase().startsWith(val.toLowerCase()) && c.toLowerCase() !== val.toLowerCase()
            );
            setSuggestions(filtered);
        } else {
            setSuggestions([]);
        }
    }, [inputValue]);

    useEffect(() => {
        if (terminalRef.current) {
            terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
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
                if (e.data === ': heartbeat') return;

                try {
                    const data = JSON.parse(e.data);
                    if (data.error) {
                        setStream(prev => [...prev, { timestamp: new Date().toISOString(), layer: "SYSTEM", status: "ERROR", detail_action: `> [ERROR] ${data.error}. Retrying...` }].slice(-50));
                        return;
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
                    // silent discard
                }
            };

            eventSource.onerror = (e) => {
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

    const handleCommandSubmit = async (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            if (commandHistory.length > 0) {
                const newIndex = Math.min(historyIndex + 1, commandHistory.length - 1);
                setHistoryIndex(newIndex);
                setInputValue(commandHistory[commandHistory.length - 1 - newIndex]);
            }
            return;
        }

        if (e.key === 'ArrowDown') {
            e.preventDefault();
            if (historyIndex > 0) {
                const newIndex = historyIndex - 1;
                setHistoryIndex(newIndex);
                setInputValue(commandHistory[commandHistory.length - 1 - newIndex]);
            } else {
                setHistoryIndex(-1);
                setInputValue("");
            }
            return;
        }

        if (e.key === 'Enter' && inputValue.trim() !== '') {
            const cmd = inputValue.trim();
            setInputValue("");
            setHistoryIndex(-1);
            setSuggestions([]);
            setCommandHistory(prev => [...prev, cmd].slice(-20));

            if (cmd.toLowerCase() === 'clear') {
                setStream([]);
                return;
            }

            // Optimistic UI Append with command feedback
            setStream(prev => {
                const updated = [...prev.slice(-50), {
                    timestamp: new Date().toISOString(),
                    layer: "Admin",
                    status: "EXEC",
                    detail_action: `nexus_admin@soc:~$ ${cmd}`
                }];

                if (cmd.startsWith('@nexus')) {
                    updated.push({
                        timestamp: new Date().toISOString(),
                        layer: "Reasoning",
                        status: "THINKING",
                        detail_action: "[NEXUS-AI] Analysing cosmic threat vectors and forensic data..."
                    });
                }
                return updated;
            });

            try {
                const res = await fetch('http://localhost:8080/api/cli/execute', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ command: cmd })
                });

                if (res.ok) {
                    const data = await res.json();
                    const rawOutput = data.output || data.response || "[EMPTY_RESPONSE]";

                    // Check if we have a THINKING state to stream into
                    let isThinkingMsg = false;
                    setStream(prev => {
                        const next = [...prev];
                        for (let i = next.length - 1; i >= 0; i--) {
                            if (next[i].status === "THINKING") {
                                isThinkingMsg = true;
                                next[i] = {
                                    ...next[i],
                                    status: "OK",
                                    detail_action: "" // Cleared to prepare for letter typing
                                };
                                return next;
                            }
                        }
                        return next;
                    });

                    if (isThinkingMsg) {
                        simulateStreamingText(rawOutput, (currentText) => {
                            setStream(prev => {
                                const next = [...prev];
                                for (let i = next.length - 1; i >= 0; i--) {
                                    if (next[i].status === "OK" && next[i].layer === "Reasoning") {
                                        next[i] = { ...next[i], detail_action: currentText };
                                        break;
                                    }
                                }
                                return next;
                            });
                        }, () => {});
                    } else {
                        // Regular CLI output - stream letter-by-letter as well
                        const streamMsg = {
                            timestamp: new Date().toISOString(),
                            layer: "System",
                            status: "OK",
                            detail_action: ""
                        };
                        setStream(prev => [...prev.slice(-50), streamMsg]);

                        simulateStreamingText(rawOutput, (currentText) => {
                            setStream(prev => {
                                const next = [...prev];
                                if (next.length > 0) {
                                    next[next.length - 1] = {
                                        ...next[next.length - 1],
                                        detail_action: currentText
                                    };
                                }
                                return next;
                            });
                        }, () => {});
                    }
                } else {
                    setStream(prev => {
                        const next = prev.filter(item => item.status !== 'THINKING');
                        return [...next.slice(-50), {
                            timestamp: new Date().toISOString(),
                            layer: "System",
                            status: "ERROR",
                            detail_action: "[ERROR] Command routing failed."
                        }];
                    });
                }
            } catch (err) {
                setStream(prev => {
                    const next = prev.filter(item => item.status !== 'THINKING');
                    return [...next.slice(-50), {
                        timestamp: new Date().toISOString(),
                        layer: "System",
                        status: "ERROR",
                        detail_action: "[ERROR] Execution offline."
                    }];
                });
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
                    )
                    }
                </div>
            </div>
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar bg-black font-mono" ref={terminalRef}>
                <div className="flex flex-col gap-1 w-full text-[11px] leading-relaxed">
                    <div className="text-cyan-600 mb-2 whitespace-pre font-black leading-none tracking-tighter">
                        {`N EX US   C O R E   O S   v7.2`}
                    </div>
                    <div className="text-cyan-500/50 mb-4">{`> Loading secure cognitive streams...`}</div>

                    {stream.map((log, index) => (
                        <div key={index} className="flex">
                            <span className={`w-full whitespace-pre-wrap break-words ${log.status === 'ERROR' ? 'text-red-500' :
                                    log.status === 'THINKING' ? 'text-fuchsia-400 animate-pulse font-bold' :
                                        log.layer === 'Self-Repair' ? 'text-emerald-400 font-bold' :
                                            log.layer === 'Reasoning' ? 'text-fuchsia-400' :
                                                log.layer === 'Admin' ? 'text-blue-400 font-bold' :
                                                    'text-green-400'
                                }`}>
                                {log.detail_action}
                            </span>
                        </div>
                    ))}

                    {/* Suggestions Box Overlay */}
                    {suggestions.length > 0 && (
                        <div className="bg-[#05080c] border border-cyan-800/40 rounded-lg p-2 mt-3 mb-1 flex flex-col gap-1 text-[10px] text-cyan-400/80 animate-pulse font-mono shadow-[0_0_10px_rgba(6,182,212,0.1)] w-fit max-w-xs">
                            <div className="text-cyan-500 font-bold border-b border-cyan-900/30 pb-0.5 mb-1 uppercase tracking-wider text-[9px]">
                                Command Suggestions:
                            </div>
                            {suggestions.map((item, i) => (
                                <div 
                                    key={i} 
                                    className="cursor-pointer hover:bg-cyan-950/40 hover:text-cyan-300 px-2 py-0.5 rounded transition-colors"
                                    onClick={() => {
                                        setInputValue(item);
                                        setSuggestions([]);
                                    }}
                                >
                                    {item}
                                </div>
                            ))}
                        </div>
                    )}

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

                </div>
            </div>
        </div>
    );
}
