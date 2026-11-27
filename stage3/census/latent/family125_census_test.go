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

func TestLatentTracingNamespaceValue(t *testing.T) {
	t.Parallel()
	const source = "namespace tracingEnabled { export function read(): number { return 7; } }\n" +
		"let tracing: typeof tracingEnabled | undefined;\n" +
		"function startTracing(): void {        tracing = tracingEnabled; console.log(`${tracing.read()}`); }\n" +
		"startTracing();\n"
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
			if unit.Name == "startTracing" && unit.Status == "attempted" {
				attempted = true
			}
		}
	})
	if !attempted {
		t.Fatal("startTracing was not attempted")
	}
}

func TestLatentScoutNamespacePublication(t *testing.T) {
	t.Parallel()
	const source = `let published: typeof Tracing | undefined;
namespace Tracing {
    export function enable(): void { published = Tracing; }
    export function read(): number { return 7; }
}
Tracing.enable();
console.log(String(published?.read()));
`
	path := filepath.Join(t.TempDir(), "125-namespace-publication.ts")
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
			if strings.HasSuffix(unit.Where, "/125-namespace-publication.ts:2:1") {
				t.Logf("unit=%s name=%q kind=%s status=%s", unit.Where, unit.Name, unit.Kind, unit.Status)
				attempted = unit.Status == "attempted"
			}
		}
	})
	if !attempted {
		t.Fatal("unit at 2:1 was not attempted")
	}
}
