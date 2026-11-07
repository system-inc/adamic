package regexp

import (
	"reflect"
	"testing"
)

// Mutate compiled programs, never the implementation. Each mutant must still
// execute successfully and differ from Node on one of the new inputs.
func TestMatcherOct6Mutants(test *testing.T) {
	test.Parallel()
	mutants := []struct {
		name      string
		caseIndex int
		mutate    func(*Program)
	}{
		{"keep stale nested captures", 0, func(program *Program) {
			for index := range program.code {
				program.code[index].clear = nil
			}
		}},
		{"reverse alternation priority in lookbehind", 7, func(program *Program) {
			for _, instruction := range program.code {
				if instruction.look != nil {
					for index := range instruction.look.code {
						if instruction.look.code[index].op == opSplit {
							instruction.look.code[index].x, instruction.look.code[index].y = instruction.look.code[index].y, instruction.look.code[index].x
						}
					}
				}
			}
		}},
		{"ignore repeated backreferences", 15, func(program *Program) {
			for index := range program.code {
				if program.code[index].op == opReference {
					program.code[index].references = nil
				}
			}
		}},
		{"case sensitive non-ASCII backreference", 21, func(program *Program) {
			for index := range program.code {
				if program.code[index].op == opReference {
					program.code[index].flags.IgnoreCase = false
				}
			}
		}},
	}
	cases := oct6Cases()
	expected := nodeResults(test, cases)
	for _, mutant := range mutants {
		test.Run(mutant.name, func(test *testing.T) {
			executionCase := cases[mutant.caseIndex]
			program, compileError := Compile(executionCase.Pattern, executionCase.Flags)
			if compileError != nil {
				test.Fatal(compileError)
			}
			control, matcherResultError := matcherResult(program.New(), executionCase.Input)
			if matcherResultError != nil || !reflect.DeepEqual(control, expected[mutant.caseIndex]) {
				test.Fatalf("control: %v %+v", matcherResultError, control)
			}
			mutant.mutate(program)
			exp := program.New()
			exp.StepLimit = 100000
			got, matcherResultError := matcherResult(exp, executionCase.Input)
			if matcherResultError != nil {
				test.Fatalf("mutant must execute, not hit a limit: %v", matcherResultError)
			}
			if reflect.DeepEqual(got, expected[mutant.caseIndex]) {
				test.Fatal("mutant survived")
			}
			test.Logf("caught: mutant=%+v Node=%+v", got, expected[mutant.caseIndex])
		})
	}
	loops := oct6Loops()
	want := oct6NodeLoops(test, loops)
	for _, mutant := range []struct {
		name   string
		index  int
		mutate func(*Program)
	}{
		{"global ignores lastIndex", 1, func(program *Program) { program.flags.Global = false }},
		{
			"sticky searches across a gap",
			6,
			func(program *Program) { program.flags.Sticky = false; program.flags.Global = true },
		},
		{"empty pattern consumes a unit", 3, func(program *Program) {
			program.code = append(
				[]instruction{{op: opSet, direction: 1, set: characterSet{ranges: []RuneRange{{0, 0xffff}}}}},
				program.code...,
			)
		}},
	} {
		test.Run(mutant.name, func(test *testing.T) {
			loop := loops[mutant.index]
			program, compileError := Compile(loop.Pattern, loop.Flags)
			if compileError != nil {
				test.Fatal(compileError)
			}
			mutant.mutate(program)
			exp := program.New()
			exp.LastIndex = loop.LastIndex
			exp.StepLimit = 100000
			caught := false
			for step := 0; step < loop.Steps; step++ {
				got, matcherResultError := matcherResult(exp, loop.Input)
				if matcherResultError != nil {
					test.Fatal(matcherResultError)
				}
				if !reflect.DeepEqual(got, want[mutant.index][step]) {
					caught = true
					test.Logf("caught at step %d: mutant=%+v Node=%+v", step, got, want[mutant.index][step])
					break
				}
			}
			if !caught {
				test.Fatal("mutant survived")
			}
		})
	}
}
