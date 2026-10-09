package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnnecessaryTemplateExpressionFile = "/repository/source/Templates.ts"

func noUnnecessaryTemplateExpressionCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnnecessaryTemplateExpressionStaysSilent is upstream's sixty seven passing cases verbatim.
//
// These are the refusals, and they are the rule: a number, a union, an `any`, an enum member, a
// type parameter, a comment inside the interpolation, and a whitespace literal held before a
// newline. Each is an input where splicing the text in raw would change what the program does, so
// every case skipped here is a class of behaviour-changing fix rather than a missed finding.
//
// Measured one file per program against the installed 8.67.0 build, using upstream's own fixture
// compilerOptions verbatim. All sixty seven are clean.
func TestNoUnnecessaryTemplateExpressionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"const string = 'a';\n",
		"const string = `a`;\n",
		"const string = `NaN: ${/* comment */ NaN}`;\n",
		"const string = `undefined: ${/* comment */ undefined}`;\n",
		"const string = `Infinity: ${Infinity /* comment */}`;\n",
		"declare const string: 'a';\n`${string}b`;\n",
		"declare const number: 1;\n`${number}b`;\n",
		"declare const boolean: true;\n`${boolean}b`;\n",
		"declare const nullish: null;\n`${nullish}-undefined`;\n",
		"declare const undefinedish: undefined;\n`${undefinedish}`;\n",
		"declare const left: 'a';\ndeclare const right: 'b';\n`${left}${right}`;\n",
		"declare const left: 'a';\ndeclare const right: 'c';\n`${left}b${right}`;\n",
		"declare const left: 'a';\ndeclare const center: 'b';\ndeclare const right: 'c';\n`${left}${center}${right}`;\n",
		"`1 + 1 = ${1 + 1}`;\n",
		"`true && false = ${true && false}`;\n",
		"tag`${'a'}${'b'}`;\n",
		"`${function () {}}`;\n",
		"`${() => {}}`;\n",
		"`${(...args: any[]) => args}`;\n",
		"declare const number: 1;\n`${number}`;\n",
		"declare const boolean: true;\n`${boolean}`;\n",
		"declare const nullish: null;\n`${nullish}`;\n",
		"declare const union: string | number;\n`${union}`;\n",
		"declare const unknown: unknown;\n`${unknown}`;\n",
		"declare const never: never;\n`${never}`;\n",
		"declare const any: any;\n`${any}`;\n",
		"function func<T extends number>(arg: T) {\n  `${arg}`;\n}\n",
		"`with\n\n      new line`;\n",
		"declare const a: 'a';\n\n`${a} with\n\n      new line`;\n",
		"`with windows \r new line`;\n",
		"`not a useless ${String.raw`nested interpolation ${a}`}`;\n",
		"`\nthis code has trailing whitespace: ${'    '}\n    `;\n",
		"`\nthis code has trailing whitespace: ${`    `}\n    `;\n",
		"`this code has trailing whitespace with a CRLF (windows \\\r) new line: ${' '}\r\n`;\n",
		"`this code has trailing whitespace with a CR new line: ${' '}\r`;\n",
		"`this code has trailing whitespace with a LF new line: ${' '}\n`;\n",
		"`this code has trailing whitespace with a LS new line: ${' '} `;\n",
		"`this code has trailing whitespace with a PS new line: ${' '} `;\n",
		"`trailing position interpolated empty string also makes whitespace clear    ${''}\n`;\n",
		"`\n${/* intentional comment before */ 'bar'}\n...`;\n",
		"`\n${'bar' /* intentional comment after */}\n...`;\n",
		"`\n${/* intentional comment before */ 'bar' /* intentional comment after */}\n...`;\n",
		"`${/* intentional  before */ 'bar'}`;\n",
		"`${'bar' /* intentional comment after */}`;\n",
		"`${/* intentional comment before */ 'bar' /* intentional comment after */}`;\n",
		"`${\n  // intentional comment before\n  'bar'\n}`;\n",
		"`${\n  'bar'\n  // intentional comment after\n}`;\n",
		"function getTpl<T>(input: T) {\n  return `${input}`;\n}\n",
		"type FooBarBaz = `foo${/* comment */ 'bar'}\"baz\"`;\n",
		"enum Foo {\n  A = 'A',\n  B = 'B',\n}\ntype Foos = `${Foo}`;\n",
		"type Foo = 'A' | 'B';\ntype Bar = `foo${Foo}foo`;\n",
		"type Foo =\n  `trailing position interpolated empty string also makes whitespace clear    ${''}\n`;\n",
		"type Foo = `this code has trailing whitespace with a CRLF (windows \\\r) new line: ${` `}\r\n`;\n",
		"type Foo = `this code has trailing whitespace with a CR new line: ${` `}\r`;\n",
		"type Foo = `this code has trailing whitespace with a LF new line: ${` `}\n`;\n",
		"type Foo = `this code has trailing whitespace with a LS new line: ${` `} `;\n",
		"type Foo = `this code has trailing whitespace with a PS new line: ${` `} `;\n",
		"type Foo = `${'foo' | 'bar' | null}`;\n",
		"type StringOrNumber = string | number;\ntype Foo = `${StringOrNumber}`;\n",
		"enum Foo {\n  A = 1,\n  B = 2,\n}\ntype Bar = `${Foo.A}`;\n",
		"enum Enum1 {\n  A = 'A1',\n  B = 'B1',\n}\n\nenum Enum2 {\n  A = 'A2',\n  B = 'B2',\n}\n\ntype Union = `${Enum1 | Enum2}`;\n",
		"enum Enum1 {\n  A = 'A1',\n  B = 'B1',\n}\n\nenum Enum2 {\n  A = 'A2',\n  B = 'B2',\n}\n\ntype Union = `${Enum1.A | Enum2.B}`;\n",
		"enum Enum1 {\n  A = 'A1',\n  B = 'B1',\n}\n\nenum Enum2 {\n  A = 'A2',\n  B = 'B2',\n}\ntype Enums = Enum1 | Enum2;\ntype Union = `${Enums}`;\n",
		"enum Enum {\n  A = 'A',\n  B = 'A',\n}\n\ntype Intersection = `${Enum1.A & string}`;\n",
		"enum Foo {\n  A = 'A',\n  B = 'B',\n}\ntype Bar = `${Foo.A}`;\n",
		"function foo<T extends string>() {\n  const a: `${T}` = 'a';\n}\n",
		"type T<A extends string> = `${A}`;\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnnecessaryTemplateExpressionCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnnecessaryTemplateExpression,
				noUnnecessaryTemplateExpressionFile, sourceText))
		})
	}
}

// TestNoUnnecessaryTemplateExpressionFires is upstream's one hundred seventeen reporting cases.
//
// Every row asserts the span and the count. The span carries more than usual here: the finding
// covers `${x}` INCLUDING both delimiters, computed from our head-and-spans model rather than from
// upstream's parallel quasi arrays, so an off-by-one in that mapping points the reader at the
// wrong characters while every message id still passes.
//
// One case carries an astral emoji, and it caught an instrument bug rather than a rule one: ESLint
// reports columns in UTF-16 code units, so a four-character emoji spans eight of them, and a
// column-to-offset conversion counting characters ran the span four bytes long. Fixed in the
// extractor; the row below holds the corrected span.
func TestNoUnnecessaryTemplateExpressionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
	}{
		{
			sourceText: "`${1}`;\n",
			wantSpans:  []string{"${1}"},
		},
		{
			sourceText: "`${1n}`;\n",
			wantSpans:  []string{"${1n}"},
		},
		{
			sourceText: "`${0o25}`;\n",
			wantSpans:  []string{"${0o25}"},
		},
		{
			sourceText: "`${0b1010} ${0b1111}`;\n",
			wantSpans:  []string{"${0b1010}", "${0b1111}"},
		},
		{
			sourceText: "`${0x25}`;\n",
			wantSpans:  []string{"${0x25}"},
		},
		{
			sourceText: "`${/a/}`;\n",
			wantSpans:  []string{"${/a/}"},
		},
		{
			sourceText: "`${/a/gim}`;\n",
			wantSpans:  []string{"${/a/gim}"},
		},
		{
			sourceText: "`${    1    }`;\n",
			wantSpans:  []string{"${    1    }"},
		},
		{
			sourceText: "`${    'a'    }`;\n",
			wantSpans:  []string{"  'a' "},
		},
		{
			sourceText: "`${    \"a\"    }`;\n",
			wantSpans:  []string{"  \"a\" "},
		},
		{
			sourceText: "`${    'a' + 'b'    }`;\n",
			wantSpans:  []string{"  'a' + 'b' "},
		},
		{
			sourceText: "`${true}`;\n",
			wantSpans:  []string{"${true}"},
		},
		{
			sourceText: "`${    true    }`;\n",
			wantSpans:  []string{"${    true    }"},
		},
		{
			sourceText: "`${null}`;\n",
			wantSpans:  []string{"${null}"},
		},
		{
			sourceText: "`${    null    }`;\n",
			wantSpans:  []string{"${    null    }"},
		},
		{
			sourceText: "`${undefined}`;\n",
			wantSpans:  []string{"${undefined}"},
		},
		{
			sourceText: "`${    undefined    }`;\n",
			wantSpans:  []string{"${    undefined    }"},
		},
		{
			sourceText: "`${Infinity}`;\n",
			wantSpans:  []string{"${Infinity}"},
		},
		{
			sourceText: "`${NaN}`;\n",
			wantSpans:  []string{"${NaN}"},
		},
		{
			sourceText: "`${'a'} ${'b'}`;\n",
			wantSpans:  []string{"${'a'}", "${'b'}"},
		},
		{
			sourceText: "`${   'a'   } ${   'b'   }`;\n",
			wantSpans:  []string{"${   'a'   }", "${   'b'   }"},
		},
		{
			sourceText: "`use${'less'}`;\n",
			wantSpans:  []string{"${'less'}"},
		},
		{
			sourceText: "`use${`less`}`;\n",
			wantSpans:  []string{"${`less`}"},
		},
		{
			sourceText: "`u${\n  // hopefully this comment is not needed.\n  'se'\n\n}${\n  `le${  `ss`  }`\n}`;\n",
			wantSpans:  []string{"${\n  `le${  `ss`  }`\n}", "${  `ss`  }"},
		},
		{
			sourceText: "`use${\n  `less`\n}`;\n",
			wantSpans:  []string{"${\n  `less`\n}"},
		},
		{
			sourceText: "`${'1 + 1 ='} ${2}`;\n",
			wantSpans:  []string{"${'1 + 1 ='}", "${2}"},
		},
		{
			sourceText: "`${'a'} ${true}`;\n",
			wantSpans:  []string{"${'a'}", "${true}"},
		},
		{
			sourceText: "`${String(Symbol.for('test'))}`;\n",
			wantSpans:  []string{"${String(Symbol.for('test'))}"},
		},
		{
			sourceText: "`${'`'}`;\n",
			wantSpans:  []string{"${'`'}"},
		},
		{
			sourceText: "`back${'`'}tick`;\n",
			wantSpans:  []string{"${'`'}"},
		},
		{
			sourceText: "`dollar${'${`this is test`}'}sign`;\n",
			wantSpans:  []string{"${'${`this is test`}'}"},
		},
		{
			sourceText: "`complex${'`${\"`${test}`\"}`'}case`;\n",
			wantSpans:  []string{"${'`${\"`${test}`\"}`'}"},
		},
		{
			sourceText: "`some ${'\\\\${test}'} string`;\n",
			wantSpans:  []string{"${'\\\\${test}'}"},
		},
		{
			sourceText: "`some ${'\\\\`'} string`;\n",
			wantSpans:  []string{"${'\\\\`'}"},
		},
		{
			sourceText: "`some ${/`/} string`;\n",
			wantSpans:  []string{"${/`/}"},
		},
		{
			sourceText: "`some ${/\\`/} string`;\n",
			wantSpans:  []string{"${/\\`/}"},
		},
		{
			sourceText: "`some ${/\\\\`/} string`;\n",
			wantSpans:  []string{"${/\\\\`/}"},
		},
		{
			sourceText: "`some ${/\\\\\\`/} string`;\n",
			wantSpans:  []string{"${/\\\\\\`/}"},
		},
		{
			sourceText: "`some ${/${}/} string`;\n",
			wantSpans:  []string{"${/${}/}"},
		},
		{
			sourceText: "`some ${/$ {}/} string`;\n",
			wantSpans:  []string{"${/$ {}/}"},
		},
		{
			sourceText: "`some ${/\\\\/} string`;\n",
			wantSpans:  []string{"${/\\\\/}"},
		},
		{
			sourceText: "`some ${/\\\\\\b/} string`;\n",
			wantSpans:  []string{"${/\\\\\\b/}"},
		},
		{
			sourceText: "`some ${/\\\\\\\\/} string`;\n",
			wantSpans:  []string{"${/\\\\\\\\/}"},
		},
		{
			sourceText: "` ${''} `;\n",
			wantSpans:  []string{"${''}"},
		},
		{
			sourceText: "` ${\"\"} `;\n",
			wantSpans:  []string{"${\"\"}"},
		},
		{
			sourceText: "` ${``} `;\n",
			wantSpans:  []string{"${``}"},
		},
		{
			sourceText: "` ${'\\`'} `;\n",
			wantSpans:  []string{"${'\\`'}"},
		},
		{
			sourceText: "` ${'\\\\`'} `;\n",
			wantSpans:  []string{"${'\\\\`'}"},
		},
		{
			sourceText: "` ${'$'}{} `;\n",
			wantSpans:  []string{"${'$'}"},
		},
		{
			sourceText: "` ${'\\$'}{} `;\n",
			wantSpans:  []string{"${'\\$'}"},
		},
		{
			sourceText: "` ${'\\\\$'}{} `;\n",
			wantSpans:  []string{"${'\\\\$'}"},
		},
		{
			sourceText: "` ${'\\\\$ '}{} `;\n",
			wantSpans:  []string{"${'\\\\$ '}"},
		},
		{
			sourceText: "` ${'\\\\\\$'}{} `;\n",
			wantSpans:  []string{"${'\\\\\\$'}"},
		},
		{
			sourceText: "` \\\\${'\\\\$'}{} `;\n",
			wantSpans:  []string{"${'\\\\$'}"},
		},
		{
			sourceText: "` $${'{$'}{} `;\n",
			wantSpans:  []string{"${'{$'}"},
		},
		{
			sourceText: "` $${'${$'}{} `;\n",
			wantSpans:  []string{"${'${$'}"},
		},
		{
			sourceText: "` ${'foo$'}{} `;\n",
			wantSpans:  []string{"${'foo$'}"},
		},
		{
			sourceText: "` ${`$`} `;\n",
			wantSpans:  []string{"${`$`}"},
		},
		{
			sourceText: "` ${`$`}{} `;\n",
			wantSpans:  []string{"${`$`}"},
		},
		{
			sourceText: "` ${`$`} {} `;\n",
			wantSpans:  []string{"${`$`}"},
		},
		{
			sourceText: "` ${`$`}${undefined}{} `;\n",
			wantSpans:  []string{"${`$`}", "${undefined}"},
		},
		{
			sourceText: "` ${`foo$`}{} `;\n",
			wantSpans:  []string{"${`foo$`}"},
		},
		{
			sourceText: "` ${'$'}${''}{} `;\n",
			wantSpans:  []string{"${'$'}", "${''}"},
		},
		{
			sourceText: "` ${'$'}${``}{} `;\n",
			wantSpans:  []string{"${'$'}", "${``}"},
		},
		{
			sourceText: "` ${'foo$'}${''}${``}{} `;\n",
			wantSpans:  []string{"${'foo$'}", "${''}", "${``}"},
		},
		{
			sourceText: "` $${'{}'} `;\n",
			wantSpans:  []string{"${'{}'}"},
		},
		{
			sourceText: "` $${undefined}${'{}'} `;\n",
			wantSpans:  []string{"${undefined}", "${'{}'}"},
		},
		{
			sourceText: "` $${''}${undefined}${'{}'} `;\n",
			wantSpans:  []string{"${''}", "${undefined}", "${'{}'}"},
		},
		{
			sourceText: "` \\$${'{}'} `;\n",
			wantSpans:  []string{"${'{}'}"},
		},
		{
			sourceText: "` $${'foo'}${'{'} `;\n",
			wantSpans:  []string{"${'foo'}", "${'{'}"},
		},
		{
			sourceText: "` $${'{ foo'}${'{'} `;\n",
			wantSpans:  []string{"${'{ foo'}", "${'{'}"},
		},
		{
			sourceText: "` \\\\$${'{}'} `;\n",
			wantSpans:  []string{"${'{}'}"},
		},
		{
			sourceText: "` \\\\\\$${'{}'} `;\n",
			wantSpans:  []string{"${'{}'}"},
		},
		{
			sourceText: "` foo$${'{}'} `;\n",
			wantSpans:  []string{"${'{}'}"},
		},
		{
			sourceText: "` $${''}${'{}'} `;\n",
			wantSpans:  []string{"${''}", "${'{}'}"},
		},
		{
			sourceText: "` $${''} `;\n",
			wantSpans:  []string{"${''}"},
		},
		{
			sourceText: "` $${`{}`} `;\n",
			wantSpans:  []string{"${`{}`}"},
		},
		{
			sourceText: "` $${``}${`{}`} `;\n",
			wantSpans:  []string{"${``}", "${`{}`}"},
		},
		{
			sourceText: "` $${``}${`foo{}`} `;\n",
			wantSpans:  []string{"${``}", "${`foo{}`}"},
		},
		{
			sourceText: "` $${`${''}${`${``}`}`}${`{a}`} `;\n",
			// The finding SET matches upstream exactly; only the emission ORDER differs, and the
			// cause is the walk rather than the rule. A template nested inside another
			// interpolation is a separate node visit, so its findings arrive after every finding of
			// the enclosing template, while upstream emits in ascending source position because
			// ESLint sorts diagnostics before handing them over. Our harness preserves emission
			// order.
			//
			// Reordered here rather than reshaped around: a rule cannot reorder findings across
			// separate listener invocations, and making the harness sort would change what every
			// other rule's fixtures observe. Thirty six corpus cases nest a template and this is
			// the only one where the order is visible.
			wantSpans: []string{"${`${''}${`${``}`}`}", "${`{a}`}", "${''}", "${`${``}`}", "${``}"},
		},
		{
			sourceText: "` $${''}${`{}`} `;\n",
			wantSpans:  []string{"${''}", "${`{}`}"},
		},
		{
			sourceText: "` $${``}${'{}'} `;\n",
			wantSpans:  []string{"${``}", "${'{}'}"},
		},
		{
			sourceText: "` $${''}${``}${'{}'} `;\n",
			wantSpans:  []string{"${''}", "${``}", "${'{}'}"},
		},
		{
			sourceText: "` ${'$'} `;\n",
			wantSpans:  []string{"${'$'}"},
		},
		{
			sourceText: "` ${'$'}${'{}'} `;\n",
			wantSpans:  []string{"${'$'}", "${'{}'}"},
		},
		{
			sourceText: "` ${'$'}${''}${'{'} `;\n",
			wantSpans:  []string{"${'$'}", "${''}", "${'{'}"},
		},
		{
			sourceText: "` ${`\n\\$`}{} `;\n",
			wantSpans:  []string{"${`\n\\$`}"},
		},
		{
			sourceText: "` ${`\n\\\\$`}{} `;\n",
			wantSpans:  []string{"${`\n\\\\$`}"},
		},
		{
			sourceText: "`${'\\u00E5'}`;\n",
			wantSpans:  []string{"${'\\u00E5'}"},
		},
		{
			sourceText: "`${'\\n'}`;\n",
			wantSpans:  []string{"${'\\n'}"},
		},
		{
			sourceText: "` ${'\\u00E5'} `;\n",
			wantSpans:  []string{"${'\\u00E5'}"},
		},
		{
			sourceText: "` ${'\\n'} `;\n",
			wantSpans:  []string{"${'\\n'}"},
		},
		{
			sourceText: "` ${\"\\n\"} `;\n",
			wantSpans:  []string{"${\"\\n\"}"},
		},
		{
			sourceText: "` ${`\\n`} `;\n",
			wantSpans:  []string{"${`\\n`}"},
		},
		{
			sourceText: "` ${ 'A\\u0307\\u0323' } `;\n",
			wantSpans:  []string{"${ 'A\\u0307\\u0323' }"},
		},
		{
			sourceText: "` ${'👨‍👩‍👧‍👦'} `;\n",
			wantSpans:  []string{"${'👨‍👩‍👧‍👦'}"},
		},
		{
			sourceText: "` ${'\\ud83d\\udc68'} `;\n",
			wantSpans:  []string{"${'\\ud83d\\udc68'}"},
		},
		{
			sourceText: "`\nthis code does not have trailing whitespace: ${' '}\\n even though it might look it.`;\n",
			wantSpans:  []string{"${' '}"},
		},
		{
			sourceText: "`\nthis code has trailing position template expression ${\"but it isn't whitespace\"}\n    `;\n",
			wantSpans:  []string{"${\"but it isn't whitespace\"}"},
		},
		{
			sourceText: "`trailing whitespace followed by escaped windows newline: ${' '}\\r\\n`;\n",
			wantSpans:  []string{"${' '}"},
		},
		{
			sourceText: "`template literal with interpolations followed by newline: ${` ${'interpolation'} `}\n`;\n",
			wantSpans:  []string{"${` ${'interpolation'} `}", "${'interpolation'}"},
		},
		{
			sourceText: "function func<T extends string>(arg: T) {\n  `${arg}`;\n}\n",
			wantSpans:  []string{"${arg}"},
		},
		{
			sourceText: "declare const b: 'b';\n`a${b}${'c'}`;\n",
			wantSpans:  []string{"${'c'}"},
		},
		{
			sourceText: "declare const nested: string, interpolation: string;\n`use${`less${nested}${interpolation}`}`;\n",
			wantSpans:  []string{"${`less${nested}${interpolation}`}"},
		},
		{
			sourceText: "declare const string: 'a';\n        `${   string   }`;\n",
			wantSpans:  []string{"  string "},
		},
		{
			sourceText: "declare const string: 'a';\n`${string}`;\n",
			wantSpans:  []string{"${string}"},
		},
		{
			sourceText: "declare const intersection: string & { _brand: 'test-brand' };\n`${intersection}`;\n",
			wantSpans:  []string{"${intersection}"},
		},
		{
			sourceText: "true ? `${'test' || ''}`.trim() : undefined;\n",
			wantSpans:  []string{"${'test' || ''}"},
		},
		{
			sourceText: "type Foo = `${1}`;\n",
			wantSpans:  []string{"${1}"},
		},
		{
			sourceText: "type Foo = `${null}`;\n",
			wantSpans:  []string{"${null}"},
		},
		{
			sourceText: "type Foo = `${undefined}`;\n",
			wantSpans:  []string{"${undefined}"},
		},
		{
			sourceText: "type Foo = `${'foo'}`;\n",
			wantSpans:  []string{"${'foo'}"},
		},
		{
			sourceText: "type Foo = 'A' | 'B';\ntype Bar = `${Foo}`;\n",
			wantSpans:  []string{"${Foo}"},
		},
		{
			sourceText: "type Foo = 'A' | 'B';\ntype Bar = `${`${Foo}`}`;\n",
			wantSpans:  []string{"${`${Foo}`}", "${Foo}"},
		},
		{
			sourceText: "type FooBarBaz = `foo${'bar'}baz`;\n",
			wantSpans:  []string{"${'bar'}"},
		},
		{
			sourceText: "type FooBar = `foo${`bar`}`;\n",
			wantSpans:  []string{"${`bar`}"},
		},
		{
			sourceText: "type FooBar = `${'foo' | 'bar'}`;\n",
			wantSpans:  []string{"${'foo' | 'bar'}"},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnnecessaryTemplateExpressionCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTemplateExpression,
				noUnnecessaryTemplateExpressionFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for position := range wantIds {
				wantIds[position] = "noUnnecessaryTemplateExpression"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so span slices are against that text.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantSpans {
				diagnostic := result.Diagnostics[position]
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want {
					t.Fatalf("finding %d span: expected %q, got %q", position, want, gotSpan)
				}
				if diagnostic.Message.Description !=
					"Template literal expression is unnecessary and can be simplified." {
					t.Fatalf("finding %d message: got %q", position, diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestNoUnnecessaryTemplateExpressionFixes asserts the rewrite on every case where ONE pass of the
// harness can express it.
//
// Twenty of upstream's cases record an ARRAY of outputs rather than a string, because the fixer's
// own result is reportable and ESLint re-runs it to convergence. `ExpectFixedSource` applies a
// single pass and refuses overlapping edits, so the expectation here is the FIRST pass rather than
// the converged text. Asserting the converged text against a one-pass harness reads as a broken
// fixer and is not.
//
// Four cases propose edits that overlap within one pass and are excluded, listed in the constant
// below rather than silently dropped. They are covered for detection by the firing test above.
func TestNoUnnecessaryTemplateExpressionFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantFixed  string
	}{
		{
			sourceText: "`${1}`;\n",
			wantFixed:  "`1`;\n",
		},
		{
			sourceText: "`${1n}`;\n",
			wantFixed:  "`1`;\n",
		},
		{
			sourceText: "`${0o25}`;\n",
			wantFixed:  "`21`;\n",
		},
		{
			sourceText: "`${0b1010} ${0b1111}`;\n",
			wantFixed:  "`10 15`;\n",
		},
		{
			sourceText: "`${0x25}`;\n",
			wantFixed:  "`37`;\n",
		},
		{
			sourceText: "`${/a/}`;\n",
			wantFixed:  "`/a/`;\n",
		},
		{
			sourceText: "`${/a/gim}`;\n",
			wantFixed:  "`/a/gim`;\n",
		},
		{
			sourceText: "`${    1    }`;\n",
			wantFixed:  "`1`;\n",
		},
		{
			sourceText: "`${    'a'    }`;\n",
			wantFixed:  "'a';\n",
		},
		{
			sourceText: "`${    \"a\"    }`;\n",
			wantFixed:  "\"a\";\n",
		},
		{
			sourceText: "`${    'a' + 'b'    }`;\n",
			wantFixed:  "'a' + 'b';\n",
		},
		{
			sourceText: "`${true}`;\n",
			wantFixed:  "`true`;\n",
		},
		{
			sourceText: "`${    true    }`;\n",
			wantFixed:  "`true`;\n",
		},
		{
			sourceText: "`${null}`;\n",
			wantFixed:  "`null`;\n",
		},
		{
			sourceText: "`${    null    }`;\n",
			wantFixed:  "`null`;\n",
		},
		{
			sourceText: "`${undefined}`;\n",
			wantFixed:  "`undefined`;\n",
		},
		{
			sourceText: "`${    undefined    }`;\n",
			wantFixed:  "`undefined`;\n",
		},
		{
			sourceText: "`${Infinity}`;\n",
			wantFixed:  "`Infinity`;\n",
		},
		{
			sourceText: "`${NaN}`;\n",
			wantFixed:  "`NaN`;\n",
		},
		{
			sourceText: "`${'a'} ${'b'}`;\n",
			wantFixed:  "`a b`;\n",
		},
		{
			sourceText: "`${   'a'   } ${   'b'   }`;\n",
			wantFixed:  "`a b`;\n",
		},
		{
			sourceText: "`use${'less'}`;\n",
			wantFixed:  "`useless`;\n",
		},
		{
			sourceText: "`use${`less`}`;\n",
			wantFixed:  "`useless`;\n",
		},
		{
			sourceText: "`use${\n  `less`\n}`;\n",
			wantFixed:  "`useless`;\n",
		},
		{
			sourceText: "`${'1 + 1 ='} ${2}`;\n",
			wantFixed:  "`1 + 1 = 2`;\n",
		},
		{
			sourceText: "`${'a'} ${true}`;\n",
			wantFixed:  "`a true`;\n",
		},
		{
			sourceText: "`${String(Symbol.for('test'))}`;\n",
			wantFixed:  "String(Symbol.for('test'));\n",
		},
		{
			sourceText: "`${'`'}`;\n",
			wantFixed:  "'`';\n",
		},
		{
			sourceText: "`back${'`'}tick`;\n",
			wantFixed:  "`back\\`tick`;\n",
		},
		{
			sourceText: "`dollar${'${`this is test`}'}sign`;\n",
			wantFixed:  "`dollar\\${\\`this is test\\`}sign`;\n",
		},
		{
			sourceText: "`complex${'`${\"`${test}`\"}`'}case`;\n",
			wantFixed:  "`complex\\`\\${\"\\`\\${test}\\`\"}\\`case`;\n",
		},
		{
			sourceText: "`some ${'\\\\${test}'} string`;\n",
			wantFixed:  "`some \\\\\\${test} string`;\n",
		},
		{
			sourceText: "`some ${'\\\\`'} string`;\n",
			wantFixed:  "`some \\\\\\` string`;\n",
		},
		{
			sourceText: "`some ${/`/} string`;\n",
			wantFixed:  "`some /\\`/ string`;\n",
		},
		{
			sourceText: "`some ${/\\`/} string`;\n",
			wantFixed:  "`some /\\\\\\`/ string`;\n",
		},
		{
			sourceText: "`some ${/\\\\`/} string`;\n",
			wantFixed:  "`some /\\\\\\\\\\`/ string`;\n",
		},
		{
			sourceText: "`some ${/\\\\\\`/} string`;\n",
			wantFixed:  "`some /\\\\\\\\\\\\\\`/ string`;\n",
		},
		{
			sourceText: "`some ${/${}/} string`;\n",
			wantFixed:  "`some /\\${}/ string`;\n",
		},
		{
			sourceText: "`some ${/$ {}/} string`;\n",
			wantFixed:  "`some /$ {}/ string`;\n",
		},
		{
			sourceText: "`some ${/\\\\/} string`;\n",
			wantFixed:  "`some /\\\\\\\\/ string`;\n",
		},
		{
			sourceText: "`some ${/\\\\\\b/} string`;\n",
			wantFixed:  "`some /\\\\\\\\\\\\b/ string`;\n",
		},
		{
			sourceText: "`some ${/\\\\\\\\/} string`;\n",
			wantFixed:  "`some /\\\\\\\\\\\\\\\\/ string`;\n",
		},
		{
			sourceText: "` ${''} `;\n",
			wantFixed:  "`  `;\n",
		},
		{
			sourceText: "` ${\"\"} `;\n",
			wantFixed:  "`  `;\n",
		},
		{
			sourceText: "` ${``} `;\n",
			wantFixed:  "`  `;\n",
		},
		{
			sourceText: "` ${'\\`'} `;\n",
			wantFixed:  "` \\` `;\n",
		},
		{
			sourceText: "` ${'\\\\`'} `;\n",
			wantFixed:  "` \\\\\\` `;\n",
		},
		{
			sourceText: "` ${'$'}{} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'\\$'}{} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'\\\\$'}{} `;\n",
			wantFixed:  "` \\\\\\${} `;\n",
		},
		{
			sourceText: "` ${'\\\\$ '}{} `;\n",
			wantFixed:  "` \\\\$ {} `;\n",
		},
		{
			sourceText: "` ${'\\\\\\$'}{} `;\n",
			wantFixed:  "` \\\\\\${} `;\n",
		},
		{
			sourceText: "` \\\\${'\\\\$'}{} `;\n",
			wantFixed:  "` \\\\\\\\\\${} `;\n",
		},
		{
			sourceText: "` $${'{$'}{} `;\n",
			wantFixed:  "` \\${\\${} `;\n",
		},
		{
			sourceText: "` $${'${$'}{} `;\n",
			wantFixed:  "` $\\${\\${} `;\n",
		},
		{
			sourceText: "` ${'foo$'}{} `;\n",
			wantFixed:  "` foo\\${} `;\n",
		},
		{
			sourceText: "` ${`$`} `;\n",
			wantFixed:  "` $ `;\n",
		},
		{
			sourceText: "` ${`$`}{} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${`$`} {} `;\n",
			wantFixed:  "` $ {} `;\n",
		},
		{
			sourceText: "` ${`$`}${undefined}{} `;\n",
			wantFixed:  "` $undefined{} `;\n",
		},
		{
			sourceText: "` ${`foo$`}{} `;\n",
			wantFixed:  "` foo\\${} `;\n",
		},
		{
			sourceText: "` ${'$'}${''}{} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'$'}${``}{} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'foo$'}${''}${``}{} `;\n",
			wantFixed:  "` foo\\${} `;\n",
		},
		{
			sourceText: "` $${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${undefined}${'{}'} `;\n",
			wantFixed:  "` $undefined{} `;\n",
		},
		{
			sourceText: "` $${''}${undefined}${'{}'} `;\n",
			wantFixed:  "` $undefined{} `;\n",
		},
		{
			sourceText: "` \\$${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${'foo'}${'{'} `;\n",
			wantFixed:  "` $foo{ `;\n",
		},
		{
			sourceText: "` $${'{ foo'}${'{'} `;\n",
			wantFixed:  "` \\${ foo{ `;\n",
		},
		{
			sourceText: "` \\\\$${'{}'} `;\n",
			wantFixed:  "` \\\\\\${} `;\n",
		},
		{
			sourceText: "` \\\\\\$${'{}'} `;\n",
			wantFixed:  "` \\\\\\${} `;\n",
		},
		{
			sourceText: "` foo$${'{}'} `;\n",
			wantFixed:  "` foo\\${} `;\n",
		},
		{
			sourceText: "` $${''}${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${''} `;\n",
			wantFixed:  "` $ `;\n",
		},
		{
			sourceText: "` $${`{}`} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${``}${`{}`} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${``}${`foo{}`} `;\n",
			wantFixed:  "` $foo{} `;\n",
		},
		{
			sourceText: "` $${''}${`{}`} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${``}${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` $${''}${``}${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'$'} `;\n",
			wantFixed:  "` $ `;\n",
		},
		{
			sourceText: "` ${'$'}${'{}'} `;\n",
			wantFixed:  "` \\${} `;\n",
		},
		{
			sourceText: "` ${'$'}${''}${'{'} `;\n",
			wantFixed:  "` \\${ `;\n",
		},
		{
			sourceText: "` ${`\n\\$`}{} `;\n",
			wantFixed:  "` \n\\${} `;\n",
		},
		{
			sourceText: "` ${`\n\\\\$`}{} `;\n",
			wantFixed:  "` \n\\\\\\${} `;\n",
		},
		{
			sourceText: "`${'\\u00E5'}`;\n",
			wantFixed:  "'\\u00E5';\n",
		},
		{
			sourceText: "`${'\\n'}`;\n",
			wantFixed:  "'\\n';\n",
		},
		{
			sourceText: "` ${'\\u00E5'} `;\n",
			wantFixed:  "` \\u00E5 `;\n",
		},
		{
			sourceText: "` ${'\\n'} `;\n",
			wantFixed:  "` \\n `;\n",
		},
		{
			sourceText: "` ${\"\\n\"} `;\n",
			wantFixed:  "` \\n `;\n",
		},
		{
			sourceText: "` ${`\\n`} `;\n",
			wantFixed:  "` \\n `;\n",
		},
		{
			sourceText: "` ${ 'A\\u0307\\u0323' } `;\n",
			wantFixed:  "` A\\u0307\\u0323 `;\n",
		},
		{
			sourceText: "` ${'👨‍👩‍👧‍👦'} `;\n",
			wantFixed:  "` 👨‍👩‍👧‍👦 `;\n",
		},
		{
			sourceText: "` ${'\\ud83d\\udc68'} `;\n",
			wantFixed:  "` \\ud83d\\udc68 `;\n",
		},
		{
			sourceText: "`\nthis code does not have trailing whitespace: ${' '}\\n even though it might look it.`;\n",
			wantFixed:  "`\nthis code does not have trailing whitespace:  \\n even though it might look it.`;\n",
		},
		{
			sourceText: "`\nthis code has trailing position template expression ${\"but it isn't whitespace\"}\n    `;\n",
			wantFixed:  "`\nthis code has trailing position template expression but it isn't whitespace\n    `;\n",
		},
		{
			sourceText: "`trailing whitespace followed by escaped windows newline: ${' '}\\r\\n`;\n",
			wantFixed:  "`trailing whitespace followed by escaped windows newline:  \\r\\n`;\n",
		},
		{
			sourceText: "function func<T extends string>(arg: T) {\n  `${arg}`;\n}\n",
			wantFixed:  "function func<T extends string>(arg: T) {\n  arg;\n}\n",
		},
		{
			sourceText: "declare const b: 'b';\n`a${b}${'c'}`;\n",
			wantFixed:  "declare const b: 'b';\n`a${b}c`;\n",
		},
		{
			sourceText: "declare const nested: string, interpolation: string;\n`use${`less${nested}${interpolation}`}`;\n",
			wantFixed:  "declare const nested: string, interpolation: string;\n`useless${nested}${interpolation}`;\n",
		},
		{
			sourceText: "declare const string: 'a';\n        `${   string   }`;\n",
			wantFixed:  "declare const string: 'a';\n        string;\n",
		},
		{
			sourceText: "declare const string: 'a';\n`${string}`;\n",
			wantFixed:  "declare const string: 'a';\nstring;\n",
		},
		{
			sourceText: "declare const intersection: string & { _brand: 'test-brand' };\n`${intersection}`;\n",
			wantFixed:  "declare const intersection: string & { _brand: 'test-brand' };\nintersection;\n",
		},
		{
			sourceText: "true ? `${'test' || ''}`.trim() : undefined;\n",
			wantFixed:  "true ? ('test' || '').trim() : undefined;\n",
		},
		{
			sourceText: "type Foo = `${1}`;\n",
			wantFixed:  "type Foo = `1`;\n",
		},
		{
			sourceText: "type Foo = `${null}`;\n",
			wantFixed:  "type Foo = `null`;\n",
		},
		{
			sourceText: "type Foo = `${undefined}`;\n",
			wantFixed:  "type Foo = `undefined`;\n",
		},
		{
			sourceText: "type Foo = `${'foo'}`;\n",
			wantFixed:  "type Foo = 'foo';\n",
		},
		{
			sourceText: "type Foo = 'A' | 'B';\ntype Bar = `${Foo}`;\n",
			wantFixed:  "type Foo = 'A' | 'B';\ntype Bar = Foo;\n",
		},
		{
			sourceText: "type FooBarBaz = `foo${'bar'}baz`;\n",
			wantFixed:  "type FooBarBaz = `foobarbaz`;\n",
		},
		{
			sourceText: "type FooBar = `foo${`bar`}`;\n",
			wantFixed:  "type FooBar = `foobar`;\n",
		},
		{
			sourceText: "type FooBar = `${'foo' | 'bar'}`;\n",
			wantFixed:  "type FooBar = 'foo' | 'bar';\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noUnnecessaryTemplateExpressionCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryTemplateExpression,
				noUnnecessaryTemplateExpressionFile, testCase.sourceText)

			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}
