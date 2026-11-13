"""Audit the complete proof and render its report; refuse an incomplete lane."""
from pathlib import Path
from collections import Counter
import hashlib,json,subprocess
p=Path(__file__).parent
repo=Path('/workspace/timer-ambient-delete')
base='f37253b347c4f7e7bc223ed05fe04edb8cef0cd4'
head='2cfa44d66203f6ed8f59910371fbd519dd4df27d'
assert subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],text=True).strip()==head
assert subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD^'],text=True).strip()==base
assert not subprocess.check_output(['git','-C',str(repo),'status','--porcelain'],text=True)
inputs=json.loads((p/'adaptation-inputs.json').read_text())
assert inputs=={str(f.relative_to(repo)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((repo/'stage3/adapt').rglob('*')) if f.is_file() and '__pycache__' not in f.parts}
zero=['full-apply','full-apply-replay','contract-41-final','stage-replays','stock-road5-79','adamic-road5-79','stock-final-79','adamic-final-79','build-after','node-witness-before','node-witness-after','census-before-exact','census-after','census-delta','source-scope']
for name in zero:assert json.loads((p/(name+'.status.json')).read_text())['child_exit']==0,name
for name in ['mutant-property-slot','mutant-restored-concrete-pair','binder-first-stop-final','binder-cli-first-stop']:assert json.loads((p/(name+'.status.json')).read_text())['child_exit']==1,name
assert (p/'node-witness-before.stdout').read_bytes()==(p/'node-witness-after.stdout').read_bytes()
assert '41 timer-head check: ambient timer declaration restored' in (p/'mutant-restored-concrete-pair.stderr').read_text()
property=json.loads((p/'mutant-property-slot.json').read_text())
assert len(property['diagnostics'])==2 and all(d['code']==2322 for d in property['diagnostics'])
ledger=json.loads(Path('/workspace/scratch/timer-road5-lane/ledger.json').read_text())
manifest=json.loads((repo/'stage3/lane/shards.json').read_text())
assert ledger['all_commands_executed'] and len(ledger['units'])==len(manifest['units'])==398
expected={u['name']:u['expect'] for u in manifest['units']}
counts=Counter()
for unit in ledger['units']:
 assert unit['exit']==0 and unit['status']=='pass',unit
 assert unit['counts']==expected[unit['name']],unit
 assert unit['seconds']<30,unit
 counts.update(unit['counts'])
assert dict(counts)=={'pass':106366,'fail':1,'skip':0},counts
(p/'lane-ledger.json').write_text(json.dumps(ledger,indent=2)+'\n')
delta=json.loads((p/'census-delta.json').read_text());assert (delta['before'],delta['after'],delta['retained'])==(236,230,228)
first=json.loads((p/'binder-first-stop-final.stdout').read_text());assert first['checked_roots']==79 and first['entry_roots']==1 and not first['has_ir']
assert first['failure_type']=='*lower.Refused' and 'sys.ts:142:81:' in first['first_stop']
summary={'base':base,'head':head,'all_required_proofs_complete':True,'lane_shards':398,'lane_counts':dict(counts),'lane_maximum_seconds':ledger['maximum'],'census':{k:delta[k] for k in ['before','after','retained']},'gone_records':len(delta['gone']),'new_records':len(delta['new']),'binder_first_stop':first}
(p/'SUMMARY.json').write_text(json.dumps(summary,indent=2)+'\n')
report=f'''# Ambient timer deletion, road 5

Adamic task 995spev, ruled final Nov 13 2026 at 07:51:37Z. Language's road and
named slot premise, plus Adamic's real Node cancellation witness, are implemented.

Base: `{base}`. Change: `{head}`.
The change is one commit on the base, pushed without force to
`typescript/timer-ambient-delete-2`. Evidence is kept only on
`typescript/timer-ambient-delete-evidence-2`, in this directory.

## Change and premises

41 deletes the two complete ambient declaration lines as an atomic group.
Both absent is idempotent; half-applied, drifted, duplicate and moved owners
refuse before writing. Its own timer-head check also rejects restored heads.
93's two getNodeSystem cancellation forwarders take
`Parameters<typeof clearTimeout>[0]`. Its after-owner pin and both crypto
getNodeSystem pins are updated. The scheduling forwarders and the `unknown`
System method slot are unchanged. No cast, new ambient declaration or host
provider widening is introduced.

Stock TypeScript sees global clearTimeout's parameter as
`NodeJS.Timeout | string | number | undefined`; Adamic's existing pinned provider
serves `unknown`. Both accept the parameter annotation. Stock acceptance at the
System slot rests on method-parameter bivariance; Adamic's relation is exact.
The property-slot mutant fails loudly with TS2322 at both forwarders. This
premise is recorded in both the timer README and upstream ledger. The README
also carries the requested dynamic-this and callback-length premises.

## Acceptance

| Check | Result |
|---|---|
| Gate first: actual adamic types, 79 roots | Exit 0, no diagnostics |
| Gate first: stock TypeScript 6.0.3, 79 roots | Exit 0, no diagnostics |
| Actual full-apply source: both 79-root checks | Both exit 0, no diagnostics |
| Fresh full apply | Exit 0, 360.581 seconds |
| Keyed full-apply replay | Exit 0, all tracked source bytes identical |
| Own-stage replays: 41, 93, crypto, timer closures | Zero edits; all source bytes and mtimes unchanged |
| Stock ordinary compiler and harness build | Exit 0 |
| Adapted lane | 398/398 shards pass, in bounded pieces |
| Lane counts | 106,366 passing, one expected API failure, zero pending |
| Restored exact concrete ambient pair | Exit 1 at 41's own timer-head check |
| Property-style System slot | Exit 1; exactly two stock TS2322 diagnostics |
| Node cancellation witness, before/after | Both exit 0; stdout byte-identical |
| Full-mode sys.ts census pair | Both exit 0, zero checker diagnostics |

The largest successful shard took {ledger['maximum']:.3f} seconds. Raw attempts,
per-shard reports, observations and baselines are preserved in lane-raw.tar.gz;
the ledger retains setup failures and any retries. Each piece and compiler
command is bounded to 1700 seconds and reports its child's actual exit.

## Node witness

node-cancel-witness.cjs loads each built compiler's real `sys.clearTimeout`,
which is getNodeSystem's forwarder. A real Node handle is cancelled and its
callback never fires. The string `not-a-timer-id`, a plain object and undefined
each return undefined without throwing. This is measured through the actual
built host, rather than an extracted expression. Before and after stdout is
identical and preserved in the two witness receipts.

## Census and next ordinary stop

The exact-CRLF f37253b3 control has 236 records: 35 Refused, 89 NotYet and
112 Boundary. The deletion head has 230: 33 Refused, 87 NotYet and 110 Boundary.
There are 228 retained records, eight gone and two new. The two new records are
both at `src/compiler/sys.ts:1520:84`, the global call in the expression-bodied
cancellation arrow, with lowering reason:

```
node:timers.clearTimeout used as a value; call it as a statement
```

census-delta.json retains every paired/gone/new record, including multiplicity,
raw coordinates, named unit scope and named finding scope. Boundaries are not
compiler stops. These are measurement-only lowering observations, not native
compilation successes. Both runs use LATENT_FULL=1, exact LATENT_ONLY_FILE and
LATENT_ASSERT_NO_OUTPUT=1 with no speculative or replay flags. The standard
no-output overlay and its commands are archived; physical Go evidence uses
.go.txt names.

After checking all 79 roots, normal one-entry lowering from binder.ts exits 1
with no IR. The production CLI independently reports the same first stop:

```
{first['first_stop'].replace('/workspace/scratch/timer-road5-after/','')}
```

This establishes the next stop for binder.ts; it does not assert that all 73
ordinary roots share that next stop.

## Source identity, commands and retained setup attempts

source-scope.json proves that sys.ts is the sole changed source file: exactly
the two ambient-line deletions and the two parameter annotation replacements.
The final apply preserves CRLF. The initial exploratory gate used normalized
sys.ts line endings; its text equals the actual applied source, and both
checkers were then rerun successfully on the actual final tree. The definitive
before census is census-before-exact.jsonl, restored byte-for-byte from the
base's recorded final stage. Source archives and SHA-256 pairs are preserved.

run-command.py records each command, working directory, elapsed time and exact
child exit. Failed setup receipts are retained separately: the first 41 proof
hit Node's default source-output buffer limit, the first lane piece found the
build-only node_modules link where the lane product expects it absent, and the
first first-stop harness passed 79 entries to the one-entry lowering API.
These were corrected in the definitive suffixed receipts without changing the
adaptation or compiler. Before the exact-byte census, the earlier normalized
control run is retained too. No evidence file is placed under stage3/adapt.
'''
(p/'README.md').write_text(report)
print(json.dumps(summary,indent=2))
