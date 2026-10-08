# Wave 30 checker audit

Label: wave-30. Audit basis: area/stage1-lint c4bdc23fa86d55cf7e579989201c11258f4d3a62; cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. Archive branch: codex/typeaware-wave-30, merged and pushed at 8431009c828acd09ca199610aea2124421c3e716.

This is an audit, not a port. RuleContext.checker, parser/checker node correlation and JSX parsing are available. All fourteen retained claims stopped on the missing bridge facts or shared analysis helpers below. No new claims, private checker, copied shared helper, descriptor or port implementation was added. The previously released nexus/concurrency-no-lost-update claim is excluded.

## Missing bridge answers and shared helpers

The Go calls exist upstream. What is missing is their complete answer on the area bridge, or the corresponding shared native analysis. A call below can require more payload than the area exposes for a related question. This lists the blocking calls actually audited; it is not a complete inventory of every dependency that later implementation may reveal.

| Exact symbol as cohere calls it | Rules blocked | Required payload |
| --- | --- | --- |
| `ctx.TypeChecker.GetSymbolAtLocation` | `nexus/consistency-no-iso-string-date-cut`, `nexus/correctness-no-callback-in-parse-try`, `nexus/correctness-no-collection-misuse`, `nexus/correctness-no-discarded-pure-result`, `nexus/correctness-no-uncleared-race-timeout`, `nexus/correctness-no-process-exit-after-output`, `nexus/correctness-require-blocking-standard-streams`, `react/jsx-fragments`, `react/jsx-no-constructed-context-values`, `react/jsx-no-undef` | Symbol presence and identity, all declaration spans/files/kinds, declaration parent/initializer/import shapes. Not just a ValueDeclaration filename. |
| `ctx.TypeChecker.GetShorthandAssignmentValueSymbol` | `nexus/correctness-no-callback-in-parse-try`, `nexus/correctness-no-uncleared-race-timeout` | Value-symbol identity for shorthand assignments, comparable with other reference identities. |
| `ctx.TypeChecker.GetAliasedSymbol` | `nexus/correctness-no-process-exit-after-output`, `nexus/correctness-require-blocking-standard-streams` | Aliased target identity and declarations, including imports across files. |
| `type_checking.IsSymbolFromDefaultLibrary` | `nexus/consistency-no-iso-string-date-cut`, `nexus/correctness-no-collection-misuse`, `nexus/correctness-no-uncleared-race-timeout` | Default-library provenance for the symbol at a value/member expression. type-origin on a type symbol is not a substitute. |
| `type_checking.IsSourceFileDefaultLibrary` | `nexus/correctness-no-discarded-pure-result`, `nexus/correctness-no-uncleared-race-timeout` | Default-library status for the referenced declaration file, alongside its owner shape. |
| `checker.Checker_getAwaitedType` | `nexus/correctness-no-discarded-outcome` | Awaited type graph of the returned value. |
| `checker.Type_symbol; symbol.Declarations; declaration.Parent` | `nexus/correctness-no-discarded-outcome` | Type-literal declaration identity/ancestry, enclosing union arms and alias identity/source path. type-symbol currently exposes only a name. |
| `writers.ctx.TypeChecker.GetResolvedSignature; writers.ctx.TypeChecker.GetReturnTypeOfSignature` | `nexus/correctness-no-process-exit-after-output` | Resolved target declaration/body and its selected signature return type. signature-shape exposes parameters but not the declaration/body; call-returns enumerates signatures instead of selecting the resolved one. |
| `high_level_intermediate_representation.ForFunction` | `react-hooks/purity` | Shared native React HIR, single assignment and render/capture analysis. |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | `react-hooks/refs` | Shared native React HIR, single assignment and capture analysis. |
| `hir.ForFunction; hir.CloneFunction; hir.AnalyzePreservedManualMemoization` | `react-hooks/preserve-manual-memoization` | Shared native React HIR, SSA, reactive scopes and memoization/capture pipeline. |

## Per-rule source locations and reproducers

Locations below refer to the pinned cohere source above. Reproducers identify the requested behavior; they are not newly certified witnesses. Historical standalone evidence remains on the archive and does not certify the unified checker.

| Rule | Upstream file:line | Blocking call/data |
| --- | --- | --- |
| `nexus/consistency-no-iso-string-date-cut` | `cohere/internal/lint/rules/nexus/consistency_no_iso_string_date_cut.go:248` | `GetSymbolAtLocation(identifier), symbol.Declarations and declaration initializer/parent` |
| `nexus/correctness-no-callback-in-parse-try` | `cohere/internal/lint/rules/nexus/correctness_no_callback_in_parse_try.go:160` | `GetShorthandAssignmentValueSymbol(parent), with GetSymbolAtLocation at 147/158` |
| `nexus/correctness-no-collection-misuse` | `cohere/internal/lint/rules/nexus/correctness_no_collection_misuse.go:429` | `GetSymbolAtLocation(object), IsSymbolFromDefaultLibrary(ctx.Program, symbol)` |
| `nexus/correctness-no-discarded-outcome` | `cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome.go:135` | `checker.Checker_getAwaitedType(ctx.TypeChecker, valueType); symbol.Declarations at 177` |
| `nexus/correctness-no-discarded-pure-result` | `cohere/internal/lint/rules/nexus/correctness_no_discarded_pure_result.go:153` | `GetSymbolAtLocation(name); declaration.Parent at 158; IsSourceFileDefaultLibrary at 167` |
| `nexus/correctness-no-uncleared-race-timeout` | `cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:337` | `GetShorthandAssignmentValueSymbol(parent), GetSymbolAtLocation at 335` |
| `nexus/correctness-no-process-exit-after-output` | `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:413` | `GetAliasedSymbol(symbol), GetSymbolAtLocation at 411` |
| `nexus/correctness-require-blocking-standard-streams` | `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:586` | `GetAliasedSymbol(symbol), declaration traversal at 606` |
| `react/jsx-fragments` | `cohere/internal/lint/rules/react/jsx_fragments.go:290` | `GetSymbolAtLocation(identifier), symbol.Declarations at 294` |
| `react/jsx-no-constructed-context-values` | `cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:320` | `GetSymbolAtLocation(identifier), declaration.AsVariableDeclaration().Initializer at 328` |
| `react/jsx-no-undef` | `cohere/internal/lint/rules/react/jsx_no_undef.go:101` | `GetSymbolAtLocation(reference), jsxNoUndefDeclaredInFile at 102` |
| `react-hooks/purity` | `cohere/internal/lint/rules/react/purity.go:250` | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` |
| `react-hooks/refs` | `cohere/internal/lint/rules/react/refs.go:184` | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` |
| `react-hooks/preserve-manual-memoization` | `cohere/internal/lint/rules/react/preserve_manual_memoization.go:195` | `hir.CloneFunction(hir.ForFunction(ctx, functionNode)); hir.AnalyzePreservedManualMemoization at 199` |

### nexus/consistency-no-iso-string-date-cut

```ts
const iso = new Date().toISOString(); const day = iso.slice(0, 10);
```

### nexus/correctness-no-callback-in-parse-try

```ts
export function readSettings(text: string, callback: (error: Error | null, settings?: unknown) => void): void { try { callback(null, JSON.parse(text)); } catch { callback(new Error("settings are not JSON")); } }
```

### nexus/correctness-no-collection-misuse

```ts
Object.keys(new Map([["a", 1]]));
```

### nexus/correctness-no-discarded-outcome

```ts
import { parseJson } from "./nexus/source/structured-text/json/Json"; parseJson("{}", "webhook payload");
```

### nexus/correctness-no-discarded-pure-result

```ts
const items = [1, 2]; items.map(value => value + 1);
```

### nexus/correctness-no-uncleared-race-timeout

```ts
declare function work(): Promise<void>; Promise.race([work(), new Promise<never>((_resolve, reject) => setTimeout(() => reject(new Error("timeout")), 1))]);
```

### nexus/correctness-no-process-exit-after-output

```ts
console.log("starting"); process.exit(1);
```

### nexus/correctness-require-blocking-standard-streams

```ts
console.log("done"); process.exit(0);
```

### react/jsx-fragments

```tsx
import { Fragment as F } from "react"; const value = <F />;
```

### react/jsx-no-constructed-context-values

```tsx
import { createContext } from "react"; const Ctx = createContext({a: 0}); function Component(){ return <Ctx.Provider value={{a: 1}}/>; }
```

### react/jsx-no-undef

```tsx
const value = <Missing />;
```

### react-hooks/purity

```tsx
function Component(){ return <div>{Math.random()}</div>; }
```

### react-hooks/refs

```tsx
function Component(props){ return <div>{props.ref.current}</div>; }
```

### react-hooks/preserve-manual-memoization

```tsx
function Component(props){ const data = useMemo(() => props.items.edges.nodes ?? [], [props.items?.edges?.nodes]); return <Foo data={data}/>; }
```

The process snippets need the existing Node declaration/project fixtures at stage1/cohere/typeaware/testdata/wave_30_exit_controls.json and wave_30_blocking_controls.json on the archive. The discarded-outcome snippet needs the Nexus declaration fixture at the exact listed source path. Reproduce the previous positive standalone controls from the archive with:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/typeaware -run '^TestWave30(AgreementAndMutants|NextAgreementAndMutants|ThirdAgreementAndMutants|ProcessExitAgreement|BlockingStreamsAgreement|JsxPrerequisites|ReactPrerequisites)$' -count=1 -timeout 30m -v > /tmp/wave-30-parked.log 2>&1
```

The command currently requires the archive bridge API migrations below before it can compile. It is a historical standalone control recipe, not a successful rerun at this pin.

## Why existing questions do not supply these answers

`symbol-origin` returns only the filename of Symbol.ValueDeclaration. An empty filename cannot distinguish no symbol from a symbol without a value declaration. `type-origin` supplies a type symbol and its declaration-file provenance, not an arbitrary value-expression symbol. `declarations` accepts only classes/interfaces. These cannot answer import/variable declaration traversal, exact symbol identity or shorthand binding identity. The archive-only richer questions were not copied into the new checker harness.

## Legacy bridge migration inventory

These are obsolete upstream API assumptions in the old archive, distinct from the missing area question payloads above. No shared migration was performed in this audit.

| Old API assumption | Current pinned API | Archive source sites |
| --- | --- | --- |
| `SourceFile.Path()` | `SourceFile.PathKey()` | bridge/tsgo/checker/declaration_facts.go:22; type_leaf_facts.go:56 |
| `SourceFile.FileName()` returns a plain string | Returns `tspath.RootedFilePath`; `.AsString()` is needed at string wire boundaries | declaration_facts.go:17; facts.go:303; program_module_resolution.go:17,64; resolved_call_target.go:33; symbol_declaration_paths.go:27; type_declaration_ancestry.go:41; type_leaf_facts.go:54 |
| `Program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)` accepts a path | `Program.GetSourceFileForResolvedModule(resolved *module.ResolvedModule)` | bridge/tsgo/checker/program_module_resolution.go:63; current cohere/TypeScript/tsc/internal/compiler/program.go:2098 |

The archive also contains question implementations absent from the audited area: `symbol-lineage`, `symbol-identities`, `type-declaration-ancestry`, `type-leaf-facts`, `process-node-fields`, `resolved-call-target`, `symbol-declaration-paths`, `program-module-resolution`, `source-parse-context`. Their Go implementations and owned Adamic decoders remain on codex/typeaware-wave-30. Landing needs to choose shared question ownership and migrate the path APIs before moving them; they are not installed by this audit.

## Observed gate result

The required archive merge was pushed. The all-input lint package failed at build time; no tests executed: 0 pass, 0 test fail, 0 skip, one package build failure. Command: `go test -json -count=1 -timeout=60m ./stage1/cohere/lint`. Inputs: GOMAXPROCS=4; GOFLAGS=-buildvcs=false; ADAMIC_LINT_BENCH=1; ADAMIC_TYPESCRIPT_SOURCE set to clean TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8; both profile inputs set to one fresh directory.

Retry wall time 30.206776066s; nproc 5; one-minute load 0.951 at start, 3.416 at end. Build errors include declaration_facts.go:17 (RootedFilePath passed as string), declaration_facts.go:22 (SourceFile.Path removed) and program_module_resolution.go:63 (RootedFilePath passed where *module.ResolvedModule is required). Setup also failed with these diagnostics and disk exhaustion. No test guard was changed, no skip was introduced or relaxed, and no upstream cases or new mutants were reported as passed.

Original local evidence: /workspace/unpark-wave30-lint-retry.jsonl, /workspace/unpark-wave30-lint-retry-summary.json, /workspace/unpark-setup.log and /workspace/unpark-wave30-push.log. This single file embeds the audit and reproducers so its useful content does not depend on those machine-local files.
