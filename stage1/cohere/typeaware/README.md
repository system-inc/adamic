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

## The next five rules

`suite.ts` runs the unary-minus rule and five more production cohere rules under
**their default options**:

| Rule | Checker question |
| --- | --- |
| `related-getter-setter-pairs` | Getter-to-setter assignability |
| `no-unsafe-declaration-merging` | Named and local symbols' declarations |
| `no-unsafe-argument` | Resolved signature, rest parameters, argument/reference types |
| `restrict-plus-operands` | Constrained, widened types and union/intersection parts |
| `no-unnecessary-boolean-literal-compare` | Boolean/nullable constrained types and strict null checks |

One `tsgoProgram` loads all manifest roots. Adamic parses each file once, builds
its numeric parent index, and dispatches all six rules in one shared rule walk.
The checker owns the corresponding Go AST; every query identifies the exact
Adamic node by kind/span. Generic assignment recursion, tuple/rest consumption,
accessor pairing, declaration selection, operand tests, nullable defaults,
messages, spans and fixes are decided by Adamic. No lint verdict comes from Go.
Nullable boolean comparisons stay silent under the production defaults.
The command has no rule-option/configuration surface beyond that fixed suite.

```sh
/tmp/adamic build stage1/cohere/typeaware/suite.ts -o /tmp/suite --tsgo /tmp/tsgo.a
/tmp/suite /path/tsconfig.json /path/manifest
ADAMIC_TSGO_TIMING=1 /tmp/suite /path/tsconfig.json /path/manifest --count
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
ADAMIC_SIX_ARTIFACTS=/tmp/six-validation \
ADAMIC_TYPEAWARE_BENCH=1 \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run TestSix > /tmp/six.log 2>&1
```

The six-rule protocol retains file headers and diagnostic fields, appending each
boolean fix's byte start/end and escaped replacement after the two repair counts.
Its summary is `findings N`. Lines sort lexically by their complete canonical
representation, preserving duplicates. Output is diagnostics and proposed fixes;
no suppression or edit engine is invoked on either side.

`testdata/oracle_six.go` has its own program loader and one shared Go AST walk.
It calls all six unmodified production `Run` implementations, including each
rule's declared program view and shared file cache. It compares complete finding
bytes and fix bytes, not only counts. Table-source extraction is followed by an
independent Go parse filter: nonsource message strings and intentionally broken
fixtures are excluded and counted explicitly.

The added fact API and ownership are documented in [facts.md](../../../bridge/tsgo/facts.md).
The six-rule measurements and mutant evidence are in [SIX_RULE_REPORT.md](SIX_RULE_REPORT.md).
The older [REPORT.md](REPORT.md) remains evidence for the first rule.

## Profiled query costs

The optimized suite requests `raw-shape`, `type-shape` and `signature-shape`
where a decision needs flags/relations but no display name. It asks `name` for a
live type identity only while building a reported argument message. Plus still
requests named facts for its RegExp decision. Numeric frames retain canonical,
safe-integer and ownership checks while avoiding temporary numeric strings.
The shared walk avoids empty unary-result slices, metadata searches avoid
capturing closures, and sorting renders one complete key per finished finding.
`Diagnostic.written()` remains fresh if a repair is edited later.

Use `ADAMIC_TSGO_PROFILE=/absolute/path/cpu.pprof` on a separate profiling run.
Go CPU samples, native leaf addresses, C input/call/output intervals and Go
allocation counts distinguish rendering/serialization from crossing overhead.
[PROFILE_REPORT.md](PROFILE_REPORT.md) records before/after numbers, source pins,
mutants, regressions and why the measured result did not justify batching.
Native phase and leaf tools live in `bridge/tsgo/profile/`; normal builds do not
instrument native phases. Earlier six-rule measurements remain in
[SIX_RULE_REPORT.md](SIX_RULE_REPORT.md).


## Ten rules selected by finding volume

`volume_suite.ts` runs the original six and the next ten default TypeScript
rules. Selection began with Go counts of all 62 checker-dependent
`@typescript-eslint` rules. The counter now also records all 197 checker-dependent
rules across every family. These are the ten highest combined counts among the
remaining `@typescript-eslint` rules on the pinned compiler corpus and pre-port
repository manifest; higher-volume rules in other families remain unported:
unsafe type assertions, unsafe member access, nullish coalescing, shadowing,
unsafe enum comparisons, unsafe assignment, confusing void expressions,
consistent returns, exhaustive switches and unbound methods.

The runner loads once, parses each file with Adamic's parser, and emits complete
canonical findings. After fix count and suggestion count, every fix contains
byte start/end and escaped replacement. Each suggestion contains message ID,
escaped message, fix count and those fix triples. Fix ordering and duplicate
findings are preserved; lines sort by their complete canonical representation.
This compares proposed edits without applying them or running suppressions.

```sh
/tmp/adamic build stage1/cohere/typeaware/volume_suite.ts -o /tmp/volume --tsgo /tmp/tsgo.a
/tmp/volume /path/tsconfig.json /path/manifest
ADAMIC_TSGO_TIMING=1 /tmp/volume /path/tsconfig.json /path/manifest --count
ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript \
ADAMIC_VOLUME_ARTIFACTS=/tmp/volume-validation \
ADAMIC_VOLUME_REPOSITORY_MANIFEST=/path/pre-port-repository.manifest \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run '^TestVolumeAgreementAndMutants$' > /tmp/volume.log 2>&1
```

`testdata/count_volume.go` and `testdata/oracle_volume.go` build through overlays
inside the pinned cohere module. They independently load programs and call its
unchanged production rules, with one shared AST walk and production program views
and file caches. They import no bridge code. The declaration roots from the
config are retained when explicit manifests replace its ordinary roots; `.a`
files are read as TypeScript with the compiler's non-TS-extension option.

[Volume report](VOLUME_REPORT.md) records selection counts, byte comparisons,
sanitizer and mutant evidence, timings, exact commands and limits. The additional
compiler questions and their ownership are in
[facts.md](../../../bridge/tsgo/facts.md). The fixed suite accepts default rule
options and uses the original strict configs for both measured corpora. It explicitly
refuses `noImplicitThis: false` until those special diagnostic variants are ported.

## Profiling the sixteen-rule suite

The span index in the Go bridge caches parsed AST locations, with source-file
identity, both byte bounds and exact kind preserved. It does not cache checker
facts. The native shadow rule indexes candidate scopes by binding name, keeping
scope order, containment and the first binding in a scope. The supported rule
defaults and diagnostic protocol are unchanged.

[Volume profile report](VOLUME_PROFILE_REPORT.md) records disjoint phase costs,
bridge overhead, before/after/Go measurements, byte comparisons and mutants.
`bridge/tsgo/profile/volume_phases.py` instruments scratch generated C only;
`volume_bench.py --before /path/unchanged-native` measures an unchanged baseline
alongside the current native binary and independent production Go oracle.
