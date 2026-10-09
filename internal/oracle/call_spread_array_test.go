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

func TestCallSpreadArray(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/call_spread_array.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	for name, run := range map[string]run{"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, run); difference != "" {
			t.Errorf("%s: %s; Node %q, got %q; stderr %q", name, difference, truth.stdout, run.stdout, run.stderr)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

func TestCallSpreadEvaluatedTwiceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/call_spread_array.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	target := -1
	for index, function := range program.Functions {
		if function.Name == "spreadOnce" {
			target = index
		}
	}
	if target < 0 {
		t.Fatal("spread expression function missing")
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
		if !ok || program.Functions[insertion.Function].Name != "array_unshift_arguments" {
			continue
		}
		packed := insertion.Arguments[1].(ir.ArrayLiteral)
		for index, value := range packed.Elements {
			call, ok := value.(ir.Call)
			if ok && call.Function == target {
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
		t.Fatal("spread mutation site missing")
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant caught by %q", difference)
	}
	t.Logf("Node %q; evaluated twice %q", truth.stdout, got.stdout)
}
