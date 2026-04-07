"use client";

import { useMemo } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  type Node,
  type Edge,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

interface Props {
  nodes: Node[];
  edges: Edge[];
  deadlock: boolean;
  deadlockCycle: number[];
}

export default function ResourceAllocationGraph({
  nodes,
  edges,
  deadlock,
  deadlockCycle,
}: Props) {
  const cycleSet = useMemo(
    () => new Set(deadlockCycle.map((pid) => `P${pid}`)),
    [deadlockCycle]
  );

  const styledNodes = useMemo<Node[]>(() => {
    let pIdx = 0;
    let rIdx = 0;
    return nodes.map((n) => {
      const rawType = (n as Record<string, unknown>).type as string | undefined;
      const isProcess = rawType === "process" || n.id.startsWith("P");
      const inCycle = deadlock && cycleSet.has(n.id);

      let x: number, y: number;
      if (isProcess) {
        x = 80 + pIdx * 180;
        y = 60;
        pIdx++;
      } else {
        x = 80 + rIdx * 220;
        y = 300;
        rIdx++;
      }

      return {
        ...n,
        type: "default",
        position: n.position ?? { x, y },
        data: { label: n.data?.label ?? n.id },
        style: {
          background: inCycle
            ? "#451a1a"
            : isProcess
            ? "#1e3a5f"
            : "#4a3728",
          color: "#e2e8f0",
          border: inCycle
            ? "2px solid #ef4444"
            : isProcess
            ? "2px solid #3b82f6"
            : "2px solid #f59e0b",
          borderRadius: isProcess ? "50%" : "8px",
          width: isProcess ? 60 : 80,
          height: isProcess ? 60 : 40,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: "13px",
          fontWeight: 600,
        },
      };
    });
  }, [nodes, deadlock, cycleSet]);

  const styledEdges = useMemo<Edge[]>(() => {
    return edges.map((e) => {
      const sourceInCycle = deadlock && cycleSet.has(e.source);
      const targetInCycle = deadlock && cycleSet.has(e.target);
      const inCycle = sourceInCycle || targetInCycle;
      const isAllocation = e.data?.label === "Allocation" || e.label === "Allocation";

      return {
        ...e,
        animated: inCycle,
        style: {
          stroke: inCycle ? "#ef4444" : isAllocation ? "#3b82f6" : "#f59e0b",
          strokeWidth: 2,
        },
      };
    });
  }, [edges, deadlock, cycleSet]);

  return (
    <div className="w-full h-full rounded-xl overflow-hidden border border-zinc-800 bg-zinc-950">
      <ReactFlow
        nodes={styledNodes}
        edges={styledEdges}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        panOnDrag={false}
        zoomOnScroll={false}
        zoomOnPinch={false}
        zoomOnDoubleClick={false}
        preventScrolling={false}
        proOptions={{ hideAttribution: true }}
      >
        <Background color="#27272a" gap={20} />
        <Controls
          showInteractive={false}
          className="!bg-zinc-900 !border-zinc-700 !shadow-lg"
        />
      </ReactFlow>
    </div>
  );
}
