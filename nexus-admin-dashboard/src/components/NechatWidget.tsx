"use client"

import React, { useState, useRef, useEffect } from "react"
import { MessageSquare, X, Send, Command, Loader2, Minimize2, Maximize2 } from "lucide-react"

interface Message {
    id: number;
    text: string;
    sender: "user" | "nechat";
}

interface NechatWidgetProps {
    activeDomain: string;
}

// Simple Markdown parser for basic **bold** and `code` formatting
const parseMarkdown = (text: string) => {
    // Bold
    let parsed = text.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    // Inline Code
    parsed = parsed.replace(/`(.*?)`/g, '<code class="bg-gray-800 text-blue-300 px-1 rounded font-mono text-xs">$1</code>');
    // Newlines to BR
    parsed = parsed.replace(/\n/g, '<br/>');
    return <span dangerouslySetInnerHTML={{ __html: parsed }} />;
};

export default function NechatWidget({ activeDomain }: NechatWidgetProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [isMinimized, setIsMinimized] = useState(false);
    const [messages, setMessages] = useState<Message[]>([
        { id: 1, text: "Halo, Admin! Saya NECHAT, Asisten intelijen SOC Anda. Berdasarkan Qwen3-235B, saya sedang memantau log sekuriti. Ada yang bisa saya analisis hari ini?", sender: "nechat" }
    ]);
    const [input, setInput] = useState("");
    const [isTyping, setIsTyping] = useState(false);

    const endOfMessagesRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        endOfMessagesRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages, isTyping, isOpen]);

    const handleSend = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!input.trim()) return;

        const query = input;
        setInput("");

        // Add user message to UI
        setMessages(prev => [...prev, { id: Date.now(), text: query, sender: "user" }]);
        setIsTyping(true);

        try {
            const res = await fetch("http://localhost:8080/api/nechat", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ query, domain: activeDomain })
            });

            if (!res.ok) throw new Error("API Offline or CORS error");

            const data = await res.json();

            setMessages(prev => [...prev, { id: Date.now(), text: data.reply, sender: "nechat" }]);
        } catch (error) {
            console.error(error);
            setMessages(prev => [...prev, {
                id: Date.now(),
                text: "⚠️ **Gangguan Komunikasi:** Tidak dapat menghubungi AI Cortex di Backend. Pastikan Gateway berjalan normal dan port 8080 aktif.",
                sender: "nechat"
            }]);
        } finally {
            setIsTyping(false);
        }
    };

    if (!isOpen) {
        return (
            <button
                onClick={() => setIsOpen(true)}
                className="fixed bottom-6 right-6 p-4 bg-emerald-600/20 hover:bg-emerald-600/40 border border-emerald-500/50 backdrop-blur shadow-2xl rounded-full text-emerald-400 focus:outline-none transition-transform hover:scale-105 z-50 flex items-center justify-center"
            >
                <Command className="w-6 h-6 animate-pulse" />
            </button>
        )
    }

    return (
        <div className={`fixed right-6 bottom-6 flex flex-col bg-[#0b0e14]/95 backdrop-blur-xl border border-gray-800/80 rounded-2xl shadow-2xl z-50 overflow-hidden transition-all duration-300 ${isMinimized ? 'w-72 h-14' : 'w-80 md:w-96 h-[500px]'}`}>

            {/* Header */}
            <div className="bg-[#0f141d] px-4 py-3 border-b border-gray-800/80 flex items-center justify-between cursor-pointer" onClick={() => setIsMinimized(!isMinimized)}>
                <div className="flex items-center gap-2">
                    <div className="bg-emerald-500/10 p-1.5 rounded-lg">
                        <Command className="w-5 h-5 text-emerald-500" />
                    </div>
                    <div>
                        <h3 className="text-sm font-bold text-gray-200 tracking-wide">NECHAT ASSIST</h3>
                        <p className="text-[10px] text-emerald-500 flex items-center gap-1 font-mono">
                            <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse"></span>
                            Qwen3-235B
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-1">
                    <button className="p-1 text-gray-400 hover:text-white rounded" onClick={(e) => { e.stopPropagation(); setIsMinimized(!isMinimized); }}>
                        {isMinimized ? <Maximize2 className="w-4 h-4" /> : <Minimize2 className="w-4 h-4" />}
                    </button>
                    <button className="p-1 text-gray-400 hover:text-red-400 rounded" onClick={(e) => { e.stopPropagation(); setIsOpen(false); }}>
                        <X className="w-4 h-4" />
                    </button>
                </div>
            </div>

            {/* Chat Area */}
            {!isMinimized && (
                <>
                    <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gradient-to-b from-[#0b0e14] to-[#06080b]">
                        {messages.map((msg) => (
                            <div key={msg.id} className={`flex ${msg.sender === 'user' ? 'justify-end' : 'justify-start'}`}>
                                <div className={`max-w-[85%] rounded-2xl px-4 py-2.5 text-sm shadow-sm ${msg.sender === 'user'
                                    ? 'bg-blue-600/20 border border-blue-500/30 text-blue-100 rounded-br-sm'
                                    : 'bg-gray-800/40 border border-gray-700/50 text-gray-300 rounded-bl-sm'
                                    }`}>
                                    {msg.sender === 'nechat' ? parseMarkdown(msg.text) : msg.text}
                                </div>
                            </div>
                        ))}

                        {isTyping && (
                            <div className="flex justify-start">
                                <div className="max-w-[85%] rounded-2xl rounded-bl-sm px-4 py-3 text-sm bg-gray-800/40 border border-gray-700/50 text-emerald-400 flex items-center gap-2">
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                    <span className="font-mono text-xs tracking-widest">ANALYZING LOGS...</span>
                                </div>
                            </div>
                        )}
                        <div ref={endOfMessagesRef} />
                    </div>

                    {/* Input Area */}
                    <div className="p-3 bg-[#0a0d11] border-t border-gray-800/80">
                        <form onSubmit={handleSend} className="flex gap-2 relative">
                            <input
                                type="text"
                                value={input}
                                onChange={(e) => setInput(e.target.value)}
                                placeholder="Ask NECHAT about current threats..."
                                className="flex-1 bg-gray-900/50 border border-gray-700/50 rounded-xl px-4 py-2 text-sm text-gray-200 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 placeholder-gray-500 transition-all font-mono"
                                disabled={isTyping}
                            />
                            <button
                                type="submit"
                                disabled={!input.trim() || isTyping}
                                className="bg-emerald-600 hover:bg-emerald-500 text-white p-2 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center shrink-0"
                            >
                                <Send className="w-4 h-4" />
                            </button>
                        </form>
                    </div>
                </>
            )}
        </div>
    )
}
