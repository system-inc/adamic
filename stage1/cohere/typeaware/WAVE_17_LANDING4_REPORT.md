Built: rebased existing wave-17 work onto main b8fb957a; withdrew duplicate React claims, with no new production ports.
Commits: rebased source tip e6d4109b, from pushed 72ce121c; the withdrawal claim remains historical 14713630.
Checks: all eight selected lint suites PASS 888.958s; checker, full vet, formatting, numeric listeners and uncached Node pass.
Mutants: existing rule/raw-question/handle mutants caught again; shared-regex literal byte 1134; details in mutant-evidence.txt.
Limits: original parked scopes remain; no full repository gate, no duplicate React ports, no unclaimed rules in the origin audit.

Main advanced to b8fb957aa839a9e8cb0b54279dd9864fa317bd30 after the withdrawal
was pushed. All 31 own commits rebased cleanly onto it; the patch correspondence
is recorded in rebase-map.log. Main adds inherited-static-field emission and
its fixture. Those protected files were accepted from main during rebase and
were not edited by this worker. There are no submodule pin changes. The freshly
verified own remote still has 72ce121c217224e637c39069f05b32fbfb284877; the final
owned-branch push uses that exact lease. No main or area branch is pushed.

The collision/withdrawal audit is in WAVE_17_SIXTH_WITHDRAWAL_REPORT.md and
validation-wave-17-sixth-withdrawal. Its 309-control experiment belongs to the
older c01907a7 snapshot and is not offered as a production port. The duplicate
sources remain in /workspace/wave-17-sixth/withdrawn-implementation.patch.
The all-origin audit inspected 583 refs and 33 Markdown claim blobs and found
zero unclaimed entries in the 197-rule checker-dependent ranking. No new
reservation was made after discovering the competing wave-11 claim.

## Current-base validation

| Selected suite | Seconds | Result |
| --- | ---: | --- |
| TestCoverageAgreementAndMutants | 154.68 | PASS |
| TestWave17FifthAgreementAndMutants | 158.45 | PASS |
| TestWave17FourthAgreementAndMutants | 144.71 | PASS |
| TestWave17HeadJudgmentsAndParserBoundary | 79.80 | PASS |
| TestWave17NextAgreementAndMutants | 118.81 | PASS |
| TestWave17AgreementMutantAndJSXBoundary | 66.46 | PASS |
| TestWave17ThirdAgreementAndMutants | 163.57 | PASS |
| TestWave17UnicodeUpper | 2.49 | PASS |

Every selected differential suite preserves complete canonical diagnostics,
including message IDs/text, byte spans, every fix and every suggestion. The
existing positive controls, supported option variants, frozen 287 repository
roots and 77 compiler roots, normal/sanitized runs, decision/raw-question
mutants and released-handle checks all pass again. The head suite explicitly
proves that end-to-end JSX remains blocked by the shared parser refusal; its
projected controls are not relabeled as native-parser coverage. Existing
component/props/analysis and configured-regex limitations remain as stated in
the earlier wave reports. No shared generator, harness or parser was changed.

The separate shared-regex literal mutant changes the placeholder's first
identifier class from [A-Za-z_] to [A-Z_]. Rebuilt under ASan/UBSan, it exits 0
with empty stderr and differs from the independent Go oracle at byte 1134.
Existing mutants and their named catching checks are copied verbatim into
mutant-evidence.txt; complete process stdout/stderr is in command-output.tar.gz.
Failed compiler builds from the withdrawn prototype are not counted as mutants.

Checker tests PASS 0.603s. The filtered uncached Node oracle PASS 5.949s, and
main's new inherited-static-field fixture separately PASS 0.443s. All 15 numeric
listener manifests match production keys; its kind-0 mutant is caught. Full
go vet ./... and gofmt report no output. The full repository test gate was not
run. Log files are preserved, and no test process output was piped.

## Current-base native time against Go

Three alternating rounds after all selected tests stopped, preserving complete
stdout equality in every round. Medians are whole-process seconds:

| Population | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.754273 | 0.384396 | 7.17 |
| repository | 0.353659 | 0.145049 | 2.44 |

These measurements cover the fifth-batch ordinary-parser runner at the compared
corpus/default scope, not projected JSX, complete React configuration, or all
inventory rules. Native remains slower. Each round's phase counters, query
counts and diagnostic SHA-256 are in measurements.json. The original three-rule
suite also logged one compiler observation: native 1.762463s versus Go 0.290863s.

Setup succeeded again: Go ready 0s, clang ready 0s, Node ready 0s, submodules
ready 0s, build cache warm 87s, done 87s. nproc is 5; cgroup quota is 4 CPUs,
memory 17.6 GB. Go 1.27.1, clang 20.1.8 and Node 24.19.0 are unchanged. Each
command sourced /workspace/adamic-tools/env.sh. Disk space was recovered by
removing only identified ELF/ar artifacts from old owned scratch directories;
cleanup.json lists all 139 files and 5,780,745,836 bytes. Source, patches, logs,
controls and manifests were retained.

Commands, with output redirected to the preserved logs:

```sh
git fetch --no-recurse-submodules origin refs/heads/main:refs/remotes/origin/main
git rebase origin/main
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_ARTIFACTS=/workspace/wave-17-landing4/original ADAMIC_WAVE17_HEAD_ARTIFACTS=/workspace/wave-17-landing4/heads ADAMIC_WAVE17_NEXT_ARTIFACTS=/workspace/wave-17-landing4/next ADAMIC_WAVE17_THIRD_ARTIFACTS=/workspace/wave-17-landing4/third ADAMIC_WAVE17_FOURTH_ARTIFACTS=/workspace/wave-17-landing4/fourth ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-landing4/fifth ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-17-landing4/coverage ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^(TestWave17|TestCoverageAgreementAndMutants)' -count=1 -v -timeout=30m
go vet ./...
gofmt -l cmd internal stage1/cohere/typeaware
go test ./bridge/tsgo/checker -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions|method_closures|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)[.]a$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/inherited_static_field_read[.]a$' -count=1 -v
python3 stage1/cohere/typeaware/wave_17_listeners/verify.py
python3 stage1/cohere/typeaware/validation-wave-17-landing/measure.py /workspace/wave-17-landing4/fifth /workspace/wave-17-landing4/timing /workspace/wave-17-typescript
```
