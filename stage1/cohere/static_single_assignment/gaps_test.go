package static_single_assignment

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// gaps are GAPS.md's smallest programs, each with what Node prints running it, and where stage 0 stands
// on it: refused with the words GAPS.md records (refused, a Refused stage 0 shouldn't give, or notYet),
// or closed (neither), when it must lower and print natively what it prints on Node, leaking nothing.
// So when a gap moves, this test says so, and the port's workaround for it can go.
var gaps = []struct {
	path    string
	refused string
	notYet  string
	stdout  string
}{
	{path: "gaps/1_block_arrow_returning_class.ts", refused: "a value without nominal ancestry seen as Box", stdout: "2\n"},
	{path: "gaps/2_call_through_optional_chain.ts", notYet: "a call through ?. (an optional call)", stdout: "7 undefined\n"},
}

func TestEachGapStandsWhereGapsMdSaysItDoes(t *testing.T) {
	t.Parallel()
	for _, gap := range gaps {
		t.Run(gap.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(gap.path)
			if err != nil {
				t.Fatal(err)
			}
			// What the program means, whatever stage 0 makes of it.
			nodeRun := onNode(t, path)
			if nodeRun.exitCode != 0 || string(nodeRun.stdout) != gap.stdout {
				t.Fatalf("on Node: exit %d, stdout %q; GAPS.md records %q", nodeRun.exitCode, nodeRun.stdout, gap.stdout)
			}

			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			switch {
			case gap.refused != "":
				var refused *lower.Refused
				if err == nil {
					t.Fatalf("this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)")
				}
				if !errors.As(err, &refused) || refused.What != gap.refused {
					t.Fatalf("stage 0 refuses another way now (%v), where GAPS.md records %q", err, gap.refused)
				}
				return
			case gap.notYet != "":
				var notYet *lower.NotYet
				if err == nil {
					t.Fatalf("this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)")
				}
				if !errors.As(err, &notYet) || notYet.What != gap.notYet {
					t.Fatalf("stage 0 refuses another way now (%v), where GAPS.md records %q", err, gap.notYet)
				}
				return
			}
			if err != nil {
				t.Fatalf("stage 0 refuses this gap now (%v): record that in GAPS.md, where it says the program lowers", err)
			}
			nativeRun, sanitized := natively(t, lowered)
			if nativeRun.exitCode != 0 || string(nativeRun.stdout) != gap.stdout {
				t.Errorf("natively: exit %d, stdout %q, stderr %q; Node prints %q", nativeRun.exitCode, nativeRun.stdout, nativeRun.stderr, gap.stdout)
			}
			if leaked := leaks(t, lowered, sanitized); leaked != "" {
				t.Errorf("leaks:\n%s", leaked)
			}
		})
	}
}
