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

func (lowering *lowering) regexConstant(node *ast.Node) (ir.Expression, error) {
	pattern, flags := "", ""
	var evaluated []ir.Expression
	if node.Kind == ast.KindRegularExpressionLiteral {
		var written bool
		pattern, flags, written = regexLiteralParts(node)
		if !written {
			return nil, lowering.notYet(node, "a malformed regular expression literal")
		}
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
			return nil, lowering.notYet(node, "RegExp with more than two arguments")
		}
		if len(args) > 0 {
			var ok bool
			pattern, ok = lowering.constantPattern(args[0], 0)
			if !ok {
				return nil, lowering.notYet(args[0], "RegExp with a nonconstant pattern")
			}
		}
		if len(args) > 1 {
			var ok bool
			flags, ok = lowering.constantPattern(args[1], 0)
			if !ok {
				return nil, lowering.notYet(args[1], "RegExp with nonconstant flags")
			}
		}
		for _, arg := range args {
			value, expressionError := lowering.expression(arg)
			if expressionError != nil {
				return nil, expressionError
			}
			evaluated = append(evaluated, value)
		}
	}
	program, err := regex.Compile(pattern, flags)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid RegExp: %w", lowering.program.Where(node), err)
	}
	index := len(lowering.result.Regexps)
	declarations, err := program.NativeDeclarations(fmt.Sprintf("adamic_regex_%d", index))
	if err != nil {
		if input, literal := regexLiteralTestInput(node); literal {
			declarations, err = regex.NativeLiteralTestDeclarations(
				pattern,
				flags,
				input,
				fmt.Sprintf("adamic_regex_%d", index),
			)
		}
	}
	if err != nil {
		return nil, lowering.notYet(node, err.Error())
	}
	lowering.result.Regexps = append(
		lowering.result.Regexps,
		ir.RegExpProgram{Pattern: pattern, Flags: flags, Declarations: declarations},
	)
	source := escapeRegexSource(pattern, flags)
	ordered := ""
	for _, flag := range "dgimsuvy" {
		if strings.ContainsRune(flags, flag) {
			ordered += string(flag)
		}
	}
	return ir.RegExpNew{
		Index:     index,
		Source:    lowering.constant(source),
		Flags:     lowering.constant(ordered),
		Arguments: evaluated,
	}, nil
}

// regexLiteralParts is a RegExp literal's pattern, between its slashes, and its flags, after the last one.
func regexLiteralParts(node *ast.Node) (string, string, bool) {
	text := node.Text()
	end := strings.LastIndex(text, "/")
	if end <= 0 {
		return "", "", false
	}
	return text[1:end], text[end+1:], true
}

func (lowering *lowering) constantPattern(node *ast.Node, depth int) (string, bool) {
	if depth > 32 {
		return "", false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	case ast.KindIdentifier:
		if lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			return "", true
		}
		symbol := lowering.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration &&
				declaration.Parent != nil &&
				declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
				declaration.AsVariableDeclaration().Initializer != nil {
				return lowering.constantPattern(declaration.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	case ast.KindBinaryExpression:
		expression := node.AsBinaryExpression()
		if expression.OperatorToken.Kind == ast.KindPlusToken {
			a, ok := lowering.constantPattern(expression.Left, depth+1)
			c, other := lowering.constantPattern(expression.Right, depth+1)
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

func (lowering *lowering) regexBuiltin(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if lowering.isLibraryGlobal(callee, "RegExp") {
		value, regexConstantError := lowering.regexConstant(node)
		return value, true, regexConstantError
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	name := callee.Name().Text()
	proven := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(receiver))
	args := node.AsCallExpression().Arguments.Nodes
	method := name
	result := ir.Type(0)
	switch {
	case lowering.isLibraryType(proven, "RegExp"):
		if name == "test" {
			result = ir.Boolean
		} else if name == "exec" {
			result = ir.Array
		} else {
			return nil, true, lowering.notYet(node, "RegExp."+name)
		}
		if len(args) != 1 {
			return nil, true, lowering.notYet(node, "RegExp."+name+" without exactly one string")
		}
	case lowering.isLibraryType(proven, "RegExpStringIterator"):
		if name != "next" || len(args) != 0 {
			return nil, true, lowering.notYet(node, "RegExp iterator."+name)
		}
		method = "next"
		result = ir.Object
	default:
		of, _ := lowering.representation(proven)
		if of != ir.String ||
			len(args) == 0 ||
			!lowering.isLibraryType(
				lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(args[0])),
				"RegExp",
			) {
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
			return nil, true, lowering.notYet(node, "String."+name+" with a RegExp")
		}
	}
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, true, lowering.notYet(node, "an optional RegExp call")
	}
	var arguments []ir.Expression
	for _, arg := range args {
		expression, expressionError := lowering.expression(arg)
		if expressionError != nil {
			return nil, true, expressionError
		}
		arguments = append(arguments, expression)
	}
	if (name == "exec" || name == "test") && arguments[0].Type() != ir.String {
		return nil, true, lowering.notYet(node, "RegExp input other than a string")
	}
	if name == "replace" || name == "replaceAll" {
		if len(arguments) == 2 && arguments[1].Type() == ir.Closure {
			return lowering.regexReplacement(node, value, arguments, name, args[1])
		}
		if len(arguments) != 2 || arguments[1].Type() != ir.String {
			return nil, true, lowering.notYet(node, "regex replacement other than a string")
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
				arguments[1] = ir.Coalesce{
					Value:    arguments[1],
					Fallback: ir.NumberConstant{Value: 4294967295},
					Of:       ir.Number,
				}
			}
		}
		if len(arguments) != 2 || arguments[1].Type() != ir.Number {
			return nil, true, lowering.notYet(node, "regex split limit other than a number")
		}
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
		character := pattern[at]
		escape := ""
		switch character {
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
			if character == '[' && (sets || depth == 0) {
				depth++
			}
			if character == ']' && depth > 0 {
				depth--
			}
		}
		if character == '/' && !escaped && depth == 0 {
			result.WriteByte('\\')
		}
		result.WriteByte(character)
		if character == '\\' {
			escaped = !escaped
		} else {
			escaped = false
		}
	}
	return result.String()
}

func (lowering *lowering) regexGroups(node *ast.Node) bool {
	proven := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(node))
	for _, info := range lowering.checker.GetIndexInfosOfType(proven) {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(file.AsSourceFile().FileName().AsString(), ".regexp.") {
				return true
			}
		}
	}
	return false
}

func (lowering *lowering) regexUnsupportedUse(node *ast.Node) error {
	proven := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(node))
	if !lowering.isLibraryType(proven, "RegExp", "RegExpStringIterator") && !lowering.regexGroups(node) {
		return nil
	}
	parent := node.Parent
	if parent == nil {
		return nil
	}
	if parent.Kind == ast.KindSpreadAssignment {
		return lowering.notYet(parent, "spreading a RegExp or its iterator")
	}
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.Parent != nil &&
		parent.Parent.Kind == ast.KindBinaryExpression {
		assignment := parent.Parent.AsBinaryExpression()
		if assignment.Left == parent &&
			assignment.OperatorToken.Kind == ast.KindEqualsToken &&
			(parent.Name().Text() != "lastIndex" || lowering.regexGroups(node)) {
			return lowering.notYet(parent, "overriding a RegExp or iterator property")
		}
	}
	return nil
}

// These library fields have explicit native storage. Other library members remain inherited
// prototype reads, so recognizing regex metadata must not admit arbitrary own-field loads.
func (lowering *lowering) regexRuntimeProperty(receiver *ast.Node, name string) bool {
	if lowering.regexGroups(receiver) {
		// Named-group dictionaries have dynamic own keys; regexGroups lowering checks their value type.
		return true
	}
	proven := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(receiver))
	switch {
	case lowering.isLibraryType(proven, "RegExp"):
		switch name {
		case "lastIndex", "source", "flags", "global", "ignoreCase", "multiline", "unicode", "sticky", "hasIndices", "unicodeSets", "dotAll":
			return true
		}
	case lowering.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray"):
		return name == "index" || name == "input" || name == "groups" || name == "indices"
	}
	return false
}
