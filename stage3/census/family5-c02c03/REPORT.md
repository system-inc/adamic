# Family 5 C02/C03: main-control continuation

Stopped: Stop: main-control-c02 is green where c02 is red; user instructed a stop

Main `fbe5b72ba115a8f3060d4b5229fd578574f4505f`; C02 `28a2e22015d0cf51667fcefa7acb246cab553817`; C03 `not created`. TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. All proof children in this continuation ran one at a time on the same four-core/16 GiB box.

## Main controls and candidate lanes

| Run | Real exit | Passing | Failing | Pending | Classification |
|---|---:|---:|---:|---:|---|
| main-control-c02 | 0 | 106366 | 1 | 0 | green |
| c02 | 1 | 106354 | 3 | 0 | red control or unresolved red |

c02: `compiler tests compiler tests for tests/cases/compiler/recursiveConditionalTypes.ts "before all" hook for "Correct errors for tests/cases/compiler/recursiveConditionalTypes.ts"`; Timeout of 40000ms exceeded. For async tests and hooks, ensure "done()" is called; if returning a Promise, ensure it resolves.; elapsed 42167.509889 ms.

c02: `conformance tests conformance tests for tests/cases/conformance/types/literal/templateLiteralTypes1.ts "before all" hook for "Correct errors for tests/cases/conformance/types/literal/templateLiteralTypes1.ts"`; Timeout of 40000ms exceeded. For async tests and hooks, ensure "done()" is called; if returning a Promise, ensure it resolves.; elapsed 43073.104485 ms.

Box-red acceptance follows the user's explicit instruction: a candidate's timeout-red lane is accepted only beside a timeout-red main control on this same box. No timeout or worker-count override was made. A green main control stops the candidate. Ordinary lane and npm child exits remain recorded as their real nonzero values.

Complete continuation evidence: [main-control](main-control/REPORT.md).

## Every actual proof child

| Proof receipt | Result | Real child exit |
|---|---|---:|
| main/lane-alone | PASS | 0 |

Stock checks enforce 79 roots and zero diagnostics in each configuration. Public compares use cmp on independently built declarations. Product commands use -timeout 45m. Full censuses use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, exact absolute LATENT_ONLY_FILE, and the tsc.ts entry. Complete streams and stderr are retained.

## Census outcomes and every non-gone finding


## Publication

{}

Each landing commit is verified to contain only its adaptation directory and one appended ledger line, with C03 directly parented by C02. No internal/ change. The prior five missing results remain accounted for in the earlier report: six test bodies were skipped after one failed before-all hook, which itself counted as one failure; net loss five.

A required proof absent from this receipt table was not run or is not complete; no success is implied. If a census comparison did not complete, target clearance and zero new findings remain unestablished. Previous reports/evidence remain in Git history and previous-control-report.md.
