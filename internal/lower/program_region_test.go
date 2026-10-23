package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"testing"
)

func TestProgramRegionScoutParseSelectionEmpty(t *testing.T) {
	t.Parallel()
	source, err := load.Load([]string{"../oracle/testdata/program_region/parse.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := LowerWithOptions(context.Background(), source, Options{ProgramRegion: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.ProgramTypes) != 0 {
		t.Fatalf("acyclic parse.a selected %d identities", len(program.ProgramTypes))
	}
	for _, local := range program.Locals {
		if local.ProgramRegion {
			t.Fatalf("parser cell %s selected", local.Name)
		}
	}
	for _, fn := range program.Functions {
		if fn.ProgramRegion {
			t.Fatalf("parser function %s selected", fn.Name)
		}
	}
	for _, class := range program.Classes {
		if class.ProgramRegion {
			t.Fatalf("parser class %s selected", class.Name)
		}
	}
}
