package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A checked view's fields are checked at a read under the view's receiver type, and a value the view admitted can be
// seen as another type after, so const h: HasText = l; h.text read a lying layout's text raw: on e3e7dc1f, Node
// printed 5 for view_widened.ts, release native segfaulted and sanitized native died in AddressSanitizer, while the
// JavaScript backend printed 5 as Node does. The same held on every road Language found (#80cs7rz): an array element,
// a Map value, a closure capture, a parameter, a return, an object field, destructuring, a template, a conditional's
// arm, a class field and satisfies, and a ?. read and a method read through the wider type besides, the same with a
// checked predicate as the entry instead of the cast, through a type written inline, and for the value a checked read
// takes out of a viewed one. A read of a field
// a view checks, through a receiver type that doesn't check it, is now NotYet at the read when the receiver's flow
// may come from the view, failing closed when the reach is unknown (internal/lower/view_widened.go). The fixtures sit in
// reland_refused, since the flow and fresh corpora lower every program in testdata. Mutant: checkViewWidenedReads
// returning nil at once; with it applied, this test goes red, and ADAMIC_VIEW_WIDENED_MUTANT=1 holds the mutant's
// admitted programs to Node on sanitized and release native to show the crash.
func TestViewWidenedReadNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, field, through, node string }{
		{"view_widened.ts", "5:106", "text", "HasText", "5\n"},
		{"view_widened_array.ts", "5:131", "text", "HasText", "5\n"},
		{"view_widened_map.ts", "5:181", "text", "HasText", "5\n"},
		{"view_widened_closure.ts", "5:118", "text", "HasText", "5\n"},
		{"view_widened_parameter.ts", "5:46", "text", "HasText", "5\n"},
		{"view_widened_return.ts", "6:90", "text", "HasText", "5\n"},
		{"view_widened_field.ts", "6:112", "text", "HasText", "5\n"},
		{"view_widened_destructure.ts", "5:102", "text", "HasText", "5\n"},
		{"view_widened_template.ts", "5:111", "text", "HasText", "t=5\n"},
		{"view_widened_conditional.ts", "5:166", "text", "HasText", "5\n"},
		{"view_widened_class_field.ts", "5:84", "text", "HasText", "5\n"},
		{"view_widened_satisfies.ts", "5:138", "text", "HasText", "5\n"},
		{"view_widened_optional.ts", "5:144", "text", "HasText", "5\n"},
		// Node throws calling the lying say, a number: a callable field read through Says is read raw too.
		{"view_widened_method.ts", "5:103", "say", "Says", ""},
		// A type written inline at the widening: no view covers it, so only this rule sees the read.
		{"view_widened_anonymous.ts", "5:164", "text", "{ readonly kind: 9; readonly pos: number; readonly text: string; }", "5\n"},
		// The value inside a viewed value, which a checked read of o.inner takes out.
		{"view_widened_inner.ts", "6:114", "text", "HasText", "5\n"},
		{"view_widened_inner_parameter.ts", "6:46", "text", "HasText", "5\n"},
		// A predicate admits the value with no CheckedCast: it enters where isLit(n) narrows n.
		{"view_widened_predicate.ts", "6:82", "text", "HasText", "5\n"},
		{"view_widened_predicate_array.ts", "6:107", "text", "HasText", "5\n"},
		{"view_widened_predicate_parameter.ts", "6:46", "text", "HasText", "5\n"},
		{"view_widened_predicate_closure.ts", "6:94", "text", "HasText", "5\n"},
		{"view_widened_predicate_field.ts", "7:88", "text", "HasText", "5\n"},
		{"view_widened_predicate_anonymous.ts", "6:140", "text", "{ readonly kind: 9; readonly pos: number; readonly text: string; }", "5\n"},
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
			what := "a read of field " + probe.field + " through type " + probe.through + ", which may hold a checked view's value"
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.file+":"+probe.at) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want the widened read NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}

// The other roads of the lying value already hold, and stay as they are. A read through Lit itself, a closure over
// the Lit and a generic instantiated at Lit stop with the located field-read failure on every backend. `in`,
// Object.keys and hasOwnProperty read no field's value, and a read through Node of a field Node already held (the
// source's own pos) reads what the source guaranteed: each lowers and agrees with Node. The whole-object readers
// (Object.values and Object.entries, String and a template of the object, JSON.stringify), element access and for-in,
// Object.hasOwn and a getter in the contract stop before they lower, at those rules.
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
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if probe.failed == "" {
					if problem := disagreement(truth, got); problem != "" {
						t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
					}
				} else if got.exitCode == 0 || !strings.Contains(string(got.stderr), probe.failed) {
					t.Fatalf("%s: want %q, got exit %d, stdout %q, stderr %q", name, probe.failed, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
	for _, probe := range []struct{ file, at, what string }{
		{"view_widened_held_values.ts", "5:113", "refuses Object.values"},
		{"view_widened_held_entries.ts", "5:115", "refuses Object.entries"},
		{"view_widened_held_string.ts", "5:91", "String conversion of an object"},
		{"view_widened_held_template_object.ts", "5:89", "a template interpolating an object"},
		{"view_widened_held_json.ts", "5:121", "JSON.stringify object references"},
		{"view_widened_held_element.ts", "5:106", "view member read by element access is not yet checked"},
		{"view_widened_held_for_in.ts", "5:125", "view member read by element access is not yet checked"},
		{"view_widened_held_has_own.ts", "5:127", "Object.hasOwn on a shape not proven by a plain literal"},
		{"view_widened_held_getter.ts", "5:24", "a getter in a checked field contract"},
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
// no cast reached (#yq6fkp5's over-refusal, fixed on its base). An honest array element read beside a view isn't:
// the element's reach is unknown, and unknown fails closed.
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
	t.Run("an honest element beside a view", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "main.ts")
		source := `interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }
const xs: HasText[] = [{ text: "arr" }];
for (const x of xs) console.log(x.text);
const real = { kind: 9, pos: 2, text: "real" };
show(real);
`
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := lowered(t, path)
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "main.ts:6:33") || !strings.Contains(err.Error(), "a read of field text through type HasText, which may hold a checked view's value") {
			t.Fatalf("want the unknown reach to fail closed at the element's read, got %v", err)
		}
	})
}

// The price of failing closed, ruled accepted (#80cs7rz): four programs that lowered before are NotYet at a read their
// value may reach a view by, and the per-read check wins them back. objects-good.ts (TestCheckedViewObjects) and
// 03_void_zero.a (TestCheckedViewV2MovedStage3NotYet) are pinned where they were; these two are pinned here, one moved
// from review/agree to review/refused and one out of TestPredicateDirectionUsesLower.
func TestViewWidenedPriceAccepted(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, field, through string }{
		// #80cs7rz: found comes from nodes.find(isB), a library call the flow doesn't follow: an unknown reach.
		{"internal/oracle/testdata/review/refused/9xam9ds_predicate_closed_enum_tag.a", "14:43", "value", "B"},
		// #80cs7rz: declaration is a for-of element, an unknown reach, stored into symbol.valueDeclaration.
		{"internal/oracle/testdata/predicate_direction_uses/inferred_call.ts", "46:66", "name", "Declaration"},
	} {
		t.Run(filepath.Base(probe.file), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, probe.file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || len(truth.stdout) == 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			what := filepath.Base(probe.file) + ":" + probe.at + ": stage 0 can't lower a read of field " + probe.field + " through type " + probe.through + ", which may hold a checked view's value"
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want %q, got %v", what, err)
			}
		})
	}
}
