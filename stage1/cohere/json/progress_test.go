package json

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// jsonGuard is json's one policy for every child it runs: childguard's 30-minute first output and
// 60-minute ceiling, and a 5-minute stall, json's proven policy before childguard. The slowest isolated
// ASan case takes about 24 s on an idle 5-core box and about 4x that under load, so the 2-minute
// default would leave 1.25x headroom on a loaded gate box.
var jsonGuard = childguard.Options{Stall: 5 * time.Minute}

func TestProgressGuard(t *testing.T) {
	t.Parallel()
	// Each case leaves only the window it tests able to fire (the others an hour), and a talking child's
	// gaps are a small fraction of its windows, so load can't decide a verdict: at load 827 on Cloud a
	// 200 ms stall window killed the talking child (Oct 8, 293f6c9d).
	never := time.Hour
	for _, fixture := range []struct {
		name, source, message string
		policy                childguard.Options
	}{
		{"never prints", "exec sleep 600", "stalled: no first output", childguard.Options{FirstOutput: 5 * time.Second, Stall: never, Ceiling: never}},
		{"never returns", "printf ready; exec sleep 600", "stalled: no output", childguard.Options{FirstOutput: never, Stall: 5 * time.Second, Ceiling: never}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			command := exec.Command("sh", "-c", fixture.source)
			_, err := childguard.CombinedOutput(command, fixture.policy)
			if err == nil || !strings.Contains(err.Error(), fixture.message) || !strings.Contains(err.Error(), "sh") {
				t.Fatalf("guard failed by name: %v", err)
			}
			t.Logf("caught %s: %v", fixture.name, err)
		})
	}
	// Talking for 4 s against a 3 s first-output window and a 60 s stall window: the first-output window
	// must stop counting once the child talks.
	talking := childguard.Options{FirstOutput: 3 * time.Second, Stall: time.Minute, Ceiling: never}
	t.Run("progress outlasts startup", func(t *testing.T) {
		t.Parallel()
		command := exec.Command("sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x; sleep 0.5; done")
		output, err := childguard.CombinedOutput(command, talking)
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("progress was killed: %q %v", output, err)
		}
	})
	t.Run("stderr is progress", func(t *testing.T) {
		t.Parallel()
		command := exec.Command("sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x >&2; sleep 0.5; done")
		output, err := childguard.CombinedOutput(command, talking)
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("stderr progress was killed: %q %v", output, err)
		}
	})
	t.Run("overall backstop", func(t *testing.T) {
		t.Parallel()
		command := exec.Command("sh", "-c", "while :; do printf x; sleep 0.05; done")
		_, err := childguard.CombinedOutput(command, childguard.Options{FirstOutput: never, Stall: never, Ceiling: time.Second})
		if err == nil || !strings.Contains(err.Error(), "ceiling:") {
			t.Fatalf("backstop survived: %v", err)
		}
	})
}
