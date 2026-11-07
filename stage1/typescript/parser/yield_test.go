package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYieldLookaheadAgrees(test *testing.T) {
	test.Parallel()
	cases := []string{
		"yield 0;", "yield 1n;", "yield 'x';", "yield true;", "yield null;", "yield this;", "yield value;",
		"yield(0);", "yield?.(0);", "yield + 0;", "yield * value;", "yield;",
		"yield\n0;", "yield\r\nvalue;", "yield\u2028value;", "yield/*\n*/0;",
		"function* named(){yield 0;yield* values;}",
	}
	directory := test.TempDir()
	var manifest strings.Builder
	for index, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("yield-%d.ts", index))
		if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if writeError := os.WriteFile(path, []byte(manifest.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	want := execute(test, "", goOracle(test), "--manifest", path).output
	port, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{
		{"Node", node(test, port, path, false).output},
		{"native", execute(test, "", buildPort(test, port, true), "--manifest", path).output},
	} {
		if diff := difference(side.output, want); diff != "" {
			test.Fatalf("%s: %s", side.name, diff)
		}
	}
	test.Logf("%d yield lookahead cases, %d bytes identical to Go on Node and sanitized native", len(cases), len(want))
}
