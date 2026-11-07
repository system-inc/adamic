package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"testing"
)

func TestProgramRegionScoutParseSelectionEmpty(test *testing.T) {
	test.Parallel()
	source, loadError := load.Load([]string{"../oracle/testdata/program_region/parse.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerWithOptionsError := LowerWithOptions(context.Background(), source, Options{ProgramRegion: true})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
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
