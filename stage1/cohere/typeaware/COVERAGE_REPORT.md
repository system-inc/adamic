# Ten inventory-selected bridge rule ports

The new runner is `coverage_suite.ts`. Adamic parses each file with the stage 1
TypeScript parser, selects nodes and applies all ten rule decisions and repairs.
It creates one checker program for the complete root manifest, retains it across
all files and rules, then releases it. Go supplies raw checker facts through the
existing C-archive ABI; it does not call cohere rule implementations on behalf of
native Adamic. No native emitter or lowerer files were changed.

## Selection before implementation

Base branch commit: `d00818a99833ae10fc080d46d968589d68bfef08`. Cohere is
pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`; its typescript-go
checkout is `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.

Inventory: `origin/codex/lint-inventory`, commit
`73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf`. Its `type-aware / binding bridge`
wave contains **204** entries, rather than the 203 in the request. I counted all
204 names, including entries whose production registration does not currently
set `NeedsChecker`. That matters for `method-signature-style`: it is in the
inventory wave, but requires no new checker operation in this port.

`testdata/count_coverage.go` is built as an overlay in cohere and invokes the
unchanged production registry, with its own loader and walk and no bridge import.
The full counts and descending combined-volume ranking are in
[validation-coverage](validation-coverage). The prior sixteen implemented rules
are excluded, leaving 188 candidates. These are the first ten, with lexical ties:

| Rule | Compiler | Repository | Combined |
| --- | ---: | ---: | ---: |
| no-use-before-define | 7,777 | 66 | 7,843 |
| adamic/no-unchecked-cast | 3,585 | 1 | 3,586 |
| @typescript-eslint/method-signature-style | 1,508 | 0 | 1,508 |
| no-implicit-coercion | 924 | 0 | 924 |
| nexus/correctness-no-caller-data-mutation | 760 | 60 | 820 |
| no-param-reassign | 669 | 9 | 678 |
| adamic/invariant-mutable | 575 | 0 | 575 |
| adamic/no-optional-widening | 287 | 0 | 287 |
| nexus/consistency-no-property-alias | 258 | 24 | 282 |
| @typescript-eslint/no-unused-vars | 246 | 20 | 266 |
| **New ten** | **16,589** | **180** | **16,769** |

Next is `nexus/correctness-no-implicit-return`, 251 compiler / 0 repository.

The compiler population is the same 77 `src/compiler` files in TypeScript
v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. The repository
population is the prior frozen manifest: 287 files, 212 `.a` and 75 `.ts`,
excluding the intentional-invalid roots. Selection and agreement use exactly
those roots, not a changing manifest polluted by the ports under construction.
Configured declaration roots are retained; `.a` files are read as TypeScript.
Portable relative manifests are committed with the evidence. The controls compile
and execute the new port sources themselves as native Adamic.

## Checker extension and independence

There is no new C function or changed output-ownership contract. New
`tsgo_inspect` questions expose binding and alias declarations, symbol identities,
name resolution, literal values, global symbol and annotation types, type
identity, property flags/types/lookup, call signature parameters and returns,
contextual arguments, annotated returns, declaration/JSDoc records and inherited
container types. `bridge/tsgo/facts.md` specifies every wire field and
`bridge/tsgo/tsgo.h` states the ownership and borrowed identity lifetimes.

Opaque symbol IDs have their own program-scoped namespace. Private TypeScript
property keys can contain invalid UTF-8: the bridge returns a valid UTF-8 display
name and an opaque symbol ID, and performs property lookup with the original key.
The display string is not silently treated as the original property identity.
`reference-shape` explicitly expands deferred type references; earlier shape
questions keep their previous behavior, protecting existing rules.

`TestCoverageCheckerQuestions` holds the new answers to direct checker operations,
pinned flags and malformed-identity rejection, including Unicode literal framing,
readonly and optional properties, JSDoc, contextual argument types and deferred
`Map<string, State>` arguments. Native rules decode and validate the answers.

The diagnostic oracle, `testdata/oracle_coverage.go`, runs the ten unchanged Go
cohere rules with default options. It imports no bridge code. Comparisons cover
file headings, byte ranges, full rule IDs, message IDs and text, fix counts and
edit ranges/text, suggestions and their descriptions, and the final count. Both
sides sort complete diagnostic strings per file. Equal counts alone are never an
agreement check. The oracle preserves production behavior even when surprising:
for a union target, the optional-widening rule can report the property recorded by
a later alternative with the type pair from an earlier failed alternative; the
native implementation and a targeted control reproduce those exact bytes.

## Rule mutants

Each compiled mutant exits normally with changed diagnostic bytes. The
independent production Go byte oracle catches it; compilation failures are not
counted as mutant kills.

| Rule | Mutation | First differing byte in controls |
| --- | --- | ---: |
| no-use-before-define | Remove the above-declaration check | 68 |
| no-unchecked-cast | Reverse the any/unknown target exemption | 1,604 |
| method-signature-style | Drop method-signature judgments | 2,287 |
| no-implicit-coercion | Replace a literal value with `false` | 1,198 |
| no-caller-data-mutation | Substitute getter flags for setter flags | 10,416 |
| no-param-reassign | Lose the parameter-binding anchor | 380 |
| invariant-mutable | Reverse the assignability question | 12,987 |
| no-optional-widening | Substitute the missing property name | 15,798 |
| no-property-alias | Add one to the diagnostic end position | 16,298 |
| no-unused-vars | Treat an arrow body read as its declaration name | 17,151 |

An initial coercion mutation used a literal constant that TypeScript rejected;
it was revised to a runtime string expression and then caught by the byte
oracle. An initial property-alias getter-flag mutation survived because a second
declaration-kind check made it redundant; the final range mutant above was run
and caught. These initial attempts are not claimed as successful kills.

A released program queried with a new question must panic with exit 70 and
`invalid or released checker handle`. A registry mutant retaining the released
program exits 0 and is caught by that expectation. C-owned buffers and copies are
checked under ASan/UBSan and leak checks; ASan does not instrument the Go heap.

## Toolchain and scope limits

First setup failed while linking stage 0 with `mapping output file failed: no
space left on device`. Removing only identified obsolete bridge scratch binaries
and archives freed space; the unchanged setup then passed. Retry timing lines:
Go 1.27.1 ready 0 s; clang 20.1.8 ready 1 s; Node v24.19.0 ready 1 s;
submodules ready 1 s; build cache warm 15 s; setup done 15 s.
`nproc` reports 5; cgroup quota is 4 CPU cores. Both setup logs are preserved.

This unit validates production default rule options on the two pinned populations
and 35 targeted control sources plus their helper. It does not claim agreement
for every configurable rule option, JSX project, ambient-module export variant,
class-expression inherited mutation contract or every upstream rule fixture.
Those extensions require their own byte oracle coverage. The new runner does not
replace the inventory's independent syntax-only implementation elsewhere. Native
performance remains substantially slower for these new rules; this is a coverage
unit and contains no claim that the earlier whole-process speed gap is solved.


## Validation commands and observed output

All test output was redirected to log files. The targeted gate is used rather
than claiming a full repository gate. Commands below ran from the repository
root after sourcing `/workspace/adamic-tools/env.sh`.

```sh
ADAMIC_COVERAGE_ARTIFACTS=/tmp/tsgo-coverage/controls-final \
ADAMIC_COVERAGE_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest \
ADAMIC_COVERAGE_COMPILER_MANIFEST=/tmp/tsgo-profile/final/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' \
    -count=1 -v > /tmp/tsgo-coverage/coverage-test-final.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v \
    > /tmp/tsgo-coverage/checker-test.log 2>&1

go test ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
    -count=1 -timeout=10m -v > /tmp/tsgo-coverage/node-oracle-test.log 2>&1
```

Initial coverage suite: **PASS**, 501.447 s. Its 34 controls: 52 findings and 23,419 identical
bytes, normal and ASan. Repository: 180 findings and 85,151 identical bytes,
normal and ASan. Compiler: 16,589 findings and 7,120,228 identical bytes, normal
and ASan. Native stderr was empty for every normal/sanitizer comparison.
Ten rule mutants and the released-registry mutant were caught as described above.

Checker package: **PASS**, 0.112 s, all five tests including every new question.
Final review added an explicit `annotation-shape` refusal for a value node. A
mutant disabling that kind check fails with `annotation-shape accepted a value
node`; the normal direct-checker test passes. This is an additional checker
question mutant beyond the ten rule mutants.
Filtered Node oracle: **PASS**, 19.983 s, eight selected fixtures (Go's subtest
filter also selected `method_closures` and `generic_functions`) and the one-byte
mutant. Fixtures additionally check the JavaScript backend and native leak checks.

The full stream SHA-256 values are recorded separately for each Go/native/ASan
run in `validation-coverage/diagnostic-hashes.json`. Compiler streams share
`cb4d875ac853b1e27561cbcce3b68eccc4d0818e701c9ab2a4e3d8494c614ab3`;
repository streams share
`24d3e29e8d5eb9f4c89f03de301aa5e0d443ebc61d99afe1afadc94d1f4cbad3`.


Existing sixteen-rule regression command:

```sh
ADAMIC_VOLUME_ARTIFACTS=/tmp/tsgo-coverage/old16 \
ADAMIC_VOLUME_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
go test ./stage1/cohere/typeaware -run '^TestVolumeAgreementAndMutants$' \
    -count=1 -timeout=30m -v > /tmp/tsgo-coverage/old16-test.log 2>&1
```

**PASS**, 411.081 s. The original sixteen rules retain 14,232 compiler findings
and 46 repository findings, with 6,717,107 and 31,862 identical bytes respectively,
normal and ASan. The controls retain 73 findings / 24,218 identical bytes, normal
and ASan. All fifteen prior checker-question mutants were caught by the production
Go byte oracle: assignable-types, widened-shape, enum-types, type-symbol,
scope-locals, call-returns, property-shape, contextual-shape, symbol-origin,
type-origin, property-info, call-count, call-parameters, apparent-shape and
base-shapes. The original released-handle probe and registry mutant passed their
expectations too. The log records every mutation's first differing byte.


C ownership and compiler-query gate:

```sh
TMPDIR=/workspace/tsgo-coverage-scratch ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript \
go test ./bridge/tsgo/... -count=1 -timeout=15m -v \
    > /tmp/tsgo-coverage/bridge-test.log 2>&1
```

**PASS**, bridge package 59.267 s, checker package 0.148 s. There were 1,600
independently answered positions across `checker.ts`, `parser.ts`, `types.ts` and
`utilities.ts`: 54,982 identical bytes under ASan/UBSan/LeakSanitizer. The C API
ownership checks also pass (100 queries, buffers survive program release, stale
and zero handles rejected, distinct second handle, Unicode output). The original
input-length and output-length +1 mutants both trigger ASan heap-buffer-overflow;
retaining a released handle fails the stale-handle assertion; querying the wrong
position fails the Go byte oracle at byte 6; disabling link opt-in fails the
refusal expectation; omitting C output frees and allocating a region result on
the heap both trigger LeakSanitizer.

The first bridge gate attempt failed during mutant archive construction with
`ar: unable to copy file .../a.out.a; reason: No space left on device`. `/tmp` is a
separate 8.8 GB filesystem, even though the workspace filesystem had room. The
retry above directs transient build/test files to workspace scratch and passes;
no source workaround or skipped mutant was used. Both attempt logs are preserved.


Final refusal review caught an actual native walk defect: `Shadow.name`, reused
as a generic declaration-shape helper, did not select private/computed/literal
property names. `Flow.annotation` consequently treated an unannotated private
field name as an annotation and asked the checker about it. Before the kind
refusal, TypeScript returned its error type, hiding the incorrect question behind
unchanged findings. With the refusal enabled, the repository run panicked on
`internal/oracle/testdata/updates.a`. `Flow.annotation` now selects the complete
property-name grammar and excludes that name before choosing an annotation.
A new control includes unannotated private, computed, string and numeric fields
plus a private annotated array widening that must produce a finding. This
control agrees with Go and is rerun with the complete sanitizer/mutant suite.
The interrupted repository timing is discarded; its stderr is preserved as
`annotation-walk-refusal.stderr`.


Final coverage suite, after the annotation refusal and native field-name fix:

```sh
TMPDIR=/workspace/tsgo-coverage-scratch \
ADAMIC_COVERAGE_ARTIFACTS=/workspace/tsgo-coverage-scratch/final-controls \
ADAMIC_COVERAGE_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest \
ADAMIC_COVERAGE_COMPILER_MANIFEST=/tmp/tsgo-profile/final/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' \
    -count=1 -timeout=30m -v > /tmp/tsgo-coverage/coverage-test-complete.log 2>&1
```

**PASS**, 482.142 s. Final controls: 53 findings / 24,454 identical bytes;
repository: 180 / 85,151 bytes; compiler: 16,589 / 7,120,228 bytes. All three
comparisons pass for normal and ASan/UBSan/leak-check builds, with empty native
stderr. All ten rule mutants exit 0 and are caught by the independent Go byte
oracle at the offsets in the table above. The released program panics 70;
retaining it in the registry makes the mutant exit 0 and fail the required-panic
expectation. `final-coverage-test.log` and the final full-stream hashes are
committed alongside the earlier run evidence.


## Final cost observation

One quiet, optimized run after all verification, using the final suite's normal
binary and independent Go oracle. Both produce complete diagnostic streams to
files, including formatting, sorting, fixes and suggestions; `cmp` still passes.
Each creates one program, then runs all ten rules over its complete manifest.
Compilation and sanitizer runs are outside these intervals. No competing test or
build ran during the measurement. These are observations, not benchmark medians.

| Corpus | Native load | Native run | Native load + run | Go load | Go run | Go load + run | Native adapter queries | Mean adapter query | Native / Go findings per second |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| compiler | 0.309 s | 68.118 s | 68.427 s | 0.346 s | 7.422 s | 7.768 s | 2,222,043 | 6.602 µs | 243.5 / 2235.2 |
| repository | 0.098 s | 1.489 s | 1.586 s | 0.105 s | 0.288 s | 0.393 s | 72,186 | 6.916 µs | 120.9 / 625.8 |

`run_ns` spans successful create to release; load + run excludes process startup
and teardown. Query time includes Go checker work, serialization, C transfers,
UTF-8 decoding and frees; it is not a measurement of bare cgo crossing latency.
The compiler's summed native query intervals are 14.670 s within a 68.118 s run.
The remaining interval includes native walks, fact processing, allocation,
diagnostic rendering and output; this unit did not profile that remainder.
Go's production rule walk invokes checker operations directly, so its callback
interval is not normalized by the native adapter's different question count.

Exact nanoseconds, query counts, stream bytes and hashes are in
`validation-coverage/measurements.json`; raw timing lines are adjacent. The
compiler output remains 7,120,228 identical bytes and the repository output
85,151 identical bytes. The private-field correction removes three unnecessary
repository queries (72,189 to 72,186) and changes no finding.

The timed command pairs were:

```sh
/workspace/tsgo-coverage-scratch/final-controls/coverage-oracle \
    /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest \
    > /tmp/tsgo-coverage/final-compiler-go.stdout 2> /tmp/tsgo-coverage/final-compiler-go.stderr
ADAMIC_TSGO_TIMING=1 /workspace/tsgo-coverage-scratch/final-controls/coverage \
    /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest \
    > /tmp/tsgo-coverage/final-compiler-native.stdout 2> /tmp/tsgo-coverage/final-compiler-native.stderr
cmp /tmp/tsgo-coverage/final-compiler-go.stdout /tmp/tsgo-coverage/final-compiler-native.stdout \
    > /tmp/tsgo-coverage/final-compiler-cmp.log 2>&1
```

Repository repeats the same pair with `/workspace/adamic/tsconfig.json` and
`/tmp/tsgo-volume/repository.manifest`; both comparisons exit 0 with empty cmp
logs. The committed branch is `codex/tsgo-c-library`; no pull request is opened.
