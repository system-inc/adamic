package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhysicalInventoryBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		source string
		lines  int
	}{{"", 0}, {"package p", 1}, {"a\nb\n", 2}, {"a\r\nb", 2}} {
		if got := physicalLines([]byte(test.source)); got != test.lines {
			t.Fatalf("lines=%d want %d", got, test.lines)
		}
	}
	for name, want := range map[string]bool{"internal/p/generated.go": true, "internal/p/windows.go": true, "internal/p/a_test.go": false, "internal/p/testdata/helper.go": false, "internal/.hidden/x.go": false} {
		if production(name) != want {
			t.Fatalf("production(%q)", name)
		}
	}
}
func TestDeclarationCreditIsBoundedAndFailsClosed(t *testing.T) {
	t.Parallel()
	source := []byte("package p\n\n// selected\nfunc chosen() {\n}\n\nfunc untouched() {}\n")
	marks, err := credit(source, []string{"chosen"})
	if err != nil {
		t.Fatal(err)
	}
	if len(marks) != 3 || !marks[3] || !marks[5] || marks[7] {
		t.Fatalf("marks=%v", marks)
	}
	if _, err := credit(source, []string{"renamed"}); err == nil {
		t.Fatal("missing declaration silently credited")
	}
}
func TestUnionDoesNotCountSharedLinesTwice(t *testing.T) {
	t.Parallel()
	reports := []*report{
		{Cohere: "pin", marks: map[string]map[int]bool{"internal/p/a.go": {1: true, 2: true}}, partial: map[string]bool{}, Packages: []packageRow{{Package: "internal/p", GoLines: 4, Evidence: []string{"gap"}}}},
		{Cohere: "pin", marks: map[string]map[int]bool{"internal/p/a.go": {2: true, 3: true}}, partial: map[string]bool{}, Packages: []packageRow{{Package: "internal/p", GoLines: 4, Evidence: []string{"gap"}}}},
	}
	combined, err := union(reports)
	if err != nil {
		t.Fatal(err)
	}
	if combined.Ported != 3 || combined.Total != 4 || combined.Percent != 75 || combined.Complete != 0 {
		t.Fatalf("union=%#v", combined)
	}
	reports[1].Cohere = "different"
	if _, err := union(reports); err == nil {
		t.Fatal("union accepted incompatible pins")
	}
}
func TestGitSnapshotAndEvidenceBoundary(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	if err := os.MkdirAll(cohere, 0755); err != nil {
		t.Fatal(err)
	}
	command := func(directory string, args ...string) string {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", directory}, args...)...)
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, text string) {
		t.Helper()
		name = filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	initialize := func(directory string) {
		command(directory, "init", "-q")
		command(directory, "config", "user.name", "fixture")
		command(directory, "config", "user.email", "fixture@example.invalid")
	}
	initialize(cohere)
	write("cohere/internal/gitignore/gitignore.go", "package gitignore\n// body\n")
	write("cohere/internal/gitignore/glob.go", "package gitignore\n")
	write("cohere/internal/format/yaml/yaml.go", "package yaml\n")
	write("cohere/internal/gitignore/x_test.go", "package gitignore\n// excluded\n")
	write("cohere/internal/gitignore/testdata/helper.go", "package main\n// excluded\n")
	command(cohere, "add", ".")
	command(cohere, "commit", "-qm", "fixture")
	initialize(root)
	write("stage1/cohere/yaml/GAPS.md", "# Audit\n\nNo Adamic YAML implementation.\n")
	write("stage1/cohere/gitignore/GAPS.md", "# Gaps\n\nRecorded native and Go parity.\n")
	write("stage1/cohere/gitignore/main.ts", "console.log('a');\n")
	write("stage1/cohere/gitignore/gitignore_test.go", "package gitignore\n")
	command(root, "add", ".")
	command(root, "commit", "-qm", "ported snapshot")
	first := command(root, "rev-parse", "HEAD")
	write("stage1/cohere/gitignore/main.ts", "dirty source does not affect the Git snapshot")
	report, err := measure(root, first, map[string]*inventory{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 4 || report.Ported != 3 || report.Complete != 1 || report.NotStarted != 1 {
		t.Fatalf("snapshot=%#v", report)
	}
	command(root, "rm", "--cached", "stage1/cohere/gitignore/gitignore_test.go")
	command(root, "commit", "-qm", "without parity test")
	report, err = measure(root, "HEAD", map[string]*inventory{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Ported != 0 || report.Complete != 0 || report.Partial != 1 {
		t.Fatalf("unheld port got credit: %#v", report)
	}
	write("stage1/cohere/yaml/GAPS.md", "# Audit\n\nNo Adamic YAML implementation.\n")
	write("stage1/cohere/new_slice/GAPS.md", "# Unknown\n")
	command(root, "add", "stage1/cohere/new_slice")
	command(root, "commit", "-qm", "unknown slice")
	if _, err := measure(root, "HEAD", map[string]*inventory{}); err == nil {
		t.Fatal("unknown slice silently classified")
	}
	command(root, "update-ref", "refs/remotes/origin/main", first)
	command(root, "update-ref", "refs/remotes/origin/codex/stage1-new", "HEAD")
	command(root, "update-ref", "refs/remotes/origin/codex/stage1-old", first)
	command(root, "update-ref", "refs/remotes/origin/codex/typescript-scanner", "HEAD")
	command(root, "update-ref", "refs/remotes/origin/codex/unrelated", "HEAD")
	refs, integrated, err := pending(root)
	if err != nil || strings.Join(refs, ",") != "origin/codex/stage1-new,origin/codex/typescript-scanner" || strings.Join(integrated, ",") != "origin/codex/stage1-old" {
		t.Fatalf("pending=%v integrated=%v err=%v", refs, integrated, err)
	}

}

func TestFirstParagraphJoinsTwoLinesWithSpace(t *testing.T) {
	t.Parallel()
	text := "# Gaps\n\n  First line of evidence.  \nSecond line of evidence.\n\nAnother paragraph.\n"
	want := "First line of evidence. Second line of evidence."
	if got := firstParagraph(text); got != want {
		t.Fatalf("firstParagraph = %q, want %q", got, want)
	}
}
