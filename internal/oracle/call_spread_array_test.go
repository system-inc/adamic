package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/call_spread_array.a", true, false})
}

func TestCallSpreadArray(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/call_spread_array.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	for name, run := range map[string]run{
		"sanitized":  got,
		"release":    released(test, program),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(truth, run); difference != "" {
			test.Errorf("%s: %s; Node %q, got %q; stderr %q", name, difference, truth.stdout, run.stdout, run.stderr)
		}
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
}

func TestCallSpreadEvaluatedTwiceMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/call_spread_array.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	target := -1
	for index, function := range program.Functions {
		if function.Name == "spreadOnce" {
			target = index
		}
	}
	if target < 0 {
		test.Fatal("spread expression function missing")
	}
	// Re-evaluate the actual spread operand at its argument position. The
	// discarded first result still goes through ordinary ownership planning.
	changed := false
	for _, statement := range program.Main {
		line, ok := statement.(ir.WriteLine)
		if !ok {
			continue
		}
		text, ok := line.Value.(ir.NumberToString)
		if !ok {
			continue
		}
		insertion, ok := text.Value.(ir.Call)
		if !ok || program.Functions[program.CallTargets(insertion)[0]].Name != "array_unshift_arguments" {
			continue
		}
		packed := insertion.Arguments[1].(ir.ArrayLiteral)
		for index, value := range packed.Elements {
			call, ok := value.(ir.Call)
			if ok && len(program.CallTargets(call)) == 1 && program.CallTargets(call)[0] == target {
				packed.Elements[index] = ir.Comma{Left: call, Right: call}
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}
	if !changed {
		test.Fatal("spread mutation site missing")
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "stdout differs" {
		test.Fatalf("JavaScript mutant caught by %q", difference)
	}
	test.Logf("Node %q; evaluated twice %q", truth.stdout, got.stdout)
}
