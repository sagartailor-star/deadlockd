package engine

import "testing"

func TestCircularWaitScenarioCreatesDeadlock(t *testing.T) {
	state := InitializeSystem(1, 1)
	state.LoadScenario(BuiltInScenarios["CIRCULAR_WAIT"])

	deadlocked, cycle := DetectDeadlock(state)
	if !deadlocked {
		t.Fatal("expected CIRCULAR_WAIT to produce a deadlock")
	}
	if len(cycle) == 0 {
		t.Fatal("expected a non-empty deadlock cycle")
	}
}

func TestSafeStateScenarioIsSafeAndAcceptsSafeRequest(t *testing.T) {
	state := InitializeSystem(1, 1)
	state.LoadScenario(BuiltInScenarios["SAFE_STATE"])

	safe, sequence := IsSafeState(state)
	if !safe {
		t.Fatal("expected SAFE_STATE to be safe")
	}
	if len(sequence) != len(state.Processes) {
		t.Fatalf("expected safe sequence for %d processes, got %d", len(state.Processes), len(sequence))
	}

	sim := NewSimulation(state, 10)
	granted := sim.ProcessManualRequest(1, 0, 1)
	if !granted {
		t.Fatal("expected safe manual request to be granted")
	}

	if got := state.Available[0]; got != 2 {
		t.Fatalf("expected R0 available to drop to 2, got %d", got)
	}
	if got := state.Allocation[1][0]; got != 3 {
		t.Fatalf("expected P1 allocation for R0 to be 3, got %d", got)
	}
	if got := state.Need[1][0]; got != 0 {
		t.Fatalf("expected P1 need for R0 to drop to 0, got %d", got)
	}
}

func TestHoldAndWaitScenarioStartsWithHeldResources(t *testing.T) {
	state := InitializeSystem(1, 1)
	state.LoadScenario(BuiltInScenarios["HOLD_AND_WAIT"])

	for pid := range state.Processes {
		held := 0
		for rid := range state.Resources {
			held += state.Allocation[pid][rid]
		}
		if held == 0 {
			t.Fatalf("expected process %d to hold at least one resource", pid)
		}
	}
}

func TestReleaseHeldResourcesDoesNotCreateNegativeAllocation(t *testing.T) {
	state := InitializeSystem(1, 1)
	state.LoadScenario(BuiltInScenarios["SAFE_STATE"])

	sim := NewSimulation(state, 10)
	state.TerminateProcess(1)
	sim.releaseHeldResources(1, 0, 2)

	if got := state.Allocation[1][0]; got != 0 {
		t.Fatalf("expected terminated process allocation to stay at 0, got %d", got)
	}
	if got := state.Available[0]; got != 5 {
		t.Fatalf("expected available resources to stay at 5 after duplicate release, got %d", got)
	}
}
