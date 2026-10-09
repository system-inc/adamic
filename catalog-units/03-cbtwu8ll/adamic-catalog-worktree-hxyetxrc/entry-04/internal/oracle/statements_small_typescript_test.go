package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		rows := []string{}
		base, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range statementsSmallTypeScriptFixtures {
			typed := statementsSmallTypeScript(t, path)
			relative, err := filepath.Rel(base, typed)
			if err != nil {
				t.Fatal(err)
			}
			row := counted(t, relative, false, nil, false, false)
			rows = append(rows, strings.Replace(row, "| "+relative+" |", "| "+path+" |", 1))
		}
		return rows
	})
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
