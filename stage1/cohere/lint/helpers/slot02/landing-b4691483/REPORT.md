Rebased all thirty-six existing slot02 helpers onto lint area b46914832; shared profile snapshot comparison blocks landing readiness; no new claims.
Commits: prior published 68b6a6583568f3f454e08dcc9006c2a76456bce3; rebased source 8668d6f51084ab572d52a66281c46ecec741389d; current main c7991b90; own branch codex/lint-helpers-02 only.
Commands: setup PASS 162s/nproc 5; required compiler PASS 152.68s/446 files; registry parity PASS 117.58s; profile compilation PASS 43.88s; snapshot comparison FAIL, selected harness exit 1 after 377.315s; complete helper regression PASS 1055.022s.
Mutants: all 142 helper semantic variants caught with fresh actual-Go mismatches and identical names to the previous completed gate; all three inherited targeted mutations caught.
Not covered: full repository gate, the other sixteen mandatory external library checks, whole-rule integration of these helpers, complete CFG behavior and prior external Tailwind gaps.

# Landing scope

The user made landing the unit when a previously published branch requires a rebase. This is the only branch this worker has published. A wildcard fetch refreshed all twenty origin codex/lint-helpers* branches. Current lint area moved from b84a9d93 to b46914832d70e00847d82d5d221ab7bb24040c53, retaining current main c7991b900362796aefd111474e65eb5398e91953. All 55 local commits rebased cleanly, preserving registry migration, the explicit nil-options adapter guard and RuleContext.has. No main/area push, new claim, shared harness/registration/compiler change or regexp matcher is made. Inherited allocator changes are not reverted.

All thirty-six helper source hashes match the previous completed batch12 manifest. Existing reports and claim SHAs remain historical publication evidence; this directory records the new validation. No helper implementation or standalone oracle is altered. Prior readiness projections and consumer handoffs remain unchanged; this rebase removes no additional prerequisite.

# Validation and inputs

Test output goes directly to logs. Source /workspace/adamic-tools/env.sh first. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup Go/clang/Node/submodules ready at 0s; build cache warm/done 162s, nproc 5.

Pinned cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db. Real TypeScript input /tmp/slot02-required-typescript remains 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). A persistent profile snapshot /tmp/slot02-profile-b4691483 is generated using the inherited TestProfileArtifacts and supplied to TestProfileSnapshotsAgree. The first harness attempt was stopped before its unconfigured profile snapshot check could skip; the incomplete log is retained in evidence/superseded-missing-profile-snapshot.log and receives no passing credit. No test or skip guard is changed.

- bash cloud/setup.sh > evidence/setup.log 2>&1; nproc > evidence/nproc.log: PASS 162s, nproc 5.
- git rebase origin/area/stage1-lint > evidence/rebase.log 2>&1: exit zero, all 55 commits replayed cleanly.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=30m > evidence/helpers-final.log 2>&1: PASS 1055.022s; all 142 semantic mutants caught.
- ADAMIC_LINT_PROFILE_DIR=/tmp/slot02-profile-b4691483 ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestProfileArtifacts$' -count=1 -v -timeout=30m > evidence/profile-artifacts.log 2>&1: PASS 36.213s, release/counted/profiled binaries saved.
- ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/slot02-profile-b4691483 ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestDecorationOptionMutant|TestProfileCompilation|TestProfileSnapshotsAgree|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout=30m > evidence/inherited-harness-oracle.log 2>&1: FAIL 377.315s solely in TestProfileSnapshotsAgree (22.39s). TestRulesAgree PASS 117.58s, 2,017 unique captured source/rule/options combinations and 13,053,452 identical bytes; TestCompilerAndStage1Agree PASS 152.68s, 446 files and 20,749,193 identical bytes; TestProfileCompilation PASS 43.88s, counted allocation/frees 139/139, retains/releases 914/543, peak 99/regions 0; TestDecorationOptionMutant PASS 32.26s, TestCommentFoldMutant PASS 38.23s, TestPositionIndexMutant PASS 40.75s. Rule parity uses generated and captured rule/options rows under the inherited guard. No adapter is allowed to discard row options. The required compiler comparison includes .ts and .a files; the inherited profiling corpus walker selects .ts only, so additional .a manifest coverage is credited to the compiler comparison rather than the profile snapshot.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 8.905s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 4.716s, six probe misses/zero hits.

# Shared blocker and minimal reproducer

Observed on b46914832: TestProfileSnapshotsAgree fails before profile/native/Node comparison. Its captured OctalEscape.ts row contains the exact six source bytes `' \01'`. The real Go oracle exits 2 with `panic: invalid corpus ... [Octal escape sequences are not allowed. Use the syntax '\x01'.]; source="' \\01'"`. The completed seven-test selection remains failed; it is not called green. No nil-options-adapter guard panic occurred.

The retained raw fixture testdata/OctalEscape.a.txt and testdata/profile-invalid-corpus.txt reproduce this independently of a full corpus. Run `python3 stage1/cohere/lint/helpers/slot02/landing-b4691483/testdata/reproduce.py` from the repository root to regenerate the required absolute source path, invoke both oracle modes and assert the exact diagnostic/panic; an optional first argument supplies the oracle path. The fixture is raw invalid TypeScript text, not a new Adamic program. With the actual profile oracle generated by TestProfileArtifacts:

```
/tmp/slot02-profile-b4691483/oracle --manifest stage1/cohere/lint/helpers/slot02/landing-b4691483/testdata/profile-invalid-corpus.txt --diagnostics > evidence/profile-reproducer-diagnostics.log 2>&1
/tmp/slot02-profile-b4691483/oracle --manifest stage1/cohere/lint/helpers/slot02/landing-b4691483/testdata/profile-invalid-corpus.txt > evidence/profile-reproducer.log 2>&1
```

The diagnostics query exits 0 and prints 1; the ordinary oracle run exits 2 with the same octal-escape panic. Selecting no-var in the minimal row demonstrates that parse rejection precedes rule dispatch; it does not assert that no-var was the selected rule in the deleted full-corpus temporary directory. evidence/profile-reproducer-exit.log retains exit 2. The full failed corpus log preserves its actual OctalEscape.ts witness.

Inference supported by the code: ordinary TestRulesAgree calls recoveryRows before its manifest and passed this corpus. TestProfileSnapshotsAgree constructs its manifest without that shared recovery classification and reaches the oracle's invalid-corpus guard. The likely correction belongs to shared profile_test.go, outside the helper territory. This report does not claim the regression was introduced by b46914832; it records the observed failure on that requested base. No check is skipped, relaxed, deleted or patched here. Work stops at this blocker after the already-running helper regression completes; no new helpers are selected.

# Evidence and limits

Every helper semantic variant must compile and run cleanly before its output differs from actual Go. All first-mismatch witnesses are extracted from the fresh complete run, and names are checked against the previous 142. Source Node and sanitized native are compared for the original shared helpers and first slot02 suite; batch2 onward additionally compares emitted JavaScript. No new emitted-JavaScript coverage is claimed for those earlier suites. Inherited folding/position mutants compare Go with Node/emitted JavaScript/native; the inherited decoration-range mutant compares Node/native only.

This scoped gate does not run the full repository gate or the other sixteen required external compiler/library checks. None is claimed passed or skipped-to-green. Full rule findings/fixes/suggestions through production helper wiring, malformed arenas, callback failures/concurrency, complete CFG semantics, the separate inherited comments package and prior eight external Tailwind capture gaps remain outside this landing unit. No new rule options adapter is owned here. If an inherited guard exposes a shared adapter failure, the exact reproducer must be reported without editing shared territory. Final remote/ancestry and clean-tree checks precede publication; an exact lease protects the original remote head 68b6a6583568f3f454e08dcc9006c2a76456bce3.

Final verification: the complete helper run passes with no skipped checks and all 142 fresh mutant names matching the prior gate. All thirty-six source hashes match before and after rebase. Final fetch confirms unchanged main c7991b90, area b46914832 and own remote lease target 68b6a658. Both upstream heads are ancestors of tested source 8668d6f5. The seven-test inherited harness remains FAILED solely in TestProfileSnapshotsAgree; the required compiler comparison and all three targeted mutants pass. This branch is published with that explicit shared blocker for review, and is not represented as fully landing-ready. No further work is claimed.

Raw failed harness stdout/stderr is archived byte for byte in evidence/inherited-harness-oracle.log.gz (gzip round-trip verified). The live .log remains in the workspace; no failure or whitespace is edited out of the raw evidence. The reproducer names an explicit normal (non-recovery) mode in its final manifest field.
