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
			if gap.open {
				if err == nil {
					t.Fatalf("this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)")
				}
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("stage 0 refuses another way now: %v", err)
				}
				if notYet.What != gap.notYet {
					t.Fatalf("stage 0 refuses with %q now, where GAPS.md records %q", notYet.What, gap.notYet)
				}
				return
			}
			if err != nil {
				t.Fatalf("a closed gap is refused again: %v", err)
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
