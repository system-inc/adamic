package regexp

import (
	"errors"
	"testing"
)

func TestRegularCompileBudget(test *testing.T) {
	test.Parallel()
	pattern := "(?:)"
	for k := 0; k < 7; k++ {
		pattern = "(?:" + pattern + "){32}"
	}
	program, err := Compile(pattern, "")
	if err != nil {
		test.Fatal(err)
	}
	if program.regular != nil {
		test.Fatal("empty counted expansion must retain the VM")
	}
	matcher := program.New()
	matcher.StepLimit = 100
	if _, err := matcher.Exec(nil); !errors.Is(err, ErrStepLimit) {
		test.Fatalf("counted empty VM budget: %v", err)
	}
}

func TestRegularPriorityNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for _, pattern := range []string{`^(a)\1$`, `a.*b`, `a.*?b`, `a.*z|b`, `(a|aa)(a?)`, `(a|(b))+`, `(a?){2,4}?`} {
		for _, input := range []string{"aaxb", "abzz", "abxx", "aa", "aba", ""} {
			units := make([]uint16, len(input))
			for k := range input {
				units[k] = uint16(input[k])
			}
			cases = append(cases, executionCase{Pattern: pattern, Input: units, Source: "regular priority"})
		}
	}
	compareExecutionCases(test, cases, true)
}
