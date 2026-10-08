package native

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstantTablesMatchNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "tables.a")
	source, err := os.ReadFile("testdata/constant_tables.a")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, source, 0644); err != nil {
		t.Fatal(err)
	}
	code := C(namedProgram(t, entry, nil))
	if !strings.Contains(code, "static const double adamic_number_table_") || !strings.Contains(code, "static const adamic_value adamic_record_table_") {
		t.Fatal("fixture did not emit constant templates")
	}
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, entry)
	for _, split := range []bool{false, true} {
		for _, sanitize := range []bool{false, true} {
			binary := filepath.Join(directory, "tables")
			if err := Build(code, binary, Options{Split: split, Jobs: 3, Sanitize: sanitize}); err != nil {
				t.Fatal(err)
			}
			if got := runWithInput(t, "", binary); got != want {
				t.Fatalf("split=%t sanitize=%t: native %q Node %q", split, sanitize, got, want)
			}
		}
	}
}

func TestConstantRecordConstructorsStayBounded(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "wide.a")
	var fields, rows []string
	for field := 0; field < 256; field++ {
		fields = append(fields, fmt.Sprintf("field%d:%d", field, field))
	}
	for row := 0; row < 16; row++ {
		rows = append(rows, "{"+strings.Join(fields, ",")+"}")
	}
	source := "import { panic } from 'adamic'; const rows = [" + strings.Join(rows, ",") + "]; console.log((rows[7] ?? panic('missing')).field255.toString());"
	if err := os.WriteFile(entry, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	code := C(namedProgram(t, entry, nil))
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range declarations {
		if declaration.function && strings.HasPrefix(declaration.name, "adamic_initialize_") {
			body := code[declaration.tokens[declaration.body].end:declaration.tokens[len(declaration.tokens)-1].start]
			if strings.Count(body, "\n") > 132 {
				t.Fatalf("unbounded constant record constructor: %s", declaration.name)
			}
		}
	}
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, entry)
	binary := filepath.Join(directory, "wide")
	if err := Build(code, binary, Options{Split: true, Jobs: 3, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q Node %q", got, want)
	}
}
