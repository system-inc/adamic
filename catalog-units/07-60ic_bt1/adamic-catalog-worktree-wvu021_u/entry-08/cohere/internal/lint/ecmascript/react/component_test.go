package react

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// firstNodeOfKind parses source text and returns the first node of the given kind.
func firstNodeOfKind(t *testing.T, sourceText string, kinds ...ast.Kind) *ast.Node {
	t.Helper()
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/source/Thing.tsx",
		PathKey:  "/repository/source/Thing.tsx",
	}, sourceText, core.ScriptKindTSX)
	if file == nil {
		t.Fatal("the parser returned no source file")
	}

	wanted := make(map[ast.Kind]bool, len(kinds))
	for _, kind := range kinds {
		wanted[kind] = true
	}

	var found *ast.Node
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if found != nil || node == nil {
			return true
		}
		if wanted[node.Kind] {
			found = node
			return true
		}
		node.ForEachChild(walk)
		return found != nil
	}
	file.AsNode().ForEachChild(walk)

	if found == nil {
		t.Fatalf("no node of the wanted kind in %q", sourceText)
	}
	return found
}

// The object is deliberately not required to be React, because a file importing the function
// directly writes a bare call and that is the common modern spelling. The one exception is
// document, which shares the property name and constructs a DOM node instead.
func TestIsCreateElementCall(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		want       bool
	}{
		{"a bare call", "export const A = createElement('div');\n", true},
		{"the React namespace", "export const A = React.createElement('div');\n", true},
		{"another namespace still counts", "export const A = Preact.createElement('div');\n", true},
		{"a computed member", "export const A = React['createElement']('div');\n", true},
		{"parentheses around the object", "export const A = (React).createElement('div');\n", true},
		// The discriminating case. It shares the property name and is not a React element.
		{"document is rejected", "export const A = document.createElement('div');\n", false},
		{"document computed is rejected", "export const A = document['createElement']('div');\n", false},
		{"a different property", "export const A = React.cloneElement(child);\n", false},
		{"a different bare name", "export const A = createFragment('div');\n", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			node := firstNodeOfKind(t, testCase.sourceText, ast.KindCallExpression)
			if got := IsCreateElementCall(node); got != testCase.want {
				t.Fatalf("want %v, got %v", testCase.want, got)
			}
		})
	}
}

// Both factory spellings count, and the namespaced form checks the object is React specifically,
// because Foo.createClass is somebody else's factory.
func TestIsEs5ComponentCall(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		want       bool
	}{
		{"the standalone package", "export const A = createReactClass({});\n", true},
		{"the React method", "export const A = React.createClass({});\n", true},
		{"a bare createClass", "export const A = createClass({});\n", true},
		{"another namespace is rejected", "export const A = Foo.createClass({});\n", false},
		{"an unrelated call", "export const A = makeThing({});\n", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			node := firstNodeOfKind(t, testCase.sourceText, ast.KindCallExpression)
			if got := IsEs5ComponentCall(node); got != testCase.want {
				t.Fatalf("want %v, got %v", testCase.want, got)
			}
		})
	}
}

// A class with no extends clause is the case this gate mostly exists to stay silent on: it is 2 of
// the 8 passing fixtures upstream ships for no-direct-mutation-state.
func TestIsEs6ComponentClass(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		want       bool
	}{
		{"extends React.Component", "export class Thing extends React.Component {}\n", true},
		{"extends React.PureComponent", "export class Thing extends React.PureComponent {}\n", true},
		{"extends a bare Component", "export class Thing extends Component {}\n", true},
		{"extends a bare PureComponent", "export class Thing extends PureComponent {}\n", true},
		{"a class expression", "export const Thing = class extends React.Component {};\n", true},
		// The discriminating case.
		{"no extends clause", "export class Hello {}\n", false},
		{"extends something else", "export class Thing extends Base {}\n", false},
		{"extends another namespace", "export class Thing extends Foo.Component {}\n", false},
		{"implements rather than extends", "export class Thing implements Component {}\n", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			node := firstNodeOfKind(t, testCase.sourceText,
				ast.KindClassDeclaration, ast.KindClassExpression)
			if got := IsEs6ComponentClass(node); got != testCase.want {
				t.Fatalf("want %v, got %v", testCase.want, got)
			}
		})
	}
}

// Both definitions are searched in one walk, so the answer does not depend on which era wrote the
// component. The node itself is considered, so a rule that already matched a class can ask without
// stepping to the parent.
func TestEnclosingComponent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		startKind  ast.Kind
		wantFound  bool
	}{
		{
			"a method inside a class component",
			"export class Thing extends React.Component {\n    render() {\n        return null;\n    }\n}\n",
			ast.KindMethodDeclaration, true,
		},
		{
			"a method inside a plain class",
			"export class Thing {\n    render() {\n        return null;\n    }\n}\n",
			ast.KindMethodDeclaration, false,
		},
		{
			"a function inside an es5 component",
			"export const A = createReactClass({\n    render: function() {\n        return null;\n    },\n});\n",
			ast.KindFunctionExpression, true,
		},
		{
			"a free function",
			"export function run() {\n    return null;\n}\n",
			ast.KindFunctionDeclaration, false,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			node := firstNodeOfKind(t, testCase.sourceText, testCase.startKind)
			got := EnclosingComponent(node)
			if (got != nil) != testCase.wantFound {
				t.Fatalf("want found=%v, got %v", testCase.wantFound, got != nil)
			}
		})
	}
}

// A rule reaching a shared helper with a node it did not check is a bug in the rule, and a panic in
// a shared package takes the whole run down rather than that one rule's finding.
func TestHelpersSurviveNilAndWrongKinds(t *testing.T) {
	t.Parallel()
	if IsCreateElementCall(nil) || IsEs5ComponentCall(nil) || IsEs6ComponentClass(nil) {
		t.Fatal("want nil to answer false everywhere")
	}
	if EnclosingComponent(nil) != nil {
		t.Fatal("want nil to have no enclosing component")
	}

	// A node of the wrong kind answers rather than panics.
	identifier := firstNodeOfKind(t, "export const A = value;\n", ast.KindIdentifier)
	if IsCreateElementCall(identifier) || IsEs5ComponentCall(identifier) || IsEs6ComponentClass(identifier) {
		t.Fatal("want an identifier to answer false everywhere")
	}
}

// TestIsNamespacedMember covers the shape four rules were matching by hand.
func TestIsNamespacedMember(t *testing.T) {
	t.Parallel()
	anyName := func(string) bool { return true }
	named := func(want string) func(string) bool {
		return func(got string) bool { return got == want }
	}

	for _, testCase := range []struct {
		name       string
		sourceText string
		matches    func(string) bool
		want       bool
	}{
		{"a namespaced member", "React.useEffect;", named("useEffect"), true},
		// Parentheses on the receiver are skipped, which is the one place the four lifted copies
		// differed. `(React).useEffect(...)` is the same call, and a check on the raw node declines
		// it while looking correct.
		{"a parenthesized receiver", "(React).useEffect;", named("useEffect"), true},
		{"a doubly parenthesized receiver", "((React)).useEffect;", named("useEffect"), true},
		// The receiver must be React. A copy accepting any namespace exempted aliases the gate still
		// reports, measured in consistency-no-property-alias.
		{"another namespace", "Other.useEffect;", anyName, false},
		// The predicate decides the name, so a member React does have but the caller does not want
		// still declines.
		{"a name the predicate rejects", "React.useMemo;", named("useEffect"), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			node := firstNodeOfKind(t, testCase.sourceText, ast.KindPropertyAccessExpression)
			if got := IsNamespacedMember(node, testCase.matches); got != testCase.want {
				t.Errorf("%q: IsNamespacedMember = %v, want %v", testCase.sourceText, got, testCase.want)
			}
		})
	}
}

// TestIsNamespacedMemberDeclinesOtherKinds covers the kind guard the lifted copy did not have.
//
// Inside its rule the node always arrived from a matched switch arm, so the guard was unnecessary
// there; on a shelf any caller can pass anything, and the local version dereferences without
// checking.
func TestIsNamespacedMemberDeclinesOtherKinds(t *testing.T) {
	t.Parallel()
	anyName := func(string) bool { return true }

	if IsNamespacedMember(nil, anyName) {
		t.Error("IsNamespacedMember(nil) = true, want false")
	}

	// A computed member carries no identifier name to match, and it is a different node kind.
	elementAccess := firstNodeOfKind(t, "React['useEffect'];", ast.KindElementAccessExpression)
	if IsNamespacedMember(elementAccess, anyName) {
		t.Error("a computed member answered true")
	}

	// A call expression is the kind the four lifted copies were handed after their own switch had
	// already narrowed it. Passing one directly must decline rather than panic.
	call := firstNodeOfKind(t, "f();", ast.KindCallExpression)
	if IsNamespacedMember(call, anyName) {
		t.Error("a call expression answered true")
	}
}
