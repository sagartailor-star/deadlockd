package engine

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	BankerExecutionTime    atomic.Int64
	DetectionExecutionTime atomic.Int64
	DeadlockCount          atomic.Uint64
	ActiveGoroutines       atomic.Int64
}

type MetricsSnapshot struct {
	BankerExecutionTime    int64  `json:"banker_execution_time"`
	DetectionExecutionTime int64  `json:"detection_execution_time"`
	DeadlockCount          uint64 `json:"deadlock_count"`
	ActiveGoroutines       int64  `json:"active_goroutines"`
}

type RequestEvent struct {
	ProcessID  int
	ResourceID int
	Quantity   int
	ReplyChan  chan bool
}

// Simulation encapsulates the execution context for concurrent processes competing for system resources.
// It actively manages the process lifecycle, dispatch events via channels, and bounds tick executions.
// 
// Architecture Strategy:
// High-grade separation of concerns is maintained where the `Simulation` acts as a scheduler/arbitrator,
// and individual processes run as completely isolated goroutines utilizing buffered channels for communication.
// 
// Lifecycle constraints:
// To ensure a clean teardown, it utilizes a `context.Context` pattern and a `sync.WaitGroup` to orchestrate 
// the parallel shutdown operations of active goroutines, preventing memory/goroutine leaks.
type Simulation struct {
	State            *SystemState
	IncomingRequests chan RequestEvent // Pipeline processing automated requests
	ManualRequests   chan RequestEvent // Pipeline parsing external (user/system) commands
	ctx              context.Context
	cancel           context.CancelFunc
	tickInterval     atomic.Int64
	running          atomic.Bool
	wg               sync.WaitGroup
	onStateChange    func()            // Asynchronous callback fired on guaranteed state deltas
	metrics          Metrics
}

func (sim *Simulation) GetMetrics() MetricsSnapshot {
	return MetricsSnapshot{
		BankerExecutionTime:    sim.metrics.BankerExecutionTime.Load(),
		DetectionExecutionTime: sim.metrics.DetectionExecutionTime.Load(),
		DeadlockCount:          sim.metrics.DeadlockCount.Load(),
		ActiveGoroutines:       sim.metrics.ActiveGoroutines.Load(),
	}
}

func (sim *Simulation) SetStateChangeCallback(fn func()) {
	sim.onStateChange = fn
}

func (sim *Simulation) SetTickInterval(d time.Duration) {
	sim.tickInterval.Store(int64(d))
}

func (sim *Simulation) IsRunning() bool {
	return sim.running.Load()
}

func (sim *Simulation) getTickInterval() time.Duration {
	return time.Duration(sim.tickInterval.Load())
}

func NewSimulation(state *SystemState, ticksPerSecond int) *Simulation {
	ctx, cancel := context.WithCancel(context.Background())
	sim := &Simulation{
		State:            state,
		IncomingRequests: make(chan RequestEvent, 256),
		ManualRequests:   make(chan RequestEvent, 64),
		ctx:              ctx,
		cancel:           cancel,
	}
	sim.tickInterval.Store(int64(time.Second / time.Duration(ticksPerSecond)))
	return sim
}

func (sim *Simulation) SubmitManualRequest(pid int, rid int, qty int) bool {
	reply := make(chan bool, 1)
	evt := RequestEvent{
		ProcessID:  pid,
		ResourceID: rid,
		Quantity:   qty,
		ReplyChan:  reply,
	}
	select {
	case sim.ManualRequests <- evt:
	case <-sim.ctx.Done():
		return false
	}
	select {
	case granted := <-reply:
		return granted
	case <-sim.ctx.Done():
		return false
	}
}

func (sim *Simulation) ProcessManualRequest(pid int, rid int, qty int) bool {
	return sim.evaluateRequest(RequestEvent{
		ProcessID:  pid,
		ResourceID: rid,
		Quantity:   qty,
	})
}

func (sim *Simulation) Start(numProcesses int) {
	// First ensure cleanly stopped if it was running
	sim.cancel()
	sim.wg.Wait()

	// Create new context for new run
	sim.ctx, sim.cancel = context.WithCancel(context.Background())
	sim.running.Store(true)

	sim.wg.Add(1)
	go func() {
		defer sim.wg.Done()
		sim.StartArbiter()
	}()

	for i := 0; i < numProcesses; i++ {
		sim.wg.Add(1)
		go func(id int) {
			defer sim.wg.Done()
			sim.runProcess(id)
		}(i)
	}
}

func (sim *Simulation) Stop() {
	sim.cancel()
	sim.wg.Wait()
	sim.running.Store(false)
}

func (sim *Simulation) runProcess(id int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
	nr := len(sim.State.Resources)

	for {
		select {
		case <-sim.ctx.Done():
			return
		default:
		}

		tick := sim.getTickInterval()
		thinkTime := time.Duration(rng.Int63n(int64(tick)*3)) + tick
		select {
		case <-sim.ctx.Done():
			return
		case <-time.After(thinkTime):
		}

		candidates := make([]int, 0)
		sim.State.Mu.Lock()
		for j := 0; j < nr; j++ {
			if sim.State.Max[id][j] > 0 && sim.State.Need[id][j] > 0 {
				candidates = append(candidates, j)
			}
		}
		sim.State.Mu.Unlock()

		if len(candidates) == 0 {
			continue
		}

		resID := candidates[rng.Intn(len(candidates))]

		sim.State.Mu.Lock()
		needVal := sim.State.Need[id][resID]
		sim.State.Mu.Unlock()

		if needVal <= 0 {
			continue
		}

		qty := rng.Intn(needVal) + 1

		reply := make(chan bool, 1)
		evt := RequestEvent{
			ProcessID:  id,
			ResourceID: resID,
			Quantity:   qty,
			ReplyChan:  reply,
		}

		select {
		case <-sim.ctx.Done():
			return
		case sim.IncomingRequests <- evt:
		}

		select {
		case <-sim.ctx.Done():
			return
		case granted := <-reply:
			if granted {
				// Hold resources for a while (e.g., 2-5 ticks)
				holdTicks := rng.Intn(4) + 2
				for k := 0; k < holdTicks; k++ {
					select {
					case <-sim.ctx.Done():
						return
					case <-time.After(sim.getTickInterval()):
					}
				}

				sim.releaseHeldResources(id, resID, qty)

				if sim.onStateChange != nil {
					sim.onStateChange()
				}
			}
		}
	}
}

func (sim *Simulation) releaseHeldResources(pid int, rid int, qty int) {
	sim.State.Mu.Lock()
	defer sim.State.Mu.Unlock()

	if pid < 0 || pid >= len(sim.State.Processes) || rid < 0 || rid >= len(sim.State.Resources) {
		return
	}

	allocated := sim.State.Allocation[pid][rid]
	if allocated <= 0 {
		return
	}

	releaseQty := qty
	if allocated < releaseQty {
		releaseQty = allocated
	}

	sim.State.Allocation[pid][rid] -= releaseQty
	sim.State.Available[rid] += releaseQty
	sim.State.Need[pid][rid] += releaseQty
}

func (sim *Simulation) evaluateRequest(evt RequestEvent) bool {
	sim.State.Mu.Lock()

	// Check if request is within Need and Available
	if evt.Quantity > sim.State.Need[evt.ProcessID][evt.ResourceID] ||
		evt.Quantity > sim.State.Available[evt.ResourceID] {
		sim.State.Mu.Unlock()
		return false
	}

	// Tentative grant
	sim.State.Available[evt.ResourceID] -= evt.Quantity
	sim.State.Allocation[evt.ProcessID][evt.ResourceID] += evt.Quantity
	sim.State.Need[evt.ProcessID][evt.ResourceID] -= evt.Quantity
	sim.State.Mu.Unlock()

	bankerStart := time.Now()
	safe, _ := IsSafeState(sim.State)
	sim.metrics.BankerExecutionTime.Store(time.Since(bankerStart).Nanoseconds())

	grant := false
	if safe {
		grant = true
	} else {
		// Rollback if unsafe
		sim.State.Mu.Lock()
		sim.State.Available[evt.ResourceID] += evt.Quantity
		sim.State.Allocation[evt.ProcessID][evt.ResourceID] -= evt.Quantity
		sim.State.Need[evt.ProcessID][evt.ResourceID] += evt.Quantity
		sim.State.Mu.Unlock()
	}

	detectStart := time.Now()
	deadlocked, cycle := DetectDeadlock(sim.State)
	sim.metrics.DetectionExecutionTime.Store(time.Since(detectStart).Nanoseconds())

	if deadlocked && len(cycle) > 0 {
		sim.metrics.DeadlockCount.Add(1)
		sim.State.ResolveDeadlock(cycle)
	}

	if sim.onStateChange != nil {
		sim.onStateChange()
	}

	return grant
}

func (sim *Simulation) processRequest(evt RequestEvent) {
	grant := sim.evaluateRequest(evt)

	select {
	case evt.ReplyChan <- grant:
	case <-sim.ctx.Done():
		return
	}
}

func (sim *Simulation) StartArbiter() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			sim.metrics.ActiveGoroutines.Store(int64(runtime.NumGoroutine()))
		case evt, ok := <-sim.IncomingRequests:
			if !ok {
				return
			}
			sim.processRequest(evt)
		case evt, ok := <-sim.ManualRequests:
			if !ok {
				return
			}
			sim.processRequest(evt)
		}
	}
}
