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
func TestRecordGuardedReadAdmitted(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_guarded/admitted.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("a read no record can reach must lower: %v", loweredError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 {
		test.Fatalf("Node: %+v", node)
	}
	native, _ := natively(test, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(node, got); difference != "" {
			test.Fatalf("%s disagrees with Node: %s", backend, difference)
		}
	}
}

// A read a record may reach stays NotYet: through a type a record's type is assignable to, through
// a type parameter's constraint, through ?. on the record's own type, and through a union holding a
// record. The view rule (a), the storage type rule (b) and the narrower direction of (c) in
// entries_records.go are fail-closed extras with no witness of their own: every entrance to them
// found so far (a cast view, a checked predicate to a subtype, in-narrowing, a discriminant read) is
// refused earlier, so nothing here pins them.
func TestRecordGuardedReadRefused(test *testing.T) {
	test.Parallel()
	for name, where := range map[string]string{
		"reached_assignable": ":7:16",
		"reached_generic":    ":6:16",
		"reached_own":        ":9:15",
		"reached_narrowed":   ":8:6",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_guarded/"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, path)
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) || notYet.What != recordGuardedReason || notYet.Where != path+where {
				test.Fatalf("want NotYet %q at %s, got %v", recordGuardedReason, where, loweredError)
			}
		})
	}
}

// Two ways a record could reach a union or maybe read are stopped earlier today: a predicate that
// narrows a record to a subtype (the narrower direction of rule c's only road in), and a record
// laundered through any. Each pins its current stop. When one moves, because predicates or any start
// to lower, this goes red, and that fixture belongs with the reached_ ones, stopped at the read.
func TestRecordGuardedReadStopsEarlier(test *testing.T) {
	test.Parallel()
	for name, want := range map[string]struct{ where, what string }{
		"early_predicate": {":7:41", "a type predicate whose return is not proven"},
		"early_any":       {":5:7", "a value of type any"},
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_guarded/"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case errors.As(loweredError, &refused) &&
				refused.Where == path+want.where &&
				strings.HasPrefix(refused.What, want.what):
			case errors.As(loweredError, &notYet) &&
				notYet.Where == path+want.where &&
				strings.HasPrefix(notYet.What, want.what):
			default:
				test.Fatalf("want the earlier stop %q at %s, got %v", want.what, want.where, loweredError)
			}
		})
	}
}
