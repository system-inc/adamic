package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func viewDiagnosticLocation(test *testing.T, name, refusal, location string) {
	test.Helper()
	path := filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a")
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil ||
		!strings.Contains(err.Error(), refusal) ||
		!strings.Contains(err.Error(), "oct9_views_"+name+".a:"+location+":") {
		test.Fatalf("want %q with source location %s, got %v", refusal, location, err)
	}
}

func TestViewDiagnosticP17(test *testing.T) {
	test.Parallel()
	// Element access is refused at its key before any contract is consulted.
	viewDiagnosticLocation(test, "p17", "view member read by element access is not yet checked", "6:16")
}

// viewDiagnosticTwinLocation lowers the fixture's source as a .ts file of the same name, where its cast
// stays a checked view, so the read it was written about is reached as .a reached it before #m8dvvmf.
func viewDiagnosticTwinLocation(test *testing.T, name, refusal, location string) {
	test.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{writeAgreementSource(test, "oct9_views_"+name+".ts", string(source))})
	if err != nil {
		test.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil ||
		!strings.Contains(err.Error(), refusal) ||
		!strings.Contains(err.Error(), "oct9_views_"+name+".ts:"+location+":") {
		test.Fatalf("twin: want %q with source location %s, got %v", refusal, location, err)
	}
}

// In .a p48's cast is a downcast whose checks no proof erases, refused (#m8dvvmf) before the
// destructured read; the .ts twin pins the read's own diagnostic.
func TestViewDiagnosticP48(test *testing.T) {
	test.Parallel()
	viewDiagnosticLocation(test, "p48", "a downcast from Root to View in .a whose checks no proof erases", "5:14")
	viewDiagnosticTwinLocation(test, "p48", "unsupported collection contract", "6:9")
}

func TestViewDiagnosticP64(test *testing.T) {
	test.Parallel()
	viewDiagnosticLocation(test, "p64", "view member read by element access is not yet checked", "7:16")
}
