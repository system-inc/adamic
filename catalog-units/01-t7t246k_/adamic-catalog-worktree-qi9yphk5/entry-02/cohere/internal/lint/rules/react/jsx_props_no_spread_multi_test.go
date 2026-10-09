package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxPropsNoSpreadMultiFile is where the fixtures pretend to live. A .tsx extension, because the
// rule only ever sees JSX and upstream gates itself on `source_type().is_jsx()`.
const jsxPropsNoSpreadMultiFile = "/repository/source/Spread.tsx"

// The corpus is oxc's, copied rather than rewritten, and written into this file by a script reading
// the extractor's own dump rather than by hand.
//
// Every case is verbatim from `oxc/crates/oxc_linter/src/rules/react/jsx_props_no_spread_multi.rs`:
// 4 pass, 5 fail, one tester block, and `react_jsx_props_no_spread_multi.snap` records 5
// diagnostics from those 5 fail inputs, so one finding per input is established here rather than
// assumed. The extractor reported no discrepancy and there is only one snapshot file for this rule,
// so the multi-snapshot miscount the brief warns about does not apply.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and it is
// upstream's clean cases that catch a wrong belief. Two of the four do real work here: the third
// separates two different properties of one receiver, and the fourth is the only parenthesized
// input in either corpus and is what pins that the receivers are compared rather than the text.
func TestJsxPropsNoSpreadMultiFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"the same identifier spread twice", "\n          const props = {};\n          <App {...props} {...props} />\n        ", []string{"jsxPropsNoSpreadMultiIdentifier"}},
		{"the same property spread twice", "\n          const props = {};\n          <App {...props.foo} {...props.foo} />\n        ", []string{"jsxPropsNoSpreadMultiMemberExpression"}},
		{"one property spread twice, written with different parentheses", "\n          const props = {};\n          <App {...(props.foo).baz} {...(props.foo.baz)} />\n        ", []string{"jsxPropsNoSpreadMultiMemberExpression"}},
		{"the same identifier spread across an intervening attribute", "\n          const props = {};\n          <div {...props} a=\"a\" {...props} />\n        ", []string{"jsxPropsNoSpreadMultiIdentifier"}},
		{"the same identifier spread three times", "\n          const props = {};\n          <div {...props} {...props} {...props} />\n        ", []string{"jsxPropsNoSpreadMultiIdentifier"}},
		// Beyond upstream. Each exists because the imported corpus cannot see the distinction, and
		// each was measured on the release oxlint binary before being written down.

		// THE COUNT ASYMMETRY, and the single most important case in this file.
		//
		// Three copies of a member expression report THREE times, because the member half compares
		// every unordered pair, while three copies of an identifier report ONCE, because the
		// identifier half groups by name. Upstream never writes a third copy of a member
		// expression, so all five imported fail cases pass under either arithmetic and the corpus
		// cannot tell the two readings apart. Measured: the release binary prints three findings
		// for this input, at columns 16, 16 and 31.
		{
			"the same property spread three times, which reports once per pair",
			"const props = {};\nconst a = <App {...props.foo} {...props.foo} {...props.foo} />;\n",
			[]string{
				"jsxPropsNoSpreadMultiMemberExpression",
				"jsxPropsNoSpreadMultiMemberExpression",
				"jsxPropsNoSpreadMultiMemberExpression",
			},
		},

		// Parentheses peel on the identifier half. No imported fixture writes a parenthesized
		// identifier, so guessing this direction costs nothing at fixture time and ships a
		// divergence either way. Measured: reports.
		{
			"a parenthesized identifier against a bare one",
			"const props = {};\nconst a = <App {...(props)} {...props} />;\n",
			[]string{"jsxPropsNoSpreadMultiIdentifier"},
		},

		// The TypeScript wrappers peel through the same accessor, because oxc's
		// `get_identifier_reference` routes through `get_inner_expression`, which strips `as`,
		// `satisfies`, `!`, instantiation and type assertion alongside parentheses. Measured on
		// both spellings: reports.
		{
			"a non-null assertion against a type assertion",
			"const props = {};\nconst a = <App {...props!} {...props as any} />;\n",
			[]string{"jsxPropsNoSpreadMultiIdentifier"},
		},

		// A static computed key equals the dotted spelling. `AccessedName` is what answers this,
		// and it is the reason that helper is reached for rather than comparing node text.
		// Measured: reports.
		{
			"a string subscript against the dotted spelling",
			"const props = {};\nconst a = <App {...props[\"foo\"]} {...props.foo} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A VARIABLE subscript compares equal to itself, which is where this rule departs from
		// every other same-reference judgment in the tree. `property.AccessedName` declines a
		// variable subscript by design and `no_self_assign.go` returns false for `a[i]` against
		// itself; both are right for their question and wrong for this one, because nothing can run
		// between two spreads inside one element so `k` cannot change. Measured: reports.
		{
			"a variable subscript spread twice",
			"const props = {};\nconst a = <App {...props[k]} {...props[k]} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A `this` receiver, which is the commonest real spelling of this bug in class components
		// and which upstream's corpus never writes. Measured: reports.
		{
			"a this-receiver property spread twice",
			"const a = <App {...this.props} {...this.props} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A deeper receiver chain, to show the comparison recurses rather than stopping at the
		// first receiver. Measured: reports.
		{
			"a two-level property chain spread twice",
			"const props = {};\nconst a = <App {...props.a.b} {...props.a.b} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// The element with children rather than a self-closing one. Every imported fail case is
		// self-closing, so without this case the KindJsxOpeningElement arm is never exercised by a
		// firing input and could be deleted with the suite still green.
		{
			"an element with children rather than a self-closing one",
			"const props = {};\nconst a = <div {...props} {...props}>text</div>;\n",
			[]string{"jsxPropsNoSpreadMultiIdentifier"},
		},

		// A PARENTHESIZED PREFIX terminates the optional chain, so the outer access is ordinary and
		// this reports where the unparenthesized spelling is silent. Nothing in either corpus writes
		// this, and the rule's first version was silent here. Measured: reports.
		{
			"an optional chain whose prefix is parenthesized",
			"const a = <App {...(props?.a).b} {...(props?.a).b} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// And a terminated chain compares EQUAL to the non-optional spelling, because the
		// comparison unwraps it. Measured: reports.
		{
			"a parenthesized optional prefix against the plain spelling",
			"const a = <App {...(props?.a).b} {...(props.a).b} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A chain NESTED in a subscript is ordinary structure. The exclusion applies to the spread
		// argument alone, which is the whole reason isSameExpression uses a plain access check.
		// Measured: reports.
		{
			"an optional chain inside a subscript, against the plain spelling",
			"const a = <App {...props[x?.y]} {...props[x.y]} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A nested literal subscript that matches, which is what exercises the literal comparison
		// arm in the reporting direction. The outer subscript is non-static so AccessedName
		// declines it and the structural path runs. Measured: reports.
		{
			"matching nested string subscripts",
			"const a = <App {...props[o[\"x\"]]} {...props[o[\"x\"]]} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// A LITERAL RECEIVER, which is the only position where the literal comparison arm is
		// reachable. A nested literal subscript never falls through to it, because the tagged name
		// reader accepts every literal kind, so without a literal receiver that whole arm is dead
		// code that no fixture exercises. Found by mutation. Measured: reports.
		{
			"a string-literal receiver spread twice",
			"const a = <App {...\"abc\".length} {...\"abc\".length} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},
		{
			"a regular-expression-literal receiver spread twice",
			"const a = <App {.../r/.source} {.../r/.source} />;\n",
			[]string{"jsxPropsNoSpreadMultiMemberExpression"},
		},

		// Two independent groups on one element, which pins that the identifier map is keyed by
		// name rather than counting spreads. Measured: two findings.
		{
			"two different identifiers each spread twice",
			"const a = <App {...one} {...two} {...one} {...two} />;\n",
			[]string{
				"jsxPropsNoSpreadMultiIdentifier",
				"jsxPropsNoSpreadMultiIdentifier",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile, testCase.sourceText),
				testCase.findings...)
		})
	}
}

// The clean cases are the whole discrimination, and the imported four are joined by the shapes our
// tree can produce that oxc's cannot.
func TestJsxPropsNoSpreadMultiStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a single spread", "\n          const a = {};\n          <App {...a} />\n        "},
		{"two spreads of different identifiers", "\n          const a = {};\n          const b = {};\n          <App {...a} {...b} />\n        "},
		{"two spreads of different properties of one receiver", "\n        const props = {};\n        <App {...props.x} {...props.foo} />\n      "},
		{"parenthesized member spreads whose receivers differ", "\n        const props = {};\n        <App {...(props.foo).baz} {...(props.y.baz)} />\n      "},
		// Beyond upstream, each measured silent on the release binary.

		// OPTIONAL CHAINING, and the only case here where reproducing upstream costs an extra
		// condition rather than one fewer. oxc wraps an optional chain in a `ChainExpression`, a
		// sibling of `MemberExpression`, so the spread never enters the collection. Our parser
		// produces a plain KindPropertyAccessExpression, so a literal translation of oxc's code
		// REPORTS here. Measured: the release binary prints nothing for this input.
		//
		// The intuitive reading is that this IS the same double spread and should report. That
		// reading is wrong about upstream and is recorded here so it is not helpfully restored.
		{"an optional chain spread twice", "const a = <App {...props?.foo} {...props?.foo} />;\n"},

		// The question-dot one level down the chain, which the first version of the predicate
		// missed by checking only the outermost link. oxc's ChainExpression wraps the whole chain,
		// so this is silent there too. Measured: silent.
		{"an optional chain with the question-dot deeper in", "const a = <App {...props?.a.b} {...props?.a.b} />;\n"},

		// Parenthesizing the WHOLE chain wraps the ChainExpression and changes nothing, unlike
		// parenthesizing a prefix. Measured: silent.
		{"a wholly parenthesized optional chain", "const a = <App {...(props?.a.b)} {...(props?.a.b)} />;\n"},

		// A call is neither an identifier nor a member expression, so it enters neither collection.
		// Correct rather than a gap: two calls may return different objects.
		{"the same call spread twice", "const a = <App {...props()} {...props()} />;\n"},

		// An object literal likewise. This is the shape that is invisible to a KindJsxAttribute
		// listener and perfectly visible to this rule, and it still reports nothing, because
		// deciding two object literals are the same spread would mean evaluating them.
		{"the same object literal spread twice", "const a = <App {...{a: 1}} {...{a: 1}} />;\n"},

		// The two halves are disjoint sets and neither falls back to the other, so an identifier
		// and a property of it never pair.
		{"an identifier against a property of it", "const props = {};\nconst a = <App {...props} {...props.foo} />;\n"},

		// A static key against a variable key is oxc's (Some, None) arm and answers false whatever
		// the receivers are.
		{"a static subscript against a variable one", "const a = <App {...props[\"k\"]} {...props[k]} />;\n"},

		// Two DIFFERENT variable subscripts, which is the discrimination the non-static compare
		// exists for. Without that compare, every pair of non-static subscripts reads as equal and
		// this reports. No imported fixture writes two different variable subscripts, so the blind
		// spot was found by mutation rather than by the corpus. Measured: silent.
		{"two different variable subscripts", "const a = <App {...props[x]} {...props[y]} />;\n"},

		// Two DIFFERENT nested string subscripts. Without comparing literal TEXT, every pair of
		// same-kind literals reads as equal and this reports. Found by mutation: no imported
		// fixture writes two literals that differ. Measured: silent.
		{"different nested string subscripts", "const a = <App {...props[o[\"x\"]]} {...props[o[\"y\"]]} />;\n"},

		// A numeric literal against a string literal whose text matches. Without the kind-equality
		// guard these compare equal on text alone and this reports, which is the second blind spot
		// mutation found. oxc keys its literal arms on matched PAIRS of variants, so a number and a
		// string never meet. Measured: silent.
		{"a numeric subscript against a string one with the same text", "const a = <App {...props[o[1]]} {...props[o[\"1\"]]} />;\n"},

		// Two DIFFERENT string-literal receivers. Without comparing literal TEXT every pair of
		// same-kind literals reads as equal and this reports. Measured: silent.
		{"different string-literal receivers", "const a = <App {...\"abc\".length} {...\"abd\".length} />;\n"},

		// And two different regular-expression receivers, which is the same discrimination on a
		// second literal kind. Measured: silent.
		{"different regular-expression-literal receivers", "const a = <App {.../r/.source} {.../q/.source} />;\n"},

		// A numeric-literal receiver against a string-literal one whose text agrees. Without the
		// kind-equality guard in the structural comparison these compare equal on text alone.
		// Measured: silent.
		{"a numeric-literal receiver against a string-literal one", "const a = <App {...(1).toFixed} {...\"1\".toFixed} />;\n"},

		// Different receivers with the same property name, which is what the recursion into the
		// receiver is for. Without it this would compare equal on the key alone.
		{"the same property of different receivers", "const a = <App {...one.foo} {...two.foo} />;\n"},

		// Spreads on two different elements never see each other, since the collections are rebuilt
		// per element.
		{"the same identifier spread once on each of two elements", "const a = <div><App {...props} /><App {...props} /></div>;\n"},

		// An element with no attributes at all, and one with only ordinary attributes, which is
		// most of the real tree.
		{"an element with no attributes", "const a = <App />;\n"},
		{"an element with only ordinary attributes", "const a = <App a=\"a\" b=\"b\" />;\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile, testCase.sourceText))
		})
	}
}

// Where the finding points, which ExpectFindings cannot see.
//
// A rule reporting the right message id at the wrong node passes every case above. oxc passes both
// spreads as labels and the FIRST is the diagnostic's primary position, read from the snapshot's
// caret row and confirmed against the release binary's json output, whose first label on
// `<App {...props} {...props} />` is offset 211 length 10, the earlier spread. So the earlier spread
// is what is reported here, and the reported text is the whole `{...expression}` attribute rather
// than the expression inside it.
//
// The parenthesized case is the one worth asserting separately: after peeling, the compared node is
// the inner member expression, whose range does not cover the parentheses the reader wrote. A rule
// reporting the peeled expression would point at `props.foo` inside `{...(props.foo).baz}`, which
// is not what upstream underlines.
func TestJsxPropsNoSpreadMultiReportsTheEarlierSpread(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{
			"an identifier spread, pointing at the first of the two",
			"const a = <App {...props} {...props} />;\n",
			[]string{"{...props}"},
		},
		{
			"an identifier spread across an intervening attribute",
			"const a = <div {...props} a=\"a\" {...props} />;\n",
			[]string{"{...props}"},
		},
		{
			"a member spread, pointing at the whole attribute rather than the expression",
			"const a = <App {...props.foo} {...props.foo} />;\n",
			[]string{"{...props.foo}"},
		},
		{
			"a parenthesized member spread, whose attribute text keeps the parentheses",
			"const a = <App {...(props.foo).baz} {...(props.foo.baz)} />;\n",
			[]string{"{...(props.foo).baz}"},
		},
		{
			"three copies of a member expression, pointing at the left of each pair",
			"const a = <App {...props.a} {...props.a} {...props.a} />;\n",
			[]string{"{...props.a}", "{...props.a}", "{...props.a}"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("want %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			for index, want := range testCase.want {
				diagnostic := result.Diagnostics[index]
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Fatalf("finding %d: want reported text %q, got %q", index, want, got)
				}
			}
		})
	}
}

// The three copies of one member expression must point at three DIFFERENT spreads, which the test
// above cannot see because all three attributes render the same text.
//
// The pairs are (1,2), (1,3) and (2,3), and each reports its left member, so the offsets are those
// of the first, first and second spreads. A rule reporting the right count from the wrong pairing,
// or reporting the same node three times, passes every other assertion in this file.
func TestJsxPropsNoSpreadMultiPairsByPosition(t *testing.T) {
	t.Parallel()

	sourceText := "const a = <App {...props.a} {...props.a} {...props.a} />;\n"
	result := rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile, sourceText)
	if len(result.Diagnostics) != 3 {
		t.Fatalf("want 3 findings, got %d", len(result.Diagnostics))
	}

	first := strings.Index(sourceText, "{...props.a}")
	second := strings.Index(sourceText[first+1:], "{...props.a}") + first + 1

	wantPositions := []int{first, first, second}
	for index, want := range wantPositions {
		if got := result.Diagnostics[index].Range.Pos(); got != want {
			t.Fatalf("finding %d: want position %d, got %d", index, want, got)
		}
	}
}

// The message text, asserted exactly rather than by containment.
//
// A fixture whose predicate is weaker than the property it guards is not a guard, and the two
// messages here are near-identical prose differing only in their opening noun, so a containment
// check would pass against either one. Equality is what separates them.
func TestJsxPropsNoSpreadMultiRendersBothMessages(t *testing.T) {
	t.Parallel()

	identifierResult := rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile,
		"const a = <App {...props} {...props} />;\n")
	if len(identifierResult.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(identifierResult.Diagnostics))
	}
	if got := identifierResult.Diagnostics[0].Message.Description; got != messageJsxPropsNoSpreadMultiIdentifier.Description {
		t.Fatalf("want the identifier message, got %q", got)
	}

	memberResult := rule_testing.Run(t, JsxPropsNoSpreadMulti, jsxPropsNoSpreadMultiFile,
		"const a = <App {...props.foo} {...props.foo} />;\n")
	if len(memberResult.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(memberResult.Diagnostics))
	}
	if got := memberResult.Diagnostics[0].Message.Description; got != messageJsxPropsNoSpreadMultiMemberExpression.Description {
		t.Fatalf("want the member message, got %q", got)
	}
	if messageJsxPropsNoSpreadMultiIdentifier.Description == messageJsxPropsNoSpreadMultiMemberExpression.Description {
		t.Fatal("the two messages must differ, or the assertions above cannot separate them")
	}
}
