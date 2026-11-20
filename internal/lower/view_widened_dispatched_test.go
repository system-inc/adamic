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
// in adamic_retain, where Node prints 5 and loud (#3jj9dzg, CLI's J-qcmzfvd). The dispatched read is weighed as a
// Property read is, and a getter answers its name, so it stops NotYet in the getter's words: a checked read looks
// only at data slots (#zpzm9ta). Mutant: checkViewWidenedReads skipping dispatchedReads, and this lowers.
func TestDispatchedReadOfAViewedValueStops(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, dispatchedViewProgram("5", true), ".ts")
	var notYet *NotYet
	// textOf's line, below Shout's declaration.
	if !errors.As(err, &notYet) || !strings.Contains(notYet.Where, "main.ts:6:") || !strings.Contains(notYet.What, "a read of field text through type HasText, which may hold a checked view's value, where a getter named text may answer it") {
		t.Fatalf("want NotYet at textOf's h.text (main.ts:6:) in the getter's words, got %v", err)
	}
}

// The twin without Shout: h.text is never dispatched, so it's a widened read checked where it's made (#zpzm9ta),
// and the lie stops loud at it, exit 70 on sanitized native and the JavaScript backend, where Node prints 5.
func TestUndispatchedTwinOfAViewedValueIsChecked(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, dispatchedViewProgram("5", false), ".ts")
	if err != nil {
		t.Fatalf("want it lowered with its read of text checked, got %v", err)
	}
	if checked, _ := widenedReadsOf(t, program, "text"); checked == 0 {
		t.Fatal("no read of text is a widened read checked where it's made")
	}
	brandedStopsProgram(t, program, "adamic: panic: field read failed: h.text is not a string; expected string, found number")
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
