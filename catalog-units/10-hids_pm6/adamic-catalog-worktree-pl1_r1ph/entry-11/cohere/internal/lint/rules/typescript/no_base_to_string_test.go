package typescript

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noBaseToStringFile names the fixture file.
const noBaseToStringFile = "/repository/source/Stringify.ts"

// noBaseToStringCaseName numbers a row so a failure names which one.
func noBaseToStringCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// asTheTypedNoBaseToStringHarnessWroteIt transforms a fixture the way RunTyped transforms its input.
//
// rule_testing/program.go writes each fixture as strings.TrimSpace(contents)+"\n", so the file on
// disk is offset from the string in the Go literal above it, and a span sliced from the literal is
// shifted at both ends.
func asTheTypedNoBaseToStringHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// decodeNoBaseToStringOptionsForTest routes a fixture through the rule's own decoder.
//
// This rule has the one option surface in the batch where the default is not a zero value:
// ignoredTypeNames defaults to four builtin type names, so a decoder that returned an empty slice
// for an absent key would silently start reporting on Error, RegExp, URL and URLSearchParams. Only
// a fixture routed through the decoder can see that.
func decodeNoBaseToStringOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoBaseToStringOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// runNoBaseToString runs one case, with or without options.
func runNoBaseToString(t *testing.T, sourceText string, options string) rule_testing.Result {
	t.Helper()
	if options == "" {
		return rule_testing.RunTyped(t, NoBaseToString, noBaseToStringFile, sourceText)
	}
	return rule_testing.RunTypedWithOptions(t, NoBaseToString, noBaseToStringFile, sourceText,
		decodeNoBaseToStringOptionsForTest(t, options))
}

// TestNoBaseToStringStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All two hundred and twenty-three of upstream's passing inputs, extracted from the clone's test
// file by parsing it with the TypeScript compiler rather than by reading it, then byte verified.
// Every one was replayed through the installed 8.x build with a real type checker, one program per
// case, and all two hundred and twenty-three reported nothing.
//
// This list is where the rule's difficulty lives. Deciding that a value stringifies usefully means
// walking its type: through unions and intersections, into tuple elements and array element types,
// down a type parameter to its constraint, and finally asking whether any of three coercion methods
// is declared somewhere other than on Object itself. Every one of those steps has a passing case
// here that a wrong answer would turn into a false positive on ordinary code.
func TestNoBaseToStringStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{sourceText: "`${''}`;", options: ""},
		{sourceText: "`${'text'}`;", options: ""},
		{sourceText: "`${true}`;", options: ""},
		{sourceText: "`${false}`;", options: ""},
		{sourceText: "`${1}`;", options: ""},
		{sourceText: "`${1n}`;", options: ""},
		{sourceText: "`${[]}`;", options: ""},
		{sourceText: "`${/regex/}`;", options: ""},
		{sourceText: "`${__dirname === 'foobar'}`;", options: ""},
		{sourceText: "`${{}.constructor()}`;", options: ""},
		{sourceText: "`${() => {}}`;", options: ""},
		{sourceText: "`${function () {}}`;", options: ""},
		{sourceText: "'' + 'text';", options: ""},
		{sourceText: "'' + true;", options: ""},
		{sourceText: "'' + false;", options: ""},
		{sourceText: "'' + 1;", options: ""},
		{sourceText: "'' + 1n;", options: ""},
		{sourceText: "'' + [];", options: ""},
		{sourceText: "'' + /regex/;", options: ""},
		{sourceText: "'' + (__dirname === 'foobar');", options: ""},
		{sourceText: "'' + {}.constructor();", options: ""},
		{sourceText: "'' + (() => {});", options: ""},
		{sourceText: "'' + function () {};", options: ""},
		{sourceText: "'text' + true;", options: ""},
		{sourceText: "'text' + false;", options: ""},
		{sourceText: "'text' + 1;", options: ""},
		{sourceText: "'text' + 1n;", options: ""},
		{sourceText: "'text' + [];", options: ""},
		{sourceText: "'text' + /regex/;", options: ""},
		{sourceText: "'text' + (__dirname === 'foobar');", options: ""},
		{sourceText: "'text' + {}.constructor();", options: ""},
		{sourceText: "'text' + (() => {});", options: ""},
		{sourceText: "'text' + function () {};", options: ""},
		{sourceText: "true + false;", options: ""},
		{sourceText: "true + 1;", options: ""},
		{sourceText: "true + 1n;", options: ""},
		{sourceText: "true + [];", options: ""},
		{sourceText: "true + /regex/;", options: ""},
		{sourceText: "true + (__dirname === 'foobar');", options: ""},
		{sourceText: "true + {}.constructor();", options: ""},
		{sourceText: "true + (() => {});", options: ""},
		{sourceText: "true + function () {};", options: ""},
		{sourceText: "false + 1;", options: ""},
		{sourceText: "false + 1n;", options: ""},
		{sourceText: "false + [];", options: ""},
		{sourceText: "false + /regex/;", options: ""},
		{sourceText: "false + (__dirname === 'foobar');", options: ""},
		{sourceText: "false + {}.constructor();", options: ""},
		{sourceText: "false + (() => {});", options: ""},
		{sourceText: "false + function () {};", options: ""},
		{sourceText: "1 + 1n;", options: ""},
		{sourceText: "1 + [];", options: ""},
		{sourceText: "1 + /regex/;", options: ""},
		{sourceText: "1 + (__dirname === 'foobar');", options: ""},
		{sourceText: "1 + {}.constructor();", options: ""},
		{sourceText: "1 + (() => {});", options: ""},
		{sourceText: "1 + function () {};", options: ""},
		{sourceText: "1n + [];", options: ""},
		{sourceText: "1n + /regex/;", options: ""},
		{sourceText: "1n + (__dirname === 'foobar');", options: ""},
		{sourceText: "1n + {}.constructor();", options: ""},
		{sourceText: "1n + (() => {});", options: ""},
		{sourceText: "1n + function () {};", options: ""},
		{sourceText: "[] + /regex/;", options: ""},
		{sourceText: "[] + (__dirname === 'foobar');", options: ""},
		{sourceText: "[] + {}.constructor();", options: ""},
		{sourceText: "[] + (() => {});", options: ""},
		{sourceText: "[] + function () {};", options: ""},
		{sourceText: "/regex/ + (__dirname === 'foobar');", options: ""},
		{sourceText: "/regex/ + {}.constructor();", options: ""},
		{sourceText: "/regex/ + (() => {});", options: ""},
		{sourceText: "/regex/ + function () {};", options: ""},
		{sourceText: "(__dirname === 'foobar') + {}.constructor();", options: ""},
		{sourceText: "(__dirname === 'foobar') + (() => {});", options: ""},
		{sourceText: "(__dirname === 'foobar') + function () {};", options: ""},
		{sourceText: "({}).constructor() + (() => {});", options: ""},
		{sourceText: "({}).constructor() + function () {};", options: ""},
		{sourceText: "(() => {}) + function () {};", options: ""},
		{sourceText: "''.toString();", options: ""},
		{sourceText: "'text'.toString();", options: ""},
		{sourceText: "true.toString();", options: ""},
		{sourceText: "false.toString();", options: ""},
		{sourceText: "(1).toString();", options: ""},
		{sourceText: "1n.toString();", options: ""},
		{sourceText: "[].toString();", options: ""},
		{sourceText: "/regex/.toString();", options: ""},
		{sourceText: "(__dirname === 'foobar').toString();", options: ""},
		{sourceText: "({}).constructor().toString();", options: ""},
		{sourceText: "(() => {}).toString();", options: ""},
		{sourceText: "(function () {}).toString();", options: ""},
		{sourceText: "\ndeclare const a: {\n  [Symbol.toPrimitive](): string;\n};\n\n`${a}`;\n    ", options: ""},
		{sourceText: "\ndeclare const a: {\n  valueOf(): string;\n};\n\n`${a}`;\n    ", options: ""},
		{sourceText: "\nlet value = '';\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = 'text';\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = true;\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = false;\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = 1;\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = 1n;\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = [];\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = /regex/;\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = __dirname === 'foobar';\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = {}.constructor();\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = () => {};\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "\nlet value = function () {};\nvalue.toString();\nlet text = `${value}`;\n    ", options: ""},
		{sourceText: "String('');", options: ""},
		{sourceText: "String('text');", options: ""},
		{sourceText: "String(true);", options: ""},
		{sourceText: "String(false);", options: ""},
		{sourceText: "String(1);", options: ""},
		{sourceText: "String(1n);", options: ""},
		{sourceText: "String([]);", options: ""},
		{sourceText: "String(/regex/);", options: ""},
		{sourceText: "String(__dirname === 'foobar');", options: ""},
		{sourceText: "String({}.constructor());", options: ""},
		{sourceText: "String(() => {});", options: ""},
		{sourceText: "String(function () {});", options: ""},
		{sourceText: "\nconst String = (value: unknown) => 'safe';\nString({});\n    ", options: ""},
		{sourceText: "\nconst String = (value: unknown) => 'safe';\nfunction f() {\n  String({});\n}\n    ", options: ""},
		{sourceText: "\nfunction String(value: unknown) {\n  return 'safe';\n}\nfunction f() {\n  String({});\n}\n    ", options: ""},
		{sourceText: "\nfunction someFunction() {}\nsomeFunction.toString();\nlet text = `${someFunction}`;\n    ", options: ""},
		{sourceText: "\nfunction someFunction() {}\nsomeFunction.toLocaleString();\nlet text = `${someFunction}`;\n    ", options: ""},
		{sourceText: "unknownObject.toString();", options: ""},
		{sourceText: "unknownObject.toLocaleString();", options: ""},
		{sourceText: "unknownObject.someOtherMethod();", options: ""},
		{sourceText: "\nclass CustomToString {\n  toString() {\n    return 'Hello, world!';\n  }\n}\n'' + new CustomToString();\n    ", options: ""},
		{sourceText: "\nconst literalWithToString = {\n  toString: () => 'Hello, world!',\n};\n'' + literalWithToString;\n    ", options: ""},
		{sourceText: "\nconst printer = (inVar: string | number | boolean) => {\n  inVar.toString();\n};\nprinter('');\nprinter(1);\nprinter(true);\n    ", options: ""},
		{sourceText: "\nconst printer = (inVar: string | number | boolean) => {\n  inVar.toLocaleString();\n};\nprinter('');\nprinter(1);\nprinter(true);\n    ", options: ""},
		{sourceText: "let _ = {} * {};", options: ""},
		{sourceText: "let _ = {} / {};", options: ""},
		{sourceText: "let _ = ({} *= {});", options: ""},
		{sourceText: "let _ = ({} /= {});", options: ""},
		{sourceText: "let _ = ({} = {});", options: ""},
		{sourceText: "let _ = {} == {};", options: ""},
		{sourceText: "let _ = {} === {};", options: ""},
		{sourceText: "let _ = {} in {};", options: ""},
		{sourceText: "let _ = {} & {};", options: ""},
		{sourceText: "let _ = {} ^ {};", options: ""},
		{sourceText: "let _ = {} << {};", options: ""},
		{sourceText: "let _ = {} >> {};", options: ""},
		{sourceText: "\nfunction tag() {}\ntag`${{}}`;\n    ", options: ""},
		{sourceText: "\ninterface Brand {}\nfunction test(v: string & Brand): string {\n  return `${v}`;\n}\n    ", options: ""},
		{sourceText: "'' += new Error();", options: ""},
		{sourceText: "'' += new URL();", options: ""},
		{sourceText: "'' += new URLSearchParams();", options: ""},
		{sourceText: "\nNumber(1);\n    ", options: ""},
		{sourceText: "String(/regex/);", options: "{\"ignoredTypeNames\": [\"RegExp\"]}"},
		{sourceText: "\ntype Foo = { a: string } | { b: string };\ndeclare const foo: Foo;\nString(foo);\n      ", options: "{\"ignoredTypeNames\": [\"Foo\"]}"},
		{sourceText: "\ninterface MyError<T> {}\ndeclare const error: MyError<number>;\nerror.toString();\n      ", options: "{\"ignoredTypeNames\": [\"MyError\"]}"},
		{sourceText: "\ntype MyError<T> = {};\ndeclare const error: MyError<number>;\nerror.toString();\n      ", options: "{\"ignoredTypeNames\": [\"MyError\"]}"},
		{sourceText: "\nclass MyError<T> {}\ndeclare const error: MyError<number>;\nerror.toString();\n      ", options: "{\"ignoredTypeNames\": [\"MyError\"]}"},
		{sourceText: "\ninterface Animal {}\ninterface Serializable {}\ninterface Cat extends Animal, Serializable {}\n\ndeclare const whiskers: Cat;\nwhiskers.toString();\n      ", options: "{\"ignoredTypeNames\": [\"Animal\"]}"},
		{sourceText: "\ninterface MyError extends Error {}\n\ndeclare const error: MyError;\nerror.toString();\n      ", options: ""},
		{sourceText: "\nclass BaseError extends Error {\n  code?: string;\n}\n\nclass Boom<T> extends BaseError {\n  details: T;\n}\n\nfunction bar<T>(error: Boom<T>) {\n  console.log(error.toString());\n}\n      ", options: ""},
		{sourceText: "\nclass UnknownBase {}\nclass CustomError extends UnknownBase {}\n\ndeclare const err: CustomError;\nerr.toString();\n      ", options: "{\"ignoredTypeNames\": [\"UnknownBase\"]}"},
		{sourceText: "\ninterface Animal {}\ninterface Dog extends Animal {}\ninterface Cat extends Animal {}\n\ndeclare const dog: Dog;\ndeclare const cat: Cat;\ncat.toString();\n      ", options: "{\"ignoredTypeNames\": [\"Animal\"]}"},
		{sourceText: "\nfunction String(value) {\n  return value;\n}\ndeclare const myValue: object;\nString(myValue);\n    ", options: ""},
		{sourceText: "\nimport { String } from 'foo';\nString({});\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString(): string;\n  toString(options: { verbose: boolean }): string;\n  toString(options?: { verbose: boolean }) {\n    return 'Hello, world!';\n  }\n}\n'' + new Foo();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString(prefix?: string): string {\n    return (prefix ?? '') + 'Hello, world!';\n  }\n}\n'' + new Foo();\n    ", options: ""},
		{sourceText: "\ndeclare module 'guid' {\n  export function toString(id: number): string;\n  export function toString(id: number, format: string): string;\n}\nimport * as GUID from 'guid';\nGUID.toString(123);\n    ", options: ""},
		{sourceText: "\n['foo', 'bar'].join('');\n    ", options: ""},
		{sourceText: "\n([{ foo: 'foo' }, 'bar'] as string[]).join('');\n    ", options: ""},
		{sourceText: "\nfunction foo<T extends string>(array: T[]) {\n  return array.join();\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString() {\n    return '';\n  }\n}\n[new Foo()].join();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  join() {}\n}\nconst foo = new Foo();\nfoo.join();\n    ", options: ""},
		{sourceText: "\ndeclare const array: string[];\narray.join('');\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo)[];\narray.join('');\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] | (string & Bar)[];\narray.join('');\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] & (string & Bar)[];\narray.join('');\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const tuple: [string & Foo, string & Bar];\ntuple.join('');\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string] & [Foo];\ntuple.join('');\n    ", options: ""},
		{sourceText: "\nString(['foo', 'bar']);\n    ", options: ""},
		{sourceText: "\nString([{ foo: 'foo' }, 'bar'] as string[]);\n    ", options: ""},
		{sourceText: "\nfunction foo<T extends string>(array: T[]) {\n  return String(array);\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString() {\n    return '';\n  }\n}\nString([new Foo()]);\n    ", options: ""},
		{sourceText: "\ndeclare const array: string[];\nString(array);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo)[];\nString(array);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] | (string & Bar)[];\nString(array);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] & (string & Bar)[];\nString(array);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const tuple: [string & Foo, string & Bar];\nString(tuple);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string] & [Foo];\nString(tuple);\n    ", options: ""},
		{sourceText: "\n['foo', 'bar'].toString();\n    ", options: ""},
		{sourceText: "\n([{ foo: 'foo' }, 'bar'] as string[]).toString();\n    ", options: ""},
		{sourceText: "\nfunction foo<T extends string>(array: T[]) {\n  return array.toString();\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString() {\n    return '';\n  }\n}\n[new Foo()].toString();\n    ", options: ""},
		{sourceText: "\ndeclare const array: string[];\narray.toString();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo)[];\narray.toString();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] | (string & Bar)[];\narray.toString();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] & (string & Bar)[];\narray.toString();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const tuple: [string & Foo, string & Bar];\ntuple.toString();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string] & [Foo];\ntuple.toString();\n    ", options: ""},
		{sourceText: "\n`${['foo', 'bar']}`;\n    ", options: ""},
		{sourceText: "\n`${[{ foo: 'foo' }, 'bar'] as string[]}`;\n    ", options: ""},
		{sourceText: "\nfunction foo<T extends string>(array: T[]) {\n  return `${array}`;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  toString() {\n    return '';\n  }\n}\n`${[new Foo()]}`;\n    ", options: ""},
		{sourceText: "\ndeclare const array: string[];\n`${array}`;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo)[];\n`${array}`;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] | (string & Bar)[];\n`${array}`;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: (string & Foo)[] & (string & Bar)[];\n`${array}`;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const tuple: [string & Foo, string & Bar];\n`${tuple}`;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string] & [Foo];\n`${tuple}`;\n    ", options: ""},
		{sourceText: "\nlet objects = [{}, {}];\nString(...objects);\n    ", options: ""},
		{sourceText: "\ntype Constructable<Entity> = abstract new (...args: any[]) => Entity;\n\ninterface GuildChannel {\n  toString(): `<#${string}>`;\n}\n\ndeclare const foo: Constructable<GuildChannel & { bar: 1 }>;\nclass ExtendedGuildChannel extends foo {}\ndeclare const bb: ExtendedGuildChannel;\nbb.toString();\n    ", options: ""},
		{sourceText: "\ntype Constructable<Entity> = abstract new (...args: any[]) => Entity;\n\ninterface GuildChannel {\n  toString(): `<#${string}>`;\n}\n\ndeclare const foo: Constructable<{ bar: 1 } & GuildChannel>;\nclass ExtendedGuildChannel extends foo {}\ndeclare const bb: ExtendedGuildChannel;\nbb.toString();\n    ", options: ""},
		{sourceText: "\ntype Value = string | Value[];\ndeclare const v: Value;\n\nString(v);\n    ", options: ""},
		{sourceText: "\ntype Value = (string | Value)[];\ndeclare const v: Value;\n\nString(v);\n    ", options: ""},
		{sourceText: "\ntype Value = Value[];\ndeclare const v: Value;\n\nString(v);\n    ", options: ""},
		{sourceText: "\ntype Value = [Value];\ndeclare const v: Value;\n\nString(v);\n    ", options: ""},
		{sourceText: "\ndeclare const v: ('foo' | 'bar')[][];\nString(v);\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\n`${x})`;\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\nx.toString();\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\nx.toLocaleString();\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\n'' + x;\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\nString(x);\n    ", options: ""},
		{sourceText: "\ndeclare const x: unknown;\n'' += x;\n    ", options: ""},
		{sourceText: "\nfunction foo<T>(x: T) {\n  String(x);\n}\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\n`${x})`;\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\nx.toString();\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\nx.toLocaleString();\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\n'' + x;\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\nString(x);\n    ", options: ""},
		{sourceText: "\ndeclare const x: any;\n'' += x;\n    ", options: ""},
	}
	for index, testCase := range cases {
		t.Run(noBaseToStringCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoBaseToString(t, testCase.sourceText, testCase.options))
		})
	}
}

// TestNoBaseToStringFiresOnUpstreamFailCases is the imported failing corpus, verbatim, with the span
// and the RENDERED message of every finding asserted.
//
// The message is the reason this table asserts text rather than ids. Upstream interpolates two
// values into it, the source text of the offending expression and a three-valued certainty, and the
// corpus records both per case rather than recording line numbers. So the certainty is stated data
// here, not something recovered: fifty-one findings say the value MAY stringify uselessly and
// forty-three say it WILL, and no message id assertion can tell those apart.
//
// Both ids are represented, seventy-seven for a plain stringification and seventeen for an array
// join, because joining an array of objects has the same defect through a different door.
func TestNoBaseToStringFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		options      string
		wantIds      []string
		wantSpans    []string
		wantMessages []string
	}{
		{
			sourceText:   "\ndeclare const x: unknown;\n`${x})`;\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const x: unknown;\nx.toString();\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const x: unknown;\nx.toLocaleString();\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const x: unknown;\n'' + x;\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const x: unknown;\nString(x);\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const x: unknown;\n'' += x;\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nfunction foo<T>(x: T) {\n  String(x);\n}\n      ",
			options:      "{\"checkUnknown\": true}",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "`${{}})`;",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "({}).toString();",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "({}).toLocaleString();",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "'' + {};",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "String({});",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "'' += {};",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"{}"},
			wantMessages: []string{"'{}' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrString = Math.random() ? { a: true } : 'text';\nsomeObjectOrString.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrString"},
			wantMessages: []string{"'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrString = Math.random() ? { a: true } : 'text';\nsomeObjectOrString.toLocaleString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrString"},
			wantMessages: []string{"'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrString = Math.random() ? { a: true } : 'text';\nsomeObjectOrString + '';\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrString"},
			wantMessages: []string{"'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };\nsomeObjectOrObject.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrObject"},
			wantMessages: []string{"'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };\nsomeObjectOrObject.toLocaleString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrObject"},
			wantMessages: []string{"'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nlet someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };\nsomeObjectOrObject + '';\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"someObjectOrObject"},
			wantMessages: []string{"'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ninterface A {}\ninterface B {}\nfunction test(intersection: A & B): string {\n  return `${intersection}`;\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"intersection"},
			wantMessages: []string{"'intersection' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const foo: string | Foo;\n`${foo}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"'foo' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const foo: Bar | Foo;\n`${foo}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"'foo' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const foo: Bar & Foo;\n`${foo}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"'foo' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\n[{}, {}].join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"[{}, {}]"},
			wantMessages: []string{"Using `join()` for [{}, {}] will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = [{}, {}];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass A {\n  a: string;\n}\n[new A(), 'str'].join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"[new A(), 'str']"},
			wantMessages: []string{"Using `join()` for [new A(), 'str'] may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string | Foo)[];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo) | (string | Foo)[];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: Foo[] & Bar[];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: string[] | Foo[];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, Foo];\ntuple.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, Foo];\ntuple.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo | string, string];\ntuple.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"Using `join()` for tuple may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, string] | [Foo, Foo];\ntuple.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"Using `join()` for tuple may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, string] & [Foo, Foo];\ntuple.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = ['string', { foo: 'bar' }];\narray.join('');\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return array.join();\n}\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nString([{}, {}]);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[{}, {}]"},
			wantMessages: []string{"'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = [{}, {}];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass A {\n  a: string;\n}\nString([new A(), 'str']);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[new A(), 'str']"},
			wantMessages: []string{"'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string | Foo)[];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo) | (string | Foo)[];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: Foo[] & Bar[];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: string[] | Foo[];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, Foo];\nString(tuple);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, Foo];\nString(tuple);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo | string, string];\nString(tuple);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, string] | [Foo, Foo];\nString(tuple);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, string] & [Foo, Foo];\nString(tuple);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = ['string', { foo: 'bar' }];\nString(array);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return String(array);\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const a:\n  | {\n      [Symbol.toPrimitive](): string;\n    }\n  | {\n      other: true;\n    };\n\n`${a}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"a"},
			wantMessages: []string{"'a' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const a:\n  | {\n      valueOf(): string;\n    }\n  | {\n      other: true;\n    };\n\n`${a}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"a"},
			wantMessages: []string{"'a' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\n[{}, {}].toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[{}, {}]"},
			wantMessages: []string{"'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = [{}, {}];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass A {\n  a: string;\n}\n[new A(), 'str'].toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[new A(), 'str']"},
			wantMessages: []string{"'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string | Foo)[];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo) | (string | Foo)[];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: Foo[] & Bar[];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: string[] | Foo[];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, Foo];\ntuple.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, Foo];\ntuple.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo | string, string];\ntuple.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, string] | [Foo, Foo];\ntuple.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, string] & [Foo, Foo];\ntuple.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = ['string', { foo: 'bar' }];\narray.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return array.toString();\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\n`${[{}, {}]}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[{}, {}]"},
			wantMessages: []string{"'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = [{}, {}];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass A {\n  a: string;\n}\n`${[new A(), 'str']}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"[new A(), 'str']"},
			wantMessages: []string{"'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string | Foo)[];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: (string & Foo) | (string | Foo)[];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\nclass Bar {\n  bar: string;\n}\ndeclare const array: Foo[] & Bar[];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const array: string[] | Foo[];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, Foo];\n`${tuple}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, Foo];\n`${tuple}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo | string, string];\n`${tuple}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [string, string] | [Foo, Foo];\n`${tuple}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nclass Foo {\n  foo: string;\n}\ndeclare const tuple: [Foo, string] & [Foo, Foo];\n`${tuple}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"tuple"},
			wantMessages: []string{"'tuple' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\nconst array = ['string', { foo: 'bar' }];\n`${array}`;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return `${array}`;\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array"},
			wantMessages: []string{"'array' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  array[0].toString();\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"array[0]"},
			wantMessages: []string{"'array[0]' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(value: T) {\n  value.toString();\n}\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"value"},
			wantMessages: []string{"'value' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\ndeclare const foo: Bar | string;\nfoo.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"'foo' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return array;\n}\nfoo([{ foo: 'foo' }]).join();\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"foo([{ foo: 'foo' }])"},
			wantMessages: []string{"Using `join()` for foo([{ foo: 'foo' }]) will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Bar = Record<string, string>;\nfunction foo<T extends string | Bar>(array: T[]) {\n  return array;\n}\nfoo([{ foo: 'foo' }, 'bar']).join();\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"foo([{ foo: 'foo' }, 'bar'])"},
			wantMessages: []string{"Using `join()` for foo([{ foo: 'foo' }, 'bar']) may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Value = { foo: string } | Value[];\ndeclare const v: Value;\n\nString(v);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"v"},
			wantMessages: []string{"'v' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Value = ({ foo: string } | Value)[];\ndeclare const v: Value;\n\nString(v);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"v"},
			wantMessages: []string{"'v' may use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Value = [{ foo: string }, Value];\ndeclare const v: Value;\n\nString(v);\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"v"},
			wantMessages: []string{"'v' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ndeclare const v: { foo: string }[][];\nv.join();\n      ",
			options:      "",
			wantIds:      []string{"baseArrayJoin"},
			wantSpans:    []string{"v"},
			wantMessages: []string{"Using `join()` for v will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ninterface Dog extends Animal {}\n\ndeclare const labrador: Dog;\nlabrador.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"labrador"},
			wantMessages: []string{"'labrador' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ninterface A extends B {}\ninterface B extends A {}\n\ndeclare const a: A;\na.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"a"},
			wantMessages: []string{"'a' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ninterface Base {}\ninterface Left extends Base {}\ninterface Right extends Base {}\ninterface Diamond extends Left, Right {}\n\ndeclare const d: Diamond;\nd.toString();\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"d"},
			wantMessages: []string{"'d' will use Object's default stringification format ('[object Object]') when stringified."},
		},
		{
			sourceText:   "\ntype Mapped = { [K in 'toString']: () => string };\ndeclare const x: Mapped;\n'' + x;\n      ",
			options:      "",
			wantIds:      []string{"baseToString"},
			wantSpans:    []string{"x"},
			wantMessages: []string{"'x' will use Object's default stringification format ('[object Object]') when stringified."},
		},
	}
	for index, testCase := range cases {
		t.Run(noBaseToStringCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := runNoBaseToString(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			written := asTheTypedNoBaseToStringHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
				if reported.Message.Description != testCase.wantMessages[findingIndex] {
					t.Errorf("finding %d reads %q, wanted %q", findingIndex,
						reported.Message.Description, testCase.wantMessages[findingIndex])
				}
			}
		})
	}
}

// TestNoBaseToStringOnShapesUpstreamsCorpusDoesNotWrite covers two tests the imported cases cannot
// separate, both found by mutation.
//
// The first is the constrained type on the join arm. Upstream joins concrete arrays throughout its
// corpus and never a type parameter, so asking for the plain type instead of the constrained one is
// invisible to all three hundred and seventeen rows and then loses every finding on a generic array.
//
// The second is the local shadow of `String`. Upstream asks scope analysis whether any variable of
// that name is declared and treats "none" as the global; we have no scope index, so the symbol is
// resolved and its declarations examined, and one outside the default library means a shadow. That
// is the complement of the naive predicate the brief warns about, which answers false for exactly
// the globals a rule like this must recognize. Upstream shadows String nowhere in its corpus.
//
// Measured against the installed 8.x build, one program per case, with controls in the same runs.
func TestNoBaseToStringOnShapesUpstreamsCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
	}{
		{
			name:       "join-on-constrained-parameter",
			why:        "a join on a TYPE PARAMETER constrained to an array of objects. The join arm asks for the CONSTRAINED type rather than the plain one, and upstream's corpus joins only concrete arrays, so reading the plain type survives every imported row and then goes silent on every generic array join",
			sourceText: "function f<T extends {}[]>(x: T) {\n  x.join(', ');\n}\n",
			wantIds:    []string{"baseArrayJoin"},
		},
		{
			name:       "join-on-plain-array",
			why:        "the concrete form of the row above, which the corpus does cover, kept beside it so the pair reads as one decision",
			sourceText: "declare const x: {}[];\nx.join(', ');\n",
			wantIds:    []string{"baseArrayJoin"},
		},
		{
			name:       "shadowed-string-call",
			why:        "a LOCAL function named String, which upstream declines because its scope lookup finds a declaration. We have no scope index, so the symbol's declarations are examined instead and one outside the default library means a shadow. Upstream's corpus shadows String nowhere, so dropping the test survives every imported row and then reports on every call to a user-defined String",
			sourceText: "function String(x: unknown) { return 'x'; }\ndeclare const o: {};\nString(o);\n",
			wantIds:    []string{},
		},
		{
			name:       "global-string-call",
			why:        "the control: the real global String, which reports",
			sourceText: "declare const o: {};\nString(o);\n",
			wantIds:    []string{"baseToString"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoBaseToString(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantIds) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantIds), len(result.Diagnostics), testCase.why)
			}
		})
	}
}

// TestNoBaseToStringRequiresTheTypedHarness pins the checker guards.
//
// All three listeners start with a nil check, and the shim answers nil from a type query on a nil
// checker rather than panicking, so a rule missing those guards goes silent rather than crashing. A
// vacuous green is the more dangerous of the two failures because nothing announces it.
func TestNoBaseToStringRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const o: {};\n`${o}`;\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoBaseToString, noBaseToStringFile, sourceText))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoBaseToString, noBaseToStringFile,
		sourceText), "baseToString")
}

// TestNoBaseToStringDecoderKeepsTheBuiltinIgnoredNames pins the one default in this batch that is
// not a zero value.
//
// ignoredTypeNames defaults to four builtin type names, so a decoder returning an empty slice for an
// absent key would start reporting on every Error, RegExp, URL and URLSearchParams in the tree. An
// explicitly empty list means the opposite and must be honored, which is why the wire field is a
// pointer: absent and empty are different configurations here.
func TestNoBaseToStringDecoderKeepsTheBuiltinIgnoredNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw              string
		wantCheckUnknown bool
		wantIgnored      []string
	}{
		{raw: `{}`, wantIgnored: []string{"Error", "RegExp", "URL", "URLSearchParams"}},
		{
			raw:              `{"checkUnknown": true}`,
			wantCheckUnknown: true,
			wantIgnored:      []string{"Error", "RegExp", "URL", "URLSearchParams"},
		},
		{raw: `{"ignoredTypeNames": ["MyError"]}`, wantIgnored: []string{"MyError"}},
		{raw: `{"ignoredTypeNames": []}`, wantIgnored: []string{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded := decodeNoBaseToStringOptionsForTest(t, testCase.raw)
			settings, isSettings := decoded.(NoBaseToStringOptions)
			if !isSettings {
				t.Fatalf("the decoder returned %T rather than the rule's own options type", decoded)
			}
			if settings.CheckUnknown != testCase.wantCheckUnknown {
				t.Errorf("checkUnknown decoded to %v, wanted %v",
					settings.CheckUnknown, testCase.wantCheckUnknown)
			}
			if len(settings.IgnoredTypeNames) != len(testCase.wantIgnored) {
				t.Fatalf("ignoredTypeNames decoded to %v, wanted %v",
					settings.IgnoredTypeNames, testCase.wantIgnored)
			}
			for index, wanted := range testCase.wantIgnored {
				if settings.IgnoredTypeNames[index] != wanted {
					t.Errorf("ignoredTypeNames[%d] is %q, wanted %q",
						index, settings.IgnoredTypeNames[index], wanted)
				}
			}
		})
	}
}

// TestNoBaseToStringFallsBackWhenHandedNilOptions covers the path a rule configured as a bare
// severity string takes.
//
// A rule named as "error" with no object is handed nil, which the type assertion in Run cannot
// satisfy, so the fallback is the only thing between that configuration and an empty ignored list.
// Every other fixture reaches the rule through the decoder and none can see this line.
func TestNoBaseToStringFallsBackWhenHandedNilOptions(t *testing.T) {
	t.Parallel()

	// An Error is ignored by default, so a fallback that lost the builtin names would report here.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoBaseToString,
		noBaseToStringFile, "declare const e: Error;\n`${e}`;\n", nil))
	// And the control, so the silence above is a verdict rather than an inert rule.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoBaseToString,
		noBaseToStringFile, "declare const o: {};\n`${o}`;\n", nil), "baseToString")
}
