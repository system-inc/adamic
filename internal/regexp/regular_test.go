package regexp

import (
	"errors"
	"testing"
)

func TestRegularCompileBudget(t *testing.T) {
	t.Parallel()
	pattern := "(?:)"
	for k := 0; k < 7; k++ {
		pattern = "(?:" + pattern + "){32}"
	}
	program, err := Compile(pattern, "")
	if err != nil {
		t.Fatal(err)
	}
	if program.regular != nil {
		t.Fatal("empty counted expansion must retain the VM")
	}
	matcher := program.New()
	matcher.StepLimit = 100
	if _, err := matcher.Exec(nil); !errors.Is(err, ErrStepLimit) {
		t.Fatalf("counted empty VM budget: %v", err)
	}
}

func TestRegularPriorityNode(t *testing.T) {
	t.Parallel()
	var cases []executionCase
	for _, pattern := range []string{`a.*b`, `a.*?b`, `a.*z|b`, `(a|aa)(a?)`, `(a|(b))+`, `(a?){2,4}?`} {
		for _, input := range []string{"aaxb", "abzz", "abxx", "aa", "aba", ""} {
			units := make([]uint16, len(input))
			for k := range input {
				units[k] = uint16(input[k])
			}
			cases = append(cases, executionCase{Pattern: pattern, Input: units, Source: "regular priority"})
		}
	}
	compareExecutionCases(t, cases, true)
}
