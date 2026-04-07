package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"deadlockd/api"
	"deadlockd/engine"
)

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func main() {
	var sim *engine.Simulation
	var state *engine.SystemState

	if os.Getenv("NIGHTMARE") == "1" {
		log.Printf("[NIGHTMARE] chaos mode activated")
		sim = engine.TriggerThunderingHerd()
		state = sim.State
	} else {
		numProcesses := envInt("STRESS_PROCESSES", 5)
		numResources := envInt("STRESS_RESOURCES", 3)

		log.Printf("initializing: processes=%d resources=%d", numProcesses, numResources)

		state = engine.InitializeSystem(numProcesses, numResources)

		state.Mu.Lock()
		for j := 0; j < numResources; j++ {
			total := 3 + (j*7)%10
			state.Resources[j].TotalInstances = total
			state.Available[j] = total
		}
		for i := 0; i < numProcesses; i++ {
			for j := 0; j < numResources; j++ {
				state.Max[i][j] = state.Resources[j].TotalInstances / numProcesses
				if state.Max[i][j] < 1 {
					state.Max[i][j] = 1
				}
			}
		}
		state.Mu.Unlock()
		state.UpdateNeed()

		sim = engine.NewSimulation(state, 10)
	}
	hub := api.NewHub()

	sim.SetStateChangeCallback(func() {
		snapshot := api.BuildSnapshot(state, sim)
		data, err := json.Marshal(snapshot)
		if err != nil {
			return
		}
		select {
		case hub.Broadcast <- data:
		default:
		}
	})

	go hub.Run()

	ctx, cancel := context.WithCancel(context.Background())

	go statsLogger(ctx, sim)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", api.HandleWebSocket(hub, sim, state))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-sigChan
	log.Printf("shutting down")

	cancel()
	sim.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	close(hub.Broadcast)
}

func statsLogger(ctx context.Context, sim *engine.Simulation) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	var prevBanker int64
	var prevCount int

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := sim.GetMetrics()
			currentBanker := m.BankerExecutionTime
			prevCount++
			prevBanker += currentBanker
			avgBanker := prevBanker / int64(prevCount)

			log.Printf("[STATS] AvgBanker=%d ns (%.2f µs) | Deadlocks=%d | Goroutines=%d",
				avgBanker,
				float64(avgBanker)/1000.0,
				m.DeadlockCount,
				m.ActiveGoroutines,
			)
		}
	}
}
