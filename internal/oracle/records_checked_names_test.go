package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// A bare-name CheckedFields entry, a record's field (readiness's markRecord) or a field a class merged
// with an interface promises (mergedFieldRead), checks every read and write of that name, except through
// a receiver lowering proved can hold neither (ir.Program.PlainReceivers, lower's plainReceiver,
// #9jnjwrw). These hold the rule both ways, natively and on the JavaScript backend.

// checkedNamesPath is a fixture of records_checked_names, absolute, as a stop's Where names it.
func checkedNamesPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_checked_names", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// checkedNamesShown is a run as a failure message shows it.
func checkedNamesShown(got run) string {
	return fmt.Sprintf("exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
}

// #9jnjwrw's crash: a class instance read through a generic instantiated with its class, of a name a
// record writes, was emitted as a checked view field, which an optional number read can't be, and the
// compiler panicked ("incomplete checked field contract"). Settings is plain, so the read is a slot read
// with no view, and agrees with Node.
func TestCheckedNamesClassThroughGeneric(t *testing.T) {
	t.Parallel()
	path := checkedNamesPath(t, "class_through_generic.ts")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("must lower: %v", err)
	}
	viewed := 0
	look := func(expression ir.Expression) ir.Expression {
		if read, ok := expression.(ir.Property); ok && read.Name == "mode" && read.View != "" {
			viewed++
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), look)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), look)
	if viewed != 0 {
		t.Fatalf("%d reads of mode keep a view through a plain receiver", viewed)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "total\nmode\n2\n" {
		t.Fatalf("Node: %s", checkedNamesShown(node))
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, checkedNamesShown(node), backend, checkedNamesShown(got))
		}
	}
}

// Adamic's structural case: TypeScript lets an object literal stand where Settings is declared, so the
// receiver proof is never a claim that a value is the class's. The program refuses the literal at the
// declaration: a class type takes only what its class constructs.
func TestCheckedNamesStructuralLiteralRefused(t *testing.T) {
	t.Parallel()
	path := checkedNamesPath(t, "structural_literal.ts")
	_, err := lowered(t, path)
	const want = ":16:27: Adamic 0.1 refuses a value without nominal ancestry seen as Settings; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"
	var refused *lower.Refused
	if !errors.As(err, &refused) || err.Error() != path+want {
		t.Fatalf("want the literal refused, %q, got %v", want, err)
	}
}

// checkedNamesStops holds a fixture Node runs to the end, printing printed, to a stop natively and on
// the JavaScript backend: exit 70 after stopped, the stderr each backend's.
func checkedNamesStops(t *testing.T, name, printed, stopped, nativeStop, javascriptStop string) {
	t.Helper()
	path := checkedNamesPath(t, name)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != printed {
		t.Fatalf("Node: %s", checkedNamesShown(node))
	}
	native, _ := natively(t, program)
	for backend, probe := range map[string]struct {
		got  run
		stop string
	}{"native": {native, nativeStop}, "javascript": {onJavaScriptBackend(t, program), javascriptStop}} {
		want := run{stdout: []byte(stopped), stderr: []byte("adamic: panic: " + probe.stop + "\n"), exitCode: 70}
		if difference := disagreement(want, probe.got); difference != "" {
			t.Fatalf("%s: %s: want %s; got %s", backend, difference, checkedNamesShown(want), checkedNamesShown(probe.got))
		}
	}
}

// A record read through a widened view of its own type stays checked, since a record can be a Seen,
// and stops on the string a computed key put under total. Each backend stops in its own words, as its
// checked view read does. Mutant (f), every receiver plain, reads the record as a slot: native stops on
// "a record reached a fixed-object read", and the JavaScript backend prints seven1.
func TestCheckedNamesWidenedViewLie(t *testing.T) {
	t.Parallel()
	checkedNamesStops(t, "widened_view_lie.ts", "total\nseven1\n", "total\n",
		"field read failed: seen.total expected number, found string",
		"field read failed: seen.total is not a number; expected number, found string")
}

// A field an interface merged with a class promises and the class doesn't initialize stays checked
// through the merged class, and stops on the missing field where Node reads undefined. Mutant (f) reads
// it as a slot: native stops on "a field the checker proved is there is missing", and the JavaScript
// backend prints undefined.
func TestCheckedNamesMergedField(t *testing.T) {
	t.Parallel()
	const stop = "field read failed: box.label is not initialized; expected string, found missing"
	checkedNamesStops(t, "merged_field.ts", "1\nundefined\n", "1\n", stop, stop)
}
