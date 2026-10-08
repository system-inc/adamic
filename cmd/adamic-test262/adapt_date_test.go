package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestDateNumericLocals(t *testing.T) {
	t.Parallel()
	source := "var date = new Date(0); var result, expected; result = date.setUTCDate(2); expected = date.getTime(); assert.sameValue(result, expected);"
	adapted := adaptSource(source)
	if adapted.Counts["date-local"] != 2 || !strings.Contains(adapted.Source, "result: number | undefined") {
		t.Fatalf("got %#v", adapted)
	}
	source = "var date, text; date = new Date(0); text = date.toISOString(); assert.sameValue(text, date.toISOString());"
	adapted = adaptSource(source)
	if adapted.Counts["date-local"] != 2 || !strings.Contains(adapted.Source, "date: Date | undefined") || !strings.Contains(adapted.Source, "text: string | undefined") {
		t.Fatalf("Date/string locals: %#v", adapted)
	}
	for _, source := range []string{
		"var date = new Date(0); var result; result = date.getTime(); result = 'text'; console.log(result);",
		"var date = new Date(0); var result; result = date.getTime(); result = unknownCall();",
		"var date = new Date(0); var result; result = date.getTime(); result += 'text';",
	} {
		if adapted := adaptSource(source); adapted.Counts["date-local"] != 0 {
			t.Fatalf("uncertain writes annotated: %#v", adapted)
		}
	}
}

func TestSentinelHarnessAgreesWithOriginal(t *testing.T) {
	t.Parallel()
	// This checks ordering and failure text too. SameValue's NaN and signed-zero
	// behavior stays in the ordinary harness when neither operand is a sentinel.
	source := `let observations = "";
function value(): number { observations += "v"; return -0; }
function message(): string { observations += "m"; return "message"; }
assert.notSameValue(value(), undefined, message());
assert.notSameValue(null, new Date(0));
assert.notSameValue(new Date(0), undefined, "literal message");
assert.sameValue(null, null);
assert.sameValue(Date.prototype, Date.prototype);
assert.sameValue(NaN, NaN);
assert.notSameValue(-0, 0);
try { assert.sameValue(1, undefined); } catch (error) { if (error instanceof Error) observations += error.message; }
console.log(observations);`
	// An unrewritten free-function program is the old primitive harness, with the
	// null-only assertions removed because that old harness did not support null.
	original := strings.NewReplacer("assert.notSameValue", "assertNotSameValue", "assert.sameValue", "assertSameValue").Replace(source)
	original = strings.ReplaceAll(original, "assertNotSameValue(null, new Date(0));", "if (null === new Date(0)) throw new Error('Expected NotSameValue');")
	original = strings.ReplaceAll(original, "assertSameValue(null, null);", "if (null !== null) throw new Error('Expected SameValue');")
	original = strings.ReplaceAll(original, "assertSameValue(Date.prototype, Date.prototype);", "if (Date.prototype !== Date.prototype) throw new Error('Expected SameValue');")
	var outputs []string
	for index, body := range []string{original, rewriteHarnessCalls(source)} {
		path := filepath.Join(t.TempDir(), "program.mts")
		if err := os.WriteFile(path, []byte(prelude+"\n"+body), 0644); err != nil {
			t.Fatal(err)
		}
		if index == 1 {
			checkedPath := strings.TrimSuffix(path, ".mts") + ".ts"
			if err := os.WriteFile(checkedPath, []byte(prelude+"\n"+body), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := load.LoadOverlay([]string{checkedPath}, map[string]string{checkedPath: prelude + "\n" + body}); err != nil {
				t.Fatalf("rewritten harness must type-check: %v", err)
			}
		}
		output, err := exec.Command("node", "--disable-warning=ExperimentalWarning", path).CombinedOutput()
		if err != nil {
			t.Fatalf("program %d: %v: %s", index, err, output)
		}
		outputs = append(outputs, string(output))
	}
	if outputs[0] != outputs[1] || outputs[1] != "vmExpected SameValue\n" {
		t.Fatalf("original %q adapted %q", outputs[0], outputs[1])
	}
	for _, source := range []string{
		"function check(undefined) { assert.sameValue(-0, undefined); } check(0);",
		"assert.sameValue(-0, undefined); var undefined = 0;",
		"function check(undefined) { return `${assert.sameValue(-0, undefined)}`; }",
		"function check(Date) { assert.sameValue(Date.prototype, NaN); }",
		"function check(Date) { return `${assert.sameValue(Date.prototype, NaN)}`; }",
	} {
		if strings.Contains(rewriteHarnessCalls(source), "assertIdentity(") {
			t.Fatalf("shadowed intrinsic rewritten: %s", source)
		}
	}
}
