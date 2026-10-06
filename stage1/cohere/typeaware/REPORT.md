Native Adamic now implements cohere's no-unsafe-unary-minus rule.
The scanner/parser branch is merged; all 77 compiler files match Go cohere.
Compiler corpus: 2 findings, 245 queries; generated probes: 36 findings, 61 queries.
Median native throughput: 1.263 compiler findings/s; warmed API query: 4.518 us.
Wrong-node and released-program mutants are caught; all focused checks pass.

# Milestone 5: the first type-aware rule

Implementation commit: `4e77bbc66a80bd773bd0f8ab6d54be58f2f3b6c5`.
Continued `codex/tsgo-c-library` from foundation `ae82c7a`. The requested
`origin/codex/typescript-scanner` head, `6636e8f`, was merged cleanly in
`ed8c201c491f6a55ac29222069a0d9f423b77fc9`. That brings both the whole-file
parser and the existing syntax-only cohere slice. No protected compiler file
was edited. No pull request opened.

The implementation is in `unary_minus.ts`, driven by `main.ts`. It traverses
the actual Adamic parser's SourceFile tree, queries each unary minus operand
once, and decides from the constrained union parts' flags. The first unsafe
part supplies the exact TypeToString text in cohere's message. Finding spans
cover the unary expression, trimmed by the Adamic scanner. Stable sort retains
source order. There are no options, fixes or suggestions for this rule.

One additive bridge function, `tsgo_type_parts` / `tsgoTypeParts`, exposes
exact node selection, base constraint resolution, ordered union parts, raw
TypeFlags and TypeToString. No unary-minus predicate or findings are computed
in the bridge. Inputs and output keep the existing explicit UTF-8 lengths,
C ownership, monotonic handle registry and serialization. The result is copied
to an owned Adamic string and its C buffer freed. Frame lengths count UTF-16
units inside the explicit UTF-8 buffer, so Adamic can consume arbitrary type
text without separator escaping. Empty/malformed frames panic before a finding.
A pinned-flags check holds the Adamic mask to the Go checker's enum.

## Independent findings oracle

The overlay driver in `testdata/oracle.go` builds inside cohere and calls its
unchanged `NoUnsafeUnaryMinus.Run`. Its loader and AST walk are independent of
the bridge. It imports no bridge implementation and contains no copied rule
predicate. Both sides use the same original compiler config and explicit roots.

Pins:

- cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- typescript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- Original TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Every .ts file recursively under `src/compiler` is compared, including
`_namespaces`: 77 files. Canonical bytes include file headers, UTF-8 start/end,
rule name, ID, exact escaped message, fix/suggestion counts, total findings and
total queries. Native uses an ASan/UBSan-instrumented C boundary and runtime;
normal completion also passes Linux's leak check.

| Corpus | Files | Findings | Queries | Identical bytes |
| --- | ---: | ---: | ---: | ---: |
| TypeScript compiler | 77 | 2 | 245 | 4,959 |
| Table-derived and additional probes | 53 | 36 | 61 | 7,782 |

Byte counts include absolute path headers and therefore depend on the scratch
path. The two compiler findings are in `checker.ts`, line 9011 (`-name`,
bytes 503384:503389) and line 39994
(`-(node.operand as NumericLiteral).text`, bytes 2364455:2364493).
Both messages name `string`. These are observations from Go cohere, not a
claim that the compiler's deliberate coercions need changing.

Generated probes cover the production rule's tables, number/bigint unions,
any/never, generic constraints, boolean's first union part, unknown/void,
boxed primitives, enum members, parentheses, nested unary minus, calls,
property/element access, flow narrowing, imports, CRLF and Unicode/non-BMP text.
The generated nonempty controls prevent a vacuous all-clean result.

## Mutants and failure checks

All mutants build successfully. Finding mutants finish normally; no compiler
or sanitizer failure is counted as an oracle catch.

| Mutant or probe | What catches it |
| --- | --- |
| Ask the unary expression's type instead of its operand's | Go cohere byte oracle: first mismatch byte 258; mutant emits 0 instead of 36 findings |
| Query a released program | Native panic, exit 70: `invalid or released checker handle` |
| Remove registry deletion on release | Released-query expectation detects a normal exit 0 instead of 70 |
| Return an empty parts buffer | Exit 70: `checker returned no type parts` |
| Return a missing frame header | Exit 70: `invalid checker type frame` |
| Return a negative frame length | Exit 70: `invalid checker type length or flags` |
| Query a nonexistent exact kind/span | Native panic: `no exact Identifier node` |
| Ignore the requested kind during lookup | Exact-node refusal expectation detects exit 0 |
| Use the new library call without linking the archive | Lowering refusal: `unlinked typescript-go library call` |

The foundation's seven mutants also pass unchanged in behavior: input/output
length +1 caught by ASan; stale handle assertion; wrong-position type oracle;
removed link guard; missing C output free caught by LSan; region adapter heap
allocation caught by LSan. Its 1,600-position oracle still matches 54,982 bytes.

Two implementation findings are recorded rather than hidden. The first driver
attempt hit the parser worker's documented nested-constructor lowering gap; the
driver now constructs its Parser/Scanner before constructing the linter.
Also, typescript-go's GetNodeAtPosition deliberately skips token kinds,
including Identifier. Exact lookup now descends the AST by kind and span and
includes tokens. The existing foundation regression initially failed because
its wrong-position mutant's search text acquired a second occurrence; its
anchor was narrowed to the original result expression, then every mutant passed.

## Measurements

Three alternating rounds, release native and Go binaries, no other test suite
running. CPU: AMD EPYC 7763, Linux amd64, `nproc=5`, cgroup quota 4 CPUs.
Load before/after: `0.62 2.71 2.52` / `0.65 2.68 2.51`.
Use medians; all samples are in [timing.log](validation/timing.log).

Whole-process count mode includes startup, file reads, program load, AST
traversal, Adamic parsing/offset mapping and linting. Compilation, findings
printing and sanitizer overhead are excluded.

| Corpus | Native seconds | Go seconds | Native findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| Compiler, 2 findings | 1.584099 | 0.353968 | 1.263 | 5.650 |
| Generated, 36 findings | 0.022129 | 0.021937 | 1,626.833 | 1,641.088 |

Native is 4.48 times slower on the compiler corpus. The generated corpus is
near parity within observed variation. No speedup is claimed. For the median
compiler native run, load takes 211.931 ms, all 245 API queries 98.773 ms, of
which the first is 87.663 ms. About 1.27 seconds is outside those recorded API
intervals; no profiling was done to attribute that remainder.

A separate probe performs 10,000 repeated queries of the same node and compares
all returned frame-length sums with direct Go: both print 90,000. Subtract the
first query, then divide by 9,999:

| Warm query | Median cost |
| --- | ---: |
| Native Adamic bridge call | 4.518 us |
| Direct Go constrained-type/frame helper | 2.973 us |
| Difference | 1.545 us |

Native's interval includes path/kind copies, the C/Go crossing, exact node
lookup, checker acquisition/release, framing and conversion to an owned Adamic
string. Direct Go keeps its file, node and checker available. Thus the difference
is the measured adapter cost, not an isolated cgo instruction cost. Native's
first query includes lazy checker setup; it is excluded from the warm cost.
The production Go rule time does not include its prior checker acquisition, so
it must not be directly compared with native's accumulated query time.

## Commands and outputs

All test output went to files. Setup succeeded:

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (41s)
setup: done in 41s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Run `bash cloud/setup.sh`, then source `/workspace/adamic-tools/env.sh`.
The exact final unit command was:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
ADAMIC_TYPEAWARE_ARTIFACTS=/tmp/tsgo-typeaware-final \
ADAMIC_TYPEAWARE_BENCH=1 \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware > /tmp/tsgo-typeaware-final.log 2>&1
```

PASS, 96.311s; [typeaware.log](validation/typeaware.log). Its timings overlapped
with parser regression and are superseded by the isolated samples. Reproduce
the isolated measurements after other suites finish:

```sh
python3 stage1/cohere/typeaware/validation/measure.py /tmp/tsgo-typeaware-final /tmp/tsgo-typescript > /tmp/tsgo-typeaware-isolated-timing.log 2>&1
```

The touched-package command was:

```sh
ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript go test -count=1 -v -timeout 30m ./bridge/tsgo ./internal/load ./internal/lower ./internal/native > /tmp/tsgo-typeaware-regression.log 2>&1
```

Load/lower/native passed in 0.803s/4.829s/67.139s. Bridge initially failed only
its nonunique mutant anchor as described above; the corrected final command:

```sh
ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript go test -count=1 -v -timeout 30m ./bridge/tsgo > /tmp/tsgo-typeaware-bridge-final.log 2>&1
```

PASS, 90.352s; [bridge.log](validation/bridge.log).

Merged-slice regression:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript go test -v -count=1 -timeout 30m ./stage1/typescript/parser ./stage1/typescript/scanner ./stage1/cohere/lint > /tmp/tsgo-typeaware-parser-regression.log 2>&1
```

PASS: parser 291.439s, scanner 67.720s, syntax-only lint 140.037s.
[parser-scanner-lint.log](validation/parser-scanner-lint.log) retains all
generated, compiler, gap and mutant outputs, including 44,766,682 identical
whole compiler tree bytes.

Filtered Node oracle:

```sh
go test -count=1 -v -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(closures\.a|method_closures\.a|generic_functions\.a|regions\.a|regions_throw\.a)$|TestTheOracleCatchesOneByte' > /tmp/tsgo-typeaware-filtered-oracle.log 2>&1
```

PASS, 15.942s; [node-oracle.log](validation/node-oracle.log).
`go vet ./...`, `gofmt -l cmd internal bridge stage1/cohere/typeaware` and
`git diff --check` passed with no output. Cohere format and lint commands:

```sh
/tmp/tsgo-lint-cohere --format-only stage1/cohere/typeaware/*.ts stage1/cohere/typeaware/testdata/*.ts
/tmp/tsgo-lint-cohere --no-fix --no-cache stage1/cohere/typeaware/*.ts stage1/cohere/typeaware/testdata/*.ts
```

Zero findings, four of four Adamic-ready; [cohere.log](validation/cohere.log).
Raw regression logs retain trailing tabs for empty canonical fields. The local
validation attributes disable whitespace diagnostics on those data files only;
source whitespace checks remain enabled.

## Limits

The complete repository gate was not rerun for this continuation; the previous
foundation's full gate took about twelve minutes. Touched compiler/bridge
packages, every merged parser/scanner/lint package, the new rule and a filtered
Node oracle were run instead. No protected emitter/native/lowering/oracle
file was edited.

Coverage is one rule's canonical diagnostics, not the cohere CLI's suppression,
configuration, display or fix engine. The port is not a general semantic checker.
JSX, diagnostic parity, mutable source files during a run, concurrent callers,
other platforms, c-shared and complete semantic diagnostic retrieval are not
covered. Semantic diagnostics are not requested, and source/config imports may
resolve to error types just as they do for the independently loaded Go program;
this is parity on the pinned corpus/config, not a claim every project dependency
was installed or every type error examined.

The external checker retains Go's runtime and collector. Adamic values remain
counted or region-owned; sanitizer checks cover native/C ownership, not
reclamation of Go heap roots. No performance optimization was attempted here.
