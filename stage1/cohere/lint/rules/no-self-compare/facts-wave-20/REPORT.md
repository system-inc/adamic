# Wave 20 facts assessment

Base: `d845dccde413c89643293e808626344d12e3f023`, branch `lint-rules/facts-wave-20`. Cohere pin `7945d102a6c18dd36adf9114a758ce646e8b2359`. This branch starts from the supplied facts commit; no old bridge extension or private analysis was merged, copied or reused.

Read CHECKER_FACTS.md whole and the complete no-class-assign descriptor, listener, oracle, mutant and three witnesses as the declaration model. The shared checker supplies node-symbol-details, binding-declarations and provenance; those dependencies are no longer reported missing.

## All claimed rules

Three claims are already completed by wave 1 according to the user and were not ported again: `prefer-rest-params`, `react/jsx-no-undef`, `nexus/correctness-no-uncleared-race-timeout`. Their completion SHAs/case counts are not supplied here and are not invented.

All twelve remaining rules stop on a missing shared helper or question. This is a source inspection result, not a live unsupported-question probe or a runtime parity claim. A shared analysis request is a request for native Adamic analysis, not for a Go lint verdict.

| Rule | Exact blocked Go call | Cohere file:line | Missing helper/question | Exact upstream ProgramReads |
| --- | --- | --- | --- | --- |
| @typescript-eslint/no-floating-promises | `type_checking.IsPromiseLike` | `cohere/internal/lint/rules/typescript/no_floating_promises.go:372` | Shared native promise/default-library predicate; complete symbol declarations do not implement it. | `ReadsCompilerOptions | ReadsDefaultLibrary` |
| @typescript-eslint/no-implied-eval | `type_checking.IsBuiltinSymbolLike` | `cohere/internal/lint/rules/typescript/no_implied_eval.go:280` | Shared native builtin/base-type predicate for FunctionConstructor (and Function at line 236). | `ReadsCompilerOptions | ReadsDefaultLibrary` |
| @typescript-eslint/no-meaningless-void-operator | `type_checking.IsThenableType` | `cohere/internal/lint/rules/typescript/no_meaningless_void_operator.go:191` | Shared native thenable predicate; node-symbol-details is not callback/rest-parameter type projection. | `none declared` |
| nexus/correctness-no-process-exit-after-output | `control_flow_graph.Build` | `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:511` | Shared native source-node CFG and event hooks; alias targets and resolved-callee declarations are also still missing. | `none declared` |
| nexus/correctness-require-blocking-standard-streams | `control_flow_graph.Build` | `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:840` | Shared native source-node CFG; runtime module graph/import resolutions remain separate missing questions. | `ReadsCompilerOptions | ReadsModuleResolution | ReadsOtherFiles` |
| prefer-promise-reject-errors | `property.AccessedName` | `cohere/internal/lint/rules/core/prefer_promise_reject_errors.go:247` | Shared native property.Static evaluation of computed keys; new symbol presence/identity closes the old binding blocker, not static evaluation. | `none declared` |
| prefer-regex-literals | `reference.NewTracker` | `cohere/internal/lint/rules/core/prefer_regex_literals.go:196` | Shared native global-reference tracker, followed by constant-string/regex analysis; WritesToBinding is only write analysis. | `none declared` |
| react/jsx-fragments | `jsx.ElementParts` | `cohere/internal/lint/rules/react/jsx_fragments.go:207` | Shared JSX tag/attribute accessor. Bare-fragment binding also requires initializer/import syntax, which node-symbol-details does not serialize. | `none declared` |
| react/jsx-no-constructed-context-values | `walk.ctx.TypeChecker.GetResolvedSignature` | `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:488` | Raw resolved signature declaration/body metadata (path, spans, flags) and guarded foreign-node type queries. The ordinary signature type graph does not contain the declaration/body. Shared memo/escape analyses remain missing. | `none declared` |
| react-hooks/set-state-in-effect | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | `cohere/internal/lint/rules/react/set_state_in_effect.go:267` | Shared native lint source-to-HIR, SSA, captures and manual-memoization analysis. | `none declared` |
| react-hooks/set-state-in-render | `high_level_intermediate_representation.ForFunction` | `cohere/internal/lint/rules/react/set_state_in_render.go:183` | Shared native lint source-to-HIR, SSA and capture analysis. | `none declared` |
| react-hooks/static-components | `high_level_intermediate_representation.ForFunction` | `cohere/internal/lint/rules/react/static_components.go:120` | Shared native lint source-to-HIR, SSA and capture analysis. | `none declared` |

## Program-read contract

No descriptor was added or changed, and no program read was over-declared. For prefer-promise-reject-errors, react/jsx-fragments and react/jsx-no-constructed-context-values, cohere declares no ProgramReads. The model helper declarationAnswer in checker_declarations.a always calls askFile with ReadsOtherFiles, and Checker.askFile rejects a missing declaration. Copying that path into these rules would violate their upstream read contract. Current-file ask(index, node-symbol-details) is a distinct existing route and is not reported missing; it does not supply foreign initializer/body syntax or the absent shared helpers above. Any future port must use a read-compatible route or wait for a shared helper change rather than add ReadsOtherFiles. In particular, the constructed-context-values foreign callee-body reproducer needs ReadsOtherFiles for an askFile foreign-node/body query, because the body is in helper.ts; cohere declares no such read. That proposed route is a separate read-contract blocker, even after resolved-callee metadata becomes available.

The two TypeScript rules explicitly declare ReadsCompilerOptions and ReadsDefaultLibrary. The blocking-standard-streams rule declares ReadsCompilerOptions, ReadsModuleResolution and ReadsOtherFiles. All other remaining rule declarations above have no ProgramReads field.

## Reproducer shapes

### nexus/correctness-no-process-exit-after-output

```tsx
console.log("x"); process.exit(0);
```

### nexus/correctness-require-blocking-standard-streams

```tsx
#!/usr/bin/env node
process.exit(0);
```

### @typescript-eslint/no-floating-promises

```tsx
declare function f(): Promise<void>; f();
```

### @typescript-eslint/no-implied-eval

```tsx
new Function("return 1");
```

### @typescript-eslint/no-meaningless-void-operator

```tsx
declare const p: Promise<void>; void p;
```

### prefer-promise-reject-errors

```tsx
Promise[`re${"ject"}`]("bad");
```

### prefer-regex-literals

```tsx
const R = RegExp; new R("x");
```

### react-hooks/set-state-in-effect

```tsx
function App() { const [x, setX] = useState(0); useEffect(() => { setX(1); }, []); return <div>{x}</div>; }
```

### react-hooks/set-state-in-render

```tsx
function App() { const [x, setX] = useState(0); setX(1); return <div>{x}</div>; }
```

### react-hooks/static-components

```tsx
function App() { function Child() { return <div />; } return <Child />; }
```

### react/jsx-fragments

```tsx
import React from "react"; const x = <React.Fragment />;
```

### react/jsx-no-constructed-context-values

```tsx
// helper.ts
export function make() { return {}; }
// main.tsx
import { make } from "./helper";
function App() { const value = React.useMemo(() => make(), []); return <Ctx.Provider value={value} />; }
```

## Requested questions and native helpers

For IsPromiseLike and IsBuiltinSymbolLike, migrate the native checking helper and any exact type-symbol declaration/default-library provenance it needs (a type-identity-based type-symbol-details contract), with ReadsDefaultLibrary rather than adding ReadsOtherFiles. For IsThenableType, expose exact first callback parameter projection, including rest-number-index types, then supply the shared native checking helper. For property.AccessedName, reference.NewTracker, jsx.ElementParts, control_flow_graph.Build and the HIR calls, the missing contract is the native shared helper itself; a Go lint decision is not an acceptable question. For constructed context values, the missing raw question is resolved signature declaration/body metadata and a guarded same-program foreign-node selector; node-symbol-details does not include declaration initializer/body syntax.

## Existing legacy thenable implementation

`stage1/cohere/typeaware/returns.ts:36` contains `Returns.thenable`, but it is bound to the legacy `Rules` driver whose ask directly invokes tsgoInspect (`rules.ts:104`); it is not a shared lint RuleContext helper. No Rules adapter, direct bridge route or copied method was introduced. Moreover, cohere IsThenableType calls `IsCallback` at `cohere/internal/lint/checking/types.go:209`; IsCallback indexes a rest parameter by the number type at `types.go:175`. The facts slice call-parameters question returns the apparent first-parameter type without this rest-array indexing (`bridge/tsgo/checker/facts.go:389`). An exact callback-parameter/rest-index projection question plus a shared native thenable helper is still needed.

## Validation scope

Newly matched upstream cases: zero for each of the twelve blocked rules. New semantic mutants compiled/run/caught: none. No partial descriptor or listener is registered. Full package certification runs once with every input set and -timeout 3h; its result will be attached here. Shared tests/checker/helper files remain untouched.

## Toolchain

Setup PASS: Go ready 0.218s, Node ready 0.207s, submodules ready 0.438s, markdown dependencies ready 0.660s, clang ready 1.298s, Go build ready 288.015s, build cache warm 290.101s, done 291.224s. nproc 5; CPU quota four cores (400000/100000). Registry generation PASS (80 descriptors); go vet ./stage1/cohere/lint/... PASS (empty output). Existing owned Go adapter gofmt validation is recorded; no Go source was edited.

Full gate command: `GOFLAGS=-buildvcs=false GOMAXPROCS=4 ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=/workspace/wave20-facts-profiles ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/wave20-facts-profiles go test -json -count=1 -timeout 3h ./stage1/cohere/lint`. The source input is the existing clean TypeScript v6.0.3 checkout, pin 050880ce59e30b356b686bd3144efe24f875ebc8. No skip was introduced, relaxed or suppressed.

## Observed gate result

The single full package invocation ended without a final package event or exit status. Its process is no longer running and the runner did not write its completion summary; the cause is not established. Certification is **incomplete**, not green. Last output: 2026-10-08T17:52:21.914373312Z. Top-level observed results: {'pass': 14, 'skip': 1}. Including subtests: {'pass': 14, 'skip': 1}. Started but unfinished: TestOptionAndComparatorGaps, TestRulesAgree. Tests not reached are not counted as passes or skips. Observed wall time through the last load sample: 3422.561 seconds; nproc 5; initial load [12.7041015625, 6.6484375, 2.67529296875], last load [2.20751953125, 2.05419921875, 2.15185546875]. See evidence/summary.json and evidence/lint.jsonl. No second package invocation was made.

The one recorded skip is TestCheckerBridgeRefusalPending, checker_pending_test.go:49: awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer. No input-related skip was recorded before interruption. Shared code was not edited to remove or bypass this pending test.

Baseline source-byte checker mutants were caught in native, Node and emitted JavaScript (evidence/baseline-mutants.txt). These are shared harness controls, not newly ported rule mutants. TestOwnedWitnesses and TestMutants have no completion events in this run. TestRulesAgree began but did not finish. Partial typed-runtime findings comparisons and timings are preserved in evidence/lint.jsonl and do not certify any blocked rule.
