Rebased wave 14 onto current main and adapted two owned Repair constructions.
Rebase head 75caf8c1; main e8ba3d5d; previous pushed head 68a2070a.
All four owned oracle suites, bridge packages, filtered Node oracle and Go vet pass.
Nineteen semantic mutants, released-handle mutants and seven bridge mutants caught.
No new claims; native pattern validation, undefined labels and real JSX remain incomplete.

The requested landing cap was this unit. Only codex/typeaware-wave-14 was pushed
by this worker. Its eleven unique commits rebased cleanly onto origin/main
e8ba3d5d81de4d3773c723914fccd4c76248b965. The remote branch still held
68a2070a13d0d1356733595e5f8f31c691224e99 when checked before publishing.
The explicit request to rebase and push this branch supersedes CLAUDE.md's
ordinary prohibition on rewriting history for this operation. Publishing uses
an explicit lease for that old SHA and names only codex/typeaware-wave-14.

Current main's nominal class check refused two structural Repair values. The
owned no_array_delete.a and no_misleading_character_class.a now construct
Repair instances. Their positions, text and suggestion contents agree with Go.
No shared harness, generator, parser or protected compiler file was edited.

All commands source /workspace/adamic-tools/env.sh. Setup printed ready at 0 s
for Go, clang, Node and submodules, cache warm at 113 s and done at 113 s.
nproc reports 5, with cpu.max 400000 100000 and 17.6 GB available memory.

The initial combined invocation used the original frozen compiler and repository
manifests through ADAMIC_TYPESCRIPT_SOURCE, ADAMIC_WAVE14_COMPILER_MANIFEST,
ADAMIC_WAVE14_REPOSITORY_MANIFEST and the corresponding NEXT manifest variables:

```sh
go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
```

Continuation PASS 128.62 s; render judgments PASS 49.19 s. The other two suites
failed at the nominal Repair check before running their comparisons. That failed
log is retained. After the two owned changes, the same population settings ran:

```sh
go test ./stage1/cohere/typeaware \
  -run '^TestWave14AgreementAndMutants$|^TestWave14ThirdAgreementAndMutants$' \
  -count=1 -v -timeout=30m
```

Original PASS 80.64 s; third PASS 225.54 s; package PASS 306.179 s. The original
suite compares 33 control findings and three compiler findings. The third suite
compares 54 controls / 60 findings and 62 constructor controls / 58 findings.
Both frozen corpora also agree in every suite. All complete canonical finding,
fix and suggestion bytes agree with the independent production Go rules in
normal and ASan/UBSan/LeakSanitizer builds. Supported native stderr is empty.
Different scratch path lengths explain byte counts differing from older reports.

The 83 previously supported upstream constructor inputs also agree on 89
findings, 38,351 bytes, SHA256
b47e0caed3215599fe7cc8c14919902322f7c1633fea6c748dd0d9e8aaab6d92.
The excluded strict-module legacy-octal input remains excluded. An initial
standalone runner incorrectly required empty Go stderr; Go emits timing
telemetry. The corrected runner preserves it and requires empty native stderr.
Both runner logs are retained.

Every semantic mutant exits 0 with empty stderr; the full Go byte comparison
catches it. Mutants and first differing bytes from this run:

| Mutant | Byte |
| --- | ---: |
| Array delete judgment | 55 |
| Stringification judgment | 1536 |
| Extraneous class judgment | 4348 |
| Listener assertion judgment | 161 |
| Module namespace origin | 4969 |
| Leaked render judgment | 85 |
| Cooked mapping | 57 |
| Constant writes ignored | 8680 |
| Constructor flags ignored | 459 |
| Overridden literal checked twice | 10234 |
| Constant finding deduplication removed | 6670 |
| Alias traced as global root | 12739 |
| Global writes ignored | 21885 |
| Label membership | 52 |
| Flag precedence | 5120 |
| Unicode printable quoting | 5717 |
| Surrogate decoding | 18009 |
| Combining class judgment | 6867 |
| Scope meaning Value changed to Variable | 1350 |

Original, continuation and third suites separately prove released handles panic
70 and that retaining a released registry entry exits 0 and fails that assertion.

Additional commands, all redirected directly to retained log files:

```sh
ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript \
  go test ./bridge/tsgo/... -count=1 -v -timeout=15m
go test ./bridge/tsgo/checker -run '^TestWave14|ScopeSymbols' -count=1 -v -timeout=10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -v -timeout=10m
go vet ./...
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
git diff --check
```

Bridge PASS 67.930 s, checker PASS 0.194 s. It compares 1,600 positions across
four compiler files, 54,982 bytes under sanitizers. Its seven mutants are caught:
input and output lengths by ASan, retained released handle by its assertion,
wrong type position by byte comparison, removed link opt-in by refusal assertion,
missing C buffer free by LeakSanitizer and heap region result by LeakSanitizer.
The separate owned question tests PASS 0.449 s. Uncached Node oracle PASS
10.730 s, including nine matched fixtures and the one-byte mutant. Go vet and
Go formatting print nothing. The complete repository gate was not run.

Quiet alternating three-round whole-process native / Go medians, after byte
agreement and all builds/tests, using the existing count-mode benchmark:

| Population | Native seconds | Go seconds | Ratio |
| --- | ---: | ---: | ---: |
| Compiler | 1.834315 | 0.347441 | 5.28x |
| Repository | 0.249979 | 0.128145 | 1.95x |
| Controls | 0.022329 | 0.029605 | 0.75x |
| Constructors | 0.025764 | 0.033688 | 0.76x |
| Upstream | 0.024694 | 0.041000 | 0.60x |

Compilation and sanitizers are excluded from timing. Constructor and upstream
runs use --class-only in both implementations. Raw rounds and the measurement
script are retained in validation-wave-14-landing.

Real JSX still fails the shared parser. undefined labels also fail that parser.
no-invalid-regexp's missing native ECMAScript rewrite and regexp2-compatible
validation/error messages remain substantive owned implementation, with a
Go-positive new RegExp('[') witness and explicit native NotYet panic 70.
No further rules are claimed. Nondefault options, new-rule emitted JavaScript
comparison and the full repository gate are not claimed covered by this unit.
