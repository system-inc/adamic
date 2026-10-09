package program

import (
	"context"
	"path"
	"testing"
)

// Behaviour the pinned compiler must have, tested against the checker this package links rather than
// trusted from the commit it was pinned for. Each case names the upstream report it guards, so a
// failure here says which fix is missing rather than only that a count changed.

// microsoft/TypeScript#64593, fixed by #64594. Reverse-mapped inference copied a class's private
// members into the inferred type as public ones, so since #63932 (intersection properties take the
// most permissive accessibility) `getAccount() as Account` reported TS2352. tsc 6.0.3 accepts it.
//
// cohere's compiler is System, Inc.'s fork, system-inc/TypeScript on its adamic branch, which carries
// the fix while the PR is open, because upstream main still has the defect. This is what fails if the
// pin moves back to upstream before Microsoft merges, and it stays after they do. It was found on two
// Base authentication casts of this shape.
func TestReverseMappedInferenceDoesNotExposePrivateMembers(t *testing.T) {
	t.Parallel()
	source := `class Entity {
    private secret = 1;
    markChanged(field: string): void {}
}
class Account extends Entity {
    emailAddress = '';
}
declare function getAccount<T extends object = object>(): Readonly<Account & T>;
export const account = getAccount() as Account;
`
	if diagnostics := checkInMemory(t, source); diagnostics != 0 {
		t.Fatalf("the linked checker reports %d diagnostics on the #64593 repro, expected 0: the pinned "+
			"compiler is missing the fix from #64594", diagnostics)
	}
}

// checkInMemory type-checks one file under --strict, built entirely through the overlay with an
// explicit file list, so it reads nothing from the disk and writes nothing anywhere.
func checkInMemory(t *testing.T, source string) int {
	t.Helper()
	root := path.Join(t.TempDir(), "regression")
	configFileName := path.Join(root, "tsconfig.json")
	graph, err := Build(Options{
		ConfigFileName:   configFileName,
		CurrentDirectory: root,
		SingleThreaded:   true,
		Overlay: map[string]string{
			configFileName: `{ "compilerOptions": { "strict": true, "target": "es2022", "module": "esnext", ` +
				`"noEmit": true, "skipLibCheck": true }, "files": ["probe.ts"] }`,
			path.Join(root, "probe.ts"): source,
		},
	})
	if err != nil {
		t.Fatalf("building the in-memory program: %v", err)
	}
	return len(graph.AllDiagnostics(context.Background()))
}
