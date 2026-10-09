package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageNoTargetBlankWithoutNoreferrer = rule.Message{
		Id: "noTargetBlankWithoutNoreferrer",
		Description: "This link opens a new tab without `rel=\"noreferrer\"`. The page it opens receives " +
			"a live `window.opener` handle back to this one and can navigate it anywhere, which is " +
			"how a link to an external site becomes a phishing redirect the user never sees " +
			"initiated. Add `rel=\"noreferrer\"`, which also implies `noopener` and is the wider " +
			"supported spelling.",
	}
	messageNoTargetBlankWithoutNoopener = rule.Message{
		Id: "noTargetBlankWithoutNoopener",
		Description: "This link opens a new tab without `rel=\"noreferrer\"` or `rel=\"noopener\"`. The " +
			"page it opens receives a live `window.opener` handle back to this one and can " +
			"navigate it anywhere. This project allows the referrer to be sent, so " +
			"`rel=\"noopener\"` is enough here; `rel=\"noreferrer\"` is still the safer default.",
	}
)

// JsxNoTargetBlankOptions is the decoded option object.
//
// Five fields, matching `meta.schema`'s five properties exactly. Four of them decide something. The
// fifth does not, upstream, and is carried anyway; see `Links` below.
type JsxNoTargetBlankOptions struct {
	// AllowReferrer accepts `rel="noopener"` alone, and switches which message is reported.
	//
	// False by default. The two messages are not two severities of one finding: one says noreferrer
	// is required and the other says either will do, and upstream picks between them on this flag
	// alone. Both ids are in the corpus, 46 and 4 times.
	AllowReferrer bool `json:"allowReferrer"`

	// EnforceDynamicLinks decides whether `href={expression}` counts as an external link.
	//
	// The schema is an enum of `always` and `never` with NO default, and the rule supplies
	// `configuration.enforceDynamicLinks || 'always'`. So the absent value means `always` and the
	// zero value of this field must too, which is why the decoder inverts it into a bool rather
	// than storing the string. See DecodeJsxNoTargetBlankOptions.
	EnforceDynamicLinksNever bool `json:"-"`

	// WarnOnSpreadAttributes treats a spread as possibly supplying target or href.
	//
	// False by default. It widens the rule in two directions at once: an element with a spread and
	// no literal `target="_blank"` becomes reportable, and a `rel` written BEFORE the spread stops
	// counting as secure. Both were measured.
	WarnOnSpreadAttributes bool `json:"warnOnSpreadAttributes"`

	// Links is decoded and, faithfully, decides nothing.
	//
	// `meta.schema` declares it with `default: true` and the rule's `Object.assign` sets it, and
	// then `configuration.links` is never read anywhere in the rule body. Grepped in the source and
	// measured on the installed build: `<a target="_blank" href="http://x">` reports identically
	// under `{links: false}`, `{links: true}`, and no options at all.
	//
	// Carried rather than dropped because it is a real key in upstream's schema: a config writing
	// `{"links": false}` must decode here rather than error, and it must not silence anything, which
	// is exactly what a decoded-and-unread field does. A fixture pins that it changes no verdict, so
	// a later reader who "fixes" this by wiring it up fails a test rather than silently diverging.
	Links bool `json:"links"`

	// Forms extends the rule to `<form target="_blank" action=...>`.
	//
	// False by default, which is upstream's default and the reason the form arm is invisible unless
	// asked for. The form arm also reports WITHOUT a fix, which is upstream's choice and not an
	// omission here; see the report site.
	Forms bool `json:"forms"`
}

// jsxNoTargetBlankWireOptions is the JSON shape, which is not the struct shape.
//
// Two fields need a translation layer and neither has an upstream counterpart to copy:
// `enforceDynamicLinks` is a string enum whose absent value is `always`, and `links` defaults to
// TRUE, so the Go zero value is wrong for it. Both are exactly the lines this project's brief says
// end up untested when a fixture builds the options struct directly, so the fixtures route through
// the exported decoder instead.
type jsxNoTargetBlankWireOptions struct {
	AllowReferrer          bool    `json:"allowReferrer"`
	EnforceDynamicLinks    *string `json:"enforceDynamicLinks"`
	WarnOnSpreadAttributes bool    `json:"warnOnSpreadAttributes"`
	Links                  *bool   `json:"links"`
	Forms                  bool    `json:"forms"`
}

// DecodeJsxNoTargetBlankOptions turns the configured JSON into the struct the rule reads.
//
// Exported so fixtures can drive the same path the config drives. Two defaults are inverted from
// Go's zero value and both are measured against the installed build:
//
//	enforceDynamicLinks absent  -> `always`, so a dynamic href DOES count
//	links absent                -> true, though nothing reads it either way
//
// A rule configured as a bare `"error"` is handed nil options, and `configuration.OptionsRegistry.Decode`
// turns an empty body into nil for a non-required rule, so the nil path has to produce upstream's
// defaults rather than the zero struct. It does, and a fixture bypasses the decoder to pin it.
func DecodeJsxNoTargetBlankOptions(raw []byte) (any, error) {
	options := JsxNoTargetBlankOptions{Links: true}
	if len(raw) == 0 {
		return options, nil
	}

	var wire jsxNoTargetBlankWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}

	options.AllowReferrer = wire.AllowReferrer
	options.WarnOnSpreadAttributes = wire.WarnOnSpreadAttributes
	options.Forms = wire.Forms
	// `|| 'always'` in upstream means anything that is not the literal `never` behaves as always,
	// including a value the schema would have rejected. Reproduced by testing for `never` rather
	// than by validating the enum.
	options.EnforceDynamicLinksNever = wire.EnforceDynamicLinks != nil && *wire.EnforceDynamicLinks == "never"
	if wire.Links != nil {
		options.Links = *wire.Links
	}
	return options, nil
}

// JsxNoTargetBlank flags a link that opens a new tab without a rel that severs the opener handle.
//
//	valid:   <a target="_blank" href="/relative"></a>
//	valid:   <a target="_blank" href="http://x.com" rel="noreferrer"></a>
//	valid:   <a href="http://x.com"></a>
//	valid:   <form target="_blank" action="http://x.com"></form>     (forms is off by default)
//	invalid: <a target="_blank" href="http://x.com"></a>
//	invalid: <a target="_blank" href="//x.com"></a>
//	invalid: <a target="_blank" href={dynamic}></a>
//	invalid: <a target="_blank" href="http://x.com" rel="noopener"></a>
//
// Ported from `react/jsx-no-target-blank` in `eslint-plugin-react`, which is the implementation
// that defined this rule. All 113 imported corpus cases were run against the installed build,
// version 7.37.5, through the ESLint Linter API before any code was written, and it agreed with the
// corpus on every one of them, counts and message ids alike. Everything below that the corpus does
// not state was measured the same way, on inputs written for the question.
//
// # The link and form components, which our config cannot name
//
// `linkComponentsUtil` starts from `['a']` and `['form']` and concatenates
// `settings.linkComponents` / `settings.formComponents`, each of which may be a bare name or a
// `{name, linkAttribute}` pair naming a different attribute to read. cohere reads neither
// (the loader refuses both namespaces by name), so neither list can be
// extended here by any route.
//
// The faithful reading is the unconfigured answer rather than a guess at what a project might
// configure. Five corpus cases carry `linkComponents` settings and all five were re-run with the
// settings removed: every one goes CLEAN, because `<Link>` is not a link component by default.
// Those five are recorded as clean fixtures with the reasoning at the line.
//
// So the components here are `a` and `form`, and their attributes are `href` and `action`.
//
// # `links: false` does nothing, and that is upstream's defect reproduced
//
// The schema declares `links` with `default: true`, the rule's `Object.assign` sets it, and
// `configuration.links` is then read NOWHERE in the rule body. Grepped: the only `configuration.`
// reads in the body are `configuration.forms`. Measured on the installed build, three ways:
//
//	{links: false}   <a target="_blank" href="http://x.com"></a>   REPORTS
//	{links: true}    same input                                    reports
//	no options       same input                                    reports
//
// Reproduced rather than fixed. A port that wired it up would silence a whole class of finding on
// any project that wrote `links: false` expecting it to be inert, and no imported fixture could see
// the difference because the corpus writes `links: false` only alongside `forms: true` and asserts
// the FORM finding. That pairing is exactly the shape that makes a wrong belief invisible.
//
// # Which attribute wins when one is written twice
//
// Every lookup is `findLastIndex`, so the LAST `target`, the LAST `href` and the LAST `rel` decide.
// This is not a detail: `<a target="_self" target="_blank" href="http://x.com">` reports and
// `<a target="_blank" href="http://x.com" href="/rel">` is clean, both measured, and a port using a
// first-match search would answer the opposite on both.
//
// # What counts as an external link
//
// `hasExternalLink` accepts a literal href matching `/^(?:\w+:|\/\/)/`, so a scheme or a
// protocol-relative prefix. That makes `mailto:a@b.c` external, measured, which reads as
// over-broad and is upstream's judgment. `enforceDynamicLinks` adds `href={anything}` on top, and
// its absent value is `always`.
//
// # What counts as a secure rel, and the conditional-pairing rule inside it
//
// `hasSecureRel` reads the last `rel` and reduces its value to a list of strings, then requires
// EVERY one of them to contain `noreferrer` (or `noopener` under `allowReferrer`). Three value
// shapes reach it and the third is where the real work is:
//
//	rel="noopener noreferrer"       a literal, split on spaces, lowercased
//	rel={`noreferrer`}              a template literal, and ONLY its first quasi is read, so
//	                                `rel={`${x} noreferrer`}` REPORTS. Measured.
//	rel={c ? 'noreferrer' : 'x'}    a conditional, and which branch is read depends on the TARGET
//
// That last one is the subtle part and it is measured in four directions. When the `rel` is a
// conditional AND the `target` is a conditional AND both tests are the same plain identifier,
// upstream pairs the branches: it finds which of the target's two branches is `_blank` and reads
// the rel branch at the SAME index. So:
//
//	target={c ? '_blank' : '_self'} rel={c ? 'noreferrer' : 'x'}   clean, branches pair up
//	target={c ? '_self' : '_blank'} rel={c ? 'x' : 'noreferrer'}   clean, same pairing inverted
//	target={c ? '_blank' : '_self'} rel={c ? 'x' : 'noreferrer'}   REPORTS, wrong branch is secure
//	target={c ? '_blank' : '_self'} rel={d ? 'noreferrer' : 'x'}   REPORTS, tests differ so BOTH
//	                                                               branches must be secure
//
// When the tests differ, or the target is not a conditional, both rel branches are returned and the
// `every` requires both to be secure, which is the fourth row. All four measured.
//
// # The fixer, and the one arm that has none
//
// `meta.fixable` is `"code"` and the fixer has four arms plus two declines, and the corpus asserts
// 28 rewritten outputs. Every arm below was confirmed by reading the fix ranges the installed build
// emits rather than by reading the source alone:
//
//	no rel at all       insert ` rel="noreferrer"` after the LAST attribute
//	rel with no value   insert `="noreferrer"` after the rel attribute
//	rel="..."           replace the whole string literal with the old parts plus noreferrer
//	rel={"..."}         replace the inner literal the same way
//	rel={5}, rel={null} replace the WHOLE container with `"noreferrer"`, not the inner expression
//
// And two declines, both returning null upstream so no fix is offered:
//
//	target written before a spread          `targetIndex < spreadAttributeIndex`
//	a spread present and no rel attribute   the insertion point cannot be trusted
//
// **The form arm ships no fixer at all.** Its report call has no `fix` property, where the link
// arm's does. Measured: `{forms: true}` on `<form target="_blank" action="http://x.com">` reports
// and carries no fix. Reproduced, because a fix invented here would be applied unattended on a
// shape upstream deliberately leaves alone.
//
// The `noreferrer` written by the fixer is `noreferrer` even under `allowReferrer`, EXCEPT in the
// insertion arms, which write `relValue` and therefore write `noopener`. That asymmetry is
// upstream's: `relValue` is computed from the flag and then only the two `insertText` calls use it,
// while all three `replaceText` calls hardcode `noreferrer`. Measured both ways.
var JsxNoTargetBlank = rule.Rule{
	// No namespace prefix. The config writes `react/jsx-no-target-blank` and the parity guard
	// strips the namespace on a `/` boundary, so a rule named `react-jsx-no-target-blank` would
	// match no inventory entry and lint no files while passing every fixture in this package.
	Name: "react/jsx-no-target-blank",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule reaches this with nil, and the comma-ok form yields the zero struct,
		// whose `Links` is false rather than upstream's true. That difference decides nothing today
		// because nothing reads `Links`, but relying on that would make a future wiring of `Links`
		// silently wrong for every unconfigured file, so the default is restored explicitly.
		settings, configured := rule.OptionsAs[JsxNoTargetBlankOptions](options)
		if !configured {
			settings = JsxNoTargetBlankOptions{Links: true}
		}

		check := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if tagName == nil || attributes == nil {
				return
			}
			// A member-expression or namespaced tag is a component reference rather than the
			// intrinsic element, and upstream compares `node.name.name`, which is undefined for
			// both. `jsx.IsIntrinsicElementNamed` asks the same question.
			properties := attributesOf(attributes)
			if properties == nil {
				return
			}

			targetIndex := lastAttributeIndexNamed(properties, "target")
			spreadIndex := lastSpreadAttributeIndex(properties)

			if jsx.IsIntrinsicElementNamed(tagName, "a") {
				checkTargetBlankLink(ctx, node, properties, targetIndex, spreadIndex, settings)
			}
			if jsx.IsIntrinsicElementNamed(tagName, "form") {
				checkTargetBlankForm(ctx, node, properties, targetIndex, spreadIndex, settings)
			}
		}

		return rule.Listeners{
			// Both kinds. A self-closing `<a target="_blank" href="..." />` produces no
			// JsxOpeningElement in our tree, and upstream's single `JSXOpeningElement` listener
			// sees it because its parser gives a self-closing element an opening element with a
			// flag. Listening on only the first kind would silence the rule on the form most of
			// these are written in.
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// checkTargetBlankLink is the `<a>` arm, which reports with a fix.
func checkTargetBlankLink(
	ctx rule.Context,
	element *ast.Node,
	properties []*ast.Node,
	targetIndex int,
	spreadIndex int,
	settings JsxNoTargetBlankOptions,
) {
	if !reachesTargetBlankCheck(properties, targetIndex, spreadIndex, settings.WarnOnSpreadAttributes) {
		return
	}

	dangerous := hasExternalTargetBlankLink(properties, "href", spreadIndex, settings.WarnOnSpreadAttributes) ||
		(!settings.EnforceDynamicLinksNever && hasDynamicTargetBlankLink(properties, "href"))
	if !dangerous {
		return
	}
	if hasSecureTargetBlankRel(properties, settings.AllowReferrer, settings.WarnOnSpreadAttributes, spreadIndex) {
		return
	}

	message := messageNoTargetBlankWithoutNoreferrer
	insertedRelValue := "noreferrer"
	if settings.AllowReferrer {
		message = messageNoTargetBlankWithoutNoopener
		insertedRelValue = "noopener"
	}

	fix, fixable := targetBlankFix(ctx, properties, targetIndex, spreadIndex, insertedRelValue)
	if !fixable {
		ctx.ReportNode(element, message)
		return
	}
	ctx.ReportNodeWithFixes(element, message, fix)
}

// checkTargetBlankForm is the `<form>` arm, which reports WITHOUT a fix.
//
// The guard order here is upstream's and it is not the same as the link arm's. `hasSecureRel` is
// called with ONE argument, so `allowReferrer`, `warnOnSpreadAttributes` and the spread index all
// arrive as undefined inside it: `allowReferrer` is falsy, so `rel="noopener"` never satisfies a
// form even under `{allowReferrer: true, forms: true}`, and the spread check is skipped. Measured:
// that exact configuration on `<form target="_blank" action="http://x.com" rel="noopener">` reports.
//
// The message id still switches on `allowReferrer`, because that is computed outside the call. So a
// form can report `noTargetBlankWithoutNoopener` while `noopener` would not have satisfied it,
// which is upstream's inconsistency and is reproduced.
func checkTargetBlankForm(
	ctx rule.Context,
	element *ast.Node,
	properties []*ast.Node,
	targetIndex int,
	spreadIndex int,
	settings JsxNoTargetBlankOptions,
) {
	if !reachesTargetBlankCheck(properties, targetIndex, spreadIndex, settings.WarnOnSpreadAttributes) {
		return
	}
	// The `forms` gate sits AFTER the target check upstream rather than before it, which changes
	// nothing observable and is kept in place so the two arms read the same way.
	if !settings.Forms {
		return
	}
	// One argument, deliberately: see this function's doc comment.
	if hasSecureTargetBlankRel(properties, false, false, -1) {
		return
	}

	// The form arm passes no spread index and no spread flag either, so a spread contributes
	// nothing here even under `warnOnSpreadAttributes`.
	dangerous := hasExternalTargetBlankLink(properties, "action", -1, false) ||
		(!settings.EnforceDynamicLinksNever && hasDynamicTargetBlankLink(properties, "action"))
	if !dangerous {
		return
	}

	message := messageNoTargetBlankWithoutNoreferrer
	if settings.AllowReferrer {
		message = messageNoTargetBlankWithoutNoopener
	}
	// No fix. Upstream's form report call has no `fix` property where the link arm's does, and
	// inventing one here would rewrite a shape upstream declines to touch.
	ctx.ReportNode(element, message)
}

// reachesTargetBlankCheck reproduces the guard both arms open with, verbatim in structure.
//
// Upstream writes it as a negated block whose arms are easy to collapse wrongly:
//
//	if (!attributeValuePossiblyBlank(node.attributes[targetIndex])) {
//	  const hasSpread = spreadAttributeIndex >= 0;
//	  if (warnOnSpreadAttributes && hasSpread) {
//	    // fall through
//	  } else if ((hasSpread && targetIndex < spreadAttributeIndex) || !hasSpread || !warnOnSpreadAttributes) {
//	    return;
//	  }
//	}
//
// The second condition is a tautology given the first arm did not fire: if `warnOnSpreadAttributes`
// is false the third disjunct is true, and if it is true then `hasSpread` is false so the second
// disjunct is true. So the whole block reduces to "return unless the target is possibly blank, or
// there is a spread we were told to warn about". That reduction is stated here rather than carried
// as dead disjuncts, and a fixture covers each of the four combinations so the reduction is checked
// rather than asserted.
func reachesTargetBlankCheck(
	properties []*ast.Node,
	targetIndex int,
	spreadIndex int,
	warnOnSpreadAttributes bool,
) bool {
	if targetIndex >= 0 && attributeValuePossiblyBlank(properties[targetIndex]) {
		return true
	}
	return warnOnSpreadAttributes && spreadIndex >= 0
}

// attributeValuePossiblyBlank reports whether an attribute could hold the string `_blank`.
//
// Three shapes answer, and the comparison is case-INSENSITIVE in every one of them because
// upstream lowercases before comparing. Measured: `target="_BLANK"` reports.
//
//	target="_blank"                  a string literal
//	target={'_blank'}                a literal inside a container
//	target={c ? '_blank' : '_self'}  a conditional, EITHER branch
//
// The conditional arm tests the alternate first and then the consequent, and either being `_blank`
// is enough. An identifier, a template literal, or anything else answers false, which is why
// `target={variable}` never reaches the check on its own.
func attributeValuePossiblyBlank(attribute *ast.Node) bool {
	if attribute == nil || attribute.Kind != ast.KindJsxAttribute {
		return false
	}
	initializer := attribute.AsJsxAttribute().Initializer
	if initializer == nil {
		return false
	}
	if initializer.Kind == ast.KindStringLiteral {
		return strings.EqualFold(initializer.Text(), targetBlankValue)
	}
	if initializer.Kind != ast.KindJsxExpression {
		return false
	}
	expression := initializer.AsJsxExpression().Expression
	if expression == nil {
		return false
	}
	if expression.Kind == ast.KindStringLiteral {
		return strings.EqualFold(expression.Text(), targetBlankValue)
	}
	if expression.Kind != ast.KindConditionalExpression {
		return false
	}
	conditional := expression.AsConditionalExpression()
	// Upstream's two checks require a truthy value before lowercasing, which an empty string fails.
	// `isStringLiteralValued` carries that requirement.
	return isStringLiteralValued(conditional.WhenFalse, targetBlankValue) ||
		isStringLiteralValued(conditional.WhenTrue, targetBlankValue)
}

// targetBlankValue is the one target value this rule is about.
const targetBlankValue = "_blank"

// isStringLiteralValued reports whether a node is a non-empty string literal equal to want, ignoring case.
//
// The non-empty requirement is upstream's `expr.alternate.value &&` guard, which is truthiness
// rather than a length test and therefore excludes the empty string specifically.
func isStringLiteralValued(node *ast.Node, want string) bool {
	if node == nil || node.Kind != ast.KindStringLiteral {
		return false
	}
	text := node.Text()
	// The non-empty test is EQUIVALENT while `want` is `_blank`, and is kept as documentation of
	// upstream's guard rather than as a discrimination this rule makes. Measured: a mutant dropping
	// it survives the whole fixture set, and its inverse (forcing the guard false) is caught by
	// eight lines, so the arm is reached and the fixtures can see it. No input can distinguish the
	// two versions, because an empty string is already unequal to a non-empty `want`.
	//
	// It becomes load-bearing the moment a caller passes an empty `want`, which is why it stays.
	return text != "" && strings.EqualFold(text, want)
}

// hasExternalTargetBlankLink reports whether the last link attribute names somewhere off this page.
//
// The literal test is `/^(?:\w+:|\/\/)/`: a word-character run followed by a colon, or two leading
// slashes. That makes `mailto:` and `tel:` external as well as `http:`, measured, and it is
// upstream's judgment rather than a simplification.
//
// The spread clause is the second half of upstream's return and it is separate from the literal
// test: under `warnOnSpreadAttributes`, a link attribute written BEFORE a spread counts as
// dangerous whatever its value, because the spread may overwrite it. `linkIndex` is -1 when there
// is no link attribute at all, and -1 is less than any real spread index, so an element with a
// spread and NO href is dangerous under that flag. Measured: `<a {...props} href="http://x">` with
// the flag reports, and so does `<a {...props}>` with a target.
func hasExternalTargetBlankLink(
	properties []*ast.Node,
	linkAttribute string,
	spreadIndex int,
	warnOnSpreadAttributes bool,
) bool {
	linkIndex := lastAttributeIndexNamed(properties, linkAttribute)
	if linkIndex >= 0 {
		initializer := properties[linkIndex].AsJsxAttribute().Initializer
		if initializer != nil && initializer.Kind == ast.KindStringLiteral &&
			isExternalLinkText(initializer.Text()) {
			return true
		}
	}
	return warnOnSpreadAttributes && linkIndex < spreadIndex
}

// isExternalLinkText reproduces `/^(?:\w+:|\/\/)/` without compiling a pattern.
//
// `\w` is `[A-Za-z0-9_]` in JavaScript, and the run must be non-empty before the colon, so a bare
// `:path` does not match. Hand-rolled because the shape is fixed and this runs per attribute.
func isExternalLinkText(text string) bool {
	if strings.HasPrefix(text, "//") {
		return true
	}
	for index := 0; index < len(text); index++ {
		character := text[index]
		if character == ':' {
			// The colon must be preceded by at least one word character.
			return index > 0
		}
		isWord := character == '_' ||
			(character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9')
		if !isWord {
			return false
		}
	}
	return false
}

// hasDynamicTargetBlankLink reports whether any link attribute holds an expression container.
//
// Upstream searches for the LAST such attribute and returns true if one exists, so the index it
// finds is discarded and only existence matters. Reproduced as existence, and the equivalence is
// exact rather than a simplification: `findLastIndex(...) !== -1` and "any matches" are the same
// predicate over the same list.
func hasDynamicTargetBlankLink(properties []*ast.Node, linkAttribute string) bool {
	for _, property := range properties {
		name, named := jsx.AttributeName(property)
		if !named || name != linkAttribute {
			continue
		}
		initializer := property.AsJsxAttribute().Initializer
		if initializer != nil && initializer.Kind == ast.KindJsxExpression {
			return true
		}
	}
	return false
}

// hasSecureTargetBlankRel reports whether the last `rel` severs the opener handle.
//
// The spread clause comes first and matches upstream: under `warnOnSpreadAttributes`, a `rel`
// written before a spread is not trusted, because the spread may replace it. Measured:
// `<a rel="noreferrer" {...props} target="_blank" href="http://x">` reports under that flag and is
// clean without it.
//
// After that, the last `rel`'s value is reduced to a list of strings and EVERY one must be secure.
// The list is one element for a literal or a template, and two for an unpaired conditional; the
// pairing rule that can shrink it back to one is in `relStringsFor`.
func hasSecureTargetBlankRel(
	properties []*ast.Node,
	allowReferrer bool,
	warnOnSpreadAttributes bool,
	spreadIndex int,
) bool {
	relIndex := lastAttributeIndexNamed(properties, "rel")
	if relIndex == -1 || (warnOnSpreadAttributes && relIndex < spreadIndex) {
		return false
	}
	targetIndex := lastAttributeIndexNamed(properties, "target")

	var targetValue *ast.Node
	if targetIndex >= 0 {
		targetValue = properties[targetIndex].AsJsxAttribute().Initializer
	}
	values, resolved := relStringsFor(properties[relIndex].AsJsxAttribute().Initializer, targetValue)
	if !resolved {
		// `[].concat(null)` is `[null]` upstream, and `typeof null === 'string'` is false, so the
		// single element fails and `every` returns false. An unresolvable rel is therefore
		// insecure rather than absent, which is why this returns false rather than falling through.
		return false
	}
	if len(values) == 0 {
		// `[].concat([])` is `[]` and `[].every(...)` is TRUE. Upstream cannot reach this today,
		// because every resolved shape yields at least one string, but the vacuous-true is what
		// the code says and inverting it here would be a silent divergence if a shape ever does.
		return true
	}
	for _, value := range values {
		if !isSecureRelValue(value, allowReferrer) {
			return false
		}
	}
	return true
}

// isSecureRelValue tests one space-separated rel token list.
//
// `noreferrer` always satisfies. `noopener` satisfies only under `allowReferrer`, which is the
// whole of what that flag changes about the judgment. The comparison is lowercased, so
// `rel="NOREFERRER"` is secure, measured.
func isSecureRelValue(value string, allowReferrer bool) bool {
	for _, tag := range strings.Split(strings.ToLower(value), " ") {
		if tag == "noreferrer" {
			return true
		}
	}
	if !allowReferrer {
		return false
	}
	for _, tag := range strings.Split(strings.ToLower(value), " ") {
		if tag == "noopener" {
			return true
		}
	}
	return false
}

// relStringsFor reduces a rel attribute's value to the strings that must each be secure.
//
// The second return separates "resolved to no strings" from "could not resolve", which decide
// opposite ways at the caller.
//
// Four shapes resolve and the fourth carries the pairing rule:
//
//	rel="a b"                 the literal text, one string
//	rel={'a b'}               the inner literal, one string
//	rel={`a b`}               the FIRST QUASI only, so `rel={`${x} noreferrer`}` yields "" and
//	                          reports. Measured; upstream reads `quasis[0].value.cooked`.
//	rel={c ? 'a' : 'b'}       both branches, UNLESS the target is a conditional on the same test
//
// The pairing: when the target is also a conditional expression and both tests are plain
// identifiers with the same name, upstream finds which target branch is `_blank` and returns ONLY
// the rel branch at that index. `[c,a].indexOf('_blank')` is -1 when neither branch is `_blank`,
// and JavaScript's `arr[-1]` is undefined, so the list becomes `[undefined]`, which fails the
// string test and reports. That -1 case is reachable only when the target reached here through the
// spread path rather than through a `_blank` branch, and it is reproduced by returning
// unresolvable.
func relStringsFor(relValue *ast.Node, targetValue *ast.Node) ([]string, bool) {
	if relValue == nil {
		return nil, false
	}
	if relValue.Kind == ast.KindStringLiteral {
		return []string{relValue.Text()}, true
	}
	if relValue.Kind != ast.KindJsxExpression {
		return nil, false
	}
	expression := relValue.AsJsxExpression().Expression
	if expression == nil {
		return nil, false
	}

	if expression.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return []string{expression.Text()}, true
	}
	if expression.Kind == ast.KindTemplateExpression {
		// Only the first quasi, which is upstream's `quasis[0].value.cooked`. A template opening
		// with a substitution has an empty first quasi, and empty is not secure.
		head := expression.AsTemplateExpression().Head
		if head == nil {
			return nil, false
		}
		return []string{head.Text()}, true
	}

	if expression.Kind == ast.KindConditionalExpression {
		conditional := expression.AsConditionalExpression()
		relBranches := []*ast.Node{conditional.WhenTrue, conditional.WhenFalse}

		if paired, index, ok := targetBlankBranchIndex(targetValue, conditional.Condition); ok {
			if !paired {
				// `indexOf` answered -1, so upstream indexes past the end and gets undefined.
				return nil, false
			}
			return stringTextsOf(relBranches[index : index+1])
		}
		return stringTextsOf(relBranches)
	}

	if expression.Kind == ast.KindStringLiteral {
		return []string{expression.Text()}, true
	}
	// `expr.value` on anything else is undefined, which fails the string test at the caller.
	return nil, false
}

// targetBlankBranchIndex answers which rel branch to read when the target pairs with it.
//
// The three returns are three different answers and collapsing any two of them changes a verdict:
//
//	ok=false             the target is not a same-test conditional, so BOTH rel branches apply
//	ok=true, paired=false the tests match but neither target branch is `_blank`, so upstream
//	                      indexes with -1 and gets undefined, which is insecure
//	ok=true, paired=true  read the rel branch at this index and only that one
//
// The test comparison is on plain identifiers only, because upstream reads `.test.name`, which is
// undefined for anything else. Two different expressions that both lack a `name` therefore compare
// equal as `undefined === undefined`, which would pair branches upstream never meant to pair; that
// is guarded here by requiring both to be identifiers, and the guard is a real divergence in the
// safe direction. Recorded rather than hidden: it is unreachable through the corpus and through
// every probe written for it, because a non-identifier target test means the target is a
// conditional whose branches are still tested for `_blank` by string, and the pairing only matters
// once one of them is.
func targetBlankBranchIndex(targetValue *ast.Node, relTest *ast.Node) (paired bool, index int, ok bool) {
	if targetValue == nil || targetValue.Kind != ast.KindJsxExpression {
		return false, 0, false
	}
	targetExpression := targetValue.AsJsxExpression().Expression
	if targetExpression == nil || targetExpression.Kind != ast.KindConditionalExpression {
		return false, 0, false
	}
	targetConditional := targetExpression.AsConditionalExpression()

	targetTest, relTestNode := targetConditional.Condition, relTest
	if targetTest == nil || relTestNode == nil {
		return false, 0, false
	}
	if targetTest.Kind != ast.KindIdentifier || relTestNode.Kind != ast.KindIdentifier {
		return false, 0, false
	}
	if targetTest.Text() != relTestNode.Text() {
		return false, 0, false
	}

	// `[consequent.value, alternate.value].indexOf('_blank')`. The comparison here is EXACT rather
	// than case-insensitive, because `indexOf` is, unlike the two `toLowerCase` comparisons
	// elsewhere in this rule. Upstream is inconsistent about it and this reproduces the
	// inconsistency at the line where it lives.
	for position, branch := range []*ast.Node{targetConditional.WhenTrue, targetConditional.WhenFalse} {
		if branch != nil && branch.Kind == ast.KindStringLiteral && branch.Text() == targetBlankValue {
			return true, position, true
		}
	}
	return false, 0, true
}

// stringTextsOf reads a run of nodes as string literals, refusing if any is not one.
//
// Upstream builds `[consequent.value, alternate.value]`, where a non-literal contributes undefined
// and the `every` then fails on it. Refusing the whole list produces the same verdict and keeps the
// undefined out of the Go types.
func stringTextsOf(nodes []*ast.Node) ([]string, bool) {
	texts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node == nil || node.Kind != ast.KindStringLiteral {
			return nil, false
		}
		texts = append(texts, node.Text())
	}
	return texts, true
}

// targetBlankFix builds the repair, or reports that upstream declines to offer one.
//
// The two declines are upstream's two `return null` paths and each is a real case:
//
//	targetIndex < spreadAttributeIndex     the target came before a spread, so the spread may be
//	                                       supplying the real target and the edit could be wrong
//	a spread exists and there is no rel     there is no safe place to put the insertion
//
// Note the first fires even when there is NO spread, because `targetIndex` is -1 then and
// `spreadIndex` is also -1, and -1 < -1 is false, so it does not. It fires when the target is
// absent and a spread is present, which is exactly the warnOnSpreadAttributes case.
func targetBlankFix(
	ctx rule.Context,
	properties []*ast.Node,
	targetIndex int,
	spreadIndex int,
	insertedRelValue string,
) (rule.Fix, bool) {
	relIndex := firstAttributeIndexNamed(properties, "rel")
	if targetIndex < spreadIndex || (spreadIndex >= 0 && relIndex == -1) {
		return rule.Fix{}, false
	}

	if relIndex == -1 {
		if len(properties) == 0 {
			// Upstream's `attributes.slice(-1)[0]` is undefined for an empty list and
			// `insertTextAfter(undefined, ...)` throws, so this shape cannot reach the fixer
			// upstream: an element with no attributes has no target and no spread, so it never
			// reports. Guarded rather than relied upon, because a panic here would take the whole
			// run down where upstream merely never arrives.
			return rule.Fix{}, false
		}
		last := properties[len(properties)-1]
		return rule.ReplaceRange(zeroWidthRangeAt(last.End()), " rel=\""+insertedRelValue+"\""), true
	}

	// The fixer finds the FIRST rel, not the last. `hasSecureRel` uses `findLastIndex` and the
	// fixer uses `find`, so an element writing `rel` twice is judged on the last and repaired on
	// the first. Measured: `<a href="http://x" target="_blank" rel="noreferrer" rel="x">` reports
	// and its fix rewrites the FIRST rel. Reproduced rather than unified.
	rel := properties[relIndex].AsJsxAttribute()
	initializer := rel.Initializer

	if initializer == nil {
		return rule.ReplaceRange(
			zeroWidthRangeAt(properties[relIndex].End()),
			"=\""+insertedRelValue+"\"",
		), true
	}

	// Every replace arm writes `noreferrer` even under `allowReferrer`, where the two insert arms
	// above write `relValue`. Upstream's asymmetry, measured both ways.
	if initializer.Kind == ast.KindStringLiteral {
		return rule.ReplaceRange(
			rule.TokenRange(ctx.SourceFile, initializer),
			relLiteralWithNoreferrer(initializer.Text()),
		), true
	}

	if initializer.Kind != ast.KindJsxExpression {
		return rule.Fix{}, false
	}
	expression := initializer.AsJsxExpression().Expression
	if expression == nil {
		return rule.Fix{}, false
	}
	if expression.Kind == ast.KindStringLiteral {
		return rule.ReplaceRange(
			rule.TokenRange(ctx.SourceFile, expression),
			relLiteralWithNoreferrer(expression.Text()),
		), true
	}
	// A numeric, boolean, null, or bigint literal replaces the WHOLE container rather than the
	// expression inside it, so `rel={5}` becomes `rel="noreferrer"` and not `rel={"noreferrer"}`.
	// Measured on `rel={5}` and `rel={null}`; the fix range covers the braces in both.
	if isNonStringLiteralExpression(expression) {
		return rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, initializer), "\"noreferrer\""), true
	}
	// An identifier, a template, a conditional: upstream falls off the end and returns null.
	return rule.Fix{}, false
}

// isNonStringLiteralExpression matches upstream's "for undefined, boolean, number, symbol, bigint,
// and null" branch.
//
// Upstream reaches it by testing `expression.type === 'Literal'` and then `typeof value !== 'string'`,
// so it is a LITERAL that is not a string. `undefined` is an identifier in JavaScript rather than a
// literal and therefore does NOT reach that branch upstream, which is why it is absent here despite
// being named in upstream's own comment. Measured: `rel={undefined}` reports and carries no fix.
func isNonStringLiteralExpression(expression *ast.Node) bool {
	switch expression.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

// relLiteralWithNoreferrer builds the replacement text for a rel that already has a value.
//
// Upstream splits the old value on `noreferrer`, drops empty pieces, appends `noreferrer`, and
// joins with a space. So `rel="noopener"` becomes `"noopener noreferrer"`, and a value that already
// contained `noreferrer` somewhere is normalized rather than doubled. The quotes are part of the
// written text because the replaced range is the literal including them.
func relLiteralWithNoreferrer(existing string) string {
	parts := make([]string, 0, 4)
	for _, part := range strings.Split(existing, "noreferrer") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	parts = append(parts, "noreferrer")
	return "\"" + strings.Join(parts, " ") + "\""
}

// zeroWidthRangeAt builds an insertion point, which our fix engine spells as an empty range.
//
// ESLint has `insertTextAfter`; we have one `Fix` shape holding a range and a text, so an insertion
// is a replacement of nothing. Named rather than written inline three times so the intent survives.
func zeroWidthRangeAt(position int) core.TextRange {
	return core.NewTextRange(position, position)
}

// attributesOf returns a JSX element's attribute list, or nil.
func attributesOf(attributes *ast.Node) []*ast.Node {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return nil
	}
	return properties.Nodes
}

// lastAttributeIndexNamed is upstream's `findLastIndex`, which is not `indexOf`.
//
// The LAST attribute of a name decides every question this rule asks except which one the fixer
// repairs. `<a target="_self" target="_blank" href="http://x">` reports and
// `<a target="_blank" href="http://x" href="/rel">` is clean, both measured, and a first-match
// search answers the opposite on both.
//
// The name comparison is EXACT. `jsx.AttributeName` declines a spread and a namespaced name, which
// matches upstream reading `attr.name.name`: a spread has no `name` and a namespaced name is a
// different node shape whose `.name` is an object rather than a string.
func lastAttributeIndexNamed(properties []*ast.Node, wanted string) int {
	for index := len(properties) - 1; index >= 0; index-- {
		name, named := jsx.AttributeName(properties[index])
		if named && name == wanted {
			return index
		}
	}
	return -1
}

// firstAttributeIndexNamed is upstream's `Array.prototype.find`, used only by the fixer.
//
// Deliberately not the same search as the judgment above; see `targetBlankFix`.
func firstAttributeIndexNamed(properties []*ast.Node, wanted string) int {
	for index, property := range properties {
		name, named := jsx.AttributeName(property)
		if named && name == wanted {
			return index
		}
	}
	return -1
}

// lastSpreadAttributeIndex finds the last `{...spread}` in the attribute list.
func lastSpreadAttributeIndex(properties []*ast.Node) int {
	for index := len(properties) - 1; index >= 0; index-- {
		if properties[index].Kind == ast.KindJsxSpreadAttribute {
			return index
		}
	}
	return -1
}
