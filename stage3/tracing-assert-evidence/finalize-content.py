from pathlib import Path
import json
r=Path('/workspace/scratch/baobab-evidence');repo=Path('/workspace/adamic');lane=json.loads((r/'lane-summary.json').read_text());a=repo/'stage3/adapt/99-correctness-tracing-assert'
text=(r/'PROOF.draft.md').read_text()
counts=lane['oracleCountsCheckedByEachGoUnit'];controls=lane['controls']
section=f'''prepare_shards.py succeeds. All {lane['units']} real TestShard_ Go units pass
with exit 0. Each Go unit checks its observed oracle counts against its manifest
expectation: {counts['pass']} passing, {counts['fail']} expected failing and
{counts['skip']} skipped tests in total. The counter preload records
{lane['processRecords']} completed Node processes, and every manifest unit has
counter-loaded process coverage.

stopTracing's first assert executes **{lane['assertCalls']} times** in the entire
lane. Boolean(tracing) versus !!tracing differences: **{lane['differences']}**;
assertion-outcome differences: **{lane['differences']}**. {'The oracle never calls stopTracing. This is a zero count, not a coverage claim for the changed assertion. The separate positive controls supply direct execution evidence.' if lane['oracleNeverCallsStopTracing'] else 'Every recorded call has equal conversions and equal assertion outcomes.'}

The counted tree is a temporary copy of the prepared artifact. The source
adaptation is never instrumented. The prepared artifact's 747 source files and
10 normal JavaScript files match the validated candidate byte for byte. Only
private emitted bundle copies receive a counter before the first assert.
The VM-aware preload persists zero rows as well as actual call rows. A
measurement-only Go overlay selects that copied product and carries the preload
through the gate's filtered tool environment; production Go files remain intact.

The first attempt is invalid: all 398 units failed before running their oracle
because the shared artifact's node_modules collided with the product-path
link. It is retained separately as attempt1 and excluded from all counts.
Removing that directory only from the temporary copy restores the gate's product
layout. The successful full rerun uses parallel=2, without a test filter beyond
all TestShard_ units. No failed subset is spliced into the successful run.

Prepared artifact key: {lane['preparedArtifactKey']}.
The lane ran against frozen inputs before this final README and proof summary
were added. Adapter code, pins and all runtime source bytes remain unchanged.
'''
text=text.replace('LANE_PENDING',section).replace('EVIDENCE_LINK','Raw evidence, scripts, counter rows and complete logs: [typescript/tracing-assert-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-assert-evidence/stage3/tracing-assert-evidence).')
assert 'PENDING' not in text
(a/'PROOF.md').write_text(text);(a/'README.md').write_text((r/'README.final.md').read_text())
p=repo/'stage3/upstream/LEDGER.md';t=p.read_text();old='  and all other stopTracing bytes are preserved. Verification is pending.';assert t.count(old)==1
t=t.replace(old,'  and all other stopTracing bytes are preserved. Both 79-root type checks and\n  the complete sharded lane pass; the lane records '+str(lane['assertCalls'])+' assertion calls and '+str(lane['differences'])+' outcome\n  differences, with positive Node controls. See the adaptation PROOF.md.')
p.write_text(t)
print('final README, proof and ledger prepared')
