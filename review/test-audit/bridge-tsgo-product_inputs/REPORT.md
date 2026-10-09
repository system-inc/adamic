All 20 requested names exist at origin/main b3f83786a0ea00c23d47d774d3b6c98f9dd71336.
Grouped as one 19-member product-input family and one separate product-coverage setup check.
Bounded verdicts: input family sacred on M1; coverage setup-check on S2 and S3.
Six production mutants, three construction breaks and three separate empty-answer probes were run.
Evidence is committed on test-audit/bridge-tsgo-product_inputs; production files are restored.

```json
[
  {
    "test": "TestBridgeProductInput family",
    "package": "bridge/tsgo",
    "file": "bridge/tsgo/product_inputs_test.go",
    "seconds": 3.983,
    "oracle": "self-written version one/version two and key inequality, using synthetic dependencies and products; no real bridge result or external authority comparison",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M2: product_inputs_test.go:151: api served stale product: \"version one\", want version two",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBridgeProductInput family",
      "TestBridgeProductUnitsCoverEveryProduct",
      "TestBridgeProductCacheIsVerified"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./bridge/tsgo/ -run ^(TestBridgeProductUnitsCoverEveryProduct|TestBridgeProductInput_.*|TestBridgeProductCacheIsVerified)$; ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u002/cache/M1; product_inputs_test.go:151: api served stale product: \"version one\", want version two",
    "members": [
      "TestBridgeProductInput_TsgoArchive",
      "TestBridgeProductInput_TsgoSanitizedArchive",
      "TestBridgeProductInput_LengthArchive",
      "TestBridgeProductInput_StaleArchive",
      "TestBridgeProductInput_WrongArchive",
      "TestBridgeProductInput_LeakArchive",
      "TestBridgeProductInput_Stage0",
      "TestBridgeProductInput_Oracle",
      "TestBridgeProductInput_ApiDriver",
      "TestBridgeProductInput_LengthDriver",
      "TestBridgeProductInput_StaleDriver",
      "TestBridgeProductInput_LeakDriver",
      "TestBridgeProductInput_SanitizedNative",
      "TestBridgeProductInput_Native",
      "TestBridgeProductInput_WrongNative",
      "TestBridgeProductInput_HealthyRegion",
      "TestBridgeProductInput_RegionStage0",
      "TestBridgeProductInput_RegionNative",
      "TestBridgeProductInput_LinkageTest"
    ]
  },
  {
    "test": "TestBridgeProductUnitsCoverEveryProduct",
    "package": "bridge/tsgo",
    "file": "bridge/tsgo/product_inputs_test.go",
    "seconds": 0.01,
    "oracle": "self-written count 19 and presence of product bindings; setup construction, not checker output",
    "oracle_kind": "self",
    "kills": [
      "S2",
      "S3"
    ],
    "unique_kills": [
      "S2",
      "S3"
    ],
    "last_proven_fail": "S3: product_inputs_test.go:81: TestBridgeABI fetches unknown product unknown-product",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestBridgeProductInput family",
      "TestBridgeProductUnitsCoverEveryProduct",
      "TestBridgeProductCacheIsVerified"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./bridge/tsgo/ -run ^(TestBridgeProductUnitsCoverEveryProduct|TestBridgeProductInput_.*|TestBridgeProductCacheIsVerified)$; ADAMIC_MUTANT=S2 ADAMIC_BUILD_CACHE_DIR=/tmp/u002/cache/S2; product_inputs_test.go:71: product coverage: 18 products, 19 units, want 19"
  }
]
```

| ID | Origin file:line | Change | Failed grouped rows |
|---|---|---|---|
| M1 | internal/buildcache/buildcache.go:176 | change constant: content digest uses empty bytes | TestBridgeProductInput family |
| M2 | internal/buildcache/buildcache.go:132 | drop entire flag loop | TestBridgeProductCacheIsVerified, TestBridgeProductInput family |
| M3 | internal/buildcache/buildcache.go:135 | drop entire toolchain loop | [] |
| M4 | internal/buildcache/buildcache.go:126 | drop name field | [] |
| M5 | internal/buildcache/buildcache.go:125 | change version constant | [] |
| M6 | internal/buildcache/buildcache.go:84 | flip first hit condition to false | [] |
| S1 | bridge/tsgo/products_test.go:38 | setup construction: drop recipe file | [] |
| S2 | bridge/tsgo/products_test.go:23 | setup construction: drop one product | TestBridgeProductUnitsCoverEveryProduct |
| S3 | bridge/tsgo/product_units_test.go:20 | setup construction: change abi dependency constant | TestBridgeProductUnitsCoverEveryProduct |

Survivors:
M3: unguarded in this bounded set. Distinct toolchain inputs produce distinct baseline keys, but both produce 48dd3204a4ae4339533e70ce27e8956162411d26dec0e8a5a261837895a83c05 under M3.
M4: unguarded in this bounded set. Distinct names produce distinct baseline keys, but both produce 8f754268f9847614630fb441be4efbb18a8da122bedcd724d86e85f8985b7aa7 under M4.
M5: observed changed key, survivor. Base key 5355a5a61d0fedc2fe682f353093cf1aa3eca9b8631bc08b611031a7b664966f becomes 2372385525e561909c21c6d902cdf2484fe6fa7227a59beb542e455f18bc25eb. A cache schema version change is not itself evidence of incorrect caching.
M6: equivalent candidate. Both reuse witnesses print builds=1 same=true product=answer error=<nil>. The second cache-hit check preserves reuse after locking; no changed functional output demonstrated.
S1: equivalent candidate. Both recipe witnesses report recipe_present=true files=1119 and digest=f41b9377728e8ed695b0774eb40ab7ae56f2f14ed2759864088b6f4767af19da. Go list rediscovers the recipe after its explicit map entry is removed.

Probes, excluded from mutant kills and uniqueness:
P1 buildcache.Key at buildcache.go:120 returns empty string, nil; kills family and adjacent cache row.
P2 buildcache.Get at buildcache.go:54 returns empty directory, nil; kills family and adjacent cache row.
P3 bridgeProductInputs at products_test.go:28 returns empty Inputs; kills family only.
Coverage does not call these entries and was not probed, so vacuous is null.

Brief ambiguities and time costs:
1. Twenty supplied roots collapse to two rows under the family rule. Nineteen wrappers have one checker and one registry, although there is no literal generator in the current file. Family medians are for running all nineteen together, not a sum of individually timed tests.
2. Coverage checks the TestProduct_* bindings, not the input wrappers. Grouping it into the input family would hide its distinct setup role. It remains a separate setup-check.
3. Product-input tests exercise suite construction and production cache behavior together. Recipe and registry mutations are construction breaks under the explicit setup-check exception. Production cache mutations are separate and determine the input family's bounded sacred verdict.
4. The full package exhausted 90.025 binary seconds in its first sanitized archive build. It was not a red assertion baseline. Selecting all consumers would repeat cold native builds beyond the budget, so the matrix covers the requested members plus the adjacent cache-verification row. Kills outside that explicit set remain unknown. No package-wide or repo-wide uniqueness is claimed.
5. The recipe's explicit entry is redundant for this discovered dependency set, as the survivor's identical output demonstrates. Removing it cannot prove guard weakness here.
6. A constant version change is observable without breaking reuse. Surviving M5 must not be mistaken for a missing semantic correctness check.
7. Standalone empty-body probes initially left unused imports; these were removed and the repaired diffs passed apply-check and go vet. The production M1 diff uses a blank read binding with assignment, preserving the file-read error behavior without an unused content binding. Whole flag and toolchain loops were dropped, avoiding unused loop variables.
8. Initial switched build time was not instrumented. A later repeated build is measured separately and is not presented as the initial cold build duration.
9. A missing pathlib qualifier interrupted the first restoration script. Files were then restored explicitly and a second clean verification was run. The earlier run named restored-clean had still used selectors disabled, so only the second run is the final restoration evidence.
10. No scoped member has a corpus or SDK opt-in. Other corpus consumers were outside the bounded matrix; their opt-in modes were not measured or enabled. No skips occurred among matrix members.

All standalone diffs applied against the origin index and compiled with go vet using original-source overlays. Saved .go.fixture files prevent evidence copies from becoming accidental repository packages. Validation paths were updated after renaming to fixtures. No other packages' tests were run. The diagnostic programs only exercised the mutated functions to witness survivors.

Timing:
```json
{
  "setup": {
    "npm_seconds": 0.6680773869993573,
    "npm_exit": 0,
    "toolchain": "warm env verified; setup skipped",
    "nproc": 5
  },
  "build": {
    "repeated_switched_build_wall": 0.13601380999898538,
    "exit": 0,
    "initial_build_wall": null,
    "note": "Initial switch build passed but its wall time was not instrumented."
  },
  "baseline_binary": 90.025,
  "bounded_clean_binary": 7.531,
  "matrix_wall": 64.35905635000381,
  "matrix_binary": 43.055,
  "timing_runs_wall": 31.970487974002026,
  "recorded_at": "2026-10-09T09:20:30.408296+00:00"
}
```
