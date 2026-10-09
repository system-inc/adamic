package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoWriteOnlyCollectionId = "writeOnlyCollection"

// correctnessNoWriteOnlyCollectionText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-write-only-collection.json`.
var correctnessNoWriteOnlyCollectionText = policy.MessageOf("nexus/correctness-no-write-only-collection", correctnessNoWriteOnlyCollectionId)

func correctnessNoWriteOnlyCollectionMessage() rule.Message {
	return rule.Message{Id: correctnessNoWriteOnlyCollectionId, Description: correctnessNoWriteOnlyCollectionText.Render(nil)}
}

// CorrectnessNoWriteOnlyCollection reports a local array, Map or Set that is filled and never read.
//
//	invalid: const bmRanks = new Map<number, number>(); rows.forEach(function(row, index) { bmRanks.set(row.rowid, index); });
//	invalid: const seen: string[] = []; for(const row of rows) { seen.push(row.id); }
//	valid:   const ranks = new Map<number, number>(); rows.forEach(...ranks.set(...)...); return ranks.get(id);
//	valid:   const seen = new Set<string>(); for(...) { if(seen.has(id)) continue; seen.add(id); }
//	valid:   const list: string[] = []; list.push(a); return { list };
//
// # Where it came from
//
// `modules/os/database/AhraOsProfileCommitsStore.ts` in ahra, `searchProfileCommitsHybrid`:
// `bmRanks` and `vecRanks` are `new Map<number, number>()`, each filled with `.set(row.rowid, rank)`
// during rank fusion and never read. The ranks the result carries come from `hitData` instead. Found
// by the JavaScript-catalog pass of the new-rules sweep (`#tevhg3f`, after sonarjs
// `no-unused-collection`), built as part of `#j03vwm6`.
//
// # The shape, exactly
//
//  1. A `const` declared inside a function (or a static block or property initializer), with a plain
//     identifier name, initialized to an array literal or to `new Map(...)` / `new Set(...)` with the
//     default library's constructor. `as` and `satisfies` around the initializer are looked through.
//     A module-level binding is not read: it can be exported by a later `export { name }`, which this
//     rule would have to resolve through an alias.
//  2. Every reference to the binding, found by symbol (a shorthand `{ name }` resolves to the
//     binding, and is a read), is a write. A write is a call of a mutating method that is a
//     statement on its own, so its result is dropped (for an array `push`, `unshift`, `pop`, `shift`,
//     `splice`; for a Map `set`, `delete`, `clear`; for a Set `add`, `delete`, `clear`, each resolved
//     to the default library's declaration so a typed wrapper with the same names is not read as
//     one), or the target of a plain `=` assignment through an element (`list[index] = value`).
//  3. At least one such write. A collection never referenced at all is an unused variable, which
//     `no-unused-vars` owns.
//
// Anything else is a read, and the collection stays silent: iterating it, spreading it, `.length`,
// `.size`, `.has`, `.get`, returning it, passing it, storing it, a mutating call whose result is used
// (`const length = list.push(x)`), a write in an arrow's expression body (the arrow returns the
// result, and whether its caller reads it is not this rule's question), `typeof list` in a type. Each
// is a missed finding at worst, never a false one.
//
// # No fix
//
// Deleting the collection and its writes is right when it is dead, and wrong when a reader was meant
// to exist. Which one is the author's call.
var CorrectnessNoWriteOnlyCollection = rule.Rule{
	Name: "nexus/correctness-no-write-only-collection",

	// Symbol identity finds every reference to the binding through shadowing and shorthand properties,
	// and ties each mutating method to the default library's declaration.
	NeedsTypeChecker: true,

	// Whether `Map`, `Set` and the mutating methods are the default library's, through type_checking.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				correctnessNoWriteOnlyCollectionCheck(ctx, node)
			},
		}
	},
}

// correctnessNoWriteOnlyCollectionWriters are the mutating methods of each collection, by the default
// library interface that declares them, whose only effect when their result is dropped is the write.
var correctnessNoWriteOnlyCollectionWriters = map[string]map[string]bool{
	"Array": {"push": true, "unshift": true, "pop": true, "shift": true, "splice": true},
	"Map":   {"set": true, "delete": true, "clear": true},
	"Set":   {"add": true, "delete": true, "clear": true},
}

func correctnessNoWriteOnlyCollectionCheck(ctx rule.Context, declaration *ast.Node) {
	variable := declaration.AsVariableDeclaration()
	name := declaration.Name()
	list := declaration.Parent
	if name == nil || name.Kind != ast.KindIdentifier || variable.Initializer == nil ||
		list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsConst == 0 {
		return
	}
	root := control_flow_graph.RootOf(declaration)
	if root == nil || root.Kind == ast.KindSourceFile {
		return
	}
	kind := correctnessNoWriteOnlyCollectionKind(ctx, variable.Initializer)
	if kind == "" {
		return
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0] != declaration {
		return
	}

	writes := 0
	read := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if read {
			return true
		}
		if node.Kind == ast.KindIdentifier && node != name && node.Text() == name.Text() {
			resolved := ctx.TypeChecker.GetSymbolAtLocation(node)
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
				if valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); valueSymbol != nil {
					resolved = valueSymbol
				}
			}
			if resolved == symbol {
				if !correctnessNoWriteOnlyCollectionIsWrite(ctx, node) {
					read = true
					return true
				}
				writes++
			}
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	if read || writes == 0 {
		return
	}
	ctx.ReportNode(name, correctnessNoWriteOnlyCollectionMessage())
}

// correctnessNoWriteOnlyCollectionKind names the collection an initializer makes, "Array", "Map" or
// "Set", and "" for anything else.
func correctnessNoWriteOnlyCollectionKind(ctx rule.Context, initializer *ast.Node) string {
	value := initializer
	for {
		value = ast.SkipParentheses(value)
		switch value.Kind {
		case ast.KindAsExpression:
			value = value.AsAsExpression().Expression
			continue
		case ast.KindSatisfiesExpression:
			value = value.AsSatisfiesExpression().Expression
			continue
		}
		break
	}
	switch value.Kind {
	case ast.KindArrayLiteralExpression:
		return "Array"
	case ast.KindNewExpression:
		constructor := ast.SkipParentheses(value.AsNewExpression().Expression)
		if constructor.Kind != ast.KindIdentifier || (constructor.Text() != "Map" && constructor.Text() != "Set") {
			return ""
		}
		if !type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(constructor)) {
			return ""
		}
		return constructor.Text()
	}
	return ""
}

// correctnessNoWriteOnlyCollectionIsWrite says whether one reference to the collection only writes
// it: a dropped call of a default-library mutating method, or the target of `name[key] = value`.
func correctnessNoWriteOnlyCollectionIsWrite(ctx rule.Context, reference *ast.Node) bool {
	parent := reference.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		if access.Expression != reference || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		call := parent.Parent
		if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != parent {
			return false
		}
		statement := call.Parent
		for statement != nil && statement.Kind == ast.KindParenthesizedExpression {
			statement = statement.Parent
		}
		if statement == nil || statement.Kind != ast.KindExpressionStatement {
			return false
		}
		return correctnessNoWriteOnlyCollectionIsWriter(ctx, access.Name())
	case ast.KindElementAccessExpression:
		if parent.AsElementAccessExpression().Expression != reference {
			return false
		}
		outer := parent
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		assignment := outer.Parent
		if assignment == nil || assignment.Kind != ast.KindBinaryExpression {
			return false
		}
		binary := assignment.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == outer
	}
	return false
}

// correctnessNoWriteOnlyCollectionIsWriter says whether a method name resolves to a mutating method
// of the default library's Array, Map or Set, with every declaration there.
func correctnessNoWriteOnlyCollectionIsWriter(ctx rule.Context, name *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil {
			return false
		}
		writers := correctnessNoWriteOnlyCollectionWriters[owner.Name().Text()]
		if !writers[name.Text()] {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}
