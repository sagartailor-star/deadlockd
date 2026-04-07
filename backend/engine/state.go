package engine

import "sync"

// Process represents an individual execution unit in the simulation.
// It maps the internal ID to an existential state ("Ready", "Terminated").
type Process struct {
	ID     int
	Status string
}

// Resource represents an allocatable system entity.
// It maintains the absolute upper bound of instances available system-wide.
type Resource struct {
	ID             int
	TotalInstances int
}

// SystemState encapsulates the entire runtime state of the DeadlockD simulator.
// It holds the core matrices required for executing Banker's Algorithm and the Deadlock Detection Algorithm.
//
// Thread Safety:
// Due to the highly concurrent nature of the simulation (where goroutines continually dispatch requests),
// all mutating matrix and vector operations MUST acquire the `Mu` mutex.
// Starvation and unbounded access duration are mitigated via strict, short-lived critical sections
// across the simulation tick bounds constraint.
type SystemState struct {
	Processes  []Process
	Resources  []Resource
	Available  []int     // Vector: Number of currently available instances for each resource
	Max        [][]int   // Matrix: Maximum potential claim per process per resource
	Allocation [][]int   // Matrix: Currently held instances per process per resource
	Need       [][]int   // Matrix: Remaining required instances per process per resource (Max - Allocation)
	Mu         sync.Mutex // Protects concurrent mutative access across state slices
}


func InitializeSystem(numProcesses int, numResources int) *SystemState {
	processes := make([]Process, numProcesses)
	for i := 0; i < numProcesses; i++ {
		processes[i] = Process{
			ID:     i,
			Status: "Ready",
		}
	}

	resources := make([]Resource, numResources)
	for i := 0; i < numResources; i++ {
		resources[i] = Resource{
			ID:             i,
			TotalInstances: 0,
		}
	}

	available := make([]int, numResources)

	maxMatrix := make([][]int, numProcesses)
	allocation := make([][]int, numProcesses)
	need := make([][]int, numProcesses)
	for i := 0; i < numProcesses; i++ {
		maxMatrix[i] = make([]int, numResources)
		allocation[i] = make([]int, numResources)
		need[i] = make([]int, numResources)
	}

	return &SystemState{
		Processes:  processes,
		Resources:  resources,
		Available:  available,
		Max:        maxMatrix,
		Allocation: allocation,
		Need:       need,
	}
}

func (s *SystemState) RecalculateNeed() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for i := range s.Processes {
		for j := range s.Resources {
			s.Need[i][j] = s.Max[i][j] - s.Allocation[i][j]
		}
	}
}

func (s *SystemState) UpdateNeed() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	np := len(s.Processes)
	nr := len(s.Resources)
	for i := 0; i < np; i++ {
		for j := 0; j < nr; j++ {
			s.Need[i][j] = s.Max[i][j] - s.Allocation[i][j]
		}
	}
}
