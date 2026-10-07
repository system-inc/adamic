Built no-new-func, no-new-native-nonconstructor and no-new-wrappers in separate native .a files.
Claim 1bb22f33 was pushed before code; implementation is a63e3ffc, after pushed completion c78d9bf9.
TestWave18ConstructorAgreementAndMutants passed in 78.786s: 65 findings over 96 controls plus their helper, both corpora, and sanitizers.
Each rule mutant compiled, exited 0 with empty stderr, and was caught by complete Go bytes; stale registry and Node one-byte mutants were caught too.
Not covered: full repository gate, emitted-JavaScript lint-rule comparison, error-recovery syntax, and integration of the original Next visitor's JSX parser dependency.

## Ownership and implementation

Work continues on codex/typeaware-wave-18. Before this claim, the prior original and core
rule sources and evidence were pushed through c78d9bf9. The Next visitor's published native
JSX parser dependency remains the precise shared integration gap documented in
WAVE_18_TITLE_REPORT.md; no shared parser edits were made here.

The fresh audit fetched 341 origin refs, checked 33 distinct claim Markdown blobs and 117
claimed rule names, and excluded 25 ranked ports on origin/main
(ef3d907ecdc4c771b016f7d9c52372def057a340) and origin/codex/tsgo-c-library
(5afbdb83da2ed7ad9815657cd3f6ececd5294bf6). The ranking combines the two VOLUME_REPORT
count populations with lexical ties. These are the first three remaining names, each with zero
recorded compiler/repository findings. Audit code and full selection evidence are saved outside
claims, in validation-wave-18-constructor. The claim push succeeded before implementation.

NoNewFunc handles new and direct calls of Function, plus apply/bind/call reached by dots,
quoted subscripts and no-substitution templates. It reproduces parentheses, optional access,
shadowing, qualified-receiver exclusions, and reporting on the invoking expression rather than
an outer call of the result. NoNewNativeNonconstructor handles Symbol and BigInt, reports the
callee identifier, and deliberately does not strip callee parentheses, matching production Go.
NoNewWrappers handles String, Number and Boolean, strips callee parentheses, and reports the
whole new expression. Dynamic method names and qualified constructor names stay silent.

ConstructorBindingFacts decodes the existing node-symbol-details question. The exact Go
predicate requires a symbol with declarations and checks the FIRST declaration's file, not all
declarations and not default-library membership. This differs from the binary rule's global
predicate in the previous batch, so this decoder is separate. Raw compiler metadata crosses
the bridge; Adamic owns every lint decision, message and report span. No new bridge question,
shared registration line, shared harness edit, parser edit or protected compiler edit was needed.
All implementation files are .a. The driver and tests are isolated new files in this wave's scope.

## Agreement and checks

The independent Go executable imports no bridge code. It uses its own checker program and
AST walk and calls the unchanged production registry rules with their production program views.
The complete protocol compares ranges, rule names, message IDs and descriptions, every fix,
and each suggestion's message and edits. These three production rules have no fixes or
suggestions; their zero fields remain part of each comparison.

Controls preserve source strings extracted from the production Go tables, including repeated
span cases, and add Unicode/CRLF, nested shadows, imported names, optional/property access,
parenthesized callees and receivers, multiple findings and dynamic method names. A helper
module is an explicit root. There are 96 control sources plus that helper, not 96 distinct
upstream scenarios. Separate production Go tests pass in 0.073s. Positive finding counts are
31 no-new-func, 10 no-new-native-nonconstructor and 24 no-new-wrappers.

No-library controls explicitly require zero findings rather than silently treating absent globals
as proven globals. Custom non-default-library ambient declarations restore all 65 findings,
proving that declaration-file origin, rather than default-library origin, is the production test.
All three configurations run normally and under ASan/UBSan with empty native stderr.

| Population | Findings | Identical bytes, normal and ASan/UBSan |
| --- | ---: | ---: |
| Default controls and helper, 97 roots | 65 | 28763 |
| No-library controls and helper | 0 | 5936 |
| Custom ambient controls and helper | 65 | 28763 |
| Frozen repository, 287 files | 0 | 18485 |
| TypeScript compiler, 77 files | 0 | 5318 |

Linux sanitizer runs include leak checking; Go heap allocations are not ASan-instrumented.
The released checker probe exits 70 with exactly
`adamic: panic: invalid or released checker handle`. Removing registry deletion in an isolated
Go overlay makes the same probe exit 0 with empty stderr, and the lifetime expectation catches it.
No shared bridge file is modified by that experiment.

| Mutant | Execution | Independent catch |
| --- | --- | --- |
| no-new-func indirect receiver global predicate reversed | Exit 0, empty stderr | Go bytes, first difference 1128 |
| no-new-native-nonconstructor reports whole new instead of callee | Exit 0, empty stderr | Go bytes, first difference 11577 |
| no-new-wrappers global predicate reversed | Exit 0, empty stderr | Go bytes, first difference 13462 |
| Released registry deletion removed | Exit 0, empty stderr | Expected stale-handle refusal 70 |
| Existing Node one-byte result mutant | Builds and runs | Independent Node oracle |

The filtered Node oracle passes in 11.704s on closures, method_closures, generic_functions,
regions and regions_throw, with native sanitizers, emitted JavaScript and the one-byte mutant.
It validates the compiler foundation, not an emitted-JavaScript lint-rule comparison.
Go vet passes with empty output; new Go files are gofmt-clean. All five new .a source files
are formatted with the pinned cohere native formatter and stable on a second pass. The shared
.a-aware CLI lint/profile gate is not claimed here. No full repository Go test gate was run.

## Observed timings and reproduction

Toolchain setup refresh succeeds: Go, clang, Node and submodules ready in 0s; cache warm
17s; setup done in 17s. nproc is 5, cgroup CPU quota is four cores, memory 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0. No submodule pins change. The compiler corpus stays
TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Both frozen manifests,
source hashes, raw diagnostic streams and command stdout/stderr are preserved in the evidence.

Timing rounds inside the validation test overlap another verification process and are retained
only as observations. The table uses a separate quiet run after builds and tests ended: three
alternating native/Go rounds, whole-process medians in seconds, with loading and traversal.

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.201842 | 0.115793 | 1.74 |
| Compiler | 1.396603 | 0.302121 | 4.62 |

Native makes zero checker queries on the repository and two on the compiler. Neither corpus
has findings for these rules; native remains slower, and no broader speed-parity claim is made.
Quiet round stdout/stderr and measurements.json are saved separately from validation timings.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE18_CONSTRUCTOR_ARTIFACTS=/workspace/wave-18-constructor-validation \
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest \
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave18ConstructorAgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave-18-batch3-test.log 2>&1
# In the pinned cohere checkout:
go test ./internal/lint/rules/core -run '^TestNoNew(Func|NativeNonconstructor|Wrappers)' \
  -count=1 -v > /tmp/wave-18-batch3-production.log 2>&1
# Back in Adamic:
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  -count=1 -timeout=15m -v > /tmp/wave-18-batch3-node-oracle.log 2>&1
go vet ./... > /tmp/wave-18-batch3-vet.log 2>&1
```

No further rules are claimed in this turn. No pull request is opened. Prior default-option
scope limits remain as recorded in the earlier wave reports; no additional current-rule shared
harness gap blocks these three native ports.
