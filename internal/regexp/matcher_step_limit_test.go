package regexp

import (
	"errors"
	"testing"
)

func TestMatcherStepLimitBoundary(t *testing.T) {
	t.Parallel()
	program, err := Compile("ab", "")
	if err != nil {
		t.Fatal(err)
	}
	// Two character instructions followed by accept require exactly three steps.
	// A budget of two must interrupt before accept, rather than report no match.
	for _, test := range []struct {
		name      string
		limit     uint64
		wantSteps uint64
		wantLimit bool
	}{
		{"exactly-limit", 3, 3, false},
		{"one-step-past-limit", 2, 2, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			budget := executionBudget{limit: test.limit}
			_, matched, err := program.run([]uint16{'a', 'b'}, 0, []int{0, -1}, &budget)
			if test.wantLimit {
				if !errors.Is(err, ErrStepLimit) || matched {
					t.Fatalf("limit %d: matched=%t error=%v; want ErrStepLimit before accept", test.limit, matched, err)
				}
			} else if err != nil || !matched {
				t.Fatalf("limit %d: matched=%t error=%v; want a match at exactly the limit", test.limit, matched, err)
			}
			if budget.steps != test.wantSteps {
				t.Fatalf("limit %d: completed %d steps, want %d", test.limit, budget.steps, test.wantSteps)
			}
			// The public matcher must propagate exhaustion as an error, never a miss.
			matcher := program.New()
			matcher.StepLimit = test.limit
			match, err := matcher.ExecString("ab")
			if test.wantLimit {
				if !errors.Is(err, ErrStepLimit) || match != nil {
					t.Fatalf("ExecString limit %d: match=%+v error=%v; want nil, ErrStepLimit", test.limit, match, err)
				}
			} else if err != nil || match == nil || len(match.Captures) != 1 || match.Captures[0] != (Capture{Start: 0, End: 2}) {
				t.Fatalf("ExecString limit %d: match=%+v error=%v; want capture [0,2]", test.limit, match, err)
			}
		})
	}
}
