package react

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDangerOptions is the rule's option surface.
//
// Upstream's `meta.schema` declares one object with one key, `customComponentNames`, an array of
// unique strings with no default, checked against the installed build rather than taken from the
// inventory column. Note the schema does NOT set `additionalProperties: false`, so an unknown key is
// accepted and ignored upstream; decoding into a struct does the same thing here.
type NoDangerOptions struct {
	// CustomComponentNames holds glob patterns naming components to check in addition to the DOM
	// elements the rule always checks. Absent means the empty list, which is upstream's `|| []`.
	CustomComponentNames []string
}

type noDangerWireOptions struct {
	CustomComponentNames []string `json:"customComponentNames"`
}

// DefaultNoDangerOptions is what an unconfigured rule uses: no custom components, DOM only.
func DefaultNoDangerOptions() NoDangerOptions {
	return NoDangerOptions{}
}

// DecodeNoDangerOptions is hand-rolled rather than `rule.DecodeOptionsInto`.
//
// The generic decoder ERRORS on empty input, and a rule configured as a bare `"error"` is handed
// nil options. For a non-required rule the config layer turns that error into nil, and the type
// assertion on nil then yields the zero value, so the generic helper would happen to work here.
// It works by accident rather than by design, and the accident is exactly the failure the port
// brief describes costing a rule 3,407 inert registrations. Answering the default explicitly on
// empty input makes the nil path a decision rather than a coincidence, and
// `TestNoDangerDecodesOptions` exercises it directly.
func DecodeNoDangerOptions(raw []byte) (any, error) {
	options := DefaultNoDangerOptions()
	if len(raw) == 0 {
		return options, nil
	}

	// Lenient on purpose: eslint-plugin-react 7.37.5's schema leaves this object open (no
	// `additionalProperties`), so ESLint loads an extra key and so must this.
	var wire noDangerWireOptions
	if err := json.Unmarshal(raw, &wire); err != nil {
		return options, err
	}
	options.CustomComponentNames = wire.CustomComponentNames
	return options, nil
}

// dangerousPropertyName is the one attribute this rule knows about.
//
// Upstream keeps a one-element list and builds a lookup object from it, which is a shape ready for a
// second entry rather than a set with one member today. A single constant says the same thing and a
// second name should turn this back into a set rather than being appended as a second comparison.
const dangerousPropertyName = "dangerouslySetInnerHTML"

// NoDanger flags a dangerous property written on a DOM element.
//
//	valid:   <App dangerouslySetInnerHTML={{ __html: "" }} />
//	valid:   <div className="bar"></div>
//	invalid: <div dangerouslySetInnerHTML={{ __html: "" }}></div>
//	invalid: <App dangerouslySetInnerHTML={{ __html: "" }} />    with customComponentNames ["*"]
//
// `dangerouslySetInnerHTML` hands React a raw string and tells it to skip escaping, so any value
// reaching it that a user can influence is a cross-site scripting hole. React named it this way on
// purpose. Render the content as children, or sanitize it and say at the call site who did.
//
// Ported from `eslint-plugin-react`'s `no-danger.js`. `meta.fixable` is absent, and there is nothing
// to write anyway: the repair is deciding whether the string is trusted, which no linter can do.
//
// # What counts as a DOM element, and it is a first-character test on a RENDERED name
//
// Upstream asks `jsxUtil.isDOMComponent`, which renders the tag through `jsx-ast-utils`'s
// `elementType` and tests `/^[a-z]/` against the result. That is a test on the rendered string
// rather than on the node kind, and two of its consequences look like defects and are not.
// Measured on the installed build:
//
//	<div dangerouslySetInnerHTML={...} />        REPORTS   the ordinary case
//	<x-custom dangerouslySetInnerHTML={...} />   REPORTS   a custom element still starts lowercase
//	<a.b dangerouslySetInnerHTML={...} />        REPORTS   rendered "a.b", first character is a
//	<div:ns dangerouslySetInnerHTML={...} />     REPORTS   rendered "div:ns", same
//	<App dangerouslySetInnerHTML={...} />        SILENT    a component, by the capital
//	<A.B dangerouslySetInnerHTML={...} />        SILENT    rendered "A.B"
//
// The dotted and namespaced lines report because `elementType` joins the parts with `.` or `:` and
// the regex only ever looks at character zero. A port testing "is the tag a lowercase Identifier"
// would be silent on both and pass every imported fixture, because the corpus writes neither.
//
// # The attribute name test is exact and case-sensitive
//
//	dangerouslySetInnerHTML       REPORTS
//	dangerouslysetinnerhtml       SILENT
//	DANGEROUSLYSETINNERHTML       SILENT
//	ns:dangerouslySetInnerHTML     SILENT   a namespaced attribute has no plain name
//
// The value is never examined. A bare `<div dangerouslySetInnerHTML />` and a string-valued
// `<div dangerouslySetInnerHTML="x" />` both report, measured, which matters because reaching for
// the expression would be the natural way to write this and would panic on the string form.
//
// # customComponentNames, its glob dialect, and two upstream defects
//
// The option holds minimatch patterns compared against a name upstream builds itself, and that
// name is NOT the one `elementType` renders. The rule computes
// `nodeName.name || nodeName.object.name + "." + nodeName.property.name`, which is a different and
// shallower reading, and the difference produces two real bugs. Both were measured by reproducing
// that expression against real parses and then confirmed against the running rule:
//
//	<App />       -> "App"            fine
//	<A.B />       -> "A.B"            fine
//	<A.B.C />     -> "undefined.C"    DEFECT: the nested object is not recursed
//	<ns:App />    -> the name NODE    DEFECT: not a string at all
//
// The third line is confirmed behaviorally: with `customComponentNames: ["A.B.C"]` the element
// `<A.B.C dangerouslySetInnerHTML={...} />` is SILENT, and with `["undefined.C"]` it REPORTS.
//
// The fourth line CRASHES upstream. minimatch is handed an object, calls `.split` on it, and throws
// a TypeError that takes down the whole lint run for that file. Reproduced directly against the
// installed build with `<Ns:App dangerouslySetInnerHTML={...} />` and any `customComponentNames`.
//
// The CAPITAL namespace is load-bearing and I had it wrong first; a fixture caught it. A lowercase
// `<ns:App />` renders as `ns:App`, whose first character is lowercase, so the DOM branch above
// already accepts it and the option path is never reached. Measured: `<ns:App>` reports with NO
// options at all, and only `<Ns:App>` gets far enough to crash. So the crash needs a tag that is
// namespaced AND capitalised, which is why it survived upstream long enough to find.
//
// Neither defect is reproduced here, and both declines are deliberate rather than silent:
//
//	the nested member renders fully, so `<A.B.C />` matches the pattern `A.B.C`
//	a namespaced tag renders as `ns:App` and matches or does not, without crashing
//
// This is the brief's "a port is not obliged to carry a defect it can see", and the crash is the
// clearest case of it available: reproducing a TypeError faithfully would mean taking a file away
// from all 289 rules, since the walk recovers per FILE rather than per rule. The intuitive reading
// and the measured one are both recorded so the next reader meets a stated decision. The cost is a
// divergence on two shapes nobody writes, in the direction of reporting slightly more and crashing
// never.
//
// # The glob dialect, measured rather than assumed
//
// Upstream uses minimatch; this uses `path.Match`, which is also the format walk's glob
// (`internal/format/formatfiles/enumerate.go`). They are different implementations, so the substitution was
// measured rather than argued: minimatch's answers were enumerated over 204 pattern-and-name pairs
// covering every pattern the corpus uses plus the dotted, namespaced and character-class shapes this
// rule can produce, and `filepath.Match` was run over the same table, on Unix, where it is
// `path.Match`. **Zero disagreements.**
//
// The probe asserted its verdicts rather than logging them, because a probe that only logs reports
// `ok` while carrying disagreements. It was deleted after the measurement; the number is recorded
// here because the table is the evidence and the count is what survives.
//
// One agreement worth naming, since it looks wrong: `MUI*` does NOT match `TextMUI`, and upstream's
// corpus has exactly that pair in its VALID list. Both matchers say false for the same reason, and a
// port that "fixed" it would break an imported clean case.
var NoDanger = rule.Rule{
	Name: "react/no-danger",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoDangerOptions](options)
		if !ok {
			// A rule configured as a bare severity reaches here with nil rather than with the
			// struct. Falling back to the documented default keeps that path a decision.
			settings = DefaultNoDangerOptions()
		}

		checkAttributes := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if tagName == nil || attributes == nil {
				return
			}
			properties := attributes.AsJsxAttributes().Properties
			if properties == nil {
				return
			}

			renderedName := renderJsxTagName(tagName)
			if !isDomElementName(renderedName) && !matchesAnyPattern(settings.CustomComponentNames, renderedName) {
				return
			}

			for _, property := range properties.Nodes {
				// No kind test here, deliberately. A first version guarded on
				// `KindJsxAttribute` before calling `jsx.AttributeName`, and a mutation
				// neutralizing that guard SURVIVED. It is subsumed rather than unreached:
				// `AttributeName`'s own first line rejects every kind but `KindJsxAttribute`, so
				// the readable flag can only ever be true for one, and no input can distinguish the
				// two versions. Deleted rather than covered by a fixture, since a fixture over a
				// branch that cannot change an answer asserts nothing.
				//
				// The behavior it looked like it was providing is real and still holds: a spread
				// carries no name to compare, so `<div {...props} />` is silent, measured upstream,
				// even though the spread may well supply the dangerous property. The rule reads
				// syntax rather than values, and that case is pinned in the attribute boundary
				// table.
				name, readable := jsx.AttributeName(property)
				if !readable || name != dangerousPropertyName {
					continue
				}
				ctx.ReportNode(property, rule.Message{
					Id: "dangerousProp",
					Description: fmt.Sprintf(
						"This element sets %s, which hands React a raw string and tells it to "+
							"skip escaping. React named the property this way on purpose: any "+
							"value reaching it that a user can influence is a cross-site "+
							"scripting hole, and the danger is invisible at the render site "+
							"because the markup arrives from somewhere else. Render the content "+
							"as children instead, or sanitize it and say at the call site who "+
							"did.",
						name,
					),
				})
			}
		}

		return rule.Listeners{
			// Both kinds. Every reporting fixture upstream ships is one or the other, and our
			// parser gives a self-closing element its own kind with no opening element inside it.
			ast.KindJsxOpeningElement:     checkAttributes,
			ast.KindJsxSelfClosingElement: checkAttributes,
		}
	},
}

// renderJsxTagName renders a tag the way `jsx-ast-utils`'s `elementType` does.
//
// A plain identifier is its own text. A member access joins its parts with `.`, RECURSING through
// the object, so `<A.B.C />` renders `A.B.C`. A namespaced name joins with `:`.
//
// The recursion is where this deliberately diverges from one of upstream's two readers. `elementType`
// recurses and the rule's own inline `functionName` expression does not, so upstream renders the
// same element two different ways depending on which question it is asking, and the shallow one
// produces `undefined.C`. This function is the recursing reading, used for both questions. See the
// rule's doc comment for the measurements and for why the defect is declined rather than reproduced.
//
// **`Text()` is never called on the tag node itself, deliberately.** `ast.Node.Text` has no case for
// a PropertyAccessExpression and reaches its unhandled-case panic, measured with a probe, and the
// walk recovers per FILE rather than per rule. The parts are read individually instead.
func renderJsxTagName(tagName *ast.Node) string {
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.Text()

	case ast.KindJsxNamespacedName:
		namespaced := tagName.AsJsxNamespacedName()
		name := namespaced.Name()
		if namespaced.Namespace == nil || name == nil {
			return ""
		}
		return namespaced.Namespace.Text() + ":" + name.Text()

	case ast.KindPropertyAccessExpression:
		access := tagName.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return ""
		}
		object := access.Expression
		if object == nil {
			return ""
		}
		// `<this.Foo />` is a real shape our parser produces and upstream's `elementType` cannot,
		// since its tree has no ThisExpression in a tag name. Rendered as an empty object part,
		// which makes it a component by the capital and keeps it out of the DOM branch either way.
		rendered := renderJsxTagName(object)
		return rendered + "." + name.Text()
	}
	return ""
}

// isDomElementName reproduces upstream's `COMPAT_TAG_REGEX`, which is `/^[a-z]/`.
//
// A first-character test on the RENDERED name, not a check for an intrinsic element. So `x-custom`,
// `a.b` and `div:ns` are all DOM elements to this rule, and all three are measured reporting
// upstream. The comparison is ASCII rather than `unicode.IsLower`, matching the regex's own
// character class: upstream's `[a-z]` does not accept a lowercase non-ASCII letter, so neither does
// this. Nothing in any corpus writes one, which is precisely why the divergence would never surface
// once shipped.
func isDomElementName(rendered string) bool {
	if rendered == "" {
		return false
	}
	first := rendered[0]
	return first >= 'a' && first <= 'z'
}

// matchesAnyPattern reports whether a rendered tag name matches any configured glob.
//
// `path.Match` stands in for minimatch. Measured equivalent over 204 pattern-and-name pairs
// covering every pattern the corpus uses and every tag shape this rule can produce, with zero
// disagreements; see the rule's doc comment.
//
// A malformed pattern makes `filepath.Match` return an error, where minimatch treats the pattern as
// literal text. Both answer "no match" for the shapes anyone writes, and the error is discarded
// rather than surfaced because a rule has no channel for a configuration complaint and reporting on
// every element in the file would be worse than ignoring one bad pattern.
func matchesAnyPattern(patterns []string, rendered string) bool {
	if rendered == "" {
		return false
	}
	for _, pattern := range patterns {
		// path.Match, which reads `\` as an escape on every platform. filepath.Match does not on
		// Windows, so a pattern like `MUI\*` meant something else there than where it was measured.
		if matched, err := path.Match(pattern, rendered); err == nil && matched {
			return true
		}
	}
	return false
}
