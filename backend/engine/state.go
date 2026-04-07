package engine

import "sync"

type Process struct {
	ID     int
	Status string
}

type Resource struct {
	ID             int
	TotalInstances int
}

type SystemState struct {
	Processes  []Process
	Resources  []Resource
	Available  []int
	Max        [][]int
	Allocation [][]int
	Need       [][]int
	Mu         sync.Mutex
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
