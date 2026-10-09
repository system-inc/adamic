package react

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// forbidDomPropsMessage builds the finding for one forbidden prop.
//
// Upstream carries one message id and one interpolated slot, plus an escape hatch: an entry may
// supply its own `message`, in which case the finding carries that text and NO message id at all.
// The two are told apart by truthiness, so an EMPTY message string takes the default branch rather
// than reporting an empty finding. Measured on the installed build.
//
// A finding with no id would be invisible to `ExpectFindings`, which asserts ids, so the custom
// branch keeps the id and swaps only the text. That is a deliberate divergence in the shape of the
// report rather than in the decision: the same inputs report, with the same text, and the id stays
// assertable. Recorded here because a reader comparing to upstream will see the difference.
func forbidDomPropsMessage(prop string, custom string) rule.Message {
	if custom != "" {
		return rule.Message{Id: "propIsForbidden", Description: custom}
	}
	return rule.Message{
		Id:          "propIsForbidden",
		Description: fmt.Sprintf("Prop %q is forbidden on DOM Nodes", prop),
	}
}

// ForbidDomPropsOptions is the decoded option object.
//
// One property, `forbid`, whose items are a union of a bare string and an object. The union is why
// this needs a hand-rolled decoder rather than `rule.DecodeOptionsInto`.
type ForbidDomPropsOptions struct {
	// Forbid is the list of prop names to reject on DOM nodes.
	//
	// Order matters and later entries win. Upstream builds a lookup keyed by prop name and assigns
	// into it in list order, so a name written twice keeps the LAST entry's restrictions. Measured.
	Forbid []ForbidDomPropsEntry
}

// ForbidDomPropsEntry is one prop on the forbidden list.
type ForbidDomPropsEntry struct {
	// PropName is the attribute name to reject, matched exactly.
	PropName string

	// DisallowedFor narrows the entry to a set of tag names.
	//
	// Nil means every DOM tag. An EMPTY non-nil list means NO tag, which is not the same thing and
	// is the edge this field's pointer-free spelling would lose. Upstream tests the list for
	// truthiness and an empty array is truthy in JavaScript, so it falls through to a membership
	// test that nothing can satisfy. `<div className="x" />` with `disallowedFor: []` is CLEAN.
	// Measured on the installed build. See DisallowedForWasGiven.
	DisallowedFor []string

	// DisallowedForWasGiven separates an absent list from an empty one.
	//
	// Go cannot tell `nil` from `[]string{}` reliably once a slice has round-tripped, and the two
	// mean opposite things here: absent means every tag, empty means no tag. This flag carries the
	// distinction the decoder saw.
	DisallowedForWasGiven bool

	// Message replaces the finding's text when it is non-empty.
	Message string
}

// forbidDomPropsWireEntry decodes the object arm of the schema's union.
type forbidDomPropsWireEntry struct {
	PropName      string    `json:"propName"`
	DisallowedFor *[]string `json:"disallowedFor"`
	Message       string    `json:"message"`
}

// forbidDomPropsWireOptions is the JSON shape.
type forbidDomPropsWireOptions struct {
	Forbid []json.RawMessage `json:"forbid"`
}

// DecodeForbidDomPropsOptions turns the configured JSON into the struct the rule reads.
//
// Exported so fixtures drive the same path the config drives, which is what puts the union handling
// and the empty-list distinction under test rather than assumed.
//
// A rule configured as a bare severity is handed nil options, and the empty body has to produce
// upstream's default of an EMPTY forbid list rather than an error. An empty list is the rule doing
// nothing, which is upstream's behaviour with no options: it reads the configured list or a shared
// empty default. That is the safe direction for this rule, since the whole judgment is supplied by
// the config.
func DecodeForbidDomPropsOptions(raw []byte) (any, error) {
	options := ForbidDomPropsOptions{}
	if len(raw) == 0 {
		return options, nil
	}

	var wire forbidDomPropsWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}

	for _, item := range wire.Forbid {
		var asString string
		if err := json.Unmarshal(item, &asString); err == nil {
			options.Forbid = append(options.Forbid, ForbidDomPropsEntry{PropName: asString})
			continue
		}
		// Lenient on purpose: eslint-plugin-react 7.37.5's schema leaves a forbid entry object open (no
		// `additionalProperties`), so ESLint loads `{"propName": "id", "extra": 1}` and so must this.
		var asObject forbidDomPropsWireEntry
		if err := json.Unmarshal(item, &asObject); err != nil {
			return options, fmt.Errorf("decoding a forbid entry: %w", err)
		}
		entry := ForbidDomPropsEntry{
			PropName: asObject.PropName,
			Message:  asObject.Message,
		}
		// The pointer is what separates an absent key from an empty array, and the two mean
		// opposite things. Upstream's `value.disallowedFor || null` maps absent to null, meaning
		// every tag, while an empty array survives as a truthy list that matches no tag.
		if asObject.DisallowedFor != nil {
			entry.DisallowedFor = *asObject.DisallowedFor
			entry.DisallowedForWasGiven = true
		}
		options.Forbid = append(options.Forbid, entry)
	}
	return options, nil
}

// forbidDomPropsIsDomNodeName reports whether a JSX tag name is a DOM node rather than a component.
//
// Upstream's test is `tag[0] !== tag[0].toUpperCase()`, which asks whether the first character
// CHANGES under upper-casing rather than whether it is lowercase. The difference shows on a
// character with no case: an underscore upper-cases to itself, so `<_div someProp="x" />` is CLEAN.
// Measured on the installed build, and no corpus case writes it.
//
// A digit behaves the same way and for the same reason, though the parser will not accept a tag
// starting with one. The predicate is written as upstream wrote it rather than as "is lowercase",
// because the two disagree on exactly the characters nobody thinks to test.
//
// This is deliberately NOT `jsx.IsIntrinsicElementNamed`, which asks whether a tag is one specific
// named element. The question here is the general one, and no shelf helper asks it.
func forbidDomPropsIsDomNodeName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	// ASCII upper-casing, matching upstream over a name the JSX grammar keeps to identifier
	// characters. A character outside a-z is unchanged, which is what makes the underscore clean.
	upper := first
	if first >= 'a' && first <= 'z' {
		upper = first - 32
	}
	return first != upper
}

// forbidDomPropsIsForbidden reports whether one prop on one tag is on the list.
//
// Upstream's two-part test: the prop must be on the list at all, and either the entry names no tag
// restriction or the tag must be among the named ones.
func forbidDomPropsIsForbidden(entry ForbidDomPropsEntry, tagName string) bool {
	if !entry.DisallowedForWasGiven {
		return true
	}
	for _, allowed := range entry.DisallowedFor {
		if allowed == tagName {
			return true
		}
	}
	return false
}

// ForbidDomProps flags a prop the project's configuration has forbidden on DOM nodes.
//
//	valid:   <div className="x" />                  with no forbid list
//	valid:   <Foo id="x" />                         a component, not a DOM node
//	valid:   <span id="x" />                        with id disallowedFor ["div"]
//	valid:   <div id="x" />                         with id disallowedFor [] (matches no tag)
//	invalid: <div id="x" />                         with forbid ["id"]
//	invalid: <div id />                             the value is never read
//	invalid: <div id={expression} />                likewise
//
// Ported from `react/forbid-dom-props` in `eslint-plugin-react`. One option holding a union-typed
// list, one message, no fixer.
//
// # The clone and the installed build disagree, and this port follows the installed build
//
// The clone's working tree carries a `disallowedValues` feature and a second message id,
// `propIsForbiddenWithValue`, that the INSTALLED build does not have. Both claim version 7.37.5;
// the clone is the unreleased default branch. Verified two ways: the installed rule file contains
// neither identifier, and replaying upstream's own corpus against the installed build disagrees on
// exactly the three cases that exercise the feature. Upstream's `valid 8` and `valid 10` report,
// and its `invalid 10` reports six findings where the corpus says five.
//
// This port reproduces the INSTALLED build, because that is the artifact our gate compares against
// and the one this repository actually runs. The consequence is that a `disallowedValues` key in a
// configuration is accepted by the schema and ignored, exactly as the installed build ignores it.
// The three corpus cases are carried at the installed build's verdict with the clone's verdict
// named beside each.
//
// # The value is never read, which is the opposite of a sibling rule
//
// `jsx-no-script-url` decides from an attribute's string value and therefore declines a bare prop
// and an expression container. This rule decides from the NAME alone, so `<div id />` and
// `<div id={x} />` both report. Measured on the installed build. Reading the value here would
// silently narrow the rule to string-valued props.
//
// The clone's newer body does read the value, unguarded, which would crash on a bare prop. Not
// reproduced, since the installed build does not have that code at all.
//
// # An empty disallowedFor list matches nothing
//
// Upstream tests the list for truthiness and an empty array is truthy, so it falls through to a
// membership test nothing satisfies. An absent list means every DOM tag; an empty one means none.
// The decoder keeps the two apart with a flag, because Go's zero value cannot. Measured.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue rather than
// upstream behaviour. This rule is ported from the authority, which has no such gate.
var ForbidDomProps = rule.Rule{
	Name: "react/forbid-dom-props",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[ForbidDomPropsOptions](options)

		// Assignment in list order reproduces upstream's map construction, so a repeated prop name
		// keeps the last entry.
		forbidden := make(map[string]ForbidDomPropsEntry, len(settings.Forbid))
		for _, entry := range settings.Forbid {
			forbidden[entry.PropName] = entry
		}

		return rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				// The attribute must carry a plain identifier name. This declines a spread and a
				// namespaced name, matching upstream reading the name's own name.
				attributeName, named := jsx.AttributeName(node)
				if !named {
					return
				}

				// The parent chain is attribute, then the attribute list, then the element.
				// Upstream reads the attribute's parent directly because its tree has no
				// attribute-list node in between.
				attributes := node.Parent
				if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
					return
				}
				tagName, _ := jsx.ElementParts(attributes.Parent)
				// A member or namespaced tag declines here, which is upstream's answer: it reads a
				// plain name that neither shape carries, and its guard then rejects the undefined.
				if tagName == nil || tagName.Kind != ast.KindIdentifier {
					return
				}
				tag := tagName.Text()
				if !forbidDomPropsIsDomNodeName(tag) {
					return
				}

				entry, listed := forbidden[attributeName]
				if !listed || !forbidDomPropsIsForbidden(entry, tag) {
					return
				}

				// Upstream reports on the whole attribute, so the span covers the name and any
				// value together.
				ctx.ReportNode(node, forbidDomPropsMessage(attributeName, entry.Message))
			},
		}
	},
}
