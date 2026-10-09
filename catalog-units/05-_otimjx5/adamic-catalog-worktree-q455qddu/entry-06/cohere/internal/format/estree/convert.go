package estree

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	ecmascripttext "github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

/*
 * typescript-estree's convert.js, ported over typescript-go's AST.
 *
 * Upstream is @typescript-eslint/typescript-estree 8.65.0, dist/convert.js, the version the Prettier
 * fork bundles. Each case below is the upstream case of the same SyntaxKind, in upstream's order, and
 * builds the same properties. Properties nobody reads still get built where upstream builds them,
 * because the printer reads nodes generically (a visitor-key walk, `key in node`), and a property that
 * is missing here but present upstream is the kind of difference nobody finds by reading.
 *
 * Where typescript-go's tree differs from TypeScript's, the difference is absorbed here and named:
 *
 *   - Modifiers and decorators share one ModifierList. Upstream asks ts.getModifiers and
 *     ts.getDecorators separately; modifiers() and decorators() below split the list the same way.
 *   - A nested namespace (`namespace A.B {}`) carries a zero-width export modifier the parser
 *     synthesizes and flags Reparsed. TypeScript's tree has no such modifier, so reparsed modifiers
 *     are dropped.
 *   - `declare global` is a ModuleDeclaration whose Keyword is GlobalKeyword, where TypeScript sets
 *     NodeFlags.GlobalAugmentation; `namespace` versus `module` is the Keyword too.
 *
 * Upstream's syntax checks (check-syntax-errors.js) are not ported. They reject programs TypeScript
 * parsed but ESTree cannot represent, the oracle fails on those files, and a file the oracle fails on
 * leaves the differential's denominator. Parse errors themselves are refused before conversion.
 */

// Converter turns one typescript-go source file into an ESTree Program.
type Converter struct {
	sourceFile   *ast.SourceFile
	text         string
	allowPattern bool
	scan         *scanner.Scanner
	unsupported  error

	// typeCastCommentEnds are the ends of the file's type cast comments, in order, when converting
	// for the babel parser (ConvertForBabel). Nil converts for typescript.
	typeCastCommentEnds []int

	// nodes is where the tree's nodes come from, or nil for the heap.
	nodes *Arena
}

// ConvertForBabel is Convert for a file Prettier would parse with babel (see ParseJavaScript). The
// one difference in the tree is postprocess's: Babel keeps a parenthesized expression as a node
// when a type cast comment, `/** @type {T} */ (value)`, comes right before it.
func ConvertForBabel(sourceFile *ast.SourceFile, nodes *Arena) (*Node, []*Node, error) {
	// postprocess merges nestled JSDoc comments before it looks for casts, so this does too.
	ends := []int{}
	for _, comment := range mergeNestledJsdocComments(collectComments(sourceFile)) {
		if IsTypeCastComment(comment) {
			ends = append(ends, LocEnd(comment))
		}
	}
	return convert(sourceFile, ends, nodes)
}

// followsTypeCastComment is postprocess's test for keeping a ParenthesizedExpression: the last type
// cast comment that ends at or before start is separated from it by whitespace alone.
func (converter *Converter) followsTypeCastComment(start int) bool {
	ends := converter.typeCastCommentEnds
	index := sort.SearchInts(ends, start+1) - 1
	return index >= 0 && TrimJavaScript(converter.text[ends[index]:start]) == ""
}

// IsTypeCastComment is upstream's isTypeCastComment, utilities/is-type-cast-comment.js. Upstream
// memoizes it per comment; the answer reads only the comment's value, so computing it each time is
// the same.
func IsTypeCastComment(comment *Node) bool {
	return comment.Is("Block", "CommentBlock", "MultiLine") &&
		strings.HasPrefix(comment.String("value"), "*") &&
		// TypeScript expects the type to be enclosed in curly brackets, however
		// Closure Compiler accepts types in parens and even without any delimiters at all.
		// That's why we just search for "@type" and "@satisfies".
		typeCastCommentPattern.MatchString(comment.String("value"))
}

var typeCastCommentPattern = regexp.MustCompile(`@(?:type|satisfies)\b`)

// Convert parses nothing: it converts a parsed source file, and returns the Program with its comments.
//
// The file must have parsed without diagnostics. Upstream throws the first parse diagnostic before
// converting, and so does this. The tree's nodes come from nodes, or the heap when it is nil.
func Convert(sourceFile *ast.SourceFile, nodes *Arena) (program *Node, comments []*Node, err error) {
	return convert(sourceFile, nil, nodes)
}

func convert(sourceFile *ast.SourceFile, typeCastCommentEnds []int, nodes *Arena) (program *Node, comments []*Node, err error) {
	if diagnostics := sourceFile.Diagnostics(); len(diagnostics) > 0 {
		return nil, nil, fmt.Errorf("parse error at %d: %s", diagnostics[0].Pos(), diagnostics[0].String())
	}
	converter := &Converter{sourceFile: sourceFile, text: sourceFile.Text(), typeCastCommentEnds: typeCastCommentEnds, nodes: nodes}
	converter.scan = scanner.NewScanner()
	converter.scan.SetText(converter.text)

	defer func() {
		if recovered := recover(); recovered != nil {
			program, comments, err = nil, nil, fmt.Errorf("converting to ESTree: %v", recovered)
		}
	}()

	program = converter.converter(sourceFile.AsNode(), nil, false)
	if converter.unsupported != nil {
		return nil, nil, converter.unsupported
	}
	comments = collectComments(sourceFile)
	return program, comments, nil
}

// getStart is node.getStart(ast): the node's start after its leading trivia.
func (converter *Converter) getStart(node *ast.Node) int {
	return scanner.GetTokenPosOfNode(node, converter.sourceFile, false)
}

// getRange is upstream's getRange.
func (converter *Converter) getRange(node *ast.Node) [2]int {
	return [2]int{converter.getStart(node), node.End()}
}

// getText is node.getText().
func (converter *Converter) getText(node *ast.Node) string {
	return converter.text[converter.getStart(node):node.End()]
}

// token is one scanned token: what findNextToken and getFirstToken return in upstream.
type token struct {
	kind  ast.Kind
	start int
	end   int
}

// tokenAt scans the token whose full start is pos, which is upstream's findNextToken: the token
// starting at the end of the previous one, trivia included.
func (converter *Converter) tokenAt(pos int) token {
	converter.scan.ResetPos(pos)
	kind := converter.scan.Scan()
	return token{kind: kind, start: converter.scan.TokenStart(), end: converter.scan.TokenEnd()}
}

// firstToken is node.getFirstToken().
func (converter *Converter) firstToken(node *ast.Node) token {
	return converter.tokenAt(converter.getStart(node))
}

// createNode is upstream's createNode: the range defaults to the TypeScript node's.
func (converter *Converter) createNode(node *ast.Node, nodeType string, keysAndValues ...any) *Node {
	nodeRange := converter.getRange(node)
	return converter.nodes.node(nodeType, nodeRange[0], nodeRange[1], keysAndValues)
}

// createNodeWithRange is createNode with an explicit range, upstream's `range: [...]` in the data.
func (converter *Converter) createNodeWithRange(nodeType string, nodeRange [2]int, keysAndValues ...any) *Node {
	return converter.nodes.node(nodeType, nodeRange[0], nodeRange[1], keysAndValues)
}

// noNodes is an empty node list made a property value once. A []*Node{} given to createNode is boxed
// into its value on every node that has one, which was 1.5M of the conversion's allocations on ahra for
// identifiers' decorators alone (#fyw36kf). Sharing one is safe: nothing can be stored in an empty list,
// and appending to it moves the result to new memory.
var noNodes any = []*Node{}

// fixParentLocation is upstream's fixParentLocation.
func fixParentLocation(result *Node, childRange [2]int) {
	if childRange[0] < result.Range[0] {
		result.Range[0] = childRange[0]
	}
	if childRange[1] > result.Range[1] {
		result.Range[1] = childRange[1]
	}
}

// modifiers is getModifiers(node): the modifier keywords, without decorators.
func modifiers(node *ast.Node) []*ast.Node {
	if node == nil || !ast.CanHaveModifiers(node) {
		return nil
	}
	list := node.Modifiers()
	if list == nil {
		return nil
	}
	var result []*ast.Node
	for _, modifier := range list.Nodes {
		if modifier.Kind == ast.KindDecorator || modifier.Flags&ast.NodeFlagsReparsed != 0 {
			continue
		}
		result = append(result, modifier)
	}
	return result
}

// decorators is getDecorators(node).
func decorators(node *ast.Node) []*ast.Node {
	if node == nil || !ast.CanHaveDecorators(node) {
		return nil
	}
	list := node.Modifiers()
	if list == nil {
		return nil
	}
	var result []*ast.Node
	for _, modifier := range list.Nodes {
		if modifier.Kind == ast.KindDecorator {
			result = append(result, modifier)
		}
	}
	return result
}

// hasModifier is upstream's hasModifier.
func hasModifier(kind ast.Kind, node *ast.Node) bool {
	for _, modifier := range modifiers(node) {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}

// getTSNodeAccessibility is upstream's getTSNodeAccessibility. Upstream returns undefined when there
// is none, so nil here, which a Node stores as a present property with no value.
func getTSNodeAccessibility(node *ast.Node) any {
	for _, modifier := range modifiers(node) {
		switch modifier.Kind {
		case ast.KindPublicKeyword:
			return "public"
		case ast.KindProtectedKeyword:
			return "protected"
		case ast.KindPrivateKeyword:
			return "private"
		}
	}
	return nil
}

// getDeclarationKind is upstream's getDeclarationKind.
func getDeclarationKind(node *ast.Node) string {
	switch {
	case node.Flags&ast.NodeFlagsLet != 0:
		return "let"
	case node.Flags&ast.NodeFlagsAwaitUsing == ast.NodeFlagsAwaitUsing:
		return "await using"
	case node.Flags&ast.NodeFlagsConst != 0:
		return "const"
	case node.Flags&ast.NodeFlagsUsing != 0:
		return "using"
	}
	return "var"
}

func isComputedProperty(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindComputedPropertyName
}

func isOptional(node *ast.Node) bool {
	return node.QuestionToken() != nil
}

// canContainDirective is upstream's canContainDirective.
func canContainDirective(node *ast.Node) bool {
	if node.Kind == ast.KindBlock {
		switch node.Parent.Kind {
		case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindArrowFunction,
			ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindMethodDeclaration:
			return true
		default:
			return false
		}
	}
	return true
}

// convertChild is upstream's convertChild: patterns are not allowed below it.
func (converter *Converter) convertChild(node *ast.Node, parent *ast.Node) *Node {
	return converter.converter(node, parent, false)
}

// convertPattern is upstream's convertPattern.
func (converter *Converter) convertPattern(node *ast.Node, parent *ast.Node) *Node {
	return converter.converter(node, parent, true)
}

// convertChildren is upstream's convertChildren. A nil result stays in the list, as upstream's map
// keeps it, which is how array holes survive.
func (converter *Converter) convertChildren(nodes []*ast.Node, parent *ast.Node) []*Node {
	result := make([]*Node, len(nodes))
	for index, node := range nodes {
		result[index] = converter.converter(node, parent, false)
	}
	return result
}

// converter is upstream's converter.
func (converter *Converter) converter(node *ast.Node, parent *ast.Node, allowPattern bool) *Node {
	if node == nil {
		return nil
	}
	pattern := converter.allowPattern
	converter.allowPattern = allowPattern
	parentNode := parent
	if parentNode == nil {
		parentNode = node.Parent
	}
	result := converter.convertNode(node, parentNode)
	converter.allowPattern = pattern
	return result
}

// convertBindingNameWithTypeAnnotation is upstream's convertBindingNameWithTypeAnnotation.
func (converter *Converter) convertBindingNameWithTypeAnnotation(name *ast.Node, tsType *ast.Node, parent *ast.Node) *Node {
	id := converter.convertPattern(name, nil)
	if tsType != nil {
		annotation := converter.convertTypeAnnotation(tsType, parent)
		id.Set("typeAnnotation", annotation)
		fixParentLocation(id, annotation.Range)
	}
	return id
}

// convertBodyExpressions is upstream's convertBodyExpressions.
func (converter *Converter) convertBodyExpressions(nodes []*ast.Node, parent *ast.Node) []*Node {
	allowDirectives := canContainDirective(parent)
	result := make([]*Node, 0, len(nodes))
	for _, statement := range nodes {
		child := converter.convertChild(statement, nil)
		if allowDirectives {
			if child != nil && child.Child("expression") != nil && statement.Kind == ast.KindExpressionStatement &&
				statement.Expression().Kind == ast.KindStringLiteral {
				raw := child.Child("expression").String("raw")
				child.Set("directive", raw[1:len(raw)-1])
				result = append(result, child)
				continue
			}
			allowDirectives = false
		}
		if child != nil {
			result = append(result, child)
		}
	}
	return result
}

// convertChainExpression is upstream's convertChainExpression.
func (converter *Converter) convertChainExpression(node *Node, tsNode *ast.Node) *Node {
	var child *Node
	isOptionalChain := false
	switch node.Type() {
	case "MemberExpression":
		child, isOptionalChain = node.Child("object"), node.Bool("optional")
	case "CallExpression":
		child, isOptionalChain = node.Child("callee"), node.Bool("optional")
	default:
		child = node.Child("expression")
	}

	// isChildUnwrappableOptionalChain: (x?.y).z is semantically different, so .z is not optional.
	isChildUnwrappable := child.Is("ChainExpression") && tsNode.Expression().Kind != ast.KindParenthesizedExpression
	if !isChildUnwrappable && !isOptionalChain {
		return node
	}
	if isChildUnwrappable && child.Is("ChainExpression") {
		newChild := child.Child("expression")
		switch node.Type() {
		case "MemberExpression":
			node.Set("object", newChild)
		case "CallExpression":
			node.Set("callee", newChild)
		default:
			node.Set("expression", newChild)
		}
	}
	return converter.createNode(tsNode, "ChainExpression", "expression", node)
}

// convertTypeAnnotation is upstream's convertTypeAnnotation: an intermediary TSTypeAnnotation whose
// range starts at the colon, or at the arrow in function and constructor types.
func (converter *Converter) convertTypeAnnotation(child *ast.Node, parent *ast.Node) *Node {
	offset := 1
	if parent != nil && (parent.Kind == ast.KindFunctionType || parent.Kind == ast.KindConstructorType) {
		offset = 2
	}
	start := child.Pos() - offset
	return New("TSTypeAnnotation", start, child.End(), "typeAnnotation", converter.convertChild(child, nil))
}

// convertTypeArguments is upstream's convertTypeArguments.
func (converter *Converter) convertTypeArguments(node *ast.Node) *Node {
	list := node.TypeArgumentList()
	if list == nil {
		return nil
	}
	greaterThan := converter.tokenAt(list.End())
	return converter.createNodeWithRange("TSTypeParameterInstantiation", [2]int{list.Pos() - 1, greaterThan.end},
		"params", converter.convertChildren(list.Nodes, nil))
}

// convertTypeParameters is upstream's convertTypeParameters.
func (converter *Converter) convertTypeParameters(node *ast.Node) *Node {
	list := node.TypeParameterList()
	if list == nil {
		return nil
	}
	greaterThan := converter.tokenAt(list.End())
	return converter.createNodeWithRange("TSTypeParameterDeclaration", [2]int{list.Pos() - 1, greaterThan.end},
		"params", converter.convertChildren(list.Nodes, nil))
}

// convertParameters is upstream's convertParameters.
func (converter *Converter) convertParameters(parameters []*ast.Node) []*Node {
	result := make([]*Node, 0, len(parameters))
	for _, parameter := range parameters {
		converted := converter.convertChild(parameter, nil)
		converted.Set("decorators", converter.convertChildren(decorators(parameter), nil))
		result = append(result, converted)
	}
	return result
}

func (converter *Converter) parameterList(node *ast.Node) []*ast.Node {
	list := node.ParameterList()
	if list == nil {
		return nil
	}
	return list.Nodes
}

// returnType is upstream's `node.type && this.convertTypeAnnotation(node.type, node)`, which is
// undefined when there is no type.
func (converter *Converter) returnType(node *ast.Node) *Node {
	if node.Type() == nil {
		return nil
	}
	return converter.convertTypeAnnotation(node.Type(), node)
}

// convertImportAttributes is upstream's convertImportAttributes.
func (converter *Converter) convertImportAttributes(attributes *ast.Node) []*Node {
	if attributes == nil {
		return []*Node{}
	}
	return converter.convertChildren(attributes.AsImportAttributes().Attributes.Nodes, nil)
}

// convertJSXIdentifier is upstream's convertJSXIdentifier.
func (converter *Converter) convertJSXIdentifier(node *ast.Node) *Node {
	return converter.createNode(node, "JSXIdentifier", "name", converter.getText(node))
}

// convertJSXNamespaceOrIdentifier is upstream's convertJSXNamespaceOrIdentifier, for TypeScript 5.1
// and later where JsxNamespacedName is its own node.
func (converter *Converter) convertJSXNamespaceOrIdentifier(node *ast.Node) *Node {
	if node.Kind == ast.KindJsxNamespacedName {
		namespaced := node.AsJsxNamespacedName()
		return converter.createNode(node, "JSXNamespacedName",
			"name", converter.createNode(namespaced.Name(), "JSXIdentifier", "name", namespaced.Name().Text()),
			"namespace", converter.createNode(namespaced.Namespace, "JSXIdentifier", "name", namespaced.Namespace.Text()))
	}
	text := converter.getText(node)
	if colon := strings.Index(text, ":"); colon > 0 {
		nodeRange := converter.getRange(node)
		return converter.createNodeWithRange("JSXNamespacedName", nodeRange,
			"name", converter.createNodeWithRange("JSXIdentifier", [2]int{nodeRange[0] + colon + 1, nodeRange[1]}, "name", text[colon+1:]),
			"namespace", converter.createNodeWithRange("JSXIdentifier", [2]int{nodeRange[0], nodeRange[0] + colon}, "name", text[:colon]))
	}
	return converter.convertJSXIdentifier(node)
}

// convertJSXTagName is upstream's convertJSXTagName.
func (converter *Converter) convertJSXTagName(node *ast.Node, parent *ast.Node) *Node {
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		return converter.createNode(node, "JSXMemberExpression",
			"object", converter.convertJSXTagName(access.Expression, parent),
			"property", converter.convertJSXIdentifier(access.Name()))
	}
	return converter.convertJSXNamespaceOrIdentifier(node)
}

// convertMethodSignature is upstream's convertMethodSignature.
func (converter *Converter) convertMethodSignature(node *ast.Node) *Node {
	kind := "method"
	switch node.Kind {
	case ast.KindGetAccessor:
		kind = "get"
	case ast.KindSetAccessor:
		kind = "set"
	}
	return converter.createNode(node, "TSMethodSignature",
		"accessibility", getTSNodeAccessibility(node),
		"computed", isComputedProperty(node.Name()),
		"key", converter.convertChild(node.Name(), nil),
		"kind", kind,
		"optional", isOptional(node),
		"params", converter.convertParameters(converter.parameterList(node)),
		"readonly", hasModifier(ast.KindReadonlyKeyword, node),
		"returnType", converter.returnType(node),
		"static", hasModifier(ast.KindStaticKeyword, node),
		"typeParameters", converter.convertTypeParameters(node))
}

// fixExports is upstream's fixExports.
func (converter *Converter) fixExports(node *ast.Node, result *Node) *Node {
	isNamespaceNode := node.Kind == ast.KindModuleDeclaration && node.Name().Kind != ast.KindStringLiteral
	var nodeModifiers []*ast.Node
	if isNamespaceNode {
		nodeModifiers = getNamespaceModifiers(node)
	} else {
		nodeModifiers = modifiers(node)
	}
	if len(nodeModifiers) == 0 || nodeModifiers[0].Kind != ast.KindExportKeyword {
		return result
	}

	exportKeyword := nodeModifiers[0]
	var nextModifier *ast.Node
	if len(nodeModifiers) > 1 {
		nextModifier = nodeModifiers[1]
	}
	declarationIsDefault := nextModifier != nil && nextModifier.Kind == ast.KindDefaultKeyword
	var varToken token
	if declarationIsDefault {
		varToken = converter.tokenAt(nextModifier.End())
	} else {
		varToken = converter.tokenAt(exportKeyword.End())
	}
	result.Range[0] = varToken.start

	if declarationIsDefault {
		return converter.createNodeWithRange("ExportDefaultDeclaration", [2]int{converter.getStart(exportKeyword), result.Range[1]},
			"declaration", result,
			"exportKind", "value")
	}

	isType := result.Type() == "TSInterfaceDeclaration" || result.Type() == "TSTypeAliasDeclaration"
	isDeclare := result.Bool("declare")
	exportKind := "value"
	if isType || isDeclare {
		exportKind = "type"
	}
	return converter.createNodeWithRange("ExportNamedDeclaration", [2]int{converter.getStart(exportKeyword), result.Range[1]},
		"attributes", noNodes,
		"declaration", result,
		"exportKind", exportKind,
		"source", nil,
		"specifiers", noNodes)
}

// getNamespaceModifiers is upstream's getNamespaceModifiers: nested namespaces use the topmost
// namespace's modifiers.
func getNamespaceModifiers(node *ast.Node) []*ast.Node {
	nodeModifiers := modifiers(node)
	moduleDeclaration := node
	for len(nodeModifiers) == 0 && moduleDeclaration.Parent != nil && moduleDeclaration.Parent.Kind == ast.KindModuleDeclaration {
		if parentModifiers := modifiers(moduleDeclaration.Parent); len(parentModifiers) > 0 {
			nodeModifiers = parentModifiers
		}
		moduleDeclaration = moduleDeclaration.Parent
	}
	return nodeModifiers
}

// isThisInTypeQuery is upstream's isThisInTypeQuery.
func isThisInTypeQuery(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier || scanner.IdentifierToKeywordKind(node.AsIdentifier()) != ast.KindThisKeyword {
		return false
	}
	for node.Parent.Kind == ast.KindQualifiedName && node.Parent.AsQualifiedName().Left == node {
		node = node.Parent
	}
	return node.Parent.Kind == ast.KindTypeQuery
}

// keywordTypeNames are the keyword types upstream converts to `TS${SyntaxKind[kind]}`.
var keywordTypeNames = map[ast.Kind]string{
	ast.KindAnyKeyword:       "TSAnyKeyword",
	ast.KindBigIntKeyword:    "TSBigIntKeyword",
	ast.KindBooleanKeyword:   "TSBooleanKeyword",
	ast.KindNeverKeyword:     "TSNeverKeyword",
	ast.KindNumberKeyword:    "TSNumberKeyword",
	ast.KindObjectKeyword:    "TSObjectKeyword",
	ast.KindStringKeyword:    "TSStringKeyword",
	ast.KindSymbolKeyword:    "TSSymbolKeyword",
	ast.KindUnknownKeyword:   "TSUnknownKeyword",
	ast.KindVoidKeyword:      "TSVoidKeyword",
	ast.KindUndefinedKeyword: "TSUndefinedKeyword",
	ast.KindIntrinsicKeyword: "TSIntrinsicKeyword",
}

// identifier builds upstream's Identifier object literal.
func (converter *Converter) identifierNode(nodeRange [2]int, name string) *Node {
	return converter.createNodeWithRange("Identifier", nodeRange,
		"decorators", noNodes,
		"name", name,
		"optional", false,
		"typeAnnotation", nil)
}

// convertNode is upstream's convertNode.
func (converter *Converter) convertNode(node *ast.Node, parent *ast.Node) *Node {
	switch node.Kind {
	case ast.KindSourceFile:
		// Prettier tries sourceType "module" first for a .ts or .tsx path, and typescript-estree then
		// sets every file's external module indicator, so the Program is always a module here.
		sourceFile := node.AsSourceFile()
		sourceType := "module"
		return converter.createNodeWithRange("Program", [2]int{converter.getStart(node), sourceFile.EndOfFileToken.End()},
			"body", converter.convertBodyExpressions(sourceFile.Statements.Nodes, node),
			"comments", nil,
			"sourceType", sourceType,
			"tokens", nil)

	case ast.KindBlock:
		return converter.createNode(node, "BlockStatement",
			"body", converter.convertBodyExpressions(node.Statements(), node))

	case ast.KindIdentifier:
		if isThisInTypeQuery(node) {
			return converter.createNode(node, "ThisExpression")
		}
		return converter.identifierNode(converter.getRange(node), node.Text())

	case ast.KindPrivateIdentifier:
		return converter.createNode(node, "PrivateIdentifier", "name", node.Text()[1:])

	case ast.KindWithStatement:
		withStatement := node.AsWithStatement()
		return converter.createNode(node, "WithStatement",
			"body", converter.convertChild(withStatement.Statement, nil),
			"object", converter.convertChild(withStatement.Expression, nil))

	case ast.KindReturnStatement:
		return converter.createNode(node, "ReturnStatement",
			"argument", converter.convertChild(node.Expression(), nil))

	case ast.KindLabeledStatement:
		labeled := node.AsLabeledStatement()
		return converter.createNode(node, "LabeledStatement",
			"body", converter.convertChild(labeled.Statement, nil),
			"label", converter.convertChild(labeled.Label, nil))

	case ast.KindContinueStatement:
		return converter.createNode(node, "ContinueStatement",
			"label", converter.convertChild(node.AsContinueStatement().Label, nil))

	case ast.KindBreakStatement:
		return converter.createNode(node, "BreakStatement",
			"label", converter.convertChild(node.AsBreakStatement().Label, nil))

	case ast.KindIfStatement:
		ifStatement := node.AsIfStatement()
		return converter.createNode(node, "IfStatement",
			"alternate", converter.convertChild(ifStatement.ElseStatement, nil),
			"consequent", converter.convertChild(ifStatement.ThenStatement, nil),
			"test", converter.convertChild(ifStatement.Expression, nil))

	case ast.KindSwitchStatement:
		switchStatement := node.AsSwitchStatement()
		return converter.createNode(node, "SwitchStatement",
			"cases", converter.convertChildren(switchStatement.CaseBlock.AsCaseBlock().Clauses.Nodes, nil),
			"discriminant", converter.convertChild(switchStatement.Expression, nil))

	case ast.KindCaseClause, ast.KindDefaultClause:
		clause := node.AsCaseOrDefaultClause()
		var test *Node
		if node.Kind == ast.KindCaseClause {
			test = converter.convertChild(clause.Expression, nil)
		}
		return converter.createNode(node, "SwitchCase",
			"consequent", converter.convertChildren(clause.Statements.Nodes, nil),
			"test", test)

	case ast.KindThrowStatement:
		return converter.createNode(node, "ThrowStatement",
			"argument", converter.convertChild(node.Expression(), nil))

	case ast.KindTryStatement:
		tryStatement := node.AsTryStatement()
		return converter.createNode(node, "TryStatement",
			"block", converter.convertChild(tryStatement.TryBlock, nil),
			"finalizer", converter.convertChild(tryStatement.FinallyBlock, nil),
			"handler", converter.convertChild(tryStatement.CatchClause, nil))

	case ast.KindCatchClause:
		catchClause := node.AsCatchClause()
		var param *Node
		if catchClause.VariableDeclaration != nil {
			declaration := catchClause.VariableDeclaration
			param = converter.convertBindingNameWithTypeAnnotation(declaration.Name(), declaration.Type(), nil)
		}
		return converter.createNode(node, "CatchClause",
			"body", converter.convertChild(catchClause.Block, nil),
			"param", param)

	case ast.KindWhileStatement:
		whileStatement := node.AsWhileStatement()
		return converter.createNode(node, "WhileStatement",
			"body", converter.convertChild(whileStatement.Statement, nil),
			"test", converter.convertChild(whileStatement.Expression, nil))

	case ast.KindDoStatement:
		doStatement := node.AsDoStatement()
		return converter.createNode(node, "DoWhileStatement",
			"body", converter.convertChild(doStatement.Statement, nil),
			"test", converter.convertChild(doStatement.Expression, nil))

	case ast.KindForStatement:
		forStatement := node.AsForStatement()
		return converter.createNode(node, "ForStatement",
			"body", converter.convertChild(forStatement.Statement, nil),
			"init", converter.convertChild(forStatement.Initializer, nil),
			"test", converter.convertChild(forStatement.Condition, nil),
			"update", converter.convertChild(forStatement.Incrementor, nil))

	case ast.KindForInStatement:
		forIn := node.AsForInOrOfStatement()
		return converter.createNode(node, "ForInStatement",
			"body", converter.convertChild(forIn.Statement, nil),
			"left", converter.convertPattern(forIn.Initializer, nil),
			"right", converter.convertChild(forIn.Expression, nil))

	case ast.KindForOfStatement:
		forOf := node.AsForInOrOfStatement()
		return converter.createNode(node, "ForOfStatement",
			"await", forOf.AwaitModifier != nil && forOf.AwaitModifier.Kind == ast.KindAwaitKeyword,
			"body", converter.convertChild(forOf.Statement, nil),
			"left", converter.convertPattern(forOf.Initializer, nil),
			"right", converter.convertChild(forOf.Expression, nil))

	case ast.KindFunctionDeclaration:
		function := node.AsFunctionDeclaration()
		nodeType := "FunctionDeclaration"
		if function.Body == nil {
			nodeType = "TSDeclareFunction"
		}
		result := converter.createNode(node, nodeType,
			"async", hasModifier(ast.KindAsyncKeyword, node),
			"body", converter.convertChild(function.Body, nil),
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"expression", false,
			"generator", function.AsteriskToken != nil,
			"id", converter.convertChild(node.Name(), nil),
			"params", converter.convertParameters(converter.parameterList(node)),
			"returnType", converter.returnType(node),
			"typeParameters", converter.convertTypeParameters(node))
		return converter.fixExports(node, result)

	case ast.KindVariableDeclaration:
		declaration := node.AsVariableDeclaration()
		init := converter.convertChild(declaration.Initializer, nil)
		id := converter.convertBindingNameWithTypeAnnotation(node.Name(), declaration.Type, node)
		return converter.createNode(node, "VariableDeclarator",
			"definite", declaration.ExclamationToken != nil,
			"id", id,
			"init", init)

	case ast.KindVariableStatement:
		declarationList := node.AsVariableStatement().DeclarationList
		result := converter.createNode(node, "VariableDeclaration",
			"declarations", converter.convertChildren(declarationList.AsVariableDeclarationList().Declarations.Nodes, nil),
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"kind", getDeclarationKind(declarationList))
		return converter.fixExports(node, result)

	case ast.KindVariableDeclarationList:
		return converter.createNode(node, "VariableDeclaration",
			"declarations", converter.convertChildren(node.AsVariableDeclarationList().Declarations.Nodes, nil),
			"declare", false,
			"kind", getDeclarationKind(node))

	case ast.KindExpressionStatement:
		return converter.createNode(node, "ExpressionStatement",
			"directive", nil,
			"expression", converter.convertChild(node.Expression(), nil))

	case ast.KindThisKeyword:
		return converter.createNode(node, "ThisExpression")

	case ast.KindArrayLiteralExpression:
		elements := node.AsArrayLiteralExpression().Elements.Nodes
		if converter.allowPattern {
			patterns := make([]*Node, len(elements))
			for index, element := range elements {
				patterns[index] = converter.convertPattern(element, nil)
			}
			return converter.createNode(node, "ArrayPattern",
				"decorators", noNodes,
				"elements", patterns,
				"optional", false,
				"typeAnnotation", nil)
		}
		return converter.createNode(node, "ArrayExpression",
			"elements", converter.convertChildren(elements, nil))

	case ast.KindObjectLiteralExpression:
		properties := node.AsObjectLiteralExpression().Properties.Nodes
		if converter.allowPattern {
			patterns := make([]*Node, len(properties))
			for index, property := range properties {
				patterns[index] = converter.convertPattern(property, nil)
			}
			return converter.createNode(node, "ObjectPattern",
				"decorators", noNodes,
				"optional", false,
				"properties", patterns,
				"typeAnnotation", nil)
		}
		return converter.createNode(node, "ObjectExpression",
			"properties", converter.convertChildren(properties, nil))

	case ast.KindPropertyAssignment:
		assignment := node.AsPropertyAssignment()
		return converter.createNode(node, "Property",
			"computed", isComputedProperty(node.Name()),
			"key", converter.convertChild(node.Name(), nil),
			"kind", "init",
			"method", false,
			"optional", false,
			"shorthand", false,
			"value", converter.converter(assignment.Initializer, node, converter.allowPattern))

	case ast.KindShorthandPropertyAssignment:
		shorthand := node.AsShorthandPropertyAssignment()
		if shorthand.ObjectAssignmentInitializer != nil {
			return converter.createNode(node, "Property",
				"computed", false,
				"key", converter.convertChild(node.Name(), nil),
				"kind", "init",
				"method", false,
				"optional", false,
				"shorthand", true,
				"value", converter.createNode(node, "AssignmentPattern",
					"decorators", noNodes,
					"left", converter.convertPattern(node.Name(), nil),
					"optional", false,
					"right", converter.convertChild(shorthand.ObjectAssignmentInitializer, nil),
					"typeAnnotation", nil))
		}
		return converter.createNode(node, "Property",
			"computed", false,
			"key", converter.convertChild(node.Name(), nil),
			"kind", "init",
			"method", false,
			"optional", false,
			"shorthand", true,
			"value", converter.convertChild(node.Name(), nil))

	case ast.KindComputedPropertyName:
		return converter.convertChild(node.Expression(), nil)

	case ast.KindPropertyDeclaration:
		property := node.AsPropertyDeclaration()
		isAbstract := hasModifier(ast.KindAbstractKeyword, node)
		isAccessor := hasModifier(ast.KindAccessorKeyword, node)
		nodeType := "PropertyDefinition"
		switch {
		case isAccessor && isAbstract:
			nodeType = "TSAbstractAccessorProperty"
		case isAccessor:
			nodeType = "AccessorProperty"
		case isAbstract:
			nodeType = "TSAbstractPropertyDefinition"
		}
		key := converter.convertChild(node.Name(), nil)
		nameKind := node.Name().Kind
		optional := (key.Is("Literal") || nameKind == ast.KindIdentifier || nameKind == ast.KindComputedPropertyName ||
			nameKind == ast.KindPrivateIdentifier) && isOptional(node)
		var value *Node
		if !isAbstract {
			value = converter.convertChild(property.Initializer, nil)
		}
		var typeAnnotation *Node
		if property.Type != nil {
			typeAnnotation = converter.convertTypeAnnotation(property.Type, node)
		}
		return converter.createNode(node, nodeType,
			"accessibility", getTSNodeAccessibility(node),
			"computed", isComputedProperty(node.Name()),
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"decorators", converter.convertChildren(decorators(node), nil),
			"definite", property.PostfixToken != nil && property.PostfixToken.Kind == ast.KindExclamationToken,
			"key", key,
			"optional", optional,
			"override", hasModifier(ast.KindOverrideKeyword, node),
			"readonly", hasModifier(ast.KindReadonlyKeyword, node),
			"static", hasModifier(ast.KindStaticKeyword, node),
			"typeAnnotation", typeAnnotation,
			"value", value)

	case ast.KindGetAccessor, ast.KindSetAccessor, ast.KindMethodDeclaration:
		if (node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor) &&
			(node.Parent.Kind == ast.KindInterfaceDeclaration || node.Parent.Kind == ast.KindTypeLiteral) {
			return converter.convertMethodSignature(node)
		}
		return converter.convertMethod(node, parent)

	case ast.KindConstructor:
		return converter.convertConstructor(node)

	case ast.KindFunctionExpression:
		function := node.AsFunctionExpression()
		return converter.createNode(node, "FunctionExpression",
			"async", hasModifier(ast.KindAsyncKeyword, node),
			"body", converter.convertChild(function.Body, nil),
			"declare", false,
			"expression", false,
			"generator", function.AsteriskToken != nil,
			"id", converter.convertChild(node.Name(), nil),
			"params", converter.convertParameters(converter.parameterList(node)),
			"returnType", converter.returnType(node),
			"typeParameters", converter.convertTypeParameters(node))

	case ast.KindSuperKeyword:
		return converter.createNode(node, "Super")

	case ast.KindArrayBindingPattern:
		elements := node.AsBindingPattern().Elements.Nodes
		patterns := make([]*Node, len(elements))
		for index, element := range elements {
			patterns[index] = converter.convertPattern(element, nil)
		}
		return converter.createNode(node, "ArrayPattern",
			"decorators", noNodes,
			"elements", patterns,
			"optional", false,
			"typeAnnotation", nil)

	case ast.KindOmittedExpression:
		return nil

	case ast.KindObjectBindingPattern:
		elements := node.AsBindingPattern().Elements.Nodes
		patterns := make([]*Node, len(elements))
		for index, element := range elements {
			patterns[index] = converter.convertPattern(element, nil)
		}
		return converter.createNode(node, "ObjectPattern",
			"decorators", noNodes,
			"optional", false,
			"properties", patterns,
			"typeAnnotation", nil)

	case ast.KindBindingElement:
		return converter.convertBindingElement(node, parent)

	case ast.KindArrowFunction:
		arrow := node.AsArrowFunction()
		return converter.createNode(node, "ArrowFunctionExpression",
			"async", hasModifier(ast.KindAsyncKeyword, node),
			"body", converter.convertChild(arrow.Body, nil),
			"expression", arrow.Body.Kind != ast.KindBlock,
			"generator", false,
			"id", nil,
			"params", converter.convertParameters(converter.parameterList(node)),
			"returnType", converter.returnType(node),
			"typeParameters", converter.convertTypeParameters(node))

	case ast.KindYieldExpression:
		yield := node.AsYieldExpression()
		return converter.createNode(node, "YieldExpression",
			"argument", converter.convertChild(yield.Expression, nil),
			"delegate", yield.AsteriskToken != nil)

	case ast.KindAwaitExpression:
		return converter.createNode(node, "AwaitExpression",
			"argument", converter.convertChild(node.Expression(), nil))

	case ast.KindNoSubstitutionTemplateLiteral:
		rawText := converter.text[converter.getStart(node)+1 : node.End()-1]
		cooked := node.Text()
		cookedValue := &cooked
		if node.Parent.Kind == ast.KindTaggedTemplateExpression && !isValidEscape(rawText) {
			cookedValue = nil
		}
		return converter.createNode(node, "TemplateLiteral",
			"expressions", noNodes,
			"quasis", []*Node{converter.createNode(node, "TemplateElement",
				"tail", true,
				"value", &TemplateValue{Cooked: cookedValue, Raw: rawText})})

	case ast.KindTemplateExpression:
		template := node.AsTemplateExpression()
		expressions := []*Node{}
		quasis := []*Node{converter.convertChild(template.Head, nil)}
		for _, span := range template.TemplateSpans.Nodes {
			templateSpan := span.AsTemplateSpan()
			expressions = append(expressions, converter.convertChild(templateSpan.Expression, nil))
			quasis = append(quasis, converter.convertChild(templateSpan.Literal, nil))
		}
		return converter.createNode(node, "TemplateLiteral",
			"expressions", expressions,
			"quasis", quasis)

	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		return converter.createNode(node, "TaggedTemplateExpression",
			"quasi", converter.convertChild(tagged.Template, nil),
			"tag", converter.convertChild(tagged.Tag, nil),
			"typeArguments", converter.convertTypeArguments(node))

	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		tail := node.Kind == ast.KindTemplateTail
		endOffset := 2
		if tail {
			endOffset = 1
		}
		rawText := converter.text[converter.getStart(node)+1 : node.End()-endOffset]
		var isTagged bool
		if node.Kind == ast.KindTemplateHead {
			isTagged = node.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
		} else {
			isTagged = node.Parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
		}
		cooked := node.Text()
		cookedValue := &cooked
		if isTagged && !isValidEscape(rawText) {
			cookedValue = nil
		}
		return converter.createNode(node, "TemplateElement",
			"tail", tail,
			"value", &TemplateValue{Cooked: cookedValue, Raw: rawText})

	case ast.KindSpreadAssignment, ast.KindSpreadElement:
		if converter.allowPattern {
			return converter.createNode(node, "RestElement",
				"argument", converter.convertPattern(node.Expression(), nil),
				"decorators", noNodes,
				"optional", false,
				"typeAnnotation", nil,
				"value", nil)
		}
		return converter.createNode(node, "SpreadElement",
			"argument", converter.convertChild(node.Expression(), nil))

	case ast.KindParameter:
		return converter.convertParameter(node, parent)

	case ast.KindClassDeclaration, ast.KindClassExpression:
		return converter.convertClass(node)

	case ast.KindModuleBlock:
		return converter.createNode(node, "TSModuleBlock",
			"body", converter.convertBodyExpressions(node.Statements(), node))

	case ast.KindImportDeclaration:
		return converter.convertImportDeclaration(node)

	case ast.KindNamespaceImport:
		return converter.createNode(node, "ImportNamespaceSpecifier",
			"local", converter.convertChild(node.Name(), nil))

	case ast.KindImportSpecifier:
		specifier := node.AsImportSpecifier()
		imported := specifier.PropertyName
		if imported == nil {
			imported = node.Name()
		}
		importKind := "value"
		if specifier.IsTypeOnly {
			importKind = "type"
		}
		return converter.createNode(node, "ImportSpecifier",
			"imported", converter.convertChild(imported, nil),
			"importKind", importKind,
			"local", converter.convertChild(node.Name(), nil))

	case ast.KindImportClause:
		local := converter.convertChild(node.Name(), nil)
		return converter.createNodeWithRange("ImportDefaultSpecifier", local.Range, "local", local)

	case ast.KindExportDeclaration:
		export := node.AsExportDeclaration()
		exportKind := "value"
		if export.IsTypeOnly {
			exportKind = "type"
		}
		if export.ExportClause != nil && export.ExportClause.Kind == ast.KindNamedExports {
			return converter.createNode(node, "ExportNamedDeclaration",
				"attributes", converter.convertImportAttributes(export.Attributes),
				"declaration", nil,
				"exportKind", exportKind,
				"source", converter.convertChild(export.ModuleSpecifier, nil),
				"specifiers", converter.convertChildren(export.ExportClause.AsNamedExports().Elements.Nodes, node))
		}
		var exported *Node
		if export.ExportClause != nil && export.ExportClause.Kind == ast.KindNamespaceExport {
			exported = converter.convertChild(export.ExportClause.Name(), nil)
		}
		return converter.createNode(node, "ExportAllDeclaration",
			"attributes", converter.convertImportAttributes(export.Attributes),
			"exported", exported,
			"exportKind", exportKind,
			"source", converter.convertChild(export.ModuleSpecifier, nil))

	case ast.KindExportSpecifier:
		specifier := node.AsExportSpecifier()
		local := specifier.PropertyName
		if local == nil {
			local = node.Name()
		}
		exportKind := "value"
		if specifier.IsTypeOnly {
			exportKind = "type"
		}
		return converter.createNode(node, "ExportSpecifier",
			"exported", converter.convertChild(node.Name(), nil),
			"exportKind", exportKind,
			"local", converter.convertChild(local, nil))

	case ast.KindExportAssignment:
		assignment := node.AsExportAssignment()
		if assignment.IsExportEquals {
			return converter.createNode(node, "TSExportAssignment",
				"expression", converter.convertChild(assignment.Expression, nil))
		}
		return converter.createNode(node, "ExportDefaultDeclaration",
			"declaration", converter.convertChild(assignment.Expression, nil),
			"exportKind", "value")

	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		var operatorKind ast.Kind
		var operand *ast.Node
		if node.Kind == ast.KindPrefixUnaryExpression {
			operatorKind, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
		} else {
			operatorKind, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
		}
		operator := scanner.TokenToString(operatorKind)
		if operator == "++" || operator == "--" {
			return converter.createNode(node, "UpdateExpression",
				"argument", converter.convertChild(operand, nil),
				"operator", operator,
				"prefix", node.Kind == ast.KindPrefixUnaryExpression)
		}
		return converter.createNode(node, "UnaryExpression",
			"argument", converter.convertChild(operand, nil),
			"operator", operator,
			"prefix", node.Kind == ast.KindPrefixUnaryExpression)

	case ast.KindDeleteExpression:
		return converter.unary(node, "delete")
	case ast.KindVoidExpression:
		return converter.unary(node, "void")
	case ast.KindTypeOfExpression:
		return converter.unary(node, "typeof")

	case ast.KindTypeOperator:
		operator := node.AsTypeOperatorNode()
		return converter.createNode(node, "TSTypeOperator",
			"operator", scanner.TokenToString(operator.Operator),
			"typeAnnotation", converter.convertChild(operator.Type, nil))

	case ast.KindBinaryExpression:
		return converter.convertBinary(node)

	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		object := converter.convertChild(access.Expression, nil)
		property := converter.convertChild(access.Name(), nil)
		result := converter.createNode(node, "MemberExpression",
			"computed", false,
			"object", object,
			"optional", access.QuestionDotToken != nil,
			"property", property)
		return converter.convertChainExpression(result, node)

	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		object := converter.convertChild(access.Expression, nil)
		property := converter.convertChild(access.ArgumentExpression, nil)
		result := converter.createNode(node, "MemberExpression",
			"computed", true,
			"object", object,
			"optional", access.QuestionDotToken != nil,
			"property", property)
		return converter.convertChainExpression(result, node)

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		arguments := call.Arguments.Nodes
		if call.Expression.Kind == ast.KindImportKeyword {
			var options *Node
			if len(arguments) > 1 {
				options = converter.convertChild(arguments[1], nil)
			}
			var source *Node
			if len(arguments) > 0 {
				source = converter.convertChild(arguments[0], nil)
			}
			return converter.createNode(node, "ImportExpression",
				"options", options,
				"source", source)
		}
		callee := converter.convertChild(call.Expression, nil)
		convertedArguments := converter.convertChildren(arguments, nil)
		typeArguments := converter.convertTypeArguments(node)
		result := converter.createNode(node, "CallExpression",
			"arguments", convertedArguments,
			"callee", callee,
			"optional", call.QuestionDotToken != nil,
			"typeArguments", typeArguments)
		return converter.convertChainExpression(result, node)

	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		typeArguments := converter.convertTypeArguments(node)
		var arguments []*ast.Node
		if newExpression.Arguments != nil {
			arguments = newExpression.Arguments.Nodes
		}
		return converter.createNode(node, "NewExpression",
			"arguments", converter.convertChildren(arguments, nil),
			"callee", converter.convertChild(newExpression.Expression, nil),
			"typeArguments", typeArguments)

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return converter.createNode(node, "ConditionalExpression",
			"alternate", converter.convertChild(conditional.WhenFalse, nil),
			"consequent", converter.convertChild(conditional.WhenTrue, nil),
			"test", converter.convertChild(conditional.Condition, nil))

	case ast.KindMetaProperty:
		meta := node.AsMetaProperty()
		first := converter.firstToken(node)
		return converter.createNode(node, "MetaProperty",
			"meta", converter.identifierNode([2]int{first.start, first.end}, scanner.TokenToString(meta.KeywordToken)),
			"property", converter.convertChild(meta.Name(), nil))

	case ast.KindDecorator:
		return converter.createNode(node, "Decorator",
			"expression", converter.convertChild(node.Expression(), nil))

	case ast.KindStringLiteral:
		value := node.Text()
		if parent.Kind == ast.KindJsxAttribute {
			value = ecmascripttext.UnescapeStringLiteralText(value)
		}
		return converter.createNode(node, "Literal",
			"raw", converter.getText(node),
			"value", value)

	case ast.KindNumericLiteral:
		raw := converter.getText(node)
		return converter.createNode(node, "Literal",
			"raw", raw,
			"value", numberValue(raw))

	case ast.KindBigIntLiteral:
		nodeRange := converter.getRange(node)
		rawValue := converter.text[nodeRange[0]:nodeRange[1]]
		bigint := strings.ReplaceAll(rawValue[:len(rawValue)-1], "_", "")
		return converter.createNodeWithRange("Literal", nodeRange,
			"bigint", bigintString(bigint),
			"raw", rawValue,
			"value", nil)

	case ast.KindRegularExpressionLiteral:
		text := node.Text()
		slash := strings.LastIndex(text, "/")
		return converter.createNode(node, "Literal",
			"raw", text,
			"regex", &Regex{Pattern: text[1:slash], Flags: text[slash+1:]},
			"value", nil)

	case ast.KindTrueKeyword:
		return converter.createNode(node, "Literal", "raw", "true", "value", true)
	case ast.KindFalseKeyword:
		return converter.createNode(node, "Literal", "raw", "false", "value", false)
	case ast.KindNullKeyword:
		return converter.createNode(node, "Literal", "raw", "null", "value", nil)

	case ast.KindEmptyStatement:
		return converter.createNode(node, "EmptyStatement")
	case ast.KindDebuggerStatement:
		return converter.createNode(node, "DebuggerStatement")

	case ast.KindJsxElement:
		element := node.AsJsxElement()
		return converter.createNode(node, "JSXElement",
			"children", converter.convertChildren(element.Children.Nodes, nil),
			"closingElement", converter.convertChild(element.ClosingElement, nil),
			"openingElement", converter.convertChild(element.OpeningElement, nil))

	case ast.KindJsxFragment:
		fragment := node.AsJsxFragment()
		return converter.createNode(node, "JSXFragment",
			"children", converter.convertChildren(fragment.Children.Nodes, nil),
			"closingFragment", converter.convertChild(fragment.ClosingFragment, nil),
			"openingFragment", converter.convertChild(fragment.OpeningFragment, nil))

	case ast.KindJsxSelfClosingElement:
		element := node.AsJsxSelfClosingElement()
		return converter.createNode(node, "JSXElement",
			"children", noNodes,
			"closingElement", nil,
			"openingElement", converter.createNodeWithRange("JSXOpeningElement", converter.getRange(node),
				"attributes", converter.convertChildren(element.Attributes.AsJsxAttributes().Properties.Nodes, nil),
				"name", converter.convertJSXTagName(element.TagName, node),
				"selfClosing", true,
				"typeArguments", converter.convertTypeArguments(node)))

	case ast.KindJsxOpeningElement:
		element := node.AsJsxOpeningElement()
		return converter.createNode(node, "JSXOpeningElement",
			"attributes", converter.convertChildren(element.Attributes.AsJsxAttributes().Properties.Nodes, nil),
			"name", converter.convertJSXTagName(element.TagName, node),
			"selfClosing", false,
			"typeArguments", converter.convertTypeArguments(node))

	case ast.KindJsxClosingElement:
		return converter.createNode(node, "JSXClosingElement",
			"name", converter.convertJSXTagName(node.AsJsxClosingElement().TagName, node))

	case ast.KindJsxOpeningFragment:
		return converter.createNode(node, "JSXOpeningFragment")
	case ast.KindJsxClosingFragment:
		return converter.createNode(node, "JSXClosingFragment")

	case ast.KindJsxExpression:
		jsxExpression := node.AsJsxExpression()
		var expression *Node
		if jsxExpression.Expression != nil {
			expression = converter.convertChild(jsxExpression.Expression, nil)
		} else {
			expression = converter.createNodeWithRange("JSXEmptyExpression", [2]int{converter.getStart(node) + 1, node.End() - 1})
		}
		if jsxExpression.DotDotDotToken != nil {
			return converter.createNode(node, "JSXSpreadChild", "expression", expression)
		}
		return converter.createNode(node, "JSXExpressionContainer", "expression", expression)

	case ast.KindJsxAttribute:
		attribute := node.AsJsxAttribute()
		return converter.createNode(node, "JSXAttribute",
			"name", converter.convertJSXNamespaceOrIdentifier(attribute.Name()),
			"value", converter.convertChild(attribute.Initializer, nil))

	case ast.KindJsxText:
		start, end := node.Pos(), node.End()
		text := converter.text[start:end]
		return converter.createNodeWithRange("JSXText", [2]int{start, end},
			"raw", text,
			"value", ecmascripttext.UnescapeStringLiteralText(text))

	case ast.KindJsxSpreadAttribute:
		return converter.createNode(node, "JSXSpreadAttribute",
			"argument", converter.convertChild(node.Expression(), nil))

	case ast.KindQualifiedName:
		qualified := node.AsQualifiedName()
		return converter.createNode(node, "TSQualifiedName",
			"left", converter.convertChild(qualified.Left, nil),
			"right", converter.convertChild(qualified.Right, nil))

	case ast.KindTypeReference:
		// typescript-go parses an interface's extends and a class's implements as type references,
		// converting `A.B` to a qualified name, where TypeScript keeps ExpressionWithTypeArguments.
		// Upstream's ExpressionWithTypeArguments case is what these must become.
		if parent.Kind == ast.KindInterfaceDeclaration || parent.Kind == ast.KindHeritageClause {
			nodeType := "TSClassImplements"
			if parent.Kind == ast.KindInterfaceDeclaration {
				nodeType = "TSInterfaceHeritage"
			}
			return converter.createNode(node, nodeType,
				"expression", converter.convertEntityNameAsExpression(node.AsTypeReferenceNode().TypeName),
				"typeArguments", converter.convertTypeArguments(node))
		}
		return converter.createNode(node, "TSTypeReference",
			"typeArguments", converter.convertTypeArguments(node),
			"typeName", converter.convertChild(node.AsTypeReferenceNode().TypeName, nil))

	case ast.KindTypeParameter:
		parameter := node.AsTypeParameterDeclaration()
		var defaultType any
		if parameter.DefaultType != nil {
			defaultType = converter.convertChild(parameter.DefaultType, nil)
		}
		var constraint any
		if parameter.Constraint != nil {
			constraint = converter.convertChild(parameter.Constraint, nil)
		}
		return converter.createNode(node, "TSTypeParameter",
			"const", hasModifier(ast.KindConstKeyword, node),
			"constraint", constraint,
			"default", defaultType,
			"in", hasModifier(ast.KindInKeyword, node),
			"name", converter.convertChild(node.Name(), nil),
			"out", hasModifier(ast.KindOutKeyword, node))

	case ast.KindThisType:
		return converter.createNode(node, "TSThisType")

	case ast.KindAnyKeyword, ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword,
		ast.KindNumberKeyword, ast.KindObjectKeyword, ast.KindStringKeyword, ast.KindSymbolKeyword,
		ast.KindUnknownKeyword, ast.KindVoidKeyword, ast.KindUndefinedKeyword, ast.KindIntrinsicKeyword:
		return converter.createNode(node, keywordTypeNames[node.Kind])

	case ast.KindNonNullExpression:
		nonNull := converter.createNode(node, "TSNonNullExpression",
			"expression", converter.convertChild(node.Expression(), nil))
		return converter.convertChainExpression(nonNull, node)

	case ast.KindTypeLiteral:
		return converter.createNode(node, "TSTypeLiteral",
			"members", converter.convertChildren(node.Members(), nil))

	case ast.KindArrayType:
		return converter.createNode(node, "TSArrayType",
			"elementType", converter.convertChild(node.AsArrayTypeNode().ElementType, nil))

	case ast.KindIndexedAccessType:
		indexed := node.AsIndexedAccessTypeNode()
		return converter.createNode(node, "TSIndexedAccessType",
			"indexType", converter.convertChild(indexed.IndexType, nil),
			"objectType", converter.convertChild(indexed.ObjectType, nil))

	case ast.KindConditionalType:
		conditional := node.AsConditionalTypeNode()
		return converter.createNode(node, "TSConditionalType",
			"checkType", converter.convertChild(conditional.CheckType, nil),
			"extendsType", converter.convertChild(conditional.ExtendsType, nil),
			"falseType", converter.convertChild(conditional.FalseType, nil),
			"trueType", converter.convertChild(conditional.TrueType, nil))

	case ast.KindTypeQuery:
		return converter.createNode(node, "TSTypeQuery",
			"exprName", converter.convertChild(node.AsTypeQueryNode().ExprName, nil),
			"typeArguments", converter.convertTypeArguments(node))

	case ast.KindMappedType:
		mapped := node.AsMappedTypeNode()
		typeParameter := mapped.TypeParameter.AsTypeParameterDeclaration()
		var optional any = false
		if mapped.QuestionToken != nil {
			if mapped.QuestionToken.Kind == ast.KindQuestionToken {
				optional = true
			} else {
				optional = scanner.TokenToString(mapped.QuestionToken.Kind)
			}
		}
		var readonly any
		if mapped.ReadonlyToken != nil {
			if mapped.ReadonlyToken.Kind == ast.KindReadonlyKeyword {
				readonly = true
			} else {
				readonly = scanner.TokenToString(mapped.ReadonlyToken.Kind)
			}
		}
		var typeAnnotation *Node
		if mapped.Type != nil {
			typeAnnotation = converter.convertChild(mapped.Type, nil)
		}
		return converter.createNode(node, "TSMappedType",
			"constraint", converter.convertChild(typeParameter.Constraint, nil),
			"key", converter.convertChild(mapped.TypeParameter.Name(), nil),
			"nameType", converter.convertChild(mapped.NameType, nil),
			"optional", optional,
			"readonly", readonly,
			"typeAnnotation", typeAnnotation)

	case ast.KindParenthesizedExpression:
		// Under babel, Prettier keeps Babel's ParenthesizedExpression after a Closure-style type cast
		// comment (parse/postprocess/index.js); typescript-estree never makes one.
		if start := converter.getStart(node); converter.followsTypeCastComment(start) {
			return converter.createNodeWithRange("ParenthesizedExpression", converter.getRange(node),
				"expression", converter.convertChild(node.Expression(), parent))
		}
		return converter.convertChild(node.Expression(), parent)

	case ast.KindTypeAliasDeclaration:
		alias := node.AsTypeAliasDeclaration()
		result := converter.createNode(node, "TSTypeAliasDeclaration",
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"id", converter.convertChild(node.Name(), nil),
			"typeAnnotation", converter.convertChild(alias.Type, nil),
			"typeParameters", converter.convertTypeParameters(node))
		return converter.fixExports(node, result)

	case ast.KindMethodSignature:
		return converter.convertMethodSignature(node)

	case ast.KindPropertySignature:
		signature := node.AsPropertySignatureDeclaration()
		var typeAnnotation *Node
		if signature.Type != nil {
			typeAnnotation = converter.convertTypeAnnotation(signature.Type, node)
		}
		return converter.createNode(node, "TSPropertySignature",
			"accessibility", getTSNodeAccessibility(node),
			"computed", isComputedProperty(node.Name()),
			"key", converter.convertChild(node.Name(), nil),
			"optional", isOptional(node),
			"readonly", hasModifier(ast.KindReadonlyKeyword, node),
			"static", hasModifier(ast.KindStaticKeyword, node),
			"typeAnnotation", typeAnnotation)

	case ast.KindIndexSignature:
		return converter.createNode(node, "TSIndexSignature",
			"accessibility", getTSNodeAccessibility(node),
			"parameters", converter.convertChildren(converter.parameterList(node), nil),
			"readonly", hasModifier(ast.KindReadonlyKeyword, node),
			"static", hasModifier(ast.KindStaticKeyword, node),
			"typeAnnotation", converter.returnType(node))

	case ast.KindConstructorType:
		return converter.createNode(node, "TSConstructorType",
			"abstract", hasModifier(ast.KindAbstractKeyword, node),
			"params", converter.convertParameters(converter.parameterList(node)),
			"returnType", converter.returnType(node),
			"typeParameters", converter.convertTypeParameters(node))

	case ast.KindFunctionType, ast.KindConstructSignature, ast.KindCallSignature:
		nodeType := "TSFunctionType"
		switch node.Kind {
		case ast.KindConstructSignature:
			nodeType = "TSConstructSignatureDeclaration"
		case ast.KindCallSignature:
			nodeType = "TSCallSignatureDeclaration"
		}
		return converter.createNode(node, nodeType,
			"params", converter.convertParameters(converter.parameterList(node)),
			"returnType", converter.returnType(node),
			"typeParameters", converter.convertTypeParameters(node))

	case ast.KindExpressionWithTypeArguments:
		nodeType := "TSInstantiationExpression"
		switch parent.Kind {
		case ast.KindInterfaceDeclaration:
			nodeType = "TSInterfaceHeritage"
		case ast.KindHeritageClause:
			nodeType = "TSClassImplements"
		}
		return converter.createNode(node, nodeType,
			"expression", converter.convertChild(node.Expression(), nil),
			"typeArguments", converter.convertTypeArguments(node))

	case ast.KindInterfaceDeclaration:
		interfaceDeclaration := node.AsInterfaceDeclaration()
		interfaceExtends := []*Node{}
		if interfaceDeclaration.HeritageClauses != nil {
			for _, clause := range interfaceDeclaration.HeritageClauses.Nodes {
				heritage := clause.AsHeritageClause()
				if heritage.Token != ast.KindExtendsKeyword {
					continue
				}
				for _, heritageType := range heritage.Types.Nodes {
					interfaceExtends = append(interfaceExtends, converter.convertChild(heritageType, node))
				}
			}
		}
		members := interfaceDeclaration.Members
		result := converter.createNode(node, "TSInterfaceDeclaration",
			"body", converter.createNodeWithRange("TSInterfaceBody", [2]int{members.Pos() - 1, node.End()},
				"body", converter.convertChildren(members.Nodes, nil)),
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"extends", interfaceExtends,
			"id", converter.convertChild(node.Name(), nil),
			"typeParameters", converter.convertTypeParameters(node))
		return converter.fixExports(node, result)

	case ast.KindTypePredicate:
		predicate := node.AsTypePredicateNode()
		result := converter.createNode(node, "TSTypePredicate",
			"asserts", predicate.AssertsModifier != nil,
			"parameterName", converter.convertChild(predicate.ParameterName, nil),
			"typeAnnotation", nil)
		if predicate.Type != nil {
			annotation := converter.convertTypeAnnotation(predicate.Type, node)
			annotation.Range = annotation.Child("typeAnnotation").Range
			result.Set("typeAnnotation", annotation)
		}
		return result

	case ast.KindImportType:
		return converter.convertImportType(node)

	case ast.KindEnumDeclaration:
		enum := node.AsEnumDeclaration()
		members := converter.convertChildren(enum.Members.Nodes, nil)
		result := converter.createNode(node, "TSEnumDeclaration",
			"body", converter.createNodeWithRange("TSEnumBody", [2]int{enum.Members.Pos() - 1, node.End()}, "members", members),
			"const", hasModifier(ast.KindConstKeyword, node),
			"declare", hasModifier(ast.KindDeclareKeyword, node),
			"id", converter.convertChild(node.Name(), nil))
		return converter.fixExports(node, result)

	case ast.KindEnumMember:
		member := node.AsEnumMember()
		var initializer *Node
		if member.Initializer != nil {
			initializer = converter.convertChild(member.Initializer, nil)
		}
		return converter.createNode(node, "TSEnumMember",
			"id", converter.convertChild(node.Name(), nil),
			"initializer", initializer)

	case ast.KindModuleDeclaration:
		return converter.convertModuleDeclaration(node)

	case ast.KindParenthesizedType:
		return converter.convertChild(node.AsParenthesizedTypeNode().Type, nil)

	case ast.KindUnionType:
		return converter.createNode(node, "TSUnionType",
			"types", converter.convertChildren(node.AsUnionTypeNode().Types.Nodes, nil))

	case ast.KindIntersectionType:
		return converter.createNode(node, "TSIntersectionType",
			"types", converter.convertChildren(node.AsIntersectionTypeNode().Types.Nodes, nil))

	case ast.KindAsExpression:
		asExpression := node.AsAsExpression()
		return converter.createNode(node, "TSAsExpression",
			"expression", converter.convertChild(asExpression.Expression, nil),
			"typeAnnotation", converter.convertChild(asExpression.Type, nil))

	case ast.KindInferType:
		return converter.createNode(node, "TSInferType",
			"typeParameter", converter.convertChild(node.AsInferTypeNode().TypeParameter, nil))

	case ast.KindLiteralType:
		literal := node.AsLiteralTypeNode().Literal
		if literal.Kind == ast.KindNullKeyword {
			return converter.createNode(literal, "TSNullKeyword")
		}
		return converter.createNode(node, "TSLiteralType",
			"literal", converter.convertChild(literal, nil))

	case ast.KindTypeAssertionExpression:
		assertion := node.AsTypeAssertion()
		return converter.createNode(node, "TSTypeAssertion",
			"expression", converter.convertChild(assertion.Expression, nil),
			"typeAnnotation", converter.convertChild(assertion.Type, nil))

	case ast.KindImportEqualsDeclaration:
		importEquals := node.AsImportEqualsDeclaration()
		importKind := "value"
		if importEquals.IsTypeOnly {
			importKind = "type"
		}
		return converter.fixExports(node, converter.createNode(node, "TSImportEqualsDeclaration",
			"id", converter.convertChild(node.Name(), nil),
			"importKind", importKind,
			"moduleReference", converter.convertChild(importEquals.ModuleReference, nil)))

	case ast.KindExternalModuleReference:
		return converter.createNode(node, "TSExternalModuleReference",
			"expression", converter.convertChild(node.AsExternalModuleReference().Expression, nil))

	case ast.KindNamespaceExportDeclaration:
		return converter.createNode(node, "TSNamespaceExportDeclaration",
			"id", converter.convertChild(node.Name(), nil))

	case ast.KindAbstractKeyword:
		return converter.createNode(node, "TSAbstractKeyword")

	case ast.KindTupleType:
		return converter.createNode(node, "TSTupleType",
			"elementTypes", converter.convertChildren(node.AsTupleTypeNode().Elements.Nodes, nil))

	case ast.KindNamedTupleMember:
		named := node.AsNamedTupleMember()
		member := converter.createNode(node, "TSNamedTupleMember",
			"elementType", converter.convertChild(named.Type, node),
			"label", converter.convertChild(node.Name(), node),
			"optional", named.QuestionToken != nil)
		if named.DotDotDotToken != nil {
			member.Range[0] = member.Child("label").Range[0]
			return converter.createNode(node, "TSRestType", "typeAnnotation", member)
		}
		return member

	case ast.KindOptionalType:
		return converter.createNode(node, "TSOptionalType",
			"typeAnnotation", converter.convertChild(node.AsOptionalTypeNode().Type, nil))

	case ast.KindRestType:
		return converter.createNode(node, "TSRestType",
			"typeAnnotation", converter.convertChild(node.AsRestTypeNode().Type, nil))

	case ast.KindTemplateLiteralType:
		template := node.AsTemplateLiteralTypeNode()
		types := []*Node{}
		quasis := []*Node{converter.convertChild(template.Head, nil)}
		for _, span := range template.TemplateSpans.Nodes {
			typeSpan := span.AsTemplateLiteralTypeSpan()
			types = append(types, converter.convertChild(typeSpan.Type, nil))
			quasis = append(quasis, converter.convertChild(typeSpan.Literal, nil))
		}
		return converter.createNode(node, "TSTemplateLiteralType",
			"quasis", quasis,
			"types", types)

	case ast.KindClassStaticBlockDeclaration:
		return converter.createNode(node, "StaticBlock",
			"body", converter.convertBodyExpressions(node.AsClassStaticBlockDeclaration().Body.Statements(), node))

	case ast.KindImportAttribute:
		attribute := node.AsImportAttribute()
		return converter.createNode(node, "ImportAttribute",
			"key", converter.convertChild(node.Name(), nil),
			"value", converter.convertChild(attribute.Value, nil))

	case ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		return converter.createNode(node, "TSSatisfiesExpression",
			"expression", converter.convertChild(satisfies.Expression, nil),
			"typeAnnotation", converter.convertChild(satisfies.Type, nil))
	}

	// Upstream's deeplyCopy copies the TypeScript node under `TS${kind}`. Nothing in our code reaches it,
	// and a copied TypeScript node is not something the printer can print the way upstream does, so
	// the file is refused rather than formatted wrong.
	if converter.unsupported == nil {
		converter.unsupported = fmt.Errorf("unsupported syntax %s at %d", node.Kind.String(), converter.getStart(node))
	}
	return converter.createNode(node, "TS"+strings.TrimPrefix(node.Kind.String(), "Kind"))
}

// convertEntityNameAsExpression converts an entity name the way upstream converts the property
// access it was parsed from: a qualified name becomes a MemberExpression.
func (converter *Converter) convertEntityNameAsExpression(node *ast.Node) *Node {
	if node.Kind != ast.KindQualifiedName {
		return converter.convertChild(node, nil)
	}
	qualified := node.AsQualifiedName()
	return converter.createNode(node, "MemberExpression",
		"computed", false,
		"object", converter.convertEntityNameAsExpression(qualified.Left),
		"optional", false,
		"property", converter.convertChild(qualified.Right, nil))
}

func (converter *Converter) unary(node *ast.Node, operator string) *Node {
	return converter.createNode(node, "UnaryExpression",
		"argument", converter.convertChild(node.Expression(), nil),
		"operator", operator,
		"prefix", true)
}

// convertMethod is upstream's GetAccessor, SetAccessor and MethodDeclaration case outside type members.
func (converter *Converter) convertMethod(node *ast.Node, parent *ast.Node) *Node {
	body := node.Body()
	functionType := "FunctionExpression"
	if body == nil {
		functionType = "TSEmptyBodyFunctionExpression"
	}
	var asterisk bool
	if node.Kind == ast.KindMethodDeclaration {
		asterisk = node.AsMethodDeclaration().AsteriskToken != nil
	}
	method := converter.createNodeWithRange(functionType, [2]int{node.ParameterList().Pos() - 1, node.End()},
		"async", hasModifier(ast.KindAsyncKeyword, node),
		"body", converter.convertChild(body, nil),
		"declare", false,
		"expression", false,
		"generator", asterisk,
		"id", nil,
		"params", noNodes,
		"returnType", converter.returnType(node),
		"typeParameters", converter.convertTypeParameters(node))
	if typeParameters := method.Child("typeParameters"); typeParameters != nil {
		fixParentLocation(method, typeParameters.Range)
	}

	var result *Node
	if parent.Kind == ast.KindObjectLiteralExpression {
		method.Set("params", converter.convertChildren(converter.parameterList(node), nil))
		result = converter.createNode(node, "Property",
			"computed", isComputedProperty(node.Name()),
			"key", converter.convertChild(node.Name(), nil),
			"kind", "init",
			"method", node.Kind == ast.KindMethodDeclaration,
			"optional", isOptional(node),
			"shorthand", false,
			"value", method)
	} else {
		isAbstract := hasModifier(ast.KindAbstractKeyword, node)
		method.Set("params", converter.convertParameters(converter.parameterList(node)))
		methodDefinitionType := "MethodDefinition"
		if isAbstract {
			methodDefinitionType = "TSAbstractMethodDefinition"
		}
		result = converter.createNode(node, methodDefinitionType,
			"accessibility", getTSNodeAccessibility(node),
			"computed", isComputedProperty(node.Name()),
			"decorators", converter.convertChildren(decorators(node), nil),
			"key", converter.convertChild(node.Name(), nil),
			"kind", "method",
			"optional", isOptional(node),
			"override", hasModifier(ast.KindOverrideKeyword, node),
			"static", hasModifier(ast.KindStaticKeyword, node),
			"value", method)
	}

	switch {
	case node.Kind == ast.KindGetAccessor:
		result.Set("kind", "get")
	case node.Kind == ast.KindSetAccessor:
		result.Set("kind", "set")
	case !result.Bool("static") && node.Name().Kind == ast.KindStringLiteral && node.Name().Text() == "constructor" &&
		result.Type() != "Property":
		result.Set("kind", "constructor")
	}
	return result
}

// convertConstructor is upstream's Constructor case.
func (converter *Converter) convertConstructor(node *ast.Node) *Node {
	nodeModifiers := modifiers(node)
	var constructorToken token
	if len(nodeModifiers) > 0 {
		constructorToken = converter.tokenAt(nodeModifiers[len(nodeModifiers)-1].End())
	} else {
		constructorToken = converter.firstToken(node)
	}

	body := node.Body()
	functionType := "FunctionExpression"
	if body == nil {
		functionType = "TSEmptyBodyFunctionExpression"
	}
	constructor := converter.createNodeWithRange(functionType, [2]int{node.ParameterList().Pos() - 1, node.End()},
		"async", false,
		"body", converter.convertChild(body, nil),
		"declare", false,
		"expression", false,
		"generator", false,
		"id", nil,
		"params", converter.convertParameters(converter.parameterList(node)),
		"returnType", converter.returnType(node),
		"typeParameters", converter.convertTypeParameters(node))
	if typeParameters := constructor.Child("typeParameters"); typeParameters != nil {
		fixParentLocation(constructor, typeParameters.Range)
	}

	var constructorKey *Node
	if constructorToken.kind == ast.KindStringLiteral {
		constructorKey = converter.createNodeWithRange("Literal", [2]int{constructorToken.start, constructorToken.end},
			"raw", converter.text[constructorToken.start:constructorToken.end],
			"value", "constructor")
	} else {
		constructorKey = converter.identifierNode([2]int{constructorToken.start, constructorToken.end}, "constructor")
	}

	isStatic := hasModifier(ast.KindStaticKeyword, node)
	methodType := "MethodDefinition"
	if hasModifier(ast.KindAbstractKeyword, node) {
		methodType = "TSAbstractMethodDefinition"
	}
	kind := "constructor"
	if isStatic {
		kind = "method"
	}
	return converter.createNode(node, methodType,
		"accessibility", getTSNodeAccessibility(node),
		"computed", false,
		"decorators", noNodes,
		"key", constructorKey,
		"kind", kind,
		"optional", false,
		"override", false,
		"static", isStatic,
		"value", constructor)
}

// convertBindingElement is upstream's BindingElement case.
func (converter *Converter) convertBindingElement(node *ast.Node, parent *ast.Node) *Node {
	element := node.AsBindingElement()
	if parent.Kind == ast.KindArrayBindingPattern {
		arrayItem := converter.convertChild(node.Name(), parent)
		if element.Initializer != nil {
			return converter.createNode(node, "AssignmentPattern",
				"decorators", noNodes,
				"left", arrayItem,
				"optional", false,
				"right", converter.convertChild(element.Initializer, nil),
				"typeAnnotation", nil)
		}
		if element.DotDotDotToken != nil {
			return converter.createNode(node, "RestElement",
				"argument", arrayItem,
				"decorators", noNodes,
				"optional", false,
				"typeAnnotation", nil,
				"value", nil)
		}
		return arrayItem
	}

	var result *Node
	if element.DotDotDotToken != nil {
		argument := element.PropertyName
		if argument == nil {
			argument = node.Name()
		}
		result = converter.createNode(node, "RestElement",
			"argument", converter.convertChild(argument, nil),
			"decorators", noNodes,
			"optional", false,
			"typeAnnotation", nil,
			"value", nil)
	} else {
		key := element.PropertyName
		if key == nil {
			key = node.Name()
		}
		result = converter.createNode(node, "Property",
			"computed", element.PropertyName != nil && element.PropertyName.Kind == ast.KindComputedPropertyName,
			"key", converter.convertChild(key, nil),
			"kind", "init",
			"method", false,
			"optional", false,
			"shorthand", element.PropertyName == nil,
			"value", converter.convertChild(node.Name(), nil))
	}
	if element.Initializer != nil {
		result.Set("value", converter.createNodeWithRange("AssignmentPattern", [2]int{converter.getStart(node.Name()), element.Initializer.End()},
			"decorators", noNodes,
			"left", converter.convertChild(node.Name(), nil),
			"optional", false,
			"right", converter.convertChild(element.Initializer, nil),
			"typeAnnotation", nil))
	}
	return result
}

// convertParameter is upstream's Parameter case.
func (converter *Converter) convertParameter(node *ast.Node, parent *ast.Node) *Node {
	declaration := node.AsParameterDeclaration()
	var parameter, result *Node
	switch {
	case declaration.DotDotDotToken != nil:
		parameter = converter.createNode(node, "RestElement",
			"argument", converter.convertChild(node.Name(), nil),
			"decorators", noNodes,
			"optional", false,
			"typeAnnotation", nil,
			"value", nil)
		result = parameter
	case declaration.Initializer != nil:
		parameter = converter.convertChild(node.Name(), nil)
		result = converter.createNodeWithRange("AssignmentPattern", [2]int{converter.getStart(node.Name()), declaration.Initializer.End()},
			"decorators", noNodes,
			"left", parameter,
			"optional", false,
			"right", converter.convertChild(declaration.Initializer, nil),
			"typeAnnotation", nil)
		if len(modifiers(node)) > 0 {
			// AssignmentPattern should not contain modifiers in range
			result.Range[0] = parameter.Range[0]
		}
	default:
		parameter = converter.convertChild(node.Name(), parent)
		result = parameter
	}

	if declaration.Type != nil {
		annotation := converter.convertTypeAnnotation(declaration.Type, node)
		parameter.Set("typeAnnotation", annotation)
		fixParentLocation(parameter, annotation.Range)
	}
	if declaration.QuestionToken != nil {
		if declaration.QuestionToken.End() > parameter.Range[1] {
			parameter.Range[1] = declaration.QuestionToken.End()
		}
		parameter.Set("optional", true)
	}
	if len(modifiers(node)) > 0 {
		return converter.createNode(node, "TSParameterProperty",
			"accessibility", getTSNodeAccessibility(node),
			"decorators", noNodes,
			"override", hasModifier(ast.KindOverrideKeyword, node),
			"parameter", result,
			"readonly", hasModifier(ast.KindReadonlyKeyword, node),
			"static", hasModifier(ast.KindStaticKeyword, node))
	}
	return result
}

// convertClass is upstream's ClassDeclaration and ClassExpression case.
func (converter *Converter) convertClass(node *ast.Node) *Node {
	classType := "ClassExpression"
	if node.Kind == ast.KindClassDeclaration {
		classType = "ClassDeclaration"
	}
	classLike := node.ClassLikeData()
	var extendsClause, implementsClause *ast.HeritageClause
	if classLike.HeritageClauses != nil {
		for _, clause := range classLike.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			if heritage.Token == ast.KindExtendsKeyword && extendsClause == nil {
				extendsClause = heritage
			}
			if heritage.Token == ast.KindImplementsKeyword && implementsClause == nil {
				implementsClause = heritage
			}
		}
	}

	var members []*ast.Node
	for _, member := range classLike.Members.Nodes {
		if member.Kind != ast.KindSemicolonClassElement {
			members = append(members, member)
		}
	}
	var implements []*ast.Node
	if implementsClause != nil {
		implements = implementsClause.Types.Nodes
	}
	var superClass *Node
	if extendsClause != nil && len(extendsClause.Types.Nodes) > 0 {
		superClass = converter.convertChild(extendsClause.Types.Nodes[0].Expression(), nil)
	}
	result := converter.createNode(node, classType,
		"abstract", hasModifier(ast.KindAbstractKeyword, node),
		"body", converter.createNodeWithRange("ClassBody", [2]int{classLike.Members.Pos() - 1, node.End()},
			"body", converter.convertChildren(members, nil)),
		"declare", hasModifier(ast.KindDeclareKeyword, node),
		"decorators", converter.convertChildren(decorators(node), nil),
		"id", converter.convertChild(node.Name(), nil),
		"implements", converter.convertChildren(implements, nil),
		"superClass", superClass,
		"superTypeArguments", nil,
		"typeParameters", converter.convertTypeParameters(node))
	if extendsClause != nil && len(extendsClause.Types.Nodes) > 0 && extendsClause.Types.Nodes[0].TypeArgumentList() != nil {
		result.Set("superTypeArguments", converter.convertTypeArguments(extendsClause.Types.Nodes[0]))
	}
	return converter.fixExports(node, result)
}

// convertImportDeclaration is upstream's ImportDeclaration case.
func (converter *Converter) convertImportDeclaration(node *ast.Node) *Node {
	declaration := node.AsImportDeclaration()
	phase := importClausePhase(declaration.ImportClause)
	var phaseValue any
	if phase == "defer" {
		phaseValue = "defer"
	}
	specifiers := []*Node{}
	result := converter.createNode(node, "ImportDeclaration",
		"attributes", converter.convertImportAttributes(declaration.Attributes),
		"importKind", "value",
		"phase", phaseValue,
		"source", converter.convertChild(declaration.ModuleSpecifier, nil),
		"specifiers", specifiers)
	if declaration.ImportClause != nil {
		clause := declaration.ImportClause.AsImportClause()
		if phase == "type" {
			result.Set("importKind", "type")
		}
		if declaration.ImportClause.Name() != nil {
			specifiers = append(specifiers, converter.convertChild(declaration.ImportClause, nil))
		}
		if clause.NamedBindings != nil {
			switch clause.NamedBindings.Kind {
			case ast.KindNamespaceImport:
				specifiers = append(specifiers, converter.convertChild(clause.NamedBindings, nil))
			case ast.KindNamedImports:
				specifiers = append(specifiers, converter.convertChildren(clause.NamedBindings.AsNamedImports().Elements.Nodes, nil)...)
			}
		}
		result.Set("specifiers", specifiers)
	}
	return result
}

// importClausePhase is upstream's getImportClausePhaseModifier.
func importClausePhase(clause *ast.Node) string {
	if clause == nil {
		return ""
	}
	switch clause.AsImportClause().PhaseModifier {
	case ast.KindTypeKeyword:
		return "type"
	case ast.KindDeferKeyword:
		return "defer"
	}
	return ""
}

// convertImportType is upstream's ImportType case.
func (converter *Converter) convertImportType(node *ast.Node) *Node {
	importType := node.AsImportTypeNode()
	nodeRange := converter.getRange(node)
	if importType.IsTypeOf {
		first := converter.firstToken(node)
		next := converter.tokenAt(first.end)
		nodeRange[0] = next.start
	}

	var options *Node
	if importType.Attributes != nil {
		attributes := importType.Attributes.AsImportAttributes()
		var properties []*Node
		for _, attribute := range attributes.Attributes.Nodes {
			properties = append(properties, converter.createNode(attribute, "Property",
				"computed", false,
				"key", converter.convertChild(attribute.Name(), nil),
				"kind", "init",
				"method", false,
				"optional", false,
				"shorthand", false,
				"value", converter.convertChild(attribute.AsImportAttribute().Value, nil)))
		}
		value := converter.createNode(importType.Attributes, "ObjectExpression", "properties", properties)
		comma := converter.tokenAt(importType.Argument.End())
		openBrace := converter.tokenAt(comma.end)
		afterAttributes := converter.tokenAt(importType.Attributes.End())
		closeBrace := afterAttributes
		if afterAttributes.kind == ast.KindCommaToken {
			closeBrace = converter.tokenAt(afterAttributes.end)
		}
		withOrAssert := converter.tokenAt(openBrace.end)
		withOrAssertName := "with"
		if withOrAssert.kind == ast.KindAssertKeyword {
			withOrAssertName = "assert"
		}
		options = converter.createNodeWithRange("ObjectExpression", [2]int{openBrace.start, closeBrace.end},
			"properties", []*Node{converter.createNodeWithRange("Property", [2]int{withOrAssert.start, importType.Attributes.End()},
				"computed", false,
				"key", converter.identifierNode([2]int{withOrAssert.start, withOrAssert.end}, withOrAssertName),
				"kind", "init",
				"method", false,
				"optional", false,
				"shorthand", false,
				"value", value)})
	}

	argument := converter.convertChild(importType.Argument, nil)
	typeArguments := converter.convertTypeArguments(node)
	result := converter.createNodeWithRange("TSImportType", nodeRange,
		"options", options,
		"qualifier", converter.convertChild(importType.Qualifier, nil),
		"source", argument.Child("literal"),
		"typeArguments", typeArguments)
	if importType.IsTypeOf {
		return converter.createNode(node, "TSTypeQuery",
			"exprName", result,
			"typeArguments", nil)
	}
	return result
}

// convertModuleDeclaration is upstream's ModuleDeclaration case.
func (converter *Converter) convertModuleDeclaration(node *ast.Node) *Node {
	module := node.AsModuleDeclaration()
	isDeclare := hasModifier(ast.KindDeclareKeyword, node)
	isGlobal := module.Keyword == ast.KindGlobalKeyword
	result := converter.createNode(node, "TSModuleDeclaration")

	switch {
	case isGlobal:
		result.Set("body", converter.convertChild(module.Body, nil))
		result.Set("declare", false)
		result.Set("global", false)
		result.Set("id", converter.convertChild(node.Name(), nil))
		result.Set("kind", "global")
	case node.Name().Kind == ast.KindStringLiteral:
		body := converter.convertChild(module.Body, nil)
		result.Set("kind", "module")
		if body != nil {
			result.Set("body", body)
		}
		result.Set("declare", false)
		result.Set("global", false)
		result.Set("id", converter.convertChild(node.Name(), nil))
	default:
		innermost := node
		name := converter.identifierNode([2]int{converter.getStart(node.Name()), node.Name().End()}, node.Name().Text())
		for innermost.Body() != nil && innermost.Body().Kind == ast.KindModuleDeclaration && innermost.Body().Name() != nil {
			innermost = innermost.Body()
			isDeclare = isDeclare || hasModifier(ast.KindDeclareKeyword, innermost)
			nextName := innermost.Name()
			right := converter.identifierNode([2]int{converter.getStart(nextName), nextName.End()}, nextName.Text())
			name = converter.createNodeWithRange("TSQualifiedName", [2]int{name.Range[0], right.Range[1]},
				"left", name,
				"right", right)
		}
		kind := "module"
		if innermost.AsModuleDeclaration().Keyword == ast.KindNamespaceKeyword {
			kind = "namespace"
		}
		result.Set("body", converter.convertChild(innermost.Body(), nil))
		result.Set("declare", false)
		result.Set("global", false)
		result.Set("id", name)
		result.Set("kind", kind)
	}

	result.Set("declare", isDeclare)
	if isGlobal {
		result.Set("global", true)
	}
	return converter.fixExports(node, result)
}

// convertBinary is upstream's BinaryExpression case.
func (converter *Converter) convertBinary(node *ast.Node) *Node {
	binary := node.AsBinaryExpression()
	operatorKind := binary.OperatorToken.Kind
	if operatorKind == ast.KindCommaToken {
		expressions := []*Node{}
		left := converter.convertChild(binary.Left, nil)
		if left.Is("SequenceExpression") && binary.Left.Kind != ast.KindParenthesizedExpression {
			expressions = append(expressions, left.List("expressions")...)
		} else {
			expressions = append(expressions, left)
		}
		expressions = append(expressions, converter.convertChild(binary.Right, nil))
		return converter.createNode(node, "SequenceExpression", "expressions", expressions)
	}

	expressionType, operator := binaryExpressionType(operatorKind)
	if converter.allowPattern && expressionType == "AssignmentExpression" {
		return converter.createNode(node, "AssignmentPattern",
			"decorators", noNodes,
			"left", converter.convertPattern(binary.Left, node),
			"optional", false,
			"right", converter.convertChild(binary.Right, nil),
			"typeAnnotation", nil)
	}
	return converter.createNode(node, expressionType,
		"operator", operator,
		"left", converter.converter(binary.Left, node, expressionType == "AssignmentExpression"),
		"right", converter.convertChild(binary.Right, nil))
}

// binaryExpressionType is upstream's getBinaryExpressionType.
func binaryExpressionType(kind ast.Kind) (string, string) {
	operator := scanner.TokenToString(kind)
	switch kind {
	case ast.KindAmpersandAmpersandEqualsToken, ast.KindAmpersandEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindBarBarEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken,
		ast.KindEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindMinusEqualsToken, ast.KindPercentEqualsToken, ast.KindPlusEqualsToken,
		ast.KindQuestionQuestionEqualsToken, ast.KindSlashEqualsToken:
		return "AssignmentExpression", operator
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return "LogicalExpression", operator
	}
	return "BinaryExpression", operator
}

// isValidEscape is upstream's #isValidEscape.
func isValidEscape(text string) bool {
	if !strings.Contains(text, `\x`) && !strings.Contains(text, `\u`) {
		return true
	}
	for index := 0; index+1 < len(text); index++ {
		if text[index] != '\\' {
			continue
		}
		switch text[index+1] {
		case 'u':
			rest := text[index+2:]
			if !strings.HasPrefix(rest, "{") && !(len(rest) >= 4 && isHexDigits(rest[:4])) {
				return false
			}
		case 'x':
			rest := text[index+2:]
			if !(len(rest) >= 2 && isHexDigits(rest[:2])) {
				return false
			}
		}
	}
	return true
}

func isHexDigits(text string) bool {
	for index := 0; index < len(text); index++ {
		character := text[index]
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f' || character >= 'A' && character <= 'F') {
			return false
		}
	}
	return true
}

// numberValue is JavaScript's Number(text) for a numeric literal's source.
func numberValue(raw string) float64 {
	text := strings.ReplaceAll(raw, "_", "")
	lower := strings.ToLower(text)
	var base int
	switch {
	case strings.HasPrefix(lower, "0x"):
		base = 16
	case strings.HasPrefix(lower, "0o"):
		base = 8
	case strings.HasPrefix(lower, "0b"):
		base = 2
	}
	if base != 0 {
		value, err := strconv.ParseUint(text[2:], base, 64)
		if err != nil {
			return math.NaN()
		}
		return float64(value)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return math.NaN()
	}
	return value
}

// bigintString is String(BigInt(text)): the decimal digits of a bigint literal's value.
func bigintString(text string) string {
	lower := strings.ToLower(text)
	base := 10
	digits := text
	switch {
	case strings.HasPrefix(lower, "0x"):
		base, digits = 16, text[2:]
	case strings.HasPrefix(lower, "0o"):
		base, digits = 8, text[2:]
	case strings.HasPrefix(lower, "0b"):
		base, digits = 2, text[2:]
	}
	value, ok := new(big.Int).SetString(digits, base)
	if !ok {
		return text
	}
	return value.String()
}
