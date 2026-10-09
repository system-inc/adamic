Ported statements, declarations and all 42 kinds in Go's TypeNode inventory.
77 compiler files: 44,766,682 identical canonical bytes on Go, Node and native.
68 generated files, two JSDoc probes and obsolete assertions also match.
Three tree mutants and a count-only mutant are caught on Node and native.
Best of five: native 1,134,442, Go 5,665,349, Node 1,357,630 nodes/s.

# Milestone 2 whole-file parser continuation

Continued `codex/typescript-scanner` from the expression slice. Everything in
this continuation stays inside `stage1/typescript/parser/`. No compiler,
runtime, scanner or other worker's source was edited. No pull request opened.
Green steps were committed and pushed separately:

- `db1f91c`: statement module and explicit callback workaround; gap 5 proved.
- `d7ddef4`: whole-file protocol, real imports and exports, compiler agreement.
- `d11f976`: declarations, types, generated forms, gaps and mutant evidence.

The expression-slice results remain in [REPORT.md](REPORT.md). Its driver mode
and existing tests are preserved. The new answer is the whole SourceFile,
including declarations and statements at top level, all expression/type
children, and the end-of-file token. Static imports and exports are now parsed.

The independent oracle builds an overlay program inside cohere's typescript-go
module, invoking its unmodified `parser.ParseSourceFile`. The pinned oracle
commit is `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Compiler sources remain in
scratch, not committed: TypeScript tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, under
`/workspace/scratch/typescript-6.0.3`. The test checks the pin before collecting
all 77 `.ts` files recursively from `src/compiler`.

## What is compared

Use `main.ts <file> --whole`, or `main.ts --manifest <file> --whole`.
The manifest lists source paths, one per line. `--count` suppresses trees and
prints the total reachable node count for the manifest.

Preorder lines retain the expression protocol: kind, byte position/end,
optional-chain flag, literal flags, selected list count/trailing/multiline,
operator, cooked text and template raw text. Whole-file mode adds declaration
semantics: variable declaration-list flags, import phase, type-only booleans
and export-equals. Module keywords and attribute tokens are also retained.
Go's ForEachChild supplies the child order. Unicode and CRLF probes verify
conversion from the port's UTF-16 positions to Go's UTF-8 byte positions.
See [GAPS.md](GAPS.md) for the exact fields and their limits.

The port covers variable/function/class/interface/type alias/enum/module and
namespace declarations, members and heritage, static blocks, decorators and
parameter decorators; import/export clauses, equals/default, namespace,
type-only/defer and attributes; blocks, empty statements, if, all for forms,
while/do, switch, try/catch/finally, labels, return/throw, break/continue,
debugger and with. It includes using and await using declarations.

Type coverage is checked against `ast.IsTypeNodeKind` in the Go oracle, not a
handwritten inventory: every one of its 42 kinds must occur in generated
oracle trees that the port matches. This covers the 24 ordinary kinds, twelve
keyword types, ExpressionWithTypeArguments and five JSDoc wrappers. The
optional and variadic JSDoc wrappers require annotation grammar; two reduced
`/** @type {...} */` probes compare their type subtrees with Go's lazy JSDoc
parser via `--doc-types`. This is an annotation extractor, not a general
JSDoc comment/tag parser.

## Observations

| Corpus | Identical canonical bytes | Implementations |
| --- | ---: | --- |
| 77 whole compiler files | 44,766,682 | Go, Node, sanitized native and release native |
| 68 generated whole files | 93,323 | Go, Node, sanitized native |
| Two JSDoc annotation types | 250 | Go, Node, sanitized native |
| Obsolete assert import/export/import-type forms | 1,021 | Go, Node, sanitized native |

Total new canonical agreement: 44,861,276 bytes. The obsolete-assertion probe
uses an explicit oracle option which requires that Go emit only diagnostic
2880. Ordinary tests still reject every parse diagnostic. This compares the
obsolete forms' trees and does not claim diagnostic parity.

Generated inputs also cover contextual type/as ambiguities, dotted and ambient
modules, ambient module type-literal attributes, const enums, variances,
accessors versus keyword-named members, computed decorator members versus
indexed decorator arguments, ASI and empty files. Each ordinary generated
input is also tested after a Unicode comment prefix and CRLF conversion.
Qualified primitive-keyword type names, this/asserts predicates, infer
constraint speculation, tuples, mapped modifiers/remapping, import attributes
inside types and template types have explicit probes.

Six reduced gap programs and tests remain. The two new gaps are optional
parameters on function values (refused) and untyped empty-array branches in a
conditional (NotYet). Required callbacks plus ordinary method defaults, and an
explicit `number[]` annotation, work around them. The statement module also
uses real function-property callbacks to avoid the previously proved class
method/interface dispatch bug and type-only import-cycle refusal. No internal
fix was made. [GAPS.md](GAPS.md) distinguishes refusals from the observed bug.

## Mutants

All three new tree mutants compile and finish normally on Node and native.
Byte comparison against Go catches each:

| Mutation | First disagreement |
| --- | --- |
| For-of emitted as for-in | ForInStatement versus ForOfStatement, position 0, end 25 |
| Type-only import phase lost | ImportClause Unknown versus TypeKeyword, position 32, end 41 |
| Keyof emitted as readonly | TypeOperator ReadonlyKeyword versus KeyOfKeyword, position 60, end 68 |

A fourth mutant changes the count traversal's own-node contribution from one
to zero. All whole-tree bytes stay identical, but Node and native report zero
instead of Go's 12. The count comparison catches it. The original expression
precedence, parenthesized-arrow and optional-chain mutants, expression count
mutant and scanner's three mutants remain in the regression suite.

## Throughput

Compilation and tree printing are outside the timing. Before timing, the
release native and Node builds each match all 44,766,682 Go compiler-tree
bytes. Every warm-up and timed run must return Go's exact count: **887,803
reachable whole-file nodes**. Timed work includes process startup, file reads,
parsing and count traversal. Five runs per implementation rotate ordering;
these are the fastest runs, with the full samples retained in
[whole-performance.log](validation/whole-performance.log).

| Implementation | Best seconds | Nodes/s |
| --- | ---: | ---: |
| Go | 0.156707557 | 5,665,349 |
| Node | 0.653935711 | 1,357,630 |
| Native, clang -O2 | 0.782590196 | 1,134,442 |

Native loses by **4.994x to Go** and **1.197x to Node**. Node-count rates from
the earlier expression slice use a different denominator and are not a speedup
measurement for this continuation.

Machine: linux/amd64, AMD EPYC 9V74, `nproc` 5, cgroup quota 4 CPUs
(`cpu.max=400000 100000`), memory cap 16 GiB. Load at benchmark start was
`0.79 0.97 1.04`; sample load was `0.81` to `0.83` for the one-minute value.
No other test suite ran during measurement. Toolchain: Go 1.27.1,
clang 20.1.8, Node 24.19.0. Benchmark binaries and the manifest are retained
under `/workspace/scratch/parser-whole-benchmark`, not committed.

## Commands and retained outputs

All test output went to files, never pipes. Setup used
`bash cloud/setup.sh > /tmp/typescript-parser-whole-setup.log 2>&1`, followed
by `source /workspace/adamic-tools/env.sh`. Timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (10s)
setup: done in 10s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The focused final command was:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -v -run '^(TestWholeCompilerAgrees|TestWholeGeneratedAgrees|TestEveryTypeNodeKindAgrees|TestWholeMutants|TestWholeCountCheckCatchesMutant|TestConditionalEmptyArrayGap)$' -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-whole-final.log 2>&1
```

PASS, 111.678s; [whole-final.log](validation/whole-final.log).
The obsolete assertion probe ran separately:

```sh
go test -count=1 -v -run '^TestObsoleteImportAttributesAgrees$' ./stage1/typescript/parser > /tmp/typescript-parser-whole-obsolete-assert.log 2>&1
```

PASS, 16.606s; [whole-obsolete-assert.log](validation/whole-obsolete-assert.log).
The release agreement and best-of-five command was:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_PARSER_BENCH=1 ADAMIC_PARSER_ARTIFACTS=/workspace/scratch/parser-whole-benchmark go test -count=1 -v -run '^TestWholePerformance$' -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-whole-performance.log 2>&1
```

PASS, 20.386s; [whole-performance.log](validation/whole-performance.log).
Formatting and lint:

```sh
/workspace/scratch/cohere --format-only stage1/typescript/parser/*.ts > /tmp/typescript-parser-whole-format.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts > /tmp/typescript-parser-whole-lint.log 2>&1
```

Passed, six of six files Adamic-ready, zero findings;
[whole-cohere.log](validation/whole-cohere.log). Go vet and the exact final
package regression results are retained with this report. The filtered
compiler oracle command was:

```sh
go test -count=1 -v -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(closures\.a|method_closures\.a|generic_functions\.a)$|TestTheOracleCatchesOneByte' ./internal/oracle > /tmp/typescript-parser-whole-filtered-oracle.log 2>&1
```

PASS, 8.494s; [whole-filtered-oracle.log](validation/whole-filtered-oracle.log).
It checks closure dispatch, method closures, generics and the one-byte mutant.

The final full package regression and vet commands were:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -v -timeout 30m ./stage1/typescript/parser ./stage1/typescript/scanner > /tmp/typescript-parser-whole-regression-final.log 2>&1
go vet ./... > /tmp/typescript-parser-whole-vet-final.log 2>&1
```

PASS: parser 236.628s, scanner 40.224s. Go vet exited zero with no output.
[whole-regression-final.log](validation/whole-regression-final.log) and
[whole-vet-final.log](validation/whole-vet-final.log) preserve the outputs.
The original 1,676 expression inputs still match 1,432,520 bytes; 58 focused
inputs match 17,463 bytes; expression subtrees in the 77 compiler files match
28,836,875 bytes. All six gap tests and every old/new mutant pass. The scanner
regression includes the new sources and matches 24,041,908 answer bytes.

## Limits

Agreement covers the documented canonical fields, not every Go AST field.
Context/transform flags, source-file metadata, parent pointers and list range
or trailing metadata beyond the printed fields are omitted. This is a parser
slice, not a checker. General error recovery and diagnostic parity, JSX and
complete JSDoc comment/tag parsing are not covered. Kind coverage proves the
listed probes, not exhaustive TypeScript grammar conformance. The complete
repository gate was not rerun for this continuation: touched parser/scanner
packages, a filtered internal oracle and `go vet ./...` were used. The earlier
full gate remains documented in the historical expression report.
