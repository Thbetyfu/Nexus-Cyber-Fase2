"use client"

/* 
   NEXUS_UX_STABILITY_COVENANT [LOCKED-BY-ANTIGRAVITY]
   - Peraturan 1: Dilarang memunculkan fokus otomatis pada widget chat.
   - Peraturan 2: Gunakan scrollTop untuk log internal, jangan scrollIntoView.
*/
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
    const [messages, setMessages] = useState<Message[]>([
        { id: 1, text: "Halo, Admin! Saya NECHAT, Asisten intelijen SOC Anda. Berdasarkan Qwen3-235B, saya sedang memantau log sekuriti. Ada yang bisa saya analisis hari ini?", sender: "nechat" }
    ]);
    const [input, setInput] = useState("");
    const [isTyping, setIsTyping] = useState(false);

    const chatContainerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (chatContainerRef.current) {
            chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
        }
    }, [messages, isTyping]);

    const handleSend = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!input.trim()) return;

        const query = input;
        setInput("");

        setMessages(prev => [...prev, { id: Date.now(), text: query, sender: "user" }]);
        setIsTyping(true);

        try {
            const minLoadingPromise = new Promise(resolve => setTimeout(resolve, 1000));
            
            const fetchPromise = fetch("http://localhost:8080/api/nechat", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ query, domain: activeDomain })
            });

            const [res] = await Promise.all([fetchPromise, minLoadingPromise]);

            if (!res.ok) throw new Error("API Offline");

            const data = await res.json();
            const replyText = data.reply && data.reply.trim() !== "" 
                ? data.reply 
                : "⚠️ **Sistem Alpaca Gagal Merespons.**";

            setMessages(prev => [...prev, { id: Date.now(), text: replyText, sender: "nechat" }]);
        } catch (error) {
            setMessages(prev => [...prev, {
                id: Date.now(),
                text: "⚠️ **Gangguan Komunikasi:** Tidak dapat menghubungi Nexus Gateway.",
                sender: "nechat"
            }]);
        } finally {
            setIsTyping(false);
        }
    };

    return (
        <div className="flex flex-col h-full w-full bg-[#0b0e14]/50 overflow-hidden">
            {/* Chat Area */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gradient-to-b from-[#0b0e14] to-[#06080b] custom-scrollbar" ref={chatContainerRef}>
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
                            <span className="font-mono text-[10px] tracking-widest">ANALYZING LOGS...</span>
                        </div>
                    </div>
                )}
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
        </div>
    )
}

