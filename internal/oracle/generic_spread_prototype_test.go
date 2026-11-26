package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A spread copies own properties only, so a class instance's methods, getters and setters, its
// prototype's, aren't in the copy, while a spread of a type parameter's value is typed as that type
// parameter, so a generic function instantiated with a class types the copy as the class (#y4gjeb8). Main
// lowered the reads anyway: generic_method_call printed 20 natively where Node throws, generic_method_typeof
// read function on both backends where Node reads undefined, and generic_setter stopped natively on a compiler
// bug. Each now stops at compile time at the read, naming the member, the spread and the class
// (checkGenericSpreadCopyReads, internal/lower/generic_spread_copy_reads.go), and so does generic_getter, which
// spreadCopyRead stopped already. private_read is the call road's memory hazard: its weigh reads a private field
// and a data field the copy lays out elsewhere, so main printed 14 natively, ASan silent, where Node throws, and it
// stops at compile time too. same (the generic returning its argument as is), plain_source (a generic spread of a
// plain literal, whose weigh is its own), class_parameter_call (a method call on a class-typed parameter in a
// program with no generic spread) and control_instance (private_read's class, a real instance through the same
// generic) lower and agree with Node. Mutants (internal/lower/generic_spread_copy_reads.mutants.md): the stop
// answering nil lowers generic_method_call and private_read again, and private_read then dies natively on the
// runtime backstop's invariant (adamic_method_this), printing 14 only with the backstop dropped too.
func TestGenericSpreadPrototypeMember(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, stop string }{
		{"generic_method_call.ts", "threw\n", "generic_method_call.ts:16:21: " + genericSpreadReadStop("weigh", "Scale", "method")},
		{"generic_method_typeof.ts", "t undefined\n", "generic_method_typeof.ts:15:40: " + genericSpreadReadStop("weigh", "Scale", "method")},
		{"generic_setter.ts", "undefined undefined\n", "generic_setter.ts:15:20: " + genericSpreadReadStop("weight", "Sized", "setter")},
		{"generic_getter.ts", "undefined undefined\n", "generic_getter.ts:15:20: " + genericSpreadReadStop("weight", "Sized", "getter")},
		{"private_read.ts", "threw\n", "private_read.ts:18:21: " + genericSpreadReadStop("weigh", "Scale", "method")},
		{"generic_same.ts", "20\n", ""},
		{"generic_plain_source.ts", "x 5\n", ""},
		{"class_parameter_call.ts", "40\n", ""},
		{"control_instance.ts", "20\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_spread_prototype", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.stop) || !strings.Contains(err.Error(), probe.name) {
					t.Fatalf("want a not yet with %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// genericSpreadReadStop is the words a read of name from the probes' copy, { ...source }, stops with.
func genericSpreadReadStop(name, class, kind string) string {
	return "stage 0 can't lower a read of " + name + " from { ...source }, which may be a generic spread's copy of a " + class + " instance, whose " + name + " is a " + kind + " on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does (#y4gjeb8) yet"
}
