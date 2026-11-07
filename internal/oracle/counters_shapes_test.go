package oracle

import (
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

func counterShapes(test *testing.T) (*ir.Program, string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/counters_shapes.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	return program, path
}

func TestCounterShapesProofAndFallbacks(test *testing.T) {
	test.Parallel()
	program, _ := counterShapes(test)
	expected := map[string]string{
		"squareCounter":             "guarded",
		"productCounter":            "guarded",
		"parameterCounter":          "guarded",
		"leftCounter":               "integer",
		"rightCounter":              "integer",
		"sumCounter":                "integer",
		"multipliedCounter":         "integer",
		"guardedCounter":            "guarded",
		"edgeCounter":               "guarded",
		"descendingCounter":         "guarded",
		"changingBoundCounter":      "double",
		"changingStepCounter":       "double",
		"dependencyOuterCounter":    "guarded",
		"dependentCounter":          "guarded",
		"largeCounter":              "integer",
		"smallCounter":              "integer",
		"largeStepCounter":          "guarded",
		"overflowingProductCounter": "double",
		"squareEdgeCounter":         "guarded",
		"independentOuterCounter":   "guarded",
		"independentInnerCounter":   "guarded",
		"typedSquareCounter":        "integer",
		// Its bound is a product of typed-array lengths, which can pass 2^53, so the proof declines it and the guarded
		// fallback does not take .length.
		"typedProductCounter":  "double",
		"localCounter":         "guarded",
		"fractionBoundCounter": "double",
		"fractionStepCounter":  "double",
		"capturedStepCounter":  "double",
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
			test.Errorf("%s: %s, want %s", local.Name, got, want)
		}
	}
	for name := range expected {
		if !found[name] {
			test.Errorf("missing %s", name)
		}
	}
	// Trace actual integer declarations: fractional/nonfinite bounds, bad steps,
	// and updates crossing 2^53 must execute the original double version.
	code := native.C(program)
	for _, local := range program.Locals {
		if local.Name == "independentInnerCounter" {
			declaration := "int64_t " +
				emittedLocalSymbol(test, code, "adamic_local_", "independentInnerCounter") +
				" ="
			if strings.Count(code, declaration) != 1 {
				test.Fatal("independent nested guards duplicated the integer path")
			}
		}
	}
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n#include <stdio.h>", 1)
	for _, name := range []string{
		"guardedCounter",
		"edgeCounter",
		"descendingCounter",
		"squareCounter",
		"parameterCounter",
		"dependencyOuterCounter",
		"dependentCounter",
		"largeStepCounter",
		"squareEdgeCounter",
		"independentOuterCounter",
		"independentInnerCounter",
		"localCounter",
	} {
		declaration := regexp.MustCompile(`(?m)^(\s*int64_t adamic_local_\d+_` + name + ` = [^\n]+;)$`)
		if !declaration.MatchString(code) {
			test.Fatalf("missing integer path: %s", name)
		}
		code = declaration.ReplaceAllString(code, `${1}`+"\n\tfprintf(stderr, \"integer:"+name+"\\n\");")
	}
	binary := filepath.Join(test.TempDir(), "trace")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := execute(test, binary)
	expectedTrace := "integer:squareCounter\ninteger:parameterCounter\ninteger:guardedCounter\ninteger:edgeCounter\ninteger:descendingCounter\ninteger:dependencyOuterCounter\ninteger:dependentCounter\ninteger:dependentCounter\ninteger:dependentCounter\ninteger:squareEdgeCounter\ninteger:squareEdgeCounter\ninteger:squareEdgeCounter\ninteger:independentOuterCounter\ninteger:independentInnerCounter\ninteger:independentInnerCounter\ninteger:localCounter\n"
	if result.exitCode != 0 || string(result.stderr) != expectedTrace {
		test.Fatalf("wrong fallback trace: exit %d stderr %s", result.exitCode, result.stderr)
	}
}

func TestCounterNonInvariantStepMutant(test *testing.T) {
	test.Parallel()
	program, path := counterShapes(test)
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
		test.Fatal("missing non-invariant loop")
	}
	if program.Locals[counter].Counter || program.Locals[counter].CounterGuard != nil {
		test.Fatal("non-invariant step was accepted")
	}
	// Forge precisely the plan that dropping the step-invariance requirement
	// would admit. The entry step is 1, but the body changes it to 0.5.
	program.Locals[counter].CounterGuard = &ir.CounterGuard{
		Bound: ir.NumberConstant{Value: 4}, Step: ir.Read{Local: step, Of: ir.Number},
		Ascending: true, Inclusive: true,
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := execute(test, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("mutant must execute cleanly: %+v", result)
	}
	if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
		test.Fatalf("non-invariant mutant survived: %s", difference)
	}
	test.Log("Node caught the non-invariant step mutant")
}

// Corrupt only one representation's initial value. These mutants must finish
// without sanitizer diagnostics and disagree with Node, so both branches must
// actually execute on real inputs rather than merely appear in generated C.
func counterPathMutation(test *testing.T, representation, initial string) {
	test.Helper()
	program, path := counterShapes(test)
	truth := onNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	code := native.C(program)
	declaration := regexp.MustCompile(`(?m)(` + representation + ` adamic_local_\d+_guardedCounter = )[^;\n]+;`)
	if len(declaration.FindAllString(code, -1)) != 1 {
		test.Fatal("missing unique counter path", representation)
	}
	code = declaration.ReplaceAllString(code, `${1}`+initial+";")
	binary := filepath.Join(test.TempDir(), "counter-path-mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal("mutant must compile:", err)
	}
	result := execute(test, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("mutant must execute cleanly: %+v", result)
	}
	if difference := disagreement(truth, result); difference != "stdout differs" {
		test.Fatalf("%s path mutant survived: %s", representation, difference)
	}
	test.Log(representation + " path mutation caught by Node stdout difference")
}

func TestCounterIntegerPathMutant(test *testing.T) {
	test.Parallel()
	counterPathMutation(test, "int64_t", "1")
}

func TestCounterDoubleFallbackMutant(test *testing.T) {
	test.Parallel()
	counterPathMutation(test, "double", "0.25")
}
