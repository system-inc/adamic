package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferFunctionType flags an interface or type literal whose only member is a call signature.
//
//	valid:   interface Foo { (): void; bar: number; }
//	valid:   interface Foo extends Bar { (): void; }
//	valid:   type Foo = { (): void; new (): Bar; };
//	invalid: interface Foo { (): string; }              type Foo = () => string;
//	invalid: type Foo = { (): string; };                type Foo = () => string;
//	invalid: interface Foo<T> { (bar: T): string; }     type Foo<T> = (bar: T) => string;
//
// A one-member interface holding only a call signature is a function type wearing four extra lines
// of syntax. `type Foo = () => string` says the same thing, reads at a glance, and does not invite
// a reader to look for the other members.
//
// # Two messages, and the second one exists to avoid giving bad advice
//
// The ordinary finding is `functionTypeOverCallableType`. But an interface whose signature mentions
// `this` cannot be rewritten as a function type without changing what `this` means, so the rule
// swaps in `unexpectedThisOnFunctionOnlyInterface`, points at the `this` rather than the member,
// and offers no repair. Upstream reports only the FIRST `this` when there are several, deliberately,
// so the reader gets one message instead of a pile.
//
// A `this` inside a NESTED type literal does not count, because that code is already invalid
// TypeScript and upstream declines to claim the `this` refers to the interface when the compiler
// says it refers to nothing.
//
// # Three shape differences from the tree upstream walks, all measured
//
// `export interface Foo` is an ExportNamedDeclaration wrapping an interface upstream, and here it is
// an interface CARRYING a modifier. Same for `export default`. Upstream's fixable test reads
// `node.parent.type === ExportDefaultDeclaration` and its comment relocation reads
// `node.parent.type === ExportNamedDeclaration`; both are modifier tests here, and a port copying
// the parent test would never fire either branch.
//
// An interface's `extends` entry is a `KindTypeReference` here and an expression upstream. Probed
// rather than assumed, after the first version of this rule's probe crashed on the assumption: the
// `Function` test has to read a type reference's name, not an expression.
//
// A member's `Pos()` includes its leading comment and the newline before it, while upstream's
// `member.range[0]` is the token start with comments fetched separately. The fixer below takes the
// member text from `rule.TokenRange` for exactly this reason; using `Pos()` would fold the comment
// into the text and then insert it a second time.
//
// # The repair, and the one case it declines
//
// The rewrite turns `(args): Return` into `(args) => Return`, wraps the result in parentheses when
// the surrounding type is a union, intersection, or array, and for an interface rebuilds the whole
// declaration as a type alias carrying the original type parameters. Comments attached to the
// member move with it, and for an exported interface they move ABOVE the export rather than
// between `export` and `type`, which is the only place they are still legal.
//
// Upstream declines to fix `export default interface`, because the replacement would have to become
// `export default type`, which is not a thing. That decline is reproduced.
//
// One further decline is this port's own, and it is stated rather than silent. A member separated by
// a COMMA rather than a semicolon is legal, and upstream's fixer only strips a semicolon, so
// `type Foo = { (): string, };` is rewritten to `type Foo = () => string,;`. Measured on the
// installed 8.67.0 build and then handed to the TypeScript parser, which rejects it with
// "';' expected". The finding is reproduced and the repair is withheld, because a fix is applied
// unattended and this one turns working code into a syntax error.
var PreferFunctionType = rule.Rule{
	Name: "@typescript-eslint/prefer-function-type",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				declaration := node.AsInterfaceDeclaration()
				if declaration.Members == nil || len(declaration.Members.Nodes) != 1 {
					return
				}
				if interfaceHasOneSupertype(declaration) {
					return
				}
				checkPreferFunctionTypeMember(ctx, declaration.Members.Nodes[0], node,
					collectInterfaceThisTypes(declaration.Members.Nodes[0]))
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				literal := node.AsTypeLiteralNode()
				if literal.Members == nil || len(literal.Members.Nodes) != 1 {
					return
				}
				// A type literal never reports the `this` message: upstream's tracking array is
				// populated only while inside an interface, and it passes null here.
				checkPreferFunctionTypeMember(ctx, literal.Members.Nodes[0], node, nil)
			},
		}
	},
}

// interfaceHasOneSupertype is upstream's `hasOneSupertype`, whose name says the opposite of what it
// answers: it returns true when the interface has a supertype that DISQUALIFIES it, so the caller
// declines. Reproduced under upstream's name so the two read the same.
//
// No supertypes means the interface is a candidate. More than one means it is not. Exactly one is a
// candidate only when that one is the global `Function`, which contributes nothing a function type
// does not already have.
func interfaceHasOneSupertype(declaration *ast.InterfaceDeclaration) bool {
	if declaration.HeritageClauses == nil {
		return false
	}
	types := []*ast.Node{}
	for _, clause := range declaration.HeritageClauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage.Types == nil {
			continue
		}
		types = append(types, heritage.Types.Nodes...)
	}
	if len(types) == 0 {
		return false
	}
	if len(types) != 1 {
		return true
	}
	// Upstream tests the name only, without resolving it, so a locally declared `Function` is
	// treated the same as the global one. Reproduced rather than improved: resolving would report
	// an input upstream is silent on.
	//
	// An interface's extends entry parses as a type reference here rather than as an expression,
	// which is the difference this branch exists to bridge.
	entry := types[0]
	if entry.Kind == ast.KindTypeReference {
		typeName := entry.AsTypeReferenceNode().TypeName
		return typeName == nil || typeName.Kind != ast.KindIdentifier || typeName.Text() != "Function"
	}
	if entry.Kind == ast.KindExpressionWithTypeArguments {
		expression := entry.AsExpressionWithTypeArguments().Expression
		return expression == nil || expression.Kind != ast.KindIdentifier ||
			expression.Text() != "Function"
	}
	return true
}

// collectInterfaceThisTypes gathers every `this` in the member, skipping any inside a nested type
// literal.
//
// Upstream keeps a running array and a nesting counter across visitor callbacks; the same answer
// comes out of one walk of the single member, since the array is reset per interface anyway and
// only this member is ever consulted.
func collectInterfaceThisTypes(member *ast.Node) []*ast.Node {
	found := []*ast.Node{}
	var visit func(*ast.Node, int)
	visit = func(node *ast.Node, literalNesting int) {
		if node.Kind == ast.KindThisType && literalNesting == 0 {
			found = append(found, node)
		}
		childNesting := literalNesting
		if node.Kind == ast.KindTypeLiteral {
			childNesting++
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, childNesting)
			return false
		})
	}
	visit(member, 0)
	return found
}

// checkPreferFunctionTypeMember is upstream's `checkMember`, the single body both anchors share.
func checkPreferFunctionTypeMember(
	ctx rule.Context,
	member *ast.Node,
	node *ast.Node,
	thisTypes []*ast.Node,
) {
	returnType := preferFunctionTypeReturnType(member)
	if returnType == nil {
		// A call or construct signature with no annotated return type is not reported. Upstream
		// tests `member.returnType != null` for the same reason: the rewrite needs a colon to turn
		// into an arrow, and without one there is nothing to move.
		return
	}

	isInterface := node.Kind == ast.KindInterfaceDeclaration

	if isInterface && len(thisTypes) > 0 {
		// Pointing at the whole member here would make the message confusing, and reporting every
		// `this` would make it repetitive. Upstream picks the first and says why at the line.
		ctx.ReportNode(thisTypes[0], rule.Message{
			Id: "unexpectedThisOnFunctionOnlyInterface",
			Description: "`this` refers to the function type '" +
				node.AsInterfaceDeclaration().Name().Text() +
				"', did you intend to use a generic `this` parameter like " +
				"`<Self>(this: Self, ...) => Self` instead?",
		})
		return
	}

	literalOrInterface := "Type literal"
	if isInterface {
		literalOrInterface = "Interface"
	}
	message := rule.Message{
		Id: "functionTypeOverCallableType",
		Description: literalOrInterface +
			" only has a call signature, you should use a function type instead.",
	}

	fix, hasFix := preferFunctionTypeFix(ctx, member, node, returnType, isInterface)
	if !hasFix {
		ctx.ReportNode(member, message)
		return
	}
	ctx.ReportNodeWithFixes(member, message, fix)
}

// preferFunctionTypeReturnType answers the member's annotated return type, or nil when the member is
// not a signature the rule cares about.
func preferFunctionTypeReturnType(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindCallSignature:
		return member.AsCallSignatureDeclaration().Type
	case ast.KindConstructSignature:
		return member.AsConstructSignatureDeclaration().Type
	}
	return nil
}

// preferFunctionTypeFix builds the replacement text, or reports that there is none to build.
//
// The second return is false for `export default interface`, which upstream declines because the
// only replacement would be `export default type`, and that is not valid TypeScript.
func preferFunctionTypeFix(
	ctx rule.Context,
	member *ast.Node,
	node *ast.Node,
	returnType *ast.Node,
	isInterface bool,
) (rule.Fix, bool) {
	if isInterface && preferFunctionTypeHasModifier(node, ast.KindDefaultKeyword) {
		return rule.Fix{}, false
	}

	source := ctx.SourceFile.Text()

	// The member's own text, without the leading trivia. Upstream slices from `member.range[0]`,
	// which is the token start; our `Pos()` would include the comment above it and the newline,
	// and the comment is handled separately below.
	memberRange := rule.TokenRange(ctx.SourceFile, member)
	memberText := source[memberRange.Pos():memberRange.End()]

	// The colon that opens the return annotation becomes the arrow. Offsets are relative to the
	// member's own start, as upstream's are.
	colonOffset := rule.TokenRange(ctx.SourceFile, returnType).Pos() - memberRange.Pos() - 1
	for colonOffset > 0 && colonOffset < len(memberText) && memberText[colonOffset] != ':' {
		// Upstream computes this as `returnType.range[0] - start - 1` and relies on the colon
		// sitting exactly one character before the type. That holds when the annotation is written
		// `): T` and not when it is written `) : T` or with a comment between. Walking back to the
		// colon reproduces upstream's intent on the shapes it handles and does not break on the
		// ones its arithmetic would land in the middle of.
		colonOffset--
	}
	if colonOffset <= 0 || colonOffset >= len(memberText) || memberText[colonOffset] != ':' {
		return rule.Fix{}, false
	}

	suggestion := memberText[:colonOffset] + " =>" + memberText[colonOffset+1:]

	// A trailing semicolon belongs to the member inside the braces and has to be re-attached after
	// the whole type alias rather than left in the middle of it.
	lastCharacter := ""
	if strings.HasSuffix(suggestion, ";") {
		lastCharacter = ";"
		suggestion = suggestion[:len(suggestion)-1]
	} else if strings.HasSuffix(suggestion, ",") {
		// A member may be separated by a comma instead, and upstream's fixer only knows about the
		// semicolon. Measured on the installed 8.67.0 build: `type Foo = { (): string, };` is
		// rewritten to `type Foo = () => string,;`, which the TypeScript parser rejects with
		// "';' expected". That is a fix, so the edit engine would apply it with nobody watching.
		//
		// This port reports the case, as upstream does, and declines the repair. Rewriting the
		// comma to a semicolon instead would be the obvious improvement and is deliberately not
		// taken: it would produce a repair upstream never writes, on a shape nothing here has
		// asked for, and a fix that differs from the reference is the thing this port is least
		// able to defend. Reporting without a repair is the subset that can be shown correct.
		return rule.Fix{}, false
	}

	// A function type binds loosely, so it needs parentheses wherever the surrounding type would
	// otherwise swallow the arrow.
	if preferFunctionTypeShouldWrap(node.Parent) {
		suggestion = "(" + suggestion + ")"
	}

	memberComments := preferFunctionTypeCommentsFor(ctx, member, node)

	if isInterface {
		declaration := node.AsInterfaceDeclaration()
		name := declaration.Name()
		if declaration.TypeParameters != nil {
			// The type parameters have to survive, so the alias is rebuilt from the source text
			// spanning the name through the closing angle bracket rather than from the name alone.
			nameStart := rule.TokenRange(ctx.SourceFile, name).Pos()
			// TypeParameters.End() is the last parameter's end, and the closing angle bracket is the
			// next token after it, past any trivia and a trailing comma. Upstream's
			// `node.typeParameters.range[1]` is past the bracket, because estree models the parameter
			// LIST as a node with its own delimiters. Taking the bracket as one byte past End() held
			// for `interface Foo<T>` and broke on a list written one parameter per line, as Prettier
			// writes a long one: `TFlags extends Flags = DefaultFlags,\n>` cut the alias at the comma
			// and the fix did not parse (TS1139, trpc's mutationOptions.ts, #kq9vtva).
			tail := scanner.SkipTrivia(source, declaration.TypeParameters.End())
			if tail < len(source) && source[tail] == ',' {
				tail = scanner.SkipTrivia(source, tail+1)
			}
			if tail >= len(source) || source[tail] != '>' {
				return rule.Fix{}, false
			}
			tail++
			suggestion = "type " + source[nameStart:tail] + " = " + suggestion + lastCharacter
		} else {
			suggestion = "type " + name.Text() + " = " + suggestion + lastCharacter
		}
	}

	// The node being replaced is the whole declaration for an interface, and the literal for a type
	// literal.
	//
	// Upstream replaces `node.range`, which for an exported interface EXCLUDES the `export` because
	// estree puts the interface inside an ExportNamedDeclaration. Here the modifier is part of the
	// interface, so its own token range already starts at `export` and covers exactly what upstream
	// covers plus the keyword upstream leaves in place. The first version of this rule added a
	// modifier adjustment on top of that and deleted the `export` from two of upstream's own
	// outputs; the range needs no adjustment at all.
	nodeRange := rule.TokenRange(ctx.SourceFile, node)

	if isInterface && isExportedInterface(node) {
		// A comment cannot sit between `export` and `type`, so it has to go above the whole
		// statement. Upstream gets that with a second edit inserting before the parent export
		// declaration, because in estree the `export` is that parent and the node being replaced
		// starts at `interface`.
		//
		// Here the `export` is a modifier ON the interface, so the node's own range already covers
		// it and there is no separate parent to insert before. One edit replacing the whole thing
		// with the comment, the keyword, and the alias produces the identical text. The route
		// differs because the trees differ; the output is upstream's, byte for byte, and the
		// fixtures assert it.
		commentText := ""
		for _, comment := range memberComments {
			commentText += comment.Text + "\n"
		}
		suggestion = commentText + "export " + suggestion
	} else {
		// Every other shape keeps the comment immediately before the rewritten text, on its own
		// line when it was on its own line and inline when it was inline.
		// The same helper the comments shelf uses to fill Comment.StartLine, so the two line
		// numbers are computed the same way and are comparable.
		memberLine, _ := scanner.GetECMALineAndByteOffsetOfPosition(ctx.SourceFile, memberRange.Pos())
		for _, comment := range memberComments {
			separator := "\n"
			if comment.StartLine == memberLine {
				separator = " "
			}
			suggestion = comment.Text + separator + suggestion
		}
	}

	return rule.ReplaceRange(core.NewTextRange(nodeRange.Pos(), nodeRange.End()), suggestion), true
}

// preferFunctionTypeCommentsFor collects the comments upstream moves with the member.
//
// Upstream asks for `getCommentsBefore(member)` and `getCommentsAfter(member)`. Here that is every
// comment lying inside the declaration's braces but outside the member's own text, which is the same
// set for a one-member body and is expressible without a token-level comment index.
func preferFunctionTypeCommentsFor(
	ctx rule.Context,
	member *ast.Node,
	node *ast.Node,
) []comments.Comment {
	memberRange := rule.TokenRange(ctx.SourceFile, member)
	// The declaration's own token range, not its Pos: `Pos()` reaches back over leading trivia, so
	// a comment written ABOVE the interface falls inside it. That is not a comment on the member and
	// upstream does not move it; collecting it duplicated a file-leading comment in one of
	// upstream's own outputs.
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	within := []comments.Comment{}
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() < nodeRange.Pos() || comment.Range.End() > nodeRange.End() {
			continue
		}
		if comment.Range.Pos() >= memberRange.Pos() && comment.Range.End() <= memberRange.End() {
			// Inside the signature itself, so it travels with the member text already.
			continue
		}
		within = append(within, comment)
	}
	return within
}

// preferFunctionTypeShouldWrap is upstream's `shouldWrapSuggestion`.
//
// A union, an intersection, and an array type each bind more tightly than an arrow, so
// `{} | { (): void }` has to become `{} | (() => void)` rather than `{} | () => void`, which parses
// as a union inside a return type.
func preferFunctionTypeShouldWrap(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindUnionType, ast.KindIntersectionType, ast.KindArrayType:
		return true
	}
	return false
}

// preferFunctionTypeHasModifier answers whether the node carries one specific modifier keyword.
func preferFunctionTypeHasModifier(node *ast.Node, kind ast.Kind) bool {
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

// isExportedInterface answers whether the declaration carries an `export` modifier.
//
// Upstream asks whether the PARENT is an ExportNamedDeclaration. Here `export` is a modifier on the
// interface itself, so the same question is a modifier test. A port copying the parent test would
// never fire this branch and would leave a comment sitting between `export` and `type`, which does
// not parse.
func isExportedInterface(node *ast.Node) bool {
	return preferFunctionTypeHasModifier(node, ast.KindExportKeyword)
}
