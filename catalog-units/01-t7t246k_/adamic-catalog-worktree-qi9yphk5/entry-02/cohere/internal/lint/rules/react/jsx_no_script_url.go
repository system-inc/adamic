package react

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// jsxNoScriptUrlMessage is the single finding this rule produces.
//
// One message id, no interpolation, so there is nothing to render and the text is a constant.
// Upstream's wording is kept because it names the reason rather than the rule: a future React
// version blocks these, so the finding is about a deprecation with a security motive rather than
// about style.
var jsxNoScriptUrlMessage = rule.Message{
	Id: "noScriptURL",
	Description: "A future version of React will block javascript: URLs as a security precaution. " +
		"Use event handlers instead if you can. If you need to generate unsafe HTML, try using " +
		"dangerouslySetInnerHTML instead.",
}

// JsxNoScriptUrlOptions is the decoded option surface.
//
// Upstream's schema is an `anyOf` over two positional lists, which is why this needs a hand-rolled
// decoder:
//
//	[[{name, props}], {includeFromSettings}]    the legacy array, optionally followed by the object
//	[{includeFromSettings}]                     the object alone
//
// `includeFromSettings` has no field here. Upstream reads `settings.linkComponents` from ESLint's
// shared settings when it is true, and cohere does not read settings.linkComponents, so `true` cannot be
// honoured and the decoder refuses it by name; `false` is what the rule already does and is
// accepted.
type JsxNoScriptUrlOptions struct {
	// Legacy is the list of component-and-attribute pairs to check, from the array arm.
	//
	// Order matters and later entries win, which the corpus does not state and which was measured
	// against the installed build both ways. An entry naming a component twice keeps the LAST
	// entry's attribute list, and an entry naming `a` REPLACES the built-in `href` default rather
	// than adding to it.
	Legacy []JsxNoScriptUrlComponent
}

// JsxNoScriptUrlComponent is one component-and-attributes pair from the legacy array.
type JsxNoScriptUrlComponent struct {
	// Name is the JSX tag name to check, matched exactly.
	Name string

	// Props are the attribute names on that tag whose string values are examined.
	Props []string
}

// jsxNoScriptUrlWireComponent is the JSON shape of one legacy entry.
type jsxNoScriptUrlWireComponent struct {
	Name  string   `json:"name"`
	Props []string `json:"props"`
}

// jsxNoScriptUrlWireObject is the JSON shape of the object arm.
type jsxNoScriptUrlWireObject struct {
	IncludeFromSettings bool `json:"includeFromSettings"`
}

// DecodeJsxNoScriptUrlOptions turns upstream's option list into the struct the rule reads.
//
// Exported so fixtures drive the same path the config drives, which is what puts the positional
// union under test rather than assumed.
//
// The rule registers with `DecodeOptionList`, so it is handed upstream's own list:
//
//	[[{"name": "Foo", "props": ["to"]}], {"includeFromSettings": false}]
//
// The config layer used to hand a decoder the first option element alone, so this decoder read the
// list from under a cohere-only `{"positional": [...]}` wrapper. The wrapper is gone and refused:
// it was never upstream's spelling and the config layer no longer needs it.
//
// The array arm and the object arm are told apart the way upstream tells them apart, by whether the
// first element is an array. An object first takes no second element. Each legacy entry and the
// object are read strictly, a key outside the schema refused, and `includeFromSettings: true` is
// refused because nothing here can honour it; see JsxNoScriptUrlOptions.
//
// A rule configured as a bare severity is handed nil options, and the empty body has to produce
// upstream's default of the built-in pair alone rather than an error.
func DecodeJsxNoScriptUrlOptions(list []byte) (any, error) {
	options := JsxNoScriptUrlOptions{}
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return options, err
	}

	objectElement := elements[0]
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	if bytes.HasPrefix(bytes.TrimSpace(elements[0]), []byte("[")) {
		var legacy []jsxNoScriptUrlWireComponent
		if err := jsxNoScriptUrlDecodeStrictly(elements[0], &legacy); err != nil {
			return JsxNoScriptUrlOptions{}, fmt.Errorf("jsx-no-script-url element 1: %w", err)
		}
		for _, entry := range legacy {
			options.Legacy = append(options.Legacy, JsxNoScriptUrlComponent{
				Name:  entry.Name,
				Props: entry.Props,
			})
		}
		if len(elements) < 2 {
			return options, nil
		}
		objectElement = elements[1]
	} else if len(elements) > 1 {
		return JsxNoScriptUrlOptions{}, fmt.Errorf(
			"jsx-no-script-url: an options object first takes no second element, so %s would "+
				"never be read", elements[1])
	}

	var object jsxNoScriptUrlWireObject
	if err := jsxNoScriptUrlDecodeStrictly(objectElement, &object); err != nil {
		return JsxNoScriptUrlOptions{}, fmt.Errorf("jsx-no-script-url options object: %w", err)
	}
	if object.IncludeFromSettings {
		return JsxNoScriptUrlOptions{}, fmt.Errorf("jsx-no-script-url: includeFromSettings: true " +
			"reads ESLint's shared settings.linkComponents, and cohere does not read settings.linkComponents, so " +
			"it would be accepted and never honoured. List the link components in this rule's own " +
			`first option instead: ["error", [{"name": "Link", "props": ["to"]}]]`)
	}
	return options, nil
}

// jsxNoScriptUrlDecodeStrictly decodes one element and refuses keys its schema does not declare.
func jsxNoScriptUrlDecodeStrictly(element json.RawMessage, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(element))
	decoder.DisallowUnknownFields()
	return decoder.Decode(into)
}

// jsxNoScriptUrlIsInterleavingWhitespace reports whether a character may sit BETWEEN the letters of
// the protocol.
//
// Upstream's regex writes a carriage-return, newline and tab class between every letter, so only
// those three are accepted there. A plain space inside the word does NOT match, which is the
// distinction a port is most likely to get wrong by reusing the wider prefix class below. Measured
// against the installed build: a space inside the word is clean and a newline inside it reports.
func jsxNoScriptUrlIsInterleavingWhitespace(character byte) bool {
	return character == '\r' || character == '\n' || character == '\t'
}

// jsxNoScriptUrlIsLeadingWhitespace reports whether a character may sit BEFORE the protocol.
//
// Upstream's prefix class spans every code unit from zero through 0x1F, plus the space, which is
// wider than the interleaving class. Measured at the edges against the installed build: 0x01, 0x09,
// 0x1F and the space all report, and 0x7F does not.
//
// The comparison is over bytes rather than runes, and that is safe rather than a shortcut: every
// character in the accepted set is below 0x80, so no byte of a multi-byte sequence can be mistaken
// for one. A non-ASCII prefix character fails the test on its first byte, which is the right answer.
func jsxNoScriptUrlIsLeadingWhitespace(character byte) bool {
	return character <= 0x1F || character == ' '
}

// jsxNoScriptUrlHasJavaScriptProtocol reports whether a string value is a `javascript:` URL, by the
// same reading React's own sanitizer uses.
//
// Upstream ports the pattern verbatim from react-dom's `sanitizeURL.js`: a prefix class, then each
// letter of the word separated by an optional carriage-return, newline or tab class, then the
// colon, matched case-insensitively and anchored at the start.
//
// Written as a scan rather than as a pattern because the shape is a fixed sequence with optional
// separators, which a scan states more directly, and because the pattern is anchored so no
// backtracking is involved. The case-insensitivity is upstream's, and it is ASCII-only here for the
// same reason the byte test above is safe: every letter in the word is ASCII, so no Unicode case
// folding can produce one.
//
// Note the tail. The separator class applies after the final letter as well, before the colon, so a
// tab sitting between the word and its colon still reports. Measured.
func jsxNoScriptUrlHasJavaScriptProtocol(value string) bool {
	index := 0
	for index < len(value) && jsxNoScriptUrlIsLeadingWhitespace(value[index]) {
		index++
	}

	for _, wanted := range []byte{'j', 'a', 'v', 'a', 's', 'c', 'r', 'i', 'p', 't', ':'} {
		if index >= len(value) {
			return false
		}
		character := value[index]
		if wanted != ':' {
			// ASCII case folding, matching upstream's case-insensitive flag over letters that are
			// all ASCII.
			character |= 0x20
		}
		if character != wanted {
			return false
		}
		index++
		// The separator class applies after every element including the final letter, which is
		// what lets a tab sit between the word and its colon.
		for index < len(value) && jsxNoScriptUrlIsInterleavingWhitespace(value[index]) {
			index++
		}
	}
	return true
}

// jsxNoScriptUrlLinkComponents builds the tag-to-attributes lookup the rule consults.
//
// Upstream builds a map seeded with the built-in `a` to `href` pair and then assigns each legacy
// entry into it, so two behaviours follow from that assignment and neither is in the corpus. Both
// were measured against the installed build:
//
//	an entry named `a` REPLACES the built-in href default   `<a href="javascript:">` goes clean
//	a name written twice keeps the LAST entry               only the second prop list applies
//
// A port that merged instead of replaced, or that kept the first entry, would answer the opposite
// on both and every corpus case would still pass.
func jsxNoScriptUrlLinkComponents(options JsxNoScriptUrlOptions) map[string][]string {
	components := map[string][]string{"a": {"href"}}
	for _, entry := range options.Legacy {
		components[entry.Name] = entry.Props
	}
	return components
}

// JsxNoScriptUrl flags a `javascript:` URL written into a link attribute in JSX.
//
//	valid:   <a href="https://reactjs.org"></a>
//	valid:   <a href="#"></a>
//	valid:   <a href={"javascript:"}></a>          not a string literal attribute value
//	valid:   <a href />                            no value to read
//	valid:   <Foo href="javascript:"></Foo>        Foo is not a configured link component
//	invalid: <a href="javascript:"></a>
//	invalid: <a href="javascript:void(0)"></a>
//	invalid: <Foo to="javascript:"></Foo>          with Foo configured for `to`
//
// Ported from `react/jsx-no-script-url` in `eslint-plugin-react`, read from the clone at
// `lib/rules/jsx-no-script-url.js`. One positional option surface, one message, no fixer. All 22
// corpus cases were replayed against the installed build, version 7.37.5, through the ESLint Linter
// API before any code was written, and it agreed on all 22, counts and message ids alike.
// Everything below that the corpus does not state was measured the same way.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue rather than
// upstream behaviour. This rule is ported from the authority, which has no such gate, so a `.ts`
// file holding JSX is examined. A fixture pins that.
//
// # The shared-settings half of the option surface has no substrate here, and it is refused
//
// Upstream reads `settings.linkComponents` from ESLint's shared settings when `includeFromSettings`
// is true. cohere has no shared-settings surface anywhere, so that half of the rule cannot be
// expressed, and a config writing `includeFromSettings: true` is refused at startup rather than
// accepted and ignored. Measured, this is what the missing surface costs on upstream's own corpus:
//
//	valid 9, 10        no `includeFromSettings: true`, so imported unchanged
//	valid 11           clean with the settings, and its configuration is now refused
//	invalid 6, 7       report ONLY because of settings, and their configuration is now refused
//	invalid 8          two findings upstream, its configuration is now refused
//
// Those four are recorded in `TestJsxNoScriptUrlRefusesIncludeFromSettings` with the upstream
// verdict named beside each, rather than deleted. Deleting them would hide a real gap; the gap is a
// missing configuration surface, not a defect in the judgment. Anything a user configures through the
// legacy array is reproduced exactly.
//
// # What counts as a value
//
// Upstream requires the attribute value to be a Literal, so only a plain string attribute is read.
// An expression container declines even when its content is a string literal, which upstream's
// corpus pins directly, and a bare `<a href />` has no value at all.
// `jsx.StringAttributeValue` asks the same question, but it searches an attribute list by name;
// this rule anchors on the attribute itself, so it reads the initializer directly.
//
// # A namespaced attribute is invisible, and that is upstream's answer too
//
// Upstream reads the attribute's name and then that name's own name, which is absent for a
// namespaced attribute, so `<a xlink:href="javascript:">` is clean. `jsx.AttributeName` declines
// the same shape for the same reason, so the shelf helper and upstream agree here. Measured on the
// installed build.
//
// # The tag name must be a plain identifier
//
// The parent's name is read the same way, so a member tag and a namespaced tag are both clean,
// since neither carries a plain name under its name. Both measured. A port reading source text for
// the tag would report on inputs upstream ignores.
var JsxNoScriptUrl = rule.Rule{
	Name: "react/jsx-no-script-url",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[JsxNoScriptUrlOptions](options)
		components := jsxNoScriptUrlLinkComponents(settings)

		return rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				// The attribute must carry a plain identifier name. This declines a spread and a
				// namespaced name, matching upstream reading the name's own name.
				attributeName, named := jsx.AttributeName(node)
				if !named {
					return
				}

				// The parent chain is attribute, then the attribute list, then the opening or
				// self-closing element. Upstream reads the attribute's parent directly because its
				// tree has no attribute-list node in between. Walking two levels here reaches the
				// same element rather than a different one.
				//
				// The KIND half of this guard is unreachable and is kept as crash protection
				// rather than as a filter, which is why a mutant removing it survives every
				// fixture. Probed with a rule collecting the parent kind of every JsxAttribute
				// over nine attributes across eight sources, including a spread beside an
				// attribute, a nested element inside an attribute value, and four malformed
				// inputs the parser recovers from: the parent is `KindJsxAttributes` in all nine
				// and nothing else appeared. The NIL half is what earns its place, since the line
				// below dereferences this node.
				//
				// The enumeration is what makes the verdict expire: it names error recovery and a
				// nested-element value as the shapes considered. A parser change that gives an
				// attribute another parent voids it.
				attributes := node.Parent
				if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
					return
				}
				tagName, _ := jsx.ElementParts(attributes.Parent)
				// A member or namespaced tag name declines here, which is upstream's answer: it
				// reads a plain name that neither shape carries.
				if tagName == nil || tagName.Kind != ast.KindIdentifier {
					return
				}

				allowed, configured := components[tagName.Text()]
				if !configured {
					return
				}
				matched := false
				for _, candidate := range allowed {
					if candidate == attributeName {
						matched = true
						break
					}
				}
				if !matched {
					return
				}

				// Only a plain string literal is read, matching upstream's Literal test. An
				// expression container and an absent value both decline.
				initializer := node.AsJsxAttribute().Initializer
				if initializer == nil || initializer.Kind != ast.KindStringLiteral {
					return
				}
				if !jsxNoScriptUrlHasJavaScriptProtocol(initializer.Text()) {
					return
				}

				// Upstream reports on the whole attribute, so the span covers the name and the
				// value together rather than just one. Measured: an attribute at column 4 reports
				// through the closing quote of its value.
				ctx.ReportNode(node, jsxNoScriptUrlMessage)
			},
		}
	},
}
