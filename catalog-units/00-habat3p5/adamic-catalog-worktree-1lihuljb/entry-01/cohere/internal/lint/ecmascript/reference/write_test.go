package reference_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// answerFor drives WritesToBinding over every occurrence of `target` in a source and returns the
// answers in source order, so a test asserts what the function said about real parsed nodes rather
// than about a synthesized tree.
func answerFor(t *testing.T, source string, target string) []bool {
	t.Helper()
	var answers []bool
	rule_testing.Run(t, rule.Rule{
		Name: "write-probe",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindIdentifier: func(node *ast.Node) {
					if node.Text() != target {
						return
					}
					// The declaration's own name is not an occurrence anyone asks about.
					if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
						return
					}
					answers = append(answers, reference.WritesToBinding(node))
				},
			}
		},
	}, "probe.ts", source)
	return answers
}

func expectOne(t *testing.T, source string, want bool) {
	t.Helper()
	got := answerFor(t, "let a, b, c, d, o, xs, foo;\n"+source, "a")
	if len(got) != 1 {
		t.Fatalf("%q: expected exactly one occurrence of `a`, got %d", source, len(got))
	}
	if got[0] != want {
		t.Errorf("%q: WritesToBinding = %v, want %v", source, got[0], want)
	}
}

// TestPlainWrites covers the shapes ast.IsWriteAccess already gets right, so a regression in the
// pass-through arms shows up as one of these flipping rather than as silence.
func TestPlainWrites(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"a = 1;",
		"a += 1;",
		"a ??= 1;",
		"a ||= 1;",
		"a &&= 1;",
		"a++;",
		"++a;",
		"a--;",
		"--a;",
		"for (a of xs) {}",
		"for (a in xs) {}",
		"[a] = xs;",
		"({a} = o);",
		"({b: a} = o);",
		"(a) = 1;",
		"((a)) = 1;",
	} {
		expectOne(t, source, true)
	}
}

// TestReads covers occurrences that name the binding without assigning to it. These are what keep
// the unguarded climb honest: each one walks through a pass-through wrapper and must still answer
// false, because the climb ends at something that is not an assignment.
func TestReads(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"foo(a);",
		"b = a;",
		"a.x = 1;",
		"foo(...a);",
		"const xs2 = [...a];",
		"const o2 = {...a};",
		"foo([a]);",
		"foo({a});",
		"typeof a;",
		"delete a.b;",
		"for (const k in a) {}",
		"for (const k of a) {}",
		// The default-value position, which upstream carries as a clean case for no-const-assign.
		"({ files = a } = o);",
		// A computed key names the property to take rather than the binding to write.
		"({[a]: b} = o);",
		// A key merely names a property.
		"({a: b} = o);",
		// The NAME side of a property access. `o.a = 1` writes to a property of `o`; the identifier
		// `a` names a member and no binding called `a` is touched. Note `a.x = 1` above covers the
		// object side, which is a read; this is the other half and it was missing.
		//
		// `ast.IsWriteAccess` answers true here because the access as a whole is a write target, so
		// without the guard every rule in this family reports any property write whose member name
		// collides with a binding it watches. Measured: `no-global-assign` produced 625 findings on
		// our own tree against oxlint's 0, nearly all of them `document.cookie = ...`.
		"o.a = 1;",
		"o.a += 1;",
		"foo.bar.a = 1;",
		"this.a = 1;",
	} {
		expectOne(t, source, false)
	}
}

// TestRestElementsInAssignmentTargets is the measured gap in ast.IsWriteAccess.
//
// Every one of these is a real write that the shelf accessor classifies as a read, because its
// accessKind switch has no arm for KindSpreadElement or KindSpreadAssignment. Three rules measured
// this independently and each wrote its own workaround; this is what the shared version has to get
// right for all of them.
func TestRestElementsInAssignmentTargets(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"[...a] = xs;",
		"({...a} = o);",
		"[b, ...a] = xs;",
		"({b, ...a} = o);",
		"[[...a]] = xs;",
		"[{...a}] = xs;",
		"({x: [...a]} = o);",
		"({x: {...a}} = o);",
		"for ([...a] of xs) {}",
		"for ({...a} of xs) {}",
		// The deepest case either upstream corpus carries: `a` sits under spread, array, spread,
		// array before reaching the assignment. A climb that walked only spreads stopped at the
		// inner array literal and got that literal's answer, which is false precisely because it is
		// reached through a spread.
		"[b, c, ...[d, ...a]] = [1,2,3,4,5];",
		"[, {foo: b, ...a}] = foo();",
	} {
		expectOne(t, source, true)
	}
}

// TestParenthesizedRestTargets is the disagreement between the three lifted implementations.
//
// `no-class-assign` answers true for both; `no-const-assign` and `prefer-const` answer false,
// because each gates its climb on the immediate parent being a spread or a literal and a
// parenthesis is neither. Node confirms both are real writes: `[...(a)] = [1,2]` leaves a as
// [1,2], and `({...(a)} = o)` leaves a holding the whole object.
//
// The shared version takes no-class-assign's answer, which is the union and the correct one.
func TestParenthesizedRestTargets(t *testing.T) {
	t.Parallel()
	expectOne(t, "[...(a)] = xs;", true)
	expectOne(t, "({...(a)} = o);", true)
	// A parenthesis further up the chain, which no implementation got wrong but which the
	// pass-through arm has to keep answering.
	expectOne(t, "([...a]) = xs;", true)
}

// TestSpreadOutsideAnAssignmentStaysARead pins the other direction of the unguarded climb. A
// spread in a value position walks the same wrappers and must still answer false, because the climb
// reaches a call or a declaration rather than an assignment operator.
func TestSpreadOutsideAnAssignmentStaysARead(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"foo(...[...a]);",
		"const xs2 = [b, ...a];",
		"const o2 = {b, ...a};",
		"foo({x: [...a]});",
		"const o3 = {x: {...a}};",
	} {
		expectOne(t, source, false)
	}
}

// TestNilIsNotAWrite covers the guard a shared function needs and a rule-local one did not: inside
// a rule the identifier always came from a walk, and on a shelf any caller can pass anything.
func TestNilIsNotAWrite(t *testing.T) {
	t.Parallel()
	if reference.WritesToBinding(nil) {
		t.Error("WritesToBinding(nil) = true, want false")
	}
}

// TestShorthandPropertyPositions pins the two slots of a shorthand property, which is the one place
// this function deliberately answers differently from the six-rule implementation it replaces.
//
// A shorthand carries the identifier in Name() or in ObjectAssignmentInitializer, and the two mean
// opposite things. The six rules list the shorthand as a transparent wrapper with no position test
// and therefore call the default-value read a write; ast.IsWriteAccess distinguishes them, so this
// function gets both right by NOT listing the shorthand as a wrapper.
//
// Verified with node: `let a = "ORIGINAL"; ({ files = a } = {})` leaves a untouched. It is also
// upstream's own clean case for no-const-assign.
func TestShorthandPropertyPositions(t *testing.T) {
	t.Parallel()
	// Name() position: the binding is what gets written.
	expectOne(t, "({a} = o);", true)
	// ObjectAssignmentInitializer position: the binding supplies a default and is only read.
	expectOne(t, "({ files = a } = o);", false)
	// Both positions in one statement, so the two answers are produced by one parse.
	got := answerFor(t, "let a, b, o;\n({ a, files = b } = o);", "a")
	if len(got) != 1 || !got[0] {
		t.Errorf("({ a, files = b } = o): write position answered %v, want [true]", got)
	}
	got = answerFor(t, "let a, b, o;\n({ a, files = b } = o);", "b")
	if len(got) != 1 || got[0] {
		t.Errorf("({ a, files = b } = o): default position answered %v, want [false]", got)
	}
}

// TestDefaultValuesInOtherPatterns covers the same read-not-write question in the array and nested
// forms, so the discrimination is not pinned only to the object shorthand.
func TestDefaultValuesInOtherPatterns(t *testing.T) {
	t.Parallel()
	// An array element default. `a` supplies the fallback and is read.
	expectOne(t, "[b = a] = xs;", false)
	// A property-value default.
	expectOne(t, "({x: b = a} = o);", false)
	// The written binding in the same statement is still a write.
	got := answerFor(t, "let a, b, o, xs;\n[a = b] = xs;", "a")
	if len(got) != 1 || !got[0] {
		t.Errorf("[a = b] = xs: target answered %v, want [true]", got)
	}
}

// TestIsUpdateOperator covers the two operators that write back and a sample of those that do not.
func TestIsUpdateOperator(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		operator ast.Kind
		want     bool
	}{
		{ast.KindPlusPlusToken, true},
		{ast.KindMinusMinusToken, true},
		{ast.KindMinusToken, false},
		{ast.KindPlusToken, false},
		{ast.KindExclamationToken, false},
		{ast.KindTildeToken, false},
		{ast.KindTypeOfKeyword, false},
	} {
		if got := reference.IsUpdateOperator(c.operator); got != c.want {
			t.Errorf("IsUpdateOperator(kind %d) = %v, want %v", int(c.operator), got, c.want)
		}
	}
}

// TestShelfAccessorStillDeclinesRestElements is a differential guard on the shim rather than on this
// package, and it is the reason the climb above exists.
//
// `ast.IsWriteAccess` classifies a rest element in a destructuring assignment target as a read,
// because typescript-go's `accessKind` switch has no arm for `KindSpreadElement` or
// `KindSpreadAssignment`. Three rules measured that independently and each wrote a different
// workaround before the union was lifted here.
//
// If a future vendored compiler closes that gap, this test fails, and the failure is the useful
// signal: the climb becomes redundant and should be deleted rather than left as code nobody can
// justify. Without it the climb would sit there forever, correct but unexplainable, and the next
// reader would have no way to learn whether it still does anything.
//
// The assertion is deliberately on the SHIM's answer rather than on ours. Asserting our answer would
// pass either way and prove nothing about why the wrapper exists.
func TestShelfAccessorStillDeclinesRestElements(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"[...a] = xs;",
		"({...a} = o);",
		"[b, ...a] = xs;",
		"[b, c, ...[d, ...a]] = [1,2,3,4,5];",
	} {
		full := "let a, b, c, d, o, xs;\n" + source
		var shimAnswers []bool
		rule_testing.Run(t, rule.Rule{
			Name: "shim-differential",
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindIdentifier: func(node *ast.Node) {
						if node.Text() != "a" {
							return
						}
						if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
							return
						}
						shimAnswers = append(shimAnswers, ast.IsWriteAccess(node))
					},
				}
			},
		}, "probe.ts", full)

		if len(shimAnswers) != 1 {
			t.Fatalf("%q: expected one occurrence of `a`, got %d", source, len(shimAnswers))
		}
		if shimAnswers[0] {
			t.Errorf("%q: ast.IsWriteAccess now reports a rest element as a write. The gap this "+
				"package's climb exists to close has been fixed upstream: delete the spread and "+
				"literal arms of WritesToBinding and this test with them.", source)
		}
		// Ours must answer true whatever the shim says, which is the whole point of the wrapper.
		expectOne(t, source, true)
	}
}
