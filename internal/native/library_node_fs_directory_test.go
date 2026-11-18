package native

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
)

// A host call native doesn't bind is a compiler bug said out loud, named by its module and member,
// before any branch reads it: a workers: module is JavaScript's alone (workers:KvNamespace.get), a
// Promise form settles only under JavaScript's await, and an unknown module once fell into node:fs.
// Mutant: the refusal dropped (TestNativeWorkersHostCallMutantIsCaught).
func TestNodeHostCallRefusesWorkersAndPromises(t *testing.T) {
	t.Parallel()
	emits := func(call ir.NodeHostCall) (failure string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				failure = fmt.Sprint(recovered)
			}
		}()
		e := &emitter{program: &ir.Program{}, scopes: [][]string{nil}}
		e.nodeHostCall(call)
		return ""
	}
	for _, call := range []ir.NodeHostCall{
		{Module: "workers:KvNamespace", Member: "get", Arguments: []ir.Expression{ir.Undefined{Of: ir.Object}, ir.Undefined{Of: ir.String}}, Returns: ir.Promise, Throws: true},
		{Module: "node:fs", Member: "readdirSync", Arguments: []ir.Expression{ir.Undefined{Of: ir.String}}, Returns: ir.Promise, Throws: true},
		{Module: "node:os", Member: "tmpdir", Returns: ir.String},
	} {
		want := "native: compiler bug: no native emission for host operation " + call.Module + "." + call.Member
		if failure := emits(call); failure != want {
			t.Fatalf("native host call %s.%s emitted, want the named refusal %q, got %q", call.Module, call.Member, want, failure)
		}
	}
}

func nativeWorkersHostCallMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "native-workers-host-call-mutant.test", "./internal/native", nil, []buildcache.Mutation{{
		File:   "internal/native/library_node_fs_directory.go",
		Before: `if (call.Module != "node:path" && call.Module != "node:fs") || call.Returns == ir.Promise {`,
		After:  `if false {`,
	}})
}

func TestProduct_NativeWorkersHostCallMutant(t *testing.T) {
	t.Parallel()
	nativeWorkersHostCallMutant(t)
}

func TestNativeWorkersHostCallMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := nativeWorkersHostCallMutant(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run", `^TestNodeHostCallRefusesWorkersAndPromises$`, "-test.count=1", "-test.v")
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), "native host call workers:KvNamespace.get emitted") {
		t.Fatalf("native refusal mutant must red TestNodeHostCallRefusesWorkersAndPromises: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("native refusal mutant caught: %s", stdout)
}
