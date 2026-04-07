package engine

import (
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func logMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("[NIGHTMARE][%s] Alloc=%d MB | TotalAlloc=%d MB | Sys=%d MB | NumGC=%d | Goroutines=%d",
		label,
		m.Alloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC,
		runtime.NumGoroutine(),
	)
}

func TriggerThunderingHerd() *Simulation {
	const numProcesses = 50000
	const numResources = 5000

	log.Printf("[NIGHTMARE] initializing: processes=%d resources=%d", numProcesses, numResources)
	logMemStats("PRE_INIT")

	state := InitializeSystem(numProcesses, numResources)

	state.Mu.Lock()
	for j := 0; j < numResources; j++ {
		state.Resources[j].TotalInstances = 2
		state.Available[j] = 2
	}
	for i := 0; i < numProcesses; i++ {
		for j := 0; j < numResources; j++ {
			state.Max[i][j] = 1
		}
	}
	state.Mu.Unlock()
	state.UpdateNeed()

	logMemStats("POST_INIT")

	sim := NewSimulation(state, 1000)

	sim.wg.Add(1)
	go func() {
		defer sim.wg.Done()
		sim.StartArbiter()
	}()

	log.Printf("[NIGHTMARE] arbiter started, launching thundering herd")

	var gate sync.WaitGroup
	gate.Add(1)

	var herdWg sync.WaitGroup

	for i := 0; i < numProcesses; i++ {
		herdWg.Add(1)
		go func(pid int) {
			defer herdWg.Done()
			gate.Wait()

			rid := rand.Intn(numResources)
			reply := make(chan bool, 1)
			evt := RequestEvent{
				ProcessID:  pid,
				ResourceID: rid,
				Quantity:   1,
				ReplyChan:  reply,
			}

			select {
			case sim.IncomingRequests <- evt:
				select {
				case <-reply:
				case <-sim.ctx.Done():
				case <-time.After(5 * time.Second):
				}
			case <-sim.ctx.Done():
			case <-time.After(5 * time.Second):
			}
		}(i)
	}

	logMemStats("PRE_HERD")
	log.Printf("[NIGHTMARE] releasing %d goroutines simultaneously", numProcesses)
	gate.Done()

	go func() {
		herdWg.Wait()
		log.Printf("[NIGHTMARE] all herd goroutines completed")
		logMemStats("POST_HERD")
	}()

	go func() {
		time.Sleep(30 * time.Second)
		logMemStats("30S_MARK")
	}()

	return sim
}
