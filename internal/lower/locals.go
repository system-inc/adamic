// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// variables lowers const and let declarations, each to a local of its own.
func (l *lowering) variables(list *ast.Node) ([]ir.Statement, error) {
	if list.Flags&ast.NodeFlagsBlockScoped == 0 && !assertionVarList(list) && !namespaceVariable(list) {
		return nil, &Refused{Where: l.program.Where(list), What: "var", Fix: "use const or let"}
	}
	statements := []ir.Statement{}
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		name := declaration.Name()
		if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
			// const [a, b] = tuple, and const { x, y } = object (collections.go).
			destructured, err := l.destructure(name, declaration.AsVariableDeclaration().Initializer)
			if err != nil {
				return nil, err
			}
			for index, statement := range destructured {
				if declared, ok := statement.(ir.Declare); ok && l.result.Locals[declared.Local].NamespaceVar {
					destructured[index] = ir.Assign{Local: declared.Local, Value: declared.Value}
				}
			}
			statements = append(statements, destructured...)
			continue
		}
		if !ast.IsIdentifier(name) {
			return nil, l.notYet(name, "a destructuring declaration")
		}

		if list.Flags&ast.NodeFlagsBlockScoped == 0 && !namespaceVariable(list) {
			if symbol := l.symbol(name); symbol != nil && len(symbol.Declarations) > 1 {
				return nil, &Refused{Where: l.program.Where(declaration), What: "repeated var assertion declarations", Fix: "use one declaration and subsequent assignments"}
			}
		}
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		if list.Flags&ast.NodeFlagsBlockScoped == 0 {
			l.result.Locals[local].Hoisted = true
		}
		if l.uninitializedDeclaration(declaration) {
			l.result.Locals[local].Uninitialized = true
			statements = append(statements, ir.Declare{Local: local, Uninitialized: true})
			continue
		}

		if initializer := declaration.AsVariableDeclaration().Initializer; assertionInitializer(initializer) {
			prefix, present, value, err := l.lazyAssertion(initializer, l.result.Locals[local].Type)
			if err != nil {
				return nil, err
			}
			if known, ok := present.(ir.BooleanConstant); ok && known.Value && len(prefix) == 0 {
				statements = append(statements, ir.Declare{Local: local, Value: value})
				continue
			}
			l.result.Locals[local].Uninitialized = true
			l.result.Locals[local].InitializerExpression = sourceExpression(initializer)
			statements = append(statements, prefix...)
			statements = append(statements, ir.Declare{Local: local, Uninitialized: true}, ir.If{Condition: present, Then: []ir.Statement{ir.Assign{Local: local, Value: value}}})
			continue
		}
		var value ir.Expression
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			if l.initializing == nil {
				l.initializing = map[int]*ast.Node{}
			}
			l.initializing[local] = name
			if l.detachedOwnDeclaration(declaration) {
				value = ir.BooleanConstant{Value: true}
			} else {
				value, err = l.expression(initializer)
			}
			delete(l.initializing, local)
			if err != nil {
				return nil, err
			}
		} else if !l.result.Locals[local].NamespaceVar && l.includesUndefined(l.checker.GetTypeAtLocation(name)) {
			// An optional declaration can be read before assignment. Its first value
			// is undefined, not the backend's unobservable storage placeholder. Namespace
			// var was initialized when hoisted; a later declaration must not reset it.
			value = ir.Undefined{Of: ir.Object}
		}
		if closure, literal := value.(ir.MakeClosure); literal && list.Flags&ast.NodeFlagsConst != 0 {
			l.result.Locals[local].ConstantClosure = closure.Function + 1
		}
		if l.result.Locals[local].NamespaceVar {
			if value != nil {
				statements = append(statements, ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type)})
			}
		} else {
			if value == nil && l.result.Locals[local].NamespaceState && l.includesUndefined(l.checker.GetTypeAtLocation(name)) {
				value = fit(ir.Undefined{Of: l.result.Locals[local].Type}, l.result.Locals[local].Type)
			}
			statements = append(statements, l.initializeLocal(local, fit(value, l.result.Locals[local].Type))...)
		}
	}
	return statements, nil
}

// skipped reports whether an element of an array binding pattern is a hole, as in const [, b]: the
// checker's tree has a binding element with no name there.
func skipped(binding *ast.Node) bool {
	return binding.Kind == ast.KindOmittedExpression || binding.Name() == nil
}

// declareLocal makes a new local for a declared name, typed by what the checker proved for it.
func (l *lowering) declareLocal(name *ast.Node) (int, error) {
	symbol := l.symbol(name)
	if symbol == nil {
		return 0, l.notYet(name, "the checker gave a declaration no symbol")
	}
	valueType := ir.Object
	if l.detachedOwnAlias(name) != nil {
		valueType = ir.Boolean
	} else if l.detachedOwnObjectParameter(name) {
		valueType = ir.Object
	} else if !l.alwaysUndefined[symbol] && !l.caught[symbol] {
		var err error
		if valueType, err = l.typeOf(name); err != nil {
			return 0, err
		}
	}
	if l.locals == nil {
		l.locals = map[*ast.Symbol]int{}
	}
	if local, isDeclared := l.locals[symbol]; isDeclared {
		// A global, registered before anything was lowered.
		return local, nil
	}
	l.locals[symbol] = len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name.Text(), Type: valueType, Function: l.functionIndex})
	proven := l.checker.GetTypeAtLocation(name)
	if l.detachedOwnAlias(name) != nil {
		// This local holds only the readiness marker. The intrinsic cannot escape,
		// so cycle analysis must not treat it as a user closure capturing cells.
		proven = l.checker.GetBooleanType()
	}
	l.noteLocal(l.locals[symbol], proven, name)
	return l.locals[symbol], nil
}

// noteAlso keeps another checker type a local has: a this shared by more than one instantiation.
func (l *lowering) noteAlso(local int, proven *checker.Type) {
	if l.localAlso == nil {
		l.localAlso = map[int][]*checker.Type{}
	}
	l.localAlso[local] = append(l.localAlso[local], proven)
}

// noteLocal keeps a local's checker type and declaration for the cycle finder.
func (l *lowering) noteLocal(local int, proven *checker.Type, node *ast.Node) {
	if l.localTypes == nil {
		l.localTypes, l.localNodes = map[int]*checker.Type{}, map[int]*ast.Node{}
	}
	l.localTypes[local], l.localNodes[local] = l.concrete(proven), node
	if l.instance != nil {
		l.instance.templates = append(l.instance.templates, template{local: local, proven: proven})
	}
}

// symbol is what a name refers to, the same whichever file names it: an import resolves to what it
// imports, and an exported declaration to its export symbol, so one declaration is one symbol.
func (l *lowering) symbol(node *ast.Node) *ast.Symbol {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	return l.checker.GetExportSymbolOfSymbol(symbol)
}

// local finds the local an identifier refers to, and notes a capture if it belongs to another
// function.
func (l *lowering) local(identifier *ast.Node) (int, bool) {
	symbol := l.symbol(identifier)
	local, isLocal := l.locals[symbol]
	if isLocal {
		l.touch(local)
		if declared := l.result.Locals[local]; declared.Captured && slotless(declared.Type) && l.unlowerable == nil {
			// A cell holds one adamic_value, and number | undefined needs two words. Lower says so once
			// it's done, since the capture is found here, where nothing can return an error.
			l.unlowerable = l.notYet(identifier, "a "+typeName(declared.Type)+" variable a function value captures")
		}
	}
	return local, isLocal
}

// touch notes that the function being lowered reads or writes a local. One declared in another
// function, and not a global, is captured: it moves into a cell, and every closure being lowered
// between its function and this one carries that cell in its environment.
func (l *lowering) touch(local int) {
	declared := l.result.Locals[local]
	if declared.Ready != 0 {
		l.touch(declared.Ready - 1)
	}
	if declared.Global || declared.Function == l.functionIndex {
		return
	}
	l.result.Locals[local].Captured = true
	if name, isInitializing := l.initializing[local]; isInitializing && l.unlowerable == nil {
		// const f = () => f(): the function value is made before the variable's cell is, so it has
		// nothing to capture yet.
		l.unlowerable = l.notYet(name, "a function value that captures the variable its own initializer declares")
	}
	start := 0
	for position, closure := range l.closures {
		if closure == declared.Function {
			start = position + 1
		}
	}
	for _, closure := range l.closures[start:] {
		function := &l.result.Functions[closure]
		if !slices.Contains(function.Environment, local) {
			function.Environment = append(function.Environment, local)
		}
	}
}

// checked guards switch lexical bindings, captured cells, and globals reached before initialization.
func (l *lowering) checked(local int) bool {
	return l.result.Locals[local].NamespaceState || l.result.Locals[local].Ready != 0 || (l.result.Locals[local].Global && (l.function != nil || l.cyclicModules)) || (l.function != nil && l.result.Locals[local].Preallocated && l.result.Locals[local].Captured)
}

func (l *lowering) constant(value string) int {
	if l.strings == nil {
		l.strings = map[string]int{}
	}
	if index, isKnown := l.strings[value]; isKnown {
		return index
	}
	l.strings[value] = len(l.result.Strings)
	l.result.Strings = append(l.result.Strings, value)
	return l.strings[value]
}

// localRead preserves checker narrowing for both private and qualified singleton reads.
func (l *lowering) localRead(node *ast.Node, local int) (ir.Expression, error) {
	read := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.result.Locals[local].NamespaceState || l.checkedModuleRead(node, local), Readiness: sourceExpression(node)})
	if l.result.Locals[local].Type == ir.Union {
		// Where the checker has narrowed it to fewer members held one way, it's read as that.
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		observing := comparedWithUndefined(node) || (parent != nil && parent.Kind == ast.KindTypeOfExpression)
		if narrowed, isKnown := l.representation(l.checker.GetTypeAtLocation(node)); isKnown && narrowed != ir.Union && !observing {
			// Calls and captured writes can invalidate the checker's narrowing. Check the
			// held member before casting it, with ordinary IR shared by both backends.
			name := "object"
			switch narrowed.Present() {
			case ir.Number:
				name = "number"
			case ir.Boolean:
				name = "boolean"
			case ir.String:
				name = "string"
			case ir.Closure:
				name = "function"
			}
			if name == "object" {
				// typeof cannot distinguish differently held object members.
				declared := l.concrete(l.checker.GetTypeOfSymbol(l.symbol(node)))
				members := []*checker.Type{declared}
				if declared.Flags()&checker.TypeFlagsUnion != 0 {
					members = declared.Types()
				}
				for _, member := range members {
					if held, known := l.representation(member); known && held != narrowed && (held == ir.Object || held == ir.Array || held == ir.Map) {
						return nil, l.notYet(node, "a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables")
					}
				}
			}
			b := l.libraryArrayBuilder([]ir.Expression{read})
			held := b.read(b.parameters[0])
			matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant(name)}})
			if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
				matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
			}
			message := "union member where the checker narrowed it away: a call since the narrowing put it back"
			b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
			read = b.finish("narrowed_union_member", ir.Narrow{Value: held, To: narrowed})
		}
	}
	if declared := l.result.Locals[local].Type; declared.IsMaybe() {
		// Where the checker has narrowed it to what it holds, it's read as that.
		if narrowed, _ := l.representation(l.checker.GetTypeAtLocation(node)); narrowed == declared.Present() && !l.acceptsUndefined(node) {
			read = ir.Unwrap{Value: read}
		}
	}
	return l.defined(node, read), nil
}
