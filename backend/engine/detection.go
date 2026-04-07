package engine

func DetectDeadlock(state *SystemState) (bool, []int) {
	state.Mu.Lock()
	np := len(state.Processes)
	nr := len(state.Resources)

	need := make([][]int, np)
	alloc := make([][]int, np)
	for i := 0; i < np; i++ {
		need[i] = make([]int, nr)
		alloc[i] = make([]int, nr)
		copy(need[i], state.Need[i])
		copy(alloc[i], state.Allocation[i])
	}

	avail := make([]int, nr)
	copy(avail, state.Available)
	state.Mu.Unlock()

	graph := make(map[int][]int, np)
	for i := 0; i < np; i++ {
		for k := 0; k < nr; k++ {
			if need[i][k] > 0 && avail[k] == 0 {
				for j := 0; j < np; j++ {
					if j != i && alloc[j][k] > 0 {
						graph[i] = append(graph[i], j)
					}
				}
			}
		}
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make([]int, np)
	parent := make([]int, np)
	for i := 0; i < np; i++ {
		parent[i] = -1
	}

	type frame struct {
		node int
		idx  int
	}

	for start := 0; start < np; start++ {
		if color[start] != white {
			continue
		}

		stack := make([]frame, 0, np)
		stack = append(stack, frame{node: start, idx: 0})
		color[start] = gray

		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			neighbors := graph[top.node]

			if top.idx >= len(neighbors) {
				color[top.node] = black
				stack = stack[:len(stack)-1]
				continue
			}

			next := neighbors[top.idx]
			top.idx++

			if color[next] == gray {
				cycle := make([]int, 0)
				cycle = append(cycle, next)
				for i := len(stack) - 1; i >= 0; i-- {
					cycle = append(cycle, stack[i].node)
					if stack[i].node == next {
						break
					}
				}
				return true, cycle
			}

			if color[next] == white {
				color[next] = gray
				parent[next] = top.node
				stack = append(stack, frame{node: next, idx: 0})
			}
		}
	}

	return false, nil
}
