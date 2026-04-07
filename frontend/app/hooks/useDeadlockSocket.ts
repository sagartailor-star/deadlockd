"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import type { SystemSnapshot, ServerMessage } from "../types";

const MAX_RECONNECT_DELAY = 30000;
const BASE_RECONNECT_DELAY = 1000;

function resolveWebSocketUrl(): string {
  const envUrl = process.env.NEXT_PUBLIC_WS_URL;
  if (envUrl) {
    return envUrl;
  }

  if (typeof window === "undefined") {
    return "ws://localhost:8080/ws";
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  if (window.location.port === "3000") {
    return `${protocol}//${window.location.hostname}:8080/ws`;
  }

  return `${protocol}//${window.location.host}/ws`;
}

export function useDeadlockSocket() {
  const [snapshot, setSnapshot] = useState<SystemSnapshot | null>(null);
  const [connected, setConnected] = useState(false);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const unmountedRef = useRef(false);

  const connect = useCallback(function connectSocket() {
    if (unmountedRef.current) return;

    const ws = new WebSocket(resolveWebSocketUrl());
    socketRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      reconnectAttempt.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data);
        if (msg.type === "STATE_UPDATE" && msg.payload) {
          setSnapshot(msg.payload);
        }
      } catch {
        /* ignore malformed */
      }
    };

    ws.onclose = () => {
      setConnected(false);
      socketRef.current = null;
      if (!unmountedRef.current) {
        const delay = Math.min(
          BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttempt.current),
          MAX_RECONNECT_DELAY
        );
        reconnectAttempt.current++;
        reconnectTimer.current = setTimeout(connectSocket, delay);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, []);

  useEffect(() => {
    unmountedRef.current = false;
    connect();

    return () => {
      unmountedRef.current = true;
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
      if (socketRef.current) {
        socketRef.current.close();
        socketRef.current = null;
      }
    };
  }, [connect]);

  const sendCommand = useCallback(
    (type: string, payload?: Record<string, unknown>) => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.send(JSON.stringify({ type, ...payload }));
      }
    },
    []
  );

  const loadScenario = useCallback(
    (name: string) => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.send(JSON.stringify({ type: "LOAD_SCENARIO", name }));
      }
    },
    []
  );

  const sendManualRequest = useCallback(
    (pid: number, rid: number, qty: number) => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.send(
          JSON.stringify({ type: "MANUAL_REQUEST", pid, rid, qty })
        );
      }
    },
    []
  );

  return { snapshot, connected, sendCommand, loadScenario, sendManualRequest };
}
