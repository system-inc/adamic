package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var conditionRepresentations = []string{"number", "string", "object", "array", "function", "optional_reference", "optional_number", "optional_boolean", "optional_string", "null", "union", "sites", "evaluation", "ledger78"}

func init() {
	for _, name := range conditionRepresentations {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/conditions_" + name + ".a", true, false})
	}
}

func TestTypeScriptConditionsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range conditionRepresentations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/conditions_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			// Store source as .a in the repository; exercise the TypeScript loader explicitly.
			path := filepath.Join(t.TempDir(), "conditions.ts")
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if name == "ledger78" {
				if err := conditionLedgerDifference(source, truth.stdout); err != nil {
					t.Fatal(err)
				}
			}
			native, binary := nativelyUncached(t, program)
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if name == "ledger78" {
				t.Log("78 original conditions, 241 samples; Node, both backends and leak checks agree")
			} else {
				t.Logf("Node stdout %q; both backends agree", truth.stdout)
			}
		})
	}
}

func TestConditionNaNMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/conditions_number.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	changed := 0
	mutate := func(node any) any {
		if call, ok := node.(ir.NumberCall); ok && call.Function == "toBoolean" && call.Arguments[0].Type() == ir.Number {
			changed++
			// A naive nonzero comparison incorrectly makes NaN truthy.
			return ir.Binary{Operator: ir.NotEqual, Left: call.Arguments[0], Right: ir.NumberConstant{Value: 0}}
		}
		return node
	}
	for i := range program.Functions {
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, mutate)
	}
	if changed == 0 {
		t.Fatal("NaN mutant changed no conditions")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("mutant was not caught only by stdout: %#v", got)
		}
	}
	t.Logf("NaN-as-truthy mutant caught in both backends by stdout; changed %d conditions", changed)
}

// Check the independent historical row list as well as the executable reduction manifest.
func conditionLedgerDifference(source, stdout []byte) error {
	root := filepath.Join(repository, "docs/verification/conditions")
	original, err := os.ReadFile(filepath.Join(root, "ledger-sites.json"))
	if err != nil {
		return err
	}
	var sites []struct{ Where string }
	if err := json.Unmarshal(original, &sites); err != nil {
		return err
	}
	coverage, err := os.ReadFile(filepath.Join(root, "ledger-coverage.json"))
	if err != nil {
		return err
	}
	var recorded struct {
		Sites []struct {
			Where, Expression, Context string
			Function                   string `json:"reduction_function"`
			Samples                    []string
		}
	}
	if err := json.Unmarshal(coverage, &recorded); err != nil {
		return err
	}
	if len(sites) != 78 || len(recorded.Sites) != 78 {
		return fmt.Errorf("want exactly 78 original and reduced sites")
	}
	expected := map[string]bool{}
	for _, site := range sites {
		if expected[site.Where] {
			return fmt.Errorf("duplicate original site %s", site.Where)
		}
		expected[site.Where] = true
	}
	samples := 0
	for _, site := range recorded.Sites {
		if !expected[site.Where] {
			return fmt.Errorf("unexpected or repeated reduction %s", site.Where)
		}
		delete(expected, site.Where)
		if !strings.Contains(string(source), "function "+site.Function+"(") {
			return fmt.Errorf("missing function %s", site.Function)
		}
		condition := ""
		switch site.Context {
		case "IfStatement":
			condition = "if (" + site.Expression + ")"
		case "WhileStatement":
			condition = "while (" + site.Expression + ")"
		case "ConditionalExpression":
			condition = "return " + site.Expression + " ? true : false;"
		default:
			return fmt.Errorf("uncovered context %s", site.Context)
		}
		if !strings.Contains(string(source), condition) {
			return fmt.Errorf("missing original condition %s", site.Where)
		}
		for index := range site.Samples {
			label := fmt.Sprintf("%s sample %d: ", site.Where, index)
			if strings.Count(string(stdout), label) != 1 {
				return fmt.Errorf("missing or duplicate sample %s", label)
			}
			samples++
		}
	}
	if len(expected) != 0 || samples != 241 || strings.Count(string(stdout), "\n") != samples {
		return fmt.Errorf("incomplete 78-site, 241-sample observation")
	}
	return nil
}

func TestConditionLedgerMissingSiteMutant(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/conditions_ledger78.a"))
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	removed := 0
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(line, "console.log(`src/compiler/builder.ts:642:13 sample ") {
			removed++
			continue
		}
		lines = append(lines, line)
	}
	if removed != 3 {
		t.Fatalf("missing-site mutant removed %d samples", removed)
	}
	path := filepath.Join(t.TempDir(), "missing-site.ts")
	mutantSource := []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(path, mutantSource, 0644); err != nil {
		t.Fatal(err)
	}
	mutant := onNode(t, path)
	if mutant.exitCode != 0 {
		t.Fatalf("mutant failed execution: %#v", mutant)
	}
	if err := conditionLedgerDifference(mutantSource, mutant.stdout); err == nil || !strings.Contains(err.Error(), "missing or duplicate sample src/compiler/builder.ts:642:13") {
		t.Fatalf("missing-site mutant escaped intended coverage check: %v", err)
	}
	t.Log("missing-site source mutant caught by the independent historical-site coverage check")
}

func TestConditionLedgerNullMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/conditions_ledger78.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	changed := 0
	for i := range program.Functions {
		if program.Functions[i].Name != "conditionSite01" {
			continue
		}
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
			if call, ok := node.(ir.NumberCall); ok && call.Function == "toBoolean" && call.Arguments[0].Type() == ir.Array {
				changed++
				return ir.BooleanConstant{Value: true}
			}
			return node
		})
	}
	if changed != 1 {
		t.Fatalf("null-as-truthy mutant changed %d conditions", changed)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("null-as-truthy mutant escaped stdout comparison: %#v", got)
		}
	}
	t.Log("null-as-truthy mutant at commandLineParser.ts:4176:9 caught only by stdout in both backends")
}
