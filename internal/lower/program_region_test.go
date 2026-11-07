package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"testing"
)

func TestProgramRegionScoutParseSelectionEmpty(test *testing.T) {
	test.Parallel()
	source, err := load.Load([]string{"../oracle/testdata/program_region/parse.a"})
	if err != nil {
		test.Fatal(err)
	}
	program, err := LowerWithOptions(context.Background(), source, Options{ProgramRegion: true})
	if err != nil {
		test.Fatal(err)
	}
	if len(program.ProgramTypes) != 0 {
		test.Fatalf("acyclic parse.a selected %d identities", len(program.ProgramTypes))
	}
	for _, local := range program.Locals {
		if local.ProgramRegion {
			test.Fatalf("parser cell %s selected", local.Name)
		}
	}
	for _, fn := range program.Functions {
		if fn.ProgramRegion {
			test.Fatalf("parser function %s selected", fn.Name)
		}
	}
	for _, class := range program.Classes {
		if class.ProgramRegion {
			test.Fatalf("parser class %s selected", class.Name)
		}
	}
}
