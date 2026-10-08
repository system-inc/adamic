# Wave 01 checker audit

Audit branch base: `1f9e223d8ef6d0a914407d4b37d0e56568521f4d` (origin/area/stage1-lint).
Blocker observations were made at area `c4bdc23fa86d55cf7e579989201c11258f4d3a62`; this documentation-only consolidation does not rerun ports or certify newer questions.
Cohere paths below are relative to `cohere/internal/lint/rules/`. These are the exact call-site spellings, including local receiver names.
No rule code, shared helper, checker question or descriptor is added. All 12 rules have zero newly certified upstream cases and zero new mutant executions on the unified harness.

## Missing questions and shared analysis

| Go call or field | Rules blocked | Cohere call sites |
| --- | --- | --- |
| `checker.Checker_functionHasImplicitReturn` | nexus/correctness-no-implicit-return | `nexus/correctness_no_implicit_return.go:96` |
| `ast.GetJSDocDeprecatedTag` | @typescript-eslint/no-deprecated | `typescript/no_deprecated.go:192` |
| `ctx.TypeChecker.GetSymbolsInScope` | no-else-return | `core/no_else_return.go:372` |
| `ctx.TypeChecker.GetSymbolAtLocation` | nexus/correctness-require-child-process-error-listener; nexus/performance-no-independent-await-in-loop; symbol-description; react-hooks/globals | `nexus/correctness_require_child_process_error_listener.go:135; nexus/performance_no_independent_await_in_loop.go:711; core/no_new_native_nonconstructor.go:104; react/globals.go:617` |
| `ctx.TypeChecker.GetShorthandAssignmentValueSymbol` | nexus/correctness-require-response-status-check; nexus/performance-no-independent-await-in-loop | `nexus/correctness_require_response_status_check.go:486; nexus/performance_no_independent_await_in_loop.go:709` |
| `checker.SkipAlias` | nexus/performance-no-independent-await-in-loop | `nexus/performance_no_independent_await_in_loop.go:479` |
| `symbol.Declarations` (field, not call) | symbol-description | `core/no_new_native_nonconstructor.go:105` |
| `checker.Checker_getResolvedSignature; resolved.Target()` | require-await | `core/require_await.go:988,992` |
| `control_flow_graph.Build` | require-atomic-updates; nexus/correctness-require-response-status-check | `core/require_atomic_updates.go:402; nexus/correctness_require_response_status_check.go:521` |
| `high_level_intermediate_representation.MayHoldComponentOrHook` | react-hooks/immutability | `react/immutability.go:166` |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | react-hooks/no-deriving-state-in-effects | `react/no_deriving_state_in_effects.go:142` |

## Per-rule reason and reproducer

### nexus/correctness-no-implicit-return

No function end-flow/implicit-return answer is exposed.

```typescript
function f(x:boolean){if(x)return 1;}
```

### nexus/correctness-require-child-process-error-listener

No stable resolved binding-symbol identity/declaration answer is exposed, nor shorthand/export symbol identities.

```typescript
import {spawn} from 'node:child_process'; const child=spawn('x');
```

### nexus/correctness-require-response-status-check

No shorthand binding-symbol identity answer or shared control_flow_graph.Build implementation is exposed.

```typescript
async function f(){const response=await fetch('/');return {response};}
```

### nexus/performance-no-independent-await-in-loop

No stable resolved/alias-skipped symbol identities are exposed.

```typescript
async function f(xs:number[]){for(const x of xs){await Promise.resolve(1);}}
```

### no-else-return

scope-locals exposes binder tables, not GetSymbolsInScope at an arbitrary if with declarations and lexical scope ownership.

```typescript
function f(x:boolean){let a; if(x)return 1;else {let a=2;}}
```

### react-hooks/globals

No resolved identifier declaration/lexical ownership answer is exposed.

```typescript
let g=0;function Component(){g=1;return null;}
```

### react-hooks/immutability

Shared React HIR/SSA/capture analysis remains missing.

```typescript
function Component({value}){value.x=1;return null;}
```

### react-hooks/no-deriving-state-in-effects

Shared memo-erased React HIR/SSA/capture analysis remains missing.

```typescript
function Component(){const [x,setX]=useState(0);useEffect(()=>{setX(1)},[]);return null;}
```

### require-atomic-updates

Shared native control-flow graph builder with read/write/suspend hooks is missing; no private copy was made.

```typescript
let x=0;async function f(){x+=await Promise.resolve(1);}
```

### require-await

signature-shape does not expose the declared generic target and its type parameters/substitutions needed to avoid inferred self-demand.

```typescript
[1,2].map(async n=>n+1);
```

### symbol-description

declarations is restricted to class/interface nodes. symbol-origin does not expose declaration presence or IsDeclarationFile; see shared resolvesToAGlobal at lines 104-109.

```typescript
Symbol();
const local=()=>{const Symbol=()=>1;Symbol();};
```

### @typescript-eslint/no-deprecated

No resolved-symbol declaration JSDoc/deprecation answer is exposed; symbol-origin gives only a filename.

```typescript
/** @deprecated */ function old(){}; old();
```

## Legacy bridge APIs absent from area

These are legacy `Rules.ask` question names used in the old branch, not a request to recreate a private checker. Absence was checked against the current area Go dispatcher and its `additionalQuestions` registrations. The directly identified call-site list follows so integration can migrate the consumers as well as the implementations.

| Legacy question | Old branch consumers |
| --- | --- |
| `annotated-return-shape` | `stage1/cohere/typeaware/flow.ts:301` |
| `annotation-shape` | `stage1/cohere/typeaware/casts.ts:85`; `stage1/cohere/typeaware/flow.ts:316`; `stage1/cohere/typeaware/flow.ts:355`; `stage1/cohere/typeaware/property_alias.ts:174` |
| `binding-declarations` | `stage1/cohere/typeaware/property_alias.ts:120`; `stage1/cohere/typeaware/reassign.ts:87` |
| `call-symbol-shape` | `stage1/cohere/typeaware/wave_01_fourth/symbol_description/call_symbol_shape.a:10`; `stage1/cohere/typeaware/wave_01_fourth/symbol_description/call_symbol_shape.a:11` |
| `contextual-argument` | `stage1/cohere/typeaware/flow.ts:330` |
| `declaration-context` | `stage1/cohere/typeaware/wave_01_next/declaration_context.a:10` |
| `declaration-details` | `stage1/cohere/typeaware/caller.ts:78` |
| `function-return-flow` | `stage1/cohere/typeaware/function_return_flow.a:20` |
| `function-signatures` | `stage1/cohere/typeaware/flow.ts:79` |
| `literal-value` | `stage1/cohere/typeaware/coercion.ts:55`; `stage1/cohere/typeaware/no_deprecated.a:267`; `stage1/cohere/typeaware/wave_01_next/child_process_error_listener.a:73` |
| `loop-header` | `stage1/cohere/typeaware/wave_01_next/loop_header.a:4` |
| `node-symbol-details` | `stage1/cohere/typeaware/caller.ts:70` |
| `property-declarations` | `stage1/cohere/typeaware/caller.ts:191` |
| `property-exists` | `stage1/cohere/typeaware/flow.ts:92` |
| `reference-access` | `stage1/cohere/typeaware/wave_01_next/reference_access.a:4` |
| `reference-shape` | `stage1/cohere/typeaware/flow.ts:97` |
| `resolved-name` | `stage1/cohere/typeaware/coercion.ts:143` |
| `scope-symbol-declarations` | `stage1/cohere/typeaware/scope_symbol_declarations.a:5` |
| `symbol-documentation` | `stage1/cohere/typeaware/symbol_documentation.a:12`; `stage1/cohere/typeaware/symbol_documentation.a:47` |
| `symbol-identities` | `stage1/cohere/typeaware/unused.ts:40` |
| `symbol-shape` | `stage1/cohere/typeaware/caller.ts:160`; `stage1/cohere/typeaware/property_alias.ts:132`; `stage1/cohere/typeaware/property_alias.ts:170` |
| `syntax-flow-graph` | `stage1/cohere/typeaware/wave_01_next/syntax_flow_graph.a:6` |
| `type-metadata` | `stage1/cohere/typeaware/checker_facts.ts:45` |
| `type-properties` | `stage1/cohere/typeaware/checker_facts.ts:52` |
| `type-symbol-details` | `stage1/cohere/typeaware/caller.ts:74` |

The wave-specific implementations are in old branch `bridge/tsgo/checker/{function_return_flow,scope_symbol_declarations,symbol_documentation,declaration_context,loop_header,reference_access,syntax_flow_graph,call_symbol_shape}.go`; the additional dispatcher is `wave_01_questions.go`. Symbol/declaration fact extensions are in `{symbol_identity,declaration_facts,facts}.go`. `syntax-flow-graph` depended on `wave_01_next_flow/`, a Go CFG implementation, not the required shared native analysis. These implementations were not copied into lint.

## Existing evidence scope

Prior audit and all-input lint evidence: branch `codex/typeaware-wave-01-checker`, commit `2ed30ea6e1c9ed1b9fa1854e4b94944c1e35cb9f`, under `stage1/cohere/lint/rules/no-ex-assign/parked-claims/`. That run passed 34 top-level tests, failed 0, skipped only `TestCheckerBridgeRefusalPending`, and caught all 75 inherited registered mutants. It does not certify a port of these blocked rules. This turn is an audit consolidation only.
