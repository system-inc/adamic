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

// A host call native doesn't emit is refused by name before emission, as an ordinary error: a workers:
// module is JavaScript's alone (workers:KvNamespace.get), a Promise form settles only under
// JavaScript's await, and an unknown module once fell into node:fs. HostCallRefusal names each, once,
// wherever it sits in the program, and the emitter's panic stays behind it as the backstop.
// Mutant: HostCallRefusal's test dropped (TestNativeWorkersHostCallMutantIsCaught).
func TestNodeHostCallRefusesWorkersAndPromises(t *testing.T) {
	t.Parallel()
	program := func(call ir.NodeHostCall) *ir.Program {
		return &ir.Program{Functions: []ir.Function{{Name: "handle", Returns: ir.String, Body: []ir.Statement{
			ir.Declare{Local: 0, Value: call},
			ir.Return{Value: ir.Undefined{Of: ir.String}},
		}}}}
	}
	for _, call := range []ir.NodeHostCall{
		{Module: "workers:KvNamespace", Member: "get", Arguments: []ir.Expression{ir.Undefined{Of: ir.Object}, ir.Undefined{Of: ir.String}}, Returns: ir.Promise, Throws: true},
		{Module: "node:fs", Member: "readdirSync", Arguments: []ir.Expression{ir.Undefined{Of: ir.String}}, Returns: ir.Promise, Throws: true},
		{Module: "node:os", Member: "tmpdir", Returns: ir.String},
	} {
		want := "native: refused host operation " + call.Module + "." + call.Member + ": "
		if err := HostCallRefusal(program(call)); err == nil || !strings.HasPrefix(err.Error(), want) || !strings.Contains(err.Error(), "fix:") {
			t.Fatalf("native host call %s.%s admitted, want the named refusal %q, got %v", call.Module, call.Member, want, err)
		}
		if _, err := WASI(program(call), -1); err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("native host call %s.%s: WASI emitted it, want the named refusal, got %v", call.Module, call.Member, err)
		}
		if _, err := WASIExports(program(call), nil); err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("native host call %s.%s: WASIExports emitted it, want the named refusal, got %v", call.Module, call.Member, err)
		}
		backstop := func() (failure string) {
			defer func() {
				if recovered := recover(); recovered != nil {
					failure = fmt.Sprint(recovered)
				}
			}()
			e := &emitter{program: &ir.Program{}, scopes: [][]string{nil}}
			e.nodeHostCall(call)
			return ""
		}()
		if backstop != "native: compiler bug: no native emission for host operation "+call.Module+"."+call.Member {
			t.Fatalf("native host call %s.%s: the emitter's backstop panic is gone: %q", call.Module, call.Member, backstop)
		}
	}
	if err := HostCallRefusal(program(ir.NodeHostCall{Module: "node:path", Member: "dirname", Arguments: []ir.Expression{ir.Undefined{Of: ir.String}}, Returns: ir.String})); err != nil {
		t.Fatalf("native host call node:path.dirname refused: %v", err)
	}
}

func nativeWorkersHostCallMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "native-workers-host-call-mutant.test", "./internal/native", nil, []buildcache.Mutation{{
		File:   "internal/native/library_node_fs_directory.go",
		Before: `if call, isCall := value.Interface().(ir.NodeHostCall); isCall && !nativeHostCall(call) {`,
		After:  `if call, isCall := value.Interface().(ir.NodeHostCall); isCall && false && !nativeHostCall(call) {`,
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
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), "native host call workers:KvNamespace.get admitted") {
		t.Fatalf("native refusal mutant must red TestNodeHostCallRefusesWorkersAndPromises: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("native refusal mutant caught: %s", stdout)
}
