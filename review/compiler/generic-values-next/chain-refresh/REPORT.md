Built the refreshed step 16 delivery branch by merging the requested chain head; both sides' guards remain intact.
Source commit: 6f800def962b4e404b3ee794e5b25a199b575880, parents c905de3a and 2a28375f; the evidence commit changes no compiler or fixture source.
Checks: 16 focused Node oracles, complete admission proof, full lower in two bounded shards, reader guard, counts and lane checks all passed.
Mutants: result, identity, alias escape, address-only, widening refusal, JSON layout/shape and admission wrong-result catchers all succeeded.
Uncovered: no full gate or unrelated whole-package suite; integration owns the final gate.

Delivery branch: compiler/generic-values-next. Exact merged chain head:
2a28375f9c8c487d29abb89f38489858fc9d5c92, containing c81ba261 and main 88adf555.
The merge applied without conflicts. No main or area branch was changed.
The chain refresh changes build-cache/test-audit inputs, but its lower, native,
JavaScript, IR and counts inputs equal c81ba261 byte for byte; tree/blob IDs are
recorded in guard-input-equivalence.json. The coupled function-to-object/unknown
refusal and both address-only mutants remain together in generic_values_test.go,
with their dependency comments and shared helper intact.

Counts verification passed in 90.580 s, without update-counts. The table did not
move, so the requested conditional regeneration was unnecessary. Every row is
attributed in counts-attribution.json: 1,320 inherited unchanged from 2a28375f,
five generic-value admissions; 1,325 total, zero unattributed rows.

Verification commands, all with output redirected to individual logs:

- export GOPROXY='https://proxy.golang.org|direct'; timeout 180 bash cloud/setup.sh.
  The shared-cache attempt reached dependency enumeration and exhausted its
  hard deadline. Its partial timings are in setup.log.gz. A local-cache retry
  reported "cat: write error: No space left on device". Three abandoned Go
  temporary build directories were removed; no repository sources were removed.
- export ADAMIC_GOCACHE_OFF=1 GOMAXPROCS=4 GOFLAGS='-p=2'; timeout 600 bash cloud/setup.sh.
  Successful setup: Node 0.026 s, Go 0.026 s, markdown dependencies 0.076 s,
  submodules 0.075 s, clang 0.155 s, build ready 59.471 s, total 59.707 s.
  nproc 5; cpu.max 400000 100000. Node v24.19.0, Go go1.27.1, clang 20.1.8.
- source /workspace/adamic-tools/env.sh; export TMPDIR=/workspace/generic-values-proof-tmp.
  The env file retains canonical -buildvcs=false -trimpath; shared cache stays off.
- timeout 1000 python3 review/compiler/generic-values-next/chain-refresh/checks.py.
  This compiles only the four needed test binaries with separate 300 s build
  bounds, then executes each from its package directory. Build times: oracle
  8.200 s, lower 8.160 s, IR 12.081 s, native 7.748 s. Runtime checks below have
  85 s test and 90 s outer bounds, except the existing counts census.
- oracle: -test.run=^TestGenericValue -test.count=1 -test.timeout=85s -test.v.
  All sixteen tests passed in 2.083 s, including both backends held to Node,
  per-adapter identity and wrong-result mutants, alias escape and address-only
  Map/array mutants. Individual leaf timings are in generic-values.log.gz.
- native: -test.run=^TestGenericValueLibraryIdentityViews$ -test.count=1
  -test.timeout=85s -test.v. Release and ASan/UBSan views agree with Node;
  pass 0.295 s.
- IR: -test.run=^TestCallTargetReaders$ -test.count=1 -test.timeout=85s -test.v.
  Pass 19.013 s.
- lower: -test.list=^Test, then -test.run='^(SHARD_NAMES)$' -test.count=1
  -test.timeout=85s -test.v. All 354 top-level tests partitioned exhaustively
  as 177/177 in lower-shards.json, including their subtests. Both shards pass,
  30.096 s and 47.396 s, each under its 90 s outer bound.
- counts: -test.run=^TestCountsAreRecorded$ -test.count=1 -test.timeout=600s
  -test.v, outer bound 650 s. Pass 90.580 s, no regeneration or changed rows.

Complete admission-delta proof against the requested base:

- python3 review/compiler/generic-values/admission_manifest.py /workspace/adamic HEAD.
- timeout 600 /tmp/generic-values-admission-delta --base 2a28375f --head 6f800def
  --manifest review/compiler/generic-values-next/chain-refresh/admission-manifest.json
  --manifest-generator cloud/admission-corpus/manifest.py
  --manifest-generator-revision 543925aa --json --workers 4
  --compile-timeout 45s --timeout 10s.

The pinned official generator/tool is compiler/admission-delta 543925aa, with
its reproducible step 16 manifest extension. All manifest blobs are verified
against the committed head. Complete corpus true; budget zero; 1,129 programs,
983 accepted by both, 141 refused by both, five newly accepted, zero omitted or
new refusals. All five agree on Node, JavaScript and native stdout, exit 0 and
empty stderr. Pass 143.098 s. Every admitted path and stdout is recorded in
admission-summary.json; complete observations are in admission-delta.json.gz.
The proof pins the source merge intentionally; this evidence-only commit
preserves all compiler inputs, fixture blobs and the proven admission set.

Mutants on the merged head:

- timeout 450 python3 review/compiler/generic-values-next/chain-refresh/mutants.py.
  Each temporary mutation is restored in finally, byte for byte. Disabling
  the closure widening refusal makes all four coupled tests fail with
  "must stop before emission: <nil>" (31.689 s including Go rebuild).
  Replacing JSON's object with a closure and changing its empty shape count to
  one each fail TestGenericValueLibraryIdentityViews (26.692 s / 18.601 s).
  Patches, expected failures and individual leaf times are recorded alongside
  mutants-results.json. Restored focused tests pass: oracle 2.876 s,
  native identity 0.413 s. No delivery source mutation remains.
- Existing permanent mutant tests prove wrong comparer/default/utility results,
  wrong source identity in both backends, alias escape into a higher-rank slot,
  and address-only Map/array selection are caught on this head.
- Actual admission tool with the same base/head and manifest, --corpus-filter diff,
  verified real base/head lowerer binaries, and admission_wrong_result.py supplied
  as --head-binary. The wrapper changes only emitted JavaScript's two concrete
  comparer instances to return false. Verdict fail in 7.801 s; JavaScript stdout
  differs while native and Node agree, all exit 0 with empty stderr.
  mutant-compilers.json records the actual cached compiler paths. This seven-
  program filtered mutant is separate from the complete, passing proof.

Lane checks after source commit:

export ADAMIC_GOCACHE_OFF=1 GOMAXPROCS=4 GOFLAGS='-p=2'; source /workspace/adamic-tools/env.sh;
git fetch -q origin main devtools/fast-gate cloud/merge-tree;
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -

Pass 16.9 s: gofmt/tools on 302 Go files, t.Parallel on 23 test packages,
a-check on 105 .a files, vet on 23 packages. No skip or failure.
The mandatory lane command runs again after the evidence commit, before push;
its output is reported with the final delivery SHA.
