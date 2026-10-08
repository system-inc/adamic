# Wave 15 checker and helper audit

Audit only. Area inspected: `1f9e223d8ef6d0a914407d4b37d0e56568521f4d`. Worker evidence: `codex/typeaware-wave-15-checker` at `f59b2ce5d6d03a41c7b804b5ab48f802b931ce13`. Pinned cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`. This branch starts at the area and adds only this file. No checker, helper, rule or harness implementation is changed.

`@typescript-eslint/prefer-find` has no missing checker question: its 53 captured upstream cases, witness and compiling/executing mutant passed using context.checker. The 17 other claimed rules below remain blocked. Existing granular questions are not declared wholly absent merely because the rules need richer payloads. In particular GetSymbolAtLocation and GetResolvedSignature already back symbol-origin and signature-shape; their additional identity/declaration facts are what is missing. Shared native analyses must be implemented natively, not replaced by a checker returning Go lint verdicts.

## Missing fact contracts and native helpers

Call-site paths are relative to `cohere/internal/lint/rules/`, except the explicitly named checking and compiler paths.

| Exact Go call or helper | Rules blocked | Call site | Missing contract |
| --- | --- | --- | --- |
| `checker.Signature_parameters; checker.Signature_declaration; checker.Checker_getTypeOfSymbol; ctx.TypeChecker.GetSignatureFromDeclaration` | @typescript-eslint/no-useless-default-assignment | `typescript/no_useless_default_assignment.go:370, :380, :396, :445` | All contextual/declared signature parameters, optional/rest declaration facts, type identities and signature declaration identity. contextual-shape and call-parameters expose types but not this contract. |
| `type_checking.GetTypeName` | @typescript-eslint/require-array-sort-compare | `typescript/require_array_sort_compare.go:169` | Shared type-name normalization, including type-parameter annotation/declaration constraints; name returns TypeToString instead. |
| `ctx.TypeChecker.GetNonNullableType` | nexus/correctness-no-global-listener-target-assertion | `nexus/correctness_no_global_listener_target_assertion.go:257` | Nonnullable type identities usable as both operands of IsTypeAssignableTo, rather than guessing from displayed names. |
| `ctx.TypeChecker.GetSymbolAtLocation` | nexus/correctness-no-global-listener-target-assertion; nexus/correctness-no-mock-on-module-namespace; no-invalid-regexp; no-throw-literal; prefer-arrow-callback; react/jsx-fragments; react/jsx-no-constructed-context-values; react/jsx-no-undef | `nexus/correctness_no_global_listener_target_assertion.go:237; nexus/correctness_no_mock_on_module_namespace.go:179; core/no_invalid_regexp.go:143; core/no_undef_init.go:163; core/prefer_arrow_callback.go:343; react/jsx_fragments.go:290; react/jsx_no_constructed_context_values.go:320, :570; react/jsx_no_undef.go:101` | Stable resolved symbol identity and all declarations, source identity/IsDeclarationFile, imports and binding names, initializer/container details. symbol-origin returns one filename; declarations accepts only class/interface selectors. |
| `rule.IsDeclaredOnlyInDeclarationFiles` | no-invalid-regexp | `core/no_invalid_regexp.go:143` | Shared global predicate requires complete declarations and IsDeclarationFile. A local name comparison or one origin file cannot replace it. |
| `identifierIsShadowed` | no-throw-literal | `core/no_undef_init.go:158` | Shared shadow predicate requires every declaration and its source IsDeclarationFile bit. |
| `type_checking.IsSourceFileDefaultLibrary` | nexus/correctness-no-global-listener-target-assertion | `nexus/correctness_no_global_listener_target_assertion.go:221; checking/builtins.go:25` | Default-library ownership of every symbol declaration, using program path identity and library options. |
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

## Reproducers and original stopping calls

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

### no-throw-literal

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_undef_init.go:163`.

```typescript
function f(undefined: Error) { throw undefined; }
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

### prefer-arrow-callback

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(inner) == declared` at `cohere/internal/lint/rules/core/prefer_arrow_callback.go:343`.

```typescript
[1].map(function f(x) { return f(x); });
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

### react/jsx-fragments

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_fragments.go:290`.

```typescript
import { Fragment as F } from 'react'; const v = <F/>;
```

### react/jsx-no-constructed-context-values

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:320`.

```typescript
function C() { const dep = {}; const value = useMemo(() => ({}), [dep]); return <Ctx.Provider value={value}/>; }
```

### react/jsx-no-undef

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(reference)` at `cohere/internal/lint/rules/react/jsx_no_undef.go:101`.

```typescript
function C() { return <Missing/>; }
```

### @typescript-eslint/require-array-sort-compare

Exact upstream call: `type_checking.GetTypeName(ctx.TypeChecker, argument)` at `cohere/internal/lint/rules/typescript/require_array_sort_compare.go:169`.

```typescript
function f<T extends string>(a: T[]) { a.sort(); }
```

## Scope and evidence

The previous full lint invocation passed 114 tests/subtests, failed 0, and skipped only TestCheckerBridgeRefusalPending, with all 76 registered mutants caught. That proves the prefer-find port and inherited registry, not the 17 blocked ports. The 875-file shared corpus manifest is syntax-only; no checker-enabled full-corpus parity is claimed.

The historical merge 5fed7283f is not an ancestor of the checker port or either new branch. Its RootedFilePath/SourceFile.Path build errors are an old pin adaptation problem, separate from this missing-facts inventory; they are not evidence that the current area checker fails. Blocker documents and historical logs are excluded from the clean lint-rules/prefer-find port branch.
