package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYieldLookaheadAgrees(t *testing.T) {
	t.Parallel()
	cases := []string{
		"yield 0;", "yield 1n;", "yield 'x';", "yield true;", "yield null;", "yield this;", "yield value;",
		"yield(0);", "yield?.(0);", "yield + 0;", "yield * value;", "yield;",
		"yield\n0;", "yield\r\nvalue;", "yield\u2028value;", "yield/*\n*/0;",
		"function* named(){yield 0;yield* values;}",
	}
	directory := t.TempDir()
	var manifest strings.Builder
	for index, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("yield-%d.ts", index))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", goOracle(t), "--manifest", path).output
	port, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{
		{"Node", node(t, port, path, false).output},
		{"native", execute(t, "", buildPort(t, port, true), "--manifest", path).output},
	} {
		if diff := difference(side.output, want); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("%d yield lookahead cases, %d bytes identical to Go on Node and sanitized native", len(cases), len(want))
}
