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
  const [size, setSize] = useState({ 
    width: typeof width === 'number' ? width : parseInt(String(width)) || 600, 
    height: typeof height === 'number' ? height : parseInt(String(height)) || 400 
  });
  const [isResizing, setIsResizing] = useState(false);
  const dragControls = useDragControls();
  const windowRef = useRef<HTMLDivElement>(null);

  const handleResizeStart = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsResizing(true);
    
    const startX = e.clientX;
    const startY = e.clientY;
    const startWidth = size.width;
    const startHeight = size.height;

    const onMouseMove = (moveEvent: MouseEvent) => {
      const newWidth = Math.max(350, startWidth + (moveEvent.clientX - startX));
      const newHeight = Math.max(250, startHeight + (moveEvent.clientY - startY));
      setSize({ width: newWidth, height: newHeight });
    };

    const onMouseUp = () => {
      setIsResizing(false);
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
    };

    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
  };

  if (isMinimized) return null;

  return (
    <motion.div
      ref={windowRef}
      drag={!isMaximized && !isResizing}
      dragControls={dragControls}
      dragMomentum={false}
      dragListener={false}
      dragConstraints={{ top: 0, left: 0, right: typeof window !== 'undefined' ? window.innerWidth - 100 : 1000, bottom: typeof window !== 'undefined' ? window.innerHeight - 100 : 800 }}
      dragElastic={0}
      initial={{ opacity: 0, scale: 0.95, x: initialX, y: initialY }}
      animate={{ 
        opacity: 1, 
        scale: 1,
        x: isMaximized ? 0 : undefined,
        y: isMaximized ? 0 : undefined,
        width: isMaximized ? "100vw" : size.width,
        height: isMaximized ? "calc(100vh - 56px)" : size.height,
        zIndex: zIndex,
        borderRadius: isMaximized ? "0px" : "12px",
        boxShadow: isActive 
          ? "0 25px 50px -12px rgba(59, 130, 246, 0.2), 0 0 20px rgba(59, 130, 246, 0.1)" 
          : "0 10px 15px -3px rgba(0, 0, 0, 0.4)"
      }}
      exit={{ opacity: 0, scale: 0.9, transition: { duration: 0.2 } }}
      onMouseDown={onFocus}
      className={`absolute flex flex-col bg-[#0c0f14]/95 backdrop-blur-3xl border transition-colors duration-300 ${
        isActive ? "border-blue-500/50" : "border-gray-800/80"
      } ${isMaximized ? "z-[9999]" : ""}`}
      style={{
        position: isMaximized ? "fixed" : "absolute",
        top: isMaximized ? 0 : undefined,
        left: isMaximized ? 0 : undefined,
        touchAction: "none"
      }}
    >
      {/* Title Bar */}
      <div 
        className={`h-11 flex items-center justify-between px-4 cursor-grab active:cursor-grabbing select-none shrink-0 transition-colors ${
          isActive ? "bg-[#0f172a]" : "bg-[#090b0e]"
        } border-b border-white/5`}
        onPointerDown={(e) => {
          if (!isMaximized) {
            dragControls.start(e);
            onFocus();
          }
        }}
        onDoubleClick={() => setIsMaximized(!isMaximized)}
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
            className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-white/5 text-gray-500 transition-all"
            onClick={() => {/* minimize logic usually handled by parent */}}
          >
            <Minus size={14} />
          </button>
          <button 
            onClick={() => setIsMaximized(!isMaximized)}
            className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-white/5 text-gray-500 transition-all"
          >
            {isMaximized ? (
              <div className="relative w-3 h-3 border border-current">
                <div className="absolute -top-1 -right-1 w-2 h-2 border border-current bg-[#0c0f14]" />
              </div>
            ) : (
              <Square size={11} />
            )}
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

      {/* Resize Handle */}
      {!isMaximized && (
        <div 
          className="absolute bottom-0 right-0 w-4 h-4 cursor-nwse-resize group z-50"
          onMouseDown={handleResizeStart}
        >
          <div className="absolute bottom-1.5 right-1.5 w-2.5 h-2.5 border-r-2 border-b-2 border-gray-600/50 group-hover:border-blue-500 transition-colors" />
        </div>
      )}
    </motion.div>
  );
};


export default WindowFrame;
