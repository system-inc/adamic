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
func TestSelection(t *testing.T) {
	root := t.TempDir()
	paths := []string{"pkg/testdata/b.md", "other/b.md", "pkg/fixtures/a.md", "other/a.md", "pkg/testdata/a.md"}
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "source_test.go"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	t.Setenv("ADAMIC_GATE_CHANGED", "")
	first, err := Select(root, paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"other/a.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(first.Paths, want) {
		t.Fatalf("stride: %v want %v", first.Paths, want)
	}
	second, err := Select(root, paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("selection is not deterministic")
	}
	reversed := append([]string(nil), paths...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	second, err = Select(root, reversed, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("selection depends on input order")
	}
	if got := first.Log("TestCorpus"); got != "gate-sample: TestCorpus 2/5 files, stride 3, offset 0" {
		t.Fatal(got)
	}
	control, err := Select(root, paths, 3, "other/b.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"other/a.md", "other/b.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(control.Paths, want) {
		t.Fatalf("fixed control dropped: %v", control.Paths)
	}
	seen := map[string]bool{}
	for offset := 0; offset < 3; offset++ {
		t.Setenv("ADAMIC_GATE_SAMPLE", fmt.Sprintf("%08x%s", offset, strings.Repeat("0", 32)))
		rotated, err := Select(root, paths, 3)
		if err != nil {
			t.Fatal(err)
		}
		if rotated.Offset != offset {
			t.Fatalf("offset %d want %d", rotated.Offset, offset)
		}
		for _, name := range rotated.Paths {
			seen[name] = true
		}
	}
	if len(seen) != len(paths) {
		t.Fatalf("rotation covers %d/%d", len(seen), len(paths))
	}
	t.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	list := filepath.Join(root, "changed.txt")
	if err := os.WriteFile(list, []byte("other/b.md\npkg/new.ts\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_GATE_CHANGED", list)
	selected, err := Select(root, paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected.Paths) != len(paths) {
		t.Fatalf("changed file and package fixtures must all be included: %v", selected.Paths)
	}
	if err := os.MkdirAll(filepath.Join(root, "pkg", "testdata"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "testdata", "adapter.go"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(list, []byte("pkg/testdata/adapter.go\n"), 0644); err != nil {
		t.Fatal(err)
	}
	selected, err = Select(root, paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"other/a.md", "pkg/fixtures/a.md", "pkg/testdata/a.md", "pkg/testdata/b.md"}; !reflect.DeepEqual(selected.Paths, want) {
		t.Fatalf("fixture adapter lost owning package: %v", selected.Paths)
	}
	// Changed paths outside the stride must also work without any owning package.
	if err := os.WriteFile(list, []byte("other/b.md\n"), 0644); err != nil {
		t.Fatal(err)
	}
	selected, err = Select(root, paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"other/a.md", "other/b.md", "pkg/testdata/a.md"}; !reflect.DeepEqual(selected.Paths, want) {
		t.Fatalf("changed file: %v", selected.Paths)
	}
	for _, sha := range []string{"", "abcd", strings.Repeat("g", 40), strings.Repeat("0", 39), strings.Repeat("0", 41)} {
		t.Setenv("ADAMIC_GATE_SAMPLE", sha)
		if _, err := Select(root, paths, 3); err == nil {
			t.Fatalf("accepted malformed switch %q", sha)
		}
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestUnsetIsWholeCorpus(t *testing.T) {
	old, set := os.LookupEnv("ADAMIC_GATE_SAMPLE")
	if err := os.Unsetenv("ADAMIC_GATE_SAMPLE"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if set {
			os.Setenv("ADAMIC_GATE_SAMPLE", old)
		} else {
			os.Unsetenv("ADAMIC_GATE_SAMPLE")
		}
	})
	t.Setenv("ADAMIC_GATE_CHANGED", "/does/not/exist")
	paths := []string{"b.md", "a.md"}
	selected, err := Select(t.TempDir(), paths, 3)
	if err != nil {
		t.Fatal(err)
	}
	if selected.Sample || !reflect.DeepEqual(selected.Paths, paths) {
		t.Fatalf("unset changed full corpus: %+v", selected)
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestBadChangedList(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	t.Setenv("ADAMIC_GATE_CHANGED", filepath.Join(root, "missing"))
	if _, err := Select(root, []string{"a.md"}, 3); err == nil {
		t.Fatal("missing changed list silently ignored")
	}
	list := filepath.Join(root, "changed.txt")
	t.Setenv("ADAMIC_GATE_CHANGED", list)
	for _, path := range []string{"../outside.md", filepath.Join(root, "absolute.md")} {
		if err := os.WriteFile(list, []byte(path+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Select(root, []string{"a.md"}, 3); err == nil {
			t.Fatalf("accepted non-repository path %q", path)
		}
	}
}

// Not parallel: this test changes the process-wide sampling environment.
func TestHexOffset(t *testing.T) {
	t.Setenv("ADAMIC_GATE_CHANGED", "")
	for _, prefix := range []string{"0000000a", "0000000A"} {
		t.Setenv("ADAMIC_GATE_SAMPLE", prefix+strings.Repeat("0", 32))
		selected, err := Select(t.TempDir(), []string{"a", "b", "c"}, 3)
		if err != nil {
			t.Fatal(err)
		}
		if selected.Offset != 1 || len(selected.Paths) != 1 || selected.Paths[0] != "b" {
			t.Fatalf("offset must parse hexadecimal: %+v", selected)
		}
	}
}

func TestGeneratedSelection(t *testing.T) {
	s := Selection{Sample: true, Stride: 3, Offset: 1}.Generated(8, 0, 7)
	want := []string{GeneratedKey(0), GeneratedKey(1), GeneratedKey(4), GeneratedKey(7)}
	if !reflect.DeepEqual(s.Paths, want) || s.Total != 8 {
		t.Fatalf("indexed selection: %#v", s)
	}
	full := Selection{Stride: 3}.Generated(8)
	if len(full.Paths) != 8 {
		t.Fatal("unset dropped generated inputs")
	}
	rotated := Selection{Sample: true, Stride: 3, Offset: 2}.Generated(8)
	if reflect.DeepEqual(s.Paths, rotated.Paths) {
		t.Fatal("offset did not rotate")
	}
}
