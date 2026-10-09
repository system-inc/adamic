package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// GroupedAccessorPairsOrder is upstream's three-valued first option.
type GroupedAccessorPairsOrder string

const (
	// GroupedAccessorPairsAnyOrder requires only that the pair be adjacent, and is the default.
	GroupedAccessorPairsAnyOrder GroupedAccessorPairsOrder = "anyOrder"
	// GroupedAccessorPairsGetBeforeSet additionally requires the getter first.
	GroupedAccessorPairsGetBeforeSet GroupedAccessorPairsOrder = "getBeforeSet"
	// GroupedAccessorPairsSetBeforeGet additionally requires the setter first.
	GroupedAccessorPairsSetBeforeGet GroupedAccessorPairsOrder = "setBeforeGet"
)

// GroupedAccessorPairsOptions carries upstream's whole option surface.
//
// The wire shape is a LIST, unlike most rules here, because upstream's schema is positional:
// `["error", "getBeforeSet", { enforceForTSTypes: true }]`. The rule registers with
// `DecodeOptionList`, so the decoder is handed that list with the severity removed.
type GroupedAccessorPairsOptions struct {
	Order GroupedAccessorPairsOrder

	// EnforceForTSTypes also checks the accessor signatures in an interface body or a type literal.
	// Off by default, as upstream's is.
	EnforceForTSTypes bool
}

// DecodeGroupedAccessorPairsOptions turns the configured array into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, and this rule is a case that helper cannot
// serve. The first option is a bare STRING whose default is "anyOrder", so a zero value would be the
// empty string, which matches none of the three arms: the rule would register on every file, be
// handed nil by the live config's bare "error", and silently stop enforcing order while every
// fixture passing an explicit option stayed green.
//
// A value outside the enum is refused rather than folded into an arm. Upstream gets that refusal
// from its schema before the rule runs; there is no schema layer here, so it lives here.
func DecodeGroupedAccessorPairsOptions(list []byte) (any, error) {
	options := GroupedAccessorPairsOptions{Order: GroupedAccessorPairsAnyOrder}
	configured, err := rule.OptionElements(list, 2)
	if err != nil || len(configured) == 0 {
		return options, err
	}

	var order string
	if err := json.Unmarshal(configured[0], &order); err != nil {
		return options, fmt.Errorf("grouped-accessor-pairs takes a string first, %q: %w",
			string(configured[0]), err)
	}
	switch GroupedAccessorPairsOrder(order) {
	case GroupedAccessorPairsAnyOrder, GroupedAccessorPairsGetBeforeSet, GroupedAccessorPairsSetBeforeGet:
		options.Order = GroupedAccessorPairsOrder(order)
	default:
		return options, fmt.Errorf("grouped-accessor-pairs takes %q, %q or %q, got %q",
			GroupedAccessorPairsAnyOrder, GroupedAccessorPairsGetBeforeSet,
			GroupedAccessorPairsSetBeforeGet, order)
	}
	if len(configured) < 2 {
		return options, nil
	}

	// The second element's only key is `enforceForTSTypes`.
	var second struct {
		EnforceForTSTypes bool `json:"enforceForTSTypes"`
	}
	if err := rule.UnmarshalOptions(configured[1], &second); err != nil {
		return options, fmt.Errorf("grouped-accessor-pairs element 2: %w", err)
	}
	options.EnforceForTSTypes = second.EnforceForTSTypes
	return options, nil
}

var messageGroupedAccessorPairsNotGrouped = rule.Message{
	Id: "notGrouped",
	Description: "Accessor pair should be grouped. A getter and its setter describe one property, " +
		"and splitting them puts half the definition somewhere a reader looking at the other half " +
		"will not see it, so a change to one silently stops matching the other.",
}

var messageGroupedAccessorPairsInvalidOrder = rule.Message{
	Id: "invalidOrder",
	Description: "Accessor pair is in the wrong order for this project's convention. The pair is " +
		"grouped, which is the substance; this is the ordering the configuration asks for on top " +
		"of that.",
}

// GroupedAccessorPairs requires a getter and its setter to sit next to each other.
//
//	valid:   ({ get a(){}, set a(foo){} })
//	valid:   class A { get a(){} static set a(foo){} }
//	valid:   ({ get a(){}, get a(){} })
//	invalid: ({ get a(){}, b:1, set a(foo){} })
//	invalid: ({ set a(foo){}, get a(){} })      under getBeforeSet
//
// A getter and its setter are one property written in two halves, so separating them hides half the
// definition from anyone reading the other half.
//
// # Four lists, not one
//
// Upstream checks an object literal's properties as one list, and a class body as TWO: static
// members and instance members separately. That is what makes
// `class A { get a(){} static set a(foo){} }` clean rather than a split pair, since those are two
// different properties that happen to share a name. Reproduced by running the class scan twice with
// opposite static filters, exactly as upstream does.
//
// # Which pairs count
//
// Only a key with exactly ONE getter and ONE setter. A duplicate getter means the code is already
// broken in a way this rule has no opinion about, and upstream skips it rather than guessing which
// of the two the setter belongs to.
//
// Keys are matched by their static name where one exists. A computed key has no static name, and
// upstream falls back to comparing the key's TOKEN LIST, so `get [a]` and `set [a]` pair while
// `get [a]` and `set [b]` do not. That is reproduced here by comparing the key's source text, which
// is the same comparison over the same bytes for every shape the corpus writes.
//
// # The span is the accessor head
//
// Not the whole member: `set a`, `static get a`, `set [a]`, running from the `get`/`set` or `static`
// keyword through the key and stopping before the parameter list. Upstream computes it with
// `getFunctionHeadLoc` over the accessor's function value; our parser has no separate value node, so
// the same span is the member's own token start through its key's end. Probed against seven shapes
// including a static, a computed key, a private name, and numeric and string keys.
//
// # Type members, under enforceForTSTypes
//
// The option extends the same judgment to the accessor signatures in an interface body or a type
// literal, nested ones included. Upstream checks only its `TSMethodSignature`s of kind get or set
// there; this parser gives those the same GetAccessor and SetAccessor kinds a class accessor has,
// and the list scan already counts only those, so the scan runs unchanged over the members. Type
// members carry no `static`, so the list is scanned once. Spans and names match the class case,
// measured against the installed eslint at 10.8.1 over ten shapes (#d21war2).
//
// One kept divergence, in the message only: a signature whose key is computed and not literal,
// `get [k](): any`, renders upstream as `getter 'null'`, because upstream reads the signature's name
// through `getStaticPropertyName` and quotes the null it gets back. A class accessor with the same
// key renders as a bare `getter`, and so does a signature here. The finding and its span agree.
var GroupedAccessorPairs = rule.Rule{
	Name: "grouped-accessor-pairs",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare "error" is handed nil, so the fallback here is what keeps the
		// order check from comparing against an empty string that matches no arm.
		settings, isSettings := rule.OptionsAs[GroupedAccessorPairsOptions](options)
		if !isSettings {
			settings = GroupedAccessorPairsOptions{Order: GroupedAccessorPairsAnyOrder}
		}

		listeners := rule.Listeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkGroupedAccessorList(ctx, settings.Order,
					node.AsObjectLiteralExpression().Properties.Nodes, nil)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkGroupedAccessorClassBody(ctx, settings.Order, node.AsClassDeclaration().Members.Nodes)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				checkGroupedAccessorClassBody(ctx, settings.Order, node.AsClassExpression().Members.Nodes)
			},
		}
		if settings.EnforceForTSTypes {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				checkGroupedAccessorList(ctx, settings.Order, node.AsInterfaceDeclaration().Members.Nodes, nil)
			}
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				checkGroupedAccessorList(ctx, settings.Order, node.AsTypeLiteralNode().Members.Nodes, nil)
			}
		}
		return listeners
	},
}

// checkGroupedAccessorClassBody runs the scan twice, once per staticness.
//
// Upstream's two `checkList` calls over the same class body. Splitting them is what makes a static
// accessor and an instance accessor of the same name two different properties rather than a split
// pair, and `class A { get a(){} static set a(foo){} }` is upstream's clean case for it.
func checkGroupedAccessorClassBody(ctx rule.Context, order GroupedAccessorPairsOrder, members []*ast.Node) {
	wantStatic := false
	checkGroupedAccessorList(ctx, order, members, &wantStatic)
	wantStatic = true
	checkGroupedAccessorList(ctx, order, members, &wantStatic)
}

// groupedAccessorEntry is one key's getters and setters, with the positions they appeared at.
type groupedAccessorEntry struct {
	key         string
	getterIndex int
	setterIndex int
	getterNode  *ast.Node
	setterNode  *ast.Node
	getterCount int
	setterCount int
}

// checkGroupedAccessorList gathers accessors by key and reports the split or misordered pairs.
//
// `wantStatic` nil means "check every member", which is the object-literal case. Non-nil selects one
// staticness, which is how the class body is scanned twice.
func checkGroupedAccessorList(ctx rule.Context, order GroupedAccessorPairsOrder,
	members []*ast.Node, wantStatic *bool) {
	entries := []*groupedAccessorEntry{}

	for index, member := range members {
		if member.Kind != ast.KindGetAccessor && member.Kind != ast.KindSetAccessor {
			continue
		}
		if wantStatic != nil && groupedAccessorIsStatic(member) != *wantStatic {
			continue
		}
		key, hasKey := groupedAccessorKeyText(ctx, member)
		if !hasKey {
			continue
		}

		var entry *groupedAccessorEntry
		for _, candidate := range entries {
			if candidate.key == key {
				entry = candidate
				break
			}
		}
		if entry == nil {
			entry = &groupedAccessorEntry{key: key, getterIndex: -1, setterIndex: -1}
			entries = append(entries, entry)
		}

		if member.Kind == ast.KindGetAccessor {
			entry.getterCount++
			entry.getterIndex = index
			entry.getterNode = member
			continue
		}
		entry.setterCount++
		entry.setterIndex = index
		entry.setterNode = member
	}

	for _, entry := range entries {
		// Exactly one of each. A duplicate accessor is code that is already wrong in a way this rule
		// has no opinion about, and upstream declines rather than guessing which one pairs.
		if entry.getterCount != 1 || entry.setterCount != 1 {
			continue
		}

		formerNode, latterNode := entry.getterNode, entry.setterNode
		if entry.setterIndex < entry.getterIndex {
			formerNode, latterNode = entry.setterNode, entry.getterNode
		}

		distance := entry.getterIndex - entry.setterIndex
		if distance < 0 {
			distance = -distance
		}
		if distance > 1 {
			reportGroupedAccessorPair(ctx, messageGroupedAccessorPairsNotGrouped,
				"Accessor pair %s and %s should be grouped.", formerNode, latterNode)
			continue
		}

		misordered := (order == GroupedAccessorPairsGetBeforeSet && entry.getterIndex > entry.setterIndex) ||
			(order == GroupedAccessorPairsSetBeforeGet && entry.getterIndex < entry.setterIndex)
		if misordered {
			reportGroupedAccessorPair(ctx, messageGroupedAccessorPairsInvalidOrder,
				"Expected %[2]s to be before %[1]s.", formerNode, latterNode)
		}
	}
}

// reportGroupedAccessorPair reports on the LATTER accessor's head, naming both.
//
// The format string takes the former name first and the latter second in both cases; the ordering
// message reverses them with an explicit index rather than by swapping the arguments, so the two
// call sites pass the same pair in the same order and only the template differs. Upstream does the
// same through its message data.
func reportGroupedAccessorPair(ctx rule.Context, message rule.Message, format string,
	formerNode *ast.Node, latterNode *ast.Node) {
	rendered := rule.Message{
		Id: message.Id,
		Description: fmt.Sprintf(format,
			groupedAccessorNameWithKind(formerNode), groupedAccessorNameWithKind(latterNode)),
	}
	ctx.ReportRange(groupedAccessorHeadRange(ctx, latterNode), rendered)
}

// groupedAccessorHeadRange is the accessor's head: the keyword through the key.
//
// `rule.TokenRange` starts past leading trivia and past nothing else, so for a static member it
// begins at `static` and for a plain one at `get` or `set`. The end is the key's own end, which
// stops before the parameter list. Measured against seven shapes.
func groupedAccessorHeadRange(ctx rule.Context, member *ast.Node) core.TextRange {
	start := rule.TokenRange(ctx.SourceFile, member).Pos()
	name := member.Name()
	if name == nil {
		// Not reachable for a get or set accessor, which the grammar requires to have a key, and
		// the callers only pass those. Guarded rather than asserted because a nil here would be a
		// panic that costs every rule this file, not just this one.
		return rule.TokenRange(ctx.SourceFile, member)
	}
	return core.NewTextRange(start, name.End())
}

// groupedAccessorIsStatic reports whether a class member carries `static`.
func groupedAccessorIsStatic(member *ast.Node) bool {
	modifiers := member.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindStaticKeyword {
			return true
		}
	}
	return false
}

// groupedAccessorKeyText returns the text two accessors are matched on.
//
// Upstream matches on `getStaticPropertyName` where the syntax settles the name, and falls back to
// comparing the key's TOKEN LIST otherwise. `property.Name` with `property.Static` is this tree's
// lift of that same decision, probed against nine key spellings before being built on and agreeing
// with the installed eslint on every one.
//
// The static half is wider than it looks and getting it wrong was this port's first defect. A
// computed key whose expression is a LITERAL still has a static name, so `set 123` pairs with
// `get [123]`, `set ['abc']` pairs with a template-keyed getter, and a computed key pairs with a plain one
// across the bracket. Four of upstream's own cases assert exactly that, and a first draft that
// prefixed computed keys to keep them apart failed all four.
//
// A computed key with a NON-literal expression settles nothing, and there upstream compares tokens
// so that `get [a]` and `set [a]` pair while `get [a]` and `set [b]` do not. Comparing the key's
// source bytes is that same comparison for every shape the corpus writes.
//
// The two halves are tagged apart, because a computed `[a]` compared by text must never collide
// with a static name that happens to read the same way.
func groupedAccessorKeyText(ctx rule.Context, member *ast.Node) (string, bool) {
	name := member.Name()
	if name == nil {
		return "", false
	}
	if staticName, settled := property.Name(name, property.Static); settled {
		// A private name is tagged apart from every other spelling. `property.Name` returns `#abc`
		// for `#abc`, hash included, and its doc comment says that can never collide with the
		// ordinary property `abc`. True, and not the whole story: it DOES collide with the string
		// key `'#abc'`, which upstream treats as a different property. Measured against the
		// installed eslint at 10.8.1, `class A { get '#abc'(){} b(){} set #abc(foo){} }` is CLEAN
		// while both same-spelling pairs report, and four of upstream's clean cases are exactly this
		// shape. The helper is right about what it documents; this rule needs one distinction more.
		if name.Kind == ast.KindPrivateIdentifier {
			return "private:" + staticName, true
		}
		return "static:" + staticName, true
	}
	// Not statically named, so compare the written bytes. `.Text()` PANICS on a computed property
	// name rather than returning empty, measured on a probe here before the rule existed, so this
	// branch must not reach for it.
	//
	// The `computed:` tag is DECORATIVE rather than load-bearing, and that was measured rather than
	// assumed. A mutation rewriting it to `static:` survives every fixture, including three written
	// specifically to kill it, and the reason is that the two branches already cannot collide: this
	// one returns the key's source text WITH its brackets, so `[a]` never equals the static branch's
	// `a` whatever either is tagged. There is no input that distinguishes the two versions.
	//
	// Kept because it says which question was asked, and because narrowing this range to the bracket
	// interior would make it load-bearing overnight with nothing to notice.
	tokenRange := rule.TokenRange(ctx.SourceFile, name)
	return "computed:" + ctx.SourceFile.Text()[tokenRange.Pos():tokenRange.End()], true
}

// groupedAccessorNameWithKind renders an accessor the way upstream's `getFunctionNameWithKind` does.
//
// Five forms, all measured against the installed eslint at 10.8.1 rather than inferred:
//
//	getter 'a'           a plain key, quoted
//	static setter 'a'    a static class member
//	private getter #p    a private name, NOT quoted and carrying its own hash
//	getter               a computed key, which cannot be named at all
//	getter '1'           a numeric key, quoted like a string
func groupedAccessorNameWithKind(member *ast.Node) string {
	kind := "getter"
	if member.Kind == ast.KindSetAccessor {
		kind = "setter"
	}

	name := member.Name()
	if name == nil {
		return kind
	}

	prefix := ""
	if groupedAccessorIsStatic(member) {
		prefix = "static "
	}

	if name.Kind == ast.KindPrivateIdentifier {
		// `.Text()` already carries the `#`, so adding one produces `##p`. Measured, and the
		// rendered-text fixtures are what catch it.
		return prefix + "private " + kind + " " + name.Text()
	}

	// The SAME static-name question the pairing uses, which is why a computed key that resolves is
	// named rather than left bare. A first draft rendered every computed key as a bare kind and
	// failed four of upstream's own cases: `` get [`abc`] `` renders as `getter 'abc'`, because
	// upstream reads the name through `getStaticPropertyName` here too.
	if staticName, settled := property.Name(name, property.Static); settled {
		return prefix + kind + " '" + staticName + "'"
	}
	// A computed key with a non-literal expression has no name upstream can render, so the kind
	// alone is the whole string.
	return prefix + kind
}
