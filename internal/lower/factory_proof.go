package lower

import "github.com/system-inc/adamic/internal/ir"

type factoryState struct {
	aliases     map[int]int
	fields      map[int]map[string]bool
	escaped     map[int]map[string]bool
	returns     []map[string]bool
	unsupported bool
}

func factoryCopyFields(fields map[string]bool) map[string]bool {
	result := map[string]bool{}
	for name, ready := range fields {
		result[name] = ready
	}
	return result
}
func factoryIntersection(left, right map[string]bool) map[string]bool {
	result := factoryCopyFields(left)
	for name := range result {
		result[name] = result[name] && right[name]
	}
	return result
}
func (state *factoryState) clone() *factoryState {
	result := &factoryState{aliases: map[int]int{}, fields: map[int]map[string]bool{}, escaped: map[int]map[string]bool{}, unsupported: state.unsupported}
	for local, root := range state.aliases {
		result.aliases[local] = root
	}
	for root, fields := range state.fields {
		result.fields[root] = factoryCopyFields(fields)
	}
	for root, fields := range state.escaped {
		result.escaped[root] = factoryCopyFields(fields)
	}
	return result
}
func (state *factoryState) root(value ir.Expression) (int, bool) {
	if read, ok := value.(ir.Read); ok {
		root, found := state.aliases[read.Local]
		return root, found
	}
	return 0, false
}
func (state *factoryState) escape(root int) {
	fields := state.fields[root]
	if earlier, ok := state.escaped[root]; ok {
		fields = factoryIntersection(earlier, fields)
	}
	state.escaped[root] = factoryCopyFields(fields)
}

// A property read observes its value, not its receiver's entire construction.
// Every other use of an object reference is conservatively an escape. Calls
// invalidate local completion facts, including possible captured aliases.
func (state *factoryState) expression(value ir.Expression) {
	if value == nil {
		return
	}
	walk(value, func(node any) bool {
		if property, ok := node.(ir.Property); ok {
			if _, direct := state.root(property.Object); !direct {
				state.expression(property.Object)
			}
			return false
		}
		if read, ok := node.(ir.Read); ok {
			if root, found := state.aliases[read.Local]; found {
				state.escape(root)
			}
		}
		switch node.(type) {
		case ir.Call, ir.CallClosure, ir.MakeClosure:
			state.unsupported = true
		}
		return true
	})
}

// FactoryCompletion reports fields written before every escape of a fresh
// returned object. Unknown control or effects retain checks. It never infers
// completion from an asserted type, and different aliases share one root.
func FactoryCompletion(program *ir.Program) map[int]map[string]bool {
	result := map[int]map[string]bool{}
	for index, function := range program.Functions {
		if function.Returns != ir.Object {
			continue
		}
		state := &factoryState{aliases: map[int]int{}, fields: map[int]map[string]bool{}, escaped: map[int]map[string]bool{}}
		factoryStatements(program, state, function.Body)
		if state.unsupported || len(state.returns) == 0 {
			continue
		}
		ready := factoryCopyFields(state.returns[0])
		for _, returned := range state.returns[1:] {
			ready = factoryIntersection(ready, returned)
		}
		result[index] = ready
	}
	return result
}

func factoryStatements(program *ir.Program, state *factoryState, body []ir.Statement) bool {
	for _, statement := range body {
		switch statement := statement.(type) {
		case ir.Declare:
			if literal, fresh := statement.Value.(ir.ObjectLiteral); fresh && literal.Spread == nil && literal.Class == 0 && !program.Locals[statement.Local].Captured {
				fields := map[string]bool{}
				for _, field := range literal.Fields {
					state.expression(field.Value)
					fields[field.Name] = !field.Uninitialized
				}
				state.aliases[statement.Local] = statement.Local
				state.fields[statement.Local] = fields
			} else if root, found := state.root(statement.Value); found && !program.Locals[statement.Local].Captured && !program.Locals[statement.Local].Global {
				state.aliases[statement.Local] = root
			} else {
				state.expression(statement.Value)
				delete(state.aliases, statement.Local)
			}
		case ir.Assign:
			root, found := state.root(statement.Value)
			if found && !program.Locals[statement.Local].Captured && !program.Locals[statement.Local].Global {
				state.aliases[statement.Local] = root
			} else {
				state.expression(statement.Value)
				delete(state.aliases, statement.Local)
			}
		case ir.SetProperty:
			state.expression(statement.Value)
			if root, found := state.root(statement.Object); found {
				state.fields[root][statement.Name] = !statement.Uninitialized
			} else {
				state.expression(statement.Object)
			}
		case ir.Evaluate:
			state.expression(statement.Value)
		case ir.Return:
			if root, found := state.root(statement.Value); found {
				state.escape(root)
				state.returns = append(state.returns, factoryCopyFields(state.escaped[root]))
			} else {
				state.returns = append(state.returns, map[string]bool{})
				state.expression(statement.Value)
			}
			return false
		case ir.If:
			state.expression(statement.Condition)
			left, right := state.clone(), state.clone()
			leftLives := factoryStatements(program, left, statement.Then)
			rightLives := factoryStatements(program, right, statement.Else)
			state.returns = append(state.returns, left.returns...)
			state.returns = append(state.returns, right.returns...)
			state.unsupported = state.unsupported || left.unsupported || right.unsupported
			for root, fields := range left.escaped {
				if other, ok := state.escaped[root]; ok {
					fields = factoryIntersection(fields, other)
				}
				state.escaped[root] = fields
			}
			for root, fields := range right.escaped {
				if other, ok := state.escaped[root]; ok {
					fields = factoryIntersection(fields, other)
				}
				state.escaped[root] = fields
			}
			if !leftLives && !rightLives {
				return false
			}
			active := left
			if !leftLives {
				active = right
			}
			state.aliases = active.aliases
			state.fields = active.fields
			if leftLives && rightLives {
				for local, root := range state.aliases {
					if other, ok := right.aliases[local]; !ok || other != root {
						delete(state.aliases, local)
					}
				}
				for root, fields := range state.fields {
					state.fields[root] = factoryIntersection(fields, right.fields[root])
				}
			}
		case ir.Block:
			if !factoryStatements(program, state, statement.Body) {
				return false
			}
		default:
			state.unsupported = true
		}
	}
	return true
}
