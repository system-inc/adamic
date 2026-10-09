package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/counters_shapes.a", true, false})
}

func counterShapes(t *testing.T) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/counters_shapes.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func TestCounterShapesProofAndFallbacks(t *testing.T) {
	t.Parallel()
	program, _ := counterShapes(t)
	expected := map[string]string{
		"squareCounter": "guarded", "productCounter": "guarded", "parameterCounter": "guarded",
		"leftCounter": "integer", "rightCounter": "integer", "sumCounter": "integer", "multipliedCounter": "integer",
		"guardedCounter": "guarded", "edgeCounter": "guarded", "descendingCounter": "guarded",
		"changingBoundCounter": "double", "changingStepCounter": "double",
		"dependencyOuterCounter": "guarded", "dependentCounter": "guarded",
		"largeCounter": "integer", "smallCounter": "integer", "largeStepCounter": "guarded",
		"overflowingProductCounter": "double", "squareEdgeCounter": "guarded",
		"independentOuterCounter": "guarded", "independentInnerCounter": "guarded",
		"typedSquareCounter": "integer", "typedProductCounter": "integer",
		"localCounter": "guarded", "fractionBoundCounter": "double", "fractionStepCounter": "double", "capturedStepCounter": "double",
	}
	found := map[string]bool{}
	for _, local := range program.Locals {
		want, ok := expected[local.Name]
		if !ok {
			continue
		}
		found[local.Name] = true
		got := "double"
		if local.Counter {
			got = "integer"
		}
		if local.CounterGuard != nil {
			got = "guarded"
		}
		if got != want {
			t.Errorf("%s: %s, want %s", local.Name, got, want)
		}
	}
	for name := range expected {
		if !found[name] {
			t.Errorf("missing %s", name)
		}
	}
	// Trace actual integer declarations: fractional/nonfinite bounds, bad steps,
	// and updates crossing 2^53 must execute the original double version.
	code := native.C(program)
	for id, local := range program.Locals {
		if local.Name == "independentInnerCounter" {
			declaration := fmt.Sprintf("int64_t adamic_local_%d_independentInnerCounter =", id)
			if strings.Count(code, declaration) != 1 {
				t.Fatal("independent nested guards duplicated the integer path")
			}
		}
	}
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n#include <stdio.h>", 1)
	for _, name := range []string{"guardedCounter", "edgeCounter", "descendingCounter", "squareCounter", "parameterCounter", "dependencyOuterCounter", "dependentCounter", "largeStepCounter", "squareEdgeCounter", "independentOuterCounter", "independentInnerCounter", "localCounter"} {
		declaration := regexp.MustCompile(`(?m)^(\s*int64_t adamic_local_\d+_` + name + ` = [^\n]+;)$`)
		if !declaration.MatchString(code) {
			t.Fatalf("missing integer path: %s", name)
		}
		code = declaration.ReplaceAllString(code, `${1}`+"\n\tfprintf(stderr, \"integer:"+name+"\\n\");")
	}
	binary := filepath.Join(t.TempDir(), "trace")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	expectedTrace := "integer:squareCounter\ninteger:parameterCounter\ninteger:guardedCounter\ninteger:edgeCounter\ninteger:descendingCounter\ninteger:dependencyOuterCounter\ninteger:dependentCounter\ninteger:dependentCounter\ninteger:dependentCounter\ninteger:squareEdgeCounter\ninteger:squareEdgeCounter\ninteger:squareEdgeCounter\ninteger:independentOuterCounter\ninteger:independentInnerCounter\ninteger:independentInnerCounter\ninteger:localCounter\n"
	if result.exitCode != 0 || string(result.stderr) != expectedTrace {
		t.Fatalf("wrong fallback trace: exit %d stderr %s", result.exitCode, result.stderr)
	}
}

func TestCounterNonInvariantStepMutant(t *testing.T) {
	t.Parallel()
	program, path := counterShapes(t)
	step, counter := -1, -1
	for _, function := range program.Functions {
		if function.Name == "changingStep" {
			step = function.Parameters[0]
		}
	}
	for id, local := range program.Locals {
		if local.Name == "changingStepCounter" {
			counter = id
		}
	}
	if step < 0 || counter < 0 {
		t.Fatal("missing non-invariant loop")
	}
	if program.Locals[counter].Counter || program.Locals[counter].CounterGuard != nil {
		t.Fatal("non-invariant step was accepted")
	}
	// Forge precisely the plan that dropping the step-invariance requirement
	// would admit. The entry step is 1, but the body changes it to 0.5.
	program.Locals[counter].CounterGuard = &ir.CounterGuard{
		Bound: ir.NumberConstant{Value: 4}, Step: ir.Read{Local: step, Of: ir.Number},
		Ascending: true, Inclusive: true,
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must execute cleanly: %+v", result)
	}
	if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
		t.Fatalf("non-invariant mutant survived: %s", difference)
	}
	t.Log("Node caught the non-invariant step mutant")
}
