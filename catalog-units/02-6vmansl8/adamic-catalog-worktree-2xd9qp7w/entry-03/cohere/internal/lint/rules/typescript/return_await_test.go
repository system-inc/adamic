package typescript

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const returnAwaitFile = "/repository/source/Returning.ts"

func returnAwaitCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// returnAwaitOptionsFor routes a case through the rule's own decoder.
//
// Building the options struct directly would leave the decoder untested, and the decoder is where
// the DEFAULT lives. That default is `in-try-catch` rather than `always`, so a rule configured as
// a plain "error" arrives with nil options and must land on it; the zero value is an empty mode
// matching no arm, which would make the rule silently inert.
func returnAwaitOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" || optionsJson == "null" {
		return DefaultReturnAwaitSettings()
	}
	decoded, err := DecodeReturnAwaitOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestReturnAwaitStaysSilent is upstream's fifty one passing cases verbatim.
//
// Each carries its own option, because this rule's verdict for one input moves with the mode: the
// same source is valid under `never` and reports under `always`. Measured one file per program
// against the installed 8.67.0 build, using upstream's own fixture compilerOptions verbatim.
func TestReturnAwaitStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{
			sourceText:  "return;\n",
			optionsJson: "",
		},
		{
			sourceText:  "function test() {\n  return;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "function test() {\n  return 1;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  return;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  return 1;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "const test = () => 1;\n",
			optionsJson: "",
		},
		{
			sourceText:  "const test = async () => 1;\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  return Promise.resolve(1);\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  try {\n    return await Promise.resolve(1);\n  } catch (e) {\n    return await Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "const fn = (): any => null;\nasync function test() {\n  return await fn();\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "const fn = (): unknown => null;\nasync function test() {\n  return await fn();\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test(unknownParam: unknown) {\n  try {\n    return await unknownParam;\n  } finally {\n    console.log('In finally block');\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  if (Math.random() < 0.33) {\n    return await Promise.resolve(1);\n  } else if (Math.random() < 0.5) {\n    return Promise.resolve(2);\n  }\n\n  try {\n  } catch (e) {\n    return await Promise.resolve(3);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"error-handling-correctness-only\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    const one = await Promise.resolve(1);\n    return one;\n  } catch (e) {\n    const two = await Promise.resolve(2);\n    return two;\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "function test() {\n  return 1;\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  return 1;\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "const test = () => 1;\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "const test = async () => 1;\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  return Promise.resolve(1);\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    return await Promise.resolve(1);\n  } catch (e) {\n    return await Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    throw 'foo';\n  } catch (e) {\n    return Promise.resolve(1);\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    throw 'foo';\n  } catch (e) {\n    throw 'foo2';\n  } finally {\n    return Promise.resolve(1);\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    const one = await Promise.resolve(1);\n    return one;\n  } catch (e) {\n    const two = await Promise.resolve(2);\n    return two;\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "async function test() {\n  return Promise.resolve(1);\n}\n",
			optionsJson: "\"never\"",
		},
		{
			sourceText:  "const test = async () => Promise.resolve(1);\n",
			optionsJson: "\"never\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"never\"",
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "\"always\"",
		},
		{
			sourceText:  "const test = async () => await Promise.resolve(1);\n",
			optionsJson: "\"always\"",
		},
		{
			sourceText:  "async function test() {\n  try {\n    return await Promise.resolve(1);\n  } catch (e) {\n    return await Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"always\"",
		},
		{
			sourceText:  "declare function foo(): Promise<boolean>;\n\nfunction bar(baz: boolean): Promise<boolean> | boolean {\n  if (baz) {\n    return true;\n  } else {\n    return foo();\n  }\n}\n",
			optionsJson: "\"always\"",
		},
		{
			sourceText:  "async function test(): Promise<string> {\n  const res = await Promise.resolve('{}');\n  try {\n    return JSON.parse(res);\n  } catch (error) {\n    return res;\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function test() {\n  const res = await Promise.resolve('{}');\n  try {\n    async function nested() {\n      return Promise.resolve('ok');\n    }\n    return await nested();\n  } catch (error) {\n    return Promise.resolve('error');\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function f() {\n  try {\n  } catch {\n    try {\n    } catch {\n      return Promise.reject();\n    }\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function f() {\n  try {\n  } finally {\n    try {\n    } catch {\n      return Promise.reject();\n    }\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function f() {\n  try {\n  } finally {\n    try {\n    } finally {\n      try {\n      } catch {\n        return Promise.reject();\n      }\n    }\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  using something = bleh;\n  return await Promise.resolve(2);\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  await using something = bleh;\n  return await Promise.resolve(2);\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  using something = bleh;\n  {\n    return await Promise.resolve(2);\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  return Promise.resolve(2);\n  using something = bleh;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  return await Promise.resolve(2);\n  using something = bleh;\n}\n",
			optionsJson: "\"always\"",
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function returnAwait() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  return await asyncFn();\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  async function innerFunction() {\n    return asyncFn();\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  const innerFunction = async () => asyncFn();\n}\n",
			optionsJson: "\"in-try-catch\"",
		},
		{
			sourceText:  "using foo = 1 as any;\nreturn Promise.resolve(42);\n",
			optionsJson: "",
		},
		{
			sourceText:  "{\n  using foo = 1 as any;\n  return Promise.resolve(42);\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function wrapper<T>(value: T) {\n  return await value;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function wrapper<T extends unknown>(value: T) {\n  return await value;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "async function wrapper<T extends any>(value: T) {\n  return await value;\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "class C<T> {\n  async wrapper<T>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "class C<R> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
		},
		{
			sourceText:  "class C<R extends unknown> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
		},
	}
	for index, testCase := range cases {
		t.Run(returnAwaitCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, ReturnAwait,
				returnAwaitFile, testCase.sourceText,
				returnAwaitOptionsFor(t, testCase.optionsJson)))
		})
	}
}

// returnAwaitFinding is one expected finding with every layer it can be wrong at.
type returnAwaitFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestReturnAwaitFires is upstream's fifty two reporting cases verbatim.
//
// Every row asserts the span, the id and the whole rendered message. The id carries the judgment:
// `requiredPromiseAwait` and `disallowedPromiseAwait` are opposite verdicts on the same syntax,
// separated only by whether the position affects error handling, and `nonPromiseAwait` is a third
// answer about the type rather than the context.
//
// Cases whose repair upstream ships as a FIX assert the rewritten file in TestReturnAwaitFixes
// below. Cases whose repair is a SUGGESTION are asserted separately, because the harness applies
// fixes and not suggestions, and collapsing the two would have the edit engine rewriting which
// frame catches a rejection.
func TestReturnAwaitFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		optionsJson  string
		wantFindings []returnAwaitFinding
	}{
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  const foo = 1;\n  return await { foo };\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { foo }",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  const foo = 1;\n  return await foo;\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await foo",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await 1;\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await { a: 1 };\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { a: 1 }",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await { a: 1 }.a;\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { a: 1 }.a",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await ({ a: 1 });\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await ({ a: 1 })",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => (await { a: 1 });\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { a: 1 }",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "declare const cond: boolean;\nconst test = async () => (cond ? await { a: 1 } : 2);\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { a: 1 }",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await /* comment */ 1;\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await /* comment */ 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await /* comment */ { a: 1 };\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await /* comment */ { a: 1 }",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await Promise.resolve(1);\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await { then(cb: () => void) {} };\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await { then(cb: () => void) {} }",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"error-handling-correctness-only\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(1)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(1)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(1)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await 1;\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "const test = async () => await Promise.resolve(1);\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"never\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  try {\n    return await Promise.resolve(1);\n  } catch (e) {\n    return await Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson: "\"never\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
				{
					wantSpan:    "await Promise.resolve(2)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "\"never\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve(1)",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  return Promise.resolve(1);\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(1)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "const test = async () => Promise.resolve(1);\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(1)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo()) ? bar() : baz();\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "baz()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo())\n    ? (\n      bar ? bar() : baz()\n    ) : baz ? baz() : bar();\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "baz()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "baz()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function buzz() {\n  return (await foo()) ? await 1 : bar();\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nconst buzz = async () => ((await foo()) ? bar() : baz());\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
				{
					wantSpan:    "baz()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nconst buzz = async () => ((await foo()) ? await 1 : bar());\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await 1",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
				{
					wantSpan:    "bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test<T>(): Promise<T> {\n  const res = await fetch('...');\n  try {\n    return res.json() as Promise<T>;\n  } catch (err) {\n    throw Error('Request Failed.');\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "res.json() as Promise<T>",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  try {\n    const callback1 = function () {};\n    const callback2 = async function () {};\n    function callback3() {}\n    async function callback4() {}\n    const callback5 = () => {};\n    const callback6 = async () => {};\n    return Promise.resolve('try');\n  } finally {\n    return Promise.resolve('finally');\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve('try')",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function bar() {}\nasync function foo() {\n  try {\n    return undefined || bar();\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "undefined || bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function bar() {}\nasync function foo() {\n  try {\n    return bar() || undefined || bar();\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "bar() || undefined || bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function bar() {}\nasync function func1() {\n  try {\n    return null ?? bar();\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "null ?? bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function bar() {}\nasync function func2() {\n  try {\n    return 1 && bar();\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "1 && bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "const foo = {\n  bar: async function () {},\n};\nasync function func3() {\n  try {\n    return foo.bar();\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "foo.bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "class X {\n  async bar() {\n    return;\n  }\n  async func2() {\n    try {\n      return this.bar();\n    } catch {}\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "this.bar()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function test() {\n  const res = await Promise.resolve('{}');\n  try {\n    async function nested() {\n      return Promise.resolve('ok');\n    }\n    return await nested();\n  } catch (error) {\n    return await Promise.resolve('error');\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await Promise.resolve('error')",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function f() {\n  try {\n    try {\n    } finally {\n      // affects error handling of outer catch\n      return Promise.reject();\n    }\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.reject()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function f() {\n  try {\n    try {\n    } catch {\n      // affects error handling of outer catch\n      return Promise.reject();\n    }\n  } catch {}\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.reject()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "async function f() {\n  try {\n  } catch {\n    try {\n    } finally {\n      try {\n      } catch {\n        return Promise.reject();\n      }\n    }\n  } finally {\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.reject()",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    using something = bleh;\n    if (anotherCondition) {\n      return Promise.resolve(2);\n    }\n  }\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    await using something = bleh;\n    if (anotherCondition) {\n      return Promise.resolve(2);\n    }\n  }\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    using something = bleh;\n  } else if (anotherCondition) {\n    return Promise.resolve(2);\n  }\n}\n",
			optionsJson: "\"always\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "Promise.resolve(2)",
					wantId:      "requiredPromiseAwait",
					wantMessage: "Returning an awaited promise is required in this context.",
				},
			},
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  async function innerFunction() {\n    return await asyncFn();\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await asyncFn()",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  const innerFunction = async () => await asyncFn();\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await asyncFn()",
					wantId:      "disallowedPromiseAwait",
					wantMessage: "Returning an awaited promise is not allowed in this context.",
				},
			},
		},
		{
			sourceText:  "async function wrapper<T extends number>(value: T) {\n  return await value;\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await value",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "class C<T> {\n  async wrapper<T extends string>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await value",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
		{
			sourceText:  "class C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
			wantFindings: []returnAwaitFinding{
				{
					wantSpan:    "await value",
					wantId:      "nonPromiseAwait",
					wantMessage: "Returning an awaited value that is not a promise is not allowed.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(returnAwaitCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ReturnAwait, returnAwaitFile,
				testCase.sourceText, returnAwaitOptionsFor(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so span slices are against that text rather
			// than the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Id != want.wantId {
					t.Fatalf("finding %d id: expected %q, got %q", position, want.wantId,
						diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestReturnAwaitFixes asserts the repair on every case upstream applies unattended.
//
// These go through rule_testing.ExpectFixedSource, which replays the fixes into the source and
// compares the WHOLE rewritten file, so a repair that edits the right span with the wrong text or
// the wrong span with the right text both fail here and neither is visible to a message id.
//
// Four rows below take their expected output from the CLONE rather than from the installed build,
// and they are the only four where the two disagree. Enumerated rather than assumed: every other
// fix vector is byte-identical between 8.67 and 8.68.
//
// The parenthesis rows are the reason this port follows the 8.68 clone rather than the installed
// 8.67 build. Removing `await` from an arrow body that begins with `{` leaves source parsing as a
// Block rather than an object literal, so the function silently stops returning anything. The
// installed build produces exactly that; these rows assert the corrected output.
func TestReturnAwaitFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
		wantFixed   string
	}{
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "",
			wantFixed:   "async function test() {\n  return 1;\n}\n",
		},
		{
			sourceText:  "async function test() {\n  const foo = 1;\n  return await { foo };\n}\n",
			optionsJson: "",
			wantFixed:   "async function test() {\n  const foo = 1;\n  return { foo };\n}\n",
		},
		{
			sourceText:  "async function test() {\n  const foo = 1;\n  return await foo;\n}\n",
			optionsJson: "",
			wantFixed:   "async function test() {\n  const foo = 1;\n  return foo;\n}\n",
		},
		{
			sourceText:  "const test = async () => await 1;\n",
			optionsJson: "",
			wantFixed:   "const test = async () => 1;\n",
		},
		{
			sourceText:  "const test = async () => await { a: 1 };\n",
			optionsJson: "",
			wantFixed:   "const test = async () => ({ a: 1 });\n",
		},
		{
			sourceText:  "const test = async () => await { a: 1 }.a;\n",
			optionsJson: "",
			wantFixed:   "const test = async () => ({ a: 1 }.a);\n",
		},
		{
			sourceText:  "const test = async () => await ({ a: 1 });\n",
			optionsJson: "",
			wantFixed:   "const test = async () => ({ a: 1 });\n",
		},
		{
			sourceText:  "const test = async () => (await { a: 1 });\n",
			optionsJson: "",
			wantFixed:   "const test = async () => ({ a: 1 });\n",
		},
		{
			sourceText:  "declare const cond: boolean;\nconst test = async () => (cond ? await { a: 1 } : 2);\n",
			optionsJson: "",
			wantFixed:   "declare const cond: boolean;\nconst test = async () => (cond ? { a: 1 } : 2);\n",
		},
		{
			sourceText:  "const test = async () => await /* comment */ 1;\n",
			optionsJson: "",
			wantFixed:   "const test = async () => /* comment */ 1;\n",
		},
		{
			sourceText:  "const test = async () => await /* comment */ { a: 1 };\n",
			optionsJson: "",
			wantFixed:   "const test = async () => /* comment */ ({ a: 1 });\n",
		},
		{
			sourceText:  "const test = async () => await Promise.resolve(1);\n",
			optionsJson: "",
			wantFixed:   "const test = async () => Promise.resolve(1);\n",
		},
		{
			sourceText:  "const test = async () => await { then(cb: () => void) {} };\n",
			optionsJson: "",
			wantFixed:   "const test = async () => ({ then(cb: () => void) {} });\n",
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "",
			wantFixed:   "async function test() {\n  return Promise.resolve(1);\n}\n",
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "async function test() {\n  return 1;\n}\n",
		},
		{
			sourceText:  "const test = async () => await 1;\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "const test = async () => 1;\n",
		},
		{
			sourceText:  "const test = async () => await Promise.resolve(1);\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "const test = async () => Promise.resolve(1);\n",
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "async function test() {\n  return Promise.resolve(1);\n}\n",
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"never\"",
			wantFixed:   "async function test() {\n  return 1;\n}\n",
		},
		{
			sourceText:  "async function test() {\n  return await Promise.resolve(1);\n}\n",
			optionsJson: "\"never\"",
			wantFixed:   "async function test() {\n  return Promise.resolve(1);\n}\n",
		},
		{
			sourceText:  "async function test() {\n  return await 1;\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function test() {\n  return 1;\n}\n",
		},
		{
			sourceText:  "async function test() {\n  return Promise.resolve(1);\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function test() {\n  return await Promise.resolve(1);\n}\n",
		},
		{
			sourceText:  "const test = async () => Promise.resolve(1);\n",
			optionsJson: "\"always\"",
			wantFixed:   "const test = async () => await Promise.resolve(1);\n",
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo()) ? bar() : baz();\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo()) ? await bar() : await baz();\n}\n",
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo())\n    ? (\n      bar ? bar() : baz()\n    ) : baz ? baz() : bar();\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nasync function qux() {}\nasync function buzz() {\n  return (await foo())\n    ? (\n      bar ? await bar() : await baz()\n    ) : baz ? await baz() : await bar();\n}\n",
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function buzz() {\n  return (await foo()) ? await 1 : bar();\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function foo() {}\nasync function bar() {}\nasync function buzz() {\n  return (await foo()) ? 1 : await bar();\n}\n",
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nconst buzz = async () => ((await foo()) ? bar() : baz());\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function foo() {}\nasync function bar() {}\nasync function baz() {}\nconst buzz = async () => ((await foo()) ? await bar() : await baz());\n",
		},
		{
			sourceText:  "async function foo() {}\nasync function bar() {}\nconst buzz = async () => ((await foo()) ? await 1 : bar());\n",
			optionsJson: "\"always\"",
			wantFixed:   "async function foo() {}\nasync function bar() {}\nconst buzz = async () => ((await foo()) ? 1 : await bar());\n",
		},
		{
			sourceText:  "async function test() {\n  const res = await Promise.resolve('{}');\n  try {\n    async function nested() {\n      return Promise.resolve('ok');\n    }\n    return await nested();\n  } catch (error) {\n    return await Promise.resolve('error');\n  }\n}\n",
			optionsJson: "",
			wantFixed:   "async function test() {\n  const res = await Promise.resolve('{}');\n  try {\n    async function nested() {\n      return Promise.resolve('ok');\n    }\n    return await nested();\n  } catch (error) {\n    return Promise.resolve('error');\n  }\n}\n",
		},
		{
			sourceText:  "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    using something = bleh;\n  } else if (anotherCondition) {\n    return Promise.resolve(2);\n  }\n}\n",
			optionsJson: "\"always\"",
			wantFixed:   "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    using something = bleh;\n  } else if (anotherCondition) {\n    return await Promise.resolve(2);\n  }\n}\n",
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  async function innerFunction() {\n    return await asyncFn();\n  }\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  async function innerFunction() {\n    return asyncFn();\n  }\n}\n",
		},
		{
			sourceText:  "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  const innerFunction = async () => await asyncFn();\n}\n",
			optionsJson: "\"in-try-catch\"",
			wantFixed:   "declare function asyncFn(): Promise<unknown>;\nasync function outerFunction() {\n  using _ = {\n    [Symbol.dispose]: () => {\n      console.log('dispose');\n    },\n  };\n\n  const innerFunction = async () => asyncFn();\n}\n",
		},
		{
			sourceText:  "async function wrapper<T extends number>(value: T) {\n  return await value;\n}\n",
			optionsJson: "",
			wantFixed:   "async function wrapper<T extends number>(value: T) {\n  return value;\n}\n",
		},
		{
			sourceText:  "class C<T> {\n  async wrapper<T extends string>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
			wantFixed:   "class C<T> {\n  async wrapper<T extends string>(value: T) {\n    return value;\n  }\n}\n",
		},
		{
			sourceText:  "class C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n",
			optionsJson: "",
			wantFixed:   "class C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return value;\n  }\n}\n",
		},
	}
	for index, testCase := range cases {
		t.Run(returnAwaitCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ReturnAwait, returnAwaitFile,
				testCase.sourceText, returnAwaitOptionsFor(t, testCase.optionsJson))

			// The harness writes the fixture trimmed, so the expected rewrite is transformed the
			// same way the input was rather than compared against the Go literal.
			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}

// TestReturnAwaitOffersSuggestionsRatherThanFixes covers the half the harness cannot apply.
//
// Where the position affects error handling, upstream ships the same edit as a SUGGESTION rather
// than a fix, because adding or removing the await changes which frame catches a rejection. That
// distinction is the whole safety property of this rule: shipping these as fixes would have the
// edit engine silently rewriting control flow.
//
// ExpectFixedSource cannot see a suggestion, so these rows assert that the finding carries a
// suggestion and NO fix, plus the suggestion text. A row that arrived as a fix would pass every
// message-id assertion in the firing test above.
func TestReturnAwaitOffersSuggestionsRatherThanFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		optionsJson     string
		wantSuggestions []string
	}{
		{
			sourceText:      "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson:     "\"error-handling-correctness-only\"",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow.", "Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson:     "\"always\"",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow.", "Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function test() {\n  try {\n    return Promise.resolve(1);\n  } catch (e) {\n    return Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson:     "\"in-try-catch\"",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow.", "Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function test() {\n  try {\n    return await Promise.resolve(1);\n  } catch (e) {\n    return await Promise.resolve(2);\n  } finally {\n    console.log('cleanup');\n  }\n}\n",
			optionsJson:     "\"never\"",
			wantSuggestions: []string{"Remove `await` before the expression. Use caution as this may impact control flow.", "Remove `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function test<T>(): Promise<T> {\n  const res = await fetch('...');\n  try {\n    return res.json() as Promise<T>;\n  } catch (err) {\n    throw Error('Request Failed.');\n  }\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function test() {\n  try {\n    const callback1 = function () {};\n    const callback2 = async function () {};\n    function callback3() {}\n    async function callback4() {}\n    const callback5 = () => {};\n    const callback6 = async () => {};\n    return Promise.resolve('try');\n  } finally {\n    return Promise.resolve('finally');\n  }\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function bar() {}\nasync function foo() {\n  try {\n    return undefined || bar();\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function bar() {}\nasync function foo() {\n  try {\n    return bar() || undefined || bar();\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function bar() {}\nasync function func1() {\n  try {\n    return null ?? bar();\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function bar() {}\nasync function func2() {\n  try {\n    return 1 && bar();\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "const foo = {\n  bar: async function () {},\n};\nasync function func3() {\n  try {\n    return foo.bar();\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "class X {\n  async bar() {\n    return;\n  }\n  async func2() {\n    try {\n      return this.bar();\n    } catch {}\n  }\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function f() {\n  try {\n    try {\n    } finally {\n      // affects error handling of outer catch\n      return Promise.reject();\n    }\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function f() {\n  try {\n    try {\n    } catch {\n      // affects error handling of outer catch\n      return Promise.reject();\n    }\n  } catch {}\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "async function f() {\n  try {\n  } catch {\n    try {\n    } finally {\n      try {\n      } catch {\n        return Promise.reject();\n      }\n    }\n  } finally {\n  }\n}\n",
			optionsJson:     "",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    using something = bleh;\n    if (anotherCondition) {\n      return Promise.resolve(2);\n    }\n  }\n}\n",
			optionsJson:     "\"always\"",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
		{
			sourceText:      "declare const bleh: any;\nasync function f() {\n  if (cond) {\n    await using something = bleh;\n    if (anotherCondition) {\n      return Promise.resolve(2);\n    }\n  }\n}\n",
			optionsJson:     "\"always\"",
			wantSuggestions: []string{"Add `await` before the expression. Use caution as this may impact control flow."},
		},
	}
	for index, testCase := range cases {
		t.Run(returnAwaitCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ReturnAwait, returnAwaitFile,
				testCase.sourceText, returnAwaitOptionsFor(t, testCase.optionsJson))

			var got []string
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("finding %q carries a fix; upstream offers only a suggestion here, and a "+
						"fix would let the engine rewrite control flow unattended", diagnostic.Message.Id)
				}
				for _, suggestion := range diagnostic.Suggestions {
					got = append(got, suggestion.Message.Description)
				}
			}
			if len(got) != len(testCase.wantSuggestions) {
				t.Fatalf("expected %d suggestions %q, got %d %q", len(testCase.wantSuggestions),
					testCase.wantSuggestions, len(got), got)
			}
			for position, want := range testCase.wantSuggestions {
				if got[position] != want {
					t.Fatalf("suggestion %d: expected %q, got %q", position, want, got[position])
				}
			}
		})
	}
}
