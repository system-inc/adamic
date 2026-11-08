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

func (notYet *jsonDecodeNullableNotYet) Error() string {
	return notYet.NotYet.Error() +
		"\nfix: nullable JSON fields come with the representation of T | null; decode the field as a discriminated union or leave it out"
}
func (notYet *jsonDecodeNullableNotYet) Unwrap() error { return &notYet.NotYet }

// Check the complete data graph before representation selection can reject a nullable union.
func (lowering *lowering) jsonDecodeContainsNull(subjectType *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[subjectType] {
		return false
	}
	seen[subjectType] = true
	if subjectType.Flags()&checker.TypeFlagsNull != 0 {
		return true
	}
	if subjectType.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range subjectType.Types() {
			if lowering.jsonDecodeContainsNull(member, seen) {
				return true
			}
		}
	} else if subjectType.Flags()&checker.TypeFlagsObject != 0 {
		if lowering.isArrayType(subjectType) {
			return lowering.jsonDecodeContainsNull(lowering.elementTypeOfArray(subjectType), seen)
		}
		if checker.IsTupleType(subjectType) {
			for _, element := range lowering.checker.GetTypeArguments(subjectType) {
				if lowering.jsonDecodeContainsNull(element, seen) {
					return true
				}
			}
		} else {
			for _, field := range lowering.checker.GetPropertiesOfType(subjectType) {
				if lowering.jsonDecodeContainsNull(lowering.checker.GetTypeOfSymbol(field), seen) {
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
func (lowering *lowering) isJsonSchemaCall(node *ast.Node) bool {
	return node.Kind == ast.KindCallExpression && (lowering.isPreludeFunction(
		node.AsCallExpression().Expression,
		"decodeJson",
	) || lowering.isPreludeFunction(node.AsCallExpression().Expression, "encodeJson"))
}

// jsonSchemaCallRefusal refuses a decodeJson or encodeJson call whose type has no schema.
func (lowering *lowering) jsonSchemaCallRefusal(node *ast.Node) error {
	if lowering.isJsonSchemaCall(node) {
		_, jsonSchemaCallTypeError := lowering.jsonSchemaCallType(node)
		if jsonSchemaCallTypeError != nil {
			return jsonSchemaCallTypeError
		}
	}
	return nil
}

func init() { refusalRules.register(1500, "jsonSchemaCallRefusal", (*lowering).jsonSchemaCallRefusal) }

func (lowering *lowering) jsonSchemaCallType(node *ast.Node) (ir.JSONDecodeSchema, error) {
	mode := jsonSchemaDecode
	if lowering.isPreludeFunction(node.AsCallExpression().Expression, "encodeJson") {
		mode = jsonSchemaEncode
	}
	return lowering.jsonSchemaType(node, mode)
}

func (lowering *lowering) decodeJsonType(node *ast.Node) (ir.JSONDecodeSchema, error) {
	return lowering.jsonSchemaType(node, jsonSchemaDecode)
}

func (lowering *lowering) jsonSchemaType(node *ast.Node, mode jsonSchemaMode) (ir.JSONDecodeSchema, error) {
	schema := ir.JSONDecodeSchema{}
	call := node.AsCallExpression()
	operation, argument := mode.functionName(), "text"
	if mode == jsonSchemaEncode {
		argument = "value"
	}
	if call.TypeArguments == nil || len(call.TypeArguments.Nodes) != 1 {
		return schema, &Refused{
			Where: lowering.program.Where(node),
			What:  operation + " requires a type argument",
			Fix:   "name the type: " + operation + "<YourType>(" + argument + ")",
		}
	}
	rootType := lowering.checker.GetTypeFromTypeNode(call.TypeArguments.Nodes[0])
	if lowering.jsonDecodeContainsNull(rootType, map[*checker.Type]bool{}) {
		return schema, &jsonDecodeNullableNotYet{NotYet{
			Where: lowering.program.Where(node),
			What:  operation + "<" + lowering.checker.TypeToString(rootType) + "> containing null",
		}}
	}
	return lowering.jsonDecodeSchema(node, rootType, mode, lowering.constant)
}

// typeString is l.checker.TypeToString(subjectType), held per type for the lowering. TypeToString takes no enclosing
// node and no flags, so its answer is a function of the type alone, and a held answer is the answer. A schema node's
// expected text is asked once a node and a node once a function's return type, so tsc's entry census asked 1,866,154
// times for 3,790 types, and each answer's synthesized type nodes stayed live in the checker: 7.57 against 5.04 GiB at
// its peak (#wy2dyf0, #rw2vjhn). A lowering is one goroutine's, as sourceIdentities is.
func (lowering *lowering) typeString(subjectType *checker.Type) string {
	if text, found := lowering.typeStrings[subjectType]; found {
		return text
	}
	if lowering.typeStrings == nil {
		lowering.typeStrings = map[*checker.Type]string{}
	}
	text := lowering.checker.TypeToString(subjectType)
	lowering.typeStrings[subjectType] = text
	return text
}

// jsonDecodeSchema is the one checker-type descriptor builder. Boundary mode
// preserves undefined for fields; decode and encode retain JSON admission rules.
func (lowering *lowering) jsonDecodeSchema(
	node *ast.Node,
	rootType *checker.Type,
	mode jsonSchemaMode,
	intern func(string) int,
) (ir.JSONDecodeSchema, error) {
	schema := ir.JSONDecodeSchema{}
	operation := mode.functionName()
	seen := map[*checker.Type]int{}
	var visit func(*checker.Type) (int, error)
	refuse := func(subjectType *checker.Type) error {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  operation + " cannot prove " + lowering.checker.TypeToString(subjectType) + " is JSON data",
			Fix:   "name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant",
		}
	}
	visit = func(subjectType *checker.Type) (int, error) {
		if mode == jsonSchemaBoundary {
			subjectType = lowering.concrete(subjectType)
			held, _ := lowering.representation(subjectType)
			if lowering.weakTarget(subjectType) != nil || held == ir.Weak {
				return 0, refuse(subjectType)
			}
		}
		// An open generic is refused even if it has a data constraint.
		if subjectType.Flags()&checker.TypeFlagsTypeParameter != 0 {
			return 0, refuse(subjectType)
		}
		if index, ok := seen[subjectType]; ok {
			return index, nil
		}
		index := len(schema.Nodes)
		seen[subjectType] = index
		schema.Nodes = append(schema.Nodes, ir.JSONDecodeNode{})
		decodeNode := ir.JSONDecodeNode{Expected: lowering.typeString(subjectType)}
		flags := subjectType.Flags()
		var literal ir.Expression
		var of ir.Type
		var ok bool
		if flags&checker.TypeFlagsStringLiteral != 0 {
			text, isText := subjectType.AsLiteralType().Value().(string)
			literal, of, ok = ir.StringConstant{Index: intern(text)}, ir.String, isText
		} else {
			literal, of, ok = lowering.literalConstant(subjectType)
		}
		if ok {
			decodeNode.Kind = "literal"
			decodeNode.Of = of
			switch typed := literal.(type) {
			case ir.StringConstant:
				decodeNode.Literal = subjectType.AsLiteralType().Value().(string)
				decodeNode.LiteralUnits = ir.JsonLiteralUnits(decodeNode.Literal)
			case ir.NumberConstant:
				decodeNode.Number = typed.Value
				decodeNode.NumberText = strconv.FormatFloat(typed.Value, 'g', -1, 64)
				if decodeNode.NumberText == "+Inf" {
					decodeNode.NumberText = "Infinity"
				}
				if decodeNode.NumberText == "-Inf" {
					decodeNode.NumberText = "-Infinity"
				}
			case ir.BooleanConstant:
				decodeNode.Boolean = typed.Value
			}
		} else if mode == jsonSchemaBoundary && flags&checker.TypeFlagsUndefined != 0 {
			decodeNode.Kind = "undefined"
		} else if flags&checker.TypeFlagsNumber != 0 {
			decodeNode.Kind = "number"
			decodeNode.Of = ir.Number
		} else if flags&checker.TypeFlagsString != 0 {
			decodeNode.Kind = "string"
			decodeNode.Of = ir.String
		} else if flags&checker.TypeFlagsBoolean != 0 {
			decodeNode.Kind = "boolean"
			decodeNode.Of = ir.Boolean
		} else if flags&checker.TypeFlagsUnion != 0 {
			decodeNode.Kind = "union"
			of, ok := lowering.representation(subjectType)
			if !ok && mode != jsonSchemaBoundary {
				return 0, refuse(subjectType)
			}
			decodeNode.Of = of
			objects := []*checker.Type{}
			for _, member := range subjectType.Types() {
				child, visitError := visit(member)
				if visitError != nil {
					return 0, visitError
				}
				decodeNode.Children = append(decodeNode.Children, child)
				k := schema.Nodes[child].Kind
				if k == "object" {
					objects = append(objects, member)
				} else if mode != jsonSchemaBoundary &&
					k != "null" &&
					k != "literal" &&
					k != "number" &&
					k != "string" &&
					k != "boolean" {
					return 0, refuse(subjectType)
				}
			}
			if mode != jsonSchemaBoundary && len(objects) > 0 {
				if len(objects)+1 < len(decodeNode.Children) || len(objects) == 1 && len(decodeNode.Children) != 2 {
					return 0, refuse(subjectType)
				}
				if len(objects) > 1 {
					for _, field := range lowering.checker.GetPropertiesOfType(objects[0]) {
						distinct := map[string]bool{}
						valid := true
						for _, member := range objects {
							property := lowering.checker.GetPropertyOfType(member, field.Name)
							literal := lowering.fieldLiteral(member, field.Name)
							if property == nil || property.Flags&ast.SymbolFlagsOptional != 0 || literal == nil {
								valid = false
								break
							}
							key := lowering.checker.TypeToString(literal)
							if distinct[key] {
								valid = false
								break
							}
							distinct[key] = true
						}
						if valid {
							decodeNode.Discriminant = field.Name
							decodeNode.DiscriminantUnits = ir.JsonLiteralUnits(field.Name)
							break
						}
					}
					if decodeNode.Discriminant == "" {
						return 0, refuse(subjectType)
					}
				}
			}
		} else if flags&checker.TypeFlagsObject != 0 {
			if mode == jsonSchemaEncode && isClassInstance(subjectType) {
				return 0, lowering.notYet(
					node,
					"encodeJson of a class instance is not yet supported; describe the data with an interface",
				)
			}
			if isClassInstance(subjectType) ||
				lowering.isLibraryType(subjectType, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Date", "RegExp") ||
				len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall)) != 0 ||
				len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindConstruct)) != 0 ||
				len(lowering.checker.GetIndexInfosOfType(subjectType)) != 0 &&
					!lowering.isArrayType(subjectType) &&
					!checker.IsTupleType(subjectType) {
				return 0, refuse(subjectType)
			}
			if checker.IsTupleType(subjectType) {
				decodeNode.Kind = "tuple"
				decodeNode.Of = ir.Object
				for index, element := range lowering.checker.GetTypeArguments(subjectType) {
					child, visitError := visit(element)
					if visitError != nil {
						return 0, visitError
					}
					decodeNode.Fields = append(
						decodeNode.Fields,
						ir.JSONDecodeField{Name: strconv.Itoa(index), Node: child},
					)
				}
			} else if lowering.isArrayType(subjectType) {
				decodeNode.Kind = "array"
				decodeNode.Of = ir.Array
				child, visitError := visit(lowering.elementTypeOfArray(subjectType))
				if visitError != nil {
					return 0, visitError
				}
				decodeNode.Children = []int{child}
			} else {
				decodeNode.Kind = "object"
				decodeNode.Of = ir.Object
				for _, field := range lowering.checker.GetPropertiesOfType(subjectType) {
					if strings.ContainsRune(field.Name, 0) {
						return 0, refuse(subjectType)
					}
					ft := lowering.checker.GetTypeOfSymbol(field)
					optional := field.Flags&ast.SymbolFlagsOptional != 0
					if mode != jsonSchemaBoundary && optional && ft.Flags()&checker.TypeFlagsUnion != 0 {
						members := []*checker.Type{}
						for _, member := range ft.Types() {
							if member.Flags()&checker.TypeFlagsUndefined != 0 {
								if member == checker.Checker_undefinedType(lowering.checker) {
									return 0, refuse(ft)
								}
								continue // The checker's missing-property type, not JSON undefined.
							}
							members = append(members, member)
						}
						ft = lowering.checker.GetUnionType(members)
					}
					child, visitError := visit(ft)
					if visitError != nil {
						return 0, visitError
					}
					decodeNode.Fields = append(decodeNode.Fields, ir.JSONDecodeField{
						Name:      field.Name,
						NameUnits: ir.JsonLiteralUnits(field.Name),
						Node:      child,
						Optional:  optional,
					})
				}
			}
		} else {
			return 0, refuse(subjectType)
		}
		schema.Nodes[index] = decodeNode
		return index, nil
	}
	root, visitError := visit(rootType)
	schema.Root = root
	return schema, visitError
}
func (lowering *lowering) decodeJson(node *ast.Node) (ir.Expression, bool, error) {
	if !lowering.isPreludeFunction(node.AsCallExpression().Expression, "decodeJson") {
		return nil, false, nil
	}
	schema, decodeJsonTypeError := lowering.decodeJsonType(node)
	if decodeJsonTypeError != nil {
		return nil, true, decodeJsonTypeError
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || hasSpread(node) {
		return nil, true, lowering.notYet(node, "decodeJson requires one text argument")
	}
	text, expressionError := lowering.expression(args[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	return ir.JSONDecode{Text: text, Schema: schema}, true, nil
}

// DecodeJsonSources supplies only types to the source oracle. It does not lower expressions:
// Node still executes the original source, with a descriptor inserted as a hidden argument.
func DecodeJsonSources(callerContext context.Context, program *load.Program) (map[string]string, error) {
	entry := program.Files()[0]
	typeChecker, release := program.Checker(callerContext, entry)
	defer release()
	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}}
	modules, moduleOrderError := lowering.moduleOrder(entry)
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
			if lowering.isJsonSchemaCall(node) {
				descriptor, jsonSchemaCallTypeError := lowering.jsonSchemaCallType(node)
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
					found = lowering.notYet(node, "decodeJson requires one text argument")
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
		sort.Slice(
			insertions,
			func(left, right int) bool { return insertions[left].position > insertions[right].position },
		)
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
func (lowering *lowering) decodedResultValue(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "value" {
		return false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := lowering.symbol(receiver)
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
		lowering.isPreludeFunction(initializer.AsCallExpression().Expression, "decodeJson")
}
