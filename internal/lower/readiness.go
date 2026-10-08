package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"reflect"
	"slices"
)

func sourceExpression(node *ast.Node) string {
	file := ast.GetSourceFileOfNode(node)
	return file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
}

// The two builtin literal assertions reserve or deinitialize a slot.
// A shadowed identifier named undefined remains an ordinary assertion.
func (l *lowering) uninitializedInitializer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindNonNullExpression {
		return false
	}
	operand := ast.SkipParentheses(node.AsNonNullExpression().Expression)
	if operand.Kind == ast.KindNullKeyword {
		return true
	}
	if operand.Kind != ast.KindIdentifier || operand.Text() != "undefined" {
		return false
	}
	symbol := l.symbol(operand)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return true
	}
	return load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

func (l *lowering) uninitializedDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return node.AsVariableDeclaration().ExclamationToken != nil || l.uninitializedInitializer(node.AsVariableDeclaration().Initializer)
	case ast.KindPropertyDeclaration:
		return node.PostfixToken() != nil && node.PostfixToken().Kind == ast.KindExclamationToken || l.uninitializedInitializer(node.AsPropertyDeclaration().Initializer)
	}
	return false
}

// readiness uses the existing CFG, including its exceptional edges. Literal assertion
// assignments clear readiness; calls invalidate slots that a captured writer can clear.
// Captures and globals participate in this bit analysis even though value SSA excludes them.
func readiness(program *ir.Program) {
	prepareViewArrayReferenceWrites(program)
	objectWrites := false
	inspect := func(value any) bool {
		if write, ok := value.(ir.SetProperty); ok && write.WriteContract != 0 && write.Value.Type() == ir.Object {
			objectWrites = true
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if objectWrites {
		for _, contract := range program.ViewContracts {
			for _, field := range contract.Fields {
				program.CheckedFields[field.Name] = true
			}
		}
	}
	program.FreshViewWrites = map[int]bool{}
	if len(program.CheckedFields) != 0 {
		for _, proof := range fresh.ProveWrites(program) {
			if proof.Kind == fresh.WriteField && proof.Site != 0 {
				previous, known := program.FreshViewWrites[proof.Site]
				program.FreshViewWrites[proof.Site] = proof.Unaliased && (!known || previous)
			}
		}
	}
	eraseTagViewWrites(program)
	defer eraseProvenViewChecks(program)
	deinitialized := map[int]bool{}
	collect := func(node any) bool {
		if assign, ok := node.(ir.Assign); ok && assign.Uninitialized {
			deinitialized[assign.Local] = true
		}
		return true
	}
	walk(program.Main, collect)
	for _, function := range program.Functions {
		walk(function.Body, collect)
	}
	invalidate := func(state []bool) {
		for local := range deinitialized {
			if program.Locals[local].Captured || program.Locals[local].Global {
				state[local] = false
			}
		}
	}
	names := map[string]bool{}
	walk(program.Main, func(node any) bool {
		if value, ok := node.(ir.ObjectLiteral); ok {
			for _, field := range value.Fields {
				if field.Uninitialized {
					names[field.Name] = true
				}
			}
		}
		if value, ok := node.(ir.SetProperty); ok && value.Uninitialized {
			names[value.Name] = true
		}
		return true
	})
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if value, ok := node.(ir.ObjectLiteral); ok {
				for _, field := range value.Fields {
					if field.Uninitialized {
						names[field.Name] = true
					}
				}
			}
			if value, ok := node.(ir.SetProperty); ok && value.Uninitialized {
				names[value.Name] = true
			}
			return true
		})
	}
	marked := len(names) != 0
	for _, local := range program.Locals {
		marked = marked || local.Uninitialized
	}
	if !marked {
		clear := func(body []ir.Statement) []ir.Statement {
			for index, statement := range body {
				body[index] = readinessStatement(statement, program, names, nil, nil)
			}
			return body
		}
		program.Main = clear(program.Main)
		for index := range program.Functions {
			program.Functions[index].Body = clear(program.Functions[index].Body)
		}
		return
	}
	for function := -1; function < len(program.Functions); function++ {
		// The graph's locations still point into the original body, while the local metadata
		// copy lets its existing builder track cells and globals without changing value SSA.
		tracked := *program
		tracked.Locals = slices.Clone(program.Locals)
		for i := range tracked.Locals {
			tracked.Locals[i].Global, tracked.Locals[i].Captured = false, false
		}
		graph := flow.Build(&tracked, function)
		fieldSlots := map[fieldReadiness]int{}
		statements := program.Main
		if function >= 0 {
			statements = program.Functions[function].Body
		}
		walk(statements, func(node any) bool {
			if property, ok := node.(ir.Property); ok && names[property.Name] {
				if local, ok := readinessObject(property.Object); ok {
					key := fieldReadiness{local, property.Name}
					if _, found := fieldSlots[key]; !found {
						fieldSlots[key] = len(program.Locals) + len(fieldSlots)
					}
				}
			}
			return true
		})
		count := len(program.Locals) + len(fieldSlots)
		ins, outs, exceptional := map[flow.BlockId][]bool{}, map[flow.BlockId][]bool{}, map[flow.BlockId][]bool{}
		for _, block := range graph.Blocks {
			ins[block.Id], outs[block.Id], exceptional[block.Id] = make([]bool, count), make([]bool, count), make([]bool, count)
			// Must analysis starts at top away from entry; loops converge by intersection.
			if block.Id != graph.Entry {
				for i := 0; i < count; i++ {
					ins[block.Id][i], outs[block.Id][i], exceptional[block.Id][i] = true, true, true
				}
			}
		}
		entry := make([]bool, count)
		for i, local := range program.Locals {
			entry[i] = !local.Uninitialized
		}
		if function >= 0 {
			for _, parameter := range program.Functions[function].Parameters {
				entry[parameter] = true
			}
		}
		changed := true
		for changed {
			changed = false
			for _, block := range graph.Blocks {
				state := slices.Clone(entry)
				if block.Id != graph.Entry {
					for i := range state {
						state[i] = true
					}
					for _, pred := range block.Predecessors {
						previous := outs[pred]
						parent, _ := graph.Block(pred)
						if edge, ok := parent.Terminal.(*flow.MayThrow); ok && edge.Handler == block.Id {
							previous = exceptional[pred]
						}
						for i := range state {
							state[i] = state[i] && previous[i]
						}
					}
				}
				if !slices.Equal(ins[block.Id], state) {
					ins[block.Id] = slices.Clone(state)
					changed = true
				}
				beforeLast := slices.Clone(state)
				for _, id := range block.Instructions {
					instruction := graph.Instructions[id]
					if readinessCalls(instruction) {
						invalidate(state)
						for _, slot := range fieldSlots {
							state[slot] = false
						}
					}
					beforeLast = slices.Clone(state)
					readinessWrite(program, graph, instruction, state, fieldSlots)
				}
				if !slices.Equal(outs[block.Id], state) {
					outs[block.Id] = state
					changed = true
				}
				if !slices.Equal(exceptional[block.Id], beforeLast) {
					exceptional[block.Id] = beforeLast
					changed = true
				}
			}
		}
		for _, block := range graph.Blocks {
			state := slices.Clone(ins[block.Id])
			for _, id := range block.Instructions {
				instruction := graph.Instructions[id]
				if readinessCalls(instruction) {
					invalidate(state)
					for _, slot := range fieldSlots {
						state[slot] = false
					}
				}
				if instruction.Part == 0 {
					*instruction.At = readinessStatement(*instruction.At, program, names, state, fieldSlots)
				}
				readinessWrite(program, graph, instruction, state, fieldSlots)
			}
		}
	}
}

func readinessWrite(program *ir.Program, graph *flow.Function, instruction *flow.Instruction, state []bool, fields map[fieldReadiness]int) {
	for _, place := range instruction.Defines {
		local := int(graph.Identifiers[place.Identifier].Declaration) - 1
		ready := true
		if declare, ok := (*instruction.At).(ir.Declare); ok {
			ready = !declare.Uninitialized
		}
		if assign, ok := (*instruction.At).(ir.Assign); ok {
			ready = !assign.Uninitialized
		}
		state[local] = ready
		for field, slot := range fields {
			if field.local == local {
				state[slot] = false
			}
		}
	}
	if set, ok := (*instruction.At).(ir.SetProperty); ok {
		if set.Uninitialized {
			for field, slot := range fields {
				if field.name == set.Name {
					state[slot] = false
				}
			}
		}
		if local, ok := readinessObject(set.Object); ok {
			if slot, found := fields[fieldReadiness{local, set.Name}]; found {
				state[slot] = !set.Uninitialized
			}
		}
	}
}

// Transform just this instruction's expressions. Nested statements have their own graph locations.
func readinessStatement(statement ir.Statement, program *ir.Program, fields map[string]bool, ready []bool, fieldSlots map[fieldReadiness]int) ir.Statement {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Interface {
			if value.IsNil() || ready != nil && value.Type() == statementType {
				return value
			}
			mapped := transform(value.Elem())
			node := mapped.Interface()
			switch expression := node.(type) {
			case ir.Read:
				if program.Locals[expression.Local].Uninitialized && !ready[expression.Local] {
					if expression.Readiness == "" {
						expression.Readiness = program.Locals[expression.Local].Name
					}
					if origin := program.Locals[expression.Local].InitializerExpression; origin != "" {
						expression.Readiness = origin
					}
					expression.Checked = false
				} else {
					expression.Readiness = ""
					if program.Locals[expression.Local].Uninitialized {
						expression.Checked = false
					}
				}
				node = expression
			case ir.ArrayIndex:
				node = markProgramViewArrayRead(program, expression)
			case ir.ArraySearch:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ArrayJoin:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ArrayMap:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ArrayVisit:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ArrayReduce:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ArrayPop:
				expression.ViewRead = markProgramViewArrayUse(program, expression.ViewRead)
				node = expression
			case ir.ObjectLiteral:
				if expression.Spread != nil && len(fields) > 0 {
					expression.NoReuse = true
					if expression.SpreadReadiness == "" {
						expression.SpreadReadiness = "object spread"
					}
				} else {
					expression.SpreadReadiness = ""
				}
				node = expression
			case ir.SetProperty:
				node = eraseFreshViewWrite(program, expression)
			case ir.ObjectCall:
				if len(fields) == 0 || (expression.Method != "values" && expression.Method != "entries" && expression.Method != "assign") {
					expression.Readiness = ""
				}
				node = expression
			case ir.Property:
				if expression.DictionaryKey == nil && !program.CheckedFields[expression.Name] && !primitiveBindingConversion(program, expression) {
					expression.View = ""
					expression.ViewType = ""
					expression.ViewAllowed = nil
					expression.ViewContract = 0
					expression.ViewTypeID = 0
				} else if expression.ViewTypeID != 0 && !expression.DictionaryPrimitive {
					// A function read can precede the cast that interns its contract.
					expression.ViewContract = program.ViewContractTypes[expression.ViewTypeID]
				}
				proven := false
				if local, ok := readinessObject(expression.Object); ok {
					if slot, found := fieldSlots[fieldReadiness{local, expression.Name}]; found {
						proven = ready[slot]
					}
				}
				if fields[expression.Name] && !proven {
					if expression.Readiness == "" {
						expression.Readiness = expression.Name
					}
				} else {
					expression.Readiness = ""
				}
				node = expression
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(node))
			return result
		}
		switch value.Kind() {
		case reflect.Pointer:
			// Dictionary read adapters share the same receiver readiness facts
			// as their enclosing record operation, including generic parameters.
			if value.Type() != reflect.TypeOf((*ir.Property)(nil)) || value.IsNil() {
				return value
			}
			result := reflect.New(value.Type().Elem())
			result.Elem().Set(transform(value.Elem()))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			if loop, ok := result.Interface().(ir.ForOf); ok {
				loop.ViewRead = markProgramViewArrayUse(program, loop.ViewRead)
				result.Set(reflect.ValueOf(loop))
			}
			return result
		case reflect.Array:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() || ready != nil && value.Type() == reflect.TypeOf([]ir.Statement{}) {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		}
		return value
	}
	result := transform(reflect.ValueOf(statement)).Interface().(ir.Statement)
	if write, ok := result.(ir.SetProperty); ok {
		result = eraseFreshViewWrite(program, write)
	}
	if loop, ok := result.(ir.ForOf); ok {
		loop.ViewRead = markProgramViewArrayUse(program, loop.ViewRead)
		result = loop
	}
	if assign, ok := result.(ir.Assign); ok && program.Locals[assign.Local].Uninitialized {
		assign.Checked = false
		result = assign
	}
	return result
}

// A field fact belongs to one binding's current object, never just to its field name.
type fieldReadiness struct {
	local int
	name  string
}

func readinessObject(value ir.Expression) (int, bool) {
	switch value := value.(type) {
	case ir.Read:
		return value.Local, true
	case ir.Defined:
		return readinessObject(value.Value)
	case ir.Narrow:
		return readinessObject(value.Value)
	}
	return 0, false
}
func readinessCalls(instruction *flow.Instruction) bool {
	var node any = instruction.Expression
	if node == nil {
		switch statement := (*instruction.At).(type) {
		case ir.SetProperty, ir.SetIndex:
			node = statement
		}
	}
	return readinessHasCalls(node)
}

func readinessHasCalls(node any) bool {
	calls := false
	walk(node, func(node any) bool {
		switch node.(type) {
		case ir.ObjectCall, ir.Call, ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArraySort, ir.ArrayFrom, ir.MapForEach:
			calls = true
		}
		return true
	})
	return calls
}

// assertionInitializer recognizes syntax only; its operand is evaluated at the declaration.
func assertionInitializer(node *ast.Node) bool {
	return node != nil && ast.SkipParentheses(node).Kind == ast.KindNonNullExpression
}

func assertionVarList(list *ast.Node) bool {
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		if !ast.IsIdentifier(declaration.Name()) || !assertionInitializer(declaration.AsVariableDeclaration().Initializer) && declaration.AsVariableDeclaration().ExclamationToken == nil {
			return false
		}
	}
	return len(list.AsVariableDeclarationList().Declarations.Nodes) > 0
}

// lazyAssertion holds the original representation once and assigns only when present.
func (l *lowering) lazyAssertion(node *ast.Node, to ir.Type) ([]ir.Statement, ir.Expression, ir.Expression, error) {
	operand := ast.SkipParentheses(node).AsNonNullExpression().Expression
	value, err := l.expression(operand)
	if err != nil {
		return nil, nil, nil, err
	}
	// Initializer assertions use deferred readiness checks, but still belong in the check report.
	l.recordNonNullCheck(node, !value.Type().IsMaybe() && !l.includesUndefined(l.checker.GetTypeAtLocation(operand)) && !l.includesNull(l.checker.GetTypeAtLocation(operand)))
	for {
		switch narrowed := value.(type) {
		case ir.Unwrap:
			value = narrowed.Value
		case ir.Defined:
			value = narrowed.Value
		case ir.Narrow:
			value = narrowed.Value
		default:
			goto stored
		}
	}
stored:
	if weak, ok := value.(ir.WeakTarget); ok {
		weak.Present = false
		value = weak
	}
	if !value.Type().IsMaybe() && !value.Type().IsReference() {
		return nil, ir.BooleanConstant{Value: true}, fit(value, to), nil
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "assertion_initializer", Type: value.Type(), Function: l.functionIndex})
	read := ir.Read{Local: local, Of: value.Type()}
	present := ir.Expression(ir.BooleanConstant{Value: true})
	if value.Type().IsMaybe() {
		present = ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: read}}
	} else if value.Type().IsReference() {
		present = ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}}
	}
	assigned := fit(read, to)
	if read.Type() == ir.Union && to != ir.Union {
		assigned = ir.Narrow{Value: read, To: to}
	}
	return []ir.Statement{ir.Declare{Local: local, Value: value}}, present, assigned, nil
}

func eraseFreshViewWrite(program *ir.Program, write ir.SetProperty) ir.SetProperty {
	if write.WriteContract == 0 || !program.FreshViewWrites[write.Site] {
		return write
	}
	literal, ok := write.Object.(ir.ObjectLiteral)
	if !ok || literal.Spread != nil || literal.Class != 0 {
		return write
	}
	found := false
	for _, field := range literal.Fields {
		found = found || field.Name == write.Name
	}
	if !found {
		literal.Fields = append(literal.Fields, ir.Field{Name: write.Name, Value: zeroValue(freshViewSlotType(program, write)), Contract: write.TargetContract, Uninitialized: true})
	}
	write.Object, write.WriteProven = literal, true
	return write
}

func freshViewSlotType(program *ir.Program, write ir.SetProperty) ir.Type {
	if write.TargetContract != 0 {
		of := program.ViewContracts[write.TargetContract-1].Of
		if of == ir.MaybeBoolean {
			return ir.Boolean
		}
		return of
	}
	return write.Value.Type()
}

// The stored value is inaccessible until a subsequent write sets readiness.
func uninitializedValue(of ir.Type) ir.Expression {
	if of.IsMaybe() {
		return ir.MaybeOf{Of: of}
	}
	switch of {
	case ir.Number:
		return ir.NumberConstant{}
	case ir.Boolean:
		return ir.BooleanConstant{}
	}
	return ir.Undefined{Of: of}
}
