#!/usr/bin/env python3
"""Verify step 32 evidence without rerunning compiler packages."""
import gzip,json,re,sys
from pathlib import Path
here=Path(__file__).resolve().parent
out=Path(sys.argv[1]) if len(sys.argv)>1 else here/'evidence/step32'
def read(name):
    return gzip.decompress((out/(name+'.gz')).read_bytes()).decode()
data=json.loads((out/'result.json').read_text())
assert data['compiler_commit']=='88fe8de47cba6928539f1b35d4fd84c1ad06f34d'
assert data['skipped_merge']=='8f32e51e8fc41b8f1177453213ca5453ce764486'
assert data['combined_mode_measured'] is False
assert len((out/'conflicts.txt').read_text().splitlines())==7
assert 'CONFLICT' in (out/'merge.txt').read_text()
assert 'vcs.modified=false' in read('binary.txt')
assert data['compiler_commit'] in read('binary.txt')
for profile in ('reference','current'):
    assert read(profile+'-0.stderr')==read(profile+'-1.stderr')
    assert read(profile+'-0.stdout')==read(profile+'-1.stdout')
    assert data[profile+'_exits']==[1,1]
    assert data[profile+'_first_stop']==read(profile+'-0.stderr').splitlines()[0]
    assert ':1246:69: error TS2345:' in data[profile+'_first_stop']
assert read('node-entry.stdout')=='Version 6.0.3\n'
assert read('node-entry.stderr')==''
assert read('node-entry.exit')=='0\n'
baseline=json.loads((here/'evidence/main-efe9f404/stops.json').read_text())
comparison=json.loads((out/'comparison.json').read_text())
assert len(comparison['stops'])==15
for old,row in zip(baseline,comparison['stops']):
    assert row['ordinal']==old['ordinal']
    raw=read(f"snapshot-{old['ordinal']:02d}-snapshot-types.stderr")
    pattern=re.escape('/'+old['file'])+f":{old['line']}:{old['column']}: (.*)"
    actual=[m.group(1) for line in raw.splitlines() if (m:=re.search(pattern,line))]
    assert row['snapshot_diagnostics']==actual
    assert row['node_stdout']==read(f"snapshot-{old['ordinal']:02d}-probe-node.stdout")
    status='remain' if old['message'] in row['snapshot_diagnostics'] else ('changed' if row['snapshot_diagnostics'] else 'disappear')
    assert row['status']==status
    assert row['node_exit']==0 and row['node_stdout']==old['node_stdout']
    assert row['placeholder_induced']==(old['ordinal'] in (9,14))
walk=json.loads((out/'walk.json').read_text())
assert 1<=len(walk)<=16
for row in walk:
    n=row['ordinal']
    assert read(f'walk-{n:02d}-0.stderr')==read(f'walk-{n:02d}-1.stderr')
    assert row['split_0_exit']==row['split_1_exit']==1
    assert row['message'] in read(f'walk-{n:02d}-0.stderr').splitlines()[0]
    assert row['owner'] in ('compiler','runtime','library','adaptation')
    assert row['probe_types_exit']==row['probe_build_exit']==1
    assert row['probe_first_stop'] is not None
    assert row['node_exit']==0 and row['node_stderr']==''
    assert row['node_stdout']==row['expected_stdout']
    assert row['mutant_exit']==0 and row['mutant_stdout']!=row['expected_stdout']
print(f"15 saved sites verified; {len(walk)} walked stops; Node witnesses and stdout mutants verified; conflicting mode excluded")
