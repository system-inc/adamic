Rebased wave-16 onto current origin/main f8013f0ba; added fifteen numeric rule.json manifests in owned directories.
Source commits: rebased implementation aeb6039ad2babd4434b8338e2e81d2406c85c500; manifest/validation tools f37bcc46a30908791317119b7de7e416fa1ebaa8.
Commands/output: five oracle suites PASS 425.666s; bridge PASS 66.168s; numeric probes and 15 manifests pass; 60 timing processes pass; vet/gofmt clean.
Mutants: 15 semantic, 45 JSON metadata, 15 declaration, one compiled numeric, five retained-registry and seven bridge foundation mutations caught.
Uncovered: native React source-to-HIR/SSA integration, node-local numeric dispatch and ordinary Nexus dispatch; no new claims, main/area pushes or shared edits.

# Landing on f8013f0ba

This unit has pushed only codex/typeaware-wave-16. Its old published tip was
50236940095db0a09b2b6bf464c9147ac0185703. An all-head fetch found main advanced
from e8ba3d5d to f8013f0baac41ddc340d76f83bddde38536a8f07. Rebase replayed all
25 owned commits without conflicts. A later main fetch confirmed that same base.
The final push uses an exact lease on the old own-branch tip, as required for the
explicitly requested rebase of a published branch. No push targets main or area/.

The rebase brings main's checking/runtime fixes into the compiler used by the
native gates. No owned rule behavior was changed. All fifteen numeric manifests
live in wave16_listeners/<module>/rule.json and contain the production Go name,
relative .a module path and numeric kinds. This follows the type-aware metadata
shape inspected on wave 04. The validator regenerates independent Go enum/listener
expectations using the existing owned checker, then compares the module export
and each JSON manifest. Name, module and kind mutations are each rejected for
each manifest, totaling 45 JSON-only witnesses. The existing fifteen declaration
mutations are rerun as well. These are metadata witnesses, not semantic rule mutants.

No shared loader, generator or test harness was edited. Current main's ParseNode
still exposes kind as a string, without a numeric field or supplied-node rule
interface. Legacy bodies still scan/refetch nodes and use their old string
predicates. The manifests do not claim that numeric node-local execution is done
or that the current driver consumes these declarations. New source rules are not
being added under an invented API. The shared Diagnostic landing SHA has not been
named in this message; no unrelated branch was cherry-picked as its substitute.

## Commands and results

All test output was redirected directly to logs in
/workspace/wave16-artifacts/landing-f801, copied into the committed evidence.

```
source /workspace/adamic-tools/env.sh
export ADAMIC_WAVE16_ARTIFACTS=/workspace/wave16-artifacts/landing-f801/first
export ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS=/workspace/wave16-artifacts/landing-f801/followup
export ADAMIC_WAVE16_THIRD_ARTIFACTS=/workspace/wave16-artifacts/landing-f801/third
export ADAMIC_WAVE16_FOURTH_ARTIFACTS=/workspace/wave16-artifacts/landing-f801/fourth
export ADAMIC_WAVE16_FIFTH_ARTIFACTS=/workspace/wave16-artifacts/landing-f801/fifth
export ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest
export ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript
go test ./stage1/cohere/typeaware -run '^TestWave16.*AgreementAndMutants$' -count=1 -v -timeout 20m > /workspace/wave16-artifacts/landing-f801/oracles.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /workspace/wave16-artifacts/landing-f801/bridge.log 2>&1
python3 stage1/cohere/typeaware/validate_wave16_rule_json.py /workspace/wave16-artifacts/landing-f801 > /workspace/wave16-artifacts/landing-f801/manifests.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /workspace/wave16-artifacts/landing-f801/vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave16*test.go stage1/cohere/typeaware/testdata/oracle_wave16*.go > /workspace/wave16-artifacts/landing-f801/gofmt.log
```

The five oracle suites pass in 425.666s. Normal and ASan/UBSan/LeakSanitizer runs
match complete findings, fixes and suggestions on all controls, compiler77 and
repository287. Control counts remain 65/22/50/33/231 for original/follow-up/third/
fourth/fifth. The original set has two repository findings and one compiler
finding; the others have zero. Fifth-set option profiles retain 225/243/235
findings. Every released-handle probe requires panic 70; each of five retained-
registry mutants exits zero and is caught. Nexus suites still use the supplied
shared-question dispatch build overlay, not ordinary integrated registration.

All fifteen semantic mutants compile, exit zero and leave stderr empty. Only
independent Go diagnostic bytes catch them:

| Mutant | First differing byte | Findings |
| --- | ---: | ---: |
| throw-range | 130 | 231 |
| backreference-range | 9283 | 231 |
| arrow-fix | 6056 | 231 |
| listener-general | 723 | 23 |
| render-number | 7165 | 14 |
| mock-range | 3576 | 22 |
| function-range | 132 | 33 |
| native-range | 4598 | 33 |
| wrapper-range | 5183 | 33 |
| import-write | 66 | 59 |
| exponent-base | 15678 | 65 |
| nan-comma | 12079 | 63 |
| shell-heredoc | 8299 | 53 |
| sql-string | 11784 | 31 |
| alert-range | 130 | 50 |

The bridge package passes in 66.168s and checker package in 0.186s. Foundation
coverage includes 100 C ABI queries and 162 source positions / 3261 identical
bytes under sanitizers. Seven foundation mutants are caught: input/output lengths
by ASan, retained stale handle by its assertion, source-file type position by Go
byte 6, removed link guard by refusal, and removed output free / heap-region
allocation by LeakSanitizer. Independent raw rule-question tests also pass.

The numeric probe was rebuilt using this base's stage 0 and both checker archives.
Normal and sanitized complete stdout match the independent Go numeric values,
with empty stderr. The existing no-alert numeric-kind mutation compiles and exits
zero with empty stderr, and the comparison catches byte 254. The mutation is
metadata-only, additional to the qualifying semantic rule mutant.

308 new oracle streams are retained compressed losslessly or as stderr, together
with logs, fresh timing output, manifest expectations, source hashes and the
refreshed dependency inventory. Native/build binaries are excluded. Vet, gofmt
and whitespace logs are empty. No full repository gate or emitted-JavaScript
rule comparison was run in this landing pass. Sanitizers cover native/C code,
not the Go heap.

## Native versus Go time

benchmark.py runs after both gates, serially, with three samples per population/
set and alternating native/Go order. The table gives process-wall medians in
count mode, including program loading and analysis but excluding compilation.
All 30 count pairs match; strict finding/fix/suggestion agreement comes from the
separate oracle gate above. Timing samples and binary hashes are retained.

```
python3 stage1/cohere/typeaware/validation-wave16-listeners/benchmark.py /workspace/wave16-artifacts/landing-f801 /workspace/wave16-artifacts/landing-f801/timing > /workspace/wave16-artifacts/landing-f801/timing.log 2>&1
```

| Population | Set | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | first | 3.137812 | 0.352900 | 8.89x |
| compiler | followup | 2.248557 | 0.292663 | 7.68x |
| compiler | third | 3.788573 | 1.353641 | 2.80x |
| compiler | fourth | 1.552722 | 0.292772 | 5.30x |
| compiler | fifth | 2.382991 | 0.347649 | 6.85x |
| repository | first | 0.297812 | 0.130996 | 2.27x |
| repository | followup | 0.319797 | 0.122681 | 2.61x |
| repository | third | 0.509335 | 0.201616 | 2.53x |
| repository | fourth | 0.278161 | 0.183914 | 1.51x |
| repository | fifth | 0.358146 | 0.141028 | 2.54x |

Native remains slower in each row. These manifests are ignored by today's driver;
no dispatch speedup is claimed. Toolchain setup was retained, not repeated:
original Go 0s, clang 1s, Node 1s, submodules 1s, cache/done 85s. Environment was
sourced from /workspace/adamic-tools/env.sh; nproc again reports 5.

## Outstanding claims

The retained three React source claims are set-state-in-effect, set-state-in-render
and static-components. Wave 21 has prepared-HIR native validator cores, but its
current report still explicitly lacks source lowering/SSA, exact compilation-unit
gates and memo-shadowing integration. Wave 04's partial kernels also continue to
refuse full source execution. The source adapter is not supplied by .a loading,
profile compilation or suggestion serialization. Duplicating prepared-HIR cores
would not complete that adapter. The separately reproduced static-components
creation-site message nondeterminism also remains in WAVE16_REACT_REFERENCE_REPORT.md.

The fresh inventory covers 507 origin heads and finds 0 declarations of the five named native entry points.
That name-based inventory is evidence, not proof about differently named
implementations; current dependency reports were inspected separately. All three
claims remain incomplete, with no qualifying native source mutant/sanitizer/
handle/timing result asserted for them. No next set was claimed. This unit stops
at the documented prerequisite boundaries rather than editing shared files.
