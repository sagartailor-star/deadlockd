"use client";

import { useDeadlockSocket } from "./hooks/useDeadlockSocket";
import ResourceAllocationGraph from "./components/ResourceAllocationGraph";
import ControlPanel from "./components/ControlPanel";
import SandboxPanel from "./components/SandboxPanel";
import StatePanel from "./components/StatePanel";
import MetricsPanel from "./components/MetricsPanel";

export default function Home() {
  const { snapshot, sendCommand, connected, loadScenario, sendManualRequest } =
    useDeadlockSocket();

  const graphNodes = snapshot?.nodes ?? [];
  const graphEdges = snapshot?.edges ?? [];
  const deadlock = snapshot?.deadlock_status ?? false;
  const deadlockCycle = snapshot?.deadlock_cycle ?? [];

  return (
    <main className="min-h-screen bg-zinc-950 text-zinc-100">
      <div className="flex min-h-screen flex-col lg:h-screen lg:flex-row">
        <div className="min-w-0 p-4 lg:flex-[7] lg:overflow-hidden">
          <div className="flex h-full min-h-[26rem] flex-col gap-3 sm:min-h-[32rem] lg:min-h-full">
            <div className="flex items-center gap-3 px-1">
              <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <h1 className="text-lg font-semibold tracking-tight text-zinc-200">
                deadlockd
              </h1>
              <span className="text-xs text-zinc-600 font-mono">
                Resource Allocation Graph
              </span>
            </div>
            <div className="min-h-0 flex-1">
              <ResourceAllocationGraph
                nodes={graphNodes}
                edges={graphEdges}
                deadlock={deadlock}
                deadlockCycle={deadlockCycle}
              />
            </div>
          </div>
        </div>

        <div className="min-w-0 border-t border-zinc-800/50 p-4 lg:flex-[3] lg:border-t-0 lg:border-l lg:max-h-screen lg:overflow-y-auto lg:scrollbar-thin">
          <div className="flex flex-col gap-4">
            <ControlPanel
              snapshot={snapshot}
              sendCommand={sendCommand}
              loadScenario={loadScenario}
              connected={connected}
            />
            <SandboxPanel
              sendManualRequest={sendManualRequest}
              snapshot={snapshot}
            />
            <StatePanel snapshot={snapshot} />
            <MetricsPanel metrics={snapshot?.metrics} />
          </div>
        </div>
      </div>
    </main>
  );
}
