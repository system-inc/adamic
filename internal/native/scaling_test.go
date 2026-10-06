package native

import (
	"os/exec"
	"strings"
	"testing"
)

func TestParallelScaling(t *testing.T) {
	for _, build := range parallelBuilds(t) {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "scaling.c", build.options)
			for _, threads := range []string{"1", "4"} {
				out, _ := parallelRun(t, binary, threads)
				if out != "scaling clean\n" {
					t.Fatal(out)
				}
				parallelLeaks(t, binary, build.options, threads)
			}
			command := exec.Command(binary, "high_pointer")
			command.Env = parallelEnvironment("1", false)
			out, err := command.CombinedOutput()
			failure, ok := err.(*exec.ExitError)
			if !ok || failure.ExitCode() != 70 || !strings.Contains(string(out), "shape address exceeds 48 bits") {
				t.Fatalf("pointer guard: %v: %s", err, out)
			}
		})
	}
}
