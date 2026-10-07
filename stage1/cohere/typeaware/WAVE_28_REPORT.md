Built: Verify optional parity, block-scoped-var, and getter-return in separate .a files.
Commits: claim ef4a0ad7; implementation and evidence are committed together after this report.
Commands/results: wave gate PASS 118.688s; bridge PASS 113.666s; checker PASS 0.220s; filtered Node oracle PASS 37.026s; vet and Go formatting pass.
Mutants: all three rule mutants differ only in compared output; released-registry mutant, seven bridge mutants, and Node's one-byte mutant are caught.
Not covered: full repository gate, nondefault rule options, speed parity, and cohere CLI lint of .a on this pinned branch.

Branch: `codex/typeaware-wave-28`. Base: `0d540f413625f016f20fea39761c7b184f335de6` on `origin/codex/tsgo-c-library`. The claim was committed and pushed before implementation. Filtering the by-volume ranking against the base's 26 ports leaves 172 candidates; remaining positions 82, 83, 84 are `base/correctness-require-verify-optional-parity`, `block-scoped-var`, `getter-return`. Each has zero compiler and repository volume. No implementation or claim matched on the 268 fetched origin refs, representing 64 distinct stage1 trees. No rule was skipped.

The three rule files use the base's `raw-type`, `symbol-identities`, and `node-symbol-details` questions. No new checker question was needed. No existing tracked file, shared registration, compiler emitter, lowerer, or oracle file changed. `wave_28_suite.a` loads one checker program and parses each source once. The independent Go runner calls the pinned, unmodified production rules through the registry, with its own loader and AST walk, and imports no bridge code. Both runners sort and serialize complete diagnostic spans, names, IDs, messages, fixes, and suggestions. Production defaults for these three rules emit no fixes or suggestions; the comparison checks their zero lengths too.

The comparison found and held two implementation corrections: the pinned checker's Void flag is 16, and speculative parser nodes such as a temporary type-only import token are not members of the final AST. Rules now visit linked nodes only. The type-only import control deliberately declares a variable named `type`, so a spelling filter cannot conceal that distinction. Earlier failures also found default diagnostic namespace prefixes and an omitted destructuring element. All are corrected in the passing run. The bridge refuses mismatched exact nodes instead of returning an approximate answer.

The corpus is frozen to the base's `validation-coverage` manifests: 287 repository roots and all 77 TypeScript compiler roots. TypeScript is commit `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3), downloaded as the exact codeload archive after a full clone proved slow. Cohere is `715ba94f3608a6500086b1076ce5cb7e51b836db`; its typescript-go is `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Root contents are SHA-256 recorded in `validation-wave-28/*-roots.json`; manifests are copied unchanged from the base. Imported files and ambient declarations use the configured programs. The 255 controls are mechanically extracted from the pinned production Go tests plus additional Unicode/CRLF, destructuring, symbol identity, descriptor/global shadow, generics, nullable types, and import controls. All 255 parsed; none was excluded. Tests use the established config and explicit .a roots with AllowNonTsExtensions, preserving config declaration files.

| Population | Roots | Findings | Identical complete stdout bytes | Native sanitizer stderr |
| --- | ---: | ---: | ---: | ---: |
| Controls | 255 | 126 | 45,419 | 0 |
| Repository | 287 | 0 | 18,485 | 0 |
| Compiler | 77 | 0 | 5,010 | 0 |

Controls produce 12 parity, 47 scope, and 67 getter findings. Normal native and ASan/UBSan/LSan output equals independent Go output byte for byte for every population. The empty corpora alone would not establish rule correctness; positive controls and mutants supply that evidence. `validation-wave-28/comparisons.json` records hashes, lengths, stderr lengths, and mutant offsets. The `.findings` artifacts preserve the complete canonical streams, including absolute run paths. Those path prefixes affect hashes on another machine.

| Mutant actually run | Catch |
| --- | --- |
| Parity uses only root flags instead of traversing union alternatives | Exit 0, empty stderr; Go bytes differ at 36,675 |
| Scope containment accepts every identifier in the source | Exit 0, empty stderr; Go bytes differ at 49 |
| Getter if-statement exit uses OR instead of AND | Exit 0, empty stderr; Go bytes differ at 21,640 |
| Wave released registry keeps the released program live | Mutant exits 0; required panic 70 is absent |
| Bridge input length +1 | ASan heap-buffer-overflow |
| Bridge output length +1 | ASan heap-buffer-overflow |
| Bridge released handle stays live | Stale-handle assertion |
| Bridge type comes from SourceFile position | Independent Go byte mismatch at 6 |
| Bridge removes link opt-in guard | Refusal test |
| Bridge omits C output free | LeakSanitizer |
| Bridge allocates region result on heap | LeakSanitizer |
| Node oracle changes one output byte | Differential oracle mismatch |

The unmodified released-handle probe exits 70 with exactly `adamic: panic: invalid or released checker handle`. The full bridge gate additionally proves 100 C ABI queries, output ownership after release, rejected zero/stale handles, a distinct subsequent handle, and 162 independent source positions producing 3,261 identical bytes under sanitizers. Complete gate logs and the released probe streams are committed in `validation-wave-28`.

Three fresh-process rounds, alternating subject order, measure the same full finding stream with warm filesystem caches. Every timed stdout is byte-compared again. These are observations on this machine, not a general speed claim.

| Population | Median native seconds | Median Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.341393 | 0.194896 | 1.75 |
| Compiler | 9.166505 | 1.891799 | 4.85 |

Native is slower. Compiler native makes 193,228 checker calls; bridge query time is about 1.2s of about 8.9s reported native run time. Repository makes zero checker calls for these rules. No profiling experiment here attributes the remaining cost to a particular component. `timing.json` preserves every process and internal timing observation.

Setup: `bash cloud/setup.sh > /tmp/wave-28-setup.log 2>&1`, then `source /workspace/adamic-tools/env.sh`. Go 1.27.1 ready 0s; clang 20.1.8 ready 1s; Node v24.19.0 ready 1s; submodules ready 1s; build cache warm 114s; done 114s. `nproc` is 5; CPU quota is `400000 100000`; reported memory is 17.6 GB. Setup succeeded. The setup log is included.

Commands run from the repository with the setup environment sourced:

```bash
ADAMIC_WAVE28_ARTIFACTS=/workspace/wave-28-validation \
ADAMIC_WAVE28_REPOSITORY_MANIFEST=/workspace/wave-28-artifacts/repository.manifest \
ADAMIC_WAVE28_COMPILER_MANIFEST=/workspace/wave-28-artifacts/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-28-corpus \
go test ./stage1/cohere/typeaware -run Wave28 -count=1 -timeout=20m -v \
  > /tmp/wave-28-final-test2.log 2>&1

go test ./bridge/tsgo -count=1 -timeout=15m -v > /tmp/wave-28-bridge.log 2>&1
go test ./bridge/tsgo/checker -count=1 -timeout=10m -v > /tmp/wave-28-checker.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /tmp/wave-28-node-oracle.log 2>&1

gofmt -l stage1/cohere/typeaware/wave_28_test.go \
  stage1/cohere/typeaware/testdata/oracle_wave_28.go > /tmp/wave-28-gofmt.log
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /tmp/wave-28-final-vet.log 2>&1
```

Wave, bridge, checker, and filtered oracle all exit 0 with the pass times above. Go formatting and vet exit 0 with empty logs. The Node filter runs eight fixtures (functions, generic_functions, closures, method_closures, maps_and_text, sorting, string_index, lone_surrogates), each through Node, JavaScript backend, and sanitized native, plus the one-byte mutant. The full `go test ./...` and full `go vet ./...` gates were not run; validation is the touched packages and this filtered oracle. The wave test without corpus environment variables reports that corpora were not supplied, so reproduce the full gate with the manifests above, materializing their portable root paths locally.

The pinned production cohere CLI was built from `./command/cohere` and run with `--no-cache --no-fix --lint` on the four new .a entry/source files. It exits 1: `.a is not a TypeScript or JavaScript file, so no tsconfig can put it in the program`. The branch's raw config discovery also rejects an all-.a input config. Explicit roots work in the established bridge and independent Go harness, as the passing comparisons show. The committed lint refusal is not a clean lint result. No .ts files were created as a workaround, and no submodule or shared CLI was changed. Nondefault rule option modes, arbitrary corpora, and speed parity remain outside this evidence.
