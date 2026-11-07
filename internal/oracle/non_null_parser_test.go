package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestParserNonNullSourceExtension(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/non_null_refused/parser_index.a")
	_, err := lowered(test, path)
	assertAdamicNonNullRefusal(test, err)
	source, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	checkedPath := filepath.Join(test.TempDir(), "parser_index.ts")
	if err := os.WriteFile(checkedPath, source, 0644); err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, checkedPath)
	if err != nil {
		test.Fatal(err)
	}
	if program.NonNullChecks.Checked != 1 || program.NonNullChecks.Proven != 0 {
		test.Fatalf("indexed assertion must remain checked: %+v", program.NonNullChecks)
	}
	truth := onNode(test, checkedPath)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" || len(truth.stderr) != 0 {
		test.Fatalf("unexpected Node result: %+v", truth)
	}
	native, _ := nativelyUncached(test, program)
	for _, got := range []run{native, onJavaScriptBackend(test, program)} {
		if difference := disagreement(truth, got); difference != "" {
			test.Fatal(difference)
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
		test.Fatalf("parser indexed-value mutant changed %d arrays", changes)
	}
	mutant, _ := nativelyUncached(test, program)
	for _, got := range []run{mutant, onJavaScriptBackend(test, program)} {
		if difference := disagreement(truth, got); difference != "stdout differs" {
			test.Fatalf("parser indexed-value mutant escaped Node: %q", difference)
		}
	}
	test.Log("parser indexed-value mutant caught by Node stdout in both backends")
}
