# Milestone 2 expression parser slice

Historical expression-slice report. The whole-file continuation is recorded in
[WHOLE_REPORT.md](WHOLE_REPORT.md).

Built on the existing scanner, in Adamic's subset, compiled through stage 0.
No compiler, runtime, scanner or other worker's files changed. The branch is
`codex/typescript-scanner`; green steps were committed and pushed individually.

The driver parses all 77 TypeScript 6.0.3 compiler files and selects maximal
expression roots using typescript-go's `ast.IsExpressionNode` rule. Go, Node
running the same TS, and native under ASan, UBSan and LeakSanitizer agree on
28,836,875 canonical bytes. Another 1,676 generated inputs agree on 1,432,520
bytes. The original 58 focused inputs also agree and hold the three mutants.

Statements and declarations are parsed to reach the compiler's expressions.
Every child of a selected expression is compared, including statements, types
and declarations inside arrow, function and class expression bodies. Top-level
statement and declaration trees themselves are not claimed as a ported answer.
Static imports and exports are structurally skipped. This is a parser slice,
not a complete TypeScript parser or checker.

Primary/unary/binary/conditional expressions, assignment and comma, arrows,
functions, call/member/element/new, templates, objects/arrays/spread, optional
chains, non-null expressions, assertions/as/satisfies and expression type
arguments are included. Supporting type and statement grammar is described in
[GAPS.md](GAPS.md). Positions are byte offsets including leading trivia, matching
Go. The port internally uses UTF-16 and converts at the printing boundary.
Children follow Go's ForEachChild order. The protocol retains OptionalChain,
literal flags, operator kind, cooked literal/identifier text, template raw text,
and literal/container/call metadata. Other parser context, recovery, JSDoc and
binder flags are excluded explicitly, not silently normalized after comparison.

The oracle is [testdata/oracle.go](testdata/oracle.go), compiled inside the
unmodified typescript-go module using a Go build overlay. The port's node table
contains numeric child indexes and has no owning parent/child pointer cycle.
The driver reads either a source file or a manifest of absolute paths:

```
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/typescript/parser/main.ts -o /tmp/parser
/tmp/parser input.ts
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/parser/main.ts input.ts
/tmp/parser --manifest compiler.manifest --count
```

The corpus pin is TypeScript tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, cloned into
`/workspace/scratch/typescript-6.0.3`, not committed. The oracle/parser original
is cohere's typescript-go commit `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
The compiler-corpus test checks the pin and recursively includes every `.ts`
under `src/compiler`. Generated cases include all 625 pairs of 25 binary
operators, sixteen assignment forms, 1,000 nested expressions with seed 720,
and 35 targeted regressions. The random nesting uses the first ten operators;
the pair matrix independently exercises all twenty-five. Every generated input
must be diagnostic-free in Go; inputs are not filtered after a comparison
failure. The source tree and corpus generators are reviewable in the tests.

| Green step | Pushed commit | Observation |
| --- | --- | --- |
| Primary and precedence | `df909aa52267a066e2c59d524f7c8916d4573ef1` | 14 inputs, 3,194 bytes |
| Suffixes and containers | `854d6bd89d86d564f7fb3c90887d53b2b2eb41d2` | 30 inputs, 8,080 bytes |
| Arrows, functions and types | `48bfd18d25261561cf604061de1c4ac71c99e4b4` | 58 inputs, 17,463 bytes |
| Whole compiler traversal | `5824126a5555214d70a48fb155d120cd715b2aed` | 77 files, 28,812,163 bytes |
| Generated combinations | `bbc78f98c91ae10f56b7fa580afcb7898052f70a` | 1,676 inputs, 1,432,520 bytes; compiler 28,836,875 bytes |
| Throughput measurement | `a8ccd394e6923448727f199953a88e0715348b67` | Corrected source view; 557,010 equal nodes on every sample |
| Isolated count mutant | `83c540ba65bdb663ca0c18ec39fb6563a2259435` | Same AST bytes; wrong count caught on Node and native |

The compiler-byte increase in the last step adds raw no-substitution template
text to the canonical protocol on all three sides. It does not change any
scanner output. Historical counts in this table are the protocol at each step.

All mutants compile and finish normally under Node and sanitized native. A
crash, lowering error or clang error is not credited as catching a mutant:

| Mutant | Input demonstrating it | Comparison that catches it |
| --- | --- | --- |
| Multiplication precedence 14 changed to 12 | `x + y * z` | BinaryExpression child shape differs |
| Ordinary parentheses classified as an arrow | `(x)` | Root ParenthesizedExpression becomes ArrowFunction |
| Optional-chain property flag forced false | `a?.b.c?.[x]?.(y)` | Required flag 32 becomes 0 |

The arrow mutant changes the returned classification of an ordinary
parenthesized expression. It does not claim that a detector mutant which merely
panics while expecting `=>` has passed the wrong-tree test. Exact first differing
lines for all six successful executions are retained in
[validation/generated-and-compiler.log](validation/generated-and-compiler.log).
A fourth mutant proves the benchmark count guard: it changes only countTree's
per-node contribution from 1 to 0. Native and Node both finish normally with
identical printed AST bytes, but count 0 instead of Go's 18. Tree comparison
alone cannot catch this mutant; the count check does. Its output is retained in
[validation/count-mutant.log](validation/count-mutant.log).

The 625 operator pairs independently caught the original `??` rank: Go gives it
the same precedence as `||`. Other generated failures led to parser-directed
regex/template rescans in lookahead, correct generic optional propagation,
relational-versus-type-argument disambiguation, and async parameter contexts.

Four gaps have reduced proving programs and executable gap tests:

| Gap | Observation | Workaround |
| --- | --- | --- |
| Strong parent/child ownership cycle | Node prints root; stage 0 refuses the cycle-capable array | Indexed node table |
| Spread arguments to Array.push | Node prints 1,2,3; stage 0 reports NotYet SpreadElement | Append indexes in a loop |
| Import cycle with a type-only edge | Node prints 1; stage 0 refuses an import cycle | Shared snapshot type; concrete Scanner dependency |
| Class method called through an interface | Node prints 1; sanitized native exits 70 with missing-field compiler panic | Call concrete Scanner methods in lookahead |

The final gap is an observed compiler/runtime bug, not a language refusal.
[GAPS.md](GAPS.md) gives the evidence and exact proving files. Internal code is
unchanged; a compiler unit can use that regression directly.

Timing counts the same 557,010 selected-tree nodes on every implementation.
It includes process startup, file reading, scanning, whole-file parsing and
selected-tree counting. It excludes printing/escaping the trees, UTF-16 to byte
mapping, build time and manifest creation. One warm-up precedes five samples;
implementation order rotates between samples. All warm-ups and samples must
print the identical node count. Native uses the normal stage-0 clang `-O2`
flags without sanitizers; sanitizer checks are separate. Go uses ordinary
`go build`; Node executes the TS through the repository's runner.

| Implementation | Best seconds | Nodes per second |
| --- | ---: | ---: |
| Go | 0.153166816 | 3,636,623 |
| Node | 0.567490607 | 981,532 |
| Native Adamic | 0.693479657 | 803,210 |

Observed: native takes 4.528 times Go's time and 1.222 times Node's. This slice
loses to both; the result is not a claim that identical counted nodes imply
identical implementation work. Go builds its complete AST, parents and parser
context flags. The port builds indexed nodes and traversal scaffolding, skips
static import/export declarations, and does not implement recovery or JSDoc.
Those differences limit what the rate comparison says about the runtime alone.
No cost attribution or compiler performance fix is claimed by this unit.

Machine: Linux amd64, AMD EPYC 9V74, `nproc` 5, cgroup CPU quota 4 cores
(`400000 100000`), cgroup memory limit 16 GiB. Go 1.27.1, Node 24.19.0,
clang 20.1.8. One-minute host load during the samples was 1.81 to 1.88;
benchmark-start load was 1.88/1.21/0.60. Every duration and observed load is in
[validation/performance.log](validation/performance.log). There were no
concurrent tests from this worker during the reported timing run. Host load
alone cannot establish absence of other workers or CPU contention.

The first timing attempt was discarded: the raw-text oracle addition converted
source bytes to a new string for every expression root, copying entire files.
The oracle now uses the already parsed SourceFile.Text. After this correction,
the compiler corpus still agrees on all 28,836,875 bytes. The table above uses
only the corrected run; no discarded timing is included.

Commands and retained outputs:

```
bash cloud/setup.sh > /tmp/typescript-parser-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# setup: go ready (0s)
# setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
# setup: node ready (0s)
# setup: submodules ready (0s)
# setup: build cache warm (11s)
# setup: done in 11s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc
# 5

/workspace/scratch/cohere --format-only stage1/typescript/parser/*.ts > /tmp/typescript-parser-format.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts > /tmp/typescript-parser-cohere.log 2>&1
# 276 rules, 5 checked, 100% Adamic-ready, no findings

gofmt -l cmd internal stage1/typescript/parser > /tmp/typescript-parser-gofmt.log 2>&1
# no output
go vet ./... > /tmp/typescript-parser-vet.log 2>&1
# exit 0, no output
git diff --check
# exit 0, no output

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -v -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-step5.log 2>&1
# PASS, 74.839s; generated, compiler, gaps and all three mutants

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -v -run '^TestCompilerExpressionsAgree$' -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-oracle-source-view.log 2>&1
# PASS, 19.784s; corrected oracle, 28,836,875 identical bytes

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_PARSER_BENCH=1 ADAMIC_PARSER_ARTIFACTS=/workspace/scratch/parser-performance go test -count=1 -v -run '^TestPerformance$' -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-performance.log 2>&1
# PASS, 12.896s; all warm-ups and fifteen samples have 557,010 nodes

go test -count=1 -v -run '^TestNodeCountCheckCatchesMutant$' -timeout 30m ./stage1/typescript/parser > /tmp/typescript-parser-count-mutant.log 2>&1
# PASS, 24.807s; identical AST bytes, Node/native 0 nodes versus Go 18

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -timeout 30m ./... > /tmp/typescript-parser-full-gate.log 2>&1
# exit 0; all packages passed
# internal/oracle 681.359s; stage1/typescript/parser 170.721s; scanner 111.729s
```

The full gate includes the pinned compiler corpus. The later isolated count
mutant test was run separately, since it was added after that gate started.
No production source changed during the full gate.

Logs are retained under [validation/](validation/). Benchmark binaries and the
manifest are scratch artifacts at `/workspace/scratch/parser-performance`, not
committed. The benchmark test is opt-in and reproduces the measurement above.

Not covered: parser recovery/diagnostic parity, malformed source, JSX,
decorators, JSDoc trees, static import/export AST fields, non-UTF-8 input,
other node flags, complete top-level statement/declaration tree equivalence,
and general TypeScript syntax beyond the compiler and generated corpora.
The expression parser is not used to validate or type-check Adamic itself;
stage 0 continues to use typescript-go's parser.
