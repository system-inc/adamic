Case-level sharding replaces the 43-second watcher file task with 16 exact upstream grep units.<br>
The gate's exact version-1 manifest contains 151 commands, each bound to its inputs hash.<br>
All 151 cold commands passed; the longest took 24.650 seconds on four CPUs.<br>
The original expected counts, API sanction and step12 failure reporting remain intact.<br>
Developer tools owns gate wiring; no full Adamic gate or macOS confirmation was run.

This follows `a4afcd66` on `codex/step34-lane-shards`, which already merged the
step12 diagnostics as `12963b97`. No compiler, adapter, expected.json,
sanctioned-api.json or external gate file is changed. The integration location
is `cloud/fast-gate/run.py stage3()` on `devtools/fast-gate`; Kirk explicitly
assigned that code to developer tools.

[SHARD_CONTRACT.md](SHARD_CONTRACT.md) describes artifact preparation, the exact
manifest schema and the separate post-unit merge. The command argv carries each
unit's inputs hash, so shards.json has only the requested fields. The prepared
artifact key is `e41a430a94a8b7b011d31cd1c9aaabfda03d4f8c9dc3f67d13eb8859aa434b24`.
Its bundled stock API parser has SHA-256
`1512f1871eebe448dac94546c5a6337daa1cb64ff20c80f3d9e93f85a0f48dce`.
Each command rehashes the whole artifact and the stock parser. Test results and
Node's compile cache are never reused.

The 134 file shards keep the existing deterministic longest-first assignment
and its measured weights. All root unit suites from one source file remain
together. The watcher file is the authorized exception: each of its 16 uniquely
named cases gets one exact, anchored upstream `--tests` grep, including all its
describe names. Original setup and cleanup hooks, waits, assertion behavior and
upstream timeouts remain intact. No case is skipped.

`stage3-lane-byte-comparisons` owns the complete publicApi.ts file: 30 passing and
one failing test. It checks the exact reference/local API bytes and applies the
unchanged stock TypeScript 6.0.3 declaration sanction. Other upstream baseline
bytes, including JavaScript, are compared in their assigned file shards. The
adapter-specific JavaScript equivalence proofs still execute in the unchanged
apply artifact producer, rather than being silently omitted or relaxed.

The test-only entry calls upstream's own `runConsoleTests` function. A real
probe found that `npm test --built` still builds its dependencies, so that is not
the gate command. The probe changed a cached scratch build and the next payload
check correctly refused reuse. That scratch artifact was preserved separately;
a clean artifact was produced in 134.318 seconds, outside the test-unit tier.
The original 43.102-second whole-file observation remains in SHARDS_REPORT.md.

A first cold batch also found a genuine harness mismatch in file shard 060:
its native IPC driver did not inherit upstream's `NODE_ENV=development` setup.
This disables normal TypeScript assertions and changes the tsserver project
removal baseline. The original guard rejected it with full Mocha text and diff.
The driver now supplies the same development environment as upstream; a focused
fixture and a dropped-environment mutant hold it. The actual corrected shard
passed in 13.086 seconds with 868 passing, zero failing and zero pending. The
pre-fix partial batch was discarded; every manifest command was restarted with
new bound inputs hashes and fresh results.

This is the same Codex Linux x64 box with Node v24.19.0, nproc 5 and cgroup
`cpu.max=400000 100000`, hence four CPUs. The final measurement runs four
manifest commands concurrently, each with a distinct STAGE3_CACHE. Cold means
fresh processes, `NODE_DISABLE_COMPILE_CACHE=1`, inputs verified by hash, and no
prior result reuse. The OS page cache is retained. Reported command walls
include Python startup, input/tool probes, complete artifact verification,
worker startup, execution, comparisons and report publication. These are
observations on this box, not predictions about other machines.

The same-artifact **unsharded** reference is the real
`stage3/lane/run.sh RESULTS --artifact ARTIFACT` command. It retains all runners,
eight upstream workers, no filter and the original full lane checker. It passed
with 106366 passing, one sanctioned failure and zero pending: 300.248 seconds
for upstream tests, 300.334 seconds for the oracle and 307.257 seconds for the
lane including artifact verification. This is a reference measurement, not a
manifest unit. Install/build exit evidence is explicitly attributed to the
artifact producer and not claimed to have run again.

The complete listing in `evidence/case-shards/tests-and-units.jsonl.gz` assigns
106367 observed results once. The merge compares actual title identities with
multiplicity, not just totals, and rechecks exact failure titles, baseline paths
and complete diff bytes. It requires the same artifact identity and runs the
unchanged full lane checker against the real reference. Missing or swapped
tests cannot hide behind an equal count.

The final cold batch passed all 151 commands. The longest was
`stage3-lane-file-007`, 24.649585 seconds (653 passing, zero failing and zero
pending). The longest watcher selector took 16.141195 seconds. No individual
watcher case exceeded 30 seconds. The byte-comparison unit took 12.668878
seconds, with 30 passing and the sole sanctioned failure. The local four-way
measurement coordinator took 719.867632 seconds; it is not a test unit or a
measurement of the full gate.

The normal merge passed with exactly 106366 passing, one failing and zero
pending, matching the real unsharded reference and all 106367 actual test
identities. See [cold-ledger.json](evidence/case-shards/cold-ledger.json),
[merged.json](evidence/case-shards/merged.json) and the complete raw per-unit
observations in [cold-units.tar.gz](evidence/case-shards/cold-units.tar.gz).

The planted mutation changes only the private prepared upstream runner. After
watcher cleanup it asserts that the polling file still contains 10 although the
case changed it to 100. All 16 watcher selectors ran against that runner; only
`stage3-lane-symlink-case-03` failed, in 11.766099 seconds. The other 15 passed;
the longest mutated selector took 15.623775 seconds. The captured original
Mocha event names `unittests:: sys:: symlinkWatching:: with ts.sys:: watchFile
using polling watchFile using polling`. Its assertion message and stack are
preserved. This assertion has no baseline change, so its baseline diff is empty.
The inherited step12 fallback title parser also emits a shortened diagnostic
for this nested `::` title; there is exactly one captured Mocha failure event
and one counted failed test.

The mutant merge uses those 16 executed results together with the 135 unchanged
executed file/API results. Its full identity union still matches the reference,
but 106365 passing and two failures reject the verdict, including exactly the
planted watcher failure. This does not claim the other 135 commands were rerun
with the watcher-only mutation. See [proof.json](evidence/case-shards/proof.json),
[merged-mutant.json](evidence/case-shards/merged-mutant.json) and
[watcher-mutant.diff](evidence/case-shards/watcher-mutant.diff). The mutated
runner SHA-256 is
`d6c18a5abbfe1563019a24770bef4aa39758c42054b77f03cdb58643efde0f41`.

The final artifact cache verification passed in 7.697 seconds with a cache hit,
the same artifact key and stock-parser digest. The shared build remained
unchanged through all normal units and private mutant views.

Every test/measurement command writes stdout and stderr to a log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step34-followup-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export STAGE3_ARTIFACT=/workspace/scratch/step34-artifacts/e41a430a94a8b7b011d31cd1c9aaabfda03d4f8c9dc3f67d13eb8859aa434b24
export PYTHONDONTWRITEBYTECODE=1
python3 stage3/lane/prepare_shards.py /workspace/scratch/step34-artifacts > /tmp/step34-shared-repreparation.log 2>&1
python3 stage3/lane/shard_plan.py "$STAGE3_ARTIFACT" > /tmp/step34-manifest-generation.log 2>&1
bash stage3/lane/run.sh /workspace/scratch/step34-followup/whole --artifact "$STAGE3_ARTIFACT" > /tmp/step34-same-artifact-whole.log 2>&1
python3 stage3/lane/measure_shards.py /workspace/scratch/step34-followup/cold-verified --concurrency 4 > /tmp/step34-cold-verified.log 2>&1
python3 stage3/lane/merge_shards.py /workspace/scratch/step34-followup/cold-verified --whole /workspace/scratch/step34-followup/whole --output /workspace/scratch/step34-followup/merged.json > /tmp/step34-merge.log 2>&1
python3 stage3/lane/measure_shards.py /workspace/scratch/step34-followup/mutant-cases --concurrency 4 --only-prefix stage3-lane-symlink-case- --mutant-runner /workspace/scratch/step34-followup/watcher-mutant-run.js > /tmp/step34-mutant-cases.log 2>&1
python3 stage3/lane/merge_shards.py /workspace/scratch/step34-followup/mutant-union --whole /workspace/scratch/step34-followup/whole --output /workspace/scratch/step34-followup/merged-mutant.json > /tmp/step34-mutant-merge.log 2>&1
python3 stage3/lane/prepare_shards.py /workspace/scratch/step34-artifacts > /tmp/step34-final-cache.log 2>&1
# Both mutant commands above exit 1 as required.
# Each fixture file is its own unittest discover command, preserving a separate log and process wall.
python3 -m unittest discover -s stage3/lane -p test_worker_environment.py -v > /tmp/step34-worker-environment.log 2>&1
python3 stage3/lane/prove_artifact.py > /tmp/step34-followup-artifact-mutants.log 2>&1
python3 stage3/lane/prove_plan.py > /tmp/step34-followup-plan-mutants.log 2>&1
python3 stage3/lane/prove_shards.py > /tmp/step34-followup-shard-mutants.log 2>&1
python3 stage3/lane/prove_case_observer.py > /tmp/step34-case-observer-mutant.log 2>&1
python3 stage3/lane/prove_worker_environment.py > /tmp/step34-worker-environment-mutant.log 2>&1
```

The required setup passed: Node ready 0.026 s, Go ready 0.029 s, submodules ready
0.082 s, Markdown dependencies ready 0.083 s (verified-lock skip step 0.013 s),
clang ready 0.187 s, Go build ready 41.413 s, test binaries deferred 41.515 s,
build cache warm 41.517 s and done 41.547 s. nproc is 5; the four-CPU quota is
reported above. The environment file is `/workspace/adamic-tools/env.sh`.

Review strengthened the merge to compare raw diff bytes, including line endings.
A dedicated newline-tampering mutant proves that check independently. All units
were measured again after the resulting runtime inputs hashes changed.

The earlier combined 62-fixture invocation passed but took 33.750 seconds
alongside the unsharded suite. It is not a valid 30-second unit. Final fixture
verification runs by test file: all 68 fixtures passed; the longest file was
`test_check.py` at 9.329388 seconds. The 12 shard/shared-artifact fixtures passed
in 1.066790 seconds. No new Adamic fixture is added; lane counts.md is refreshed
to 68. The original combined attempt and final per-file logs and timing ledger
are retained in evidence/case-shards/fixtures/.

| New guard mutant actually run | Isolated catch |
|---|---|
| Remove the anchors from the watcher grep | extra suffixed case matches and the exact-selection assertion fails |
| Ignore the unit's 30-second budget | over-budget observation with every other predicate green is accepted, failing the fixture |
| Ignore selected case identity | a substituted case with unchanged counts escapes and fails the fixture |
| Ignore the unit inputs hash | changed harness bytes are accepted, failing the required rejection |
| Ignore stock parser corruption | changed API parser bytes are accepted, failing the required rejection |
| Ignore a failed producer phase | build exit 7 is accepted, failing the required rejection |
| Ignore changed ready input metadata | inconsistent input identity is accepted, failing the required rejection |
| Ignore prepared compiler payload corruption | modified compiler bytes are accepted, failing the required rejection |
| Ignore exact raw diff bytes | newline-only tampering is accepted despite every other check being green, failing the fixture |
| Ignore the executed union comparison | a swapped test with identical counts escapes; the isolated union fixture fails |
| Drop a selected-case Mocha event | original event observation is missing and the Node fixture fails |
| Drop NODE_ENV=development from native workers | the original assertion environment is absent and the Node fixture fails |

All five existing artifact guard mutants and three file-plan guard mutants were
rerun and caught. The existing dropped-IPC and dropped-error-text mutants remain
covered by the per-file fixtures. Their individual catches, plus the two earlier
real source mutants, are documented in SHARDS_REPORT.md.

Not covered: the full Adamic gate, wiring in the developer-tools branch, macOS
case counts or a universal timing bound for arbitrary machines. The shared
artifact must be produced/restored before manifest units start. The full gate's
under-five-minute wall is owned by its integration and scheduling; this branch
reports complete per-command measurements without claiming a full-gate run.
