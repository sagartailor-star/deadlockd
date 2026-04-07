"use client";

import { Timer, Bug, Cpu, Activity } from "lucide-react";
import type { MetricsData } from "../types";

interface Props {
  metrics: MetricsData | undefined;
}

function formatNanos(ns: number): string {
  if (ns < 1_000) return `${ns} ns`;
  if (ns < 1_000_000) return `${(ns / 1_000).toFixed(2)} µs`;
  return `${(ns / 1_000_000).toFixed(2)} ms`;
}

export default function MetricsPanel({ metrics }: Props) {
  if (!metrics) {
    return (
      <div className="flex items-center justify-center h-full p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
        <p className="text-zinc-500 text-sm">No metrics yet</p>
      </div>
    );
  }

  const cards = [
    {
      label: "Banker Time",
      value: formatNanos(metrics.banker_execution_time),
      icon: Timer,
      color: "text-blue-400",
    },
    {
      label: "Detection Time",
      value: formatNanos(metrics.detection_execution_time),
      icon: Activity,
      color: "text-violet-400",
    },
    {
      label: "Deadlocks",
      value: metrics.deadlock_count.toString(),
      icon: Bug,
      color: metrics.deadlock_count > 0 ? "text-amber-400" : "text-zinc-400",
    },
    {
      label: "Goroutines",
      value: metrics.active_goroutines.toString(),
      icon: Cpu,
      color: metrics.active_goroutines > 100 ? "text-amber-400" : "text-emerald-400",
    },
  ];

  return (
    <div className="p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
      <h2 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider mb-3">
        Metrics
      </h2>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {cards.map((card) => (
          <div
            key={card.label}
            className="flex min-w-0 flex-col gap-1 p-3 bg-zinc-800/60 rounded-lg border border-zinc-700/50"
          >
            <div className="flex items-center gap-1.5">
              <card.icon className={`w-3.5 h-3.5 ${card.color}`} />
              <span className="break-words text-[10px] leading-snug text-zinc-500 uppercase tracking-wider">
                {card.label}
              </span>
            </div>
            <span className={`text-sm font-mono font-semibold ${card.color}`}>
              {card.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
