#!/usr/bin/env python3
"""Render the continuation only from complete matrix, oracle and census receipts."""
import json
from pathlib import Path
import subprocess
p=Path(__file__).resolve().parent;repo=p.parents[2]
def read(name):return json.loads((p/'composition'/name).read_text())
labels=['x','32x','95hx','both'];matrix={k:read('matrix/'+k+'.json') for k in labels};oracles={k:read('oracle/'+k+'/union.json') for k in ['x','32x']};latent=read('latent/summary.json')
assert all(v['status']=='pass' and v['types']['diagnostics']==0 and v['javascript']['pass'] for v in matrix.values())
assert all(v['status']=='pass' and v['units']==398 and v['identical_test_identities_and_statuses'] and v['identical_sanctioned_API_bytes'] for v in oracles.values())
assert latent['status']=='pass'
final_identity=read('oracle-final-source-identity.json');assert final_identity['status']=='pass' and final_identity['matrix_unchanged']
for name in ['owners.json','inventory.json','ownership.json']:
 assert not subprocess.check_output(['git','diff','a6d5f494','--',str((p/name).relative_to(repo))],cwd=repo)
rows='\n'.join('| '+label+' | '+str(v['source']['expectedSites'])+' | '+str(v['source']['replay']['left'])+' | 0 | Pass | '+', '.join(v['source']['importsRemoved'])+' |' for label,v in matrix.items())
cw='\n'.join('| '+r['name']+' | '+r['before']['where']+' | '+r['after']['where']+' |' for r in latent['comparison'])
timings=[]
for label in ['x','32x']:
 for name in ['preparation','whole']:
  value=read('oracle/'+label+'/'+name+'.json');timings.append((label+' '+name,value['metrics']['seconds']))
 for i in range(10):timings.append((label+' shards '+str(i),read('oracle/'+label+'/piece-'+str(i)+'.json')['metrics']['seconds']))
assert all(seconds<1500 for _,seconds in timings)
text='''# Fold composition continuation: complete

This continuation starts from typescript/maybebind-adapt-1,
a6d5f494e46262f805e5be76a4d7c108fe6a3a6c, on main
322f80565edaa978d928ca559803c578ed67c105. It is published plainly to
typescript/maybebind-adapt-2. All shipping changes remain in 99x's directory.
The old proof and Packed evidence section below remain historical audit data.
The current continuation's complete receipts are under composition/.

## Adapter rule and ownership

For every ownership.json row, 99x counts only the inventoried original core
calls under that owner. Every original must remain, or every original must
be absent. A nonempty proper subset or duplicate refuses and names the owner.
No replacement expression is validated or edited. Other helper call shapes
inside a completely folded owner stay opaque; the control proves this too.
The overlapping createProgram/getEmitHost rows each keep this rule.

The 28 own sites and ten exact before/after owner texts and SHA-256 pins are
unchanged. owners.json, inventory.json and ownership.json have no changes
from a6d5f494. Changed, missing, duplicate and half-adapted own owners still
refuse before writes, as do mixed own-file/tree states. Full programs use
stock TypeScript 6.0.3 symbol identity to distinguish core calls and binder's
local helper. No cast is added; core.ts's body is unchanged.

After the 28 own edits, each of the five eligible source files keeps its
maybeBind import exactly when a core helper call remains. inspect uses the
file's symbol-resolved core-call count rather than inventory.imports. The
historical import inventory stays as a provenance record, not an acceptance
rule. A complete 32 fold lets 99x remove watch.ts's import. A complete program
owner fold lets 99x remove program.ts's import; the synthetic complete-fold
control proves that case. checker.ts has no 99x site, so its source and import
are untouched by 99x. Checker import cleanup belongs to 95h's fold.

## Four complete tree combinations

The bare main control and all four candidates run fresh full applies with
--build-product. No fold branch is used as the shipping branch's base. Only
32's and 95h's adaptation directories are overlaid in isolated test snapshots.
The exact fold inputs are:

- 32: 997331ef1ec86480f47f3710fd43e1150407dafc.
- Completed 95h: c2714b29c5937f553d79e98e9fae80aab8ed995a.

Labels x, 32x, 95hx and both mean main plus 99x, with the corresponding earlier
folds. The current 95h fold rewrites the two checker sites; its four program
sites stay present under their owner pins. Thus both has four originals left,
which is an accepted all-present program group.

| Candidate | Total rewrites against main | Handed-off originals remaining | Types diagnostics, 79 roots | Nine JavaScript files | Imports removed against main |
| --- | ---: | ---: | ---: | --- | --- |
MATRIX_ROWS

The bare control also passes 79-root types. Each candidate's own 99x count is
28; the larger total includes the earlier folds. Every complete source file
is compared against exactly those substitutions and authorized unused import
removals, preserving each input's CRLF convention. Replaying 99x over each
candidate writes zero files and preserves all 79 source byte strings and
nanosecond mtimes. Per-candidate summaries are composition/matrix/*.json;
full commands, types listings, build logs, source checks, declaration hashes
and all nine raw JavaScript diff files are in the matching .tar.gz archives.
All three measured declaration products are byte-identical.

Seven JavaScript files are byte-identical in every comparison. Each compiler
bundle has precisely the expected site count. The complete normalized ASTs
agree after reconstructing only inventoried bindings and ordinary generated
temporary bookkeeping. No helper deletion or unrelated source edit is waived.
The two 95h combinations additionally use the exact symbol/observation proof
from 95h's rename-proof.cjs, copied as composition-renames.cjs with only the
owner lookup adapted. It proves the binder's maybeBind2 to maybeBind3 rename
at its declaration and two direct calls, and the checker getter arrow's _a to
_a2 temporary at its declaration, assignment and call receiver. Every use
resolves to the unique binding, has the same corresponding context, and has
no escape, alias or .name observation. All nine files are scanned for string,
template and property-name observations. The checker arrow differs only by
its proved rename. No other non-temporary binding rename is exempted. Generated temporary
renumbering is matched by its checker symbols and complete normalized uses. Complete
location/use maps are javascript/rename-map.json inside the 95hx and both
matrix archives. Their raw hunks remain available alongside the normalized
comparison, including all normal optional-chain temporary changes.

The production type checker and latent selector are built from this branch's
unchanged main 322f8056 Go sources. Exact type-check commands, root lists and
binary hashes are retained in each matrix archive. No unrelated compiler
change from a fold branch enters these measurements.

## Whole lanes and all 796 shard runs

Both x and 32x use the unchanged standard artifact preparation, whole lane,
shard manifest, runners, timeouts, expected counts and sanctioned API lines.
Each prepared artifact's complete 746-file source map and all nine emitted
JavaScript products equal the final checked tree byte for byte. The final
call counter also recognizes parenthesized callees; all four full applies,
types and emitted comparisons were rerun after that validation change.
composition/oracle-final-source-identity.json proves the exact source and
product identity, so the recorded oracle runs test the final products.
The initial matrix is retained in initial-matrix.tar.gz. Final full-apply
snapshots' five functional inputs match this branch byte for byte; the earlier
preparation snapshots remain identified by preparation.json and provenance.json.

Each whole lane passes with 106,366 passing, one sanctioned public API failure,
and zero pending. All 398 standard shards pass for each composition. The
independent union checker requires exactly the same 106,367 test identities
and statuses as that composition's whole lane, and exactly equal sanctioned
API diff bytes. Main's unchanged check_results passes too. The union receipts
are composition/oracle/x/union.json and composition/oracle/32x/union.json.

Preparation and whole runs have a 1440-second outer bound. Shards run in ten
pieces per composition, at most 40 standard shards per piece, with two shards
concurrent and unchanged per-shard limits. Each completed piece is committed
and plainly pushed immediately. All preparation, whole and shard pieces
complete in less than 25 minutes:

| Piece | Seconds |
| --- | ---: |
TIMING_ROWS

Each whole run and shard piece has a small unpacked summary plus packed raw
reports, logs, observations and exact API diff. Per-shard result files are
inside piece-N.tar.gz, never expanded in this branch. Standard product-path
shards use symlink-only views of the same prepared artifact and exact locked
dependencies. Successful redundant scratch views are removed after recording
results. No test expectation or runner is edited.

## Controls and CW051

composition/contract.log.gz records all original strict malformed-owner,
missing-import, partial-tree, unrelated-edit, CRLF and byte/mtime replay
controls. composition/composition-mutant.json and its dedicated compressed
log record five partial handed-off owner groups, retained unused imports in
builder.ts, watch.ts and program.ts, and a complete-fold success control.
Every rejection verifies zero source writes and unchanged mtimes. The
complete-fold control removes program/watch imports, preserves checker bytes,
accepts an opaque foreign-owner replacement, and replays with zero writes.
An additional complete-fold control keeps program.ts's import for a
parenthesized core helper call and verifies a zero-write replay.

The six CW051 units are measured freshly on main and main plus 32 plus 99x.
The scheduling-only latent selector uses main's existing measurement harness;
full/speculative modes stay disabled. Every piece selects exactly one unit,
reports checker_rejected false and completes below 25 minutes. Raw file-wide
refusal scans stay separate from the selected unit's lowering stop. Binary
hashes, exact references, current unit positions, reasons and timers are in
composition/latent/. All stops below are NotYet:

| Unit | Main before | Main plus 32 plus 99x |
| --- | --- | --- |
CW_ROWS

The checker unit remains at its method-presence boundary. The five host units
move from core's optional call to the direct optional-call form Apple's head
lowers. No native-lowering completion is claimed by this measurement.

## Packed continuation evidence

Every new raw matrix, latent, whole-lane and shard-piece result is packed.
For NAME.tar.gz, NAME.sha256 lists every member's original SHA-256 and its
path relative to the archive root. NAME.tar.gz.sha256 records the archive's
own hash. Extract into a fresh directory and run sha256sum -c against the
member list to validate original bytes. Summaries and union receipts remain
unpacked. Links into a raw result refer to members of its corresponding
archive. composition/superseded.tar.gz retains the interrupted first attempt,
superseded by the complete opaque-owner rule before any acceptance claim.
No stopped or superseded run is counted as a pass.

The historical proof below documents the original squashed delivery only.

---

'''
text=text.replace('MATRIX_ROWS',rows).replace('CW_ROWS',cw).replace('TIMING_ROWS','\n'.join('| '+label+' | '+f'{seconds:.3f}'+' |' for label,seconds in timings))
assert '\u2013' not in text and '\u2014' not in text
proof=p/'PROOF.md';old=proof.read_text();marker='# 99x proof: 28 rewrites, 18 earlier-owner handoffs';assert marker in old;proof.write_text(text+old[old.index(marker):])
f=p/'README.md';old=f.read_text();old=old.replace('Full apply, type, emitted\nJavaScript, oracle and CW051 composition measurements are in progress.', 'All four composition applies, 79-root type checks and nine-file JavaScript\ncomparisons pass. Both required whole lanes and all 398 shards for each pass,\nwith exact test-identity/status and API-byte unions. All six CW051 units are\nmeasured before and after main plus 32 plus 99x. See PROOF.md for the table\nand composition/ for packed current receipts.')
f.write_text(old)
out=dict(status='pass',base='a6d5f494e46262f805e5be76a4d7c108fe6a3a6c',main='322f80565edaa978d928ca559803c578ed67c105',own_rewrites=28,owner_pins_unchanged=True,matrix={k:dict(expected_rewrites=v['source']['expectedSites'],remaining_originals=v['source']['replay']['left'],types=v['types'],javascript_pass=v['javascript']['pass'],imports_removed=v['source']['importsRemoved']) for k,v in matrix.items()},oracles=oracles,CW051=latent['comparison'],contracts='pass')
(p/'outcome.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps(dict(status='pass',proof_bytes=len(text),matrix=list(matrix),oracle_shards=796)))
