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
		mutations = []buildcache.Mutation{{File: file, Before: "if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember && nodeInspectorRequireThrows(initializer) {\n\t\t\treturn l.nodeCommonJSRequireRoot(initializer), false\n\t\t}", After: "if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember { return initializer, false }"}}
	case "Let":
		mutations = []buildcache.Mutation{{File: file, Before: "if list.Flags&ast.NodeFlagsConst != 0 {", After: "if list.Flags&(ast.NodeFlagsConst|ast.NodeFlagsLet) != 0 {"}}
	case "Receiver":
		mutations = []buildcache.Mutation{{File: file, Before: "if require == nil {\n\t\treturn nil, nil\n\t}", After: "// mutant: every inspector receiver is doomed"}}
	case "ReadDrop":
		mutations = []buildcache.Mutation{{File: file, Before: "read, err := l.expression(root)", After: "_, err := ir.Expression(nil), error(nil)"}, {File: file, Before: "ir.Evaluate{Value: read}, ", After: ""}}
	case "ValueHook":
		mutations = []buildcache.Mutation{{File: "internal/lower/expression.go", Before: "\tif value, handled, err := l.inspectorDoomedValue(node); handled {\n\t\treturn value, err\n\t}\n", After: "\t// mutant: omit the value trap hook\n"}}
	case "Fallthrough":
		mutations = []buildcache.Mutation{{File: file, Before: "func (l *lowering) checkInspectorDoomedContracts() error {", After: "func (l *lowering) checkInspectorDoomedContracts() error {\n return nil // mutant: a contract skip may fall through without its Panic"}}
	case "AnyLet":
		mutations = []buildcache.Mutation{{File: file, Before: "if list.Flags&ast.NodeFlagsConst != 0 {", After: "if list.Flags&(ast.NodeFlagsConst|ast.NodeFlagsLet) != 0 {"}}
	case "WriteScan":
		mutations = []buildcache.Mutation{{File: file, Before: "function.Body().ForEachChild(visit)", After: "_ = visit // mutant: ignore assignments"}}
	case "FlowGate":
		mutations = []buildcache.Mutation{{File: file, Before: "if viaLet {", After: "if false && viaLet {"}}
	case "DeclaredType":
		mutations = []buildcache.Mutation{{File: file, Before: "if initializer == nil && !inspectorDeclaredUndefined(l.checker.GetTypeOfSymbol(symbol)) {", After: "if false && initializer == nil && !inspectorDeclaredUndefined(l.checker.GetTypeOfSymbol(symbol)) {"}}
	case "DerivedWrite":
		mutations = []buildcache.Mutation{{File: file, Before: "if doomed, viaLet := l.inspectorDoomedBinding(root, seen); doomed != nil && (!viaLet || ast.IsIdentifier(ast.SkipOuterExpressions(value, ast.OEKAll))) {", After: "if doomed, _ := l.inspectorDoomedBinding(root, seen); doomed != nil {"}}
	case "FifthWrite":
		mutations = []buildcache.Mutation{{File: "internal/lower/inspector_doomed_let_test.go", Before: "return string(data)", After: `return strings.Replace(string(data), "return disable();", "function fifth(other: Session): void { activeSession = other; } return disable();", 1)`}}

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

func TestInspectorLetAnyLetMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "AnyLet", "TestInspectorLetParameterStaysRefused", "unclassified let must stay refused")
}
func TestInspectorLetWriteScanMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "WriteScan", "TestInspectorLetParameterStaysRefused", "unclassified let must stay refused")
}
func TestInspectorLetFlowGateMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "FlowGate", "TestInspectorLetOptionalStaysRefused", "unclassified let must stay refused")
}
func TestInspectorLetDeclaredTypeMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "DeclaredType", "TestInspectorLetLanguageClosure", "Language closure must stay refused")
}
func TestInspectorLetFifthWriteMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "FifthWrite", "TestInspectorDoomedLetTSCShape", "tsc shape post must be doomed")
}

func TestProduct_InspectorLetAnyLetMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "AnyLet")
}

func TestProduct_InspectorLetWriteScanMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "WriteScan")
}

func TestProduct_InspectorLetFlowGateMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "FlowGate")
}

func TestProduct_InspectorLetDeclaredTypeMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "DeclaredType")
}

func TestProduct_InspectorLetFifthWriteMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "FifthWrite")
}

func TestInspectorLetDerivedWriteMutantIsCaught(t *testing.T) {
	t.Parallel()
	inspectorDoomedCaught(t, "DerivedWrite", "TestInspectorLetDerivedWriteStaysRefused", "unclassified let must stay refused")
}

func TestProduct_InspectorLetDerivedWriteMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedMutant(t, "DerivedWrite")
}
