package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestCheckedViewCallableFactoryHigherOrder(t *testing.T) {
	signature := "(hint: number, node: Node, emitCallback: (hint: number, node: Node) => void) => void"
	for _, variant := range []string{"good", "boundary", "wrong-arity", "wrong-nested"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/share-factory/rank-165/"+variant)
			truth := onNode(t, path)
			want := "done\n"
			if variant == "boundary" {
				want = "8\ndone\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != want || len(truth.stderr) != 0 {
				t.Fatalf("source Node: %#v want %q", truth, want)
			}
			sanitized, binary := nativelyUncached(t, program)
			if variant == "wrong-nested" {
				t.Logf("sanitized nested exit=%d stdout=%q stderr=%q", sanitized.exitCode, sanitized.stdout, sanitized.stderr)
			}
			for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if variant == "good" || variant == "boundary" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatalf("backend%d: %s", index, difference)
					}
				} else {
					found := "function with incompatible parameter representations"
					if variant == "wrong-arity" {
						found = "function with arity 0"
					}
					message := "adamic: panic: field read failed: context.onEmitNode expected " + signature + ", found " + found + "\n"
					t.Logf("backend%d negative exit=%d stdout=%q stderr=%q", index, got.exitCode, got.stdout, got.stderr)
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
						t.Fatalf("higher-order negative did not stop: %#v want %q", got, message)
					}
				}
			}
			if variant == "good" || variant == "boundary" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				if variant == "boundary" {
					higherOrderArgumentMutant(t, program)
				}
				return
			}
			if variant == "wrong-arity" {
				changed := 0
				for index := range program.ViewContracts {
					contract := &program.ViewContracts[index]
					if contract.Kind == ir.ViewCallable && len(contract.Parameters) == 3 {
						contract.Parameters = nil
						changed++
					}
				}
				if changed != 1 {
					t.Fatalf("arity mutant changed %d", changed)
				}
			} else {
				// The direct nested-comparison source mutants are run by run-nested-mutants.py.
				return
			}
			mutantSanitized, mutantBinary := nativelyUncached(t, program)
			for index, got := range []run{releasedUncached(t, program), mutantSanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("mutant%d must execute cleanly: %s", index, difference)
				}
				t.Logf("backend%d %s omission mutant caught: stdout %q instead of exit 70", index, variant, got.stdout)
			}
			if report := leaksUncached(t, program, mutantBinary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func higherOrderArgumentMutant(t *testing.T, program *ir.Program) {
	t.Helper()
	bad := -1
	for index, function := range program.Functions {
		if function.Closure && len(function.Parameters) == 0 {
			bad = index
		}
	}
	if bad < 0 {
		t.Fatal("missing boundary producer")
	}
	changed := 0
	rewrite := func(value any) any {
		if call, ok := value.(ir.CallClosure); ok && len(call.CallableArguments) == 3 && call.CallableArguments[2] != 0 {
			call.Arguments[2] = ir.MakeClosure{Function: bad}
			changed++
			return call
		}
		return value
	}
	program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
	for index := range program.Functions {
		program.Functions[index].Body = rewriteUnionTargetStatements(program.Functions[index].Body, rewrite)
	}
	if changed != 1 {
		t.Fatalf("boundary witness changed %d calls", changed)
	}
	sanitized, _ := nativelyUncached(t, program)
	message := "adamic: panic: field read failed: call argument 3 expected (hint: number, node: Node) => void, found function with arity 0\n"
	for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
			t.Fatalf("callback argument negative%d: %#v want %q", index, got, message)
		}
	}
	rewrite = func(value any) any {
		if call, ok := value.(ir.CallClosure); ok && len(call.CallableArguments) == 3 {
			call.CallableArguments = nil
			return call
		}
		return value
	}
	program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
	for index := range program.Functions {
		program.Functions[index].Body = rewriteUnionTargetStatements(program.Functions[index].Body, rewrite)
	}
	mutantSanitized, binary := nativelyUncached(t, program)
	for index, got := range []run{releasedUncached(t, program), mutantSanitized, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "wrong\ndone\n" {
			t.Fatalf("boundary mutant%d did not execute cleanly: %#v", index, got)
		}
		t.Logf("backend%d argument descriptor omission caught: stdout %q instead of exit 70", index, got.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
