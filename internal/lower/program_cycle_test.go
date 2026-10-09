package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Check the ruled selector's actual types before asserting ownership or the
// canonical-cache boundary. A mode-aware expectation must not hide a nonmember.
func lowerCycleSource(t *testing.T, source string, members ...string) (*ir.Program, error) {
	t.Helper()
	if os.Getenv("ADAMIC_PROGRAM_REGION") != "1" {
		return lowerSource(t, source)
	}
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	defer release()
	discovery, err := lowerChecked(loaded, check, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	plan := discovery.programPlan
	for _, member := range members {
		kind, name, _ := strings.Cut(member, ":")
		found := false
		switch kind {
		case "type", "counted", "cell":
			for local, proven := range discovery.localTypes {
				if proven == nil || discovery.result.Locals[local].Name != name {
					continue
				}
				selected := plan.types[int(proven.Id())]
				if kind == "cell" {
					selected = plan.cells[local]
				}
				label := check.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
				t.Logf("selector %s %s: %s selected=%t", kind, name, label, selected)
				if kind == "counted" {
					selected = !selected
				}
				found = found || selected
			}
		case "closure":
			for _, record := range discovery.closureRecords {
				if discovery.result.Functions[record.function].Name != name {
					continue
				}
				label := check.TypeToStringEx(record.proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
				selected := plan.functions[record.function]
				t.Logf("selector closure %s: %s type=%t allocation=%t", name, label, plan.types[int(record.proven.Id())], selected)
				found = found || selected
			}
		}
		if !found {
			t.Fatalf("ruled selector did not satisfy %s", member)
		}
	}
	built, err := lowerChecked(loaded, check, entry, plan, false)
	if err != nil {
		return nil, err
	}
	return built.result, nil
}

func cycleOwnershipMissing(program *ir.Program) bool {
	if program.ProgramRegion {
		return len(program.ProgramTypes) == 0
	}
	return len(program.GraphTypes) == 0
}

func requireProgramCanonicalRefusal(t *testing.T, err error) {
	t.Helper()
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "Program canonical closure before canonical-cache adoption" {
		t.Fatalf("want selected canonical closure boundary, got %v", err)
	}
}
