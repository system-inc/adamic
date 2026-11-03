package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
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

// In a .ts program with records, a union or optional field read whose receiver a record may be, as far
// as the types say, tests the receiver's storage as it runs (lower's weakFieldRead): a fixed object
// reads its slot, a record its key, held to the field's type. Each fixture is one of tsc's seven weak
// reads in miniature, read off fixed objects, and agrees with Node natively and on the JavaScript
// backend. Adamic refuses a record seen as a weak type at the relation, so the record path is reached
// through the one road that gets past it (TestWeakReadRoads), a generic: generic_record reads a record's
// key and an absent one, and generic_open_enum a record's 37 under the whole enum Kind, open by the
// October 19 ruling. Each read is the weak_field_read helper, not another path, and its storage test and
// value check are each a checked site in the report --explain-checks prints. Mutant (b), the storage test
// dropped, turns generic_record red natively (a record read as a slot stops on "a record reached a
// fixed-object read"); mutant (d), the open-enum exemption dropped, turns generic_open_enum red.
func TestWeakReadAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"optional_plain", "optional_chain", "partial_member", "enum_union", "enum_open", "generic_record", "generic_open_enum"} {
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
			if sites["weak field read storage"] != helpers || sites["weak field read record value"] != helpers {
				t.Fatalf("checked sites %v for %d weak reads, want each one's storage test and value check", sites, helpers)
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

// weakReadStops holds a fixture Node runs to the end, printing printed, to a stop natively and on the
// JavaScript backend: exit 70 after stopped, the stop's message want.
func weakReadStops(t *testing.T, name, printed, stopped, want string) {
	t.Helper()
	path := weakReadPath(t, name)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the read must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != printed {
		t.Fatalf("Node: %s", weakReadShown(node))
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != stopped || string(got.stderr) != "adamic: panic: "+want+"\n" {
			t.Fatalf("%s: want %q printed, then a stop with %q, got %s", backend, stopped, want, weakReadShown(got))
		}
	}
}

// The literal-membership test tightens every dictionary read recordReadType holds, not only weak ones:
// a 3 a computed key put under mode, typed 1 | 2, read the ordinary way, now stops where native read
// it as the field and printed 3, as Node does. Mutant (c), the membership test dropped, prints 3.
func TestRecordLiteralMembershipStops(t *testing.T) {
	t.Parallel()
	weakReadStops(t, "literal_miss.ts", "read\n3\n", "read\n", "field read failed: options.mode is not a 1 | 2; expected 1 | 2, found number 3, which a computed-key or index-signature write may have stored")
}

// moduleNameResolver.ts's shape, packageJsonContent.exports, a field held as a Union (object | string),
// stays NotYet at the read: recordReadDeclared makes no value check for a Union-represented field yet,
// and through a weak type a record's values may be of any type, so a number would be read as
// object | string. A fixed-receiver twin isn't kept: the read is NotYet by its receiver's type, whatever
// the program passes. Mutant (e), the NotYet removed, admits the unchecked read, and this goes red.
func TestWeakReadUnionFieldStaysNotYet(t *testing.T) {
	t.Parallel()
	path := weakReadPath(t, "union_field.ts")
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+":10:16" {
		t.Fatalf("want NotYet %q at :10:16, got %v", recordGuardedReason, err)
	}
}

// The same read in a .a file, of a fixed object, stays NotYet with the reason it had before, at the read:
// its receiver's type is one a record's is assignable to. .a is to refuse a record seen as a weak type at
// the relation, so a read that reaches here in .a would be a hole in that rule.
func TestWeakReadStaysNotYetInA(t *testing.T) {
	t.Parallel()
	path := weakReadPath(t, "weak_read.a")
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+":10:16" {
		t.Fatalf("want NotYet %q at :10:16, got %v", recordGuardedReason, err)
	}
}

// roads/generic.ts is the road that reaches the record path: a type parameter constrained by the weak
// type, instantiated with the record's type, whose one body reads T by its constraint. These hold the
// record path through it, exit, stdout and stderr exactly, natively and on the JavaScript backend: a
// string under 1 | 2 stops on the value check, a 5 under Kind.A | Kind.B on its literal membership, and a
// class instance handed the same generic at the kind test. generic_record and generic_open_enum, which
// agree with Node, are TestWeakReadAgreesWithNode's. The mutants in proof.sh name these: (a), the value
// check dropped, wrong_type and wrong_literal; (b), the storage test dropped, class_instance (and
// generic_record natively); (c), the literal membership dropped, wrong_literal.
func TestWeakReadGenericStops(t *testing.T) {
	t.Parallel()
	for name, probe := range map[string]struct{ printed, stopped, want string }{
		"wrong_type":     {"total\nmode\nfast\n", "total\nmode\n", "field read failed: t.mode is not a 1 | 2; expected 1 | 2, found string, which a computed-key or index-signature write may have stored"},
		"wrong_literal":  {"total\nkind\n5\n", "total\nkind\n", "field read failed: t.kind is not a Kind.A | Kind.B; expected Kind.A | Kind.B, found number 5, which a computed-key or index-signature write may have stored"},
		"class_instance": {"total\nmode\n1\n2\n", "total\nmode\n1\n", "field read failed: t.mode reads T storage; expected a fixed object or a record, found a class instance"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			weakReadStops(t, "generic_"+name+".ts", probe.printed, probe.stopped, probe.want)
		})
	}
}

// The road hunt: each way a program might hand a record to a weak read, written as a program, Lower's
// list (json_parse through returned) and Language's (callback through union). generic reaches the record
// path, and the generic_ fixtures pin it; any other road that lowers with a weak_field_read in it reaches
// the record path through the language too, and fails loudly. Every other one must stop, Refused or NotYet; this first version asserts only that it stops,
// and logs the stop's whole text, to be pinned exactly next.
func TestWeakReadRoads(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"json_parse", "json_decode", "generic", "cast", "cast_unknown", "spread", "returned", "callback", "array_element", "array_view", "map_view", "field", "union"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, weakReadPath(t, filepath.Join("roads", name+".ts")))
			if name == "generic" {
				// The known road, which the generic_ fixtures pin.
				if err != nil {
					t.Fatalf("generic, the known road, must lower to the weak read: %v", err)
				}
				for _, function := range program.Functions {
					if function.Name == "weak_field_read" {
						return
					}
				}
				t.Fatal("generic, the known road, lowers with no weak_field_read")
			}
			if err == nil {
				for _, function := range program.Functions {
					if function.Name == "weak_field_read" {
						t.Fatalf("ROAD REACHES %s", name)
					}
				}
				t.Fatalf("%s lowers, with no weak read and no stop", name)
			}
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case errors.As(err, &refused):
				t.Logf("%s stops, Refused at %s: %s; fix: %s (%v)", name, refused.Where, refused.What, refused.Fix, err)
			case errors.As(err, &notYet):
				t.Logf("%s stops, NotYet at %s: %s (%v)", name, notYet.Where, notYet.What, err)
			default:
				t.Fatalf("%s stops with neither Refused nor NotYet: %v", name, err)
			}
		})
	}
}
