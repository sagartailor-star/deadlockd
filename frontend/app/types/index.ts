import type { Node as ReactFlowNode, Edge as ReactFlowEdge } from "@xyflow/react";

export interface MetricsData {
  banker_execution_time: number;
  detection_execution_time: number;
  deadlock_count: number;
  active_goroutines: number;
}

export interface SystemSnapshot {
  available: number[];
  allocation: number[][];
  need: number[][];
  deadlock_status: boolean;
  deadlock_cycle: number[];
  safe_state: boolean;
  safe_sequence: number[];
  nodes: ReactFlowNode[];
  edges: ReactFlowEdge[];
  metrics: MetricsData;
}

export interface ServerMessage {
  type: string;
  payload: SystemSnapshot;
}

export interface ClientCommand {
  type: string;
  pid?: number;
  ticks?: number;
  name?: string;
}
