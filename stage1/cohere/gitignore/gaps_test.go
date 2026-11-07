package gitignore

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// gaps are GAPS.md's smallest programs, each with what Node prints running it. An open gap must still
// be refused by stage 0 with the words GAPS.md records; a closed one must lower and print natively what
// it prints on Node, leaking nothing. So when a gap closes, this test says so, and the port's
// workaround for it can go.
var gaps = []struct {
	path   string
	open   bool
	notYet string
	stdout string
}{
	{"gaps/1_from_char_code.ts", false, "", "hé\n"},
	{"gaps/2_bitwise.ts", false, "", "25 800 15 8 201 203 -201\n"},
	{"gaps/3_later_function.ts", false, "", "7\n"},
	{"gaps/3_later_method.ts", false, "", "3\n"},
	{"gaps/4_function_value.ts", false, "", "true\n"},
	{"gaps/5_array_from.ts", false, "", "4\n"},
	{"gaps/6_boolean_element.ts", false, "", "true true true\n"},
	{"gaps/7_maybe_number_compared.ts", false, "", "true\n"},
	{"gaps/8_tuple_value.ts", false, "", "bc 1\n"},
	{"gaps/9_return_panic.ts", false, "", "1\n"},
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
			if gap.open {
				if lowerError == nil {
					test.Fatalf(
						"this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				var notYet *lower.NotYet
				if !errors.As(lowerError, &notYet) {
					test.Fatalf("stage 0 refuses another way now: %v", lowerError)
				}
				if notYet.What != gap.notYet {
					test.Fatalf("stage 0 refuses with %q now, where GAPS.md records %q", notYet.What, gap.notYet)
				}
				return
			}
			if lowerError != nil {
				test.Fatalf("a closed gap is refused again: %v", lowerError)
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
