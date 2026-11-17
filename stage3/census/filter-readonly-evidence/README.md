# Filter feasibility: stop before an adapter

Task #ev14ch9. Base 03c140ec28a249b8b0f13eb24e34fc3d3df05266; upstream
050880ce59e30b356b686bd3144efe24f875ebc8 (TypeScript 6.0.3).
The report-only count was pushed before any source probe. No compiler, stage3
adapter, lane sanction, ledger, or landing commit is changed. No Loom submission.

## Count and limits

[Complete result-use inventory](../filter-overload-count/REPORT.md): 157 resolved
core-filter calls in 713 src roots; 89 select writable-result overloads. The
whole-checkout audit additionally resolves 15,480 call expressions in 6,487
outside-src roots reaching core through 20,614 modules, with zero additional calls.

| Overload | Calls | Reads only | Writable parameter | Writable slot/return | Identity |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 5 | 2 | 0 | 3 | 0 |
| 2 | 60 | 16 | 7 | 41 | 2 |
| 5 | 11 | 4 | 2 | 5 | 0 |
| 6 | 13 | 4 | 2 | 8 | 0 |

Non-read columns overlap. The 63 distinct non-read calls are quoted in the count
report. No-drop reachability is established locally for 14 yes and 4 no cases;
45 remain unknown. The alias subtask is explicitly incomplete, and an unknown is
never clearance to copy. Inferred mutable locals count as writable slots, even
where changing inference to readonly would itself need no explicit local edit.

## Road decision

(a) Retyping narrowing results readonly does not establish their element type.
The exact complete owner isolated at line 261 is refused by scan-all in
lowering.declareModule, both before and after:

```
before: overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined
after:  overload 1 of filter result readonly U[] cannot be served by implementation result readonly T[] | undefined
```

These are isolated-owner observations, not a successful readonly sys.ts run.
The readonly sys.ts probe fails loading on four compiler caller diagnostics.
Caller annotation fixes cannot change the isolated owner's own result refusal.

(b) Removing overloads 1 and 5 loses needed narrowing and adds 17 stock-checker
diagnostics. The no-caller-needs-it condition fails. Keeping another narrowing
overload would retain the array-result declaration problem.

(c) A checked view or new per-overload element proof is beyond an established
adaptation shape. Stop here as requested. No typescript/filter-readonly-1 branch,
no ledger line and no stage3 adaptation are created.

## Proof receipts and explicit non-results

| Check | Result |
| --- | --- |
| Fresh detached exact-base worktree | Used; recursive shallow submodules initialized |
| stage3/api npm ci | Passed, TypeScript 6.0.3 |
| Unchanged stage3/apply.sh | Passed; apply.log.gz |
| Virtual readonly stock comparison | 6 new diagnostics over the same 713-root baseline (2765 -> 2771) |
| Virtual dropped-overload comparison | 17 new diagnostics (2765 -> 2782) |
| Full single-file latent core, CommonJSBundle, no-output guards | Both complete: 14 findings each, zero new/gone/changed; all 14 in lowering phase |
| Generic filter in those latent runs | Deferred; absence of 261:1 there is not a proof of clearance |
| Latent compiler-root checker diagnostics | Baseline 0; readonly probe 4, retained in latent-delta.json |
| scan-all exact isolated owner | Before and readonly both stop at 261:1 in lowering.declareModule |
| scan-all readonly sys.ts | Load rejected, four compiler caller diagnostics; no declaration observation |
| scan-all baseline sys.ts | Cancelled without a result after isolated-owner obstruction; no outcome claimed |
| Core emitted JavaScript, comments retained and removed | Byte-identical, two hashes in source-probe.json.gz |
| Public API baseline | AST finds zero filter declarations; hash retained |
| Source drift | Baseline unchanged; disposable probe changes only core.ts |
| Restoring mutant | No adapter/mutant built; isolated before owner retains original refusal |
| Stock 79-root zero-new acceptance | Not achieved: readonly probe has four compiler diagnostics |
| adamic types zero | Not run for a landing adaptation; none exists |
| Adapted-tree product test, whole stage3/lane | Not run; no landing candidate |
| Upstream npm ci / hereby dts / generated d.ts equality | Not run; no landing candidate |
| Owners drift/half-adaptation/replay controls | Not built; no adapter |

The stock 713-root comparison intentionally preserves the same options and missing
ambient environment in every run; it is not the diagnostic-free 79-root acceptance
proof. It is used only for relative feasibility diagnostics. All raw observations
are retained, including the failed readonly load. The original baseline tree is
never edited. Only scratch source probes and report/evidence files are written.

## Reproduction

Use the count tools in ../filter-overload-count and the exact pinned unchanged
adapted tree. variants.cjs performs only virtual source reads and no emit:

```sh
node stage3/census/filter-overload-count/variants.cjs TREE base base.json
node stage3/census/filter-overload-count/variants.cjs TREE readonly readonly.json
node stage3/census/filter-overload-count/variants.cjs TREE drop drop.json
```

Build the existing latent/make_overlay.py overlay outside the checkout and its
latent/tool entry point. For each baseline or disposable readonly source tree:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=TREE/src/compiler/core.ts \
  LATENT TREE/src/compiler core.jsonl
```

Build scan-all with its existing make_overlay.py, as its README describes. Its
existing driver enforces both no-output APIs. Run on sys.ts to retain the load
failure; run on isolated-base.ts.txt and isolated-readonly.ts.txt copied to fresh
core.ts paths to observe the result-contract obstruction. The isolated texts are
exact full filter owners with leading blank lines preserving line 261. The only
probe source edit is retained in readonly-only.patch.
