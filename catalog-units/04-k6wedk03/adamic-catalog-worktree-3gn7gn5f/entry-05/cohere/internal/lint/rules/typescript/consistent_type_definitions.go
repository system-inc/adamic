package typescript

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageConsistentTypeDefinitionsInterfaceOverType is reported when a type alias whose body is an
// object literal could be an interface.
var messageConsistentTypeDefinitionsInterfaceOverType = rule.Message{
	Id: "interfaceOverType",
	Description: "Use an `interface` instead of a `type` here. Both describe the same object " +
		"shape, but an interface is the one this project writes: it can be reopened by a later " +
		"declaration, it shows up under its own name in compiler errors rather than being " +
		"expanded inline, and mixing the two spellings for the same job makes a reader ask " +
		"whether the difference was meant.",
}

// messageConsistentTypeDefinitionsTypeOverInterface is the mirror.
var messageConsistentTypeDefinitionsTypeOverInterface = rule.Message{
	Id: "typeOverInterface",
	Description: "Use a `type` instead of an `interface` here. Both describe the same object " +
		"shape, but a type alias is the one this project writes: it cannot be reopened by a " +
		"declaration somewhere else in the program, so what you read at the definition is all " +
		"there is, and mixing the two spellings for the same job makes a reader ask whether the " +
		"difference was meant.",
}

// ConsistentTypeDefinitionsStyle is which spelling the project prefers.
type ConsistentTypeDefinitionsStyle string

const (
	// ConsistentTypeDefinitionsInterface reports a type alias of an object literal.
	ConsistentTypeDefinitionsInterface ConsistentTypeDefinitionsStyle = "interface"
	// ConsistentTypeDefinitionsType reports an interface declaration.
	ConsistentTypeDefinitionsType ConsistentTypeDefinitionsStyle = "type"
)

// ConsistentTypeDefinitionsOptions is the rule's whole configuration.
//
// Upstream's schema is a single positional string enum, so the value the config layer hands the
// decoder is a bare JSON string rather than an object.
type ConsistentTypeDefinitionsOptions struct {
	// Style is which of the two spellings to enforce. Upstream defaults to `interface`.
	Style ConsistentTypeDefinitionsStyle
}

// DefaultConsistentTypeDefinitionsSettings is upstream's `defaultOptions: ['interface']`.
//
// Spelled out rather than left to the zero value, which would be the empty string and would match
// neither arm, making the rule silent on everything while looking configured.
func DefaultConsistentTypeDefinitionsSettings() ConsistentTypeDefinitionsOptions {
	return ConsistentTypeDefinitionsOptions{Style: ConsistentTypeDefinitionsInterface}
}

// DecodeConsistentTypeDefinitionsOptions reads this rule's configuration.
//
// Hand-rolled because the wire value is a bare string and because the default is not the zero
// value. A rule configured as a plain "error" reaches here with empty input, and a plain string
// field would read that as "", which matches neither arm and silences the rule on every file --
// indistinguishable, in a violation count, from a clean tree.
func DecodeConsistentTypeDefinitionsOptions(raw []byte) (any, error) {
	settings := DefaultConsistentTypeDefinitionsSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var style string
	if err := json.Unmarshal(raw, &style); err != nil {
		return settings, fmt.Errorf(
			"consistent-type-definitions takes the string \"interface\" or \"type\": %w", err)
	}
	switch ConsistentTypeDefinitionsStyle(style) {
	case ConsistentTypeDefinitionsInterface, ConsistentTypeDefinitionsType:
		settings.Style = ConsistentTypeDefinitionsStyle(style)
		return settings, nil
	}
	return settings, fmt.Errorf(
		"consistent-type-definitions takes \"interface\" or \"type\", got %q", style)
}

// ConsistentTypeDefinitions reports a type declaration written in the spelling the project did not
// choose.
//
//	valid:   type U = string;                     under either option, it is not an object shape
//	valid:   type V = { x: number } | { y: string };   a union, not an object literal
//	valid:   interface I { x: number }            under 'interface'
//	invalid: type T = { x: number };              under 'interface'
//	invalid: interface I { x: number }            under 'type'
//
// # Only an object LITERAL body is interchangeable, which is what the selector says
//
// Upstream's listener key is the esquery selector
// `TSTypeAliasDeclaration[typeAnnotation.type='TSTypeLiteral']`, and the filter is the whole point:
// `type U = string` and `type V = {x: number} | {y: string}` are not expressible as interfaces, so
// they are clean under `interface` and upstream never reports them. Here the same test is a kind
// check on the alias's type node, `ast.KindTypeLiteral`, which is what that selector's attribute
// clause compares. Measured across the corpus's four such valid cases.
//
// # The fixer is the port, and its shape is dictated by tokens rather than nodes
//
// Both repairs rewrite the KEYWORDS around a declaration rather than any node's own text, so both
// are built from spans computed against the source rather than from `ReplaceNode`. Upstream reaches
// them through `sourceCode.getTokenBefore`; there is no token list here, so each span is found by
// scanning the declaration's own text, which cannot run past the declaration.
//
//	type  T  =  { x: number };        ->    interface  T   { x: number }
//	     ^^^^ replaced with `interface`      ^^^ the `=` and everything to the `{` becomes one space
//	                             ^ the `;` and anything after the `}` is removed
//
// The `interface` -> `type` direction adds two cases the other does not have, and both are in the
// corpus. An `extends` clause has no interface equivalent, so it becomes an intersection appended
// after the body (`interface I extends A {}` -> `type I = {} & A`), and a default-exported
// interface has to be split, because `export default type I = ...` is not legal syntax.
//
// # `declare global` is a deliberate DECLINE, not a missing case
//
// An interface inside `declare global` is reported and NOT fixed. Upstream returns a null fixer
// there, citing its own issue 2707: a type alias inside a global augmentation does not do what the
// interface did, so the repair would change meaning rather than spelling. Two corpus cases carry
// `output: null` for exactly this, and the installed build was driven over both to confirm the
// decline is real rather than a fixer that happens to fail.
//
// Reproducing a decline matters more than it looks: a fixer that repairs a case upstream refuses to
// touch is a defect no message-id fixture can see, because the finding is identical either way.
var ConsistentTypeDefinitions = rule.Rule{
	Name: "@typescript-eslint/consistent-type-definitions",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[ConsistentTypeDefinitionsOptions](options)
		if !configured {
			settings = DefaultConsistentTypeDefinitionsSettings()
		}

		if settings.Style == ConsistentTypeDefinitionsInterface {
			return rule.Listeners{
				ast.KindTypeAliasDeclaration: func(node *ast.Node) {
					checkConsistentTypeDefinitionsAlias(ctx, node)
				},
			}
		}
		return rule.Listeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkConsistentTypeDefinitionsInterface(ctx, node)
			},
		}
	},
}

// checkConsistentTypeDefinitionsAlias handles `type T = { ... }` under the `interface` option.
func checkConsistentTypeDefinitionsAlias(ctx rule.Context, node *ast.Node) {
	alias := node.AsTypeAliasDeclaration()
	if alias == nil || alias.Type == nil {
		return
	}
	// Upstream's `[typeAnnotation.type='TSTypeLiteral']`. A union, a primitive or a reference is not
	// expressible as an interface and is left alone.
	//
	// The parentheses are looked THROUGH, which upstream gets for free and this tree does not.
	// typescript-go keeps `(...)` as a `KindParenthesizedType` node and nests one per pair, where
	// ESTree drops them entirely -- so upstream's selector sees `TSTypeLiteral` directly on
	// `type Foo = ({ a: string })` and a kind check here sees `KindParenthesizedType`. Measured with
	// an ancestry probe: three corpus cases carry them, one of them nine deep, and all three were
	// silent before this unwrap.
	body := consistentTypeDefinitionsUnwrapParentheses(alias.Type)
	if body.Kind != ast.KindTypeLiteral {
		return
	}
	name := alias.Name()
	if name == nil {
		return
	}

	fixes := consistentTypeDefinitionsAliasFixes(ctx, node, alias, body)
	if len(fixes) == 0 {
		ctx.ReportNode(name, messageConsistentTypeDefinitionsInterfaceOverType)
		return
	}
	ctx.ReportNodeWithFixes(name, messageConsistentTypeDefinitionsInterfaceOverType, fixes...)
}

// consistentTypeDefinitionsAliasFixes builds the three edits that turn an alias into an interface.
//
// Returns nothing rather than a partial repair if any span cannot be located, which is the same
// decision upstream expresses by throwing on a missing token: a fixer that applies two of three
// edits writes source that does not parse.
func consistentTypeDefinitionsAliasFixes(
	ctx rule.Context,
	node *ast.Node,
	alias *ast.TypeAliasDeclaration,
	body *ast.Node,
) []rule.Fix {
	if ctx.SourceFile == nil {
		return nil
	}
	text := ctx.SourceFile.Text()
	statement := rule.TokenRange(ctx.SourceFile, node)
	if statement.Pos() < 0 || statement.End() > len(text) {
		return nil
	}

	// The `type` keyword opens the declaration once modifiers are past, so it is the first `type`
	// at or after the name-preceding region rather than anywhere in the statement.
	keyword := consistentTypeDefinitionsKeywordRange(text, statement, "type")
	if keyword == nil {
		return nil
	}

	// The literal itself, not the parentheses around it. Upstream's repair strips them -- its
	// `output` for `type Foo = (({ a: string }))` is `interface Foo { a: string }` -- and taking the
	// span from the unwrapped node makes the surrounding parens fall inside the spans that are
	// replaced and removed, which strips them for the same reason.
	bodyRange := rule.TokenRange(ctx.SourceFile, body)

	// Everything from the end of whatever precedes the `=` through the start of the body becomes a
	// single space. Upstream computes this as [beforeEqualsToken.end, typeAnnotation.start] so that
	// a comment sitting before the `=` survives, which is why the left edge is the token BEFORE the
	// equals rather than the equals itself.
	equals := consistentTypeDefinitionsEqualsBefore(text, statement, bodyRange.Pos())
	if equals < 0 {
		return nil
	}
	beforeEquals := consistentTypeDefinitionsEndOfPreviousToken(text, statement.Pos(), equals)
	if beforeEquals < 0 {
		return nil
	}

	return []rule.Fix{
		rule.ReplaceRange(*keyword, "interface"),
		rule.ReplaceRange(core.NewTextRange(beforeEquals, bodyRange.Pos()), " "),
		// The `;` and anything else trailing the body. Upstream removes to the end of the whole
		// declaration, so a semicolon on a later line goes too.
		rule.RemoveRange(core.NewTextRange(bodyRange.End(), statement.End())),
	}
}

// checkConsistentTypeDefinitionsInterface handles `interface I { ... }` under the `type` option.
func checkConsistentTypeDefinitionsInterface(ctx rule.Context, node *ast.Node) {
	name := node.Name()
	if name == nil {
		return
	}

	// `declare global { interface I {} }` is reported and deliberately not repaired. See the rule's
	// doc comment; upstream cites its issue 2707.
	if consistentTypeDefinitionsIsInsideDeclareGlobal(node) {
		ctx.ReportNode(name, messageConsistentTypeDefinitionsTypeOverInterface)
		return
	}

	fixes := consistentTypeDefinitionsInterfaceFixes(ctx, node)
	if len(fixes) == 0 {
		ctx.ReportNode(name, messageConsistentTypeDefinitionsTypeOverInterface)
		return
	}
	ctx.ReportNodeWithFixes(name, messageConsistentTypeDefinitionsTypeOverInterface, fixes...)
}

// consistentTypeDefinitionsInterfaceFixes builds the edits that turn an interface into an alias.
func consistentTypeDefinitionsInterfaceFixes(ctx rule.Context, node *ast.Node) []rule.Fix {
	if ctx.SourceFile == nil {
		return nil
	}
	declaration := node.AsInterfaceDeclaration()
	if declaration == nil {
		return nil
	}
	text := ctx.SourceFile.Text()
	statement := rule.TokenRange(ctx.SourceFile, node)
	if statement.Pos() < 0 || statement.End() > len(text) {
		return nil
	}

	keyword := consistentTypeDefinitionsKeywordRange(text, statement, "interface")
	if keyword == nil {
		return nil
	}

	// The head ends after the type parameters when there are any, and after the name otherwise.
	// Upstream: `node.typeParameters ?? node.id`.
	headEnd := rule.TokenRange(ctx.SourceFile, node.Name()).End()
	if parameters := declaration.TypeParameters; parameters != nil && len(parameters.Nodes) > 0 {
		last := parameters.Nodes[len(parameters.Nodes)-1]
		// The closing `>` sits past the last parameter, so the head runs to it rather than to the
		// parameter's own end.
		headEnd = consistentTypeDefinitionsCloseAngleAfter(text, rule.TokenRange(ctx.SourceFile, last).End(), statement.End())
		if headEnd < 0 {
			return nil
		}
	}

	bodyStart := consistentTypeDefinitionsBodyBraceStart(text, headEnd, statement.End())
	if bodyStart < 0 {
		return nil
	}

	fixes := []rule.Fix{
		rule.ReplaceRange(*keyword, "type"),
		rule.ReplaceRange(core.NewTextRange(headEnd, bodyStart), " = "),
	}

	// An `extends` clause becomes an intersection appended after the body, because a type alias has
	// no heritage syntax.
	//
	// Built as ONE insertion rather than one per heritage type, and that is load-bearing rather than
	// tidy. Upstream emits a separate `insertTextAfter` per type and ESLint applies same-position
	// insertions in the order proposed; the edit engine here applies fixes back to front so that an
	// earlier replacement cannot move offsets a later one was computed against, which reverses a run
	// of insertions sharing a position. Measured: `interface A extends B, C {}` repaired to
	// `type A = {...} & C & B`, valid TypeScript that names the constraints in the wrong order and
	// that no message-id fixture can see.
	bodyEnd := statement.End()
	heritage := consistentTypeDefinitionsHeritageTypes(declaration)
	if len(heritage) > 0 {
		var appended strings.Builder
		for _, clause := range heritage {
			clauseRange := rule.TokenRange(ctx.SourceFile, clause)
			appended.WriteString(" & ")
			appended.WriteString(text[clauseRange.Pos():clauseRange.End()])
		}
		fixes = append(fixes, rule.Fix{
			Range: core.NewTextRange(bodyEnd, bodyEnd),
			Text:  appended.String(),
		})
	}

	// `export default interface I {}` has to be split: `export default type I = {}` is not legal,
	// so the `export default` is removed from the front and re-attached after the body by name.
	if consistentTypeDefinitionsHasModifier(node, ast.KindDefaultKeyword) &&
		consistentTypeDefinitionsHasModifier(node, ast.KindExportKeyword) {
		nameText := node.Name().Text()
		fixes = append(fixes,
			rule.RemoveRange(core.NewTextRange(statement.Pos(), *consistentTypeDefinitionsPtr(keyword.Pos()))),
			rule.Fix{
				Range: core.NewTextRange(bodyEnd, bodyEnd),
				Text:  "\nexport default " + nameText,
			},
		)
	}

	return fixes
}

// consistentTypeDefinitionsUnwrapParentheses looks through any number of parenthesized-type nodes.
//
// This exists purely for a parser difference. In ESTree a parenthesized type is not represented at
// all, so upstream's selector and its fixer both see the literal directly. typescript-go keeps one
// `KindParenthesizedType` per pair and nests them, so `((({...})))` is three nodes deep. Upstream's
// corpus carries exactly that case and expects it both reported and repaired with the parentheses
// removed.
func consistentTypeDefinitionsUnwrapParentheses(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Kind == ast.KindParenthesizedType {
		inner := current.AsParenthesizedTypeNode()
		if inner == nil || inner.Type == nil {
			return current
		}
		current = inner.Type
	}
	return current
}

// consistentTypeDefinitionsPtr is a small helper so a computed int can be taken by address inline.
func consistentTypeDefinitionsPtr(value int) *int { return &value }

// consistentTypeDefinitionsHeritageTypes returns the types an interface extends, in source order.
func consistentTypeDefinitionsHeritageTypes(declaration *ast.InterfaceDeclaration) []*ast.Node {
	types := []*ast.Node{}
	clauses := declaration.HeritageClauses
	if clauses == nil {
		return types
	}
	for _, clause := range clauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage == nil || heritage.Types == nil {
			continue
		}
		types = append(types, heritage.Types.Nodes...)
	}
	return types
}

// consistentTypeDefinitionsIsInsideDeclareGlobal answers upstream's
// `isCurrentlyTraversedNodeWithinModuleDeclaration`.
//
// Upstream walks every ancestor looking for a `TSModuleDeclaration` that is both `declare` and of
// kind `global`. Here that is a `KindModuleDeclaration` carrying a declare modifier whose name is
// the identifier `global` -- measured, because `declare global` and `declare module 'x'` are the
// same node kind and only the name separates them.
func consistentTypeDefinitionsIsInsideDeclareGlobal(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindModuleDeclaration {
			continue
		}
		if !consistentTypeDefinitionsHasModifier(current, ast.KindDeclareKeyword) {
			continue
		}
		name := current.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "global" {
			return true
		}
	}
	return false
}

// consistentTypeDefinitionsHasModifier answers whether a declaration carries a modifier keyword.
func consistentTypeDefinitionsHasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}

// consistentTypeDefinitionsKeywordRange finds the span of a declaration's own leading keyword.
//
// Scanned rather than read off a token list, which this tree does not expose to a rule. The scan is
// bounded by the declaration and looks for the keyword as a whole word, so a modifier or a type
// name containing the same letters cannot be mistaken for it.
func consistentTypeDefinitionsKeywordRange(
	text string,
	statement core.TextRange,
	keyword string,
) *core.TextRange {
	for offset := statement.Pos(); offset+len(keyword) <= statement.End(); offset++ {
		if !strings.HasPrefix(text[offset:], keyword) {
			continue
		}
		if offset > statement.Pos() && consistentTypeDefinitionsIsWordByte(text[offset-1]) {
			continue
		}
		after := offset + len(keyword)
		if after < len(text) && consistentTypeDefinitionsIsWordByte(text[after]) {
			continue
		}
		span := core.NewTextRange(offset, after)
		return &span
	}
	return nil
}

// consistentTypeDefinitionsIsWordByte answers whether a byte can appear inside an identifier.
func consistentTypeDefinitionsIsWordByte(b byte) bool {
	return b == '_' || b == '$' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') ||
		b >= 0x80
}

// consistentTypeDefinitionsEqualsBefore finds the `=` that introduces an alias's body.
//
// Searched BACKWARD from the body, which is what makes it the right `=`: a default in a type
// parameter list (`type T<U = string> = {...}`) puts an earlier `=` in the statement, and a forward
// scan would find that one and produce a repair that deletes the parameter list.
func consistentTypeDefinitionsEqualsBefore(text string, statement core.TextRange, bodyStart int) int {
	for offset := bodyStart - 1; offset >= statement.Pos(); offset-- {
		if text[offset] == '=' {
			return offset
		}
	}
	return -1
}

// consistentTypeDefinitionsEndOfPreviousToken returns the position just past the last non-space
// byte before `before`.
//
// This is upstream's `getTokenBefore(equalsToken, {includeComments: true}).range[1]`. Comments count
// as tokens there, so a comment before the `=` is preserved rather than swallowed by the
// replacement, and scanning back over whitespace only reproduces that.
func consistentTypeDefinitionsEndOfPreviousToken(text string, lowerBound int, before int) int {
	offset := before - 1
	for offset >= lowerBound && consistentTypeDefinitionsIsSpaceByte(text[offset]) {
		offset--
	}
	if offset < lowerBound {
		return -1
	}
	return offset + 1
}

// consistentTypeDefinitionsIsSpaceByte answers whether a byte is ASCII whitespace.
func consistentTypeDefinitionsIsSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

// consistentTypeDefinitionsCloseAngleAfter finds the `>` closing a type parameter list.
//
// Scanned forward from the last parameter with a depth counter, because a parameter's own
// constraint can carry nested angle brackets: `interface I<T extends Map<string, number>> {}`.
func consistentTypeDefinitionsCloseAngleAfter(text string, from int, limit int) int {
	depth := 0
	for offset := from; offset < limit && offset < len(text); offset++ {
		switch text[offset] {
		case '<':
			depth++
		case '>':
			if depth == 0 {
				return offset + 1
			}
			depth--
		case '{':
			// The body started, so the list was never closed. Refuse rather than guess.
			return -1
		}
	}
	return -1
}

// consistentTypeDefinitionsBodyBraceStart finds the `{` opening an interface body.
func consistentTypeDefinitionsBodyBraceStart(text string, from int, limit int) int {
	for offset := from; offset < limit && offset < len(text); offset++ {
		if text[offset] == '{' {
			return offset
		}
	}
	return -1
}
