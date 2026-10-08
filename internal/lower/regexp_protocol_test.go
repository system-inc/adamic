package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRegExpProtocolRuntimeConstruction(t *testing.T) {
	for _, source := range []string{
		`function build(pattern: string): RegExp { return new RegExp(pattern); }`,
		`function build(flags: string): RegExp { return new RegExp("a", flags); }`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("runtime construction must lower: %v", err)
		}
	}
}

func TestRegExpProtocolRefusals(t *testing.T) {
	_, escapeErr := lowerSource(t, `function diagnostic(error: {message: string}): string { return error.message; } try { new RegExp("("); } catch (error) { if (error instanceof Error) console.log(diagnostic(error)); }`)
	var escaped *NotYet
	if !errors.As(escapeErr, &escaped) || !strings.Contains(escapeErr.Error(), "V8 diagnostic wording") {
		t.Fatalf("escaping diagnostic must refuse: %v", escapeErr)
	}
	_, err := lowerSource(t, `try { new RegExp("("); } catch (error) { if (error instanceof Error) console.log(error.message); }`)
	var refusal *NotYet
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "V8 diagnostic wording") {
		t.Fatalf("unsupported diagnostic observation must refuse: %v", err)
	}
}

func TestRegExpProtocolOverrideRefusals(t *testing.T) {
	for _, source := range []string{
		`const regex = /a/; Object.assign(regex, {exec: (text: string): RegExpExecArray | null => null}); regex[Symbol.match]('a');`,
		`try { throw new Error('x'); } catch (error) { if (error instanceof Error) { error.constructor = TypeError; if (error.constructor === TypeError) {} } }`,
		`const regex = /a/g; Object.freeze(regex); regex[Symbol.match]('a');`,
		`const hook = { [Symbol.match]: (_text: string): RegExpMatchArray | null => null }; 'a'.match(hook);`,
		`const hook = { [Symbol.replace]: (_text: string, _replacement: string): string => 'x' }; 'a'.replace(hook, 'b');`,
		`const hook = { [Symbol.search]: (_text: string): number => 4 }; 'a'.search(hook);`,
		`const hook = { [Symbol.split]: (_text: string): string[] => ['x'] }; 'a'.split(hook);`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		var refused *Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Fatalf("an unsupported protocol override must refuse: %s: %v", source, err)
		}
	}
}

func TestRegExpProtocolSplitLimitProof(t *testing.T) {
	_, err := lowerSource(t, `function split(limit: number): (string | undefined)[] { return /(?!\W)/u[Symbol.split]('🌍', limit); }`)
	var refusal *NotYet
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "Smi/HeapNumber") {
		t.Fatalf("unproved Symbol split limit must refuse: %v", err)
	}
}
