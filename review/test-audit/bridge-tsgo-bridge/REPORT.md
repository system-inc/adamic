Audited three rows at origin/main b3f83786a0ea00c23d47d774d3b6c98f9dd71336; none moved or vanished.
Verdicts: one bounded sacred row and two setup-check rows.
Three production mutants, six construction edits, and two separate empty-answer probes.
One production survivor: M2 removes all fallback panics and the selected rows still pass.
All eleven final standalone diffs apply and pass go vet; nproc = 5.

```json
[
  {
    "test": "TestTSGoRequiresLink",
    "package": "bridge/tsgo",
    "file": "bridge/tsgo/bridge_test.go:22",
    "seconds": 0.072,
    "oracle": "Self-written requirement: unlinked Lower must return some error; linked Lower must succeed and UsesTSGo must be true. The negative assertion does not verify diagnostic identity.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [
      "M1",
      "M3"
    ],
    "last_proven_fail": "M3 bridge_test.go:38: opted-in bridge calls lost",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestTSGoRequiresLink",
      "TestBridgeUnitsCoverEveryPiece",
      "TestBridgeProductCacheIsVerified"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u001/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./bridge/tsgo/ -run ^(TestTSGoRequiresLink|TestBridgeUnitsCoverEveryPiece|TestBridgeProductCacheIsVerified)$ > M3.log 2>&1; bridge_test.go:38: opted-in bridge calls lost",
    "timing_samples": [
      0.072,
      0.066,
      0.076
    ]
  },
  {
    "test": "TestBridgeUnitsCoverEveryPiece",
    "package": "bridge/tsgo",
    "file": "bridge/tsgo/coverage_test.go:16",
    "seconds": 0.024,
    "oracle": "Self-written 36 bindings, helper identities, query counts and arithmetic spacing, corpus-mode active counts, and exactly one shard owner for counts 1..40.",
    "oracle_kind": "self",
    "kills": [
      "C1",
      "C2",
      "C3"
    ],
    "unique_kills": [
      "C1",
      "C2",
      "C3"
    ],
    "last_proven_fail": "C3 coverage_test.go:90: active bridge coverage count: 0, want 16",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTSGoRequiresLink",
      "TestBridgeUnitsCoverEveryPiece",
      "TestBridgeProductCacheIsVerified"
    ],
    "evidence": "ADAMIC_MUTANT=C3 ADAMIC_BUILD_CACHE_DIR=/tmp/u001/cache/C3 timeout 120 go test -json -count=1 -timeout 90s ./bridge/tsgo/ -run ^(TestTSGoRequiresLink|TestBridgeUnitsCoverEveryPiece|TestBridgeProductCacheIsVerified)$ > C3.log 2>&1; coverage_test.go:90: active bridge coverage count: 0, want 16",
    "timing_samples": [
      0.026,
      0.021,
      0.024
    ],
    "mutation_kind": "construction edits permitted for setup-check; no production-mutant verdict"
  },
  {
    "test": "TestBridgeProductCacheIsVerified",
    "package": "bridge/tsgo",
    "file": "bridge/tsgo/coverage_test.go:116",
    "seconds": 0.015,
    "oracle": "Self-written build count/path equality and rejection of corrupt bytes and an extra manifest entry. Rejections check only non-nil error, not its identity.",
    "oracle_kind": "self",
    "kills": [
      "C4",
      "C5",
      "C6"
    ],
    "unique_kills": [
      "C4",
      "C5",
      "C6"
    ],
    "last_proven_fail": "C6 coverage_test.go:148: corrupt product was accepted",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTSGoRequiresLink",
      "TestBridgeUnitsCoverEveryPiece",
      "TestBridgeProductCacheIsVerified"
    ],
    "evidence": "ADAMIC_MUTANT=C6 ADAMIC_BUILD_CACHE_DIR=/tmp/u001/cache/C6 timeout 120 go test -json -count=1 -timeout 90s ./bridge/tsgo/ -run ^(TestTSGoRequiresLink|TestBridgeUnitsCoverEveryPiece|TestBridgeProductCacheIsVerified)$ > C6.log 2>&1; coverage_test.go:148: corrupt product was accepted",
    "timing_samples": [
      0.021,
      0.012,
      0.015
    ],
    "mutation_kind": "construction edits permitted for setup-check; no production-mutant verdict"
  }
]
```

| ID | Origin file:line | Change | Failed rows in bounded matrix |
|---|---|---|---|
| M1 | internal/lower/tsgo.go:27 | Invert TSGo opt-in guard. | TestTSGoRequiresLink |
| M2 | internal/lower/tsgo.go:90 | Drop all three panic-setup statements; whole-block deletion avoids unused message. |  |
| M3 | internal/native/tsgo.go:18 | UsesTSGo recognized-name return true becomes false. | TestTSGoRequiresLink |
| C1 | bridge/tsgo/units_test.go:141 | Query cap 400 becomes 399. | TestBridgeUnitsCoverEveryPiece (corpus only) |
| C2 | bridge/tsgo/units_test.go:48 | Shard divisor count becomes count+1. | TestBridgeUnitsCoverEveryPiece |
| C3 | bridge/tsgo/units_test.go:44 | Corpus activation OR becomes AND. | TestBridgeUnitsCoverEveryPiece |
| C4 | bridge/tsgo/product_cache_test.go:16 | Drop the complete bridgeProductGet verification block. | TestBridgeProductCacheIsVerified |
| C5 | bridge/tsgo/product_cache_test.go:31 | Disable the manifest-count guard. | TestBridgeProductCacheIsVerified |
| C6 | bridge/tsgo/product_cache_test.go:39 | Disable the byte-digest guard. | TestBridgeProductCacheIsVerified |
| P1 | internal/lower/lower.go:20 | Return nil, nil at Lower entry. | TestTSGoRequiresLink |
| P2 | internal/native/tsgo.go:15 | Return false at UsesTSGo entry. | TestTSGoRequiresLink |

Survivors:

M2: all three bridge stubs lose their fallback panic: before panics 1/1/1, after 0/0/0. Final standalone M2 passes all three selected rows in sample and corpus mode. This is unguarded within the bounded unit; catches elsewhere remain unknown. Command: go run ./review/test-audit/bridge-tsgo-bridge/survivor-observe.go before and with M2.diff applied afterward. Full outputs: survivor-before.log and survivor-after.log.

C1 survives sample mode with identical 162 positions, but fails corpus mode at 399 versus 400; it is not a survivor of the combined matrix.

Brief feedback and audit costs:

- The row taxonomy matters here. Two of three named rows test construction, not bridge execution. I used the setup-check exception for C1..C6 and did not use those edits to declare a production test sacred. Three production mutants for one production row plus three edits per setup row is the interpretation of the per-row target.
- Vacuity has no specified aggregate rule for construction checks with multiple helper entries. Both construction rows are null, unprobed. P1 probes Lower and P2 probes UsesTSGo. Both fail TestTSGoRequiresLink. P1 stops at the negative assertion, so the positive half was not observed under P1.
- A coverage test belongs to a shared suite but is also explicitly listed as a row and fits setup-check. The brief's broad family wording could absorb it into the 36-piece bridge suite. I kept the three explicitly requested rows as the unit, with no family expansion.
- Warm tools were not warm bridge products at this new commit. The whole-package baseline spent its entire 90-second test budget building tsgo-asan.a. It timed out, not an assertion-red baseline. I stopped its orphaned Go build tree and narrowed to the three named rows. Whole-package uniqueness remains unknown.
- Corpus opt-in expects src/compiler layout. The existing compiler sources were under cohere/TypeScript/tsc/testdata/fixtures/compiler. A /tmp/u001/corpus/src/compiler symlink layout enabled the selected rows without downloading another corpus. Both unmutated modes passed. No selected row skipped. Other package rows were not completed, so their skip or failure outcomes are unknown.
- The sample is only 162 bytes, so 400 to 399 is unchanged on it. Corpus mode exposes the changed cap and kills C1. A default-mode-only run would miss that construction defect.
- Dropping M2's single prepend left message unused. Its switch compiled because the non-mutant branch retained the use, but its standalone patch failed vet. I repaired it by deleting the complete three-statement panic setup and reran the final standalone mutation in both modes. The failing initial patch is kept as .txt, not a replayable diff. The initial switched-source artifact predates this repair; diffs/M2.diff and final M2 logs are authoritative.

My execution mistakes also cost time: the first test-list command used stage3/api as cwd and failed before a valid root-cwd inventory was made; the first switch generator ran before sourcing env.sh for gofmt and was immediately formatted afterward; the first corpus runner reused log names, so I preserved corpus output and reran the default matrix; the M2 metadata updater had a duplicate command key and stopped after a passing standalone run, then the corrected replay completed both modes. No verdict rests on a failed driver or an invalid patch.

Timing and limits:

Setup skipped, 0 setup seconds, because /workspace/adamic-tools/env.sh worked. Warm verification command took 0.431 s. npm ci in stage3/api reported 872 ms for three packages. Go 1.27.1, Node 24.19.0, nproc 5.
Explicit switched binary build: 20.357 wall s. Formatting afterward invalidated that build cache and the first matrix control took another 21.523 wall s including Go build and test; this was an avoidable rebuild. Final standalone vet validation totals 9.794 wall s, excluding the initial failed M2 vet (0.510 s).
Retained baseline and three-alone-per-row commands total 30.488 wall s, including compile/tool invocation. Retained default matrix totals 30.134 wall s, corpus matrix 47.727 wall s. The first matrix overwritten during the mode mistake was 52.991 wall s as recorded in session output and was repeated. Whole-package clean test binary: 90.141 s, timeout. Additional M2 repair/replay times are in their timing JSON files; these include a 17.657 s compile-and-test pass before the metadata-driver error. Test binary medians, not command walls, are the seconds in rows.json. These phase totals overlap no test workload, except the early coverage inventory ran alongside the last timing commands; no claimed speedup is based on them.
No native product rebuilt for the selected matrix. Native products outside it, Go cohere, other packages, the full bridge suite, and repo-wide uniqueness are not audited. The survivor witness measures changed Lower output, not a native execution. Source restored and the selected rows pass afterward in restored.log. All file:line references use the starting commit. No PR or main push.
