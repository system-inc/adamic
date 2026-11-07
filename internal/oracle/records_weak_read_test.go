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

// weakReadPath is a fixture of records_weak_read, absolute, as a NotYet's Where names it.
func weakReadPath(test *testing.T, name string) string {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_weak_read", name),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
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
// backend. Each read is the weak_field_read helper, not another path, and its storage test and value
// check are each a checked site in the report --explain-checks prints. The record path is pinned below
// the language, by TestWeakReadRecordPath.
func TestWeakReadAgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"optional_plain", "optional_chain", "partial_member", "enum_union", "enum_open"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := weakReadPath(test, name+".ts")
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("a weak read in .ts must lower: %v", loweredError)
			}
			helpers := 0
			for _, function := range program.Functions {
				if function.Name == "weak_field_read" {
					helpers++
				}
			}
			if helpers == 0 {
				test.Fatal("no weak_field_read: the read took another path than the storage test")
			}
			sites := map[string]int{}
			for _, site := range program.PlaceholderChecks {
				if site.Origin == "record storage" && site.Status == "checked" {
					sites[site.Use]++
				}
			}
			if sites["weak field read storage"] != helpers || sites["weak field read record value"] != helpers {
				test.Fatalf(
					"checked sites %v for %d weak reads, want each one's storage test and value check",
					sites,
					helpers,
				)
			}
			node := onNode(test, path)
			if node.exitCode != 0 {
				test.Fatalf("Node: %s", weakReadShown(node))
			}
			native, _ := natively(test, program)
			for backend, got := range map[string]run{
				"native":     native,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(node, got); difference != "" {
					test.Fatalf(
						"%s disagrees with Node (%s): Node %s; %s %s",
						backend,
						difference,
						weakReadShown(node),
						backend,
						weakReadShown(got),
					)
				}
			}
		})
	}
}

// weakReadStops holds a fixture Node runs to the end, printing printed, to a stop natively and on the
// JavaScript backend: exit 70 after stopped, the stop's message want.
func weakReadStops(test *testing.T, name, printed, stopped, want string) {
	test.Helper()
	path := weakReadPath(test, name)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("the read must lower: %v", loweredError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != printed {
		test.Fatalf("Node: %s", weakReadShown(node))
	}
	native, _ := natively(test, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 70 || string(got.stdout) != stopped || string(got.stderr) != "adamic: panic: "+want+"\n" {
			test.Fatalf("%s: want %q printed, then a stop with %q, got %s", backend, stopped, want, weakReadShown(got))
		}
	}
}

// The literal-membership test tightens every dictionary read recordReadType holds, not only weak ones:
// a 3 a computed key put under mode, typed 1 | 2, read the ordinary way, now stops where native read
// it as the field and printed 3, as Node does. Mutant (c), the membership test dropped, prints 3.
func TestRecordLiteralMembershipStops(test *testing.T) {
	test.Parallel()
	weakReadStops(
		test,
		"literal_miss.ts",
		"read\n3\n",
		"read\n",
		"field read failed: options.mode is not a 1 | 2; expected 1 | 2, found number 3, which a computed-key or index-signature write may have stored",
	)
}

// moduleNameResolver.ts's shape, packageJsonContent.exports, a field held as a Union (object | string),
// stays NotYet at the read: recordReadDeclared makes no value check for a Union-represented field yet,
// and through a weak type a record's values may be of any type, so a number would be read as
// object | string. A fixed-receiver twin isn't kept: the read is NotYet by its receiver's type, whatever
// the program passes. Mutant (e), the NotYet removed, admits the unchecked read, and this goes red.
func TestWeakReadUnionFieldStaysNotYet(test *testing.T) {
	test.Parallel()
	path := weakReadPath(test, "union_field.ts")
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+":10:16" {
		test.Fatalf("want NotYet %q at :10:16, got %v", recordGuardedReason, loweredError)
	}
}

// The same read in a .a file, of a fixed object, stays NotYet with the reason it had before, at the read:
// its receiver's type is one a record's is assignable to. .a is to refuse a record seen as a weak type at
// the relation, so a read that reaches here in .a would be a hole in that rule.
func TestWeakReadStaysNotYetInA(test *testing.T) {
	test.Parallel()
	path := weakReadPath(test, "weak_read.a")
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+":10:16" {
		test.Fatalf("want NotYet %q at :10:16, got %v", recordGuardedReason, loweredError)
	}
}

// weakReadRecordCase is one rewrite of record_path.ts: the call to function gets, in place of its
// fixed object, a record made of fields, or the class instance settings, and the program then prints
// stdout and stops with panic, or runs to the end when panic is empty.
type weakReadRecordCase struct {
	function string
	fields   func(program *ir.Program) []ir.Field
	stdout   string
	panic    string
}

// The record path, below the language. Adamic refuses a record seen as a weak type at every relation
// TestWeakReadRoads tries, and #xywvn06 closes the one that got past, a generic's type argument, so no
// program is to hand a record to a weak read, and the record path stays as a second line of defense. This lowers
// record_path.ts, whose reads are of fixed objects, then rewrites one call's argument in the IR to a
// record table, an entries record literal as a const table is made, or to a class instance, and runs it
// natively and on the JavaScript backend, exit, stdout and stderr exact. A record takes the record branch and reads its
// key, an absent key reads undefined, the value check stops a wrong type and a literal outside a strict subset
// of an enum's members, a whole numeric enum takes any number (it's open, the October 19 ruling), and a
// class instance stops at the kind test. Each mutant in proof.sh names the case it turns red: (a) the
// value check dropped, wrong_type and wrong_literal; (b) the storage test dropped, record_key (native
// reads a record as a slot and stops on "a record reached a fixed-object read") and class_instance;
// (c) the literal membership dropped, wrong_literal; (d) the open-enum exemption dropped, open_enum.
func TestWeakReadRecordPath(test *testing.T) {
	test.Parallel()
	number := func(value float64) ir.Expression { return ir.Box{Value: ir.NumberConstant{Value: value}} }
	for name, probe := range map[string]weakReadRecordCase{
		"fixed": {stdout: "total\nmode\n3\n2\n1\n5\n"},
		"record_key": {function: "readMode", fields: func(*ir.Program) []ir.Field {
			return []ir.Field{{Name: "mode", Value: number(1)}}
		}, stdout: "total\nmode\n3\n1\n1\n5\n"},
		"absent_key": {function: "readMode", fields: func(*ir.Program) []ir.Field {
			return []ir.Field{{Name: "total", Value: number(2)}}
		}, stdout: "total\nmode\n3\nundefined\n1\n5\n"},
		"wrong_type": {
			function: "readMode",
			fields: func(program *ir.Program) []ir.Field {
				program.Strings = append(program.Strings, "fast")
				return []ir.Field{{Name: "mode", Value: ir.StringConstant{Index: len(program.Strings) - 1}}}
			},
			stdout: "total\nmode\n3\n",
			panic:  "field read failed: options.mode is not a 1 | 2; expected 1 | 2, found string, which a computed-key or index-signature write may have stored",
		},
		"wrong_literal": {
			function: "readKind",
			fields: func(*ir.Program) []ir.Field {
				return []ir.Field{{Name: "kind", Value: number(5)}}
			},
			stdout: "total\nmode\n3\n2\n",
			panic:  "field read failed: options.kind is not a Kind.A | Kind.B; expected Kind.A | Kind.B, found number 5, which a computed-key or index-signature write may have stored",
		},
		"open_enum": {function: "readOpen", fields: func(*ir.Program) []ir.Field {
			return []ir.Field{{Name: "open", Value: number(37)}}
		}, stdout: "total\nmode\n3\n2\n1\n37\n"},
		"class_instance": {
			function: "readMode",
			stdout:   "total\nmode\n3\n",
			panic:    "field read failed: options.mode reads Options storage; expected a fixed object or a record, found a class instance",
		},
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := weakReadPath(test, "record_path.ts")
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("record_path.ts must lower: %v", loweredError)
			}
			if name == "fixed" {
				if node := onNode(test, path); node.exitCode != 0 || string(node.stdout) != probe.stdout {
					test.Fatalf("Node: %s", weakReadShown(node))
				}
			} else {
				var argument ir.Expression
				if probe.fields != nil {
					argument = ir.ObjectLiteral{Record: true, Fields: probe.fields(program)}
				} else {
					settings := -1
					for index, local := range program.Locals {
						if local.Name == "settings" && local.Type == ir.Object {
							settings = index
						}
					}
					if settings < 0 {
						test.Fatal("no local settings to hand the read")
					}
					argument = ir.Read{Local: settings, Of: ir.Object}
				}
				rewritten := 0
				mutate := func(expression ir.Expression) ir.Expression {
					call, ok := expression.(ir.Call)
					if !ok || program.Functions[call.Function].Name != probe.function {
						return expression
					}
					if len(call.Arguments) != 1 {
						test.Fatalf("%s is called with %d arguments, want 1", probe.function, len(call.Arguments))
					}
					rewritten++
					call.Arguments = []ir.Expression{argument}
					return call
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
				if rewritten != 1 {
					test.Fatalf("rewrote %d calls to %s, want 1", rewritten, probe.function)
				}
			}
			want := run{stdout: []byte(probe.stdout)}
			if probe.panic != "" {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.panic + "\n")
			}
			native, _ := natively(test, program)
			for backend, got := range map[string]run{
				"native":     native,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf("%s: %s: want %s; got %s", backend, difference, weakReadShown(want), weakReadShown(got))
				}
			}
		})
	}
}

// The road hunt: each way a program might hand a record to a weak read, written as a program, Lower's
// list (json_parse through returned) and Language's (callback through union). A road that lowers with a
// weak_field_read in it reaches the record path through the language, and fails loudly. Every other one
// must stop, Refused or NotYet; this first version asserts only that it stops, and logs the stop's whole
// text, to be pinned exactly next. The one road that reached, a record bound to a type parameter
// constrained by the weak type, is #xywvn06's, and that head pins its refusal at the call.
func TestWeakReadRoads(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"json_parse",
		"json_decode",
		"cast",
		"cast_unknown",
		"spread",
		"returned",
		"callback",
		"array_element",
		"array_view",
		"map_view",
		"field",
		"union",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			program, loweredError := lowered(test, weakReadPath(test, filepath.Join("roads", name+".ts")))
			if loweredError == nil {
				for _, function := range program.Functions {
					if function.Name == "weak_field_read" {
						test.Fatalf("ROAD REACHES %s", name)
					}
				}
				test.Fatalf("%s lowers, with no weak read and no stop", name)
			}
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case errors.As(loweredError, &refused):
				test.Logf(
					"%s stops, Refused at %s: %s; fix: %s (%v)",
					name,
					refused.Where,
					refused.What,
					refused.Fix,
					loweredError,
				)
			case errors.As(loweredError, &notYet):
				test.Logf("%s stops, NotYet at %s: %s (%v)", name, notYet.Where, notYet.What, loweredError)
			default:
				test.Fatalf("%s stops with neither Refused nor NotYet: %v", name, loweredError)
			}
		})
	}
}
