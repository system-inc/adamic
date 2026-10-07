Built: no-eval, no-extend-native and no-func-assign in separate .a files, including allowIndirect and native-prototype exceptions.
Commits: previous six ports and reports pushed at 6b1e4e2f; new claim 2c15df8c pushed before code; implementation 92b75cd0 pushed on codex/typeaware-wave-23.
Commands and outputs: final dedicated gate PASS 82.153s; bridge PASS 84.309s; checker PASS 0.207s; filtered Node oracle PASS 19.735s; vet, formatting and diff-check logs empty.
Mutants: five rule/option mutants caught only by Go diagnostic bytes; released-registry mutant caught by refusal; seven bridge foundation mutants and the Node one-byte mutant caught.
Not covered: full repository test gate, emitted-JavaScript execution of these checker-linked rules, all option/compiler combinations, suppression/edit application, or cohere CLI support for .a inputs.

The previous six rules were finished, tested and pushed before this continuation.
`git push` first reported everything up to date. An explicit all-heads fetch was
needed because the remote's default fetch refspec only includes main. Selection
checked 330 fetched origin refs, 26 existing ports on main and the bridge branch,
and 108 claimed ranked rules across 33 distinct Markdown claim blobs. These three
were the first of 64 remaining entries in the combined VOLUME_REPORT ranking,
descending by volume with lexical ties; their original counts are all zero.
No conflicting claim was found. Binary artifacts in other workers' claim
folders were excluded from Markdown scanning. The exact snapshot, claim text,
exclusions and remaining ranking are in [selection.json](validation-wave23-core/selection.json).
No further rules are claimed in this continuation.

Main remains ef3d907ecdc4c771b016f7d9c52372def057a340 and the bridge branch
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Cohere remains pinned to
715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go to
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the external TypeScript v6.0.3
corpus to 050880ce59e30b356b686bd3144efe24f875ebc8. The branch continues from
its original 0d540f4 base. No submodule pin changed.

[no_eval.a](no_eval.a) preserves the Go decisions for direct calls, wrappers,
optional calls, indirect values and static global-object chains. Direct calls
still report when eval is locally shadowed. With allowIndirect, optional calls
and all indirect accesses are quiet. Reports use the outer callee span for direct
calls and the property-name/argument span for global-object accesses. Go's
implementation leaves this.eval quiet and ignores shorthand eval properties;
the port preserves those decisions.

[no_extend_native.a](no_extend_native.a) uses Go's exact 49-name set, all assignment
operators, the first argument of Object.defineProperty/defineProperties and the
whole assignment/call span. It preserves shadowing, exception filtering and the
quiet cases for updates, deletion, deeper properties, aliases and Reflect.
[no_func_assign.a](no_func_assign.a) anchors declarations and named expressions
on the checker's first same-file declaration, then checks structural writes
against declaration identity. It handles overloads, hoisting, shadows,
destructuring, shorthand assignment values and for-in/of heads. A merged variable
and function declaration also matches Go's declaration-order decision.

The native decisions stay in Adamic. [core_references.a](core_references.a)
decodes existing raw symbol declaration metadata and static member syntax.
For these core rules, a global identifier means its first declaration is in
any declaration file, not exclusively a bundled library. A separate noLib
control with custom .d.ts declarations proves this distinction. No new bridge
question or registration edit was needed. All new Adamic implementation files
are .a; the current native loader compiled them successfully. The only new Go
files are the dedicated validation test and independent production-rule oracle.
The shared harness, registration generator and protected compiler files are
untouched. There is no shared-harness blocker for these native ports.

[wave_23_core_suite.a](wave_23_core_suite.a) loads one checker program and runs
the three rules over each native parsed root. The independent Go oracle invokes
the unchanged pinned production registry rules and imports no bridge code.
Both serialize complete canonical finding fields, ordered fixes and complete
suggestions, sort per file and retain duplicates. These rules have no fixes or
suggestions in production; every compared population has zero of both.

| Population | Roots | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| ESNext controls | 258 | 208 | 124,610 |
| Custom declarations, noLib | 1 | 5 | 2,725 |
| TypeScript compiler | 77 | 0 | 5,318 |
| Frozen repository | 287 | 0 | 18,485 |

The 208 control findings split into 54 eval, 95 native-extension and 59 function
assignment findings. The controls extract source strings from the pinned Go
fixture tables, then add all 49 protected builtin names, nine lowercase/unknown
negative names, every assignment operator, wrappers, optional calls, shadowed
bindings, overloads, merged declarations, initializer references, global-object
chains, destructuring, Unicode and CRLF. Fixture strings run under the runner's
configurations rather than asserting every original row's options and expected
IDs. The frozen repository and compiler populations are unchanged from the
previous six ports. Empty corpus findings are supported by positive controls
and live byte mutants, not relied on alone.

Four option configurations also match in normal and sanitizer runs:

| Options | Findings | Identical bytes |
| --- | ---: | ---: |
| allowIndirect | 169 | 102,625 |
| exceptions Array and Object | 168 | 107,099 |
| unknown exception | 208 | 124,610 |
| allowIndirect; exceptions String and Array | 137 | 88,619 |

| Mutant | Observation and catcher |
| --- | --- |
| Eval: report unwrapped callee rather than outer callee | Exit 0, empty stderr; Go bytes differ at 4,110 |
| Native extension: substitute Object in every message | Exit 0, empty stderr; Go bytes differ at 8,768 |
| Function assignment: accept any resolved declaration of a matching name | Exit 0, empty stderr; Go bytes differ at 71,837 |
| Eval option: ignore the optional-call exclusion | Exit 0, empty stderr; Go bytes differ at 50,096 |
| Native exception option: ignore the exception list | Exit 0, empty stderr; Go bytes differ at 3,130 |
| Released registry: retain a released program | Mutant exits 0, empty stderr; healthy program exits 70 with the required invalid-or-released-handle panic |

No compilation failure, crash or sanitizer failure counts as a rule-mutant kill.
The foundation gate separately catches input and output length mutants with
ASan, an omitted C free and heap-instead-of-region allocation with LeakSanitizer,
a retained registry entry with a stale-handle assertion, source-file instead of
node type lookup with a byte mismatch at 6, and a removed link guard with the
required refusal test. Its external corpus compares 1,600 positions over four
compiler files, 54,982 identical bytes under ASan/UBSan/LSan. It also checks 100
C ABI queries, output survival after release, zero/stale handles and distinct
program handles. Sanitizers instrument native/C memory, not the external Go heap.
The filtered Node gate compares six compiler fixtures against Node and emitted
JavaScript, and catches its deliberate one-byte oracle mismatch. That is compiler
regression evidence, not JavaScript execution of these checker-linked lint ports.

Three quiet alternating whole-process runs, with no builds or tests active,
include program load, parsing, rule decisions, rendering and teardown, and
compare the complete finding stream on every round:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.225 s | 0.399 s | 8.08 |
| Repository | 0.283 s | 0.126 s | 2.25 |

Native query counts are 9,190 and 633 respectively. Native is slower on this
machine. No causal attribution is inferred from those timing measurements.
Builds and sanitizers are excluded from this quiet comparison. Full rounds and
phase logs are in [measurements.json](validation-wave23-core/measurements.json).

All commands sourced /workspace/adamic-tools/env.sh. The existing setup was
reused: Go 1.27.1 ready 0s; clang 20.1.8 ready 0s; Node 24.19.0 ready 0s;
submodules 0s; cache warm 83s; done 83s. The original setup log is preserved.
`nproc` was rechecked and reports 5. Every test writes output to a log file.
The precise gates were:

```sh
ADAMIC_WAVE23_CORE_ARTIFACTS=/workspace/wave-23/core-complete \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave23CoreAgreementAndMutants$' -count=1 -timeout=15m -v > /workspace/wave-23/core-complete/gate.log 2>&1

ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/core-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/core-node.log 2>&1
go vet ./... > /workspace/wave-23/core-vet-final.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/core-gofmt-final.log 2>&1
git diff --check > /workspace/wave-23/core-diff-check.log 2>&1

python3 stage1/cohere/typeaware/validation-wave23-core/measure.py /workspace/wave-23/core-complete/native /workspace/wave-23/core-complete/wave23-core-oracle /workspace/wave-23/typescript /workspace/wave-23/core-bench /workspace/adamic > /workspace/wave-23/core-bench.log 2>&1
```

The first default-only gate passed in 111.080s while foundation checks ran.
After option support and expanded controls, the next gate passed in 74.712s.
The final gate adds custom declarations and two option mutants and passes in
82.153s. These logs, compressed full outputs, matching hashes, root source hashes,
portable manifests, full fixture sources, mutants and measurements are preserved
in [validation-wave23-core](validation-wave23-core).

No existing bridge or shared compiler implementation changed during this
continuation, so the previous six ports retain their pushed regression evidence.
The full `go test ./...` gate was not run. The pinned cohere CLI's .a extension
refusal remains recorded in the previous report; no new CLI lint pass is claimed.
Emitted-JavaScript execution of checker-linked rules and shared profile work
remain owned by codex/lint-harness-dot-a. No PR was opened.
