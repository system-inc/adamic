package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func testLeavesFormatAndEscaping(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"shards.json": `{"TestZ":{"literal":"probe_test.go"},"TestA":{"literal":"probe_test.go"}}`,
		"probe_test.go": `package probe
func TestZ() { for _, name := range []string{"z", "a+b words", "path/a.b"} { _ = name } }
func TestA() { for _, name := range []string{"first"} { _ = name } }
`,
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var got bytes.Buffer
	if err := packageLeaves(&got, "example/probe", dir); err != nil {
		t.Fatal(err)
	}
	want := "example/probe TestA first ^TestA$/^first$\n" +
		"example/probe TestZ a+b_words ^TestZ$/^a\\+b_words$\n" +
		"example/probe TestZ path/a.b ^TestZ$/^path$/^a\\.b$\n" +
		"example/probe TestZ z ^TestZ$/^z$\n"
	if got.String() != want {
		t.Fatalf("got %q, want %q", got.String(), want)
	}
	var again bytes.Buffer
	if err := packageLeaves(&again, "example/probe", dir); err != nil {
		t.Fatal(err)
	}
	if got.String() != again.String() {
		t.Fatal("leaves output changed")
	}
}
