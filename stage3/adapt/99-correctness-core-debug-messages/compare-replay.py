from pathlib import Path
import json,difflib,hashlib
s=Path('/workspace/scratch/debug-messages');raw=s/'raw';records={side:json.loads((s/('replay-raw-'+side)/'records.json').read_text()) for side in ['before','after']}
assert len(records['before'])==len(records['after'])==96
assert [r['adapter'] for r in records['before']]==[r['adapter'] for r in records['after']]
differences=[]
for before,after in zip(records['before'],records['after']):
 if before!=after:differences.append({'adapter':before['adapter'],'before':before,'after':after})
assert [r['adapter'] for r in differences]==['99-correctness-core-debug-messages'],[r['adapter'] for r in differences]
ours=differences[0];assert ours['before']['exit']==ours['after']['exit']==0
assert len(ours['before']['writes'])==1 and not ours['after']['writes']
assert ours['before']['writes'][0]['file']=='src/compiler/debug.ts'
assert json.loads(ours['before']['stdout_stderr'])['edits']==4
assert json.loads(ours['after']['stdout_stderr'])['edits']==0
(raw/'paired-replay.diff').write_text(''.join(difflib.unified_diff(json.dumps(records['before'],indent=2).splitlines(True),json.dumps(records['after'],indent=2).splitlines(True),fromfile='before/records.json',tofile='after/records.json')))
result={'complete':True,'adapters':96,'identicalOtherAdapters':95,'differentAdapters':[r['adapter'] for r in differences],'ourBeforeEdits':4,'ourAfterEdits':0,'ourBeforeChildExit':0,'ourAfterChildExit':0,'normalization':'Only each disposable tree root in stdout/stderr. No result, reason, exit, write or write contents normalized.','commonRefusals':[r['adapter'] for r in records['before'] if r['exit']!=0]}
(raw/'paired-replay-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
