package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Register here because this unit does not own oracle_test.go. Both the differential oracle
// and the recorded-count gate see the same fixture.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{
		"internal/oracle/testdata/discriminant_construction.a", true, false,
	})
}

func TestDiscriminantWitnessRefusedBeforeLowering(t *testing.T) {
	t.Parallel()
	// Exact witness from coverage/proven-guards-relations da12012. Its predicate is still
	// refused independently on main; the write diagnostic must win before that refusal.
	source := "type Leaf = { kind: 'leaf'; readonly value: number };\ntype Branch = { kind: 'branch'; readonly label: string };\ntype SyntaxNode = Leaf | Branch;\nfunction change(node: SyntaxNode): void { node.kind = 'branch'; }\nfunction isBranch(node: SyntaxNode): node is Branch { return node.kind === 'branch'; }\nfunction show(node: SyntaxNode): void { change(node); if (isBranch(node)) { console.log(`branch ${node.label}`); } }\nshow({ kind: 'leaf', value: 4 });\n"
	path := filepath.Join(t.TempDir(), "witness.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	observation := onNode(t, path)
	if observation.exitCode != 0 || string(observation.stdout) != "branch undefined\n" || len(observation.stderr) != 0 {
		t.Fatalf("Node witness changed: %#v", observation)
	}
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want discriminant refusal, got %v", err)
	}
	want := path + ":4:43: Adamic 0.1 refuses a write to discriminant field 'kind' after construction; changing variant means building a new object"
	if refused.Error() != want {
		t.Fatalf("want %q, got %v", want, err)
	}
}

func TestDiscriminantAcceptedProgramsMatchNode(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Leaf { kind: 'leaf'; constructor() { this.kind = 'leaf'; this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);",
		"type Node = { kind: 'leaf' } | { kind: string; value: number };\nfunction change(node: { kind: string; value: number }): void { node.kind = 'new'; }\nconst node = { kind: 'old', value: 1 }; change(node); console.log(node.kind);",
		"class Plain { kind: 'leaf' = 'leaf'; change(): void { this.kind = 'leaf'; } }\nconst node = new Plain(); node.change(); console.log(node.kind);",
	} {
		path := filepath.Join(t.TempDir(), "accepted.a")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		sanitized, binary := natively(t, program)
		for _, observed := range []run{released(t, program), sanitized, onJavaScriptBackend(t, program)} {
			if difference := disagreement(truth, observed); difference != "" {
				t.Fatal(difference)
			}
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
}

func TestDiscriminantBackstopUnchanged(t *testing.T) {
	t.Parallel()
	// This unit changes only admission. Keep the native missing-field backstop untouched.
	source, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/object.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "compiler bug: a field the checker proved is there is missing") {
		t.Fatal("missing field backstop changed")
	}
}
