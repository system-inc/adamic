package typescript

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRedundantTypeConstituents flags union and intersection members the others swallow.
//
//	valid:   type T = string | number;
//	valid:   type T = boolean | string;
//	valid:   function f(): never | string;      never is meaningful in a return position
//	invalid: type T = string | any;             any overrides everything
//	invalid: type T = number | never;           never contributes nothing to a union
//	invalid: type T = string | 'literal';       the literal is already a string
//	invalid: type T = string & 'literal';       here the PRIMITIVE is the redundant one
//
// The whole question is subtype relationships the checker has already computed, so the rule reads
// type flags rather than reasoning about syntax.
//
// # The absorption is directional, and the direction flips between the two containers
//
// In a UNION the wider type wins, so `'a' | string` is `string` and the LITERAL is reported. In an
// INTERSECTION the narrower type wins, so `'a' & string` is `'a'` and the PRIMITIVE is reported.
// Same pair of members, opposite finding. Getting this backwards names the wrong member while
// reading entirely plausible, which is why the two arms are written separately below rather than
// shared behind a flag.
//
// # Six combinations of the three special types, not three plus symmetry
//
// `any`, `never` and `unknown` each behave differently in each container, and upstream handles all
// six. Reproduced rather than derived:
//
//	union         any overrides       unknown overrides     never is overridden
//	intersection  any overrides       unknown is overridden never overrides
//
// The `never` union case carries one further exception: inside a function's return annotation
// `never | string` is meaningful rather than redundant, so it is not reported there.
//
// # Our parser inserts a node upstream never sees, and it changes the walk
//
// estree has no parenthesized type, so upstream's handler receives `(string | any) | number` as two
// members whose first has a union TYPE. Ours inserts a `KindParenthesizedType` between, and without
// unwrapping it the member reads as opaque and the finding is lost.
//
// Measured against the installed 8.67.0 build before this rule was written, because the corpus
// writes almost no parenthesized form and so has no opinion:
//
//	(string | any) | number     TWO findings, spans `string | any` and `any`
//	string | any | number       ONE finding, span `any`
//	((string | any)) | number   TWO findings, same spans, so the unwrap must loop
//	(any)                       no union node at all, so nothing fires
//	(string) | (any)            ONE finding, span `any`
//
// The two findings in the first case come from two different listener calls: the outer union
// reports the member whose type is a union, and the inner union reports `any` on its own. So the
// unwrap is for member IDENTITY only and the spans fall out; a probe predicting exactly that agreed
// with all five measured shapes, and disabling the unwrap took it to two of five.
//
// # `boolean` is stored as a union and must not be split
//
// TypeScript represents `boolean` as `false | true` internally, so splitting every union-flagged
// type would make `boolean | string` look like three members and report nothing sensible. Upstream
// has `unionTypePartsUnlessBoolean` for exactly this and it is reproduced. Probed here: `boolean`
// carries the Union flag, and so does a type alias resolving to it.
//
// # Cost
//
// Two anchors, both uncommon outside type-heavy files, and each member's flags are cached per node
// because a type reference resolving to a large union would otherwise be re-read once per sibling.
var NoRedundantTypeConstituents = rule.Rule{
	Name: "@typescript-eslint/no-redundant-type-constituents",

	// Every judgment is a type-flag question about a member the checker has already resolved.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The declaration governs registration; the harness can still build a Context by hand, and
		// a nil dereference here costs every rule its verdict on the whole file.
		if ctx.TypeChecker == nil {
			return nil
		}

		cache := map[*ast.Node][]redundantTypePart{}

		return rule.Listeners{
			ast.KindUnionType: func(node *ast.Node) {
				checkRedundantUnion(ctx, node, cache)
			},
			ast.KindIntersectionType: func(node *ast.Node) {
				checkRedundantIntersection(ctx, node, cache)
			},
		}
	},
}

// redundantTypePart is one resolved member: the flags that decide absorption and the name the
// message prints.
type redundantTypePart struct {
	flags    checker.TypeFlags
	typeName string
}

// primitiveOf maps a literal flag onto the primitive that absorbs it in a union.
//
// A template literal type absorbs into `string`, which is upstream's table and is not obvious: a
// template literal is a string at the type level however many interpolations it carries.
var redundantPrimitiveOf = map[checker.TypeFlags]checker.TypeFlags{
	checker.TypeFlagsBigIntLiteral:   checker.TypeFlagsBigInt,
	checker.TypeFlagsBooleanLiteral:  checker.TypeFlagsBoolean,
	checker.TypeFlagsNumberLiteral:   checker.TypeFlagsNumber,
	checker.TypeFlagsStringLiteral:   checker.TypeFlagsString,
	checker.TypeFlagsTemplateLiteral: checker.TypeFlagsString,
}

// redundantPrimitiveNames is how each primitive is spelled in a message.
var redundantPrimitiveNames = map[checker.TypeFlags]string{
	checker.TypeFlagsBigInt:  "bigint",
	checker.TypeFlagsBoolean: "boolean",
	checker.TypeFlagsNumber:  "number",
	checker.TypeFlagsString:  "string",
}

// redundantPrimitiveFlags is the iteration order for primitives, fixed so a message listing several
// is deterministic. Go map order is randomised and a rule whose output depends on it produces a
// test that passes most of the time, which is worse than one that fails.
var redundantPrimitiveFlags = []checker.TypeFlags{
	checker.TypeFlagsBigInt,
	checker.TypeFlagsBoolean,
	checker.TypeFlagsNumber,
	checker.TypeFlagsString,
}

// redundantLiteralFlags is the matching order for literal flags.
var redundantLiteralFlags = []checker.TypeFlags{
	checker.TypeFlagsBigIntLiteral,
	checker.TypeFlagsBooleanLiteral,
	checker.TypeFlagsNumberLiteral,
	checker.TypeFlagsStringLiteral,
	checker.TypeFlagsTemplateLiteral,
}

// This rule uses `unwrapParenthesizedType` from consistent_indexed_object_style.go in this package.
//
// It is the same decision: loop rather than one step because `((T))` nests, and its own nil check
// rather than `ast.SkipParentheses`, which is for expressions and dereferences its argument. It is
// also slightly safer than the version this rule first wrote, returning the wrapper rather than nil
// when a malformed inner type is missing. A package is one namespace, so a second copy would not
// have compiled, and the collision is what surfaced the existing one.

// redundantKeywordFlags is upstream's keywordNodeTypesToTsTypes: the flags a bare keyword carries
// without asking the checker.
//
// Reading these syntactically rather than through the checker is upstream's choice and it matters:
// the checker widens some of them in context, and the rule wants what the author wrote.
func redundantKeywordFlags(node *ast.Node) (checker.TypeFlags, bool) {
	switch node.Kind {
	case ast.KindAnyKeyword:
		return checker.TypeFlagsAny, true
	case ast.KindBigIntKeyword:
		return checker.TypeFlagsBigInt, true
	case ast.KindBooleanKeyword:
		return checker.TypeFlagsBoolean, true
	case ast.KindNeverKeyword:
		return checker.TypeFlagsNever, true
	case ast.KindNumberKeyword:
		return checker.TypeFlagsNumber, true
	case ast.KindStringKeyword:
		return checker.TypeFlagsString, true
	case ast.KindUnknownKeyword:
		return checker.TypeFlagsUnknown, true
	}
	return 0, false
}

// typePartsOf resolves one member into the parts that participate in absorption.
//
// Cached per node, because a member that is a type reference to a large union would otherwise be
// re-resolved once for every sibling it is compared against.
func typePartsOf(ctx rule.Context, node *ast.Node, cache map[*ast.Node][]redundantTypePart) []redundantTypePart {
	if existing, seen := cache[node]; seen {
		return existing
	}
	computed := computeTypeParts(ctx, node)
	cache[node] = computed
	return computed
}

// computeTypeParts is upstream's getTypeNodeTypePartFlags.
func computeTypeParts(ctx rule.Context, node *ast.Node) []redundantTypePart {
	node = unwrapParenthesizedType(node)
	if node == nil {
		return nil
	}

	if flags, isKeyword := redundantKeywordFlags(node); isKeyword {
		return []redundantTypePart{{flags: flags, typeName: describeTypeNodeName(ctx, node)}}
	}

	// A nested union contributes every member, which is what makes `(a | b) | c` behave like
	// `a | b | c` for absorption while still reporting at the inner node.
	if node.Kind == ast.KindUnionType {
		parts := []redundantTypePart{}
		for _, member := range node.AsUnionTypeNode().Types.Nodes {
			parts = append(parts, computeTypeParts(ctx, member)...)
		}
		return parts
	}

	if node.Kind == ast.KindLiteralType {
		literal := node.AsLiteralTypeNode().Literal
		if literal != nil {
			switch literal.Kind {
			case ast.KindStringLiteral:
				return []redundantTypePart{{flags: checker.TypeFlagsStringLiteral,
					typeName: describeTypeNodeName(ctx, node)}}
			case ast.KindNumericLiteral:
				return []redundantTypePart{{flags: checker.TypeFlagsNumberLiteral,
					typeName: describeTypeNodeName(ctx, node)}}
			case ast.KindBigIntLiteral:
				return []redundantTypePart{{flags: checker.TypeFlagsBigIntLiteral,
					typeName: describeTypeNodeName(ctx, node)}}
			case ast.KindTrueKeyword, ast.KindFalseKeyword:
				return []redundantTypePart{{flags: checker.TypeFlagsBooleanLiteral,
					typeName: describeTypeNodeName(ctx, node)}}
			}
		}
	}

	// Anything else is asked of the checker and split into its union members, except that `boolean`
	// is stored as `false | true` and must stay whole.
	resolved := ctx.TypeChecker.GetTypeAtLocation(node)
	if resolved == nil {
		return nil
	}
	parts := []redundantTypePart{}
	for _, part := range unionPartsUnlessBoolean(resolved) {
		parts = append(parts, redundantTypePart{
			flags:    part.Flags(),
			typeName: describeResolvedType(ctx, part),
		})
	}
	return parts
}

// unionPartsUnlessBoolean splits a union into members, keeping `boolean` whole.
//
// TypeScript stores `boolean` as exactly the two-member union `false | true`, so a naive split turns
// one member into two literal members and the absorption reads backwards. Upstream detects the
// shape by testing for exactly two members that are the false and true literals.
func unionPartsUnlessBoolean(subject *checker.Type) []*checker.Type {
	if !type_checking.IsUnionType(subject) {
		return []*checker.Type{subject}
	}
	members := subject.Types()
	// Matched by FLAG rather than through the shelf's true/false predicates. Those require an
	// INTRINSIC type, and a boolean reached through an alias is not one: measured on
	// `type B = boolean; type T = B | false`, both members report the BooleanLiteral flag while
	// IsFalseLiteralType answers false for each. With the predicate version the guard never fired,
	// `B` split into two literals, no primitive was ever seen, and the case reported nothing.
	//
	// Testing the flag is also enough to identify the shape: a two-member union of boolean literals
	// is `boolean` and nothing else, since `false | false` collapses and `true | false` is
	// normalised to the same type.
	if len(members) == 2 &&
		type_checking.IsTypeFlagSet(members[0], checker.TypeFlagsBooleanLiteral) &&
		type_checking.IsTypeFlagSet(members[1], checker.TypeFlagsBooleanLiteral) {
		return []*checker.Type{subject}
	}
	return members
}

// describeTypeNodeName renders a member written syntactically, which is upstream's
// describeLiteralTypeNode.
//
// A string literal is printed with double quotes whatever the source used, because upstream renders
// through JSON stringification. `'a'` therefore appears in the message as `"a"`.
func describeTypeNodeName(ctx rule.Context, node *ast.Node) string {
	switch node.Kind {
	case ast.KindAnyKeyword:
		return "any"
	case ast.KindBooleanKeyword:
		return "boolean"
	case ast.KindNeverKeyword:
		return "never"
	case ast.KindNumberKeyword:
		return "number"
	case ast.KindStringKeyword:
		return "string"
	case ast.KindUnknownKeyword:
		return "unknown"
	case ast.KindLiteralType:
		literal := node.AsLiteralTypeNode().Literal
		if literal == nil {
			return "literal type"
		}
		switch literal.Kind {
		case ast.KindStringLiteral:
			return quoteForMessage(literal.Text())
		case ast.KindNumericLiteral, ast.KindBigIntLiteral:
			nodeRange := rule.TokenRange(ctx.SourceFile, literal)
			return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
		case ast.KindTrueKeyword:
			return "true"
		case ast.KindFalseKeyword:
			return "false"
		}
	case ast.KindTemplateLiteralType:
		return "template literal type"
	}
	return "literal type"
}

// describeResolvedType renders a member the checker resolved, which is upstream's describeLiteralType.
func describeResolvedType(ctx rule.Context, subject *checker.Type) string {
	// The ERROR type first, and this ordering is load-bearing rather than tidy. An unresolvable
	// name carries the `any` flag as well as being the error type, so an `IsTypeAnyType` test
	// placed above this one answers "any" and the caller's `typeName != "any"` check never fires,
	// which silently downgrades `errorTypeOverrides` to `overrides`. Measured on `NotKnown | 0`:
	// any=true AND errorType=true on the same type.
	if type_checking.IsIntrinsicErrorType(subject) {
		return ctx.TypeChecker.TypeToString(subject)
	}
	if type_checking.IsTypeAnyType(subject) {
		return "any"
	}
	if type_checking.IsTypeFlagSet(subject, checker.TypeFlagsNever) {
		return "never"
	}
	if type_checking.IsTypeUnknownType(subject) {
		return "unknown"
	}
	if type_checking.IsTypeFlagSet(subject, checker.TypeFlagsTemplateLiteral) {
		return "template literal type"
	}
	// A boolean literal is matched by FLAG rather than through the shelf's true/false predicates.
	// Those require an intrinsic type, and one reached through an alias is not: measured on
	// `type B = false; type T = B & boolean`, the member reports BooleanLiteral set while
	// IsFalseLiteralType answers false, which sent the name to "literal type" and the message to
	// "overridden by the literal type" instead of "by the false".
	if type_checking.IsTypeFlagSet(subject, checker.TypeFlagsBooleanLiteral) {
		return ctx.TypeChecker.TypeToString(subject)
	}
	if type_checking.IsTypeFlagSet(subject, checker.TypeFlagsStringLiteral|
		checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) {
		// The checker prints a string literal already quoted, and upstream re-quotes the raw value,
		// so both land on the same double-quoted rendering.
		return ctx.TypeChecker.TypeToString(subject)
	}
	return "literal type"
}

// quoteForMessage renders a string with double quotes and the escapes a message needs.
func quoteForMessage(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			builder.WriteString("\\\"")
		case '\\':
			builder.WriteString("\\\\")
		case '\n':
			builder.WriteString("\\n")
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

// isInsideReturnType answers whether a union sits in a function's return annotation.
//
// `never` in a union is normally redundant, and in a return position it is not: `never | string`
// there says something about a function that can fail to return.
func isInsideReturnType(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	// A return annotation hangs directly off the function-like declaration, so the union's parent
	// IS the function rather than a separate annotation node the way estree models it.
	if !ast.IsFunctionLike(parent) {
		return false
	}
	return parent.Type() == node
}

// checkRedundantUnion is upstream's TSUnionType handler.
func checkRedundantUnion(ctx rule.Context, node *ast.Node, cache map[*ast.Node][]redundantTypePart) {
	members := node.AsUnionTypeNode().Types
	if members == nil {
		return
	}

	type literalSighting struct {
		typeNode *ast.Node
		name     string
	}
	seenLiterals := map[checker.TypeFlags][]literalSighting{}
	seenPrimitives := map[checker.TypeFlags]bool{}

	for _, rawMember := range members.Nodes {
		member := unwrapParenthesizedType(rawMember)
		if member == nil {
			continue
		}
		for _, part := range typePartsOf(ctx, member, cache) {
			// `any` and `unknown` swallow a union outright.
			if part.flags == checker.TypeFlagsAny || part.flags == checker.TypeFlagsUnknown {
				messageId := "overrides"
				if part.flags == checker.TypeFlagsAny && part.typeName != "any" {
					// The checker's error type carries the `any` flag, and calling it `any` in a
					// message sends the reader looking for an `any` nobody wrote.
					messageId = "errorTypeOverrides"
				}
				ctx.ReportNode(member, redundantMessage(messageId, "union", part.typeName, "", ""))
				continue
			}
			// `never` contributes nothing to a union, except in a return annotation.
			if part.flags == checker.TypeFlagsNever && !isInsideReturnType(node) {
				ctx.ReportNode(member, redundantMessage("overridden", "union", "never", "", ""))
				continue
			}

			for _, literalFlag := range redundantLiteralFlags {
				if part.flags == literalFlag {
					primitive := redundantPrimitiveOf[literalFlag]
					seenLiterals[primitive] = append(seenLiterals[primitive],
						literalSighting{typeNode: member, name: part.typeName})
					break
				}
			}
			for _, primitiveFlag := range redundantPrimitiveFlags {
				if part.flags&primitiveFlag != 0 {
					seenPrimitives[primitiveFlag] = true
				}
			}
		}
	}

	// A literal is redundant when the union also carries its primitive. Reported once per type node
	// per primitive, with every literal that node contributed listed together.
	type overriddenEntry struct {
		typeNode  *ast.Node
		primitive checker.TypeFlags
		names     []string
	}
	entries := []overriddenEntry{}
	for _, primitiveFlag := range redundantPrimitiveFlags {
		if !seenPrimitives[primitiveFlag] {
			continue
		}
		for _, sighting := range seenLiterals[primitiveFlag] {
			found := false
			for index := range entries {
				if entries[index].typeNode == sighting.typeNode &&
					entries[index].primitive == primitiveFlag {
					entries[index].names = append(entries[index].names, sighting.name)
					found = true
					break
				}
			}
			if !found {
				entries = append(entries, overriddenEntry{
					typeNode: sighting.typeNode, primitive: primitiveFlag,
					names: []string{sighting.name},
				})
			}
		}
	}
	// Grouped by type node so a node contributing several literals reports once per primitive, in
	// the order the nodes appear rather than in map order.
	sort.SliceStable(entries, func(a, b int) bool {
		return entries[a].typeNode.Pos() < entries[b].typeNode.Pos()
	})
	for _, entry := range entries {
		ctx.ReportNode(entry.typeNode, redundantMessage("literalOverridden", "union", "",
			strings.Join(entry.names, " | "), redundantPrimitiveNames[entry.primitive]))
	}
}

// checkRedundantIntersection is upstream's TSIntersectionType handler.
//
// The absorption runs the other way here: the narrower type wins, so a primitive beside a literal of
// the same kind is the redundant one.
func checkRedundantIntersection(ctx rule.Context, node *ast.Node, cache map[*ast.Node][]redundantTypePart) {
	members := node.AsIntersectionTypeNode().Types
	if members == nil {
		return
	}

	seenLiterals := map[checker.TypeFlags][]string{}
	seenPrimitives := map[checker.TypeFlags][]*ast.Node{}
	unionMembers := []*ast.Node{}
	unionParts := map[*ast.Node][]redundantTypePart{}

	for _, rawMember := range members.Nodes {
		member := unwrapParenthesizedType(rawMember)
		if member == nil {
			continue
		}
		parts := typePartsOf(ctx, member, cache)

		for _, part := range parts {
			// `any` and `never` swallow an intersection; `unknown` is swallowed BY it. That
			// asymmetry against the union arm is upstream's and is the reason the two arms are
			// written out rather than shared.
			handled := false
			for _, pair := range []struct {
				messageId string
				flag      checker.TypeFlags
			}{
				{"overrides", checker.TypeFlagsAny},
				{"overrides", checker.TypeFlagsNever},
				{"overridden", checker.TypeFlagsUnknown},
			} {
				if part.flags != pair.flag {
					continue
				}
				messageId := pair.messageId
				if part.flags == checker.TypeFlagsAny && part.typeName != "any" {
					messageId = "errorTypeOverrides"
				}
				ctx.ReportNode(member,
					redundantMessage(messageId, "intersection", part.typeName, "", ""))
				handled = true
				break
			}
			if handled {
				continue
			}

			for _, literalFlag := range redundantLiteralFlags {
				if part.flags == literalFlag {
					primitive := redundantPrimitiveOf[literalFlag]
					seenLiterals[primitive] = append(seenLiterals[primitive], part.typeName)
					break
				}
			}
			for _, primitiveFlag := range redundantPrimitiveFlags {
				if part.flags == primitiveFlag {
					seenPrimitives[primitiveFlag] = append(seenPrimitives[primitiveFlag], member)
				}
			}
		}

		// More than one part means the member resolved to a union, which the intersection arm
		// handles separately: every member of that union has to be absorbed for the report to fire.
		if len(parts) >= 2 {
			unionMembers = append(unionMembers, member)
			unionParts[member] = parts
		}
	}

	if len(unionMembers) > 0 {
		for _, unionMember := range unionMembers {
			var primitive checker.TypeFlags
			resolved := true
			for _, part := range unionParts[unionMember] {
				candidate, isLiteral := redundantPrimitiveOf[part.flags]
				if !isLiteral || len(seenPrimitives[candidate]) == 0 {
					resolved = false
					break
				}
				primitive = candidate
			}
			if !resolved {
				continue
			}
			names := []string{}
			for _, part := range unionParts[unionMember] {
				names = append(names, part.typeName)
			}
			ctx.ReportNode(unionMember, redundantMessage("primitiveOverridden", "intersection", "",
				strings.Join(names, " | "), redundantPrimitiveNames[primitive]))
		}
		return
	}

	for _, primitiveFlag := range redundantPrimitiveFlags {
		literals := seenLiterals[primitiveFlag]
		if len(literals) == 0 {
			continue
		}
		for _, typeNode := range seenPrimitives[primitiveFlag] {
			ctx.ReportNode(typeNode, redundantMessage("primitiveOverridden", "intersection", "",
				strings.Join(literals, " | "), redundantPrimitiveNames[primitiveFlag]))
		}
	}
}

// redundantMessage renders one of the rule's five messages.
//
// The slots differ per message, so this takes all of them and each arm uses what it needs, which
// keeps the five texts in one place where they can be compared against upstream's.
func redundantMessage(messageId, container, typeName, literal, primitive string) rule.Message {
	switch messageId {
	case "errorTypeOverrides":
		return rule.Message{Id: messageId,
			Description: "'" + typeName + "' is an 'error' type that acts as 'any' and overrides " +
				"all other types in this " + container + " type."}
	case "literalOverridden":
		return rule.Message{Id: messageId,
			Description: literal + " is overridden by " + primitive + " in this union type."}
	case "overridden":
		return rule.Message{Id: messageId,
			Description: "'" + typeName + "' is overridden by other types in this " + container + " type."}
	case "primitiveOverridden":
		return rule.Message{Id: messageId,
			Description: primitive + " is overridden by the " + literal + " in this intersection type."}
	default:
		return rule.Message{Id: "overrides",
			Description: "'" + typeName + "' overrides all other types in this " + container + " type."}
	}
}
