package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoCommentTextnodesFile is where the fixtures pretend to live.
//
// A .tsx extension because every case holds JSX. The rule has no suffix gate, which the suffix
// cases below pin by writing the same reporting source to three extensions.
const jsxNoCommentTextnodesFile = "/repository/source/JsxNoCommentTextnodes.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-no-comment-textnodes.js, read by
// evaluating the two arrays in the tester and emitting Go raw strings from the decoded values, so
// no escape sequence was typed on the way here. Upstream carries 17 valid and 7 invalid cases, each
// invalid one naming exactly one messageId, which is 7 findings from 7 inputs.
//
// Upstream's feature markers are handled rather than dropped. "fragment" carries no meaning for us,
// since fragments are ordinary syntax to this parser. "no-ts-old" excludes a legacy parser we do
// not have. "no-ts" appears on one valid case, `<></* valid *//>`, which upstream reports both of
// its TypeScript parsers as failing to parse; ours parses it and produces no JSX text node at all,
// so it is imported as clean, which is upstream's own verdict for it. Probed alongside two sibling
// fragment forms, all three of which produce zero text nodes here.

// TestJsxNoCommentTextnodesFires runs the seven failing cases from upstream, one finding each.
func TestJsxNoCommentTextnodesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"invalid 0 a line comment as the only child, inline", `
        class Comp1 extends Component {
          render() {
            return (<div>// invalid</div>);
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 1 a line comment as the only child of a fragment", `
        class Comp1 extends Component {
          render() {
            return (<>// invalid</>);
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 2 a block comment as the only child, inline", `
        class Comp1 extends Component {
          render() {
            return (<div>/* invalid */</div>);
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 3 a line comment on its own line", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                // invalid
              </div>
            );
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 4 a block comment between two text lines", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                asdjfl
                /* invalid */
                foo
              </div>
            );
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 5 a comment line between two expression containers", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {'asdjfl'}
                // invalid
                {'foo'}
              </div>
            );
          }
        }
      `, []string{"putCommentInBraces"}},
		{"invalid 6 a bare slash-star as the only child", `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `, []string{"putCommentInBraces"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoCommentTextnodesStaysSilent runs the seventeen passing cases from upstream.
//
// These are the false positives upstream already thought about, and three of them are the reason
// the rule reads the raw source rather than a decoded value: the entity-escaped cases hold the
// ampersand form in the source and the slash form only after decoding. A port reading a cooked
// value would report all three and no fixture written from the corpus could see the difference,
// because the corpus asserts them as clean either way.
func TestJsxNoCommentTextnodesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0 a braced comment inside a div", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {/* valid */}
              </div>
            );
          }
        }
      `},
		{"valid 1 a braced comment inside a fragment", `
        class Comp1 extends Component {
          render() {
            return (
              <>
                {/* valid */}
              </>
            );
          }
        }
      `},
		{"valid 2 a braced comment on one line", `
        class Comp1 extends Component {
          render() {
            return (<div>{/* valid */}</div>);
          }
        }
      `},
		{"valid 3 a braced comment assigned to a local", `
        class Comp1 extends Component {
          render() {
            const bar = (<div>{/* valid */}</div>);
            return bar;
          }
        }
      `},
		{"valid 4 a braced comment in a factory property", `
        var Hello = createReactClass({
          foo: (<div>{/* valid */}</div>),
          render() {
            return this.foo;
          },
        });
      `},
		{"valid 5 three braced comments", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {/* valid */}
                {/* valid 2 */}
                {/* valid 3 */}
              </div>
            );
          }
        }
      `},
		{"valid 6 an empty div", `
        class Comp1 extends Component {
          render() {
            return (
              <div>
              </div>
            );
          }
        }
      `},
		{"valid 7 a file with no jsx at all", `
        var foo = require('foo');
      `},
		{"valid 8 a braced comment inside a component element", `
        <Foo bar='test'>
          {/* valid */}
        </Foo>
      `},
		{"valid 9 a url after a non-breaking space", `
        <strong>
          &nbsp;https://www.example.com/attachment/download/1
        </strong>
      `},
		{"valid 10 a real comment between attributes", `
        <Foo /* valid */ placeholder={'foo'}/>
      `},
		{"valid 11 a comment inside a fragment opening tag", `
        </* valid */></>
      `},
		{"valid 12 a comment inside a self-closing fragment", `
        <></* valid *//>
      `},
		{"valid 13 a comment inside an expression container", `
        <Foo title={'foo' /* valid */}/>
      `},
		{"valid 14 entity-escaped line comment in a pre", `<pre>&#x2F;&#x2F; TODO: Write perfect code</pre>`},
		{"valid 15 entity-escaped block comment in a pre", `<pre>&#x2F;&#42; TODO: Write perfect code &#42;&#x2F;</pre>`},
		{"valid 16 entity-escaped slashes inside a span", `
        <div>
          <span className="pl-c"><span className="pl-c">&#47;&#47;</span> ...</span><br />
        </div>
      `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoCommentTextnodesHasNoFileSuffixGate pins the absence of the gate three siblings carry.
//
// Those siblings were ported from oxc, which gates the whole rule on the file being read as JSX.
// The authority here is eslint-plugin-react, which has no such gate. Measured by driving the
// installed build over this source under .tsx, .jsx and .js: all three report. There is no .ts row
// because a .ts file cannot hold JSX at all, which makes the gate unobservable there rather than
// wrong.
func TestJsxNoCommentTextnodesHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := "const a = <div>// invalid</div>;\n"
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.jsx",
		"/repository/source/Suffix.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, fileName, source)
			rule_testing.ExpectFindings(t, result, "putCommentInBraces")
		})
	}
}

// TestJsxNoCommentTextnodesLineAnchoring pins the `m` flag, which the corpus exercises only
// implicitly.
//
// Upstream's pattern is `/^\s*\/(\/|\*)/m`, and the `m` is what makes `^` match at every line start
// rather than only at the node's start. Four of the seven failing corpus cases depend on it, but
// they depend on it incidentally: each would also pass a port that scanned for `//` anywhere. This
// table separates the two readings, and every row was measured against the installed build.
func TestJsxNoCommentTextnodesLineAnchoring(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// The control. If this stops reporting the rest of the table measures nothing.
		{"a comment at the node start", "const a = <div>// invalid</div>;\n", []string{"putCommentInBraces"}},

		// The `m` flag: the comment is on the third line of one text node.
		{"a comment on a later line", "const a = <div>\n  hello\n  // invalid\n</div>;\n", []string{"putCommentInBraces"}},

		// `\s*` allows leading whitespace, so indentation does not hide it.
		{"an indented comment", "const a = <div>\n      // invalid\n</div>;\n", []string{"putCommentInBraces"}},

		// `^` is the whole reason this is clean: the line does not BEGIN with the slashes. A port
		// scanning for `//` anywhere in the text would report here, and every corpus case would
		// still pass.
		{"slashes after other text on the same line", "const a = <div>hello // not a comment</div>;\n", nil},

		// A single slash is not the start of either comment form.
		{"a lone slash", "const a = <span>/</span>;\n", nil},

		// Two slashes with nothing after them still open a line comment.
		{"a bare double slash", "const a = <span>//</span>;\n", []string{"putCommentInBraces"}},

		// A block comment opener with no closer, which is upstream's own last failing case in a
		// different wrapper.
		{"a bare block comment opener", "const a = <span>/*</span>;\n", []string{"putCommentInBraces"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoCommentTextnodesReportsOncePerTextNode pins the finding count, which the corpus cannot.
//
// The regex is a test rather than a scan, so a text node holding several comment lines reports
// once. Splitting the same two lines with an expression container makes them two text nodes and
// produces two findings. Every case in the corpus holds at most one comment in one node, so the
// distinction is invisible there, and a port scanning for every match would pass all seven while
// reporting twice on the first row here.
func TestJsxNoCommentTextnodesReportsOncePerTextNode(t *testing.T) {
	t.Parallel()

	t.Run("two comment lines in one text node report once", func(t *testing.T) {
		t.Parallel()
		source := "const a = <div>\n  // one\n  // two\n</div>;\n"
		result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, source)
		rule_testing.ExpectFindings(t, result, "putCommentInBraces")
	})

	t.Run("the same two lines split by a container report twice", func(t *testing.T) {
		t.Parallel()
		source := "const a = <div>\n  // one\n  {x}\n  // two\n</div>;\n"
		result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, source)
		rule_testing.ExpectFindings(t, result, "putCommentInBraces", "putCommentInBraces")
	})
}

// TestJsxNoCommentTextnodesIgnoresStringLiterals pins the listener this port deliberately omits.
//
// Upstream registers `Literal` alongside `JSXText` and then filters on the parent's type, requiring
// it to contain "JSX" while being neither a JSXAttribute nor a JSXExpressionContainer. Those two
// are a string literal's only JSX parents, so the filter excludes every literal it could see and
// the listener can never report. Probed in this parser: an attribute value's parent is
// KindJsxAttribute and a braced string's is KindJsxExpression, matching upstream's two exclusions
// exactly. Both rows measured silent against the installed build.
//
// The rule registers no literal listener at all as a result. That is a divergence in mechanism and
// none in decision, and these cases are what keeps it honest: if the omission were wrong, they
// would be the fixtures that said so.
func TestJsxNoCommentTextnodesIgnoresStringLiterals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a line comment inside an attribute string", `const a = <div className="// nope" />;`},
		{"a block comment inside an attribute string", `const a = <div className="/* nope */" />;`},
		{"a line comment inside a braced string", `const a = <div>{"// ok"}</div>;`},
		{"a line comment inside a plain string outside jsx", `const a = "// nope";`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoCommentTextnodesReadsRawSource pins that the predicate runs over the source slice.
//
// Upstream calls `getText(context, node)`, which is the raw source rather than any decoded value,
// and the comment above that call says why: one parser hands back a wrong `raw`. The consequence
// for us is the entity cases, which the corpus already ships as clean. This adds the mirror image,
// where the raw text really does hold the slashes, so the pair separates "reads raw" from "reports
// nothing on a pre element".
func TestJsxNoCommentTextnodesReadsRawSource(t *testing.T) {
	t.Parallel()

	t.Run("entity-escaped slashes are clean", func(t *testing.T) {
		t.Parallel()
		source := "const a = <pre>&#x2F;&#x2F; TODO</pre>;\n"
		result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, source)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("literal slashes in the same element report", func(t *testing.T) {
		t.Parallel()
		source := "const a = <pre>// TODO</pre>;\n"
		result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, source)
		rule_testing.ExpectFindings(t, result, "putCommentInBraces")
	})
}

// TestJsxNoCommentTextnodesSpans asserts where the finding points.
//
// Upstream passes the text node itself to `report`, so the span is the whole text node including
// its surrounding whitespace rather than the comment inside it. A message-id assertion cannot see
// that, and the second row is the one that shows the difference: the reported text is far wider
// than the comment. `rule_testing.Run` writes the source verbatim, unlike `RunTyped` which trims
// it, so the literal here is the file on disk and slicing it directly is correct.
func TestJsxNoCommentTextnodesSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{
			"an inline comment underlines exactly the text node",
			"const a = <div>// invalid</div>;\n",
			"// invalid",
		},
		{
			"a comment on its own line underlines the whole text node",
			"const a = <div>\n  hello\n  // invalid\n</div>;\n",
			"\n  hello\n  // invalid\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoCommentTextnodes, jsxNoCommentTextnodesFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantText {
				t.Errorf("span text = %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// TestJsxNoCommentTextnodesMessage asserts the message by identity.
//
// A rule.Message is {Id, Description} with no interpolation, so there is nothing to render.
// Asserting against a literal typed here rather than against the rule's own constant is deliberate:
// comparing a finding to the constant it was reported with is an equality that moves on both sides
// under mutation and cannot fail.
func TestJsxNoCommentTextnodesMessage(t *testing.T) {
	t.Parallel()

	if messageJsxNoCommentTextnodes.Id != "putCommentInBraces" {
		t.Errorf("message id = %q", messageJsxNoCommentTextnodes.Id)
	}
	if len(messageJsxNoCommentTextnodes.Description) < 80 {
		t.Errorf("description is too short to say why: %q", messageJsxNoCommentTextnodes.Description)
	}
}
