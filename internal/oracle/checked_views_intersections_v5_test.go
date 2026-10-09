package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func v5IntersectionSourceCounterfactual(t *testing.T, kind, fixture, member string) {
	t.Helper()
	program, path := interfaceFixture(t, "lane7/"+fixture)
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
		t.Fatal("source Node: " + difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value." + member + " is not a number; expected number, found boolean\n")}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal("baseline: " + difference)
		}
	}
	intersectionSourceMutant(t, program, kind)
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "true\n" {
			t.Fatalf("counterfactual must execute valid code: %#v", got)
		}
		if disagreement(want, got) == "" {
			t.Fatal("refusal pin did not catch mutant")
		}
	}
	t.Logf("%s mutant caught by native and JavaScript exit-70 pins", kind)
}
func TestV5IntersectionSourceSkip(t *testing.T) {
	t.Parallel()
	v5IntersectionSourceCounterfactual(t, "skip", "v5-root-wrong", "flags")
}
func TestV5IntersectionSourceShape(t *testing.T) {
	t.Parallel()
	v5IntersectionSourceCounterfactual(t, "shape", "v5-root-wrong", "flags")
}
func TestV5IntersectionSourceNested(t *testing.T) {
	t.Parallel()
	v5IntersectionSourceCounterfactual(t, "nested", "v5-root-nested", "autoGenerate.id")
}

func v5IntersectionFixtureCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"good", "absent", "wrong", "nested", "brand-good", "brand-wrong", "root-only-wrong", "v5-root-wrong", "v5-root-nested"} {
		rows = append(rows, counted(t, "stage3/interface-downcasts/lane7/"+name+".a", false, nil, false, false))
	}
	return rows
}
func init() { additionalFixtureCounts = append(additionalFixtureCounts, v5IntersectionFixtureCounts) }

// Not parallel: the update writes measured fixture rows to the shared counts table.
func TestV5IntersectionCounts(t *testing.T) {
	rows := v5IntersectionFixtureCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(text, row+"\n") {
				t.Errorf("unrecorded fixture: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(text, "\n## Predicate direction counts", 2)
	lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
	for _, row := range rows {
		key := strings.Split(row, " | ")[0] + " | "
		found := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i] = row
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, row)
		}
	}
	text = strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		text += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}
