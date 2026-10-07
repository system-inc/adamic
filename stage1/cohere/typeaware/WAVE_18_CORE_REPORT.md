Built no-class-assign, no-const-assign and no-constant-binary-expression as native .a visitors.
Claim 1a7f2a3a was pushed before code; implementation is 1cba7ee7, after original completion 1b6147b0.
Comparison passed in 69.447s: 199 controls, 145 default findings, 150 enabled-option findings, and both frozen corpora.
Three rule mutants were caught only by Go byte comparison; the stale-handle registry mutant and Node one-byte mutant were caught too.
Not covered: the full repository gate, emitted-JavaScript lint-rule comparison, and integration of the original Next visitor's shared JSX parser dependency.

## Ownership and scope

On codex/typeaware-wave-18, the original Next visitor and its evidence were pushed in 1b6147b0
before selecting another batch. The published native JSX slice is a dependency, as recorded in
WAVE_18_TITLE_REPORT.md. Original await-thenable and class-literal-property-style default ports
and their bridge questions remain as previously committed and tested in WAVE_18_REPORT.md.

The selection audit fetched 328 origin refs and checked 33 distinct claim Markdown blobs, 105
claimed rule names, and native ports on origin/main ef3d907ecdc4c771b016f7d9c52372def057a340
and origin/codex/tsgo-c-library 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. The combined
compiler/repository by-volume population contains 197 rules; 25 ranked rules are already ported
on those baselines (the 26th baseline port is outside this checker-dependent ranking).
The first remaining names are these three core rules, all with zero recorded corpus findings.
The withdrawn Nexus candidates are claimed by other workers and were skipped. Selection code
and the full audit are preserved in validation-wave-18-core, outside the claims directory.

No shared harness, registration generator, parser, submodule pin or protected compiler file is
changed by this batch. Each rule has its own .a source. CoreBindingFacts only decodes existing
binding-declarations and node-symbol-details questions; no bridge extension or registration is
needed. Name resolution and declaration metadata cross the bridge, and Adamic makes the lint
decisions on its native AST. The isolated driver and test use existing build/comparison helpers.

## Behavior and evidence

NoClassAssign anchors each declared class name and reports structurally proven writes whose
checker declaration identity matches, preserving class-expression scope and shadows.
NoConstAssign gathers all names in const, using and await-using declarations, including nested
binding patterns, then reports each matching write once. Both reuse the existing native write
climb for compound assignments, increments, rest targets, property/default reads and for-heads.
NoConstantBinaryExpression separately judges truthiness, nullishness, strict/loose boolean
comparison and fresh object identity; global names are distinguished from source-file shadows.
The optional relational arm is implemented and independently compared with its Go options on.

The independent Go executable calls unchanged production rules through the production registry,
loads its own TypeScript program and walks its own AST. It imports no bridge code. Its serializer
compares ranges, rule names, message IDs and complete descriptions, all fixes, and all suggestions
and suggestion edits. These three production rules emit no fixes or suggestions; their zero
fields are still part of every compared diagnostic.

Controls preserve extracted source strings from the production Go fixtures and add Unicode/CRLF,
nested shadows, rest/default targets, templates, constructor/global shadows, conditional objects,
sparse arrays, numeric identities and relational comparisons. The 199 controls include repeated
production source rows and identifier-only negative rows from fixture data; this is not a claim
of 199 distinct upstream scenarios. Default positive counts are 36 class, 61 const and 48 binary.
The enabled relational arm adds five findings. Production Go fixtures pass separately in 0.080s.

A boundary comparison initially failed for `(a || 0n) || b`: Go's numericLiteralSign scans the
retained `n` suffix and therefore treats even `0n` as a truthy logical identity. The native port
now reproduces this production quirk explicitly, and the boundary control passes. This reports
observed parity, not a claim that that judgment is JavaScript's runtime truthiness.

| Population | Findings | Identical bytes, normal and ASan/UBSan |
| --- | ---: | ---: |
| Default controls, 199 files | 145 | 77967 |
| Relational option controls, 199 files | 150 | 79357 |
| Frozen repository, 287 files | 0 | 18485 |
| TypeScript compiler, 77 files | 0 | 5318 |

All normal and sanitized native runs exit 0 with empty stderr. Linux ASan includes leak checking;
UBSan uses no recovery. The stale binding-declarations handle probe exits 70 with exactly
`adamic: panic: invalid or released checker handle`. Removing the registry deletion makes that
same probe exit 0 with empty stderr, proving the lifetime check detects the registry defect.

| Mutant | Successful mutant execution | Catch |
| --- | --- | --- |
| Class declaration identity equality reversed | Exit 0, empty stderr | Complete Go comparison, byte 60 |
| Const anchor membership reversed | Exit 0, empty stderr | Complete Go comparison, byte 573 |
| Binary nullishness predicate reversed | Exit 0, empty stderr | Complete Go comparison, byte 42804 |
| Released-handle registry deletion removed in an overlay | Exit 0, empty stderr | Expected stale-handle exit 70 |
| Existing Node one-byte result mutant | Compiles and runs | Independent Node oracle |

The filtered compiler oracle passed in 12.692s on closures, method_closures, generic_functions,
regions and regions_throw, comparing native and emitted JavaScript against Node, including its
one-byte mutant. This checks the compiler foundation, not an emitted-JavaScript lint-rule suite.
Go vet passed with empty output. All new Go sources are gofmt-clean, and all new Adamic sources
were formatted with the pinned cohere native formatter and proven stable on a second pass.
The shared .a-aware CLI lint gate is not claimed here.

## Timing and reproduction

Setup refresh printed Go, clang, Node and submodules ready in 0s, cache warming 17s and total
17s; nproc is 5 with a four-CPU cgroup quota and 17.6 GB memory. The earlier initial setup took
76s. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. The TypeScript compiler corpus remains
v6.0.3 at the previously pinned 050880ce59e30b356b686bd3144efe24f875ebc8. Frozen manifests,
source hashes, command stdout/stderr and raw checker load/query/run timings are preserved.

Whole-process medians over three alternating native/Go rounds, in seconds:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.344546 | 0.182503 | 1.89 |
| Compiler | 1.949142 | 0.436043 | 4.47 |

Both include program loading and their own traversal. Native makes 3945 checker queries on the
repository and 20005 on the compiler. Native is slower in this observation; these zero-finding
corpora do not establish broad performance parity. No timings from failed comparisons are used.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE18_CORE_ARTIFACTS=/workspace/wave-18-core-validation-final \
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest \
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave18CoreAgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave-18-core-final-test.log 2>&1
# In the pinned cohere checkout:
go test ./internal/lint/rules/core -run '^TestNo(ClassAssign|ConstAssign|ConstantBinaryExpression)' \
  -count=1 -v > /tmp/wave-18-core-upstream-test.log 2>&1
# Back in Adamic:
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  -count=1 -timeout=15m -v > /tmp/wave-18-core-node-oracle.log 2>&1
go vet ./... > /tmp/wave-18-core-vet.log 2>&1
```

No additional batch is claimed. No pull request is opened. The original Next rule still needs its
published JSX parser dependency integrated into the base parser; its isolated native agreement,
sanitizers, controls and mutant are already pushed and documented separately. Nondefault
class-literal-property-style options remain outside the previously reported default port.
