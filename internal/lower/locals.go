// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
)

// variables lowers const and let declarations, each to a local of its own.
func (lowering *lowering) variables(list *ast.Node) ([]ir.Statement, error) {
	if list.Flags&ast.NodeFlagsBlockScoped == 0 && !assertionVarList(list) && !namespaceVariable(list) {
		return nil, &Refused{Where: lowering.program.Where(list), What: "var", Fix: "use const or let"}
	}
	statements := []ir.Statement{}
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		name := declaration.Name()
		if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
			// const [a, b] = tuple, and const { x, y } = object (collections.go).
			destructured, destructureError := lowering.destructure(
				name,
				declaration.AsVariableDeclaration().Initializer,
			)
			if destructureError != nil {
				return nil, destructureError
			}
			for index, statement := range destructured {
				if declared, ok := statement.(ir.Declare); ok && lowering.result.Locals[declared.Local].NamespaceVar {
					destructured[index] = ir.Assign{Local: declared.Local, Value: declared.Value}
				}
			}
			statements = append(statements, destructured...)
			continue
		}
		if !ast.IsIdentifier(name) {
			return nil, lowering.notYet(name, "a destructuring declaration")
		}

		if list.Flags&ast.NodeFlagsBlockScoped == 0 && !namespaceVariable(list) {
			if symbol := lowering.symbol(name); symbol != nil && len(symbol.Declarations) > 1 {
				return nil, &Refused{
					Where: lowering.program.Where(declaration),
					What:  "repeated var assertion declarations",
					Fix:   "use one declaration and subsequent assignments",
				}
			}
		}
		local, err := lowering.declareLocal(name)
		if err != nil {
			return nil, err
		}
		if list.Flags&ast.NodeFlagsBlockScoped == 0 {
			lowering.result.Locals[local].Hoisted = true
		}
		if lowering.result.Locals[local].NamespaceVar && declaration.Initializer() == nil {
			continue // Hoisted storage must not overwrite an earlier assignment.
		}
		if lowering.uninitializedDeclaration(declaration) {
			lowering.result.Locals[local].Uninitialized = true
			initializer := declaration.AsVariableDeclaration().Initializer
			declaration := ir.Declare{Local: local, Uninitialized: true}
			if lowering.result.Locals[local].Placeholder != "" {
				declaration.Value = lowering.placeholderInitialValue(initializer, lowering.result.Locals[local].Type)
			}
			statements = append(statements, declaration)
			continue
		}

		var value ir.Expression
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			if lowering.initializing == nil {
				lowering.initializing = map[int]*ast.Node{}
			}
			lowering.initializing[local] = name
			if lowering.detachedOwnDeclaration(declaration) && lowering.detachedOwnMethod(initializer) {
				value = ir.BooleanConstant{Value: true}
			} else {
				value, err = lowering.expression(initializer)
			}
			delete(lowering.initializing, local)
			if err != nil {
				return nil, err
			}
		} else if !lowering.result.Locals[local].NamespaceVar &&
			(lowering.includesUndefined(lowering.checker.GetTypeAtLocation(name)) ||
				lowering.evolvingObject(name) != nil) {
			// An optional declaration can be read before assignment. Its first value
			// is undefined, not the backend's unobservable storage placeholder. Namespace
			// var was initialized when hoisted; a later declaration must not reset it.
			value = ir.Undefined{Of: ir.Object}
		}
		if closure, literal := value.(ir.MakeClosure); literal && list.Flags&ast.NodeFlagsConst != 0 {
			lowering.result.Locals[local].ConstantClosure = closure.Function + 1
		}
		if lowering.result.Locals[local].NamespaceVar {
			if value != nil {
				statements = append(
					statements,
					ir.Assign{Local: local, Value: fit(value, lowering.result.Locals[local].Type)},
				)
			}
		} else {
			statements = append(
				statements,
				lowering.initializeLocal(local, fit(value, lowering.result.Locals[local].Type))...,
			)
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
func (lowering *lowering) declareLocal(name *ast.Node) (int, error) {
	symbol := lowering.symbol(name)
	if symbol == nil {
		return 0, lowering.notYet(name, "the checker gave a declaration no symbol")
	}
	valueType := ir.Object
	inferred := lowering.evolvingObject(name)
	if lowering.detachedOwnAlias(name) != nil {
		valueType = ir.Boolean
	} else if lowering.detachedOwnObjectParameter(name) {
		valueType = ir.Object
	} else if !lowering.alwaysUndefined[symbol] && !lowering.caught[symbol] {
		var typeOfError error
		if valueType, typeOfError = lowering.typeOf(name); typeOfError != nil {
			if inferred == nil {
				return 0, typeOfError
			}
			valueType = ir.Object
		}
	}
	if origin := lowering.placeholderOrigin(name); origin != "" {
		valueType = placeholderStorage(valueType)
	}
	unsetField := ""
	if declaration := name.Parent; declaration != nil &&
		declaration.Kind == ast.KindVariableDeclaration &&
		declaration.Type() == nil &&
		declaration.Initializer() != nil &&
		declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
		unsetField = lowering.savedFactoryOrigin(declaration.Initializer(), map[*ast.Symbol]bool{})
		if unsetField != "" {
			valueType = ir.Union
		}
	}
	if lowering.locals == nil {
		lowering.locals = map[*ast.Symbol]int{}
	}
	if local, isDeclared := lowering.locals[symbol]; isDeclared {
		// A global, registered before anything was lowered.
		return local, nil
	}
	lowering.locals[symbol] = len(lowering.result.Locals)
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{
		UnsetField:  unsetField,
		Name:        name.Text(),
		Type:        valueType,
		Function:    lowering.functionIndex,
		Placeholder: lowering.placeholderOrigin(name),
		Uninitialized: lowering.placeholderOrigin(name) != "" && (name.Parent.Kind != ast.KindParameter ||
			lowering.uninitializedInitializer(name.Parent.AsParameterDeclaration().Initializer)),
	})
	proven := lowering.checker.GetTypeAtLocation(name)
	if lowering.detachedOwnAlias(name) != nil {
		// This local holds only the readiness marker. The intrinsic cannot escape,
		// so cycle analysis must not treat it as a user closure capturing cells.
		proven = lowering.checker.GetBooleanType()
	} else if inferred != nil {
		proven = inferred
	}
	lowering.noteLocal(lowering.locals[symbol], proven, name)
	return lowering.locals[symbol], nil
}

// noteAlso keeps another checker type a local has: a this shared by more than one instantiation.
func (lowering *lowering) noteAlso(local int, proven *checker.Type) {
	if lowering.localAlso == nil {
		lowering.localAlso = map[int][]*checker.Type{}
	}
	lowering.localAlso[local] = append(lowering.localAlso[local], proven)
}

// noteLocal keeps a local's checker type and declaration for the cycle finder.
func (lowering *lowering) noteLocal(local int, proven *checker.Type, node *ast.Node) {
	if lowering.localTypes == nil {
		lowering.localTypes, lowering.localNodes = map[int]*checker.Type{}, map[int]*ast.Node{}
	}
	lowering.localTypes[local], lowering.localNodes[local] = lowering.concrete(proven), node
	lowering.result.Locals[local].Source = lowering.sourceIdentity(node)
	if lowering.instance != nil {
		lowering.instance.templates = append(lowering.instance.templates, template{local: local, proven: proven})
	}
}

// symbol is what a name refers to, the same whichever file names it: an import resolves to what it
// imports, and an exported declaration to its export symbol, so one declaration is one symbol.
func (lowering *lowering) symbol(node *ast.Node) *ast.Symbol {
	symbol := lowering.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = lowering.checker.GetAliasedSymbol(symbol)
	}
	return lowering.checker.GetExportSymbolOfSymbol(symbol)
}

// local finds the local an identifier refers to, and notes a capture if it belongs to another
// function.
func (lowering *lowering) local(identifier *ast.Node) (int, bool) {
	return lowering.localOf(identifier, lowering.symbol(identifier))
}

// localOf is local for a symbol the identifier refers to other than the one it names: in
// { a } = value, a names the pattern's property, and the variable written is its value symbol.
func (lowering *lowering) localOf(identifier *ast.Node, symbol *ast.Symbol) (int, bool) {
	local, isLocal := lowering.locals[symbol]
	if isLocal {
		lowering.touch(local)
		if declared := lowering.result.Locals[local]; declared.Placeholder != "" &&
			declared.Captured &&
			declared.Preallocated &&
			declared.Ready == 0 &&
			lowering.unlowerable == nil {
			lowering.unlowerable = lowering.notYet(
				identifier,
				"an unset-capable capture before its declaration without a separate declaration-readiness cell",
			)
		}
		if declared := lowering.result.Locals[local]; declared.Captured &&
			slotless(declared.Type) &&
			!(declared.Placeholder != "" && declared.Type == ir.Union) &&
			lowering.unlowerable == nil {
			// A placeholder union is one counted heap reference, which a cell already holds.
			// Other slotless shapes retain their existing capture refusal. Lower says so once
			// it's done, since the capture is found here, where nothing can return an error.
			lowering.unlowerable = lowering.notYet(
				identifier,
				"a "+typeName(declared.Type)+" variable a function value captures",
			)
		}
	}
	return local, isLocal
}

// touch notes that the function being lowered reads or writes a local. One declared in another
// function, and not a global, is captured: it moves into a cell, and every closure being lowered
// between its function and this one carries that cell in its environment.
func (lowering *lowering) touch(local int) {
	declared := lowering.result.Locals[local]
	if declared.Ready != 0 {
		lowering.touch(declared.Ready - 1)
	}
	if declared.Global || declared.Function == lowering.functionIndex {
		return
	}
	lowering.result.Locals[local].Captured = true
	if name, isInitializing := lowering.initializing[local]; isInitializing && lowering.unlowerable == nil {
		// const f = () => f(): the function value is made before the variable's cell is, so it has
		// nothing to capture yet.
		lowering.unlowerable = lowering.notYet(
			name,
			"a function value that captures the variable its own initializer declares",
		)
	}
	start := 0
	for position, closure := range lowering.closures {
		if closure == declared.Function {
			start = position + 1
		}
	}
	for _, closure := range lowering.closures[start:] {
		function := &lowering.result.Functions[closure]
		if !slices.Contains(function.Environment, local) {
			function.Environment = append(function.Environment, local)
		}
	}
}

// checked guards switch lexical bindings, captured cells, and globals reached before initialization.
func (lowering *lowering) checked(local int) bool {
	return lowering.result.Locals[local].NamespaceState ||
		lowering.result.Locals[local].Ready != 0 ||
		(lowering.result.Locals[local].Global && (lowering.function != nil || lowering.cyclicModules)) ||
		(lowering.function != nil &&
			lowering.result.Locals[local].Preallocated &&
			lowering.result.Locals[local].Captured)
}

func (lowering *lowering) constant(value string) int {
	if lowering.strings == nil {
		lowering.strings = map[string]int{}
	}
	if index, isKnown := lowering.strings[value]; isKnown {
		return index
	}
	lowering.strings[value] = len(lowering.result.Strings)
	lowering.result.Strings = append(lowering.result.Strings, value)
	return lowering.strings[value]
}

// localRead preserves checker narrowing for both private and qualified singleton reads.
func (lowering *lowering) localRead(node *ast.Node, local int) (ir.Expression, error) {
	lowering.narrowedRead(node, local)
	if origin := lowering.result.Locals[local].Placeholder; origin != "" {
		return lowering.placeholderRead(node, ir.Read{
			Local:   local,
			Of:      lowering.result.Locals[local].Type,
			Checked: lowering.result.Locals[local].NamespaceState || lowering.checkedModuleRead(node, local),
			Unset:   true,
		}, origin), nil
	}
	read := ir.Expression(ir.Read{
		Local:     local,
		Of:        lowering.result.Locals[local].Type,
		Checked:   lowering.result.Locals[local].NamespaceState || lowering.checkedModuleRead(node, local),
		Readiness: sourceExpression(node),
	})
	if origin := lowering.result.Locals[local].UnsetField; origin != "" {
		return lowering.unsetSpecialization(node, read, lowering.checker.GetTypeOfSymbol(lowering.symbol(node)), origin)
	}
	if lowering.result.Locals[local].Type == ir.Union {
		if lowering.nullableObjectUnion(lowering.checker.GetTypeOfSymbol(lowering.symbol(node))) {
			return lowering.checkedNullableObject(node, read), nil
		}
		// Where the checker has narrowed it to fewer members held one way, it's read as that.
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		observing := comparedWithUndefined(node) ||
			(parent != nil && parent.Kind == ast.KindTypeOfExpression) ||
			unionEqualityObservation(node) ||
			lowering.checker.GetTypeAtLocation(node).Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
		if narrowed, isKnown := lowering.representation(
			lowering.checker.GetTypeAtLocation(node),
		); isKnown && narrowed != ir.Union && !observing {
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
			nullable, excludesArray := false, false
			if name == "object" {
				// typeof cannot distinguish object members: null, another class, a typed array
				// or a promise all say object. Beside the narrowed members, only undefined, null
				// (tested for below) and the kinds typeof tells from an object may sit, so any
				// other member, or one whose holding isn't known, stops here. Arrays use their
				// own brand check; a class beside an array also excludes that brand below.
				declared := lowering.concrete(lowering.checker.GetTypeOfSymbol(lowering.symbol(node)))
				members := []*checker.Type{declared}
				if declared.Flags()&checker.TypeFlagsUnion != 0 {
					members = declared.Types()
				}
				observed := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
				kept := []*checker.Type{observed}
				if observed.Flags()&checker.TypeFlagsUnion != 0 {
					kept = observed.Types()
				}
				for _, member := range members {
					held, known := lowering.representation(member)
					if narrowed == ir.Array {
						// Tuples and erased views can hide a layout other than adamic_array.
						if !lowering.arrayPredicateDomain(member) {
							return nil, lowering.notYet(
								node,
								"a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables",
							)
						}
						continue
					}
					if narrowed == ir.Object && isClassInstance(observed) && known && held == ir.Array {
						excludesArray = true
						continue
					}
					if known && held == narrowed && (slices.Contains(kept, member) ||
						member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0) {
						// A narrowed member, or undefined held as a missing object, which typeof says
						// undefined for.
						continue
					}
					if known && held == narrowed && member.Flags()&checker.TypeFlagsNull != 0 {
						nullable = true
						continue
					}
					if known {
						switch held.Present() {
						case ir.Number, ir.Boolean, ir.String, ir.Closure, ir.UndefinedRepresentation:
							continue
						}
					}
					return nil, lowering.notYet(
						node,
						"a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables",
					)
				}
			}
			builder := lowering.libraryArrayBuilder([]ir.Expression{read})
			held := builder.read(builder.parameters[0])
			matches := ir.Expression(ir.Binary{
				Operator: ir.Equal,
				Left:     ir.TypeOf{Value: held},
				Right:    ir.StringConstant{Index: lowering.constant(name)},
			})
			if narrowed == ir.Array {
				matches = ir.ArrayIsArray{Value: held}
			}
			if excludesArray {
				matches = ir.Binary{
					Operator: ir.And,
					Left:     matches,
					Right:    ir.Unary{Operator: ir.Not, Operand: ir.ArrayIsArray{Value: held}},
				}
			}
			if nullable {
				// typeof says object for null too.
				matches = ir.Binary{
					Operator: ir.And,
					Left:     matches,
					Right:    ir.Unary{Operator: ir.Not, Operand: ir.IsNull{Value: held}},
				}
			}
			if lowering.includesUndefined(lowering.checker.GetTypeAtLocation(node)) {
				matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
			}
			message := "union member where the checker narrowed it away: a call since the narrowing put it back"
			builder.body = append(builder.body, ir.If{
				Condition: ir.Unary{Operator: ir.Not, Operand: matches},
				Then:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: lowering.constant(message)}}},
			})
			read = builder.finish("narrowed_union_member", ir.Narrow{Value: held, To: narrowed})
			lowering.result.Functions[builder.function].CheckedUnionNarrow = true
		}
	}
	if declared := lowering.result.Locals[local].Type; declared.IsMaybe() {
		// Where the checker has narrowed it to what it holds, it's read as that.
		if narrowed, _ := lowering.representation(
			lowering.checker.GetTypeAtLocation(node),
		); narrowed == declared.Present() && !lowering.acceptsUndefined(node) {
			read = ir.Unwrap{Value: read}
		}
	}
	return lowering.defined(node, read), nil
}
