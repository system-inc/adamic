package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// arrowBodyStyleFile is where the fixtures pretend to live.
const arrowBodyStyleFile = "/repository/source/ArrowBodyStyle.ts"

// arrowBodyStyleCase is one imported corpus row.
//
// `wantFixedSource` empty means the case is reported and deliberately NOT repaired, which upstream
// records as `output: null` and which is a decision rather than an omission: seven of its cases are
// in that state, and a fixer that repaired any of them would be a defect no message-id fixture
// could see.
type arrowBodyStyleCase struct {
	sourceText      string
	options         any
	wantIds         []string
	wantFixedSource string
}

// runArrowBodyStyle drives one case, routing options through the rule's own exported decoder.
//
// Through the decoder rather than by building the struct, because the decoder is where the default
// mode is applied and where an unknown mode is rejected, and neither line has an upstream
// counterpart to inherit correctness from.
func runArrowBodyStyle(t *testing.T, testCase arrowBodyStyleCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, ArrowBodyStyle, arrowBodyStyleFile, testCase.sourceText)
	}
	decoded, err := DecodeArrowBodyStyleOptions([]byte(arrowBodyStyleUpstreamList(testCase.options.(ArrowBodyStyleOptions))))
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, ArrowBodyStyle, arrowBodyStyleFile,
		testCase.sourceText, decoded)
}

// arrowBodyStyleUpstreamList writes a case's options the way upstream's config writes them, so
// every case crosses the decoder on the list shape the config layer delivers.
func arrowBodyStyleUpstreamList(options ArrowBodyStyleOptions) string {
	mode := map[ArrowBodyStyleMode]string{
		"":                     "as-needed",
		ArrowBodyStyleAsNeeded: "as-needed",
		ArrowBodyStyleAlways:   "always",
		ArrowBodyStyleNever:    "never",
	}[options.Mode]
	if !options.RequireReturnForObjectLiteral {
		return `["` + mode + `"]`
	}
	return `["` + mode + `", {"requireReturnForObjectLiteral": true}]`
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/arrow-body-style.js` was loaded with its RuleTester stubbed so every case came
// out as data, then each was replayed against the INSTALLED rule in `node_modules` to record both
// what it reports and what its fixer writes. All 87 findings and all 63 fix outputs reproduced, so
// everything below is a measurement rather than a transcription.
//
// Upstream's options are a positional list whose second element is only legal beside one spelling
// of the first. Each case names the struct, and `arrowBodyStyleUpstreamList` writes it back as that
// list, so the decoder reads upstream's spelling. Only the Go spelling of the mode moved; the
// decision each selects is identical, and the mapping is applied by the generator rather than by hand.
func arrowBodyStyleFiresCases() []arrowBodyStyleCase {
	return []arrowBodyStyleCase{
		{"for (var foo = () => { return a in b ? bar : () => {} } ;;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var foo = () => (a in b ? bar : () => {}) ;;);"},
		{"a in b; for (var f = () => { return c };;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "a in b; for (var f = () => c;;);"},
		{"for (a = b => { return c in d ? e : f } ;;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (a = b => (c in d ? e : f) ;;);"},
		{"for (var f = () => { return a };;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var f = () => a;;);"},
		{"for (var f;f = () => { return a };);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var f;f = () => a;);"},
		{"for (var f = () => { return a in c };;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var f = () => (a in c);;);"},
		{"for (var f;f = () => { return a in c };);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var f;f = () => a in c;);"},
		{"for (;;){var f = () => { return a in c }}", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (;;){var f = () => a in c}"},
		{"for (a = b => { return c = d in e } ;;);", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (a = b => (c = d in e) ;;);"},
		{"for (var a;;a = b => { return c = d in e } );", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "for (var a;;a = b => c = d in e );"},
		{"for (let a = (b, c, d) => { return vb && c in d; }; ;);", nil, []string{"unexpectedSingleBlock"}, "for (let a = (b, c, d) => (vb && c in d); ;);"},
		{"for (let a = (b, c, d) => { return v in b && c in d; }; ;);", nil, []string{"unexpectedSingleBlock"}, "for (let a = (b, c, d) => (v in b && c in d); ;);"},
		{"function foo(){ for (let a = (b, c, d) => { return v in b && c in d; }; ;); }", nil, []string{"unexpectedSingleBlock"}, "function foo(){ for (let a = (b, c, d) => (v in b && c in d); ;); }"},
		{"for ( a = (b, c, d) => { return v in b && c in d; }; ;);", nil, []string{"unexpectedSingleBlock"}, "for ( a = (b, c, d) => (v in b && c in d); ;);"},
		{"for ( a = (b) => { return (c in d) }; ;);", nil, []string{"unexpectedSingleBlock"}, "for ( a = (b) => (c in d); ;);"},
		{"for (let a = (b, c, d) => { return vb in dd ; }; ;);", nil, []string{"unexpectedSingleBlock"}, "for (let a = (b, c, d) => (vb in dd ); ;);"},
		{"for (let a = (b, c, d) => { return vb in c in dd ; }; ;);", nil, []string{"unexpectedSingleBlock"}, "for (let a = (b, c, d) => (vb in c in dd ); ;);"},
		{"do{let a = () => {return f in ff}}while(true){}", nil, []string{"unexpectedSingleBlock"}, "do{let a = () => f in ff}while(true){}"},
		{"do{for (let a = (b, c, d) => { return vb in c in dd ; }; ;);}while(true){}", nil, []string{"unexpectedSingleBlock"}, "do{for (let a = (b, c, d) => (vb in c in dd ); ;);}while(true){}"},
		{"scores.map(score => { return x in +(score / maxScore).toFixed(2)});", nil, []string{"unexpectedSingleBlock"}, "scores.map(score => x in +(score / maxScore).toFixed(2));"},
		{"const fn = (a, b) => { return a + x in Number(b) };", nil, []string{"unexpectedSingleBlock"}, "const fn = (a, b) => a + x in Number(b);"},
		{"var foo = () => 0", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return 0}"},
		{"var foo = () => 0;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return 0};"},
		{"var foo = () => ({});", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return {}};"},
		{"var foo = () => (  {});", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return   {}};"},
		{"(() => ({}))", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "(() => {return {}})"},
		{"(() => ( {}))", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "(() => {return  {}})"},
		{"var foo = () => { return 0; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => 0;"},
		{"var foo = () => { return 0 };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => 0;"},
		{"var foo = () => { return bar(); };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => bar();"},
		{"var foo = () => {};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedEmptyBlock"}, ""},
		{"var foo = () => {\nreturn 0;\n};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedSingleBlock"}, "var foo = () => 0;"},
		{"var foo = () => { return { bar: 0 }; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedObjectBlock"}, "var foo = () => ({ bar: 0 });"},
		{"var foo = () => { return ({ bar: 0 }); };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => ({ bar: 0 });"},
		{"var foo = () => { return a, b }", nil, []string{"unexpectedSingleBlock"}, "var foo = () => (a, b)"},
		{"var foo = () => { return };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return ( /* a */ {ok: true} /* b */ ) };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => ( /* a */ {ok: true} /* b */ );"},
		{"var foo = () => { return '{' };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = () => '{';"},
		{"var foo = () => { return { bar: 0 }.bar; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedObjectBlock"}, "var foo = () => ({ bar: 0 }.bar);"},
		{"var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedOtherBlock"}, ""},
		{"var foo = () => { bar };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedOtherBlock"}, ""},
		{"var foo = () => { return 0; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"unexpectedSingleBlock"}, "var foo = () => 0;"},
		{"var foo = () => { return bar(); };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"unexpectedSingleBlock"}, "var foo = () => bar();"},
		{"var foo = () => ({});", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"expectedBlock"}, "var foo = () => {return {}};"},
		{"var foo = () => ({ bar: 0 });", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{"expectedBlock"}, "var foo = () => {return { bar: 0 }};"},
		{"var foo = () => (((((((5)))))));", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return (((((((5)))))))};"},
		{"var foo = () => { return bar }\n[1, 2, 3].map(foo)", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return bar }\n(1).toString();", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return bar };\n[1, 2, 3].map(foo)", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{"unexpectedSingleBlock"}, "var foo = () => bar;\n[1, 2, 3].map(foo)"},
		{"var foo = /* a */ ( /* b */ ) /* c */ => /* d */ { /* e */ return /* f */ 5 /* g */ ; /* h */ } /* i */ ;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}, []string{"unexpectedSingleBlock"}, "var foo = /* a */ ( /* b */ ) /* c */ => /* d */  /* e */  /* f */ 5 /* g */  /* h */  /* i */ ;"},
		{"var foo = /* a */ ( /* b */ ) /* c */ => /* d */ ( /* e */ 5 /* f */ ) /* g */ ;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = /* a */ ( /* b */ ) /* c */ => /* d */ {return ( /* e */ 5 /* f */ )} /* g */ ;"},
		{"var foo = () => {\nreturn bar;\n};", nil, []string{"unexpectedSingleBlock"}, "var foo = () => bar;"},
		{"var foo = () => {\nreturn bar;};", nil, []string{"unexpectedSingleBlock"}, "var foo = () => bar;"},
		{"var foo = () => {return bar;\n};", nil, []string{"unexpectedSingleBlock"}, "var foo = () => bar;"},
		{"\n              var foo = () => {\n                return foo\n                  .bar;\n              };\n            ", nil, []string{"unexpectedSingleBlock"}, "\n              var foo = () => foo\n                  .bar;\n            "},
		{"\n              var foo = () => {\n                return {\n                  bar: 1,\n                  baz: 2\n                };\n              };\n            ", nil, []string{"unexpectedObjectBlock"}, "\n              var foo = () => ({\n                  bar: 1,\n                  baz: 2\n                });\n            "},
		{"var foo = () => ({foo: 1}).foo();", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return {foo: 1}.foo()};"},
		{"var foo = () => ({foo: 1}.foo());", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return {foo: 1}.foo()};"},
		{"var foo = () => ( {foo: 1} ).foo();", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "var foo = () => {return  {foo: 1} .foo()};"},
		{"\n              var foo = () => ({\n                  bar: 1,\n                  baz: 2\n                });\n            ", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "\n              var foo = () => {return {\n                  bar: 1,\n                  baz: 2\n                }};\n            "},
		{"\n              parsedYears = _map(years, (year) => (\n                  {\n                      index : year,\n                      title : splitYear(year)\n                  }\n              ));\n            ", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "\n              parsedYears = _map(years, (year) => {\n                  return {\n                      index : year,\n                      title : splitYear(year)\n                  }\n              });\n            "},
		{"const createMarker = (color) => ({ latitude, longitude }, index) => {};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{"expectedBlock"}, "const createMarker = (color) => {return ({ latitude, longitude }, index) => {}};"},
	}
}

func arrowBodyStyleSilentCases() []arrowBodyStyleCase {
	return []arrowBodyStyleCase{
		{"var foo = () => {};", nil, []string{}, ""},
		{"var foo = () => 0;", nil, []string{}, ""},
		{"var addToB = (a) => { b =  b + a };", nil, []string{}, ""},
		{"var foo = () => { /* do nothing */ };", nil, []string{}, ""},
		{"var foo = () => {\n /* do nothing */ \n};", nil, []string{}, ""},
		{"var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};", nil, []string{}, ""},
		{"var foo = () => ({});", nil, []string{}, ""},
		{"var foo = () => bar();", nil, []string{}, ""},
		{"var foo = () => { bar(); };", nil, []string{}, ""},
		{"var foo = () => { b = a };", nil, []string{}, ""},
		{"var foo = () => { bar: 1 };", nil, []string{}, ""},
		{"var foo = () => { return 0; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{}, ""},
		{"var foo = () => { return bar(); };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, []string{}, ""},
		{"var foo = () => 0;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{}, ""},
		{"var foo = () => ({ foo: 0 });", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, []string{}, ""},
		{"var foo = () => {};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => 0;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var addToB = (a) => { b =  b + a };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => { /* do nothing */ };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => {\n /* do nothing */ \n};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => bar();", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => { bar(); };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
		{"var foo = () => { return { bar: 0 }; };", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded, RequireReturnForObjectLiteral: true}, []string{}, ""},
	}
}

func TestArrowBodyStyleFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrowBodyStyleFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runArrowBodyStyle(t, testCase), testCase.wantIds...)
		})
	}
}

func TestArrowBodyStyleStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrowBodyStyleSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runArrowBodyStyle(t, testCase))
		})
	}
}

// What the fixer WRITES, which is half this rule and which no message-id fixture can see.
//
// Every invalid case in a fixable rule's corpus carries the exact text the repair must produce, and
// this rule has 56 of them plus 7 deliberate declines. A fixer that repairs the right span with the
// wrong text, or repairs a case upstream refuses to touch, passes every assertion above.
//
// `ExpectFixedSource` compares the whole rewritten file, so the expectation is upstream's `output`
// verbatim. Worth stating because the neighbouring hazard is real and points the other way: the
// TYPED harness writes each fixture as `TrimSpace(source) + "\n"` (`rule_testing/program.go:160`),
// so a typed rule's expectation must be transformed the same way or every byte-correct repair fails
// on a trailing newline. `Run` does not trim, this rule needs no checker, and transforming here
// instead broke all 56 rows before the difference was measured rather than assumed.
func TestArrowBodyStyleFixesTheSource(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrowBodyStyleFiresCases() {
		if testCase.wantFixedSource == "" {
			continue
		}
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runArrowBodyStyle(t, testCase),
				testCase.wantFixedSource)
		})
	}
}

// The seven cases upstream reports and deliberately declines to repair.
//
// Two mechanisms produce them. A block that does real work has no concise equivalent, so under the
// Never mode the finding stands with nothing to offer. And a following line beginning with one of
// `+"`"+`([/`+"`"+`, a backtick, plus or minus would be joined onto the shortened expression once the closing
// brace stops separating them, so the repair is declined rather than made safe.
//
// Asserted as carrying NO fix at all, rather than as a fix that happens to be a no-op: those are
// different artifacts and only the first is what upstream ships.
func TestArrowBodyStyleDeclinesToFix(t *testing.T) {
	t.Parallel()

	declined := 0
	for _, testCase := range arrowBodyStyleFiresCases() {
		if testCase.wantFixedSource != "" {
			continue
		}
		declined++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, testCase)
			if len(result.Diagnostics) == 0 {
				t.Fatal("wanted a finding")
			}
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carried %d fixes, wanted none", index,
						len(diagnostic.Fixes))
				}
			}
		})
	}
	if declined != 7 {
		t.Errorf("the corpus carried %d declined repairs, wanted 7", declined)
	}
}

// Where the finding points, which no message-id fixture can see.
//
// Upstream reports the arrow function node but overrides the location to the BODY, which is a
// distinction invisible from the message id and easy to get wrong in the direction that still looks
// plausible: pointing at the whole arrow reads fine in a terminal and puts the caret on `var`.
func TestArrowBodyStylePointsAtTheBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantSpan   string
	}{
		{"var foo = () => { return 0; };", nil, "{ return 0; }"},
		{"var foo = () => {};", ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}, "{}"},
		{"var foo = () => 0;", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways}, "0"},
		{"var foo = () => ({ bar: 1 });", ArrowBodyStyleOptions{Mode: ArrowBodyStyleAlways},
			"({ bar: 1 })"},
		{"var foo = () => { return { bar: 1 }; };", nil, "{ return { bar: 1 }; }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, arrowBodyStyleCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			span := result.Diagnostics[0].Range
			if got := testCase.sourceText[span.Pos():span.End()]; got != testCase.wantSpan {
				t.Errorf("pointed at %q, wanted %q", got, testCase.wantSpan)
			}
		})
	}
}

// The four braced-body messages, which the corpus exercises but which no id assertion pins to a
// shape.
//
// Upstream chooses between them by what the block holds, because the remedy differs in each case,
// and a rule reporting one id for all four would pass a count-based fixture set. These are the four
// discriminating shapes, one per id.
func TestArrowBodyStyleChoosesTheMessageByShape(t *testing.T) {
	t.Parallel()

	never := ArrowBodyStyleOptions{Mode: ArrowBodyStyleNever}
	cases := []arrowBodyStyleCase{
		{"var foo = () => {};", never, []string{"unexpectedEmptyBlock"}, ""},
		{"var foo = () => { bar(); };", never, []string{"unexpectedOtherBlock"}, ""},
		{"var foo = () => { return; };", never, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return 0; };", never, []string{"unexpectedSingleBlock"}, ""},
		{"var foo = () => { return { bar: 1 }; };", never, []string{"unexpectedObjectBlock"}, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runArrowBodyStyle(t, testCase), testCase.wantIds...)
		})
	}
}

// The decoder, which reads upstream's option list and whose second element is only legal beside one
// spelling of the first. The default, the mode-dependent second element and the rejection of an
// unknown mode are written here rather than inherited, and none would be exercised by a fixture that
// built the options struct directly.
func TestDecodeArrowBodyStyleOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input selects upstream's default mode", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeArrowBodyStyleOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode := decoded.(ArrowBodyStyleOptions).Mode; mode != ArrowBodyStyleAsNeeded {
			t.Errorf("mode came back %q, wanted AsNeeded", mode)
		}
	})

	t.Run("the second element is read beside as-needed", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeArrowBodyStyleOptions(
			[]byte(`["as-needed", {"requireReturnForObjectLiteral": true}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ArrowBodyStyleOptions)
		if options.Mode != ArrowBodyStyleAsNeeded {
			t.Errorf("mode came back %q, wanted AsNeeded", options.Mode)
		}
		if !options.RequireReturnForObjectLiteral {
			t.Error("the second element was handed over and not read")
		}
	})

	t.Run("a second element beside always is refused rather than ignored", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeArrowBodyStyleOptions(
			[]byte(`["always", {"requireReturnForObjectLiteral": true}]`)); err == nil {
			t.Error("an option the always mode never reads decoded")
		}
	})

	t.Run("a third element is refused", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeArrowBodyStyleOptions(
			[]byte(`["as-needed", {"requireReturnForObjectLiteral": true}, "never"]`)); err == nil {
			t.Error("a third element decoded and would have been dropped")
		}
	})

	t.Run("an unknown key in the second element is refused", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeArrowBodyStyleOptions(
			[]byte(`["as-needed", {"requireReturnForObjectLiterals": true}]`)); err == nil {
			t.Error("a misspelled key decoded, so the option it carried would be ignored")
		}
	})

	t.Run("an unknown mode is rejected rather than silently disabling the rule", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeArrowBodyStyleOptions([]byte(`["AsNeeded"]`)); err == nil {
			t.Error("the Go constant's spelling decoded; it is not one of upstream's three modes")
		}
	})

	t.Run("the retired object spelling is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeArrowBodyStyleOptions([]byte(`{"mode":"Always"}`)); err == nil {
			t.Error("the pre-list object spelling decoded, which the config layer never delivers")
		}
	})
}

// A rule configured as a bare severity is handed nil, and must still enforce the default mode.
//
// `options.(T)` on nil yields the zero value, whose empty mode matches no arm in the rule, so
// without the fallback the rule would register on every file and report nothing. Every fixture
// above reaches the rule through the decoder, so none of them can see this.
func TestArrowBodyStyleWithNilOptionsUsesAsNeeded(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, ArrowBodyStyle, arrowBodyStyleFile,
		"var foo = () => { return 0; };"), "unexpectedSingleBlock")
	// Always and Never both differ from AsNeeded here, so this pins WHICH default was chosen
	// rather than merely that some default was.
	rule_testing.ExpectClean(t, rule_testing.Run(t, ArrowBodyStyle, arrowBodyStyleFile,
		"var foo = () => 0;"))
	rule_testing.ExpectClean(t, rule_testing.Run(t, ArrowBodyStyle, arrowBodyStyleFile,
		"var foo = () => { bar(); };"))
}

// A comment after the body is trivia, and must not decide the automatic-semicolon-insertion guard.
//
// Written for a defect this port had: the guard scanned bytes and stopped at the first
// non-whitespace character, so the `/` opening a trailing block comment read as the division
// operator and every repair with a trailing comment was declined. Upstream asks for the next TOKEN,
// which skips comments, and its own nine-comment case is repaired.
//
// The three rows separate the two things a `/` can start. Every verdict was measured against the
// installed rule before it was written down.
func TestArrowBodyStyleSkipsCommentsWhenCheckingTheNextToken(t *testing.T) {
	t.Parallel()

	cases := []arrowBodyStyleCase{
		// A comment then a semicolon: repaired, and the comment survives.
		{"var foo = () => { return bar } /* c */ ;", nil, []string{"unexpectedSingleBlock"},
			"var foo = () => bar /* c */ ;"},
		// A comment then a token that WOULD join on: declined, so the comment did not mask it.
		{"var foo = () => { return bar } /* c */ + 1;", nil, []string{"unexpectedSingleBlock"}, ""},
		// A line comment, whose ending is a newline rather than a delimiter.
		{"var foo = () => { return bar } // c\n[1,2].map(foo)", nil,
			[]string{"unexpectedSingleBlock"}, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if testCase.wantFixedSource == "" {
				for index, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d carried a repair, wanted none", index)
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// The `in` operator, whose handling replaces upstream's frame stack and whose breadth is the part
// most likely to be quietly wrong.
//
// Upstream sets a flag on every enclosing ARROW frame when it meets an `in`, and nothing else
// pushes a frame, so an `in` inside a nested `function`, a class method, or a bracketed subscript
// reaches the outer arrow too. This port asks the subtree instead, which is the same answer reached
// without a stack, and these rows are what pin that the two agree. All six were measured against
// the installed rule.
//
// The clause discrimination is the other half: only a for statement's INITIALIZER makes `in`
// ambiguous, so the same arrow in the condition clause needs no parentheses.
func TestArrowBodyStyleParenthesizesTheInOperatorLikeUpstream(t *testing.T) {
	t.Parallel()

	cases := []arrowBodyStyleCase{
		// The initializer clause: wrapped.
		{"for (var f = () => { return a in c };;);", nil, []string{"unexpectedSingleBlock"},
			"for (var f = () => (a in c);;);"},
		// The condition clause: not wrapped, though the arrow and the operator are identical.
		{"for (var f;f = () => { return a in c };);", nil, []string{"unexpectedSingleBlock"},
			"for (var f;f = () => a in c;);"},
		// The body: not wrapped either.
		{"for (;;){var f = () => { return a in c }}", nil, []string{"unexpectedSingleBlock"},
			"for (;;){var f = () => a in c}"},
		// Outside any for statement: not wrapped.
		{"var f = () => { return a in c };", nil, []string{"unexpectedSingleBlock"},
			"var f = () => a in c;"},
		// An `in` the grammar could never confuse, inside brackets, still wraps -- because
		// upstream's flag does not care where the operator sat. Reproduced, not tightened.
		{"for (var f = () => { return a[b in c] };;);", nil, []string{"unexpectedSingleBlock"},
			"for (var f = () => (a[b in c]);;);"},
		// An `in` inside a nested `function`, which pushes no frame upstream, reaches out too.
		{"for (var f = () => { return g(function(){ return a in c }) };;);", nil,
			[]string{"unexpectedSingleBlock"},
			"for (var f = () => (g(function(){ return a in c }));;);"},
		// The control: no `in` anywhere, so the initializer clause alone changes nothing.
		{"for (var f = () => { return a };;);", nil, []string{"unexpectedSingleBlock"},
			"for (var f = () => a;;);"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// A comment on ONE side of the returned value, which the imported corpus does not separate.
//
// Written for a surviving mutant. The repair takes a comment-preserving branch when trivia sits
// anywhere in the ceremony being deleted, and the check is two halves: comments before the value
// and comments after it. Upstream's only comment case threads them through both sides at once, so
// disabling either half alone left every fixture green while the other half covered for it.
//
// Each row here has a comment on exactly one side, so each half is load-bearing on its own row.
// All four outputs were measured against the installed rule before being written down: the
// preserving branch removes only the tokens and leaves the whitespace, which is why the repaired
// source carries doubled spaces that the ordinary branch would have collapsed.
func TestArrowBodyStyleKeepsCommentsOnEitherSide(t *testing.T) {
	t.Parallel()

	cases := []arrowBodyStyleCase{
		// Before the return keyword: only the first half of the check sees this.
		{"var foo = () => { /* leading */ return 5; };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () =>  /* leading */  5 ;"},
		{"var foo = () => { /* leading */ return 5 };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () =>  /* leading */  5 ;"},
		// Between the keyword and the value, which is also before the value.
		{"var foo = () => { return /* mid */ 5; };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () =>   /* mid */ 5 ;"},
		// Between the value and its semicolon: the trailing window starts AFTER the semicolon, so
		// this takes the ordinary branch and the comment survives inside the kept span. The
		// doubled spaces in the rows below are what distinguish the two branches by eye.
		{"var foo = () => { return 5 /* trailing */; };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () => 5 /* trailing */;"},
		// After the semicolon: this is what the trailing half of the check exists for, and the
		// only shape that reaches it. Without these two rows the whole second half could be
		// disabled with every fixture staying green, which is how it was found.
		{"var foo = () => { return 5; /* after */ };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () =>   5 /* after */ ;"},
		{"var foo = () => { return 5 /* after */ };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () =>   5 /* after */ ;"},
		// The control: no comment at all takes the ordinary branch, which collapses the space.
		{"var foo = () => { return 5; };", nil, []string{"unexpectedSingleBlock"},
			"var foo = () => 5;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// No two fixes may share a boundary, because the engine drops one of them when they do.
//
// This exists because of a defect that shipped past all 87 imported cases. The repair emitted each
// parenthesis as its own zero-width insertion, at a position that was also the edge of a removal.
// `ExpectFixedSource` applies every proposed fix in order and produced the right text; the real
// engine treats a shared boundary as an overlap, dropped the insertion, and rewrote
// `() => { return { a: 1 }; }` into `() => { a: 1 }` -- which still parses, as a labelled block, and
// means something else. A fix producing source that parses and does not mean the same thing is the
// one failure the engine structurally cannot refuse.
//
// So the check is structural rather than textual: assert directly that the ranges are pairwise
// disjoint. A test that only compared the rewritten source would keep passing under the same defect,
// which is exactly what happened.
func TestArrowBodyStyleProposesDisjointFixes(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrowBodyStyleFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runArrowBodyStyle(t, testCase)
			type span struct{ start, end int }
			var spans []span
			for _, diagnostic := range result.Diagnostics {
				for _, fix := range diagnostic.Fixes {
					spans = append(spans, span{fix.Range.Pos(), fix.Range.End()})
				}
			}
			for first := range spans {
				for second := first + 1; second < len(spans); second++ {
					left, right := spans[first], spans[second]
					if left.start > right.start {
						left, right = right, left
					}
					// Touching counts. Two edits meeting at one offset, including a zero-width one
					// sitting on the edge of a removal, are what the engine refuses.
					if right.start <= left.end && !(left.start == left.end &&
						right.start == right.end && left.start != right.start) {
						if right.start < left.end || left.start == left.end ||
							right.start == right.end {
							t.Errorf("fixes [%d,%d) and [%d,%d) touch or overlap, so the engine "+
								"will drop one", left.start, left.end, right.start, right.end)
						}
					}
				}
			}
		})
	}
}
