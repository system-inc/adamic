package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var statementsSmallTypeScriptFixtures = []string{
	"internal/oracle/testdata/statements_small_typescript/nonnull.a",
	"internal/oracle/testdata/statements_small_typescript/missing.a",
	"internal/oracle/testdata/statements_small_typescript/absent.a",
}

// The .a files preserve the source witnesses. Only their scratch .ts copies
// use the TypeScript assertion frontend; production .a assertions stay refused.
func statementsSmallTypeScript(test *testing.T, path string) string {
	test.Helper()
	source, err := os.ReadFile(filepath.Join(repository, path))
	if err != nil {
		test.Fatal(err)
	}
	typed := filepath.Join(test.TempDir(), strings.TrimSuffix(filepath.Base(path), ".a")+".ts")
	if err := os.WriteFile(typed, source, 0644); err != nil {
		test.Fatal(err)
	}
	return typed
}

// Give the TypeScript adapter its logical name before lowering. Replacing a
// scratch path in emitted C afterward leaves that path's string lengths behind.
func statementsSmallTypeScriptCounted(test *testing.T, path string) (string, func() string) {
	test.Helper()
	typed := statementsSmallTypeScript(test, path)
	logical := "<repository>/" + strings.TrimSuffix(path, ".a") + ".ts"
	checked, err := load.LoadDisplayed([]string{typed}, func(name string) string {
		if name == typed {
			return logical
		}
		return name
	})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	return countPortable(test, source, ""), func() string {
		return countedRun(test, path, typed, source, false, nil, false, false)
	}
}

func init() {
	for _, path := range statementsSmallTypeScriptFixtures {
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{path, func(test *testing.T) (string, func() string) {
				return statementsSmallTypeScriptCounted(test, path)
			}},
		)
	}
}

func TestStatementsSmallTypeScriptCountInputsAreStable(test *testing.T) {
	test.Parallel()
	recorded, err := recordedCounts()
	if err != nil {
		test.Fatal(err)
	}
	for _, path := range statementsSmallTypeScriptFixtures {
		first, measure := statementsSmallTypeScriptCounted(test, path)
		second, _ := statementsSmallTypeScriptCounted(test, path)
		if first != second {
			test.Errorf("%s: count inputs depend on scratch location", path)
		}
		if strings.Contains(first, "TestStatementsSmallTypeScript") {
			test.Errorf("%s: count inputs name the test's scratch location", path)
		}
		if row := measure(); row != recorded[path] {
			test.Errorf("counts changed: got %s want %s", row, recorded[path])
		}
	}
}

func TestStatementsSmallTypeScriptIncrement(test *testing.T) {
	test.Parallel()
	for _, path := range statementsSmallTypeScriptFixtures {
		test.Run(filepath.Base(path), func(test *testing.T) {
			typed := statementsSmallTypeScript(test, path)
			program, err := lowered(test, typed)
			if err != nil {
				test.Fatal(err)
			}
			node := onNode(test, typed)
			want := node
			if filepath.Base(path) != "nonnull.a" {
				if difference := disagreement(run{stdout: []byte("before\nafter\n")}, node); difference != "" {
					test.Fatal("source Node: " + difference)
				}
				if len(program.NonNullChecks.Sites) != 1 || program.NonNullChecks.Checked != 1 {
					test.Fatalf("want one checked assertion, got %#v", program.NonNullChecks)
				}
				site := program.NonNullChecks.Sites[0]
				if site.Expression != "state.index!" || !strings.HasSuffix(site.Where, ":3:1") {
					test.Fatalf("wrong assertion site %#v", site)
				}
				want = run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed at " +
					site.Where +
					": state.index! is null or undefined\n"), exitCode: 70}
			}
			native, _ := nativelyUncached(test, program)
			for _, got := range []run{native, onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf(
						"%s: got exit %d stdout %q stderr %q; want exit %d stdout %q stderr %q",
						difference,
						got.exitCode,
						got.stdout,
						got.stderr,
						want.exitCode,
						want.stdout,
						want.stderr,
					)
				}
			}
		})
	}
}
