Built: no-new-func, no-new-native-nonconstructor and no-new-wrappers, each in its own .a file.
Commits: previous nine ports pushed at 2203c2f0; claim 34e1e65a pushed before code; implementation 2cc0d773 pushed on codex/typeaware-wave-23.
Commands and outputs: dedicated gate PASS 92.266s; bridge PASS 76.821s; checker PASS 0.165s; filtered Node oracle PASS 20.219s; vet, formatting and whitespace logs empty.
Mutants: three rule mutants exit 0 with empty stderr and are caught only by Go finding bytes; released registry caught by refusal; seven foundation mutants and the Node byte mutant caught.
Not covered: full repository test gate, emitted-JavaScript execution of these checker-linked rules, all compiler configurations, malformed/error-recovery source, suppression/edit application or cohere CLI support for .a inputs.

The previous nine ports were fully tested and pushed before selection. The first
push reported everything up to date. Fetch explicitly requested every origin
head, because the configured default refspec includes only main. Selection
checked 341 fetched origin refs, the same 26 production ports on main and the
bridge branch, and 117 claimed ranked rules across 33 distinct Markdown claim
blobs. These were the first three of 55 remaining entries in VOLUME_REPORT's
combined descending count ranking, with lexical ties. All original counts are
zero. Already-claimed candidates were skipped, including no-global-assign,
no-implicit-globals, no-implied-eval, no-import-assign, no-invalid-regexp,
no-label-var and no-misleading-character-class. There was no conflicting claim
for these three. [selection.json](validation-wave23-constructor/selection.json)
preserves the exact refs, claim texts, exclusions and remaining ranking.

Main remains ef3d907ecdc4c771b016f7d9c52372def057a340 and the bridge branch
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go remains
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the external TypeScript v6.0.3
corpus remains 050880ce59e30b356b686bd3144efe24f875ebc8. Work continues on the
same branch, originally based on 0d540f4. No submodule pin changed.

[no_new_func.a](no_new_func.a) checks direct Function calls/construction and the
exact apply/bind/call method set, including static string/template accesses and
optional calls. It skips parentheses on callee and receiver, preserves checker
shadowing on both arms, and reports the invoking expression. For a bound result
that is immediately called, that is the inner bind call, not the outer call.
References to Function.bind and nonconstructing methods remain quiet.

[no_new_native_nonconstructor.a](no_new_native_nonconstructor.a) checks only
unwrapped Identifier callees Symbol and BigInt and reports that identifier.
Parenthesized callees remain quiet because that is Go cohere's decision.
[no_new_wrappers.a](no_new_wrappers.a) skips parentheses before checking String,
Number and Boolean, and reports the whole new expression. Both preserve global
resolution, local/import shadows, exact message interpolation and case sensitivity.
Calls without new remain quiet. These rules have no configurable options, fixes
or suggestions in production.

All decisions and messages are native Adamic. Existing
[core_references.a](core_references.a) supplies raw symbol declaration metadata
and static member syntax. A global means the first declaration is in a declaration
file, not exclusively a bundled library. No new checker question or registration
edit was needed. A noLib configuration with custom .d.ts declarations produces
seven positive findings, proving that origin distinction. New Adamic source
files are .a and compile through the existing native loader. The shared harness,
registration generator, bridge implementation and protected compiler files are
untouched. No shared-harness gap blocks these native ports.

[wave_23_constructor_suite.a](wave_23_constructor_suite.a) loads one checker
program, parses each root natively, runs all three rules and emits sorted full
canonical diagnostics while retaining duplicates. The independent Go oracle
loads its own program and invokes the unchanged pinned production registry rules.
It imports no bridge code. Every finding field, fix and suggestion is compared;
all populations have zero repairs and suggestions, matching production.

| Population | Roots | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| ESNext controls | 89 | 54 | 24,979 |
| Custom declarations, noLib | 1 | 7 | 2,289 |
| TypeScript compiler | 77 | 0 | 5,318 |
| Frozen repository | 287 | 0 | 18,485 |

Controls comprise 88 executable sources and one imported helper module, all .a.
Sources are extracted from the pinned Go fixture tables and supplemented with
all accepted names/methods, nested calls, parentheses, optional invocation,
static templates, dynamic/computed negative keys, aliases, nonidentifier
receivers, source declarations, parameter/class/function/import shadows, new
expressions without argument lists, assertion wrappers, Unicode and CRLF.
The 54 findings split into 32 Function, 8 nonconstructor and 14 wrapper findings.
The custom declaration control adds 2, 2 and 3 respectively. The two frozen
corpus manifests remain unchanged from prior ports. Their empty findings are
supported by positive controls and live mutants, not treated alone as proof.

| Mutant | Observation and catcher |
| --- | --- |
| Function: remove the indirect receiver's global-origin check | Exit 0, empty stderr; Go bytes differ at 8,911 |
| Nonconstructor: extend the reported callee end to the new-expression end | Exit 0, empty stderr; Go bytes differ at 1,837 |
| Wrapper: shorten the new-expression end to the callee end | Exit 0, empty stderr; Go bytes differ at 122 |
| Registry: keep a released program live | Mutant exits 0, empty stderr; healthy program exits 70 with the required invalid-or-released-handle panic |

No build failure or crash counts as a rule-mutant kill. The foundation separately
catches input and output length mutations with ASan, removed C-buffer frees and
heap-instead-of-region allocation with LeakSanitizer, retained released entries
with stale-handle assertions, source-file instead of node type lookup with the
Go byte oracle at byte 6, and a removed link opt-in guard with the required
refusal. It compares 1,600 positions across four compiler files, 54,982 identical
bytes under ASan/UBSan/LSan, and checks 100 C ABI queries, buffer survival after
release, zero/stale handles and distinct second handles. Sanitizers instrument
native/C memory, not the Go heap. The filtered Node gate selects six fixtures,
compares native, Node and emitted JavaScript behavior, and catches its deliberate
one-byte oracle mismatch. That is compiler regression evidence, not JavaScript
execution of these checker-linked lint rules.

Three quiet alternating whole-process rounds, with no builds or tests active,
include program load, parsing, decisions, rendering and teardown. Complete finding
streams are compared on every round:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.567 s | 0.320 s | 4.89 |
| Repository | 0.217 s | 0.124 s | 1.75 |

Native query counts are 2 and 0 respectively. Native is slower on this machine;
these measurements do not establish why. Build and sanitizer times are excluded.
[measurements.json](validation-wave23-constructor/measurements.json) preserves
all rounds and phase logs.

The existing successful setup was reused: Go 1.27.1 ready 0s; clang 20.1.8 ready
0s; Node 24.19.0 ready 0s; submodules 0s; cache warm 83s; setup done 83s.
The original setup log is preserved. nproc was rechecked and reports 5.
All commands sourced /workspace/adamic-tools/env.sh. Every test output went to a
log file, never a pipe. Exact gates:

```sh
ADAMIC_WAVE23_CONSTRUCTOR_ARTIFACTS=/workspace/wave-23/constructor-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave23ConstructorAgreementAndMutants$' -count=1 -timeout=15m -v > /workspace/wave-23/constructor-final/gate.log 2>&1

ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/constructor-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/constructor-node.log 2>&1
go vet ./... > /workspace/wave-23/constructor-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/constructor-gofmt.log 2>&1
git diff --check > /workspace/wave-23/constructor-diff.log 2>&1

python3 stage1/cohere/typeaware/validation-wave23-constructor/measure.py /workspace/wave-23/constructor-final/native /workspace/wave-23/constructor-final/wave23-constructor-oracle /workspace/wave-23/typescript /workspace/wave-23/constructor-bench /workspace/adamic > /workspace/wave-23/constructor-bench.log 2>&1
```

Full compressed outputs, matching hashes, root/config hashes, portable manifests,
fixture sources, mutant outputs and logs are in
[validation-wave23-constructor](validation-wave23-constructor). Existing shared
code did not change, so the prior nine ports retain their pushed regression
evidence. The full go test ./... gate was not run. The pinned cohere CLI's .a
extension refusal remains recorded in the prior report; no new CLI lint pass is
claimed. Shared emitted-JavaScript/profile integration remains owned by
codex/lint-harness-dot-a. No PR was opened; no further rules were claimed.
