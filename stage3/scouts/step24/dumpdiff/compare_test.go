package dumpdiff

import (
	"errors"
	"io"
	"strings"
	"testing"
)

const dump = "file\tx.ts\nSourceFile 0 4 0\t\nIdentifier 0 1 0\tx\ndiagnostics 1\ndiagnostic 1 1110 1 1\tType expected.\njsDocDiagnostics 1\njsDocDiagnostic 1 1110 1 1\tType expected.\n"

func TestExactRecords(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, actual, path string
		line               int
	}{
		{"equal", dump, "", 0},
		{"nodeEnd", strings.Replace(dump, "Identifier 0 1", "Identifier 0 2", 1), "x.ts/preorder/1", 3},
		{"flags", strings.Replace(dump, "Identifier 0 1 0", "Identifier 0 1 4", 1), "x.ts/preorder/1", 3},
		{"text", strings.Replace(dump, "\tx\n", "\ty\n", 1), "x.ts/preorder/1", 3},
		{"header", strings.Replace(dump, "x.ts", "y.ts", 1), "x.ts/header", 1},
		{"diagnostic", strings.Replace(dump, "diagnostic 1 1110", "diagnostic 1 1111", 1), "x.ts/parseDiagnostics/0", 5},
		{"jsdoc", strings.Replace(dump, "jsDocDiagnostic 1 1110", "jsDocDiagnostic 1 1111", 1), "x.ts/jsDocDiagnostics/0", 7},
		{"newline", strings.TrimSuffix(dump, "\n"), "x.ts/jsDocDiagnostics/0", 7},
		{"truncated", "", "x.ts/header", 1},
		{"extra", dump + "file\ty.ts\n", "EOF", 8},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, err := Compare(strings.NewReader(dump), strings.NewReader(c.actual))
			if err != nil {
				t.Fatal(err)
			}
			if c.line == 0 {
				if d != nil {
					t.Fatal(d)
				}
				return
			}
			if d == nil || d.Line != c.line || d.ReferencePath != c.path {
				t.Fatalf("got %+v, want line %d path %s", d, c.line, c.path)
			}
		})
	}
}
func TestLongRecordAndReset(t *testing.T) {
	t.Parallel()
	input := dump + "file\ty.ts\nSourceFile 0 1 0\t" + strings.Repeat("x", 100000) + "\n"
	changed := input[:len(input)-2] + "y\n"
	d, err := Compare(strings.NewReader(input), strings.NewReader(changed))
	if err != nil || d == nil {
		t.Fatalf("missing difference: %v", err)
	}
	if d.ReferencePath != "y.ts/preorder/0" {
		t.Fatalf("wrong path: %s", d.ReferencePath)
	}
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, errors.New("broken") }
func TestReadError(t *testing.T) {
	t.Parallel()
	for _, readers := range [][2]io.Reader{{badReader{}, strings.NewReader("")}, {strings.NewReader(""), badReader{}}} {
		if _, err := Compare(readers[0], readers[1]); err == nil {
			t.Fatal("lost read error")
		}
	}
}
