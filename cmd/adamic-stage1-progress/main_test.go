package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

func TestPhysicalInventoryBoundaries(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		source string
		lines  int
	}{{"", 0}, {"package p", 1}, {"a\nb\n", 2}, {"a\r\nb", 2}} {
		if got := physicalLines([]byte(testCase.source)); got != testCase.lines {
			test.Fatalf("lines=%d want %d", got, testCase.lines)
		}
	}
	for name, want := range map[string]bool{
		"internal/p/generated.go":       true,
		"internal/p/windows.go":         true,
		"internal/p/a_test.go":          false,
		"internal/p/testdata/helper.go": false,
		"internal/.hidden/x.go":         false,
	} {
		if production(name) != want {
			test.Fatalf("production(%q)", name)
		}
	}
}
func TestDeclarationCreditIsBoundedAndFailsClosed(test *testing.T) {
	test.Parallel()
	source := []byte("package p\n\n// selected\nfunc chosen() {\n}\n\nfunc untouched() {}\n")
	marks, creditError := credit(source, []string{"chosen"})
	if creditError != nil {
		test.Fatal(creditError)
	}
	if len(marks) != 3 || !marks[3] || !marks[5] || marks[7] {
		test.Fatalf("marks=%v", marks)
	}
	if _, creditError := credit(source, []string{"renamed"}); creditError == nil {
		test.Fatal("missing declaration silently credited")
	}
}
func TestUnionDoesNotCountSharedLinesTwice(test *testing.T) {
	test.Parallel()
	reports := []*report{
		{
			Cohere:   "pin",
			marks:    map[string]map[int]bool{"internal/p/a.go": {1: true, 2: true}},
			partial:  map[string]bool{},
			Packages: []packageRow{{Package: "internal/p", GoLines: 4, Evidence: []string{"gap"}}},
		},
		{
			Cohere:   "pin",
			marks:    map[string]map[int]bool{"internal/p/a.go": {2: true, 3: true}},
			partial:  map[string]bool{},
			Packages: []packageRow{{Package: "internal/p", GoLines: 4, Evidence: []string{"gap"}}},
		},
	}
	combined, unionError := union(reports)
	if unionError != nil {
		test.Fatal(unionError)
	}
	if combined.Ported != 3 || combined.Total != 4 || combined.Percent != 75 || combined.Complete != 0 {
		test.Fatalf("union=%#v", combined)
	}
	reports[1].Cohere = "different"
	if _, unionError := union(reports); unionError == nil {
		test.Fatal("union accepted incompatible pins")
	}
}
func TestGitSnapshotAndEvidenceBoundary(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	cohere := filepath.Join(root, "cohere")
	if createDirectoryError := os.MkdirAll(cohere, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	command := func(directory string, args ...string) string {
		test.Helper()
		c := toolchain.GitCommand(append([]string{"-C", directory}, args...)...)
		out, combinedOutputError := c.CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("git %v: %v %s", args, combinedOutputError, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, text string) {
		test.Helper()
		name = filepath.Join(root, name)
		if createDirectoryError := os.MkdirAll(filepath.Dir(name), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(name, []byte(text), 0644); writeError != nil {
			test.Fatal(writeError)
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
	report, measureError := measure(root, first, map[string]*inventory{})
	if measureError != nil {
		test.Fatal(measureError)
	}
	if report.Total != 4 || report.Ported != 3 || report.Complete != 1 || report.NotStarted != 1 {
		test.Fatalf("snapshot=%#v", report)
	}
	command(root, "rm", "--cached", "stage1/cohere/gitignore/gitignore_test.go")
	command(root, "commit", "-qm", "without parity test")
	report, measureError = measure(root, "HEAD", map[string]*inventory{})
	if measureError != nil {
		test.Fatal(measureError)
	}
	if report.Ported != 0 || report.Complete != 0 || report.Partial != 1 {
		test.Fatalf("unheld port got credit: %#v", report)
	}
	write("stage1/cohere/yaml/GAPS.md", "# Audit\n\nNo Adamic YAML implementation.\n")
	write("stage1/cohere/new_slice/GAPS.md", "# Unknown\n")
	command(root, "add", "stage1/cohere/new_slice")
	command(root, "commit", "-qm", "unknown slice")
	if _, measureError := measure(root, "HEAD", map[string]*inventory{}); measureError == nil {
		test.Fatal("unknown slice silently classified")
	}
	command(root, "update-ref", "refs/remotes/origin/main", first)
	command(root, "update-ref", "refs/remotes/origin/codex/stage1-new", "HEAD")
	command(root, "update-ref", "refs/remotes/origin/codex/stage1-old", first)
	command(root, "update-ref", "refs/remotes/origin/codex/typescript-scanner", "HEAD")
	command(root, "update-ref", "refs/remotes/origin/codex/unrelated", "HEAD")
	refs, integrated, pendingError := pending(root)
	if pendingError != nil ||
		strings.Join(refs, ",") != "origin/codex/stage1-new,origin/codex/typescript-scanner" ||
		strings.Join(integrated, ",") != "origin/codex/stage1-old" {
		test.Fatalf("pending=%v integrated=%v err=%v", refs, integrated, pendingError)
	}

}

func TestFirstParagraphJoinsTwoLinesWithSpace(test *testing.T) {
	test.Parallel()
	text := "# Gaps\n\n  First line of evidence.  \nSecond line of evidence.\n\nAnother paragraph.\n"
	want := "First line of evidence. Second line of evidence."
	if got := firstParagraph(text); got != want {
		test.Fatalf("firstParagraph = %q, want %q", got, want)
	}
}
