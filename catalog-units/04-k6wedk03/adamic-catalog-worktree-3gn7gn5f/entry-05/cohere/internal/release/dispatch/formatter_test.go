package dispatch

import (
	"strings"
	"testing"
)

// The formatter's identity keys the format record, so it must move with everything that can change
// what the formatter prints and with nothing else. A missed input keeps a record across a printer
// change, which skips files the new printer would rewrite; an extra one throws the record away on a
// lint commit, which is the cost this identity exists to remove.

func TestFormatterIdentityMovesWithEveryInputAndNotWithOrder(t *testing.T) {
	t.Parallel()

	inputs := []FormatterInput{
		{Name: "github.com/system-inc/cohere/internal/format/native/native.go", Sum: strings.Repeat("a", 64)},
		{Name: "github.com/microsoft/TypeScript/tsc/internal/parser/parser.go", Sum: strings.Repeat("b", 64)},
		{Name: FormatterPassLoop, Sum: strings.Repeat("c", 64)},
	}
	base := hashFormatterInputs("go1.27.0 darwin arm64", inputs)

	reversed := []FormatterInput{inputs[2], inputs[1], inputs[0]}
	if hashFormatterInputs("go1.27.0 darwin arm64", reversed) != base {
		t.Fatal("the same inputs in another order hashed differently, so go list's order would discard the record")
	}
	if hashFormatterInputs("go1.28.0 darwin arm64", inputs) == base {
		t.Fatal("a toolchain change did not move the identity")
	}
	for index := range inputs {
		changed := append([]FormatterInput(nil), inputs...)
		changed[index].Sum = strings.Repeat("d", 64)
		if hashFormatterInputs("go1.27.0 darwin arm64", changed) == base {
			t.Errorf("an edit to %s did not move the identity", inputs[index].Name)
		}
		renamed := append([]FormatterInput(nil), inputs...)
		renamed[index].Name += ".moved"
		if hashFormatterInputs("go1.27.0 darwin arm64", renamed) == base {
			t.Errorf("moving %s did not move the identity", inputs[index].Name)
		}
	}
	if hashFormatterInputs("go1.27.0 darwin arm64", inputs[:2]) == base {
		t.Fatal("dropping an input did not move the identity")
	}
}
