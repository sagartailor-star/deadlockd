package api

import (
	"encoding/json"
	"net/http"
	"time"

	"deadlockd/engine"

	"github.com/gorilla/websocket"
)

type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ClientCommand struct {
	Type  string `json:"type"`
	PID   int    `json:"pid,omitempty"`
	Ticks int    `json:"ticks,omitempty"`
	Name  string `json:"name,omitempty"`
	RID   int    `json:"rid,omitempty"`
	Qty   int    `json:"qty,omitempty"`
}

type metricsPayload struct {
	BankerExecutionTime    int64  `json:"banker_execution_time"`
	DetectionExecutionTime int64  `json:"detection_execution_time"`
	DeadlockCount          uint64 `json:"deadlock_count"`
	ActiveGoroutines       int64  `json:"active_goroutines"`
}

type snapshotPayload struct {
	Available      []int          `json:"available"`
	Allocation     [][]int        `json:"allocation"`
	Need           [][]int        `json:"need"`
	DeadlockStatus bool           `json:"deadlock_status"`
	DeadlockCycle  []int          `json:"deadlock_cycle"`
	SafeState      bool           `json:"safe_state"`
	SafeSequence   []int          `json:"safe_sequence"`
	Nodes          []engine.Node  `json:"nodes"`
	Edges          []engine.Edge  `json:"edges"`
	Metrics        metricsPayload `json:"metrics"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles the HTTP upgrade routing for the frontend.
// It constructs the real-time WebSocket connection to the Next.js React client.
// 
// Protocol Flow:
// 1. Initial successful connection triggers immediate snapshot sync guaranteeing consistency.
// 2. Continuous event handling loop decodes 'ClientCommand' requests spanning lifecycle manipulation, rate limits, Sandbox execution and standard simulation execution. 
// 3. Real-time broadcast pushes delta frames (`ServerMessage->STATE_UPDATE`) mapping the system architecture over a volatile, unthrottled bus architecture.
func HandleWebSocket(
	hub *Hub,
	sim *engine.Simulation,
	state *engine.SystemState,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(conn)
		defer hub.Unregister(conn)

		initialSnapshot := BuildSnapshot(state, sim)
		initialData, err := json.Marshal(initialSnapshot)
		if err != nil {
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, initialData); err != nil {
			return
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var cmd ClientCommand
			if json.Unmarshal(msg, &cmd) != nil {
				continue
			}

			switch cmd.Type {
			case "START_SIM":
				np := len(state.Processes)
				sim.Start(np)
			case "STOP_SIM":
				sim.Stop()
			case "TERMINATE_PROCESS":
				state.TerminateProcess(cmd.PID)
			case "UPDATE_TICKS":
				if cmd.Ticks > 0 {
					sim.SetTickInterval(time.Second / time.Duration(cmd.Ticks))
				}
			case "LOAD_SCENARIO":
				scn, ok := engine.BuiltInScenarios[cmd.Name]
				if !ok {
					continue
				}
				sim.Stop()
				state.LoadScenario(scn)
			case "MANUAL_REQUEST":
				go func() {
					if sim.IsRunning() {
						sim.SubmitManualRequest(cmd.PID, cmd.RID, cmd.Qty)
					} else {
						sim.ProcessManualRequest(cmd.PID, cmd.RID, cmd.Qty)
					}
					snap := BuildSnapshot(state, sim)
					data, err := json.Marshal(snap)
					if err != nil {
						return
					}
					select {
					case hub.Broadcast <- data:
					default:
					}
				}()
				continue
			default:
				continue
			}

			snapshot := BuildSnapshot(state, sim)
			data, err := json.Marshal(snapshot)
			if err != nil {
				continue
			}

			select {
			case hub.Broadcast <- data:
			default:
			}
		}
	}
}

func BuildSnapshot(state *engine.SystemState, sim *engine.Simulation) ServerMessage {
	state.Mu.Lock()
	np := len(state.Processes)
	nr := len(state.Resources)

	avail := make([]int, nr)
	copy(avail, state.Available)

	alloc := make([][]int, np)
	need := make([][]int, np)
	for i := 0; i < np; i++ {
		alloc[i] = make([]int, nr)
		need[i] = make([]int, nr)
		copy(alloc[i], state.Allocation[i])
		copy(need[i], state.Need[i])
	}
	state.Mu.Unlock()

	deadlocked, cycle := engine.DetectDeadlock(state)
	safe, safeSeq := engine.IsSafeState(state)
	nodes, edges := state.GenerateGraphSnapshot()

	if cycle == nil {
		cycle = []int{}
	}
	if safeSeq == nil {
		safeSeq = []int{}
	}

	m := sim.GetMetrics()

	return ServerMessage{
		Type: "STATE_UPDATE",
		Payload: snapshotPayload{
			Available:      avail,
			Allocation:     alloc,
			Need:           need,
			DeadlockStatus: deadlocked,
			DeadlockCycle:  cycle,
			SafeState:      safe,
			SafeSequence:   safeSeq,
			Nodes:          nodes,
			Edges:          edges,
			Metrics: metricsPayload{
				BankerExecutionTime:    m.BankerExecutionTime,
				DetectionExecutionTime: m.DetectionExecutionTime,
				DeadlockCount:          m.DeadlockCount,
				ActiveGoroutines:       m.ActiveGoroutines,
			},
		},
	}
}
