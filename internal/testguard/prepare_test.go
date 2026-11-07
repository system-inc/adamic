package testguard

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func TestPrepareCPUWithChildGuard(test *testing.T) {
	test.Parallel()
	command := exec.CommandContext(test.Context(), "/bin/sh", "-c", "ulimit -t")
	var output bytes.Buffer
	command.Stdout = &output
	if err := PrepareCPU(command, 7*time.Second); err != nil {
		test.Fatal(err)
	}
	if err := childguard.Run(
		command,
		childguard.Options{FirstOutput: Ceiling, Stall: Ceiling, Ceiling: Ceiling},
	); err != nil {
		test.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != "7" {
		test.Fatalf("CPU limit = %q, want 7 seconds", got)
	}
}
