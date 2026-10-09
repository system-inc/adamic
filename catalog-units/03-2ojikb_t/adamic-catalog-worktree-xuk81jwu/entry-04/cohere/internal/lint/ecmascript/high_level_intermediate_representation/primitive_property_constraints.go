package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

type primitiveConstraintKey struct {
	function   *Function
	identifier static_single_assignment.IdentifierId
}

type primitiveConstraintKind uint8

const (
	primitiveConstraintUnknown primitiveConstraintKind = iota
	primitiveConstraintPrimitive
	primitiveConstraintOther
)

type primitiveConstraintState struct {
	parents  map[primitiveConstraintKey]primitiveConstraintKey
	kinds    map[primitiveConstraintKey]primitiveConstraintKind
	constant map[primitiveConstraintKey]bool
	visited  map[*Function]bool
}

func (state *primitiveConstraintState) root(key primitiveConstraintKey) primitiveConstraintKey {
	if parent, exists := state.parents[key]; exists {
		parent = state.root(parent)
		state.parents[key] = parent
		return parent
	}
	return key
}

func (state *primitiveConstraintState) bind(key primitiveConstraintKey, kind primitiveConstraintKind) {
	root := state.root(key)
	if state.kinds[root] == primitiveConstraintUnknown {
		state.kinds[root] = kind
	}
}

func (state *primitiveConstraintState) alias(left, right primitiveConstraintKey) {
	left, right = state.root(left), state.root(right)
	if left == right {
		return
	}
	leftKind, rightKind := state.kinds[left], state.kinds[right]
	if leftKind != primitiveConstraintUnknown && rightKind != primitiveConstraintUnknown && leftKind != rightKind {
		return
	}
	state.parents[left] = right
	if rightKind == primitiveConstraintUnknown {
		state.kinds[right] = leftKind
	}
}

func (state *primitiveConstraintState) visit(function *Function) {
	if function == nil || state.visited[function] {
		return
	}
	state.visited[function] = true
	key := func(place Place) primitiveConstraintKey { return primitiveConstraintKey{function, place.Identifier} }
	if function.Kind == FunctionKindComponent {
		for _, parameter := range function.Params {
			state.bind(key(parameter), primitiveConstraintOther)
		}
	}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, phi := range block.Phis {
			state.bind(key(phi.Place), primitiveConstraintOther)
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			left := key(instruction.LValue)
			switch value := instruction.Value.(type) {
			case *Primitive, *TemplateLiteral, *UnaryExpression:
				state.bind(left, primitiveConstraintPrimitive)
			case *LoadLocal:
				state.alias(left, key(value.Place))
			case *StoreLocal:
				state.alias(left, key(value.Value))
				state.alias(key(value.LValue), key(value.Value))
				if value.Kind == InstructionKindConst {
					state.constant[key(value.LValue)] = true
				}
			case *LoadContext:
				if state.constant[key(value.Place)] {
					state.alias(left, key(value.Place))
				}
			case *StoreContext:
				if value.Kind == InstructionKindConst {
					state.alias(key(value.LValue), key(value.Value))
					state.constant[key(value.LValue)] = true
				}
			case *BinaryExpression:
				switch value.Operator {
				case "+", "-", "/", "%", "*", "**", "&", "|", ">>", "<<", "^", ">", "<", ">=", "<=", "|>":
					state.bind(key(value.Left), primitiveConstraintPrimitive)
					state.bind(key(value.Right), primitiveConstraintPrimitive)
				}
				state.bind(left, primitiveConstraintPrimitive)
			case *PrefixUpdate:
				state.bind(key(value.Value), primitiveConstraintPrimitive)
				state.bind(key(value.LValue), primitiveConstraintPrimitive)
				state.bind(left, primitiveConstraintPrimitive)
			case *PostfixUpdate:
				state.bind(key(value.Value), primitiveConstraintPrimitive)
				state.bind(key(value.LValue), primitiveConstraintPrimitive)
				state.bind(left, primitiveConstraintPrimitive)
			case *LoadGlobal, *ObjectExpression, *ArrayExpression, *RegExpLiteral, *ObjectMethod, *NewExpression:
				state.bind(left, primitiveConstraintOther)
			case *PropertyLoad:
				if value.Property == "current" || state.kinds[state.root(key(value.Object))] != primitiveConstraintUnknown {
					state.bind(left, primitiveConstraintOther)
				}
			case *ComputedLoad:
				state.bind(left, primitiveConstraintOther)
			case *CallExpression:
				state.bind(key(value.Callee), primitiveConstraintOther)
				if value.CalleeOrigin.Module == "react" && (value.CalleeOrigin.Export == "useRef" || value.CalleeOrigin.Export == "useState" || value.CalleeOrigin.Export == "useReducer" || value.CalleeOrigin.Export == "useCallback") {
					state.bind(left, primitiveConstraintOther)
				}
			case *MethodCall:
				if value.CalleeOrigin.Module == "react" && (value.CalleeOrigin.Export == "useRef" || value.CalleeOrigin.Export == "useState" || value.CalleeOrigin.Export == "useReducer" || value.CalleeOrigin.Export == "useCallback") {
					state.bind(left, primitiveConstraintOther)
				}
			case *FunctionExpression:
				state.bind(left, primitiveConstraintOther)
				if int(value.Function) >= len(function.Functions) {
					continue
				}
				nested := function.Functions[value.Function]
				if nested != nil && len(nested.Context) == len(value.Captures) {
					for index, capture := range value.Captures {
						if state.constant[key(capture)] {
							context := primitiveConstraintKey{nested, nested.Context[index].Identifier}
							state.constant[context] = true
							state.alias(context, key(capture))
						}
					}
				}
				state.visit(nested)
			}
		}
	}
}

func inferPrimitivePropertyReads(function *Function) map[static_single_assignment.IdentifierId]bool {
	state := primitiveConstraintState{
		parents:  map[primitiveConstraintKey]primitiveConstraintKey{},
		kinds:    map[primitiveConstraintKey]primitiveConstraintKind{},
		constant: map[primitiveConstraintKey]bool{},
		visited:  map[*Function]bool{},
	}
	state.visit(function)
	result := map[static_single_assignment.IdentifierId]bool{}
	if function != nil {
		for _, instruction := range function.Instructions {
			if instruction == nil {
				continue
			}
			if _, property := instruction.Value.(*PropertyLoad); property && state.kinds[state.root(primitiveConstraintKey{function, instruction.LValue.Identifier})] == primitiveConstraintPrimitive {
				result[instruction.LValue.Identifier] = true
			}
		}
	}
	return result
}
