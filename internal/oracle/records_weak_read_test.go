package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// weakReadPath is a fixture of records_weak_read, absolute, as a NotYet's Where names it.
func weakReadPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_weak_read", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// weakReadShown is a run as a failure message shows it.
func weakReadShown(got run) string {
	return fmt.Sprintf("exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
}

// In a .ts program with records, a union or optional field read whose receiver a record may be
// tests the receiver's storage as it runs (lower's weakFieldRead): a fixed object reads its slot, a
// record its key, held to the field's type. Each fixture is one of tsc's seven weak reads in
// miniature, read once off a fixed object and once off a record laundered into the weak type, and
// agrees with Node natively and on the JavaScript backend. Each read is the weak_field_read helper,
// not another path, and its storage test (and, for a field the record's union can lie about, its value
// check) is a checked site in the report --explain-checks prints.
// Mutant (b), the storage test dropped so every receiver is read as a slot: the record rows stop
// natively on the runtime's "a record reached a fixed-object read".
func TestWeakReadAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, valueChecked := range map[string]bool{"optional_plain": true, "optional_chain": true, "union_field": false, "partial_member": true, "enum_union": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := weakReadPath(t, name+".ts")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a weak read in .ts must lower: %v", err)
			}
			helpers := 0
			for _, function := range program.Functions {
				if function.Name == "weak_field_read" {
					helpers++
				}
			}
			if helpers == 0 {
				t.Fatal("no weak_field_read: the read took another path than the storage test")
			}
			sites := map[string]int{}
			for _, site := range program.PlaceholderChecks {
				if site.Origin == "record storage" && site.Status == "checked" {
					sites[site.Use]++
				}
			}
			if sites["weak field read storage"] != helpers || valueChecked != (sites["weak field read record value"] == helpers) {
				t.Fatalf("checked sites %v for %d weak reads, want each storage test and, value checked %t, each value check", sites, helpers, valueChecked)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: %s", weakReadShown(node))
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, weakReadShown(node), backend, weakReadShown(got))
				}
			}
		})
	}
}

// A record holding a string under a field typed 1 | 2 stops at the read, naming it, with recordReadType's
// own check, natively and on the JavaScript backend; Node prints the string. Mutant (a), the record's
// value read without the check: native takes the string for a number, the JavaScript backend prints
// it, and neither stops.
func TestWeakReadWrongValueStops(t *testing.T) {
	t.Parallel()
	path := weakReadPath(t, "wrong_type.ts")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the weak read must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "total\nmode\n2\nfast\n" {
		t.Fatalf("Node: %s", weakReadShown(node))
	}
	const want = "field read failed: options.mode is not a 1 | 2; expected 1 | 2, found string"
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "total\nmode\n2\n" || !strings.Contains(string(got.stderr), want) {
			t.Fatalf("%s: want the fixed read, then a stop with %q, got %s", backend, want, weakReadShown(got))
		}
	}
}

// A class instance seen as the weak type is neither a fixed object nor a record: the read stops by name
// before any slot is read, natively and on the JavaScript backend, where Node reads the instance's field.
// Mutant (b) reads it as a slot, and this goes red too. An array isn't pinned: tsc's weak-type check
// refuses one for a weak type that shares none of its properties.
func TestWeakReadOtherHeapStops(t *testing.T) {
	t.Parallel()
	path := weakReadPath(t, "class_instance.ts")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the weak read must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "total\nmode\n1\n2\n" {
		t.Fatalf("Node: %s", weakReadShown(node))
	}
	const want = "field read failed: options.mode reads Options storage; expected a fixed object or a record, found a class instance"
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "total\nmode\n1\n" || !strings.Contains(string(got.stderr), want) {
			t.Fatalf("%s: want the record read, then a stop with %q, got %s", backend, want, weakReadShown(got))
		}
	}
}

// The same read in a .a file stays NotYet with the reason it had before, at the read. .a is to refuse a
// record seen as a weak type at the relation, so a read that reaches here in .a would be a hole in that
// rule.
func TestWeakReadStaysNotYetInA(t *testing.T) {
	t.Parallel()
	path := weakReadPath(t, "weak_read.a")
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+":9:16" {
		t.Fatalf("want NotYet %q at :9:16, got %v", recordGuardedReason, err)
	}
}
