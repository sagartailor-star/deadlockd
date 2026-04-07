"use client";

import { useState } from "react";
import {
  Play,
  Square,
  Skull,
  Gauge,
  CircleDot,
  ShieldCheck,
  ShieldAlert,
  Wifi,
  WifiOff,
  FlaskConical,
  Upload,
} from "lucide-react";
import type { SystemSnapshot } from "../types";

const SCENARIOS = ["CIRCULAR_WAIT", "SAFE_STATE", "HOLD_AND_WAIT"] as const;

interface Props {
  snapshot: SystemSnapshot | null;
  sendCommand: (type: string, payload?: Record<string, unknown>) => void;
  loadScenario: (name: string) => void;
  connected: boolean;
}

export default function ControlPanel({
  snapshot,
  sendCommand,
  loadScenario,
  connected,
}: Props) {
  const [ticks, setTicks] = useState(10);
  const [selectedScenario, setSelectedScenario] = useState<string>(SCENARIOS[0]);
  const [simRunning, setSimRunning] = useState(false);

  const handleTickChange = (val: number) => {
    setTicks(val);
    sendCommand("UPDATE_TICKS", { ticks: val });
  };

  return (
    <div className="flex flex-col gap-4 p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">
          Control Panel
        </h2>
        <div className="flex flex-wrap items-center gap-1.5">
          {connected ? (
            <Wifi className="w-4 h-4 text-emerald-400" />
          ) : (
            <WifiOff className="w-4 h-4 text-red-400" />
          )}
          <span
            className={`text-xs font-medium ${
              connected ? "text-emerald-400" : "text-red-400"
            }`}
          >
            {connected ? "Connected" : "Disconnected"}
          </span>
        </div>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row">
        <button
          onClick={() => {
            sendCommand("START_SIM");
            setSimRunning(true);
          }}
          className="relative flex flex-1 items-center justify-center gap-2 px-3 py-2.5 bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 rounded-lg border border-emerald-600/30 transition-colors text-sm font-medium"
        >
          {simRunning && (
            <span className="absolute top-2 right-2 flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
          )}
          <Play className="w-4 h-4" />
          Start
        </button>
        <button
          onClick={() => {
            sendCommand("STOP_SIM");
            setSimRunning(false);
          }}
          className="flex flex-1 items-center justify-center gap-2 px-3 py-2.5 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-lg border border-red-600/30 transition-colors text-sm font-medium"
        >
          <Square className="w-4 h-4" />
          Stop
        </button>
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-1.5 text-zinc-400 text-xs">
          <FlaskConical className="w-3.5 h-3.5" />
          Scenario
        </div>
        <div className="flex flex-col gap-2 sm:flex-row">
          <select
            value={selectedScenario}
            onChange={(e) => setSelectedScenario(e.target.value)}
            className="min-w-0 flex-1 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-xs text-zinc-300 font-mono focus:outline-none focus:border-blue-500 transition-colors"
          >
            {SCENARIOS.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <button
            onClick={() => loadScenario(selectedScenario)}
            className="flex items-center justify-center gap-1.5 px-3 py-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 rounded-lg border border-blue-600/30 transition-colors text-xs font-medium sm:w-auto"
          >
            <Upload className="w-3.5 h-3.5" />
            Load
          </button>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-zinc-400 text-xs">
            <Gauge className="w-3.5 h-3.5" />
            Ticks/sec
          </div>
          <span className="text-xs font-mono text-zinc-300">{ticks}</span>
        </div>
        <input
          type="range"
          min={1}
          max={50}
          value={ticks}
          onChange={(e) => handleTickChange(Number(e.target.value))}
          className="w-full h-1.5 bg-zinc-700 rounded-full appearance-none cursor-pointer accent-blue-500"
        />
      </div>

      <div className="space-y-1.5">
        <div className="flex flex-wrap items-center gap-1.5">
          {snapshot?.deadlock_status ? (
            <ShieldAlert className="w-4 h-4 text-red-400" />
          ) : (
            <ShieldCheck className="w-4 h-4 text-emerald-400" />
          )}
          <span
            className={`text-sm font-medium ${
              snapshot?.deadlock_status ? "text-red-400" : "text-emerald-400"
            }`}
          >
            {snapshot?.deadlock_status ? "Deadlock Detected" : "No Deadlock"}
          </span>
        </div>
        {snapshot?.safe_state && snapshot.safe_sequence.length > 0 && (
          <p className="break-all text-xs text-zinc-500 font-mono">
            Safe: {"<"}
            {snapshot.safe_sequence.join(", ")}
            {">"}
          </p>
        )}
      </div>

      {snapshot && (
        <div className="space-y-1.5">
          <h3 className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
            Processes
          </h3>
          <div className="max-h-40 overflow-y-auto space-y-1 pr-1 scrollbar-thin">
            {snapshot.allocation.map((_, i) => (
              <div
                key={i}
                className="flex items-center justify-between px-3 py-1.5 bg-zinc-800/60 rounded-lg"
              >
                <div className="flex items-center gap-2">
                  <CircleDot
                    className={`w-3.5 h-3.5 ${
                      snapshot.deadlock_cycle.includes(i)
                        ? "text-red-400"
                        : "text-blue-400"
                    }`}
                  />
                  <span className="text-xs font-mono text-zinc-300">
                    P{i}
                  </span>
                </div>
                <button
                  onClick={() => sendCommand("TERMINATE_PROCESS", { pid: i })}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-red-400 hover:bg-red-600/20 rounded transition-colors"
                >
                  <Skull className="w-3 h-3" />
                  Kill
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
