# Wave 14 third batch

This is a partial port, with the supported comparisons passing. It does not
claim three universally complete rules. The preceding six implementations were
pushed at 5bd31ee5; the real JSX gap remains documented in WAVE_14_RENDER_REPORT.md.

Claim 350d4776c60ef9a70dc33a81b14c890936cd02d0 was pushed before code. After
fetching 335 origin refs, the first three unclaimed rows in VOLUME_REPORT.md's
combined compiler/repository ranking were no-invalid-regexp, no-label-var and
no-misleading-character-class, each with zero findings. Selection excluded the
26 base ports and reservations on every origin branch, including incomplete
reservations. There were 58 eligible rows before this claim. No further batch
is claimed while these gaps remain.

## Implemented behavior

Each rule has its own .a file and uses the existing canonical diagnostic
protocol, including every repair and suggestion field:

- no-label-var asks for value symbols at the exact labeled statement and matches
  the name natively. Parameters, outer/local bindings, functions, classes,
  hoisting, builtins, block exclusions and Unicode spans have positive/negative
  controls. The undefined-label frontend case below is still blocked.
- no-invalid-regexp selects the global constructor from raw declaration-file
  facts, handles call/new argument selection, validates flags with Go's duplicate
  and u/v precedence, and renders Go-compatible rune quotations. Its 741 Unicode
  printable ranges come from pinned Go 1.27.1 strconv.IsPrint; the generator is
  testdata/regexp_printable_wave_14.go. Pattern validation is an explicit gap.
- no-misleading-character-class implements literal class scanning, escapes,
  ranges, nested v classes, breakers, Unicode/non-Unicode surrogate treatment,
  all six message IDs, byte spans and proposed Unicode-flag edits. Whole-pattern
  class parse failure discards accumulated findings. Constructor tracking and
  cooked-to-raw mappings are explicit gaps.

wave_14_third.a runs all three with one program and one native parse per file.
The --class-only flag isolates the constructor-gap witness from the preceding
pattern validator's refusal; it is not used for corpus agreement or timing.
The independent Go oracle invokes the three unchanged production rules with
their default options, its own loader, AST walk, program views and file cache.
It imports no bridge code and copies no rule predicate.

The only checker extension is the raw scope-symbols question, in its own Go and
Adamic files, with one shared switch registration. No shared generator, test
harness, parser, protected compiler file or production rule registration changed.

## Demonstrated blockers

Every boundary has an independent Go-positive .a witness and a native exit-70
refusal, retained in validation-wave-14-third:

| Witness | Go finding | Native boundary |
| --- | --- | --- |
| `undefined: for(;;) {break undefined;}` | identifierClashWithLabel | Shared parser expects a semicolon at the colon instead of recognizing this label |
| `new RegExp('[')` | invalidRegexp | Native ECMAScript rewrite/regexp2-compatible pattern validation and errors are missing |
| `new RegExp('[Á]')` | combiningClass | Native constructor reference tracking and cooked-to-raw source mapping are missing |
| `RegExp(p, '\ud800')` | invalidRegexp | Go/native lone-surrogate string decoding is not held yet, so the native rule refuses |

The constructor probe uses --class-only. The regex dependencies and Unicode
decoder boundary are not fixed by .a module loading in the shared harness.
Only the undefined-label boundary is the demonstrated shared-parser blocker
in this batch. These paths are not counted as passing byte comparisons.
Constructor aliases, globalThis tracking, arbitrary constant-flow cases and
rule options beyond defaults remain outside this partial slice.

## Validation

Toolchain setup succeeded: Go, clang, Node and submodules ready at 0 s;
build cache warm 22 s; setup done 22 s. nproc is 5; cgroup quota is four cores.
Go is 1.27.1, clang 20.1.8 and Node v24.19.0. Environment:
source /workspace/adamic-tools/env.sh. Setup and all test output went to files.

The final targeted rule command is:

```sh
ADAMIC_WAVE14_THIRD_ARTIFACTS=/workspace/wave-14-third-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript \
ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest \
ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest \
go test ./stage1/cohere/typeaware \
  -run '^TestWave14ThirdAgreementAndMutants$' -count=1 -v -timeout=30m \
  > /tmp/wave-14-third-final.log 2>&1
```

There are 51 supported controls, 49 complete findings and 17,789 identical bytes,
normal and ASan/UBSan/LeakSanitizer, with empty native stderr. Controls include
all six literal IDs and nonempty suggestions, flags precedence, supplementary
Unicode, nonprintable BMP/private-use/noncharacter flag quoting, escaped
surrogates, ranges, nested classes and join chains. The Go-positive boundary
witnesses are additional inputs, not part of this supported control count.

The pinned compiler population is 77 src/compiler roots in TypeScript v6.0.3
at 050880ce59e30b356b686bd3144efe24f875ebc8. The repository is the same frozen
287-file population used by the base branch. Both have zero findings in these
three rules. All file headings, final counts and complete streams agree:
compiler 5,318 bytes, repository 18,485 bytes, normal and all three sanitizers.
This zero is backed by the positive controls, rather than treated as sufficient
evidence on its own. Source hashes and complete compressed streams are retained.

Every rule mutant compiles, exits 0 and has empty stderr. The full Go byte
comparison alone catches the inverted label-name test, lost duplicate-flag
precedence and inverted combining-mark detector. Additional mutants invert
Unicode rune quoting and change the raw scope meaning from Value to Variable.
Final rule suite PASS, 89.340 s. Byte offsets: label 50, flags 5076, unicode-quote 5667, class 6811, scope meaning 1338. A stale program
panics 70 with invalid or released checker handle; retaining it in the registry
exits 0 and fails that expectation.

```sh
TMPDIR=/workspace/wave-14-third-bridge-scratch \
ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript \
go test ./bridge/tsgo/... -count=1 -v -timeout=15m \
  > /tmp/wave-14-third-bridge.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v -timeout=15m \
  > /tmp/wave-14-third-checker.log 2>&1

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -v -timeout=10m > /tmp/wave-14-third-node.log 2>&1

go vet ./... > /tmp/wave-14-third-vet.log 2>&1
```

Bridge PASS, 71.938 s; checker PASS, 0.252 s in that run. The bridge covers
100 ownership queries and 1,600 independently answered positions across four
compiler files, with 54,982 identical bytes under ASan/UBSan/LeakSanitizer.
Input/output length +1 mutants trigger ASan; wrong position differs at byte 6;
retaining a stale handle fails its assertion; removing link opt-in fails the
refusal expectation; omitting output frees or giving a region result heap
ownership triggers LeakSanitizer. Filtered Node oracle PASS, 15.903 s: eight
fixtures, including generic_functions and method_closures selected by Go's
subtest matching, and the one-byte mutant. The new rule programs themselves
were compared natively to Go; they were not compared as emitted JavaScript.

## Timing and limits

Quiet, alternating three-round whole-process medians and raw load/query/run
timings are in validation-wave-14-third/benchmark. Both pinned corpora and the
49-finding positive controls are measured with --count after verification.
The final timing table is recorded in benchmark/medians.json.

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.532114 s | 0.321177 s | 4.77x |
| repository | 0.209672 s | 0.112689 s | 1.86x |
| controls | 0.020854 s | 0.028292 s | 0.74x |

The compiler and repository runs have zero findings; the controls have 49.
These measure this supported slice, not throughput through the refused paths.
Benchmark timing does not include compilation or sanitizer execution.

The first benchmark attempt did not execute because automatic permission review
timed out. The authorized local-only command succeeded on its single retry;
that timeout was not a source failure or an unsafe-action finding.

The complete repository gate, every upstream fixture/options matrix, new-rule
JavaScript emission, real JSX and the four refused boundaries above are not
covered. The regex rules are partial native dependencies, not complete ports
hidden behind zero corpus counts. Prior wave reports retain their original
evidence; these new files are not registered into the shared production suite.
