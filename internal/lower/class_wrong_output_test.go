package lower

import (
	"os"
	"strings"
	"testing"
)

func TestClassWrongOutputIteratorReceiver(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/class_wrong_output_refused/iterators_override_this.a")
	if err != nil {
		t.Fatal(err)
	}
	prefix := strings.Split(string(source), "for(const value")[0]
	for _, use := range []string{
		"for (const value of new ScaledIterator()) { console.log(`${value}`); break; }",
		"console.log([...new ScaledIterator()].join(','));",
		"const viewed: BaseIterator = new ScaledIterator(); console.log(Array.from(viewed).join(','));",
		"const [first = -1] = new ScaledIterator(); console.log(`${first}`);",
	} {
		_, err := lowerSource(t, prefix+use)
		if err != nil {
			t.Fatalf("dynamic receiver protocol refused: %s: %v", use, err)
		}
	}
	for _, use := range []string{
		"console.log([...new BaseIterator()].join(','));",
		"class Unchanged extends BaseIterator {} console.log([...new Unchanged()].join(','));",
	} {
		if _, err := lowerSource(t, prefix+use); err != nil {
			t.Fatalf("unchanged protocol refused: %v", err)
		}
	}
}

func TestClassWrongOutputKeysRepair(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a")
	if err != nil {
		t.Fatal(err)
	}
	repaired := strings.Replace(string(source), "Object.keys(view)", "Object.keys({ value: view.value })", 1)
	program, err := lowerSource(t, repaired)
	if err != nil {
		t.Fatalf("explicit string-key copy refused: %v", err)
	}
	requireLoweredOutput(t, program)
}

func TestClassWrongOutputPrivateRepair(t *testing.T) {
	t.Parallel()
	source := "class Box { #secret: string; constructor(secret: string) { this.#secret = secret; } reveal(): string { return this.#secret; } static peek(box: Box): string { return box.reveal(); } } console.log(Box.peek(new Box(`s${1}`)));"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatalf("instance delegation refused: %v", err)
	}
	requireLoweredOutput(t, program)
}
