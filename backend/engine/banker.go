package engine

// IsSafeState evaluates if a safe sequence exists for the current simulation state,
// effectively proving that all processes could eventually terminate without deadlock.
//
// Algorithm implementation adheres directly to Banker's Algorithm criteria:
// 1. A simulated 'work' vector resolves potential free resources.
// 2. Iteratively searches for processes capable of executing (Need <= Work).
// 3. Claims the allocated resources back.
//
// Complexity:
// - Time Complexity: O(P^2 * R), where P is number of processes and R is the number of resources.
// - Space Complexity: O(P * R) transient space to prevent mutating core state.
func IsSafeState(state *SystemState) (bool, []int) {
	state.Mu.Lock()
	np := len(state.Processes)
	nr := len(state.Resources)

	work := make([]int, nr)
	copy(work, state.Available)

	need := make([][]int, np)
	alloc := make([][]int, np)
	for i := 0; i < np; i++ {
		need[i] = make([]int, nr)
		alloc[i] = make([]int, nr)
		copy(need[i], state.Need[i])
		copy(alloc[i], state.Allocation[i])
	}
	state.Mu.Unlock()

	finish := make([]bool, np)
	safeSequence := make([]int, 0, np)

	for count := 0; count < np; count++ {
		found := false
		for i := 0; i < np; i++ {
			if finish[i] {
				continue
			}
			canRun := true
			for j := 0; j < nr; j++ {
				if need[i][j] > work[j] {
					canRun = false
					break
				}
			}
			if canRun {
				for j := 0; j < nr; j++ {
					work[j] += alloc[i][j]
				}
				finish[i] = true
				safeSequence = append(safeSequence, i)
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, safeSequence
}
