package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoScriptUrlFile is where the fixtures pretend to live.
//
// A .tsx extension because every case holds JSX. The rule has no suffix gate, which a case below
// pins by writing reporting source to a plain `.ts` file.
const jsxNoScriptUrlFile = "/repository/source/JsxNoScriptUrl.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-no-script-url.js, read by evaluating
// the two arrays in the tester with a stubbed RuleTester and emitting Go literals from the decoded
// values, so no escape sequence was typed on the way here and nothing could cook one. Upstream
// carries 12 valid and 10 invalid cases, and the invalid ones name 11 findings between them.
//
// All 22 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and the corpus and the running rule agreed on every one.
//
// # Some cases depend on a configuration surface cohere does not have
//
// Upstream reads `settings.linkComponents` from ESLint's shared settings when a case configures
// `includeFromSettings: true`. cohere has no shared-settings surface, so those four cases' options
// are refused at configuration, recorded with upstream's verdicts in
// TestJsxNoScriptUrlRefusesIncludeFromSettings. The cases with settings but no flag are kept at
// the verdict this port produces, measured by replaying them against the installed build with the
// settings stripped, which is exactly the configuration this port can express.

// jsxNoScriptUrlOptions decodes a raw option body the way the config layer does.
//
// Fixtures route through the exported decoder rather than building the struct, which is what puts
// the positional union and the built-in default under test. An empty body is the nil-options case:
// upstream still checks `a` and `href`, so the answer is the built-in pair rather than an error.
func jsxNoScriptUrlOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeJsxNoScriptUrlOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestJsxNoScriptUrlFires runs every case this port reports on.
func TestJsxNoScriptUrlFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantCount  int
	}{
		{"invalid 0", `<a href="javascript:"></a>`, ``, 1},
		{"invalid 1", `<a href="javascript:void(0)"></a>`, ``, 1},
		// The interleaved-whitespace case, carrying real newline, carriage return and tab
		// characters rather than typed escapes.
		{"invalid 2", "<a href=\"j\n\n\na\rv\tascript:\"></a>", ``, 1},
		{"invalid 3", `<Foo to="javascript:"></Foo>`, `[[{"name":"Foo","props":["to","href"]}]]`, 1},
		{"invalid 4", `<Foo href="javascript:"></Foo>`, `[[{"name":"Foo","props":["to","href"]}]]`, 1},
		{"invalid 5", `<a href="javascript:void(0)"></a>`, `[[{"name":"Foo","props":["to","href"]}]]`, 1},
		// Upstream's invalid 8 is this source with `includeFromSettings: true` added and reports
		// twice; its configuration is refused here, see TestJsxNoScriptUrlRefusesIncludeFromSettings.
		// Without the flag upstream answers with one finding, and so does this port.
		{"invalid 9", `
      <div>
        <Foo href="javascript:"></Foo>
        <Bar link="javascript:"></Bar>
      </div>
    `, `[[{"name":"Bar","props":["link"]}]]`, 1},

		// Cases upstream's corpus does not write, each measured against the installed build before
		// being written here.

		// The prefix class spans every code unit below 0x20 plus the space, which is wider than
		// the class allowed between the letters. All four report.
		{"prefix 0x01", "<a href=\"\x01javascript:x\"></a>", ``, 1},
		{"prefix 0x09 tab", "<a href=\"\tjavascript:x\"></a>", ``, 1},
		{"prefix 0x1f", "<a href=\"\x1fjavascript:x\"></a>", ``, 1},
		{"prefix space", `<a href=" javascript:x"></a>`, ``, 1},
		// A newline BETWEEN letters is in the interleaving class and reports. Its partner in the
		// silent table is the same shape with a space, which is not.
		{"newline inside the word", "<a href=\"ja\nvascript:x\"></a>", ``, 1},
		// The separator class applies after the final letter too, before the colon.
		{"tab before the colon", "<a href=\"javascript\t:x\"></a>", ``, 1},
		// Case-insensitive, upstream's flag.
		{"mixed case", `<a href="JaVaScRiPt:x"></a>`, ``, 1},
		{"upper case", `<a href="JAVASCRIPT:void(0)"></a>`, ``, 1},
		// A self-closing element is a different node kind here and upstream's one selector covers
		// both, so this pins that the attribute listener reaches it.
		{"self-closing anchor", `<a href="javascript:" />`, ``, 1},
		// A legacy entry naming a component twice keeps the LAST entry's prop list. Measured both
		// ways; the partner case in the silent table uses the first entry's prop and is clean.
		{"repeated name keeps the last props", `<Foo to="javascript:"></Foo>`, `[[{"name":"Foo","props":["href"]},{"name":"Foo","props":["to"]}]]`, 1},
		// An entry named `a` REPLACES the built-in href pair rather than adding to it, so the
		// replacement's own prop reports. Its partner below shows `href` going clean.
		{"an entry named a replaces the default", `<a to="javascript:"></a>`, `[[{"name":"a","props":["to"]}]]`, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := jsxNoScriptUrlOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxNoScriptUrl, jsxNoScriptUrlFile, testCase.sourceText, options)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "noScriptURL"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestJsxNoScriptUrlStaysSilent runs every case this port declines.
func TestJsxNoScriptUrlStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"valid 0", `<a href="https://reactjs.org"></a>`, ``},
		{"valid 1", `<a href="mailto:foo@bar.com"></a>`, ``},
		{"valid 2", `<a href="#"></a>`, ``},
		{"valid 3", `<a href=""></a>`, ``},
		{"valid 4", `<a name="foo"></a>`, ``},
		// An expression container declines even though its content is the very string that would
		// report as a plain attribute value. Upstream tests the value node's type, not its content.
		{"valid 5", `<a href={"javascript:"}></a>`, ``},
		{"valid 6", `<Foo href="javascript:"></Foo>`, ``},
		{"valid 7", `<a href />`, ``},
		{"valid 8", `<Foo other="javascript:"></Foo>`, `[[{"name":"Foo","props":["to","href"]}]]`},
		// Clean upstream WITH its settings and clean here without them, so the missing surface
		// costs nothing on these two. Measured both ways. Upstream's valid 11 carries
		// `includeFromSettings: true` and is in TestJsxNoScriptUrlRefusesIncludeFromSettings.
		{"valid 9", `<Foo href="javascript:"></Foo>`, ``},
		{"valid 10", `<Foo href="javascript:"></Foo>`, `[[],{"includeFromSettings":false}]`},

		// Cases upstream's corpus does not write, each measured against the installed build.

		// A space between the letters is NOT in the interleaving class, unlike a newline. This is
		// the distinction a port reusing the wider prefix class would get wrong, and its partner
		// in the firing table is the same shape with a newline.
		{"space inside the word", `<a href="ja vascript:x"></a>`, ``},
		// 0x7f sits above the prefix class, whose upper bound is 0x1f.
		{"prefix 0x7f", "<a href=\"\x7fjavascript:x\"></a>", ``},
		{"no colon at all", `<a href="javascript"></a>`, ``},
		// A namespaced attribute carries no plain name, so upstream never reads its value.
		{"namespaced attribute", `<a xlink:href="javascript:"></a>`, ``},
		// A spread has no name to match.
		{"spread only", `<a {...properties} />`, ``},
		// Neither a member nor a namespaced tag carries a plain name, so both decline.
		{"member tag name", `<a.b href="javascript:"></a.b>`, ``},
		{"namespaced tag name", `<svg:a href="javascript:"></svg:a>`, ``},
		// A capitalised tag is a different name from the built-in `a`, matched exactly.
		{"uppercase tag", `<A href="javascript:"></A>`, ``},
		// `form` and `action` are the FORM component defaults, which this rule never consults.
		{"a form action is not a link", `<form action="javascript:"></form>`, ``},
		// A template literal is not a Literal upstream, so it declines like the expression case.
		{"template literal value", "<a href={`javascript:`}></a>", ``},
		// The partner to the replacement case above: once `a` is redefined, its built-in `href`
		// pair is gone and the same source goes clean.
		{"an entry named a drops the built-in href", `<a href="javascript:"></a>`, `[[{"name":"a","props":["to"]}]]`},
		// The partner to the repeated-name case: the FIRST entry's prop no longer applies.
		{"repeated name drops the first props", `<Foo href="javascript:"></Foo>`, `[[{"name":"Foo","props":["href"]},{"name":"Foo","props":["to"]}]]`},
		// An empty prop list matches nothing.
		{"empty props list", `<Foo href="javascript:"></Foo>`, `[[{"name":"Foo","props":[]}]]`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := jsxNoScriptUrlOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxNoScriptUrl, jsxNoScriptUrlFile, testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoScriptUrlReportsOnTheWholeAttribute pins WHERE the finding points.
//
// Upstream reports on the JSXAttribute, so the span covers the name and the value together rather
// than either alone. Measured on the installed build: the attribute in this source reports from
// column 4 through the closing quote. A port anchoring on the value node, or on the element, would
// pass every message-id fixture above while pointing somewhere the reader was never shown.
func TestJsxNoScriptUrlReportsOnTheWholeAttribute(t *testing.T) {
	t.Parallel()

	const source = `<a href="javascript:void(0)"></a>`

	result := rule_testing.RunWithOptions(t, JsxNoScriptUrl, jsxNoScriptUrlFile, source, jsxNoScriptUrlOptions(t, ``))
	rule_testing.ExpectFindings(t, result, "noScriptURL")

	// The harness does not trim for an untyped run, so the literal above and the file on disk are
	// the same bytes and this slice is safe.
	span := result.Diagnostics[0].Range
	reported := source[span.Pos():span.End()]
	if reported != `href="javascript:void(0)"` {
		t.Fatalf("finding spans %q, want the whole attribute", reported)
	}
}

// TestJsxNoScriptUrlMessageText asserts the rendered finding exactly.
//
// The message carries no format verbs, so there is nothing to interpolate and the assertion is
// equality against a literal typed here rather than against the rule's own constant, which would
// move with any mutation of it.
func TestJsxNoScriptUrlMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, JsxNoScriptUrl, jsxNoScriptUrlFile, `<a href="javascript:"></a>`, jsxNoScriptUrlOptions(t, ``))
	rule_testing.ExpectFindings(t, result, "noScriptURL")

	const want = "A future version of React will block javascript: URLs as a security precaution. " +
		"Use event handlers instead if you can. If you need to generate unsafe HTML, try using " +
		"dangerouslySetInnerHTML instead."
	if got := jsxNoScriptUrlMessage.Description; got != want {
		t.Fatalf("message description is %q, want %q", got, want)
	}
	if got := jsxNoScriptUrlMessage.Id; got != "noScriptURL" {
		t.Fatalf("message id is %q, want %q", got, "noScriptURL")
	}
}

// TestJsxNoScriptUrlHasNoFileSuffixGate pins that the rule reads every extension that can carry JSX.
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue. Upstream has no
// such gate anywhere, and a gate here would blind the rule to most of a tree.
//
// `.ts` is deliberately absent, and it is absent for a PARSER reason rather than a rule one, which
// is a distinction worth stating because the two look identical from a fixture. TypeScript reads
// `<a href="...">` in a `.ts` file as a type assertion, not as JSX, so no attribute node is
// produced for any rule to see. Probed with a control across all four extensions: `.tsx`, `.jsx`
// and `.js` each report one finding on this source and `.ts` reports none, with the same rule and
// the same options. A sibling makes the ungated point using `React.createElement` source, which
// needs no JSX; this rule anchors on a JSX attribute and has no such arm, so the honest coverage is
// the three extensions where the node can exist.
func TestJsxNoScriptUrlHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const source = `<a href="javascript:"></a>`

	for _, fileName := range []string{
		"/repository/source/JsxNoScriptUrl.tsx",
		"/repository/source/JsxNoScriptUrl.js",
		"/repository/source/JsxNoScriptUrl.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, JsxNoScriptUrl, fileName, source, jsxNoScriptUrlOptions(t, ``))
			rule_testing.ExpectFindings(t, result, "noScriptURL")
		})
	}
}

// TestDecodeJsxNoScriptUrlOptions covers the positional union and the built-in default.
//
// The two slots have different shapes and Go has no single type for that, so the decoder tries the
// array arm first, matching upstream's own test on whether the first option is an array. An empty
// body must yield the built-in pair rather than an error, which is the nil-options path a rule
// configured as a bare severity takes.
func TestDecodeJsxNoScriptUrlOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty body keeps the built-in pair", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeJsxNoScriptUrlOptions(nil)
		if err != nil {
			t.Fatalf("decoding an empty body: %v", err)
		}
		options, ok := decoded.(JsxNoScriptUrlOptions)
		if !ok {
			t.Fatalf("decoded to %T, want JsxNoScriptUrlOptions", decoded)
		}
		if len(options.Legacy) != 0 {
			t.Fatalf("legacy list is %v, want empty", options.Legacy)
		}
		components := jsxNoScriptUrlLinkComponents(options)
		if got := strings.Join(components["a"], ","); got != "href" {
			t.Fatalf("built-in pair is %q, want %q", got, "href")
		}
	})

	t.Run("the array arm alone", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeJsxNoScriptUrlOptions([]byte(`[[{"name":"Foo","props":["to","href"]}]]`))
		if err != nil {
			t.Fatalf("decoding the array arm: %v", err)
		}
		options := decoded.(JsxNoScriptUrlOptions)
		if len(options.Legacy) != 1 || options.Legacy[0].Name != "Foo" {
			t.Fatalf("legacy list is %v, want one entry named Foo", options.Legacy)
		}
		if got := strings.Join(options.Legacy[0].Props, ","); got != "to,href" {
			t.Fatalf("props are %q, want %q", got, "to,href")
		}
	})

	t.Run("the object arm alone, at its default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeJsxNoScriptUrlOptions([]byte(`[{"includeFromSettings":false}]`))
		if err != nil {
			t.Fatalf("decoding the object arm: %v", err)
		}
		if options := decoded.(JsxNoScriptUrlOptions); len(options.Legacy) != 0 {
			t.Fatalf("legacy list is %v, want empty", options.Legacy)
		}
	})

	t.Run("both arms in order", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeJsxNoScriptUrlOptions([]byte(`[[{"name":"Bar","props":["link"]}],{"includeFromSettings":false}]`))
		if err != nil {
			t.Fatalf("decoding both arms: %v", err)
		}
		options := decoded.(JsxNoScriptUrlOptions)
		if len(options.Legacy) != 1 || options.Legacy[0].Name != "Bar" {
			t.Fatalf("legacy list is %v, want one entry named Bar", options.Legacy)
		}
	})

	t.Run("an empty array slot is not the object arm", func(t *testing.T) {
		t.Parallel()
		// `[[], {...}]` is upstream's own spelling in two corpus cases. An empty JSON array is the
		// array arm, so the object must still be read from the second position: the refusal of
		// `true` there is what proves it was read at all.
		if _, err := DecodeJsxNoScriptUrlOptions([]byte(`[[],{"includeFromSettings":true}]`)); err == nil {
			t.Fatal("the object in the second position was not read, so its `true` was not refused")
		}
		if _, err := DecodeJsxNoScriptUrlOptions([]byte(`[[],{"includeFromSettings":false}]`)); err != nil {
			t.Fatalf("the same list at the default was refused: %v", err)
		}
	})

	t.Run("shapes the config layer does not deliver, or upstream refuses, are refused", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{
			// The wrapper this decoder used to read.
			`{"positional":[[{"name":"Foo","props":["to"]}]]}`,
			// An object first takes no second element.
			`[{"includeFromSettings":false},[{"name":"Foo","props":["to"]}]]`,
			// A third element.
			`[[],{"includeFromSettings":false},{}]`,
			// Keys outside the schema, in a legacy entry and in the object.
			`[[{"name":"Foo","prop":["to"]}]]`,
			`[{"includeFromSetting":false}]`,
		} {
			if decoded, err := DecodeJsxNoScriptUrlOptions([]byte(raw)); err == nil {
				t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
			}
		}
	})
}

// TestJsxNoScriptUrlRefusesIncludeFromSettings records upstream's four corpus cases that carry
// `includeFromSettings: true`, with upstream's verdict beside each.
//
// Upstream reads `settings.linkComponents` when the flag is true, and cohere has no shared-settings
// surface, so the flag cannot be honoured. These used to be accepted and run as though the settings
// were empty, which made invalid 6 and 7 silent and invalid 8 report once where upstream reports
// twice, with nothing telling the author their flag had no effect. They are refused at configuration
// now, so the gap is a startup error naming the flag rather than a quiet difference in findings.
func TestJsxNoScriptUrlRefusesIncludeFromSettings(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`[[],{"includeFromSettings":true}]`,                                     // valid 11, clean upstream
		`[[{"name":"Bar","props":["to","href"]}],{"includeFromSettings":true}]`, // invalid 6, one finding upstream
		`[{"includeFromSettings":true}]`,                                        // invalid 7, one finding upstream
		`[[{"name":"Bar","props":["link"]}],{"includeFromSettings":true}]`,      // invalid 8, two findings upstream
	} {
		_, err := DecodeJsxNoScriptUrlOptions([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), "includeFromSettings") {
			t.Errorf("%s: want a refusal naming includeFromSettings, got %v", raw, err)
			continue
		}
		// The refusal says why (no shared settings) and where the components go instead (ruled on
		// #e06zm4b), so an author copying an ESLint config is told what to write.
		for _, want := range []string{"does not read settings.linkComponents", `[{"name": "Link", "props": ["to"]}]`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the refusal does not say %q: %v", raw, want, err)
			}
		}
	}
}

// TestJsxNoScriptUrlProtocolScan probes the predicate directly at its edges.
//
// The scan is the whole judgment and its two character classes differ, which no corpus case
// separates: upstream writes only one interleaved case and no prefix case at all. Every expectation
// here was measured by driving the installed rule on the same value.
func TestJsxNoScriptUrlProtocolScan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		value string
		want  bool
	}{
		{"javascript:", true},
		{"javascript:void(0)", true},
		{"JAVASCRIPT:", true},
		{"JaVaScRiPt:", true},
		{" javascript:", true},
		{"\x00javascript:", true},
		{"\x01javascript:", true},
		{"\x1fjavascript:", true},
		{"\x7fjavascript:", false},
		// A non-breaking space is not in the prefix class, whose members are all below 0x80.
		// This is why the scan may compare bytes: a multi-byte character fails on its first
		// byte, which is the right answer. Written as an escape so it cannot be mistaken for
		// an ordinary space, and measured against the installed build.
		{" javascript:", false},
		{"  javascript:", true},
		{"ja\nvascript:", true},
		{"ja\rvascript:", true},
		{"ja\tvascript:", true},
		{"ja vascript:", false},
		{"javascript\t:", true},
		{"javascript", false},
		{"", false},
		{"j", false},
		{"https://reactjs.org", false},
		{"mailto:foo@bar.com", false},
		{"#", false},
		{"xjavascript:", false},
		// The prefix class may not appear between the letters, and the interleaving class may not
		// appear before them. This pair is the asymmetry stated as two inputs.
		{"\x01ja\x01vascript:", false},
	}

	for _, testCase := range cases {
		if got := jsxNoScriptUrlHasJavaScriptProtocol(testCase.value); got != testCase.want {
			t.Errorf("protocol scan of %q is %v, want %v", testCase.value, got, testCase.want)
		}
	}
}
