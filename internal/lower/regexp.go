package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	regex "github.com/system-inc/adamic/internal/regexp"
	"strings"
)

func (l *lowering) regexConstant(node *ast.Node) (ir.Expression, error) {
	pattern, flags := "", ""
	var evaluated []ir.Expression
	if node.Kind == ast.KindRegularExpressionLiteral {
		text := node.Text()
		end := strings.LastIndex(text, "/")
		if end <= 0 {
			return nil, l.notYet(node, "a malformed regular expression literal")
		}
		pattern, flags = text[1:end], text[end+1:]
	} else {
		var args []*ast.Node
		if node.Kind == ast.KindNewExpression {
			if node.AsNewExpression().Arguments != nil {
				args = node.AsNewExpression().Arguments.Nodes
			}
		} else {
			args = node.AsCallExpression().Arguments.Nodes
		}
		if len(args) > 2 {
			return nil, l.notYet(node, "RegExp with more than two arguments")
		}
		if len(args) > 0 {
			var ok bool
			pattern, ok = l.constantPattern(args[0], 0)
			if !ok {
				return nil, l.notYet(args[0], "RegExp with a nonconstant pattern")
			}
		}
		if len(args) > 1 {
			var ok bool
			flags, ok = l.constantPattern(args[1], 0)
			if !ok {
				return nil, l.notYet(args[1], "RegExp with nonconstant flags")
			}
		}
		for _, arg := range args {
			value, err := l.expression(arg)
			if err != nil {
				return nil, err
			}
			evaluated = append(evaluated, value)
		}
	}
	program, err := regex.Compile(pattern, flags)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid RegExp: %w", l.program.Where(node), err)
	}
	index := len(l.result.Regexps)
	declarations, err := program.NativeDeclarations(fmt.Sprintf("adamic_regex_%d", index))
	if err != nil {
		return nil, l.notYet(node, err.Error())
	}
	l.result.Regexps = append(l.result.Regexps, ir.RegExpProgram{Pattern: pattern, Flags: flags, Declarations: declarations})
	source := escapeRegexSource(pattern, flags)
	ordered := ""
	for _, flag := range "dgimsuvy" {
		if strings.ContainsRune(flags, flag) {
			ordered += string(flag)
		}
	}
	return ir.RegExpNew{Index: index, Source: l.constant(source), Flags: l.constant(ordered), Arguments: evaluated}, nil
}
func (l *lowering) constantPattern(node *ast.Node, depth int) (string, bool) {
	if depth > 32 {
		return "", false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	case ast.KindIdentifier:
		if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			return "", true
		}
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil {
				return l.constantPattern(declaration.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	case ast.KindBinaryExpression:
		b := node.AsBinaryExpression()
		if b.OperatorToken.Kind == ast.KindPlusToken {
			a, ok := l.constantPattern(b.Left, depth+1)
			c, other := l.constantPattern(b.Right, depth+1)
			return joinPatternStrings(a, c), ok && other
		}
	}
	return "", false
}

// joinPatternStrings mirrors adamic_string_put: canonical WTF-8 strings can
// form a new surrogate pair only at the concatenation boundary.
func joinPatternStrings(left, right string) string {
	if len(left) < 3 || len(right) < 3 {
		return left + right
	}
	high, low := left[len(left)-3:], right[:3]
	if high[0] != 0xed || high[1] < 0xa0 || high[1] > 0xaf || high[2] < 0x80 || high[2] > 0xbf ||
		low[0] != 0xed || low[1] < 0xb0 || low[1] > 0xbf || low[2] < 0x80 || low[2] > 0xbf {
		return left + right
	}
	highUnit := rune(0xd000) | rune(high[1]&0x3f)<<6 | rune(high[2]&0x3f)
	lowUnit := rune(0xd000) | rune(low[1]&0x3f)<<6 | rune(low[2]&0x3f)
	point := 0x10000 + ((highUnit - 0xd800) << 10) + (lowUnit - 0xdc00)
	return left[:len(left)-3] + string(point) + right[3:]
}

func (l *lowering) regexBuiltin(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if l.isLibraryGlobal(callee, "RegExp") {
		value, err := l.regexConstant(node)
		return value, true, err
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	name := callee.Name().Text()
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	args := node.AsCallExpression().Arguments.Nodes
	method := name
	result := ir.Type(0)
	switch {
	case l.isLibraryType(proven, "RegExp"):
		if name == "test" {
			result = ir.Boolean
		} else if name == "exec" {
			result = ir.Array
		} else {
			return nil, true, l.notYet(node, "RegExp."+name)
		}
		if len(args) != 1 {
			return nil, true, l.notYet(node, "RegExp."+name+" without exactly one string")
		}
	case l.isLibraryType(proven, "RegExpStringIterator"):
		if name != "next" || len(args) != 0 {
			return nil, true, l.notYet(node, "RegExp iterator."+name)
		}
		method = "next"
		result = ir.Object
	default:
		of, _ := l.representation(proven)
		if of != ir.String || len(args) == 0 || !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "RegExp") {
			return nil, false, nil
		}
		switch name {
		case "match", "split":
			result = ir.Array
		case "matchAll":
			result = ir.Object
		case "search":
			result = ir.Number
		case "replace", "replaceAll":
			result = ir.String
		default:
			return nil, true, l.notYet(node, "String."+name+" with a RegExp")
		}
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	var arguments []ir.Expression
	for _, arg := range args {
		v, e := l.expression(arg)
		if e != nil {
			return nil, true, e
		}
		arguments = append(arguments, v)
	}
	if (name == "exec" || name == "test") && arguments[0].Type() != ir.String {
		return nil, true, l.notYet(node, "RegExp input other than a string")
	}
	if name == "replace" || name == "replaceAll" {
		if len(arguments) != 2 || arguments[1].Type() != ir.String {
			return nil, true, l.notYet(node, "regex replacement other than a string")
		}
	}
	if name == "split" {
		if len(arguments) == 1 {
			arguments = append(arguments, ir.NumberConstant{Value: 4294967295})
		}
		if len(arguments) == 2 {
			if _, undefined := arguments[1].(ir.Undefined); undefined {
				arguments[1] = ir.NumberConstant{Value: 4294967295}
			}
			if arguments[1].Type() == ir.MaybeNumber {
				arguments[1] = ir.Coalesce{Value: arguments[1], Fallback: ir.NumberConstant{Value: 4294967295}, Of: ir.Number}
			}
		}
		if len(arguments) != 2 || arguments[1].Type() != ir.Number {
			return nil, true, l.notYet(node, "regex split limit other than a number")
		}
	}
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, true, l.notYet(node, "an optional RegExp call")
	}
	return ir.RegExpCall{Value: value, Arguments: arguments, Method: method, Returns: result}, true, nil
}

// EscapeRegExpPattern preserves escapes already present in the pattern.
func escapeRegexSource(pattern, flags string) string {
	if pattern == "" {
		return "(?:)"
	}
	var result strings.Builder
	escaped, sets := false, strings.Contains(flags, "v")
	depth := 0
	// Scan bytes so WTF-8 lone surrogates retain their exact code units.
	for at := 0; at < len(pattern); at++ {
		c := pattern[at]
		escape := ""
		switch c {
		case '\n':
			escape = "n"
		case '\r':
			escape = "r"
		case 0xe2:
			if at+2 < len(pattern) && pattern[at+1] == 0x80 {
				if pattern[at+2] == 0xa8 {
					escape = "u2028"
					at += 2
				}
				if escape == "" && pattern[at+2] == 0xa9 {
					escape = "u2029"
					at += 2
				}
			}
		}
		if escape != "" {
			if !escaped {
				result.WriteByte('\\')
			}
			result.WriteString(escape)
			escaped = false
			continue
		}
		if !escaped {
			if c == '[' && (sets || depth == 0) {
				depth++
			}
			if c == ']' && depth > 0 {
				depth--
			}
		}
		if c == '/' && !escaped && depth == 0 {
			result.WriteByte('\\')
		}
		result.WriteByte(c)
		if c == '\\' {
			escaped = !escaped
		} else {
			escaped = false
		}
	}
	return result.String()
}

func (l *lowering) regexGroups(node *ast.Node) bool {
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))
	for _, info := range l.checker.GetIndexInfosOfType(proven) {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(file.AsSourceFile().FileName(), ".regexp.") {
				return true
			}
		}
	}
	return false
}

func (l *lowering) regexUnsupportedUse(node *ast.Node) error {
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))
	if !l.isLibraryType(proven, "RegExp", "RegExpStringIterator") && !l.regexGroups(node) {
		return nil
	}
	parent := node.Parent
	if parent == nil {
		return nil
	}
	if parent.Kind == ast.KindSpreadAssignment {
		return l.notYet(parent, "spreading a RegExp or its iterator")
	}
	if (parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) && parent.Parent != nil && parent.Parent.Kind == ast.KindBinaryExpression {
		assignment := parent.Parent.AsBinaryExpression()
		lastIndex := parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "lastIndex"
		if parent.Kind == ast.KindElementAccessExpression {
			key := ast.SkipParentheses(parent.AsElementAccessExpression().ArgumentExpression)
			lastIndex = key.Kind == ast.KindStringLiteral && key.Text() == "lastIndex"
		}
		if assignment.Left == parent && ast.IsAssignmentOperator(assignment.OperatorToken.Kind) && (!lastIndex || l.regexGroups(node)) {
			return l.notYet(parent, "overriding a RegExp or iterator property")
		}
	}
	return nil
}

// These library fields have explicit native storage. Other library members remain inherited
// prototype reads, so recognizing regex metadata must not admit arbitrary own-field loads.
func (l *lowering) regexRuntimeProperty(receiver *ast.Node, name string) bool {
	if l.regexGroups(receiver) {
		// Named-group dictionaries have dynamic own keys; regexGroups lowering checks their value type.
		return true
	}
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	switch {
	case l.isLibraryType(proven, "RegExp"):
		switch name {
		case "lastIndex", "source", "flags", "global", "ignoreCase", "multiline", "unicode", "sticky", "hasIndices", "unicodeSets", "dotAll":
			return true
		}
	case l.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray"):
		return name == "index" || name == "input" || name == "groups" || name == "indices"
	}
	if name != "done" && name != "value" {
		return false
	}
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if !l.isLibraryType(member, "IteratorYieldResult", "IteratorReturnResult") {
			return false
		}
	}
	return true
}
