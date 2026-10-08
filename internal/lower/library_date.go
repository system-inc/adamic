package lower

import (
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) dateClock(node *ast.Node, what string) error {
	return &Refused{Where: l.program.Where(node), What: what, Fix: "the wall clock is nondeterministic, so the oracle cannot compare it with Node; pass an explicit timestamp"}
}
func (l *lowering) dateRefusal(node *ast.Node) error {
	if node.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Date") && node.Name().Text() == "now" {
		return l.dateClock(node, "Date.now")
	}
	if node.Kind == ast.KindNewExpression && l.isLibraryGlobal(node.AsNewExpression().Expression, "Date") && (node.AsNewExpression().Arguments == nil || len(node.AsNewExpression().Arguments.Nodes) == 0) {
		return l.dateClock(node, "new Date() without arguments")
	}
	if node.Kind == ast.KindCallExpression && l.isLibraryGlobal(node.AsCallExpression().Expression, "Date") {
		return l.dateClock(node, "Date()")
	}
	return nil
}

// Dynamic strings can select local legacy forms, so all parsing uses the explicit UTC contract.
func (l *lowering) dateString(node *ast.Node) (ir.Expression, error) {
	if os.Getenv("TZ") != "UTC" {
		return nil, l.notYet(node, "Date parsing without TZ=UTC fixed for native and Node")
	}
	return l.stringConversion(node)
}
func (l *lowering) dateArguments(node *ast.Node, written []*ast.Node) ([]ir.Expression, error) {
	var args []ir.Expression
	for _, arg := range written {
		if arg.Kind == ast.KindSpreadElement {
			return nil, l.notYet(node, "Date arguments with spread")
		}
		value, err := l.libraryNumber(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}
	return args, nil
}
func (l *lowering) newDate(node *ast.Node) (ir.Expression, error) {
	written := node.AsNewExpression().Arguments.Nodes
	if len(written) == 1 {
		if l.isLibraryType(l.checker.GetTypeAtLocation(written[0]), "Date") {
			value, err := l.expression(written[0])
			return ir.DateCall{Method: "copy", Receiver: value, Returns: ir.Object}, err
		}
		of, _ := l.representation(l.checker.GetTypeAtLocation(written[0]))
		if of == ir.String {
			value, err := l.dateString(written[0])
			return ir.DateCall{Method: "newString", Arguments: []ir.Expression{value}, Returns: ir.Object}, err
		}
		value, err := l.libraryNumber(written[0])
		return ir.DateCall{Method: "new", Arguments: []ir.Expression{value}, Returns: ir.Object}, err
	}
	if os.Getenv("TZ") != "UTC" {
		return nil, l.notYet(node, "local Date construction without TZ=UTC fixed for native and Node")
	}
	args, err := l.dateArguments(node, written)
	if err != nil {
		return nil, err
	}
	return ir.DateCall{Method: "components", Arguments: args, Returns: ir.Object}, nil
}

var dateGetters = map[string]int{"getFullYear": 0, "getMonth": 1, "getDate": 2, "getDay": 3, "getHours": 4, "getMinutes": 5, "getSeconds": 6, "getMilliseconds": 7, "getTimezoneOffset": 8, "getYear": 9}
var dateSetters = map[string]int{"setFullYear": 0, "setMonth": 1, "setDate": 2, "setHours": 4, "setMinutes": 5, "setSeconds": 6, "setMilliseconds": 7, "setYear": 9, "setTime": 10}

func dateLocalName(name string) string { return strings.Replace(name, "UTC", "", 1) }
func (l *lowering) libraryDateCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	written := node.AsCallExpression().Arguments.Nodes
	if name == "hasOwnProperty" && (l.datePrototype(receiver) || l.isLibraryGlobal(receiver, "Date")) {
		if len(written) != 1 || hasSpread(node) {
			return nil, true, l.notYet(node, "Date hasOwnProperty with missing, extra or spread arguments")
		}
		key, err := l.stringConversion(written[0])
		method := "constructorHasOwn"
		if l.datePrototype(receiver) {
			method = "prototypeHasOwn"
		}
		return ir.DateCall{Method: method, Arguments: []ir.Expression{key}, Returns: ir.Boolean}, true, err
	}
	if value, known, err := l.dateStringMethod(node); known {
		return value, true, err
	}
	// Explicit .call binds the intrinsic's this without materializing Date.prototype.
	if name == "call" {
		method := ast.SkipParentheses(receiver)
		if method.Kind == ast.KindPropertyAccessExpression && l.datePrototype(method.AsPropertyAccessExpression().Expression) {
			if len(written) == 0 || !l.isLibraryType(l.checker.GetTypeAtLocation(written[0]), "Date") {
				if method.Name().Text() == "toJSON" {
					return nil, true, l.notYet(node, "Date.toJSON on a generic receiver (ToPrimitive and dynamic toISOString lookup are not built)")
				}
				return nil, true, l.notYet(node, "Date prototype method requires a Date internal slot")
			}
			name = method.Name().Text()
			receiver = written[0]
			written = written[1:]
		}
	}
	if l.isLibraryGlobal(receiver, "Date") {
		if name == "UTC" {
			args, err := l.dateArguments(node, written)
			return ir.DateCall{Method: "UTC", Arguments: args, Returns: ir.Number}, true, err
		}
		if name == "parse" && len(written) == 1 {
			value, err := l.dateString(written[0])
			return ir.DateCall{Method: "parse", Arguments: []ir.Expression{value}, Returns: ir.Number}, true, err
		}
		return nil, true, l.notYet(node, "Date."+name)
	}
	if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date") || !l.libraryMember(callee) {
		return nil, false, nil
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	local := dateLocalName(name)
	if _, ok := dateGetters[local]; ok {
		if !strings.Contains(name, "UTC") && os.Getenv("TZ") != "UTC" {
			return nil, true, l.notYet(node, "local Date method without TZ=UTC fixed for native and Node")
		}
		args, err := l.dateArguments(node, written)
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.Number}, true, err
	}
	if _, ok := dateSetters[local]; ok {
		if !strings.Contains(name, "UTC") && name != "setTime" && os.Getenv("TZ") != "UTC" {
			return nil, true, l.notYet(node, "local Date method without TZ=UTC fixed for native and Node")
		}
		args, err := l.dateArguments(node, written)
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.Number}, true, err
	}
	if name == "getTime" || name == "valueOf" {
		args, err := l.dateArguments(node, written)
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.Number}, true, err
	}
	if name == "toISOString" && len(written) == 0 {
		return ir.DateCall{Receiver: value, Method: name, Returns: ir.String}, true, nil
	}
	if name == "toString" || name == "toUTCString" || name == "toDateString" || name == "toTimeString" {
		if name != "toUTCString" && os.Getenv("TZ") != "UTC" {
			return nil, true, l.notYet(node, "local Date formatting without TZ=UTC fixed for native and Node")
		}
		args, err := l.dateArguments(node, written)
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.String}, true, err
	}
	if name == "toJSON" {
		args := []ir.Expression{}
		for _, argument := range written {
			if argument.Kind == ast.KindSpreadElement {
				return nil, true, l.notYet(node, "Date.toJSON spread arguments")
			}
			evaluated, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			args = append(args, evaluated)
		}
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.String}, true, nil
	}
	return nil, true, l.notYet(node, "Date."+name)
}

// Date is nominal at every structural boundary because its scalar internal slot is not an own field.
func (l *lowering) dateViewsMatch(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	// Date internal slots cannot be supplied by structural objects or lost through a method interface.
	if l.isLibraryType(from, "Date") != l.isLibraryType(to, "Date") {
		return false
	}
	same := func(inside, viewed *checker.Type) bool {
		return l.dateViewsMatch(inside, viewed, visited)
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if !same(l.checker.GetTypeOfSymbol(fromParameters[index]), l.checker.GetTypeOfSymbol(toParameters[index])) {
				return false
			}
		}
		return same(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]))
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := l.typeArguments(from), l.typeArguments(to)
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if !same(fromArguments[index], toArguments[index]) {
				return false
			}
		}
	default:
		for _, viewed := range l.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil && !same(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)) {
				return false
			}
		}
	}
	return true
}

func (l *lowering) datePrototype(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Date")
}
func (l *lowering) libraryDateBoundMethod(node *ast.Node) bool {
	if !l.datePrototype(node.AsPropertyAccessExpression().Expression) {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && called(parent)
}

// Null strings use the existing reference null pointer, but never the undefined union tag.
func (l *lowering) dateNullableString(proven *checker.Type) bool {
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsStringLike) == 0 {
			return false
		}
	}
	return true
}

// Nullable strings and present strings need distinct generic bodies even though both are pointers.
func (l *lowering) dateTypeKey(proven *checker.Type, held ir.Type) string {
	if held == ir.String && l.includesNull(proven) {
		return "nullable-string"
	}
	return typeName(held)
}
