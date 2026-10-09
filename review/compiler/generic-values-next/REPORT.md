Built the final step 16 delivery branch, coupling address-only mutants to widening refusals and merging the requested chain.
Source commit: fdac37a7485a10d5ef7dd8d0bf5e647761b8e901, parents 40a93702 and c81ba261; this evidence commit changes no compiler or fixture source.
Checks: all generic-value oracles, complete admission proof, both full-lower shards, reader guard, JSON identity and lanes passed.
Mutants: disabled widening guard fails all four coupled tests; address-only and identity mutations are caught; admission wrong-result mutant fails the real gate.
Uncovered: no full gate or unrelated whole-package suite; integration owns the final gate.

compiler/generic-values-next was cut from runtime-cleared 40a93702 and merged
exact origin/compiler/optional-presence-next c81ba26193ce58fe7eb8c15d3ea1e46a58052fef.
The merge had no conflicts. The generic body/default/alias and higher-rank
refusal guards remain, alongside optional-presence readiness/storage guards.
The merge commit has both requested parents; no main or area branch was changed.

The object and unknown widening tests live next to both address-only mutants in
internal/oracle/generic_values_test.go. Both mutants invoke the same named
helper genericValueWideningRefused before testing their emitted-code mutation.
Comments explain that address-only selection is sound because a closure cannot
be widened into those slots, and that admitting such widening requires revisiting
the selection. Node accepts each planted widening program and prints 2; lowering
returns the exact dynamic-function-descriptor NotYet and no IR program.

The new test leaves passed in 0.53 s (object widening), 0.54 s (unknown widening),
2.11 s (Map mutant) and 2.73 s (array mutant) on their initial green run.
Disabling only the lowerer's closure widening guard admits both planted programs
and makes all four coupled tests fail with "must stop before emission: <nil>".
The guard was restored, and all sixteen generic-value tests passed again.
The Map mutant substitutes address-only constructors; the array mutant
substitutes address-only equality. Both compile under ASan/UBSan and terminate
cleanly, but differ from Node's stdout. The existing per-adapter source-identity
mutant is caught in native and JavaScript, as are all result/default mutants.
The alias escape mutant still produces the ruled higher-rank refusal.

Counts were regenerated exactly once on the merged compiler. All 1,325 rows
are attributed in counts-attribution.json: 1,320 equal c81ba261, five equal the
cleared generic-values admissions, no removals, no unexplained numeric changes.
Those five are comparer, indexed default, utility default, immutable alias and
the existing generics/04_function_value.a. Inherited optional-presence and main
count changes are retained. Counts regeneration passed in 154.930 s.

The complete admission-delta proof uses c81ba261 as base and fdac37a7 as head.
The pinned official generator/tool is compiler/admission-delta 543925aa, plus
the same reproducible step 16 manifest extension used by the prior unit.
The manifest and all program blobs are verified against head. Classification
is lowering-only, four workers, 45 s compile and 10 s runtime command deadlines.
Budget zero checks every newly admitted program; complete_corpus is true.

Results: verdict pass, 1,129 programs; 983 accepted by both, 141 refused by both,
five newly admitted, zero omitted or new refusals. All five admissions agree on
source Node, JavaScript and native stdout, exit 0 and empty stderr. The seven
revision-diff programs are included, including both runtime measurement probes.
Full raw observations are in admission-delta.json.gz; admission-summary.json
lists every admitted program and its common stdout. Total proof time 106.021 s.

The planted emitted-JavaScript comparer returns false for both instances;
native and source Node remain unchanged. The actual admission tool, using the
same base/head and verified lowerer binaries, reports verdict fail in 7.439 s.
Comparer JavaScript stdout differs from Node/native; every execution exits 0.
This mutant run filters to the seven-program diff and is not the complete proof.
The reusable mutation wrapper is review/compiler/generic-values/admission_wrong_result.py;
mutant-compilers.json records the checked real compiler paths.

Commands (every test's output was redirected to a log, with an outer hard bound):

- export GOPROXY='https://proxy.golang.org|direct'; timeout 180 bash cloud/setup.sh;
  source /workspace/adamic-tools/env.sh. nproc 5, cpu.max 400000 100000.
  Setup: Node 0.024 s, Go 0.026 s, markdown dependencies 0.084 s,
  submodules 0.088 s, clang 0.204 s, go build 67.784 s; total 68.003 s.
- go test ./internal/oracle -run '^TestGenericValue' -count=1 -timeout 90s -v:
  all sixteen tests pass; initial green 5.522 s, restored final 1.302 s.
- go test ./internal/native -run '^TestGenericValueLibraryIdentityViews$'
  -count=1 -timeout 85s -v: pass, 0.939 s.
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 85s:
  pass, 45.837 s.
- go test ./internal/lower -list '^Test' -run '^$' -timeout 85s:
  354 top-level tests inventoried, partitioned exhaustively in lower-shards.json.
  Each shard runs go test ./internal/lower -run '^(SHARD_NAMES)$' -count=1
  -timeout 85s -v, with a 90 s outer command deadline. Both 177-test shards pass,
  48.860 s and 34.516 s, including every original top-level test's subtests.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1
  -timeout 210s -args -update-counts: one regeneration, pass, 154.930 s.
- python3 review/compiler/generic-values/admission_manifest.py /workspace/adamic HEAD.
- timeout 600 /tmp/generic-values-admission-delta --base c81ba261 --head fdac37a7
  --manifest review/compiler/generic-values-next/admission-manifest.json
  --manifest-generator cloud/admission-corpus/manifest.py
  --manifest-generator-revision 543925aa --json --workers 4
  --compile-timeout 45s --timeout 10s: complete pass.
- Admission mutant: same flags/revisions plus --corpus-filter diff, supplied base
  compiler/lowerer and actual head lowerer, and admission_wrong_result.py as
  --head-binary; GENERIC_VALUES_HEAD_COMPILER selects the real head compiler.
- git fetch -q origin main devtools/fast-gate cloud/merge-tree;
  git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -:
  pass after source commit, 22.4 s; gofmt/tools on 302 Go files, t.Parallel on
  23 test packages, a-check on 104 .a files, vet on 23 packages.
  The mandatory check runs again after the evidence commit, before push.

The complete proof pins the merged source commit intentionally: its subsequent
evidence commit adds only review metadata and logs, preserving compiler bytes,
fixture blobs and the proven admission set. The branch is prepared for integration
after the requested chain; it has not been merged into main.
