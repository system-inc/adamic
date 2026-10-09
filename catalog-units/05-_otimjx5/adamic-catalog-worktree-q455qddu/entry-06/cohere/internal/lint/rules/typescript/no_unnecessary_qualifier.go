package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnnecessaryQualifier flags a namespace or enum qualifier that names the scope you are already in.
//
//	valid:   namespace X { export type T = number; } namespace Y { export const x: X.T = 3; }
//	valid:   namespace X { export type T = number; namespace Y { type T = string; const x: X.T = 0; } }
//	valid:   enum Foo { One } namespace Foo { export function bar() { return Foo.One; } }
//	invalid: namespace A { export type B = number; const x: A.B = 3; }
//	invalid: namespace A { export const x = 3; export const y = A.x; }
//	invalid: enum A { B, C = A.B }
//
// Writing `A.B` from inside `A` says something the reader has to check: it looks like a reach into
// another namespace and it is not. The name is already in scope, so the qualifier adds letters and
// removes the signal that a qualifier is supposed to carry.
//
// # Three questions, and each one can decline
//
// A qualifier is unnecessary only when all three hold. The qualifier resolves to a symbol whose
// declaration is one of the namespaces or enums this node is lexically inside. The accessed name
// resolves to a symbol. And the name the checker finds in scope AT THE QUALIFIER is the same symbol
// the access resolved to, compared through `GetExportSymbolOfSymbol` because a namespace member's
// local symbol and its exported symbol are different objects.
//
// That last comparison is the whole rule and it is what makes the shadowing case clean:
//
//	namespace X { export type T = number; namespace Y { type T = string; const x: X.T = 0; } }
//
// `T` IS in scope at that point, and it is the wrong `T`. Measured with a probe: the scope lookup
// finds a `T` and the export-symbol comparison answers false, so dropping the comparison would
// report a qualifier that is load-bearing and the fix would change what the code means.
//
// # The nesting suppression, and why this port drops the subtree rather than tracking a flag
//
// Upstream keeps a `currentFailedNamespaceExpression` set on report and cleared on the reported
// node's `:exit`, so a finding on `A.B.C.D` suppresses the nested findings on `A.B.C` and `A.B`. We
// have no exit visitor, and we do not need one: this rule does its own recursive walk, so declining
// to recurse into a node it just reported is the same suppression expressed as control flow.
//
// The equivalence is not free and was probed rather than assumed. What upstream suppresses is
// everything inside the reported node's subtree, so the two differ if a reportable qualifier can
// live inside a reported one without being part of its qualifier chain. The candidate shape is a
// type argument:
//
//	namespace A { export type B<T> = T; export type C = number; const x: A.B<A.C> = 3; }
//
// Measured: both report, because a type reference's arguments are children of the TYPE REFERENCE
// and not of the qualified name, in our parser as in estree. So the subtree of a reported
// `KindQualifiedName` or `KindPropertyAccessExpression` holds nothing but its own chain, and
// skipping it is exactly upstream's suppression.
//
// # What our harness cannot reach
//
// Upstream's ninth reporting case is `import * as Foo from './foo'; declare module './foo' { const
// x: Foo.T = 3; }`, which needs `./foo` to resolve so the namespace import aliases to a real module
// symbol. Driven through the installed 8.67.0 build with no `./foo` on disk, UPSTREAM IS SILENT
// TOO, so this is a case whose verdict depends on a module existing rather than a shape our harness
// cannot express. The alias branch it exercises is reached by a different fixture, recorded there.
//
// # Cost, and why the walk starts at the outermost enum or namespace
//
// The checker is asked only at a qualified name or a non-computed property access whose object is
// an entity name, which is nearly every `a.b` in a file. Outside every enum and namespace none of
// them can report: the first question compares the qualifier's declarations against the stack of
// enclosing declarations, and that stack is empty there, so the answer is no before the checker's
// answer is read. This rule walked the whole file from the source file and asked anyway, 0.9s of
// CPU on ahra, whose files have almost no enums and no namespaces (#hekjpw3).
//
// So the walk starts at each outermost enum or namespace, with that declaration as the stack's
// first entry, which is the stack the whole-file walk had at that point. A nested one is reached
// by its outermost one's walk, and skipped by the listener so no subtree is walked twice.
var NoUnnecessaryQualifier = rule.Rule{
	Name: "@typescript-eslint/no-unnecessary-qualifier",

	// Every question the rule asks is a resolution question: what does this qualifier name, what is
	// visible at this point, and are those the same symbol.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		walkOutermost := func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
				if ancestor.Kind == ast.KindEnumDeclaration || ancestor.Kind == ast.KindModuleDeclaration {
					return
				}
			}

			// namespacesInScope is upstream's push/pop stack of the enum and namespace
			// declarations this position is lexically inside. Upstream pushes the ESTree
			// declaration node; we push the same declaration our checker hands back from a
			// symbol, so the identity comparison below is over the same objects.
			namespacesInScope := []*ast.Node{}

			var visit func(*ast.Node)
			visit = func(current *ast.Node) {
				switch current.Kind {
				case ast.KindEnumDeclaration, ast.KindModuleDeclaration:
					namespacesInScope = append(namespacesInScope, current)
					current.ForEachChild(func(child *ast.Node) bool {
						visit(child)
						return false
					})
					namespacesInScope = namespacesInScope[:len(namespacesInScope)-1]
					return

				case ast.KindQualifiedName:
					qualifiedName := current.AsQualifiedName()
					if qualifiedName.Left == nil || qualifiedName.Right == nil {
						break
					}
					// A type-position qualifier cannot itself be parenthesized in the grammar
					// (`A.(B)` is not a type), so there is nothing to unwrap on the left. The
					// parenthesized TYPE case `const x: (A.T) = 3` puts the parens around the
					// whole type reference, above this node, and reaches here unwrapped.
					if unnecessaryQualifierReport(ctx, namespacesInScope,
						qualifiedName.Left, qualifiedName.Right) {
						// Reported: upstream suppresses everything inside this node, and so
						// does declining to recurse. See the doc comment.
						return
					}

				case ast.KindPropertyAccessExpression:
					propertyAccess := current.AsPropertyAccessExpression()
					if propertyAccess.Expression == nil {
						break
					}
					// Upstream's selector is `MemberExpression[computed=false]` and it then
					// requires the object to be an entity name expression, which is an
					// identifier or a chain of non-computed accesses over one. A computed
					// access is a different node kind here, so the first half is the kind.
					//
					// The unwrap is ours and it costs findings without it. estree has no
					// parenthesized node, so upstream's `node.object` for `(A).x` is the bare
					// identifier and the case reports; our parser hands back a
					// KindParenthesizedExpression, which the entity-name predicate declines.
					// Measured on the installed 8.67.0 build rather than reasoned about,
					// because the corpus writes no parenthesized form: `(A).x`, `(A.B).x`,
					// `((A).B).x` and `(A.T)` in type position ALL report there. A loop
					// rather than one step, since `((A))` nests.
					qualifier := unwrapQualifierParentheses(propertyAccess.Expression)
					if qualifier == nil || !isEntityNameExpressionForQualifier(qualifier) {
						break
					}
					name := propertyAccess.Name()
					if name == nil {
						break
					}
					if unnecessaryQualifierReport(ctx, namespacesInScope, qualifier, name) {
						return
					}
				}

				current.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}

			visit(node)
		}
		return rule.Listeners{
			ast.KindEnumDeclaration:   walkOutermost,
			ast.KindModuleDeclaration: walkOutermost,
		}
	},
}

// unnecessaryQualifierReport asks upstream's three questions and reports if all three hold.
//
// Returns whether it reported, which is what drives the subtree suppression at the call site.
func unnecessaryQualifierReport(
	ctx rule.Context,
	namespacesInScope []*ast.Node,
	qualifier *ast.Node,
	name *ast.Node,
) bool {
	namespaceSymbol := ctx.TypeChecker.GetSymbolAtLocation(qualifier)
	if namespaceSymbol == nil || !qualifierSymbolIsNamespaceInScope(ctx, namespacesInScope, namespaceSymbol) {
		return false
	}

	accessedSymbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if accessedSymbol == nil {
		return false
	}

	// The scope lookup is anchored at the QUALIFIER rather than at the name, which is upstream's
	// choice and is load-bearing: asking at the name would ask what is visible after the dot, which
	// is the namespace's own member table and would answer yes for every qualifier.
	scopeSymbols := ctx.TypeChecker.GetSymbolsInScope(qualifier, accessedSymbol.Flags)
	var fromScope *ast.Symbol
	for _, scopeSymbol := range scopeSymbols {
		if scopeSymbol.Name == name.Text() {
			fromScope = scopeSymbol
			break
		}
	}
	if fromScope == nil {
		return false
	}

	// A namespace member has a local symbol and an exported symbol, and the scope lookup answers
	// with the local one while the access resolves to the exported one. Comparing them directly
	// would answer false on every real finding. Comparing without this normalization also loses the
	// shadowing case in the other direction, so the call is doing two jobs.
	if ctx.TypeChecker.GetExportSymbolOfSymbol(fromScope) != accessedSymbol {
		return false
	}

	message := rule.Message{
		Id:          "unnecessaryQualifier",
		Description: "Qualifier is unnecessary since '" + name.Text() + "' is in scope.",
	}

	// The fix removes the qualifier and the dot, leaving the bare name. Both ends are token starts,
	// so the leading trivia of the qualifier stays where it is and the indentation survives. Using
	// Pos() on either node instead would eat the preceding whitespace.
	//
	// The decline below is a divergence from upstream and it is deliberate. Upstream computes this
	// same span and, when the qualifier sits inside parentheses the removal does not also remove,
	// it writes source that does not parse. Measured on the installed 8.67.0 build:
	//
	//	(A).x     ->  `const y = (x;`     Parsing error: ')' expected
	//	((A).B).x ->  `const y = ((B).x;` Parsing error: ')' expected
	//	(A.B).x   ->  `const y = (B).x;`  fine, the parens enclose the whole edit
	//	(A.T)     ->  `const x: (T) = 3;` fine, same reason
	//
	// A fix is applied unattended, so shipping the first two would rewrite working code into a
	// syntax error with nobody watching. The FINDING is upstream's and is reproduced; only the
	// repair is withheld, and only on the shapes where upstream's own repair is broken. The test
	// for it is positional rather than a paren count: if an open parenthesis lies between the
	// qualifier's start and the enclosing expression's start, its partner is outside the removal.
	removal := core.NewTextRange(
		rule.TokenRange(ctx.SourceFile, qualifier).Pos(),
		rule.TokenRange(ctx.SourceFile, name).Pos(),
	)
	if qualifierRemovalWouldStrandAParenthesis(qualifier) {
		ctx.ReportNode(qualifier, message)
		return true
	}

	ctx.ReportNodeWithFixes(qualifier, message, rule.RemoveRange(removal))
	return true
}

// qualifierSymbolIsNamespaceInScope answers whether any of the symbol's declarations is one of the
// namespaces or enums we are lexically inside, following an alias if the symbol is one.
//
// The alias step is what makes `import B = A; namespace A { const y = B.x; }` report: `B` is an
// alias symbol whose own declaration is the import statement rather than the namespace, so the
// identity loop above finds nothing and only the recursion reaches the real declaration. Measured
// on the installed 8.67.0 build and pinned by a fixture; without this branch that case is silent.
// The recursion is upstream's and terminates because an alias chain is finite and the checker
// resolves it to a non-alias.
func qualifierSymbolIsNamespaceInScope(
	ctx rule.Context,
	namespacesInScope []*ast.Node,
	symbol *ast.Symbol,
) bool {
	for _, declaration := range symbol.Declarations {
		for _, namespaceNode := range namespacesInScope {
			if declaration == namespaceNode {
				return true
			}
		}
	}

	if symbol.Flags&ast.SymbolFlagsAlias == 0 {
		return false
	}
	aliased := ctx.TypeChecker.GetAliasedSymbol(symbol)
	if aliased == nil || aliased == symbol {
		return false
	}
	return qualifierSymbolIsNamespaceInScope(ctx, namespacesInScope, aliased)
}

// isEntityNameExpressionForQualifier is upstream's `isEntityNameExpression`: an identifier, or a
// non-computed property access whose own object is one.
//
// Named for the rule rather than bare, because `internal/rules/typescript` is one namespace and a
// helper called `isEntityNameExpression` is the sort of name a sibling rule will also want.
//
// # This predicate is currently subsumed, and it is kept anyway
//
// A mutant neutralizing it survives, and the reason is not a fixture blind spot. For the predicate
// to decide anything, some object it rejects would have to resolve to a symbol whose declarations
// include a namespace or enum we are lexically inside, so that the two guards below it both say
// report. Probed over fourteen shapes that can sit left of a dot: call, element access, `as`, angle
// assertion, non-null, conditional, comma, nullish coalesce, await, array index, object literal,
// optional chain, `this`, and a property access on a literal. Every one of them either resolves to
// NO SYMBOL, which the `namespaceSymbol == nil` guard declines, or resolves to something whose
// declarations are not in the stack, which the identity loop declines.
//
// So it is redundant against this parser today. It is kept because it is upstream's, because the
// redundancy rests on what the checker happens to answer for those kinds rather than on anything
// guaranteed, and because deleting a guard on that basis is how the recorded verdict in this tree
// expired once already. The measurement is here rather than the deletion.
func isEntityNameExpressionForQualifier(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindPropertyAccessExpression:
		expression := node.AsPropertyAccessExpression().Expression
		return expression != nil && isEntityNameExpressionForQualifier(expression)
	}
	return false
}

// unwrapQualifierParentheses peels every parenthesis off a property access's object.
//
// A loop rather than one step, because `((A)).x` nests. Written out rather than calling
// ast.SkipParentheses, which dereferences its argument: the caller's node is optional and that
// helper is how this project lost 167 files to a nil panic.
func unwrapQualifierParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// qualifierRemovalWouldStrandAParenthesis answers whether removing the qualifier and its dot would
// delete an opening parenthesis whose partner survives.
//
// The shape is `(A).x`, where the parenthesis opens before the qualifier and closes after it but
// before the dot. Removing from the qualifier's start to the name's start takes the `)` with it and
// leaves the `(`. Walking up from the unwrapped qualifier is the cheapest test that distinguishes
// this from `(A.B).x`, where the parenthesis opens above the whole property access and is untouched.
func qualifierRemovalWouldStrandAParenthesis(qualifier *ast.Node) bool {
	return qualifier.Parent != nil && qualifier.Parent.Kind == ast.KindParenthesizedExpression
}
