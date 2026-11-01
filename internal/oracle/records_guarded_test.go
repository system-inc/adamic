package oracle

import (
	"errors"
	"path/filepath"
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
// a type parameter's constraint, through ?. on the record's own type, and through a discriminant that
// narrows a union holding a record to a member narrower than it. The view rule (a) and the storage
// type rule (b) in entries_records.go are fail-closed extras with no witness of their own: (c)
// implies both today, so nothing here pins them.
func TestRecordGuardedReadRefused(t *testing.T) {
	t.Parallel()
	for name, where := range map[string]string{"reached_assignable": ":7:16", "reached_generic": ":6:16", "reached_own": ":9:15", "reached_narrowed": ":8:9"} {
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
