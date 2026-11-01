package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

const recordGuardedReason = "a union or optional field read in a program with record storage"

// In a program with records, a union or optional field read lowers when no record can be its
// receiver, and matches Node natively and on the JavaScript backend (entries_records.go).
func TestRecordGuardedReadAdmitted(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_guarded/admitted.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("a read no record can reach must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 {
		t.Fatalf("Node: %+v", node)
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("%s disagrees with Node: %s", backend, difference)
		}
	}
}

// A read a record may reach stays NotYet: through a type a record's type is assignable to, through
// a type parameter's constraint, through ?. on the record's own type, and through a union holding a
// record. The view rule (a), the storage type rule (b) and the narrower direction of (c) in
// entries_records.go are fail-closed extras with no witness of their own: every entrance to them
// found so far (a cast view, a checked predicate to a subtype, in-narrowing, a discriminant read) is
// refused earlier, so nothing here pins them.
func TestRecordGuardedReadRefused(t *testing.T) {
	t.Parallel()
	for name, where := range map[string]string{"reached_assignable": ":7:16", "reached_generic": ":6:16", "reached_own": ":9:15", "reached_narrowed": ":8:6"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_guarded/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+where {
				t.Fatalf("want NotYet %q at %s, got %v", recordGuardedReason, where, err)
			}
		})
	}
}

// Two ways a record could reach a union or maybe read are stopped earlier today: a predicate that
// narrows a record to a subtype (the narrower direction of rule c's only road in), and a record
// laundered through any. Each pins its current stop. When one moves, because predicates or any start
// to lower, this goes red, and that fixture belongs with the reached_ ones, stopped at the read.
func TestRecordGuardedReadStopsEarlier(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]struct{ where, what string }{
		"early_predicate": {":7:41", "a type predicate whose return is not proven"},
		"early_any":       {":5:7", "a value of type any"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_guarded/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case errors.As(err, &refused) && refused.Where == path+want.where && strings.HasPrefix(refused.What, want.what):
			case errors.As(err, &notYet) && notYet.Where == path+want.where && strings.HasPrefix(notYet.What, want.what):
			default:
				t.Fatalf("want the earlier stop %q at %s, got %v", want.what, want.where, err)
			}
		})
	}
}
