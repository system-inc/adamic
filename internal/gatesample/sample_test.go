package gatesample

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Not parallel: this test changes the process-wide sampling environment.
func TestSelection(test *testing.T) {
	root := test.TempDir()
	paths := []string{"pkg/testdata/b.md", "other/b.md", "pkg/fixtures/a.md", "other/a.md", "pkg/testdata/a.md"}
	if createDirectoryError := os.MkdirAll(filepath.Join(root, "pkg"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(root, "pkg", "source_test.go"), nil, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	test.Setenv("ADAMIC_GATE_CHANGED", "")
	first, selectError := Select(root, paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if want := []string{"other/a.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(first.Paths, want) {
		test.Fatalf("stride: %v want %v", first.Paths, want)
	}
	second, selectError := Select(root, paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if !reflect.DeepEqual(first, second) {
		test.Fatal("selection is not deterministic")
	}
	reversed := append([]string(nil), paths...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	second, selectError = Select(root, reversed, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if !reflect.DeepEqual(first, second) {
		test.Fatal("selection depends on input order")
	}
	if got := first.Log("TestCorpus"); got != "gate-sample: TestCorpus 2/5 files, stride 3, offset 0" {
		test.Fatal(got)
	}
	control, selectError := Select(root, paths, 3, "other/b.md")
	if selectError != nil {
		test.Fatal(selectError)
	}
	if want := []string{"other/a.md", "other/b.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(control.Paths, want) {
		test.Fatalf("fixed control dropped: %v", control.Paths)
	}
	seen := map[string]bool{}
	for offset := 0; offset < 3; offset++ {
		test.Setenv("ADAMIC_GATE_SAMPLE", fmt.Sprintf("%08x%s", offset, strings.Repeat("0", 32)))
		rotated, selectError := Select(root, paths, 3)
		if selectError != nil {
			test.Fatal(selectError)
		}
		if rotated.Offset != offset {
			test.Fatalf("offset %d want %d", rotated.Offset, offset)
		}
		for _, name := range rotated.Paths {
			seen[name] = true
		}
	}
	if len(seen) != len(paths) {
		test.Fatalf("rotation covers %d/%d", len(seen), len(paths))
	}
	test.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	list := filepath.Join(root, "changed.txt")
	if writeError := os.WriteFile(list, []byte("other/b.md\npkg/new.ts\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_GATE_CHANGED", list)
	selected, selectError := Select(root, paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if len(selected.Paths) != len(paths) {
		test.Fatalf("changed file and package fixtures must all be included: %v", selected.Paths)
	}
	if createDirectoryError := os.MkdirAll(filepath.Join(root, "pkg", "testdata"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(root, "pkg", "testdata", "adapter.go"), nil, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(list, []byte("pkg/testdata/adapter.go\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	selected, selectError = Select(root, paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if want := []string{
		"other/a.md",
		"pkg/fixtures/a.md",
		"pkg/testdata/a.md",
		"pkg/testdata/b.md",
	}; !reflect.DeepEqual(selected.Paths, want) {
		test.Fatalf("fixture adapter lost owning package: %v", selected.Paths)
	}
	// Changed paths outside the stride must also work without any owning package.
	if writeError := os.WriteFile(list, []byte("other/b.md\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	selected, selectError = Select(root, paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if want := []string{"other/a.md", "other/b.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(selected.Paths, want) {
		test.Fatalf("changed file: %v", selected.Paths)
	}
	for _, sha := range []string{
		"",
		"abcd",
		strings.Repeat("g", 40),
		strings.Repeat("0", 39),
		strings.Repeat("0", 41),
	} {
		test.Setenv("ADAMIC_GATE_SAMPLE", sha)
		if _, selectError := Select(root, paths, 3); selectError == nil {
			test.Fatalf("accepted malformed switch %q", sha)
		}
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestUnsetIsWholeCorpus(test *testing.T) {
	old, set := os.LookupEnv("ADAMIC_GATE_SAMPLE")
	if unsetenvError := os.Unsetenv("ADAMIC_GATE_SAMPLE"); unsetenvError != nil {
		test.Fatal(unsetenvError)
	}
	test.Cleanup(func() {
		if set {
			os.Setenv("ADAMIC_GATE_SAMPLE", old)
		} else {
			os.Unsetenv("ADAMIC_GATE_SAMPLE")
		}
	})
	test.Setenv("ADAMIC_GATE_CHANGED", "/does/not/exist")
	paths := []string{"b.md", "a.md"}
	selected, selectError := Select(test.TempDir(), paths, 3)
	if selectError != nil {
		test.Fatal(selectError)
	}
	if selected.Sample || !reflect.DeepEqual(selected.Paths, paths) {
		test.Fatalf("unset changed full corpus: %+v", selected)
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestBadChangedList(test *testing.T) {
	root := test.TempDir()
	test.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	test.Setenv("ADAMIC_GATE_CHANGED", filepath.Join(root, "missing"))
	if _, selectError := Select(root, []string{"a.md"}, 3); selectError == nil {
		test.Fatal("missing changed list silently ignored")
	}
	list := filepath.Join(root, "changed.txt")
	test.Setenv("ADAMIC_GATE_CHANGED", list)
	for _, path := range []string{"../outside.md", filepath.Join(root, "absolute.md")} {
		if writeError := os.WriteFile(list, []byte(path+"\n"), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		if _, selectError := Select(root, []string{"a.md"}, 3); selectError == nil {
			test.Fatalf("accepted non-repository path %q", path)
		}
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestHexOffset(test *testing.T) {
	test.Setenv("ADAMIC_GATE_CHANGED", "")
	for _, prefix := range []string{"0000000a", "0000000A"} {
		test.Setenv("ADAMIC_GATE_SAMPLE", prefix+strings.Repeat("0", 32))
		selected, selectError := Select(test.TempDir(), []string{"a", "b", "c"}, 3)
		if selectError != nil {
			test.Fatal(selectError)
		}
		if selected.Offset != 1 || len(selected.Paths) != 1 || selected.Paths[0] != "b" {
			test.Fatalf("offset must parse hexadecimal: %+v", selected)
		}
	}
}

func TestGeneratedSelection(test *testing.T) {
	test.Parallel()
	s := Selection{Sample: true, Stride: 3, Offset: 1}.Generated(8, 0, 7)
	want := []string{GeneratedKey(0), GeneratedKey(1), GeneratedKey(4), GeneratedKey(7)}
	if !reflect.DeepEqual(s.Paths, want) || s.Total != 8 {
		test.Fatalf("indexed selection: %#v", s)
	}
	full := Selection{Stride: 3}.Generated(8)
	if len(full.Paths) != 8 {
		test.Fatal("unset dropped generated inputs")
	}
	rotated := Selection{Sample: true, Stride: 3, Offset: 2}.Generated(8)
	if reflect.DeepEqual(s.Paths, rotated.Paths) {
		test.Fatal("offset did not rotate")
	}
}
