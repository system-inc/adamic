package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Read checked parameter types directly: unrelated body admission must not hide this boundary.
func TestClockStringNullRepresentation(t *testing.T) {
	for _, probe := range []struct {
		name, declared string
		admit          bool
	}{
		{"nullable-string", "string | null", true},
		{"nullable-literals", "'text' | '' | null", true},
		{"three-state", "string | null | undefined", false},
		{"nullable-number", "number | null", false},
		{"mixed-nullable", "string | number | null", false},
		{"nullable-object", "{ value: string } | null", false},
		{"present-only", "string", false},
		{"present-literals", "'text' | ''", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "probe.a")
			if err := os.WriteFile(path, []byte("function probe(value: "+probe.declared+"): void {}"), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			entry := program.Files()[0]
			checked, release := program.Checker(context.Background(), entry)
			defer release()
			parameter := entry.Statements.Nodes[0].AsFunctionDeclaration().Parameters.Nodes[0]
			proven := checked.GetTypeAtLocation(parameter)
			l := &lowering{program: program, checker: checked}
			if got := l.clockNullableString(proven); got != probe.admit {
				t.Fatalf("admission for %s = %v, want %v", probe.declared, got, probe.admit)
			}
			if probe.admit {
				if kind, known := l.representation(proven); !known || kind != ir.String {
					t.Fatalf("got %v, %v; want string representation", kind, known)
				}
			}
			if probe.name == "three-state" {
				if _, known := l.representation(proven); known {
					t.Fatal("three-state null/undefined union must remain unrepresented")
				}
			}
		})
	}
}
