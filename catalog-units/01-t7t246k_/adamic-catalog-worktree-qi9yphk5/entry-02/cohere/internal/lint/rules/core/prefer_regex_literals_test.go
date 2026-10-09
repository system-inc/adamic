package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus below is ESLint's own, imported verbatim from
// tests/lib/rules/prefer-regex-literals.js at eslint 10.8.1: 70 clean cases and 181 reporting ones,
// carrying 192 findings and 163 suggestions between them.
//
// It was extracted mechanically rather than retyped. The extractor loaded the upstream test file as
// a module, captured the case objects, and then replayed all 251 of them through the installed
// eslint rule with the Linter API; every one reproduced its stated verdict, message ids, and
// suggestion count, so the table below is measured against the running rule rather than against a
// reading of it. Controlling the extractor by perturbing one expected count turned the run red,
// which is what makes the clean run mean something.
//
// `upstreamLanguageOptions` records a case's own languageOptions where it had any. Most of them do
// not change the verdict here; the eight that do are called out at the rows and in the rule's doc
// comment under the ecmaVersion divergence.

const preferRegexLiteralsFile = "prefer_regex_literals.ts"

// preferRegexLiteralsSuggestion is one offered repair: the message id, and the whole file as it
// would read if a human accepted it.
//
// Asserting the resulting SOURCE rather than the fix text is the point. A fix writing the right
// string over the wrong span passes a text comparison and is a real defect; only replaying the edit
// into the file can see it. `rule_testing.ExpectFixedSource` does this for fixes and has no
// suggestion equivalent, so the applier below is hand-rolled, which the brief says to budget for.
type preferRegexLiteralsSuggestion struct {
	id     string
	output string
}

type preferRegexLiteralsRow struct {
	source                  string
	ids                     []string
	suggestions             []preferRegexLiteralsSuggestion
	options                 string
	upstreamLanguageOptions string
}

// preferRegexLiteralsClean holds every case upstream expects to report nothing.
//
// These are the false positives upstream already thought about, and they are the half of the corpus
// that catches a port. Eleven of them are a local binding shadowing the global, which is the whole
// reason this rule reads the checker; another eleven are string concatenations, which upstream
// deliberately declines to constant-fold.
var preferRegexLiteralsClean = []preferRegexLiteralsRow{
	{source: "/abc/"},
	{source: "/abc/g"},
	{source: "new RegExp(pattern)"},
	{source: "new RegExp('\\\\p{Emoji_Presentation}\\\\P{Script_Extensions=Latin}' + '', `ug`)"},
	{source: "new RegExp('\\\\cA' + '')"},
	{source: "RegExp(pattern, 'g')"},
	{source: "new RegExp(f('a'))"},
	{source: "RegExp(prefix + 'a')"},
	{source: "new RegExp('a' + suffix)"},
	{source: "RegExp(`a` + suffix);"},
	{source: "new RegExp(String.raw`a` + suffix);"},
	{source: "RegExp('a', flags)"},
	{source: "const flags = 'gu';RegExp('a', flags)"},
	{source: "RegExp('a', 'g' + flags)"},
	{source: "new RegExp(String.raw`a`, flags);"},
	{source: "RegExp(`${prefix}abc`)"},
	{source: "new RegExp(`a${b}c`);"},
	{source: "new RegExp(`a${''}c`);"},
	{source: "new RegExp(String.raw`a${b}c`);"},
	{source: "new RegExp(String.raw`a${''}c`);"},
	{source: "new RegExp('a' + 'b')"},
	{source: "RegExp(1)"},
	{source: "new RegExp('(\\\\p{Emoji_Presentation})\\\\1' + '', `ug`)"},
	{source: "RegExp(String.raw`\\78\\126` + '\\\\5934', '' + `g` + '')"},
	{source: "func(new RegExp(String.raw`a${''}c\\d`, 'u'),new RegExp(String.raw`a${''}c\\d`, 'u'))"},
	{source: "new RegExp('\\\\[' + \"b\\\\]\")"},
	{source: "new RegExp(/a/, flags);", options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/, `u${flags}`);", options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/);", options: "{}"},
	{source: "new RegExp(/a/);", options: "{\"disallowRedundantWrapping\":false}"},
	{source: "new RegExp;"},
	{source: "new RegExp();"},
	{source: "RegExp();"},
	{source: "new RegExp('a', 'g', 'b');"},
	{source: "RegExp('a', 'g', 'b');"},
	{source: "new RegExp(`a`, `g`, `b`);"},
	{source: "RegExp(`a`, `g`, `b`);"},
	{source: "new RegExp(String.raw`a`, String.raw`g`, String.raw`b`);"},
	{source: "RegExp(String.raw`a`, String.raw`g`, String.raw`b`);"},
	{source: "new RegExp(/a/, 'u', 'foo');", options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(String`a`);"},
	{source: "RegExp(raw`a`);"},
	{source: "new RegExp(f(String.raw)`a`);"},
	{source: "RegExp(string.raw`a`);"},
	{source: "new RegExp(String.Raw`a`);"},
	{source: "new RegExp(String[raw]`a`);"},
	{source: "RegExp(String.raw.foo`a`);"},
	{source: "new RegExp(String.foo.raw`a`);"},
	{source: "RegExp(foo.String.raw`a`);"},
	{source: "new RegExp(String.raw);"},
	{source: "let String; new RegExp(String.raw`a`);"},
	{source: "function foo() { var String; new RegExp(String.raw`a`); }"},
	{source: "function foo(String) { RegExp(String.raw`a`); }"},
	{source: "if (foo) { const String = bar; RegExp(String.raw`a`); }"},
	{source: "new Regexp('abc');"},
	{source: "Regexp(`a`);"},
	{source: "new Regexp(String.raw`a`);"},
	{source: "let RegExp; new RegExp('a');"},
	{source: "function foo() { var RegExp; RegExp('a', 'g'); }"},
	{source: "function foo(RegExp) { new RegExp(String.raw`a`); }"},
	{source: "if (foo) { const RegExp = bar; RegExp('a'); }"},
	{source: "class C { #RegExp; foo() { globalThis.#RegExp('a'); } }"},
	{source: "new RegExp('[[A--B]]' + a, 'v')"},
}

// preferRegexLiteralsReporting holds every case upstream expects to report, with its message ids and
// the exact source each offered suggestion would produce.
var preferRegexLiteralsReporting = []preferRegexLiteralsRow{
	{source: "new RegExp('abc');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "RegExp('abc');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "new RegExp('abc', 'g');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "RegExp('abc', 'g');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "new RegExp(`abc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "RegExp(`abc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "new RegExp(`abc`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "RegExp(`abc`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "new RegExp(String.raw`abc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "new RegExp(String.raw`abc\nabc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc\\nabc/;"},
	}},
	{source: "new RegExp(String.raw`\tabc\nabc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\tabc\\nabc/;"},
	}},
	{source: "RegExp(String.raw`abc`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/;"},
	}},
	{source: "new RegExp(String.raw`abc`, String.raw`g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "RegExp(String.raw`abc`, String.raw`g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/g;"},
	}},
	{source: "new RegExp(String['raw']`a`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}},
	{source: "new RegExp('');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/(?:)/;"},
	}},
	{source: "RegExp('', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/(?:)/;"},
	}},
	{source: "new RegExp(String.raw``);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/(?:)/;"},
	}},
	{source: "new RegExp('a', `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/g;"},
	}},
	{source: "RegExp(`a`, 'g');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/g;"},
	}},
	{source: "RegExp(String.raw`a`, 'g');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/g;"},
	}},
	{source: "new RegExp(String.raw`\\d`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\d/g;"},
	}},
	{source: "new RegExp(String.raw`\\\\d`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\\\d/g;"},
	}},
	{source: "new RegExp(String['raw']`\\\\d`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\\\d/g;"},
	}},
	{source: "new RegExp(String[\"raw\"]`\\\\d`, `g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\\\d/g;"},
	}},
	{source: "RegExp('a', String.raw`g`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/g;"},
	}},
	{source: "new globalThis.RegExp('a');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2020}"},
	{source: "globalThis.RegExp('a');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2020}"},
	{source: "new RegExp(/a/);", ids: []string{"unexpectedRedundantRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/, 'u');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/u;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/g, '');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/;"},
		{id: "replaceWithIntendedLiteralAndFlags", output: "/a/g;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/g, 'g');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/g;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/ig, 'g');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/g;"},
		{id: "replaceWithIntendedLiteralAndFlags", output: "/a/ig;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/g, 'ig');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/ig;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/i, 'g');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/g;"},
		{id: "replaceWithIntendedLiteralAndFlags", output: "/a/ig;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/i, 'i');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/i;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/, `u`);", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/u;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/, `gi`);", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/gi;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp('a');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/, String.raw`u`);", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/u;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp(/a/ /* comment */);", ids: []string{"unexpectedRedundantRegExp"}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "(a)\nnew RegExp(/b/);", ids: []string{"unexpectedRedundantRegExp"}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "(a)\nnew RegExp(/b/, 'g');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "a/RegExp(/foo/);", ids: []string{"unexpectedRedundantRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "a/ /foo/;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "RegExp(/foo/)in a;", ids: []string{"unexpectedRedundantRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/foo/ in a;"},
	}, options: "{\"disallowRedundantWrapping\":true}"},
	{source: "new RegExp((String?.raw)`a`);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/a/;"},
	}},
	{source: "new RegExp('+');", ids: []string{"unexpectedRegExp"}},
	{source: "new RegExp('*');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('+');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('*');", ids: []string{"unexpectedRegExp"}},
	{source: "new RegExp('+', 'g');", ids: []string{"unexpectedRegExp"}},
	{source: "new RegExp('*', 'g');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('+', 'g');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('*', 'g');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('abc', 'd');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/abc/d;"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2022}"},
	{source: "RegExp('\\\\\\\\', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\\\/;"},
	}},
	{source: "RegExp('\\n', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n/;"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2022}"},
	{source: "RegExp('\\n\\n', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n\\n/;"},
	}},
	{source: "RegExp('\\t', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\t/;"},
	}},
	{source: "RegExp('\\t\\t', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\t\\t/;"},
	}},
	{source: "RegExp('\\r\\n', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\r\\n/;"},
	}},
	{source: "RegExp('\\u1234', 'g')", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('\\u{1234}', 'g')", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('\\u{11111}', 'g')", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('\\v', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\v/;"},
	}},
	{source: "RegExp('\\v\\v', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\v\\v/;"},
	}},
	{source: "RegExp('\\f', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\f/;"},
	}},
	{source: "RegExp('\\f\\f', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\f\\f/;"},
	}},
	{source: "RegExp('\\\\b', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\b/;"},
	}},
	{source: "RegExp('\\\\b\\\\b', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\b\\b/;"},
	}},
	{source: "new RegExp('\\\\B\\\\b', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\B\\b/;"},
	}},
	{source: "RegExp('\\\\w', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\w/;"},
	}},
	{source: "new globalThis.RegExp('\\\\W', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\W/;"},
	}, upstreamLanguageOptions: "{\"globals\":{\"globalThis\":\"readonly\"}}"},
	{source: "RegExp('\\\\s', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\s/;"},
	}},
	{source: "new RegExp('\\\\S', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\S/"},
	}},
	{source: "globalThis.RegExp('\\\\d', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\d/;"},
	}, upstreamLanguageOptions: "{\"globals\":{\"globalThis\":\"readonly\"}}"},
	{source: "globalThis.RegExp('\\\\D', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\D/"},
	}, upstreamLanguageOptions: "{\"globals\":{\"globalThis\":\"readonly\"}}"},
	{source: "globalThis.RegExp('\\\\\\\\\\\\D', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\\\\\D/"},
	}, upstreamLanguageOptions: "{\"globals\":{\"globalThis\":\"readonly\"}}"},
	{source: "new RegExp('\\\\D\\\\D', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\D\\D/"},
	}},
	{source: "new globalThis.RegExp('\\\\0\\\\0', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\0\\0/;"},
	}, upstreamLanguageOptions: "{\"globals\":{\"globalThis\":\"writable\"}}"},
	{source: "new RegExp('\\\\0\\\\0', '');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\0\\0/;"},
	}},
	{source: "new RegExp('\\0\\0', 'g');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('\\\\0\\\\0\\\\0', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\0\\0\\0/"},
	}},
	{source: "RegExp('\\\\78\\\\126\\\\5934', '')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\78\\126\\5934/"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2022}"},
	{source: "a in(RegExp('abc'))", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "a in(/abc/)"},
	}},
	{source: "x = y\n            RegExp(\"foo\").test(x) ? bar() : baz()", ids: []string{"unexpectedRegExp"}},
	{source: "func(new RegExp(String.raw`\\w{1, 2`, 'u'),new RegExp(String.raw`\\w{1, 2`, 'u'))", ids: []string{"unexpectedRegExp", "unexpectedRegExp"}},
	{source: "x = y;\n            RegExp(\"foo\").test(x) ? bar() : baz()", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "x = y;\n            /foo/.test(x) ? bar() : baz()"},
	}},
	{source: "typeof RegExp(\"foo\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "typeof /foo/"},
	}},
	{source: "RegExp(\"foo\") instanceof RegExp(String.raw`blahblah`, 'g') ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false", ids: []string{"unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/foo/ instanceof RegExp(String.raw`blahblah`, 'g') ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false"},
		{id: "replaceWithLiteral", output: "RegExp(\"foo\") instanceof /blahblah/g ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false"},
		{id: "replaceWithLiteral", output: "RegExp(\"foo\") instanceof RegExp(String.raw`blahblah`, 'g') ? typeof /(\\p{Emoji_Presentation})\\1/ug : false"},
	}},
	{source: "[   new RegExp(`someregular`)]", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "[   /someregular/]"},
	}},
	{source: "const totallyValidatesEmails = new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")\n            if (typeof totallyValidatesEmails === 'object') {\n                runSomethingThatExists(Regexp('stuff'))\n            }", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "const totallyValidatesEmails = /\\S+@(\\S+\\.)+\\S+/\n            if (typeof totallyValidatesEmails === 'object') {\n                runSomethingThatExists(Regexp('stuff'))\n            }"},
	}},
	{source: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'", ids: []string{"unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp", "unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "!/^Hey, /u && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && /jk$/ && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~/^Sup, / || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || /hi/ + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + /person/ === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -/hi again/ ? 5 * new RegExp('abc') : 'notregbutstring'"},
		{id: "replaceWithLiteral", output: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * /abc/ : 'notregbutstring'"},
	}},
	{source: "#!/usr/bin/sh\n            RegExp(\"foo\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "#!/usr/bin/sh\n            /foo/"},
	}},
	{source: "async function abc(){await new RegExp(\"foo\")}", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "{\"ecmaVersion\":8,\"sourceType\":\"module\"}"},
	{source: "function* abc(){yield new RegExp(\"foo\")}", ids: []string{"unexpectedRegExp"}},
	{source: "function* abc(){yield* new RegExp(\"foo\")}", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "function* abc(){yield* /foo/}"},
	}},
	{source: "console.log({ ...new RegExp('a') })", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "console.log({ .../a/ })"},
	}},
	{source: "delete RegExp('a');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "delete /a/;"},
	}},
	{source: "void RegExp('a');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "void /a/;"},
	}},
	{source: "new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")**RegExp('a')", ids: []string{"unexpectedRegExp", "unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\S+@(\\S+\\.)+\\S+/**RegExp('a')"},
		{id: "replaceWithLiteral", output: "new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")**/a/"},
	}},
	{source: "new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")%RegExp('a')", ids: []string{"unexpectedRegExp", "unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\S+@(\\S+\\.)+\\S+/%RegExp('a')"},
		{id: "replaceWithLiteral", output: "new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")%/a/"},
	}},
	{source: "a in RegExp('abc')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "a in /abc/"},
	}},
	{source: "\n            /abc/ == new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ == /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ === new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ === /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ != new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ != /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ !== new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ !== /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ > new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ > /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ < new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ < /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ >= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ >= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ <= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ <= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ << new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ << /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ >> new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ >> /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ >>> new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ >>> /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ ^ new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ ^ /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ & new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ & /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            /abc/ | new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            /abc/ | /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            null ?? new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            null ?? /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc *= new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc *= /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            console.log({a: new RegExp('sup')})\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            console.log({a: /sup/})\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            console.log(() => {new RegExp('sup')})\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            console.log(() => {/sup/})\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            function abc() {new RegExp('sup')}\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            function abc() {/sup/}\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            function abc() {return new RegExp('sup')}\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            function abc() {return /sup/}\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc <<= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc <<= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc >>= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc >>= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc >>>= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc >>>= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc ^= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc ^= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc &= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc &= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc |= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc |= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc ??= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc ??= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc &&= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc &&= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc ||= new RegExp('cba');\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc ||= /cba/;\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc **= new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc **= /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc /= new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc /= /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc += new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc += /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc -= new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc -= /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            abc %= new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            abc %= /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "\n            () => new RegExp('blah')\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            () => /blah/\n            "},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2021}"},
	{source: "a/RegExp(\"foo\")in b", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "a/ /foo/ in b"},
	}},
	{source: "a/RegExp(\"foo\")instanceof b", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "a/ /foo/ instanceof b"},
	}},
	{source: "do RegExp(\"foo\")\nwhile (true);", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "do /foo/\nwhile (true);"},
	}},
	{source: "for(let i;i<5;i++) { break\nnew RegExp('search')}", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "for(let i;i<5;i++) { break\n/search/}"},
	}},
	{source: "for(let i;i<5;i++) { continue\nnew RegExp('search')}", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "for(let i;i<5;i++) { continue\n/search/}"},
	}},
	{source: "\n            switch (value) {\n                case \"possibility\":\n                    console.log('possibility matched')\n                case RegExp('myReg').toString():\n                    console.log('matches a regexp\\' toString value')\n                    break;\n            }\n            ", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "\n            switch (value) {\n                case \"possibility\":\n                    console.log('possibility matched')\n                case /myReg/.toString():\n                    console.log('matches a regexp\\' toString value')\n                    break;\n            }\n            "},
	}},
	{source: "throw new RegExp('abcdefg') // fail with a regular expression", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "throw /abcdefg/ // fail with a regular expression"},
	}},
	{source: "for (value of new RegExp('something being searched')) { console.log(value) }", ids: []string{"unexpectedRegExp"}},
	{source: "(async function(){for await (value of new RegExp('something being searched')) { console.log(value) }})()", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "{\"ecmaVersion\":2018}"},
	{source: "for (value in new RegExp('something being searched')) { console.log(value) }", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "for (value in /something being searched/) { console.log(value) }"},
	}},
	{source: "if (condition1 && condition2) new RegExp('avalue').test(str);", ids: []string{"unexpectedRegExp"}},
	{source: "debugger\nnew RegExp('myReg')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "debugger\n/myReg/"},
	}},
	{source: "RegExp(\"\\\\\\n\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n/"},
	}},
	{source: "RegExp(\"\\\\\\t\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\t/"},
	}},
	{source: "RegExp(\"\\\\\\f\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\f/"},
	}},
	{source: "RegExp(\"\\\\\\v\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\v/"},
	}},
	{source: "RegExp(\"\\\\\\r\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\r/"},
	}},
	{source: "new RegExp(\"\t\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\t/"},
	}},
	{source: "new RegExp(\"/\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\//"},
	}},
	{source: "new RegExp(\"\\.\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/./"},
	}},
	{source: "new RegExp(\"\\\\.\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\./"},
	}},
	{source: "new RegExp(\"\\\\\\n\\\\\\n\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n\\n/"},
	}},
	{source: "new RegExp(\"\\\\\\n\\\\\\f\\\\\\n\")", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n\\f\\n/"},
	}},
	{source: "new RegExp(\"\\u000A\\u000A\");", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/\\n\\n/;"},
	}},
	{source: "new RegExp('mysafereg' /* comment explaining its safety */)", ids: []string{"unexpectedRegExp"}},
	{source: "new RegExp('[[A--B]]', 'v')", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "/[[A--B]]/v"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "new RegExp('a', 'uv')", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "new RegExp(/a/, 'v')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/v"},
	}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "new RegExp(/a/g, 'v')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/v"},
		{id: "replaceWithIntendedLiteralAndFlags", output: "/a/gv"},
	}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "new RegExp(/a/u, 'v')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/v"},
	}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "new RegExp(/a/v, 'u')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteralAndFlags", output: "/a/u"},
	}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "{\"ecmaVersion\":2024}"},
	{source: "var regex = new RegExp('foo', 'u');", ids: []string{"unexpectedRegExp"}, suggestions: []preferRegexLiteralsSuggestion{
		{id: "replaceWithLiteral", output: "var regex = /foo/u;"},
	}, upstreamLanguageOptions: "{\"ecmaVersion\":2015}"},
}

// preferRegexLiteralsDecidedByGlobalsConfig holds the four upstream clean cases whose verdict is not
// the rule's to make.
//
// Each is clean upstream ONLY because the ESLint configuration removes a name from the global scope,
// either through `languageOptions.globals` or through a `/* globals X:off */` directive comment.
// With `RegExp` or `String` declared off, upstream's ReferenceTracker finds no global reference and
// the rule never looks at the call.
//
// cohere has no such surface. Globals come from the TypeScript standard library, which always
// declares `RegExp` and `String`, and there is no configuration that can un-declare one. So these
// report here, and pinning them as reporting is the honest record: the brief's instruction for a
// case decided above the rule is to record it at the layer that actually decides it rather than to
// weaken the rule until the case goes green, which would turn a fact about the config into a fact
// about the rule.
//
// The pairing is what makes this legible. Two of the four are byte-identical to rows in the
// reporting table and differ only in the globals setting, which is upstream stating outright that
// the config is the whole difference.
var preferRegexLiteralsDecidedByGlobalsConfig = []preferRegexLiteralsRow{
	{source: "/* globals String:off */ new RegExp(String.raw`a`);", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('a', String.raw`g`);", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "{\"globals\":{\"String\":\"off\"}}"},
	{source: "/* globals RegExp:off */ new RegExp('a');", ids: []string{"unexpectedRegExp"}},
	{source: "RegExp('a');", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "{\"globals\":{\"RegExp\":\"off\"}}"},
}

// TestPreferRegexLiteralsReportsWhatTheGlobalsConfigWouldHaveSilenced pins those four as reporting.
//
// Asserting the divergence rather than deleting the cases is the point: if cohere ever grows a
// surface for declaring a global off, this test is what tells the next reader these four were a
// known consequence of not having one, and it fails the moment the behaviour changes in either
// direction.
func TestPreferRegexLiteralsReportsWhatTheGlobalsConfigWouldHaveSilenced(t *testing.T) {
	t.Parallel()

	for _, row := range preferRegexLiteralsDecidedByGlobalsConfig {
		t.Run(row.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runPreferRegexLiterals(t, row), row.ids...)
		})
	}
}

// preferRegexLiteralsDecidedByEcmaVersion holds the thirteen upstream cases whose verdict depends on
// which edition of JavaScript the file claims to be.
//
// ESLint carries `languageOptions.ecmaVersion` per config, and this rule consults it twice: through
// `getRegexppEcmaVersion`, which decides whether a pattern or a flag PARSES, and through the global
// scope, which decides whether `globalThis` exists as a name at all. cohere has no equivalent. A
// program is compiled once against one set of compiler options, and a rule cannot ask what language
// edition an individual file claims, so validity here is judged at the latest edition.
//
// The rows below record what cohere does, with upstream's own gate in the note. They split three
// ways and the directions are opposite, which is why they are one table rather than one sentence:
//
//   - `globalThis` did not exist before ES2020, so upstream is SILENT on the three cases pinned at
//     ecmaVersion 5, 2015 and 2017 and cohere REPORTS them. Reporting is the safe direction: the
//     code is a `RegExp` construction either way, and every file in this codebase targets an
//     edition where `globalThis` is real.
//   - The `d`, `u`, `v` and `uv` flags each arrived in a known year, so upstream WITHHOLDS the
//     suggestion when a case is pinned before it and cohere OFFERS one. Also the safe direction: a
//     suggestion is never applied without a human, and the literal it proposes is valid today.
//   - Inline modifiers, `(?i:foo)`, are an ES2025 addition, so this one runs the OTHER way. Cohere
//     withholds where upstream at ecmaVersion 2025 offers. Measured directly against eslint 10.8.1:
//     the same input yields no suggestion at 2024 and one at 2025, so cohere currently matches the
//     2024 reading. This is the one row where closing the gap would mean teaching the pattern
//     scanner a syntax it does not know, rather than reading a version we do not have.
//
// One row is not about the version at all and rides along because it is the same kind of fact.
// `new window['RegExp'](...)` reports upstream only because the config declares `window` a global;
// with no DOM library in the fixture program the checker resolves nothing, so cohere is silent. It
// is pinned here rather than deleted so the difference stays visible.
var preferRegexLiteralsDecidedByLanguageEdition = []preferRegexLiteralsRow{
	// Upstream is silent because `globalThis` is not a name before ES2020. Cohere reports.
	{source: "new globalThis.RegExp('a');", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 5, silent upstream"},

	// Upstream withholds the suggestion because the flag does not exist at the pinned edition.
	// Cohere offers one.
	{source: "new RegExp(/a/, 'd');", ids: []string{"unexpectedRedundantRegExpWithFlags"}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "ecmaVersion 2021, no suggestion upstream"},
	{source: "RegExp('abc', 'u');", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 3, no suggestion upstream"},
	{source: "new RegExp('abc', 'd');", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 2021, no suggestion upstream"},
	{source: "new RegExp('[[A--B]]', 'v')", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 2023, no suggestion upstream"},
	{source: "new RegExp('[[A&&&]]', 'v')", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 2024, no suggestion upstream"},
	{source: "new RegExp(/a/, 'v')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "ecmaVersion 2023, no suggestion upstream"},
	{source: "new RegExp(/[[A--B]]/v, 'u')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "ecmaVersion 2024, no suggestion upstream"},
	{source: "new RegExp(/[[A--B]]/v, 'g')", ids: []string{"unexpectedRedundantRegExpWithFlags"}, options: "{\"disallowRedundantWrapping\":true}", upstreamLanguageOptions: "ecmaVersion 2024, one suggestion upstream"},

	// The other direction: an ES2025 syntax the pattern scanner does not know, so cohere withholds
	// where upstream at 2025 offers.
	{source: "new RegExp('(?i:foo)bar')", ids: []string{"unexpectedRegExp"}, upstreamLanguageOptions: "ecmaVersion 2025, one suggestion upstream"},
}

// TestPreferRegexLiteralsRecordsTheLanguageEditionDivergence pins those cases at what cohere does.
//
// Asserting them rather than deleting them is the point. Each is a real behavioural difference from
// upstream, and a difference nobody wrote down is one the next reader rediscovers as a bug. If
// cohere ever grows a per-file language edition, this test is what fails and tells them these rows
// were a known consequence of not having one.
//
// Only the finding is asserted here, not the suggestion. The finding is the part that is stable
// across editions for these rows; the suggestion is exactly what the edition decides, and it is
// described at each row rather than pinned, so this test does not have to be rewritten every time
// the pattern scanner learns a new syntax.
func TestPreferRegexLiteralsRecordsTheLanguageEditionDivergence(t *testing.T) {
	t.Parallel()

	for _, row := range preferRegexLiteralsDecidedByLanguageEdition {
		t.Run(row.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runPreferRegexLiterals(t, row), row.ids...)
		})
	}
}

// TestPreferRegexLiteralsFollowsWindowAsTheGlobalObject pins the `window` row.
//
// The calls come from the shelf's ReferenceTracker, which takes an undeclared `window` for the runtime's
// global object, as it does for every rule on it. ESLint's corpus runs this row with `window` declared a
// browser global and reports it, so the two agree here; without that config ESLint is silent, which is
// the one edge the ESLint table leaves out.
func TestPreferRegexLiteralsFollowsWindowAsTheGlobalObject(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile,
		"new window['RegExp']('\\x56\\x78\\x45', '');"), "unexpectedRegExp")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile,
		"window.RegExp('a');"), "unexpectedRegExp")
}

func TestPreferRegexLiteralsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, row := range preferRegexLiteralsClean {
		t.Run(row.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runPreferRegexLiterals(t, row))
		})
	}
}

func TestPreferRegexLiteralsFires(t *testing.T) {
	t.Parallel()

	for _, row := range preferRegexLiteralsReporting {
		t.Run(row.source, func(t *testing.T) {
			t.Parallel()
			result := runPreferRegexLiterals(t, row)
			rule_testing.ExpectFindings(t, result, row.ids...)
			expectPreferRegexLiteralsSuggestions(t, row, result)
		})
	}
}

// expectPreferRegexLiteralsSuggestions asserts what each finding OFFERS, not merely that it fired.
//
// This runs inside the Fires loop rather than as its own pass over the same table, and that is a
// cost decision rather than a stylistic one: the typed harness builds a whole program per call, so a
// second loop over 181 cases is 181 more programs and about three minutes. Folding the assertion in
// keeps one program per case.
//
// It is a separate assertion for a reason the brief makes plainly. A message-id assertion cannot see
// a suggestion that rewrites the right span with the wrong text, and it cannot see a suggestion
// offered where upstream withholds one. Thirty-three of the corpus's findings deliberately carry no
// suggestion, and those rows are asserted here as offering none.
func expectPreferRegexLiteralsSuggestions(t *testing.T, row preferRegexLiteralsRow, result rule_testing.Result) {
	t.Helper()

	var offered []preferRegexLiteralsSuggestion
	for _, diagnostic := range result.Diagnostics {
		for _, suggestion := range diagnostic.Suggestions {
			offered = append(offered, preferRegexLiteralsSuggestion{
				id:     suggestion.Message.Id,
				output: applyPreferRegexLiteralsSuggestion(t, row.source, suggestion),
			})
		}
	}

	if len(offered) != len(row.suggestions) {
		t.Fatalf("wanted %d suggestions, got %d: %+v", len(row.suggestions), len(offered), offered)
	}
	for index, want := range row.suggestions {
		got := offered[index]
		if got.id != want.id {
			t.Errorf("suggestion %d: wanted id %q, got %q", index, want.id, got.id)
		}
		// Both sides are normalised the same way before comparing, and the reason is the harness
		// rather than either side being wrong.
		//
		// The rule saw `strings.TrimSpace(source)+"\n"`, so the applier's result carries the
		// harness's trailing newline and has lost the fixture's leading indentation. Upstream's
		// `output` is the untrimmed source with the repair applied. Trimming both is what compares
		// the REPAIR rather than the whitespace two different writers happened to put around it.
		//
		// Getting this wrong in either direction is invisible in the right way and loud in the
		// wrong one: comparing raw failed thirty-five indented cases and read as a broken fixer,
		// and transforming only the expectation failed a hundred and fifty-seven on a trailing
		// newline. The rule did not change between those runs.
		if strings.TrimSpace(got.output) != strings.TrimSpace(want.output) {
			t.Errorf("suggestion %d: wanted output %q, got %q", index, want.output, got.output)
		}
	}
}

// applyPreferRegexLiteralsSuggestion replays one suggestion's edits into the source and returns the
// result.
//
// Edits are applied from the end backwards so an earlier edit does not shift a later one's offsets.
// Every suggestion this rule proposes carries exactly one edit, but the loop does not assume that,
// because an applier that silently mishandles two would be a wrong instrument rather than a failing
// one.
func applyPreferRegexLiteralsSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	// Edits are offsets into the text the RULE saw, and that is not the literal in the table above:
	// the typed harness writes each fixture as `strings.TrimSpace(contents)+"\n"`
	// (internal/rule_testing/program.go:160). Replaying an edit into the untrimmed literal lands it
	// several bytes early, which cost real time here and presented as a broken fixer rather than as
	// a harness difference: thirty-five of upstream's preceding-token cases are written with leading
	// indentation, and every one of them spliced into the middle of its own call. The rule was
	// correct throughout.
	edited := harnessSourceForPreferRegexLiterals(source)
	for index := len(suggestion.Fixes) - 1; index >= 0; index-- {
		fix := suggestion.Fixes[index]
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(edited) || start > end {
			t.Fatalf("suggestion edit [%d,%d) is outside the source of length %d", start, end, len(edited))
		}
		edited = edited[:start] + fix.Text + edited[end:]
	}
	return edited
}

// harnessSourceForPreferRegexLiterals mirrors what the typed harness writes to disk.
//
// `internal/rule_testing/program.go:160` writes each fixture as `strings.TrimSpace(contents)+"\n"`,
// so the text the rule saw is not the literal in the table above. Kept as one named function rather
// than inlined at both call sites, so the expectation and the applier cannot drift apart.
func harnessSourceForPreferRegexLiterals(source string) string {
	return strings.TrimSpace(source) + "\n"
}

func runPreferRegexLiterals(t *testing.T, row preferRegexLiteralsRow) rule_testing.Result {
	t.Helper()
	if row.options == "" {
		return rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile, row.source)
	}
	// Routed through the rule's own registered decoder rather than by building the options struct,
	// so the wire shape and the json tag are under test too. cohere's config layer unwraps the
	// `[severity, options]` tuple before dispatch, so the decoder receives the bare object even
	// though upstream's corpus writes it inside an array.
	decoded, err := rule.DecodeOptionsInto[PreferRegexLiteralsOptions]()([]byte(row.options))
	if err != nil {
		t.Fatalf("decoding options %s: %v", row.options, err)
	}
	return rule_testing.RunTypedWithOptions(t, PreferRegexLiterals, preferRegexLiteralsFile, row.source, decoded)
}

// TestPreferRegexLiteralsNeedsTheTypedHarness pins the checker declaration.
//
// A typed rule handed a nil checker goes silent rather than crashing, which is the more dangerous
// failure: every StaysSilent case would pass vacuously and the suite would stay green over a rule
// that had stopped working. This asserts the typed harness is what makes it report, so a later
// revert of `NeedsTypeChecker` fails loudly.
func TestPreferRegexLiteralsNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "new RegExp('abc');"
	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile, source),
		"unexpectedRegExp")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, PreferRegexLiterals, preferRegexLiteralsFile, source))
}

// TestPreferRegexLiteralsReportsTheWholeCall asserts WHERE the finding points.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule anchored on the callee rather
// than the call would pass every row above while pointing at the wrong span, and the suggestion it
// carries would then replace the wrong bytes. Upstream reports the whole call expression.
func TestPreferRegexLiteralsReportsTheWholeCall(t *testing.T) {
	t.Parallel()

	const source = "const pattern = new RegExp('abc', 'g');"
	result := rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "new RegExp('abc', 'g')" {
		t.Errorf("wanted the whole call, got %q", reported)
	}
}

// TestPreferRegexLiteralsExplainsWhyItMatters asserts the rendered message text exactly.
//
// Asserting the id alone cannot see a reworded description, and a rule whose message merely restates
// its own name is the first one somebody disables. Compared against a literal typed here rather than
// against the rule's own constant, since comparing a diagnostic to the constant it was built from is
// an equality that moves on both sides under mutation and therefore proves nothing.
func TestPreferRegexLiteralsExplainsWhyItMatters(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, PreferRegexLiterals, preferRegexLiteralsFile, "new RegExp('abc');")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	description := result.Diagnostics[0].Message.Description
	if !strings.HasPrefix(description, "This builds a regular expression by handing constant strings") {
		t.Errorf("message does not open by naming the mistake: %q", description)
	}
	if !strings.Contains(description, "Write the regular expression literal.") {
		t.Errorf("message does not say what to do instead: %q", description)
	}
}
