package typescript

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentIndexedObjectStyleOptions is the rule's option surface.
//
// One positional string rather than an object, which is unusual here: upstream's schema is a bare
// enum, so the configured value arrives as `"record"` or `"index-signature"` rather than as a
// keyed object.
type ConsistentIndexedObjectStyleOptions struct {
	// Mode is `record` or `index-signature`.
	Mode string
}

// DefaultConsistentIndexedObjectStyleSettings is upstream's `defaultOptions`.
func DefaultConsistentIndexedObjectStyleSettings() ConsistentIndexedObjectStyleOptions {
	return ConsistentIndexedObjectStyleOptions{Mode: "record"}
}

// DecodeConsistentIndexedObjectStyleOptions reads the rule's configuration.
//
// The wire value is a bare STRING rather than an object, which is why this cannot use
// `rule.DecodeOptionsInto` at all: there is no struct to decode into. Empty input and `null` are a
// rule configured as a bare `"error"`, and get upstream's default.
//
// Anything else is refused, naming the value. This read every value other than the two mode strings
// as the default, an object and a misspelling included, on the reasoning that a typo should not
// disable the rule. It did not disable it; it ran a mode nobody wrote, and the config loaded clean.
// Upstream's schema is a string enum and refuses those at load, and the config layer now refuses a
// decoder's error by rule name, so refusing is what tells the author (#p9s1131).
func DecodeConsistentIndexedObjectStyleOptions(raw []byte) (any, error) {
	options := DefaultConsistentIndexedObjectStyleSettings()

	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return options, nil
	}

	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return options, fmt.Errorf(`expected a mode string, "record" or "index-signature", got %s`, trimmed)
	}
	if mode != "index-signature" && mode != "record" {
		return options, fmt.Errorf(`mode %q is not "record" or "index-signature"`, mode)
	}
	options.Mode = mode
	return options, nil
}

// ConsistentIndexedObjectStyle requires either `Record<K, V>` or an index signature, consistently.
//
//	valid (record mode):   type T = Record<string, number>;
//	valid (record mode):   type T = { [key: string]: number; other: string };
//	valid (record mode):   type Foo = { [key: string]: Foo };
//	invalid (record mode): type T = { [key: string]: number };
//	invalid (index mode):  type T = Record<string, number>;
//
// The two spellings mean the same thing, so the rule is about consistency rather than correctness.
// It is rated stylistic upstream and it is stylistic here.
//
// # The repair rebuilds a type from sub-node text, which is where the danger is
//
// This project has twice shipped a fixer that silently destroyed type information, both times
// because the repair reconstructed a span rather than deleting one. This rule reconstructs by
// design, so every modifier that can live inside the replaced span is enumerated and carried:
//
//	an index signature's `readonly`   becomes `Readonly<Record<K, V>>`
//	a mapped type's `readonly` / `+readonly`   becomes `Readonly<...>`
//	a mapped type's `?` / `+?`   becomes `Partial<...>`
//	a mapped type's `-?`   becomes `Required<...>`
//	a mapped type's `-readonly`   has NO Record equivalent, so the repair is DECLINED
//	an interface's type parameters   are carried onto the generated type alias
//	an interface's `export` / `declare`   stay OUTSIDE the replaced span, so they survive
//
// That last decline is upstream's and it is the shape this brief keeps asking for: there is no
// builtin `Mutable<T>`, so a minus-readonly mapped type cannot be written as a Record at all.
// Reporting without a repair is the subset that can be shown correct.
//
// # A comment inside the replaced span downgrades the fix to a SUGGESTION
//
// The repair reads only the key type and the value type, so a comment sitting anywhere else inside
// the node would be dropped by the rewrite. Upstream detects that and offers the change rather than
// applying it, which is a decision worth reproducing exactly: an unattended fix that deletes a
// comment is the same class of loss as one that deletes a type.
//
// # Circularity, and what makes a self-reference safe to leave alone
//
// `type Foo = { [key: string]: Foo }` cannot become a Record, because the alias would reference
// itself through a type argument rather than through a member and TypeScript rejects it. Upstream
// answers this with `isDeeplyReferencingType`, walking its scope manager's definitions.
//
// This port resolves identifiers through the CHECKER instead and compares symbols against the
// enclosing declaration's own symbol, which answers the same question without a scope manager.
// Probed before the rule was written, including the case that separates it from a naive name
// match: `namespace A { export type Foo = 1 }` beside `type Foo = { [k: string]: A.Foo }` is NOT
// circular, because `A.Foo` resolves to a different symbol, and upstream reports it.
//
// # Cost
//
// Four anchors, all of them type nodes, which are rare outside declaration files. The checker is
// consulted only after a node has already been shown to be a single-member index signature or a
// convertible mapped type.
var ConsistentIndexedObjectStyle = rule.Rule{
	Name: "@typescript-eslint/consistent-indexed-object-style",

	// The circularity test resolves an identifier to its declaration.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[ConsistentIndexedObjectStyleOptions](options)
		if !isSettings {
			settings = DefaultConsistentIndexedObjectStyleSettings()
		}

		textOf := func(node *ast.Node) string {
			if node == nil {
				return ""
			}
			nodeRange := rule.TokenRange(ctx.SourceFile, node)
			return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
		}

		// hasUnpreservedComments answers upstream's helper of the same name: is there a comment
		// inside the replaced range that does not sit inside one of the sub-nodes the repair copies
		// out?
		//
		// It takes the RANGE the repair replaces rather than a node, because for an interface those
		// differ: the modifiers sit inside the node and outside the replacement, so a comment between
		// `export` and `interface` survives the rewrite and must not downgrade it.
		//
		// A comment that survives the rewrite is one wholly contained in a preserved sub-node, so
		// the test is "no preserved range contains it".
		//
		// The comment ranges are scanned out of the source text here rather than read from the
		// shelf's `comments.ForFile`, and that is a measured decision rather than a preference.
		// Probed on upstream's own case, a line comment sitting between a mapped type's opening
		// brace and its key: `comments.ForFile` returned ZERO comments for it. That helper anchors
		// its scan on node positions and documents a known family of gaps around empty lists; this
		// comment falls in a different gap, before the key of a mapped type. The question here is
		// narrower than the one that helper answers, being only "is any byte in this range inside a
		// comment", so it is answered directly.
		hasUnpreservedComments := func(nodeRange core.TextRange, preserved ...*ast.Node) bool {
			text := ctx.SourceFile.Text()

			survives := func(start int, end int) bool {
				for _, target := range preserved {
					if target == nil {
						continue
					}
					targetRange := rule.TokenRange(ctx.SourceFile, target)
					if start >= targetRange.Pos() && end <= targetRange.End() {
						return true
					}
				}
				return false
			}

			for index := nodeRange.Pos(); index < nodeRange.End() && index < len(text); {
				after := scanner.SkipTrivia(text, index)
				if after > index {
					// SkipTrivia consumed whitespace and comments together, so the run is rescanned
					// to find where a comment actually starts. Only a comment matters here; skipped
					// whitespace is not something the rewrite could drop.
					for position := index; position < after && position+1 < len(text); position++ {
						if text[position] != '/' {
							continue
						}
						if text[position+1] != '/' && text[position+1] != '*' {
							continue
						}
						commentEnd := after
						if commentEnd > nodeRange.End() {
							commentEnd = nodeRange.End()
						}
						if !survives(position, commentEnd) {
							return true
						}
						break
					}
					index = after
					continue
				}
				index++
			}
			return false
		}

		// referencesDeclaration answers upstream's `isDeeplyReferencingType`.
		//
		// The question is not "does this subtree mention the name" but "can this type be written as
		// a Record without the alias referring to itself through a type argument". Those differ, and
		// the difference is most of what this predicate is:
		//
		//	interface Foo { [k: string]: Foo }        circular, silent
		//	interface Foo { [k: string]: Foo[] }      NOT circular, reports
		//	interface Foo { [k: string]: () => Foo }  NOT circular, reports
		//	interface Foo { [k: string]: { foo: Foo } }   NOT circular, reports
		//
		// An array, a function type and a nested object literal all put the self-reference behind a
		// constructor that a Record can legally hold, so the rewrite compiles. A union, an
		// intersection, a conditional or a bare alias hop does not, so those are followed.
		//
		// The traversal therefore recurses only through the TYPE-TRANSPARENT kinds upstream lists
		// and stops everywhere else, rather than walking every child. A walk over every child was
		// the first version of this and it got eight of upstream's cases wrong in both directions at
		// once: it missed `Foo1` to `Foo2` to `Foo1`, and it wrongly silenced all four wrapper
		// shapes above.
		//
		// Indirection is followed by resolving an identifier to its declaration through the checker,
		// which is what replaces upstream's scope-manager definition lookup. `visited` guards the
		// cycle that following definitions necessarily creates.
		var referencesDeclarationFrom func(node *ast.Node, anchor *ast.Symbol, visited map[*ast.Node]bool) bool
		referencesDeclarationFrom = func(node *ast.Node, anchor *ast.Symbol, visited map[*ast.Node]bool) bool {
			if node == nil || visited[node] {
				return false
			}
			visited[node] = true

			recurse := func(child *ast.Node) bool {
				return referencesDeclarationFrom(child, anchor, visited)
			}

			switch node.Kind {
			case ast.KindIdentifier:
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return false
				}
				if symbol == anchor {
					return true
				}
				// Not the anchor, so follow what this name is declared as. That is the hop upstream
				// makes through `refVar.defs`, and it is what makes a two-interface cycle circular.
				for _, declaration := range symbol.Declarations {
					if recurse(declaration) {
						return true
					}
				}
				return false

			case ast.KindTypeAliasDeclaration:
				return recurse(node.AsTypeAliasDeclaration().Type)

			case ast.KindInterfaceDeclaration:
				declaration := node.AsInterfaceDeclaration()
				if declaration.Members == nil {
					return false
				}
				for _, member := range declaration.Members.Nodes {
					if recurse(member) {
						return true
					}
				}
				return false

			case ast.KindTypeLiteral:
				literal := node.AsTypeLiteralNode()
				if literal.Members == nil {
					return false
				}
				for _, member := range literal.Members.Nodes {
					if recurse(member) {
						return true
					}
				}
				return false

			case ast.KindIndexSignature:
				return recurse(node.AsIndexSignatureDeclaration().Type)

			case ast.KindUnionType:
				for _, constituent := range node.AsUnionTypeNode().Types.Nodes {
					if recurse(constituent) {
						return true
					}
				}
				return false

			case ast.KindIntersectionType:
				for _, constituent := range node.AsIntersectionTypeNode().Types.Nodes {
					if recurse(constituent) {
						return true
					}
				}
				return false

			case ast.KindConditionalType:
				conditional := node.AsConditionalTypeNode()
				return recurse(conditional.CheckType) || recurse(conditional.ExtendsType) ||
					recurse(conditional.TrueType) || recurse(conditional.FalseType)

			case ast.KindIndexedAccessType:
				indexed := node.AsIndexedAccessTypeNode()
				return recurse(indexed.ObjectType) || recurse(indexed.IndexType)

			case ast.KindMappedType:
				return recurse(node.AsMappedTypeNode().Type)

			case ast.KindParenthesizedType:
				return recurse(node.AsParenthesizedTypeNode().Type)

			case ast.KindTypeReference:
				reference := node.AsTypeReferenceNode()
				if recurse(reference.TypeName) {
					return true
				}
				if reference.TypeArguments != nil {
					for _, argument := range reference.TypeArguments.Nodes {
						if recurse(argument) {
							return true
						}
					}
				}
				return false
			}

			// Every other kind is opaque: an array, a function type, a tuple and a property
			// signature all hold their operand behind a constructor a Record can carry, so a
			// self-reference inside one is not circular. Upstream reaches the same verdict by simply
			// not naming those kinds in its switch.
			return false
		}

		referencesDeclaration := func(body *ast.Node, declarationName *ast.Node) bool {
			if ctx.TypeChecker == nil || declarationName == nil || body == nil {
				return false
			}
			anchor := ctx.TypeChecker.GetSymbolAtLocation(declarationName)
			if anchor == nil {
				return false
			}
			// The anchoring declaration is pre-visited so that following the anchor's own name back
			// to its declaration does not read as a reference to itself.
			return referencesDeclarationFrom(body, anchor, map[*ast.Node]bool{declarationName: true})
		}

		// reportIndexSignature is upstream's `checkMembers`: one member, an index signature, whose
		// key parameter carries a type.
		//
		// `prefix` and `postfix` wrap the generated Record so that an interface becomes a whole
		// type alias while a type literal is replaced in place.
		//
		// `replacedRange` is both what the finding underlines and what the repair overwrites, and for
		// an interface it starts at the `interface` keyword, AFTER the modifiers. That is what keeps
		// `export` and `declare`: the repair rebuilds only `type <name><generics> = ...`, so it must
		// replace only the text that spelling covers. Replacing from the first modifier is how this
		// fixer once turned `export interface X` into `type X` and broke six imports.
		reportIndexSignature := func(members []*ast.Node, node *ast.Node, declarationName *ast.Node,
			prefix string, postfix string, repairIsSafe bool, replacedRange core.TextRange) {
			if len(members) != 1 {
				return
			}
			member := members[0]
			if member.Kind != ast.KindIndexSignature {
				return
			}
			signature := member.AsIndexSignatureDeclaration()
			if signature.Parameters == nil || len(signature.Parameters.Nodes) != 1 {
				return
			}
			parameter := signature.Parameters.Nodes[0]
			if parameter.Kind != ast.KindParameter {
				return
			}
			keyType := parameter.AsParameterDeclaration().Type
			valueType := signature.Type
			if keyType == nil || valueType == nil {
				return
			}

			if declarationName != nil && referencesDeclaration(node, declarationName) {
				return
			}

			record := "Record<" + textOf(keyType) + ", " + textOf(valueType) + ">"
			if indexSignatureIsReadonly(signature) {
				// The `readonly` modifier has a Record spelling, so it is carried rather than
				// dropped. This is the enumeration the two fixers that destroyed type information
				// in this project did not do.
				record = "Readonly<" + record + ">"
			}
			replacement := prefix + record + postfix

			if !repairIsSafe {
				ctx.ReportRange(replacedRange, buildPreferRecordMessage())
				return
			}
			fix := rule.ReplaceRange(replacedRange, replacement)
			if hasUnpreservedComments(replacedRange, keyType, valueType) {
				ctx.ReportRangeWithSuggestions(replacedRange, buildPreferRecordMessage(), rule.Suggestion{
					Message: buildPreferRecordSuggestionMessage(),
					Fixes:   []rule.Fix{fix},
				})
				return
			}
			ctx.ReportRangeWithFixes(replacedRange, buildPreferRecordMessage(), fix)
		}

		listeners := rule.Listeners{}

		if settings.Mode == "index-signature" {
			listeners[ast.KindTypeReference] = func(node *ast.Node) {
				reference := node.AsTypeReferenceNode()
				if reference.TypeName == nil || reference.TypeName.Kind != ast.KindIdentifier {
					return
				}
				if reference.TypeName.Text() != "Record" {
					return
				}
				if reference.TypeArguments == nil || len(reference.TypeArguments.Nodes) != 2 {
					return
				}
				keyArgument := reference.TypeArguments.Nodes[0]
				valueArgument := reference.TypeArguments.Nodes[1]

				// Only the three primitive key kinds can be written as an index signature; anything
				// else is offered as a suggestion rather than applied, because the rewrite would not
				// compile.
				repairIsSafe := keyArgument.Kind == ast.KindStringKeyword ||
					keyArgument.Kind == ast.KindNumberKeyword ||
					keyArgument.Kind == ast.KindSymbolKeyword

				replacement := "{ [key: " + textOf(keyArgument) + "]: " + textOf(valueArgument) + " }"
				fix := rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement)

				if repairIsSafe && !hasUnpreservedComments(rule.TokenRange(ctx.SourceFile, node), keyArgument, valueArgument) {
					ctx.ReportNodeWithFixes(node, buildPreferIndexSignatureMessage(), fix)
					return
				}
				ctx.ReportNodeWithSuggestions(node, buildPreferIndexSignatureMessage(), rule.Suggestion{
					Message: buildPreferIndexSignatureSuggestionMessage(),
					Fixes:   []rule.Fix{fix},
				})
			}
			return listeners
		}

		// interfaceReplacedRange is the interface's span WITHOUT its leading modifiers.
		//
		// `rule.TokenRange` starts at the first token, which for `export interface Foo` is `export`.
		// Upstream's node begins at the `interface` keyword, because estree hangs the export off a
		// separate wrapper node, so both its finding and its repair leave `export` alone. The one
		// corpus case that shows the span is an export-default interface, whose finding underlines
		// only the interface; the repair inherits the same start, which is what carries the
		// modifiers. Upstream's estree does put `declare` inside its node and drops it; this keeps it,
		// since `declare type` is equally valid and the rewrite has no reason to touch it.
		interfaceReplacedRange := func(node *ast.Node) core.TextRange {
			nodeRange := rule.TokenRange(ctx.SourceFile, node)
			if node.Modifiers() != nil && len(node.Modifiers().Nodes) != 0 {
				last := node.Modifiers().Nodes[len(node.Modifiers().Nodes)-1]
				return core.NewTextRange(
					scanner.SkipTrivia(ctx.SourceFile.Text(), rule.TokenRange(ctx.SourceFile, last).End()),
					nodeRange.End())
			}
			return nodeRange
		}

		listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
			declaration := node.AsInterfaceDeclaration()
			if declaration.Members == nil {
				return
			}

			// Type parameters are carried onto the generated alias. Dropping them would turn
			// `interface Foo<T> { [k: string]: T }` into a type alias naming an unbound `T`.
			generics := ""
			if declaration.TypeParameters != nil && len(declaration.TypeParameters.Nodes) != 0 {
				rendered := make([]string, 0, len(declaration.TypeParameters.Nodes))
				for _, parameter := range declaration.TypeParameters.Nodes {
					rendered = append(rendered, textOf(parameter))
				}
				generics = "<" + strings.Join(rendered, ", ") + ">"
			}

			// An interface that extends something, or that is the subject of an export-default,
			// cannot be rewritten as a type alias without changing what the file means, so the
			// finding is reported with no repair at all.
			repairIsSafe := !interfaceHasHeritage(declaration) && !isExportDefaultSubject(node)

			name := declaration.Name()
			if name == nil {
				return
			}
			reportIndexSignature(declaration.Members.Nodes, node, name,
				"type "+name.Text()+generics+" = ", ";", repairIsSafe, interfaceReplacedRange(node))
		}

		listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
			literal := node.AsTypeLiteralNode()
			if literal.Members == nil {
				return
			}
			var declarationName *ast.Node
			if parent := enclosingTypeAliasDeclaration(node); parent != nil {
				declarationName = parent.AsTypeAliasDeclaration().Name()
			}
			reportIndexSignature(literal.Members.Nodes, node, declarationName, "", "", true,
				rule.TokenRange(ctx.SourceFile, node))
		}

		listeners[ast.KindMappedType] = func(node *ast.Node) {
			mapped := node.AsMappedTypeNode()
			if mapped.TypeParameter == nil {
				return
			}
			constraint := mapped.TypeParameter.AsTypeParameterDeclaration().Constraint
			if constraint == nil {
				return
			}

			// A `keyof` constraint preserves the source type's own modifiers through the mapped
			// type, which a Record does not, so the two are not equivalent and upstream declines.
			//
			// The exemption applies ONLY to an unparenthesized keyof. Upstream writes that as
			// `!isParenthesized(constraint)`, and it looks like an oversight until you see the case
			// built for it: `{ [k in (keyof ParseResult)]: unknown }` REPORTS and repairs to
			// `Record<keyof ParseResult, unknown>`. Our parser gives the parenthesized form a
			// `KindParenthesizedType` wrapper, so the operator test simply does not match it, which
			// reproduces upstream's behavior without a separate parenthesis check.
			if constraint.Kind == ast.KindTypeOperator &&
				constraint.AsTypeOperatorNode().Operator == ast.KindKeyOfKeyword {
				return
			}

			// If the key is used to compute the value, the mapping is not uniform and no Record can
			// express it.
			if mappedKeyIsUsedInValue(ctx, mapped) {
				return
			}

			if parent := enclosingTypeAliasDeclaration(node); parent != nil {
				if referencesDeclaration(node, parent.AsTypeAliasDeclaration().Name()) {
					return
				}
			}

			valueText := "any"
			if mapped.Type != nil {
				valueText = textOf(mapped.Type)
			}
			// The key text is read from INSIDE any parentheses. estree has no parenthesized type
			// node, so upstream's `getText(constraint)` on `(keyof P)` yields `keyof P`, and the
			// repair it records is `Record<keyof ParseResult, unknown>` rather than one carrying the
			// brackets. Measured on the one corpus case that writes a parenthesized constraint.
			keyText := textOf(unwrapParenthesizedType(constraint))
			record := "Record<" + keyText + ", " + valueText + ">"

			// The modifiers, in upstream's order: optionality wraps first, then readonly.
			if mapped.QuestionToken != nil {
				switch mapped.QuestionToken.Kind {
				case ast.KindMinusToken:
					record = "Required<" + record + ">"
				default:
					record = "Partial<" + record + ">"
				}
			}
			declineRepair := false
			if mapped.ReadonlyToken != nil {
				if mapped.ReadonlyToken.Kind == ast.KindMinusToken {
					// There is no builtin `Mutable<T>`, so a minus-readonly mapped type has no
					// Record spelling. Upstream reports it and offers nothing, and so does this.
					declineRepair = true
				} else {
					record = "Readonly<" + record + ">"
				}
			}

			if declineRepair {
				ctx.ReportNode(node, buildPreferRecordMessage())
				return
			}
			fix := rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), record)
			if hasUnpreservedComments(rule.TokenRange(ctx.SourceFile, node), constraint, mapped.Type) {
				ctx.ReportNodeWithSuggestions(node, buildPreferRecordMessage(), rule.Suggestion{
					Message: buildPreferRecordSuggestionMessage(),
					Fixes:   []rule.Fix{fix},
				})
				return
			}
			ctx.ReportNodeWithFixes(node, buildPreferRecordMessage(), fix)
		}

		return listeners
	},
}

// indexSignatureIsReadonly reports whether an index signature carries the `readonly` modifier.
func indexSignatureIsReadonly(signature *ast.IndexSignatureDeclaration) bool {
	if signature.Modifiers() == nil {
		return false
	}
	for _, modifier := range signature.Modifiers().Nodes {
		if modifier.Kind == ast.KindReadonlyKeyword {
			return true
		}
	}
	return false
}

// interfaceHasHeritage reports whether an interface extends anything.
func interfaceHasHeritage(declaration *ast.InterfaceDeclaration) bool {
	return declaration.HeritageClauses != nil && len(declaration.HeritageClauses.Nodes) != 0
}

// isExportDefaultSubject reports whether a declaration is the subject of `export default`.
func isExportDefaultSubject(node *ast.Node) bool {
	if node.Modifiers() == nil {
		return false
	}
	hasExport, hasDefault := false, false
	for _, modifier := range node.Modifiers().Nodes {
		switch modifier.Kind {
		case ast.KindExportKeyword:
			hasExport = true
		case ast.KindDefaultKeyword:
			hasDefault = true
		}
	}
	return hasExport && hasDefault
}

// enclosingTypeAliasDeclaration answers upstream's `findParentDeclaration`.
//
// The walk stops at a type annotation, matching upstream: a type literal used as an annotation
// somewhere inside an alias is not the alias's own body and cannot be circular with it.
func enclosingTypeAliasDeclaration(node *ast.Node) *ast.Node {
	child := node
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindTypeAliasDeclaration {
			return current
		}

		// Upstream stops the walk at a type ANNOTATION, and that stop is load-bearing rather than
		// tidiness. estree gives an index signature's value a `TSTypeAnnotation` wrapper, so for
		// `type Foo = { [k: string]: { [k: string]: Foo } }` the inner literal's walk halts there,
		// finds no declaration, runs no circularity check, and REPORTS. The outer literal does find
		// `Foo`, is circular, and stays silent. One input, two nested nodes, opposite verdicts.
		//
		// Our tree has no annotation wrapper: the value type is the index signature's own `Type`
		// field. So the equivalent stop is "I arrived here as the type of something", which is what
		// the checks below express. Without it the inner literal inherits the outer alias and this
		// case goes silent, which is how the rule was written first.
		switch current.Kind {
		case ast.KindIndexSignature:
			if current.AsIndexSignatureDeclaration().Type == child {
				return nil
			}
		case ast.KindPropertySignature:
			if current.AsPropertySignatureDeclaration().Type == child {
				return nil
			}
		case ast.KindParameter, ast.KindPropertyDeclaration, ast.KindVariableDeclaration:
			return nil
		}

		child = current
		current = current.Parent
	}
	return nil
}

// mappedKeyIsUsedInValue reports whether the mapped type's key appears in its value type.
//
// `{ [K in Keys]: Foo[K] }` maps each key to a different type, which no Record can express, so
// upstream declines. It asks its scope manager whether the key variable has any type reference;
// here the value subtree is walked for an identifier spelled the same, which answers the same
// question because a mapped type's key is bound to exactly that subtree.
func mappedKeyIsUsedInValue(ctx rule.Context, mapped *ast.MappedTypeNode) bool {
	if mapped.Type == nil || mapped.TypeParameter == nil {
		return false
	}
	keyName := mapped.TypeParameter.AsTypeParameterDeclaration().Name()
	if keyName == nil {
		return false
	}
	wanted := keyName.Text()
	found := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node.Kind == ast.KindIdentifier && node.Text() == wanted {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(mapped.Type)
	return found
}

// unwrapParenthesizedType steps through parenthesized type wrappers.
//
// Written as a loop with its own nil check rather than a helper call: `ast.SkipParentheses` is for
// expressions and dereferences its argument, and `((T))` nests so one step would not be enough.
func unwrapParenthesizedType(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Kind == ast.KindParenthesizedType {
		inner := current.AsParenthesizedTypeNode().Type
		if inner == nil {
			return current
		}
		current = inner
	}
	return current
}

func buildPreferRecordMessage() rule.Message {
	return rule.Message{Id: "preferRecord", Description: "A record is preferred over an index signature."}
}

func buildPreferRecordSuggestionMessage() rule.Message {
	return rule.Message{Id: "preferRecordSuggestion", Description: "Change into a record instead of an index signature."}
}

func buildPreferIndexSignatureMessage() rule.Message {
	return rule.Message{Id: "preferIndexSignature", Description: "An index signature is preferred over a record."}
}

func buildPreferIndexSignatureSuggestionMessage() rule.Message {
	return rule.Message{Id: "preferIndexSignatureSuggestion", Description: "Change into an index signature instead of a record."}
}
