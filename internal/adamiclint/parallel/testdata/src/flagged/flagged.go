// Every test in this package's _test.go files is serial with no reason given, and must be flagged.
package flagged

// lookalike has a Parallel method of its own, which is not testing.T's.
type lookalike struct{}

func (lookalike) Parallel() {}
