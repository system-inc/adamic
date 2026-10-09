package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testShardDeclarations(t *testing.T) {
	dir := t.TempDir()
	source := `package probe
func TestParent() {
 for _, row := range []string{"one", "two words"} { _ = row }
}
func TestTable() {
 cases := []struct{name string}{{"table one"}, {"table two"}}
 for _, row := range cases { _ = row }
}
var fixtures = []struct{name string}{{%q}}
`
	fixture := filepath.Join(dir, "fixture.a")
	for name, data := range map[string]string{
		"fixture.a":     "fixture",
		"probe_test.go": fmt.Sprintf(source, fixture),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name, data, parent, problem string
		want                        []string
	}{
		{name: "literal", data: `{"TestParent":{"literal":"probe_test.go"}}`, parent: "TestParent", want: []string{"one", "two_words"}},
		{name: "table", data: `{"TestTable":{"literal":"probe_test.go","table":"cases"}}`, parent: "TestTable", want: []string{"table_one", "table_two"}},
		{name: "rows", data: `{"TestParent":{"rows":"probe_test.go","variable":"fixtures"}}`, parent: "TestParent", want: []string{fixture}},
		{name: "glob", data: `{"TestParent":{"glob":"*.a","minimum":1}}`, parent: "TestParent", want: []string{"fixture.a"}},
		{name: "absent parent", data: `{}`, parent: "TestParent"},
		{name: "no file", parent: "TestParent"},
		{name: "wrong key case", data: `{"TestParent":{"Literal":"probe_test.go"}}`, parent: "TestParent", problem: "unknown field"},
		{name: "unknown key", data: `{"TestParent":{"literal":"probe_test.go","typo":true}}`, parent: "TestParent", problem: "unknown field"},
		{name: "two kinds", data: `{"TestParent":{"literal":"probe_test.go","glob":"*.a","minimum":1}}`, parent: "TestParent", problem: "exactly one"},
		{name: "missing file", data: `{"TestParent":{"literal":"missing_test.go"}}`, parent: "TestParent", problem: "missing_test.go"},
		{name: "missing rows file", data: `{"TestParent":{"rows":"missing_test.go","variable":"fixtures"}}`, parent: "TestParent", problem: "missing_test.go"},
		{name: "minimum", data: `{"TestParent":{"glob":"*.a","minimum":2}}`, parent: "TestParent", problem: "matched 1 files, minimum 2"},
		{name: "empty glob", data: `{"TestParent":{"glob":"*.missing","minimum":0}}`, parent: "TestParent", problem: "enumeration is empty"},
		{name: "empty literal", data: `{"TestAbsent":{"literal":"probe_test.go"}}`, parent: "TestAbsent", problem: "cannot enumerate"},
		{name: "empty table", data: `{"TestTable":{"literal":"probe_test.go","table":"absent"}}`, parent: "TestTable", problem: "cannot enumerate"},
		{name: "empty rows", data: `{"TestParent":{"rows":"probe_test.go","variable":"absent"}}`, parent: "TestParent", problem: "no fixture rows"},
		{name: "missing variable", data: `{"TestParent":{"rows":"probe_test.go"}}`, parent: "TestParent", problem: "exactly one"},
		{name: "missing minimum", data: `{"TestParent":{"glob":"*.a"}}`, parent: "TestParent", problem: "exactly one"},
		{name: "table without literal", data: `{"TestParent":{"table":"cases"}}`, parent: "TestParent", problem: "exactly one"},
		{name: "negative minimum", data: `{"TestParent":{"glob":"*.a","minimum":-1}}`, parent: "TestParent", problem: "nonnegative"},
		{name: "bad glob", data: `{"TestParent":{"glob":"[","minimum":0}}`, parent: "TestParent", problem: "syntax error"},
		{name: "wrong type", data: `{"TestParent":{"literal":1}}`, parent: "TestParent", problem: "cannot unmarshal"},
		{name: "bad JSON", data: `{`, parent: "TestParent", problem: "unexpected end"},
		{name: "array", data: `[]`, parent: "TestParent", problem: "cannot unmarshal"},
		{name: "null document", data: `null`, parent: "TestParent", problem: "expected an object"},
		{name: "null declaration", data: `{"TestParent":null}`, parent: "TestParent", problem: "exactly one"},
		{name: "empty declaration", data: `{"TestParent":{}}`, parent: "TestParent", problem: "exactly one"},
		{name: "null field", data: `{"TestParent":{"literal":null}}`, parent: "TestParent", problem: "exactly one"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(dir, "shards.json")
			if tc.data == "" {
				if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(file, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := children(dir, tc.parent)
			if tc.problem != "" {
				if err == nil || !strings.Contains(err.Error(), tc.problem) || !strings.Contains(err.Error(), file) || !strings.Contains(err.Error(), tc.parent) {
					t.Fatalf("got %v, %v; want error containing %q, file and parent", got, err, tc.problem)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, %v; want %v", got, err, tc.want)
			}
		})
	}
}
