package graphql

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// gaps are GAPS.md's smallest programs, each with what Node prints running it, and where stage 0 stands
// on it: refused with the words GAPS.md records (notYet), lowered to C that clang refuses with the words
// it records (badC), or closed (neither), when it must lower and print natively what it prints on Node,
// leaking nothing. So when a gap moves, this test says so, and the port's workaround for it can go.
var gaps = []struct {
	path   string
	notYet string
	badC   string
	stdout string
}{
	{path: "gaps/1_throwing_function_value.ts", stdout: "too long: bb\n"},
	{path: "gaps/2_error_made_elsewhere.ts", notYet: "throwing an Error that isn't made where it's thrown or caught by the catch around it", stdout: "Unexpected token at 3.\n"},
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
			if gap.notYet != "" {
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
				t.Fatalf("stage 0 refuses this gap now (%v): record that in GAPS.md, where it says the program lowers", err)
			}
			if gap.badC != "" {
				buildErr := native.Build(native.C(lowered), filepath.Join(t.TempDir(), "gap"), native.Options{Sanitize: true})
				if buildErr == nil {
					t.Fatalf("clang compiles this gap's C now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)")
				}
				if !strings.Contains(buildErr.Error(), gap.badC) {
					t.Fatalf("clang refuses this gap's C another way now: %v", buildErr)
				}
				return
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
