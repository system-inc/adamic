package enterhelper

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: consumer capture pins process environment and bounds sanitizer builds.
func TestThrowableIdentifierMatchesGo(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	capture := t.TempDir()
	script, _ := filepath.Abs("testdata/capture_throwable.py")
	run(t, root, "python3", script, capture)
	entry, _ := filepath.Abs("throwable_main.a")
	binary, js := build(t, entry)
	runner := filepath.Join(root, "oracle/node.mjs")
	sides := func(entry, binary, js, input string) [][]byte {
		return [][]byte{
			run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, input),
			run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, js, input),
			run(t, "", binary, input),
		}
	}
	consumers := []string{"array-callback-return", "consistent-return", "no-unreachable-loop", "react-hooks_rules-of-hooks", "controls"}
	var allInput, allWant []byte
	for _, consumer := range consumers {
		input := filepath.Join(capture, consumer+".input")
		want, err := os.ReadFile(filepath.Join(capture, consumer+".want"))
		if err != nil {
			t.Fatal(err)
		}
		rows, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		allInput = append(allInput, rows...)
		allWant = append(allWant, want...)
		for side, got := range sides(entry, binary, js, input) {
			if !bytes.Equal(got, want) {
				t.Fatalf("%s side %d disagrees with actual Go isThrowableIdentifier", consumer, side)
			}
		}
		t.Logf("%s: %d actual Go calls, %d bytes per backend", consumer, bytes.Count(rows, []byte{'\n'}), len(want))
	}
	input := filepath.Join(t.TempDir(), "all.input")
	if err := os.WriteFile(input, allInput, 0644); err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile("is_throwable_identifier.a")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := os.ReadFile("throwable_main.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct{ name, from, to string }{
		{"declaration name polarity reversed", "return !context.nameIsNode;", "return context.nameIsNode;"},
		{"binding rest exclusion removed", "context.bindingHasRest || context.bindingPropertyIsNode", "context.bindingPropertyIsNode"},
		{"JSX tag identity inverted", "context.tagNameIsNode &&", "!context.tagNameIsNode &&"},
	} {
		directory := t.TempDir()
		mutant := strings.Replace(string(helper), change.from, change.to, 1)
		if mutant == string(helper) {
			t.Fatal("missing mutant anchor")
		}
		os.WriteFile(filepath.Join(directory, "is_throwable_identifier.a"), []byte(mutant), 0644)
		entry := filepath.Join(directory, "throwable_main.a")
		os.WriteFile(entry, driver, 0644)
		binary, js := build(t, entry)
		for side, got := range sides(entry, binary, js, input) {
			if bytes.Equal(got, allWant) {
				t.Fatalf("%s survives side %d", change.name, side)
			}
			t.Logf("%s compiles and runs; comparison alone catches side %d", change.name, side)
		}
	}
}
