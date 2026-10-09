package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestParserNonNullSourceExtension(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/non_null_refused/parser_index.a")
	_, err := lowered(t, path)
	assertAdamicNonNullRefusal(t, err)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	checkedPath := filepath.Join(t.TempDir(), "parser_index.ts")
	if err := os.WriteFile(checkedPath, source, 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, checkedPath)
	if err != nil {
		t.Fatal(err)
	}
	if program.NonNullChecks.Checked != 1 || program.NonNullChecks.Proven != 0 {
		t.Fatalf("indexed assertion must remain checked: %+v", program.NonNullChecks)
	}
	truth := onNode(t, checkedPath)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node result: %+v", truth)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if literal, ok := node.(ir.ArrayLiteral); ok && len(literal.Elements) == 1 {
			literal.Elements = []ir.Expression{ir.NumberConstant{Value: 2}}
			changes++
			return literal
		}
		return node
	})
	if changes != 1 {
		t.Fatalf("parser indexed-value mutant changed %d arrays", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "stdout differs" {
			t.Fatalf("parser indexed-value mutant escaped Node: %q", difference)
		}
	}
	t.Log("parser indexed-value mutant caught by Node stdout in both backends")
}
