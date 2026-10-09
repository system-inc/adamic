package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoImportAssign = rule.Message{
	Id: "noImportAssign",
	Description: "This writes to a binding an import declaration introduced. An imported binding is " +
		"an immutable view of the exporting module's own variable, so the assignment throws a " +
		"TypeError at runtime rather than rebinding anything, and a module is always strict so there " +
		"is no mode in which it fails quietly. A namespace object is sealed the same way, so writing " +
		"to one of its properties, deleting one, or handing it to Object.assign fails for the same " +
		"reason. Change the value in the module that exports it, or copy it into a local variable and " +
		"write to that.",
}

// NoImportAssign flags a write to a name an import declaration bound.
//
//	valid:   import mod from 'mod'; mod.prop = 0
//	valid:   import * as mod from 'mod'; mod.named.prop = 0
//	valid:   import mod from 'mod'; { let mod = 0; mod = 1 }
//	valid:   import * as mod from 'mod'; Object.assign(mod.prop, obj)
//	valid:   import * as mod from 'mod'; var Object; Object.assign(mod, obj)
//	invalid: import mod from 'mod'; mod = 0
//	invalid: import {named} from 'mod'; [named] = foo
//	invalid: import * as mod from 'mod'; mod.named = 0
//	invalid: import * as mod from 'mod'; Object.assign(mod, obj)
//	invalid: import * as mod from 'mod'; delete mod?.prop
//
// An import binding is not a variable holding a copy. It is a live, immutable view of the exporting
// module's own binding, so the assignment has nowhere to write. Module code is always strict, so
// there is no sloppy mode in which the write is merely ignored: it throws a TypeError at the line
// that makes it. TypeScript catches most of these already, which is the reason this rule is worth
// less in a typed tree than in a plain one; it does not catch the `Object.assign` family, which is
// the case this rule is here for.
//
// # Three shapes, and only the first is a rebinding
//
// Rebinding the local name (`mod = 0`) fails for any import. Writing to a property of a namespace
// object (`mod.named = 0`) fails only for `import * as`, because only that shape produces a module
// namespace exotic object, which is sealed and whose properties are non-configurable and
// non-writable. A default or named import is an ordinary value, so `mod.prop = 0` mutates whatever
// the module exported and is upstream's very first clean case. Passing a namespace object as the
// first argument of a well-known mutation function (`Object.assign(mod, obj)`) is the same seal
// reached through a different door, and it is the one shape a type checker does not catch, since the
// signature of `Object.assign` accepts it happily.
//
// # Why this reads the checker
//
// The discrimination is name resolution, and three of upstream's clean cases are written to prove
// it. `import mod from 'mod'; { let mod = 0; mod = 1 }` writes to a local spelled exactly like the
// import; so does the namespace form, and so does its member variant. Their text is identical to the
// failing shapes and only the binding differs, so a rule matching on the name reports all three.
//
// Upstream answers with `get_resolved_references`, a find-all-references index we do not have. The
// question our checker does answer is which declaration a given identifier binds to, so this anchors
// on the import declaration the listener already received and asks, for each write in the file,
// whether the resolved symbol's first declaration is one of that declaration's own specifier nodes.
//
// Node identity rather than declaration kind. A shadow inside a block is a `KindVariableDeclaration`
// and the anchor is a `KindImportSpecifier`, so kind looks sufficient here and is not: two imports
// in the same file both declare at `KindImportSpecifier`, and a kind test would let a write to one
// be attributed to the other. `import {a} from 'x'; import {b} from 'y'; b = 0` would then report
// twice, once per declaration, where upstream reports once. Identity separates them and kind cannot.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names and nothing about whether the occurrence
// writes. `mod.prop = 0` and `foo(mod)` both resolve to the import and neither rebinds it, so the
// rule needs the structural half too. For the plain rebinding that is a straight conjunction and
// `reference.WritesToBinding` is the structural half, shared with five sibling rules.
//
// The namespace shapes are where the two halves are not a simple conjunction, which is the thing
// that makes this rule bigger than its sibling. `mod.named = 0` does not write to `mod` by that
// test, and correctly so, since the same test has to keep `mod.prop = 0` clean for a default import.
// The write being flagged is a write to a property of the thing the identifier names, so the
// structural question is asked one level up: is the member expression whose object is this
// identifier the target of an assignment, an update, a delete, or a for-in/of head. That is
// `writesThroughMemberExpression`, and it is a different question from `reference.WritesToBinding`
// rather than a variant of it.
//
// # What upstream misses and this reproduces
//
// `import * as mod from 'mod'; obj[mod] = 0` is clean upstream and clean here: `mod` is the computed
// key, not the object, so nothing about the namespace is written. So is
// `Object.assign(obj, mod, other)`, since only the first argument is mutated. Both are upstream
// telling a porter which near misses it decided against, and both have a fixture.
//
// The one place this port is narrower than the corpus suggests: upstream's `Object` guard asks
// whether the name has any binding at all, which is a scope question. This asks the checker where
// `Object` is declared and requires the declaration to come from a library file rather than from the
// file being linted. Measured on upstream's own two clean cases for it, which shadow `Object` with a
// `var` in and out of a block, and both stay clean.
//
// No fix. The repair is to change the exporting module or to introduce a local, and the rule cannot
// know which was meant.
var NoImportAssign = rule.Rule{
	Name: "no-import-assign",

	// See the doc above: three of upstream's clean cases shadow an imported name with a local and are
	// textually identical to failing forms.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Each import declaration on the shared walk, including one inside a `declare module` block,
		// which can hold its own. The writes to its names are read from the file's identifier index.
		//
		// Each declaration once walked the whole file for writes to its names (198ms on ahra). Then
		// every import was anchored from the file node and the file walked once against all of them,
		// still 57ms of rule CPU on a cold ahra run (#fcac58b). The index is built once per file for
		// every rule that asks, so each name's writes are read from the few identifiers spelled like
		// it, and the rule walks nothing itself.
		//
		// The anchors already judged, so that two import declarations naming one binding (a parse
		// that recovers from the duplicate) report each write once.
		judged := map[*ast.Node]bool{}
		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built, and
				// this rule can answer nothing without one. A mutant removing this survives the
				// fixtures, because typescript-go tolerates a nil receiver on this path and returns no
				// symbol; that is an implementation detail of a vendored compiler rather than a
				// promise, so the guard stays and the rule does not depend on it.
				if ctx.TypeChecker == nil {
					return
				}
				bindings := imports.BindingsOf(node)
				reportWritesToImport(ctx, bindings.Default, false, judged)
				// `import { a as b }` binds `b`, and `Name()` on the specifier is the local name rather
				// than the imported one. Upstream's `named12 as foo` case proves it matters: `foo = 0`
				// reports and `named12 = 0` does not, because `named12` binds nothing.
				for _, specifier := range bindings.Named {
					reportWritesToImport(ctx, specifier.Name(), false, judged)
				}
				if bindings.Namespace != nil {
					reportWritesToImport(ctx, bindings.Namespace.Name(), true, judged)
				}
			},
		}
	},
}

// reportWritesToImport reports every write to the binding one imported name creates.
//
// `import 'mod'` and `import {} from 'mod'`, both clean upstream, name nothing and so judge nothing.
//
// The member and mutation-function shapes apply only to `import * as`. A default or named import is
// an ordinary value and mutating it is upstream's first ten clean cases.
func reportWritesToImport(ctx rule.Context, name *ast.Node, isNamespace bool, judged map[*ast.Node]bool) {
	if name == nil {
		return
	}
	// Anchored on the declaration its own symbol reports, so both sides of the comparison are answers
	// to the same question. Asking the AST for one side and the checker for the other would compare
	// two things that happen to agree today.
	declaration := declarationAnchoredAt(ctx, name)
	if declaration == nil || judged[declaration] {
		return
	}
	judged[declaration] = true

	for _, occurrence := range reference.IdentifiersNamed(ctx, name.Text()) {
		// The structural question first, and the checker only for an identifier that already sits
		// somewhere a write happens. Both orders report the same findings, and the order matters for
		// what the rule costs: this rule anchors on import declarations, so it runs in nearly every
		// file in a real tree, while the structural tests are pure AST and the checker call takes a
		// per-file lock.
		//
		// `blame` is the node to report, which differs by shape: a rebinding points at the identifier
		// and a namespace property write points at the whole member expression, both matching
		// upstream's snapshot.
		var blame *ast.Node
		switch {
		case reference.WritesToBinding(occurrence):
			blame = occurrence
		case !isNamespace:
			// The namespace-only shapes apply to nothing else. `import mod from 'mod'; mod.prop = 0`
			// reaches here and is upstream's first clean case.
		default:
			if member := writesThroughMemberExpression(occurrence); member != nil {
				blame = member
			} else if isArgumentOfWellKnownMutationFunction(ctx, occurrence) {
				blame = occurrence
			}
		}
		if blame != nil && resolvedDeclarationOf(ctx, occurrence) == declaration {
			ctx.ReportNode(blame, messageNoImportAssign)
		}
	}
}

// resolvedDeclarationOf reads the declaration node an identifier occurrence binds to.
//
// A shorthand property in a destructuring target resolves through the plain accessor to the
// property's own symbol rather than to the value being written, which reads exactly like a correctly
// declined shadow. `({named} = obj)` on an imported `named` is a real failing shape, so that case is
// asked through the accessor built for it.
func resolvedDeclarationOf(ctx rule.Context, identifier *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}
	// An import binding is declared in this file, so only this file's declarations can be one.
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if len(declarations) == 0 {
		return nil
	}
	return declarations[0]
}

// writesThroughMemberExpression returns the member expression through which an identifier's
// properties are written, or nil.
//
// This is the namespace half of the rule and it is a different question from `reference.WritesToBinding`
// rather than a variant of it: `mod.named = 0` writes to a property of what `mod` names, while `mod`
// itself is only read. Asking the same question one level up is what separates the two, and getting
// that wrong in either direction breaks a clean case rather than a failing one. Widening it to any
// member expression would report `mod.named.prop = 0`, which upstream keeps clean because the write
// lands on `mod.named`'s object and not on the namespace.
//
// The identifier must be the member expression's object. `obj[mod] = 0` puts it in the key position,
// upstream keeps it clean, and there is a fixture.
//
// Returns the member expression because that is the span upstream reports, wider than the identifier
// in every one of these twelve cases.
func writesThroughMemberExpression(identifier *ast.Node) *ast.Node {
	member := identifier.Parent
	if member == nil {
		return nil
	}
	switch member.Kind {
	case ast.KindPropertyAccessExpression:
		// `mod.named` and, with the same node kind, `mod?.named`. Optional chaining does not change
		// what is written, and upstream's `delete mod?.prop` is a failing case.
		if member.AsPropertyAccessExpression().Expression != identifier {
			return nil
		}
	case ast.KindElementAccessExpression:
		// `mod['named'] = 0`. Upstream has no fixture for this and its own code handles it through
		// the computed-member arm, so the shape is deliberate rather than incidental.
		if member.AsElementAccessExpression().Expression != identifier {
			return nil
		}
	default:
		return nil
	}
	if isWriteTarget(member) {
		return member
	}
	return nil
}

// isWriteTarget reports whether a member expression sits where something is written to it.
//
// Split out from `writesThroughMemberExpression` because the member expression is found by one
// question and judged by another, and because `delete` belongs only here. Deleting a binding is a
// syntax error, so `reference.WritesToBinding` has no reason to know the operator; deleting a namespace
// property is upstream's `delete mod12.named` and fails.
//
// The climb through destructuring wrappers mirrors `reference.WritesToBinding` and exists for the same
// reason: `[mod6.named] = foo` and `({ ...mod11.named } = foo)` reach their assignment through an
// array or object literal rather than directly.
func isWriteTarget(member *ast.Node) bool {
	child := member
	for parent := member.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil &&
				ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
				ast.SkipParentheses(binary.Left) == child

		case ast.KindPrefixUnaryExpression:
			unary := parent.AsPrefixUnaryExpression()
			return reference.IsUpdateOperator(unary.Operator) && ast.SkipParentheses(unary.Operand) == child

		case ast.KindPostfixUnaryExpression:
			unary := parent.AsPostfixUnaryExpression()
			return reference.IsUpdateOperator(unary.Operator) && ast.SkipParentheses(unary.Operand) == child

		case ast.KindDeleteExpression:
			// `delete mod.named`. Only reachable from this side of the rule.
			return ast.SkipParentheses(parent.AsDeleteExpression().Expression) == child

		case ast.KindForInStatement, ast.KindForOfStatement:
			// `for (mod.named in foo);` assigns on every iteration. `for (var foo in mod.named);`
			// reads it instead, and upstream keeps that clean; the initializer test is what
			// separates them.
			return parent.AsForInOrOfStatement().Initializer == child

		case ast.KindParenthesizedExpression,
			ast.KindArrayLiteralExpression,
			ast.KindSpreadElement,
			ast.KindSpreadAssignment,
			ast.KindObjectLiteralExpression:
			// The destructuring wrappers. A destructuring assignment target parses as an array or
			// object literal rather than as a pattern, so the question passes through unchanged and
			// the binary-expression arm decides it. That is what keeps
			// `var obj = { ...mod.named };` clean: the same object literal is not assigned to, so
			// the climb ends at a declaration.

		case ast.KindPropertyAssignment:
			// `({ bar: mod9.named } = foo)` writes through the value side. `({ [mod.named]: bar } =
			// foo)` names it as a computed key and upstream keeps that clean, which is exactly the
			// initializer test.
			if parent.AsPropertyAssignment().Initializer != child {
				return false
			}

		default:
			// Anything else ends the climb: a call argument, a further property access
			// (`mod.named.prop`), a plain expression statement. None of them writes.
			return false
		}
		// Advance unwrapped. Every arm above compares against `ast.SkipParentheses(...)` of the
		// parent's own operand, which yields the member expression rather than the parenthesis
		// around it, so carrying a `KindParenthesizedExpression` up as `child` makes every one of
		// those comparisons fail. `(mod.named) = 0` is a write and was read as a read before this
		// line existed, and so were `((mod.named)) = 0`, `(mod['named']) = 0` and
		// `delete (mod.named)`. All four throw a TypeError at runtime, which is the whole subject
		// of this rule.
		//
		// Neither upstream's corpus nor ours had a parenthesized namespace write, which is why it
		// shipped silently: parentheses are a real node here and absent from ESTree, so no imported
		// corpus can contain one. `no-class-assign` hit the same defect on a plain assignment
		// target, fixed it the same way, and recorded it at its own line; that climb has since been
		// replaced by `reference.WritesToBinding`, which walks parentheses, so this was the last
		// hand-rolled climb still carrying the bug. The rebinding half of this rule already routes
		// through the shelf and was never affected, which is why `(mod) = 0` reported while
		// `(mod.named) = 0` did not.
		child = ast.SkipParentheses(parent)
	}
	return false
}

// objectMutationMethods are the `Object` statics that mutate their first argument.
var objectMutationMethods = map[string]bool{
	"assign":         true,
	"defineProperty": true,
	// Upstream's regex is `definePropert(?:y|ies)`, so the plural is a separate name rather than a
	// prefix match. `Object.seal` and `Object.preventExtensions` are deliberately absent and both
	// have a clean fixture: sealing an already-sealed namespace object changes nothing.
	"defineProperties": true,
	"freeze":           true,
	"setPrototypeOf":   true,
}

// reflectMutationMethods are the `Reflect` statics that mutate their first argument.
//
// Not the same set as Object's. `Reflect.deleteProperty` has no `Object` counterpart and
// `Reflect.assign` does not exist, so a shared list would be wrong in both directions.
// `Reflect.preventExtensions` is absent for the same reason as `Object.preventExtensions` and has
// its own clean fixture.
var reflectMutationMethods = map[string]bool{
	"defineProperty": true,
	"deleteProperty": true,
	"set":            true,
	"setPrototypeOf": true,
}

// isArgumentOfWellKnownMutationFunction reports whether an identifier is the first argument of a
// call that mutates it.
//
// This is the case the rule exists for in a typed tree. TypeScript already refuses `mod = 0` and
// `mod.named = 0` on a namespace import, but `Object.assign(mod, obj)` type-checks cleanly and fails
// at runtime, so this arm is the part that earns its keep here.
//
// Only the first argument. `Object.assign(obj, mod, other)` reads `mod` into `obj` and is clean
// upstream, with a fixture.
func isArgumentOfWellKnownMutationFunction(ctx rule.Context, identifier *ast.Node) bool {
	call := identifier.Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callExpression := call.AsCallExpression()
	if callExpression.Arguments == nil || len(callExpression.Arguments.Nodes) == 0 ||
		callExpression.Arguments.Nodes[0] != identifier {
		return false
	}

	// `(Object?.defineProperty)(mod, key, d)` wraps the callee in parentheses and upstream reports
	// it, so the parentheses come off before the shape is read.
	callee := ast.SkipParentheses(callExpression.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		// `Object[assign](mod, obj)` is an element access with a non-literal key, clean upstream and
		// clean here with a fixture.
		return false
	}
	access := callee.AsPropertyAccessExpression()
	object := access.Expression
	name := access.Name()
	if object == nil || object.Kind != ast.KindIdentifier || name == nil ||
		name.Kind != ast.KindIdentifier {
		return false
	}

	switch object.Text() {
	case "Object":
		if !objectMutationMethods[name.Text()] {
			return false
		}
	case "Reflect":
		if !reflectMutationMethods[name.Text()] {
			return false
		}
	default:
		return false
	}

	return resolvesToAGlobalObject(ctx, object)
}

// resolvesToAGlobalObject reports whether an identifier names the real global rather than a local of
// the same name.
//
// Upstream asks whether the name has any binding, which is escope's scope question. Ours is the
// checker's: where is this declared. The real `Object` declares in a bundled library file, and a
// `var Object` in the file being linted declares in that file, so the source file of the declaration
// separates them. Measured on upstream's two clean cases for this rather than assumed; both shadow
// `Object` with a `var`, one inside a block and one at the top level, and both stay clean.
//
// An identifier the checker cannot resolve answers false. Nothing in this tree reaches that, since a
// `lib.d.ts` is always present, and answering false there is the direction that costs a missed
// finding rather than a false positive.
func resolvesToAGlobalObject(ctx rule.Context, identifier *ast.Node) bool {
	// A shadowing `var Object` declares in the file under lint. Any declaration there disproves the
	// global, so this refuses on the first one rather than requiring all of them to be local: the real
	// `Object` is declared several times across the lib files and a shadow adds one more, so "every
	// declaration is local" would answer wrongly.
	return rule.IsDeclaredOnlyInDeclarationFiles(ctx.TypeChecker.GetSymbolAtLocation(identifier))
}
