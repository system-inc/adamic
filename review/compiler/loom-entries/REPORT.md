Imported the exact entries fixture bucket; compiler implementation already exists on main.
Commit: the single commit on compiler/loom-entries, based on 80640d7a8b8e456040758f4aa33791e45eceb4ce.
Checks: 12/12 acceptance variants pass; provenance, lower proof, registration, reader guard and vet pass.
Mutants: six source mutants caught by Node goldens; dropped compiler check caught by acceptance.
Incomplete: shared fixture registration is red; .a refusal prose conflicts with pinned checked-compilation expectations.

Task #tsdh6nb. This is a reviewable partial port, not a green submission. No PR or Loom submission.

## What main already has

The implementation from 05a60066 is already present through main commits be356537 and
8c962b67. entries_provenance.go and entries_records.go match the supplied tip exactly.
Main's subsequent edits are retained, including exact-shape handling for assign/freeze,
parallel test declarations, IR/backend changes and product-based count measurements.
No compiler source hunk was replaced. All 33 files from fixture commit 8d864f9c are
byte-identical, including historical evidence, observations and counts. No TypeScript
hunk changed. The acceptance skip line is untouched.

## Observed blockers

TestFixtureDirectoriesHaveTopLevelTests fails: fixture directory entries has no top-level
test. Main's directory runner requires testFixtureDirectory directly; simply adding that
registration would exercise historical status.json outcomes (2 Refused, 4 NotYet) against
current behavior (4 Compiles, 2 named Refused). The bucket's old refused index-signature
headers also describe closed gaps. These are historical observations, not failed ruled
expectations: all 12 explicit acceptance contracts pass.

The shared runner's CheckedStop representation constructs expected stdout as empty, while
03_hidden_other_type.a intentionally prints "before entries\n" before its check. Thus
refreshing status alone cannot express that checked stop in the shared runner. This
integration needs fixture-owner observation/header refresh and a runner change, or an
explicit ownership decision for the dedicated oracle acceptance test. Those shared tests
were not changed to make room for this bucket.

The brief additionally demands refusal in .a when provenance fails. The pinned .a
expectations demand checked compilation for 03, 04 and incompatible mutants of 02/04.
Conservative choice: preserve those explicit source contracts and current main behavior;
do not silently change the expectations or claim the prose refusal requirement is met.

## Exact checks

All commands source /workspace/adamic-tools/env.sh. Test output is redirected to the
corresponding log in this directory. No whole package suite or full gate ran.

- timeout 180 bash cloud/setup.sh: exit 0. submodules 0.062s, Go 0.164s, Node 0.218s,
  markdown 0.373s, clang 0.664s, shared cache 1.591s, build 49.830s, deferred binaries
  49.942s, warm cache 49.944s, done 49.974s. nproc=5; cgroup quota=4 CPUs.
  Go 1.27.1, Node 24.19.0, Linux clang 20.1.8.
- timeout 90 go test ./internal/oracle -run '^TestEntriesAcceptance$' -count=1
  -timeout 90s -parallel 4 -v: exit 0, 19.820s. All six baselines and six mutants pass;
  leaves range 0.19 to 18.81s. Source Node, native sanitizers, release and generated JS.
- timeout 90 go test ./internal/lower -run
  '^TestEntries(AllocationProof|RecordBoundaries)$|^TestNativeAgreementRowsNameRegisteredPrograms$'
  -timeout 90s -count=1 -parallel 4 -v: exit 0, 0.155s.
- timeout 90 go test ./internal/oracle -run '^TestEntries(Provenance|RuntimeReadiness)$'
  -timeout 90s -count=1 -parallel 4 -v: exit 0, 4.034s. Includes built-in unchecked
  mutants in both backends and dropped literal membership mutant.
- timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
  -count=1: exit 0, 16.461s.
- timeout 90 go vet ./internal/lower ./internal/oracle ./stage3/fixtures: exit 0.
- timeout 90 go test ./stage3/fixtures -run '^TestFixtureDirectoriesHaveTopLevelTests$'
  -timeout 90s -count=1: exit 1, 0.010s, missing entries top-level test.

No new Go test leaves or internal/oracle fixtures were added. Existing oracle count rows
were already carried by main, so no count regeneration was necessary. No package Go source
changed; compiler readers outside these focused checks were not rerun.

## Mutation evidence

source-mutants.log: all six exact edits from mutants.json run on Node (per process timeout
15s, total timeout 60s); their complete observations match the stored mutant observations
and differ from the baseline goldens. Scanner removes await; hidden-number becomes string;
hidden-string becomes number; alias adds string; getter becomes data; symbol becomes string.

compiler-mutant.log and drop-enumeration-check.log: temporarily replace the single
call.Checked assignment with false. TestEntriesAcceptance/03_hidden_other_type.a/baseline
fails in 2.91s on checked=false and missing panic (native exits 0). No compilation failure.
Source restored in finally; patch evidence has .patch extension. Mutant command had 60s
Go timeout and 75s subprocess deadline, under a 90s outer deadline.

The auxiliary bucket check.py acceptance attempt was stopped because its go-run builds
overlapped the temporary compiler mutant. It is not claimed as evidence; the earlier
complete Go acceptance run finished before mutation. No new test was rerun after exceeding
60s, and no new test leaf was added. Historical fixture README/evidence paths are retained
verbatim as requested, not introduced into compiler products.
