# Family 5 C02/C03: sequential lane attribution

**Stopped at the required eight-worker lane. Neither landing is pushed.** The isolated C02 rerun exited **1**, with **106,354 passing, 3 failing, 0 pending**. The required result remains 106,366 passing and one sanctioned failure.

The measurements support **box/load sensitivity**. Neither original timeout recurred in the isolated whole lane; two different compilation hooks timed out. All four cases passed individually on both shas. All ten emitted JavaScript files and public `typescript.d.ts` are byte-identical between independently built main and C02 trees. A shared-path repeat removed the initial timing spread. This does not waive the red lane. Main's complete eight-worker lane was not run; paired cases and identical runtime artifacts are the attribution evidence.

Main: `fbe5b72ba115a8f3060d4b5229fd578574f4505f`. C02 local candidate: `28a2e22015d0cf51667fcefa7acb246cab553817`. C03 has no landing commit. Both adapted inputs use TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. The 748 source-file manifests differ only at types.ts; replacing C02's pinned Node owner with its pinned original makes that complete file byte-identical to main.

## Idle-box lane

The [initial process snapshot](sequential/idle-before-lane.txt) shows no other proof workload. The box has a **four-core CPU quota** (`400000 100000`), **16 GiB** memory, Linux x64 and Node v24.19.0. The unmodified lane used its standard **eight workers** and standard **40,000 ms** Mocha timeout. No product build or census ran alongside it. [Serial audit](sequential/serial-audit.json): 18 real-child receipts, no inferred overlap above its 10 ms rounding tolerance.

| Run | Passing | Failing | Pending | Total reported | Real lane exit | Child wall s |
|---|---:|---:|---:|---:|---:|---:|
| Previous concurrent run | 106359 | 3 | 0 | 106362 | 1 | 1864.450 |
| New isolated run | 106354 | 3 | 0 | 106357 | 1 | 808.298 |

The new oracle install and build exited **0**; its npm test child and the validated lane exited **1**. The test phase took **692.117 s**. [Whole lane report](sequential/c02/lane-receipts/report.json), [oracle report](sequential/c02/lane-receipts/oracle/report.json), complete compressed logs and raw Mocha errors are retained.

| Failure in isolated lane | Cause | Observed elapsed ms |
|---|---|---:|
| `compiler tests compiler tests for tests/cases/compiler/recursiveConditionalTypes.ts "before all" hook for "Correct errors for tests/cases/compiler/recursiveConditionalTypes.ts"` | timeout | 42167.510 |
| `conformance tests conformance tests for tests/cases/conformance/types/literal/templateLiteralTypes1.ts "before all" hook for "Correct errors for tests/cases/conformance/types/literal/templateLiteralTypes1.ts"` | timeout | 43073.104 |
| `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` | baseline-content | 2480.737 |

Both additional failures were `Timeout of 40000ms exceeded. For async tests and hooks, ensure "done()" is called; if returning a Promise, ensure it resolves.` The remaining failure is the sanctioned API acknowledgement baseline.

## Each case alone, main and C02

The original pair and both new failures were each measured serially, C02 then main, using the official oracle's `--tests` selection. Its documented targeted path uses one worker. No timeout override was set. Each side was checked out at its stated repository sha and used its matching adapted source tree and locked build.

Timing is the fourslash **test body**, or the compiler/conformance **before-all compilation hook**. A scratch preload derived from the existing error observer records pass/hook timings as well as errors; it is identical on both sides. [Observer](sequential/measure-runnables.cjs), [commands and orchestration](sequential/stage1.py), raw events, phase durations and actual exits are retained.

| Case | main s | C02 s | main oracle/test exit | C02 oracle/test exit |
|---|---:|---:|---:|---:|
| `codeFixClassImplementInterfaceNoTruncation.ts` | 4.964 | 14.632 | 0 / 0 | 0 / 0 |
| `relationComplexityError.ts` | 11.331 | 11.485 | 0 / 0 | 0 / 0 |
| `recursiveConditionalTypes.ts` | 5.870 | 5.119 | 0 / 0 | 0 / 0 |
| `templateLiteralTypes1.ts` | 7.406 | 17.071 | 0 / 0 | 0 / 0 |

Each compiler/conformance selection ran all six suite bodies successfully; the fourslash selection ran its one body successfully. No isolated timeout occurred on either side. The initial fourslash and template-literal timing spreads were not treated as evidence of a C02 regression.

## Shared-path check of the two timing spreads

Main's first tree was on /tmp and C02's on /workspace. The two timing outliers were repeated **main then C02** at the same /workspace path. Before each run, the complete source manifest matched that side and the executable JavaScript/public declaration manifest matched both independent builds. The restoring Node adapter selected main's source state; its normal replay selected C02. Standard targeted npm tests then ran one at a time. This input selection does not establish the latent mutant proof.

| Case, same path | main s | C02 s | main test exit | C02 test exit |
|---|---:|---:|---:|---:|
| `codeFixClassImplementInterfaceNoTruncation.ts` | 6.550 | 5.845 | 0 | 0 |
| `templateLiteralTypes1.ts` | 8.630 | 8.196 | 0 | 0 |

[Shared-path method](sequential/common-path.py), [raw result](sequential/common-path-result.json), and [exact C02 restoration](sequential/restoration.json). All prepare/restore children also exited 0. After the final run, all source files again match C02 and all emitted JavaScript/public declarations still match both sides. The original timing spread did not persist with the shared-path check. These measurements identify load/path sensitivity, not its particular scheduler, cache or GC mechanism.

## Where the five results went

The original `relationComplexityError.ts` suite registers **six** bodies:

1. Correct errors
2. Correct module resolution tracing
3. Correct sourcemap content
4. Correct JS output
5. Correct Sourcemap output
6. Correct type/symbol baselines

Its before-all hook failed. Mocha calls `done()` immediately on that hook error and never calls `runTests` for the suite. The six bodies therefore produced no pass, failure or pending events; the failed hook produced one failure. Net reported-total loss: **6 − 1 = 5**. The fourslash body's timeout replaced one pass with one failure and did not change the total.

```text
Original passes: 106366 - 6 skipped bodies - 1 failed fourslash body = 106359
Original failures: 1 sanctioned + 1 failed hook + 1 failed body = 3
Original total: 106367 - 6 skipped bodies + 1 failed hook = 106362
```

[Six-test registration](sequential/accounting/compilerRunner-suite.txt), [Mocha hook error path](sequential/accounting/mocha-before-all.txt), [original raw timeout observations](sequential/accounting/original-timeouts.json), and [machine-readable arithmetic](sequential/accounting/missing-five.json).

The isolated run instead lost two six-test suites to failed before-all hooks:

```text
Isolated passes: 106366 - 12 skipped bodies = 106354
Isolated failures: 1 sanctioned + 2 failed hooks = 3
Isolated total: 106367 - 12 skipped bodies + 2 failed hooks = 106357
```

[Isolated-run accounting](sequential/accounting/isolated-run-total.json).

## Proof gates and stop

| Requested line | Current result | Real child exit |
|---|---|---:|
| C02 whole lane, idle box, standard eight workers | FAIL: 106354 / 3 / 0 | 1 |
| Four isolated cases on main | PASS, all cases | 0 each |
| Four isolated cases on C02 | PASS, all cases | 0 each |
| Two shared-path cases on main | PASS, both | 0 each |
| Two shared-path cases on C02 | PASS, both | 0 each |
| TestProduct_Stage3AdaptedTree, -timeout 45m | NOT STARTED in this continuation; prior attempt was interrupted | — |
| Paired full G02 censuses | NOT STARTED in this continuation; prior comparison remains incomplete | — |
| Latent mutant site-restoration proof | NOT STARTED | — |
| C02 landing | HELD: lane remains red | — |
| C03 build, proofs and landing | NOT STARTED | — |

The prior same-candidate stock tsc (both configurations, 79 roots, zero diagnostics), adamic types, whole apply and public declaration comparison receipts remain PASS/0. They were not rerun here. No claim of target clearance, zero new latent findings, or reached-stop classification is made.

No timeout or worker-count change was made to obtain a green lane. A required gate is still red on this box, so this continuation stops without a landing push. The prior report and its complete partial evidence remain in [previous-stop-report.md](previous-stop-report.md).

## Complete new child receipt index

| Receipt | Result | Real child exit |
|---|---|---:|
| c02/lane-alone | FAIL | 1 |
| common-targeted/c02/codeFixClassImplementInterfaceNoTruncation/prepare | PASS | 0 |
| common-targeted/c02/codeFixClassImplementInterfaceNoTruncation/test | PASS | 0 |
| common-targeted/c02/templateLiteralTypes1/prepare | PASS | 0 |
| common-targeted/c02/templateLiteralTypes1/test | PASS | 0 |
| common-targeted/main/codeFixClassImplementInterfaceNoTruncation/prepare | PASS | 0 |
| common-targeted/main/codeFixClassImplementInterfaceNoTruncation/test | PASS | 0 |
| common-targeted/main/templateLiteralTypes1/prepare | PASS | 0 |
| common-targeted/main/templateLiteralTypes1/test | PASS | 0 |
| common-targeted/restore | PASS | 0 |
| targeted/c02/codeFixClassImplementInterfaceNoTruncation | PASS | 0 |
| targeted/c02/recursiveConditionalTypes | PASS | 0 |
| targeted/c02/relationComplexityError | PASS | 0 |
| targeted/c02/templateLiteralTypes1 | PASS | 0 |
| targeted/main/codeFixClassImplementInterfaceNoTruncation | PASS | 0 |
| targeted/main/recursiveConditionalTypes | PASS | 0 |
| targeted/main/relationComplexityError | PASS | 0 |
| targeted/main/templateLiteralTypes1 | PASS | 0 |

Each status file records exact commands, cwd, environment and child wall time. Full logs/events are gzip-compressed losslessly. No landing adaptation, ledger or internal/ file was changed during this continuation.
