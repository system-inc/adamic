package lower

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

func realpathUseCountMutant(test *testing.T) string {
	test.Helper()
	return buildcache.GoTestMutated(
		test,
		"",
		"realpath-use-mutant.test",
		"./internal/lower",
		nil,
		[]buildcache.Mutation{{
			File:   "internal/lower/library_node_realpath_uses.go",
			Before: "if len(arguments) != 1 {",
			After:  "if false {",
		}},
	)
}

func TestProduct_NativeRealpathUseMutant(test *testing.T) {
	test.Parallel()
	realpathUseCountMutant(test)
}

func TestNodeRealpathNativeUseMutantIsCaught(test *testing.T) {
	test.Parallel()
	binary := realpathUseCountMutant(test)
	ctx, cancel := context.WithTimeout(test.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		binary,
		"-test.run",
		`^TestNodeRealpathNativeClosedUses$/^two_arguments\.a$`,
		"-test.count=1",
		"-test.v",
	)
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if !errors.As(captureError, &exited) ||
		exited.ExitCode() != 1 ||
		!strings.Contains(string(stdout), "want native use refusal zero or multiple arguments") {
		test.Fatalf(
			"argument-count mutant must red TestNodeRealpathNativeClosedUses/two_arguments.a: err=%v stdout=%s stderr=%s",
			captureError,
			stdout,
			stderr,
		)
	}
	test.Logf("argument-count mutant caught: %s", stdout)
}
