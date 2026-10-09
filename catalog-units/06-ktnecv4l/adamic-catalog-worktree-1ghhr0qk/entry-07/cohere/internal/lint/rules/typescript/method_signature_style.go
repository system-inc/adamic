package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageMethodSignatureStyleErrorMethod is upstream's `errorMethod`.
func messageMethodSignatureStyleErrorMethod() rule.Message {
	return rule.Message{
		Id: "errorMethod",
		Description: "This is written as a shorthand method signature, which TypeScript checks " +
			"bivariantly: a parameter can be narrowed by an implementer and the compiler will not " +
			"object, so a call that type-checks can still be wrong at runtime. A function property " +
			"is checked the strict way and catches that.",
	}
}

// messageMethodSignatureStyleErrorProperty is upstream's `errorProperty`.
func messageMethodSignatureStyleErrorProperty() rule.Message {
	return rule.Message{
		Id: "errorProperty",
		Description: "This project writes interface members as shorthand methods, and this one is " +
			"a property holding a function type. The two differ in how the compiler checks " +
			"parameters, so mixing them in one codebase means a reader cannot tell which rule " +
			"applies without looking at the punctuation.",
	}
}

// messageMethodSignatureStyleConvertToMethod is upstream's `convertToMethodSignature`.
func messageMethodSignatureStyleConvertToMethod() rule.Message {
	return rule.Message{
		Id: "convertToMethodSignature",
		Description: "Convert to a method signature. There is no syntax for a readonly method, so " +
			"this drops the `readonly` modifier and the member becomes reassignable.",
	}
}

// MethodSignatureStyle requires one spelling for a signature, a shorthand method or a property.
//
//	valid (property): interface I { m: (a: string) => void; }
//	valid (method):   interface I { m(a: string): void; }
//	valid (property): interface I { m(): this; }        a this return has no property spelling
//	invalid (property): interface I { m(a: string): void; }   becomes m: (a: string) => void;
//	invalid (method):   interface I { m: (a: string) => void; } becomes m(a: string): void;
//
// Ported from `@typescript-eslint/method-signature-style`, reading the clone at
// `packages/eslint-plugin/src/rules/method-signature-style.ts` and measuring every verdict and every
// repair against the installed 8.67.0 build driven through the ESLint 10.8.1 Linter API.
//
// # The repair rebuilds a signature, and it is type-safe because it copies rather than re-renders
//
// This is the shape that lost type information twice in this project: a fixer that constructs a
// replacement drops whatever it does not think to re-render. Upstream avoids it by splicing RAW
// SOURCE for the two parts that can hold anything: the parameter list is copied from its opening
// parenthesis to its closing one, and the type parameter list is copied whole. Only the key, the
// arrow and the return type are assembled.
//
// So everything inside those spans survives without being enumerated. Measured on the installed
// build, and every row also checked through the TypeScript compiler to confirm the member's type is
// unchanged:
//
//	m<T extends object>(a: T): T;   m: <T extends object>(a: T) => T;   generics and constraints
//	m(a?: string): void;            m: (a?: string) => void;            an optional parameter
//	m(...a: string[]): void;        m: (...a: string[]) => void;        a rest parameter
//	m?(a: string): void;            m?: (a: string) => void;            an optional MEMBER
//	m(/* c */ a: string): void;     m: (/* c */ a: string) => void;     a comment in the list
//	[k](a: string): void;           [k]: (a: string) => void;           a computed key
//
// # An absent return type becomes an explicit `any`, and that is not a widening
//
// `getMethodReturnType` returns the literal string `any` when there is no annotation, because a
// method signature with no return type already IS `any` and the property form has nowhere to put an
// implicit one. Measured through the compiler: `m(a: string);` and `m: (a: string) => any;` are both
// `(a: string) => any`. Reproduced rather than declined.
//
// # Three places upstream declines, all reproduced
//
// A `this` return type has no property spelling, since `this` in a function type means something
// else, so the finding fires and the repair is withheld. A member inside a module declaration is
// reported without a repair, which upstream does not explain and which is reproduced rather than
// improved. And a `readonly` function property converts to a method that is no longer readonly,
// which is a behavior change, so upstream offers it as a SUGGESTION rather than a fix.
//
// The `readonly` case is the one worth naming: it is the same judgment this project applies when it
// declines a fixer that would drop a type, arrived at independently by upstream, and it is why this
// rule carries both `fixable` and `hasSuggestions`.
//
// # Overloads merge into an intersection
//
// Two method signatures sharing a key cannot become two properties, so the repair writes one
// property whose type is an intersection of function types and deletes the others. The finding still
// fires once per signature, so a pair of overloads reports twice and the first repair does all the
// work.
var MethodSignatureStyle = rule.Rule{
	Name: "@typescript-eslint/method-signature-style",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The zero value is not the default and this is the normal path, not a defensive branch. A
		// rule configured as a bare `"error"` is handed nil options, because `DecodeOptionsInto`
		// errors on empty input and the config layer turns that into nil. A bare type assertion
		// would then yield the empty string, matching neither arm, and the rule would register on
		// every file and report nothing while every decoder-routed fixture stayed green.
		parsed, decoded := rule.OptionsAs[MethodSignatureStyleOptions](options)
		if !decoded {
			parsed = DefaultMethodSignatureStyleOptions()
		}

		if parsed.Style == MethodSignatureStyleProperty {
			return rule.Listeners{
				ast.KindMethodSignature: func(node *ast.Node) {
					reportMethodSignatureShouldBeProperty(ctx, node)
				},
			}
		}

		return rule.Listeners{
			ast.KindPropertySignature: func(node *ast.Node) {
				reportMethodSignaturePropertyShouldBeMethod(ctx, node)
			},
		}
	},
}

// reportMethodSignatureShouldBeProperty is upstream's `TSMethodSignature` visitor under `property`.
func reportMethodSignatureShouldBeProperty(ctx rule.Context, node *ast.Node) {
	signature := node.AsMethodSignatureDeclaration()

	// A getter or setter signature is `kind !== 'method'` upstream and is not this rule's business.
	// typescript-go gives those their own node kinds, so reaching this listener already excludes
	// them and no test is written.

	key, keyed := methodSignatureKeyText(ctx, node, signature.Name())
	if !keyed {
		return
	}

	// Upstream's `skipFix`: a `this` return type has no meaning in a function type, so the finding
	// fires and the repair is withheld.
	skipFix := methodSignatureReturnMentionsThis(signature.Type)

	// A member inside a module declaration is reported without a repair. Upstream gives no reason
	// and this reproduces the decision rather than improving on it.
	if methodSignatureIsInsideModule(node) {
		ctx.ReportNode(node, messageMethodSignatureStyleErrorMethod())
		return
	}

	siblings := methodSignatureSiblingsOf(node)
	var duplicates []*ast.Node
	for _, sibling := range siblings {
		if sibling == node || sibling.Kind != ast.KindMethodSignature {
			continue
		}
		siblingKey, siblingKeyed := methodSignatureKeyText(ctx, sibling,
			sibling.AsMethodSignatureDeclaration().Name())
		if siblingKeyed && siblingKey == key {
			duplicates = append(duplicates, sibling)
		}
	}

	if skipFix {
		ctx.ReportNode(node, messageMethodSignatureStyleErrorMethod())
		return
	}

	if len(duplicates) > 0 {
		// Every signature in the group reports, and only the FIRST one carries the repair.
		//
		// Upstream attaches the same whole-group rewrite to every signature, so its two findings
		// carry two overlapping fixes. That works there because ESLint applies one fix per pass and
		// re-lints, and the second finding is gone by the next pass. Our engine flattens a run's
		// fixes into independent proposals and refuses an overlap outright, so attaching both would
		// mean neither lands.
		//
		// Measured: upstream's `cohereAndFix` on `m(a: string): void; m(a: number): void;` writes
		// `m: ((a: string) => void) & ((a: number) => void);`, which is exactly what the first
		// signature's repair alone produces here. The judgment is unchanged, both signatures still
		// report, and only the redundant second copy of the repair is dropped.
		first := node
		for _, duplicate := range duplicates {
			if duplicate.Pos() < first.Pos() {
				first = duplicate
			}
		}
		if first != node {
			ctx.ReportNode(node, messageMethodSignatureStyleErrorMethod())
			return
		}

		fixes, buildable := methodSignatureOverloadFixes(ctx, node, duplicates, key)
		if !buildable {
			ctx.ReportNode(node, messageMethodSignatureStyleErrorMethod())
			return
		}
		ctx.ReportNodeWithFixes(node, messageMethodSignatureStyleErrorMethod(), fixes...)
		return
	}

	parameters, parametersReadable := methodSignatureParametersText(ctx, node,
		signature.Parameters, signature.TypeParameters)
	if !parametersReadable {
		ctx.ReportNode(node, messageMethodSignatureStyleErrorMethod())
		return
	}
	returnType := methodSignatureReturnTypeText(ctx, signature.Type)
	delimiter := methodSignatureDelimiterOf(ctx, node)

	ctx.ReportNodeWithFixes(node, messageMethodSignatureStyleErrorMethod(),
		rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node),
			key+": "+parameters+" => "+returnType+delimiter))
}

// methodSignatureOverloadFixes builds upstream's intersection repair for a duplicated key.
//
// Every signature sharing the key becomes one function type, joined with `&`, written at the FIRST
// of them in source order, and the rest are deleted along with the whitespace that followed them.
func methodSignatureOverloadFixes(
	ctx rule.Context,
	node *ast.Node,
	duplicates []*ast.Node,
	key string,
) ([]rule.Fix, bool) {
	ordered := make([]*ast.Node, 0, len(duplicates)+1)
	ordered = append(ordered, node)
	ordered = append(ordered, duplicates...)
	// Upstream sorts by start position, so the intersection reads in source order regardless of
	// which signature the listener happens to be visiting.
	for outer := 1; outer < len(ordered); outer++ {
		candidate := ordered[outer]
		inner := outer - 1
		for inner >= 0 && ordered[inner].Pos() > candidate.Pos() {
			ordered[inner+1] = ordered[inner]
			inner--
		}
		ordered[inner+1] = candidate
	}

	var builder strings.Builder
	for index, member := range ordered {
		signature := member.AsMethodSignatureDeclaration()
		parameters, readable := methodSignatureParametersText(ctx, member,
			signature.Parameters, signature.TypeParameters)
		if !readable {
			return nil, false
		}
		if index > 0 {
			builder.WriteString(" & ")
		}
		// Upstream writes `(${params} => ${returnType})`, and `params` already carries its own
		// parentheses, so the rendered constituent is `((a: string) => void)` from ONE pair written
		// here. Writing two produced `((( ... ` and left the parentheses unbalanced.
		builder.WriteString("(")
		builder.WriteString(parameters)
		builder.WriteString(" => ")
		builder.WriteString(methodSignatureReturnTypeText(ctx, signature.Type))
		builder.WriteString(")")
	}

	// One replacement over the whole group rather than a rewrite plus a removal per duplicate.
	//
	// Upstream emits exactly that pair, and ESLint then MERGES a finding's fixes into a single
	// range running from the first edit's start to the last edit's end, splicing the untouched text
	// between them back in. Measured on a three-overload case: upstream's rule yields a rewrite and
	// two removals, and the fix ESLint actually applies is one replacement of `[19,61)` whose text
	// ends in the newline and indentation that preceded the closing brace.
	//
	// Our engine does not merge, it flattens fixes into independent proposals, so reproducing
	// upstream's shape literally deletes the whitespace between the signatures and leaves the
	// closing brace indented as if a member were still there. Writing the merged span directly is
	// the same edit ESLint performs and it cannot half-apply.
	last := ordered[len(ordered)-1]
	groupStart := rule.TokenRange(ctx.SourceFile, ordered[0]).Pos()
	groupEnd := last.End()
	if groupStart > groupEnd || groupEnd > len(ctx.SourceFile.Text()) {
		return nil, false
	}

	return []rule.Fix{
		rule.ReplaceRange(ctx.SourceFile.AsNode().Loc.WithPos(groupStart).WithEnd(groupEnd),
			key+": "+builder.String()+methodSignatureDelimiterOf(ctx, last)),
	}, true
}

// reportMethodSignaturePropertyShouldBeMethod is upstream's `TSPropertySignature` visitor.
func reportMethodSignaturePropertyShouldBeMethod(ctx rule.Context, node *ast.Node) {
	property := node.AsPropertySignatureDeclaration()

	typeNode := property.Type
	if typeNode == nil || typeNode.Kind != ast.KindFunctionType {
		return
	}
	functionType := typeNode.AsFunctionTypeNode()

	key, keyed := methodSignatureKeyText(ctx, node, property.Name())
	if !keyed {
		return
	}
	parameters, readable := methodSignatureParametersText(ctx, typeNode,
		functionType.Parameters, functionType.TypeParameters)
	if !readable {
		ctx.ReportNode(node, messageMethodSignatureStyleErrorProperty())
		return
	}
	returnType := methodSignatureReturnTypeText(ctx, functionType.Type)
	delimiter := methodSignatureDelimiterOf(ctx, node)

	replacement := key + parameters + ": " + returnType + delimiter
	fix := rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement)

	// There is no syntax for a readonly method signature, so the conversion drops the modifier and
	// the member becomes reassignable. That is a behavior change rather than a spelling change, so
	// upstream offers it as a suggestion a human chooses rather than as a fix the engine applies
	// unattended. Reproducing the distinction is part of the port.
	if methodSignatureHasModifier(node, ast.KindReadonlyKeyword) {
		ctx.ReportNodeWithSuggestions(node, messageMethodSignatureStyleErrorProperty(),
			rule.Suggestion{
				Message: messageMethodSignatureStyleConvertToMethod(),
				Fixes:   []rule.Fix{fix},
			})
		return
	}

	ctx.ReportNodeWithFixes(node, messageMethodSignatureStyleErrorProperty(), fix)
}

// methodSignatureKeyText is upstream's `getMethodKey`.
//
// The key's own source text, with the brackets already on it for a computed key, and a `?` appended
// for an optional member. Upstream re-adds the brackets because TSESTree's `key` is the inner
// expression; our `Name()` for a computed member IS the bracketed node, so its text already carries
// them and re-adding would double them.
func methodSignatureKeyText(ctx rule.Context, member *ast.Node, name *ast.Node) (string, bool) {
	text, readable := methodSignatureNodeText(ctx, name)
	if !readable {
		return "", false
	}
	if methodSignatureQuestionTokenOf(member) != nil {
		text += "?"
	}
	return text, true
}

// methodSignatureQuestionTokenOf answers a member's optionality token, or nil.
func methodSignatureQuestionTokenOf(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindMethodSignature:
		return member.AsMethodSignatureDeclaration().PostfixToken
	case ast.KindPropertySignature:
		return member.AsPropertySignatureDeclaration().PostfixToken
	}
	return nil
}

// methodSignatureParametersText is upstream's `getMethodParams`.
//
// The parameter list is copied as RAW SOURCE from its opening parenthesis to its closing one, and
// the type parameter list is copied whole and prepended. Nothing in either is re-rendered, which is
// what makes an optional parameter, a rest parameter, a generic constraint, a default and a comment
// all survive the rewrite. A fixer that enumerated those would lose whichever it forgot, which is
// the failure this project shipped twice.
//
// An empty parameter list is written as `()` rather than sliced, matching upstream, because there is
// no first parameter to scan back from.
func methodSignatureParametersText(
	ctx rule.Context,
	owner *ast.Node,
	parameters *ast.NodeList,
	typeParameters *ast.NodeList,
) (string, bool) {
	rendered := "()"

	if parameters != nil && len(parameters.Nodes) > 0 {
		sourceText := ctx.SourceFile.Text()
		first := rule.TokenRange(ctx.SourceFile, parameters.Nodes[0]).Pos()
		last := parameters.Nodes[len(parameters.Nodes)-1].End()

		open := methodSignatureScanBackFor(sourceText, first,
			rule.TokenRange(ctx.SourceFile, owner).Pos(), '(')
		closing := methodSignatureScanForwardFor(sourceText, last, owner.End(), ')')
		if open < 0 || closing < 0 {
			return "", false
		}
		rendered = sourceText[open : closing+1]
	}

	if typeParameters != nil && len(typeParameters.Nodes) > 0 {
		sourceText := ctx.SourceFile.Text()
		first := rule.TokenRange(ctx.SourceFile, typeParameters.Nodes[0]).Pos()
		last := typeParameters.Nodes[len(typeParameters.Nodes)-1].End()

		open := methodSignatureScanBackFor(sourceText, first,
			rule.TokenRange(ctx.SourceFile, owner).Pos(), '<')
		closing := methodSignatureScanForwardFor(sourceText, last, owner.End(), '>')
		if open < 0 || closing < 0 {
			return "", false
		}
		rendered = sourceText[open:closing+1] + rendered
	}

	return rendered, true
}

// methodSignatureScanBackFor finds the last occurrence of a byte before an offset, or -1.
//
// Upstream asks the token stream for the opening parenthesis before the first parameter, which this
// tree has no equivalent for. Scanning the raw text is the same answer for well-formed source and it
// degrades to a decline rather than a panic on anything else, which is what the caller's boolean is
// for. The floor keeps the scan inside the member so a malformed signature cannot walk off into the
// file.
func methodSignatureScanBackFor(sourceText string, from int, floor int, want byte) int {
	if from > len(sourceText) {
		from = len(sourceText)
	}
	for offset := from - 1; offset >= floor && offset >= 0; offset-- {
		if sourceText[offset] == want {
			return offset
		}
	}
	return -1
}

// methodSignatureScanForwardFor finds the first occurrence of a byte at or after an offset, or -1.
func methodSignatureScanForwardFor(sourceText string, from int, ceiling int, want byte) int {
	if ceiling > len(sourceText) {
		ceiling = len(sourceText)
	}
	for offset := from; offset < ceiling; offset++ {
		if sourceText[offset] == want {
			return offset
		}
	}
	return -1
}

// methodSignatureReturnTypeText is upstream's `getMethodReturnType`.
//
// An absent annotation renders as the literal `any`, because a method signature with no return type
// already IS `any` and the property form has nowhere to put an implicit one. Measured through the
// compiler: `m(a: string);` and `m: (a: string) => any;` are the same type.
func methodSignatureReturnTypeText(ctx rule.Context, returnType *ast.Node) string {
	if returnType == nil {
		return "any"
	}
	text, readable := methodSignatureNodeText(ctx, returnType)
	if !readable {
		return "any"
	}
	return text
}

// methodSignatureDelimiterOf is upstream's `getDelimiter`.
//
// A member's trailing `;` or `,` is preserved, because a type literal written with commas would
// otherwise lose one and stop parsing. The last token is read off the source rather than from a
// token stream: the member's own end is past its delimiter when it has one.
func methodSignatureDelimiterOf(ctx rule.Context, member *ast.Node) string {
	sourceText := ctx.SourceFile.Text()
	end := member.End()
	if end <= 0 || end > len(sourceText) {
		return ""
	}
	switch sourceText[end-1] {
	case ';':
		return ";"
	case ',':
		return ","
	}
	return ""
}

// methodSignatureReturnMentionsThis is upstream's `returnTypeReferencesThisType`.
//
// A `this` type in a return position means the implementing type, which a function type cannot
// express, so the repair is withheld. The search is over the whole annotation rather than its root,
// because `this[]` and `Promise<this>` both count.
func methodSignatureReturnMentionsThis(returnType *ast.Node) bool {
	if returnType == nil {
		return false
	}

	found := false
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node.Kind == ast.KindThisType {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(returnType)
	return found
}

// methodSignatureIsInsideModule is upstream's `isNodeParentModuleDeclaration`.
//
// The walk stops at the source file, so a member at the top level answers false. Upstream gives no
// reason for the exemption and it is reproduced rather than improved.
func methodSignatureIsInsideModule(node *ast.Node) bool {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindModuleDeclaration {
			return true
		}
		if ancestor.Kind == ast.KindSourceFile {
			return false
		}
	}
	return false
}

// methodSignatureSiblingsOf returns the members of the interface or type literal holding a member.
//
// `Members()` PANICS on a kind with no member list, so the kinds are named rather than tried. The
// walk recovers per FILE rather than per rule, so one panic would cost every rule that file.
func methodSignatureSiblingsOf(node *ast.Node) []*ast.Node {
	parent := node.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeLiteral:
		return parent.Members()
	}
	return nil
}

// methodSignatureHasModifier answers whether a member carries one modifier keyword.
func methodSignatureHasModifier(node *ast.Node, keyword ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == keyword {
			return true
		}
	}
	return false
}

// methodSignatureNodeText slices a node's own source text, trivia excluded.
func methodSignatureNodeText(ctx rule.Context, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	sourceText := ctx.SourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(sourceText) || trimmed.Pos() > trimmed.End() {
		return "", false
	}
	return sourceText[trimmed.Pos():trimmed.End()], true
}

// MethodSignatureStyleSetting is upstream's one option, `'method' | 'property'`.
type MethodSignatureStyleSetting string

const (
	// MethodSignatureStyleProperty is `"property"`, upstream's default: a shorthand method reports.
	MethodSignatureStyleProperty MethodSignatureStyleSetting = "property"

	// MethodSignatureStyleMethod is `"method"`: a property holding a function type reports.
	MethodSignatureStyleMethod MethodSignatureStyleSetting = "method"
)

// MethodSignatureStyleOptions is the rule's configuration.
//
// Upstream's schema is a bare STRING rather than an object, matching
// `class-literal-property-style` in this package, and `defaultOptions` is `['property']`.
type MethodSignatureStyleOptions struct {
	Style MethodSignatureStyleSetting
}

// DefaultMethodSignatureStyleOptions is upstream's configured-nothing behavior.
//
// Exported because a fixture asserting the default has to be able to name it, and because a caller
// starting from the Go zero value would get an empty style, which matches neither arm and silences
// the rule on every input while looking configured.
func DefaultMethodSignatureStyleOptions() MethodSignatureStyleOptions {
	return MethodSignatureStyleOptions{Style: MethodSignatureStyleProperty}
}

// DecodeMethodSignatureStyleOptions reads upstream's bare-string option.
//
// Hand-written rather than `rule.DecodeOptionsInto` for the same two reasons as its sibling: the
// wire value is a JSON string rather than an object, and the default is `property` rather than a Go
// zero value, so a generic decode would hand the rule an empty style that silences it while every
// fixture stayed green.
func DecodeMethodSignatureStyleOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[string]()(raw)
	if err != nil {
		return DefaultMethodSignatureStyleOptions(), err
	}

	wire, _ := decoded.(string)
	switch MethodSignatureStyleSetting(wire) {
	case MethodSignatureStyleProperty:
		return MethodSignatureStyleOptions{Style: MethodSignatureStyleProperty}, nil
	case MethodSignatureStyleMethod:
		return MethodSignatureStyleOptions{Style: MethodSignatureStyleMethod}, nil
	}
	return DefaultMethodSignatureStyleOptions(), nil
}
