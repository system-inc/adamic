package regexp

import (
	"errors"
	"testing"
)

func TestRegularCompileBudget(test *testing.T) {
	test.Parallel()
	pattern := "(?:)"
	for index := 0; index < 7; index++ {
		pattern = "(?:" + pattern + "){32}"
	}
	program, compileError := Compile(pattern, "")
	if compileError != nil {
		test.Fatal(compileError)
	}
	if program.regular != nil {
		test.Fatal("empty counted expansion must retain the VM")
	}
	matcher := program.New()
	matcher.StepLimit = 100
	if _, execError := matcher.Exec(nil); !errors.Is(execError, ErrStepLimit) {
		test.Fatalf("counted empty VM budget: %v", execError)
	}
}

func TestRegularPriorityNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for _, pattern := range []string{`^(a)\1$`, `a.*b`, `a.*?b`, `a.*z|b`, `(a|aa)(a?)`, `(a|(b))+`, `(a?){2,4}?`} {
		for _, input := range []string{"aaxb", "abzz", "abxx", "aa", "aba", ""} {
			units := make([]uint16, len(input))
			for index := range input {
				units[index] = uint16(input[index])
			}
			cases = append(cases, executionCase{Pattern: pattern, Input: units, Source: "regular priority"})
		}
	}
	compareExecutionCases(test, cases, true)
}
