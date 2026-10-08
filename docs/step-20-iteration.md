# Step 20: the iteration protocol

## Evidence and counting

Delivery base: `dcdbb9098f77f30ad41790c56df1bd63ad462b63`, the merged
`origin/compiler/area-next-fixtures` tip. The delivery branch is `codex/scout-iteration`.
The ranking branch `6c4fc1afb019d0a16fea97082e45fef8fcbe1746` is evidence only;
it is not merged into this branch.

A root in these tables is a distinct diagnostic `where` for an exact kind/reason.
An attempted unit is counted separately: recovery can report the same site from
several enclosing units. These are diagnostic roots, not successfully compiled
entry roots. The inventory retains every site and unit in JSON.

Hidden bytes below are historical, outermost-boundary credits from the ranking's
compiler `ed6e29751ee47d86fad450cd1674139883bc0f70`. They are estimates of bytes
revealed if that reason alone were fixed, not bytes compiled by this change.
Zero means zero credited bytes; unmeasured means the ranking has no such row.
The inherited full census also predates this delivery base (`74fb6490...`);
neither artifact is presented as a fresh base measurement.

### Historical hidden ranking

| Kind | Exact reason | Diagnostic roots | Attempted units | Hidden bytes | Diagnostic witnesses |
| --- | --- | ---: | ---: | ---: | --- |
| NotYet | `for...of over an object` | 120 | 124 | 13,952 | `utilities.ts:10736:13`; `binder.ts:375:17`; `binder.ts:432:13` |
| NotYet | `a for...of destructuring an object` | 4 | 4 | 1,106 | `checker.ts:26333:13`; `checker.ts:42067:17`; `program.ts:798:13` |
| NotYet | `for...of over a union of differently held members` | 1 | 1 | 123 | `core.ts:438:5` |
| Refused | `a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)` | 1 | 1 | 109 | `parser.ts:2356:5` |
| NotYet | `a YieldExpression as a statement` | 1 | 1 | 26 | `core.ts:1045:9` |
| NotYet | `iterating a value` | 1 | 1 | 0 | `programDiagnostics.ts:219:17` |

Rows with fewer than three distinct sites list all real sites. Repeating or
inventing witnesses would conceal the size of the evidence. Full boundary records,
ranking provenance and artifact hashes are in [hidden.json](step-20-iteration/hidden.json).
The first row has 141 raw records, 120 distinct boundaries and 25 boundaries with
nonzero credits. The destructuring row has four records and two credited boundaries.

### Refusal table

The ranking bundles the refusal table from `d35a81d36fdafccf827bad0f572d311b2a0d4deb`.
It records `a generator function` at six sites and `yield (generators)` at eight;
all are original-source sites. Neither has a hidden-byte ranking row, so their
bytes are **unmeasured**. Its published examples are `checker.ts:21677`,
`checker.ts:21685` for generator functions and `checker.ts:21681`,
`checker.ts:21693` for yield. The table supplies only two witnesses per reason;
the fresh census below supplies additional real sites where observable.
The owner is `internal/lower/refusals.go` / `lowering.refuse`; the step 20 ruling authorizes replacing these
refusals with correctly implemented generator behavior.

### Scope exclusions

`reading iterator` and `reading iteratorValueStatement` describe variables in
TypeScript's emitter, not proof that those roots need runtime iterator support.
They are retained as contextual candidates in the census JSON, excluded from
step retirement totals. `for...in over an array ... use for...of for elements`
is property enumeration, outside this step. `YieldExpression` AST model views
are not runtime generators. Name matches alone do not assign a root to step 20.

### Fresh base census

The guarded measurement covers all **81 resolved non-library source files** on the
delivery base, from TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`
(v6.0.3), adapted by `stage3/apply.sh`. The entry is `src/tsc/tsc.ts`.
The checker reports **324 diagnostics**: this is a measurement on a rejected
entry, not a claim that the TypeScript compiler builds. No output was emitted.

| Kind | Exact reason | Roots | Units | Historical hidden bytes | Witnesses |
| --- | --- | ---: | ---: | ---: | --- |
| NotYet | `a YieldExpression as a statement` | 1 | 1 | 26 | `core.ts:1045:9` |
| NotYet | `a for...of destructuring an object` | 4 | 4 | 1,106 | `checker.ts:26333:18`; `checker.ts:42067:22`; `program.ts:798:18` |
| NotYet | `for...of over a union of differently held members` | 1 | 3 | 123 | `core.ts:2173:29` |
| NotYet | `for...of over an object` | 125 | 129 | 13,952 | `binder.ts:1296:36`; `binder.ts:2061:33`; `binder.ts:2809:33` |
| NotYet | `iterating a value` | 1 | 1 | 0 | `programDiagnostics.ts:219:39` |
| Refused | `a generator function` | 12 | 12 | unmeasured | `checker.ts:21677:5`; `checker.ts:21685:5`; `checker.ts:21781:30` |
| Refused | `a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read` | 14 | 1 | unmeasured | `sourcemap.ts:504:48`; `sourcemap.ts:505:52`; `sourcemap.ts:511:52` |
| Refused | `optional property return in ArrayIterator<JSDocLink | JSDocLinkCode | JSDocLinkPlain | JSDocText> absent from structural source ArrayIterator<JSDocComment>, which can hide fields` | 1 | 1 | unmeasured | `parser.ts:9508:101` |
| Refused | `optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<JSDocComment>, which can hide fields` | 13 | 5 | unmeasured | `parser.ts:1054:92`; `parser.ts:1060:89`; `parser.ts:1073:89` |
| Refused | `optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<JsonObjectExpressionStatement>, which can hide fields` | 14 | 9 | unmeasured | `commandLineParser.ts:2445:13`; `commandLineParser.ts:2498:65`; `commandLineParser.ts:2503:65` |
| Refused | `yield (generators)` | 15 | 12 | unmeasured | `checker.ts:21681:13`; `checker.ts:21693:17`; `checker.ts:21782:33` |

The ArrayIterator optional-return refusals are iteration-related structural view
boundaries. They must preserve protocol state and exception behavior; removing
a structural check alone would not implement them. The IteratorResult refusal
separately exposes incompatible yield/completion payloads. Historical symbol-key
reflection remains in the table above even though recovery did not reach it here.
`strictBuiltinIteratorReturn` in a CompilerOptions property name is an unrelated
configuration view, excluded along with emitter variable-name matches.

### First shape by census

The source contains 651 for-of expressions. Genuine arrays account for 466
and already have a fast path; 143 are NodeArray views, six are
SortedReadonlyArray views and six are JSDocArray views. These are source loop
counts, separate from diagnostic roots. Matching diagnostic and expression
locations by file and line gives the refused shapes in [shapes.json](step-20-iteration/shapes.json).
NodeArray is the largest unsupported concrete shape and is implemented first.

### Reproduction and audit

The archive [base-census.jsonl.gz](step-20-iteration/base-census.jsonl.gz) has
one checker header and exactly one record per resolved source file.
[base.json](step-20-iteration/base.json) retains every diagnostic site and attempted unit.
The decompressed SHA-256 is
`d1336db1f97fc3ff86471d028d84df4e5a099b2fb11fdedb5a0fa0fa2b7b4c49`.

The initial full process was interrupted by an environment restart after 29
complete file records. The remaining 52 files were measured individually with
a scratch overlay filtering only the outer per-file emission loop; checker files,
declaration registration and attempted-unit recovery remained unchanged. Every
header matches exactly. A repeated corePublic.ts record matched the original
full run exactly. Interrupted processes have no claimed successful exit status.

Commands run (each redirected to its own log):

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/scout-overlay
python3 stage3/meter/entry_overlay.py /tmp/scout-overlay /tmp/scout-entry-overlay
go build -buildvcs=false -overlay=/tmp/scout-entry-overlay/overlay.json -o /tmp/scout-census ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/scout-census /tmp/scout-adapted/src/tsc/tsc.ts /tmp/scout-full.jsonl
python3 docs/step-20-iteration/census.py docs/step-20-iteration/base-census.jsonl.gz /tmp/base.json --compiler dcdbb9098f77f30ad41790c56df1bd63ad462b63
python3 docs/step-20-iteration/hidden_audit.py
```

The fresh extractor audits 14 exact candidate reasons. Mutants that increment
roots, increment units, invent a witness, or omit a source file each fail their
corresponding assertion. The historical audit verifies six reasons against the
pinned ranking; byte, root, witness and boundary mutations each fail. No byte
credits from different compiler bases are added together.
