Built: numeric listener declarations in all fifteen completed wave-16 rule modules; shared numeric node dispatch remains pending.
Commits: implementation b88d26044124a8e92f1ccb24b4e7dd871f39fb61, based on current origin/main e8ba3d5d; evidence follows on this same branch.
Commands/output: five native oracle suites PASS 423.026s; numeric checks and normal/sanitized probes pass; 60 timing processes and 30 count pairs pass; vet clean.
Mutants: fifteen semantic mutants, fifteen static numeric mutations, one compiled numeric mutation and five retained-registry mutations caught by independent checks.
Uncovered: complete numeric per-node handlers, shared Nexus dispatch integration, three React source ports and their source frontend; no new claims or shared edits.

# Wave 16 numeric listener declarations

This unit's only pushed branch is codex/typeaware-wave-16. Fresh all-head fetch
and a subsequent main fetch leave main at e8ba3d5d81de4d3773c723914fccd4c76248b965.
The branch already descends from that base; another rebase was unnecessary.
Only the owned rule modules, probe, evidence and this report changed. Nothing
is pushed to main or an area branch. No shared generator, harness, parser or
protected compiler file was edited.

Each owned rule module exports listenerKinds(): readonly number[]. Its literal
numeric values come from the pinned Go parser's ast/kind_generated.go, not an
invented numbering or a different TypeScript release. The declarations contain
no node lookup or string-kind read. check.py independently extracts Go production
listener registrations and resolves their keys through that enum. It validates
all fifteen declarations and changes a numeric value in each as a static witness.
results.json preserves the independent names, numbers and source hashes.
Run `python3 stage1/cohere/typeaware/validation-wave16-listeners/check.py`
with stdout/stderr redirected to check.log to reproduce the numeric checks.

The owned wave16_listeners_probe.a calls all fifteen exported declarations.
Its complete native stdout matches the Go-derived expected stream normally and
under ASan/UBSan/LeakSanitizer, with empty stderr. A no-alert declaration-only
mutation changes CallExpression 214 to NewExpression 215. It compiles, exits
zero, leaves stderr empty, and the independent numeric bytes catch byte 254.
Its source and output are retained. These numeric witnesses supplement, rather
than replace, the qualifying semantic mutants below.

## Dispatch boundary

The shared ParseNode currently has readonly kind: string and no numeric kind
field. The incoming kind-indexed driver is absent on this base. The declarations
are ready for it, but today's legacy run methods retain their string predicates,
node lookups and scans. Numeric per-node handlers have not been implemented or
represented as complete. Finishing that conversion needs the shared numeric node
representation/driver contract, outside this unit's allowed rule directories.
No speed improvement from these declarations is claimed. SourceFile declarations
for the contextual rules match their production Go registrations.

## Oracle rerun

All output was redirected to validation-wave16-listeners/oracles.log:

```
source /workspace/adamic-tools/env.sh
export ADAMIC_WAVE16_ARTIFACTS=/workspace/wave16-artifacts/listeners/first
export ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS=/workspace/wave16-artifacts/listeners/followup
export ADAMIC_WAVE16_THIRD_ARTIFACTS=/workspace/wave16-artifacts/listeners/third
export ADAMIC_WAVE16_FOURTH_ARTIFACTS=/workspace/wave16-artifacts/listeners/fourth
export ADAMIC_WAVE16_FIFTH_ARTIFACTS=/workspace/wave16-artifacts/listeners/fifth
export ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest
export ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript
go test ./stage1/cohere/typeaware -run '^TestWave16.*AgreementAndMutants$' -count=1 -v -timeout 20m \
 > stage1/cohere/typeaware/validation-wave16-listeners/oracles.log 2>&1
```

PASS 423.026s. The frozen repository has 287 roots and compiler corpus 77 roots.
Normal and sanitized findings, fixes and suggestions match the full production
Go output bytes. Original/follow-up/third/fourth/fifth controls have 65/22/50/33/231
findings. The original set retains two repository and one compiler finding; the
other sets have zero there. The fifth set's three extra option profiles retain
225/243/235 findings. All five released-handle probes panic 70 as required; all
five retained-registry mutants exit zero and are caught. The Nexus sets still
use the previously supplied shared-dispatch build overlay: ordinary shared
question dispatch is not integrated on this branch.

Every qualifying semantic mutant compiles, exits zero and has empty stderr;
the independent full Go byte comparison alone catches it:

| Mutant | First differing byte | Findings |
| --- | ---: | ---: |
| throw-range | 124 | 231 |
| backreference-range | 9259 | 231 |
| arrow-fix | 6041 | 231 |
| listener-general | 717 | 23 |
| render-number | 7123 | 14 |
| mock-range | 3549 | 22 |
| function-range | 126 | 33 |
| native-range | 4589 | 33 |
| wrapper-range | 5171 | 33 |
| import-write | 63 | 59 |
| exponent-base | 15651 | 65 |
| nan-comma | 12061 | 63 |
| shell-heredoc | 8278 | 53 |
| sql-string | 11751 | 31 |
| alert-range | 124 | 50 |

308 fresh oracle stdout/stderr streams are retained under streams/; stdout is
compressed losslessly. provenance.json records byte counts, hashes, original
paths, source commit and base. Controls are reproducible from the unchanged
owned Go suites and the existing landing evidence. No build binaries are added.
The earlier full bridge PASS 69.592s and seven foundation mutants remain in
WAVE16_LANDING_REPORT.md; bridge implementation is unchanged and that entire
gate was not repeated here. Targeted go vet ./stage1/cohere/typeaware is clean;
git diff --check is clean. The full repository Go gate was not run.

## Native time against Go

benchmark.py runs serially after the oracle gate, alternates native/Go order,
and takes three process-wall samples per set/population. These are count-mode
whole-process medians, including each pipeline's source/program loading and
analysis, with compilation excluded. All thirty count pairs match; this timing
check does not replace the full diagnostic comparisons above. Raw timings,
stdout/stderr and binary hashes are retained. Native is slower in every row.

| Population | Set | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | first | 3.122423 | 0.365894 | 8.53x |
| compiler | followup | 2.002449 | 0.285192 | 7.02x |
| compiler | third | 3.592577 | 1.338157 | 2.68x |
| compiler | fourth | 1.532256 | 0.282076 | 5.43x |
| compiler | fifth | 2.306226 | 0.333608 | 6.91x |
| repository | first | 0.288572 | 0.122554 | 2.35x |
| repository | followup | 0.304080 | 0.121549 | 2.50x |
| repository | third | 0.443908 | 0.181894 | 2.44x |
| repository | fourth | 0.228417 | 0.119111 | 1.92x |
| repository | fifth | 0.351165 | 0.127508 | 2.75x |

The declarations are ignored by today's driver, so these measurements establish
current costs and do not show the proposed dispatch improvement. Original setup
was not repeated: Go 0s, clang 1s, Node 1s, submodules 1s, cache/done 85s. This
continuation sourced /workspace/adamic-tools/env.sh and nproc again printed 5.

## Remaining React claims

set-state-in-effect, set-state-in-render and static-components remain incomplete
source ports. Wave 21 supplies native prepared-HIR validator cores, but still
explicitly lacks native source lowering/SSA, compilation-unit gates and memo
shadowing integration. Duplicating those cores would not finish the missing
source frontend. The independent unchanged-Go static-components message ambiguity
is separately preserved in WAVE16_REACT_REFERENCE_REPORT.md. The missing frontend
and that nondeterministic creator message prevent claiming byte-exact complete
source rules. No further rule was claimed. No emitted-JavaScript rule comparison,
execution of repaired programs or qualifying React native source mutant/handle/
sanitizer/timing result is represented as covered.
