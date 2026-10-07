Built two native Adamic rules; nexus/concurrency-no-lost-update remains unimplemented and released for reassignment.
Commits: claim a628406c; implementation and evidence are the commit containing this report.
Commands: final wave test PASS 47.192s; bridge PASS 69.164s; filtered Node oracle PASS 28.295s; checker tests and vet PASS.
Mutants: ISO boundary and callback range caught by independent Go bytes; malformed checker suffix caught by direct checker test; bridge and Node mutants caught.
Not covered: lost-update port, full repository gate, complete upstream fixture/options matrix; pinned cohere CLI rejects authored .a files.

# Wave 30 report

This wave is incomplete. The two completed rules are `nexus/consistency-no-iso-string-date-cut` and `nexus/correctness-no-callback-in-parse-try`, each in its own `.a` file. The assigned `nexus/concurrency-no-lost-update` has no implementation, mutant or timing in this branch. Its 1,442-line upstream implementation uses control-flow graph traversal, suspension edges, declaration anchors, store identity, loops and overwrites. The existing native stage1 runner has no corresponding CFG analysis. Implementing that analysis remains necessary; a Go-side lint verdict would violate the required native port.

## Selection and setup

The claim records positions 88, 89 and 90 after removing existing ports from the combined volume ranking. All origin heads were fetched and checked for competing ports and claims before the claim commit was pushed. None was found. Base: `0d540f413625f016f20fea39761c7b184f335de6`. No implementation preceded the claim push.

`bash cloud/setup.sh` passed: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node 24.19.0 ready 0s, submodules 0s, build cache warm 78s, done 78s. `nproc`: 5; CPU quota 400000/100000. Environment: `/workspace/adamic-tools/env.sh`.

## Checker question and native rules

`symbol-lineage` lives in its own Go and Adamic files. Its sole shared registration is one switch case in `bridge/tsgo/checker/facts.go`. It exposes opaque symbol identity, declaration records, declaration names, ancestor kinds and raw flags, external-module/global-augmentation facts and non-nullable call-signature count. Normal mode resolves import aliases; the explicit `raw` suffix preserves alias declarations for the ISO rule. Unexpected suffixes are rejected. Lint decisions remain in Adamic.

The ISO rule recognizes range truncation, split indexing and split destructuring while checking Date/default-library declarations and single const initializers. The callback rule checks discarded catch bindings through symbol references, JSON/nexus parse identity, callable parameters, Promise executor exclusions and guarded traversal through nested try/catch/finally and CFG roots. Both production rules emit findings only, with no fixes or suggestions; the oracle nevertheless serializes all finding/fix/suggestion fields.

An initial attempt to reuse Caller pulled in an existing Unused constructor that stage0 refused for escaping `this` before field initialization. The final implementation uses the isolated lineage decoder and does not modify that existing code.

## Agreement and failure proofs

The oracle invokes the unmodified pinned Go cohere registry rules using its own loader and traversal. It does not invoke the native rules. Normal and ASan/UBSan/LSan native outputs matched its canonical bytes:

| Input | Files | Findings | Canonical bytes |
| --- | ---: | ---: | ---: |
| Controls | 25 | 17 | 10232 |
| Frozen repository manifest | 287 | 0 | 18485 |
| TypeScript src/compiler manifest | 77 | 0 | 5318 |

The compiler checkout is TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Corpus manifests come from the existing validation-volume/compiler.manifest and validation-coverage/repository.manifest. Zero findings on the two requested corpora limits positive coverage; the controls provide positive and negative cases, including shadowing, nullable/destructured callable parameters, catch shorthand references, Promise exclusions, nested handlers, computed class property names, hexadecimal/decimal literals and Unicode/CRLF byte ranges.

ISO mutant `end <= 10` to `end <= 9` compiled and exited zero; only the Go byte comparison caught byte 58. Callback mutant extends a finding end by one byte; it compiled and exited zero and only the Go byte comparison caught byte 3968. A checker mutant accepting arbitrary suffixes failed the direct test with `symbol-lineage accepted a suffix`. Released lineage handles failed with panic 70 and exact invalid-or-released-handle text.

The full bridge package gate passed normal and sanitizer checks and its mutants: input/output lengths caught by ASan, retained released handle by stale-handle assertion, wrong source position by Go byte comparison, removed link opt-in by refusal, missing free and region-on-heap allocation by LSan. The filtered Node oracle passed seven fixture cases and its one-byte mutant. Logs and compressed canonical outputs with SHA-256 hashes are in validation-wave-30.

Commands run with stdout/stderr redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE_30_ARTIFACTS=/workspace/wave-30-validation-property ADAMIC_WAVE_30_REPOSITORY_MANIFEST=/workspace/wave-30-repository.manifest ADAMIC_WAVE_30_COMPILER_MANIFEST=/workspace/wave-30-compiler.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript go test ./stage1/cohere/typeaware -run '^TestWave30AgreementAndMutants$' -count=1 -v -timeout=30m
go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go test ./bridge/tsgo/checker -count=1 -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|functions|closures)\.a$' -count=1 -timeout=10m -v
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
```

## Timing

Three alternating native/Go rounds with byte comparisons each round; median whole-process wall time, including program load. These measure the two-rule runner only. The corpus has zero findings, so these are load/traversal/query timings rather than a positive-finding throughput claim.

| Corpus | Native | Go | Native/Go |
| --- | ---: | ---: | ---: |
| compiler | 1679.043 ms | 277.787 ms | 6.04x |
| repository | 227.416 ms | 106.671 ms | 2.13x |

Raw rounds and component timings are saved in validation-wave-30/timings.json. Command: `python3 bridge/tsgo/profile/volume_bench.py <native> <oracle> <output> --corpus compiler <config> <manifest> --corpus repository <config> <manifest>`.

## Limits

The full repository gate and complete upstream fixture/options matrix were not run. The implemented nexus parser import identity and global augmentation paths do not yet have positive controls. The pinned cohere CLI returned exit 1 when asked to check the new `.a` files: “nothing to check” because it accepts TypeScript/JavaScript extensions only. Native stage0 compiles these `.a` files and the production-rule oracle loads the `.a` controls, but the standalone cohere soundness/style/formatter gate remains unverified. No `.ts` implementation was authored to work around that CLI limitation. Protected compiler files and submodule pins were not edited.
