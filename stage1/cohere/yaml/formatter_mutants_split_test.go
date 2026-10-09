package yaml

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestFormatterMutants(t *testing.T) {
	cases, _, _ := formatCases(t)
	expected := goFormat(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, file, from, to string }{
		{"root final newline lost", "printer.ts", "const hard = !(", "const hard = false && !("},
		{"colon separation lost", "printer.ts", "this.layout.text(': '), printedValue", "this.layout.text(':'), printedValue"},
		{"flow trailing comma lost", "printer.ts", "this.layout.ifBreak(this.layout.text(','), this.empty, -1)", "this.layout.ifBreak(this.layout.text(''), this.empty, -1)"},
		{"batch backslash scan misses escapes", "main.ts", "text.charCodeAt(index) !== 92", "text.charCodeAt(index) !== 13"},
		{"emoji first unit range loses endpoint", "width.ts", "code <= last", "code < last"},
		{"emoji surrogate slot shifted", "width.ts", "code - 0x100000 + 0xd800", "code - 0x100000 + 0xd900"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			entries, err := filepath.Glob("*.ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range entries {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == mutant.file {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "main.ts")
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(lowered), binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			nativeOut := run(t, "", nil, binary, "--cases", cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases)
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if bytes.Equal(side.out, expected) {
					t.Fatalf("%s missed mutant", side.name)
				}
				t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
			}
		})
	}
}
