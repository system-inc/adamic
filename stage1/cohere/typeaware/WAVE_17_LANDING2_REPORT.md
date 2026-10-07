Built: rebased all existing wave-17 work onto current main f8013f0b; no new rules claimed.
Commits: previous pushed tip 61bdd2eb, validated rebased source 6ffc1aff; evidence is committed with this report.
Commands and outputs: eight selected oracles PASS 860.616s; checker PASS 0.449s; uncached Node PASS 29.045s; vet clean; listener checks PASS; setup 99s, nproc 5.
Mutants: all existing rule, fact, tracker, handle, Unicode and Node byte mutants caught again, plus numeric-listener mutation; exact catchers are in validation-wave-17-landing2/results.json.
Not covered: full repository gate, production JSX, complete BooleanPropNaming and node-handler conversion; native remains slower than Go.

Main advanced by 25 commits since the prior landing base. Its Map/Set iterator,
narrowing, override and pragma-refusal changes warranted repeating the byte
oracle rather than relying on old observations. Rebase completed cleanly for all
22 commits, with no rule-source repairs or shared-file edits. The validated base
is f8013f0baac41ddc340d76f83bddde38536a8f07, and source tip is
6ffc1affac5b23b9d49de289cb1ca392b8dab80f. A remote check after validation and
timing confirmed main still at that base and the wave branch still at 61bdd2eb.
The exact old-to-new mapping is in rebase.json. Push uses an exact lease against
61bdd2eb, as authorized by the user's explicit rebase-and-push instruction; only
codex/typeaware-wave-17 is pushed. This takes precedence over the local default
against rewriting history. No main or area branch is pushed.

| Oracle | Seconds | Result |
| --- | ---: | --- |
| inherited inventory | 145.59 | PASS |
| fifth batch, partial PropTypes scope | 149.44 | PASS |
| fourth batch | 140.56 | PASS |
| Head decisions and parser boundary | 81.36 | PASS |
| Nexus | 121.22 | PASS |
| async client and JSX boundary | 65.80 | PASS |
| global rules | 154.31 | PASS |
| Unicode uppercase | 2.34 | PASS |

All selected rule comparisons include full findings, fixes and suggestions,
normal and ASan/UBSan builds, comparison-only rule mutants and released-handle
controls. Registry mutants finish cleanly and are caught by the required panic.
Go unicode.IsUpper agrees for all 1,114,112 code points; the boundary mutant is
caught at byte 1. Every exact catcher is preserved in results.json and the log.
The inherited inventory corpus checks are run separately because the combined
command omits their optional environment flags. Both frozen populations pass
Go/native/sanitizer byte equality: repository 85,151 bytes; compiler 7,120,921
bytes. Hashes, manifests and compressed complete streams are preserved.

The uncached Node run includes the byte mutant, functions, closures, generics,
method closures, devirtualization and call-target fixtures; Go's subtest matching
also selects regexp_cycle_closures. It reports native hits=0 misses=34 and Node
hits=0 misses=23. It checks native sanitizer output and emitted JavaScript against
Node. The direct bridge checker and numeric listener verification also pass.
Changing SourceFile 307 to Unknown 0 is caught by listener verification. This
metadata check does not imply that old rule implementations accept loaded numeric
nodes: existing entry points still scan or use table indexes. No shared driver,
registration generator or harness is changed.

Quiet three-round alternating whole-process medians after all tests finished:

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.400826 | 0.388286 | 6.18 |
| repository | 0.331502 | 0.131676 | 2.52 |

Every timed round matches complete diagnostic bytes. These fifth-batch production
measurements cover the ordinary parser, not projected JSX or the inherited ten
inventory rules. Query counts are 78,665 and 4,520 respectively. Native is slower;
no speed improvement from the listener declarations is claimed.

The full repository gate was not run. Production JSX remains an explicit parser
refusal, and BooleanPropNaming's configurable regexp matching and full
component/props integration remain incomplete. The shared unknown-question mutant
still uses an obsolete refusal anchor. These are the existing limits, not complete
rule parity claims. The branch remains a scoped, green landing candidate.

Commands, all with test output redirected to logs:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
git rebase origin/main
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_ARTIFACTS=/workspace/wave-17-landing2/original ADAMIC_WAVE17_HEAD_ARTIFACTS=/workspace/wave-17-landing2/heads ADAMIC_WAVE17_NEXT_ARTIFACTS=/workspace/wave-17-landing2/next ADAMIC_WAVE17_THIRD_ARTIFACTS=/workspace/wave-17-landing2/third ADAMIC_WAVE17_FOURTH_ARTIFACTS=/workspace/wave-17-landing2/fourth ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-landing2/fifth ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-17-landing2/coverage ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^(TestWave17|TestCoverageAgreementAndMutants)' -count=1 -v -timeout=30m
go vet ./...
go test ./bridge/tsgo/checker -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions|method_closures|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)[.]a$' -count=1 -v
python3 stage1/cohere/typeaware/wave_17_listeners/verify.py
python3 stage1/cohere/typeaware/validation-wave-17-landing/measure.py /workspace/wave-17-landing2/fifth /workspace/wave-17-landing2/timing /workspace/wave-17-typescript
```

The independent inventory binaries coverage-oracle, coverage and coverage-asan
were each run with the repository or compiler tsconfig and their unchanged frozen
manifests in /workspace/wave-17-fourth. All returned zero, native stderr was empty,
and all three complete stdout streams matched for each corpus. No fixture set
was replaced to obtain agreement.
