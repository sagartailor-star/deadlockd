"use client";

import type { SystemSnapshot } from "../types";

interface Props {
  snapshot: SystemSnapshot | null;
}

export default function StatePanel({ snapshot }: Props) {
  if (!snapshot) {
    return (
      <div className="flex items-center justify-center h-full p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
        <p className="text-zinc-500 text-sm">Waiting for data…</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-5 bg-zinc-900/80 backdrop-blur-sm rounded-xl border border-zinc-800">
      <div>
        <h3 className="text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">
          Available
        </h3>
        <div className="flex flex-wrap gap-2">
          {snapshot.available.map((val, j) => (
            <div
              key={j}
              className="px-3 py-1.5 bg-zinc-800 rounded-lg text-center"
            >
              <span className="text-[10px] text-zinc-500 block">R{j}</span>
              <span className="text-sm font-mono text-zinc-200">{val}</span>
            </div>
          ))}
        </div>
      </div>

      <div>
        <h3 className="text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">
          Allocation Matrix
        </h3>
        <div className="overflow-x-auto">
          <table className="min-w-max text-xs font-mono">
            <thead>
              <tr className="text-zinc-500">
                <th className="px-2 py-1 text-left"></th>
                {snapshot.available.map((_, j) => (
                  <th key={j} className="px-2 py-1 text-center">
                    R{j}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {snapshot.allocation.map((row, i) => (
                <tr key={i} className="border-t border-zinc-800/50">
                  <td className="px-2 py-1 text-zinc-500">P{i}</td>
                  {row.map((val, j) => (
                    <td
                      key={j}
                      className={`px-2 py-1 text-center ${
                        val > 0 ? "text-blue-400" : "text-zinc-600"
                      }`}
                    >
                      {val}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div>
        <h3 className="text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">
          Need Matrix
        </h3>
        <div className="overflow-x-auto">
          <table className="min-w-max text-xs font-mono">
            <thead>
              <tr className="text-zinc-500">
                <th className="px-2 py-1 text-left"></th>
                {snapshot.available.map((_, j) => (
                  <th key={j} className="px-2 py-1 text-center">
                    R{j}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {snapshot.need.map((row, i) => (
                <tr key={i} className="border-t border-zinc-800/50">
                  <td className="px-2 py-1 text-zinc-500">P{i}</td>
                  {row.map((val, j) => (
                    <td
                      key={j}
                      className={`px-2 py-1 text-center ${
                        val > 0 ? "text-amber-400" : "text-zinc-600"
                      }`}
                    >
                      {val}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
