# Wave 20 checker audit

Audited area: `1f9e223d8ef6d0a914407d4b37d0e56568521f4d`. Cohere source: `7945d102a6c18dd36adf9114a758ce646e8b2359`. Old wave evidence: `e88bd6651c9c6d06e7720833663adebb838ad641` on `codex/typeaware-wave-20`.

Scope: all 15 rules in `stage1/cohere/typeaware/claims/wave-20.md` on the old wave branch. This is a source-inspection audit, not a runtime parity result or an exhaustive implementation plan. These are observed first blockers and additional directly inspected binding/callee requirements; no new ports, checker questions or shared helper copies were made.

## Missing calls and shared analyses

Symbols below use the aliases and receiver names at the cohere call sites. All sites are relative to `cohere/internal/lint/rules/`. A checker method listed here is missing the required answer/addressing capability, not necessarily every use of that method.

| Kind | Exact Go symbol at call site | Rules blocked | Cohere call sites | Required capability |
| --- | --- | --- | --- | --- |
| shared helper | `type_checking.IsPromiseLike` | @typescript-eslint/no-floating-promises | `typescript/no_floating_promises.go:372` | Promise/default-library predicate. Type graph facts alone do not provide the shared helper. |
| shared helper | `type_checking.IsBuiltinSymbolLike` | @typescript-eslint/no-implied-eval | `typescript/no_implied_eval.go:280` | Builtin FunctionConstructor/base-type predicate; also called with Function at line 236. |
| shared helper | `type_checking.IsThenableType` | @typescript-eslint/no-meaningless-void-operator | `typescript/no_meaningless_void_operator.go:191` | Thenable predicate, including the callable then parameter shape. |
| shared helper | `type_checking.IsSymbolFromDefaultLibrary` | nexus/correctness-no-uncleared-race-timeout | `nexus/correctness_no_uncleared_race_timeout.go:173` | Node symbol and default-library predicate; type-symbol origin is not the same question. |
| shared analysis | `control_flow_graph.Build` | nexus/correctness-no-process-exit-after-output; nexus/correctness-require-blocking-standard-streams | `nexus/correctness_no_process_exit_after_output.go:511; nexus/correctness_require_blocking_standard_streams.go:840` | Shared parser-node CFG, with the rule-specific Hooks events. Existing compiler IR flow is not this source-node analysis. |
| shared helper | `property.AccessedName` | prefer-promise-reject-errors | `core/prefer_promise_reject_errors.go:247` | Called with property.Static; evaluate computed/static member keys rather than returning the name node. |
| shared analysis | `reference.NewTracker` | prefer-regex-literals | `core/prefer_regex_literals.go:196` | Shared reference/alias tracker for global constructor references. |
| checker question | `ctx.TypeChecker.GetSymbolAtLocation` | prefer-rest-params; prefer-promise-reject-errors; react/jsx-fragments; react/jsx-no-undef; react/jsx-no-constructed-context-values | `core/prefer_rest_params.go:114; core/prefer_promise_reject_errors.go:278,307; react/jsx_fragments.go:290; react/jsx_no_undef.go:101; react/jsx_no_constructed_context_values_stability.go:125` | Need symbol presence, stable identity and declaration list (including zero declarations and declaration source metadata). symbol-origin returns the same empty answer for absent and implicit declaration-free symbols; declarations is restricted to class/interface nodes. |
| shared helper | `jsx.ElementParts` | react/jsx-fragments; react/jsx-no-undef | `react/jsx_fragments.go:207; react/jsx_no_undef.go:93` | Shared JSX tag/attribute accessor, alongside the binding question above. |
| checker question | `walk.ctx.TypeChecker.GetResolvedSignature` | react/jsx-no-constructed-context-values | `react/jsx_no_constructed_context_values_stability.go:488` | Need signature.Declaration(), its body, source file, spans and function flags. The area signature graph does not provide the resolved declaration/body. |
| checker addressing | `walk.ctx.TypeChecker.GetTypeAtLocation` | react/jsx-no-constructed-context-values | `react/jsx_no_constructed_context_values_stability.go:283` | GetTypeAtLocation already works for current-file nodes. The missing capability is addressing expressions in a foreign resolved callee body. Checker.ask binds node indices to its current file/parser; askFile addresses its current SourceFile, not an arbitrary foreign node. |
| shared analysis | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | react-hooks/set-state-in-effect | `react/set_state_in_effect.go:267` | Source-to-HIR lowering with SSA/capture/manual-memoization analysis on lint RuleContext. |
| shared analysis | `high_level_intermediate_representation.ForFunction` | react-hooks/set-state-in-render; react-hooks/static-components | `react/set_state_in_render.go:183; react/static_components.go:120` | Source-to-HIR lowering with SSA/capture analysis on lint RuleContext. |

## Reproducers by rule

These are minimal source shapes demonstrating the requested dependency, not newly executed controls. Ambient declarations/imports needed for a complete project should be supplied normally.

### nexus/correctness-no-process-exit-after-output

```tsx
console.log("x"); process.exit(0);
```

### nexus/correctness-no-uncleared-race-timeout

```tsx
Promise.race([work(), new Promise((_resolve, reject) => setTimeout(reject, 1))]);
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

### prefer-rest-params

```tsx
arguments; // unresolved, no finding
function f() { arguments; } // resolved implicit symbol, finding
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

### react/jsx-no-undef

```tsx
const x = <Missing />;
```

## Legacy bridge selector surface absent from area

The following 26 Inspect selectors are present in the old merged wave branch but absent from the audited area facts.go switch. This is the full surface difference, not a claim that every selector is required by these 15 rules or was authored exclusively by wave 20. Inspect selectors are a separate namespace from the Go symbols above. Preserve recorded fact grammars and program-scoped identities if the landing lane migrates any of them.

| Surface | Legacy selectors | Old bridge source |
| --- | --- | --- |
| Binding/symbol identity | `binding-origin`, `symbol-identities`, `binding-declarations`, `alias-declarations`, `resolved-name` | `bridge/tsgo/checker/binding_origin.go`, `symbol_identity.go`, `facts.go` |
| Declaration metadata | `declaration-chain`, `node-symbol-details`, `declaration-details`, `type-symbol-details`, `property-declarations` | `bridge/tsgo/checker/declaration_chain.go`, `declaration_facts.go` |
| Foreign source/callee metadata | `module-records`, `resolved-callee` | `bridge/tsgo/checker/module_records.go`, `resolved_callee.go` |
| Property/callback/literal facts | `accessed-property`, `callback-parameters`, `literal-value` | `bridge/tsgo/checker/accessed_property.go`, `callback_parameters.go`, `facts.go` |
| Type metadata/property/signature facts | `type-metadata`, `type-properties`, `function-signatures`, `property-exists`, `identical-types` | `bridge/tsgo/checker/facts.go` |
| Additional graph roots | `reference-shape`, `container-bases`, `contextual-argument`, `annotated-return-shape`, `symbol-shape`, `annotation-shape` | `bridge/tsgo/checker/facts.go` |

`binding-origin` preserves raw/read symbol identities and declaration metadata, including shorthand/export targets. `resolved-callee` supplies signature-declaration metadata; `module-records` supplies source text and resolved imports. These are useful migration candidates for the blocked binding/foreign-callee capabilities, but do not themselves port the shared analyses.

## Legacy TypeScript API incompatibilities in the landing lane

The old wave branch was merged with area and failed to compile. These are compatibility changes in its legacy extensions, distinct from the absent Inspect selectors:

| Old API/use | Current incompatibility | Observed wave sites |
| --- | --- | --- |
| `SourceFile.FileName()` supplied directly as a string | Returns `tspath.RootedFilePath`; calls to `out.text` or assignments to string need an explicit conversion | `binding_origin.go:39`, `declaration_chain.go:36`, `declaration_facts.go:17`, `facts.go:299`, `module_records.go:17,66`, `resolved_callee.go:35` |
| `SourceFile.Path()` | No longer exists; area uses `SourceFile.PathKey()` for `IsSourceFileDefaultLibrary` | `declaration_chain.go:38`, `declaration_facts.go:22` |
| `Program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)` | Parameter is now `*module.ResolvedModule`, not its rooted filename | `module_records.go:65` |

Reproducer on the merged old wave branch: `GOFLAGS=-buildvcs=false GOMAXPROCS=4 go vet ./stage1/cohere/lint/...`. The full lint package also failed to build before tests: package 0 pass/1 fail/0 skip, 144.698s. Original public diagnostics remain under `stage1/cohere/lint/rules/no-self-compare/blocked-typeaware/evidence/` on that branch. No compatibility edits were made here.

## Audit validation

Checked exact upstream call sites, current area Inspect cases, RuleContext/Checker addressing, and the old wave selector difference. Only this Markdown file is committed; no porting or runtime tests in this documentation-only turn.
