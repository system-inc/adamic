package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const restrictTemplateExpressionsFile = "/repository/source/Templating.ts"

func restrictTemplateExpressionsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// restrictTemplateExpressionsSettingsFor routes an option object through the rule's own decoder.
//
// Building the struct directly would leave the two lines most likely to be wrong untested: five of
// the seven flags default to TRUE, so a decoder reading an absent key as false silently turns an
// unconfigured rule into upstream's strict preset, and the `allow` default is upstream's three
// standard-library types rather than an empty list.
//
// The empty string is the shape a bare `"error"` delivers: no bytes at all.
func restrictTemplateExpressionsSettingsFor(t *testing.T, wire string) any {
	t.Helper()
	decoded, err := DecodeRestrictTemplateExpressionsOptions([]byte(wire))
	if err != nil {
		t.Fatalf("decoding %q: %v", wire, err)
	}
	return decoded
}

// TestRestrictTemplateExpressionsStaysSilent is upstream's sixty passing cases verbatim, across
// every option setting its corpus exercises.
func TestRestrictTemplateExpressionsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wire       string
	}{
		{sourceText: "\nconst msg = `arg = ${'foo'}`;\n    ", wire: ""},
		{sourceText: "\nconst arg = 'foo';\nconst msg = `arg = ${arg}`;\n    ", wire: ""},
		{sourceText: "\nconst arg = 'foo';\nconst msg = `arg = ${arg || 'default'}`;\n    ", wire: ""},
		{sourceText: "\nfunction test<T extends string>(arg: T) {\n  return `arg = ${arg}`;\n}\n    ", wire: ""},
		{sourceText: "\nfunction test<T extends string & { _kind: 'MyBrandedString' }>(arg: T) {\n  return `arg = ${arg}`;\n}\n    ", wire: ""},
		{sourceText: "\ntag`arg = ${null}`;\n    ", wire: ""},
		{sourceText: "\nconst arg = {};\ntag`arg = ${arg}`;\n    ", wire: ""},
		{sourceText: "\nconst arg = 123;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nconst arg = 123;\nconst msg = `arg = ${arg || 'default'}`;\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nconst arg = 123n;\nconst msg = `arg = ${arg || 'default'}`;\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nfunction test<T extends number>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nfunction test<T extends number & { _kind: 'MyBrandedNumber' }>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nfunction test<T extends bigint>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nfunction test<T extends string | number>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNumber\": true}"},
		{sourceText: "\nconst arg = true;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowBoolean\": true}"},
		{sourceText: "\nconst arg = true;\nconst msg = `arg = ${arg || 'default'}`;\n      ", wire: "{\"allowBoolean\": true}"},
		{sourceText: "\nfunction test<T extends boolean>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowBoolean\": true}"},
		{sourceText: "\nfunction test<T extends string | boolean>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowBoolean\": true}"},
		{sourceText: "\nconst arg = [];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\nconst arg = [];\nconst msg = `arg = ${arg || 'default'}`;\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\nfunction test<T extends string[]>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\ndeclare const arg: [number, string];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\nconst arg = [1, 'a'] as const;\nconst msg = `arg = ${arg || 'default'}`;\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\nfunction test<T extends [string, string]>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\ndeclare const arg: [number | undefined, string];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowArray\": true, \"allowNullish\": true}"},
		{sourceText: "\ndeclare const arg: string[][];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowArray\": true}"},
		{sourceText: "\ndeclare const arg: never[];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowArray\": true, \"allowNever\": true}"},
		{sourceText: "\ndeclare const arg: any[];\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowAny\": true, \"allowArray\": true}"},
		{sourceText: "\nconst arg: any = 123;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowAny\": true}"},
		{sourceText: "\nconst arg: any = undefined;\nconst msg = `arg = ${arg || 'some-default'}`;\n      ", wire: "{\"allowAny\": true}"},
		{sourceText: "\nconst user = JSON.parse('{ \"name\": \"foo\" }');\nconst msg = `arg = ${user.name}`;\n      ", wire: "{\"allowAny\": true}"},
		{sourceText: "\nconst user = JSON.parse('{ \"name\": \"foo\" }');\nconst msg = `arg = ${user.name || 'the user with no name'}`;\n      ", wire: "{\"allowAny\": true}"},
		{sourceText: "\nconst arg = null;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowNullish\": true}"},
		{sourceText: "\ndeclare const arg: string | null | undefined;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowNullish\": true}"},
		{sourceText: "\nfunction test<T extends null | undefined>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNullish\": true}"},
		{sourceText: "\nfunction test<T extends string | null>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowNullish\": true}"},
		{sourceText: "\nconst arg = new RegExp('foo');\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowRegExp\": true}"},
		{sourceText: "\nconst arg = /foo/;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowRegExp\": true}"},
		{sourceText: "\ndeclare const arg: string | RegExp;\nconst msg = `arg = ${arg}`;\n      ", wire: "{\"allowRegExp\": true}"},
		{sourceText: "\nfunction test<T extends RegExp>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowRegExp\": true}"},
		{sourceText: "\nfunction test<T extends string | RegExp>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowRegExp\": true}"},
		{sourceText: "\ndeclare const value: never;\nconst stringy = `${value}`;\n      ", wire: "{\"allowNever\": true}"},
		{sourceText: "\nconst arg = 'hello';\nconst msg = typeof arg === 'string' ? arg : `arg = ${arg}`;\n      ", wire: "{\"allowNever\": true}"},
		{sourceText: "\nfunction test(arg: 'one' | 'two') {\n  switch (arg) {\n    case 'one':\n      return 1;\n    case 'two':\n      return 2;\n    default:\n      throw new Error(`Unrecognised arg: ${arg}`);\n  }\n}\n      ", wire: "{\"allowNever\": true}"},
		{sourceText: "\n// more variants may be added to Foo in the future\ntype Foo = { type: 'a'; value: number };\n\nfunction checkFoosAreMatching(foo1: Foo, foo2: Foo) {\n  if (foo1.type !== foo2.type) {\n    // since Foo currently only has one variant, this code is never run, and `foo1.type` has type `never`.\n    throw new Error(`expected ${foo1.type}, found ${foo2.type}`);\n  }\n}\n      ", wire: "{\"allowNever\": true}"},
		{sourceText: "\ntype All = string | number | boolean | null | undefined | RegExp | never;\nfunction test<T extends All>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ", wire: "{\"allowBoolean\": true, \"allowNever\": true, \"allowNullish\": true, \"allowNumber\": true, \"allowRegExp\": true}"},
		{sourceText: "const msg = `arg = ${Promise.resolve()}`;", wire: "{\"allow\": [{\"from\": \"lib\", \"name\": \"Promise\"}]}"},
		{sourceText: "const msg = `arg = ${new Error()}`;", wire: ""},
		{sourceText: "const msg = `arg = ${false}`;", wire: ""},
		{sourceText: "const msg = `arg = ${null}`;", wire: ""},
		{sourceText: "const msg = `arg = ${undefined}`;", wire: ""},
		{sourceText: "const msg = `arg = ${123}`;", wire: ""},
		{sourceText: "const msg = `arg = ${'abc'}`;", wire: ""},
		{sourceText: "\nclass Base {}\nclass Derived extends Base {}\nconst foo = new Base();\nconst bar = new Derived();\n`${foo}${bar}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Base\"}]}"},
		{sourceText: "\nclass Base {}\nclass Derived extends Base {}\nclass DerivedTwice extends Derived {}\nconst value = new DerivedTwice();\n`${value}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Base\"}]}"},
		{sourceText: "\ninterface Base {\n  value: string;\n}\ninterface Derived extends Base {\n  extra: number;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Base\"}]}"},
		{sourceText: "\ninterface Base {\n  value: string;\n}\ninterface Other {\n  other: number;\n}\ninterface Derived extends Base, Other {\n  extra: boolean;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Base\"}]}"},
		{sourceText: "\ninterface Base {\n  value: string;\n}\ninterface Other {\n  other: number;\n}\ninterface Derived extends Base, Other {\n  extra: boolean;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Other\"}]}"},
		{sourceText: "\ninterface Root {\n  root: string;\n}\ninterface Another {\n  another: string;\n}\ninterface Base extends Root, Another {\n  value: string;\n}\ninterface Other {\n  other: number;\n}\ninterface Derived extends Base, Other {\n  extra: boolean;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Another\"}]}"},
		{sourceText: "\ntype Custom = { value: string };\ndeclare const obj: Custom;\n`${obj}`;\n      ", wire: "{\"allow\": [{\"from\": \"file\", \"name\": \"Custom\"}]}"},

		// A GENERIC class or interface extending an allowed base. The base types live on the type
		// reference's TARGET rather than on the reference itself, so without that hop the allowlist
		// never reaches `Base` and both of these report. Measured silent on the installed build,
		// and a mutant removing the hop reports both while passing all eighty six imported cases:
		// upstream's transitive-base cases are all non-generic, so nothing imported can see it.
		{
			sourceText: "class Base {}\nclass Derived<T> extends Base {}\ndeclare const d: Derived<string>;\n`${d}`;\n",
			wire:       `{"allow":[{"from":"file","name":"Base"}]}`,
		},
		{
			sourceText: "interface Base {}\ninterface Derived<T> extends Base {}\ndeclare const d: Derived<string>;\n`${d}`;\n",
			wire:       `{"allow":[{"from":"file","name":"Base"}]}`,
		},
	}
	for index, testCase := range cases {
		t.Run(restrictTemplateExpressionsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t,
				RestrictTemplateExpressions, restrictTemplateExpressionsFile, testCase.sourceText,
				restrictTemplateExpressionsSettingsFor(t, testCase.wire)))
		})
	}
}

// TestRestrictTemplateExpressionsFires is upstream's twenty six reporting cases verbatim, with the
// message id, the rendered message text, and the span of every finding.
//
// The text is asserted by equality because the message interpolates a type name through the
// checker, and that name is the only part of the output a reader uses to understand the finding.
func TestRestrictTemplateExpressionsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wire         string
		wantFindings []struct {
			id      string
			message string
			span    string
		}
	}{
		{
			sourceText: "\nconst msg = `arg = ${123}`;\n      ",
			wire:       "{\"allowNumber\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"123\" of template literal expression.",
					span:    "123",
				},
			},
		},
		{
			sourceText: "\nconst msg = `arg = ${false}`;\n      ",
			wire:       "{\"allowBoolean\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"false\" of template literal expression.",
					span:    "false",
				},
			},
		},
		{
			sourceText: "\nconst msg = `arg = ${null}`;\n      ",
			wire:       "{\"allowNullish\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"null\" of template literal expression.",
					span:    "null",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: number[];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true, \"allowNumber\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"number[]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nconst msg = `arg = ${[, 2]}`;\n      ",
			wire:       "{\"allowArray\": true, \"allowNullish\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"(number | undefined)[]\" of template literal expression.",
					span:    "[, 2]",
				},
			},
		},
		{
			sourceText: "const msg = `arg = ${Promise.resolve()}`;",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"Promise<void>\" of template literal expression.",
					span:    "Promise.resolve()",
				},
			},
		},
		{
			sourceText: "const msg = `arg = ${new Error()}`;",
			wire:       "{\"allow\": []}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"Error\" of template literal expression.",
					span:    "new Error()",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: [number | undefined, string];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true, \"allowNullish\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"[number | undefined, string]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: object[];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"object[]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: (object | string)[];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"(string | object)[]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: object[][];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"object[][]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: (object | string)[][];\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowArray\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"(string | object)[][]\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: number;\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowNumber\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"number\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: boolean;\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowBoolean\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"boolean\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nconst arg = {};\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowBoolean\": true, \"allowNullish\": true, \"allowNumber\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"{}\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const arg: { a: string } & { b: string };\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"{ a: string; } & { b: string; }\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nfunction test<T extends {}>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ",
			wire:       "{\"allowBoolean\": true, \"allowNullish\": true, \"allowNumber\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"{}\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nfunction test<TWithNoConstraint>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ",
			wire:       "{\"allowAny\": false, \"allowBoolean\": true, \"allowNullish\": true, \"allowNumber\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"T\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nfunction test(arg: any) {\n  return `arg = ${arg}`;\n}\n      ",
			wire:       "{\"allowAny\": false, \"allowBoolean\": true, \"allowNullish\": true, \"allowNumber\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"any\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nconst arg = new RegExp('foo');\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowRegExp\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"RegExp\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nconst arg = /foo/;\nconst msg = `arg = ${arg}`;\n      ",
			wire:       "{\"allowRegExp\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"RegExp\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\ndeclare const value: never;\nconst stringy = `${value}`;\n      ",
			wire:       "{\"allowNever\": false}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"never\" of template literal expression.",
					span:    "value",
				},
			},
		},
		{
			sourceText: "\nfunction test<T extends any>(arg: T) {\n  return `arg = ${arg}`;\n}\n      ",
			wire:       "{\"allowAny\": true}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"unknown\" of template literal expression.",
					span:    "arg",
				},
			},
		},
		{
			sourceText: "\nclass Base {}\nclass Derived extends Base {}\nconst bar = new Derived();\n`${bar}`;\n      ",
			wire:       "{\"allow\": []}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"Derived\" of template literal expression.",
					span:    "bar",
				},
			},
		},
		{
			sourceText: "\ninterface Base {\n  value: string;\n}\ninterface Derived extends Base {\n  extra: number;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ",
			wire:       "{\"allow\": []}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"Derived\" of template literal expression.",
					span:    "obj",
				},
			},
		},
		{
			sourceText: "\ninterface Base {\n  value: string;\n}\ninterface Other {\n  other: number;\n}\ninterface Derived extends Base, Other {\n  extra: boolean;\n}\ndeclare const obj: Derived;\n`${obj}`;\n      ",
			wire:       "{\"allow\": []}",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"Derived\" of template literal expression.",
					span:    "obj",
				},
			},
		},

		// An array under the DEFAULT options, where allowArray is one of only two flags that default to
		// FALSE. Every one of upstream's eleven array cases supplies allowArray explicitly, so nothing in
		// the imported corpus exercises the default, and a mutant flipping it to true survived the whole
		// suite. These two are the pair that holds it: without them the rule could ship allowing arrays
		// and every fixture would still pass.
		{
			sourceText: "declare const a: string[];\n`${a}`;\n",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"string[]\" of template literal expression.",
					span:    "a",
				},
			},
		},
		// The same under a number element type, so the finding is about the array rather than about what
		// is in it: allowNumber defaults ON and does not rescue this.
		{
			sourceText: "declare const a: number[];\n`${a}`;\n",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: "Invalid type \"number[]\" of template literal expression.",
					span:    "a",
				},
			},
		},

		// `never` under the DEFAULT options, the other flag that defaults to FALSE. Upstream's four
		// never cases all supply allowNever explicitly, so a mutant flipping the default to true
		// survived every imported case.
		{
			sourceText: "declare const n: never;\n`${n}`;\n",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: `Invalid type "never" of template literal expression.`,
					span:    "n",
				},
			},
		},

		// The shape a `never` really arrives in: an exhaustiveness check whose fallthrough the
		// checker has narrowed to nothing. Worth having beside the declared one, because this is
		// what the finding looks like on real code and it is the case a reader will meet.
		{
			sourceText: "function f(x: string | number): string {\n  if (typeof x === 'string') return x;\n  if (typeof x === 'number') return String(x);\n  return `${x}`;\n}\n",
			wire:       "",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "invalidType",
					message: `Invalid type "never" of template literal expression.`,
					span:    "x",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(restrictTemplateExpressionsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
				restrictTemplateExpressionsFile, testCase.sourceText,
				restrictTemplateExpressionsSettingsFor(t, testCase.wire))

			wantIds := make([]string, 0, len(testCase.wantFindings))
			for _, want := range testCase.wantFindings {
				wantIds = append(wantIds, want.id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// RunTyped writes the fixture trimmed, so the span slice is against that text.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for findingIndex, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[findingIndex]
				if diagnostic.Message.Description != want.message {
					t.Fatalf("finding %d message: expected %q, got %q", findingIndex, want.message,
						diagnostic.Message.Description)
				}
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.span {
					t.Fatalf("finding %d span: expected %q, got %q", findingIndex, want.span, gotSpan)
				}
			}
		})
	}
}

// TestRestrictTemplateExpressionsReadsNilOptionsAsUpstreamsDefault pins the path a rule configured
// as a bare severity takes, which no fixture routed through the decoder can reach.
//
// This is the failure mode the brief describes as the expensive one: the config layer hands a
// non-required rule nil options, the type assertion in Run fails, and a fallback to the ZERO VALUE
// gives every flag its false meaning. For this rule that is not merely inert, it is upstream's
// `strict` preset, so a bare `"error"` would silently enforce a far stricter rule than upstream's
// default while every fixture supplying options stayed green. A mutant replacing the fallback with
// the zero value survives the whole eighty six case corpus.
//
// The three inputs below are the three defaults that separate the two: a number, a boolean and an
// interpolated `Error` are all allowed by default and all reported under the strict preset.
func TestRestrictTemplateExpressionsReadsNilOptionsAsUpstreamsDefault(t *testing.T) {
	t.Parallel()

	cases := []string{
		"declare const n: number;\n`${n}`;\n",
		"declare const b: boolean;\n`${b}`;\n",
		"declare const e: Error;\n`${e}`;\n",
	}
	for index, sourceText := range cases {
		t.Run(restrictTemplateExpressionsCaseName(index), func(t *testing.T) {
			t.Parallel()
			// nil rather than a decoded struct, which is what the config layer delivers.
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t,
				RestrictTemplateExpressions, restrictTemplateExpressionsFile, sourceText, nil))
		})
	}

	// The control: something no default allows still reports through the same path, so the silence
	// above is the defaults applying rather than the rule declining to run at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, "declare const o: object;\n`${o}`;\n", nil), "invalidType")
}

// TestRestrictTemplateExpressionsRequiresTheTypedHarness pins that the rule cannot work without a
// checker and declines rather than crashing when handed none.
//
// Unreachable from every other test here, because RunTyped always supplies a live checker.
func TestRestrictTemplateExpressionsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !RestrictTemplateExpressions.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: every judgment is about an interpolated value's type")
	}

	sourceText := "declare const o: object;\n`${o}`;\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, sourceText))

	// The control, so the silence above is the guard declining rather than the rule being blind to
	// this shape.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, sourceText, nil), "invalidType")
}

// TestRestrictTemplateExpressionsAllowReplacesTheDefault pins how a configured `allow` interacts
// with upstream's default list, which is a real decision and not a detail.
//
// Upstream's defaultOptions carries `allow: [{ name: ['Error', 'URL', 'URLSearchParams'], from:
// 'lib' }]`, and ESLint merges defaults key by key, so supplying `allow` REPLACES that list rather
// than adding to it while supplying any other key leaves it alone. Measured on the installed build,
// all three rows below, because the merge semantics are the sort of thing a port gets subtly wrong
// in a way no imported case catches: every one of upstream's allow cases supplies a list, so the
// interaction with the default is never exercised.
func TestRestrictTemplateExpressionsAllowReplacesTheDefault(t *testing.T) {
	t.Parallel()

	sourceText := "declare const e: Error;\n`${e}`;\n"

	// An allow list naming something else drops Error, so this reports.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, sourceText,
		restrictTemplateExpressionsSettingsFor(t, `{"allow":[{"from":"file","name":"Whatever"}]}`)),
		"invalidType")

	// An EMPTY allow list drops it too, which is the row that separates replace from merge: under
	// merge semantics an empty list would change nothing.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, sourceText,
		restrictTemplateExpressionsSettingsFor(t, `{"allow":[]}`)), "invalidType")

	// An options object with no allow key keeps the default, so this is silent.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RestrictTemplateExpressions,
		restrictTemplateExpressionsFile, sourceText,
		restrictTemplateExpressionsSettingsFor(t, `{"allowNumber":true}`)))
}
