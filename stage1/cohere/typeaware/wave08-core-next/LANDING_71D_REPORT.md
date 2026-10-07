Built: wave 08 rebased onto current main 71d7e491b, retaining lint area bb2ece564; owned profiles re-green.
Commits: tested rebased tip aaea3b0104495d77cba9e5a48c4bd601c51dda8c; this report accompanies its evidence commit.
Checks: ten completed profiles, two parked kernels, core 264/264 and Globals 111 typed programs pass normally and sanitized.
Mutants: all ten completed-rule mutants caught by Go bytes; released handles and the filtered Node oracle pass.
Uncovered: shared typed registration, two HIR rules, fresh quiet timing and the full required-input gate.

Rebased with --rebase-merges from published
51c7f65267d350a1fd93ee5c9bf56d146224332a onto origin/main
71d7e491b3c9724f7a0e2ee754592149e7f9790b. The lint-area tip
bb2ece564842c4b2f909b9f75c27e74c2efa4f29 remains an ancestor; its subsequent
merge command reports already current. Main's delta contains stage-3 adaptation
and measurement work. A git diff from the prior published tip shows no changed
files under cmd, internal, bridge or stage1/cohere. Accordingly the unchanged
stage0, core native binaries and checker archives from the previous validation
were reused. Original, process, race, Globals and mutant validators rebuild
their own executables. No shared source was edited or branch pushed outside
codex/typeaware-wave-08.

Only archived reproducible ELF/ar binaries larger than one megabyte under
/workspace/wave08-d3a were removed for scratch space: 29 artifacts,
1,169,033,642 bytes. All input sources and committed compressed evidence remain.
New test output went to files under /workspace/wave08-71d, never through a
running pipe. The configured environment was sourced for each command chain.

Commands and observed results:

- go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m,
  with original/fact artifact roots under the new scratch directory,
  ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus and both frozen
  /workspace/wave08-corpora manifests: PASS, 303.198 seconds. Controls,
  compiler 77 and repository 287 match Go normally and sanitized. Loop,
  require-import and cycle mutants compile, exit zero with empty stderr, and
  differ only under Go byte comparison at bytes 50, 7373 and 12133. Released
  handle refuses with exit 70; retaining-registry mutant violates the expected
  refusal. Declaration flags, type-syntax and resolved-module mutants fail
  direct checker comparisons.
- compare_core.py using /workspace/wave08-core/upstream and compiler root,
  normal and sanitized core binaries from /workspace/wave08-4e0:
  264/264 complete finding/fix/suggestion streams match in each run.
  validate_mutants.py with unchanged stage0 and sanitized archive catches
  atomic outdated-set, await semicolon and Symbol name-filter mutants in
  72, 99 and 20 programs, first differing bytes 119, 715 and 147. All compile,
  exit zero and emit no stderr. Four raw-question modes reject released
  handles before output with exit 70.
- wave08-next/validate_process.py with /workspace/wave08-process-full, unchanged
  stage0, both checker archives and compiler root: all 100 upstream programs
  and both corpora match normally and sanitized. Callee and blocking-order
  mutants compile, exit zero with empty stderr, and differ in six and nineteen
  programs at first bytes 164 and 1690. Released modes reject with exit 70.
- wave08-next/validate.py with unchanged stage0, both archives and compiler root:
  22 controls / 13 findings and both corpora match normally and sanitized.
  Timeout mutant compiles, exits zero with empty stderr and fails full byte
  comparison. All exercised released-question modes reject before output,
  exit 70. Exact differing byte and complete streams are retained in the logs.
- wave08-react/validate.py with /workspace/wave08-area/globals-upstream:
  111 typed upstream programs / 58 findings, 35 controls / 34 findings, JSX and
  both corpora match normally and sanitized. Ancestry mutant compiles and exits
  zero with empty stderr, differing at byte 9171 in controls and in seven
  upstream programs. The checker-free captured Go witness remains outside the
  typed native comparison. Separate sanitized node-symbol-details probe rejects
  a released handle before output with exit 70.
- wave08-react/validate_cores.py: 52 parked kernel results match Go under
  sanitizers. Join and dependency-count mutants compile and differ under Go
  byte comparison. Both HIR-dependent production entries refuse with exit 70.
  These kernel checks are not complete production-rule mutants.
- go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v:
  PASS, 0.188 seconds. Native and Node observation cache each report one hit
  and zero misses.

All four validation command chains exited zero. No selected Go check skipped.
The full gate, including all 17 externally provisioned correctness checks, was
not run or claimed green. No test was skipped, relaxed or deleted. Unchanged
shared registry, option-bearing witness and runtime checks from the preceding
landing report were not repeated beyond this relevant oracle validation.

No fresh quiet timings were run because compiler, bridge and profile sources
are identical to the preceding validated tree. LANDING_4E0_REPORT.md records
quiet native/Go medians of 2.163x through 21.678x, native slower. Current
process/race validators retain timing observations but overlapping jobs mean
those are not quiet performance comparisons. No speed improvement is claimed.

The audit covers 755 origin refs and 42 distinct claim Markdown blobs: 25
verified ports and 172 claims cover all 197 ranked rules; eligible list empty.
No new reservation is made. Complete log/stdout/stderr/manifest streams and
seven summary JSON files are retained in validation-71d. Setup remains the
previous 132-second setup, nproc 5 and quota four cores; configured tools reused.

Shared typed factory/driver integration still requires a checker-program handle
and configured root list in RuleContext; owned profiles and scratch overlays do
not certify that integration. React immutability and no-deriving-state-in-effects
remain parked for native HIR/SSA/capture analysis on #dnv6f2c. This unit adds only
owned claim/evidence files and publishes only codex/typeaware-wave-08, with an
explicit lease on the previous published SHA.
