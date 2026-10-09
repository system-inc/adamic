package main

import (
	"strings"
	"testing"
)

// The types phase's checking starts alongside the fix phase's walk (startTypeCheck), and its result is used
// only if the walk leaves the graph in place. A type error is reported exactly once either way: taken from
// the early check when nothing was rewritten, and from a check of the rebuilt graph when the fixer rewrote
// a file, whose early result described bytes that are gone (#zqsdzbq).
func TestATypeErrorFoundAlongsideTheWalkIsReportedOnce(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	for _, arguments := range [][]string{{"--no-fix"}, {"--no-fix", "--no-cache"}, {}} {
		name := strings.Join(arguments, " ")
		if name == "" {
			name = "a writing run that rewrites a file"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			fixScopeProject(t, root, map[string]string{"Broken.ts": "export const broken: number = \"text\";\n"})
			output, code := runCohere(t, binary, root, arguments...)
			if count := strings.Count(output, "error TS2322"); count != 1 {
				t.Errorf("the type error was reported %d times, want once:\n%s", count, output)
			}
			if code == 0 {
				t.Errorf("a run with a type error exited 0:\n%s", output)
			}
			if len(arguments) == 0 && !strings.Contains(output, "graph rebuilt") {
				t.Fatalf("the writing run rewrote nothing, so the early check was never thrown away and this proves nothing:\n%s", output)
			}
		})
	}
}
