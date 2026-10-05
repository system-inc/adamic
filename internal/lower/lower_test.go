package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func lowerSource(t *testing.T, source string) (*ir.Program, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Lower(context.Background(), program)
}

func TestConsoleLowersToWriteLine(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "console.log('one');\nconsole.error(`two`);\nconsole.log('one');\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []ir.Statement{
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}},
		ir.WriteLine{Stream: ir.Stderr, Value: ir.StringConstant{Index: 1}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}},
	}
	if len(program.Main) != len(want) || strings.Join(program.Strings, ",") != "one,two" {
		t.Fatalf("got %v with strings %q", program.Main, program.Strings)
	}
	for index := range want {
		if program.Main[index] != want[index] {
			t.Errorf("instruction %d: got %v, want %v", index, program.Main[index], want[index])
		}
	}
	if program.Source != "main.a" {
		t.Errorf("Source: got %q, want main.a", program.Source)
	}
}

func TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		// A local console is not the prelude's, and lowering it as console.log would print what the
		// program never asked to print.
		{"a local console", "const console = { log: (message: string): void => {} };\nconsole.log('x');\n", "main.a:1:7: stage 0 can't lower a value of type { log: (message: string) => void; } yet"},
		{"a closure", "function outer(): number {\n\tfunction inner(): number {\n\t\treturn 1;\n\t}\n\treturn inner();\n}\nconsole.log(`${outer()}`);\n", "main.a:2:2: stage 0 can't lower a function inside a function (a closure) yet"},
		{"an import", "import { panic } from 'adamic';\npanic('x');\n", "main.a:1:1: stage 0 can't lower an ImportDeclaration yet"},
		// String < compares UTF-16 code units, which the runtime doesn't do yet (docs/0.1.md, program 10).
		{"a string comparison", "console.log(`${'a' < 'b'}`);\n", "main.a:1:16: stage 0 can't lower a BinaryExpression with a string and a string yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Error(), probe.want) {
				t.Errorf("got %v, want an error ending %q", err, probe.want)
			}
		})
	}
}

// What 0.1 refuses for good is said as a refusal with its fix, never as "not yet".
func TestWhatZeroOneRefusesIsRefusedWithAFix(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"a string as a condition", "const name = 'x';\nif (name) {\n\tconsole.log(name);\n}\n", "main.a:2:5: Adamic 0.1 refuses a string as a condition; compare it explicitly, like name.length > 0 or count !== 0"},
		{"a number as a loop condition", "let count = 3;\nwhile (count) {\n\tcount -= 1;\n}\n", "main.a:2:8: Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0"},
		{"var", "var old = 1;\n", "main.a:1:1: Adamic 0.1 refuses var; use const or let"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(refused.Error(), probe.want) {
				t.Errorf("got %v, want a refusal ending %q", err, probe.want)
			}
		})
	}
}
