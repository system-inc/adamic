package typescript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for use-unknown-in-catch-callback-variable, taken verbatim from upstream's own tester.
//
// Every case below was extracted by PARSING the upstream test file with the TypeScript compiler and
// serialising each `code` and `output` string through a JSON encoder, so no source string was ever
// retyped or passed through a shell. Each extracted string was then byte-compared against the
// upstream file it came from, with a control string asserted absent, and all 57 cases plus all 26
// suggestion outputs matched. That matters more than usual here: several cases turn on a comment
// sitting in an awkward place, and one turns on a `?` between a name and its annotation, both of
// which a hand-copy silently normalises away.
//
// The whole corpus was additionally driven through the INSTALLED rule at 8.67.0 on a real type
// graph, and its verdicts agreed with upstream's recorded expectations on all 57 inputs and all 26
// suggestion outputs. So the oracle used to settle every question below is itself pinned against
// the corpus rather than trusted.
//
// Upstream's clone is at 8.68.0 and the installed build at 8.67.0. Their metadata agrees exactly:
// `schema: []`, `hasSuggestions` with no `fixable`, and the same seven message ids.
const useUnknownInCatchCallbackVariableFile = "file.ts"

// TestUseUnknownInCatchCallbackVariableStaysSilent is upstream's valid list.
//
// Fourteen of these are spread arguments, several carrying a badly typed handler inside the spread
// tuple. They are the reason the rule declines a call it cannot read positionally rather than
// looking inside, and a port that resolved the spread would report on inputs upstream passes.
func TestUseUnknownInCatchCallbackVariableStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid 0", "\nPromise.resolve().catch((err: unknown) => {\n  throw err;\n});\n    "},
		{"upstream valid 1", "\nlet x = Math.random() ? 'ca' + 'tch' : 'catch';\nPromise.resolve()[x]((err: Error) => {});\n    "},
		{"upstream valid 2", "\nPromise.resolve().then(\n  () => {},\n  (err: unknown) => {\n    throw err;\n  },\n);\n    "},
		{"upstream valid 3", "\nPromise.resolve().catch(() => {\n  throw new Error();\n});\n    "},
		{"upstream valid 4", "\nPromise.reject(new Error()).catch(\"not this rule's problem\");\n    "},
		{"upstream valid 5", "\ndeclare const crappyHandler: (() => void) | 2;\nPromise.reject(new Error()).catch(crappyHandler);\n    "},
		{"upstream valid 6", "\nPromise.resolve().catch((...args: [unknown]) => {\n  throw args[0];\n});\n    "},
		{"upstream valid 7", "\nPromise.resolve().catch((...args: [a: unknown]) => {\n  const err = args[0];\n});\n    "},
		{"upstream valid 8", "\nPromise.resolve().catch((...args: readonly unknown[]) => {\n  throw args[0];\n});\n    "},
		{"upstream valid 9", "\ndeclare const notAPromise: { catch: (f: Function) => void };\nnotAPromise.catch((...args: [a: string, string]) => {\n  throw args[0];\n});\n    "},
		{"upstream valid 10", "\ndeclare const catchArgs: [(x: unknown) => void];\nPromise.reject(new Error()).catch(...catchArgs);\n    "},
		{"upstream valid 11", "\ndeclare const catchArgs: [\n  string | (() => never),\n  (shouldntFlag: string) => void,\n  number,\n];\nPromise.reject(new Error()).catch(...catchArgs);\n    "},
		{"upstream valid 12", "\ndeclare const catchArgs: ['not callable'];\nPromise.reject(new Error()).catch(...catchArgs);\n    "},
		{"upstream valid 13", "\ndeclare const emptySpread: [];\nPromise.reject(new Error()).catch(...emptySpread);\n    "},
		{"upstream valid 14", "\nPromise.resolve().catch(\n  (\n    ...err: [unknown, string | ((number | unknown) & { b: () => void }), string]\n  ) => {\n    throw err;\n  },\n);\n    "},
		{"upstream valid 15", "\ndeclare const notAMemberExpression: (...args: any[]) => {};\nnotAMemberExpression(\n  'This helps get 100% code cov',\n  \"but doesn't test anything useful related to the rule.\",\n);\n    "},
		{"upstream valid 16", "\nPromise.resolve().catch((...[args]: [unknown]) => {\n  console.log(args);\n});\n    "},
		{"upstream valid 17", "\nPromise.resolve().catch((...{ find }: [unknown]) => {\n  console.log(find);\n});\n    "},
		{"upstream valid 18", "Promise.resolve.then();"},
		{"upstream valid 19", "Promise.resolve().then(() => {});"},
		{"upstream valid 20", "\ndeclare const singleTupleArg: [() => void];\nPromise.resolve().then(...singleTupleArg, (error: unknown) => {});\n    "},
		{"upstream valid 21", "\ndeclare const arrayArg: (() => void)[];\nPromise.resolve().then(...arrayArg, error => {});\n    "},
		{"upstream valid 22", "\ndeclare let iPromiseImAPromise: Promise<any>;\ndeclare const catchArgs: [(x: any) => void];\niPromiseImAPromise.catch(...catchArgs);\n    "},
		{"upstream valid 23", "\ndeclare const catchArgs: [\n  string | (() => never) | ((x: string) => void),\n  number,\n];\nPromise.reject(new Error()).catch(...catchArgs);\n    "},
		{"upstream valid 24", "\ndeclare const you: [];\ndeclare const cannot: [];\ndeclare const fool: [];\ndeclare const me: [(x: Error) => void] | undefined;\nPromise.resolve(undefined).catch(...you, ...cannot, ...fool, ...me!);\n    "},
		{"upstream valid 25", "\ndeclare const really: undefined[];\ndeclare const dumb: [];\ndeclare const code: (x: Error) => void;\nPromise.resolve(undefined).catch(...really, ...dumb, code);\n    "},
		{"upstream valid 26", "\ndeclare const x: ((x: any) => string)[];\nPromise.resolve('string promise').catch(...x);\n    "},
		{"upstream valid 27", "\ndeclare const x: any;\nPromise.resolve().catch(...x);\n    "},
		{"upstream valid 28", "\ndeclare const thenArgs: [() => {}, (err: any) => {}];\nPromise.resolve().then(...thenArgs);\n    "},
		{"upstream valid 29", "\ndeclare const yoloHandler: (x: any) => void;\nPromise.reject(new Error('I will reject!')).catch(yoloHandler);\n    "},
		{"upstream valid 30", "\ntype InvalidHandler = (arg: any) => void;\nPromise.resolve().catch(<InvalidHandler>(\n  function (err /* awkward spot for comment */) {\n    throw err;\n  }\n));\n    "}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestUseUnknownInCatchCallbackVariableFires is upstream's invalid list, with the message ids it
// records for each case.
//
// The last two cases report more than once, and they are the ones that pin the expression walk: a
// conditional reports both branches, and the three-finding case reaches its callbacks through a
// nullish coalescing, a logical and, a logical or, and a parenthesis around each.
func TestUseUnknownInCatchCallbackVariableFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"upstream invalid 0", "\nPromise.resolve().catch((err: Error) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 2", "\nPromise.resolve().catch((e, ...rest: []) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 3", "\nPromise.resolve().catch(\n  (err: string | ((number | unknown) & { b: () => void })) => {\n    throw err;\n  },\n);\n      ", []string{"useUnknown"}},
		{"upstream invalid 4", "\nPromise.resolve().catch(\n  (\n    ...err: [\n      unknown[],\n      string | ((number | unknown) & { b: () => void }),\n      string,\n    ]\n  ) => {\n    throw err;\n  },\n);\n      ", []string{"useUnknown"}},
		{"upstream invalid 5", "\nPromise.resolve().catch(function (err: string) {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 6", "\nPromise.resolve().catch(function (err /* awkward spot for comment */) {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 7", "\nPromise.resolve().catch(function namedCallback(err: string) {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 8", "\nPromise.resolve().catch(err => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 9", "\nPromise.resolve().then(\n  () => {},\n  error => {},\n);\n      ", []string{"useUnknown"}},
		{"upstream invalid 10", "\nPromise.resolve().catch((err?) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 11", "\nPromise.resolve().catch((err?: string) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 12", "\nPromise.resolve().catch(err/* with comment */=> {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 13", "\nPromise.resolve().catch((err = 2) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 14", "\nPromise.resolve().catch((err: any /* comment 1 */ = /* comment 2 */ 2) => {\n  throw err;\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 15", "\nPromise.resolve().catch((...args) => {\n  throw args[0];\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 16", "\nPromise.reject(new Error('I will reject!')).catch(([err]: [unknown]) => {\n  console.log(err);\n});\n      ", []string{"useUnknownArrayDestructuringPattern"}},
		{"upstream invalid 17", "\nPromise.resolve(' a string ').catch(\n  (a: any, b: () => any, c: (x: string & number) => void) => {},\n);\n      ", []string{"useUnknown"}},
		{"upstream invalid 18", "\nPromise.resolve('object destructuring').catch(({}) => {});\n      ", []string{"useUnknownObjectDestructuringPattern"}},
		{"upstream invalid 19", "\nPromise.resolve('object destructuring').catch(function ({ gotcha }) {\n  return null;\n});\n      ", []string{"useUnknownObjectDestructuringPattern"}},
		{"upstream invalid 20", "\nPromise.resolve()['catch']((x: any) => 'return');\n      ", []string{"useUnknown"}},
		{"upstream invalid 21", "\nPromise.reject().catch((...x: any) => {});\n      ", []string{"useUnknown"}},
		{"upstream invalid 22", "\nPromise.resolve().catch((...[args]: [string]) => {\n  console.log(args);\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 23", "\nPromise.resolve().catch((...{ find }: [string]) => {\n  console.log(find);\n});\n      ", []string{"useUnknown"}},
		{"upstream invalid 24", "\ndeclare const condition: boolean;\nPromise.resolve('foo').then(() => {}, condition ? err => {} : err => {});\n      ", []string{"useUnknown", "useUnknown"}},
		{"upstream invalid 25", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && (err => {})) || (err => {}),\n);\n      ", []string{"useUnknown", "useUnknown", "useUnknown"}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestUseUnknownInCatchCallbackVariableSuggestions applies each suggested repair and compares the
// whole rewritten file against the output upstream records for it.
//
// This is the assertion the harness does not supply. ExpectFixedSource applies FIXES, and every
// repair this rule offers is a SUGGESTION, which the engine never applies unattended, so the applier
// is the one in this package that await-thenable already needed.
//
// It is where this rule can be wrong invisibly. A finding carrying a repair that overwrites the
// wrong span satisfies its message id perfectly, and every id fixture above stays green over it.
// Three of these outputs turn on exactly that: the annotation replacement has to reach back over the
// colon without eating a `?`, stop before a trailing comment, and leave a default expression alone.
func TestUseUnknownInCatchCallbackVariableSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantOutput string
	}{
		{"upstream invalid 0", "\nPromise.resolve().catch((err: Error) => {\n  throw err;\n});\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch((err: unknown) => {\n  throw err;\n});\n      "},
		{"upstream invalid 2", "\nPromise.resolve().catch((e, ...rest: []) => {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch((e: unknown, ...rest: []) => {\n  throw err;\n});\n      "},
		{"upstream invalid 3", "\nPromise.resolve().catch(\n  (err: string | ((number | unknown) & { b: () => void })) => {\n    throw err;\n  },\n);\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch(\n  (err: unknown) => {\n    throw err;\n  },\n);\n      "},
		{"upstream invalid 4", "\nPromise.resolve().catch(\n  (\n    ...err: [\n      unknown[],\n      string | ((number | unknown) & { b: () => void }),\n      string,\n    ]\n  ) => {\n    throw err;\n  },\n);\n      ", "wrongRestTypeAnnotationSuggestion", "\nPromise.resolve().catch(\n  (\n    ...err: [unknown]\n  ) => {\n    throw err;\n  },\n);\n      "},
		{"upstream invalid 5", "\nPromise.resolve().catch(function (err: string) {\n  throw err;\n});\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch(function (err: unknown) {\n  throw err;\n});\n      "},
		{"upstream invalid 6", "\nPromise.resolve().catch(function (err /* awkward spot for comment */) {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch(function (err: unknown /* awkward spot for comment */) {\n  throw err;\n});\n      "},
		{"upstream invalid 7", "\nPromise.resolve().catch(function namedCallback(err: string) {\n  throw err;\n});\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch(function namedCallback(err: unknown) {\n  throw err;\n});\n      "},
		{"upstream invalid 8", "\nPromise.resolve().catch(err => {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch((err: unknown) => {\n  throw err;\n});\n      "},
		{"upstream invalid 9", "\nPromise.resolve().then(\n  () => {},\n  error => {},\n);\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().then(\n  () => {},\n  (error: unknown) => {},\n);\n      "},
		{"upstream invalid 10", "\nPromise.resolve().catch((err?) => {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch((err?: unknown) => {\n  throw err;\n});\n      "},
		{"upstream invalid 11", "\nPromise.resolve().catch((err?: string) => {\n  throw err;\n});\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch((err?: unknown) => {\n  throw err;\n});\n      "},
		{"upstream invalid 12", "\nPromise.resolve().catch(err/* with comment */=> {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch((err: unknown)/* with comment */=> {\n  throw err;\n});\n      "},
		{"upstream invalid 13", "\nPromise.resolve().catch((err = 2) => {\n  throw err;\n});\n      ", "addUnknownTypeAnnotationSuggestion", "\nPromise.resolve().catch((err: unknown = 2) => {\n  throw err;\n});\n      "},
		{"upstream invalid 14", "\nPromise.resolve().catch((err: any /* comment 1 */ = /* comment 2 */ 2) => {\n  throw err;\n});\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve().catch((err: unknown /* comment 1 */ = /* comment 2 */ 2) => {\n  throw err;\n});\n      "},
		{"upstream invalid 15", "\nPromise.resolve().catch((...args) => {\n  throw args[0];\n});\n      ", "addUnknownRestTypeAnnotationSuggestion", "\nPromise.resolve().catch((...args: [unknown]) => {\n  throw args[0];\n});\n      "},
		{"upstream invalid 17", "\nPromise.resolve(' a string ').catch(\n  (a: any, b: () => any, c: (x: string & number) => void) => {},\n);\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve(' a string ').catch(\n  (a: unknown, b: () => any, c: (x: string & number) => void) => {},\n);\n      "},
		{"upstream invalid 20", "\nPromise.resolve()['catch']((x: any) => 'return');\n      ", "wrongTypeAnnotationSuggestion", "\nPromise.resolve()['catch']((x: unknown) => 'return');\n      "},
		{"upstream invalid 21", "\nPromise.reject().catch((...x: any) => {});\n      ", "wrongRestTypeAnnotationSuggestion", "\nPromise.reject().catch((...x: [unknown]) => {});\n      "},
		{"upstream invalid 22", "\nPromise.resolve().catch((...[args]: [string]) => {\n  console.log(args);\n});\n      ", "wrongRestTypeAnnotationSuggestion", "\nPromise.resolve().catch((...[args]: [unknown]) => {\n  console.log(args);\n});\n      "},
		{"upstream invalid 23", "\nPromise.resolve().catch((...{ find }: [string]) => {\n  console.log(find);\n});\n      ", "wrongRestTypeAnnotationSuggestion", "\nPromise.resolve().catch((...{ find }: [unknown]) => {\n  console.log(find);\n});\n      "},
		{"upstream invalid 24", "\ndeclare const condition: boolean;\nPromise.resolve('foo').then(() => {}, condition ? err => {} : err => {});\n      ", "addUnknownTypeAnnotationSuggestion", "\ndeclare const condition: boolean;\nPromise.resolve('foo').then(() => {}, condition ? (err: unknown) => {} : err => {});\n      "},
		{"upstream invalid 24", "\ndeclare const condition: boolean;\nPromise.resolve('foo').then(() => {}, condition ? err => {} : err => {});\n      ", "addUnknownTypeAnnotationSuggestion", "\ndeclare const condition: boolean;\nPromise.resolve('foo').then(() => {}, condition ? err => {} : (err: unknown) => {});\n      "},
		{"upstream invalid 25", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && (err => {})) || (err => {}),\n);\n      ", "addUnknownTypeAnnotationSuggestion", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? ((err: unknown) => {}))\n    : (condition && (err => {})) || (err => {}),\n);\n      "},
		{"upstream invalid 25", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && (err => {})) || (err => {}),\n);\n      ", "addUnknownTypeAnnotationSuggestion", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && ((err: unknown) => {})) || (err => {}),\n);\n      "},
		{"upstream invalid 25", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && (err => {})) || (err => {}),\n);\n      ", "addUnknownTypeAnnotationSuggestion", "\ndeclare const condition: boolean;\ndeclare const maybeNullishHandler: null | ((err: any) => void);\nPromise.resolve('foo').catch(\n  condition\n    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))\n    : (condition && (err => {})) || ((err: unknown) => {}),\n);\n      "}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatal("want at least one finding, got none")
			}
			// The harness trims the fixture before building the program, so the offsets a fix
			// carries are against the trimmed text rather than against the literal written above.
			// The expectation has to be transformed the same way the input was.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			want := strings.TrimSpace(testCase.wantOutput) + "\n"

			// Two cases report several findings that all carry the SAME suggestion id, one per
			// callback in the expression, and upstream records a separate output for each. So the
			// row is matched by what the repair PRODUCES rather than by its id alone: picking the
			// first id match would compare one callback's repair against another's expectation and
			// fail on a rule that is right. Both the id and the output still have to agree.
			matchedTheId := false
			for _, diagnostic := range result.Diagnostics {
				for _, candidate := range diagnostic.Suggestions {
					if candidate.Message.Id != testCase.wantId {
						continue
					}
					matchedTheId = true
					if applySuggestion(t, source, candidate) == want {
						return
					}
				}
			}
			if !matchedTheId {
				t.Fatalf("no suggestion with id %q among the findings", testCase.wantId)
			}
			t.Errorf("no suggestion with id %q produced\n%q", testCase.wantId, want)
		})
	}
}

// TestUseUnknownInCatchCallbackVariableComputedKeyDivergence records the one upstream invalid case
// this rule does not reproduce, and pins the spelling it does reproduce beside it.
//
// Upstream resolves a computed member key with ESLint's `getStaticValue`, a constant folder over the
// scope. This resolves it with the checker's `getAccessedPropertyName`, which reads the key's TYPE.
// The two agree wherever a literal type survives and part ways on `let`, because a `let` initialised
// with a string literal widens to `string` and leaves the checker nothing to read.
//
// So upstream's invalid case 1 is silent here. It is kept as a case rather than deleted, asserted as
// silent with the reason at the line, so that a checker or a folding pass that later closes the gap
// fails this test loudly instead of the divergence going unnoticed.
//
// The `const` row is the control, and it is the half that makes this a measurement rather than an
// excuse: it is the same shape with the same rule and it REPORTS, which is what shows the silence
// above comes from the widening rather than from the computed access being unhandled. It is also
// something a literal-only test would lose, so it pins the reach as well as the limit.
func TestUseUnknownInCatchCallbackVariableComputedKeyDivergence(t *testing.T) {
	t.Parallel()

	t.Run("a let-bound key is silent here and reports upstream", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, "\nlet method = 'catch';\nPromise.resolve()[method]((error: Error) => {});\n      ")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a const-bound key reports, which a string-literal test would miss", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile,
			"\nconst method = 'catch';\nPromise.resolve()[method]((error: Error) => {});\n      ")
		rule_testing.ExpectFindings(t, result, "useUnknown")
	})
}

// TestUseUnknownInCatchCallbackVariableThisParameter pins an upstream DEFECT that this port
// reproduces on purpose.
//
// A `this` parameter is not a parameter. The checker's signature omits it, while both ESTree's
// params list and our Parameters() include it. Upstream reads the checker to decide whether to
// report and the syntax to decide where, so on a callback with a `this` parameter it judges the
// first real parameter and then points at `this`, offering a repair that would write
// `function (this: unknown, err: Error)`. That is broken code.
//
// Measured against the installed rule at 8.67.0, with two controls that pin the MECHANISM rather
// than the symptom, since the finding alone is consistent with several wrong explanations:
//
//	p.catch(function (this: W, err: Error) {})     reports, span "this: W", suggests over ": W"
//	p.catch(function (this: W, err: unknown) {})   CLEAN, so the decision reads `err`, not `this`
//	p.catch(function (this: W) {})                 CLEAN, so `this` is no checker parameter at all
//
// Our AST agrees with ESTree that `this` is in the parameter list, probed directly: Parameters() has
// two entries where the checker signature has one. So reproducing this costs nothing and diverging
// would be a silent improvement of the kind the porting standard refuses. It is reproduced, and it
// is written down as a defect here so the next reader does not helpfully correct it back.
//
// The span assertion is the load-bearing half. Every message id below would be satisfied by a rule
// that correctly pointed at `err`, which is what makes this a real test of the divergence.
func TestUseUnknownInCatchCallbackVariableThisParameter(t *testing.T) {
	t.Parallel()

	const source = "interface W {\n  z: number;\n}\ndeclare const p: Promise<void>;\np.catch(function (this: W, err: Error) {});"

	result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, source)
	rule_testing.ExpectFindings(t, result, "useUnknown")

	text := strings.TrimSpace(source) + "\n"
	reported := text[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "this: W" {
		t.Errorf("the finding points at %q, and upstream points at the `this` parameter %q", reported, "this: W")
	}

	got := applySuggestion(t, text, result.Diagnostics[0].Suggestions[0])
	const want = "interface W {\n  z: number;\n}\ndeclare const p: Promise<void>;\np.catch(function (this: unknown, err: Error) {});\n"
	if got != want {
		t.Errorf("the repair produced\n%q\nwant\n%q", got, want)
	}

	t.Run("a this parameter alone is clean, so this is no checker parameter", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile,
			"interface W {\n  z: number;\n}\ndeclare const p: Promise<void>;\np.catch(function (this: W) {});"))
	})

	t.Run("an unknown err is clean, so the decision reads err rather than this", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile,
			"interface W {\n  z: number;\n}\ndeclare const p: Promise<void>;\np.catch(function (this: W, err: unknown) {});"))
	})
}

// TestUseUnknownInCatchCallbackVariableRequiresTheTypedHarness asserts the rule declares the checker
// and that the plain harness cannot prove it, so a later revert to rule_testing.Run fails loudly.
//
// This rule is the silent kind rather than the panicking kind, which is the more dangerous of the
// two. Under rule_testing.Run the checker is nil, Run returns nil listeners, and every StaysSilent case
// above would pass VACUOUSLY while every Fires case failed. The count below is what makes the
// difference visible: a rule that fires on nothing is not the same as a rule that discriminates.
func TestUseUnknownInCatchCallbackVariableRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !UseUnknownInCatchCallbackVariable.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}

	const source = "Promise.resolve().catch((err: Error) => {\n  throw err;\n});"

	typed := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, source)
	rule_testing.ExpectFindings(t, typed, "useUnknown")

	untyped := rule_testing.Run(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, source)
	rule_testing.ExpectClean(t, untyped)
}

// TestUseUnknownInCatchCallbackVariableMessages pins the rendered text of every message.
//
// Three of the seven are built by concatenation from the method name and an append slot, so there is
// a format to get wrong: the `then` arm has to render "a `then` rejection callback variable" and the
// `catch` arm "a `catch` callback variable", and no message-id assertion anywhere above can see the
// difference. The literals below are typed out rather than compared against the rule's own message
// constants, which would move with the rule under mutation and assert nothing.
func TestUseUnknownInCatchCallbackVariableMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		want       string
	}{
		{
			"the catch arm",
			"Promise.resolve().catch((err: Error) => {});",
			"useUnknown",
			"Prefer the safe `: unknown` for a `catch` callback variable.",
		},
		{
			"the then arm carries the rejection append",
			"Promise.resolve().then(() => {}, (err: Error) => {});",
			"useUnknown",
			"Prefer the safe `: unknown` for a `then` rejection callback variable.",
		},
		{
			"the array destructuring arm",
			"Promise.resolve().catch(([err]: [unknown]) => {});",
			"useUnknownArrayDestructuringPattern",
			"Prefer the safe `: unknown` for a `catch` callback variable. The thrown error may not be iterable.",
		},
		{
			"the object destructuring arm",
			"Promise.resolve().catch(({}) => {});",
			"useUnknownObjectDestructuringPattern",
			"Prefer the safe `: unknown` for a `catch` callback variable. The thrown error may be nullable, or may not have the expected shape.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Id; got != testCase.wantId {
				t.Fatalf("the message id is %q, want %q", got, testCase.wantId)
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Errorf("the message reads\n%q\nwant\n%q", got, testCase.want)
			}
		})
	}
}

// TestUseUnknownInCatchCallbackVariableParentheses pins paren transparency at each of the three
// places our AST has a node where ESTree has nothing.
//
// Upstream's walk never mentions parentheses, because ESTree does not represent them, so it gets
// transparency for free at every level. Ours does not, and the corpus is nearly silent on the
// subject: only the last invalid case writes parenthesized handlers, and it happens to cover just
// the recursion site. The other two sites were found by mutation and confirmed against the
// installed rule, which reports all three shapes.
//
// The callee row is the one that earns its place. Its mutant SURVIVED the whole imported corpus
// including that last case, and the first input tried for it was not distinguishing: `(p).catch(f)`
// parenthesizes the RECEIVER, leaving the callee itself unparenthesized, so it reports either way.
// It is kept below as the control that makes that visible. Measured with the skip removed:
//
//	(p.catch)(err => {})     pristine reports, mutant SILENT   distinguishes
//	((p.catch))(err => {})   pristine reports, mutant SILENT   distinguishes
//	(p).catch(err => {})     both report                       does NOT distinguish
func TestUseUnknownInCatchCallbackVariableParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a parenthesized argument", "Promise.resolve().catch((err => {\n  throw err;\n}));"},
		{"a doubly parenthesized argument", "Promise.resolve().catch(((err => {})));"},
		{"a parenthesized callee", "declare const p: Promise<void>;\n(p.catch)(err => {});"},
		{"a doubly parenthesized callee", "declare const p: Promise<void>;\n((p.catch))(err => {});"},
		{"a parenthesized receiver, which the callee skip does not decide", "declare const p: Promise<void>;\n(p).catch(err => {});"},
		{"a parenthesized branch of a conditional", "declare const c: boolean;\nPromise.resolve().catch(c ? (err => {}) : (err => {}));"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatal("want a finding: a parenthesis is invisible to upstream and must be here too")
			}
		})
	}
}

// TestUseUnknownInCatchCallbackVariableSurvivesMalformedCalls runs the rule over call shapes where a
// node it reaches for is absent or is not what the well-formed grammar produces.
//
// A panic here would be worse than a wrong finding by a wide margin. The walk recovers per FILE
// rather than per rule, so one nil dereference does not cost this rule one verdict, it takes the
// whole file away from every rule running on it, and the run still prints a green summary line.
//
// The sources are deliberately malformed as well as merely unusual, because error recovery
// synthesizes nodes that well-formed source never produces, and those are the ones a rule written
// against the grammar has not seen.
//
// What this test does NOT prove is worth stating, because the obvious reading of it is wrong.
// `ast.SkipParentheses` dereferences its argument and this rule routes every call through a helper
// that answers nil for nil, which reads as though these shapes exercise that guard. They do not.
// Removing the guard entirely leaves this test GREEN, measured rather than assumed, because our
// parser synthesizes a missing element-access key as an identifier instead of leaving the field
// absent: `p[]` arrives as `KindIdentifier`. Probed at both skip sites over five malformed shapes,
// the nil count was zero.
//
// So this is a guard against the rule's own arithmetic and indexing on shapes it was not written
// for, which is a real thing to guard, and it is NOT evidence that the nil check is load-bearing.
//
// This runs through the TYPED harness. Under rule_testing.Run the rule returns before installing a
// listener, so the untyped harness would pass without executing a line of it.
func TestUseUnknownInCatchCallbackVariableSurvivesMalformedCalls(t *testing.T) {
	t.Parallel()

	sources := []string{
		// An element access with an empty or missing argument expression.
		"declare const p: any;\np[](err => {});\n",
		"declare const p: any;\np[]();\n",
		// A call with no arguments at all where one is indexed.
		"declare const p: Promise<void>;\np.catch();\n",
		"declare const p: Promise<void>;\np.then();\n",
		"declare const p: Promise<void>;\np.then(() => {});\n",
		// Parentheses around nothing, and an empty call.
		"declare const p: Promise<void>;\n()(err => {});\n",
		"declare const p: Promise<void>;\n(p.catch)();\n",
		// A callback whose parameter list or name is absent.
		"declare const p: Promise<void>;\np.catch(function () {});\n",
		"declare const p: Promise<void>;\np.catch(function (,) {});\n",
		"declare const p: Promise<void>;\np.catch(() => {});\n",
		// A spread with nothing in it, and a trailing spread.
		"declare const p: Promise<void>;\np.catch(...);\n",
		"declare const p: Promise<void>;\np.then(() => {}, ...);\n",
		// Optional-call and nested optional forms of the same access.
		"declare const p: Promise<void> | undefined;\np?.catch(err => {});\n",
		"declare const p: any;\np?.['catch'](err => {});\n",
		// A computed key that is not a string at all.
		"declare const p: any;\np[0](err => {});\n",
		"declare const p: any;\np[Symbol.iterator](err => {});\n",
		// A parameter that is a destructure with holes, and a default with no value.
		"declare const p: Promise<void>;\np.catch(([, ]) => {});\n",
		"declare const p: Promise<void>;\np.catch((err = ) => {});\n",
		// A conditional and a sequence with a missing operand, which the walk recurses into.
		"declare const p: Promise<void>;\ndeclare const c: boolean;\np.catch(c ? : err => {});\n",
		"declare const p: Promise<void>;\np.catch((, err => {}));\n",
	}

	for index, source := range sources {
		t.Run(fmt.Sprintf("shape-%d", index), func(t *testing.T) {
			t.Parallel()
			// A panic fails the test. The assertion is only that this returns at all: what the rule
			// concludes about a malformed shape belongs in its own fixture, and asserting it here
			// would make this guard fail for reasons that are not crashes.
			rule_testing.RunTyped(t, UseUnknownInCatchCallbackVariable, useUnknownInCatchCallbackVariableFile, source)
		})
	}
}
