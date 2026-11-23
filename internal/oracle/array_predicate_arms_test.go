package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An object arm declaring a required property no array the program can hold has is in Array.isArray's domain
// (Adamic's ruling on #s8gy89r), and the checked narrowed-union read leans on the same premise (locals.go): an
// assignment narrows the union to its array side, and the read checks that by the array brand, never typeof. Each
// program used to stop at that read, and now agrees with Node on the sanitized native build (natively: ASan and
// UBSan, recovering from nothing, as a registered fixture's), on the release build and on the JavaScript backend:
//   - object_member: `{ name: string } | readonly Extension[]`, whose name no array has.
//   - narrowed_union_map_arm: `number[] | Map<string, number>`, a library interface arm, whose size and get no array
//     has. Its false side reads no member: a read of v.size there still stops at the narrowed-union read
//     (Language's probe, J-d087fzy on #t4nbfj1), and the program reads only .length on the array side.
//
// The fixtures sit outside testdata's top level, so the flow and fresh corpora don't take them as their own.
func TestArrayPredicateArmsAgree(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, output string }{
		{"object_member", "true\n"},
		{"narrowed_union_map_arm", "1\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_predicate_arms", row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if disagreement(run{stdout: []byte(row.output)}, want) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			machine, _ := natively(t, program)
			for backend, got := range map[string]run{"native": machine, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// Tuples are held as objects so far, so a tuple arm stays out of Array.isArray's domain beside a Map arm that is in
// it, mutable or readonly (Language's pins, J-d087fzy on #t4nbfj1): the call stops at its argument.
func TestArrayPredicateTupleBesideMapRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"tuple_beside_map", "readonly_tuple_beside_map"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union_refused", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); disagreement(run{stdout: []byte("true\n")}, node) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.NotYet
			const reason = "Array.isArray on a tuple or an erased object/any/unknown view"
			if !errors.As(err, &refused) || refused.What != reason || refused.Where != path+":5:34" {
				t.Fatalf("want NotYet %q at :5:34, got %v", reason, err)
			}
		})
	}
}
