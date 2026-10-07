Built: hoisted namespace singleton state, scoped enums, ordered nested bodies and live mutable exports; escape and callable/class merges stay explicit NotYet.
Commits: aff39ff, b0e9b15, f5161d8, 083e7a9, 2f40563, and the sixth commit containing this report.
Validation: affected packages pass; uncached namespace oracle passes in 4.052s with 13 fixtures and eight runtime mutants; counts pass in 16.752s.
Mutants: eight runtime mutations plus five refusal-guard removals caught; evidence and exploratory failures are recorded below.
Not covered: unchanged parser.ts native execution, original compiler bodies/types, namespace containers, callable/class merges, arbitrary initialization calls, overloads and destructured state.

## Base and setup

origin/main at ef3d907 did not contain internal/lower/namespaces.go, so this branch uses origin/codex/parameter-properties-namespaces at adc45ca. Stage 3 census was read from origin/codex/tsc-census at 429c117, including REPORT.md, the namespace bucket and r23. Enum implementation was integrated from Adamic's origin/codex/enums at 7127080, adapted to the current split files. No code was copied from cohere. No protected compiler files were edited. No PR was opened.

CLAUDE.md, the namespace documentation and the named implementation were read before editing. The cloud runtime skill was used for environment inspection. The initial setup failed during a checkout/build overlap with missing cmd/adamic/tsgo.go and cmd/adamic-meter/optional.go. Retrying `bash cloud/setup.sh` succeeded. Timing: Go 0s, clang 1s, Node 1s, submodules 1s, build cache warm 109s, total 109s. `nproc` reported 5; CPU quota was 4. Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0. Commands sourced /workspace/adamic-tools/env.sh. Setup logs: /tmp/namespaces-setup.log and /tmp/namespaces-setup-retry.log.

## Six steps

1. aff39ff: direct namespace var storage is hoisted. Reads of checker types including undefined observe undefined before assignment. Reads typed without undefined use ready checks rather than uninitialized native storage. A later var declaration without an initializer does not reset an assigned value. Parser state fixture covers initialization, reset, and the repeated declaration.
2. b0e9b15: constant enums, const enums and nested enum scopes extend the existing enum lowering, preserving its closed-domain refusal checks. Parser, nested JSDocParser and IncrementalParser enum fixtures agree with Node.
3. f5161d8: nested executable bodies and admitted statements run at module evaluation in source order. Arbitrary initialization calls, premature qualified reads and var hidden inside control flow remain NotYet.
4. 083e7a9: exported let/var uses shared singleton storage, including private, qualified and imported reads/writes, compound assignment and increment. Fixtures cover Debug state and cross-module updates.
5. 2f40563: tracingEnabled escape stays NotYet with the runtime-container reason. An object snapshot cannot preserve live mutable aliases, identity, receivers and staged exports.
6. This commit: separate callable merge, constructor merge and namespace-class reasons, refusal fixtures and mutants, all-ten shape census, final verification and documentation. Callable and class merges remain NotYet because their runtime identities are not implemented.

## Independent observations and their limits

The stock TypeScript 6.0.3 source is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Regeneration command:

```
NAMESPACE_TYPESCRIPT=/tmp/namespaces-stock/node_modules/typescript node stage3/namespaces/shape-census.cjs /tmp/namespaces-stock/source/src/compiler stage3/namespaces/shapes
```

The generator, census.json and results.json are committed. Declaration shapes preserve names, export flags, declaration kinds, missing initializers, eager calls, generic arity, overload signatures, destructuring, nested declarations and relevant merges. External types, function bodies and enum expressions are normalized. This establishes which isolated declaration shapes lower, not which original implementations run.

| Namespace | Shape result | Observed first blocker |
| --- | --- | --- |
| BuilderState | Lowers | None in normalized shape |
| JsxNames | Lowers | None in normalized shape |
| ReactNames | Lowers | None in normalized shape |
| BinaryExpressionState | Lowers | None in normalized shape |
| Parser.JSDocParser | Lowers | None in normalized shape |
| Parser | NotYet | Scanner/factory initialization call |
| IncrementalParser | NotYet | Bodyless overload signature |
| Debug | NotYet | Cache initialization call |
| Debug.log | NotYet | Callable namespace container |
| tracingEnabled | NotYet | Escaped namespace container |

Parser has 437 functions, 30 var declarations, two enums, a nested namespace, 18 overload signatures and a destructured factory var binding. Supporting its mutable state and enums is necessary but insufficient for the day 3 proof. Debug also has a class, overloads and a callable merge. Full compiler types and bodies were not compiled. The inference that the five accepted shapes can lower their original implementations is not established.

Census r23 was saved as a scratch .a file, built and run natively: stdout `1`, matching independent Node stdout `1` (`cmp` passed). The checker reports the uninitialized singleton fixture's token as number and read as () => number. Source Node reads undefined; native and checked JavaScript deliberately fail their ready check, matching Adamic's existing contract rather than silently returning a number. Source tracing escape prints 0, and the callable log example prints call then true, but Adamic refuses both shapes.

## Verification

All output went to log files, never through a pipe. The broad affected-package command was:

```
go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle
```

/tmp/namespaces-packages.log: lower 52.389s, native 273.998s, flow 168.080s, fresh 67.465s, oracle 242.005s, all pass; JavaScript and IR have no package tests. Full `go test ./...` was not run. After the final additional fixture assertions and hoisting mutant, focused lower/enum/shape tests passed in 2.884s (/tmp/namespaces-final-lower.log), and the final uncached oracle passed:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNamespace|TestNativeAgreesWithNode/internal/oracle/testdata/namespaces|TestNativeAgreesWithNode/stage3/namespaces/shapes' -count=1 -timeout 30m -v
```

/tmp/namespaces-final-uncached2.log: 13 fixtures and eight runtime mutants pass, 4.052s; native cache hits 0/misses 38, Node hits 0/misses 34. Recorded counts were regenerated with TestCountsAreRecorded (16.752s, /tmp/namespaces-counts-final3.log); a final read-only verification passed in 16.934s (/tmp/namespaces-counts-verify.log). `go vet ./...`, gofmt and `git diff --check` pass. A first incorrectly filtered oracle command matched no fixtures; it was replaced and is not counted as validation.

## Mutant evidence

| Mutation | What caught it |
| --- | --- |
| Wrong scoped function result | Independent Node stdout, clean native exit and sanitizers |
| Wrong scoped constant | Independent Node stdout, clean native exit and sanitizers |
| Parser enum SourceElements changed to 99 | Independent Node stdout, clean native exit and sanitizers |
| Swap the last two initialized namespace/module outputs | Independent Node stdout, clean native exit and sanitizers |
| Outside exported state write changed from 3 to 30 | Independent Node stdout, clean native exit and sanitizers |
| Drop assignment initializing parser string state | Independent Node comparison catches ready-check failure |
| Remove all hoisted token declarations | Independent Node comparison catches premature read failure |
| Remove singleton ready check | Checked JavaScript exit comparison catches native failure-contract difference |
| Remove escaped-container guard | Specific NotYet reason regression fails; later unbound-name refusal is insufficient |
| Remove callable merge guard | Specific callable-container reason regression fails |
| Remove constructor merge guard | Regression fails because formerly refused class merge lowers |
| Remove namespace-class guard | Regression fails because formerly refused class declaration lowers |
| Remove control-flow var guard | Regression fails because expected NotYet becomes a generic var Refused |

Runtime mutant evidence is in /tmp/namespaces-final-uncached2.log. Guard mutations were run against the real compiler, one at a time, then restored: /tmp/namespaces-escape-mutant.log, /tmp/namespaces-function-merge-mutant2.log, /tmp/namespaces-class-merge-mutant2.log, /tmp/namespaces-namespace-class-mutant.log, /tmp/namespaces-block-var-mutant.log. A guard mutant proves a diagnostic boundary, not correct runtime object semantics.

Exploratory mutations are not omitted: changing only the function diagnostic prefix survived because the asserted semantic reason was unchanged; actual guard removal replaced it. Removing only one hoisted declaration survived because the repeated var declaration retained another; the final mutant removes all copies and is caught. An early output reorder ran before state initialization and triggered readiness instead of the intended clean stdout disagreement; the final reorder swaps initialized outputs. These experiments do not count as additional caught mutants. All compiler mutations were restored before final verification.
