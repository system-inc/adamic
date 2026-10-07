Rebased thirty-six existing helpers onto current lint area d3a37422 containing main b6b1538b; no new claims; the shared profile snapshot blocker is reproduced again, and no new claims are made.
Commits: prior published 801edcc3ca278bd2d49327d3938a756ad59d3388; rebased source 475c1c4bd50175b747c3ce5673de16b80bf5caaf; own branch codex/lint-helpers-02 only.
Commands: setup PASS 53s/nproc 5; eight-test inherited harness FAIL 373.930s solely in profile snapshots; required compiler PASS 135.16s/446 files; profile artifacts/compilation and registry parity PASS; complete helper regression PASS 1012.998s.
Mutants: all 142 helper semantic variants and three inherited targeted variants caught with fresh independent Go mismatches.
Not covered: full repository gate, the other sixteen mandatory external library checks, integrated helper findings/fixes/suggestions, full CFG semantics and prior external Tailwind gaps.

# Landing scope

This is the only branch published by this worker. The work-in-progress cap makes rebasing and validating it the unit; no new helpers or rules are selected. All twenty origin codex/lint-helpers* branches were fetched. Main advanced from c7991b90 to b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 with the typeof/null dispatch and lookup-presence fix. Lint area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 merges that main and retains b46914832's registry migration, options-adapter guard and RuleContext.has. The owned branch rebased cleanly. All thirty-six helper source hashes match the previous batch12 manifest; no helper implementation or standalone helper oracle changed.

No shared compiler, harness, registration, inventory or rule file is edited. No main/area branch is pushed. Inherited allocator changes are retained. No new Adamic .ts, regex matcher, node-kind dispatcher or options adapter is written. Earlier reports and commit hashes remain historical publication evidence. Readiness and consumer handoffs are unchanged; no additional prerequisite is removed by a rebase.

# Validation commands

All test output goes directly to log files; source /workspace/adamic-tools/env.sh before Go. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5. Setup Go/clang/Node ready at 0s, submodules ready 1s, cache warm/done 53s. Pinned real TypeScript /tmp/slot02-required-typescript is 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3); cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db.

- bash cloud/setup.sh > evidence/setup.log 2>&1; nproc > evidence/nproc.log: PASS 53s, nproc 5.
- git rebase origin/area/stage1-lint > evidence/rebase.log 2>&1: exit zero.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=30m > evidence/helpers-final.log 2>&1: PASS 1012.998s, all 142 mutants. TestSlot02Batch12 PASS 305.79s.
- ADAMIC_LINT_PROFILE_DIR=/tmp/slot02-profile-d3a37422 ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/slot02-profile-d3a37422 ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestDecorationOptionMutant|TestProfileArtifacts|TestProfileCompilation|TestProfileSnapshotsAgree|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout=30m > evidence/inherited-harness-oracle.log 2>&1: FAIL 373.930s solely in TestProfileSnapshotsAgree (26.36s); registry parity PASS 90.00s, 13,051,687 identical bytes; required compiler PASS 135.16s, 446 files/20,749,193 identical bytes; profile artifacts PASS 41.68s; profile compilation PASS 38.25s; decoration/fold/position mutants PASS 37.31s/42.40s/32.91s. Raw output is archived losslessly in evidence/inherited-harness-oracle.log.gz; the live .log remains in the workspace. ProfileArtifacts creates the snapshot before ProfileSnapshotsAgree. The failing shared test stays in this selection; no skip or guard is changed. The required compiler corpus includes .a and .ts; the inherited profile snapshot walker selects .ts, with additional .a coverage credited to the required compiler comparison.

# Prior blocker and reproducer

At b46914832 the ordinary registry rule comparison passed, but TestProfileSnapshotsAgree failed in the real Go oracle on captured OctalEscape.ts source `' \01'`, exit 2, before any profile comparison. Its manifest omits the recovery classification applied by TestRulesAgree through recoveryRows. That shared profile_test.go file is unchanged on the newly fetched area; this observation motivates a fresh run, rather than claiming a fix landed. No nil-options adapter guard failure was observed in the earlier run.

testdata/reproduce.py copies the prior minimal reproducer without changing the earlier landing evidence. It regenerates TypeScript-go's required normalized absolute path to the six-byte raw-text fixture, calls --diagnostics and ordinary --manifest, and asserts diagnostics 1 plus exit 2 with the exact octal-escape panic. Invoke it with the newly generated oracle as its first argument, after TestProfileArtifacts completes:

```
python3 stage1/cohere/lint/helpers/slot02/landing-d3a37422/testdata/reproduce.py /tmp/slot02-profile-d3a37422/oracle
```

The fresh script writes only this landing directory's manifest and evidence, preserving historical evidence in landing-b4691483. Fixing shared recovery classification is outside this territory. The failure is not claimed introduced by the typeof/null merge. A shared failure remains an explicit landing blocker; new claims stop until it is resolved.

# Witnesses and limits

Every helper semantic mutant must compile and run cleanly before differing from actual Go. Fresh names and first mismatches are compared with the previous 142. The original shared helpers and first slot02 suite compare Go/source Node/sanitized native; batch2 onward also compares emitted JavaScript. No extra emitted coverage is claimed for earlier suites. Inherited fold/position mutants compare Node/emitted JavaScript/native; the decoration-range mutant compares source Node/native only.

The full repository gate and other sixteen mandatory external compiler/library checks are not run or claimed skipped-to-green. Production helper dependency wiring, whole-rule findings/fixes/suggestions, complete CFG semantics, arbitrary malformed arenas, callback failures/concurrency, the separate inherited comments package and prior external Tailwind live/corpus gaps remain outside this scoped gate. Final upstream and own-branch lease checks precede publication. Only codex/lint-helpers-02 is pushed, with an exact lease against 801edcc3ca278bd2d49327d3938a756ad59d3388; a surviving shared profile failure is retained in the report rather than edited out of the suite.

Fresh blocker observation: TestProfileSnapshotsAgree again passes an unclassified captured OctalEscape.ts row to the Go oracle and receives the octal-escape panic before profile comparisons. The newly generated oracle independently reports parse diagnostic 1, then exits 2 with the same panic under testdata/reproduce.py; evidence/reproducer-run.log records the successful assertion of that expected failure. No nil-options guard panic occurs and no selected check is skipped. Registry parity and compiler parity agree across Go/source Node/emitted JavaScript/native; the helper gate remains separate from the failed profiling check. Fixing profile recovery classification remains outside this territory. No check is weakened or excluded to obtain a green harness.

Supporting checks: go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1 exits zero; go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1 PASS 5.875s; ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1 PASS 10.486s, six probe misses/zero hits.

Final verification: all 142 fresh helper mutant names match the prior completed gate and all thirty-six source hashes remain unchanged. No selected helper or inherited check was skipped. Final fetch confirms current main b6b1538b and lint area d3a37422 remain ancestors of tested source 475c1c4b, with exact remote lease target 801edcc3. This source is green against its complete helper oracle and required compiler comparison. The separate shared profile snapshot comparison remains FAILED with a fresh exact reproducer and is not claimed landing-ready or repaired. Publication changes only the owned branch; no new helper or rule is claimed.
