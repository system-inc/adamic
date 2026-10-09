package corpusfiles

import (
	"testing"
	"time"
)

// Gate mutant 2 (#f3p1c0z): a test binary that dies partway. The watchdog panics the whole binary at 50 ms, the way
// TestCompilerAndStage1Agree_6484's did on Oct 9, so TestPlantedZAfter never reaches its failure. A gate that calls a
// partly run package clean (or excuses the panic as infrastructure) reads this green; a correct gate refuses it,
// because a test of the package never reached a terminal action. Never merge.
func TestPlantedWatchdog(t *testing.T) {
	t.Parallel()
	time.AfterFunc(50*time.Millisecond, func() { panic("TestPlantedWatchdog: own work exceeded 90s") })
	time.Sleep(time.Second)
}

func TestPlantedZAfter(t *testing.T) {
	t.Parallel()
	time.Sleep(2 * time.Second)
	t.Fatal("planted: this test must run for the package to be judged")
}
