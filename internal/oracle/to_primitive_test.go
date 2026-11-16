package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// String(v) and a template's ${v} over an unknown, ToPrimitive with the string hint, lowered as
// ordinary calls in one helper (lower/to_primitive.go, #09dvwny). The fixtures sit in to_primitive,
// since the flow and fresh corpora lower every program at the top of testdata. Each answer is held to
// Node natively, sanitized and released, and on the JavaScript backend, and leaks nothing; each stop is
// held to the same bytes on all three beside what Node prints running the source, so the day one is
// built it goes red here and joins the answers.
func toPrimitiveFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/to_primitive", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// cannotConvert is the TypeError when neither toString nor valueOf gives a primitive, as Node and
// exceptions.c word it.
const cannotConvert = "TypeError: Cannot convert object to primitive value"

// toPrimitiveAnswers is what each answering fixture prints on Node. arrays_acyclic.a and probes_acyclic.a
// are arrays.a's and probes.a's answers without their cycles (toPrimitiveLooseStops).
var toPrimitiveAnswers = []struct{ file, stdout string }{
	{"primitives.a", "undefined\n<undefined>\nnull\n<null>\n42\n<42>\n-0.5\n<-0.5>\ntrue\n<true>\ntext\n<text>\n"},
	{"classes.a", "[object Object]\n<[object Object]>\n" +
		"named\n<named>\n" +
		"named\n<named>\n" +
		"[object Object]\n<[object Object]>\n" +
		"3\n<3>\n" +
		"undefined\n<undefined>\n" +
		"toString ran\nstring\ntoString ran\n<string>\n" +
		cannotConvert + "\n<" + cannotConvert + ">\n" +
		cannotConvert + "\n<" + cannotConvert + ">\n"},
	{"literals.a", "x\n<x>\n" +
		"method\n<method>\n" +
		"42\n<42>\n" +
		"[object Object]\n<[object Object]>\n" +
		"[object Object]\n<[object Object]>\n" +
		cannotConvert + "\n<" + cannotConvert + ">\n"},
	{"errors.a", "Error: plain\n<Error: plain>\n" +
		"TypeError: typed\n<TypeError: typed>\n" +
		"Error\n<Error>\n" +
		"Error\n<Error>\n" +
		"Custom: renamed\n<Custom: renamed>\n" +
		"renamed\n<renamed>\n"},
	// Agrees with Node, running the source (J-xhex7eq).
	{"probes_acyclic.a", "c1|c2 2\n" +
		"a|a b|b\n" +
		"a,,,b|a,,,b\n" +
		"n!|n!\n"},
	{"records.a", "[object Object]\n<[object Object]>\n" + cannotConvert + "\n<" + cannotConvert + ">\n"},
	// Node's output, measured running the source (J-zn9yhj7).
	{"arrays_acyclic.a", "a,1,,,true,named\n<a,1,,,true,named>\n" +
		"Error: e\n<Error: e>\n" +
		"p,q\n<p,q>\n" +
		"1,a\n<1,a>\n" +
		",\n<,>\n" +
		"\n<>\n" +
		"\n<>\n" +
		"\n<>\n" +
		"h,[]\n<h,[]>\n" +
		"Error: mid-join\n<Error: mid-join>\n" +
		"t,recovered\n<t,recovered>\n"},
}

// Mutants, each its own run with the change reverted by git apply -R: the object arm skipped (the
// instance arm stops as the static class arm does), so classes.a goes red; toString and valueOf
// swapped, so classes.a's Ordered prints valueOf's line first; a literal's own toString that holds no
// function taken as one (OwnFunction answering HasOwn), so literals.a stops at { toString: 3 }; and
// the Error arm put after the plain object arm, so errors.a's Errors print "[object Object]"; the join
// stack dropped (no JoinEnter and no JoinLeave), so arrays_acyclic.a's holder, whose Back joins it
// again through viewed, joins itself until the stack runs out, each run bounded by its minute; and
// JoinLeave moved out of the finally to after the loop, so a throw leaves the array on the stack and
// arrays_acyclic.a's join after it is ""; and the scalar test dropped, which
// TestToPrimitiveStopsAlikeOnBothBackends holds: the stop fixtures join.
func TestToPrimitiveAnswersAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, probe := range toPrimitiveAnswers {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := toPrimitiveFixture(t, probe.file)
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.stdout)}, truth); difference != "" {
				t.Fatalf("source Node: %s: exit %d, stdout %q, stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("String of an unknown must lower: %v", err)
			}
			native, binary := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Errorf("leaked: %s", report)
			}
		})
	}
}

// toPrimitiveLooseStops are the two fixtures whose arrays hold themselves (arrays.a's itself, after and
// outer with inner, probes.a's a and b), held to lower's loose-slot stop at their first such push: Item
// holds object, and an Item[] written into an Item[] can reach its holder back. Each program pops its
// back-edges before the end, so it balanced natively, but the cycle is real (the runtime breaks none) and
// reaching can't credit a later pop (Adamic on #kaz4pk8). The stop lifts with the link, #yeep0sx, and
// each goes red here the day it lands, to be turned back into an answer beside its acyclic twin.
var toPrimitiveLooseStops = []struct{ file, at, what string }{
	{"arrays.a", ":48:1", "the elements of Item[], written a value of type Item[] that can reach back to the Item[] holding it through what a value held as unknown or object can be: " + looseSlotWords},
	{"probes.a", ":21:1", "the elements of Item[], written a value of type Item[] that can reach back to the Item[] holding it through what a value held as unknown or object can be: " + looseSlotWords},
}

// TestToPrimitiveCyclesStopUntilTheLink holds each of toPrimitiveLooseStops to its stop: a NotYet at its
// position, in the stop's own words.
func TestToPrimitiveCyclesStopUntilTheLink(t *testing.T) {
	t.Parallel()
	for _, pin := range toPrimitiveLooseStops {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, toPrimitiveFixture(t, pin.file))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want the loose-slot stop, a NotYet, until #yeep0sx", err)
			}
			if !strings.HasSuffix(notYet.Where, pin.file+pin.at) || notYet.What != pin.what {
				t.Fatalf("got %s: %s, want %s%s: %s", notYet.Where, notYet.What, pin.file, pin.at, pin.what)
			}
			t.Logf("stop: %v", err)
		})
	}
}

// toPrimitiveStops are the stop fixtures: which value's kind each stops at, after printing before,
// and what Node prints running the source.
var toPrimitiveStops = []struct{ file, what, node string }{
	// One stop for an array of numbers, one of booleans and one of a union type holding only numbers,
	// since natively only the first two are unboxed and the JavaScript backend can't tell any of them
	// apart: the third is the rule's over-stop, a cost (lower/to_primitive.go's arrayToString).
	{"stop_scalar_array.a", "an array of only numbers, booleans, null or undefined", "before\n1,2,3\n"},
	{"stop_boolean_array.a", "an array of only numbers, booleans, null or undefined", "before\ntrue,false\n"},
	{"stop_union_numbers.a", "an array of only numbers, booleans, null or undefined", "before\n1,2\n"},
	// Only undefined, which an unboxed number | undefined array holding nothing else is too.
	{"stop_undefined_array.a", "an array of only numbers, booleans, null or undefined", "before\n,\n"},
	// Node runs the source as its runner emits it, which indents by four spaces.
	{"stop_static_class.a", "a class object", "before\nclass Counter {\n    static count = 1;\n}\n"},
	{"stop_record_function.a", "an object whose own toString is a function this program can't name", "before\nfrom record\n"},
}

// Each stop, on both backends, in the same words byte for byte, beside Node's answer.
func TestToPrimitiveStopsAlikeOnBothBackends(t *testing.T) {
	t.Parallel()
	for _, probe := range toPrimitiveStops {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := toPrimitiveFixture(t, probe.file)
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("String of an unknown must lower and stop at run time: %v", err)
			}
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: converting " + probe.what + " to a string is not yet lowered\n"), exitCode: 70}
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": backend} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}
