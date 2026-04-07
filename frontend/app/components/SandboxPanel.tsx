"use client";

import { useState } from "react";
import { Send, Terminal } from "lucide-react";
import type { SystemSnapshot } from "../types";

interface Props {
  sendManualRequest: (pid: number, rid: number, qty: number) => void;
  snapshot: SystemSnapshot | null;
}

export default function SandboxPanel({ sendManualRequest, snapshot }: Props) {
  const [pid, setPid] = useState(0);
  const [rid, setRid] = useState(0);
  const [qty, setQty] = useState(1);
  const [feedback, setFeedback] = useState("");

  const maxPid = snapshot ? snapshot.allocation.length - 1 : 0;
  const maxRid = snapshot ? snapshot.available.length - 1 : 0;

  const valid = qty > 0 && pid >= 0 && pid <= maxPid && rid >= 0 && rid <= maxRid;

  const handleSubmit = () => {
    if (!valid) return;
    sendManualRequest(pid, rid, qty);
    setQty(1);
    setFeedback("Request Submitted!");
    setTimeout(() => setFeedback(""), 2000);
  };

  return (
    <div className="flex flex-col gap-3 p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
      <div className="flex items-center gap-1.5">
        <Terminal className="w-4 h-4 text-violet-400" />
        <h2 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">
          Sandbox
        </h2>
      </div>

      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
        <div className="flex flex-col gap-1">
          <label className="text-[10px] text-zinc-500 uppercase tracking-wider">
            PID
          </label>
          <input
            type="number"
            min={0}
            max={maxPid}
            value={pid}
            onChange={(e) => setPid(Number(e.target.value))}
            className="px-2.5 py-1.5 bg-zinc-800 border border-zinc-700 rounded-lg text-xs font-mono text-zinc-300 focus:outline-none focus:border-violet-500 transition-colors"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-[10px] text-zinc-500 uppercase tracking-wider">
            RID
          </label>
          <input
            type="number"
            min={0}
            max={maxRid}
            value={rid}
            onChange={(e) => setRid(Number(e.target.value))}
            className="px-2.5 py-1.5 bg-zinc-800 border border-zinc-700 rounded-lg text-xs font-mono text-zinc-300 focus:outline-none focus:border-violet-500 transition-colors"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-[10px] text-zinc-500 uppercase tracking-wider">
            Qty
          </label>
          <input
            type="number"
            min={1}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
            className="px-2.5 py-1.5 bg-zinc-800 border border-zinc-700 rounded-lg text-xs font-mono text-zinc-300 focus:outline-none focus:border-violet-500 transition-colors"
          />
        </div>
      </div>

      <button
        onClick={handleSubmit}
        disabled={!valid || !!feedback}
        className="flex items-center justify-center gap-2 px-3 py-2 bg-violet-600/20 hover:bg-violet-600/30 text-violet-400 rounded-lg border border-violet-600/30 transition-colors text-sm font-medium disabled:opacity-40 disabled:cursor-not-allowed"
      >
        <Send className="w-3.5 h-3.5" />
        {feedback || "Request Resource"}
      </button>
    </div>
  );
}
