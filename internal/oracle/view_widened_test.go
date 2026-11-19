package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// A checked view's fields are checked at a read under the view's receiver type, and a value the view admitted can be
// seen as another type after, so const h: HasText = l; h.text read a lying layout's text raw: on e3e7dc1f, Node
// printed 5 for view_widened.ts, release native segfaulted and sanitized native died in AddressSanitizer, while the
// JavaScript backend printed 5 as Node does. The same held on every road Language found (#80cs7rz): an array element,
// a Map value, a closure capture, a parameter, a return, an object field, destructuring, a template, a conditional's
// arm, a class field and satisfies, and a ?. read and a method read through the wider type besides, the same with a
// checked predicate as the entry instead of the cast, through a type written inline, and for the value a checked read
// takes out of a viewed one. #80cs7rz held each NotYet at the read. #zpzm9ta checks the read where it's made instead,
// by name against the read's own declared type (internal/lower/view_widened.go), so each road lowers and its lie stops
// at that read with the located failure, the same message and exit 70 on sanitized native, release native and
// JavaScript, where Node prints the field. So does a widened read of what a widened read returned, h.parent.label,
// with the lie one level down: it stops at the second read, where the lie is. A method read through the wider type
// stays NotYet in its own words, since a method read keeps its view only where readiness certifies the callable. No
// road here has an honest twin. The fixtures stay in reland_refused, since the flow and fresh corpora hold every
// program in testdata to Node, and these stop where Node runs on. Mutants: Program.CheckedRead not answering for
// WidenedReads, so readiness strips each marked read's view, or checkViewWidenedReads returning nil at once; with
// either applied this test goes red at every lowering road, and ADAMIC_VIEW_WIDENED_MUTANT=1 holds the mutant's
// programs to Node on sanitized and release native instead, to show the crash the check holds back.
func TestViewWidenedReadIsChecked(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, node, read, notYet string }{
		{"view_widened.ts", "5\n", "h.text", ""},
		{"view_widened_array.ts", "5\n", "h.text", ""},
		{"view_widened_map.ts", "5\n", "h.text", ""},
		{"view_widened_closure.ts", "5\n", "h.text", ""},
		{"view_widened_parameter.ts", "5\n", "x.text", ""},
		{"view_widened_return.ts", "5\n", "h.text", ""},
		{"view_widened_field.ts", "5\n", "b.item.text", ""},
		// A destructuring names the field it reads, not the receiver.
		{"view_widened_destructure.ts", "5\n", "text (field text)", ""},
		{"view_widened_template.ts", "t=5\n", "h.text", ""},
		{"view_widened_conditional.ts", "5\n", "h.text", ""},
		{"view_widened_class_field.ts", "5\n", "this.h.text", ""},
		{"view_widened_satisfies.ts", "5\n", "h2.text", ""},
		{"view_widened_optional.ts", "5\n", "h?.text", ""},
		// Node throws calling the lying say, a number. A method read keeps its view only where readiness certifies
		// the callable, so marking it would let it through unchecked, and it stays NotYet.
		{"view_widened_method.ts", "", "", "5:103: stage 0 can't lower a read of method say through type Says, which may hold a checked view's value"},
		// A type written inline at the widening: no view covers it, so only this rule sees the read.
		{"view_widened_anonymous.ts", "5\n", "r.text", ""},
		// The value inside a viewed value, which a checked read of o.inner takes out.
		{"view_widened_inner.ts", "5\n", "h.text", ""},
		{"view_widened_inner_parameter.ts", "5\n", "x.text", ""},
		// A widened read of an aggregate field, then a read of a field of what it returned: the first read's value is
		// weighed as inside a viewed value, so the second is checked too, and the lie one level down stops there.
		{"view_widened_nested.ts", "5\n", "h.parent.label", ""},
		// A predicate admits the value with no CheckedCast: it enters where isLit(n) narrows n.
		{"view_widened_predicate.ts", "5\n", "h.text", ""},
		{"view_widened_predicate_array.ts", "5\n", "h.text", ""},
		{"view_widened_predicate_parameter.ts", "5\n", "x.text", ""},
		{"view_widened_predicate_closure.ts", "5\n", "h.text", ""},
		{"view_widened_predicate_field.ts", "5\n", "b.item.text", ""},
		{"view_widened_predicate_anonymous.ts", "5\n", "r.text", ""},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if probe.node == "" {
				if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "is not a function") {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
			} else if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if probe.notYet != "" {
				if os.Getenv("ADAMIC_VIEW_WIDENED_MUTANT") == "1" {
					t.Skip("the method read is NotYet by its own rule, which the mutant that strips a marked read's view keeps")
				}
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.file+":"+probe.notYet) {
					t.Fatalf("want %q, got %v", probe.notYet, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want the widened read lowered as a checked read, got %v", err)
			}
			failed := "field read failed: " + probe.read + " is not a string; expected string, found number"
			if os.Getenv("ADAMIC_VIEW_WIDENED_MUTANT") == "1" {
				native, _ := nativelyUncached(t, program)
				for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program)} {
					if disagreement(truth, got) == "" || strings.Contains(string(got.stderr), failed) {
						t.Fatalf("the mutant's %s program agrees with Node or still stops at the check: %+v", name, got)
					}
					t.Logf("mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q", name, got.exitCode, got.stdout, got.stderr, truth.stdout)
				}
				return
			}
			stopsAtFieldRead(t, program, "", failed)
		})
	}
}

// stopsAtFieldRead runs a lowered program on sanitized native, release native and the JavaScript backend, and fails
// t unless each exits 70 having printed stdout, with the located field-read failure as its whole stderr.
func stopsAtFieldRead(t *testing.T, program *ir.Program, stdout, failed string) {
	t.Helper()
	want := run{exitCode: 70, stdout: []byte(stdout), stderr: []byte("adamic: panic: " + failed + "\n")}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if problem := disagreement(want, got); problem != "" {
			t.Errorf("%s: %s: want exit 70, stdout %q and %q, got exit %d, stdout %q, stderr %q", name, problem, stdout, failed, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// The other roads of the lying value already hold, and stay as they are. A read through Lit itself, a closure over
// the Lit and a generic instantiated at Lit stop with the located field-read failure on every backend. `in`,
// Object.keys and hasOwnProperty read no field's value, and a read through Node of a field Node already held (the
// source's own pos) reads what the source guaranteed: each lowers and agrees with Node, and so does JSON.stringify,
// which prints the allocation by its own slots' representation bytes (json_objects.c, #d9ep7de). The other
// whole-object readers (Object.values and Object.entries, String and a template of the object), element access and
// for-in, and Object.hasOwn stop before they lower, at those rules. A getter in the contract that only an interface declares
// is on no allocation, so it refuses nothing (#c7w6j0y), and the read through it stops at the widened read's own rule:
// a checked read looks only at data slots, so where a getter may answer the name it stays NotYet in the getter's
// words (#zpzm9ta).
func TestViewWidenedRoadsHeld(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, failed string }{
		{"view_widened_held_direct.ts", "field read failed: l.text is not a string; expected string, found number"},
		{"view_widened_held_closure.ts", "field read failed: l.text is not a string; expected string, found number"},
		{"view_widened_held_generic.ts", "field read failed: x.text is not a string; expected string, found number"},
		{"view_widened_held_in.ts", ""},
		{"view_widened_held_upcast.ts", ""},
		{"view_widened_held_keys.ts", ""},
		{"view_widened_held_has_own_property.ts", ""},
		{"view_widened_held_json.ts", ""},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			if probe.failed != "" {
				stopsAtFieldRead(t, program, "", probe.failed)
				return
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
	for _, probe := range []struct{ file, at, what string }{
		{"view_widened_held_values.ts", "5:113", "refuses Object.values"},
		{"view_widened_held_entries.ts", "5:115", "refuses Object.entries"},
		{"view_widened_held_string.ts", "5:91", "String conversion of an object"},
		{"view_widened_held_template_object.ts", "5:89", "a template interpolating an object"},
		{"view_widened_held_element.ts", "5:106", "view member read by element access is not yet checked"},
		{"view_widened_held_for_in.ts", "5:125", "view member read by element access is not yet checked"},
		{"view_widened_held_has_own.ts", "5:127", "Object.hasOwn on a shape not proven by a plain literal"},
		{"view_widened_held_getter.ts", "6:109", "a read of field text through type TextGetter, which may hold a checked view's value, where a getter named text may answer it"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.file+":"+probe.at) || !strings.Contains(err.Error(), probe.what) {
				t.Fatalf("want it refused at %s, got %v", probe.at, err)
			}
		})
	}
}

// A read through a wider type of a value no view touched is honest. Beside a view of another value, a widened local,
// a parameter and the source's own tag test and fields lower and agree with Node, on sanitized and release native as
// well as JavaScript, which can't see this class of hole. With no view in the program, an array element and a Map
// value read through the wider type do too, and so do the honest spreads into a copy that is later cast, whose source
// no cast reached (#yq6fkp5's over-refusal, fixed on its base). So does an honest array element read beside a view,
// which #80cs7rz failed closed since the element's reach is unknown: #zpzm9ta checks that read where it's made, and
// an honest element passes the check.
func TestHonestWidenedReadAgrees(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, node string }{
		{"beside a view", `interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } console.log(String(n.pos)); }
const honest: HasText = { text: "plain" };
console.log(honest.text);
function textOf(x: HasText): string { return x.text; }
console.log(textOf({ text: "param" }));
const real = { kind: 9, pos: 2, text: "real" };
show(real);
`, "plain\nparam\nreal\n2\n"},
		{"without a view", `interface HasText { readonly text: string; }
interface Lit { readonly kind: 9; readonly text: string; }
const l: Lit = { kind: 9, text: "lit" };
const h: HasText = l;
console.log(h.text);
const xs: HasText[] = [l];
for (const x of xs) console.log(x.text);
const m = new Map<string, HasText>([["k", l]]);
const g = m.get("k");
console.log(g === undefined ? "none" : g.text);
`, "lit\nlit\nlit\n"},
		{"an honest encodeJson beside a view", `import { encodeJson } from 'adamic';
interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }
const honest: HasText = { text: "plain" };
console.log(encodeJson<HasText>(honest));
const real = { kind: 9, pos: 2, text: "real" };
show(real);
`, "{\"text\":\"plain\"}\nreal\n"},
		{"an honest element beside a view", `interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }
const xs: HasText[] = [{ text: "arr" }];
for (const x of xs) console.log(x.text);
const real = { kind: 9, pos: 2, text: "real" };
show(real);
`, "arr\nreal\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("an honest widened read must lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
	for _, file := range []string{
		"stage3/interface-downcasts/v2/snapshot-maybe-number.ts",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p01_maybe_boolean_spread.ts",
		"internal/oracle/testdata/loom_views_c/snapshot-original.ts",
	} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("an honest spread source must lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// A widened read is marked checked under WidenedReadKey: its receiver type, its name and its place in the source. Two
// reads share a key only if they share all three, so on one line both's a.text, whose a may hold the viewed Lit, is
// checked and b.text, whose b only ever holds a plain literal, stays plain, each judged at its own column. The two
// instances of textOf are two lowerings of one source read, h.text through HasText at one place, so they share a key:
// the instance whose h may hold the viewed Lit demands the check, and the instance given a fresh literal is checked
// with it. Whichever reads are checked, an honest program agrees with Node on sanitized native, release native and
// JavaScript (#zpzm9ta). Mutant: WidenedReadKey without the place, and b.text is checked with a.text.
func TestViewWidenedReadKeyCollision(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function both(a: HasText, b: HasText): string { return a.text + ":" + b.text; }
function textOf<T>(h: HasText, tag: T): string { return h.text + String(tag); }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const plain: HasText = { text: "plain" }; console.log(both(l, plain)); console.log(textOf(l, 1)); } }
const real = { kind: 9, pos: 2, text: "real" };
show(real);
console.log(textOf({ text: "fresh" }, "s"));
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("real:plain\nreal1\nfreshs\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("an honest widened read must lower: %v", err)
	}
	// Each read of text by line and whether the program checks it: the keys are what behavior can't show, since an
	// honest value passes a checked read and a plain one alike.
	var line4, line5 []string
	for _, read := range sourceSlotReads(program, "text") {
		switch {
		case strings.HasPrefix(read, "4:"):
			line4 = append(line4, read)
		case strings.HasPrefix(read, "5:"):
			line5 = append(line5, read)
		}
	}
	if strings.Join(line4, ",") != "4:false,4:true" {
		t.Errorf("both's reads of text: want a.text checked and b.text plain, got %v", line4)
	}
	if len(line5) < 2 || slices.Contains(line5, "5:false") {
		t.Errorf("textOf's reads of text: want both instances' reads checked under the key they share, got %v", line5)
	}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if problem := disagreement(truth, got); problem != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// encodeJson reads every field of its value and of what that holds by its schema, raw, so it read a
// lying layout as the types the schema says, through the view's own type as well as a wider one: on
// a2cd978f, release native segfaulted and sanitized native died in AddressSanitizer for both programs
// here, where Node prints the field. It's NotYet now when anything in its value may come from a
// view's entry (#shh1v0z, view_reads.go's whole-value reader). With checkViewWidenedReads returning nil
// at once, ADAMIC_VIEW_WIDENED_MUTANT=1 holds the mutant's admitted programs to Node natively, as
// TestViewWidenedReadIsChecked does. The other mutant there, readiness stripping a marked read's view,
// leaves these NotYet, since no whole-value reader is a marked read.
func TestViewWholeValueReadNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, through, node string }{
		{"view_widened_encode.ts", "6:126", "HasText", "{\"text\":5}\n"},
		{"view_widened_encode_lit.ts", "6:100", "Lit", "{\"kind\":9,\"text\":5,\"pos\":1}\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if os.Getenv("ADAMIC_VIEW_WIDENED_MUTANT") == "1" {
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				native, _ := nativelyUncached(t, program)
				for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program)} {
					if disagreement(truth, got) == "" {
						t.Fatalf("the mutant's %s program agrees with Node: %+v", name, got)
					}
					t.Logf("mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q", name, got.exitCode, got.stdout, got.stderr, truth.stdout)
				}
				return
			}
			what := probe.file + ":" + probe.at + ": stage 0 can't lower a read of every field of a value of type " + probe.through + " by encodeJson, which may hold a checked view's value"
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want %q, got %v", what, err)
			}
		})
	}
}
