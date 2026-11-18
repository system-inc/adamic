# Family 5 C02/C03: green main control — stopped

**Main's control is green where C02's isolated lane was red. Stopped as instructed. Neither landing is pushed.** The required condition for accepting C02's red as box-red was not met. This control is not a diagnosis of a C02 regression.

## Side-by-side whole-lane results

Both runs used the same managed box, standard **eight workers**, standard **40,000 ms** case/hook timeout, Linux x64 and Node v24.19.0. The box has a four-core CPU quota and 16 GiB memory. Each run was alone, with no product test or census beside it. [Main's idle process snapshot](main-control/idle-before-main-lane.txt) and [box identity/resources](main-control/box.json) are retained.

| Run | Repository sha | Passing | Failing | Pending | Reported total | Real lane exit | Verdict |
|---|---|---:|---:|---:|---:|---:|---|
| New main control | `fbe5b72ba115a8f3060d4b5229fd578574f4505f` | 106366 | 1 | 0 | 106367 | **0** | **GREEN**, exactly the sanctioned failure |
| Prior isolated C02 | `28a2e22015d0cf51667fcefa7acb246cab553817` | 106354 | 3 | 0 | 106357 | **1** | **RED**, not accepted as box-red |

C02's row is the completed isolated run from evidence commit 7564af9b5; it was not rerun in this continuation. Both adapted inputs use TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

| Timing | Main control seconds | Prior C02 seconds |
|---|---:|---:|
| Whole lane child wall time | 1275.747 | 808.298 |
| Tests phase | 592.349 | 692.117 |

Main's whole-lane wall time includes building a missing adapted-tree cache entry. Test-phase time excludes that apply work. These are single-run observations.

Main had **zero timeout failures**, zero exceptions, no missing baselines and exactly one baseline-content failure:
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.

C02's two additional failures were:

- `recursiveConditionalTypes.ts`: before-all compilation hook, **42,167.510 ms**, 40,000 ms timeout.
- `templateLiteralTypes1.ts`: before-all compilation hook, **43,073.104 ms**, 40,000 ms timeout.

Main's original reports, full stdout/stderr, test log, raw Mocha observations and baseline diff are retained. [Main lane verdict](main-control/lane-records/main-control-c02/report.json), [main oracle report](main-control/lane-records/main-control-c02/oracle/report.json), [main real-child receipt](main-control/main/lane-alone/status.json), [C02 lane report](main-control/lane-records/c02/report.json), and [C02 real-child receipt](main-control/prior-c02-lane-status.json).

## Each proof line and real exit

| Line | Result | Real child exit |
|---|---|---:|
| Main whole stage3 lane, alone, eight workers | PASS: 106366 passing / 1 sanctioned failure | **0** |
| Main apply phase | PASS | **0** |
| Main npm ci | PASS | **0** |
| Main npm run build | PASS | **0** |
| Main npm test | One sanctioned API acknowledgement failure; no timeouts | **1** |
| Prior isolated C02 whole lane | FAIL: 106354 passing / 3 failures | **1** |
| TestProduct_Stage3AdaptedTree, -timeout 45m | NOT RUN in this continuation | — |
| Paired full G02 censuses | NOT RUN in this continuation | — |
| Latent mutant proof | NOT RUN in this continuation | — |
| C02 landing push | HELD | — |
| C03 build, proofs and push | NOT STARTED | — |

The validated main lane's exit 0 is its actual child exit: the checked-in lane accepts exactly the sanctioned API failure. Its underlying npm test exit 1 is also recorded, not hidden or rewritten.

The fresh main control still produces **all ten JavaScript files and public typescript.d.ts byte-identical to C02**: [complete hashes](main-control/fresh-main-artifacts.json). That executable identity remains true, but the user explicitly required a stop when main is green where C02 was red. No box-red waiver was applied and no further proof was started.

## Publication and retained work

- C02 remains local candidate `28a2e22015d0cf51667fcefa7acb246cab553817`.
- C03 has no landing commit.
- Neither `typescript/line-and-character-1` nor `typescript/type-constraint-1` was pushed.
- No adaptation, ledger or internal/ file was changed in this continuation.

Target clearance, zero new latent findings and reached-stop classification remain unestablished. The prior five-result gap and subsequent paired case timing work remain in [previous-control-report.md](previous-control-report.md) and the earlier evidence. The five were six skipped suite bodies replaced by one failed before-all hook in the reported count.

The [queued workflow](main-control/workflow.py) is retained for method transparency. Its product/census/mutant/C03 commands were not executed: it stopped immediately after reading the green main result. Unexecuted commands are not proof receipts. Large logs are losslessly gzip-compressed.
