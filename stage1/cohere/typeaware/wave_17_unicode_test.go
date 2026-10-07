package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

// Not parallel: native compilation shares the wave's toolchain and scratch.
func TestWave17UnicodeUpper(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if root := os.Getenv("ADAMIC_WAVE17_ARTIFACTS"); root != "" {
		directory = filepath.Join(root, "unicode")
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	table := filepath.Join(repository, "stage1/cohere/typeaware/unicode_upper.a")
	source := fmt.Sprintf("import {unicodeUpper} from '%s';\nfor(let code=0;code<=1114111;code++){if(unicodeUpper(code)){console.log(`${code}`);}}\n", table)
	entry := h.write("upper.a", source)
	binary := filepath.Join(directory, "upper")
	h.must("upper-build", exec.Command(stage0, "build", entry, "-o", binary, "--sanitize"))
	var want bytes.Buffer
	for code := rune(0); code <= unicode.MaxRune; code++ {
		if unicode.IsUpper(code) {
			fmt.Fprintln(&want, code)
		}
	}
	got := h.must("upper-run", exec.Command(binary))
	if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.Bytes()) {
		t.Fatal("native Unicode upper classification differs from Go")
	}
	t.Logf("all 1,114,112 code points agree with Go unicode.IsUpper, %d output bytes, sanitizers clean", want.Len())
	data, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "code >= 0x0041"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("uppercase mutant anchor changed")
	}
	mutantTable := h.write("mutant_upper.a", strings.Replace(string(data), anchor, "code > 0x0041", 1))
	entry = h.write("mutant.a", strings.Replace(source, table, mutantTable, 1))
	binary = filepath.Join(directory, "mutant")
	h.must("mutant-build", exec.Command(stage0, "build", entry, "-o", binary, "--sanitize"))
	got = h.must("mutant-run", exec.Command(binary))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.Bytes()) {
		t.Fatal("uppercase boundary mutant survived or sanitizer caught it")
	}
	t.Logf("uppercase boundary mutant: exit 0, empty stderr, direct Go byte oracle catches byte %d", firstDifference(got.stdout, want.Bytes()))
}
