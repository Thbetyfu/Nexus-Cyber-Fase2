"use client";

import React, { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence, useDragControls } from "framer-motion";
import { X, Minus, Square, GripVertical, Terminal } from "lucide-react";

interface WindowFrameProps {
  id: string;
  title: string;
  children: React.ReactNode;
  icon?: React.ReactNode;
  initialX?: number;
  initialY?: number;
  width?: string | number;
  height?: string | number;
  onClose?: () => void;
  zIndex: number;
  onFocus: () => void;
  isMinimized?: boolean;
  isActive: boolean;
}

const WindowFrame: React.FC<WindowFrameProps> = ({
  id,
  title,
  children,
  icon,
  initialX = 50,
  initialY = 50,
  width = 600,
  height = 400,
  onClose,
  zIndex,
  onFocus,
  isMinimized = false,
  isActive
}) => {
  const [isMaximized, setIsMaximized] = useState(false);
  const dragControls = useDragControls();

  if (isMinimized) return null;

  return (
    <motion.div
      drag={!isMaximized}
      dragControls={dragControls}
      dragMomentum={false}
      dragListener={false}
      dragConstraints={{ top: 0, left: 0, right: typeof window !== 'undefined' ? window.innerWidth - 150 : 1000, bottom: typeof window !== 'undefined' ? window.innerHeight - 100 : 800 }}
      dragElastic={0}
      initial={{ opacity: 0, scale: 0.95, x: initialX, y: initialY }}
      animate={{ 
        opacity: 1, 
        scale: 1,
        width: isMaximized ? "100vw" : width,
        height: isMaximized ? "calc(100vh - 56px)" : height,
        zIndex: zIndex,
        boxShadow: isActive 
          ? "0 25px 50px -12px rgba(59, 130, 246, 0.2), 0 0 20px rgba(59, 130, 246, 0.1)" 
          : "0 10px 15px -3px rgba(0, 0, 0, 0.4)"
      }}
      exit={{ opacity: 0, scale: 0.9, transition: { duration: 0.2 } }}
      onMouseDown={onFocus}
      className={`absolute flex flex-col bg-[#0c0f14]/95 backdrop-blur-2xl border rounded-xl overflow-hidden transition-colors duration-300 ${
        isActive ? "border-blue-500/50" : "border-gray-800/80"
      }`}
      style={{
        position: isMaximized ? "fixed" : "absolute",
        top: isMaximized ? 0 : undefined,
        left: isMaximized ? 0 : undefined,
        touchAction: "none" // Crucial for framer-motion drag
      }}
    >
      {/* Title Bar - Now acts as the drag handle */}
      <div 
        className={`h-11 flex items-center justify-between px-4 cursor-grab active:cursor-grabbing select-none shrink-0 transition-colors ${
          isActive ? "bg-[#0f172a]" : "bg-[#090b0e]"
        } border-b border-white/5`}
        onPointerDown={(e) => {
          dragControls.start(e);
          onFocus();
        }}
      >
        <div className="flex items-center gap-3">
          <div className={`${isActive ? "text-blue-400" : "text-gray-500"} transition-colors`}>
            {icon || <Terminal size={16} />}
          </div>
          <span className={`text-[11px] font-black uppercase tracking-[0.2em] transition-colors ${
            isActive ? "text-blue-100" : "text-gray-500"
          }`}>
            {title}
          </span>
        </div>

        <div className="flex items-center gap-1.5" onPointerDown={e => e.stopPropagation()}>
          <button 
            onClick={() => {/* minimize */}}
            className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-white/5 text-gray-500 transition-all"
          >
            <Minus size={14} />
          </button>
          <button 
            onClick={() => setIsMaximized(!isMaximized)}
            className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-white/5 text-gray-500 transition-all"
          >
            <Square size={11} />
          </button>
          <button 
            onClick={onClose}
            className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-red-500/20 text-gray-500 hover:text-red-400 transition-all"
          >
            <X size={14} />
          </button>
        </div>
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-auto custom-scrollbar relative bg-[#07090c]/40">
        {children}
      </div>

      {/* Resize Handle (Visual Only for now) */}
      {!isMaximized && (
        <div className="absolute bottom-1 right-1 w-3 h-3 opacity-20 pointer-events-none">
          <div className="w-full h-full border-r-2 border-b-2 border-gray-400 rounded-br-sm" />
        </div>
      )}
    </motion.div>
  );
};

export default WindowFrame;
