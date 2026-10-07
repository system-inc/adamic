Rebased the thirty-six completed helpers onto recovery fix 488385016; the shared profile blocker is cleared and no new helpers are claimed in this landing unit.
Commits: prior published 6211705b968ddfc9c96a3730162d85eb27f716b1; tested rebased source a34bbb81a69f39a3d16fb65748d8e2a7bf4ebeb5; publication goes only to codex/lint-helpers-02.
Commands: setup PASS 69s, nproc 5; requested eight-test inherited selection PASS 439.280s; options-witness check PASS 52.017s; profile snapshot PASS 108.90s with 33,612,508 identical bytes.
Mutants: recovery classification removal reproduced the exact octal-escape panic; decoration range, comment fold, position index and ignored decoded-option mutants were caught independently against Go.
Not covered: full repository gate and other sixteen external correctness checks; unchanged standalone helpers were not rerun, with their prior 142-mutant gate retained as historical evidence.

The explicit fix commit 488385016a8477bf46a091f4b25e57626a241d58 contains current main b6b1538b and lint area d3a37422. The area did not yet contain the fix when fetched. The owned branch rebased cleanly onto the fix. The inherited diff adds profile recovery classification and an options-witness check; it does not change the compiler, runtime or helpers. All thirty-six helper SHA256 values match landing-d3a37422/evidence/helper-source-manifest.json. No shared file was edited by this worker. Earlier failure reports and reproducer evidence remain unchanged as historical records.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 69s, done 69s; nproc 5. Source /workspace/adamic-tools/env.sh for the following commands. Every test writes directly to its log. Complete raw logs are archived losslessly in evidence/*.log.gz.

```
bash cloud/setup.sh > evidence/setup.log 2>&1
nproc > evidence/nproc.log
git rebase 488385016 > evidence/rebase.log 2>&1
ADAMIC_LINT_PROFILE_DIR=/tmp/slot02-profile-48838501 ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/slot02-profile-48838501 ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestDecorationOptionMutant|TestProfileArtifacts|TestProfileCompilation|TestProfileSnapshotsAgree|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout=30m > evidence/inherited-harness-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestDecodedOptionsAndMutant$' -count=1 -v -timeout=30m > evidence/options-witness.log 2>&1
```

The real TypeScript input remains pinned to 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). Registry parity passed in 93.82s, producing 13,053,452 identical bytes on Go, source Node, emitted JavaScript and native. The required compiler corpus comparison passed in 118.44s with 20,749,193 identical bytes on those four backends. Profile artifact generation passed in 40.08s; profile compilation passed in 43.15s. Profile snapshots passed in 108.90s: release, profiled, source Node and emitted JavaScript were compared to Go, with 33,612,508 identical bytes. No selected check skipped. The raw corpus byte count is this branch's observation, not the fix author's different branch count.

The baseline option witness produced 47 identical bytes across all four backends. Ignoring decoded options was caught on source Node and native. Decoration-range collapse was caught on source Node and native; comment-fold and position-index variants were caught on source Node, emitted JavaScript and native. Their fresh mismatches are in the complete logs.

The recovery mutant changes exactly `manifest(t, recoveryRows(t, oracle, rows))` to `manifest(t, rows)` through Go's file overlay. It compiles and runs TestProfileSnapshotsAgree, then fails in 22.40s with the original captured OctalEscape.ts panic and source "' \\01'". Go test exits 1 because the oracle subprocess exits 2. This is an expected mutant failure, not a remaining baseline failure. Shared profile_test.go remains unchanged. evidence/profile-recovery-mutant.go.txt is the exact overlay source; evidence/recovery-mutant-overlay.json records the absolute mapping used. Reproduce by copying that raw source to /tmp/slot02-profile-recovery-mutant.go, regenerating the mapping for this checkout's absolute profile_test.go path, and running:

```
ADAMIC_LINT_PROFILE_DIR=/tmp/slot02-profile-48838501 ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/slot02-profile-48838501 ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test -overlay=stage1/cohere/lint/helpers/slot02/landing-48838501/evidence/recovery-mutant-overlay.json ./stage1/cohere/lint -run '^TestProfileSnapshotsAgree$' -count=1 -v -timeout=30m > evidence/recovery-mutant.log 2>&1
```

This unit follows the explicit request to rerun the inherited harness selection after the shared-only fix. The compiler/runtime and standalone helper tests are byte-identical to the previously validated source, so the complete helper gate is not repeated. Its PASS 1012.998s and all 142 caught mutants remain recorded in landing-d3a37422. Whole-rule helper integration, finding/fix/suggestion serialization, complete CFG semantics and prior Tailwind live corpus gaps remain outside this unit. The full repository gate and remaining sixteen external checks are not run or claimed skipped-to-green. Only the owned branch is published with an exact lease against the prior published SHA; no main or area push and no new claim.
