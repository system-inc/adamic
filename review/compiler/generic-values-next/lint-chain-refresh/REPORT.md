Built the refreshed step 16 delivery branch, retaining both generic-value guards and source-aware runtime build paths.
Source commit: e96fd5c3055d46b49b66b73a8dad5b0a51d6aaa3, parents 9bff3458 and f5236b48; this evidence commit changes no compiler or fixture source.
Checks: 16 generic-value oracles, complete admission proof, full lower shards, reader guard, counts and lane checks passed.
Mutants: wrong results, source identity, alias escape, address-only selection, widening refusal, JSON layout/shape and the real admission gate all caught their planted errors.
Uncovered: full gate and unrelated whole-package suites remain integration's responsibility.

compiler/generic-values-next merged exact origin/compiler/chain-lint-features
f5236b48ee11d680c5288e3074c2fec6ad26a09a. The merge applied without conflicts.
The SourceFlags export and all source-aware runtime build paths are retained,
as are the generic body/default/alias and higher-rank refusal guards. The object
and unknown widening refusal tests still share their helper with the address-only
Map/array mutants, in the same file with the dependency comments intact.
No main or area branch was changed.

The lower, IR, JavaScript, runtime layout and counts inputs are unchanged from
2a28375f to f5236b48. The generic-value oracle file is unchanged from 9bff3458
to the source merge. guard-input-equivalence.json records the tree/blob IDs.
The native library API gains SourceFlags from the chain; the runtime feature
fix is included in every newly built test binary and admission compiler.

Counts verification passed in 103.007 s without update-counts.
The table is byte-for-byte unchanged, so conditional regeneration was unnecessary.
All rows are attributed in counts-attribution.json: 1,320 unchanged from f5236b48
and five generic-value admissions; 1,325 total, zero unattributed.

Setup and environment:

export GOPROXY='https://proxy.golang.org|direct' ADAMIC_GOCACHE_OFF=1 GOMAXPROCS=4 GOFLAGS='-p=2';
timeout 600 bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh;
export TMPDIR=/workspace/generic-values-proof-tmp

Setup passed: Node 0.019 s, Go 0.023 s, markdown dependencies 0.069 s,
submodules 0.069 s, clang 0.152 s, build ready 30.712 s; total 30.981 s.
nproc 5; cpu.max 400000 100000. Node v24.19.0, Go go1.27.1, clang 20.1.8.
Setup ran on the merged working tree before its merge commit was written; its
printed HEAD is therefore 9bff3458. The admission proof and checks pin e96fd5c3.
Shared cache stays disabled; sourced GOFLAGS retain -buildvcs=false -trimpath.

Commands and results (output redirected to logs, never piped):

- timeout 1000 python3 review/compiler/generic-values-next/lint-chain-refresh/checks.py.
  Each needed test binary is compiled separately with go test -c and a 300 s
  outer build bound. Builds: oracle 6.647 s, lower
  11.636 s, IR 2.167 s, native
  7.372 s. Runtime commands run from their package
  directories with 85 s test and 90 s outer bounds, except the existing census.
- oracle: -test.run=^TestGenericValue -test.count=1 -test.timeout=85s -test.v.
  All sixteen focused tests pass in 1.943 s,
  held to Node in both backends. Individual leaf times and mutant outcomes
  are recorded in generic-values.log.gz.
- native: -test.run=^TestGenericValueLibraryIdentityViews$ -test.count=1
  -test.timeout=85s -test.v. Release and ASan/UBSan agree with Node;
  pass 0.263 s.
- IR: -test.run=^TestCallTargetReaders$ -test.count=1 -test.timeout=85s -test.v.
  Pass 0.746 s.
- lower: -test.list=^Test, then -test.run='^(SHARD_NAMES)$' -test.count=1
  -test.timeout=85s -test.v. All 354 top-level tests, 177/177, including every
  subtest. Inventory equals the previous delivery. Both shards pass in
  26.355 s and 49.407 s, under 90 s.
- counts: -test.run=^TestCountsAreRecorded$ -test.count=1 -test.timeout=600s
  -test.v, outer bound 650 s. Pass 103.007 s; no update-counts.

Complete admission proof:

python3 review/compiler/generic-values/admission_manifest.py /workspace/adamic HEAD;
timeout 600 /tmp/generic-values-admission-delta --base f5236b48 --head e96fd5c3
  --manifest review/compiler/generic-values-next/lint-chain-refresh/admission-manifest.json
  --manifest-generator cloud/admission-corpus/manifest.py
  --manifest-generator-revision 543925aa --json --workers 4
  --compile-timeout 45s --timeout 10s

The pinned official generator/tool is compiler/admission-delta 543925aa with the
reproducible step 16 fixture extension. All manifest blobs are checked against
committed head. Verdict pass; complete corpus true; 1,129 programs, 983 accepted
by both, 141 refused by both, five newly accepted, zero omitted or new refusals.
Budget zero checks all five newly admitted programs on Node, JavaScript and native;
every stdout, exit 0 and empty stderr agrees. Total 163.818 s.
admission-summary.json lists each admitted path and output; full observations
are in admission-delta.json.gz. The evidence-only commit preserves all compiler
inputs, fixture blobs and the proven admission set.

Mutants:

- timeout 450 python3 review/compiler/generic-values-next/lint-chain-refresh/mutants.py.
  The widening-guard deletion fails all four coupled tests with the exact
  "must stop before emission: <nil>" catcher, in 15.852 s including rebuild.
  Invalid JSON layout and nonempty shape fail the runtime identity test in
  17.390 s and 15.058 s including rebuild. The compiled probes return exit 1;
  these mutants are caught by the runtime checks, not compiler warnings.
  All three patches and logs are recorded, and each source is restored byte
  for byte in finally. Restored controls pass: oracle 5.002 s, native 0.693 s.
- Permanent focused tests catch comparer/default/utility wrong results,
  per-adapter source-identity mutations in both backends, alias escape into a
  higher-rank slot, and address-only Map/array selection.
- timeout 240 python3 review/compiler/generic-values-next/lint-chain-refresh/admission-mutant.py.
  The compiler paths were observed in this proof, recorded, and checked for
  their comparer admission/refusal. The actual admission tool uses the same
  base/head/manifest, --corpus-filter diff, supplied compiler/lowerer binaries,
  and the existing admission_wrong_result.py wrapper as head compiler.
  JavaScript's two comparer instances return false; Node and native stay unchanged.
  Verdict fail in 18.477 s; JavaScript stdout differs,
  all three exit 0 with empty stderr. This seven-program filtered mutant is
  separate from the complete passing proof.

Mandatory lanes after source commit, with the same environment:

git fetch -q origin main devtools/fast-gate cloud/merge-tree;
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -

Pass 38.2 s: gofmt/tools on 311 Go files, t.Parallel on 24 test packages,
a-check on 105 .a files, vet on 24 packages. No skip or failure.
The command runs again after the evidence commit and before push; final output
and delivery SHA are reported in the final response.
