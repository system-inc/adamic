Built static-constructor data spreads, throwing accessor spreads, and nullish templates on main alone.
Commits: c50eaa166 (#vk7ed2m), 926175191 (#q9hz1bf accessors), fa57d7bd8 (independent template slice).
Proof: 2,171 fixture programs, six new admissions held to Node, no regressions or unresolved timeouts; 277 lower tests pass.
Mutants: all 12 expected failures caught; trailing-slot and owned-return instrumentation passes.
Deferred: #dv99xzy needs catchable lexical ReferenceError support; the null-call MakeError half needs its absent library helper. Full gate and WASI were not run.

| Task | Commit or skipped | What it needs / proof |
| --- | --- | --- |
| #dv99xzy | skipped | 71972e180 initialization-exception slice: internal/native/readiness_error.go:6, internal/lower/readiness_exceptions.go:7, and ReferenceError recognition in internal/lower/class_inheritance.go:20. Main refuses the ReferenceError catch. Using an Error catch demonstrates Node exit 0 versus native panic exit 70. |
| #vk7ed2m | c50eaa166 | No chain member. Two fresh-data fixtures agree in both backends and release builds under ASan, UBSan and LSan; broad-static-guard revert is caught. Two measured count rows included. |
| #q9hz1bf accessor slice | 926175191 | No chain member. Uses main adamic_thrown and its copy_checked path. First, middle and last failures agree; partial-copy ownership mutants fail. Four measured count rows included. |
| #q9hz1bf expression.go slice | fa57d7bd8 | No chain member. Main already has inline primitive templates; nullish spelling preserves effects. template_nullish.a was red without it. Revert, wrong spelling and dropped effects are caught. Its count row is included. |
| #q9hz1bf stringNullPrototypeCall MakeError half | skipped | Unlanded library conversion member f2ffe42d8, internal/lower/library_string.go:620 (helper), :642 (old ObjectLiteral throw). This helper does not exist on main or the requested chain head. Its isolated proposal edit has no trunk function to edit; do not import the library stack. |

Base is origin/main 5e33a17b186a8a2218d27b69b21e2de5acc5b750. There are no chain commits on the delivery branch. Earlier unpublished chain and trunk checkpoints remain local; their logs are superseded by main-* and final-* evidence.

Toward the brief's cohere step 08 port, the static-factory spread and accessor exception boundary are admitted on trunk. Recursive closure admission is deferred because its required initialization-time observation fails on trunk. No language decision was needed.

Observed admission census: 1555 admitted on main, 1561 at tip, exactly six additions, zero regressions, zero unresolved timeouts. Corpus is every tracked or new .a/.ts program under testdata, fixtures and gaps, plus dedication and examples, excluding declarations, review and cohere. Imported production modules were not part of this declared fixture census; an initial broader scan was stopped and superseded. All seven fixtures in this unit pass uncached Node source, JavaScript backend, native ASan/UBSan, release and leak checks. plain_data_spread.a is a baseline-admitted control, so it is not an admission delta. See admission-corpus.json, admission-before.json, admission-after.json, admission-new.json and final-uncached-oracle.jsonl.

Plain-data C is byte-identical, checked by cmp; SHA256 cd1a775183200bd2c65a8e95fa554966ad4774982ef99b9b4f98bec646a2487b. The IR effect test also checks fresh and saved data, saved accessor storage, unknown storage and incoming parameters; the over-broad-edge mutant fails it.

Commands actually run on trunk:

- GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. Initial setup failed during concurrent local editing with constantInitializerClosure undefined; the stable retry succeeded. Timing lines: Go 0.018 s, Node 0.024 s, submodules 0.069 s, markdown 0.074 s, clang 0.175 s, shared cache 9.262 s, build cache 59.587 s, done 59.612 s. nproc=5, cpu.max=400000/100000. setup.log and setup-retry.log preserve both attempts.
- npm ci --ignore-scripts --prefix stage3/api: added three packages, restoring pinned @types/node 25.3.3 missing from setup. Initial lower shards had setup failures; all affected shards were rerun. No lockfile changed.
- go build -o /tmp/adamic-gaps3-main ./cmd/adamic at pristine main; go build -o /tmp/adamic-gaps3-tip ./cmd/adamic at the validated source tip: exit 0. Final source matches the validated checkpoint (source-equivalence.log).
- go test ./internal/ir -run '^TestObjectSpreadAccessorEffects$' -count=1 -timeout=90s -v: pass.
- go test ./internal/lower with 12 recorded top-level regex shards, -count=1 -timeout=90s -json: all 276 existing leaves pass after the obsolete throwing-spread refusal was updated. TestClassFeaturesThrowingAccessorSpread is a separate new positive leaf, pass 0.17 s. TestClassFeaturesAccessorRefusals pass 0.13 s. lower-shards.json records exact selections; lower-shard-*.jsonl records results. The final shard 7 rerun replaces its earlier obsolete expectation failure.
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout=90s -v: pass 2.45 s in call-target-readers-final.log.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=90s -args -update-counts: pass 66.461 s, run once. Only seven added rows; no existing row moved. counts-attribution.patch records measured rows. Rows were distributed into their owning commits so each task is independently removable.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(accessor_spread_throw_(first|middle|last)|static_constructor_spread_(gap|owned)|plain_data_spread|template_nullish)\.a$' -count=1 -timeout=90s -json: all seven leaves pass. Exact leaf seconds are in verification-summary.json; first cold trunk runs were 11.52/11.53 s for static fixtures, 19.19 s for each accessor failure, 11.21 s for plain data and 0.57 s for nullish templates, all below 60 s.
- go vet ./internal/lower ./internal/ir ./internal/flow ./internal/native ./internal/oracle: exit 0, final-vet.log.
- python3 admission-delta.py before /tmp/adamic-gaps3-main, then after /tmp/adamic-gaps3-tip; per-program hard limits, complete manifests recorded. Seven initial baseline and eight tip import-heavy timeouts were resolved with bounded retries. Both sides' 62 missing-Node-types observations were rerun after setup repair. Final repair: six additions, no regressions, no unresolved timeouts.
- git fetch -q origin main devtools/fast-gate cloud/merge-tree; git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -: preliminary pass; final committed-tip output is lane-final.log.

Mutant observations, never compiler errors counted as successful kills:

| Mutant | Catcher |
| --- | --- |
| main-static-revert | Both static fixtures refuse the fresh spread |
| main-accessor-revert | First/middle/last fixtures refuse the newly admitted spread |
| main-accessor-release | LSan detects leaked partial copies |
| main-accessor-edge | UBSan detects use of the null failed copy |
| main-accessor-broad-edge | IR effect test rejects a throw edge on proven data |
| main-accessor-getter-retain | LSan detects the extra owned getter reference |
| main-accessor-return-double-release | ASan detects heap use after free |
| main-accessor-slots-revert | Sanitizers catch release through uninitialized slots |
| main-template-revert | Nullish template fixture is refused |
| main-template-spelling | Node comparison detects different stdout |
| main-template-effects | Node comparison detects missing evaluations |
| accessor-leaf-revert | New positive lower leaf fails its admission check |

The non-mutant main-accessor-slots-owned-return probe checks every trailing reference slot before release and supplies an owned getter return while an error is pending; all three field positions pass ASan and LSan. The matching double-release probe proves that reference must be released exactly once. Production sources are restored after every probe. Sources under review are .patch or .go.txt, never compilable .go.

Runtime clearance for @system_adamic_runtime: internal/native/runtime/object.c lines 31, 32, 33 and 34 add pending-error detection, release of the zero-initialized partial copy, and return NULL. The existing slot zeroing at line 16 is exercised and reverted in the slot mutant. No other object.c lines change. These lines are listed for owner clearance; no merge or PR was opened.
