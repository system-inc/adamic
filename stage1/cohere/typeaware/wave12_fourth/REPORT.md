Built: prefer-regex-literals, prefer-rest-params, and react-hooks/exhaustive-deps with production default options; twelve wave 12 rules implemented.
Commits: claim 99abb238 was pushed before implementation 974dc275; evidence follows separately.
Checks: 703 controls, 77 compiler roots and 287 repository roots match Go byte for byte, normal and sanitized.
Mutants: three rule mutants, declaration-file-bit mutation, released-registry retention, JSX guard reversal and Node one-byte mutation were caught.
Uncovered: JSX, nondefault rule options, full gate, prior 26-rule regression suite and emitted-JavaScript lint-suite comparison.

## Selection and ownership

Nine earlier rules were completed and pushed before this claim. Fetching all origin heads inspected 356 refs, 133 claimed ranking names and 25 checker-dependent ports on main/the bridge branch. The additional inventory port is not checker-dependent. The combined VOLUME_REPORT count ranking had 39 remaining rules; these were its first three, each with zero compiler/repository counts. Literal claim bodies and exact matching module filenames on all origin refs had no competing claim or port. Claim 99abb238 was pushed before code.

All native modules are .a inside this directory. Shared harness, registration generator, dispatcher, compiler and submodules were untouched. The only additions outside this directory are two separately named Go checker-question files and their direct test; each question registers in one initializer line using the existing registry.

## Implementation and independent comparison

Each rule has its own module. Adamic owns reference tracking, global/shadow classification, regex spelling and suggestion edits, implicit arguments checks, dependency collection and stability, dependency-tree recommendations, complete messages and diagnostic anchors. Regex alias tracking preserves duplicates and source traversal order. React distinguishes explicit question-dot tokens from optional-chain membership; the byte oracle caught that distinction in a suggestion before it was corrected.

The raw identifier-binding question returns symbol identity and ordered declaration filenames, declaration-file bits, kinds and spans, including implicit arguments and shorthand value symbols. It does not follow aliases. The raw source-has-jsx question returns syntax presence only. Both questions have their own Adamic adapter and Go implementation. Neither invokes a cohere rule or returns lint verdicts, messages, suggestions or fixes. ABI and release ownership remain unchanged.

The isolated Go oracle loads and walks an independent tsgo AST and invokes the pinned production rule Run methods unchanged with default options. It imports no bridge code. Full output includes UTF-8 spans, rule name, message ID/text, fixes and every suggestion including edits. Defaults follow the existing 26-rule suite: there is no rule-option configuration surface. Source rows from configurable-rule tests are reevaluated under defaults, not asserted against their option-specific original expectations.

Pins: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db; typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; TypeScript compiler 050880ce59e30b356b686bd3144efe24f875ebc8.

## Findings, fixes and suggestions

The frozen manifests retain 77 compiler roots and 287 repository roots (212 .a, 75 .ts), with configured declarations. Controls comprise 13 targeted sources and 690 literal source rows: 242 regex reference rows, 248 regex corpus rows, 24 rest-parameter rows and 176 React rows. Coverage includes aliases, global object accesses and writes, duplicate traces, casts, shadows, implicit arguments in arrows, optional dependency chains, unstable constructions, state setters, malformed dependency arrays, async effects, named callbacks, Unicode and CRLF.

| Population | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| 703 controls | 596 | 322,999 | PASS |
| 77 compiler roots | 0 | 5,010 | PASS |
| 287 repository roots | 0 | 18,485 | PASS |

Control breakdown: regex 355 findings/305 suggestions, rest 18 findings/no suggestions, React 223 findings/133 suggestions. There are zero top-level fixes under defaults. Regex suggestion replacement spans and text are compared. The pinned Go React suggestions have zero edits under defaults; those zero edit counts and full messages are compared too. Native comparison stderr is empty, including all sanitized runs.

Compressed full streams, all 703 source controls, manifests, logs and hashes are in [evidence](evidence/diagnostic-hashes.json). Controls are compressed to avoid entering repository source discovery. [Source hashes](evidence/source-hashes.json) pin the corpus and implementations. Stream paths are the absolute paths from this run; portable frozen manifests are also included.

## Commands and outputs

Initial setup remained valid: bash cloud/setup.sh passed; Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s, build-cache warm 115s, total 115s. nproc: 5, with a four-core quota. Go 1.27.1, clang 20.1.8, Node 24.19.0. Commands source /workspace/adamic-tools/env.sh. Test output was written directly to log files.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE12_FOURTH_ARTIFACTS=/workspace/wave-12/fourth/attempt7 \
go test -count=1 -v ./stage1/cohere/typeaware/wave12_fourth \
  > /workspace/wave-12-fourth-attempt7.log 2>&1
go test -count=1 -v ./bridge/tsgo/checker \
  > /workspace/wave-12-fourth-checker.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware/wave12_fourth \
  > /workspace/wave-12-fourth-vet.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  > /workspace/wave-12-fourth-node.log 2>&1
```

Owned package PASS 129.132s; checker package PASS 0.157s; vet clean; filtered Node/native/emitted-JavaScript oracle PASS 19.439s. The latter covers five existing compiler fixtures and its one-byte mutant, not emitted JavaScript of this lint suite.

Both corpus checks were separate from the package run. For each compiler/repository manifest, the attempt7 oracle, native and native-asan binaries were invoked with the same config/root arguments, stdout/stderr directed to corpus-named artifact files, then cmp checked both native streams against Go. Configs: /workspace/wave-12/corpus/src/compiler/tsconfig.json and /workspace/adamic/tsconfig.json. Roots: /workspace/wave-12/compiler.manifest and /workspace/wave-12/repository.manifest. All byte comparisons passed and normal/sanitized native stderr was empty. The test supports rerunning these via ADAMIC_WAVE12_FOURTH_COMPILER_MANIFEST, ADAMIC_WAVE12_FOURTH_REPOSITORY_MANIFEST and ADAMIC_TYPESCRIPT_SOURCE.

Initial attempts caught native build constraints, explicit-root discovery for all-.a configs, reference descriptions/options mistakenly extracted as source, and optional-chain suggestion spelling. All were corrected inside owned files. Failed logs are preserved and build failures are not counted as killed mutants.

## Mutants and release checks

| Mutation | Execution | Catch |
| --- | --- | --- |
| Regex reverses comment suggestion gate | Exit 0, empty stderr | Full bytes differ at 521 |
| Rest reverses implicit-binding declaration check | Exit 0, empty stderr | Full bytes differ at 5,617 |
| React reverses optional access in suggestion | Exit 0, empty stderr | Full bytes differ at 6,995 |
| Raw declaration-file bits forced false | Exit 0, empty stderr | Full bytes differ at 62 |
| Released registry retains program | Exit 0 instead of panic 70 | Released-handle assertion |
| Raw JSX presence reversed | Valid controls fail | Success requirement |
| Node reference output changes one byte | Output rejected | TestTheOracleCatchesOneByte |

The three rule mutants and raw binding mutant are caught solely by complete output bytes. Both new questions reject released programs with panic 70 and invalid or released checker handle. Direct checker tests compare raw symbol/declaration results against checker APIs and exercise implicit arguments, shorthand values, aliases and invalid requests. Scratch copies and Go overlays isolate all mutations.

## Native against Go

Three alternating whole-process rounds after builds finished; every round checked complete bytes. The existing benchmark script was reused unchanged. [Measurements](evidence/measurements.json) retain phases and raw rounds.

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.132s | 0.467s | 6.70x |
| Repository | 0.389s | 0.098s | 3.95x |

Compiler phase medians: native load 0.322s/run 2.650s; Go load 0.306s/run 0.147s. Repository: native load 0.087s/run 0.279s; Go load 0.080s/run 0.011s. Native remains slower; these are whole-process observations, not isolated bridge-call costs.

## Exact shared gaps and limits

The shared native parser has no JSX productions. Seventeen JSX-bearing React reference rows cannot run through it. The owned runner checks the raw Go AST first and explicitly panics 70 with wave 12 shared parser does not support JSX; a genuine .tsx probe proves the refusal. It does not silently omit JSX findings. No shared parser edits were made.

The shared emitted-JavaScript/profile harness work on codex/lint-harness-dot-a was not imported or edited. The native .a suite compiles and its owned serializer compares suggestions, but emitted-JavaScript execution of this lint suite remains uncovered. The original Go CLI still does not discover .a inputs; the independent oracle uses explicit roots.

Only production defaults were implemented/validated, as in the original suite. Nondefault regex redundant-wrapping and React option behavior are outside this fixed runner. Full repository gate, previous 26-rule regression suite, all possible regex patterns, all React reference rows, arbitrary JSX projects and rule-option matrices were not run. The corpus is frozen to the existing 77/287 roots; it is not silently expanded with this batch's generated controls.
