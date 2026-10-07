Built: native .a ports of no-new-func, no-new-native-nonconstructor and no-new-wrappers.
Commits: claim e2397129; implementation 2c6d6696; evidence is committed separately.
Commands and outputs: rule suite PASS 199.337s; bridge PASS 125.755s; Node oracle PASS 47.944s.
Mutants: three rules and two questions caught only by finding bytes; registry, seven ABI and Node mutants caught too.
Not covered: full gate, all upstream fixtures, emitted JavaScript rule runner; pinned CLI refuses .a.

# Wave 26 fourth batch

The previous nine claims were complete, tested and pushed at a385f1e4 before
this batch was selected. This brings the branch's own completed claims to twelve.
The new claim was pushed before any implementation. No fifth batch is claimed.

## Selection and implementation

Fetched all origin branches with `git fetch --prune --no-recurse-submodules origin
'+refs/heads/*:refs/remotes/origin/*'`. There were 341 remote refs. Base remains
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 and main remains
ef3d907ecdc4c771b016f7d9c52372def057a340. The combined compiler/repository
ranking excludes their ports and conservatively excludes 117 ranked rule names
mentioned in Markdown claims on any origin branch. Of 55 available rules,
no-new-func, no-new-native-nonconstructor and no-new-wrappers were first, using
lexical ties. Each has zero ranked corpus findings. All preceding candidates were
skipped because they were already ported or mentioned in origin claims. The ref,
claim-blob and rule-name audit is preserved in
[validation-wave-26-fourth/selection.json](validation-wave-26-fourth/selection.json).

Each rule has its own `.a` file. The Function rule reports direct calls and
constructors plus static apply/bind/call invocations on the global Function,
reading through parentheses only. An outer call of the bound result is not a
second report. The nonconstructor rule reports the direct Symbol/BigInt callee
and deliberately does not unwrap parentheses, matching production Go. The wrapper
rule unwraps parentheses and reports the entire new String/Number/Boolean
expression. Assertions, non-null expressions, aliases and globalThis members
are not broadened into the production direct-identifier tests.

All three use binding-origin's actual ordered symbol declarations and the first
declaration's SourceFile.IsDeclarationFile, reproducing the Go global helper.
This is not a default-library-only check. Decisions, messages and report spans
are native Adamic. Existing binding-structure transports raw syntax, and
core_access supplies static member name semantics.

The new constructor-expression question returns the raw callee kind and exact
byte range. It lives in separate Go and Adamic files, with one switch-registration
line in shared facts.go. The native decoder checks framing and requires the
returned expression to be an immediate syntax child. The question never supplies
a lint verdict. Its protocol and ownership are documented in
[WAVE_26_FOURTH_FACTS.md](WAVE_26_FOURTH_FACTS.md).

No shared registration generator, test harness, profile compilation or suggestion
serializer was changed. No protected compiler implementation was touched. The
batch has its own runner, independent Go oracle and test file using existing
helpers. The native build supports .a, so no .ts implementation was needed.

## Agreement and mutants

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Generated controls | 36 | 33 | 13641 |
| Frozen repository | 287 | 0 | 18485 |
| TypeScript src/compiler | 77 | 0 | 5318 |

Ordinary native and ASan/UBSan/LeakSanitizer runs all match the independently
loaded, unchanged production Go rules, byte for byte. The oracle imports no
bridge implementation. Comparisons include file headings, exact UTF-8 ranges,
rule and message IDs, full descriptions, fix and suggestion counts and contents,
duplicate findings, and final totals. These three production rules emit no fixes
or suggestions; all positive rows have zero of both. Native comparison stderr
is empty. Positive controls produce 16 Function findings, 6 nonconstructor
findings and 11 wrapper findings, so zero corpus counts are not the sole evidence.

Controls cover direct calls and new expressions; apply/bind/call through dot,
string and template subscripts; bound results and optional calls; nested callee
parentheses; assertions and non-null wrappers; dynamic and concatenated keys;
local parameters, classes, aliases, imports and globalThis members; constructors
without argument lists; argument substitution; Unicode trivia and CRLF spans.
Their complete text, config and declarations are in controls.json. The direct
question contract test compares callee metadata with actual compiler Expression
pointers, including parenthesized and argumentless shapes. It rejects wrong
node kinds and additional request fields. Checker tests passed separately in
1.019s and within the bridge regression in 0.971s.

All following mutants compile, exit 0 and emit empty stderr. Only the independent
production finding-byte comparison detects the changed result:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| no-new-func | Remove bind from constructor method names | 2415 |
| no-new-native-nonconstructor | Report the whole new expression instead of its callee | 7033 |
| no-new-wrappers | Stop unwrapping callee parentheses | 9400 |
| constructor-expression | Substitute the first argument for the callee | 550 |
| binding-origin | Clear declaration-file flags | 56 |

A constructor-expression request after program release panics 70 with
`invalid or released checker handle`. A mutant retaining the registry entry
exits 0 and fails that required-refusal expectation.

The bridge regression holds 100 C ABI queries, buffers surviving release,
zero/stale handle rejection, distinct subsequent handles and Unicode output.
Its independent checker comparison covers 162 positions and 3261 identical bytes
under ASan/UBSan/LeakSanitizer. All seven existing mutants were caught: input
and output lengths plus one by ASan heap-buffer-overflow; retained released
handles by the stale-handle assertion; wrong source-file type by byte mismatch
at byte 6; removed link opt-in by refusal; omitted C-buffer frees and heap
allocation of a region result by LeakSanitizer. C sanitizers do not instrument
the Go heap.

The filtered Node oracle passed eight selected native/JavaScript/Node fixtures
and caught its one-byte mutant. This checks compiler regressions; it is not an
emitted-JavaScript run of the checker-linked rule runner.

Complete stream hashes, compressed canonical outputs, portable frozen manifests
and source hashes are committed in validation-wave-26-fourth. All 77 compiler
and 287 repository source hashes were reverified. Absolute path headings are
retained in canonical outputs, so relocating roots changes those headings.

## Native time against Go

After every build and regression test finished, three isolated alternating rounds
measured count-only whole-process throughput of this complete default suite.
All rounds returned zero corpus findings. Medians are seconds; independently
computed phase medians need not sum to the process median.

| Corpus / implementation | Load | Run | Whole process |
| --- | ---: | ---: | ---: |
| Compiler native | 0.365245 | 4.831807 | 5.256821 |
| Compiler Go | 0.374341 | 0.059145 | 0.464264 |
| Repository native | 0.094431 | 0.603853 | 0.730556 |
| Repository Go | 0.086771 | 0.073007 | 0.181414 |

Native takes about 11.32 times Go's compiler process time and 4.03 times its
repository process time. It issues 1750 compiler and 1286 repository questions,
including a complete syntax projection per root and callee/origin requests.
Complete raw AST transport and native decoding remain costly. This batch does
not establish Go-speed parity. Raw alternating rounds, query metrics and phase
stderr are preserved in measurements.json and benchmark.txt. Timings inside the
agreement test were not isolated and are not used for these performance claims.

## Commands and environment

Toolchain remains Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc reports 5.
The initial setup log is validation-wave-26/setup.txt: Go ready 0s,
clang/Node ready 1s, submodules 1s, cache warm 135s, done 135s. The persisted
toolchain was sourced from /workspace/adamic-tools/env.sh. Cohere remains pinned
at 715ba94f3608a6500086b1076ce5cb7e51b836db. TypeScript corpus remains v6.0.3
at 050880ce59e30b356b686bd3144efe24f875ebc8. No submodule pins changed.
Test output went directly to log files without pipes.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE26_FOURTH_ARTIFACTS=/workspace/wave-26-fourth-validation \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
TMPDIR=/workspace/wave-26-bridge-scratch \
go test ./stage1/cohere/typeaware -run '^TestWave26FourthAgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave-26-fourth-agreement.log 2>&1
# PASS 199.337s.
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave-26-fourth-checker.log 2>&1
# PASS 1.019s.
TMPDIR=/workspace/wave-26-bridge-scratch go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /tmp/wave-26-fourth-bridge.log 2>&1
# PASS bridge 125.755s, checker 0.971s.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout 10m > /tmp/wave-26-fourth-node.log 2>&1
# PASS 47.944s, eight selected cases and one-byte mutant.
go vet ./... > /tmp/wave-26-fourth-vet.log 2>&1
# Exit 0, empty output.
# gofmt -l cmd internal and the new Go files: exit 0, empty output.
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-26-fourth-validation/wave26 /workspace/wave-26-fourth-validation/wave26-oracle /workspace/wave-26-fourth-bench --corpus compiler /workspace/wave-26-typescript/src/compiler/tsconfig.json /workspace/wave-26-compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-26-repository.manifest > /tmp/wave-26-fourth-bench.log 2>&1
```

## Limits

The full repository gate and all previous type-aware suites were not rerun.
Touched bridge packages, the complete new default-rule differential suite and
filtered Node oracle passed. All upstream fixture permutations, arbitrary
JavaScript/JSX populations, suppression and edit application are not covered.
There are no configurable options or repairs for these three production rules.
The runner consumes bridge raw syntax rather than Adamic's independent parser,
so this batch does not prove that parser on these corpora.

The JavaScript backend has no checker bridge, so this rule runner's emitted
JavaScript could not be compared. No shared harness file was changed to work
around that gap. The pinned /workspace/wave-26-cohere --no-cache --no-fix command
on the five new .a files exits 1, refusing each as not TypeScript or JavaScript.
This is recorded in cohere-refusal.txt and is not a lint pass. Native builds,
full diagnostic comparisons, question contracts, mutants, released handles and
sanitizer checks are complete despite those integration limits.
