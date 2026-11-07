package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsc's debug.ts hands a function around as a stack crawl mark: assertIsDefined(v, message,
// stackCrawlMark || assertEachIsDefined), where stackCrawlMark?: AnyFunction and AnyFunction is
// (...args: never[]) => void. Nothing calls the mark. It's compared (===), handed on through || into the
// next helper's never-rest parameter, and given to Error.captureStackTrace, and a call through a
// never-rest slot is already refused (callClosure). So a function that, by a program-wide flow proof,
// only ever goes into never-rest slots is lowered as an identity marker (#avqasqt): one closure per
// function, held by a global made before anything runs, whose body never runs the declaration. No
// forwarder or generic instance is made for it, which is what let assertEachIsDefined, a generic
// overload set, past functionValue's overload stop and an instance whose T nothing can read. Every
// other function value keeps its lowering.

// neverRestMarkerCalled is what a marker's body says if a call ever reaches it, which the proof and
// callClosure's never-rest stop keep from happening. Each marker stands for one function, so it names the
// function and where it's declared, and the reader of a fence that fires knows which of tsc's helpers was
// called (Diagnostics' should on 2848eb81).
func neverRestMarkerCalled(name, where string) string {
	return "the identity marker for " + name + " (" + where + ") was called; it stands for a function that is only ever passed as a mark"
}

// neverRestMarks is the proof's and the markers' state for one lowering.
type neverRestMarks struct {
	// references is every read of a function in the program, by its symbol, made once.
	references map[*ast.Symbol][]*ast.Node
	// answers is each root function's final answer. A followed parameter's is never kept (neverRestOnly).
	answers map[*ast.Symbol]bool
	// markers is the global holding each proven function's marker, by its symbol.
	markers map[*ast.Symbol]int
	// slot is the first never-rest slot a proof walk sent a value into. Every slot the proof admits is
	// one never-rest signature, so it is the type of everywhere a marker can go, and the markers' closure
	// records carry it (neverRestMarker).
	slot *checker.Type
}

// neverRestState is l.neverRest, made the first time it's asked for, its reference index with it.
func (l *lowering) neverRestState() *neverRestMarks {
	if l.neverRest == nil {
		l.neverRest = &neverRestMarks{references: l.neverRestReferenceIndex(), answers: map[*ast.Symbol]bool{}, markers: map[*ast.Symbol]int{}}
	}
	return l.neverRest
}

// neverRestOnly reports whether every read of the function symbol names is one the identity marker
// serves: a direct call, which isn't a value, or a value that only ever goes into never-rest slots,
// compared or handed to Error.captureStackTrace (neverRestReference). Only a function declared at a
// file's top level or in a top-level namespace's block (neverRestPlace), of the program's own, with a
// body, or an overload set with its implementation, and with no this parameter, not a generator and not
// async, is asked about; anything else answers no.
func (l *lowering) neverRestOnly(symbol *ast.Symbol) bool {
	if l.neverRestImplementation(symbol) == nil {
		return false
	}
	marks := l.neverRestState()
	if answer, known := marks.answers[symbol]; known {
		return answer
	}
	// The proof reads the declared slots, the program's answer, never those of the instantiation being
	// lowered when it's first asked, so the answer kept is the same wherever it's asked again.
	outerMapper := l.typeMapper
	l.typeMapper = nil
	defer func() { l.typeMapper = outerMapper }()
	// The walk is a greatest fixpoint. tsc's helpers hand the mark around in cycles (assertIsDefined
	// hands it to fail, assertEachIsDefined to assertIsDefined), so a parameter already being proved in
	// this walk answers yes where it's reached again, and a cycle whose only uses are on the list
	// proves. Only the root's final answer is kept, never a followed parameter's: a yes given while
	// leaning on a parameter still in progress is stale if that parameter fails later. The root's answer
	// is the conjunction of everything it reached, so any failure anywhere makes it no, which is sound.
	visited := map[*ast.Symbol]bool{}
	answer := true
	for _, reference := range marks.references[symbol] {
		if !l.neverRestReference(reference, visited) {
			answer = false
			break
		}
	}
	marks.answers[symbol] = answer
	return answer
}

// neverRestImplementation is the declaration with a body of a function the proof may be asked about, or
// nil: every declaration is a function declaration in a place neverRestPlace admits, in a program file,
// not a library's or the prelude's, none a generator, async or taking this, and exactly one has a body.
func (l *lowering) neverRestImplementation(symbol *ast.Symbol) *ast.Node {
	if symbol == nil || symbol.Flags&ast.SymbolFlagsFunction == 0 {
		return nil
	}
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil || !neverRestPlace(declaration) || !neverRestProgramFile(ast.GetSourceFileOfNode(declaration)) {
			return nil
		}
		if generatorDeclaration(declaration) || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) || neverRestThisParameter(declaration) {
			return nil
		}
		if declaration.Body() != nil {
			if implementation != nil {
				return nil
			}
			implementation = declaration
		}
	}
	return implementation
}

// neverRestPlace reports whether a function declaration sits where its every read is a name the
// reference index sees: at a file's top level, or in the block of a namespace declared at a file's top
// level by a plain name and not ambient, as tsc's debug.ts declares its helpers inside namespace Debug.
// A namespace member is read bare inside the namespace and as Debug.f outside it, and the namespace
// itself is never a runtime object: reading it as a value, or a member by a computed name, is already
// NotYet (namespaceRefusal, namespaceExpression), and so is an ESM namespace object read as a value
// (moduleNamespaceRefusal), so no read the index misses can lower. A nested function, a nested or dotted
// namespace's member and a class's member answer no.
func neverRestPlace(declaration *ast.Node) bool {
	parent := declaration.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindSourceFile {
		return true
	}
	if parent.Kind != ast.KindModuleBlock {
		return false
	}
	namespace := parent.Parent
	return namespace != nil && namespace.Kind == ast.KindModuleDeclaration && ast.IsIdentifier(namespace.Name()) && !ambientDeclaration(namespace) && namespace.Parent != nil && namespace.Parent.Kind == ast.KindSourceFile
}

// neverRestProgramFile reports whether a file is one of the program's own sources: not a declaration
// file, the library's or the prelude.
func neverRestProgramFile(file *ast.SourceFile) bool {
	return file != nil && !file.IsDeclarationFile && !load.IsLibrary(file) && !load.IsPrelude(file)
}

// neverRestThisParameter reports whether a function declaration takes a this parameter, which shifts
// every argument's position past the parameter it names.
func neverRestThisParameter(declaration *ast.Node) bool {
	parameters := declaration.Parameters()
	return len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) && parameters[0].Name().Text() == "this"
}

// neverRestReferenceIndex finds every read of a function in the program's files, by the symbol it
// reads, so each question walks a function's own reads rather than every file again. A type position
// isn't a read, and neither is a declaration's name or an import or export specifier's: export { f }
// and import { f } read nothing, and the import's own reads are found where they're written. export
// default f is a read, and the proof answers no for it.
func (l *lowering) neverRestReferenceIndex() map[*ast.Symbol][]*ast.Node {
	references := map[*ast.Symbol][]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) {
			if symbol := l.neverRestReferent(node); symbol != nil && symbol.Flags&ast.SymbolFlagsFunction != 0 && neverRestRead(node) {
				references[symbol] = append(references[symbol], node)
			}
			return false
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if neverRestProgramFile(file) {
			file.AsNode().ForEachChild(visit)
		}
	}
	return references
}

// neverRestRead reports whether an identifier reads what it names, as neverRestReferenceIndex counts
// reads. A shorthand property's name is a declaration's name and a read both, and counted as a read.
func neverRestRead(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == identifier {
		return true
	}
	if ast.IsDeclarationName(identifier) {
		return false
	}
	return parent == nil || parent.Kind != ast.KindImportSpecifier && parent.Kind != ast.KindExportSpecifier
}

// neverRestReferent is the symbol an identifier reads: l.symbol's, except that a shorthand property's
// name reads the binding it names, not the field it makes, as shorthand's lowering reads it.
func (l *lowering) neverRestReferent(identifier *ast.Node) *ast.Symbol {
	if parent := identifier.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == identifier {
		symbol := l.checker.GetShorthandAssignmentValueSymbol(parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = l.checker.GetAliasedSymbol(symbol)
		}
		if symbol == nil {
			return nil
		}
		return l.checker.GetExportSymbolOfSymbol(symbol)
	}
	return l.symbol(identifier)
}

// neverRestReference reports whether one read of a root function is one the marker serves: the callee
// of a direct call, through parentheses, f(...) or a module namespace's D.f(...), which is an ordinary
// call lowered as it always is and never a value; or a bare value whose every use neverRestUse admits.
// Read through a property access anywhere else, D.f not called, it answers no.
func (l *lowering) neverRestReference(reference *ast.Node, visited map[*ast.Symbol]bool) bool {
	use := reference
	if parent := reference.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == reference {
		use = parent
	}
	outer := use
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if call := outer.Parent; call != nil && call.Kind == ast.KindCallExpression && call.AsCallExpression().Expression == outer {
		return call.Flags&ast.NodeFlagsOptionalChain == 0
	}
	if use != reference {
		return false
	}
	return l.neverRestUse(reference, visited)
}

// neverRestUse reports whether the value at node, a read of a root function, of a followed parameter, or
// a || or ?? holding one, goes only where the marker serves, through parentheses. (a) is an argument of
// a call to one function declaration of the program with a body, at a plain parameter whose type,
// undefined and null aside, is one never-rest signature, and whose every use inside the callee is on
// this list too (neverRestParameter). (b) is an operand of === or !==. (c) is Error.captureStackTrace's
// second argument, on the library's Error. (d) is either operand of || or ??, whose whole is on this
// list. Anything else answers no, failing closed: an initializer, an assignment, a field, a return, a
// spread, a conditional's arm, typeof, !, a test, a property read, and any position not named here.
func (l *lowering) neverRestUse(node *ast.Node, visited map[*ast.Symbol]bool) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.Left != outer && binary.Right != outer {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return l.neverRestUse(parent, visited)
		}
	case ast.KindCallExpression:
		if parent.Flags&ast.NodeFlagsOptionalChain != 0 || parent.AsCallExpression().Arguments == nil {
			return false
		}
		arguments := parent.AsCallExpression().Arguments.Nodes
		position := slices.Index(arguments, outer)
		if position < 0 {
			return false
		}
		for _, argument := range arguments[:position] {
			if argument.Kind == ast.KindSpreadElement {
				return false // A spread before it moves it off the parameter its place names.
			}
		}
		if position == 1 && l.neverRestCaptureStackTrace(parent) {
			return true
		}
		return l.neverRestParameter(parent, position, visited)
	}
	return false
}

// neverRestCaptureStackTrace reports whether a call is Error.captureStackTrace's, through the library's
// Error, as nodeCaptureStackTraceRefusal admits it. Its call evaluates its arguments and nothing else
// (library_node_globals.go), so the marker given it is never called; a form that refusal stops still
// stops there.
func (l *lowering) neverRestCaptureStackTrace(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Error") && l.nodeLibraryMember(callee) == nodeCaptureStackTraceMember
}

// neverRestParameter is neverRestUse's (a): the value given at position to call goes into a never-rest
// parameter of the one function declaration the callee names, which is followed: every use of the
// parameter inside that function must be on neverRestUse's list. The callee is named by an identifier or
// a property access, never a closure, a method or a new, and one whose declaration can't be read, a this
// parameter, or a rest, destructured or defaulted parameter at position answers no. A never-rest slot is
// asked for so the slot's binding is never's (type_argument_bindings.go) and no call through it lowers
// (callClosure).
func (l *lowering) neverRestParameter(call *ast.Node, position int, visited map[*ast.Symbol]bool) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	var named *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		named = callee
	case ast.KindPropertyAccessExpression:
		named = callee.AsPropertyAccessExpression().Name()
	default:
		return false
	}
	symbol := l.symbol(named)
	if symbol == nil {
		return false
	}
	// An overload's argument lands in its implementation's parameter, the one body there is.
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
			if implementation != nil {
				return false
			}
			implementation = declaration
		}
	}
	if implementation == nil || !neverRestProgramFile(ast.GetSourceFileOfNode(implementation)) || neverRestThisParameter(implementation) {
		return false
	}
	parameters := implementation.Parameters()
	if position >= len(parameters) {
		return false
	}
	parameter := parameters[position]
	if parameter.AsParameterDeclaration().DotDotDotToken != nil || parameter.AsParameterDeclaration().Initializer != nil || !ast.IsIdentifier(parameter.Name()) {
		return false
	}
	own := l.symbol(parameter.Name())
	if own == nil {
		return false
	}
	slot := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(own))
	if slot.Flags()&checker.TypeFlagsObject == 0 {
		return false // A type parameter's slot is what each call makes it, not a never-rest one.
	}
	signatures := l.checker.GetSignaturesOfType(slot, checker.SignatureKindCall)
	if len(signatures) != 1 || !l.censusNeverRestSignature(signatures[0]) {
		return false
	}
	if l.neverRest.slot == nil {
		l.neverRest.slot = slot
	}
	if visited[own] {
		return true // Being proved in this walk: the greatest fixpoint neverRestOnly describes.
	}
	visited[own] = true
	return l.neverRestParameterUses(implementation, parameter, own, visited)
}

// neverRestParameterUses reports whether every use of a followed parameter, own, inside the function
// declaring it is on neverRestUse's list. One assigned anywhere, or read from a nested function or class,
// which captures it, answers no, and so does any read of arguments, which reads every parameter without
// naming one.
func (l *lowering) neverRestParameterUses(implementation, parameter *ast.Node, own *ast.Symbol, visited map[*ast.Symbol]bool) bool {
	var holds func(node *ast.Node, captured bool) bool
	holds = func(node *ast.Node, captured bool) bool {
		if ast.IsPartOfTypeNode(node) {
			return true
		}
		if ast.IsIdentifier(node) {
			if node.Text() == "arguments" && (node.Parent == nil || node.Parent.Kind != ast.KindPropertyAccessExpression || node.Parent.AsPropertyAccessExpression().Name() != node) {
				return false
			}
			if node == parameter.Name() || l.neverRestReferent(node) != own {
				return true
			}
			return !captured && !ast.IsAssignmentTarget(node) && l.neverRestUse(node, visited)
		}
		inner := captured || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		every := true
		node.ForEachChild(func(child *ast.Node) bool {
			every = holds(child, inner)
			return !every
		})
		return every
	}
	every := true
	implementation.ForEachChild(func(child *ast.Node) bool {
		every = holds(child, false)
		return !every
	})
	return every
}

// neverRestMarker is the identity marker of a function neverRestOnly proves, read at node: one closure
// per function, made before anything runs and held by a global of its own, as functionValue holds a
// forwarder, so every read of one function is === to every other and two functions' markers differ. Its
// body is a panic, never the declaration's: it takes no parameters, forwards nothing and checks no
// predicate, since nothing calls it. Its closure record keeps the closed-world passes whole
// (FunctionTypeTargets, the cycle finder), as every function value's does, under the never-rest slot's
// type, where the proof sends it, rather than the declaration's, so a call through an ordinary slot of
// the declaration's type never counts the marker as one of its targets (Lower's should on 4e89ac5). A
// function only ever compared, which no slot receives, keeps its declaration's type.
func (l *lowering) neverRestMarker(node *ast.Node, symbol *ast.Symbol) ir.Expression {
	marks := l.neverRestState()
	if held, isMade := marks.markers[symbol]; isMade {
		return ir.Read{Local: held, Of: ir.Closure}
	}
	implementation := l.neverRestImplementation(symbol)
	source := l.sourceIdentity(implementation)
	source.Role = "identity_marker"
	source.Declaration = append(append([]string{}, source.Declaration...), "identity_marker")
	index := len(l.result.Functions)
	called := ir.Panic{Message: ir.StringConstant{Index: l.constant(neverRestMarkerCalled(implementation.Name().Text(), l.program.Where(implementation.Name())))}}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: implementation.Name().Text() + "_identity_marker", Source: source, Closure: true, Body: []ir.Statement{called}})
	proven := marks.slot
	if proven == nil {
		proven = l.checker.GetTypeAtLocation(node)
	}
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(proven), function: index, node: node})
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "identity_marker", Source: source, Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	marks.markers[symbol] = held
	return ir.Read{Local: held, Of: ir.Closure}
}
