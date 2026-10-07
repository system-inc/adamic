package lower

import (
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Descriptor changes stay behind one library entry point. A shape slot retains
// its checker-proven type; added scalar keys are observable only by reflection.
func (l *lowering) objectDescriptorCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name != "defineProperty" && name != "defineProperties" {
		return nil, false, nil
	}
	refused := func(reason string) (ir.Expression, bool, error) {
		return nil, true, &Refused{Where: l.program.Where(node), What: "Object." + name, Fix: "property descriptors: " + reason}
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	count := 3
	if name == "defineProperties" {
		count = 2
	}
	if len(arguments) != count || hasSpread(node) {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	if hazard := l.objectDescriptorHazard(); hazard != "" {
		return nil, true, l.notYet(node, "Object."+name+" with "+hazard)
	}
	targetType := l.checker.GetTypeAtLocation(arguments[0])
	primitive := targetType.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
	if !primitive && !l.exactObject(arguments[0], 0) {
		return nil, true, l.notYet(node, "Object."+name+" requires a proven complete plain data shape")
	}
	numericKey := name == "defineProperty" && l.checker.GetTypeAtLocation(arguments[1]).Flags()&checker.TypeFlagsNumberLike != 0 && len(l.checker.GetPropertiesOfType(targetType)) == 0
	if name == "defineProperty" && !primitive && !numericKey && ast.SkipParentheses(arguments[1]).Kind != ast.KindStringLiteral {
		return refused("keys must be constant strings; coercing or dynamic keys need a proof for every affected typed slot")
	}
	if name == "defineProperty" && !primitive && l.includesUndefined(l.checker.GetTypeAtLocation(arguments[2])) && l.checker.GetTypeAtLocation(arguments[2]).Flags()&checker.TypeFlagsUndefined == 0 {
		return refused("a possibly absent descriptor needs an explicit presence proof")
	}
	values := []ir.Expression{}
	for index, argument := range arguments {
		// A literal null receiver always throws before descriptor conversion.
		// Its argument has no effects and no null value escapes this helper.
		if index == 0 && primitive && ast.SkipParentheses(argument).Kind == ast.KindNullKeyword {
			values = append(values, ir.Undefined{})
			continue
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		values = append(values, value)
	}
	if !primitive && values[0].Type() != ir.Object {
		return nil, true, l.notYet(node, "Object."+name+" on other than a plain object")
	}
	function := len(l.result.Functions)
	parameters := []int{}
	for _, value := range values {
		local := len(l.result.Locals)
		parameters = append(parameters, local)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "descriptor_argument", Type: value.Type(), Function: function})
	}
	receiver := ir.Read{Local: parameters[0], Of: values[0].Type()}
	body := []ir.Statement{}
	if primitive {
		body = append(body, ir.Throw{Value: ir.ObjectCall{Method: "typeError", Arguments: []ir.Expression{ir.StringConstant{Index: l.constant("Object." + name + " called on non-object")}}, Returns: ir.Object}})
	} else if name == "defineProperty" && (values[2].Type() == ir.Number || values[2].Type() == ir.String || values[2].Type() == ir.Boolean || l.checker.GetTypeAtLocation(arguments[2]).Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0) {
		var description ir.Expression
		switch values[2].Type() {
		case ir.Number:
			description = ir.NumberToString{Value: ir.Read{Local: parameters[2], Of: ir.Number}}
		case ir.Boolean:
			description = ir.BooleanToString{Value: ir.Read{Local: parameters[2], Of: ir.Boolean}}
		case ir.String:
			description = ir.Read{Local: parameters[2], Of: ir.String}
		default:
			word := "undefined"
			if l.checker.GetTypeAtLocation(arguments[2]).Flags()&checker.TypeFlagsNull != 0 {
				word = "null"
			}
			description = ir.StringConstant{Index: l.constant(word)}
		}
		message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Property description must be an object: ")}, description}}
		body = append(body, ir.Throw{Value: ir.ObjectCall{Method: "typeError", Arguments: []ir.Expression{message}, Returns: ir.Object}})
	} else {
		type definition struct {
			key        string
			descriptor *ast.Node
			value      ir.Expression
			keyValue   ir.Expression
		}
		definitions := []definition{}
		if name == "defineProperty" {
			key := ast.SkipParentheses(arguments[1])
			if numericKey {
				// Number formatting cannot produce a prototype member name. The
				// initially empty receiver has no unboxed slot a dynamic key can
				// overwrite; any prior additions live in tagged descriptor storage.
				if values[1].Type() != ir.Number {
					return refused("numeric keys must have a proven number representation")
				}
				definitions = append(definitions, definition{"", arguments[2], ir.Read{Local: parameters[2], Of: values[2].Type()}, ir.NumberToString{Value: ir.Read{Local: parameters[1], Of: ir.Number}}})
			} else {
				definitions = append(definitions, definition{key.Text(), arguments[2], ir.Read{Local: parameters[2], Of: values[2].Type()}, nil})
			}
		} else {
			literal := l.objectDescriptorLiteral(arguments[1], 0)
			if literal == nil || l.objectDescriptorWasMutated(literal) {
				return refused("descriptor maps must be plain literals or unchanged literal bindings")
			}
			seen := map[string]bool{}
			for _, field := range literal.AsObjectLiteralExpression().Properties.Nodes {
				if field.Kind != ast.KindPropertyAssignment {
					return refused("descriptor maps require explicit data fields")
				}
				key := field.Name().Text()
				if seen[key] {
					return refused("duplicate descriptor map keys need presence and order tracking")
				}
				seen[key] = true
				of, known := l.representation(l.checker.GetTypeAtLocation(field.AsPropertyAssignment().Initializer))
				if !known {
					return refused("every descriptor must have a proven representation")
				}
				definitions = append(definitions, definition{key, field.AsPropertyAssignment().Initializer, ir.Property{Object: ir.Read{Local: parameters[1], Of: ir.Object}, Name: key, Of: of}, nil})
			}
			// OwnPropertyKeys visits canonical array indices before other strings.
			sort.SliceStable(definitions, func(a, b int) bool {
				x, xi := objectDescriptorIndex(definitions[a].key)
				y, yi := objectDescriptorIndex(definitions[b].key)
				if xi != yi {
					return xi
				}
				return xi && x < y
			})
		}
		for _, definition := range definitions {
			key := definition.key
			if key == "__proto__" || strings.ContainsRune(key, 0) || strings.HasPrefix(key, "#") {
				return refused("reserved or NUL keys cannot be represented as public shape names")
			}
			switch key {
			case "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf", "constructor":
				return refused("overriding inherited methods requires prototype dispatch")
			}
			literal := l.objectDescriptorLiteral(definition.descriptor, 0)
			if literal == nil || l.objectDescriptorWasMutated(literal) {
				return refused("descriptors must be plain literals or unchanged literal bindings")
			}
			descriptorType := l.checker.GetTypeAtLocation(definition.descriptor)
			fields := map[string]ir.Expression{}
			for _, field := range literal.AsObjectLiteralExpression().Properties.Nodes {
				if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
					return refused("accessor and method descriptors require language-level getter/setter support")
				}
				fieldName := field.Name().Text()
				if fieldName == "get" || fieldName == "set" {
					return refused("accessor descriptors require language-level getter/setter support")
				}
				symbol := l.checker.GetPropertyOfType(descriptorType, fieldName)
				if symbol == nil {
					return refused("every descriptor field must have a proven type")
				}
				fieldType := l.checker.GetTypeOfSymbol(symbol)
				if fieldType.Flags()&checker.TypeFlagsUndefined != 0 {
					fields[fieldName] = ir.Undefined{}
					continue
				}
				of, known := l.representation(fieldType)
				if !known {
					return refused("every descriptor field must have a proven representation")
				}
				fields[fieldName] = ir.Property{Object: definition.value, Name: fieldName, Of: of}
			}
			mask := 0
			var keyValue ir.Expression = ir.StringConstant{Index: l.constant(key)}
			if definition.keyValue != nil {
				keyValue = definition.keyValue
			}
			written := []ir.Expression{receiver, keyValue}
			representation := 0
			for index, fieldName := range []string{"value", "writable", "enumerable", "configurable"} {
				value, present := fields[fieldName]
				if !present {
					written = append(written, ir.Undefined{})
					continue
				}
				if index == 0 {
					fieldType := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(descriptorType, "value"))
					undefined := fieldType.Flags()&checker.TypeFlagsUndefined != 0
					if undefined {
						value = ir.Undefined{}
					}
					if !undefined && value.Type() != ir.Number && value.Type() != ir.Boolean && value.Type() != ir.String {
						return refused("data values must be proven number, boolean, string or undefined scalars")
					}
					if field := l.checker.GetPropertyOfType(targetType, key); field != nil {
						into := l.checker.GetTypeOfSymbol(field)
						from := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(descriptorType, "value"))
						slot, known := l.representation(into)
						if !known || slot != value.Type() || !l.checker.IsTypeAssignableTo(from, into) {
							return refused("a data value must preserve the checker-proven type of its target slot")
						}
					}
				}
				mask |= 1 << index
				written = append(written, fit(value, ir.Union))
			}
			if field := l.checker.GetPropertyOfType(targetType, key); field != nil {
				switch of, _ := l.representation(l.checker.GetTypeOfSymbol(field)); of {
				case ir.Number:
					representation = 1
				case ir.Boolean:
					representation = 2
				case ir.String:
					representation = 3
				default:
					return refused("existing target slots must be present number, boolean or string data fields")
				}
				if field.Flags&ast.SymbolFlagsOptional != 0 {
					return refused("optional field presence is not represented")
				}
			}
			written = append(written, ir.NumberConstant{Value: float64(mask)}, ir.NumberConstant{Value: float64(representation)})
			errorLocal := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "descriptor_error", Type: ir.String, Function: function})
			failure := ir.Read{Local: errorLocal, Of: ir.String}
			body = append(body, ir.Declare{Local: errorLocal, Value: ir.ObjectCall{Method: "definePropertyError", Arguments: written, Returns: ir.String}},
				ir.If{Condition: ir.IsUndefined{Value: failure}, Else: []ir.Statement{ir.Throw{Value: ir.ObjectCall{Method: "typeError", Arguments: []ir.Expression{failure}, Returns: ir.Object}}}})
		}
		body = append(body, ir.Return{Value: receiver})
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "object_" + name, Parameters: parameters, Returns: values[0].Type(), Body: body})
	return ir.Call{Function: function, Arguments: values, Returns: values[0].Type()}, true, nil
}

func objectDescriptorIndex(key string) (uint64, bool) {
	index, err := strconv.ParseUint(key, 10, 32)
	return index, err == nil && index < 4294967295 && strconv.FormatUint(index, 10) == key
}

func (l *lowering) objectDescriptorLiteral(node *ast.Node, depth int) *ast.Node {
	if depth > 16 || !l.exactObject(node, 0) {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression {
		return node
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
		return nil
	}
	return l.objectDescriptorLiteral(symbol.Declarations[0].AsVariableDeclaration().Initializer, depth+1)
}

// These consumers still use the static complete shape. Refuse their combination
// with descriptor mutation until they account for hidden scalar keys and flags.
func (l *lowering) objectDescriptorHazard() string {
	hazard := ""
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return "unknown runtime shapes"
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if hazard != "" {
			return true
		}
		if node.Kind == ast.KindSpreadAssignment {
			hazard = "object spread after descriptor mutation (presence is not represented)"
			return true
		}
		if node.Kind == ast.KindForInStatement {
			hazard = "for...in after descriptor mutation (dynamic keys are not represented)"
			return true
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				property := callee.AsPropertyAccessExpression()
				name := property.Name().Text()
				if (l.isLibraryGlobal(property.Expression, "Object") && (name == "values" || name == "entries" || name == "assign")) || (l.isLibraryGlobal(property.Expression, "JSON") && name == "stringify") {
					hazard = name + " after descriptor mutation (the consumer assumes the static complete shape)"
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return hazard
}

// A static descriptor/map read cannot observe hidden keys or non-enumerable map
// fields introduced by another descriptor call. Compare literal origins across
// const aliases rather than just the spelling of the receiver.
func (l *lowering) objectDescriptorWasMutated(literal *ast.Node) bool {
	mutated := false
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if mutated {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && len(call.Arguments.Nodes) > 0 {
				property := callee.AsPropertyAccessExpression()
				name := property.Name().Text()
				if l.isLibraryGlobal(property.Expression, "Object") && (name == "defineProperty" || name == "defineProperties") && l.objectDescriptorLiteral(call.Arguments.Nodes[0], 0) == literal {
					mutated = true
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return mutated
}
