package lower

import (
	"errors"
	"strings"
	"testing"
)

// dispatchedViewProgram is CLI's program on #3jj9dzg: a checked view's value, v cast to Lit, read through HasText,
// an interface a class getter also satisfies. With shout, Shout's getter makes h.text a dispatcher's call.
func dispatchedViewProgram(text string, shout bool) string {
	source := "interface Node { readonly kind: number; readonly pos: number; }\n" +
		"interface Lit extends Node { readonly kind: 9; readonly text: string; }\n" +
		"interface HasText { readonly text: string; }\n" +
		"const v = { kind: 9, pos: 1, text: " + text + " };\n"
	if shout {
		source += "class Shout implements HasText { get text(): string { return \"loud\"; } }\n"
	}
	source += "function textOf(h: HasText): string { return h.text; }\n" +
		"function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(textOf(l)); } }\n" +
		"show(v);\n"
	if shout {
		source += "console.log(textOf(new Shout()));\n"
	}
	return source
}

// A checked view's value read through an interface a class getter also satisfies goes through the getter's
// dispatcher, whose fallback reads the slot unchecked: on main this program built and its sanitized native died
// in adamic_retain, where Node prints 5 and loud (#3jj9dzg, CLI's J-qcmzfvd). The read is weighed as a Property
// read is (#80cs7rz), so it stops NotYet at h.text, in the same words as its twin without Shout, whose read was
// never dispatched. Mutant: checkViewWidenedReads skipping dispatchedReads, and the Shout row lowers.
func TestDispatchedReadOfAViewedValueStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name  string
		shout bool
	}{{"through a getter's dispatcher", true}, {"its twin without the getter", false}} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, dispatchedViewProgram("5", row.shout), ".ts")
			var notYet *NotYet
			// textOf's line, which Shout's declaration moves down by one.
			line := "main.ts:5:"
			if row.shout {
				line = "main.ts:6:"
			}
			if !errors.As(err, &notYet) || !strings.Contains(notYet.Where, line) || !strings.Contains(notYet.What, "a read of field text through type HasText, which may hold a checked view's value") {
				t.Fatalf("want NotYet at textOf's h.text (%s) as a read that may hold a checked view's value, got %v", line, err)
			}
			t.Logf("stops at %s: %s", notYet.Where[strings.LastIndex(notYet.Where, "main.ts"):], notYet.What)
		})
	}
}

// The control: with no view in the program, the same getter dispatch lowers and agrees with Node, so the
// weighing touches only a read a view's value may reach (#3jj9dzg).
func TestDispatchedReadWithoutAViewAgrees(t *testing.T) {
	t.Parallel()
	source := "interface HasText { readonly text: string; }\n" +
		"const v: HasText = { text: 'plain' };\n" +
		"class Shout implements HasText { get text(): string { return \"loud\"; } }\n" +
		"function textOf(h: HasText): string { return h.text; }\n" +
		"console.log(textOf(v));\n" +
		"console.log(textOf(new Shout()));\n"
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}
