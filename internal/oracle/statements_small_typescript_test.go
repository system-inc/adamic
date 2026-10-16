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
func statementsSmallTypeScript(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	typed := filepath.Join(t.TempDir(), strings.TrimSuffix(filepath.Base(path), ".a")+".ts")
	if err := os.WriteFile(typed, source, 0644); err != nil {
		t.Fatal(err)
	}
	return typed
}

// Give the TypeScript adapter its logical name before lowering. Replacing a
// scratch path in emitted C afterward leaves that path's string lengths behind.
func statementsSmallTypeScriptCounted(t *testing.T, path string) (string, func() string) {
	t.Helper()
	typed := statementsSmallTypeScript(t, path)
	logical := "<repository>/" + strings.TrimSuffix(path, ".a") + ".ts"
	checked, err := load.LoadDisplayed([]string{typed}, func(name string) string {
		if name == typed {
			return logical
		}
		return name
	})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	return countPortable(t, source, ""), func() string {
		return countedRun(t, path, typed, source, false, nil, false, false)
	}
}

func init() {
	for _, path := range statementsSmallTypeScriptFixtures {
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) {
			return statementsSmallTypeScriptCounted(t, path)
		}})
	}
}

func TestStatementsSmallTypeScriptCountInputsAreStable(t *testing.T) {
	t.Parallel()
	recorded, err := recordedCounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range statementsSmallTypeScriptFixtures {
		first, measure := statementsSmallTypeScriptCounted(t, path)
		second, _ := statementsSmallTypeScriptCounted(t, path)
		if first != second {
			t.Errorf("%s: count inputs depend on scratch location", path)
		}
		if strings.Contains(first, "TestStatementsSmallTypeScript") {
			t.Errorf("%s: count inputs name the test's scratch location", path)
		}
		if row := measure(); row != recorded[path] {
			t.Errorf("counts changed: got %s want %s", row, recorded[path])
		}
	}
}

func TestStatementsSmallTypeScriptIncrement(t *testing.T) {
	t.Parallel()
	for _, path := range statementsSmallTypeScriptFixtures {
		t.Run(filepath.Base(path), func(t *testing.T) {
			typed := statementsSmallTypeScript(t, path)
			program, err := lowered(t, typed)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, typed)
			want := node
			if filepath.Base(path) != "nonnull.a" {
				if difference := disagreement(run{stdout: []byte("before\nafter\n")}, node); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				if len(program.NonNullChecks.Sites) != 1 || program.NonNullChecks.Checked != 1 {
					t.Fatalf("want one checked assertion, got %#v", program.NonNullChecks)
				}
				site := program.NonNullChecks.Sites[0]
				if site.Expression != "state.index!" || !strings.HasSuffix(site.Where, ":3:1") {
					t.Fatalf("wrong assertion site %#v", site)
				}
				want = run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed at " + site.Where + ": state.index! is null or undefined\n"), exitCode: 70}
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: got exit %d stdout %q stderr %q; want exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr, want.exitCode, want.stdout, want.stderr)
				}
			}
		})
	}
}
