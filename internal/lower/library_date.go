package lower

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

const dateRuling = "ruling 10, roadmap step 29: Date admits only UTC and ISO forms, never locale- or timezone-dependent forms"

var dateISOGrammar = regexp.MustCompile(`^(?:[0-9]{4}|[+-][0-9]{6})(?:-[0-9]{2}(?:-[0-9]{2})?)?(?:[Tt][0-9]{2}:[0-9]{2}(?::[0-9]{2}(?:\.[0-9]+)?)?(?:[Zz]|[+-][0-9]{2}:?[0-9]{2}))?$`)
var dateUTCGetters = map[string]bool{"getUTCFullYear": true, "getUTCMonth": true, "getUTCDate": true, "getUTCDay": true, "getUTCHours": true, "getUTCMinutes": true, "getUTCSeconds": true, "getUTCMilliseconds": true}

func (l *lowering) dateForbidden(node *ast.Node, name string) error {
	return &Refused{Where: l.program.Where(node), What: name, Fix: dateRuling + "; use epoch milliseconds or an explicit UTC/ISO operation"}
}
func (l *lowering) datePrototype(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Date")
}
func (l *lowering) dateMethodRead(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.AsPropertyAccessExpression()
	if l.isLibraryGlobal(access.Expression, "Date") && node.Name().Text() == "now" {
		return true
	}
	if !l.datePrototype(access.Expression) {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && called(parent)
}
func (l *lowering) dateRefusal(node *ast.Node) error {
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Date") && callee.Name().Text() == "parse" && len(call.Arguments.Nodes) == 1 && !l.dateISOArgument(call.Arguments.Nodes[0], 0) {
			return l.dateForbidden(node, "Date.parse without a proven ISO grammar and explicit date-time zone")
		}
	}

	if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		target := ast.SkipParentheses(node.AsBinaryExpression().Left)
		var receiver *ast.Node
		if target.Kind == ast.KindPropertyAccessExpression {
			receiver = target.AsPropertyAccessExpression().Expression
		}
		if target.Kind == ast.KindElementAccessExpression {
			receiver = target.AsElementAccessExpression().Expression
		}
		if receiver != nil && (l.isLibraryGlobal(receiver, "Date") || l.datePrototype(receiver) || l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date")) {
			return l.notYet(node, "Date method/prototype overrides and expandos (intrinsic operations require unmodified internal-slot receivers)")
		}
	}

	if node.Kind == ast.KindNewExpression && l.isLibraryGlobal(node.AsNewExpression().Expression, "Date") {
		arguments := node.AsNewExpression().Arguments
		if arguments != nil && len(arguments.Nodes) > 1 {
			return l.dateForbidden(node, "new Date with local-time components")
		}
		if arguments != nil && len(arguments.Nodes) == 1 {
			if held, _ := l.representation(l.checker.GetTypeAtLocation(arguments.Nodes[0])); held == ir.String && !l.dateISOArgument(arguments.Nodes[0], 0) {
				return l.dateForbidden(node, "new Date with a non-ISO or unproven string")
			}
		}
	}
	if node.Kind == ast.KindTemplateSpan && l.isLibraryType(l.checker.GetTypeAtLocation(node.AsTemplateSpan().Expression), "Date") {
		return l.dateForbidden(node, "Date string interpolation (local toString)")
	}
	if node.Kind == ast.KindCallExpression && l.isLibraryGlobal(node.AsCallExpression().Expression, "String") {
		args := node.AsCallExpression().Arguments.Nodes
		if len(args) == 1 && l.isLibraryType(l.checker.GetTypeAtLocation(args[0]), "Date") {
			return l.dateForbidden(node, "String(Date) (local toString)")
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		var receiver *ast.Node
		name := ""
		if node.Kind == ast.KindPropertyAccessExpression {
			receiver = node.AsPropertyAccessExpression().Expression
			name = node.Name().Text()
		} else {
			access := node.AsElementAccessExpression()
			receiver = access.Expression
			if access.ArgumentExpression.Kind == ast.KindStringLiteral {
				name = access.ArgumentExpression.Text()
			}
		}
		if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date") || l.datePrototype(receiver) {
			if name == "toString" || name == "toDateString" || name == "toTimeString" || strings.HasPrefix(name, "toLocale") || name == "getTimezoneOffset" || (strings.HasPrefix(name, "get") && name != "getTime" && !strings.HasPrefix(name, "getUTC")) || (strings.HasPrefix(name, "set") && name != "setTime" && !strings.HasPrefix(name, "setUTC")) {
				return l.dateForbidden(node, "Date."+name)
			}
		}
	}
	if node.Kind == ast.KindCallExpression && l.isLibraryGlobal(node.AsCallExpression().Expression, "Date") {
		return l.dateForbidden(node, "Date() (the local-time string form)")
	}
	return nil
}

// A literal type or all members of its union prove the grammar. Generic strings may
// select legacy or local-time forms, and are refused even when the build uses TZ=UTC.
func dateISOType(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsStringLiteral != 0 {
		text, ok := proven.AsLiteralType().Value().(string)
		return ok && dateISOGrammar.MatchString(text) && !strings.HasPrefix(text, "-000000")
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !dateISOType(part) {
				return false
			}
		}
		return len(proven.Types()) > 0
	}
	return false
}
func (l *lowering) dateISOArgument(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	if dateISOType(l.checker.GetTypeAtLocation(node)) {
		return true
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "toISOString" && l.libraryMember(callee) && len(call.Arguments.Nodes) == 0 && l.isLibraryType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression), "Date") {
			return true
		}
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil {
				return l.dateISOArgument(declaration.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	}
	return false
}
func (l *lowering) dateISOString(node *ast.Node) (ir.Expression, error) {
	if !l.dateISOArgument(node, 0) {
		return nil, l.dateForbidden(node, "Date string parsing without a proven ISO grammar and explicit date-time zone")
	}
	return l.expression(node)
}
func (l *lowering) newDate(node *ast.Node) (ir.Expression, error) {
	written := node.AsNewExpression().Arguments
	if written == nil || len(written.Nodes) == 0 {
		return ir.NodeFSFile{Operation: "date_new_now", Of: ir.Object}, nil
	}
	if len(written.Nodes) != 1 || written.Nodes[0].Kind == ast.KindSpreadElement {
		return nil, l.dateForbidden(node, "new Date with local-time components or spread")
	}
	argument := written.Nodes[0]
	if l.isLibraryType(l.checker.GetTypeAtLocation(argument), "Date") {
		value, err := l.expression(argument)
		return ir.NodeFSFile{Operation: "date_copy", Arguments: []ir.Expression{value}, Of: ir.Object}, err
	}
	if held, _ := l.representation(l.checker.GetTypeAtLocation(argument)); held == ir.String {
		value, err := l.dateISOString(argument)
		return ir.NodeFSFile{Operation: "date_new_iso", Arguments: []ir.Expression{value}, Of: ir.Object}, err
	}
	if held, _ := l.representation(l.checker.GetTypeAtLocation(argument)); held == ir.Union {
		return nil, l.notYet(node, "new Date of a string/number union needs constructor dispatch")
	}
	value, err := l.libraryNumber(argument)
	return ir.NodeFSFile{Operation: "date_new", Arguments: []ir.Expression{value}, Of: ir.Object}, err
}
func (l *lowering) libraryDateCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	written := call.Arguments.Nodes
	if name == "hasOwnProperty" && (l.datePrototype(receiver) || l.isLibraryGlobal(receiver, "Date")) {
		if len(written) != 1 || hasSpread(node) {
			return nil, true, l.notYet(node, "Date.hasOwnProperty with other than one string key")
		}
		key, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if key.Type() != ir.String {
			return nil, true, l.notYet(node, "Date.hasOwnProperty with a non-string key")
		}
		operation := "date_constructor_own"
		if l.datePrototype(receiver) {
			operation = "date_prototype_own"
		}
		return ir.NodeFSFile{Operation: operation, Arguments: []ir.Expression{key}, Of: ir.Boolean}, true, nil
	}
	if name == "call" {
		method := ast.SkipParentheses(receiver)
		if method.Kind == ast.KindPropertyAccessExpression && l.datePrototype(method.AsPropertyAccessExpression().Expression) {
			if len(written) == 0 || !l.isLibraryType(l.checker.GetTypeAtLocation(written[0]), "Date") {
				return nil, true, l.notYet(node, "Date prototype call requires a Date internal slot")
			}
			receiver, name = written[0], method.Name().Text()
			written = written[1:]
		}
	}
	if l.isLibraryGlobal(receiver, "Date") {
		switch name {
		case "now":
			if len(written) == 0 {
				return ir.NodeFSFile{Operation: "date_now", Of: ir.Number}, true, nil
			}
		case "parse":
			if len(written) == 1 && !hasSpread(node) {
				value, err := l.dateISOString(written[0])
				return ir.NodeFSFile{Operation: "date_parse", Arguments: []ir.Expression{value}, Of: ir.Number}, true, err
			}
		case "UTC":
			if hasSpread(node) {
				return nil, true, l.notYet(node, "Date.UTC with spread arguments")
			}
			args := []ir.Expression{}
			for _, argument := range written {
				value, err := l.libraryNumber(argument)
				if err != nil {
					return nil, true, err
				}
				args = append(args, value)
			}
			return ir.NodeFSFile{Operation: "date_UTC", Arguments: args, Of: ir.Number}, true, nil
		}
		return nil, true, l.notYet(node, "Date."+name+" with these arguments")
	}
	if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date") || !l.libraryMember(callee) {
		return nil, false, nil
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if name == "toJSON" {
		args := []ir.Expression{}
		for _, argument := range written {
			if argument.Kind == ast.KindSpreadElement {
				return nil, true, l.notYet(node, "Date.toJSON spread arguments")
			}
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			args = append(args, value)
		}
		return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.String}, true, nil
	}
	if name == "setTime" || strings.HasPrefix(name, "setUTC") {
		if _, known := dateSetters[strings.Replace(name, "UTC", "", 1)]; known {
			args, err := l.dateArguments(node, written)
			return ir.DateCall{Receiver: value, Method: name, Arguments: args, Returns: ir.Number}, true, err
		}
	}
	if len(written) != 0 {
		return nil, true, l.notYet(node, "Date."+name+" with arguments")
	}
	operation, of := "date_"+name, ir.Number
	if name == "getTime" || name == "valueOf" {
		operation = "date_time"
	} else if name == "toISOString" || name == "toUTCString" {
		of = ir.String
	} else if !dateUTCGetters[name] {
		return nil, true, l.notYet(node, "Date."+name+" is not built in the UTC/ISO surface")
	}
	return ir.NodeFSFile{Operation: operation, Arguments: []ir.Expression{value}, Of: of}, true, nil
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

var dateSetters = map[string]int{"setFullYear": 0, "setMonth": 1, "setDate": 2, "setHours": 4, "setMinutes": 5, "setSeconds": 6, "setMilliseconds": 7, "setTime": 10}
