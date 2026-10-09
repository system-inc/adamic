package lower

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func viewDiagnosticLocation(t *testing.T, name, location string) {
	t.Helper()
	path := filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a")
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "view type has an unsupported member: values") || !strings.Contains(err.Error(), "oct9_views_"+name+".a:5:14:") {
		t.Fatalf("want creation-time collection refusal, got %v", err)
	}
	// Creation refuses this contract before lowering its read. Isolate the
	// read handoff so the original read location remains independently tested.
	var read *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if strings.HasSuffix(program.Where(node), ":"+location) {
			read = node
			return true
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	if read == nil {
		t.Fatalf("missing read node at %s", location)
	}
	l := &lowering{program: program, result: &ir.Program{}}
	property := l.readViewMember(read, ir.Property{Name: "values", Of: ir.Map}, nil, nil).(ir.Property)
	if !strings.HasSuffix(property.ViewWhere, ":"+location) {
		t.Fatalf("want collection read source location %s, got %q", location, property.ViewWhere)
	}
}

func TestViewDiagnosticP17(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p17", "6:16")
}

func TestViewDiagnosticP48(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p48", "6:9")
}

func TestViewDiagnosticP64(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p64", "7:16")
}
