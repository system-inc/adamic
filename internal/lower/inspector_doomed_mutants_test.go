package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func inspectorDoomedMutant(t *testing.T, mode string) string {
	t.Helper()
	file := "internal/lower/inspector_doomed.go"
	var mutations []buildcache.Mutation
	switch mode {
	case "AnyRequire":
		mutations = []buildcache.Mutation{{File: file, Before: "if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember && nodeInspectorRequireThrows(initializer) {\n\t\t\treturn l.nodeCommonJSRequireRoot(initializer)\n\t\t}", After: "if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember { return initializer }"}}
	case "Let":
		mutations = []buildcache.Mutation{{File: file, Before: " || declaration.Parent.Flags&ast.NodeFlagsConst == 0", After: ""}}
	case "Receiver":
		mutations = []buildcache.Mutation{{File: file, Before: "if require == nil {\n\t\treturn nil, nil\n\t}", After: "// mutant: every inspector receiver is doomed"}}
	case "ReadDrop":
		mutations = []buildcache.Mutation{{File: file, Before: "read, err := l.expression(root)", After: "_, err := ir.Expression(nil), error(nil)"}, {File: file, Before: "ir.Evaluate{Value: read}, ", After: ""}}
	case "ValueHook":
		mutations = []buildcache.Mutation{{File: "internal/lower/expression.go", Before: "\tif value, handled, err := l.inspectorDoomedValue(node); handled {\n\t\treturn value, err\n\t}\n", After: "\t// mutant: omit the value trap hook\n"}}
	case "Fallthrough":
		mutations = []buildcache.Mutation{{File: file, Before: "func (l *lowering) checkInspectorDoomedContracts() error {", After: "func (l *lowering) checkInspectorDoomedContracts() error {\n return nil // mutant: a contract skip may fall through without its Panic"}}
	default:
		t.Fatalf("unknown inspector mutant %s", mode)
	}
	return buildcache.GoTestMutated(t, "", "inspector-doomed-"+mode+".test", "./internal/lower", nil, mutations)
}
func TestProduct_InspectorDoomedAnyRequireMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "AnyRequire")
}
func TestProduct_InspectorDoomedLetMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "Let")
}
func TestProduct_InspectorDoomedReceiverMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "Receiver")
}
func TestProduct_InspectorDoomedReadDropMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "ReadDrop")
}
func TestProduct_InspectorDoomedFallthroughMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "Fallthrough")
}
func inspectorDoomedCaught(t *testing.T, mode, pin string, lines ...string) {
	t.Helper()
	binary := inspectorDoomedMutant(t, mode)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, binary, "-test.run", "^"+pin+"$", "-test.count=1", "-test.v"))
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 {
		t.Fatalf("%s mutant: err=%v stdout=%s stderr=%s", mode, err, stdout, stderr)
	}
	for _, line := range lines {
		if !strings.Contains(string(stdout), line) {
			t.Fatalf("%s mutant missing catch %q: stdout=%s stderr=%s", mode, line, stdout, stderr)
		}
		t.Log(pin + " goes red: " + line)
	}
}
func TestInspectorDoomedAnyRequireMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "AnyRequire", "TestInspectorOtherRequireBindingsAreNotDoomed", "fs require must not create a doomed binding", "source-map-support require must not create a doomed binding")
}
func TestInspectorDoomedLetMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "Let", "TestInspectorLetReceiverStaysRefused", "let receiver must not be doomed")
}
func TestInspectorDoomedReceiverMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "Receiver", "TestInspectorNonDoomedMergedMemberStaysNotYet", "non-doomed Session.post must keep rule 800 NotYet")
}
func TestInspectorDoomedReadDropMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "ReadDrop", "TestInspectorDoomedEscapedReadRefused", "d1.ts must stop at the ordinary read")
	inspectorDoomedCaught(t, "ReadDrop", "TestInspectorDoomedHoistedReadKeepsDeadZone", "hoisted use lost its ordinary dead-zone read")
}
func TestInspectorDoomedFallthroughMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "Fallthrough", "TestInspectorDoomedMemberValueHasPanic", "lost trap must fail closed")
}

func TestProduct_InspectorDoomedValueHookMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "ValueHook")
}
func TestInspectorDoomedValueHookMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "ValueHook", "TestInspectorDoomedValueHookRequired", "skipped inspector contract without its doomed-use Panic")
}
