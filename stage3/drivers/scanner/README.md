# Scanner source proof

This driver imports the adapted scanner.ts and creates its scanner exactly as
parser.ts:1444 does: `createScanner(ScriptTarget.Latest, true)`. It calls `scan()`
until EOF for each input. Parser-directed rescans of slash, template or JSX
syntax are outside this lexical scanner proof.

Each line has six tab-separated fields: SyntaxKind name, full start, start, end,
token flags, and JSON-escaped token text. Positions are UTF-16 code units. EOF
is included, trivia is skipped, and token text is escaped so a multiline token
still occupies one output line. Reverse enum aliases follow stock TypeScript
6.0.3. Files are sorted by relative path, recursively, with every regular file
under src/compiler included, even JSON and generated files. files.json records
that order; each EOF separates files without adding non-token output.

After integrating stage3-base and stage3-type-imports at a3ef0dc:

```sh
source /workspace/adamic-tools/env.sh
stage3/drivers/scanner/run.sh /tmp/scanner-proof > /tmp/scanner-proof.log 2>&1
```

The output directory must be new. The runner builds a scratch copy of the apply
pipeline with adaptations 00, 10, 50 and 51. Adaptation 20 is deliberately excluded
per the October 7 instruction. apply.sh still constructs the tree and measures
the patch set. The runner regenerates diagnostics after adaptation 10 changes
the generator's type import. `--tree <already-applied-tree>` reuses an existing
scratch tree; `--compiler <binary>` selects an integrated feature compiler.
`--node-only` omits native compilation; `--inputs <tree>` selects a fixed corpus
for comparing the scanner before and after a source adaptation.

Node runs this same main.a through stock TypeScript's transpileModule, with
verbatim module syntax. This handles upstream's enums and namespaces without
using Adamic output. The loader resolves upstream .js requests to their .ts
sources and initializes the compiler barrel before importing the driver.
Starting the ESM graph directly at scanner.ts exposes parser.ts's load-time read
of textToKeywordObj before scanner initialization. Barrel initialization is an
oracle bootstrap, not evidence that native module initialization is implemented.

The runner retains separate stdout, stderr, reports, source closure edges and
input manifests. A failed native build exits 1 and leaves its full diagnostics;
it never reports a native comparison pass. When compilation succeeds it runs
the native binary with the same input paths and diffs every token byte.

Before the native build, the same `diff -u` comparison passes a copied Node
control, then rejects a scratch Node output with only the first token's end
incremented by one. This proves comparison sensitivity even while native is
blocked; the copied control is explicitly not native output.

BLOCKERS.md states the observed gate and the stopping point. Evidence retains
all checker diagnostics in returned order. It does not claim an exhaustive
Refused/NotYet traversal while checking still fails.

## Current feature profile

After gathering the raw scanner slice, set SLICE_TYPESCRIPT to stock TypeScript
6.0.3's lib/typescript.js and run:

```sh
bash stage3/drivers/scanner/adapt-slice.sh SLICE > adaptations.log 2>&1
bash stage3/drivers/scanner/run.sh OUTPUT --tree SLICE --inputs FIXED_CORPUS \
  --compiler INTEGRATED_COMPILER > scanner.log 2>&1
```

The helper selects 52-57, 59, 81, 82 and 85. Current readiness bc9f5d7 closes
58's literal-initializer workaround; numeric enums ec67b02 close 80, 83, 84 and
the proposed 86. Those legacy plans and results remain historical evidence.
56 remains selected: eef541e admits proven assertion syntax but the original
unknown parameter still cannot lower. The latest compiler tips and scratch
integration SHA are recorded in evidence/native-live-feature-refs.json.

The latest front-2 860a0d5 plus records-lowering/runtime-records attempt admits
MapLike's type-only index signature: its unchanged probe prints ok natively.
The scanner now stops at Debug.fail's Error-as-any cast, debug.ts:14:14.
Node's 509,014 tokens still match the full tree. See BLOCKERS.md and
evidence/records-run.json. No native scanner diff or ownership pass is claimed.

## Native comparison and measurement

`--oracle FULL_TREE_NODE_STDOUT` requires the slice Node stream to match the
full-tree stream before attempting native compilation. A successful native
comparison then plants exactly one changed byte in a copy of native stdout;
`diff -u` must return 1. The mutant is output-only and does not change the slice.

After a green native run, `python3 stage3/drivers/scanner/measure.py OUTPUT`
runs the native binary and the same Node driver three times each on the exact
files.json corpus, retaining every stdout, stderr, and diff. User CPU time is
the sequential RUSAGE_CHILDREN delta; report.json records all three values and
the best. Process startup, Node transpilation, and token printing are included.
The script records native binary bytes and attempts `perf stat -e instructions`
when perf is installed, retaining access failures rather than inventing counts.
Compilation failure produces no native timing, binary size, or native-output
mutant claim.

## Split compilation

The integration requires area/developer-tools. Run the scanner proof with
`ADAMIC_NATIVE_SPLIT=0`, and separately with
`ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc)`. Both binaries must match
Node byte for byte. Source emission and build timing happen only after the
checker/lowering gates pass; failed emission has no generated-C size.

To measure the emitted C without including TypeScript checking/lowering, run
`adamic c OUTPUT/main.a > scanner.c 2> emission.log` using the integrated
compiler. From that compiler checkout, run:

```sh
go run stage3/drivers/scanner/build-metrics.go scanner.c NEW_METRICS_DIRECTORY 5 > build-metrics.log 2>&1
```

Pass nproc instead of 5 on another box. The helper is ignored by ordinary Go
package builds and needs the integrated native.Options Split/Jobs API. It
records C bytes/lines, release flags, binary sizes, and wall time for unsplit,
cold-object-cache split and warm-object-cache split builds. Runtime preparation
is measured separately and excluded from those build times; split preprocessing,
cache checks and linking remain included. Run and compare all three measured
binaries on the same files.json corpus before using their timings as a proof.

The coverage dump scans every corpus file twice, first with `skipTrivia: true`,
then with `skipTrivia: false`. Every token includes `getTokenValue()` serialized
with JSON.stringify, alongside kind, full start, start, end, flags and raw text.
The value is exactly the scanner's current value, including stale values on
punctuation. Inline error rows include code, category, the callback's scanner
position (the error start), length, message text and substitution argument.
Pass headers identify the pass and input file. Absent cooked values and substitution arguments are represented as JSON `null`.

This dump does not drive rescans: regular expressions, template continuation,
JSX, and greater-than tokens. The parser slice exercises those once its
JSDoc-complete dump lands. `coverage-mutants.py` changes one escape branch,
omits exactly one error callback, and changes the single-line comment kind in
scratch copies. Each Node execution must succeed and its complete dump must
compare unequal; the escape mutation must change cooked values alone.

Observed full-tree/slice reference: **1,369,432 tokens**, **466 errors**, 81 input
files in each pass, 108,019,935 bytes. SHA-256:
`41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
The skipped-trivia pass has 509,014 tokens; the retained-trivia pass has 860,418.
All three source mutants exit zero on Node and produce unequal dumps. The
error mutant removes exactly one row. Native execution remains blocked by the
previously recorded typed captureStackTrace marker refusal; this expanded
coverage result is a Node reference proof, not a native proof.
