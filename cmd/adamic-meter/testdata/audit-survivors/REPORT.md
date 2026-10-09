# Meter audit survivors, u009, task kyjn80b

Branch: `codex/meter-audit-survivors`, based on current main `e2492670b06a4dce1deafe837158ce0366cf2bc4`.
Audit evidence: `a15e94891617aa5f6b063378179fb8551987605f`, `review/test-audit/cmd-adamic-meter/`, original base `b955d303`.

All three standalone survivor diffs were read before editing tests and then applied independently with `git apply`. Production source was restored byte-for-byte in `finally`; only test files and this testdata evidence are committed.

## Survivor proofs

| Mutant | New top-level test | Mutant failure | Restored |
|---|---|---|---|
| M10 | `TestOptionalPositionCountsUTF16Surrogates` | UTF-16 column 20 gives byte 22, want 21 | Pass |
| M18 | `TestReturnAdaptationRewritesTS7030AndPreservesBehavior` | TS7030 rewrite count is 0, want 1; overlay is empty | Pass |
| M20 | `TestOptionalResolverPreservesUncheckedIndexDiagnostic` | Resolver loses the required-string assignment's TS2322 | Pass |

M10 pins the audit coordinate against upstream TypeScript's independent UTF-16 conversion. M18 uses a real TS7030 emitted by the linked TypeScript resolver with NoImplicitReturns enabled, not a fabricated diagnostic; it pins the rewritten source, diagnostic removal, disk preservation, idempotence, and original/adapted Node behavior with a shadowed undefined parameter. M20 checks the actual semantic diagnostic and its exact assignment location, with the ordinary loader as an independent control.

## The four previously surviving rows

- `TestOptionalAdaptationPreservesLiveMethodPlacement`: now includes an eligible optional-property assignment beside the live method. It checks the exact property-only source delta, rewrite accounting, the remaining method diagnostic, and Node prototype/own-property behavior. A mutant admitting MethodDeclaration rewrites both declarations and fails this row. This is a TS2412 guard, not a TS7030 guard.
- `TestImplicitReturnsNeedNoAdaptation`: retained and explicitly labelled as a smoke/regression test for current loader acceptance and Node fallthrough/finally behavior, because these inputs emit no TS7030 and cannot prove the return rewrite path.
- `TestReturnAdaptationLeavesUnprovenContracts`: now supplies genuine strict-checker TS7030 diagnostics and one eligible annotated control. Only that control may change; a mutant admitting unknown return types changes opaque too and fails this row.
- `TestReturnAdaptationOwnsOnlyRoots`: now supplies genuine TS7030 diagnostics for both an eligible root and an eligible imported function. Only the root may change; removing the ownership guard rewrites both and fails this row.

Each of those three guard mutants failed at its intended assertion and passed after restoration. An initial supplemental `if true` contract mutant was rejected by the Go compiler as unused-variable code, so it is excluded from kills; the recorded final diff instead widens the predicate to unknown while retaining its variable use, and fails the intended test assertion.

## Cold four-CPU measurements

Each top-level test ran in a fresh process, pinned to CPUs 0-3, with GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, and -timeout 90s. Go build artifacts were fetched from the content-addressed Go cache after setup; there are no native products in this unit. OS pages were not flushed. Times are complete command wall times including build-product fetch/linking and test execution; no successful test results were reused.

```sh
go test ./cmd/adamic-meter/ -count=1 -timeout 90s -v -run '^(<Name>)$' > <Name>.log 2>&1
```

| Top-level test | Complete cold command (s) |
|---|---:|
| `TestOptionalPositionCountsUTF16Surrogates` | 2.073 |
| `TestOptionalResolverPreservesUncheckedIndexDiagnostic` | 2.079 |
| `TestReturnAdaptationRewritesTS7030AndPreservesBehavior` | 2.227 |
| `TestFixtureCorpusCountsEveryDiagnosticKind` | 2.125 |
| `TestAdaptRewritesTypeOnlyImportInMemory` | 2.073 |
| `TestJSONIsMachineReadable` | 2.077 |
| `TestNormalizeDropsIncidentalDetails` | 2.033 |
| `TestAdaptOptionalPropertiesPreservesRuntimeAndDisk` | 2.372 |
| `TestOptionalAdaptationLeavesUnsafeContractsRejected` | 2.576 |
| `TestOptionalAdaptationRechecksPresenceNarrowing` | 2.075 |
| `TestOptionalAdaptationContextualDeclarations` | 2.525 |
| `TestOptionalAdaptationOwnsOnlyNamedRoots` | 2.072 |
| `TestOptionalAdaptationPreservesLiveMethodPlacement` | 2.223 |
| `TestOptionalAdaptationUTF16AndAdamicExtension` | 2.227 |
| `TestImplicitReturnsNeedNoAdaptation` | 2.527 |
| `TestReturnAdaptationLeavesUnprovenContracts` | 2.284 |
| `TestReturnAdaptationOwnsOnlyRoots` | 2.427 |

`go test ./cmd/adamic-meter/ -count=1 -timeout 90s` passed: Go printed `ok ... 1.856s`; complete wall time was 3.978 s. No top-level test exceeded 30 or 60 seconds; none reached 90 seconds. `proofs.json` records each mutant and restored command and exit. Logs went directly to files and are retained compressed under `logs/`.

## Setup and scope

Exported GOPROXY=https://proxy.golang.org|direct before bounded setup; setup exited 0. Timing lines: Go ready 0.016 s; Node ready 0.022 s; submodules ready 0.051 s; markdown ready 0.070 s; clang ready 0.156 s; Go build ready 35.105 s; test binaries deferred 35.249 s; cache warm 35.250 s; done 35.276 s. Toolchain sourced from /workspace/adamic-tools/env.sh. `nproc` printed 5; cgroup CPU quota is 400000/100000, four CPUs, and measurements were pinned to four CPUs.

No production changes, no persistent language fixtures, no counts change, no serial-baseline change, and no PR. The full repository gate and unrelated packages were not run. The integration lane checker is run from the repository root on the committed branch before push; its result is reported in the final response.
