package regexp

import (
	"errors"
	"testing"
)

func TestMatcherStepLimitBoundary(test *testing.T) {
	test.Parallel()
	program, compileError := Compile("ab", "")
	if compileError != nil {
		test.Fatal(compileError)
	}
	// Two character instructions followed by accept require exactly three steps.
	// A budget of two must interrupt before accept, rather than report no match.
	for _, testCase := range []struct {
		name      string
		limit     uint64
		wantSteps uint64
		wantLimit bool
	}{
		{"exactly-limit", 3, 3, false},
		{"one-step-past-limit", 2, 2, true},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			budget := executionBudget{limit: testCase.limit}
			_, matched, runError := program.run([]uint16{'a', 'b'}, 0, []int{0, -1}, &budget)
			if testCase.wantLimit {
				if !errors.Is(runError, ErrStepLimit) || matched {
					test.Fatalf(
						"limit %d: matched=%t error=%v; want ErrStepLimit before accept",
						testCase.limit,
						matched,
						runError,
					)
				}
			} else if runError != nil || !matched {
				test.Fatalf(
					"limit %d: matched=%t error=%v; want a match at exactly the limit",
					testCase.limit,
					matched,
					runError,
				)
			}
			if budget.steps != testCase.wantSteps {
				test.Fatalf("limit %d: completed %d steps, want %d", testCase.limit, budget.steps, testCase.wantSteps)
			}
			// The public matcher must propagate exhaustion as an error, never a miss.
			matcher := program.New()
			matcher.StepLimit = testCase.limit
			match, execStringError := matcher.ExecString("ab")
			if testCase.wantLimit {
				if !errors.Is(execStringError, ErrStepLimit) || match != nil {
					test.Fatalf(
						"ExecString limit %d: match=%+v error=%v; want nil, ErrStepLimit",
						testCase.limit,
						match,
						execStringError,
					)
				}
			} else if execStringError != nil ||
				match == nil ||
				len(match.Captures) != 1 ||
				match.Captures[0] != (Capture{Start: 0, End: 2}) {
				test.Fatalf(
					"ExecString limit %d: match=%+v error=%v; want capture [0,2]",
					testCase.limit,
					match,
					execStringError,
				)
			}
		})
	}
}
