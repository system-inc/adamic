package testguard

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func TestPrepareCPUWithChildGuard(t *testing.T) {
	t.Parallel()
	command := exec.CommandContext(t.Context(), "/bin/sh", "-c", "ulimit -t")
	var output bytes.Buffer
	command.Stdout = &output
	if err := PrepareCPU(command, 7*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := childguard.Run(command, childguard.Options{FirstOutput: Ceiling, Stall: Ceiling, Ceiling: Ceiling}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != "7" {
		t.Fatalf("CPU limit = %q, want 7 seconds", got)
	}
}
