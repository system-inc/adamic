Built: three Nexus rules for collection misuse, discarded outcomes and discarded pure results, in separate .a files.
Commits: claim c835aa6dc806d690a24d5df89bca1696d9cfd40b; implementation cba8883eebe65ae8e3bef3fdb3c54455b00fd1e8, both pushed.
Checks: wave oracle PASS 76.813s; all bridge packages PASS; filtered Node oracle PASS; lint/type/format and Go vet PASS.
Mutants: all three rule mutants and the raw ancestry mutant compile and exit 0, killed only by Go finding bytes; stale-registry mutant killed by panic-70 expectation; foundation mutants below.
Limits: concurrent claim collision recorded; default options, frozen populations and filtered runtime gate; no further claims after Ahra's correction.

# Wave 02 continuation

The original wave-02 three rules were already completed and pushed through
1a75bc1d when the continuation request arrived. The branch remains
`codex/typeaware-wave-02` on base 0d540f413625f016f20fea39761c7b184f335de6.
No history was rewritten and no PR was opened.

After pushing the existing branch, the selection fetch inspected 320 origin refs
and 30 distinct claim blobs. VOLUME_REPORT.md's linked 197 checker-dependent
rule rows were sorted by combined compiler/repository volume descending, full
rule name ascending. Existing main/bridge ports and every name in any origin
claim file were excluded, including incomplete reservations. The first available
three were reserved in c835aa6d, and that commit was pushed before implementation:

* nexus/correctness-no-collection-misuse
* nexus/correctness-no-discarded-outcome
* nexus/correctness-no-discarded-pure-result

A later fetch found overlapping continuation claims on waves 11, 17, 18, 21,
23, 25 and 30. Wave 23's claim commit timestamp is 00:43:48 UTC, one second
before ours at 00:43:49. Commit timestamps do not prove push arrival order;
our selection fetch showed none of these continuation claims. Exclusive
ownership is not claimed. [claim-collisions.json](validation-wave-02-continuation/claim-collisions.json)
records each observed branch and claim commit for Ahra's merge coordination.

Ahra's correction arrived after these three were already reserved and the first
complete comparison/mutant suite had passed. We finished and pushed this reserved
set and did not claim more. The shared registration generator and existing test
harness files were not edited. Each new rule and helper has its own file, with
an isolated wave-owned runner and test file. The raw checker dispatcher gained
one minimal case/return registration, made before the correction. No protected
compiler files, existing Adamic registrations or submodule pins changed.

## Observed exact agreement

| Population | Files | Collection | Outcome | Pure result | Total | Identical bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Controls and imported fixtures | 22 | 36 | 12 | 13 | 61 | 25,803 |
| Frozen repository | 287 | 0 | 0 | 0 | 0 | 18,485 |
| TypeScript compiler | 77 | 0 | 0 | 0 | 0 | 5,318 |

Normal and ASan/UBSan/LeakSanitizer builds match all complete serialized fields,
including message IDs/text, UTF-8 byte ranges, fixes and suggestions. All three
production rules intentionally have no fixes or suggestions, and those empty
fields are also checked. Corpus zeroes alone would be insufficient: all three
have positive controls and compiling mutants killed solely by comparison.

The controls cover four collection-misuse categories, arrays and readonly tuples,
canonical index boundaries, literal/wide/symbol keys, every size comparison
boundary including mirrored and negative/hex/binary/octal/separator literals,
constrained generics, optional receivers, library symbols, project lookalikes,
subclasses and shadowed Object. Pure-result controls cover default-library
String/Array/ReadonlyArray declarations, union receivers, optional calls,
parentheses/chaining, generic type arguments and spreads, literal versus
callable/any/unknown arguments, callback and mutating methods, intentional void
and local interfaces. Outcome controls import all five recognized Nexus aliases,
including generic, parenthesized and optional unions, awaited and unawaited calls,
extra project arms, partial success arms, multi-outcome priority, inferred wrapper
returns, kept results and identical aliases declared outside Nexus. Unicode and
CRLF spans are compared too.

## Raw checker boundary

Only `type-declaration-ancestry` is new. Its Go file exposes raw declarations,
parent chains, union member counts and alias-declared-type links. It contains no
known Nexus filename or alias list and no lint decision. The `.a` consumer
recognizes top-level aliases by exact provenance, groups unique literal arms by
union and tests completeness. See [WAVE_02_CONTINUATION_FACTS.md](WAVE_02_CONTINUATION_FACTS.md).

The direct checker test independently compares every emitted field against the
checker symbol and AST, and checks malformed IDs, exact anchors and absent
symbols. Existing raw/constrained shapes, property shapes, literals, call counts,
symbol detail facts and the previously added awaited-shape handle the other needs.
The Go production oracle imports no bridge code and invokes the pinned rules
unchanged, with its own loader and walk.

## Mutants actually run

| Mutation | Detector and observed result |
| --- | --- |
| Collection size boundary <=0 changed to <0 | Compiles, exit 0, 53 findings; independent Go byte mismatch at 2,937 |
| Outcome complete-arm equality changed to less-than | Compiles, exit 0, 50 findings; Go byte mismatch at 20,722 |
| Pure-result library-provenance guard inverted | Compiles, exit 0, 49 findings; Go byte mismatch at 12,583 |
| Raw ancestry union member count incremented | Compiles, exit 0, 49 findings; Go byte mismatch at 20,722 |
| Released program kept in registry | Compiles, exit 0; required panic 70 is absent, killing the mutant |
| Foundation C input length off by one | ASan heap-buffer-overflow |
| Foundation output string length off by one | ASan heap-buffer-overflow |
| Foundation registry retains released handle | Stale-handle assertion |
| Foundation query uses source-file position | Independent oracle mismatch at byte 6 |
| Foundation link opt-in removed | Refusal test |
| Foundation C output free removed | LeakSanitizer leak report |
| Foundation region entry allocated on heap | LeakSanitizer unowned-result report |
| Node oracle changes one stdout byte | TestTheOracleCatchesOneByte rejects the changed bytes |

All per-rule and new-question mutants finish normally with empty stderr. Neither
compilation nor a sanitizer kills them; the diagnostic comparison is the detector.
The normal released-handle probe primes a live identity, releases the program,
then asks the new question: exit 70, exactly `invalid or released checker handle`.

## Timing observed

Three alternating count-only rounds on each frozen corpus, with no builds or
validation running concurrently. Every Go/native count agrees. Medians, seconds:

| Corpus | Native process | Go process | Native/Go | Native load | Go load | Native run | Go run |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 2.747843 | 0.762055 | 3.61x | 0.241580 | 0.220802 | 2.485982 | 0.516799 |
| Repository | 0.313968 | 0.136907 | 2.29x | 0.061196 | 0.054252 | 0.246000 | 0.071202 |

Native is slower; no speed improvement or performance parity is claimed.
Complete-process time is directly comparable. Go and native phase timers have
different boundaries, and medians of phases need not add to the process median.
All twelve runs and checker timing counters are saved in measurements.json.

## Commands, environment and outputs

All test output was saved to log files, never piped. Setup passed again:
Go 1.27.1 ready (1s), clang 20.1.8 ready (1s), Node v24.19.0 ready (1s),
submodules ready (1s), build cache warm (26s), done in 26s. `nproc`: 5;
cgroup cpu.max: 400000 100000; memory: 17.6 GB. Commands sourced
`/workspace/adamic-tools/env.sh`.

* `go test ./stage1/cohere/typeaware -run '^TestWave02ContinuationAgreementAndMutants$' -count=1 -v -timeout 30m`: PASS 76.813s, with ADAMIC_WAVE02_CONTINUATION artifact and both manifest variables set, plus ADAMIC_TYPESCRIPT_SOURCE.
* `go test -v -count=1 ./bridge/tsgo/checker -run '^TestTypeDeclarationAncestry$'`: PASS 0.047s. All checker tests also pass.
* `TMPDIR=/workspace/wave-02/scratch ADAMIC_TSGO_CORPUS=/workspace/wave-02/typescript go test -v -count=1 -timeout 15m ./bridge/tsgo/...`: PASS; TestBridge 64.86s, checker package 0.178s. Foundation oracle: 1,600 positions, four compiler files, 54,982 identical bytes under ASan/UBSan/LSan. C ABI: 100 queries with output survival, zero/stale handle rejection and nonreused handles.
* `go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures)\.a$'`: PASS 11.698s. Go's component regex also selects generic_functions and method_closures; the log names every fixture.
* Pinned cohere `--no-cache --no-fix` on eight temporary extension/import adapters: exit 0, 276 rules, eight checked, 100% Adamic-ready. Actual native rule sources remain `.a`; no shared harness was changed to obtain this gate.
* `go vet ./...`: exit 0. gofmt lists for the new Go files: empty. `git diff --check`: exit 0.
* `python3 bridge/tsgo/profile/volume_bench.py ... --rounds 3 --corpus compiler ... --corpus repository ...`: twelve successful count-agreeing measurements.

The initial native build exposed unsupported `Number(...)` conversion and an
inferred never[] conditional. The rule uses supported numeric parsing and an
explicit declaration loop instead; no compiler changes were made. An invalid
leading-zero numeric-separator control was corrected after Go's parser rejected
it. These failed runs were not counted as agreement or mutant evidence.

## Limits and artifacts

The full `go test ./...` gate, JSX/JavaScript populations and all upstream options
or fixture matrices were not run. The original three rule suite was not rerun
for this continuation; its earlier passing evidence remains on this branch.
The frozen manifest excludes new ports. No suppression or edit-application engine
is introduced. The three Go rules ignore custom options, so their default
production behavior is the implemented scope.

[validation-wave-02-continuation](validation-wave-02-continuation/README.md)
contains all passing logs, portable manifests, full compressed comparison and
mutant streams, hashes, timing rounds and claim-collision evidence. Both existing
shared harness files and the registration generator remain untouched. This unit
stops after the already-claimed set; no next claim is made.
