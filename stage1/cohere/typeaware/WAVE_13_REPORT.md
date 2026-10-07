Built: no-unassigned-vars, preserve-caught-error, and @typescript-eslint/consistent-type-exports in native Adamic, with two raw checker questions.
Commits: claim 153eb7ef (pushed before implementation); implementation 6b804574; this report and evidence are committed separately.
Commands and outputs: wave comparison PASS 243.094s; bridge PASS 223.933s; filtered Node oracle PASS 22.365s; vet, gofmt and diff checks clean.
Mutants: all three rule mutants and two checker-fact mutants caught only by independent Go byte comparison; four question refusal mutants, released-registry mutant and seven foundation mutants caught.
Not covered: nondefault rule options, full upstream fixture matrix, full go test ./..., and whole-source cohere lint/format, whose pinned CLI rejects .a discovery.

Base and selection

Started from origin/codex/tsgo-c-library at 0d540f413625f016f20fea39761c7b184f335de6. Read CLAUDE.md and the three named typeaware documents before editing. Counted the all-family volume ranking after excluding the 26 existing ports. Remaining positions 37, 38 and 39 are no-unassigned-vars (2), preserve-caught-error (2), and @typescript-eslint/consistent-type-exports (1). Checked implementation names and claim documents across 273 origin branch refs; no collisions, so no rules skipped. Claim commit was pushed before coding.

Each rule has its own .a file. export_symbol_chain and export_module_properties expose raw checker flags, declarations, ancestors and module property lookup facts; native code decides the lint verdict. Each question has a dedicated Go and .a file. The only shared source edit is two switch registration entries (four gofmt lines) in bridge/tsgo/checker/facts.go. No protected compiler files or submodule pins changed. See EXPORT_QUESTIONS.md for the wire schema and ownership.

Observed comparison evidence

The independent oracle calls unchanged production Go cohere rules and compares the entire canonical finding stream, including descriptions, spans, fixes and suggestions. It does not import bridge verdict code. Both normal and ASan/UBSan/LSan native executions matched:

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Generated controls | 51 | 32 | 18,855 |
| TypeScript src/compiler | 77 | 4 | 7,087 |
| Frozen repository | 287 | 1 | 19,144 |

Compiler corpus is TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Repository roots are the base branch's frozen coverage manifest, including 212 .a and 75 .ts roots. Portable manifests, compressed complete finding streams, SHA-256 equality records and logs are in validation-wave-13. Control source generation is in wave_13_test.go: 48 .a TypeScript subjects and three imported TypeScript helper subjects; all new executable Adamic sources are .a.

Mutants and catches

| Mutation | Catch |
| --- | --- |
| Unassigned read-without-write predicate changed to read-or-write | Compiled, exit 0, empty stderr; Go byte mismatch at 1339 |
| Caught-error cause suggestion gains trailing space | Compiled, exit 0, empty stderr; Go byte mismatch at 4116 |
| Export type insertion gains trailing space | Compiled, exit 0, empty stderr; Go byte mismatch at 14883 |
| Export symbol flags replaced with property flag | Compiled, exit 0, empty stderr; Go byte mismatch at 14601 |
| Export module lookup presence inverted | Compiled, exit 0, empty stderr; Go byte mismatch at 17024 |
| Released program retained in registry | Mutant exits 0; required stale-handle panic assertion catches it |
| Symbol-chain node-kind guard removed | TestExportCheckerQuestions rejects accepted invalid question |
| Module-properties node-kind guard removed | Same refusal assertion |
| Symbol-chain suffix guard removed | Same refusal assertion |
| Module-properties suffix guard removed | Same refusal assertion |
| Foundation C input buffer +1 | ASan heap-buffer-overflow |
| Foundation C output buffer +1 | ASan heap-buffer-overflow |
| Foundation registry retains released handle | Stale-handle assertion |
| Foundation wrong source position | Independent Go comparison at byte 6 |
| Foundation link opt-in removed | Native refusal assertion |
| Foundation C free omitted | LeakSanitizer |
| Foundation region allocation changed to heap | LeakSanitizer |

The valid released-handle probe panics with status 70 and "invalid or released checker handle". Foundation parity covers 1,600 positions in four files, 54,982 identical bytes, and lifetime checks across 100 queries. Sanitizers cover native/C memory; Go runtime-managed allocation is not a C leak assertion.

Timing observations

cloud/setup.sh passed: Go ready 0s, clang ready 1s, Node ready 1s, submodules 1s, cache 148s, done 148s. nproc reported 5; CPU quota was four cores. Environment is /workspace/adamic-tools/env.sh.

Three isolated alternating native/Go rounds, with complete output hash equality each round, measured whole-process medians:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 5.837513s | 1.772527s | 3.2933 |
| Repository | 0.328551s | 0.115479s | 2.8451 |

Native made 54,960 checker queries on compiler roots and 43 on repository roots. These ports are slower than Go in this observation. These isolated measurements supersede the timings emitted while other verification builds were active. Full three-round timings and load/query/run counters are in validation-wave-13/measurements.json.

Commands

All test stdout and stderr were redirected to log files, never piped. From the repository, after sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_WAVE13_STAGE0=/workspace/wave-13-adamic ADAMIC_WAVE13_ARCHIVE=/workspace/wave-13-tsgo.a ADAMIC_WAVE13_ARTIFACTS=/workspace/wave-13-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus ADAMIC_WAVE13_COMPILER_MANIFEST=/workspace/wave-13-compiler.manifest ADAMIC_WAVE13_REPOSITORY_MANIFEST=/workspace/wave-13-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave13AgreementAndMutants$' -count=1 -v -timeout=30m > /workspace/wave-13-test-final.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave13-corpus go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-13-bridge-test.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' -count=1 -timeout=10m -v > /workspace/wave-13-node-oracle.log 2>&1
go vet ./... > /workspace/wave-13-vet-final.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-13-gofmt-final.log 2>&1
git diff --check > /workspace/wave-13-diffcheck.log 2>&1
```

The four question guard mutations were applied with Go overlays and each compiled before failing TestExportCheckerQuestions. Logs and overlay mutation results are retained. The comparison harness constructs the rule/fact/registry mutants itself.

Limits

Default rule behavior is implemented and verified over the specified corpora and controls. Custom caught-error constructor/requireCatchParameter options and alternative consistent-type-exports option modes were not exercised. No full upstream options/JSX fixture matrix, suppression engine or application of edits was verified. The entire serialized edit and suggestion payload was compared. No full go test ./... was run; touched bridge packages, the wave gate, filtered external Node oracles, and repository-wide vet ran instead.

The pinned cohere command was built and invoked with --no-cache --no-fix on all six new .a files. It exited 1 before linting: "nothing to check: none of the named paths is in the program" because ".a is not a TypeScript or JavaScript file, so no tsconfig can put it in the program." The whole-source cohere lint/format gate therefore did not pass on this base. See validation-wave-13/cohere.log. Native compilation, exact production-rule comparison and sanitizers did pass. No extension workaround or unrelated CLI modification was made.
