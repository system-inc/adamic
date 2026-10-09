package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoClassAssign = rule.Message{
	Id: "noClassAssign",
	Description: "This assigns to a class binding. The binding is const-like inside the class body " +
		"and the assignment throws in strict mode, which every module and every class body already " +
		"is, so the line fails at runtime rather than rebinding anything. Where it does not throw " +
		"it is worse: every later use of the name gets the new value instead of the class, so " +
		"`new A()` constructs something else and `instanceof A` answers about something else. Use a " +
		"separate variable for the new value.",
}

// NoClassAssign flags an assignment to the name a class declares.
//
//	valid:   class A { } foo(A);
//	valid:   class A { b(A) { A = 0; } }
//	valid:   let A = class A { }; A = 1;
//	valid:   class A { } A.x = 0;
//	invalid: class A { } A = 0;
//	invalid: A = 0; class A { }
//	invalid: class A { b() { A = 0; } }
//	invalid: let A = class A { b() { A = 0; } }
//
// A class declaration binds its name like a `let`, so the assignment parses and the mistake is
// silent at the line that makes it. What it costs shows up somewhere else: the name now holds
// whatever was assigned, so `new A()` constructs that instead, `extends A` extends that instead, and
// `instanceof A` answers a question about that instead. Inside the class body the binding is
// immutable and the write throws outright.
//
// # Why this reads the checker
//
// The discrimination is entirely name resolution and none of it is structural. Four of upstream's
// clean cases write to a name spelled exactly like the class:
//
//	class A { b(A) { A = 0; } }              a parameter shadows it
//	class A { b() { let A; A = 0; } }        a local shadows it
//	let A = class A { }; A = 1;              the variable, not the class name
//	if (foo) { class A {} } else { class A {} } A = 1;   binds to neither
//
// Their text is identical to the failing forms, so a rule matching on the name reports all four.
// Upstream answers with `get_resolved_references`, which is a find-all-references index we do not
// have. The equivalent question our checker does answer is which declaration a given identifier
// binds to, so this anchors on the class node the listener already received and asks, for each
// write, whether the resolved symbol's first declaration is that same node.
//
// Node identity rather than declaration kind. Comparing kinds looks like it would work here, since
// the anchor is a class and a shadow is usually a parameter or a variable, but it is a near miss
// that the if/else case defeats outright: both siblings are `KindClassDeclaration`, so a kind test
// calls the second one's declaration a match for the first and reports a write that binds to
// neither. The same trap is sharper for the sibling rules on `const`, where a `const A` and a
// shadowing `let A` are both `KindVariableDeclaration` and kind cannot separate them at all.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names; it says nothing about whether the
// occurrence writes. `class A { } A.x = 0;` and `class A { } foo(A);` both resolve to the class and
// neither reassigns it, so the rule needs the structural half too and reports only where the two
// agree. `reference.WritesToBinding` is that half, shared with five sibling rules.
//
// # The one shape the checker answers differently
//
// `({A} = 0)` is a shorthand property in an object literal being used as a destructuring target, and
// `GetSymbolAtLocation` on its identifier returns the property's own symbol rather than the value
// being written. TypeScript has a separate accessor for the value side, and reaching for it is the
// difference between reporting this case and silently missing it. Measured on the corpus rather than
// assumed: the plain accessor answers "different declaration" here, which reads exactly like a
// correctly declined shadow.
//
// No fix. The repair is a new binding, which means choosing a name and deciding which later uses of
// the old one meant the class, and the rule cannot know that.
var NoClassAssign = rule.Rule{
	Name: "no-class-assign",

	// See the doc above: four of upstream's clean cases are textually identical to failing ones and
	// differ only in what the name resolves to.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		reportWritesTo := func(node *ast.Node) {
			// The engine hands every rule a nil checker when the program could not be built, and
			// this rule can answer nothing without one. Removing this survives the fixtures,
			// because typescript-go tolerates a nil receiver on this path and returns no symbol;
			// that is an implementation detail of a vendored compiler rather than a promise, so the
			// guard stays and the rule does not depend on it.
			if ctx.TypeChecker == nil {
				return
			}

			// An anonymous class expression declares no name, so there is nothing to reassign.
			// `let A = class { b() { A = 0; } }` writes to the variable, and upstream leaves it.
			name := node.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}

			declaration := declarationAnchoredAt(ctx, name)
			if declaration == nil {
				return
			}

			// The whole file, not the class subtree. A write can sit before the declaration
			// (`A = 0; class A { }`), after it, or inside it, and hoisting makes all three the same
			// binding. Anchoring the search on the file and the match on the symbol is what makes
			// the position of the write irrelevant.
			sourceFile := ast.GetSourceFileOfNode(node)
			if sourceFile == nil {
				return
			}

			var visit func(*ast.Node)
			visit = func(current *ast.Node) {
				if current == nil {
					return
				}
				// The class's own name needs no exclusion here. It is an identifier whose text
				// matches and which resolves to this very declaration, so only the structural half
				// declines it, and that is enough: the climb from a class name reaches the class
				// itself, which is not an assignment. A mutant removing an explicit guard for it
				// survived every fixture, which is what identified the guard as redundant rather
				// than as untested.
				//
				// The text comparison is a pre-filter and not a discrimination. Symbol identity
				// already implies it, since an identifier spelled differently cannot resolve to
				// this declaration. It is here because it is far cheaper than a checker call and
				// this walk visits every identifier in the file; a mutant removing it survives, and
				// correctly so.
				if current.Kind == ast.KindIdentifier &&
					current.Text() == name.Text() &&
					reference.WritesToBinding(current) &&
					resolvesToDeclaration(ctx, current, declaration) {
					ctx.ReportNode(current, messageNoClassAssign)
				}
				current.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(sourceFile.AsNode())
		}

		// Both kinds, and they behave differently on purpose. A declaration's name is visible to the
		// enclosing scope, so a write anywhere in the file can reach it. A class expression's name
		// binds only inside its own body, so `let A = class A { }; A = 1;` writes to the variable
		// and is clean while `let A = class A { b() { A = 0; } }` writes to the class name and is
		// not. Nothing here encodes that split: the checker resolves each write and the two cases
		// fall out of the same code.
		return rule.Listeners{
			ast.KindClassDeclaration: reportWritesTo,
			ast.KindClassExpression:  reportWritesTo,
		}
	},
}

// declarationAnchoredAt reads the declaration node a class's own name binds to.
//
// This is the identity the writes are compared against, and it comes from the checker rather than
// from the class node directly so that both sides of the later comparison are answers to the same
// question. Asking the AST for one side and the checker for the other would compare two things that
// happen to agree today.
func declarationAnchoredAt(ctx rule.Context, name *ast.Node) *ast.Node {
	// Only this file's declarations, as ESLint's scope manager sees one file: a binding declared in
	// another file is not one this rule can say was reassigned from a class, function or import.
	declarations := rule.DeclarationsIn(ctx.SourceFile, ctx.TypeChecker.GetSymbolAtLocation(name))
	if len(declarations) == 0 {
		return nil
	}
	return declarations[0]
}

// resolvesToDeclaration reports whether an identifier binds to a specific declaration node.
//
// A shorthand property in a destructuring target resolves to the property rather than to the value,
// so that shape is asked through the accessor built for it. `({A} = 0)` is upstream's second failing
// case and the plain accessor gets it wrong in the quiet direction, answering "some other
// declaration" exactly as a real shadow does.
//
// An identifier the checker cannot resolve answers false. `(class A {}, A = 1)` and the if/else pair
// both land there, and both are clean upstream: a name nothing declares is not provably this class,
// and reporting it would be a guess in the direction that costs a false positive.
func resolvesToDeclaration(ctx rule.Context, identifier *ast.Node, declaration *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}
	// The anchor is a declaration in this file, so only this file's declarations can be it.
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	return len(declarations) > 0 && declarations[0] == declaration
}
