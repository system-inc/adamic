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

func realpathUseCountMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "realpath-use-mutant.test", "./internal/lower", nil, []buildcache.Mutation{{
		File:   "internal/lower/library_node_realpath_uses.go",
		Before: "if len(arguments) != 1 {",
		After:  "if len(arguments) == 0 {",
	}})
}

func TestProduct_NativeRealpathUseMutant(t *testing.T) { t.Parallel(); realpathUseCountMutant(t) }

func TestNodeRealpathNativeUseMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := realpathUseCountMutant(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run", `^TestNodeRealpathNativeClosedUses$/^two_arguments\.a$`, "-test.count=1", "-test.v")
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), "want native use refusal zero or multiple arguments") {
		t.Fatalf("argument-count mutant must red TestNodeRealpathNativeClosedUses/two_arguments.a: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("argument-count mutant caught: %s", stdout)
}
