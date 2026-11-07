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
	{
		path:    "gaps/1_block_arrow_returning_class.ts",
		refused: "a value without nominal ancestry seen as Box",
		stdout:  "2\n",
	},
	{
		path:   "gaps/2_call_through_optional_chain.ts",
		notYet: "a call through ?. (an optional call)",
		stdout: "7 undefined\n",
	},
}

func TestEachGapStandsWhereGapsMdSaysItDoes(test *testing.T) {
	test.Parallel()
	for _, gap := range gaps {
		test.Run(gap.path, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(gap.path)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			// What the program means, whatever stage 0 makes of it.
			nodeRun := onNode(test, path)
			if nodeRun.exitCode != 0 || string(nodeRun.stdout) != gap.stdout {
				test.Fatalf(
					"on Node: exit %d, stdout %q; GAPS.md records %q",
					nodeRun.exitCode,
					nodeRun.stdout,
					gap.stdout,
				)
			}

			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatalf("Load: %v", loadError)
			}
			lowered, lowerError := lower.Lower(context.Background(), program)
			switch {
			case gap.refused != "":
				var refused *lower.Refused
				if lowerError == nil {
					test.Fatalf(
						"this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				if !errors.As(lowerError, &refused) || refused.What != gap.refused {
					test.Fatalf(
						"stage 0 refuses another way now (%v), where GAPS.md records %q",
						lowerError,
						gap.refused,
					)
				}
				return
			case gap.notYet != "":
				var notYet *lower.NotYet
				if lowerError == nil {
					test.Fatalf(
						"this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				if !errors.As(lowerError, &notYet) || notYet.What != gap.notYet {
					test.Fatalf(
						"stage 0 refuses another way now (%v), where GAPS.md records %q",
						lowerError,
						gap.notYet,
					)
				}
				return
			}
			if lowerError != nil {
				test.Fatalf(
					"stage 0 refuses this gap now (%v): record that in GAPS.md, where it says the program lowers",
					lowerError,
				)
			}
			sanitized := filepath.Join(gapSanitized(test, gap.path), "port")
			nativeRun := sanitizedRun(test, sanitized)
			if nativeRun.exitCode != 0 || string(nativeRun.stdout) != gap.stdout {
				test.Errorf(
					"natively: exit %d, stdout %q, stderr %q; Node prints %q",
					nativeRun.exitCode,
					nativeRun.stdout,
					nativeRun.stderr,
					gap.stdout,
				)
			}
			if leaked := leaks(test, lowered, sanitized); leaked != "" {
				test.Errorf("leaks:\n%s", leaked)
			}
		})
	}
}
