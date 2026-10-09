Built: parser discovery row 11 Debug singleton probe lowers and matches Node; current main merged into the feature branch.
Commits: integration merge ebfa1d4350fcd0119be07b81f9f09c8a5942ddd3 plus the accompanying Debug regression/report commit.
Validation: exact probe and observation regression pass native sanitizers and checked JavaScript; twelve-fixture matrix remains 5 Compiles, 5 NotYet, 2 Refused.
Mutants: wrong Debug initialization and restoring the stale namespace undefined-observation guard are both caught by independent Node comparison.
Not covered: full parser execution, full Debug implementation, callable namespace merges or escaped namespace objects; no main or area branch was pushed.

## Integration repair

Merged origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b into codex/namespaces-tsc first, without rebasing. Read the batch 3 audit from origin/compiler/landing-batch-3 at a37ebdb0913eca9d3d6e8adbe8f01bb733da52eb. Item 15 records an aborted merge with 19 conflicts and says namespace localRead must preserve unknown-narrowing runtime union checks and captured/module-cycle readiness, while namespace readiness must coexist with assertion readiness. Its historical broad regex conflict resolution was rejected by automatic review and was not executed. This turn used explicit additive resolutions, with four conflicts against current main.

The identifier path uses the shared localRead helper. That helper now uses main's acceptsUndefined observation policy, preserving runtime checked reads, ready checks and union narrowing. The wideningRefusal helper retains both main's proven-relation checks and this branch's closed-enum refusal. Namespace and enum refusal checks precede predicate checks. Both emitters retain the separate readiness bits. Main currently does not represent unknown as a runtime value; no unknown language support was added. No protected emitter/lower entry-point files were manually edited during this repair.

## Exact source and evidence

Source bytes were taken from origin/codex/stage3-parser-proof 2179dd8, stage3/drivers/parser/native-debug-namespace.a, and registered unchanged as internal/oracle/testdata/namespaces_debug_probe.a:

```typescript
// From TypeScript 6.0.3, src/compiler/debug.ts:26
namespace Debug {
    export let isDebugging = false;
}
console.log(`${Debug.isDebugging}`);
```

Node, sanitized native code and checked JavaScript all exit 0, write `false\n`, and have empty stderr. This clears the minimal row 11 probe; the discovery list reached it behind explicit scratch stubs. It does not establish that the full parser or full Debug namespace runs natively.

The additional namespaces_observed_narrowing.a regression clears an optional exported number during a call after checker narrowing. Template and parameter observations still see undefined; number/string union observations retain their guards. All three engines print `undefined\nundefined\n3\naa\n` and exit 0. An exploratory version using unknown was located NotYet; it was replaced with the supported number|string shape rather than changing the language policy.

## Mutants

| Mutation | Check that fails |
| --- | --- |
| Change isDebugging initializer to true in IR | Node stdout comparison; native exits cleanly and sanitizers remain clean |
| Restore localRead's old !comparedWithUndefined condition | Native and checked JavaScript exit 70 with the narrowed-value ready-check panic, whereas Node exits 0 and observes undefined |

The second mutation was applied to the compiler and restored in a finally block. The restored regression passes. Logs: /tmp/namespaces-debug-focused2.log, /tmp/namespaces-debug-observation-mutant.log, /tmp/namespaces-debug-observation-restored.log. The first exploratory test failure is retained in /tmp/namespaces-debug-focused.log.

## Validation and counts

Commands source /workspace/adamic-tools/env.sh. All test output goes directly to logs.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(namespaces_debug_probe|namespaces_observed_narrowing)|TestNamespaceSemanticMutants/wrong_debug_initialization' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
python3 stage3/namespaces/progress-matrix.py --label 'Debug parser row 11 and batch 3 integration repair' --scratch /tmp/namespaces-debug-matrix
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/oracle ./internal/ir ./internal/flow ./internal/fresh
```

Focused oracle passes in 1.079s, counts regenerate in 35.599s, matrix passes, go vet and gofmt pass. The broad affected-package gate passes: load 3.969s, lower 97.302s, native 457.693s, oracle 424.900s, IR 7.324s, flow 241.474s, fresh 127.962s. Log: /tmp/namespaces-debug-packages.log. go vet ./... is clean; git diff --check is clean for this change. Full go test ./... was not run.

The twelve original namespace fixtures remain 5 Compiles, 5 NotYet, 2 Refused, with fresh Node comparisons for all compiled rows. No original matrix row changed. The new Debug discovery probe is a separate Compiles row, previously Refused in the parser scratch discovery report. Original 06 and 09 refusals remain unchanged; their exact programs and judgments remain in DECISIONS.md.

New count rows (allocations, frees, retains, releases, peak, regions) are Debug probe 1/1/0/1/1/0 and observed narrowing 6/6/5/14/3/0. Regenerating inherited baselines against current main also changes retain/release counts: namespaces_parser_enums 9/17 to 8/16; namespaces 17/57 to 16/56; parameter_properties 24/66 to 20/62; parameter_properties_ownership 192/232 to 184/224. Allocation/free/peak/region counts are unchanged. These are observed post-main-merge ownership baselines, not an effect of adding the boolean probe. Outputs and leak checks are held to Node; attributing each decrement to an individual main optimization was not established.

## Setup

Initial cloud/setup.sh failed because unresolved merge markers made the Go build syntactically invalid. It was retried after resolving the merge. Retry timings: Go 0.048s, Node 0.045s, submodules 0.151s, markdown dependencies 0.160s, clang 0.335s, Go build 43.184s, cache warm 43.430s, total 43.474s. nproc is 5, cgroup quota is four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. Logs: /tmp/namespaces-debug-setup.log and /tmp/namespaces-debug-setup-retry.log. No cohere code was copied.
