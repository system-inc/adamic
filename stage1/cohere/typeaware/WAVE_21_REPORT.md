Built one of three assigned rules: native @typescript-eslint/no-mixed-enums.
Commits: claim 94246497, verified enum implementation 63746b1e; branch codex/typeaware-wave-21.
Commands: TestWave21Enums PASS 43.024s, checker PASS 0.203s, filtered Node oracle PASS 15.017s.
Mutants: reversed enum comparison caught at byte 47; retained released handle caught by required exit 70.
Not covered: no-misused-promises and no-misused-spread are unimplemented; this wave is incomplete.

## Selection and status

Base: origin/codex/tsgo-c-library at 0d540f413625f016f20fea39761c7b184f335de6.
The checkout initially fetched only main. Fetching all origin heads exposed the
specified base and allowed inspecting all fetched branches for claim files and
named ports. No matching claims or named ports were found before the claim.
Claim 94246497 was pushed before implementation began.

VOLUME_REPORT.md links the compiler-all.counts and repository-all.counts tables.
Sort their summed counts descending, break ties lexically by full rule name,
then exclude the rules already implemented on the base. Remaining positions
61, 62 and 63 are no-misused-promises, no-misused-spread and no-mixed-enums,
respectively. All three have combined volume zero. The method-signature-style
port is one of the base's 26 ports but is absent from the checker-dependent
count table, so excluding it does not remove another table row.

Only no-mixed-enums was completed. The other two reservations are explicitly
released in claims/wave-21.md. There is no claim that either is impossible,
implemented, or verified. No stub judgments are substituted for those rules.

## Implementation

no_mixed_enums.a owns literal classification, the aborting unknown case,
merged reference selection, earliest sibling ordering, cross-file precedence,
member scanning, diagnostic message and inner-parenthesis report spans.
The rule has no fixes or suggestions. Its runner is wave_21_suite.a.

The isolated enum-declarations question lives in
bridge/tsgo/checker/enum_declarations.go and enum_declarations.a. It returns
both named and local symbol declaration lists, declaration kinds/files/spans,
member counts, first initializer kinds, and aggregated raw union type flags.
Adamic selects the symbol list and reference and makes every rule decision.
The shared facts.go change is only a switch dispatch arm (two gofmt lines).
No protected compiler files or submodule pins changed. New Adamic sources
and generated Adamic controls use .a.

The independent Go oracle loads its own program, walks its own AST, and invokes
the unchanged registered production rule. It imports no bridge implementation.
It preserves default options, declaration roots, complete message bytes, ranges,
fixes and suggestions in the established canonical protocol.

## Observations

| Population | Files | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Targeted controls, including helper/augmentation | 17 | 9 | 4163 |
| TypeScript compiler | 77 | 0 | 5087 |
| Frozen repository | 287 | 0 | 18485 |

Every row agrees in normal and ASan/UBSan/LeakSanitizer builds, with empty native
stderr. Controls cover merged enums, empty merges, exported and nonexported
namespace declarations, imported augmentation, implicit numbers, nonliteral
initializer types, templates, boolean/null aborts, parentheses and Unicode/CRLF.
The TypeScript source is commit 050880ce59e30b356b686bd3144efe24f875ebc8.
The manifests are the existing validation-coverage/compiler.manifest and
repository.manifest, made absolute for this workspace. Full stream hashes,
compressed canonical streams and logs are in validation-wave-21.

The initial control run failed with a nil-pointer panic in the raw query,
which called SkipParentheses on a missing initializer. The missing-initializer
case is now checked explicitly. The initial failure log is preserved, and the
complete corrected suite passed afterward.

Two mutants run against the corrected implementation:

- Replace current !== desired with current === desired. The native mutant
  compiles and exits 0 with empty stderr. Only the independent Go comparison
  catches the changed bytes, first difference at byte 47.
- Retain a released program in the registry. The normal enum-declarations query
  after release panics 70 with invalid or released checker handle. The mutant
  exits 0, failing the required-panic expectation. This tests the handle boundary.

The filtered external Node oracle also kills its existing one-byte mutant and
passes all eight selected native/JavaScript/source fixtures with sanitizer and
leak checks. Checker package regression tests pass. Vet and formatting logs are
empty. The full repository gate was not run.

## Toolchain and commands

bash cloud/setup.sh passed. Timing lines: go ready 0s; clang ready 0s;
node ready 0s; submodules ready 0s; build cache warm 81s; done 81s.
nproc: 5; cgroup cpu.max: 400000 100000; memory: 17.6 GB.
Go 1.27.1, clang 20.1.8, Node v24.19.0. Shells sourced
/workspace/adamic-tools/env.sh. Every test output went to a log file.

```sh
ADAMIC_WAVE21_ARTIFACTS=/workspace/wave21-artifacts \
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json \
ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest \
ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave21Enums$' \
    -count=1 -timeout=30m -v > /tmp/wave-21-final.log 2>&1

go test ./bridge/tsgo/checker -count=1 -timeout=10m \
    > /tmp/wave21-checker.log 2>&1

go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave21-vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
    -count=1 -timeout=10m -v > /tmp/wave21-node-oracle.log 2>&1
```

## Cost

One quiet optimized run per implementation/corpus, after builds and verification.
These are single observations, not benchmark medians. Each emits its complete
canonical stream to files; the streams are compared byte for byte. Load and run
come from each implementation's own timing instrumentation; process time uses
the same monotonic Python timer. Exact values and hashes are in measurements.json.

| Corpus | Native load | Native run | Native process | Go load | Go run | Go process |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 0.227s | 1.135s | 1.375s | 0.246s | 0.045s | 0.310s |
| Repository | 0.067s | 0.121s | 0.193s | 0.066s | 0.040s | 0.117s |

Native compiler: 1229 queries, 0.171s summed query intervals. Repository:
zero queries because it contains no enum anchors. Whole native process is
4.43 times Go on compiler and 1.64 times Go on repository. The corpus findings
are zero, so findings per second would not be a meaningful throughput measure.
This port does not close the native performance gap.

Timing commands ran the same optimized binaries created by the test:

```sh
/workspace/wave21-artifacts/wave21-oracle CONFIG MANIFEST > go.stdout 2> go.stderr
ADAMIC_TSGO_TIMING=1 /workspace/wave21-artifacts/wave21 CONFIG MANIFEST \
    > native.stdout 2> native.stderr
cmp go.stdout native.stdout
```

No JSX corpus or complete upstream fixture/options matrix was covered. The
compiler/repository population alone cannot prove positive behavior for this
zero-volume rule; the controls and normally exiting comparison mutant provide
that evidence. The two other assigned rules remain unfinished and unverified.
