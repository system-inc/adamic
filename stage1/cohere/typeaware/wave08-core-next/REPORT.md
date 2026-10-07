Built: three complete owned native profiles for require-atomic-updates, require-await and symbol-description, with numeric handed-node listeners.
Commits: claim pushed as 2ca1f5401 before code; rebased implementation 21b55f1a2 and final codec/listener work 004e127e1, based on main c01907a70.
Checks: 264/264 production-byte comparisons normal and sanitized, prior six rules re-green, bridge ownership, released handles, numeric manifests, vet and filtered Node oracle pass.
Mutants: atomic 72 cases, await suggestion 99, Symbol 20; numeric 214-to-215 differs; missing function metadata refuses; all prior completed-rule mutants caught again.
Uncovered: shared factory/dispatcher installation, parked React JSX/HIR/SSA production parity, new stage 3 corpus and the full repository Go gate; native remains slower than Go.

The native sources are .a. Rule callbacks read numeric kinds from cached nodes
and receive only their indexed subscriptions. SourceFile is Go's atomic listener;
its native analysis builds that file's suspending function graphs. The adapter
converts the shared parser's string kinds once outside the rules. No shared
registration generator, harness, parser or protected compiler source was edited.
The raw questions use an owned dispatcher overlay. Production registration and
shared factories must be installed by integration; the work is not claimed to be
loaded by the shared runner yet. No batch 8 Diagnostic integration SHA was given.

## Scope and production parity

Captured 255 upstream fixture programs while retaining their original production
calls and assertions. One intentionally checker-free Symbol witness is outside
the typed native comparison. The remaining 254 are atomic 94 / 59 findings,
await 126 / 97 findings, and Symbol 34 / 13 findings. Four additional generic
contract witnesses distinguish promise demands from unconstrained inference for
empty and multiline property names: two clean, two reporting. The question
encodes property presence separately, and SplitN preserves the final field's
entire name. Its two newlines are data, not additional question arguments.

Each rule also matches the branch's frozen 77 compiler and 287 repository files.
Those zero-volume corpora yield no findings for these three rules. Positive
controls and compiling mutants establish activity; corpus emptiness is not used
as proof that a rule works. Earlier standalone Symbol controls additionally cover
unfollowed aliases, ambient source declarations, Unicode, comments, optional
calls and explicit type arguments. The final profile includes the missing-function
metadata refusal, so an unmatched checker node cannot silently suppress a finding.

Go decides the full finding/fix/suggestion stream through the unmodified registry.
Native Diagnostic writes match that stream, including exact suggestion IDs,
messages, repair spans and replacement text. Atomic and Symbol offer no repairs;
await has suggestions, including the three semicolon-insertion shapes. The Go
runner passes the production RequireAtomicUpdatesOptions type, with true/false
and the default exercised. The shared JSON options/factory adapter is not ported.

## Mutants and ownership

All three rule mutants compile and every control run exits zero with empty
stderr. Inverting outdated-write membership differs in 72 programs (first byte
119). Inverting the await suggestion's semicolon replacement differs in 99
(first 715), while findings remain structurally valid. Inverting the Symbol name
filter differs in 20 (first 147). Only the production byte comparison catches
these. Numeric subscription 214-to-215 compiles and exits zero with empty stderr,
then disagrees with the independent Go AST enum bytes. Three rule.json manifests
are checked against Go's kind names and native numeric exports. A compiling
missing-snapshot mutant clears function presence and is refused before output
with exit 70. All four raw questions reject released handles before output, 70.

The final touched-package gate passes with the owned overlay: bridge 172.242 s,
checker 0.416 s, raw core questions 0.059 s. It includes retained ABI outputs,
stale/zero handle checks, direct Go/native oracles, ASan/UBSan/LSan and the existing
foundation mutants. The initial combined invocation exposed a test-package
import cycle; moving the owned tests to an external package removed it and the
final combined command passed. Raw tests independently pin the library/source
flag distinction, lack of alias following, contextual signatures, then-property
edges, union branches and the Go union flag 134217728. Vet passes.

The original wave 08 gate passes on the rebased branch in 221.007 s, including
normal/sanitized controls, frozen corpora, the three compiling production-rule
mutants, released registry witness and raw checker mutations. The prior process
profile matches all 100 upstream programs and both corpora normally/sanitized,
with process-callee and blocking-order mutants and released questions caught.
The prior timeout profile matches 22 controls / 13 findings and both corpora,
normal/sanitized, with its compiling timeout mutant, flow and handle witnesses.
React remains parked with its exact blockers named in the claim file.

Filtered TestTheOracleCatchesOneByte passes in 0.515 s with fresh native and Node
cache misses. No full repository Go test gate was run. Current main adds stage 3
material and one oracle test; it does not change the checker/parser dependencies.
The worker branch was rebased cleanly before publication and no integration
branch is a push target.

## Quiet timing

Three alternating fresh-process rounds after other validation jobs completed,
including program load, native parsing, rules and serialization. Every timing
round still compares complete production bytes. These are observations; the
profiles remain slower than Go. The numeric adapter eliminates duplicate token
start scanning and the full enum walk for empty operator names. No claim of a
native speedup over Go is made.

| Rule | Corpus | Native s | Go s | Native/Go |
| --- | --- | ---: | ---: | ---: |
| atomic | compiler | 9.180676 | 0.392416 | 23.395x |
| await | compiler | 9.230457 | 0.402828 | 22.914x |
| symbol | compiler | 3.754842 | 0.410077 | 9.156x |
| atomic | repository | 1.169865 | 0.161082 | 7.263x |
| await | repository | 1.078548 | 0.168377 | 6.406x |
| symbol | repository | 0.510515 | 0.163884 | 3.115x |

Original setup succeeded: setup: go ready (0s); clang ready (0s); node ready
(0s); submodules ready (0s); build cache warm (132s); done in 132s on 5 processors.
`nproc` is 5, CPU quota four cores (400000/100000), Go 1.27.1, clang 20.1.8 and
Node 24.19.0. Environment /workspace/adamic-tools/env.sh was sourced; setup was
not repeated after rebasing.

## Exact validation commands

All outputs go to log files. README contains archive/profile build commands.
The final compiler was rebuilt with go build -o /workspace/wave08-c019/adamic
./cmd/adamic on the rebased branch. The validators below use that compiler, the
owned /workspace/wave08-core/bridge normal/sanitized archives, the full captured
/workspace/wave08-core/upstream programs and /workspace/typescript-wave08-corpus.

```sh
python3 stage1/cohere/typeaware/wave08-core-next/compare_core.py --artifacts /workspace/wave08-core/final-compare --fixtures /workspace/wave08-core/upstream --native /workspace/wave08-c019/core-native --compiler-root /workspace/typescript-wave08-corpus
python3 stage1/cohere/typeaware/wave08-core-next/compare_core.py --artifacts /workspace/wave08-core/final-compare-asan --fixtures /workspace/wave08-core/upstream --native /workspace/wave08-c019/core-native-asan --compiler-root /workspace/typescript-wave08-corpus
python3 stage1/cohere/typeaware/wave08-core-next/validate_mutants.py --artifacts /workspace/wave08-core/final-mutants --baseline /workspace/wave08-core/final-compare --stage0 /workspace/wave08-c019/adamic --archive /workspace/wave08-core/bridge/checker-asan.a --fixtures /workspace/wave08-core/upstream
python3 stage1/cohere/typeaware/wave08-core-next/validate_listeners.py --artifacts /workspace/wave08-core/listener-validation --stage0 /workspace/wave08-c019/adamic --archive /workspace/wave08-core/bridge/checker.a --fixtures /workspace/wave08-core/upstream
GOFLAGS=-overlay=/workspace/wave08-core/bridge/overlay.json go test ./bridge/tsgo ./bridge/tsgo/checker ./stage1/cohere/typeaware/wave08-core-next -count=1 -v -timeout 30m
go vet ./stage1/cohere/typeaware/wave08-core-next
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v
ADAMIC_WAVE08_ARTIFACTS=/workspace/wave08-core/original ADAMIC_WAVE08_FACT_ARTIFACTS=/workspace/wave08-core/original-facts ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus ADAMIC_WAVE08_COMPILER_MANIFEST=/workspace/wave08-corpora/compiler.manifest ADAMIC_WAVE08_REPOSITORY_MANIFEST=/workspace/wave08-corpora/repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m
python3 stage1/cohere/typeaware/wave08-next/validate_process.py --artifacts /workspace/wave08-core/prior-process --fixtures /workspace/wave08-process-full --stage0 /workspace/wave08-c019/adamic --archive /workspace/wave08-core/bridge/checker.a --asan-archive /workspace/wave08-core/bridge/checker-asan.a --compiler-root /workspace/typescript-wave08-corpus
python3 stage1/cohere/typeaware/wave08-next/validate.py --artifacts /workspace/wave08-core/prior-race --stage0 /workspace/wave08-c019/adamic --archive /workspace/wave08-core/bridge/checker.a --asan-archive /workspace/wave08-core/bridge/checker-asan.a --compiler-root /workspace/typescript-wave08-corpus
python3 stage1/cohere/typeaware/wave08-core-next/time_core.py --artifacts /workspace/wave08-core/timing --baseline /workspace/wave08-core/final-compare --native /workspace/wave08-c019/core-native --oracle /workspace/wave08-core/final-compare/oracle
```

validation-c019 contains source hashes, result metadata, complete compressed
base64 output streams and all captured upstream program/config/option inputs.
No executables or generated fixture .ts files are committed. Earlier reports
remain historical evidence for their named bases. This report accompanies the
final evidence commit; the final response identifies its SHA.
