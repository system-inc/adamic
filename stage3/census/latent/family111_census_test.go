//go:build census_overlay

package latent_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestLatentNamedSortComparator(t *testing.T) {
	t.Parallel()
	const source = `function compare(left: string, right: string): number {
    if (left < right) return -1;
    if (left > right) return 1;
    return 0;
}

function sorted(): void {
    const values: string[] = ["z", "Z", "a"];
    values.sort(compare);
    console.log(values.join(","));
}

sorted();
`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.LatentLoad([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := program.LatentDiagnostics(); len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	attempted := false
	lower.Latent(context.Background(), program, func(record lower.LatentFile) {
		for _, finding := range record.Findings {
			t.Errorf("%s %s: %s", finding.Phase, finding.Where, finding.Reason)
		}
		for _, unit := range record.Units {
			if unit.Name == "sorted" && unit.Status == "attempted" {
				attempted = true
			}
		}
	})
	if !attempted {
		t.Fatal("sorted was not attempted")
	}
}

func TestLatentScoutNamedSortComparator(t *testing.T) {
	t.Parallel()
	const source = `const enum Comparison { Less = -1, Equal = 0, Greater = 1 }
function compare(a: string | undefined, b: string | undefined): Comparison {
    return a === b ? Comparison.Equal : a === undefined ? Comparison.Less :
        b === undefined ? Comparison.Greater : a < b ? Comparison.Less : Comparison.Greater;
}
const names = ["z", "a", "Z"];
names.sort(compare);
console.log(names.join(","));
`
	path := filepath.Join(t.TempDir(), "111-named-comparator.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.LatentLoad([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := program.LatentDiagnostics(); len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	attempted := false
	lower.Latent(context.Background(), program, func(record lower.LatentFile) {
		for _, finding := range record.Findings {
			t.Errorf("phase=%s where=%s unit=%s kind=%s reason=%q text=%q",
				finding.Phase, finding.Where, finding.Unit, finding.Kind, finding.Reason, finding.Text)
		}
		for _, unit := range record.Units {
			if strings.HasSuffix(unit.Where, "/111-named-comparator.ts:7:1") {
				t.Logf("unit=%s name=%q kind=%s status=%s", unit.Where, unit.Name, unit.Kind, unit.Status)
				attempted = unit.Status == "attempted"
			}
		}
	})
	if !attempted {
		t.Fatal("unit at 7:1 was not attempted")
	}
}
