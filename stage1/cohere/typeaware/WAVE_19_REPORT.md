Built: default-option ports of consistent-generic-constructors, dot-notation and no-array-constructor, with three isolated raw checker questions.
Commits: base 0d540f413625f016f20fea39761c7b184f335de6; claim bd3aca9a; implementation 3dda5a1b772092525fe4d2bc5c2237e794ef91e9.
Commands and outputs: agreement PASS 165.432s; bridge PASS 120.764s and checker PASS 0.490s; filtered Node oracle PASS 43.187s; go vet ./... PASS; setup 137s; nproc 5.
Mutants: three rule and three fact mutants caught only by Go byte comparison; released registry mutant caught by required panic; seven existing bridge mutants and Node one-byte mutant caught by their designated checks below.
Uncovered: nondefault rule options, the existing let.if(); parser gap, direct cohere CLI lint of .a files, and the entire repository test suite; native timing is slower than Go.

## Selection and ownership

Read CLAUDE.md and the full typeaware README.md, VOLUME_REPORT.md and COVERAGE_REPORT.md before implementation. Used the task's specific bridge base instead of the generic origin/main instruction. Fetched origin branches and inspected claims and named ports before reserving the three rules. None was already ported or claimed at that snapshot; no rule was skipped. The claim commit was pushed before implementation.

VOLUME_REPORT.md links the all-family count tables. Sorting their combined compiler/repository counts descending with lexical ties, excluding the 26 base ports, yields:

| Remaining position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 55 | @typescript-eslint/consistent-generic-constructors | 0 | 0 |
| 56 | @typescript-eslint/dot-notation | 0 | 0 |
| 57 | @typescript-eslint/no-array-constructor | 0 | 0 |

[wave_19_selection.json](wave_19_selection.json) records the exact excluded set. The claim is [claims/wave-19.md](claims/wave-19.md). New Adamic source files all use `.a`. Existing parser/rule helpers retain their base filenames. No protected compiler files, submodule pins, or shared suites were modified. The only shared-file edit is exactly three single-line dispatch registrations in bridge/tsgo/checker/facts.go.

## Implementation and independent oracle

Each rule has its own file. consistent-generic-constructors moves annotation type arguments to the constructor, preserving annotation comments and adding parentheses when necessary; it respects isolatedDeclarations and built-in typed-array exemptions. dot-notation reproduces literal-key diagnostics and repairs, optional access, integer receivers, adjacency, and comment-based repair suppression; its index-signature exemption respects the compiler option. no-array-constructor distinguishes global declarations from shadows, preserves one-argument length constructors, and reproduces array-literal suggestions including source trivia.

The new bridge questions return facts, not rule decisions:

- isolated-declarations: raw compiler option on the exact SourceFile node.
- index-signature-access: raw compiler option, property presence and index key type flags on an exact element-access node.
- node-symbol-origin: exact GetSymbolAtLocation declaration files and library flags. This is necessary because value-name resolution does not identify a type-alias shadow in a type annotation.

Each question has a named Go implementation and Adamic decoder file. Direct checker tests compare the encoded facts to checker APIs, cover both option values, and reject malformed questions and wrong node kinds. Diagnostics are computed in Adamic.

The dedicated runner and independent Go executable use only these three rules. The Go oracle invokes the unchanged production rule implementations from pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and TypeScript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; it does not import the bridge fact implementations. Comparisons include ordered diagnostics, byte positions, messages, repairs, and suggestions, plus file headers and totals. Both implementations load `.a` controls explicitly while preserving configuration declaration roots.

This ports the default rule configurations used by the base suites. The production test inputs are extracted from pinned Go tests and replayed under those defaults, rather than claiming to implement the nondefault options in some original test rows.

## Agreement and memory checks

The compiler corpus is TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, using the base's frozen 77 compiler roots. The repository corpus uses the base's frozen 287 roots, so adding new ports does not silently change the population. Portable manifests are saved in validation-wave-19.

| Population | Roots/inputs | Findings | Identical output bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Targeted plus pinned production inputs | 212 | 171 | 78,044 | PASS |
| Same controls with isolatedDeclarations | 212 | 133 | 70,633 | PASS |
| Index-signature option controls | 1 | 4 | 1,765 | PASS |
| Repository | 287 | 0 | 18,485 | PASS |
| TypeScript compiler | 77 | 0 | 5,241 | PASS |

All 212 agreement inputs were independently accepted by Go's parser, so zero were dropped by parse filtering. Positive controls require a finding for each rule. Controls cover fixes and suggestions, comments, CRLF and Unicode line breaks, local typed-array type aliases and classes, function parameters versus arrows, constructor syntax, Array shadows and spreads, optional access, numeric receivers, and computed keys. An expanded comparison caught an actual CRLF annotation-comment mismatch; the fix now ends line comments at all four supported line terminators.

One additional pinned production input, `let.if();`, is excluded explicitly from agreement: Go parses it and reports zero default findings, while the existing Adamic parser exits 70 with `parser slice expected semicolon at 4`. A separate assertion proves that exact refusal. This parser gap was not fixed outside the assigned territory.

Released-handle probing calls a new question after release and requires panic 70 with `invalid or released checker handle`. The existing bridge regression also checks 100 C queries, owned outputs surviving release, stale and zero handles, distinct subsequent handles, and 1,600 positions across four files: 54,982 identical oracle bytes under sanitizers.

## Every mutant executed in the final checks

The first six compile successfully, exit 0 with empty stderr, and are killed only by comparison with the independent Go rule output. Offsets below are the first differing byte in the canonical stream.

| Mutant | Deliberate fault | Catch |
| --- | --- | --- |
| Generic rule | Require at least two annotation type arguments | Go byte comparison, byte 46 |
| Dot rule | Add an extra dot in optional access repair | Go byte comparison, byte 6,410 |
| Array rule | Suggest `[0]` instead of `[]` | Go byte comparison, byte 10,801 |
| Isolated option fact | Invert compiler option | Go byte comparison, byte 46 |
| Index key fact | Return zero index key flags | Go byte comparison, byte 1,327 |
| Node symbol fact | Return an absent symbol | Go byte comparison, byte 2,260 |
| Released registry | Keep released checker handle live | Mutant exits 0; required panic-70 check fails |
| Existing bridge input length | Input byte length plus one | ASan heap-buffer-overflow |
| Existing bridge output length | Output byte length plus one | ASan heap-buffer-overflow |
| Existing bridge stale handle | Keep released registry entry | C stale-handle assertion |
| Existing bridge node lookup | Ask type of SourceFile instead of target node | Independent checker byte oracle |
| Existing bridge linkage | Remove link opt-in guard | Unlinked-checker refusal test |
| Existing bridge C output ownership | Remove buffer free | LSan leaked buffers |
| Existing bridge region ownership | Allocate region result on heap | LSan unowned result |
| Existing Node oracle | Alter one output byte | TestTheOracleCatchesOneByte |

The initial build failure from overlapping file creation and package discovery is not a mutant kill. The earlier expanded comparison failed on the parser gap and CRLF mismatch; the passing evidence is from the subsequent complete run after the CRLF correction and explicit gap assertion.

## Native time against Go

After correctness and regressions finished, ran three alternating count-only rounds with the existing volume_bench.py. These measurements include loading and complete traversal. Count-only output agreed in every round; full diagnostic streams had already been compared above. Median wall times:

| Corpus | Go process | Native process | Native / Go | Go run | Native run |
| --- | ---: | ---: | ---: | ---: | ---: |
| Compiler | 0.406995s | 1.861266s | 4.573x | 0.058864s | 1.513268s |
| Repository | 0.171308s | 0.307864s | 1.797x | 0.066973s | 0.217622s |

Native performs 154 checker questions on compiler roots and 574 on repository roots. Both have zero findings for these selected rules, so these are traversal measurements, not findings throughput. Observed native execution is slower; no speedup is claimed. Raw rounds and medians are saved, alongside the agreement run's separate full-output timing pair.

## Commands and artifacts

Toolchain: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules 1s, cache warm 137s, and done 137s. `nproc` printed 5; cpu.max is 400000/100000, a four-core quota. Setup succeeded.

Executed from the repository with test output redirected to log files:

```bash
ADAMIC_WAVE19_ARTIFACTS=/tmp/wave19-expanded-fixed ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19AgreementAndMutants$' -v -count=1 -timeout=30m > /tmp/wave19-expanded-fixed.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave19-typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /tmp/wave19-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /tmp/wave19-node-oracle.log 2>&1
go vet ./... > /tmp/wave19-vet-all.log 2>&1
gofmt -l cmd internal > /tmp/wave19-gofmt.log
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave19-expanded-fixed/wave19 /tmp/wave19-expanded-fixed/wave19-oracle /tmp/wave19-bench --corpus compiler /workspace/wave19-typescript/src/compiler/tsconfig.json /tmp/wave19-expanded-fixed/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/wave19-expanded-fixed/repository.manifest > /tmp/wave19-bench.log 2>&1
```

The filtered Node run passed eight fixtures, including generic_functions and method_closures matched by the filter, plus the one-byte oracle mutant. Vet and the prescribed formatting check produced empty logs. New standalone Go files were gofmt'd. Shared fact registrations intentionally remain one line each as requested.

Attempted the pinned cohere CLI with `--no-cache --no-fix --lint` on all seven new `.a` sources. It exits 1 saying `.a` is not a TypeScript or JavaScript file and no tsconfig can put it in the program, despite the root sourceExtensions setting. This is not a successful lint run. Native builds load and type-check the new sources, and the dedicated production-rule oracle supports explicit `.a` roots. No `.ts` copies or CLI changes were introduced to hide this limitation.

[validation-wave-19](validation-wave-19/) contains final test logs, setup output, the CLI refusal, raw timing rounds, medians, portable manifests, control hashes, and stdout/stderr hashes. process-streams.tar.gz preserves every process stream from the final agreement/mutant run, including successful sanitized runs and the explicit parser refusal. Scratch binaries and generated controls remain at /tmp/wave19-expanded-fixed; the source-driven test recreates them.

The entire `go test ./...` gate and the older 26-rule suites were not rerun. The complete bridge packages, the new rule suite, filtered Node oracle, repository-wide vet, and prescribed formatting check are the reported coverage. JSX parser support and nondefault rule options were not added by this unit.
