package engine

type Scenario struct {
	Name        string
	Description string
	Max         [][]int
	Allocation  [][]int
	Available   []int
	Total       []int
}

var BuiltInScenarios = map[string]Scenario{
	"CIRCULAR_WAIT": {
		Name:        "CIRCULAR_WAIT",
		Description: "3 processes in circular dependency: P0→R1, P1→R2, P2→R0",
		Max: [][]int{
			{1, 1, 0},
			{0, 1, 1},
			{1, 0, 1},
		},
		Allocation: [][]int{
			{1, 0, 0},
			{0, 1, 0},
			{0, 0, 1},
		},
		Available: []int{0, 0, 0},
		Total:     []int{1, 1, 1},
	},
	"SAFE_STATE": {
		Name:        "SAFE_STATE",
		Description: "Canonical Banker's Algorithm safe state with a valid completion sequence",
		Max: [][]int{
			{7, 5, 3},
			{3, 2, 2},
			{9, 0, 2},
			{2, 2, 2},
			{4, 3, 3},
		},
		Allocation: [][]int{
			{0, 1, 0},
			{2, 0, 0},
			{3, 0, 2},
			{2, 1, 1},
			{0, 0, 2},
		},
		Available: []int{3, 3, 2},
		Total:     []int{10, 5, 7},
	},
	"HOLD_AND_WAIT": {
		Name:        "HOLD_AND_WAIT",
		Description: "4 processes already holding resources while waiting for additional instances",
		Max: [][]int{
			{1, 1, 0, 0},
			{0, 1, 1, 0},
			{0, 0, 1, 1},
			{1, 0, 0, 1},
		},
		Allocation: [][]int{
			{1, 0, 0, 0},
			{0, 1, 0, 0},
			{0, 0, 1, 0},
			{0, 0, 0, 1},
		},
		Available: []int{0, 0, 0, 0},
		Total:     []int{1, 1, 1, 1},
	},
}

func (s *SystemState) LoadScenario(scn Scenario) {
	s.Mu.Lock()

	np := len(scn.Max)
	nr := len(scn.Available)

	if len(s.Processes) != np || len(s.Resources) != nr {
		s.Processes = make([]Process, np)
		s.Resources = make([]Resource, nr)
		s.Available = make([]int, nr)
		s.Max = make([][]int, np)
		s.Allocation = make([][]int, np)
		s.Need = make([][]int, np)
		for i := 0; i < np; i++ {
			s.Max[i] = make([]int, nr)
			s.Allocation[i] = make([]int, nr)
			s.Need[i] = make([]int, nr)
		}
	}

	for i := 0; i < np; i++ {
		s.Processes[i] = Process{ID: i, Status: "Ready"}
		for j := 0; j < nr; j++ {
			s.Max[i][j] = scn.Max[i][j]
			if len(scn.Allocation) > i && len(scn.Allocation[i]) > j {
				s.Allocation[i][j] = scn.Allocation[i][j]
			} else {
				s.Allocation[i][j] = 0
			}
			s.Need[i][j] = s.Max[i][j] - s.Allocation[i][j]
		}
	}

	for j := 0; j < nr; j++ {
		total := 0
		if len(scn.Total) > j {
			total = scn.Total[j]
		} else {
			total = scn.Available[j]
		}
		s.Resources[j] = Resource{ID: j, TotalInstances: total}
		s.Available[j] = scn.Available[j]
	}

	s.Mu.Unlock()
}
