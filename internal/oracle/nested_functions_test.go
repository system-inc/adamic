package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{
		"minimal",
		"captures",
		"hoisting",
		"mutual",
		"returned",
		"array",
		"three_levels",
		"tdz",
		"tdz_write",
		"weak",
		"destructured",
		"destructured_tdz",
		"mixed",
		"pattern_parameter",
		"owned_parameter",
		"generic_capture",
		"assignment_return",
		"reduced_parameter_unused",
		"reduced_parameter_unreachable",
		"reduced_parameter_escaped",
		"callback_recursive",
		"callback_escaped",
		"callback_levels",
		"callback_late_capture",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/nested_" + name + ".a", true, false})
	}
}

// The output is unchanged when siblings capture each other's closure bindings.
// Only the leak check exposes the resulting strong cycle.
func TestNestedSiblingCycleMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/nested_mutual.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	bindings := map[int]int{}
	owner := -1
	for local, binding := range program.Locals {
		if binding.NestedFunction > 0 {
			bindings[binding.NestedFunction-1] = local
			owner = binding.Function
			program.Locals[local].Captured = true
			program.Locals[local].Preallocated = true
		}
	}
	if len(bindings) != 2 || owner < 0 {
		test.Fatal("mutant requires two siblings")
	}
	cells := []int{}
	for function := range program.Functions {
		if local, ok := bindings[function]; ok {
			cells = append(cells, local)
		}
	}
	for function := range bindings {
		target := &program.Functions[function]
		target.Environment = append(target.Environment, cells...)
		for index, statement := range target.Body {
			result, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			conditional, ok := result.Value.(ir.Conditional)
			if !ok {
				test.Fatal("mutant expects a conditional return")
			}
			call, ok := conditional.WhenNot.(ir.CallClosure)
			if !ok || call.Direct == 0 {
				test.Fatal("mutant expects a direct sibling call")
			}
			local := bindings[call.Direct-1]
			call.Direct = 0
			call.Closure = ir.Read{Local: local, Of: ir.Closure}
			conditional.WhenNot = call
			result.Value = conditional
			target.Body[index] = result
		}
	}
	prologue := []ir.Statement{}
	for _, local := range cells {
		prologue = append(prologue, ir.Declare{Local: local, Uninitialized: true})
	}
	program.Functions[owner].Body = append(prologue, program.Functions[owner].Body...)
	oracle := onNode(test, path)
	native, binary := nativelyUncached(test, program)
	if difference := disagreement(oracle, native); difference != "" {
		test.Fatalf("mutant must preserve output: %s", difference)
	}
	report := leaksUncached(test, program, binary)
	if report == "" {
		test.Fatal("sibling capture cycle escaped the leak check")
	}
	test.Logf("sibling capture mutant caught only by the leak check:\n%s", report)
}

func TestNestedCycleRefusal(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/refusals/nested_cycle.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	var refused *lower.Refused
	if errors.As(loweredError, &refused) && strings.Contains(loweredError.Error(), "adamic/cycle-capable") {
		return
	}
	if loweredError != nil {
		test.Fatalf("want cycle-capable refusal, got %v", loweredError)
	}
	oracle := onNode(test, path)
	actual, binary := nativelyUncached(test, program)
	if difference := disagreement(oracle, actual); difference != "" {
		test.Fatalf("accepted cycle differs from Node: %s", difference)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	if len(program.GraphTypes) == 0 {
		test.Fatal("cyclic canonical closure stayed counted")
	}
	if difference := disagreement(oracle, onJavaScriptBackend(test, program)); difference != "" {
		test.Fatal(difference)
	}
}

func TestNestedRebindingCheckerRefusal(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/refusals/nested_rebinding.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	_, loadError := load.Load([]string{path})
	if loadError == nil || !strings.Contains(loadError.Error(), "TS2630") {
		test.Fatalf("want checker refusal, got %v", loadError)
	}
}

func TestNestedCallbackCarrierMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/nested_callback_escaped.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.ForwardedNestedParent == 0 {
			continue
		}
		for index, statement := range function.Body {
			result, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			sum, ok := result.Value.(ir.Binary)
			if !ok {
				continue
			}
			call, ok := sum.Right.(ir.CallClosure)
			if !ok || call.Direct == 0 {
				continue
			}
			call.Closure = ir.ClosureSelf{}
			sum.Right = call
			result.Value = sum
			function.Body[index] = result
			changed = true
		}
	}
	if !changed {
		test.Fatal("carrier mutant did not change a call")
	}
	expected := onNode(test, path)
	actual, binary := nativelyUncached(test, program)
	if actual.exitCode != 0 {
		test.Fatalf("want output-only mutant, got %+v", actual)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	if disagreement(expected, actual) == "" {
		test.Fatal("wrong carrier escaped Node comparison")
	}
	test.Logf("wrong carrier caught by Node: %s", disagreement(expected, actual))
}
