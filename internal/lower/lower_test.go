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
	want := []ir.Instruction{
		ir.WriteLine{Stream: ir.Stdout, String: 0},
		ir.WriteLine{Stream: ir.Stderr, String: 1},
		ir.WriteLine{Stream: ir.Stdout, String: 0},
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
		{"a local console", "const console = { log: (message: string): void => {} };\nconsole.log('x');\n", "main.a:1:1: stage 0 can't lower a VariableStatement yet"},
		{"a computed argument", "const name = 'x';\nconsole.log(name);\n", "main.a:1:1: stage 0 can't lower a VariableStatement yet"},
		{"an import", "import { panic } from 'adamic';\npanic('x');\n", "main.a:1:1: stage 0 can't lower an ImportDeclaration yet"},
		{"a template with a value", "console.log(`${1 + 1}`);\n", "main.a:1:13: stage 0 can't lower a console argument that isn't a string constant (a TemplateExpression) yet"},
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
