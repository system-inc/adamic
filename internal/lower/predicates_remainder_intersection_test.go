package lower

import (
	"embed"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
	"testing"
)

//go:embed testdata/ruling12b/*.ts
var ruling12BPrograms embed.FS

func ruling12BSource(t *testing.T, name string) string {
	t.Helper()
	content, err := ruling12BPrograms.ReadFile("testdata/ruling12b/" + name + ".ts")
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestClosedIntersectionRemainderAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, stdout string }{
		{"b1", "name\nliteral\n"},
		{"b-three-parts", "name\nliteral2\n"},
		{"b-two-plain", "name\nliteral\nother\n"},
		{"b-tagless-part", "name\nliteral\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, ruling12BSource(t, test.name), ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != test.stdout || stderr != "" {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want exit 0 and %q", code, stdout, stderr, test.stdout)
			}
		})
	}
}

func TestClosedIntersectionRemainderWrongFalseStops(t *testing.T) {
	t.Parallel()
	overloadCheckStops(t, ruling12BSource(t, "b4"), "", "predicate isPropertyAccess", "false branch")
}

func TestClosedIntersectionRemainderRefusesAccessorPart(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, ruling12BSource(t, "b2"), ".ts")
	if err == nil || !strings.Contains(err.Error(), "checked predicate membership over an open tag domain") {
		t.Fatalf("want the open tag domain refusal, got %v", err)
	}
}

func TestClosedIntersectionRemainderRefusesClassPart(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, ruling12BSource(t, "b3"), ".ts")
	if err == nil || !strings.Contains(err.Error(), "a class instance") {
		t.Fatalf("want the class representation refusal, got %v", err)
	}

	// That refusal precedes predicate lowering. Pin the member test directly as well.
	l, types, _ := intersectionViewTypes(t)
	if l.checkedPredicateRemainderField(types["Nominal"]) {
		t.Fatal("the remainder member test must refuse the class-instance part")
	}
}

// Ruling 12 B checks each part's declaration. The pinned checker collects
// available declarations on the merged kind, so b2's accessor is visible
// there too. A mapped kind with no declaration of its own separates the checks.
func TestClosedIntersectionRemainderRefusesUndeclaredPart(t *testing.T) {
	t.Parallel()
	l, types, _ := intersectionViewTypes(t)
	member := types["MappedKind"]
	if member == nil || !l.plainObjectIntersection(member) {
		t.Fatal("MappedKind must be a plain object intersection")
	}
	merged := l.checker.GetPropertyOfType(member, "kind")
	if merged == nil || len(merged.Declarations) == 0 {
		t.Fatal("the merged kind must have a declaration")
	}
	for _, declaration := range merged.Declarations {
		if declaration.Kind != ast.KindPropertySignature && declaration.Kind != ast.KindPropertyDeclaration {
			t.Fatalf("the merged kind must have only field declarations, got %v", declaration.Kind)
		}
	}

	missing := ""
	for _, part := range member.Types() {
		if property := l.checker.GetPropertyOfType(part, "kind"); property != nil && len(property.Declarations) == 0 {
			missing = l.checker.TypeToString(part)
			break
		}
	}
	if missing == "" {
		t.Fatal("a part must have a kind with no declaration")
	}
	if l.checkedPredicateRemainderField(member) {
		t.Fatalf("admitted %s, whose kind has no declaration", missing)
	}

	_, err := lowerSourceAs(t, ruling12BSource(t, "b-mapped-kind"), ".ts")
	if err == nil || !strings.Contains(err.Error(), "checked predicate membership over an open tag domain") {
		t.Fatalf("want the open tag domain refusal, got %v", err)
	}
}

// Ruling 12 B accepts plain intersection members in a closed remainder.
// Main's ruling 12 A never road takes only plain members, so an intersection
// that passes the per-part field check must still keep main's refusal there.
func TestClosedIntersectionNeverRemainderKeepsMainsRefusal(t *testing.T) {
	t.Parallel()
	source := ruling12BSource(t, "b-empty")
	withAccessChecker(t, source, func(l *lowering, statements []*ast.Node) {
		read := namedStatement(statements, ast.KindFunctionDeclaration, "read")
		if read == nil {
			t.Fatal("missing read function")
		}
		var call *ast.CallExpression
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsCallExpression(node) && sourceExpression(node) == "isLiteralAccess(expr)" {
				call = node.AsCallExpression()
				return true
			}
			return node.ForEachChild(visit)
		}
		read.ForEachChild(visit)
		if call == nil {
			t.Fatal("missing isLiteralAccess(expr) call")
		}
		argument := call.Arguments.Nodes[0]
		incoming := l.checker.GetTypeAtLocation(argument)
		if incoming.Flags()&checker.TypeFlagsIntersection == 0 || !l.checkedPredicateRemainderField(incoming) {
			t.Fatalf("want an intersection that passes the per-part field check, got %s", l.checker.TypeToString(incoming))
		}
		if tag := l.fieldLiteral(incoming, "kind"); tag == nil || l.tagHoldsAnyNumber(tag) {
			t.Fatal("want a closed literal kind on LiteralAccess")
		}
		flow := &ast.FlowNode{Flags: ast.FlowFlagsFalseCondition, Node: call.AsNode(), Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
		remainder := l.concrete(predicateFlowType(l.checker, argument, incoming, incoming, nil, flow))
		if remainder.Flags()&checker.TypeFlagsNever == 0 {
			t.Fatalf("want a never false remainder, got %s", l.checker.TypeToString(remainder))
		}
		if got := l.checkedPredicateClosedRemainder(call, 0); got != nil {
			t.Errorf("the empty-remainder intersection guard must refuse LiteralAccess, got %s", l.checker.TypeToString(got))
		}
	})

	_, err := lowerSourceAs(t, source, ".ts")
	if err == nil || !strings.Contains(err.Error(), "checked predicate membership over an open tag domain") {
		t.Fatalf("want the open tag domain refusal, got %v", err)
	}
}
