# Frontend

This is the Next.js dashboard for `deadlockd`.

## Commands

```bash
npm install
npm run dev
npm run lint
npm run build
npm run start
```

## What this app shows

- Resource allocation graph
- Control panel for simulation commands
- Sandbox for manual resource requests
- Allocation and need matrices
- Timing and deadlock metrics

## WebSocket behavior

The client connects in this order:

1. `NEXT_PUBLIC_WS_URL` if you set it
2. `ws://<hostname>:8080/ws` when the page runs on port `3000`
3. same-origin `/ws` for proxied setups
