package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noUnknownPropertyFile is where the fixtures pretend to live.
const noUnknownPropertyFile = "/repository/source/NoUnknownProperty.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-unknown-property.js, read by evaluating
// the two arrays in the tester with a stubbed RuleTester and emitting Go literals from the decoded
// values, so no escape sequence was typed on the way here. The sources were then compared byte for
// byte against the decoded corpus with a script. Upstream carries 92 valid and 33 invalid cases.
//
// All 125 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and 121 agreed. The four disagreements are TABLE CONTENT
// rather than logic: the clone knows three property names the installed build does not and widens
// one tag list.
//
// The tables in this port come from the CLONE, so those four cases are clean here and agree with
// upstream's own expectation. Both artifacts declare 7.37.5, meaning the clone is unreleased
// commits on the same tag rather than a release we lag, and diffing the two extractions shows the
// clone is a strict superset with no removals. The four rows carry the reasoning at their site.
//
// Every other expected message id is the installed build's own, per case, so a case reporting
// twice carries two ids and their ORDER is asserted. The two artifacts agree on all 121.

// noUnknownPropertyOptions decodes a raw option body the way the config layer does.
//
// Fixtures route through the exported decoder rather than building the struct, which is what puts
// the defaults under test. An empty body is the nil-options case.
func noUnknownPropertyOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoUnknownPropertyOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestNoUnknownPropertyFires runs every case this port reports on.
func TestNoUnknownPropertyFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"invalid 0", `<div allowTransparency="true" />`, ``, []string{`unknownProp`}},
		{"invalid 1", `<div hasOwnProperty="should not be allowed property"></div>;`, ``, []string{`unknownProp`}},
		{"invalid 2", `<div abc="should not be allowed property"></div>;`, ``, []string{`unknownProp`}},
		{"invalid 3", `<div aria-fake="should not be allowed property"></div>;`, ``, []string{`unknownProp`}},
		{"invalid 4", `<div someProp="bar"></div>;`, ``, []string{`unknownProp`}},
		{"invalid 5", `<div class="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 6", `<div for="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 7", `<div accept-charset="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 8", `<div http-equiv="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 9", `<div accesskey="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 10", `<div onclick="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 11", `<div onmousedown="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 12", `<div onMousedown="bar"></div>;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 13", `<use xlink:href="bar" />;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 14", `<rect clip-path="bar" transform-origin="center" />;`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 15", `<script crossorigin nomodule />`, ``, []string{`unknownPropWithStandardName`, `unknownPropWithStandardName`}},
		{"invalid 16", `<div crossorigin />`, ``, []string{`unknownPropWithStandardName`}},
		{"invalid 17", `<div crossOrigin />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 18", `<div as="audio" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 19", `<div onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onResize={this.resize} onError={this.error} />`, ``, []string{`invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`}},
		{"invalid 20", `<div onLoad={this.load} />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 21", `<div fill="pink" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 22", `<div controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} allowFullScreen></div>`, ``, []string{`invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`, `invalidPropOnTag`}},
		{"invalid 23", `<div download="foo" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 24", `<div imageSrcSet="someImageSrcSet" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 25", `<div imageSizes="someImageSizes" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 26", `<div data-xml-anything="invalid" />`, ``, []string{`unknownProp`}},
		{"invalid 27", `<div data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;`, `{"requireDataLowercase":true}`, []string{`dataLowercaseRequired`, `dataLowercaseRequired`, `unknownProp`}},
		{"invalid 28", `<App data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;`, `{"requireDataLowercase":true}`, []string{`dataLowercaseRequired`, `dataLowercaseRequired`}},
		{"invalid 29", `<div abbr="abbr" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 30", `<div webkitDirectory="" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 31", `<div webkitdirectory="" />`, ``, []string{`invalidPropOnTag`}},
		{"invalid 32", `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash:c="customValue">
          Hello, world!
        </div>
      `, ``, []string{`unknownProp`}},

		// The only two corpus cases carrying `settings.react.version`, and each is clean UPSTREAM
		// only because of that setting. cohere has no shared-settings surface, so this port answers
		// as though there were none, and both report. Recorded here at this port's verdict with the
		// upstream one named, rather than deleted: the gap is a missing configuration surface, not
		// a defect in the judgment, and deleting the cases would hide it.
		//
		// Upstream sets 16.0.99 to ACCEPT `allowTransparency`, a name React dropped in 16.1.
		{"valid 46", `<div allowTransparency="true" />`, ``, []string{`unknownProp`}},
		// Upstream sets 19.0.0 to accept `precedence`, which the no-settings default does not reach.
		{"valid 50", `<link precedence="medium" href="https://foo.bar" rel="canonical" />`, ``, []string{`unknownProp`}},
		// Cases upstream's corpus does not write, each measured against the installed build.

		// A namespaced attribute REPORTS, which is why this rule reads source text rather than
		// calling jsx.AttributeName: that shelf helper declines a namespaced name by design, and
		// using it would silently drop every one of these.
		{"a namespaced xlink attribute", `<svg><use xlink:href="#a" /></svg>`, ``, []string{"unknownPropWithStandardName"}},
		{"a namespaced xmlns attribute", `<svg xmlns:xlink="u" />`, ``, []string{"unknownProp"}},
		{"a bogus namespaced attribute", `<div foo:bar="x" />`, ``, []string{"unknownProp"}},
		// The xml exclusion inside the data-attribute test.
		{"a data-xml prefixed attribute", `<div data-xmlFoo="x" />`, ``, []string{"unknownProp"}},
		// An ARIA name is matched EXACTLY, so a miscased one is not exempt.
		{"an unknown aria attribute", `<div aria-bogus="true" />`, ``, []string{"unknownProp"}},
		// The tag-restricted table reports when the tag is wrong.
		{"a restricted attribute on the wrong tag", `<div charset="utf-8" />`, ``, []string{"invalidPropOnTag"}},
		// The ignore list is matched against the spelling as WRITTEN, before case normalisation.
		// Naming the canonical form does not exempt the written one, which is the half a single
		// row would miss and which a mutant normalising first survives without these two.
		// `charset` is already the canonical spelling in the ignore-case list, so normalising it
		// changes nothing and that row cannot separate the two orders. `allowfullscreen` can:
		// normalisation rewrites it to `allowFullScreen`, so a port matching the ignore list AFTER
		// normalisation would exempt on the canonical spelling and not on the written one, which is
		// the exact inverse of upstream. All three rows measured against the installed build.
		{"the canonical spelling does not exempt allowfullscreen", `<div allowfullscreen="true" />`, `{"ignore":["allowFullScreen"]}`, []string{"invalidPropOnTag"}},
		{"an unignored allowfullscreen on the wrong tag reports", `<div allowfullscreen="true" />`, ``, []string{"invalidPropOnTag"}},
		{"the canonical spelling does not exempt charset", `<div charset="utf-8" />`, `{"ignore":["charSet"]}`, []string{"invalidPropOnTag"}},
		// The ARIA list is matched EXACTLY, unlike the property search below it. A mutant lowering
		// the name before the lookup survives every corpus case, because upstream writes no
		// miscased ARIA attribute anywhere.
		// A plain COMPONENT reaches the data-attribute branch, because that branch runs before the
		// html-like guard. This is the control for the member-tag row in the silent table: the two
		// differ only in the dot, and without upstream's early member-tag return the member form
		// would report here too. Both measured on the installed build.
		{"a component reaches the data branch", `<Foo data-Baz="x" />`, `{"requireDataLowercase":true}`, []string{"dataLowercaseRequired"}},
		{"a miscased aria attribute", `<div aria-Hidden="true" />`, ``, []string{"unknownProp"}},
		{"an uppercase aria attribute", `<div ARIA-HIDDEN="true" />`, ``, []string{"unknownProp"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := noUnknownPropertyOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoUnknownPropertyStaysSilent runs every case this port declines.
func TestNoUnknownPropertyStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"valid 0", `<App class="bar" />;`, ``},
		{"valid 1", `<App for="bar" />;`, ``},
		{"valid 2", `<App someProp="bar" />;`, ``},
		{"valid 3", `<Foo.bar for="bar" />;`, ``},
		{"valid 4", `<App accept-charset="bar" />;`, ``},
		{"valid 5", `<App http-equiv="bar" />;`, ``},
		{"valid 6", `<App xlink:href="bar" />;`, ``},
		{"valid 7", `<App clip-path="bar" />;`, ``},
		{"valid 8", `<App dataNotAnDataAttribute="yes" />;`, `{"requireDataLowercase":true}`},
		{"valid 9", `<div className="bar"></div>;`, ``},
		{"valid 10", `<div onMouseDown={this._onMouseDown}></div>;`, ``},
		{"valid 13", `<a href="someLink" download="foo">Read more</a>`, ``},
		{"valid 14", `<area download="foo" />`, ``},
		{"valid 15", `<img src="cat_keyboard.jpeg" alt="A cat sleeping on a keyboard" align="top" fetchPriority="high" />`, ``},
		{"valid 16", `<input type="password" required />`, ``},
		{"valid 17", `<input ref={this.input} type="radio" />`, ``},
		{"valid 18", `<input type="file" webkitdirectory="" />`, ``},
		{"valid 19", `<input type="file" webkitDirectory="" />`, ``},
		{"valid 20", `<div inert children="anything" />`, ``},
		{"valid 21", `<iframe scrolling="?" onLoad={a} onError={b} align="top" />`, ``},
		{"valid 22", `<input key="bar" type="radio" />`, ``},
		{"valid 23", `<button disabled>You cannot click me</button>;`, ``},
		{"valid 24", `<svg key="lock" viewBox="box" fill={10} d="d" stroke={1} strokeWidth={2} strokeLinecap={3} strokeLinejoin={4} transform="something" clipRule="else" x1={5} x2="6" y1="7" y2="8"></svg>`, ``},
		{"valid 25", `<g fill="#7B82A0" fillRule="evenodd"></g>`, ``},
		{"valid 26", `<mask fill="#7B82A0"></mask>`, ``},
		{"valid 27", `<symbol fill="#7B82A0"></symbol>`, ``},
		{"valid 28", `<meta property="og:type" content="website" />`, ``},
		{"valid 29", `<input type="checkbox" checked={checked} disabled={disabled} id={id} onChange={onChange} />`, ``},
		{"valid 30", `<video playsInline />`, ``},
		{"valid 31", `<img onError={foo} onLoad={bar} />`, ``},
		{"valid 32", `<picture inert={false} onError={foo} onLoad={bar} />`, ``},
		{"valid 33", `<iframe onError={foo} onLoad={bar} />`, ``},
		{"valid 34", `<script onLoad={bar} onError={foo} />`, ``},
		{"valid 35", `<source onLoad={bar} onError={foo} />`, ``},
		{"valid 36", `<link onLoad={bar} onError={foo} />`, ``},
		{"valid 37", `<link rel="preload" as="image" href="someHref" imageSrcSet="someImageSrcSet" imageSizes="someImageSizes" />`, ``},
		{"valid 38", `<object onLoad={bar} />`, ``},
		{"valid 40", `<video allowFullScreen webkitAllowFullScreen mozAllowFullScreen />`, ``},
		{"valid 41", `<iframe allowFullScreen webkitAllowFullScreen mozAllowFullScreen />`, ``},
		{"valid 42", `<table border="1" />`, ``},
		{"valid 43", `<th abbr="abbr" />`, ``},
		{"valid 44", `<td abbr="abbr" />`, ``},
		{"valid 45", `<template shadowrootmode="open" shadowrootclonable shadowrootdelegatesfocus shadowrootserializable />`, ``},
		{"valid 47", `<div onPointerDown={this.onDown} onPointerUp={this.onUp} />`, ``},
		{"valid 48", `<input type="checkbox" defaultChecked={this.state.checkbox} />`, ``},
		{"valid 49", `<div onTouchStart={this.startAnimation} onTouchEnd={this.stopAnimation} onTouchCancel={this.cancel} onTouchMove={this.move} onMouseMoveCapture={this.capture} onTouchCancelCapture={this.log} />`, ``},
		{"valid 51", `<meta charset="utf-8" />;`, ``},
		{"valid 52", `<meta charSet="utf-8" />;`, ``},
		{"valid 53", `<div class="foo" is="my-elem"></div>;`, ``},
		{"valid 54", `<div {...this.props} class="foo" is="my-elem"></div>;`, ``},
		{"valid 55", `<atom-panel class="foo"></atom-panel>;`, ``},
		{"valid 56", `<div data-foo="bar"></div>;`, ``},
		{"valid 57", `<div data-foo-bar="baz"></div>;`, ``},
		{"valid 58", `<div data-parent="parent"></div>;`, ``},
		{"valid 59", `<div data-index-number="1234"></div>;`, ``},
		{"valid 60", `<div data-e2e-id="5678"></div>;`, ``},
		{"valid 61", `<div data-testID="bar" data-under_sCoRe="bar" />;`, ``},
		{"valid 62", `<div data-testID="bar" data-under_sCoRe="bar" />;`, `{"requireDataLowercase":false}`},
		{"valid 63", `<div class="bar"></div>;`, `{"ignore":["class"]}`},
		{"valid 64", `<div someProp="bar"></div>;`, `{"ignore":["someProp"]}`},
		{"valid 65", `<div css={{flex: 1}}></div>;`, `{"ignore":["css"]}`},
		{"valid 66", `<button aria-haspopup="true">Click me to open pop up</button>;`, ``},
		{"valid 67", `<button aria-label="Close" onClick={someThing.close} />;`, ``},
		{"valid 68", `<script crossOrigin noModule />`, ``},
		{"valid 69", `<audio crossOrigin />`, ``},
		{"valid 70", `<svg focusable><image crossOrigin /></svg>`, ``},
		{"valid 71", `<details onToggle={this.onToggle}>Some details</details>`, ``},
		{"valid 72", `<path fill="pink" d="M 10,30 A 20,20 0,0,1 50,30 A 20,20 0,0,1 90,30 Q 90,60 50,90 Q 10,60 10,30 z"></path>`, ``},
		{"valid 73", `<line fill="pink" x1="0" y1="80" x2="100" y2="20"></line>`, ``},
		{"valid 74", `<link as="audio">Audio content</link>`, ``},
		{"valid 75", `<video controlsList="nodownload" controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} onResize={this.onResize}></video>`, ``},
		{"valid 76", `<audio controlsList="nodownload" controls={this.controls} crossOrigin="anonymous" disableRemotePlayback loop muted preload="none" src="something" onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onError={this.error} onResize={this.onResize}></audio>`, ``},
		{"valid 77", `<marker id={markerId} viewBox="0 0 2 2" refX="1" refY="1" markerWidth="1" markerHeight="1" orient="auto" />`, ``},
		{"valid 78", `<pattern id="pattern" viewBox="0,0,10,10" width="10%" height="10%" />`, ``},
		{"valid 79", `<symbol id="myDot" width="10" height="10" viewBox="0 0 2 2" />`, ``},
		{"valid 80", `<view id="one" viewBox="0 0 100 100" />`, ``},
		{"valid 81", `<hr align="top" />`, ``},
		{"valid 82", `<applet align="top" />`, ``},
		{"valid 83", `<marker fill="#000" />`, ``},
		{"valid 85", `
        <table align="top">
          <caption align="top">Table Caption</caption>
          <colgroup valign="top" align="top">
            <col valign="top" align="top"/>
          </colgroup>
          <thead valign="top" align="top">
            <tr valign="top" align="top">
              <th valign="top" align="top">Header</th>
              <td valign="top" align="top">Cell</td>
            </tr>
          </thead>
          <tbody valign="top" align="top" />
          <tfoot valign="top" align="top" />
        </table>
      `, ``},
		{"valid 86", `<fbt desc="foo" doNotExtract />;`, ``},
		{"valid 87", `<fbs desc="foo" doNotExtract />;`, ``},
		{"valid 88", `<math displaystyle="true" />;`, ``},
		{"valid 89", `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash="customValue">
          Hello, world!
        </div>
      `, ``},
		{"valid 90", `
        <div>
          <button popovertarget="my-popover" popovertargetaction="toggle">Open Popover</button>

          <div popover id="my-popover">Greetings, one and all!</div>
        </div>
      `, ``},
		{"valid 91", `
        <div>
          <button popoverTarget="my-popover" popoverTargetAction="toggle">Open Popover</button>

          <div id="my-popover" onBeforeToggle={this.onBeforeToggle} popover>Greetings, one and all!</div>
        </div>
      `, ``},

		// The four rows where the clone is ahead of the installed build. Both artifacts declare
		// 7.37.5; the clone is unreleased commits on the same tag, and its changelog carries all
		// four under "Unreleased" as `allow` or `add` entries. This port takes the clone's tables,
		// so these agree with upstream's own expectation and are clean.
		//
		// The choice matters only for code nobody has written yet, and it is asymmetric. Following
		// the installed build would report `onScrollEnd` as unknown on a tree running React 19.2.8,
		// which is a false positive on correct code; following the clone stays silent on four names.
		// Measured with a working control, none of the four appears anywhere in this tree today.
		{"valid 11", `<div onScrollEnd={this._onScrollEnd}></div>;`, ``},
		{"valid 12", `<div onScrollEndCapture={this._onScrollEndCapture}></div>;`, ``},
		{"valid 39", `<body onLoad={bar} />`, ``},
		{"valid 84", `<dialog closedby="something" onClose={handler} open id="dialog" returnValue="something" onCancel={handler2} />`, ``},

		// Cases upstream's corpus does not write, each measured against the installed build.

		// A hyphenated name is a plain identifier in our parser, NOT a namespaced node, and both
		// of these branches depend on reading it correctly.
		{"a well formed data attribute", `<div data-foo="x" />`, ``},
		{"an uppercase data attribute is accepted by default", `<div data-Foo="x" />`, ``},
		{"a known aria attribute", `<div aria-hidden="true" />`, ``},
		// A component, a member tag and a custom element are all outside the rule's scope.
		{"an unknown attribute on a component", `<Foo bogus="x" />`, ``},
		{"an unknown attribute on a member tag", `<Foo.Bar bogus="x" />`, ``},
		// The member-tag return happens BEFORE the data-attribute branch, so this stays clean while
		// the same attribute on a plain component reports. A mutant deleting that early return
		// survives every other fixture, because the html-like guard further down declines the same
		// inputs for every branch AFTER it. This row is the only one that can see it.
		{"a member tag is skipped before the data branch", `<Foo.Bar data-Baz="x" />`, `{"requireDataLowercase":true}`},
		{"an unknown attribute on a custom element", `<my-elem bogus="x" />`, ``},
		{"the is attribute exempts a customised built-in", `<div is="my-elem" bogus="x" />`, ``},
		{"an fbt tag is skipped", `<fbt bogus="x" />`, ``},
		{"a spread carries no name", `<div {...properties} />`, ``},
		// The tag-restricted table accepts when the tag is right.
		{"a restricted attribute on an allowed tag", `<meta charset="utf-8" />`, ``},
		// The canonical spelling passes where the DOM one reports.
		{"the canonical svg spelling", `<path strokeWidth="1" />`, ``},
		// The control for the ignore-list pair above: the written spelling exempts.
		{"the written spelling exempts charset", `<div charset="utf-8" />`, `{"ignore":["charset"]}`},
		// The control for the pair above: the WRITTEN spelling exempts, and this is the row the
		// mutant has to break.
		{"the written spelling exempts allowfullscreen", `<div allowfullscreen="true" />`, `{"ignore":["allowfullscreen"]}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := noUnknownPropertyOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnknownPropertyRepairs asserts what the fixer WRITES, not merely that it fires.
//
// Only the standard-name branch carries a repair, and it replaces the NAME node alone so the value
// survives. `ExpectFixedSource` replays the fix and compares the whole rewritten file, so a repair
// with the right text over the wrong span fails here even though its message id is correct.
func TestNoUnknownPropertyRepairs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{"class becomes className", `<div class="x" />`, `<div className="x" />`},
		{"for becomes htmlFor", `<label for="x" />`, `<label htmlFor="x" />`},
		{"stroke-width becomes strokeWidth", `<path stroke-width="1" />`, `<path strokeWidth="1" />`},
		{"http-equiv becomes httpEquiv", `<meta http-equiv="refresh" />`, `<meta httpEquiv="refresh" />`},
		// The value is untouched, including one holding an expression that must survive intact.
		{"the value survives the repair", `<div class={someExpression} />`, `<div className={someExpression} />`},
		// A namespaced name is replaced whole, colon included.
		{"a namespaced name is replaced whole", `<svg><use xlink:href="#a" /></svg>`, `<svg><use xlinkHref="#a" /></svg>`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, noUnknownPropertyOptions(t, ``))
			rule_testing.ExpectFindings(t, result, "unknownPropWithStandardName")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestNoUnknownPropertyDoesNotRepairWhatItCannotName pins that three of the four messages carry no
// repair.
//
// Only `unknownPropWithStandardName` knows what to write. A fixer on the other three would be
// guessing, and a fix is applied unattended. Asserting the ABSENCE of a repair is something no
// message-id fixture can see.
func TestNoUnknownPropertyDoesNotRepairWhatItCannotName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"an unknown name has no standard spelling", `<div bogus="x" />`, ``},
		{"a restricted attribute on the wrong tag", `<div charset="utf-8" />`, ``},
		{"an uppercase data attribute under the option", `<div data-Foo="x" />`, `{"requireDataLowercase":true}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, noUnknownPropertyOptions(t, testCase.rawOptions))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
				t.Fatalf("a repair was offered for a name the rule cannot spell: %v", fixes)
			}
		})
	}
}

// TestNoUnknownPropertyMessageText asserts every rendered message exactly.
//
// Three of the four messages interpolate, so an id assertion cannot see what the format string
// computed, and the tag-restricted one joins a whole list into its text. Asserted against literals
// typed here rather than against the rule's own constants, which would move under mutation.
func TestNoUnknownPropertyMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		want       string
	}{
		{
			"unknownProp names the attribute",
			`<div bogus="x" />`, ``,
			"Unknown property 'bogus' found",
		},
		{
			"unknownPropWithStandardName names both spellings",
			`<div class="x" />`, ``,
			"Unknown property 'class' found, use 'className' instead",
		},
		{
			"invalidPropOnTag lists every allowed tag",
			`<div charset="utf-8" />`, ``,
			"Invalid property 'charset' found on tag 'div', but it is only allowed on: meta",
		},
		{
			"dataLowercaseRequired names the lowercase form",
			`<div data-Foo="x" />`, `{"requireDataLowercase":true}`,
			"React does not recognize data-* props with uppercase characters on a DOM element. " +
				"Found 'data-Foo', use 'data-foo' instead",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, noUnknownPropertyOptions(t, testCase.rawOptions))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Fatalf("message is %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestNoUnknownPropertyAnchor pins WHERE each finding points.
//
// Upstream reports on the whole JSXAttribute for every message, so the span covers the name and the
// value together rather than either alone. The repair, by contrast, replaces only the NAME node,
// and that asymmetry is exactly the shape the brief warns about: a fix that lands correctly while
// the finding points elsewhere shows the reader a line they were never told about.
func TestNoUnknownPropertyAnchor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"the standard-name branch", `<div class="x" />`, `class="x"`},
		{"the unknown branch", `<div bogus="x" />`, `bogus="x"`},
		{"the restricted-tag branch", `<div charset="utf-8" />`, `charset="utf-8"`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, testCase.sourceText, noUnknownPropertyOptions(t, ``))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			fixture := strings.TrimSpace(testCase.sourceText) + "\n"
			span := result.Diagnostics[0].Range
			if got := fixture[span.Pos():span.End()]; got != testCase.want {
				t.Fatalf("finding spans %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestNoUnknownPropertyIgnoreOption pins that the ignore list matches the ACTUAL spelling.
//
// Upstream tests the ignore list before normalising case, so an entry has to be written the way the
// attribute is written. Naming the canonical spelling does NOT exempt the DOM one, which is the
// half a single-row fixture would miss.
func TestNoUnknownPropertyIgnoreOption(t *testing.T) {
	t.Parallel()

	t.Run("the written spelling exempts", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div class="x" />`, noUnknownPropertyOptions(t, `{"ignore":["class"]}`))
		rule_testing.ExpectClean(t, result)
	})
	t.Run("the canonical spelling does not exempt the written one", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div class="x" />`, noUnknownPropertyOptions(t, `{"ignore":["className"]}`))
		rule_testing.ExpectFindings(t, result, "unknownPropWithStandardName")
	})
	t.Run("an unrelated entry exempts nothing", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div class="x" />`, noUnknownPropertyOptions(t, `{"ignore":["other"]}`))
		rule_testing.ExpectFindings(t, result, "unknownPropWithStandardName")
	})
}

// TestNoUnknownPropertyRequireDataLowercase pins both settings of the option.
//
// The default is false, and the default matters: with it off an uppercase data attribute is
// accepted silently, which is the opposite verdict from every other unknown name.
func TestNoUnknownPropertyRequireDataLowercase(t *testing.T) {
	t.Parallel()

	t.Run("off by default", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div data-Foo="x" />`, noUnknownPropertyOptions(t, ``))
		rule_testing.ExpectClean(t, result)
	})
	t.Run("explicitly false", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div data-Foo="x" />`, noUnknownPropertyOptions(t, `{"requireDataLowercase":false}`))
		rule_testing.ExpectClean(t, result)
	})
	t.Run("on reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div data-Foo="x" />`, noUnknownPropertyOptions(t, `{"requireDataLowercase":true}`))
		rule_testing.ExpectFindings(t, result, "dataLowercaseRequired")
	})
	t.Run("on leaves a lowercase data attribute alone", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUnknownProperty, noUnknownPropertyFile, `<div data-foo="x" />`, noUnknownPropertyOptions(t, `{"requireDataLowercase":true}`))
		rule_testing.ExpectClean(t, result)
	})
}

// TestNoUnknownPropertyHasNoFileSuffixGate pins that every extension carrying JSX is examined.
//
// Three siblings in this package once gated on `.tsx` and `.jsx`, which was oxc residue and has
// since been removed from all three. Upstream gates nothing.
//
// `.ts` is absent for a PARSER reason rather than a rule one: TypeScript reads an opening angle
// bracket there as a type assertion, so no JSX attribute node exists for any rule to see.
func TestNoUnknownPropertyHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const source = `<div class="x" />`

	for _, fileName := range []string{
		"/repository/source/NoUnknownProperty.tsx",
		"/repository/source/NoUnknownProperty.jsx",
		"/repository/source/NoUnknownProperty.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnknownProperty, fileName, source, noUnknownPropertyOptions(t, ``))
			rule_testing.ExpectFindings(t, result, "unknownPropWithStandardName")
		})
	}
}

// TestDecodeNoUnknownPropertyOptions covers both options and the nil-options path.
func TestDecodeNoUnknownPropertyOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty body carries upstream's defaults", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnknownPropertyOptions(nil)
		if err != nil {
			t.Fatalf("decoding an empty body: %v", err)
		}
		options, ok := decoded.(NoUnknownPropertyOptions)
		if !ok {
			t.Fatalf("decoded to %T, want NoUnknownPropertyOptions", decoded)
		}
		if len(options.Ignore) != 0 {
			t.Fatalf("ignore list is %v, want empty", options.Ignore)
		}
		if options.RequireDataLowercase {
			t.Fatal("requireDataLowercase defaulted to true, want false")
		}
	})

	t.Run("both keys decode", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnknownPropertyOptions([]byte(`{"ignore":["a","b"],"requireDataLowercase":true}`))
		if err != nil {
			t.Fatalf("decoding both keys: %v", err)
		}
		options := decoded.(NoUnknownPropertyOptions)
		if strings.Join(options.Ignore, ",") != "a,b" {
			t.Fatalf("ignore list is %v", options.Ignore)
		}
		if !options.RequireDataLowercase {
			t.Fatal("requireDataLowercase is false, want true")
		}
	})
}

// TestNoUnknownPropertyTables guards the extracted data against silent truncation.
//
// The tables ARE the rule, so a generator that emitted half of one would leave most fixtures
// passing while the rule quietly stopped knowing names. These counts were taken from the installed
// module's own constants at extraction time.
func TestNoUnknownPropertyTables(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  int
		want int
	}{
		{"dom attribute names", len(noUnknownPropertyDomAttributeNames), 6},
		{"svg dom attribute names", len(noUnknownPropertySvgDomAttributeNames), 82},
		{"attribute tags map", len(noUnknownPropertyAttributeTags), 65},
		{"properties ignoring case", len(noUnknownPropertyPropertiesIgnoreCase), 7},
		{"aria properties", len(noUnknownPropertyAriaProperties), 53},
		// 364 two-word plus 193 one-word plus 21 pointer handlers. `precedence` is deliberately
		// absent: it is gated on React >= 19 and the no-settings default does not reach it, which
		// was established by bisecting the running rule rather than by reading its source. The
		// two-word list carries two more entries than the installed build's, `onScrollEnd` and
		// `onScrollEndCapture`, because these tables come from the clone.
		{"property names", len(noUnknownPropertyNames), 578},
	}
	for _, testCase := range cases {
		if testCase.got != testCase.want {
			t.Errorf("%s holds %d entries, want %d", testCase.name, testCase.got, testCase.want)
		}
	}

	// Spot-check the entries whose CASING is the rule, in both directions.
	if noUnknownPropertySvgDomAttributeNames["stroke-width"] != "strokeWidth" {
		t.Error("the svg map lost its stroke-width entry")
	}
	if noUnknownPropertyDomAttributeNames["class"] != "className" {
		t.Error("the dom map lost its class entry")
	}
	if _, present := noUnknownPropertySvgDomAttributeNames["strokeWidth"]; present {
		t.Error("the svg map keys on the canonical spelling, which would invert the rule")
	}
}
