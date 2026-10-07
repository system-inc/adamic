package checker

import (
	"path/filepath"
	"strconv"
	"testing"
)

func TestModuleRecordFacts(t *testing.T) {
	t.Parallel()
	program, file := wave20NextProgram(t)
	source := program.Compiler.GetSourceFile(file)
	wire, err := program.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", "module-records")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		t.Fatal(err)
	}
	if count != len(program.Compiler.GetSourceFiles()) {
		t.Fatal("missing compiler sources")
	}
	at := 3
	found := false
	for i := 0; i < count; i++ {
		path := fields[at]
		text := fields[at+2]
		records, err := strconv.Atoi(fields[at+3])
		if err != nil {
			t.Fatal(err)
		}
		at += 4
		if path == file {
			found = true
			if text != source.Text() || records != 6 {
				t.Fatalf("source or import record count %d", records)
			}
			kinds := []string{"import", "import", "export", "call", "call", "call"}
			for r := 0; r < records; r++ {
				base := at + r*4
				if fields[base] != kinds[r] {
					t.Fatal("record kind", fields[base])
				}
				if r == 1 {
					if fields[base+1] != "1" {
						t.Fatal("type-only import lost")
					}
				}
				if r == 2 || r == 4 || r == 5 {
					if fields[base+3] != "" {
						t.Fatal("unresolved record gained a target")
					}
				} else if fields[base+3] != filepath.Join(filepath.Dir(file), "helper.ts") {
					t.Fatal("wrong import resolution", r, fields[base+3])
				}
			}
		}
		at += records * 4
	}
	if !found || at != len(fields) {
		t.Fatal("incomplete module records")
	}
	if _, err := program.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", "module-records\nextra"); err == nil {
		t.Fatal("suffix accepted")
	}
}
