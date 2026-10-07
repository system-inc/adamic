Built: non-JSX `@next/next/no-async-client-component`; both Head rules remain unimplemented.<br>
Commits: claim `808be7f8`, implementation `dd5481bcb611867890b937f32a15f57b9949b2a7`.<br>
Commands/output: final wave tests PASS 56.306s; checker PASS 0.152s; filtered Node oracle PASS 13.139s.<br>
Mutants: async range +1 and uppercase boundary caught only by byte comparisons; retained released handle caught by required panic.<br>
Not covered: JSX rule ports, full repository gate, full pinned cohere CLI lint on `.a`.

# Type-aware wave 17: partial implementation

This is not three completed ports. The non-JSX async rule is implemented and
validated. `no-duplicate-head` and `no-script-component-in-head` have no native
implementation in this unit. Positive production Go witnesses demonstrate that
the shared native parser cannot supply their JSX nodes. An async component with
JSX has the same limitation. Fixing this requires JSX parser work outside this
unit's rule and bridge-question files. No shared parser, bridge, emitter,
lowerer or protected oracle file was changed. No checker question was added.

## Selection and claim

Base: `origin/codex/tsgo-c-library`,
`0d540f413625f016f20fea39761c7b184f335de6`. Branch:
`codex/typeaware-wave-17`. The clone initially fetched only main; all origin
branch tips were fetched explicitly before selection and scanned for claims and
implementations. None of these rules was already claimed or ported. The existing
nextjs syntax claim names different rules. Claim `808be7f8` was pushed before
any implementation file was written. No pull request was opened.

The ranking is reconstructed from VOLUME_REPORT.md's linked
`validation-volume/compiler-all.counts` plus `repository-all.counts`, descending
combined findings and ascending rule name for ties, excluding the base's 26
implemented names. Its 197 checker-dependent entries leave 172 candidates:
method-signature-style is already ported but absent from that checker-only table.

| Remaining position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 49 | @next/next/no-async-client-component | 0 | 0 |
| 50 | @next/next/no-duplicate-head | 0 | 0 |
| 51 | @next/next/no-script-component-in-head | 0 | 0 |

## What runs natively

`wave_17_suite.a` loads one checker program, parses each file with the existing
Adamic parser, indexes parents and bindings, runs `NoAsyncClientComponent`, emits
complete canonical diagnostics, then releases the program. Decisions, ranges and
messages are Adamic code in `no_async_client_component.a`. The existing
`binding-declarations` question supplies raw declarations without following
aliases. Adamic selects the lowest source position, preserving Go's merged
binding behavior, and recognizes only direct async default functions or indirect
async function declarations and async arrow initializers. It preserves directive
prologue position and the parenthesized-form exemptions.

`unicode_upper.a` holds the pinned Go toolchain's 152 `unicode.Upper` ranges.
It distinguishes uppercase from titlecase and supports supplementary characters.
An exhaustive native sanitized run agrees with direct Go `unicode.IsUpper` over
all 1,114,112 code points, producing 10,634 identical bytes. All new Adamic source
and generated Adamic probes are `.a` files. JSX witness inputs are `.tsx` test
source, not Adamic implementations.

The independent Go oracle, `testdata/oracle_wave_17.go`, loads its own program and
walks its own AST, invoking all three unmodified production rules through the
registry. It imports no bridge implementation. It serializes every finding,
fix and suggestion field using the existing diagnostic protocol. These three
production rules have no fixes or suggestions.

## Findings and positive controls

| Population | Roots | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| Non-JSX upstream and targeted controls | 38 | 19 | 9,588 |
| Frozen repository population | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,318 |

TypeScript is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
Cohere stays pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`.
The two manifests are the exact relative root lists in validation-coverage,
resolved against this workspace and the pinned compiler checkout. Configured
declaration roots and non-TS extension loading are retained on both sides.
The frozen repository roots exclude this unit's additions. Controls extract
verbatim non-JSX source rows from the production async rule's Go tests and add
Unicode/CRLF, supplementary capitals/lowercase, titlecase, merged declaration
order and export-assignment controls. All normal/sanitized agreement runs have
empty native stderr. Sanitizers cover C/native ownership, not Go's heap.

Zero corpus findings do not prove either missing Head port. Three separate
positive `.tsx` witnesses each produce one Go finding. Native exits 70:

| Witness | Native parser observation |
| --- | --- |
| Duplicate imported Head | expected GreaterThanToken, got SlashToken at 58 |
| Script directly inside imported Head | expected GreaterThanToken, got SlashToken at 61 |
| Async client component returning a JSX fragment | expected GreaterThanToken, got CloseBraceToken at 69 |

The boundary test expects those failures. Its PASS establishes explicit refusal,
not JSX rule parity. The Head rules have no compiled rule mutants because they
were not implemented. No silent no-op implementation is presented as a port.

## Mutants and released handles

| Mutant | Observation | What catches it |
| --- | --- | --- |
| Async finding end +1 | exit 0, empty sanitizer stderr | production Go complete finding bytes, byte 52 |
| Uppercase first range excludes U+0041 | exit 0, empty sanitizer stderr | exhaustive direct Go classification bytes, byte 1 |
| Keep a released program in the registry | exit 0, empty stderr | expected native panic 70 with invalid or released checker handle |

The normal released-program probe queries the actual question this port uses,
`binding-declarations`, and panics 70 with
`adamic: panic: invalid or released checker handle`. The registry mutant uses a
Go overlay in scratch and changes no shared repository source.
The filtered Node oracle also runs its existing one-byte mutant successfully.

## Commands and gate

Setup succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s,
build cache warm 79s, done 79s. `nproc` is 5; cgroup cpu.max is 400000 100000.
Go 1.27.1, clang 20.1.8, Node v24.19.0. Commands source
`/workspace/adamic-tools/env.sh`. Setup and all test output go to log files.
The full repository gate was not run. Actual commands:

```sh
bash cloud/setup.sh > /workspace/wave-17-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_ARTIFACTS=/workspace/wave-17-artifacts \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript \
TMPDIR=/workspace/wave-17-artifacts \
go test ./stage1/cohere/typeaware \
  -run '^TestWave17(AgreementMutantAndJSXBoundary|UnicodeUpper)$' \
  -count=1 -v -timeout=30m > /workspace/wave-17-artifacts/test-final.log 2>&1
# PASS 56.306s. Agreement/boundary test 49.68s; exhaustive Unicode test 6.62s.
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker \
  > /workspace/wave-17-artifacts/vet-final.log 2>&1
# Exit 0, empty log.
go test ./bridge/tsgo/checker -count=1 -v \
  > /workspace/wave-17-artifacts/checker-test.log 2>&1
# PASS 0.152s.
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions)\.a$' \
  -count=1 -v -timeout=10m > /workspace/wave-17-artifacts/node-oracle-test.log 2>&1
# PASS 13.139s; the filter also selects method_closures.
```

The first agreement attempt failed with `panic: invalid tsconfig`: its scratch
configuration had no conventional root before loading the `.a` manifest. The
harness now creates and configures `config-root.d.ts`, avoiding that empty-input
configuration. The corrected complete run passed, then the final formatted
sources were rerun with all comparisons and mutants.

The pinned cohere CLI's explicit `.a` lint invocation fails with `nothing to
check` and `not a TypeScript or JavaScript file`. This is a missing measurement,
not a clean lint result. Its `.a` formatting path also silently leaves the text
unchanged; that output is not counted as verification. Formatting was instead
obtained with `--fix --stdin-filepath` using virtual `.ts` paths, inspected as
formatting-only diffs and applied with patch to the real `.a` sources. No `.ts`
file was written. Native compilation and the final tests accept the resulting
`.a` files. Full cohere CLI lint of these sources remains uncovered.

## Quiet native versus Go measurements

Three interleaved rounds after all builds/tests finished; no other build or test
ran during these rounds. Both binaries emit complete streams to files and each
round's SHA-256 values agree. These are whole-process medians, including program
load, parsing, formatting/output and teardown.

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.606s | 0.303s | 5.30x |
| Repository | 0.224s | 0.118s | 1.90x |

The corpus native runs make zero adapter queries because none of these files
reaches a relevant indirect client export. These measurements describe the
loading/parsing/traversal cost on these populations, not nonzero checker-query
throughput. They do not imply a speed result for either unimplemented Head rule.
Per-round phases, times and hashes are in
[measurements.json](validation-wave-17/measurements.json). Timing command pairs:

```sh
/workspace/wave-17-artifacts/wave17-oracle CONFIG MANIFEST > GO.stdout 2> GO.stderr
ADAMIC_TSGO_TIMING=1 /workspace/wave-17-artifacts/wave17 CONFIG MANIFEST \
  > NATIVE.stdout 2> NATIVE.stderr
```

CONFIG is respectively the pinned compiler's src/compiler/tsconfig.json and the
repository tsconfig.json; MANIFEST is the corresponding artifact manifest.
[Evidence](validation-wave-17) preserves final test/regression logs, the initial
configuration failure, CLI lint failure, compressed complete finding streams,
positive JSX witness outputs/refusals, source hashes and raw quiet timing lines.
Absolute output headers describe this workspace; relocation changes those bytes.
No suppressions or edits are applied by the differential runners.
