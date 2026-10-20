package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// sameRecordStorage over two instances of one generic type compares their arguments, never their
// members. A member walk expands Promise's then and catch into fresh instantiations at every level,
// so no visited pair repeats: on TestVoidWideningIsRefused/container_promise.a it made 1.3 million
// calls, 294,120 deep, with 882,354 pairs visited, 36 s of one oracle unit (#2thvgyb). The visited
// set is the walk's own count, so the bound needs no clock. The answers are the member walk's.
// Mutant: drop the argument comparison, and 882,354 pairs are visited again.
func TestRecordStorageComparesGenericArguments(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "function store(holder: Promise<() => void>, wide: Promise<() => unknown>, same: Promise<() => void>): void {}\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked}
	parameters := map[string]*checker.Type{}
	for _, parameter := range file.Statements.Nodes[0].AsFunctionDeclaration().Parameters.Nodes {
		name := parameter.Name()
		parameters[name.Text()] = checked.GetTypeAtLocation(name)
	}
	for _, pair := range []struct {
		from, to string
		same     bool
	}{
		{"holder", "wide", true},
		{"wide", "holder", true},
		{"holder", "same", true},
	} {
		visited := map[[2]*checker.Type]bool{}
		if got := l.sameRecordStorage(parameters[pair.from], parameters[pair.to], visited); got != pair.same {
			t.Errorf("%s to %s: same storage %v, want %v", pair.from, pair.to, got, pair.same)
		}
		if len(visited) > 16 {
			t.Errorf("%s to %s: %d pairs visited, want at most 16", pair.from, pair.to, len(visited))
		}
	}
}
