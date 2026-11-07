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
	if prepareCpuError := PrepareCPU(command, 7*time.Second); prepareCpuError != nil {
		test.Fatal(prepareCpuError)
	}
	if runError := childguard.Run(
		command,
		childguard.Options{FirstOutput: Ceiling, Stall: Ceiling, Ceiling: Ceiling},
	); runError != nil {
		test.Fatal(runError)
	}
	if got := strings.TrimSpace(output.String()); got != "7" {
		test.Fatalf("CPU limit = %q, want 7 seconds", got)
	}
}
