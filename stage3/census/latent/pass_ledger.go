package latent

// PassStep is one production tail operation. Production is an optional AST
// statement key for a block or rewrite rather than a call. The overlay owns the
// single ordered table; the source guard checks it against main. Scoped selects
// the opt-in disposable observation, without changing whole-program coverage.
type PassStep struct {
	Name         string
	WholeProgram bool
	Scoped       bool
	Production   string
}

type PassEntry struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Reason      string  `json:"reason,omitempty"`
	WallSeconds float64 `json:"wall_seconds,omitempty"`
}

// OmittedPassLedger records absence of measurement. It never calls a pass or
// claims execution. A body stop takes precedence over every tail omission.
func OmittedPassLedger(steps []PassStep, loweringStopped bool) []PassEntry {
	result := make([]PassEntry, len(steps))
	for i, step := range steps {
		status := "not_run_scoped"
		if step.WholeProgram {
			status = "not_run_whole_program"
		}
		if loweringStopped {
			status = "not_reached_lowering_stop"
		}
		result[i] = PassEntry{Name: step.Name, Status: status}
	}
	return result
}
