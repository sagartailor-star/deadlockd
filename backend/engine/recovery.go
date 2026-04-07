package engine

func (s *SystemState) TerminateProcess(pid int) {
	s.Mu.Lock()
	nr := len(s.Resources)
	for j := 0; j < nr; j++ {
		s.Available[j] += s.Allocation[pid][j]
		s.Allocation[pid][j] = 0
		s.Need[pid][j] = 0
		s.Max[pid][j] = 0
	}
	s.Processes[pid].Status = "Terminated"
	s.Mu.Unlock()
}

func (s *SystemState) ResolveDeadlock(cycle []int) int {
	s.Mu.Lock()
	nr := len(s.Resources)
	victim := cycle[0]
	maxTotal := 0
	for _, pid := range cycle {
		total := 0
		for j := 0; j < nr; j++ {
			total += s.Allocation[pid][j]
		}
		if total > maxTotal {
			maxTotal = total
			victim = pid
		}
	}
	s.Mu.Unlock()

	s.TerminateProcess(victim)
	return victim
}
