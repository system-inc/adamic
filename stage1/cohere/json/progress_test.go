package json

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func TestProgressGuard(t *testing.T) {
	t.Parallel()
	policy := childguard.Options{FirstOutput: 500 * time.Millisecond, Stall: 200 * time.Millisecond, Ceiling: 5 * time.Second}
	for _, fixture := range []struct{ name, source, message string }{
		{"never prints", "exec sleep 20", "stalled: no first output"},
		{"never returns", "printf ready; exec sleep 20", "stalled: no output"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			command := exec.Command("sh", "-c", fixture.source)
			_, err := childguard.CombinedOutput(command, policy)
			if err == nil || !strings.Contains(err.Error(), fixture.message) || !strings.Contains(err.Error(), "sh") {
				t.Fatalf("guard failed by name: %v", err)
			}
			t.Logf("caught %s: %v", fixture.name, err)
		})
	}
	t.Run("progress outlasts startup", func(t *testing.T) {
		command := exec.Command("sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x; sleep 0.1; done")
		output, err := childguard.CombinedOutput(command, policy)
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("progress was killed: %q %v", output, err)
		}
	})
	t.Run("stderr is progress", func(t *testing.T) {
		command := exec.Command("sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x >&2; sleep 0.1; done")
		output, err := childguard.CombinedOutput(command, policy)
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("stderr progress was killed: %q %v", output, err)
		}
	})
	t.Run("overall backstop", func(t *testing.T) {
		short := policy
		short.Ceiling = 300 * time.Millisecond
		command := exec.Command("sh", "-c", "while :; do printf x; sleep 0.05; done")
		_, err := childguard.CombinedOutput(command, short)
		if err == nil || !strings.Contains(err.Error(), "ceiling:") {
			t.Fatalf("backstop survived: %v", err)
		}
	})
}
