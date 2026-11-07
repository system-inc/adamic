package regexp

import (
	"errors"
	"testing"
)

func TestMatcherStepLimitBoundary(test *testing.T) {
	test.Parallel()
	program, err := Compile("ab", "")
	if err != nil {
		test.Fatal(err)
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
			_, matched, err := program.run([]uint16{'a', 'b'}, 0, []int{0, -1}, &budget)
			if testCase.wantLimit {
				if !errors.Is(err, ErrStepLimit) || matched {
					test.Fatalf(
						"limit %d: matched=%t error=%v; want ErrStepLimit before accept",
						testCase.limit,
						matched,
						err,
					)
				}
			} else if err != nil || !matched {
				test.Fatalf(
					"limit %d: matched=%t error=%v; want a match at exactly the limit",
					testCase.limit,
					matched,
					err,
				)
			}
			if budget.steps != testCase.wantSteps {
				test.Fatalf("limit %d: completed %d steps, want %d", testCase.limit, budget.steps, testCase.wantSteps)
			}
			// The public matcher must propagate exhaustion as an error, never a miss.
			matcher := program.New()
			matcher.StepLimit = testCase.limit
			match, err := matcher.ExecString("ab")
			if testCase.wantLimit {
				if !errors.Is(err, ErrStepLimit) || match != nil {
					test.Fatalf(
						"ExecString limit %d: match=%+v error=%v; want nil, ErrStepLimit",
						testCase.limit,
						match,
						err,
					)
				}
			} else if err != nil ||
				match == nil ||
				len(match.Captures) != 1 ||
				match.Captures[0] != (Capture{Start: 0, End: 2}) {
				test.Fatalf("ExecString limit %d: match=%+v error=%v; want capture [0,2]", testCase.limit, match, err)
			}
		})
	}
}
