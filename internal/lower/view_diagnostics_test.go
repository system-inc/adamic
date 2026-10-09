package lower

import (
	"context"
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
	if err == nil || !strings.Contains(err.Error(), "unsupported collection contract") || !strings.Contains(err.Error(), "oct9_views_"+name+".a:"+location+":") {
		t.Fatalf("want collection refusal with source location %s, got %v", location, err)
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
