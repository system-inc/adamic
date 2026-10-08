package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Register this slice separately from the shared fixture list.
func init() {
	for _, fixture := range []struct {
		path    string
		lowers  bool
		checked bool
	}{
		{"internal/oracle/testdata/library_object_public_hash.a", false, false},
		{"internal/oracle/testdata/library_object_iterator_view.a", false, false},
		{"internal/oracle/testdata/library_object_try_assign.a", false, false},
		{"internal/oracle/testdata/library_object_prefix_names.a", true, false},
		{"internal/oracle/testdata/library_object_groups_proto2.a", false, false},
		{"internal/oracle/testdata/library_object_groups_proto.a", false, false},
		{"internal/oracle/testdata/library_object_iterator_own.a", true, false},
		{"internal/oracle/testdata/library_object_iterator_tag.a", true, false},
		{"internal/oracle/testdata/library_object_private_mangled.a", true, false},
	} {
		fixtures = append(fixtures, fixture)
	}
}

// If a refusal regresses, run the accepted artifact so the failure names Node's
// observable behavior, rather than merely noticing that lowering succeeded.
func assertObjectPrototypeRefusal(t *testing.T, fixture, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, reason) {
			t.Fatalf("want NotYet containing %q, got %v", reason, err)
		}
		return
	}
	oracle := onNode(t, path)
	native, _ := natively(t, program)
	if difference := disagreement(oracle, native); difference != "" {
		t.Fatalf("%s: node exit %d stdout %q; native exit %d stdout %q", difference, oracle.exitCode, oracle.stdout, native.exitCode, native.stdout)
	}
	t.Fatal("probe now agrees with Node; implement and register the newly supported behavior")
}

func TestObjectGroupsHaveNoPrototype(t *testing.T) {
	for _, fixture := range []string{"library_object_groups_proto.a", "library_object_groups_proto2.a"} {
		t.Run(fixture, func(t *testing.T) {
			assertObjectPrototypeRefusal(t, fixture, "null prototype")
		})
	}
}

func TestObjectAssignFailureStaysRefused(t *testing.T) {
	assertObjectPrototypeRefusal(t, "library_object_try_assign.a", "Object.assign into a potentially frozen object")
}

func TestObjectPrototypeRepresentationsStayDistinct(t *testing.T) {
	for _, probe := range []struct{ fixture, reason string }{
		{"library_object_public_hash.a", "public # name"},
		{"library_object_iterator_view.a", "collection iterator may be hidden"},
	} {
		t.Run(probe.fixture, func(t *testing.T) {
			assertObjectPrototypeRefusal(t, probe.fixture, probe.reason)
		})
	}
}
