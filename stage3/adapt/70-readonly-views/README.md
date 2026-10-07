# Readonly views, adaptation 70

This is a partial unit. The first internal wave removed 22 measured refusals;
the second removed four more, and the third removes 21 net, for 456 to 409
cumulatively before wave 4. Wave 4 removes six more, for 456 to 403 overall.
Wave 5 removes another 54 with one public owner: 456 to 349 cumulatively.
The other 349 sites are retained and individually listed in [REMAINING.md](REMAINING.md).
I decline a blanket readonly conversion because the remaining mutation chains,
generic views, inferred contracts and erased escapes have not all been proven.
Public API readonly changes are now sanctioned by @system_adamic. Each public
owner still requires the same no-write proof, plus exact API-snapshot attribution.
The internal waves below make no public API edit.

## The edit

`adapt.cjs` inserts `readonly ` before `EvaluatorResult.value` in
`src/compiler/types.ts:6264`. The enclosing exported interface is explicitly
marked `@internal`. Its type parameter is constrained to string, number or
undefined. A narrower result can therefore be read through the wider result
without giving that view a writable replacement slot.

The stock npm TypeScript 6.0.3 checker resolves the declaration, not a line
number. The script loads all upstream `src` TypeScript files except library
fragments, including compiler, services, server and test infrastructure. It
checks every property and literal/computed element access against the selected
field's declaration, including assignments, destructuring targets, updates and
deletes. Same-declaration forwarding retains that identity through parameters,
returns, aliases and generic instantiations. Contextual flows and assertions
into any other declaration or erased view are declined. A different readonly
structural declaration is also declined, because a subsequent view could erase
its readonly modifier.

The audit found 65 field references, zero writes and zero different-declaration
or erased escapes. It plans everything before editing and refuses a newly found
writer or escape. The edit inserts nine bytes, preserving CRLF and every other
source byte. Incremental scoreboard: one file, one line added and one removed.
No Adamic compiler source, TypeScript source checkout or upstream ledger was
committed. The generated `stage3/patch-set.md` was restored because it is outside
this unit's territory.

This audit is declaration-based, not a general heap alias or effects analysis.
The script deliberately owns one slot; it does not automatically certify every
other target in the census. The broader owner inventory is evidence for future
work, not an interprocedural proof for those targets.

## Counts and provenance

The branch starts from `origin/main` at `e011f8f` and merges the requested heads:

| Prerequisite | SHA |
|---|---|
| Type imports | `a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e` |
| Optional declarations | `b7e379e054fc245b5a7d002d5abc62ec5ea42b20` |
| Latent census | `70456b766243ee155104c4c9c6928cc53b7b5640` |

Merge commit: `78abe5254a5a1c66bc101bee116d266b2316eb73`.
It was pushed as the prerequisite wave to `codex/stage3-readonly-views`.
No main push, force-push, rebase or pull request was made.

TypeScript input remains v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`.
The inherited cumulative report's 487 refusals were not assumed current.
The current compiler, after adaptations 10 and 20, measures 456.

| Controlled latent measurement | Before | After |
|---|---:|---:|
| Reasons containing ` seen as ` | 456 | 434 |
| Checker diagnostics | 2,168 | 2,168 |
| Compiler files | 78 | 78 |
| Newly eligible / no longer eligible units | 0 | 0 |

All 22 removed sites are `EvaluatorResult` wider views in `utilities.ts`.
No variance refusal was added. The normalized checker diagnostics are identical;
the compiler source hash manifests differ only for `src/compiler/types.ts`.
The full before/after reasons, exact removed sites, source hashes and eligibility
comparison are in [evidence/census.json](evidence/census.json).
Normalized raw event logs are saved as `evidence/before.jsonl.gz` and
`evidence/after.jsonl.gz`.

Every number is **measured on a checker-rejected program**. The latent tool still
stops at the first lowering error within each attempted unit and omits final
ownership and backend passes. These counts do not certify successful Adamic
compilation or an exhaustive whole-program blocker count.

The first exploratory comparison had two unrelated diagnostic differences:
installing upstream dependencies resolved `source-map-support`, and upstream's
build regenerated the adapted diagnostic import. The controlled reference was
therefore copied from the built adapted tree with the same dependencies and
only the selected readonly modifier removed through the stock AST. Its census
is the before column above. Preliminary run logs are retained but are not used
as the controlled proof.

## Declines

[PUBLIC_VIEWS.md](PUBLIC_VIEWS.md) preserves the initial 164-site public inventory,
with 93 of those measured locations remaining after wave 5. Their wider target
exactly matches an exported declaration or contract without `@internal`.
The user supplied @system_adamic's sanction; retained sites remain listed
until their own proofs and API attribution pass. Wave 5 edits the first public owner. The initial other 270 were conservative declines,
not an assertion that all of their declarations are internal. Four sites have
no explicit contextual owner resolved. Every remaining site's reason is in
`REMAINING.md`; the complete stock-checker owner inventory is in
`evidence/remaining-owners.json.gz`.

Two actual writers were inspected through the checker:

- `SourceFileLike.lineMap` is written at `scanner.ts:504:35`. The RHS is
  `number[]`, which the checker confirms fits the original SourceFile slot
  `readonly number[]`. This observed write does not put undefined into it.
  The broader view has latent replacement capacity and is public, so it stays
  unchanged. All 24 measured SourceFileLike views remain.
- `clear(array: unknown[])` writes `array.length = 0` at `core.ts:312:5`.
  It removes elements, inserting no wrong-typed value. Its parameter stays
  mutable because the consumer writes through it.

`SearchResult.value` was initially declined. With Adamic's strict options,
its flows at `moduleNameResolver.ts:1851:18`, `1851:40`, `1852:36` and `1856:18`
receive a contextual `any` view. The current audit cannot prove those flows.
Wave 3 resolves that bounded inferred-local alias and edits the declaration,
as proved below. Generic, method-variance and inferred
views are not converted on the strength of a textual reason alone.

[UPSTREAM_CANDIDATES.md](UPSTREAM_CANDIDATES.md) records the inspected writes and
the absence of a verified reachable wrong-typed counterexample. No unreviewed
site is mislabeled as a reachable or latent upstream bug.

## Proof

`bash cloud/setup.sh` completed successfully in 124 seconds. Go ready: 0s;
clang ready: 1s; Node ready: 1s; submodules ready: 1s; build cache warm: 124s.
The environment file was `/workspace/adamic-tools/env.sh`.
`nproc` printed 5; cgroup CPU quota was 4 cores; memory was 17.6 GB.
Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.
See `evidence/setup.log`.

Commands actually run, with output always written to logs:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/adaptation70-overlay > /tmp/adaptation70-overlay.log 2>&1
gofmt -w /tmp/adaptation70-overlay/*.go
go build -buildvcs=false -overlay=/tmp/adaptation70-overlay/overlay.json -o /tmp/adaptation70-census ./stage3/census/latent/tool > /tmp/adaptation70-census-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/adaptation70-census > /tmp/adaptation70-census-audit.log 2>&1
STAGE3_CACHE=/tmp/adaptation70-cache bash stage3/apply.sh /tmp/adaptation70-before > /tmp/adaptation70-apply.log 2>&1
```

The latent overlay built and its audit passed. Its signature/body and attribution
mutants were caught. The overlay cannot return IR or expose rejected programs to
the production loader. No production checker options were weakened.

```sh
# npm commands ran in /tmp/adaptation70-before; tooling ran in Adamic.
(cd /tmp/adaptation70-before && npm ci --no-audit --no-fund > /tmp/adaptation70-upstream-install.log 2>&1)
(cd /tmp/adaptation70-before && npm run build > /tmp/adaptation70-before-build.log 2>&1)
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before > /tmp/adaptation70-adapt.log 2>&1
(cd /tmp/adaptation70-before && npm run build > /tmp/adaptation70-after-build.log 2>&1)
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-reference/src/compiler /tmp/adaptation70-reference.jsonl > /tmp/adaptation70-reference-census.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-before/src/compiler /tmp/adaptation70-after.jsonl > /tmp/adaptation70-after-census.log 2>&1
```

Both upstream builds exited 0. All 10 emitted JavaScript files from the complete
default build have identical SHA-256 hashes before and after. Source maps and
internal declarations are not claimed byte-identical. Hashes are in
[evidence/javascript.json](evidence/javascript.json). A second adaptation run
changed zero files, and the hashes of all 710 upstream source files were
unchanged. The final audit was repeated after tightening structural declines.

The default stage 3 oracle command is:

```sh
bash stage3/oracle/run.sh /tmp/adaptation70-before /tmp/adaptation70-oracle > /tmp/adaptation70-oracle.log 2>&1
```

The unchanged default oracle **failed**: install exited 0, build exited 0,
tests exited 1. It ran all default suites with four workers, no filter, no timeout,
and reported **106,366 passing, 1 failing, 0 pending**. Wall time was 409.612s;
test time was 387.038s. The only failing test was `unittests:: Public APIs / for
typescript.d.ts / should be acknowledged when they change`. The only baseline
diff was `api/typescript.d.ts`, consisting of prerequisite adaptation 20's
optional undefined additions. All other baseline comparisons passed.

The initial run did not achieve an empty original-baseline diff. The user has
subsequently sanctioned exactly these mechanically proven 189 additions and
instructed that this inherited difference counts as passing. Later waves accept
only that proven snapshot in the disposable tree before running the oracle.
Adaptation 20's README describes accepting a mechanically reconstructed API
snapshot in its external tree. The initial wave did not change reference baselines or edit adaptation 20.
Later waves accept only the mechanically proven sanctioned snapshot in their
disposable checkout. The original
failure, report and diff are preserved under `evidence/oracle-*`. The readonly
owner is internal and absent from emitted `built/local/typescript.d.ts`; attribution
is checked separately below. The JavaScript manifest was rechecked after the
oracle and still has zero differences.

Read-only attribution used adaptation 20's stock-checker helper, without
`--accept-api`, against a freshly built pristine pinned control and its 373-owner
ledger. It passed: the emitted API is exactly the original snapshot plus the
**189** optional-owner undefined additions, and all **60,930** other original
reference baselines are identical. This proves the observed API diff contains
no extra readonly-owner change. See `evidence/api-attribution.json`.

```sh
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/20-optional-declarations/check-baselines.cjs /tmp/adaptation70-pristine /tmp/adaptation70-before /tmp/adaptation70-optional-owners.json > /tmp/adaptation70-api-attribution.json 2> /tmp/adaptation70-api-attribution.log
```

Byte identity is claimed for adaptation 70 against its required 10+20 input.
A separate pristine-versus-complete-series comparison found differences in
`_tsc.js`, `run.js` and `typescript.js`; those already exist in the pre-70 hash
manifest. This unit makes no byte-identity claim for the entire prerequisite
series against pristine. Both builds' complete manifests are preserved in
`evidence/pristine-javascript.json`.


Four independent scratch mutants were run and restored:

| Mutant | Result |
|---|---|
| Mark public `SourceFileLike.lineMap` readonly although getLineStarts writes it | Upstream `npm run build` exited 1; TS2540 at `scanner.ts(504,46)`: Cannot assign to 'lineMap' because it is a read-only property. |
| Add a typed consumer assigning `result.value = undefined` through EvaluatorResult | Adaptation exited 1, naming `types.ts:10673:5` in its writes list; it edited no source. |
| Forward the result into `{ value: string \| number \| undefined }` and write through that view | Adaptation exited 1 with an empty direct-writes list and the distinct structural escape at `types.ts:10675:9`; it edited no source. |
| Remove EvaluatorResult's `@internal` tag | Adaptation exited 1: public declaration excluded: EvaluatorResult. |

The latter three exercise the writer, escape and public-owner guards separately.
The first is the requested actual-writing-consumer readonly mutant, caught by
the checker through upstream's own build. Sources were restored before the
oracle. Logs and `mutant.cjs` preserve the evidence and recipes.

## Re-running the evidence

`node adapt.cjs <tree>` uses stock npm typescript@6.0.3 via NODE_PATH or
`CENSUS_TYPESCRIPT`. It tolerates earlier adaptations and parses current text.
Run `survey.cjs <tree> <latent.jsonl> <output.json>` to inventory owners, and
`writers.cjs <tree> <output.json>` for the two inspected writer families.
`summarize.py BEFORE_TREE BEFORE_JSONL AFTER_TREE AFTER_JSONL OUTPUT` preserves
the controlled census and regenerates the complete remaining-site list.
Run mutants only in a disposable copy and restore its source after each run.

Not covered: a transitive write/reachability proof for every retained target,
proofs for the remaining owners, additional public API readonly edits, complete internal reachability for every retained writer,
an empty diff against all original upstream references, the full native Adamic
gate or native compilation of tsc. This unit changes only
stage 3 adaptation tooling and type declarations in the disposable input tree.

## Internal wave 2

Seven more shallow slots are readonly: ParsedPatterns.matchableStringSet and
patterns; Resolved.originalPath; ModuleResolutionState.reportDiagnostic; and
ModuleSpecifierResult.kind, moduleSpecifiers and computedWithoutCache. Every
owner's references, destructuring copies, zero writes and zero retained-view
escapes are recorded in evidence/wave2/adapt.json. Binding patterns (including
nullish fallback expressions ending in a binding pattern) copy slots out; they
do not retain the receiver. Scalar assignment to a copied local cannot replace
the original object's field. This proof concerns replacement of the shallow
slot, not mutation inside a mutable field value.

The controlled census is 434 to 430, four removed and zero added, with the
same 2,168 checker diagnostics and no eligibility changes. The four removals
are the reportDiagnostic callback view, a Resolved.originalPath view, the false
computedWithoutCache view, and a ParsedPatterns set slot. All 10 emitted JS
files are byte-identical. The second adaptation run changes zero files.
Evidence is preserved separately under evidence/wave2.

The direct-writer, structural-escape and public-owner mutants were independently
rerun and all exited 1 before editing source. They remain the checks for an
unselected public declaration until the public wave's explicit owner list is
introduced. The actual-writing-view build mutant from wave 1 remains TS2540.

ReusableBuilderProgramState.program is declined: observed writes at builder.ts
326:5 and 619:5 and deletion at 2460:31 prevent readonly certification. Those
operations are not by themselves proof of a reachable wrong-type insertion.
SearchResult.value remained declined in wave 2 because of contextual any at the
four previously listed sites; wave 3 resolves this particular alias. The broad field audit is preparatory evidence, not a
claim that these alias paths or array element owners have been proven.

Wave 2's default oracle passed: 106,367 passing, zero failures, zero pending,
empty baseline diff. Before this run, adaptation 20's check-baselines.cjs
mechanically verified exactly 189 optional additions and accepted that sanctioned
API snapshot in the disposable checkout. Every other reference baseline remained
byte-identical. No public readonly addition exists in this wave. The snapshot
verification and oracle report are in evidence/wave2.

Wave 2 commands (all output logged):

```sh
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
(cd /tmp/adaptation70-before && npm run build)
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-wave2-reference/src/compiler /tmp/adaptation70-wave2-before.jsonl
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-before/src/compiler /tmp/adaptation70-wave2-after.jsonl
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/20-optional-declarations/check-baselines.cjs /tmp/adaptation70-pristine /tmp/adaptation70-before /tmp/adaptation70-optional-owners.json --accept-api
bash stage3/oracle/run.sh /tmp/adaptation70-before /tmp/adaptation70-wave2-oracle
```


## Internal wave 3

Eleven internal TextRange parameters become Readonly<TextRange>: range on
createDiagnosticForRange, nodeIsSynthesized, moveRangeEnd, moveRangePos,
rangeIsOnSingleLine, getStartPositionOfRange and parseErrorAtRange; range1 and
range2 on rangeStartIsOnSameLineAsRangeEnd; and node on emitDetachedComments
and emitNewLineBeforeLeadingComments. SearchResult.value also becomes readonly.
These are 12 owner edits across three compiler files. No public declaration
changes in this wave.

[Per-owner proofs](evidence/wave3/adapt.json) record checker-resolved references
and transitive consumer edges. The parameter audit scans each parameter's symbol
through its entire function, including nested closures, then follows resolved
whole-object arguments into upstream function bodies. It permits primitive
property reads and declines casts, retained aliases, returned receivers, object
children, accessors and unresolved consumers. Assignments, destructuring targets,
updates, deletes, iteration targets and stock container mutators are writes.
All eleven owners and their reachable consumers have zero writes and escapes.

SearchResult's slot has 25 resolved references and zero writes or escapes.
The four contextual-any flows lead to the same inferred let result local at
moduleNameResolver.ts:1847:9. Its declaration has neither an annotation nor an
initializer. Every reference to that exact local symbol, including in nested
closures, is checked: six reads, two rebindings, no slot writes or retained
aliases. Rebinding the local does not replace the referenced object's slot.
Explicit any, assertions, other member accesses and whole-object forwarding
remain declines. This bounded proof closes the earlier SearchResult decline;
it is not a general heap alias analysis.

The clean post-restart census completed all 78 compiler files (79 event rows).
It measures **430 to 409**, with **25 removed reasons and four newly exposed
reasons**, and no eligibility changes. The before report reuses wave 2's actual
measurement only after verifying all 78 compiler-source hashes match the
controlled reference; see evidence/wave3/before-provenance.json. An interrupted
73-file run was discarded. Every remaining reason is in REMAINING.md.

Three exposed reasons concern the nested isExternalLibraryImport boolean slot
at moduleNameResolver.ts:1924:28, 1931:28 and 1950:20. The fourth, also at 1950:20,
concerns a fresh inferred object view in a logical expression. They are retained:
this wave proves the SearchResult replacement slot, not the nested payload or
that inferred view. These reasons do not establish a reachable wrong-typed
write. The 2,168 strict-checker diagnostics remain the same in number; two
existing messages acquire readonly type spelling, without a new location or
error code. Exact message changes are in evidence/wave3/diagnostic-delta.json.

The complete upstream build exited 0. All ten emitted JavaScript files match
the pre-wave hashes. Idempotence changes zero files and preserves all 710
upstream TypeScript source hashes. Adaptation 20's existing check-baselines.cjs
proves the final published API equals the original plus exactly 189 sanctioned
optional additions; all 60,930 other reference baselines are identical. There
are zero public readonly additions to attribute in this wave.

The final default oracle passed **106,367 tests, zero failures, zero pending**,
with an empty baseline diff. Install and build exited 0; test time was 248.534s,
wall time 254.144s. The oracle ran with adaptations 10, 20 and all current 70
edits, after mechanical acceptance of only adaptation 20's sanctioned snapshot.
Reports, hashes and compressed logs are under evidence/wave3.

Nine final-source mutants prove the guards and upstream check:

| Mutant | What caught it |
|---|---|
| Direct range.pos assignment | Parameter audit, exit 1 before edits |
| Assignment inside getStartPositionOfRange | Transitive caller audit, exit 1 before edits |
| Typed alias retaining a range, followed by a write | Parameter escape audit, exit 1 before edits |
| Destructuring assignment to range.pos | Parameter writer audit, exit 1 before edits |
| for-of assignment to range.pos | Parameter writer audit, exit 1 before edits |
| for-of assignment to EvaluatorResult.value | Field writer audit, exit 1 before edits |
| Inferred receiver forwarded into any and its slot written | Local-alias audit, exit 1; upstream compiler build exits 0 |
| Inferred receiver changed to explicit any | Local-alias audit, exit 1 |
| SourceFileLike.lineMap marked readonly despite its writer | Upstream compiler build exits 1, TS2540 at scanner.ts(504,46) |

The erased-writer build was rerun in a complete disposable checkout after an
initial source-only scratch lacked package.json (npm exit 254). That setup failure
is not counted as a checker result. All mutated source was restored.

Final proof commands, each with stdout and stderr captured in a log:

```sh
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
(cd /tmp/adaptation70-before && npm run build)
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-before/src/compiler /tmp/adaptation70-wave3-complete-after.jsonl
python3 stage3/adapt/70-readonly-views/summarize.py /tmp/adaptation70-wave3-reference /tmp/adaptation70-wave3-before.jsonl /tmp/adaptation70-before /tmp/adaptation70-wave3-complete-after.jsonl stage3/adapt/70-readonly-views/evidence/wave3 stage3/adapt/70-readonly-views/REMAINING.md
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/20-optional-declarations/check-baselines.cjs /tmp/adaptation70-pristine /tmp/adaptation70-before /tmp/adaptation70-optional-owners.json
bash stage3/oracle/run.sh /tmp/adaptation70-before /tmp/adaptation70-wave3-complete-oracle
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
```

Public readonly owners, their downstream consumer proof and the requested
UPSTREAM_NOTE.md remain work for a subsequent public wave. This wave does not
claim a complete write analysis of all remaining internal or public sites.


## Internal wave 4

The highest-yield newly proved internal owner is the return payload's
isExternalLibraryImport field in nodeModuleNameResolverWorker.tryResolve.
Four references, no writes, no aliases or escapes: its one readonly insertion
removes four measured reasons. The selected function and inline return member
are found by their nested AST owner path, not line numbers. The location parameter
of createMemberAccessForPropertyName becomes Readonly<TextRange>. Its entire
consumer chain is setTextRange's input location, which reads pos and end and
writes only the separate destination. The two edits touch two source files.

The controlled census is **409 to 403**: seven removed, one added, 78 files,
2,168 byte-identical normalized diagnostics, no eligibility changes. The new
reason at factory/utilities.ts:190:99 is the now-readonly location flowing into
public setTextRange's still-mutable input. That public owner is sanctioned and
queued for the next public wave. Every retained reason is in REMAINING.md.
The before measurement is wave 3's measured after input, after verification of
all 78 compiler-source hashes. A new completeness assertion rejects any missing
or duplicate compiler file. A deliberately truncated event log fails that
assertion with exit 1; no partial count is used as final evidence.

Complete upstream build: exit 0. All ten JavaScript hashes remain identical,
including after the oracle. Idempotence: zero files edited. Published API:
exactly adaptation 20's 189 sanctioned optional additions, no public readonly
addition; all 60,930 other reference baselines identical. Default oracle:
**106,367 passing, zero failing, zero pending, empty diff**, 273.957s wall time.
The payload writer mutant is rejected by the field audit before edits, and
upstream's compiler build independently catches it with **TS2540 at
moduleNameResolver.ts(1914,85)**. Earlier generic writer/escape mutants remain
applicable to the unchanged guards. Logs and owner proofs are in evidence/wave4.

Ranked internal declines, observed through stock-checker declarations and calls:

| Owner or view | Observed writes | Assessment |
|---|---|---|
| mutateMapSkippingNewValues.map (nine sites) | utilities.ts:8220:13, map.delete(key) | Actual write removes entries; no wrong-type insertion. Element replacement capacity remains latent. |
| multiMapSparseArrayAdd.map (four sites) | transformers/utilities.ts:380:9, values.push(value); 383:9, map[key] = values = [value] | Real insertion through the broader view; a wrong-type path into a particular original array was not established. Retained pending reachability proof. |
| FlowGraphNode.edges | debug.ts:1033:13 and 1034:13, source/target.edges.push(edge) | Real element insertion; shallow slot readonly would not fix it. No executed narrow-type witness yet. |
| ModuleResolutionState lookup-location slots | program.ts:1362:9 and 1363:9, replacement with string[] | Real replacement; the particular never[] caller's path to these writes is unverified. |
| SourceFileLike.lineMap | scanner.ts:504:35 | Cache writer prevents readonly; its number[] RHS fits the original slot. |
| clear.array (four sites) | core.ts:312:5, array.length = 0 | Actual mutation removes elements; no wrong-type insertion. |
| addRelatedInfo.relatedInformation (13 sites) | utilities.ts:10357:5 stores its elements in diagnostic.relatedInformation | Retained alias into the mutable diagnostic family. Array readonly alone would not prove those element views safe. |

The preparatory field and parameter inventories are recorded beside this wave.
They rank owners and report conservative escapes, including incomplete overload
and callback chains; they do not certify every remaining site. Other unproved
owners remain honest declines. No upstream candidate is inferred from a writer
location alone. Public-owner edits and their API compatibility note follow in a
separate proven wave.


## Public wave 5

The highest-yield public no-write owner is setTextRange.location, changed from
TextRange | undefined to Readonly<TextRange> | undefined. Its three symbol-bound
reads are the truthiness check and pos/end accesses. It is never retained,
returned or passed to a writer. The input proof and every previously selected
owner's repeated audit are in evidence/wave5/adapt.json. The writing destination
range stays mutable. Readonly restricts this input view; it does not freeze the
object or rule out another alias to it.

**One owner edit removes 54 refusals: 403 to 349**, zero added. All 78 compiler
files were measured. Normalized strict diagnostics are byte-identical, still
2,168, and eligibility is unchanged. The before control matches all 78 hashes
of wave 4's actually measured after tree. Every remaining reason is preserved
in REMAINING.md. Of the initial conservative 164-site public inventory, 93
locations remain; that retained subset is not a fresh exhaustive public census.

Upstream full build exits 0. All ten emitted JavaScript files are byte-identical,
rechecked after the oracle. Idempotence edits zero files and preserves all 710
source hashes. The new stock-AST check-api.cjs reconstructs the original public
snapshot from adaptation 20's owner ledger and this wave's public-owner proof.
The actual emitted API is byte-for-byte exactly **189 sanctioned optional owner
additions plus one readonly parameter addition**, 190 changed lines, and nothing
else. All 60,930 other reference baselines are unchanged. Only this mechanically
verified API snapshot is accepted in the disposable checkout before the oracle.

The default oracle passes **106,367 tests, zero failures, zero pending, empty
baseline diff**. Install/build/tests exit 0, wall time 253.217s. It uses adaptations
10, 20 and all 70 edits. No test filter or native Adamic gate was run for this
adaptation-only wave. All output is logged under evidence/wave5.

Mutants and compatibility evidence:

| Input | Observation |
|---|---|
| Add a write to the public location input | Audit exits 1 before edits; upstream compiler build exits 1 with TS2540 at factory/utilitiesPublic.ts(12,14). |
| Downstream consumer derives View from the published location parameter and assigns pos/end | Before API: zero diagnostics. After API: two TS2540 diagnostics. This is the requested compatibility witness. |
| Change unowned API version declaration from string to number in a scratch artifact | Exact API reconstruction rejects it, exit 1. |
| Replace only a scratch compiled setTextRange body with return range | Reachable-hole Node assertion rejects [0,0] instead of observed [1,2], exit 1. The initial log-format check was corrected to Node's actual diff format; the mutation itself was caught. |

UPSTREAM_NOTE.md is the requested public-API ledger-style compatibility entry,
including a draft Microsoft could accept or refuse. The separate reachable
setTextRange destination counterexample is in UPSTREAM_CANDIDATES.md. Stock 6.0.3
accepts a readonly literal-zero destination, then pristine and adapted Node
both observe [1,2] through variables typed [0,0]. Its real writer declarations
are resolved in evidence/wave5/destination-writes.json. This destination is
unchanged; no narrow instance among compiler-internal callers is claimed.
Neither note was sent upstream, and the global ledger remains untouched.

Final wave 5 commands (stdout/stderr captured in logs):

```sh
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
(cd /tmp/adaptation70-before && npm run build)
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-before/src/compiler /tmp/adaptation70-wave5-after.jsonl
python3 stage3/adapt/70-readonly-views/summarize.py /tmp/adaptation70-wave5-reference /tmp/adaptation70-wave5-before.jsonl /tmp/adaptation70-before /tmp/adaptation70-wave5-after.jsonl stage3/adapt/70-readonly-views/evidence/wave5 stage3/adapt/70-readonly-views/REMAINING.md
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/check-api.cjs /tmp/adaptation70-pristine /tmp/adaptation70-before /tmp/adaptation70-optional-owners.json /tmp/adaptation70-wave5-adapt.json --accept-api
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/consumer.cjs /tmp/adaptation70-wave5-before.d.ts /tmp/adaptation70-before/built/local/typescript.d.ts
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/narrow-range.cjs /tmp/adaptation70-pristine/built/local/typescript.d.ts /tmp/adaptation70-pristine/built/local/typescript.js
bash stage3/oracle/run.sh /tmp/adaptation70-before /tmp/adaptation70-wave5-oracle
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
```

The 349 retained reasons still include genuine writers, structural escapes,
container-element and callback contracts, public diagnostic families and inferred
views needing further owner proofs. This remains a partial unit with measured
progress and explicit declines, rather than a claim all remaining views are safe.


## Decline wave 6

No new owner edit is justified by this wave. The clean census remains **349 to
349**, all 78 files measured, zero removed or added, identical 2,168 normalized
diagnostics and no eligibility changes. Source manifests are identical. The
whole build's ten JavaScript hashes are identical after the oracle. Idempotence
changes zero files and preserves all 710 source hashes. The exact API delta
remains 189 optional additions and one readonly input addition; 60,930 other
references remain identical. No extra snapshot was accepted.

The default oracle again passes **106,367 tests, zero failures, zero pending,
empty diff**, 293.159s wall time. Logs are in evidence/wave6. No native Adamic
compiler source changed, and the full native gate was not run.

OWNER_RANKING.md records the highest unresolved groups and precise decline
reasons. The stock checker inventories 349 sites and 574 field declarations.
The ranking lists 390 non-readonly field candidates with context coverage,
reference counts, observed direct-write counts and unresolved-path counts.
Context coverage is not promised net removal; field scores overlap and are not
added. The complete context and field observations are saved as compressed JSON.

The diagnostic family's file, start and length slots share 26 contexts. Its
base slots have no direct writes, but respectively 30, 31 and 35 unresolved
structural or erased paths. Inferred diagnosticMessage locals forward the
receiver to diagnostic helpers, and an erased refactor array retains it. The
bounded SearchResult local-alias rule therefore cannot certify this family.
Readonly on just the base slot could also expose a derived slot or a deeper
refusal; no edit is made until the coupled view proof is complete.

SourceFileLike remains blocked by its cache writer. The 14 iterator-result
reasons require a proof through the iterator protocol. Rest diagnostic elements,
inferred map/array unions and sparse-array insertion likewise remain explicit
writers or unproven aliases. A zero direct-write count never establishes a
whole-program no-write proof by itself. This is an honest decline wave, not a
claim of an additional readonly adaptation or full reachability classification.

The field-audit.cjs survey is now committed with a new mutant: a typed
DiagnosticRelatedInformation consumer assigning file = undefined changes the
reported write count from zero to one at the injected source location. The
survey does not falsely certify it. The actual-writing-view readonly mutant
was separately repeated: upstream build exits 1, **TS2540 at scanner.ts(504,46)**
for SourceFileLike.lineMap. Every scratch source was restored.

The setTextRange executable fixture and UPSTREAM_CANDIDATES.md are byte-identical
to wave 5; their SHA-256 hashes are in evidence/wave6/counterexample-identity.json.
Its exact pristine/adapted observations and reachable classification are unchanged.

Commands, all with stdout/stderr captured in logs:

```sh
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/survey.cjs /tmp/adaptation70-before /tmp/adaptation70-wave5-after.jsonl /tmp/adaptation70-wave6-owners.json
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/field-audit.cjs /tmp/adaptation70-before /tmp/adaptation70-wave6-owners.json.gz /tmp/adaptation70-wave6-fields.json.gz
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation70-census /tmp/adaptation70-before/src/compiler /tmp/adaptation70-wave6-after.jsonl
python3 stage3/adapt/70-readonly-views/summarize.py /tmp/adaptation70-before /tmp/adaptation70-wave5-after.jsonl /tmp/adaptation70-before /tmp/adaptation70-wave6-after.jsonl stage3/adapt/70-readonly-views/evidence/wave6 stage3/adapt/70-readonly-views/REMAINING.md
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/adapt.cjs /tmp/adaptation70-before
NODE_PATH=/tmp/adaptation70-cache/api/node_modules node stage3/adapt/70-readonly-views/check-api.cjs /tmp/adaptation70-pristine /tmp/adaptation70-before /tmp/adaptation70-optional-owners.json /tmp/adaptation70-wave6-idempotence.json
bash stage3/oracle/run.sh /tmp/adaptation70-before /tmp/adaptation70-wave6-oracle
```
