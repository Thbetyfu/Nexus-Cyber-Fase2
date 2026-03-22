"use client"

import React, { useState } from "react"
import { Globe, Shield, X, Plus, Loader2 } from "lucide-react"

interface AddRouteModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

export default function AddRouteModal({ isOpen, onClose, onSuccess }: AddRouteModalProps) {
    const [domain, setDomain] = useState("")
    const [targetUrl, setTargetUrl] = useState("http://host.docker.internal:3001")
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setIsSubmitting(true)
        setError(null)

        try {
            const res = await fetch("http://localhost:8080/api/routes", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ domain, target_url: targetUrl }),
            })

            if (!res.ok) throw new Error("Failed to add route")

            setDomain("")
            onSuccess()
            onClose()
        } catch (err: any) {
            setError(err.message || "Something went wrong")
        } finally {
            setIsSubmitting(false)
        }
    }

    if (!isOpen) return null

    return (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/60 backdrop-blur-sm"
                onClick={onClose}
            />

            {/* Modal */}
            <div className="relative w-full max-w-md bg-[#0a0c10] border border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
                <div className="bg-slate-900/50 p-6 border-b border-slate-700/50 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="bg-blue-500/10 p-2 rounded-lg border border-blue-500/20">
                            <Shield className="w-5 h-5 text-blue-400" />
                        </div>
                        <div>
                            <h2 className="text-lg font-bold text-slate-100">Onboard New Website</h2>
                            <p className="text-xs text-slate-500">Add dynamic routing to the MTD Matrix</p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-slate-500 hover:text-white transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div>
                        <label className="block text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5">Domain Name</label>
                        <div className="relative">
                            <Globe className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                            <input
                                type="text"
                                required
                                value={domain}
                                onChange={(e) => setDomain(e.target.value)}
                                placeholder="e.kemenkeu.localhost"
                                className="w-full bg-slate-900/50 border border-slate-700/50 rounded-xl pl-10 pr-4 py-2.5 text-sm text-slate-200 focus:outline-none focus:border-blue-500/50 transition-all font-mono"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5">Target Backend URL</label>
                        <input
                            type="url"
                            required
                            value={targetUrl}
                            onChange={(e) => setTargetUrl(e.target.value)}
                            placeholder="http://10.0.0.5:80"
                            className="w-full bg-slate-900/50 border border-slate-700/50 rounded-xl px-4 py-2.5 text-sm text-slate-200 focus:outline-none focus:border-blue-500/50 transition-all font-mono"
                        />
                        <p className="mt-1.5 text-[10px] text-slate-600 font-medium">Use 'host.docker.internal' for host-side apps</p>
                    </div>

                    {error && (
                        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-xs text-red-400">
                            {error}
                        </div>
                    )}

                    <div className="pt-2">
                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-blue-600/20 flex items-center justify-center gap-2 group"
                        >
                            {isSubmitting ? (
                                <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                                <>
                                    <Plus className="w-4 h-4 group-hover:scale-110 transition-transform" />
                                    PROTECT WEBSITE
                                </>
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
