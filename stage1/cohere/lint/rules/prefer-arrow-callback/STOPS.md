# Wave 15 stops on the facts pin

Base: d845dccde413c89643293e808626344d12e3f023. Cohere: 7945d102a6c18dd36adf9114a758ce646e8b2359.

Both candidate implementations are blocked in shared infrastructure: prefer-arrow-callback on parse-refusal behavior in the converging fixer, and no-throw-literal on typed recovery metadata. No new complete port is claimed. Both use the run's checker, legacy same-node node-symbol-details grammar and the shared SymbolDetails reader. Their upstream declarations omit ProgramReads, so their descriptors declare an empty list. No foreign source is opened, no file-wide question is asked, and no compiler options or default-library query is issued by either rule. Declaration-file flags are GetSymbolAtLocation result metadata, as in cohere's shadow predicate, rather than a separate program read. Messages are verbatim.

prefer-find is already certified on lint-rules/prefer-find, 45f4f36c8fdae72f4590ab681f298fe0beb09cad; it is not transplanted here. react/jsx-no-undef is already handled by wave 1, per the task instruction, and is not duplicated. No new claims are taken.

Fourteen other rules need facts or native analyses absent at this pin. Complete symbol identities, declarations and provenance remove earlier stops, but do not supply the contracts below. Paths in this table are relative to cohere/internal/lint/rules/. These are source-audited stopping calls; none is described as a passing port.

| Exact Go call or helper | Rules stopped | Call site | Missing contract |
| --- | --- | --- | --- |
| `checker.Signature_parameters; checker.Signature_declaration; checker.Checker_getTypeOfSymbol; ctx.TypeChecker.GetSignatureFromDeclaration` | @typescript-eslint/no-useless-default-assignment | `typescript/no_useless_default_assignment.go:370, :380, :396, :445` | All contextual/declared signature parameters, optional/rest declaration facts, type identities and signature declaration identity. contextual-shape and call-parameters expose types but not this contract. |
| `type_checking.GetTypeName` | @typescript-eslint/require-array-sort-compare | `typescript/require_array_sort_compare.go:169` | Shared type-name normalization, including type-parameter annotation/declaration constraints; name returns TypeToString instead. |
| `ctx.TypeChecker.GetNonNullableType` | nexus/correctness-no-global-listener-target-assertion | `nexus/correctness_no_global_listener_target_assertion.go:257` | Nonnullable type identities usable as both operands of IsTypeAssignableTo, rather than guessing from displayed names. |
| `part.AsLiteralType().Value()` | nexus/correctness-no-leaked-number-render | `nexus/correctness_no_leaked_number_render.go:220` | Actual held literal values, including enum member zero, plus held-value status; TypeToString may return an enum member name. |
| `ctx.TypeChecker.GetResolvedSignature; signature.Declaration()` | nexus/correctness-no-mock-on-module-namespace | `nexus/correctness_no_mock_on_module_namespace.go:194, :198` | Resolved signature declaration and its complete parent/module chain. signature-shape already exposes signature types, not declaration ownership. |
| `ctx.Program.Options` | nexus/correctness-no-mock-on-module-namespace | `nexus/correctness_no_mock_on_module_namespace.go:110` | Compiler Module option needed by correctnessNoMockOnModuleNamespaceRunsAsModule; options currently exports strictness only. |
| `ctx.TypeChecker.GetSymbolsInScope` | no-label-var | `core/no_label_var.go:77` | Complete value-symbol scope chain at an arbitrary node. scope-locals exports binder locals and is a different operation. |
| `reference.NewTracker` | no-misleading-character-class; no-useless-backreference | `core/no_misleading_character_class.go:187; core/no_useless_backreference.go:100` | Shared native global-reference tracker with binding/alias/shadow semantics. |
| `reference.ConstantStringIn` | no-misleading-character-class; no-useless-backreference | `core/no_misleading_character_class.go:249, :276; core/no_useless_backreference.go:118, :124` | Shared native constant evaluator for constructor pattern and flags. |
| `reference.IsConstantRegExpIn` | no-misleading-character-class | `core/no_misleading_character_class.go:273` | Shared native constant RegExp evaluation, beyond an AST spelling check. |
| `regexsyntax.ParseRegexFlags; regexsyntax.SkipPatternEscape` | no-misleading-character-class; no-useless-backreference | `core/no_misleading_character_class.go:440, :608; core/no_useless_backreference.go:183, :307, :707` | Shared ECMAScript flags and escape analysis. |
| `regexsyntax.ParseRegexCharacterClassWithEnd` | no-misleading-character-class | `core/no_misleading_character_class.go:464` | Shared ECMAScript character-class grammar analysis. |
| `esregexp.Compile` | no-invalid-regexp | `core/no_invalid_regexp.go:263` | Shared ECMAScript validation and exact Go diagnostic descriptions. Dynamic new RegExp is also refused at internal/lower/regexp.go:39; ordinary JS syntax-error text is not a substitute. |
| `high_level_intermediate_representation.ForFunction` | react-hooks/set-state-in-render; react-hooks/static-components | `react/set_state_in_render.go:183; react/static_components.go:120` | Shared native HIR lowering, SSA/phi propagation and captures, not a Go rule verdict query. |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | react-hooks/set-state-in-effect | `react/set_state_in_effect.go:267` | The shared HIR path including memo erasure, SSA and capture analyses. |
| `high_level_intermediate_representation.MayHoldComponentOrHook` | react-hooks/set-state-in-effect; react-hooks/set-state-in-render; react-hooks/static-components | `react/set_state_in_effect.go:259; react/set_state_in_render.go:176; react/static_components.go:113` | Shared native source eligibility analysis used before requesting HIR. |
| `jsxNoConstructedContextValuesCheckMemo; (*jsxNoConstructedContextValuesStabilityWalk).functionEvaluation; (*jsxNoConstructedContextValuesStabilityWalk).anyEscapes; jsxNoConstructedContextValuesStaysHome` | react/jsx-no-constructed-context-values | `react/jsx_no_constructed_context_values.go:236; react/jsx_no_constructed_context_values_stability.go:786, :532, :603, :649` | Native memo/callback return and holder escape analysis, in addition to the missing identifier declaration/initializer facts above. |
| `declaration.AsVariableDeclaration().Initializer; variable.AsVariableDeclaration().Initializer; declaration.AsImportDeclaration().ModuleSpecifier` | react/jsx-fragments | `react/jsx_fragments.go:337, :333, :406` | Complete declaration initializer/import-container syntax. node-symbol-details includes spans and immediate parent metadata, not initializer or ancestor import module source. Same-file AST matching cannot answer a declaration belonging to another file. Cohere declares no program reads, so adding ReadsOtherFiles to fetch foreign syntax would over-declare the rule. |

## Reproducers

The following minimal inputs pin each recorded stop. React/module examples require the same ordinary declarations/import fixtures as their upstream cases; they do not provide a private checker or a private analysis implementation.

### nexus/correctness-no-global-listener-target-assertion

Exact upstream call: `ctx.TypeChecker.GetNonNullableType(known)` at `cohere/internal/lint/rules/nexus/correctness_no_global_listener_target_assertion.go:257`.

```typescript
addEventListener('click', e => (e.target as HTMLElement).click());
```

### nexus/correctness-no-leaked-number-render

Exact upstream call: `part.AsLiteralType().Value().(fmt.Stringer)` at `cohere/internal/lint/rules/nexus/correctness_no_leaked_number_render.go:220`.

```typescript
enum E { Zero = 0 } const n: E = E.Zero; const v = <div>{n && <span/>}</div>;
```

### nexus/correctness-no-mock-on-module-namespace

Exact upstream call: `ctx.TypeChecker.GetResolvedSignature(call)` at `cohere/internal/lint/rules/nexus/correctness_no_mock_on_module_namespace.go:194`.

```typescript
import * as mod from './m'; import { mock } from 'node:test'; mock.method(mod, 'run');
```

### no-invalid-regexp

Exact upstream call: `esregexp.Compile(pattern, compileFlags)` at `cohere/internal/lint/rules/core/no_invalid_regexp.go:263`.

```typescript
new RegExp('[');
```

### no-label-var

Exact upstream call: `ctx.TypeChecker.GetSymbolsInScope(node, ast.SymbolFlagsValue)` at `cohere/internal/lint/rules/core/no_label_var.go:77`.

```typescript
let label = 1; label: while (true) { break label; }
```

### no-misleading-character-class

Exact upstream call: `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` at `cohere/internal/lint/rules/core/no_misleading_character_class.go:187`.

```typescript
/[👍🏽]/u;
```

### no-useless-backreference

Exact upstream call: `reference.ConstantStringIn(ctx, arguments[0])` at `cohere/internal/lint/rules/core/no_useless_backreference.go:118`.

```typescript
const pattern = "(a\\1)"; new RegExp(pattern);
```

### @typescript-eslint/no-useless-default-assignment

Exact upstream call: `checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameterSymbol)` at `cohere/internal/lint/rules/typescript/no_useless_default_assignment.go:396`.

```typescript
const f: (a: number, b: number) => void = (a, b = 1) => {};
```

### react-hooks/set-state-in-effect

Exact upstream call: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_effect.go:267`.

```typescript
function C() { const [s, setS] = useState(0); useEffect(() => { setS(1); }, []); return s; }
```

### react-hooks/set-state-in-render

Exact upstream call: `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_render.go:183`.

```typescript
function C() { const [s, setS] = useState(0); setS(1); return s; }
```

### react-hooks/static-components

Exact upstream call: `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/static_components.go:120`.

```typescript
function C() { const Child = () => null; return <Child/>; }
```

### react/jsx-no-constructed-context-values

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:320`.

```typescript
function C() { const dep = {}; const value = useMemo(() => ({}), [dep]); return <Ctx.Provider value={value}/>; }
```

### @typescript-eslint/require-array-sort-compare

Exact upstream call: `type_checking.GetTypeName(ctx.TypeChecker, argument)` at `cohere/internal/lint/rules/typescript/require_array_sort_compare.go:169`.

```typescript
function f<T extends string>(a: T[]) { a.sort(); }
```

### react/jsx-fragments

Two script roots in one ordinary program demonstrate why matching only this file's node spans cannot provide the answer:

- globals.ts: `declare const React: any; const F = React.Fragment;`
- use.tsx: `const view = <F />;`

Go resolves F to globals.ts's VariableDeclaration, then reads its initializer at cohere/internal/lint/rules/react/jsx_fragments.go:337. The facts report that declaration's path/span but not its initializer. The port needs declaration initializer syntax, or a shared native declaration-syntax helper backed by an approved question. Adding a foreign-file read conflicts with the rule's empty upstream ProgramReads. No parser, helper or checker files are edited.

Constructed context values also retain declaration-initializer dependencies plus the memo/escape analysis in the table; symbol facts alone do not complete them. Regex rules retain their native ECMAScript/reference-analysis dependencies; no hand-written replacement regex engine is introduced.

## no-throw-literal: shared typed recovery metadata

The exact upstream case is rule_testing.RunTyped at cohere/internal/lint/rules/core/no_throw_literal_test.go:162, inside TestNoThrowLiteralHandlesShapesTheCorpusOmits: `function f() { throw; }`. Cohere expects no finding on this recovered syntax. The shared capture retains the source but does not attach its recovery marker; TestRulesAgree's typed branch passes it directly to the Go oracle at stage1/cohere/lint/lint_test.go:382. The oracle's collect calls SourceFile.Diagnostics() at stage1/cohere/lint/testdata/oracle.go:185 and panics at :186 on `Expression expected.` before it can compare findings. This is a harness input-metadata gap, not an absent symbol question or a tolerated parser mismatch. No guard is relaxed and no upstream case is excluded.

The exact candidate source/messages/adapter and witnesses are saved as proof/no-throw-literal-candidate.patch; the 48 captures are proof/no-throw-literal-upstream-cases.json. The working candidate is /workspace/scratch/wave15-no-throw-literal-parked. Applying the patch requires integration's recovery metadata fix. No blocked rule module or descriptor is in the landing registry. Five upstream cases preceding the recovery case and its own witness matched Go, source Node, emitted JavaScript and sanitized native. The shadow mutant compiled and ran and was caught on native, Node and emitted JavaScript. Complete 48-case certification is explicitly not claimed.

Failed focused evidence: proof/focused-recovery-failure.jsonl.gz. The initial full attempt was interrupted after the compiler exposed the owned arrow port's unsupported optional-chain spelling; that spelling was replaced by an explicit guard. Its log is proof/initial-optional-chain-failure.jsonl.gz. The complete all-input package invocation retains the registered arrow candidate, reports its TestRulesAgree failure, and keeps the pending base bridge-refusal skip intact. It is evidence of the stop, not a green landing claim.

## prefer-arrow-callback: shared fixer parse refusal

Exact Go helper: edit.FixText at cohere/internal/edit/engine.go:109, calling fixText and parsesWithTree at :182; parse failure refusals are recorded at :187. It discards the whole refused pass and retains original text. The shared Linter.fixed at stage1/cohere/lint/lint.ts:196-213 instead constructs Parser for the malformed intermediate source and lets its panic terminate the run. This requires shared parse-result/refusal handling, not a new per-rule checker question or a private fixer.

Upstream reproducer: `foo(bar || function() { this; }.bind(this));`, in TestPreferArrowCallbackFires at cohere/internal/lint/rules/core/prefer_arrow_callback_test.go:73 and :110. All five proposed edit ranges and texts match Go. Both edit engines reject the closing insertion at 42 because the .bind removal ends at 42. The admitted opening insertion plus missing closing insertion yield `foo(bar || (() => { this; });`. Go records overlap and parse refusals and returns the original. Native and Node replay panic at the final semicolon.

The original and upstream's fully repaired `foo(bar || (() => { this; }));` both parse unchanged on Node and sanitized native with rules disabled. Thus this is not an original-input parser gap or an invented fix-range approximation. The Go, native and replay logs under proof/arrow-refusal-* pin the difference. The registered candidate is deliberately red against this required case; no prefix is narrowed and no repair is withheld to hide it. Fourteen upstream cases preceding this one match all four runtimes. There are 83 captured cases, but 83-case agreement is not claimed. Witnesses and the self-reference mutant pass independently.
