package mediaquery

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// gaps are GAPS.md's smallest programs, each with what Node prints running it, and where stage 0 stands
// on it: not yet lowered, with the words GAPS.md records (notYet); refused by a rule of 0.1, with the
// words it records (refused); lowered to C that clang refuses with the words it records (badC); or
// closed (an entry with only stdout), when it must lower and print natively what it prints on Node,
// leaking nothing.
// So when a gap moves, this test says so, and the port's workaround for it can go.
var gaps = []struct {
	path    string
	notYet  string
	refused string
	badC    string
	stdout  string
}{
	{path: "gaps/1_generic_return.ts", stdout: "a\n"},
	{path: "gaps/1_generic_parameter.ts", stdout: "3\n"},
	{path: "gaps/2_return_undefined.ts", stdout: "none\n"},
	{path: "gaps/3_optional_chaining_array.ts", stdout: "1 0\n"},
	{path: "gaps/4_mutable_child_array.ts", stdout: "2\n"},
	{path: "gaps/5_local_children_array.ts", stdout: "2\n"},
}

func TestEachGapStandsWhereGapsMdSaysItDoes(test *testing.T) {
	test.Parallel()
	for _, gap := range gaps {
		test.Run(gap.path, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(gap.path)
			if err != nil {
				test.Fatal(err)
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

			program, err := load.Load([]string{path})
			if err != nil {
				test.Fatalf("Load: %v", err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if gap.notYet != "" {
				if err == nil {
					test.Fatalf(
						"this gap lowers now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					test.Fatalf("stage 0 refuses another way now: %v", err)
				}
				if notYet.What != gap.notYet {
					test.Fatalf("stage 0 refuses with %q now, where GAPS.md records %q", notYet.What, gap.notYet)
				}
				return
			}
			if gap.refused != "" {
				if err == nil {
					test.Fatalf(
						"0.1 accepts this program now: mark the gap closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				var refused *lower.Refused
				if !errors.As(err, &refused) {
					test.Fatalf("stage 0 refuses another way now: %v", err)
				}
				if refused.What != gap.refused {
					test.Fatalf("stage 0 refuses with %q now, where GAPS.md records %q", refused.What, gap.refused)
				}
				return
			}
			if err != nil {
				test.Fatalf(
					"stage 0 refuses this gap now (%v): record that in GAPS.md, where it says the program lowers",
					err,
				)
			}
			if gap.badC != "" {
				buildErr := native.Build(
					native.C(lowered),
					filepath.Join(test.TempDir(), "gap"),
					native.Options{Sanitize: true},
				)
				if buildErr == nil {
					test.Fatalf(
						"clang compiles this gap's C now: mark it closed here and in GAPS.md, and undo the port's workaround for it (grep -n 'gap N' *.ts)",
					)
				}
				if !strings.Contains(buildErr.Error(), gap.badC) {
					test.Fatalf("clang refuses this gap's C another way now: %v", buildErr)
				}
				return
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
