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
// Condition 2 has no mutant: a lying test stops at its own check, which predicateArguments and the checked
// forwarder or closure hold, not a guard of the certificate's (TestVisitNodesCertificateLyingTestStopsAtItsCheck).
type visitNodesMutation struct {
	name, file, before, after, pin string
	rows                           []string
	lowered                        bool
}

var visitNodesMutations = []visitNodesMutation{
	// Condition 1: the writable-view scan answers nothing.
	{"WritableView", "internal/lower/overload_certificate_visit_nodes.go", "\tif views.writable != nil {\n", "\tif false && views.writable != nil {\n", "TestVisitNodesCertificateWritableViewStops", nil, true},
	// Condition 3: the guard read from any earlier statement of the block, not the one right before.
	{"Adjacent", "internal/lower/overload_certificate_visit_nodes.go", "\treturn l.overloadGuardedLoop(implementation, guarded, test, statements[position-1], local)\n", "\tfor _, earlier := range statements[:position] {\n\t\tif l.overloadGuardedLoop(implementation, guarded, test, earlier, local) {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n", "TestVisitNodesCertificateLoopShape", []string{"a_statement_between"}, true},
	// Condition 4: the getter scan answers nothing.
	{"Getter", "internal/lower/overload_certificate_visit_nodes.go", "\t\tif getter := views.getters[field]; getter != nil {\n", "\t\tif getter := views.getters[field]; false && getter != nil {\n", "TestVisitNodesCertificateGetterStops", nil, true},
	// Condition 5: overloadCertificateSameHeld's NodeArray arm dropped, so a NodeArray is no collection it reads.
	{"Storage", "internal/lower/overload_certificate_array.go", "\tif storedNodeArray || readNodeArray {\n", "\tif false && (storedNodeArray || readNodeArray) {\n", "TestVisitNodesCertificateStorage", nil, false},
	// Condition 6: the undefined arm's return taken anywhere in its branch, not only first.
	{"FirstStatement", "internal/lower/overload_certificate_visit_nodes.go", "len(block.Statements.Nodes) == 0 || block.Statements.Nodes[0] != statement {", "len(block.Statements.Nodes) == 0 {", "TestVisitNodesCertificateUndefinedArm", []string{"a_statement_before_the_return"}, true},
	// Condition 7: #3cee4gb's refusal of a delayed mutator on a readonly receiver dropped.
	{"DelayedMutator", "internal/lower/library_method_values.go", "\tif delayedArrayMutators[name] {\n", "\tif false && delayedArrayMutators[name] {\n", "TestVisitNodesCertificateDelayedMutatorStops", nil, true},
	// Condition 8: the readonly test of the membership's fields answers no view.
	{"ReadonlyMembership", "internal/lower/overload_certificate_visit_nodes.go", "l.overloadGuardedWritableField(views, field, element, target); writable {", "l.overloadGuardedWritableField(views, field, element, target); false && writable {", "TestVisitNodesCertificateWritableMembershipStops", nil, true},
}

func visitNodesMutantBinary(t *testing.T, m visitNodesMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "visit-nodes-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: m.file, Before: m.before, After: m.after}})
}
func TestProduct_VisitNodesWritableViewMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[0])
}
func TestProduct_VisitNodesAdjacentMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[1])
}
func TestProduct_VisitNodesGetterMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[2])
}
func TestProduct_VisitNodesStorageMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[3])
}
func TestProduct_VisitNodesFirstStatementMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[4])
}
func TestProduct_VisitNodesDelayedMutatorMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[5])
}
func TestProduct_VisitNodesReadonlyMembershipMutant(t *testing.T) {
	t.Parallel()
	visitNodesMutantBinary(t, visitNodesMutations[6])
}

func TestVisitNodesMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range visitNodesMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := visitNodesMutantBinary(t, m)
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
