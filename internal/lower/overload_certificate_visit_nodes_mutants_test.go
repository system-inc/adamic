package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// Each mutant of the guarded-array certificate removes one of its conditions on #ytagqs5, and the pin
// that condition holds goes red. lowered marks a pin that stops a program, which under the mutant must
// lower outright ("got <nil>"), so its red is the condition's alone and not a stop some other road says.
// Two conditions have no mutant, since no guard of the certificate's is the only one holding their rows
// (probe J-mh34cmy): condition 2's lying test stops at its own check, which predicateArguments and the
// checked forwarder hold (TestVisitNodesCertificateLyingTestStopsAtItsCheck), and condition 4's getter is
// stopped as well by refuseReachedContractGetters (TestVisitNodesCertificateGetterStops).
type visitNodesMutation struct {
	name, file, before, after, pin string
	rows                           []string
	lowered                        bool
}

var visitNodesMutations = []visitNodesMutation{
	// Condition 1: the writable-view scan answers nothing.
	{"WritableView", "internal/lower/overload_certificate_visit_nodes.go", "\tif views.writable != nil {\n", "\tif false && views.writable != nil {\n", "TestVisitNodesCertificateWritableViewStops", []string{"(iv-a)_kept_raw"}, true},
	// Condition 1's top type hop: a source that may hold a NodeArray read as today, as none. The scan's own pin
	// goes red; the hop programs still stop at view_arrays.go or castRefusal (probes J-v8b36r7, J-3ffrmk8),
	// so no program lowers and lowered is false.
	{"MayHold", "internal/lower/overload_certificate_visit_nodes.go", "\tif proven == nil {\n\t\treturn false\n\t}\n\tif seen[proven] {\n", "\tif true || proven == nil {\n\t\treturn false\n\t}\n\tif seen[proven] {\n", "TestVisitNodesCertificateScanReadsHops", []string{"(a)_through_unknown", "(b)_through_object", "(c)_through_{}", "(d)_through_a_type_parameter", "(e)_through_an_unknown_parameter"}, false},
	// Condition 1's typed holders: the object case's property test of -8 restored, so ArrayLike<Node> and
	// Iterable<Node>, each with properties, may hold no NodeArray. The hops never lower, so lowered is false.
	{"Holder", "internal/lower/overload_certificate_visit_nodes.go", "\treturn l.checker.IsTypeAssignableTo(predicateReadonlyArray(l.checker, l.checker.GetAnyType(), true), proven)\n", "\treturn len(l.checker.GetPropertiesOfType(proven)) == 0 && !l.castSignatures(proven)\n", "TestVisitNodesCertificateScanReadsHops", []string{"through_ArrayLike<Node>", "through_Iterable<Node>"}, false},
	// Condition 3: the guard read from any earlier statement of the block, not the one right before.
	{"Adjacent", "internal/lower/overload_certificate_visit_nodes.go", "\treturn l.overloadGuardedLoop(implementation, guarded, test, statements[position-1], local)\n", "\tfor _, earlier := range statements[:position] {\n\t\tif l.overloadGuardedLoop(implementation, guarded, test, earlier, local) {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n", "TestVisitNodesCertificateLoopShape", []string{"a_statement_between"}, true},
	// Condition 5: overloadCertificateSameHeld's NodeArray arm dropped, so a NodeArray is no collection it reads.
	{"Storage", "internal/lower/overload_certificate_array.go", "\tif storedNodeArray || readNodeArray {\n", "\tif false && (storedNodeArray || readNodeArray) {\n", "TestVisitNodesCertificateStorage", nil, false},
	// Condition 6: the undefined arm's return taken anywhere in its branch, not only first.
	{"FirstStatement", "internal/lower/overload_certificate_visit_nodes.go", "len(block.Statements.Nodes) == 0 || block.Statements.Nodes[0] != statement {", "len(block.Statements.Nodes) == 0 {", "TestVisitNodesCertificateUndefinedArm", []string{"a_statement_before_the_return"}, true},
	// Condition 7: #3cee4gb's refusal of a delayed mutator on a readonly receiver dropped.
	{"DelayedMutator", "internal/lower/library_method_values.go", "\tif !delayedArrayLeavesReceiver[name] {\n", "\tif false && !delayedArrayLeavesReceiver[name] {\n", "TestVisitNodesCertificateDelayedMutatorStops", nil, true},
	// Condition 8: the readonly test of the membership's fields answers no view.
	{"ReadonlyMembership", "internal/lower/overload_certificate_visit_nodes.go", "l.overloadGuardedWritableField(views, field, element, target); writable {", "l.overloadGuardedWritableField(views, field, element, target); false && writable {", "TestVisitNodesCertificateWritableMembershipStops", []string{"a_writable_view_of_the_element"}, true},
}

// visitNodesMutantBinary builds the named mutant's test binary, a product.
func visitNodesMutantBinary(t *testing.T, name string) string {
	t.Helper()
	for _, m := range visitNodesMutations {
		if m.name == name {
			return buildcache.GoTestMutated(t, "", "visit-nodes-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: m.file, Before: m.before, After: m.after}})
		}
	}
	t.Fatalf("no visit-nodes mutant %s", name)
	return ""
}
func TestProduct_VisitNodesWritableViewMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "WritableView")
}
func TestProduct_VisitNodesMayHoldMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "MayHold")
}
func TestProduct_VisitNodesHolderMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "Holder")
}
func TestProduct_VisitNodesAdjacentMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "Adjacent")
}
func TestProduct_VisitNodesStorageMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "Storage")
}
func TestProduct_VisitNodesFirstStatementMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "FirstStatement")
}
func TestProduct_VisitNodesDelayedMutatorMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "DelayedMutator")
}
func TestProduct_VisitNodesReadonlyMembershipMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, "ReadonlyMembership")
}

func TestVisitNodesMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range visitNodesMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := visitNodesMutantBinary(t, m.name)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+m.pin+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("%s mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			for _, row := range m.rows {
				if !strings.Contains(string(stdout), "--- FAIL: "+m.pin+"/"+row+" ") {
					t.Fatalf("%s mutant left %s/%s green:\n%s", m.name, m.pin, row, stdout)
				}
			}
			if m.lowered && !strings.Contains(string(stdout), "got <nil>") {
				t.Fatalf("%s mutant turned %s red without lowering its program, so another road holds it:\n%s", m.name, m.pin, stdout)
			}
			t.Logf("%s mutant caught by %s", m.name, m.pin)
		})
	}
}
