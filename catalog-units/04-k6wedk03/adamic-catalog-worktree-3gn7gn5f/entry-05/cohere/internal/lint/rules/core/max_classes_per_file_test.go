package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// maxClassesPerFileFile is where the fixtures pretend to live.
const maxClassesPerFileFile = "/repository/source/MaxClassesPerFile.ts"

// decodedMaxClassesPerFileOptions routes a fixture's options through the rule's own exported
// decoder rather than building the options struct directly.
//
// The decoder is where this rule's two hazards live: the option is polymorphic, so an integer and
// an object must both reach the same settings, and the default is 1 rather than the zero value, so
// an absent option read as a plain int would be a maximum of ZERO and report every file holding one
// class. A fixture handing the rule a struct would exercise neither. An empty string here is the
// bare "error" the live config writes.
func decodedMaxClassesPerFileOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeMaxClassesPerFileOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/max-classes-per-file.js`, copied rather than
// rewritten: 9 of its 10 clean cases and all 9 reporting ones.
//
// The tenth clean case wraps two classes in an `eslint-disable` comment, and it is omitted
// deliberately rather than dropped. It is clean upstream because of that comment, which is the
// suppression layer's decision rather than the rule's, and `rule_testing.Run` never consults
// suppressions. Reproducing it as clean would mean breaking the rule; reporting it is what the rule
// actually does, and the same source without the disable line IS covered, as the comment-wrapped
// reporting case below.
//
// Every case string was built from a list rather than typed, so no escape could be cooked on the
// way in, and every verdict was reproduced by driving the installed eslint at 10.8.1 first.
func TestMaxClassesPerFileFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		wantCount  int
		wantMax    int
	}{
		{"two declarations", "class Foo {}\nclass Bar {}", "", 2, 1},
		{"a declaration and an expression", "class Foo {}\nconst myExpression = class {}", "", 2, 1},
		{"two expressions", "var x = class {};\nvar y = class {};", "", 2, 1},
		{"a declaration then an expression", "class Foo {}\nvar x = class {};", "", 2, 1},
		{"two declarations on one line at a maximum of one", "class Foo {} class Bar {}", "1", 2, 1},
		{"three declarations on one line at a maximum of two", "class Foo {} class Bar {} class Baz {}", "2", 3, 2},
		{"two declarations with an expression ignored at a maximum of one", "\n                class Foo {}\n                class Bar {}\n                const myExpression = class {}\n            ", "{\"ignoreExpressions\": true, \"max\": 1}", 2, 1},
		{"three declarations with an expression ignored at a maximum of two", "\n                class Foo {}\n                class Bar {}\n                class Baz {}\n                const myExpression = class {}\n            ", "{\"ignoreExpressions\": true, \"max\": 2}", 3, 2},
		{"comments wrapping two declarations", "/* comment */\nclass A {}\nclass B {}\n/* comment */", "", 2, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
				testCase.sourceText, decodedMaxClassesPerFileOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, "maximumExceeded")

			// The rendered text, not only the id. Both numbers move per finding, and passing them
			// in the wrong order renders a grammatical sentence that no id assertion can see.
			want := maxClassesPerFileMessage(testCase.wantCount, testCase.wantMax).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message text:\n got %q\nwant %q", got, want)
			}
			if !strings.HasPrefix(want, "File has too many classes (") {
				t.Fatalf("the message no longer opens with upstream's own wording: %q", want)
			}
		})
	}
}

func TestMaxClassesPerFileStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
	}{
		{"a single declaration with no options", "class Foo {}", ""},
		{"a single expression with no options", "var x = class {};", ""},
		{"no classes at all", "var x = 5;", ""},
		{"a single declaration at a maximum of one", "class Foo {}", "1"},
		{"two declarations at a maximum of two", "class Foo {}\nclass Bar {}", "2"},
		{"a single declaration under the object shape", "class Foo {}", "{\"max\": 1}"},
		{"two declarations under the object shape", "class Foo {}\nclass Bar {}", "{\"max\": 2}"},
		{"an expression ignored at a maximum of one", "\n                class Foo {}\n                const myExpression = class {}\n            ", "{\"ignoreExpressions\": true, \"max\": 1}"},
		{"an expression ignored at a maximum of two", "\n                class Foo {}\n                class Bar {}\n                const myExpression = class {}\n            ", "{\"ignoreExpressions\": true, \"max\": 2}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
				testCase.sourceText, decodedMaxClassesPerFileOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestMaxClassesPerFileSpansTheProgramBody pins where the finding points.
//
// Upstream reports on the program while overriding the location to the program's BODY: the start of
// the first statement to the end of the last. The difference from the program node itself is
// trivia, and it is load-bearing rather than cosmetic, which is why upstream's own corpus carries a
// comment-wrapped case asserting line 2 column 1 for a file whose first line is a comment.
//
// The three shapes here are the three upstream asserts, and each offset pair was converted to a
// line and column and checked against upstream's numbers before being written down.
func TestMaxClassesPerFileSpansTheProgramBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		wantStart  int
		wantEnd    int
	}{
		// upstream asserts line 1 column 1 to line 2 column 13
		{"a plain two-class file", "class Foo {}\nclass Bar {}", "", 0, 25},
		// upstream asserts line 2 column 1 to line 3 column 11
		{"a file wrapped in comments", "/* comment */\nclass A {}\nclass B {}\n/* comment */", "", 14, 35},
		// upstream asserts line 2 column 17 to line 4 column 46
		{"an indented file inside a template", "\n                class Foo {}\n                class Bar {}\n                const myExpression = class {}\n            ", "{\"ignoreExpressions\": true, \"max\": 1}", 17, 104},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
				testCase.sourceText, decodedMaxClassesPerFileOptions(t, testCase.options))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0].Range
			if reported.Pos() != testCase.wantStart || reported.End() != testCase.wantEnd {
				t.Fatalf("expected the span [%d,%d), got [%d,%d) which is %q",
					testCase.wantStart, testCase.wantEnd, reported.Pos(), reported.End(),
					testCase.sourceText[reported.Pos():reported.End()])
			}
		})
	}
}

// TestMaxClassesPerFileCountsEveryClassInTheFile covers shapes the corpus does not write and the
// rule does judge.
//
// Every case here was measured against the installed build at 10.8.1 before being written, and a
// port counting only top-level statements passes all 18 imported cases while going silent on every
// one of these. Classes declared inside a function are the ruled exception, in the test below.
func TestMaxClassesPerFileCountsEveryClassInTheFile(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
	}{
		{"two exported classes", "export class A {}\nexport class B {}"},
		{"a default export beside a declaration", "class A {}\nexport default class B {}"},
		{"two expressions passed straight to a call", "f(class {});\nf(class {});"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
				testCase.sourceText, decodedMaxClassesPerFileOptions(t, ""))
			rule_testing.ExpectFindings(t, result, "maximumExceeded")
		})
	}
}

// TestDecodeMaxClassesPerFileOptions is where this rule is hard.
//
// The option is a oneOf over an integer and an object, and its default is 1 rather than the zero
// value, so three separate decode paths can each silently invert the rule.
func TestDecodeMaxClassesPerFileOptions(t *testing.T) {
	t.Parallel()

	settingsFrom := func(t *testing.T, raw string) MaxClassesPerFileOptions {
		t.Helper()
		decoded, err := DecodeMaxClassesPerFileOptions(json.RawMessage(raw))
		if err != nil {
			t.Fatalf("the decoder refused %s: %v", raw, err)
		}
		options, isOptions := decoded.(MaxClassesPerFileOptions)
		if !isOptions {
			t.Fatalf("expected MaxClassesPerFileOptions, got %T", decoded)
		}
		return options
	}

	t.Run("empty input defaults to a maximum of one", func(t *testing.T) {
		t.Parallel()
		options := settingsFrom(t, "")
		if options.Maximum == nil || *options.Maximum != 1 {
			t.Fatalf("expected a maximum of 1, got %v", options.Maximum)
		}
		if options.IgnoreExpressions == nil || *options.IgnoreExpressions {
			t.Fatalf("expected expressions to be counted by default, got %v", options.IgnoreExpressions)
		}
	})

	t.Run("a bare integer decodes", func(t *testing.T) {
		t.Parallel()
		// The shape a config writes as ["error", 2], which no struct field can read.
		options := settingsFrom(t, "2")
		if options.Maximum == nil || *options.Maximum != 2 {
			t.Fatalf("expected a maximum of 2, got %v", options.Maximum)
		}
	})

	t.Run("an object with no max keeps the default of one", func(t *testing.T) {
		t.Parallel()
		// Upstream writes `option.max || 1`. Measured against the installed build: both {} and
		// {ignoreExpressions: true} leave the limit at one rather than at zero.
		for _, raw := range []string{"{}", `{"ignoreExpressions": true}`} {
			options := settingsFrom(t, raw)
			if options.Maximum == nil || *options.Maximum != 1 {
				t.Fatalf("%s: expected a maximum of 1, got %v", raw, options.Maximum)
			}
		}
	})

	t.Run("a maximum below one is refused", func(t *testing.T) {
		t.Parallel()
		// Upstream's schema says minimum 1, and eslint rejects both spellings at configuration load
		// rather than at lint time. Accepting them here would report every file holding a single
		// class.
		for _, raw := range []string{"0", `{"max": 0}`, "-1"} {
			if _, err := DecodeMaxClassesPerFileOptions(json.RawMessage(raw)); err == nil {
				t.Fatalf("%s should have been refused", raw)
			}
		}
	})

	t.Run("nil options reach the rule as a maximum of one", func(t *testing.T) {
		t.Parallel()
		// Past the decoder rather than through it: what a bare "error" produces after the config
		// layer has turned the decoder's error into nil. A zero read here reports the single-class
		// file below, so this case distinguishes the fallback from its absence.
		reporting := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
			"class Foo {}\nclass Bar {}", nil)
		rule_testing.ExpectFindings(t, reporting, "maximumExceeded")

		clean := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
			"class Foo {}", nil)
		rule_testing.ExpectClean(t, clean)
	})
}

// TestMaxClassesPerFileDoesNotCountAClassDeclaredInsideAFunction pins Kirk's ruling on #mbbg6js.
//
// A class declared inside any function body, a method's included, is not a design unit of the file,
// so it does not count; ESLint's core rule counts it, and Nexus's nexus/max-classes-per-file twin runs
// these rows verbatim to draw the same line. The first two cases are the ones upstream reports, measured
// at 10.8.1. The controls show the exclusion is about functions and nothing wider: a namespace is
// not a function, a bare block is not either, and an expression still answers to ignoreExpressions.
func TestMaxClassesPerFileDoesNotCountAClassDeclaredInsideAFunction(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		reports    bool
	}{
		{"a class nested in another class's method", "class A { m() { class B {} } }", false},
		{"a class inside a function", "function f(){ class A {} }\nclass B {}", false},
		{"classes inside a test callback, as the nexus type tests write them", "describe('x', () => {\n    class A {}\n    class B {}\n    it('y', () => { class C {} });\n});", false},
		{"a class inside an arrow beside a top-level one", "class A {}\nconst f = () => { class B {} };", false},
		{"a class in a namespace still counts", "class A {}\nnamespace N { class B {} }", true},
		{"a class in a bare block still counts", "class A {}\n{ class B {} }", true},
		{"a class in a static block still counts", "class A { static { class B {} } }", true},
		{"an expression inside a function still counts without ignoreExpressions", "class A {}\nfunction f() { return class {}; }", true},
		{"two top-level classes beside a nested one still report", "class A {}\nclass B {}\nfunction f() { class C {} }", true},
		// After a function the walk is at the top level again, so later classes count.
		{"two top-level classes after a function still report", "function f() { class C {} }\nclass A {}\nclass B {}", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MaxClassesPerFile, maxClassesPerFileFile,
				testCase.sourceText, decodedMaxClassesPerFileOptions(t, ""))
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "maximumExceeded")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
