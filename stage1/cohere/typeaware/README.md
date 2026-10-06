# The first type-aware cohere rule

Native Adamic implements `@typescript-eslint/no-unsafe-unary-minus`, using the
merged Adamic TypeScript parser and the milestone 5 checker archive. It makes
one bridge call per unary minus operand. The checker supplies the constrained
type's union parts and their flags/type strings. Adamic tests
`Any | Never | NumberLike | BigIntLike`, reports the first offending part, and
renders the production cohere rule's message and whole-expression span.

The rule has no options, fixes or suggestions. Safe number/bigint unions,
`any`, `never` and constrained numeric parameters stay silent. Boolean names
its first union part, `false`, as Go cohere does. Asking about the unary
expression itself would see the numeric result and hide an unsafe operand.

After setup and sourcing its printed environment:

```sh
go build -buildmode=c-archive -o /tmp/tsgo.a ./bridge/tsgo/archive
go build -o /tmp/adamic ./cmd/adamic
/tmp/adamic build stage1/cohere/typeaware/main.ts -o /tmp/typeaware --tsgo /tmp/tsgo.a
/tmp/typeaware /path/tsconfig.json /path/manifest
ADAMIC_TSGO_TIMING=1 /tmp/typeaware /path/tsconfig.json /path/manifest --count
```

The manifest contains one absolute source path per line. The program uses that
list as the tsconfig's roots. Source files must stay unchanged during the run.
Output includes a header per file and tab-separated findings: UTF-8 start/end,
rule name, message ID, escaped message, fix count and suggestion count. The last
line is `findings N queries Q`; count mode prints only that line. This is a
canonical diagnostic protocol, not the full cohere CLI display or fix engine.

The independent Go oracle is compiled through an overlay inside the pinned
cohere module. It loads its own program, walks its own AST, and invokes
`NoUnsafeUnaryMinus.Run` unchanged. It imports no bridge implementation and
does not copy the rule predicate. The same root list/config is used on both sides.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
ADAMIC_TYPEAWARE_ARTIFACTS=/tmp/typeaware-validation \
ADAMIC_TYPEAWARE_BENCH=1 \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware > /tmp/typeaware.log 2>&1
```

Use TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
Without the corpus environment variable, generated agreement and all mutants
still run. Tests never download a corpus. Benchmarking is optional; with it,
three interleaved rounds measure whole-process load/parse/lint count throughput.
The repeated-query probe compares 10,000 identical type frames with direct Go.
Subtracting the first query measures the remaining calls on an already-warm
checker, including native path copies, C crossing, exact node lookup, checker
lease, serialization and conversion to an owned Adamic string.

The new C function is `tsgo_type_parts`; the Adamic declaration is
`tsgoTypeParts`. Kind names omit Go's `Kind` prefix. Start/end are UTF-8
bytes, with leading trivia. The bridge refuses an inexact span or kind rather
than choosing a neighbor. Its enclosing result uses the ABI's explicit UTF-8
byte length and ownership. Internally each part is framed as decimal flags,
LF, decimal UTF-16 length, LF, TypeToString text. Text can contain LF; the length
separates it from the next frame without escaping or losing Unicode. Adamic
validates every frame before using a finding.

The parser uses table indexes for child links. To avoid its documented stage 0
nested-constructor gap, the driver constructs its Parser and Scanner before
passing them to UnaryMinus. No protected emitter/lowering files were edited.

See [REPORT.md](REPORT.md) for measurements, commands, mutant results and limits.
