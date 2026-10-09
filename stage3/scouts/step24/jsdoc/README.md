# Step 24: JSDoc in the parser

Scout #6z35tzs. Main/compiler base: `031a1259bc7973934792dc6cb1bd4074fc2204b9`.
TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.
This unit changes only its scout territory. It adds measurements and three bounded
native witnesses; it does not establish full native parser acceptance.

## Entry points, options and lazy work

Coordinates below are the main adapted source tree, whose hashes are recorded in
`measurement.json`.

- `parser.ts:1344` `createSourceFile` accepts a ScriptTarget or
  `CreateSourceFileOptions`, including `jsDocParsingMode` at line 1337.
  `Parser.parseSourceFile` defaults that mode to ParseAll (line 1611).
  `initializeState` passes it to the shared scanner at line 1777.
- `parser.ts:1847` `withJSDoc` eagerly parses leading comment ranges for eligible
  nodes through `JSDocParser.parseJSDocComment` (line 8852). It attaches node.jsDoc
  during parsing, not on the first subsequent tag query. Parent-link settings
  do not decide whether the comments are parsed.
- `parser.ts:1414` `parseIsolatedJSDocComment` parses one comment with optional
  start/length bounds and fixes its parent references. Its worker (line 8841)
  explicitly initializes JS mode and ParseAll. It returns its own diagnostics.
- `parser.ts:1427` `parseJSDocTypeExpressionForTests` similarly initializes JS /
  ParseAll, parses a type expression, and returns diagnostics. Its worker is at
  line 8791. This test entry is useful for direct grammar controls.
- `parser.ts:1399` `updateSourceFile` delegates to incremental parsing. The
  source-file JSDoc mode is retained through reparsing (`parser.ts:9959`). A
  directed edit of a ParseNone file remains without attached comments.
- `scanner.ts:2391` `shouldParseJSDoc` controls the preceding-comment flag.
  `types.ts:10295` defines the four modes. ParseAll parses TS/TSX/JS/JSX;
  ParseNone parses none; ParseForTypeInfo parses JS/JSX but skips TS/TSX;
  ParseForTypeErrors parses JS/JSX and only TS/TSX comments containing @see or
  @link. The 32 kind/mode/comment probes confirm this behavior. JSON parsing
  produces no attached JSDoc in all four directed probes.

There is **no deferred hydration of skipped node.jsDoc** in these compiler parser
paths. `utilitiesPublic.ts:1257` `getJSDocTagsWorker` lazily computes and caches
flattened related tags in node.jsDoc.jsDocCache. The entry-point probe observes
that cache appear after a tag query on a parent-linked tree, and verifies a ParseNone file's tag query
returns zero tags without parsing comments later. These are different operations.
Incremental parsing clears reused tag caches because parents may have changed
(`parser.ts:3171`).

Service operations can parse a comment separately when requested, for example
`src/services/classifier.ts:805` calls parseIsolatedJSDocComment while classifying
multiline comments. That on-demand service work is outside the compiler closure
and is not deferred AST materialization by createSourceFile. This scout does not
measure language-service requests.

Malformed `@param { }` gives TS1110 in JS jsDocDiagnostics, but the same TS-mode
file has no jsDocDiagnostics. Both isolated comment and type-expression entry
points report TS1110 directly. Exact results are in `evidence/entrypoints.json`.

## Corpus measurement

`node-loader.mjs` loads the **actual adapted source parser**. Stock TypeScript
6.0.3 is used only to transpile syntax, enums and namespaces, following the parser
driver's loader. It does not substitute stock createSourceFile. `prepare-inputs.py`
uses the driver's 10,406 single-file cases and verifies every source hash and
BOM-decoded input hash. ScriptKind, filename, Latest target and absent parent links
match the driver. All 10,406 node totals match `cases-reference.json`: **1,323,111**
nodes. Attached JSDoc is visited separately before normal forEachChild descendants.

The cases contain **1,758 JSDoc roots**, **1,982 tags**, **1,417 type expressions**,
**37 links**, and **12,534 nodes inside JSDoc subtrees**: **0.947%** of all case
nodes. There are 34 JSDoc diagnostics. Counting descendants includes type nodes
and identifiers, rather than counting only kinds with a JSDoc prefix.

For current main's **82 regular compiler files**, the same traversal counts
968,181 nodes, 4,157 JSDoc roots, 3,854 tags, five type expressions and 56 links.
JSDoc descendants are 13,093 nodes (**1.352%**). This is a new main-tree measurement,
not a replacement for the historical fixed 81-file reference: that reference records
4,149 roots / 3,846 tags and its original SHA remains the native acceptance target.
Current main includes adapter-created hostErrors.ts, and its source adaptations
also differ. I did not reconstruct or retime that historical 81-file adapted tree.
The four directed inputs add four roots and one diagnostic, independently recorded.

The timing hook wraps **only JSDocParser.parseJSDocComment**, with a try/finally
clock and counters. The denominator sums createSourceFile wall time. Inputs are
already in memory; imports, file IO, output and tree traversal are excluded. Five
warmed rounds alternate ParseAll/ParseNone order. The same parser is then rerun
without instrumentation: all per-input projected tree hashes, counts, diagnostics,
mode probes and type controls are identical. The hook does not change source files.

Median inclusive JSDoc-parser time share is **1.877%** across the cases
(range **1.788-2.027%**), **12.673%** in salsa, **1.897%** in remaining cases,
and **3.160%** for current main's compiler corpus. The timer includes its own small
per-comment overhead and omits work outside the wrapped function, such as comment
range collection and later tag caching. It is a scoped observation, not an exact
uninstrumented cost attribution or a native performance prediction. Full group
counts, raw timings, ParseNone controls and ranges are in `MEASUREMENTS.md` and
`measurement.json`; compressed per-input evidence is under `evidence/`.

## Latent findings and wall routing

`latent-sites.csv` lists **592 unique observed sites: 192 Refused and 400 NotYet**,
with exact diagnostic file:line:column, reason, text, function and routing. The
complete projected unit/exclusion evidence is `latent-ledger.json`.

This is a projection of the committed **full latent census**, run
`stage3/meter/runs/20261008T035244Z.latent-full`, compiler source
`74fb6490a0ba7f608012a13c9842795f15431551`, instrumentation core
`93f7f8e0862b6460aaee0ac100bdec9c5bf435fb`. It was measured on a checker-rejected
program with ordinary Load/Lower and native output disabled. It is **historical**,
not a fresh current-main census. The parser.ts and factory source hashes match
current main exactly; scanner.ts and utility source hashes differ. Their findings
retain historical coordinates. No current-main Refused claim is inferred solely
from a historical row.

Selection covers named JSDoc functions and their nested declarations in parser,
scanner, nodeFactory, nodeTests, utilities and utilitiesPublic. Matching-source
scan findings are joined to their smallest stock compiler-API declaration;
changed-source findings are selected by their attempted unit. This is a conservative
JSDoc frontend surface, including constructors/tests, not an exhaustive transitive
reachability proof. Shared generic parser/scanner helpers can have additional
blockers. The ledger preserves 264 selected units and 515 exclusion/other records;
failed constructs can hide child expressions. Absence of a row is not proof of
lowerability.

Examples with exact reasons:

| Site | Observation | Wall routing |
| --- | --- | --- |
| parser.ts:8852:43 | NotYet `a value of type HasJSDoc` | heterogeneous node shape; no numbered assignment recorded |
| parser.ts:8811:31 | Refused `a boolean \| undefined as a condition` | truthiness frontend; codex/taste-not-soundness |
| parser.ts:8813:17 | NotYet `a PrefixUnaryExpression on a boolean \| undefined` | truthiness / unary frontend |
| parser.ts:8817:28 | NotYet `a method call through a structural signature in a program with statics; use typeof the declaring class` | callable structural dispatch; owner 01a1143c |
| parser.ts:8890:13 | NotYet `reading Debug` | namespace frontend; owner 01a113e7-bfed |
| parser.ts:8913:20 | NotYet `reading result` | isolated census context boundary; replay before treating as feature gap |

All-site routing uses `stage3/meter/owners.json` without inventing dispatch numbers.
Checked-cast/refinement rows route to #b5w3ycg / step09 scout; non-null rows to owner
01a1130a; predicates to 01a1143b-d691; enums/namespaces to compiler/stage3-front;
any/unknown adaptation to the step09 hatch investigation; typed-array rows, if any,
to step12. Unknown or context-sensitive rows retain OWNER BLANK. Each row's wall
label is a routing inference, separate from the observed compiler diagnostic.

I did **not** assign an authoritative numbered wall step to every remaining reason:
the repository does not contain that complete dispatch map. I also did not recount
current main's full latent closure or claim exhaustive expression-level JSDoc
blockers. The exact recorded sites and owner identifiers support that follow-up.

## Three real reduced paths

The selected tsc function bodies are unchanged:

1. `parser.ts:3836` parseJSDocAllType: `*` consumes one token and produces AllType.
2. `parser.ts:3842` parseJSDocNonNullableType: prefix `!` wraps a number type with
   postfix false.
3. `parser.ts:3848` parseJSDocUnknownOrNullableType: all six unknown-type delimiter
   branches and the nullable `?number` branch.

Each `.a` supplies a bounded token cursor and projected factory nodes, replacing
shared scanner/factory dependencies. These dependencies are deliberate reductions:
only number leaf types and the listed tokens are supported. `verify-witnesses.cjs`
checks each original body against the actual source with the compiler API and
catches its body mutant. `type-controls.mjs` compares kind, pos/end, number payload
and postfix against the real full parser. For `?=` and `?|`, it selects the inner
UnknownType before the full parser's optional/union wrappers, recording the outer
kind and diagnostics too. Flags, arbitrary types and enclosing syntax are excluded
from this witness projection.

All three originals compile on main and match Node stdout/stderr/exit. All three
sanitizer builds run clean; counted allocations equal frees (see `counts.md`).
The mutants change AllType to UnknownType, prefix to postfix, and remove the `}`
unknown branch. **Both source Node and native mutant outputs differ from the
original oracle**, and the native mutants equal their mutated source outputs.
No silent miscompile was observed. Every new `.a` is clean in place, so the header
rule requires no expected-error header.

These are grammar-path witnesses, **not** a complete native JSDoc parser. Full
comments, tags, links, diagnostic recovery, parent flags, generic type grammar and
node-for-node native corpus equality remain unproved. Directed source Node research
covers some of them; it is not promoted to native acceptance.

## Commands and evidence

Toolchain setup ran with `GOPROXY='https://proxy.golang.org|direct'` and succeeded:
Node 0.024s; Go 0.034s; markdown ready 0.104s (step 0.010s); submodules 0.139s;
clang 0.254s; go build 55.953s; tests deferred 56.081s; cache warm 56.082s;
done **56.111s**. `nproc` **5**, CPU quota four. Env:
`/workspace/adamic-tools/env.sh`. Setup/apply/build/measure/check output went to
files. No whole-package or full repository gate was run.

From the repository root on the pinned main, with a fresh `stage3/apply.sh` tree,
the pinned upstream checkout and stock TypeScript 6.0.3 installed externally:

```sh
source /workspace/adamic-tools/env.sh
export JSDOC_TYPESCRIPT=/workspace/scratch/step09-cache/api/node_modules/typescript/lib/typescript.js
python3 stage3/scouts/step24/jsdoc/prepare-inputs.py "$PWD" TREE UPSTREAM /tmp/step24-inputs.json > /tmp/step24-inputs.log 2>&1
JSDOC_TIMING=1 node --disable-warning=ExperimentalWarning --import ./stage3/scouts/step24/jsdoc/node-loader.mjs stage3/scouts/step24/jsdoc/measure.mjs TREE /tmp/step24-inputs.json > /tmp/step24-measure.json 2> /tmp/step24-measure.stderr
JSDOC_TIMING=0 node --disable-warning=ExperimentalWarning --import ./stage3/scouts/step24/jsdoc/node-loader.mjs stage3/scouts/step24/jsdoc/measure.mjs TREE /tmp/step24-inputs.json > /tmp/step24-control.json 2> /tmp/step24-control.stderr
node --disable-warning=ExperimentalWarning --import ./stage3/scouts/step24/jsdoc/node-loader.mjs stage3/scouts/step24/jsdoc/entrypoints.mjs TREE > /tmp/step24-entrypoints.json 2> /tmp/step24-entrypoints.stderr
node --disable-warning=ExperimentalWarning --import ./stage3/scouts/step24/jsdoc/node-loader.mjs stage3/scouts/step24/jsdoc/type-controls.mjs TREE > /tmp/step24-type-controls.json 2> /tmp/step24-type-controls.stderr
python3 stage3/scouts/step24/jsdoc/latent-ledger.py "$PWD" TREE "$PWD/stage3/scouts/step24/jsdoc" > /tmp/step24-latent.log 2>&1
python3 stage3/scouts/step24/jsdoc/audit.py /tmp/step24-measure.json /tmp/step24-control.json > /tmp/step24-audit.log 2>&1
node stage3/scouts/step24/jsdoc/verify-witnesses.cjs TREE > /tmp/step24-source-audit.log 2>&1
go build -o /tmp/step24-adamic ./cmd/adamic > /tmp/step24-build.log 2>&1
python3 stage3/scouts/step24/jsdoc/check-fixtures.py --compiler /tmp/step24-adamic --output /tmp/step24-fixtures-check > /tmp/step24-fixtures-check.log 2>&1
```

The fixture comparison passes all originals, sanitizer/count runs and three source
plus native mutants. The measurement audit catches missing-case, invented-JSDoc-count
and changed-latent-reason artifacts. Raw timing/control evidence, exact entry-point
probes and full-parser type controls are committed under `evidence/`.
