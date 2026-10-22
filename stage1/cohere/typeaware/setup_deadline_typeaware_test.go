package typeaware

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Optional entry point: every selected shard prepares the same plan itself.
func TestTypeAwareAgreementAndMutants_Setup(t *testing.T) {
	t.Parallel()
	typeAwareTopPlan(t, "typeaware")
}

// A real nested subprocess must be killed before it can publish its marker.
// This exercises the unchanged own-work watchdog used by check shards.
// Not parallel: deadline probe shares the package command-group registry.
func TestTypeAwareUnitDeadline(t *testing.T) {
	mode := os.Getenv("ADAMIC_TYPEAWARE_DEADLINE_CHILD")
	marker := os.Getenv("ADAMIC_TYPEAWARE_DEADLINE_MARKER")
	if mode == "marker" {
		time.Sleep(time.Second)
		if err := os.WriteFile(marker, []byte("survived"), 0600); err != nil {
			t.Fatal(err)
		}
		return
	}
	if mode != "" {
		defer typeAwareDeadlineWithin(t, "watchdog-probe", 100*time.Millisecond, 200*time.Millisecond)()
		if mode == "quick" {
			return
		}
		command := exec.Command(os.Args[0], "-test.run=^TestTypeAwareUnitDeadline$")
		command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_DEADLINE_CHILD=marker")
		if err := typeAwareRunCommand(command); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, mode := range []string{"quick", "slow"} {
		t.Run(mode, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), "marker")
			ctx, cancel := context.WithTimeout(context.Background(), typeAwareChildLimit)
			defer cancel()
			command := typeAwareContextCommand(ctx, os.Args[0], "-test.run=^TestTypeAwareUnitDeadline$", "-test.v")
			command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_DEADLINE_CHILD="+mode, "ADAMIC_TYPEAWARE_DEADLINE_MARKER="+marker)
			// Read through childguard.Capture, to its end with no WaitDelay deciding when. Its group kill can't reach
			// the marker child, which typeAwareRunCommand starts in a group of its own, so only cooked's kill stops it.
			stdout, stderr, err := childguard.Capture(ctx, command)
			output := append(stdout, stderr...)
			if mode == "quick" {
				if err != nil || !bytes.Contains(output, []byte("cooked=false")) {
					t.Fatalf("quick unit: %v %s", err, output)
				}
				return
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !bytes.Contains(output, []byte("unit deadline exceeded name=watchdog-probe")) {
				t.Fatalf("slow unit was not cooked: %v %s", err, output)
			}
			time.Sleep(time.Second)
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("cooked unit left its subprocess alive: %v", err)
			}
		})
	}
}
