# Readonly views, adaptation 70

This is a partial unit. The first internal wave removed 22 measured refusals;
the second removes four more, for 456 to 430 cumulatively.
The other 430 sites are retained and individually listed in [REMAINING.md](REMAINING.md).
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

[PUBLIC_VIEWS.md](PUBLIC_VIEWS.md) lists 164 remaining sites whose wider target
exactly matches an exported declaration or contract without `@internal`.
The user has now supplied @system_adamic's sanction; these sites remain listed
until their own proofs and API attribution pass. The initial other 270 were conservative declines,
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

`SearchResult.value` was considered but declined. With Adamic's strict options,
its flows at `moduleNameResolver.ts:1851:18`, `1851:40`, `1852:36` and `1856:18`
receive a contextual `any` view. The current audit cannot prove those flows.
No change was made to that declaration. Generic, method-variance and inferred
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
snapshot in its external tree. This unit does not change reference baselines,
accept a snapshot to make the run green, or edit adaptation 20. The original
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
additional readonly owner edits, public API rulings, a reachable upstream bug,
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
SearchResult.value remains declined because of contextual any at the four
previously listed sites. The broad field audit is preparatory evidence, not a
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
