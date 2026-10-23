package lower

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func viewDiagnosticLocation(t *testing.T, name, refusal, location string) {
	t.Helper()
	path := filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a")
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), refusal) || !strings.Contains(err.Error(), "oct9_views_"+name+".a:"+location+":") {
		t.Fatalf("want %q with source location %s, got %v", refusal, location, err)
	}
}

func TestViewDiagnosticP17(t *testing.T) {
	t.Parallel()
	// Element access is refused at its key before any contract is consulted.
	viewDiagnosticLocation(t, "p17", "view member read by element access is not yet checked", "6:16")
}

func TestViewDiagnosticP48(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p48", "unsupported collection contract", "6:9")
}

func TestViewDiagnosticP64(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p64", "view member read by element access is not yet checked", "7:16")
}
