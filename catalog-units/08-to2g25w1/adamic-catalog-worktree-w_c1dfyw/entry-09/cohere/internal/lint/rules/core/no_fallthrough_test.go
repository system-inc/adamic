package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// fallthroughFile is where the fixtures pretend to live.
//
// A `.tsx` name rather than `.ts` because upstream's snapshot header reads `no_fallthrough.tsx` and
// one of its pass cases is JSX inside a switch. Parsed as plain TypeScript, `<td key="foo">Foo</td>`
// is read as a type assertion and the case silently stops testing what it was added to test.
const fallthroughFile = "/repository/source/Fallthrough.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// From `oxc/crates/oxc_linter/src/rules/eslint/no_fallthrough.rs`: one Tester block, 56 pass and
// 30 fail inputs, and the snapshot records 32 diagnostics. That gap is real and recovered by hand
// below: exactly two inputs report twice, and every other fail input reports once.
func TestNoFallthroughFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"a bare fall into the next case", "switch(foo) { case 0: a();\ncase 1: b() }", nil, []string{"case"}},
		{"a bare fall into default", "switch(foo) { case 0: a();\ndefault: b() }", nil, []string{"default"}},
		{"a fall into default on one line", "switch(foo) { case 0: a(); default: b() }", nil, []string{"default"}},
		// An if with no else leaves a path through, so the break in the then arm is not enough.
		{"an if whose break is conditional", "switch(foo) { case 0: if (a) { break; } default: b() }", nil, []string{"default"}},
		// A catch that swallows the throw restores the fallthrough path.
		{"a try whose catch swallows the throw", "switch(foo) { case 0: try { throw 0; } catch (err) {} default: b() }", nil, []string{"default"}},
		// A break inside a loop breaks the loop, not the switch.
		{"a break belonging to a while", "switch(foo) { case 0: while (a) { break; } default: b() }", nil, []string{"default"}},
		{"a break belonging to a do-while", "switch(foo) { case 0: do { break; } while (a); default: b() }", nil, []string{"default"}},
		// An empty clause separated by a blank line reads as a forgotten body rather than a
		// deliberate grouped case, which is the whole point of the blank-line rule.
		{"an empty clause with a blank line before default", "switch(foo) { case 0:\n\n default: b() }", nil, []string{"default"}},
		{"an empty block body", "switch(foo) { case 0: {} default: b() }", nil, []string{"default"}},
		// The comment is in a block that is not the clause's only statement, so it is not in the
		// window a fallthrough comment may occupy.
		{"a fallthrough comment in a trailing block after a call", "switch(foo) { case 0: a(); { /* falls through */ } default: b() }", nil, []string{"default"}},
		{"a fallthrough comment in a leading block", "switch(foo) { case 0: { /* falls through */ } a(); default: b() }", nil, []string{"default"}},
		{"a fallthrough comment inside an if", "switch(foo) { case 0: if (a) { /* falls through */ } default: b() }", nil, []string{"default"}},
		{"a fallthrough comment nested two blocks deep", "switch(foo) { case 0: { { /* falls through */ } } default: b() }", nil, []string{"default"}},
		{"a block holding an unrelated comment", "switch(foo) { case 0: { /* comment */ } default: b() }", nil, []string{"default"}},
		{"an unrelated line comment before default", "switch(foo) { case 0:\n // comment\n default: b() }", nil, []string{"default"}},
		// `falling through` is not one of the four accepted spellings, and oxc matches the set
		// rather than ESLint's `/falls?\s?through/iu` regular expression. Reproduced, not improved.
		{"a comment reading falling through", "switch(foo) { case 0: a(); /* falling through */ default: b() }", nil, []string{"default"}},
		{
			"a custom pattern the comment does not match",
			"switch(foo) { case 0: a();\n/* no break */\ncase 1: b(); }",
			NoFallthroughOptions{CommentPattern: "break omitted"},
			[]string{"case"},
		},
		// The last comment before the next clause is the one that decides, so a matching comment
		// followed by a non-matching one does not save the clause.
		{
			"a matching comment followed by another comment",
			"switch(foo) { case 0: a();\n/* no break */\n/* todo: fix readability */\ndefault: b() }",
			NoFallthroughOptions{CommentPattern: "no break"},
			[]string{"default"},
		},
		{
			"a matching comment followed by another comment inside a block",
			"switch(foo) { case 0: { a();\n/* no break */\n/* todo: fix readability */ }\ndefault: b() }",
			NoFallthroughOptions{CommentPattern: "no break"},
			[]string{"default"},
		},
		{"an empty clause with comments and a blank line", "switch(foo) { case 0: \n /* with comments */  \ncase 1: b(); }", nil, []string{"case"}},
		{
			"an empty clause with a blank line, allowEmptyCase off explicitly",
			"switch(foo) { case 0:\n\ncase 1: b(); }",
			NoFallthroughOptions{},
			[]string{"case"},
		},
		{"an empty clause with a blank line, an empty options object", "switch(foo) { case 0:\n\ncase 1: b(); }", NoFallthroughOptions{}, []string{"case"}},
		{
			"an empty statement body across lines",
			"switch (a) { case 1: \n ; case 2:  }",
			NoFallthroughOptions{},
			[]string{"case"},
		},
		// The one input in this corpus that reports twice on its own. `allowEmptyCase` exempts a
		// clause with *no* statements; a lone `;` is an EmptyStatement, which is a statement, so
		// both of the first two clauses fall through and the option does not reach them.
		{
			"two empty-statement bodies with allowEmptyCase on",
			"switch (a) { case 1: ; case 2: ; case 3: }",
			NoFallthroughOptions{AllowEmptyCase: true},
			[]string{"case", "case"},
		},
		// `eslint-enable` is a directive, and a directive never permits a fallthrough. The pair with
		// the `eslint-disable-next-line` pass case is what makes this a real discrimination rather
		// than a spelling: both are directives here, and neither matches the fallthrough set, so
		// this one reports for the ordinary reason.
		{
			"a directive comment that is not a fallthrough comment",
			"switch (foo) { case 0: a(); \n// eslint-enable no-fallthrough\n case 1: }",
			NoFallthroughOptions{},
			[]string{"case"},
		},
		{
			"a fallthrough comment on a clause that cannot fall through",
			"switch(foo) { case 0: a(); break; /* falls through */ case 1: b(); }",
			NoFallthroughOptions{ReportUnusedFallthroughComment: true},
			[]string{"unusedFallthroughComment"},
		},
		// A default clause falls into a following case exactly like a case does.
		{"a default falling into a case", "switch(foo) { default: a(); case 1: b(); }", nil, []string{"case"}},
		// The second input reporting twice: a default in the middle falls in from the case before it
		// and out into the case after it.
		{"a default in the middle falling both directions", "switch(foo) { case 0: a(); default: b(); case 1: c(); }", nil, []string{"default", "case"}},
		// A logical operator in the discriminant test is still an ordinary case clause. Upstream
		// carried a bug here (its issue 6417) where the extra condition instructions confused the
		// control flow graph; a structural analysis has nothing to be confused by.
		{"a case whose test uses logical or", "switch(true) { case x === 1 || x === 2: a(); case x === 3: b(); }", nil, []string{"case"}},
		{"a case whose test uses logical and", "switch(true) { case x === 1 && y: a(); case x === 3: b(); }", nil, []string{"case"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.options == nil {
				result = rule_testing.Run(t, NoFallthrough, fallthroughFile, testCase.sourceText)
			} else {
				result = rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
					testCase.sourceText, testCase.options)
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The clean cases are the whole discrimination, and each was added upstream when somebody hit it.
func TestNoFallthroughStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"a falls-through comment", "switch(foo) { case 0: a(); /* falls through */ case 1: b(); }", nil},
		{"a falls-through comment on its own line", "switch(foo) { case 0: a()\n /* falls through */ case 1: b(); }", nil},
		{"a fall-through comment", "switch(foo) { case 0: a(); /* fall through */ case 1: b(); }", nil},
		{"a fallthrough comment", "switch(foo) { case 0: a(); /* fallthrough */ case 1: b(); }", nil},
		{"an upper-case falls-through comment", "switch(foo) { case 0: a(); /* FALLS THROUGH */ case 1: b(); }", nil},
		// The block form: a clause whose only statement is a block may carry the comment inside it,
		// before the closing brace, which is a second window the plain scan never looks at.
		{"a comment inside the clause's only block", "switch(foo) { case 0: { a(); /* falls through */ } case 1: b(); }", nil},
		{"a comment inside a block on its own line", "switch(foo) { case 0: { a()\n /* falls through */ } case 1: b(); }", nil},
		{"a fall-through comment inside a block", "switch(foo) { case 0: { a(); /* fall through */ } case 1: b(); }", nil},
		{"a fallthrough comment inside a block", "switch(foo) { case 0: { a(); /* fallthrough */ } case 1: b(); }", nil},
		{"an upper-case comment inside a block", "switch(foo) { case 0: { a(); /* FALLS THROUGH */ } case 1: b(); }", nil},
		{"a comment after the block", "switch(foo) { case 0: { a(); } /* falls through */ case 1: b(); }", nil},
		// The in-block window wins even when a later comment does not match, which is the pair with
		// the fail case where the same shape has the non-matching comment *inside* the block.
		{"a matching in-block comment then an unrelated one", "switch(foo) { case 0: { a(); /* falls through */ } /* comment */ case 1: b(); }", nil},
		// The comment is the block's only content, so there is no statement to anchor a scan on.
		// This is the case that proved the shared comment scan cannot see it.
		{"a comment alone in the clause's only block", "switch(foo) { case 0: { /* falls through */ } case 1: b(); }", nil},
		{"a return", "function foo() { switch(foo) { case 0: a(); return; case 1: b(); }; }", nil},
		{"a throw", "switch(foo) { case 0: a(); throw 'foo'; case 1: b(); }", nil},
		// `continue` targets the enclosing loop and so leaves the switch.
		{"a continue inside a loop", "while (a) { switch(foo) { case 0: a(); continue; case 1: b(); } }", nil},
		{"a break", "switch(foo) { case 0: a(); break; case 1: b(); }", nil},
		{"a grouped empty clause", "switch(foo) { case 0: case 1: a(); break; case 2: b(); }", nil},
		{"grouped empty clauses before a break", "switch(foo) { case 0: case 1: break; case 2: b(); }", nil},
		{"grouped empty clauses before default", "switch(foo) { case 0: case 1: break; default: b(); }", nil},
		// Nothing follows the last clause, so it cannot fall through into anything.
		{"a last clause with a body", "switch(foo) { case 0: case 1: a(); }", nil},
		{"a last clause with a break", "switch(foo) { case 0: case 1: a(); break; }", nil},
		{"a last clause that is just a break", "switch(foo) { case 0: case 1: break; }", nil},
		// One newline is formatting; two are a blank line.
		{"a grouped empty clause across one line", "switch(foo) { case 0:\n case 1: break; }", nil},
		{"a grouped empty clause with a comment", "switch(foo) { case 0: // comment\n case 1: break; }", nil},
		{"grouped empty clauses before a return", "function foo() { switch(foo) { case 0: case 1: return; } }", nil},
		{"blocks that each return", "function foo() { switch(foo) { case 0: {return;}\n case 1: {return;} } }", nil},
		{"a last clause whose block breaks", "switch(foo) { case 0: case 1: {break;} }", nil},
		{"a switch with no clauses at all", "switch(foo) { }", nil},
		// The inner switch's break belongs to the inner switch, so the outer clause does fall
		// through, and the comment is what makes this clean.
		{"a nested switch with a fallthrough comment", "switch(foo) { case 0: switch(bar) { case 2: break; } /* falls through */ case 1: break; }", nil},
		// Statements after a return are unreachable, so the clause still exits.
		{"code after a return", "function foo() { switch(foo) { case 1: return a; a++; }}", nil},
		{"a fallthrough comment before default with a trailing comment", "switch (foo) { case 0: a(); /* falls through */ default:  b(); /* comment */ }", nil},
		{"a fallthrough comment before a default whose body starts with a comment", "switch (foo) { case 0: a(); /* falls through */ default: /* comment */ b(); }", nil},
		// Both arms of the if leave the clause, so there is no path through.
		{"an if whose both arms exit", "switch (foo) { case 0: if (a) { break; } else { throw 0; } default: b(); }", nil},
		{"a try that breaks with an empty finally", "switch (foo) { case 0: try { break; } finally {} default: b(); }", nil},
		{"a finally that breaks", "switch (foo) { case 0: try {} finally { break; } default: b(); }", nil},
		{"a catch that breaks after a throw", "switch (foo) { case 0: try { throw 0; } catch (err) { break; } default: b(); }", nil},
		// A do-while runs its body at least once, so a throw inside it always happens.
		{"a do-while whose body throws", "switch (foo) { case 0: do { throw 0; } while(a); default: b(); }", nil},
		// A default in the middle that breaks does not fall out of itself.
		{"a default in the middle that breaks", "switch(foo) { case 0: default: a(); break; case 1: b(); }", nil},
		{
			"a custom pattern matching the comment",
			"switch(foo) { case 0: a(); /* no break */ case 1: b(); }",
			NoFallthroughOptions{CommentPattern: "no break"},
		},
		{
			"a custom pattern with a character class",
			"switch(foo) { case 0: a(); /* no break: need to execute b() */ case 1: b(); }",
			NoFallthroughOptions{CommentPattern: `no break:\s?\w+`},
		},
		{
			"a custom pattern matching the last of two line comments",
			"switch(foo) { case 0: a();\n// need to execute b(), so\n// falling through\n case 1: b(); }",
			NoFallthroughOptions{CommentPattern: "falling through"},
		},
		{
			"a custom pattern before default",
			"switch(foo) { case 0: a(); /* break omitted */ default:  b(); /* comment */ }",
			NoFallthroughOptions{CommentPattern: "break omitted"},
		},
		{
			"a custom pattern matching two different comments",
			"switch(foo) { case 0: a(); /* caution: break is omitted intentionally */ case 1: b(); /* break omitted */ default: c(); }",
			NoFallthroughOptions{CommentPattern: `break[\s\w]+omitted`},
		},
		{
			"an empty clause with blank lines and allowEmptyCase on",
			"switch(foo) { case 0: \n\n\n case 1: b(); }",
			NoFallthroughOptions{AllowEmptyCase: true},
		},
		{
			"an empty clause with comments and allowEmptyCase on",
			"switch(foo) { case 0: \n /* with comments */  \n case 1: b(); }",
			NoFallthroughOptions{AllowEmptyCase: true},
		},
		{
			"an empty-statement clause that breaks, allowEmptyCase on",
			"switch (a) {\n case 1: ; break; \n case 3: }",
			NoFallthroughOptions{AllowEmptyCase: true},
		},
		{
			"an empty-statement clause that breaks, allowEmptyCase off",
			"switch (a) {\n case 1: ; break; \n case 3: }",
			NoFallthroughOptions{},
		},
		{
			"an unused fallthrough comment with reporting off",
			"switch(foo) { case 0: a(); break; /* falls through */ case 1: b(); }",
			NoFallthroughOptions{ReportUnusedFallthroughComment: false},
		},
		// Upstream issue 11340: a default in the middle whose body returns.
		{"a default in the middle that returns", "function f(name) { switch(name) { case 'a': case 'b': default: return 'x'; case 'c': return 'y'; } }", nil},
		{"a default first that breaks", "switch(foo) { default: a(); break; case 1: b(); }", nil},
		// Upstream issue 6417: a logical operator in the test with a break present.
		{"a logical-or test with a break", "switch(true) { case x === 1 || x === 2: a(); break; case x === 3: b(); }", nil},
		{"a logical-and test with a break", "switch(true) { case x === 1 && y: a(); break; case x === 3: b(); }", nil},
		{
			"clauses whose blocks each return, in JSX",
			`c.map(c => { switch (true) {
        case c.f === 'qux' && xCount > 1: { return <td key="foo">Foo</td>; }
        case c.f === 'barbaz' && isFoo: { return <td key="bar">Foobar</td>; }
        case c.f === 'baz': { return arrayOfRecords.map(r => <td key={r.id}>{r.id}</td>); }
        default: { return <td>Bar</td>; }
      } });`,
			nil,
		},
		// Upstream issue 21320: a clause whose test spans lines. The blank-line test measures the
		// gap between clauses, and a multi-line *test* is not a gap between clauses.
		{"a multi-line case test", "switch(foo) {\n  case A\n    .B:\n  case B.A:\n    break;\n}", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.options == nil {
				result = rule_testing.Run(t, NoFallthrough, fallthroughFile, testCase.sourceText)
			} else {
				result = rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
					testCase.sourceText, testCase.options)
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Where the finding points, which the message-id fixtures above cannot see.
//
// The rule carries no repair, and a rule with no repair is exactly the shape that ships pointing at
// the wrong node with a fully green suite. The span is upstream's `next_case.span`: the clause that
// is fallen *into*, from its `case` or `default` keyword through the end of its body, and not the
// clause that failed to break. Read off the snapshot's underline rather than guessed.
func TestNoFallthroughPointsAtTheClauseFallenInto(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantTexts  []string
	}{
		{"a case on the next line", "switch(foo) { case 0: a();\ncase 1: b() }", nil, []string{"case 1: b()"}},
		{"a default on one line", "switch(foo) { case 0: a(); default: b() }", nil, []string{"default: b()"}},
		{"an empty block body", "switch(foo) { case 0: {} default: b() }", nil, []string{"default: b()"}},
		// The blank line is trivia before the clause, so the finding must not start at it. This is
		// the case that would catch a range built from Pos() rather than the first token.
		{"an empty clause with a blank line", "switch(foo) { case 0:\n\n default: b() }", nil, []string{"default: b()"}},
		{"an unrelated line comment before default", "switch(foo) { case 0:\n // comment\n default: b() }", nil, []string{"default: b()"}},
		// Both findings, in source order, from the one input that reports in two directions.
		{
			"a default in the middle",
			"switch(foo) { case 0: a(); default: b(); case 1: c(); }",
			nil,
			[]string{"default: b();", "case 1: c();"},
		},
		{
			"two empty-statement bodies",
			"switch (a) { case 1: ; case 2: ; case 3: }",
			NoFallthroughOptions{AllowEmptyCase: true},
			[]string{"case 2: ;", "case 3:"},
		},
		// The unused-comment finding points at the comment instead, because the comment is the
		// thing to delete. A rule reusing the clause span for both messages passes every message-id
		// fixture above and is wrong here.
		{
			"an unused fallthrough comment",
			"switch(foo) { case 0: a(); break; /* falls through */ case 1: b(); }",
			NoFallthroughOptions{ReportUnusedFallthroughComment: true},
			[]string{"/* falls through */"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.options == nil {
				result = rule_testing.Run(t, NoFallthrough, fallthroughFile, testCase.sourceText)
			} else {
				result = rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
					testCase.sourceText, testCase.options)
			}
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantTexts), len(result.Diagnostics))
			}
			for index, wantText := range testCase.wantTexts {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantText {
					t.Fatalf("finding %d pointed at %q, wanted %q", index, reported, wantText)
				}
			}
		})
	}
}

// Cases written from reading our own code rather than upstream's, each covering a discrimination
// the imported corpus never exercises.
func TestNoFallthroughCasesFromOurOwnReading(t *testing.T) {
	t.Parallel()

	t.Run("a labeled break naming the switch leaves it", func(t *testing.T) {
		t.Parallel()
		// Upstream's corpus has no labeled break at all, and a port crediting only a bare `break`
		// would report this. A labeled break jumps out of whatever it names, and in every case that
		// is somewhere at or outside the switch, so the clause does not fall through either way.
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"outer: switch(foo) { case 0: a(); break outer; case 1: b(); }"))
	})

	t.Run("a labeled break naming an inner loop does not leave the switch", func(t *testing.T) {
		t.Parallel()
		// The other half of the pair. `break inner` leaves the loop and lands back in the clause,
		// so the clause still falls through. A port treating every labeled break as an exit stays
		// silent here and nothing upstream would notice.
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: inner: while (a) { break inner; } case 1: b(); }"), "case")
	})

	t.Run("a continue with a label leaves the switch", func(t *testing.T) {
		t.Parallel()
		// `continue` can only target a loop, and no loop is inside this clause, so it necessarily
		// leaves the switch. Upstream covers unlabeled continue only.
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"outer: while (a) { switch(foo) { case 0: a(); continue outer; case 1: b(); } }"))
	})

	t.Run("a switch nested in a clause does not lend its break to the outer clause", func(t *testing.T) {
		t.Parallel()
		// The pair with upstream's nested-switch pass case, which is clean only because of its
		// comment. Without the comment the outer clause falls through, and this pins that the inner
		// break is not credited outward.
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: switch(bar) { case 2: break; } case 1: break; }"), "case")
	})

	t.Run("a return inside a nested function does not exit the clause", func(t *testing.T) {
		t.Parallel()
		// The walk has to stop at a function boundary. A return in an inner function returns from
		// that function, and the clause carries straight on into the next one.
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"function f() { switch(foo) { case 0: (function () { return 1; }); case 1: b(); } }"), "case")
	})

	t.Run("a comment pattern is matched case-insensitively", func(t *testing.T) {
		t.Parallel()
		// oxc compiles the configured pattern with `(?iu)` prepended, so a lower-case pattern
		// matches an upper-case comment. Nothing in the corpus pairs a custom pattern with a
		// differing case, so a port compiling the pattern verbatim passes all 56 pass cases.
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: a(); /* NO BREAK */ case 1: b(); }",
			NoFallthroughOptions{CommentPattern: "no break"}))
	})

	t.Run("an invalid comment pattern does not match everything", func(t *testing.T) {
		t.Parallel()
		// A pattern that fails to compile must not silently become "matches anything", which is
		// what a port ignoring the compile error would produce: the rule would go quiet on every
		// file the option touches. Falling back to the default set keeps the rule honest.
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: a(); /* whatever */ case 1: b(); }",
			NoFallthroughOptions{CommentPattern: "("}), "case")
	})

	t.Run("a directive is never a fallthrough comment, even under a matching pattern", func(t *testing.T) {
		t.Parallel()
		// oxc rejects a comment opening with `oxlint-` or `eslint-` before consulting the pattern,
		// and the guard is only load-bearing when a pattern would otherwise match a directive. A
		// mutation deleting it survived the whole corpus and the first version of this case, both
		// of which used the default comment set: under the default set a directive fails the four
		// exact spellings anyway, so the guard never decided anything. These two configure a
		// pattern broad enough to match the directive text, which is the only shape where deleting
		// the guard changes an answer, and it is a shape a real configuration can reach.
		for _, sourceText := range []string{
			"switch(foo) { case 0: a();\n// oxlint-disable no-fallthrough\ncase 1: b(); }",
			"switch(foo) { case 0: a();\n// eslint-disable no-fallthrough\ncase 1: b(); }",
		} {
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoFallthrough, fallthroughFile,
				sourceText, NoFallthroughOptions{CommentPattern: "no-fallthrough"}), "case")
		}
	})

	t.Run("a clause whose only statement is a labeled block carries its comment", func(t *testing.T) {
		t.Parallel()
		// The in-block comment window is gated on the clause's only statement being a block. A
		// labeled block is a LabeledStatement wrapping one, so the window does not apply and the
		// comment has to be found by the ordinary between-clauses scan, which it is.
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: lbl: { a(); } /* falls through */ case 1: b(); }"))
	})

	t.Run("a suppression directive is not a fallthrough comment", func(t *testing.T) {
		t.Parallel()
		// Upstream lists this input as a *pass* case, and it is a pass case there because oxc's
		// Tester runs the whole engine and the engine consumes the directive. Our
		// `rule_testing.Run` walks one rule and applies no suppressions, so the finding is expected
		// here and `cohere` swallows it in a real run. The rule's own job is only to refuse to read
		// `eslint-disable-next-line no-fallthrough` as permission to fall through, which it does:
		// were the directive guard removed, a custom pattern could turn any suppression into an
		// excuse. Moved out of the imported clean set with the divergence stated rather than
		// deleted, so the next reader does not re-derive it.
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"switch (foo) { case 0: a(); \n// eslint-disable-next-line no-fallthrough\n case 1: }"), "case")
	})

	t.Run("declaring the rule needs no checker", func(t *testing.T) {
		t.Parallel()
		// The rule reads the statement tree only, so the untyped harness is enough. If a later
		// change reaches for the checker without declaring it, the rule goes silent under
		// `rule_testing.Run` and this case fails loudly rather than passing vacuously.
		if NoFallthrough.NeedsTypeChecker {
			t.Fatal("this rule answers structurally and must not declare a type checker")
		}
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoFallthrough, fallthroughFile,
			"switch(foo) { case 0: a();\ncase 1: b() }"), "case")
	})
}
