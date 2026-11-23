package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// The value of x++ or x-- on a numeric local is the old value, and the write is done before anything after it,
// even when a loop test then fails (Adamic on #bbssdcy, family 102, #f3n0gyp; tsc's utilities.ts:7992,
// getPreviousNonWhitespacePosition's `while (pos-- > 0)`). Each program agrees with Node on native, release
// and the JavaScript backend, with no leak; a field target keeps its stop.
// Mutant (postfix_value.mutants.md): returning the new value instead of the held old one turns every row red.
func TestPostfixUpdateValueOnANumericLocal(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"loop_test", "let pos = 2;\nlet tally = 0;\nwhile (pos-- > 0) {\n\ttally += pos;\n}\nconsole.log(`${tally} ${pos}`);\n", "1 -1\n"},
		{"argument_order", "function pair(a: number, b: number): string { return `${a} ${b}`; }\nlet x = 5;\nconsole.log(pair(x++, x));\nconsole.log(`${x}`);\n", "5 6\n6\n"},
		{"prefix_beside", "let x = 1;\nconst a = x++;\nconst b = ++x;\nconst c = x--;\nconsole.log(`${a} ${b} ${c} ${x}`);\n", "1 3 3 2\n"},
		{"in_a_function", "function count(n: number): string {\n\tlet i = n;\n\tlet out = '';\n\twhile (i-- > 0) {\n\t\tout += `${i}`;\n\t}\n\treturn `${out} ${i}`;\n}\nconsole.log(count(3));\n", "210 -1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			compiled, sanitized := natively(t, program)
			for label, run := range map[string]run{"native": compiled, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
				if difference := disagreement(truth, run); difference != "" {
					t.Errorf("%s: %s; Node %q, got %q", label, difference, truth.stdout, run.stdout)
				}
			}
			if leaked := leaks(t, program, sanitized); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
}

// Only a numeric local is admitted: a postfix update of a field as a value keeps its stop.
func TestPostfixUpdateValueOfAFieldStops(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := "const box = { n: 1 };\nconst old = box.n++;\nconsole.log(`${old} ${box.n}`);\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a postfix update value whose target is not a plain name") {
		t.Fatalf("want the field's stop, got %v", err)
	}
}
