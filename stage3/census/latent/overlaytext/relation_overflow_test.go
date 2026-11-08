package overlaytext

import (
	"os"
	"strings"
	"testing"
)

func TestLatentRelationOverflowHooksAndProductionWording(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	main, err := readMainLower("lower.go")
	if err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile("../relation_overflow.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	const wording = `"a type relation the checker gave up on as too complex, whose answer lowering read (" + overflow + ")"`
	if strings.Count(main, wording) != 1 {
		t.Fatal("production overflow NotYet wording changed; reconcile the census finding")
	}
	if !strings.Contains(string(helper), strings.ReplaceAll(wording, "overflow", "answer")) {
		t.Fatal("census does not use production overflow NotYet wording")
	}
	for file, needles := range map[string][]string{
		"../relation_overflow.go.txt": {"program.RelationOverflow(typeChecker)", "finding.PieceAtom = latentPieceOrigin", "report.Findings = append"},
		"../lower.go.txt":             {"relations := latentRelationReport(ctx, program)", "relations.End = latentReadRelationOverflow(ctx, program)", "relations.Complete(records)", "latentCheckUnitRelation(program, checker, &unit)"},
		"../units.go.txt":             {"latentCheckUnitRelation(program, checker, &unit)"},
		"../speculative.go.txt":       {"latentCheckUnitRelation(program, checker, &unit)"},
		"../main.go.txt":              {`"relation_overflow": relations`, `"findings": relations.Findings`, "program.LatentCensusEntryReach()"},
		"../make_overlay.py":          {"'internal/lower/latent_relation_overflow.go'", "'stage3/census/latent/tool/relation_overflow_test.go'"},
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, needle := range needles {
			if strings.Count(string(source), needle) != 1 {
				t.Errorf("%s: missing or repeated overflow hook %q", file, needle)
			}
		}
	}
}
