package program

import "testing"

// The default is one checker per core, from one to checkerCeiling (#xwv641q): a laptop's cores all check,
// and a machine past the knee stops adding checkers that buy no wall.
func TestTheDefaultCheckerCountIsOnePerCoreUpToTheCeiling(t *testing.T) {
	t.Parallel()
	for cores, want := range map[int]int{0: 1, 1: 1, 4: 4, 8: 8, 12: 12, 16: 12, 64: 12} {
		if got := checkerCountFor(cores); got != want {
			t.Errorf("%d cores: %d checkers, want %d", cores, got, want)
		}
	}
}
