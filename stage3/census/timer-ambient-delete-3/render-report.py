"""Audit the requested version-3 proofs and refuse an incomplete or over-budget lane."""
from pathlib import Path
from collections import Counter
import hashlib,json,subprocess
p=Path(__file__).parent;repo=Path('/workspace/timer-ambient-delete')
base='f37253b347c4f7e7bc223ed05fe04edb8cef0cd4';head='caf3b27ec873945f4c210cc20a3c2be7e6437cdd'
assert subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],text=True).strip()==head
assert subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD^'],text=True).strip()==base
assert not subprocess.check_output(['git','-C',str(repo),'status','--porcelain'],text=True)
assert json.loads((p/'adaptation-inputs.json').read_text())=={str(f.relative_to(repo)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((repo/'stage3/adapt').rglob('*')) if f.is_file() and '__pycache__' not in f.parts}
zero=['full-apply','own-stage-replays','stock-79','adamic-79','build-after','node-witness-before','node-witness-after','census-before','census-after','census-delta','census-block-fix-delta','source-scope','lane-all-pieces']
for name in zero:assert json.loads((p/(name+'.status.json')).read_text())['child_exit']==0,name
for name in ['property-slot-mutant','binder-first-stop']:assert json.loads((p/(name+'.status.json')).read_text())['child_exit']==1,name
assert (p/'node-witness-before.stdout').read_bytes()==(p/'node-witness-after.stdout').read_bytes()
stock=json.loads((p/'stock-79.json').read_text());assert stock['roots']==79 and not stock['diagnostics']
mutant=json.loads((p/'property-slot-mutant.json').read_text());assert len(mutant['diagnostics'])==2 and all(d['code']==2322 for d in mutant['diagnostics'])
delta=json.loads((p/'census-delta.json').read_text());fix=json.loads((p/'census-block-fix-delta.json').read_text());assert not delta['new'] and not fix['new']
assert (delta['before'],delta['after'],delta['retained'])==(236,226,226)
value=[r for r in fix['gone'] if r['kind']=='NotYet'];boundary=[r for r in fix['gone'] if r['kind']=='Boundary']
assert len(value)==2 and all(r['where']=='src/compiler/sys.ts:1520:84' and r['reason']=='node:timers.clearTimeout used as a value; call it as a statement' for r in value)
assert len(boundary)==2 and all(r['reason']=='failed statement; state rolled back; continued at next statement or declaration' for r in boundary)
ledger=json.loads(Path('/workspace/scratch/timer-ambient-delete-3-lane/ledger.json').read_text());manifest=json.loads((repo/'stage3/lane/shards.json').read_text());expected={u['name']:u['expect'] for u in manifest['units']}
assert ledger['all_commands_executed'] and len(ledger['units'])==len(expected)==398
counts=Counter()
for unit in ledger['units']:
 assert unit['exit']==0 and unit['status']=='pass' and unit['seconds']<30,unit
 assert unit['counts']==expected[unit['name']],unit
 counts.update(unit['counts'])
assert dict(counts)=={'pass':106366,'fail':1,'skip':0},counts
lane_total=json.loads((p/'lane-total.json').read_text());assert lane_total['seconds']<1700
(p/'lane-ledger.json').write_text(json.dumps(ledger,indent=2)+'\n')
first=json.loads((p/'binder-first-stop.stdout').read_text());assert first['checked_roots']==79 and first['entry_roots']==1 and not first['has_ir'] and first['failure_type']=='*lower.Refused'
assert 'sys.ts:142:81:' in first['first_stop'] and '(adamic/invariant-mutable)' in first['first_stop']
summary={'base':base,'head':head,'all_requested_proofs_complete':True,'lane_shards':398,'lane_counts':dict(counts),'lane_total_seconds':lane_total['seconds'],'lane_maximum_successful_shard_seconds':ledger['maximum'],'census_before':236,'census_after':226,'census_gone':10,'census_new':0,'block_fix_value_use_records_removed':2,'block_fix_rollback_boundary_records_removed':2,'binder_first_stop':first}
(p/'SUMMARY.json').write_text(json.dumps(summary,indent=2)+'\n')
apply=json.loads((p/'full-apply.status.json').read_text())
report=f'''# Ambient timer deletion 3: cancellation as a statement

Base: `{base}`. Change: `{head}`.
One commit on f37253b3, pushed without force to
`typescript/timer-ambient-delete-3`. Evidence belongs only to
`typescript/timer-ambient-delete-evidence-3`, in this directory.

The change carries road 5's paired deletion through 41, with 93's cancellation
forwarders changed to:

```ts
clearTimeout: (timeoutId: Parameters<typeof clearTimeout>[0]): void => {{ clearTimeout(timeoutId); }}
```

setTimeout's forwarder still returns the handle. The System slot remains an
unknown-parameter method. 93's after-owner pin and both crypto getNodeSystem
pins match the new text. The README and ledger retain the method-bivariance
premise and point to this evidence. The callback-this and callback-length
README premises are carried unchanged. No cast or host-provider change is added.

## Requested reruns

| Check | Result |
|---|---|
| Fresh full apply | Exit 0, {apply['seconds']:.3f} seconds |
| Own-stage replay: 41, 93, crypto, timer closures | Zero edits; all source bytes and mtimes unchanged |
| Actual adamic types, all 79 roots | Exit 0, no diagnostics |
| Stock TypeScript 6.0.3, all 79 roots | Exit 0, no diagnostics |
| Ordinary compiler and harness build | Exit 0 |
| Property-style System slot mutant | Exit 1; exactly two TS2322 diagnostics |
| Node cancellation witness, before and after | Both exit 0; stdout byte-identical |
| Full-mode sys.ts census pair against f37253b3 | Both exit 0; zero checker diagnostics; no new records |
| Binder ordinary first stop after the 79-root check | Exit 1, Refused at sys.ts:142:81; no IR |
| Named lane manifest, conservative superset | 398/398 shards pass within the total lane budget |

The property mutant changes only System's slot to
`clearTimeout?: (timeoutId: unknown) => void` and retains both new block
forwarders. Stock reports TS2322 at sys.ts:1520:13 and sys.ts:1618:13. Its exact
messages and source SHA-256 pair are preserved. This verifies the unchanged
method-parameter bivariance premise under stock; Adamic's existing global
provider accepts unknown, so its relation remains exact.

The Node witness loads the actual built compiler's sys.clearTimeout, rather
than extracting an arrow. A real handle cancels; the string `not-a-timer-id`,
a plain object and undefined each return undefined without throwing. Before
and after stdout is identical.

## Census: both named records gone, nothing new

Fresh full-mode measurements use the exact f37253b3 source and this head,
loading all 79 roots with LATENT_FULL=1, exact LATENT_ONLY_FILE and
LATENT_ASSERT_NO_OUTPUT=1. No speculative or replay flags are enabled. The
measurement-only overlay and its no-output guards are unchanged production
census tooling, pinned by the recorded toolchain and archived overlay.

Against f37253b3, records go **236 to 226**: 226 retained, ten gone and **zero
new**. Before has 35 Refused, 89 NotYet and 112 Boundary; after has 33 Refused,
85 NotYet and 108 Boundary. Every gone/retained record is preserved in
census-delta.json with raw locations, named scope and multiplicity.

A second comparison uses the exact prior expression-body census at 2cfa44d66.
That goes **230 to 226**, with exactly four gone and zero new: the two NotYet
records at sys.ts:1520:84 with reason
`node:timers.clearTimeout used as a value; call it as a statement`, and their
two rollback Boundary records at sys.ts:1537:9. Thus both requested value-use
records are absent and the block body introduces nothing new. Boundaries are
measurement events, not compiler stops. The full proof makes no claim that
tsc lowers to native code.

Binder's normal one-entry lowering, following a successful 79-root check,
still stops at:

```
{first['first_stop'].replace('/workspace/scratch/timer-ambient-delete-3-after/','')}
```

This measures binder.ts only, without claiming that all 73 ordinary roots
share the same next stop.

## Lane scope and budget

sys-suite-inventory.json lists the 16 explicitly named sys/symlinkWatching
shards from the pinned plan. lane-selection.json names all 398 manifest shards
as a conservative superset of shards that read the compiler built from sys.ts. The prior complete lane
measured 696.538 seconds, allowing this full rerun inside the requested
30-minute total. This run used ten separately receipted 40-shard pieces (the
last has 38), with four workers and an overall 1700-second lane bound.

Actual lane total: **{lane_total['seconds']:.3f} seconds**. The largest successful
shard took {ledger['maximum']:.3f} seconds. Counts: 106,366 passing, the one
expected API failure, zero pending. Per-shard identities, counts and baseline
checks passed. Raw logs, reports and observations are in lane-raw.tar.gz with
its SHA-256 receipt; lane-ledger.json retains every attempt and any retry.

## Source and reproduction

source-scope.json verifies every source file against the base's recorded final
apply stage. Only sys.ts differs, by exactly the paired ambient-line deletion
and the two block-bodied cancellation forwarders. CRLF is preserved. The before
and after sys.ts bytes and SHA-256 pair are retained. Adaptation input hashes
and checker/census binary hashes are in adaptation-inputs.json and toolchain.json.
Production Go/compiler code is unchanged, so the pinned binaries used for the
previous road-5 measurements are reused with the new source tree. Their probe
source and overlay are preserved; physical Go evidence is named .go.txt.

run-command.py propagates and records its child's actual exit, command, working
directory and elapsed time, with a 1700-second process bound. Commands and raw
stdout/stderr for every requested rerun are retained here. The property mutant
and binder stop intentionally return 1. The census count includes the two
additional rollback Boundaries removed by this fix; they are not new failures.
No evidence is stored under stage3/adapt.
'''
(p/'README.md').write_text(report)
print(json.dumps(summary,indent=2))
