package lower

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Keep the language-level NotYet classification while adding this decoder's concrete fix.
type jsonDecodeNullableNotYet struct{ NotYet }

func (n *jsonDecodeNullableNotYet) Error() string {
	return n.NotYet.Error() +
		"\nfix: nullable JSON fields come with the representation of T | null; decode the field as a discriminated union or leave it out"
}
func (n *jsonDecodeNullableNotYet) Unwrap() error { return &n.NotYet }

// Check the complete data graph before representation selection can reject a nullable union.
func (l *lowering) jsonDecodeContainsNull(t *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[t] {
		return false
	}
	seen[t] = true
	if t.Flags()&checker.TypeFlagsNull != 0 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range t.Types() {
			if l.jsonDecodeContainsNull(member, seen) {
				return true
			}
		}
	} else if t.Flags()&checker.TypeFlagsObject != 0 {
		if l.isArrayType(t) {
			return l.jsonDecodeContainsNull(l.elementTypeOfArray(t), seen)
		}
		if checker.IsTupleType(t) {
			for _, element := range l.checker.GetTypeArguments(t) {
				if l.jsonDecodeContainsNull(element, seen) {
					return true
				}
			}
		} else {
			for _, field := range l.checker.GetPropertiesOfType(t) {
				if l.jsonDecodeContainsNull(l.checker.GetTypeOfSymbol(field), seen) {
					return true
				}
			}
		}
	}
	return false
}

type jsonSchemaMode uint8

const (
	jsonSchemaDecode jsonSchemaMode = iota
	jsonSchemaEncode
	jsonSchemaBoundary
)

func (mode jsonSchemaMode) functionName() string {
	if mode == jsonSchemaEncode {
		return "encodeJson"
	}
	return "decodeJson"
}

// Only call-site dispatch inspects the prelude callee; the type walker uses mode.
func (l *lowering) isJSONSchemaCall(node *ast.Node) bool {
	return node.Kind == ast.KindCallExpression && (l.isPreludeFunction(
		node.AsCallExpression().Expression,
		"decodeJson",
	) || l.isPreludeFunction(node.AsCallExpression().Expression, "encodeJson"))
}

// jsonSchemaCallRefusal refuses a decodeJson or encodeJson call whose type has no schema.
func (l *lowering) jsonSchemaCallRefusal(node *ast.Node) error {
	if l.isJSONSchemaCall(node) {
		_, jsonSchemaCallTypeError := l.jsonSchemaCallType(node)
		if jsonSchemaCallTypeError != nil {
			return jsonSchemaCallTypeError
		}
	}
	return nil
}

func init() { refusalRules.register(1500, "jsonSchemaCallRefusal", (*lowering).jsonSchemaCallRefusal) }

func (l *lowering) jsonSchemaCallType(node *ast.Node) (ir.JSONDecodeSchema, error) {
	mode := jsonSchemaDecode
	if l.isPreludeFunction(node.AsCallExpression().Expression, "encodeJson") {
		mode = jsonSchemaEncode
	}
	return l.jsonSchemaType(node, mode)
}

func (l *lowering) decodeJsonType(node *ast.Node) (ir.JSONDecodeSchema, error) {
	return l.jsonSchemaType(node, jsonSchemaDecode)
}

func (l *lowering) jsonSchemaType(node *ast.Node, mode jsonSchemaMode) (ir.JSONDecodeSchema, error) {
	schema := ir.JSONDecodeSchema{}
	call := node.AsCallExpression()
	operation, argument := mode.functionName(), "text"
	if mode == jsonSchemaEncode {
		argument = "value"
	}
	if call.TypeArguments == nil || len(call.TypeArguments.Nodes) != 1 {
		return schema, &Refused{
			Where: l.program.Where(node),
			What:  operation + " requires a type argument",
			Fix:   "name the type: " + operation + "<YourType>(" + argument + ")",
		}
	}
	rootType := l.checker.GetTypeFromTypeNode(call.TypeArguments.Nodes[0])
	if l.jsonDecodeContainsNull(rootType, map[*checker.Type]bool{}) {
		return schema, &jsonDecodeNullableNotYet{NotYet{
			Where: l.program.Where(node),
			What:  operation + "<" + l.checker.TypeToString(rootType) + "> containing null",
		}}
	}
	return l.jsonDecodeSchema(node, rootType, mode, l.constant)
}

// typeString is l.checker.TypeToString(t), held per type for the lowering. TypeToString takes no enclosing node and no
// flags, so its answer is a function of the type alone, and a held answer is the answer. A schema node's expected text
// is asked once a node and a node once a function's return type, so tsc's entry census asked 1,866,154 times for 3,790
// types, and each answer's synthesized type nodes stayed live in the checker: 7.57 against 5.04 GiB at its peak
// (#wy2dyf0, #rw2vjhn). A lowering is one goroutine's, as sourceIdentities is.
func (l *lowering) typeString(t *checker.Type) string {
	if text, found := l.typeStrings[t]; found {
		return text
	}
	if l.typeStrings == nil {
		l.typeStrings = map[*checker.Type]string{}
	}
	text := l.checker.TypeToString(t)
	l.typeStrings[t] = text
	return text
}

// jsonDecodeSchema is the one checker-type descriptor builder. Boundary mode
// preserves undefined for fields; decode and encode retain JSON admission rules.
func (l *lowering) jsonDecodeSchema(
	node *ast.Node,
	rootType *checker.Type,
	mode jsonSchemaMode,
	intern func(string) int,
) (ir.JSONDecodeSchema, error) {
	schema := ir.JSONDecodeSchema{}
	operation := mode.functionName()
	seen := map[*checker.Type]int{}
	var visit func(*checker.Type) (int, error)
	refuse := func(t *checker.Type) error {
		return &Refused{
			Where: l.program.Where(node),
			What:  operation + " cannot prove " + l.checker.TypeToString(t) + " is JSON data",
			Fix:   "name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant",
		}
	}
	visit = func(t *checker.Type) (int, error) {
		if mode == jsonSchemaBoundary {
			t = l.concrete(t)
			held, _ := l.representation(t)
			if l.weakTarget(t) != nil || held == ir.Weak {
				return 0, refuse(t)
			}
		}
		// An open generic is refused even if it has a data constraint.
		if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
			return 0, refuse(t)
		}
		if index, ok := seen[t]; ok {
			return index, nil
		}
		index := len(schema.Nodes)
		seen[t] = index
		schema.Nodes = append(schema.Nodes, ir.JSONDecodeNode{})
		n := ir.JSONDecodeNode{Expected: l.typeString(t)}
		flags := t.Flags()
		var literal ir.Expression
		var of ir.Type
		var ok bool
		if flags&checker.TypeFlagsStringLiteral != 0 {
			text, isText := t.AsLiteralType().Value().(string)
			literal, of, ok = ir.StringConstant{Index: intern(text)}, ir.String, isText
		} else {
			literal, of, ok = l.literalConstant(t)
		}
		if ok {
			n.Kind = "literal"
			n.Of = of
			switch v := literal.(type) {
			case ir.StringConstant:
				n.Literal = t.AsLiteralType().Value().(string)
				n.LiteralUnits = ir.JSONLiteralUnits(n.Literal)
			case ir.NumberConstant:
				n.Number = v.Value
				n.NumberText = strconv.FormatFloat(v.Value, 'g', -1, 64)
				if n.NumberText == "+Inf" {
					n.NumberText = "Infinity"
				}
				if n.NumberText == "-Inf" {
					n.NumberText = "-Infinity"
				}
			case ir.BooleanConstant:
				n.Boolean = v.Value
			}
		} else if mode == jsonSchemaBoundary && flags&checker.TypeFlagsUndefined != 0 {
			n.Kind = "undefined"
		} else if flags&checker.TypeFlagsNumber != 0 {
			n.Kind = "number"
			n.Of = ir.Number
		} else if flags&checker.TypeFlagsString != 0 {
			n.Kind = "string"
			n.Of = ir.String
		} else if flags&checker.TypeFlagsBoolean != 0 {
			n.Kind = "boolean"
			n.Of = ir.Boolean
		} else if flags&checker.TypeFlagsUnion != 0 {
			n.Kind = "union"
			of, ok := l.representation(t)
			if !ok && mode != jsonSchemaBoundary {
				return 0, refuse(t)
			}
			n.Of = of
			objects := []*checker.Type{}
			for _, member := range t.Types() {
				child, visitError := visit(member)
				if visitError != nil {
					return 0, visitError
				}
				n.Children = append(n.Children, child)
				k := schema.Nodes[child].Kind
				if k == "object" {
					objects = append(objects, member)
				} else if mode != jsonSchemaBoundary &&
					k != "null" &&
					k != "literal" &&
					k != "number" &&
					k != "string" &&
					k != "boolean" {
					return 0, refuse(t)
				}
			}
			if mode != jsonSchemaBoundary && len(objects) > 0 {
				if len(objects)+1 < len(n.Children) || len(objects) == 1 && len(n.Children) != 2 {
					return 0, refuse(t)
				}
				if len(objects) > 1 {
					for _, field := range l.checker.GetPropertiesOfType(objects[0]) {
						distinct := map[string]bool{}
						valid := true
						for _, member := range objects {
							property := l.checker.GetPropertyOfType(member, field.Name)
							literal := l.fieldLiteral(member, field.Name)
							if property == nil || property.Flags&ast.SymbolFlagsOptional != 0 || literal == nil {
								valid = false
								break
							}
							key := l.checker.TypeToString(literal)
							if distinct[key] {
								valid = false
								break
							}
							distinct[key] = true
						}
						if valid {
							n.Discriminant = field.Name
							n.DiscriminantUnits = ir.JSONLiteralUnits(field.Name)
							break
						}
					}
					if n.Discriminant == "" {
						return 0, refuse(t)
					}
				}
			}
		} else if flags&checker.TypeFlagsObject != 0 {
			if mode == jsonSchemaEncode && isClassInstance(t) {
				return 0, l.notYet(
					node,
					"encodeJson of a class instance is not yet supported; describe the data with an interface",
				)
			}
			if isClassInstance(t) ||
				l.isLibraryType(t, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Date", "RegExp") ||
				len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 ||
				len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 ||
				len(l.checker.GetIndexInfosOfType(t)) != 0 && !l.isArrayType(t) && !checker.IsTupleType(t) {
				return 0, refuse(t)
			}
			if checker.IsTupleType(t) {
				n.Kind = "tuple"
				n.Of = ir.Object
				for i, element := range l.checker.GetTypeArguments(t) {
					child, visitError := visit(element)
					if visitError != nil {
						return 0, visitError
					}
					n.Fields = append(n.Fields, ir.JSONDecodeField{Name: strconv.Itoa(i), Node: child})
				}
			} else if l.isArrayType(t) {
				n.Kind = "array"
				n.Of = ir.Array
				child, visitError := visit(l.elementTypeOfArray(t))
				if visitError != nil {
					return 0, visitError
				}
				n.Children = []int{child}
			} else {
				n.Kind = "object"
				n.Of = ir.Object
				for _, field := range l.checker.GetPropertiesOfType(t) {
					if strings.ContainsRune(field.Name, 0) {
						return 0, refuse(t)
					}
					ft := l.checker.GetTypeOfSymbol(field)
					optional := field.Flags&ast.SymbolFlagsOptional != 0
					if mode != jsonSchemaBoundary && optional && ft.Flags()&checker.TypeFlagsUnion != 0 {
						members := []*checker.Type{}
						for _, member := range ft.Types() {
							if member.Flags()&checker.TypeFlagsUndefined != 0 {
								if member == checker.Checker_undefinedType(l.checker) {
									return 0, refuse(ft)
								}
								continue // The checker's missing-property type, not JSON undefined.
							}
							members = append(members, member)
						}
						ft = l.checker.GetUnionType(members)
					}
					child, visitError := visit(ft)
					if visitError != nil {
						return 0, visitError
					}
					n.Fields = append(n.Fields, ir.JSONDecodeField{
						Name:      field.Name,
						NameUnits: ir.JSONLiteralUnits(field.Name),
						Node:      child,
						Optional:  optional,
					})
				}
			}
		} else {
			return 0, refuse(t)
		}
		schema.Nodes[index] = n
		return index, nil
	}
	root, visitError := visit(rootType)
	schema.Root = root
	return schema, visitError
}
func (l *lowering) decodeJson(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isPreludeFunction(node.AsCallExpression().Expression, "decodeJson") {
		return nil, false, nil
	}
	schema, decodeJsonTypeError := l.decodeJsonType(node)
	if decodeJsonTypeError != nil {
		return nil, true, decodeJsonTypeError
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "decodeJson requires one text argument")
	}
	text, expressionError := l.expression(args[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	return ir.JSONDecode{Text: text, Schema: schema}, true, nil
}

// DecodeJsonSources supplies only types to the source oracle. It does not lower expressions:
// Node still executes the original source, with a descriptor inserted as a hidden argument.
func DecodeJsonSources(ctx context.Context, program *load.Program) (map[string]string, error) {
	entry := program.Files()[0]
	typeChecker, release := program.Checker(ctx, entry)
	defer release()
	l := &lowering{program: program, checker: typeChecker, result: &ir.Program{}}
	modules, moduleOrderError := l.moduleOrder(entry)
	if moduleOrderError != nil {
		return nil, moduleOrderError
	}
	result := map[string]string{}
	for _, module := range modules {
		type insertion struct {
			position int
			text     string
		}
		insertions := []insertion{}
		var found error
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if found != nil {
				return true
			}
			if l.isJSONSchemaCall(node) {
				descriptor, jsonSchemaCallTypeError := l.jsonSchemaCallType(node)
				if jsonSchemaCallTypeError != nil {
					found = jsonSchemaCallTypeError
					return true
				}
				encoded, encodeError := json.Marshal(descriptor)
				if encodeError != nil {
					found = encodeError
					return true
				}
				arguments := node.AsCallExpression().Arguments.Nodes
				if len(arguments) != 1 {
					found = l.notYet(node, "decodeJson requires one text argument")
					return true
				}
				// Insert before any existing trailing comma or comment.
				insertions = append(insertions, insertion{arguments[0].End(), ", " + string(encoded)})
			}
			node.ForEachChild(visit)
			return false
		}
		module.AsNode().ForEachChild(visit)
		if found != nil {
			return nil, found
		}
		sort.Slice(insertions, func(i, j int) bool { return insertions[i].position > insertions[j].position })
		source := module.Text()
		for _, insert := range insertions {
			source = source[:insert.position] + insert.text + source[insert.position:]
		}
		result[program.FileName(module)] = source
	}
	return result, nil
}

// A decoder result's readonly value field is always an own plain data value, with no hidden
// prototype fields. Keys can inspect its actual shape without trusting a structural view.
func (l *lowering) decodedResultValue(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "value" {
		return false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := l.symbol(receiver)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return false
	}
	initializer = ast.SkipParentheses(initializer)
	return initializer.Kind == ast.KindCallExpression &&
		l.isPreludeFunction(initializer.AsCallExpression().Expression, "decodeJson")
}
