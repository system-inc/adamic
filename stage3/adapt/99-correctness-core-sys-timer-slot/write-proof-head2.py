from pathlib import Path
import json
R=Path('/workspace/adamic');D=R/'stage3/adapt/99-correctness-core-sys-timer-slot';d=json.loads((D/'census-comparison.json').read_text());results=json.loads((D/'proof-results.json').read_text());lane=json.loads((D/'lane-results.json').read_text())
text='''# C-noinfer-b implementation proof

Head 1 implementation: `91af96c4157e4437d5c327d15fa20244d3ed3f26`.
Its requested replay-note follow-up: `5ba70a948`.
Pinned main: `0cb816a0ec439f63542b3f83bb13941fd79f8f41`.

| Proof | Result |
|---|---|
| Full applies | Head 1 and Head 2 exit 0 |
| Paired replay | Every existing adapter's incremental patch is identical; only the new directory adds four type edits |
| Source scope | All 746 source files checked; only sys.ts differs, exactly the four ruled annotations |
| Owner contract | Whole-owner pins, LF/CRLF, no-write replay and rejection mutants pass |
| Adamic native types | 79 compiler roots together, exit 0 in both heads |
| Stock TypeScript 6.0.3 | Adamic options and upstream solution: 0 diagnostics in both heads, 0 added |
| Own compiler and harness build | src, services, server, testRunner, harness and unittests compile, both heads exit 0 |
| JavaScript | All nine emitted compiler files compare byte for byte with Head 1 and original main |
| API generation | generate-api.cjs exits 0; shard plan runtime refreshed; generate_shards.py and --check exit 0 |
| Census | Exactly the expanded expected set; nothing else new |
| Mutant | Restoring only the two any slot signatures restores main 384:79 |

'''
text+=f"The whole lane executes all {results['laneUnits']} manifest units in pieces, with no inventory filter: **106,366 pass, one sanctioned public API failure, zero skip**. Every final unit observation has command exit 0 after checking its unchanged expected upstream exit, complete test identity multiset, counts and baseline paths. The maximum unit wall is {lane['maximum']:.3f} seconds. The API unit's upstream exit is 1 as expected; its exact normalized declaration diff equals sanctioned-api.json. The two new sanctioned members are System.setTimeout and System.clearTimeout; pre-existing API sanctions are retained. No expected.json or shard assignment moves.\n\n"
text+='''## Exact declaration changes

head2-typescript.d.ts.diff.txt contains only the two System members at built
8545-8546. head2-typescript.internal.d.ts.diff.txt contains those same two
members at internal 16663-16664; the internal file does move. The isolated
public baseline delta is public-baseline-slot.diff.txt at 8491-8492. This
adapter leaves the reference file unchanged; the API failure compares those
old reference members with the new local members at 8545-8546. The isolated
diff shows only this ruling, while the complete lane diff includes existing
sanctions too. declaration-spans.json uses the stock AST to verify each
System member and coordinate in both trees. The lane
retains its actual reference/local snapshots and complete baseline diff in
evidence.tar.gz. tsserverlibrary.d.ts does not move. Generic NoInfer is a
public reader requirement: TypeScript 5.4 or later.

## Full sys.ts census

The standard guarded latent tool loads all 79 roots with LATENT_FULL=1,
LATENT_ONLY_FILE set to the exact adapted sys.ts, and LATENT_ASSERT_NO_OUTPUT=1.
The current-main compiler and overlay provenance are in Head 1's proof.
There are zero stock checker diagnostics. Distinct walls are keyed by
main-mapped file:line:column, kind, exact reason and phase. Boundary events
are not walls. No replay or speculative lowering flags are used.

Main has 123 walls, Head 1 has 120, and Head 2 has 123. Against Head 1,
115 are kept, five gone, eight new. Against original main, nine are new,
including Head 1's expected #995spev `reading setTimeout`. The actual full
apply's census equals the preliminary candidate's complete ledger exactly.
census-comparison.json and census-kept.json retain every kept, removed and
new site. The initializer masking reversal is retained in masking-witness.json
and explained, with its inference boundary, in README.md.

| New against Head 1, main coordinates | Actual coordinates | Kind / phase | Exact reason |
|---|---|---|---|
'''
for x in d['head2']['new']:
 loc=x['where'].split(':');actual=':'.join([loc[0],str(int(loc[1])-2),loc[2]]);text+=f"| {x['where']} | {actual} | {x['kind']} / {x['phase']} | {x['reason'].replace('|','&#124;')} |\n"
text+='\n| Gone against Head 1 | Kind / phase | Exact reason |\n|---|---|---|\n'
for x in d['head2']['gone']:text+=f"| {x['where']} | {x['kind']} / {x['phase']} | {x['reason']} |\n"
text+='''
The capture message is constructed at internal/lower/locals.go:208; the rest
call message at expression.go:1704; provider method-variance refusals at
invariance.go:1042; the initializer any relation at any_relations.go:43.
These are named remaining compiler walls, not regressions hidden by this proof.

## Evidence and operational limits

proof-results.json records actual child exits and durations; artifacts.json
records all nine hashes and declaration hashes; source-scope.json records all
746 source hashes. paired-replay.json records every paired git diff exit and
incremental patch hash. evidence.tar.gz retains raw census streams, timings,
logs, exact sys.ts inputs, lane reports and observations. evidence.sha256 pins
that archive. The stock checker is the same 6.0.3 compiler-API driver used in
the accepted measurement; caller count provenance is that measurement at
`949b9c3c4c883d964bba19ee9fbdae9fd1909288`, not a new text-grep count.
Its storage-pin survey is retained separately: no direct storage or owner pins
were found in the surveyed executable rules. No existing adaptation was
weakened. The full apply and replay comparison check the current complete chain.

A proof-authoring trial put its drift whitespace outside the pinned owner,
which the adapter correctly accepts. The contract was corrected to insert a
comment inside the owner, without changing the adapter. The final no-write
contract passes. proof-authoring-retries.json records that initial child exit
and the environment retry below.

The first shard-generator invocation lacked the sourced Go PATH and exited 1;
it was rerun after sourcing /workspace/adamic-tools/env.sh and both generation
and check exited 0. This was an environment retry, not a changed generator.
The inherited finished-tree adaptation-41 replay still has the same rule-31
stop on main and Head 1; its isolated own-stage replay passes. Head 1's README
now explicitly attributes that pre-existing limitation.

The first four-way lane attempt hit the unchanged 30-second limit on three
async-function units and was stopped. The first unit passed alone in 11 seconds.
The remaining manifest was covered in bounded two-way pieces. The other two
initial timeouts passed in those pieces; no two-way timeout needed a standalone
retry. All attempts are retained with the same limits.
Final per-unit observations pass; interrupted and timeout attempts remain in
the receipts. No timing expectation was increased. Inference: the concurrent timeouts reflect
resource contention; the standalone retry and unchanged JavaScript support
that explanation, but no paired four-way main timing control was run.

All proof pieces are bounded below 30 minutes and children are reaped, with
process groups cleaned up. A separate checkpoint ref was pushed after the
full apply. Neither held head changes compiler lowering or Runtime. No main
push or force push is used.
'''
assert '\u2013' not in text and '\u2014' not in text
(D/'PROOF.md').write_text(text)
