package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestWave03SourceTokenStart(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "/* world */\r\nlet x = 1; // comment\nx;"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"target":"ES2022"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := program.Compiler.GetSourceFile(file)
	for position := 0; position <= len(source)+10; position++ {
		wire, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", fmt.Sprintf("source-token-start\n%d", position))
		if err != nil {
			t.Fatal(err)
		}
		got := decodedFields(t, wire)
		want := strconv.Itoa(scanner.SkipTrivia(source, position))
		if len(got) != 3 || got[2] != want {
			t.Fatalf("position %d: %v want %s", position, got, want)
		}
	}
	for _, suffix := range []string{"", "-1", "01", "2147483648", "1\nextra", "not-number"} {
		if _, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", "source-token-start\n"+suffix); err == nil {
			t.Fatalf("accepted suffix %q", suffix)
		}
	}
	var out fields
	if _, err := program.sourceTokenStart(&out, sf.AsNode(), "different-question\n1"); err == nil {
		t.Fatal("accepted wrong question")
	}
	t.Logf("%d exact raw trivia positions including out-of-file provenance", len(source)+11)
}
