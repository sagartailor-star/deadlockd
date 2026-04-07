package engine

import "fmt"

type Node struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

func (s *SystemState) GenerateGraphSnapshot() ([]Node, []Edge) {
	s.Mu.Lock()
	np := len(s.Processes)
	nr := len(s.Resources)

	alloc := make([][]int, np)
	need := make([][]int, np)
	statuses := make([]string, np)
	for i := 0; i < np; i++ {
		alloc[i] = make([]int, nr)
		need[i] = make([]int, nr)
		copy(alloc[i], s.Allocation[i])
		copy(need[i], s.Need[i])
		statuses[i] = s.Processes[i].Status
	}
	s.Mu.Unlock()

	nodes := make([]Node, 0, np+nr)
	edges := make([]Edge, 0, np*nr)

	for i := 0; i < np; i++ {
		if statuses[i] == "Terminated" {
			continue
		}
		nodes = append(nodes, Node{
			ID:    fmt.Sprintf("P%d", i),
			Type:  "process",
			Label: fmt.Sprintf("P%d", i),
		})
	}

	for j := 0; j < nr; j++ {
		nodes = append(nodes, Node{
			ID:    fmt.Sprintf("R%d", j),
			Type:  "resource",
			Label: fmt.Sprintf("R%d", j),
		})
	}

	for i := 0; i < np; i++ {
		for j := 0; j < nr; j++ {
			if alloc[i][j] > 0 {
				edges = append(edges, Edge{
					ID:     fmt.Sprintf("e-alloc-R%d-P%d", j, i),
					Source: fmt.Sprintf("R%d", j),
					Target: fmt.Sprintf("P%d", i),
					Label:  "Allocation",
				})
			}
			if need[i][j] > 0 {
				edges = append(edges, Edge{
					ID:     fmt.Sprintf("e-req-P%d-R%d", i, j),
					Source: fmt.Sprintf("P%d", i),
					Target: fmt.Sprintf("R%d", j),
					Label:  "Request",
				})
			}
		}
	}

	return nodes, edges
}
