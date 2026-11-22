package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A view cast's member-literal slots are checked at each read through the view, by name: present, and holding
// one of the slot's members, else the located stop (#snypn8c c6, Adamic's ruling under rulings.md:18). Each of
// tsc's shapes is a program here: a nested parent (binder 3078, emitter 1888), an overlap cast's own kind
// (binder 3369, moduleSpecifiers 915, parser 10804 and 10805, ts.ts 2191), a union nested parent (emitter
// 1868), and name.expression.kind two levels down (declarations 1808). Mutants, in
// internal/lower/view_cast_enum_slots.mutants.md: the slots dropped from the walk refuses every one of these
// at the cast.
const viewCastEnumSlotsDirectory = "internal/oracle/testdata/view_cast_enum_slots/"

func viewCastEnumSlotsFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, viewCastEnumSlotsDirectory+name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Each honest program agrees with Node natively, sanitized and on the JavaScript backend.
func TestViewCastEnumSlotsHonestAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"nested_parent.ts", "own_kind.ts", "union_parent.ts", "two_level.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, viewCastEnumSlotsDirectory+name)
		})
	}
}

// Each lie stops at the located read with the view's literal failure and exit 70, after the honest call has
// printed what Node prints. The lie's nested kind isn't the member, and it reaches the read through the cast. A union
// nested parent stops one step sooner, at the parent itself: its kind is how the union contract tells which member
// it holds, so a kind that is no member's matches none (the same check, said where the union is read).
// Node, without the checks, prints on.
func TestViewCastEnumSlotsLiesStopAtTheRead(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, node, stdout, stderr string }{
		{"nested_parent_lie.ts", "size 1\nsize 8\n", "size 1\n", "field read failed: tag.parent.kind expected Kind.Doc, found number 8"},
		{"own_kind_lie.ts", "5 size\n8 width\n", "5 size\n", "field read failed: named.kind expected Kind.Name, found number 8"},
		{"union_parent_lie.ts", "1\n8\n", "1\n", "field read failed: tag.parent matches no member of Doc | TypeLiteral; expected Doc | TypeLiteral, found object"},
		{"two_level_lie.ts", "5\n8\n", "5\n", "field read failed: declaration.name.expression.kind expected Kind.Name, found number 8"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := viewCastEnumSlotsFixture(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte("adamic: panic: " + stop.stderr + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
				}
			}
		})
	}
}

// Lower's risk rows, each either agreeing or stopping, never lowering raw: a Node subtype's full contract (a
// HasJSDoc-style union parent and a NodeArray-like readonly array field), emitter 1868's union nested slot, a
// checked parent meeting a setParent-style write elsewhere in the program, and name.expression.kind two levels
// down. A program that lowers must agree with Node honestly, and its lie must stop at the read with the view's
// literal failure; one that doesn't lower must be refused or not yet, at its honest program and its lie alike.
func TestViewCastEnumSlotsRiskRowsAgreeOrStop(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, node, stdout, stop string }{
		{"contracts", "1 0\n1 8\n", "1 0\n", "tag.parent.parent matches no member of HasDoc; "},
		{"union_parent", "1\n8\n", "1\n", "tag.parent matches no member of Doc | TypeLiteral; "},
		{"set_parent", "1\n0\n", "1\n", "tag.parent.kind expected "},
		{"two_level", "5\n8\n", "5\n", "declaration.name.expression.kind expected "},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			honest, lie := viewCastEnumSlotsFixture(t, row.name+".ts"), viewCastEnumSlotsFixture(t, row.name+"_lie.ts")
			if truth := onNode(t, lie); truth.exitCode != 0 || string(truth.stdout) != row.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, honestErr := lowered(t, honest)
			lieProgram, lieErr := lowered(t, lie)
			if honestErr != nil || lieErr != nil {
				for _, err := range []error{honestErr, lieErr} {
					var refused *lower.Refused
					var notYet *lower.NotYet
					if err == nil || !errors.As(err, &refused) && !errors.As(err, &notYet) {
						t.Fatalf("one of the pair lowers and the other doesn't, or a stop that is neither refused nor not yet: honest %v, lie %v", honestErr, lieErr)
					}
				}
				t.Logf("stops: honest %v, lie %v", honestErr, lieErr)
				return
			}
			loomViewsDAgrees(t, viewCastEnumSlotsDirectory+row.name+".ts")
			prefix := "adamic: panic: field read failed: " + row.stop
			sanitized, _ := nativelyUncached(t, lieProgram)
			for backend, got := range map[string]run{"native": releasedUncached(t, lieProgram), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, lieProgram)} {
				if got.exitCode != 70 || string(got.stdout) != row.stdout || !strings.HasPrefix(string(got.stderr), prefix) {
					t.Errorf("%s: exit %d, stdout %q, stderr %q, want exit 70, stdout %q and a stop beginning %q", backend, got.exitCode, got.stdout, got.stderr, row.stdout, prefix)
				}
			}
		})
	}
}
