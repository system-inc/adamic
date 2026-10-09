package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// AccessorPairsOptions configures which half of the pairing this rule enforces, and where.
//
// Every field is a pointer because two of the four default to TRUE upstream. The config layer hands
// a rule configured as a bare severity a nil, and a bool field behind a nil decodes to false, which
// would invert those two defaults silently. A pointer keeps "absent" and "written as false" apart,
// which is the whole reason the decoder below is hand-rolled rather than `rule.DecodeOptionsInto`.
type AccessorPairsOptions struct {
	// GetWithoutSet reports a getter with no matching setter. Defaults to FALSE upstream.
	//
	// Off by default because a read-only property is an ordinary and intentional thing, while a
	// write-only one almost never is. The asymmetry is upstream's and is the reason the two options
	// exist separately rather than as one switch.
	GetWithoutSet *bool `json:"getWithoutSet"`

	// SetWithoutGet reports a setter with no matching getter. Defaults to TRUE upstream.
	SetWithoutGet *bool `json:"setWithoutGet"`

	// EnforceForClassMembers extends the check to class bodies. Defaults to TRUE upstream.
	EnforceForClassMembers *bool `json:"enforceForClassMembers"`

	// EnforceForTSTypes extends the check to interface bodies and type literals. Defaults to FALSE.
	//
	// Off by default because a type declaring only `get prop(): T` is a legitimate way to describe a
	// readonly-through-an-accessor shape, and a setter with no getter in a type is far less
	// obviously a mistake than the same thing in a class.
	EnforceForTSTypes *bool `json:"enforceForTSTypes"`
}

// accessorPairsSettings is the decoded form, with every default already applied.
type accessorPairsSettings struct {
	getWithoutSet          bool
	setWithoutGet          bool
	enforceForClassMembers bool
	enforceForTSTypes      bool
}

// defaultAccessorPairsSettings mirrors upstream's `meta.defaultOptions`.
func defaultAccessorPairsSettings() accessorPairsSettings {
	return accessorPairsSettings{
		getWithoutSet:          false,
		setWithoutGet:          true,
		enforceForClassMembers: true,
		enforceForTSTypes:      false,
	}
}

// resolve applies each written option over the defaults, leaving an absent key alone.
func (options AccessorPairsOptions) resolve() accessorPairsSettings {
	settings := defaultAccessorPairsSettings()
	if options.GetWithoutSet != nil {
		settings.getWithoutSet = *options.GetWithoutSet
	}
	if options.SetWithoutGet != nil {
		settings.setWithoutGet = *options.SetWithoutGet
	}
	if options.EnforceForClassMembers != nil {
		settings.enforceForClassMembers = *options.EnforceForClassMembers
	}
	if options.EnforceForTSTypes != nil {
		settings.enforceForTSTypes = *options.EnforceForTSTypes
	}
	return settings
}

// DecodeAccessorPairsOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, because that helper errors on empty input and the
// config layer turns the error into nil, so a rule reached through it cannot tell "no options were
// written" from "an option was written as false". Two of these four options default to true, so that
// distinction inverts the rule rather than merely narrowing it.
func DecodeAccessorPairsOptions(raw []byte) (any, error) {
	var options AccessorPairsOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// The eight findings, split by which half is missing and by what kind of thing holds the accessor.
//
// Upstream carries eight message ids for what reads like one judgment, and the split is not
// decorative: an object literal, a class body, a type, and a property descriptor are four different
// mistakes wearing the same shape, and only the descriptor one names no accessor at all, because
// there is no accessor node to name.
var (
	messageAccessorPairsMissingGetterInObjectLiteral = "missingGetterInObjectLiteral"
	messageAccessorPairsMissingSetterInObjectLiteral = "missingSetterInObjectLiteral"
	messageAccessorPairsMissingGetterInClass         = "missingGetterInClass"
	messageAccessorPairsMissingSetterInClass         = "missingSetterInClass"
	messageAccessorPairsMissingGetterInType          = "missingGetterInType"
	messageAccessorPairsMissingSetterInType          = "missingSetterInType"
)

var messageAccessorPairsMissingGetterInDescriptor = rule.Message{
	Id: "missingGetterInPropertyDescriptor",
	Description: "This property descriptor defines a setter and no getter, so the property can be " +
		"written and reads back as undefined. That is almost always a missing getter rather than " +
		"a deliberate write-only property. Add a `get`, or use a plain data property.",
}

var messageAccessorPairsMissingSetterInDescriptor = rule.Message{
	Id: "missingSetterInPropertyDescriptor",
	Description: "This property descriptor defines a getter and no setter, so assigning to the " +
		"property is silently ignored in sloppy mode and throws in strict mode. Add a `set`, or " +
		"make the read-only intent explicit.",
}

// AccessorPairs flags a getter with no setter, or a setter with no getter, beside it.
//
//	valid:   var o = { get a() {}, set a(v) {} };
//	valid:   var o = { get a() {} };                       // getWithoutSet defaults off
//	valid:   class A { get a() {} set a(v) {} }
//	invalid: var o = { set a(v) {} };
//	invalid: class A { set a(v) {} }
//	invalid: Object.defineProperty(foo, 'bar', { set(v) {} })
//
// # What "beside it" means, and why it is not scope
//
// The pairing is positional rather than semantic: two accessors pair when they share a container and
// a key, where a container is one object literal, one class's static side, one class's instance
// side, or one type body. So `class A { get a() {} static set a(v) {} }` reports TWICE, because the
// static and instance sides are separate containers even though the class is one. That is measured
// against the installed rule rather than inferred, and it is the case a port written around "does
// the class declare both" would get wrong in the quiet direction.
//
// # Key equality has two mechanisms, and the second one is the surprise
//
// A key whose value the syntax settles compares by that value, so `get 1()` pairs with `set '1'()`
// and `0x10` pairs with `16`. A key the syntax does not settle -- a computed one -- compares by its
// TOKEN LIST instead, so `[x + 1]` pairs with `[x  +  1]` while `[x]` and `[y]` do not. Upstream
// falls back to `sourceCode.getTokens(node.key)` for exactly this, and dropping the fallback would
// leave every computed accessor unpaired and report both halves of a correct pair.
//
// Two accessors with unsettled keys therefore pair on a textual coincidence that may not hold at
// runtime, and upstream knows this. It is reproduced rather than improved on: `[foo()]` twice pairs
// here as it does upstream, even though the two calls may return different keys.
//
// # The property descriptor is a second, independent judgment on the same node
//
// An object literal passed as the descriptor argument of `Object.defineProperty`, `Reflect.
// defineProperty`, `Object.defineProperties`, or `Object.create` is checked a SECOND time under
// different rules: there the accessors are ordinary `get:`/`set:` properties rather than accessor
// syntax, so the question is whether both names are present. Both checks run on the same node and
// one input can report from each.
//
// # Why this reads the type checker
//
// The descriptor check requires `Object` and `Reflect` to be the real globals rather than local
// bindings that happen to share the name, which upstream asks through scope analysis and this asks
// through the checker. `resolvesToAGlobal` in this package already answers exactly that question.
// Two of upstream's own clean cases turn on it, and both are imported.
var AccessorPairs = rule.Rule{
	Name:             "accessor-pairs",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := defaultAccessorPairsSettings()
		if decoded, configured := rule.OptionsAs[AccessorPairsOptions](options); configured {
			settings = decoded.resolve()
		}

		// Both halves off means the rule has nothing to say, and upstream registers no listener at
		// all in that case. Reproduced so the cost is the same too.
		if !settings.getWithoutSet && !settings.setWithoutGet {
			return rule.Listeners{}
		}

		listeners := rule.Listeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkAccessorPairsObjectLiteral(ctx, node, settings)
			},
		}
		if settings.enforceForClassMembers {
			classBody := func(node *ast.Node) {
				checkAccessorPairsClassBody(ctx, node, settings)
			}
			listeners[ast.KindClassDeclaration] = classBody
			listeners[ast.KindClassExpression] = classBody
		}
		if settings.enforceForTSTypes {
			typeBody := func(node *ast.Node) {
				checkAccessorPairsTypeBody(ctx, node, settings)
			}
			listeners[ast.KindInterfaceDeclaration] = typeBody
			listeners[ast.KindTypeLiteral] = typeBody
		}
		return listeners
	},
}

// accessorPairsGroup collects the getters and setters sharing one key inside one container.
type accessorPairsGroup struct {
	// key is the settled property name when the syntax gives one.
	key string
	// keySettled says whether `key` means anything. A computed key that names no static value has
	// none, and compares through `keyTokens` instead.
	keySettled bool
	// keyTokens is the token spelling of an unsettled key, compared elementwise.
	keyTokens []accessorPairsToken
	getters   []*ast.Node
	setters   []*ast.Node
}

// accessorPairsToken is one token of a computed key, reduced to what upstream compares.
//
// Upstream compares `type` and `value`, which is why `[x + 1]` and `[x  +  1]` pair: whitespace is
// trivia and never becomes a token, so it cannot differ.
type accessorPairsToken struct {
	kind ast.Kind
	text string
}

// checkAccessorPairsList groups a container's accessors by key and reports the unpaired ones.
//
// The grouping is upstream's, including its order: groups are created in source order and each new
// accessor joins the FIRST group whose key matches, so the reports come out in the order the
// accessors were written rather than sorted by key.
func checkAccessorPairsList(
	ctx rule.Context,
	accessors []*ast.Node,
	settings accessorPairsSettings,
	messageMissingGetter string,
	messageMissingSetter string,
) []accessorPairsFinding {
	var groups []*accessorPairsGroup

	for _, accessor := range accessors {
		key, settled := accessorPairsKeyOf(ctx, accessor)
		var tokens []accessorPairsToken
		if !settled {
			tokens = accessorPairsTokensOf(ctx, accessor)
		}

		group := (*accessorPairsGroup)(nil)
		for _, candidate := range groups {
			if accessorPairsKeysEqual(candidate, key, settled, tokens) {
				group = candidate
				break
			}
		}
		if group == nil {
			group = &accessorPairsGroup{key: key, keySettled: settled, keyTokens: tokens}
			groups = append(groups, group)
		}
		if accessor.Kind == ast.KindGetAccessor {
			group.getters = append(group.getters, accessor)
		} else {
			group.setters = append(group.setters, accessor)
		}
	}

	var findings []accessorPairsFinding
	for _, group := range groups {
		if settings.setWithoutGet && len(group.setters) > 0 && len(group.getters) == 0 {
			for _, setter := range group.setters {
				findings = append(findings, accessorPairsFinding{
					accessor: setter, messageId: messageMissingGetter, missingGetter: true})
			}
		}
		if settings.getWithoutSet && len(group.getters) > 0 && len(group.setters) == 0 {
			for _, getter := range group.getters {
				findings = append(findings, accessorPairsFinding{
					accessor: getter, messageId: messageMissingSetter, missingGetter: false})
			}
		}
	}
	return findings
}

// accessorPairsFinding is one pending report, held until the caller can order them.
//
// The grouping walk produces findings grouped by KEY, and a class produces two such walks, so the
// natural order is neither upstream's nor stable. Upstream gets source order for free by visiting
// nodes rather than groups; this recovers it by sorting at the end. Measured: `class A { get a() {}
// static set a(foo) {} }` reports the getter first upstream, and grouping alone reports the static
// side first.
type accessorPairsFinding struct {
	accessor      *ast.Node
	messageId     string
	missingGetter bool
}

// reportAccessorPairsInSourceOrder emits pending findings ordered by where they appear in the file.
func reportAccessorPairsInSourceOrder(ctx rule.Context, findings []accessorPairsFinding) {
	slices.SortStableFunc(findings, func(left, right accessorPairsFinding) int {
		return left.accessor.Pos() - right.accessor.Pos()
	})
	for _, finding := range findings {
		reportAccessorPairsAccessor(ctx, finding.accessor, finding.messageId,
			finding.missingGetter)
	}
}

// accessorPairsKeysEqual is upstream's `areEqualKeys`.
//
// Two settled keys compare by value; two unsettled ones compare by token list; a settled key never
// equals an unsettled one, which is why `{ get a() {}, set [a](v) {} }` reports both halves even
// though the computed key spells the same identifier. That last arm is the one an intuitive reading
// gets wrong, and it is deliberate upstream: `[a]` names whatever the VARIABLE `a` holds, which is
// not the property `a`.
func accessorPairsKeysEqual(
	group *accessorPairsGroup,
	key string,
	settled bool,
	tokens []accessorPairsToken,
) bool {
	if group.keySettled && settled {
		return group.key == key
	}
	if group.keySettled || settled {
		return false
	}
	if len(group.keyTokens) != len(tokens) {
		return false
	}
	for index, left := range group.keyTokens {
		if left != tokens[index] {
			return false
		}
	}
	return true
}

// accessorPairsKeyOf answers the settled property name of an accessor's key, if the syntax gives one.
//
// `property.Static` is the shelf's spelling of upstream's `getStaticPropertyName`, and it agrees with
// it on every spelling this rule can meet with ONE exception, handled here rather than on the shelf:
// upstream returns null for a private name, because `#a` is neither a non-computed Identifier nor a
// Literal, so two private accessors pair through the TOKEN path rather than the value path. The shelf
// helper accepts private names, which is the more useful answer for its other callers and the wrong
// one here.
//
// Reproducing upstream's null rather than the shelf's answer changes no verdict measured -- a private
// key's token list is a single token whose text is the name, so both routes pair `#a` with `#a` and
// separate it from `a`. It is written this way so the mechanism matches, because the rendered message
// differs between the two paths and that IS visible.
func accessorPairsKeyOf(ctx rule.Context, accessor *ast.Node) (string, bool) {
	name := accessor.Name()
	if name == nil {
		return "", false
	}
	if name.Kind == ast.KindPrivateIdentifier {
		return "", false
	}
	return property.Name(name, property.Static)
}

// accessorPairsTokensOf renders an accessor's key as the token list upstream compares.
//
// Reached only when the key has no settled value, which is a computed key or a private name.
//
// # Why this walks the tree instead of running the scanner over the key's text
//
// The obvious implementation points `GetScannerForSourceFile` at the key and scans to its end. It is
// wrong on templates, and wrong in the loud direction: a scanner started mid-expression has no
// template stack, so after `${a}` it reads the closing backtick as the OPENING of a new template and
// swallows the rest of the file into one token. Measured on `[`+"`"+`${a}`+"`"+`]`, the token text came back as
// "`+"`"+`]() {}, set [`+"`"+`" -- different for two keys that are textually identical, so a correct pair
// reported both halves. Walking the parsed tree cannot make that mistake, because the parser already
// did the nesting.
//
// # Parentheses are skipped, which is fidelity rather than tidiness
//
// Upstream compares `sourceCode.getTokens(node.key)`, and in ESTree a computed key's `key` IS the
// inner expression: the brackets are not in it, and neither are the outer parentheses, because
// `(a)` parses to the Identifier `a` with no node of its own. So `[a]` and `[(a)]` produce identical
// token lists upstream and pair, which is one of upstream's own clean cases. Our parser keeps the
// parentheses as a node, so they are skipped here to ask the same question.
//
// Only the OUTER parentheses vanish this way, in both parsers: `[(a) + 1]` keeps its inner pair as
// tokens upstream and as a node here, so the two spellings stay separated on both sides.
func accessorPairsTokensOf(ctx rule.Context, accessor *ast.Node) []accessorPairsToken {
	name := accessor.Name()
	if name == nil {
		return nil
	}
	key := name
	if name.Kind == ast.KindComputedPropertyName {
		// `SkipParentheses` DEREFERENCES its argument, so the nil test comes first rather than
		// after. A malformed computed key recovers to one with no expression, and the parse the
		// error recovery hands back is not the parse the grammar describes.
		inner := name.AsComputedPropertyName().Expression
		if inner == nil {
			return nil
		}
		key = ast.SkipParentheses(inner)
	}
	if key == nil {
		return nil
	}

	var tokens []accessorPairsToken
	accessorPairsCollectTokens(ctx, key, &tokens)
	return tokens
}

// accessorPairsCollectTokens appends a structural rendering of one expression, in source order.
//
// Each node contributes a marker carrying its KIND before its children contribute theirs, and a leaf
// contributes its text. The kind marker is what separates `[a]` from `[a++]` and `[-a]`: our parser
// hangs a unary operator off the node kind rather than off a child, so a walk that emitted only
// leaves would render all three as the single token `a` and pair keys upstream keeps apart. That is
// not hypothetical -- it was a live defect here, caught by upstream's own `[a]` against `[a++]`
// case.
//
// Whitespace never contributes, because `TokenRange` strips leading trivia and no marker carries
// position. That is what pairs `[x + 1]` with `[x  +  1]`, which is upstream's behaviour and the
// reason this path exists at all.
func accessorPairsCollectTokens(ctx rule.Context, node *ast.Node, tokens *[]accessorPairsToken) {
	if node == nil {
		return
	}

	hadChild := false
	node.ForEachChild(func(child *ast.Node) bool {
		hadChild = true
		return false
	})

	if !hadChild {
		// A leaf renders as its own text, which is what makes `[a]` and `[b]` differ.
		span := rule.TokenRange(ctx.SourceFile, node)
		*tokens = append(*tokens, accessorPairsToken{
			kind: node.Kind,
			text: ctx.SourceFile.Text()[span.Pos():span.End()],
		})
		return
	}

	// A composite renders as its kind, plus the operator when the kind alone does not carry it.
	// `KindBinaryExpression` covers `+` and `-` alike, so the operator token has to be emitted; a
	// prefix or postfix expression stores its operator as a field on the node, and both spellings of
	// it are recovered from the source text of the operator itself.
	*tokens = append(*tokens, accessorPairsToken{kind: node.Kind, text: ""})
	if operator := accessorPairsOperatorTokenOf(ctx, node); operator != "" {
		*tokens = append(*tokens, accessorPairsToken{kind: node.Kind, text: operator})
	}

	node.ForEachChild(func(child *ast.Node) bool {
		// Guarded before the skip, not inside the recursion: `SkipParentheses` dereferences.
		if child == nil {
			return false
		}
		accessorPairsCollectTokens(ctx, ast.SkipParentheses(child), tokens)
		return false
	})
}

// accessorPairsOperatorTokenOf renders the operator of a composite expression, if it has one.
//
// Three kinds carry an operator the node kind does not identify. Everything else answers "", which
// costs one token slot and no correctness: two nodes of the same kind with no operator are the same
// shape, which is the whole question here.
func accessorPairsOperatorTokenOf(ctx rule.Context, node *ast.Node) string {
	var operatorKind ast.Kind
	switch node.Kind {
	case ast.KindBinaryExpression:
		if operator := node.AsBinaryExpression().OperatorToken; operator != nil {
			span := rule.TokenRange(ctx.SourceFile, operator)
			return ctx.SourceFile.Text()[span.Pos():span.End()]
		}
		return ""
	case ast.KindPrefixUnaryExpression:
		operatorKind = node.AsPrefixUnaryExpression().Operator
	case ast.KindPostfixUnaryExpression:
		operatorKind = node.AsPostfixUnaryExpression().Operator
	default:
		return ""
	}
	// The operator is a Kind rather than a node here, and rendering it as a number is enough: the
	// comparison only needs two different operators to render differently, never to read back.
	return fmt.Sprint(operatorKind)
}

// reportAccessorPairsAccessor reports one accessor, on the span upstream reports.
//
// The span runs from the accessor's first modifier-or-keyword token through the end of its key, so
// `static set a(v) {}` reports `static set a` and `set 'a b'(v) {}` reports `set 'a b'`. Measured
// against the installed rule across nine shapes rather than read off the source: upstream builds it
// from `getFunctionHeadLoc`, whose input is the function VALUE rather than the property, and the two
// descriptions of the resulting span do not obviously agree.
//
// Built from `TokenRange` rather than `Pos()`, because `Pos()` carries leading trivia and would put
// the span's start on the whitespace before `set`.
func reportAccessorPairsAccessor(
	ctx rule.Context,
	accessor *ast.Node,
	messageId string,
	missingGetter bool,
) {
	name := accessor.Name()
	if name == nil {
		return
	}
	head := rule.TokenRange(ctx.SourceFile, accessor)
	span := core.NewTextRange(head.Pos(), name.End())

	// The two halves fail differently at runtime, so they get different sentences rather than one
	// sentence with the word swapped. A write-only property reads back as undefined; a read-only one
	// swallows the assignment in sloppy mode and throws in strict mode.
	missing, consequence := "getter", "so the property can be written but always reads back as "+
		"undefined"
	if !missingGetter {
		missing, consequence = "setter", "so assigning to the property is silently ignored in "+
			"sloppy mode and throws in strict mode"
	}

	ctx.ReportRange(span, rule.Message{
		Id: messageId,
		Description: fmt.Sprintf(
			"This declares %s with no matching %s beside it, %s. Declare both halves of the "+
				"pair, or drop the accessor and use a plain property.",
			accessorPairsDescribe(accessor), missing, consequence),
	})
}

// accessorPairsDescribe renders an accessor the way upstream's `getFunctionNameWithKind` does.
//
// The parts, in upstream's order: `static`, then `private` for a private name, then `getter` or
// `setter`, then the key in single quotes. A private name renders BARE and hash-carrying (`private
// setter #a`) while a string key that happens to spell one renders quoted (`setter '#a'`), which is
// the pair of shapes that tells you the two paths are distinct rather than one path with a quoting
// bug.
//
// A key with no settled value renders with no name at all (`setter`), except inside a type body,
// where upstream interpolates the JavaScript null through a template and produces the literal text
// `'null'`. That is a defect, it is reproduced, and the reproduction is in `accessorPairsDescribeType`
// rather than here so this function stays honest.
func accessorPairsDescribe(accessor *ast.Node) string {
	var parts []string
	if ast.IsStatic(accessor) {
		parts = append(parts, "static")
	}
	name := accessor.Name()
	if name != nil && name.Kind == ast.KindPrivateIdentifier {
		parts = append(parts, "private")
	}
	if accessor.Kind == ast.KindGetAccessor {
		parts = append(parts, "getter")
	} else {
		parts = append(parts, "setter")
	}
	if name != nil && name.Kind == ast.KindPrivateIdentifier {
		return strings.Join(append(parts, name.Text()), " ")
	}
	if key, settled := property.Name(name, property.Static); settled {
		return strings.Join(append(parts, "'"+key+"'"), " ")
	}
	return strings.Join(parts, " ")
}

// checkAccessorPairsObjectLiteral runs both judgments on one object literal.
func checkAccessorPairsObjectLiteral(
	ctx rule.Context,
	node *ast.Node,
	settings accessorPairsSettings,
) {
	reportAccessorPairsInSourceOrder(ctx,
		checkAccessorPairsList(ctx, accessorPairsMembersOf(node), settings,
			messageAccessorPairsMissingGetterInObjectLiteral,
			messageAccessorPairsMissingSetterInObjectLiteral))

	if accessorPairsIsPropertyDescriptor(ctx, node) {
		checkAccessorPairsDescriptor(ctx, node, settings)
	}
}

// checkAccessorPairsClassBody runs the pairing over a class's two sides separately.
//
// Static and instance members are different containers, which is why this is two calls rather than
// one over the whole body. `class A { get a() {} static set a(v) {} }` reports twice, measured.
func checkAccessorPairsClassBody(
	ctx rule.Context,
	node *ast.Node,
	settings accessorPairsSettings,
) {
	var staticSide, instanceSide []*ast.Node
	for _, member := range accessorPairsMembersOf(node) {
		if ast.IsStatic(member) {
			staticSide = append(staticSide, member)
		} else {
			instanceSide = append(instanceSide, member)
		}
	}
	// The two sides are separate containers for PAIRING and one file for REPORTING, so they are
	// walked apart and merged back into source order. Upstream never has to do this because it
	// reports as it visits; here the grouping walk is what reorders things.
	findings := checkAccessorPairsList(ctx, staticSide, settings,
		messageAccessorPairsMissingGetterInClass, messageAccessorPairsMissingSetterInClass)
	findings = append(findings, checkAccessorPairsList(ctx, instanceSide, settings,
		messageAccessorPairsMissingGetterInClass, messageAccessorPairsMissingSetterInClass)...)
	reportAccessorPairsInSourceOrder(ctx, findings)
}

// checkAccessorPairsTypeBody runs the pairing over an interface body or a type literal.
func checkAccessorPairsTypeBody(
	ctx rule.Context,
	node *ast.Node,
	settings accessorPairsSettings,
) {
	reportAccessorPairsInSourceOrder(ctx,
		checkAccessorPairsList(ctx, accessorPairsMembersOf(node), settings,
			messageAccessorPairsMissingGetterInType, messageAccessorPairsMissingSetterInType))
}

// accessorPairsMembersOf collects the accessor children of a container, in source order.
//
// One helper for all four container kinds, because our parser gives an object literal, a class body,
// an interface body and a type literal the same accessor node kinds. Upstream needs four paths and
// two node types for the same job; this is the fidelity-to-the-decision rather than to the mechanism
// that the brief asks for.
func accessorPairsMembersOf(node *ast.Node) []*ast.Node {
	var accessors []*ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindGetAccessor || child.Kind == ast.KindSetAccessor {
			accessors = append(accessors, child)
		}
		return false
	})
	return accessors
}

// checkAccessorPairsDescriptor runs the second, independent judgment on a property descriptor.
//
// The question here is not accessor syntax at all: a descriptor names its halves with ORDINARY
// properties called `get` and `set`, so `{ get: g }` and `{ get() {} }` and `{ get }` are all a
// getter while `{ ['get']: g }` is not. Upstream reads the key name of every non-computed `init`
// property, which is what makes the computed spelling decline: a computed key names whatever the
// expression evaluates to, and `defineProperty` will read it at runtime while this cannot.
//
// The finding is reported on the descriptor object rather than on any property, because there is no
// accessor node to point at -- the whole point is that one of them is absent.
func checkAccessorPairsDescriptor(
	ctx rule.Context,
	node *ast.Node,
	settings accessorPairsSettings,
) {
	hasGetter, hasSetter := false, false
	node.ForEachChild(func(child *ast.Node) bool {
		name := accessorPairsDescriptorPropertyName(child)
		switch name {
		case "get":
			hasGetter = true
		case "set":
			hasSetter = true
		}
		return false
	})

	if settings.setWithoutGet && hasSetter && !hasGetter {
		ctx.ReportNode(node, messageAccessorPairsMissingGetterInDescriptor)
	}
	if settings.getWithoutSet && hasGetter && !hasSetter {
		ctx.ReportNode(node, messageAccessorPairsMissingSetterInDescriptor)
	}
}

// accessorPairsDescriptorPropertyName answers the name a descriptor member declares, if any.
//
// Upstream restricts this to `kind === "init"` properties, which excludes accessor syntax: a
// descriptor written `{ get set(v) {} }` declares an ACCESSOR named `set`, not a setter for the
// described property, so it must not count. Our parser separates those by node kind, which is the
// same distinction reached a shorter way.
//
// Computed keys are excluded whatever they spell, because upstream reads `key.name`, which a
// computed key does not have. That is why `{ ['set']: s }` is clean, measured.
func accessorPairsDescriptorPropertyName(member *ast.Node) string {
	var name *ast.Node
	switch member.Kind {
	case ast.KindPropertyAssignment:
		name = member.AsPropertyAssignment().Name()
	case ast.KindShorthandPropertyAssignment:
		name = member.AsShorthandPropertyAssignment().Name()
	case ast.KindMethodDeclaration:
		name = member.AsMethodDeclaration().Name()
	default:
		// A spread carries no key, and an accessor declares one of its own rather than one of the
		// described property's. Both are correctly invisible here.
		return ""
	}
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// accessorPairsIsPropertyDescriptor answers whether an object literal is being passed as one.
//
// Four call shapes qualify, and they split into two arrangements. `Object.defineProperty(o, k, D)`
// and `Reflect.defineProperty(o, k, D)` take the descriptor DIRECTLY as their third argument.
// `Object.defineProperties(o, M)` and `Object.create(o, M)` take a MAP whose values are descriptors,
// so the literal qualifies when it is the value of a property of that second argument.
//
// The receiver must resolve to the real global. `var Object = {}; Object.defineProperty(...)` is
// clean upstream and here, which is one of the two cases in the imported corpus that make this a
// checker-reading rule rather than a syntactic one.
func accessorPairsIsPropertyDescriptor(ctx rule.Context, node *ast.Node) bool {
	if accessorPairsIsArgumentOfGlobalCall(ctx, node, "Object", "defineProperty", 2) ||
		accessorPairsIsArgumentOfGlobalCall(ctx, node, "Reflect", "defineProperty", 2) {
		return true
	}

	// The map arrangement: this literal is a property's value, and the object holding it is the
	// second argument.
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAssignment {
		return false
	}
	if parent.AsPropertyAssignment().Initializer != node {
		return false
	}
	grandparent := parent.Parent
	if grandparent == nil || grandparent.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	return accessorPairsIsArgumentOfGlobalCall(ctx, grandparent, "Object", "defineProperties", 1) ||
		accessorPairsIsArgumentOfGlobalCall(ctx, grandparent, "Object", "create", 1)
}

// accessorPairsIsArgumentOfGlobalCall matches `<Global>.<method>(...)` with `node` at `index`.
//
// Optional chaining on the receiver is accepted, because `Object?.defineProperty(o, k, D)` calls the
// same function and upstream reports it -- measured against the installed rule rather than read off
// `skipChainExpression`, which is about a node type our parser does not have.
func accessorPairsIsArgumentOfGlobalCall(
	ctx rule.Context,
	node *ast.Node,
	globalName string,
	methodName string,
	index int,
) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) <= index {
		return false
	}
	if call.Arguments.Nodes[index] != node {
		return false
	}

	if call.Expression == nil {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if name, settled := property.Name(access.Name(), property.Textual); !settled ||
		name != methodName {
		return false
	}

	if access.Expression == nil {
		return false
	}
	receiver := ast.SkipParentheses(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindIdentifier ||
		receiver.Text() != globalName {
		return false
	}
	// A local binding of the same name is a different object, and calling `defineProperty` on it
	// says nothing about property descriptors. This is the whole reason the rule reads the checker.
	return resolvesToAGlobal(ctx, receiver)
}
