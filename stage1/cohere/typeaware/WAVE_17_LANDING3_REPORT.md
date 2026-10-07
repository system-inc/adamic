Built: rebased existing wave-17 work onto current main c01907a7 and replaced message interpolation's handwritten matcher with the shared regex-table literal.
Commits: previous pushed tip ce2877d3; clean 25-commit rebase f36356f1; validated interpolation source 43918ab4f587e291c8c9d916b0f51fde066b75f8; evidence accompanies this report.
Commands and outputs: eight landing oracles PASS 909.575s; changed fifth batch PASS 171.661s; checker PASS 0.304s; uncached Node PASS 3.609s; vet/listeners clean; setup 46s, nproc 5.
Mutants: existing rule/fact/handle/Unicode/Node/listener mutants caught; new message RegExp class mutant caught at byte 1134, exit zero and sanitizer stderr empty.
Not covered: full repository gate, production JSX, configured RegExp dialect parity, complete BooleanPropNaming and numeric loaded-node conversion; no new claims.

The landing-first cap took precedence over new selection: main advanced 100
commits, and all 25 existing branch commits rebased cleanly. No shared compiler,
parser, registration generator or test harness was edited or reverted. The
announced type-aware leak-helper change was not in this main update. The changed
main files are stage-3 work, documentation and a stage-3 oracle hook; compiler,
bridge and type-aware sources are identical between f8013f0b and c01907a7.
We still reran the scoped oracle set after rebasing as requested.

Main was checked again after final validation and timing and remained
c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. The wave branch was still ce2877d3,
which is the exact lease used for the rebased push. The explicit user request
authorizes this history rewrite over CLAUDE's normal default. Only
codex/typeaware-wave-17 is pushed; no main or area branch is changed.

The eight-oracle run starts on f36356f1, with the previous default RegExp literal
already present. During validation, the shared regex branch became available and
its row react/boolean_prop_naming.go:621 identified message interpolation's
remaining handwritten regex scan. The exact translated literal and flags are
now initialized once in the rule:

```
/\{\{[ \t\n\f\r]*([A-Za-z_][A-Za-z0-9_]*)[ \t\n\f\r]*\}\}/gu
```

Interpolation resets lastIndex, executes that literal, reads the captured key,
and slices using the returned UTF-16 index. Placeholder values and unknown
placeholders retain the production Go behavior. Configured pattern rows 93 and
182 prescribe new RegExp(pattern, 'u'); that constructor remains parked rather
than replaced with a handwritten matcher. The shared row snapshot and exact
origin/codex/lint-regex SHA are in shared-regex-rows.json. No shared-table file
was copied into production or edited.

The source change was tested by a separate complete fifth-batch oracle on final
source, PASS 171.661s. This distinction is explicit: the 909.575s run covers the
rebased sources, and the subsequent fifth-batch result validates the changed
module. The other rule modules were not changed. All 155 projected controls,
44 ordinary-parser controls, nested/message option variants and both frozen
corpora agree with production Go in findings, fixes and suggestions, normal and
under ASan/UBSan. Each existing comparison mutant and released-handle check
passes again. A new literal mutant changes [A-Za-z_] to [A-Z_], excluding the
known lowercase placeholder names. It compiles and finishes normally under
sanitizers; only independent Go message-option bytes catch it at byte 1134.
Mutated source, suite and complete output streams are preserved.

| Landing oracle | Seconds | Result |
| --- | ---: | --- |
| inherited inventory | 149.37 | PASS |
| fifth batch before interpolation change | 165.30 | PASS |
| fourth batch | 147.12 | PASS |
| Head decisions / parser boundary | 82.31 | PASS |
| Nexus | 130.12 | PASS |
| async client / JSX boundary | 68.68 | PASS |
| global rules | 164.35 | PASS |
| all Unicode code points | 2.32 | PASS |

The inherited inventory's optional frozen-corpus comparisons are separately
renewed with the new binaries. Go/native/sanitizer streams match: repository
85,151 bytes; compiler 7,120,921 bytes. Hashes and compressed full streams are
preserved. All per-rule, checker-fact, tracker and released-handle mutations
are recorded verbatim in results.json and oracles.log. The Unicode check covers
all 1,114,112 code points against Go unicode.IsUpper. The uncached Node run,
including the one-byte mutant, reports zero hits and 34 native / 23 Node misses;
it covers the listed functions, closures, call targets and devirtualization,
plus regexp_cycle_closures selected by Go's subtest matching.

Quiet three-round alternating medians after every test and build finished:

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.451169 | 0.371740 | 6.59 |
| repository | 0.342741 | 0.130534 | 2.63 |

All timing rounds match complete diagnostics. This measures the final fifth-batch
ordinary-parser binary, not projected JSX or all inventory rules. Native remains
slower. Full phase and round measurements are in timing/measurements.json.

Disk had only 386 MB free before validation. Only ELF binaries and ar archives
identified by their magic bytes in the old wave-17-landing and wave-17-landing2
scratch directories were removed. The exact paths and sizes are in cleanup.log.
No source, logs, manifests or diagnostic evidence was removed. This freed about
7.6 GB for the new build and checks.

The current claim scan inspected 557 refs and 33 unique Markdown claim blobs;
only eight React entries lacked both claims and source-name hits on the two
requested bases. This unit is landing and regex compatibility, so it adds no
new claims. Existing React/Next parked statuses remain. Full JSX, component and
props integration, configured Go-regex dialects and node-handler conversion are
not represented as complete parity. The full repository gate was not run.

Commands, with outputs directed to the preserved logs:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
git rebase origin/main
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_ARTIFACTS=/workspace/wave-17-landing3/original ADAMIC_WAVE17_HEAD_ARTIFACTS=/workspace/wave-17-landing3/heads ADAMIC_WAVE17_NEXT_ARTIFACTS=/workspace/wave-17-landing3/next ADAMIC_WAVE17_THIRD_ARTIFACTS=/workspace/wave-17-landing3/third ADAMIC_WAVE17_FOURTH_ARTIFACTS=/workspace/wave-17-landing3/fourth ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-landing3/fifth ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-17-landing3/coverage ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^(TestWave17|TestCoverageAgreementAndMutants)' -count=1 -v -timeout=30m
ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-landing3-regex ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^TestWave17FifthAgreementAndMutants$' -count=1 -v -timeout=30m
go vet ./...
go test ./bridge/tsgo/checker -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions|method_closures|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)[.]a$' -count=1 -v
python3 stage1/cohere/typeaware/wave_17_listeners/verify.py
python3 stage1/cohere/typeaware/validation-wave-17-landing/measure.py /workspace/wave-17-landing3-regex /workspace/wave-17-landing3-regex/timing /workspace/wave-17-typescript
```

Separate inventory corpus comparisons run coverage-oracle, coverage and
coverage-asan with the unchanged repository/compiler config and manifests from
wave-17-fourth. All exit zero, native stderr is empty and complete streams match.
The new message mutant is run with --boolean-props --message-test on the final
control config and manifest; its independent Go counterpart uses the same inputs.
