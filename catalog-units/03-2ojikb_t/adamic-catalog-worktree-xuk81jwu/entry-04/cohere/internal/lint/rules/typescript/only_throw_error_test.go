package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// onlyThrowErrorFile is the fixture name every case in this file runs under.
//
// A TypeScript extension rather than `.js`, because the corpus is full of `declare`, type
// annotations, generics and `as` casts, all of which would be parse errors under a JavaScript name
// rather than rule inputs.
const onlyThrowErrorFile = "onlyThrowError.ts"

// onlyThrowErrorSupportingFiles are the two modules some corpus cases import.
//
// Upstream keeps both under `tests/fixtures` and every case that imports one depends on the exact
// contents. `class.ts` exports a userland `Error` that shadows the built-in, which is the whole
// point of the two cases that import it: `import { Error } from './class'; throw new Error();`
// reports, because that Error is not the lib one. `./missing` is deliberately absent, so the case
// importing it resolves to the error type; that is upstream's arrangement, not an oversight here.
var onlyThrowErrorSupportingFiles = map[string]string{
	"class.ts": "export class Error {}\n",
}

// runOnlyThrowError routes every case through the typed harness, and through the OPTIONS path even
// when the case configures nothing.
//
// A case with no options passes nil rather than a zero struct, which is what a rule configured as
// bare `"error"` is actually handed in production. That is the shape the standing warning names:
// `options.(OnlyThrowErrorOptions)` on nil yields the zero value, whose three nil pointers the body
// then defaults to TRUE. Routing the unconfigured cases through here is what puts that defaulting
// under test rather than leaving it to a mutation sweep, and it matters more for this rule than for
// most: every default is permissive, so a defaulting bug turns an unconfigured rule into one that
// reports every `any` in the tree.
//
// # Why the supporting files are always present
//
// Adding `class.ts` to every program rather than only to the two cases that import it costs
// nothing and removes a per-case branch. It declares a module, so it cannot leak a global into the
// other cases.
func runOnlyThrowError(t *testing.T, sourceText string, options any) rule_testing.Result {
	t.Helper()
	files := map[string]string{onlyThrowErrorFile: sourceText}
	for name, contents := range onlyThrowErrorSupportingFiles {
		files[name] = contents
	}
	if options == nil {
		return rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, files, onlyThrowErrorFile, nil)
	}
	return rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, files, onlyThrowErrorFile, options)
}

// TestOnlyThrowErrorStaysSilent replays every clean case in upstream's corpus.
//
// These are the false positives upstream already thought about, and each one was added when
// somebody hit that bug. They are imported verbatim: extracted from the test file with a
// TypeScript AST walk rather than retyped, and byte-verified against the source afterwards, so a
// case asserts what upstream asserts rather than what this port believes.
//
// Every one of the 42 was additionally replayed against the INSTALLED @typescript-eslint rule
// before being written here, and all 42 agreed with the corpus's own verdict, so these are
// measured expectations rather than transcribed ones.
func TestOnlyThrowErrorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"upstream valid case 0", "throw new Error();", nil},
		{"upstream valid case 1", "throw new Error('error');", nil},
		{"upstream valid case 2", "throw Error('error');", nil},
		{"upstream valid case 3", "\nconst e = new Error();\nthrow e;\n    ", nil},
		{"upstream valid case 4", "\ntry {\n  throw new Error();\n} catch (e) {\n  throw e;\n}\n    ", nil},
		{"upstream valid case 5", "\nfunction foo() {\n  return new Error();\n}\nthrow foo();\n    ", nil},
		{"upstream valid case 6", "\nconst foo = {\n  bar: new Error(),\n};\nthrow foo.bar;\n    ", nil},
		{"upstream valid case 7", "\nconst foo = {\n  bar: new Error(),\n};\n\nthrow foo['bar'];\n    ", nil},
		{"upstream valid case 8", "\nconst foo = {\n  bar: new Error(),\n};\n\nconst bar = 'bar';\nthrow foo[bar];\n    ", nil},
		{"upstream valid case 9", "\nclass CustomError extends Error {}\nthrow new CustomError();\n    ", nil},
		{"upstream valid case 10", "\nclass CustomError1 extends Error {}\nclass CustomError2 extends CustomError1 {}\nthrow new CustomError2();\n    ", nil},
		{"upstream valid case 11", "throw (foo = new Error());", nil},
		{"upstream valid case 12", "throw (1, 2, new Error());", nil},
		{"upstream valid case 13", "throw 'literal' && new Error();", nil},
		{"upstream valid case 14", "throw new Error() || 'literal';", nil},
		{"upstream valid case 15", "throw foo ? new Error() : new Error();", nil},
		{"upstream valid case 16", "\nfunction* foo() {\n  let index = 0;\n  throw yield index++;\n}\n    ", nil},
		{"upstream valid case 17", "\nasync function foo() {\n  throw await bar;\n}\n    ", nil},
		{"upstream valid case 18", "\nimport { Error } from './missing';\nthrow Error;\n    ", nil},
		{"upstream valid case 19", "\nclass CustomError<T, C> extends Error {}\nthrow new CustomError<string, string>();\n    ", nil},
		{"upstream valid case 20", "\nclass CustomError<T = {}> extends Error {}\nthrow new CustomError();\n    ", nil},
		{"upstream valid case 21", "\nclass CustomError<T extends object> extends Error {}\nthrow new CustomError();\n    ", nil},
		{"upstream valid case 22", "\nfunction foo() {\n  throw Object.assign(new Error('message'), { foo: 'bar' });\n}\n    ", nil},
		{"upstream valid case 23", "\nconst foo: Error | SyntaxError = bar();\nfunction bar() {\n  throw foo;\n}\n    ", nil},
		{"upstream valid case 24", "\ndeclare const foo: Error | string;\nthrow foo as Error;\n    ", nil},
		{"upstream valid case 25", "throw new Error() as Error;", nil},
		{"upstream valid case 26", "\ndeclare const nullishError: Error | undefined;\nthrow nullishError ?? new Error();\n    ", nil},
		{"upstream valid case 27", "\ndeclare const nullishError: Error | undefined;\nthrow nullishError || new Error();\n    ", nil},
		{"upstream valid case 28", "\ndeclare const nullishError: Error | undefined;\nthrow nullishError ? nullishError : new Error();\n    ", nil},
		{"upstream valid case 29", "\nfunction fun(value: any) {\n  throw value;\n}\n    ", nil},
		{"upstream valid case 30", "\nfunction fun(value: unknown) {\n  throw value;\n}\n    ", nil},
		{"upstream valid case 31", "\nfunction fun<T extends Error>(t: T): void {\n  throw t;\n}\n    ", nil},
		{"upstream valid case 32", "\nthrow undefined;\n      ", OnlyThrowErrorOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"undefined"}}}, AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		{"upstream valid case 33", "\nclass CustomError implements Error {}\nthrow new CustomError();\n      ", OnlyThrowErrorOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"CustomError"}}}, AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		{"upstream valid case 34", "\nthrow new Map();\n      ", OnlyThrowErrorOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"Map"}}}, AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		// upstream valid case 35 is not here. It is pinned by
		// TestOnlyThrowErrorAmbientPackageIsAHarnessLimit instead, because our fixture tsconfig
		// cannot express the program shape it needs. See that test for the measurement.
		{"upstream valid case 36", "\nfunction func<T1, T2>() {\n  let err: Promise<T1> | Promise<T2>;\n  throw err;\n}\n      ", OnlyThrowErrorOptions{AllowInline: []string{"Promise"}}},
		{"upstream valid case 37", "\ntry {\n} catch (e) {\n  throw e;\n}\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		{"upstream valid case 38", "\ntry {\n} catch (eOuter) {\n  try {\n    if (Math.random() > 0.5) {\n      throw eOuter;\n    }\n  } catch (eInner) {\n    if (Math.random() > 0.5) {\n      throw eOuter;\n    } else {\n      throw eInner;\n    }\n  }\n}\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		{"upstream valid case 39", "\nPromise.reject('foo').catch(e => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}},
		{"upstream valid case 40", "\nasync function foo() {\n  throw await Promise.resolve(new Error('error'));\n}\n      ", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}},
		{"upstream valid case 41", "\nfunction* foo(): Generator<number, void, Error> {\n  throw yield 303;\n}\n      ", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runOnlyThrowError(t, testCase.sourceText, testCase.options))
		})
	}
}

// TestOnlyThrowErrorFires replays every reporting case in upstream's corpus.
//
// The expected ids come from driving the installed rule over the same inputs rather than from
// reading the corpus's `errors` arrays, and the two agreed on all 47. That matters for the two
// cases that report `undef` rather than `object`: the distinction is invisible to a fixture that
// only counts findings, and it is the one place this rule's two messages can be swapped without
// any count changing.
func TestOnlyThrowErrorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"upstream invalid case 0", "throw undefined;", nil, []string{"undef"}},
		{"upstream invalid case 1", "throw new String('');", nil, []string{"object"}},
		{"upstream invalid case 2", "throw 'error';", nil, []string{"object"}},
		{"upstream invalid case 3", "throw 0;", nil, []string{"object"}},
		{"upstream invalid case 4", "throw false;", nil, []string{"object"}},
		{"upstream invalid case 5", "throw null;", nil, []string{"object"}},
		{"upstream invalid case 6", "throw {};", nil, []string{"object"}},
		{"upstream invalid case 7", "throw 'a' + 'b';", nil, []string{"object"}},
		{"upstream invalid case 8", "\nconst a = '';\nthrow a + 'b';\n      ", nil, []string{"object"}},
		{"upstream invalid case 9", "throw (foo = 'error');", nil, []string{"object"}},
		{"upstream invalid case 10", "throw (new Error(), 1, 2, 3);", nil, []string{"object"}},
		{"upstream invalid case 11", "throw 'literal' && 'not an Error';", nil, []string{"object"}},
		{"upstream invalid case 12", "throw 'literal' || new Error();", nil, []string{"object"}},
		{"upstream invalid case 13", "throw new Error() && 'literal';", nil, []string{"object"}},
		{"upstream invalid case 14", "throw 'literal' ?? new Error();", nil, []string{"object"}},
		{"upstream invalid case 15", "throw foo ? 'not an Error' : 'literal';", nil, []string{"object"}},
		{"upstream invalid case 16", "throw foo ? new Error() : 'literal';", nil, []string{"object"}},
		{"upstream invalid case 17", "throw foo ? 'literal' : new Error();", nil, []string{"object"}},
		{"upstream invalid case 18", "throw `${err}`;", nil, []string{"object"}},
		{"upstream invalid case 19", "\nconst err = 'error';\nthrow err;\n      ", nil, []string{"object"}},
		{"upstream invalid case 20", "\nfunction foo(msg) {}\nthrow foo('error');\n      ", nil, []string{"object"}},
		{"upstream invalid case 21", "\nconst foo = {\n  msg: 'error',\n};\nthrow foo.msg;\n      ", nil, []string{"object"}},
		{"upstream invalid case 22", "\nconst foo = {\n  msg: undefined,\n};\nthrow foo.msg;\n      ", nil, []string{"undef"}},
		{"upstream invalid case 23", "\nclass CustomError {}\nthrow new CustomError();\n      ", nil, []string{"object"}},
		{"upstream invalid case 24", "\nclass Foo {}\nclass CustomError extends Foo {}\nthrow new CustomError();\n      ", nil, []string{"object"}},
		{"upstream invalid case 25", "\nconst Error = null;\nthrow Error;\n      ", nil, []string{"object"}},
		{"upstream invalid case 26", "\nimport { Error } from './class';\nthrow new Error();\n      ", nil, []string{"object"}},
		{"upstream invalid case 27", "\nclass CustomError<T extends object> extends Foo {}\nthrow new CustomError();\n      ", nil, []string{"object"}},
		{"upstream invalid case 28", "\nfunction foo<T>() {\n  const res: T;\n  throw res;\n}\n      ", nil, []string{"object"}},
		{"upstream invalid case 29", "\nfunction foo<T>(fn: () => Promise<T>) {\n  const promise = fn();\n  const res = promise.then(() => {}).catch(() => {});\n  throw res;\n}\n      ", nil, []string{"object"}},
		{"upstream invalid case 30", "\nfunction foo() {\n  throw Object.assign({ foo: 'foo' }, { bar: 'bar' });\n}\n      ", nil, []string{"object"}},
		{"upstream invalid case 31", "\nconst foo: Error | { bar: string } = bar();\nfunction bar() {\n  throw foo;\n}\n      ", nil, []string{"object"}},
		{"upstream invalid case 32", "\ndeclare const foo: Error | string;\nthrow foo as string;\n      ", nil, []string{"object"}},
		{"upstream invalid case 33", "\nfunction fun(value: any) {\n  throw value;\n}\n      ", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 34", "\nfunction fun(value: unknown) {\n  throw value;\n}\n      ", OnlyThrowErrorOptions{AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 35", "\nfunction fun<T extends number>(t: T): void {\n  throw t;\n}\n      ", nil, []string{"object"}},
		{"upstream invalid case 36", "\nfunction func<T1, T2>() {\n  let err: Promise<T1> | Promise<T2> | void;\n  throw err;\n}\n      ", OnlyThrowErrorOptions{AllowInline: []string{"Promise"}}, []string{"object"}},
		{"upstream invalid case 37", "\nclass UnknownError implements Error {}\nthrow new UnknownError();\n      ", OnlyThrowErrorOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"CustomError"}}}, AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 38", "\nlet x = 1;\nPromise.reject('foo').catch(e => {\n  throw x;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 39", "\nPromise.reject('foo').catch((...e) => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 40", "\ndeclare const x: any[];\nPromise.reject('foo').catch(...x, e => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 41", "\ndeclare const x: any[];\nPromise.reject('foo').then(...x, e => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 42", "\ndeclare const onFulfilled: any;\ndeclare const x: any[];\nPromise.reject('foo').then(onFulfilled, ...x, e => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 43", "\nPromise.reject('foo').then((...e) => {\n  throw e;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 44", "\nPromise.reject('foo').then(e => {\n  throw globalThis;\n});\n      ", OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(true), AllowThrowingAny: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 45", "\nasync function foo() {\n  throw await bar;\n}\n      ", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}, []string{"object"}},
		{"upstream invalid case 46", "\nasync function foo() {\n  throw await Promise.resolve<number>(303);\n}\n      ", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}, []string{"object"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runOnlyThrowError(t, testCase.sourceText, testCase.options), testCase.wantIds...)
		})
	}
}

// TestOnlyThrowErrorSpans asserts WHERE each finding points, which no message-id fixture can see.
//
// The span is the throw argument with its parentheses skipped, and both halves of that are
// measured against the installed rule rather than assumed. Upstream reports `throw ('error')` at
// columns 8..15 of a one-line file, which is `'error'` without its wrapper, because ESTree has no
// parenthesis node to hand the rule. Our parser does have one, so reproducing the span means
// skipping it, and a port that did not would point at `('error')` while satisfying every id
// assertion in the two tests above.
//
// # The harness trims, and a span sliced from the Go literal would be off by one
//
// `rule_testing` writes each fixture as `strings.TrimSpace(contents)+"\n"`, so a source string written
// with a leading newline is one byte shorter on disk than in the literal here. Every source in
// this test is therefore written WITHOUT leading whitespace, so the literal and the file agree and
// a slice taken against the literal is the slice the rule saw.
func TestOnlyThrowErrorSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantSpans  []string
	}{
		{"a bare string literal points at the literal", "throw 'error';", nil, []string{"'error'"}},
		{"undefined points at the identifier", "throw undefined;", nil, []string{"undefined"}},
		{"a new expression points at the whole construction", "throw new String('');", nil, []string{"new String('')"}},
		{"a conditional points at the whole conditional", "declare const foo: boolean;\nthrow foo ? 'not an Error' : 'literal';", nil, []string{"foo ? 'not an Error' : 'literal'"}},
		{"a member access points at the access", "const foo = {\n  msg: 'error',\n};\nthrow foo.msg;", nil, []string{"foo.msg"}},
		{"an await points at the await expression", "declare const bar: number;\nasync function foo() {\n  throw await bar;\n}", OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)}, []string{"await bar"}},

		// Parentheses are skipped, matching upstream's span exactly. Measured on the installed
		// rule: `throw ('error');` reports at columns 8..15, which excludes both parens.
		{"one paren layer is excluded from the span", "throw ('error');", nil, []string{"'error'"}},
		{"two paren layers are excluded from the span", "throw (('error'));", nil, []string{"'error'"}},
		{"parens around undefined are excluded", "throw (undefined);", nil, []string{"undefined"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runOnlyThrowError(t, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantSpans))
			}
			for i, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[i]
				gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d spans %q, want %q", i, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestOnlyThrowErrorMessages asserts the rendered text of both messages, exactly.
//
// Neither message interpolates anything, so there is no format string to guard and the assertion
// is an equality on the Description. It is written against literal strings typed here rather than
// against the rule's own message constants: comparing a diagnostic to the constant it was built
// from is an equality that looks correct while both sides move together under mutation, so a
// message-text mutant would survive it.
func TestOnlyThrowErrorMessages(t *testing.T) {
	t.Parallel()

	objectResult := runOnlyThrowError(t, "throw 'error';", nil)
	if len(objectResult.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(objectResult.Diagnostics))
	}
	wantObject := "This throws a value that is not an Error, so whatever catches it gets no stack trace and no " +
		"message. The failure then surfaces as a bare string or object with no record of where it came " +
		"from, which is the difference between a five-minute fix and an afternoon. Throw `new Error(...)` " +
		"instead."
	if got := objectResult.Diagnostics[0].Message.Description; got != wantObject {
		t.Errorf("object description is %q, want %q", got, wantObject)
	}
	if got := objectResult.Diagnostics[0].Message.Id; got != "object" {
		t.Errorf("object id is %q, want %q", got, "object")
	}

	undefResult := runOnlyThrowError(t, "throw undefined;", nil)
	if len(undefResult.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(undefResult.Diagnostics))
	}
	wantUndef := "This throws `undefined`, which reaches the handler as a value carrying nothing at all: no " +
		"message, no stack, no type. It is almost always an expression that was meant to produce an error " +
		"and did not. Throw `new Error(...)` instead."
	if got := undefResult.Diagnostics[0].Message.Description; got != wantUndef {
		t.Errorf("undef description is %q, want %q", got, wantUndef)
	}
	if got := undefResult.Diagnostics[0].Message.Id; got != "undef" {
		t.Errorf("undef id is %q, want %q", got, "undef")
	}
}

// TestOnlyThrowErrorParenthesizedRethrow pins the one place our parser differs from ESTree.
//
// ESTree has no parenthesis node, so upstream's `node.type !== Identifier` guard never sees a
// wrapper and `throw (e)` reaches the rethrow arm as a bare identifier. Our parser produces
// `KindParenthesizedExpression`, so a literal port of that guard would decline the rethrow and
// report a case upstream allows.
//
// The second case is the CONTROL, and it is the reason this test proves anything: turning
// `allowRethrowing` off makes the same source report, which shows the first case is silent because
// the rethrow arm fired rather than for some unrelated reason. Both verdicts were measured against
// the installed rule before being written here.
func TestOnlyThrowErrorParenthesizedRethrow(t *testing.T) {
	t.Parallel()

	source := "try {\n} catch (e) {\n  throw (e);\n}"

	rule_testing.ExpectClean(t, runOnlyThrowError(t, source, nil))

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, source, OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(false),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}), "object")
}

// TestOnlyThrowErrorRethrowShape pins the four conditions the promise arm requires.
//
// Upstream states them as one conjunction, and reading a conjunction tells you nothing about which
// conjunct is load-bearing, so each was measured separately against the installed rule by changing
// exactly one thing and watching the verdict move. The corpus covers the rest-parameter row and
// none of the other three, so without these a port could drop any of them and stay green.
//
// Every case here runs with the two throwing escapes OFF, because the handler parameter is typed
// `any`, and with them on the case would be silent through the `any` arm regardless of what the
// rethrow arm decided. That is the "change the one thing you think makes it pass" check applied
// before the fixtures were written rather than after.
func TestOnlyThrowErrorRethrowShape(t *testing.T) {
	t.Parallel()

	strict := OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(true),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// The baseline both directions are measured against.
		{"an arrow catch handler is a rethrow", "Promise.reject('x').catch(e => {\n  throw e;\n});", nil},

		// The handler must be an ARROW function. A function expression binds `this` differently and
		// upstream's check names the arrow kind explicitly.
		{"a function expression handler is not a rethrow", "Promise.reject('x').catch(function (e) {\n  throw e;\n});", []string{"object"}},

		// The binding must be the FIRST parameter.
		{"a second parameter is not the rejection value", "Promise.reject('x').catch((a, e) => {\n  throw e;\n});", []string{"object"}},

		// It must be a plain identifier parameter. A destructuring pattern binds names taken out of
		// the rejection rather than the rejection itself.
		{"a destructured parameter is not the rejection value", "Promise.reject('x').catch(({ e }) => {\n  throw e;\n});", []string{"object"}},

		// The RECEIVER must be thenable. Any object may have a method called `catch`, and the name
		// alone would let a rethrow through on something that never handled a rejection.
		{"a catch method on a non-thenable is not a rethrow", "declare const o: { catch(cb: (e: string) => void): void };\no.catch(e => {\n  throw e;\n});", []string{"object"}},

		// A literal computed key resolves. Upstream reads the member name statically and a string
		// literal subscript is one of the forms its reader accepts.
		{"a literal computed catch resolves", "Promise.reject('x')['catch'](e => {\n  throw e;\n});", nil},

		// `.then` puts the rejection handler SECOND, so a one-argument `.then` has none.
		{"a two-argument then rethrows from the second handler", "Promise.reject('x').then(() => {}, e => {\n  throw e;\n});", nil},

		// A catch clause binding is the other rethrow shape, and a nested arrow reading the outer
		// binding still resolves to the same declaration.
		{"a nested arrow reading the caught value is a rethrow", "try {\n} catch (e) {\n  const f = () => {\n    throw e;\n  };\n}", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runOnlyThrowError(t, testCase.sourceText, strict)
			if testCase.wantIds == nil {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestOnlyThrowErrorConstFoldedKeyDiverges records the one stated divergence from upstream.
//
// Upstream's member-name reader falls back to a static evaluator that resolves an identifier
// through the scope manager to a constant initializer, so `const k = 'catch'; p[k](e => ...)` is
// recognized as a catch call and the rethrow is allowed. Reproducing that means a constant-folding
// pass, which is more machinery than the whole rest of this rule.
//
// Measured on the installed rule, with `allowRethrowing: false` as the control:
//
//	const k = 'catch'; Promise.reject('x')[k](e => { throw e; })   allowRethrowing on   SILENT
//	const k = 'catch'; Promise.reject('x')[k](e => { throw e; })   allowRethrowing off  REPORTS
//
// So upstream is silent and this port reports. No case in the corpus writes that shape. This test
// asserts the DIVERGENCE rather than upstream's verdict, so that a later port of the evaluator
// fails here loudly instead of quietly agreeing.
func TestOnlyThrowErrorConstFoldedKeyDiverges(t *testing.T) {
	t.Parallel()

	source := "const k = 'catch';\nPromise.reject('x')[k](e => {\n  throw e;\n});"
	rule_testing.ExpectFindings(t, runOnlyThrowError(t, source, OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(true),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}), "object")
}

// TestOnlyThrowErrorDefaultsArePermissive pins that all three booleans default to TRUE.
//
// This is the arm a nil-options registration reaches, and it is the one most expensive to get
// wrong: every default here is permissive, so an inverted default turns an unconfigured rule into
// one that reports every `any` and every rethrow in the tree. The three cases below are each
// silent ONLY because of their own default, which is shown by the paired strict case reporting.
func TestOnlyThrowErrorDefaultsArePermissive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		strictValue OnlyThrowErrorOptions
	}{
		{
			"any is thrown freely by default",
			"function fun(value: any) {\n  throw value;\n}",
			OnlyThrowErrorOptions{AllowThrowingAny: type_checking.Ref(false)},
		},
		{
			"unknown is thrown freely by default",
			"function fun(value: unknown) {\n  throw value;\n}",
			OnlyThrowErrorOptions{AllowThrowingUnknown: type_checking.Ref(false)},
		},
		{
			"a caught value is rethrown freely by default",
			"try {\n} catch (e) {\n  throw e;\n}",
			OnlyThrowErrorOptions{AllowRethrowing: type_checking.Ref(false), AllowThrowingUnknown: type_checking.Ref(false)},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// nil options is what a rule configured as bare "error" is handed.
			rule_testing.ExpectClean(t, runOnlyThrowError(t, testCase.sourceText, nil))
			// And the same source under the option turned off, which shows the default is what
			// made the first assertion pass rather than something else about the input.
			rule_testing.ExpectFindings(t, runOnlyThrowError(t, testCase.sourceText, testCase.strictValue), "object")
		})
	}
}

// TestOnlyThrowErrorUndefBeatsTheEscapes pins the ORDER of the undefined arm.
//
// `throw undefined` reports as `undef` with no options at all, which means the undefined test sits
// ABOVE the any and unknown escapes even though all three are permissive by default. Moving it
// below either one would make this case silent while every other fixture in the file stayed green,
// because no other case reaches both arms.
//
// The union row is the measurement that keeps the flag test honest. `Error | undefined` reports
// `object` rather than `undef`, because a union's own flags are `Union` and the undefined flag is
// on a member rather than on the whole. Reading upstream's `isTypeFlagSet` alone would suggest the
// opposite, so this is measured against the installed rule and pinned here.
func TestOnlyThrowErrorUndefBeatsTheEscapes(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, "throw undefined;", nil), "undef")

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, "throw undefined;", OnlyThrowErrorOptions{
		AllowThrowingAny:     type_checking.Ref(true),
		AllowThrowingUnknown: type_checking.Ref(true),
	}), "undef")

	rule_testing.ExpectFindings(t, runOnlyThrowError(t,
		"declare const nullishError: Error | undefined;\nthrow nullishError;", nil), "object")
}

// TestOnlyThrowErrorNeedsTheTypedHarness asserts the rule declares the checker.
//
// Without the declaration the rule is handed a nil checker in production, and its guard makes it go
// completely SILENT rather than crash, so every clean fixture would pass vacuously while the rule
// found nothing at all. A test rather than a comment, so a later revert fails loudly.
func TestOnlyThrowErrorNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !OnlyThrowError.NeedsTypeChecker {
		t.Fatal("OnlyThrowError must declare NeedsTypeChecker; every arm reads the argument's type")
	}

	// The untyped harness hands the rule a nil checker, and the guard turns that into silence.
	// Asserting the silence is what pins the guard: without it this call panics.
	rule_testing.ExpectClean(t, rule_testing.Run(t, OnlyThrowError, onlyThrowErrorFile, "throw 'error';"))
}

// TestDecodeOnlyThrowErrorOptions puts the decoder itself under test.
//
// Routing option fixtures through the exported decoder rather than building the struct by hand is
// what covers the two lines with no upstream counterpart: the `from` string becoming a `uint8`
// enum, and the one heterogeneous `allow` array splitting into two typed fields. Handing
// `RunTypedFilesWithOptions` a struct would leave both untested.
func TestDecodeOnlyThrowErrorOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty object leaves every pointer nil so the rule defaults them", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeOnlyThrowErrorOptions(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options, ok := decoded.(OnlyThrowErrorOptions)
		if !ok {
			t.Fatalf("decoded to %T, want OnlyThrowErrorOptions", decoded)
		}
		if options.AllowRethrowing != nil || options.AllowThrowingAny != nil || options.AllowThrowingUnknown != nil {
			t.Errorf("absent booleans decoded to non-nil: %+v", options)
		}
		if len(options.Allow) != 0 || len(options.AllowInline) != 0 {
			t.Errorf("absent allow decoded to %+v", options)
		}
	})

	t.Run("false is distinguishable from absent", func(t *testing.T) {
		t.Parallel()
		decoded, _ := DecodeOnlyThrowErrorOptions(json.RawMessage(`{"allowThrowingAny": false}`))
		options := decoded.(OnlyThrowErrorOptions)
		if options.AllowThrowingAny == nil {
			t.Fatal("allowThrowingAny: false decoded to nil, which the rule would read as the TRUE default")
		}
		if *options.AllowThrowingAny {
			t.Error("allowThrowingAny: false decoded to true")
		}
	})

	t.Run("a bare string lands in AllowInline and an object lands in Allow", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeOnlyThrowErrorOptions(json.RawMessage(
			`{"allow": ["Promise", {"from": "lib", "name": "undefined"}]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options := decoded.(OnlyThrowErrorOptions)
		if len(options.AllowInline) != 1 || options.AllowInline[0] != "Promise" {
			t.Errorf("AllowInline is %v, want [Promise]", options.AllowInline)
		}
		if len(options.Allow) != 1 {
			t.Fatalf("Allow holds %d specifiers, want 1", len(options.Allow))
		}
		if options.Allow[0].From != type_checking.TypeOrValueSpecifierFromLib {
			t.Errorf("from decoded to %v, want the lib enum", options.Allow[0].From)
		}
		if len(options.Allow[0].Name) != 1 || options.Allow[0].Name[0] != "undefined" {
			t.Errorf("name decoded to %v, want [undefined]", options.Allow[0].Name)
		}
	})

	t.Run("name accepts an array as well as a string", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeOnlyThrowErrorOptions(json.RawMessage(
			`{"allow": [{"from": "file", "name": ["A", "B"], "path": "x.ts"}]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options := decoded.(OnlyThrowErrorOptions)
		if len(options.Allow) != 1 || len(options.Allow[0].Name) != 2 {
			t.Fatalf("Allow decoded to %+v", options.Allow)
		}
		if strings.Join(options.Allow[0].Name, ",") != "A,B" {
			t.Errorf("name decoded to %v, want [A B]", options.Allow[0].Name)
		}
		if options.Allow[0].Path != "x.ts" {
			t.Errorf("path decoded to %q, want x.ts", options.Allow[0].Path)
		}
	})

	t.Run("an unrecognized from is dropped rather than silently meaning file", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeOnlyThrowErrorOptions(json.RawMessage(
			`{"allow": [{"from": "nowhere", "name": "X"}]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options := decoded.(OnlyThrowErrorOptions)
		// `file` is the zero value, so keeping the specifier would silently mean `file`. Dropping
		// it says nothing at all, which is the honest answer for a value we cannot represent.
		if len(options.Allow) != 0 {
			t.Errorf("an unrecognized from produced %+v, want it dropped", options.Allow)
		}
	})

	t.Run("the package form carries its package name", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeOnlyThrowErrorOptions(json.RawMessage(
			`{"allow": [{"from": "package", "name": "ErrorLike", "package": "errors"}]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options := decoded.(OnlyThrowErrorOptions)
		if len(options.Allow) != 1 {
			t.Fatalf("Allow decoded to %+v", options.Allow)
		}
		if options.Allow[0].From != type_checking.TypeOrValueSpecifierFromPackage {
			t.Errorf("from decoded to %v, want the package enum", options.Allow[0].From)
		}
		if options.Allow[0].Package != "errors" {
			t.Errorf("package decoded to %q, want errors", options.Allow[0].Package)
		}
	})
}

// TestOnlyThrowErrorAmbientPackageIsAHarnessLimit pins upstream valid case 35, whose verdict our
// fixture harness cannot reproduce.
//
// The case is `import { createError } from 'errors'; throw createError();` with
// `allow: [{from: 'package', name: 'ErrorLike', package: 'errors'}]`, and upstream is SILENT.
// Upstream makes `errors` resolvable with an AMBIENT declaration in a supporting fixture file:
//
//	// @ts-ignore
//	declare module 'errors' {
//	  class ErrorLike {}
//	  export function createError(): ErrorLike;
//	}
//
// `rule_testing`'s tsconfig pins `moduleDetection: "force"`, which makes every file a module. A
// `declare module 'errors'` inside a module is a module AUGMENTATION of an existing module rather
// than an ambient declaration of a new one, so `errors` never resolves, `createError()` has the
// error type, and the throw reports.
//
// # The measurement, and why this is the harness rather than the rule
//
// Reproduced on the INSTALLED @typescript-eslint rule by varying nothing but that one compiler
// option, holding source, options and file layout fixed:
//
//	moduleDetection: "force"     REPORTS object
//	moduleDetection unset        SILENT
//
// The second row is upstream's own verdict, and it is what the whole 89-case corpus was replayed
// under: with `moduleDetection` unset, all 89 installed-rule verdicts agree with the corpus's own
// assertions, and with it set, this one case is the only disagreement.
//
// Two CONTROLS establish that the module is what fails to resolve, rather than the package
// specifier being mismatched. Under our harness, the same input is ALSO reported when `allow` is
// the bare inline name `"ErrorLike"`, which needs no file or package comparison at all and would
// match on the type's own name; and it is still reported when the supporting file is given an
// explicit `export {}`. A type that resolved to `ErrorLike` would go clean under the first.
//
// So the case is recorded as REPORTING, with the reason at the line, rather than deleted or
// green-ed by weakening the rule. Changing the shared harness's tsconfig for one case would move
// every other typed rule's fixtures, which is a far larger blast radius than one documented row.
func TestOnlyThrowErrorAmbientPackageIsAHarnessLimit(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		onlyThrowErrorFile: "import { createError } from 'errors';\nthrow createError();\n",
		"errors.ts":        "// @ts-ignore\ndeclare module 'errors' {\n  class ErrorLike {}\n\n  export function createError(): ErrorLike;\n}\n",
	}
	options := OnlyThrowErrorOptions{
		Allow: []type_checking.TypeOrValueSpecifier{
			{From: type_checking.TypeOrValueSpecifierFromPackage, Name: []string{"ErrorLike"}, Package: "errors"},
		},
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	// Upstream is silent; we report, because the module does not resolve under this tsconfig.
	rule_testing.ExpectFindings(t,
		rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, files, onlyThrowErrorFile, options), "object")

	// The control: the bare inline name needs no file or package comparison, so a type that had
	// resolved to ErrorLike would go clean here. It reports, which locates the failure in module
	// resolution rather than in the specifier matcher.
	rule_testing.ExpectFindings(t,
		rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, files, onlyThrowErrorFile, OnlyThrowErrorOptions{
			AllowInline:          []string{"ErrorLike"},
			AllowThrowingAny:     type_checking.Ref(false),
			AllowThrowingUnknown: type_checking.Ref(false),
		}), "object")
}

// TestOnlyThrowErrorSurvivesMalformedShapes pins the two guards that prevent a PANIC rather than
// deciding a finding.
//
// Both were found by the mutation sweep and neither is visible to an `ExpectFindings` fixture,
// because a panic takes the whole run down rather than producing a wrong verdict: the sweep reports
// SURVIVED identically whether the guard matters or not. What each one PREVENTS is the question,
// not what it decides.
//
//	len(arguments) < 2 in the `then` arm    without it, a ONE-argument `.then` whose arrow
//	                                        parameter is thrown indexes arguments[1] and panics
//	                                        with "index out of range [1] with length 1"
//
//	IsStringLiteralLike on a computed key   without it, `o[k()](...)` calls Node.Text on a
//	                                        CallExpression, which Node.Text does not handle and
//	                                        panics on: "Unhandled case in Node.Text"
//
// Both panics were reproduced by applying the mutation and running these exact inputs, so the
// category is measured rather than argued. Neither input appears anywhere in upstream's corpus.
func TestOnlyThrowErrorSurvivesMalformedShapes(t *testing.T) {
	t.Parallel()

	strict := OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(true),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	t.Run("a one-argument then does not index past its arguments", func(t *testing.T) {
		t.Parallel()
		// `.then` puts the rejection handler second, so this arrow is the FULFILLMENT handler and
		// the throw is not a rethrow. Upstream reports it; the point here is that we get a verdict
		// at all rather than a panic.
		rule_testing.ExpectFindings(t, runOnlyThrowError(t,
			"Promise.reject('x').then(e => {\n  throw e;\n});", strict), "object")
	})

	t.Run("a computed key that is not a literal does not read text off it", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, runOnlyThrowError(t,
			"declare const o: any;\ndeclare function k(): string;\no[k()](e => {\n  throw e;\n});", strict), "object")
	})
}

// TestOnlyThrowErrorPrivateMethodName pins that a PRIVATE name counts as a static member name.
//
// Upstream's static member reader accepts `Identifier` and `PrivateIdentifier` alike, so a thenable
// class with a `#catch` method is a catch call and a throw of its handler parameter is a rethrow.
//
// This port first restricted the name to `KindIdentifier`, which reads as the careful choice and is
// a divergence. The mutation dropping that guard SURVIVED the whole fixture set, and re-reading
// upstream rather than adding a fixture is what found it: measured on the installed rule, the case
// below is SILENT upstream and the identifier-only version reported.
//
// The class is made thenable on purpose. A plain class with a `#catch` method is silent either way,
// because the receiver test declines it, so it cannot separate the two versions and would have read
// as coverage while proving nothing.
func TestOnlyThrowErrorPrivateMethodName(t *testing.T) {
	t.Parallel()

	strict := OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(true),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	rule_testing.ExpectClean(t, runOnlyThrowError(t,
		"class C {\n"+
			"  then(cb: (v: number) => void, r?: (e: any) => void): C {\n"+
			"    return this;\n"+
			"  }\n"+
			"  #catch(cb: (e: any) => void) {}\n"+
			"  m() {\n"+
			"    this.#catch(e => {\n"+
			"      throw e;\n"+
			"    });\n"+
			"  }\n"+
			"}", strict))

	// The control: the same shape on a receiver that is NOT thenable reports, which shows the case
	// above is clean because the rethrow arm fired rather than because private names are skipped.
	rule_testing.ExpectFindings(t, runOnlyThrowError(t,
		"class C {\n"+
			"  #catch(cb: (e: any) => void) {}\n"+
			"  m() {\n"+
			"    this.#catch(e => {\n"+
			"      throw e;\n"+
			"    });\n"+
			"  }\n"+
			"}", strict), "object")
}

// TestOnlyThrowErrorSpecifierComposites pins the union and intersection semantics of `allow`.
//
// These live in the shelf's `TypeMatchesSomeSpecifier`, and neither recursion was there before this
// port: `internal/utilities/typecheck/specifier.go` was written against a tsgolint revision that
// predates them, so a composite type was tested against its own symbol, which is nil, and answered
// false. Upstream's `typeMatchesSpecifier` recurses both ways and the two directions are OPPOSITE,
// which is the asymmetry a reader would most likely collapse into one.
//
// Upstream's own corpus covers exactly one row of this, valid case 36 against invalid case 36, and
// only in the bare-string form. The object-specifier rows below were measured against the installed
// rule rather than inherited, because a mutation flipping the union's `every` to `some` survived
// the entire imported corpus:
//
//	Map<string,string> | Map<number,number>   allow [{from: lib, name: Map}]   SILENT
//	Map<string,string> | Set<number>          allow [{from: lib, name: Map}]   REPORTS
//	Map<string,string> & { a: 1 }             allow [{from: lib, name: Map}]   SILENT
//
// The middle row is the one that separates `every` from `some`; without it the flip is invisible.
func TestOnlyThrowErrorSpecifierComposites(t *testing.T) {
	t.Parallel()

	allowMap := OnlyThrowErrorOptions{
		Allow:                []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"Map"}}},
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	t.Run("a union matches only when every member matches", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runOnlyThrowError(t,
			"declare const m: Map<string, string> | Map<number, number>;\nthrow m;", allowMap))

		rule_testing.ExpectFindings(t, runOnlyThrowError(t,
			"declare const m: Map<string, string> | Set<number>;\nthrow m;", allowMap), "object")
	})

	t.Run("an intersection matches when any member matches", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runOnlyThrowError(t,
			"declare const m: Map<string, string> & { a: 1 };\nthrow m;", allowMap))
	})
}

// TestOnlyThrowErrorSpecifierDeclinesTheErrorType pins that a specifier never matches the type
// checker's own ERROR type.
//
// When a module does not resolve, the checker gives the import the intrinsic error type, whose
// intrinsic name is the string `"error"`. A specifier literally named `error` would otherwise match
// it and silently allow every unresolved import in the file, which is the opposite of what an
// allowlist is for: the point of naming a type is that you know what it is.
//
// Upstream declines it inside `typeMatchesSpecifier`, and our shelf declined it only at the top
// level of `TypeMatchesSomeSpecifier`, which the recursion added for this port can now walk past.
// A mutation removing the guard survived the whole corpus. Measured on the installed rule, all
// three rows REPORT, and the middle and last are the controls showing the input reports for
// ordinary reasons too:
//
//	allow [{from: lib, name: "error"}]   REPORTS
//	allow ["error"]                      REPORTS
//	no allow                             REPORTS
//
// Under the mutation the FIRST row goes silent and the other two do not, which is why the object
// form is the one that can see this.
func TestOnlyThrowErrorSpecifierDeclinesTheErrorType(t *testing.T) {
	t.Parallel()

	source := "import { thing } from './nonexistent-module-xyz';\nthrow thing;"

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, source, OnlyThrowErrorOptions{
		Allow:                []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"error"}}},
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}), "object")

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, source, OnlyThrowErrorOptions{
		AllowInline:          []string{"error"},
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}), "object")

	rule_testing.ExpectFindings(t, runOnlyThrowError(t, source, OnlyThrowErrorOptions{
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}), "object")
}

// TestOnlyThrowErrorDeclinesSynthesizedNodes pins that a recovered parse produces no finding.
//
// `throw;` and `throw ();` are both "Expression expected" PARSE ERRORS for upstream's parser, so
// its rule never runs on either and there is no upstream verdict to be faithful to. Our parser
// recovers, synthesizing a zero-width `KindIdentifier` whose type is `any`, so both inputs reach
// the rule. With `allowThrowingAny` at its permissive default they are silent through the `any`
// arm; with it off, a finding would be emitted with a ZERO-WIDTH span pointing at text the author
// never wrote.
//
// The second case is why the guard sits after the parenthesis skip. `throw;` puts the synthesized
// node directly under the statement, so a check on the raw argument catches it; `throw ();` wraps
// it in a real parenthesized expression which is NOT itself missing, so the same check misses.
// Measured before the guard moved: `throw ();` reported a zero-width `object` at offset 7 while
// `throw;` was clean.
//
// Neither input is in any corpus, on either side, because neither parses upstream.
func TestOnlyThrowErrorDeclinesSynthesizedNodes(t *testing.T) {
	t.Parallel()

	strict := OnlyThrowErrorOptions{
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}

	for _, source := range []string{"throw;", "throw ();", "throw (());"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runOnlyThrowError(t, source, strict))
			rule_testing.ExpectClean(t, runOnlyThrowError(t, source, nil))
		})
	}
}

// TestOnlyThrowErrorUndeclaredIdentifier pins an input on which UPSTREAM CRASHES and we do not.
//
// `throw notDeclaredAnywhere;` resolves to no symbol. Upstream's rethrow check reaches for
// `findVariable(scope, node)` and wraps it in `nullThrows(..., 'Variable ... should exist in scope
// manager')`, so an identifier that resolves to nothing takes the whole lint run down. Measured on
// the installed rule, both with `allowRethrowing: true` and at the defaults:
//
//	Non-null Assertion Failed: Variable notDeclaredAnywhere should exist in scope manager
//
// This is a defect upstream ships, not a decision, so there is no verdict to reproduce. The port's
// nil check answers "not a rethrow" and lets the type arms decide, which reports `object` because
// the unresolved identifier's type is the error type. That is the only defensible answer available.
//
// It is also a real blind spot the sweep found rather than an unreachable branch. Enumerated over
// ten malformed and unresolvable shapes, `GetSymbolAtLocation` returned nil for a NON-missing
// identifier only on this one, so nothing else in the fixture set can reach the guard.
//
// The DEFAULT case is silent, and that is worth stating because it is not what a first reading
// predicts. An unresolved identifier gets the checker's error type, whose flags report `any`, so
// the permissive `allowThrowingAny` default swallows it before the `object` arm is reached. Only
// with that option off does the finding appear, which is why the strict row is the one that puts
// the nil guard under test.
func TestOnlyThrowErrorUndeclaredIdentifier(t *testing.T) {
	t.Parallel()

	strict := OnlyThrowErrorOptions{
		AllowRethrowing:      type_checking.Ref(true),
		AllowThrowingAny:     type_checking.Ref(false),
		AllowThrowingUnknown: type_checking.Ref(false),
	}
	rule_testing.ExpectFindings(t, runOnlyThrowError(t, "throw notDeclaredAnywhere;", strict), "object")

	// Silent at the defaults: the error type reports the `any` flag.
	rule_testing.ExpectClean(t, runOnlyThrowError(t, "throw notDeclaredAnywhere;", nil))
}

// TestOnlyThrowErrorDependsOnAnotherFile pins what the rule reads, and proves the cross-file claim the
// findings cache has to honor.
//
// The cross-file case is measured: a class declared in a second module extending the built-in
// `Error`, thrown from this one. `IsErrorLike` walks the base types and asks whether each
// declaration's source file is a default library, which this file's bytes cannot answer. Flipping
// only the OTHER file, from `extends Error` to no heritage, flips the verdict here while this file is
// untouched. The cache honors that through the type fingerprint, which covers other.ts as an import,
// and the default library through the binary in its key; what the rule reads through ctx.Program is
// the compiler options and the default library, and nothing else (#b8k3bp6).
func TestOnlyThrowErrorDependsOnAnotherFile(t *testing.T) {
	t.Parallel()

	if OnlyThrowError.ProgramReads != rule.ReadsCompilerOptions|rule.ReadsDefaultLibrary {
		t.Fatalf("OnlyThrowError must declare what IsErrorLike and the allow specifiers read; declares %s", OnlyThrowError.ProgramReads)
	}

	subject := "import { Wrapped } from './other';\nthrow new Wrapped();"

	// The other file's class extends Error, so the throw is fine.
	rule_testing.ExpectClean(t, rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, map[string]string{
		onlyThrowErrorFile: subject,
		"other.ts":         "export class Wrapped extends Error {}\n",
	}, onlyThrowErrorFile, nil))

	// Byte-for-byte the same subject file, and the verdict moves because the OTHER file changed.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedFilesWithOptions(t, OnlyThrowError, map[string]string{
		onlyThrowErrorFile: subject,
		"other.ts":         "export class Wrapped {}\n",
	}, onlyThrowErrorFile, nil), "object")
}
