package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/system-inc/adamic/internal/testguard"
)

// Only a prepared leaf is launched here. Product recipes and setup never use
// this guard; each child receives its own group, leaving neighbouring units alone.
func testShardsAgreeChild(test *testing.T, name string, markers ...string) {
	test.Helper()
	command := exec.CommandContext(
		context.Background(),
		os.Args[0],
		"-test.run=^"+name+"$",
		"-test.timeout=0",
		"-test.v",
	)
	command.Env = append(os.Environ(), markers...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := testguard.Run(command, testguard.Budget, testguard.Ceiling)
	test.Logf("%s", &output)
	if err != nil {
		test.Fatalf("%s: %v", name, err)
	}
}
