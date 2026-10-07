Built: merged current lint area e667e3e1d into wave 08, retaining current main 71d7e491b.
Commits: tested combined tip a358a88d40d3a3d7f618e1e4ca363bdfd6fbf29e; this report accompanies its evidence commit.
Checks: core 264/264 normal/sanitized and shared multi-edit, path, suggestion, witness and node-table comparisons pass.
Mutants: all ten unchanged completed-rule mutants remain held by preceding Go-byte evidence; registry rejection mutants pass again.
Uncovered: six other fresh owned-profile reruns, new rule mutants/timings, shared typed integration, two HIR rules and full gate.

Previous published tip was 9381c8ee9afb5cc2e5e88e5c763051d858d52f6f. Current
origin/area/stage1-lint is e667e3e1dbdfd1b9125c3256961bfbc8ec31946b; current
origin/main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. The area tip was
merged cleanly into the owned branch, preserving both ancestors. It adds shared
support for multiple automatic edits and preserves directory paths in captured
cases and witnesses. No files under cmd, internal, bridge or
stage1/cohere/typeaware change from the previously published tree. Owned native
profiles do not import shared lint context/finding/driver modules. Their existing
compiler, core binaries and checker archives were therefore reused for fresh
core comparisons; other completed profiles retain their previous validation.
No shared implementation was edited, and no main or area branch is pushed.

Only archived reproducible ELF/ar scratch binaries larger than one megabyte
under /workspace/wave08-71d were removed: 26 artifacts, 1,059,063,359 bytes.
Sources and committed output evidence remain. All new test output went to files
under /workspace/wave08-e667, never through a running pipe. The configured
/workspace/adamic-tools/env.sh was sourced for every command chain.

Commands and observed results:

- go test ./stage1/cohere/lint/registry -count=1 -v: PASS, 0.131 seconds,
  including descriptor rejection mutants and .a module support.
- go test ./stage1/cohere/lint -run
  '^(TestRulesAgree|TestNodeTableIsLinkOnly|TestOwnedWitnesses|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix)$'
  -count=1 -v -timeout 30m: PASS. Complete suggestion serialization matches Go,
  source Node, emitted JavaScript and native on 735 bytes, 109.90 seconds.
  Automatic fix plus three unapplied suggestions preserves 721 bytes,
  46.86 seconds. Broad RulesAgree preserves 13,071,542 bytes, 89.91 seconds,
  including the changed multiple-edit model and fixture directory handling.
  Node-table guard preserves 13,085,662 bytes across 2,154 rows with and without
  unattached rows, including 2,017 unique captured cohere source/rule/options
  combinations, 40.04 seconds. Owned witnesses preserve 113,269 bytes across
  Go, source Node, emitted JavaScript and sanitized native, 41.48 seconds.
  Exact package duration is retained in the complete log.
- compare_core.py --fixtures /workspace/wave08-core/upstream --compiler-root
  /workspace/typescript-wave08-corpus, with /workspace/wave08-4e0/core-native
  and core-native-asan and separate artifact roots: 264/264 full streams agree
  in each run, including findings, fixes and suggestions. Both processes exit
  zero. Sanitized comparison retains ASan/UBSan/leak checks.
- go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v:
  PASS, 0.146 seconds. Native and Node observation caches each report one hit
  and zero misses.

All selected command chains exited zero. No selected Go check skipped. No test
or options guard was relaxed, bypassed or deleted. The full gate, including all
17 externally provisioned correctness checks, was not run or claimed green.
This landing unit changes evidence only; all ten completed-rule mutants and
released-handle refusal checks from LANDING_71D_REPORT.md remain applicable to
the unchanged owned implementations. They were not rerun in this unit. The two
parked kernel checks were not rerun either. New shared implementation coverage
is provided by the actual Go/Node/JavaScript/native comparisons above; registry
rejection mutants are separately tested.

No new quiet timings were run because compiler, bridge and owned profiles are
unchanged. LANDING_4E0_REPORT.md records quiet native/Go medians of 2.163x through
21.678x, native slower. No performance improvement is inferred.

The audit covers 772 origin refs and 42 distinct claim Markdown blobs. All 197
ranked rules are ported or claimed: 25 verified ports and 172 claims, eligible
list empty. No new reservation is made. Complete logs/stdout/stderr/manifests
and three summary JSON files are retained in validation-e667. Original setup
was 132 seconds, nproc 5 and quota four cores; configured tools reused.

Shared typed factory/driver integration still requires a checker-program handle
and configured root list in RuleContext; owned profiles and scratch overlays do
not certify this integration. React immutability and no-deriving-state-in-effects
remain parked for native HIR/SSA/capture analysis on #dnv6f2c. Only owned
claim/evidence files change. Publication targets codex/typeaware-wave-08 with
an explicit lease on its prior SHA.
