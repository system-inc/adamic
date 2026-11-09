# Go record selection over tsc, task a9e21hc

The tsc entry did not reach Go's graph selection at f0c358f2. 1 of 81 file-entry groups entered `graphTypes`; 1 completed selection. Only `src/compiler/corePublic.ts` reached selection, and it selected no component. This is a report of the reachable measurement boundary, not proof that unmeasured tsc components contain no records.

The compiler is `f0c358f2caa605525a58dad49fabb38bd92418a9`. Adaptations are main's `fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d`, Memory's merge base with main. The report branch is based on refreshed main `d7b05dc830678b13734bcf0bc4c809aa9b75c8fe`; this does not change the measured compiler or adaptation pins. The TypeScript source pin is `050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3). The entry is `src/tsc/tsc.ts`; options are parsed from `src/compiler/tsconfig.json`. The Go checker reported 82 diagnostics, all retained in [dump.json](dump.json). This is a measurement on a checker-rejected program.

## Counts and record exposure

| Observation | Normal | throughRecord stop false |
| --- | ---: | ---: |
| File-entry groups attempted | 81 | 81 |
| Groups entering graphTypes | 1 | 1 |
| Groups completing selection | 1 | 1 |
| Selected component observations with a record | 0 | 0 |
| Selected component observations without a record | 0 | 0 |

These component counts describe observations from group attempts, not a deduplicated whole-tsc total. A group that never reaches selection has unknown component counts, not zero. Whole-tsc totals with and without records remain unavailable when the entry is blocked.

No selected record component was observed, so there is no measured tsc record-component seed or literal-site list to report. The complete selected-component arrays, including any non-record components, are in the dump.

`Args` and `AdamicJsonRecoveryObject` are not observed in a selected component. Their whole-program selection and the seed that might select either remain unresolved wherever their group stops before selection. The earlier static census at `0a064458` counted 37 and 3 literal sites; those counts do not establish Go selection, and are not substituted for seeds here.

| Record type | Declaration | Selection reached for its file group | Seed |
| --- | --- | --- | --- |
| Args | `src/compiler/tracing.ts:54:5` | False | unavailable before selection |
| AdamicJsonRecoveryObject | `src/compiler/commandLineParser.ts:2439:1` | False | unavailable before selection |

## Required state and the observed boundary

`lowerWithElementKey` in lower.go obtains module order, collects viewed elements, notes inheritance, checks enum initialization, records accessor and optional facts, checks namespace initialization, runs refusal scans, links declarations and lowers module bodies. It then finishes class/accessor/primitive and iterable work, computes exceptions, checks spreads, record guarded reads, dynamic getters and generator roads, and normalizes async IR. Only then does it call `findCycles`.

`findCycles` validates generated async identities, builds the finder from the actual module ASTs, adds lowering-created instantiations, then calls `graphTypes`. The selection needs complete closure and accessor captures, statics, instances, localTypes/localAlso, captured/global/function metadata, shared frame layouts, closed-frame-input proofs, lowered IR and writeSites. `fresh.ProveWrites` derives writes from that IR. Running the type walk on empty or partially failed lowering state would omit potential seeds and strong edges.

The test calls that production path for the entry and every non-declaration source file in the entry-root compiler program. Each file is a fresh group entry and uses its ordinary ESM dependency group. It does not lower only the successful statements of a failed module. The latent census's isolated units omit final ownership and module passes; their partially reconstructed state cannot be treated as completed Go selection. All group failures are labeled below and retained verbatim in JSON.

Entry boundary: `refuse`.

```text
src/compiler/core.ts:2599:15: stage 0 can't lower the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, yet
```

| Last phase before selection | Blocked groups |
| --- | ---: |
| declareModule | 1 |
| namespaceInitialization | 3 |
| refuse | 76 |

First blocking conditions, grouped by exact observed diagnostic. These are first failures, not an exhaustive inventory of later blockers:

| Diagnostic | Groups |
| --- | ---: |
| src/compiler/core.ts:2599:15: stage 0 can't lower the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, yet | 75 |
| src/compiler/commandLineParser.ts:293:35: stage 0 can't lower reading an enum before its runtime initialization; move the call after the enum declaration yet | 1 |
| src/compiler/debug.ts:525:37: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | 1 |
| src/compiler/diagnosticInformationMap.generated.ts:13:45: stage 0 can't lower reading an enum before its runtime initialization; move the call after the enum declaration yet | 1 |
| src/compiler/hostErrors.ts:3:59: stage 0 can't lower in when a possibly absent spread has no property presence descriptors yet | 1 |
| src/compiler/semver.ts:71:9: stage 0 can't lower a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration yet | 1 |

## Control and validation

The observer is inside Go's existing `if selected` Tarjan branch and reads the existing seed map. The control only disables the final throughRecord stop; record classification, early record-slot refusal, seed generation, edges and component selection remain intact. Every group is paired with that control. Validation compares every component's members, record flags, declarations, literal sites and seeds after removing only process-local checker IDs.

The observer fixture reaches selection with a record literal built from `a.next.slice()`. Its selected component has an array-element write seed and a field-write seed in `connect`. Normal mode stops at the record-in-component check; the control selects the same component and passes that stop. Fixture results are separate from tsc counts in [fixture.json](fixture.json). No native program is emitted or executed.

Setup, npm ci, adaptation, the final test build, the measured run and validation all returned exit code 0. [run-exit.json](run-exit.json) records the child's exit code and silent-timeout status; compressed logs preserve the commands' output. The watcher stops a process group after 30 minutes without log output. No process hit that limit. The scratch compiler test and instrumentation remain uncommitted; the report commit contains evidence only.

See [METHOD.md](METHOD.md), [record_select_census_test.go.txt](record_select_census_test.go.txt), [scratch.patch](scratch.patch), [source-sha256.json](source-sha256.json), and [patch-set.md](patch-set.md). `dump.json.header.esm_order` is Go's zero-based ESMModuleOrder from tsc; every selected record member's literal sites use those indices. The fixture uses its own load index 0.

## Group coverage ledger

Every row is a distinct entry attempted in the same entry-root compiler program. The status is for the normal run; the control's selection reach and components are checked separately. Exact blockers and timings are in JSON.

| Group entry | Last phase | Selection entered | Selection completed | Components |
| --- | --- | --- | --- | ---: |
| `src/tsc/tsc.ts` | refuse | False | False | unknown |
| `src/compiler/corePublic.ts` | findCycles | True | True | 0 |
| `src/compiler/core.ts` | refuse | False | False | unknown |
| `src/compiler/debug.ts` | namespaceInitialization | False | False | unknown |
| `src/compiler/semver.ts` | refuse | False | False | unknown |
| `src/compiler/performanceCore.ts` | refuse | False | False | unknown |
| `src/compiler/performance.ts` | refuse | False | False | unknown |
| `src/compiler/_namespaces/ts.performance.ts` | refuse | False | False | unknown |
| `src/compiler/hostErrors.ts` | declareModule | False | False | unknown |
| `src/compiler/tracing.ts` | refuse | False | False | unknown |
| `src/compiler/types.ts` | namespaceInitialization | False | False | unknown |
| `src/compiler/sys.ts` | refuse | False | False | unknown |
| `src/compiler/path.ts` | refuse | False | False | unknown |
| `src/compiler/diagnosticInformationMap.generated.ts` | namespaceInitialization | False | False | unknown |
| `src/compiler/scanner.ts` | refuse | False | False | unknown |
| `src/compiler/utilitiesPublic.ts` | refuse | False | False | unknown |
| `src/compiler/utilities.ts` | refuse | False | False | unknown |
| `src/compiler/factory/baseNodeFactory.ts` | refuse | False | False | unknown |
| `src/compiler/factory/parenthesizerRules.ts` | refuse | False | False | unknown |
| `src/compiler/factory/nodeConverters.ts` | refuse | False | False | unknown |
| `src/compiler/factory/nodeFactory.ts` | refuse | False | False | unknown |
| `src/compiler/factory/emitNode.ts` | refuse | False | False | unknown |
| `src/compiler/factory/emitHelpers.ts` | refuse | False | False | unknown |
| `src/compiler/factory/nodeTests.ts` | refuse | False | False | unknown |
| `src/compiler/factory/nodeChildren.ts` | refuse | False | False | unknown |
| `src/compiler/factory/utilities.ts` | refuse | False | False | unknown |
| `src/compiler/factory/utilitiesPublic.ts` | refuse | False | False | unknown |
| `src/compiler/parser.ts` | refuse | False | False | unknown |
| `src/compiler/commandLineParser.ts` | refuse | False | False | unknown |
| `src/compiler/moduleNameResolver.ts` | refuse | False | False | unknown |
| `src/compiler/binder.ts` | refuse | False | False | unknown |
| `src/compiler/symbolWalker.ts` | refuse | False | False | unknown |
| `src/compiler/moduleSpecifiers.ts` | refuse | False | False | unknown |
| `src/compiler/_namespaces/ts.moduleSpecifiers.ts` | refuse | False | False | unknown |
| `src/compiler/checker.ts` | refuse | False | False | unknown |
| `src/compiler/visitorPublic.ts` | refuse | False | False | unknown |
| `src/compiler/sourcemap.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/utilities.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/destructuring.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/classThis.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/namedEvaluation.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/taggedTemplate.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/ts.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/classFields.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/typeSerializer.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/legacyDecorators.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/esDecorators.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2017.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2018.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2019.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2020.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2021.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/esnext.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/jsx.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2016.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/es2015.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/generators.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/module/module.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/module/system.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/module/esnextAnd2015.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/module/impliedNodeFormatDependent.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/declarations/diagnostics.ts` | refuse | False | False | unknown |
| `src/compiler/transformers/declarations.ts` | refuse | False | False | unknown |
| `src/compiler/transformer.ts` | refuse | False | False | unknown |
| `src/compiler/emitter.ts` | refuse | False | False | unknown |
| `src/compiler/watchUtilities.ts` | refuse | False | False | unknown |
| `src/compiler/program.ts` | refuse | False | False | unknown |
| `src/compiler/programDiagnostics.ts` | refuse | False | False | unknown |
| `src/compiler/builderStatePublic.ts` | refuse | False | False | unknown |
| `src/compiler/builderState.ts` | refuse | False | False | unknown |
| `src/compiler/builder.ts` | refuse | False | False | unknown |
| `src/compiler/builderPublic.ts` | refuse | False | False | unknown |
| `src/compiler/resolutionCache.ts` | refuse | False | False | unknown |
| `src/compiler/watch.ts` | refuse | False | False | unknown |
| `src/compiler/watchPublic.ts` | refuse | False | False | unknown |
| `src/compiler/tsbuild.ts` | refuse | False | False | unknown |
| `src/compiler/tsbuildPublic.ts` | refuse | False | False | unknown |
| `src/compiler/executeCommandLine.ts` | refuse | False | False | unknown |
| `src/compiler/expressionToTypeNode.ts` | refuse | False | False | unknown |
| `src/compiler/_namespaces/ts.ts` | refuse | False | False | unknown |
| `src/tsc/_namespaces/ts.ts` | refuse | False | False | unknown |
