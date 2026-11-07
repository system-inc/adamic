from pathlib import Path
import json
s=Path('/workspace/scratch/arrayof');own='99-correctness-core-arrayof';a=json.loads((s/'replay-before.json').read_text());b=json.loads((s/'replay-after.json').read_text());assert [r['adapter'] for r in a]==[r['adapter'] for r in b];expected=sorted(p.name for p in Path('/workspace/adamic/stage3/adapt').iterdir() if p.is_dir());assert [r['adapter'] for r in a]==expected
changed=[]
for x,y in zip(a,b):
 if x!=y:changed.append(x['adapter'])
 if x['adapter']!=own:assert x==y,(x,y)
 else:
  assert x['exit']==y['exit']==0;assert not x['stderr'] and not y['stderr'];assert json.loads(x['stdout'])==dict(files=1,edits=1,javascriptChanging=True);assert json.loads(y['stdout'])==dict(files=0,edits=0,javascriptChanging=True)
assert changed==[own],changed
result=dict(pass_=True,adapters=len(a),changedOutcomes=changed,allOtherOutcomesByteIdenticalAfterTreePathNormalization=True,ownBefore=dict(files=1,edits=1),ownAfter=dict(files=0,edits=0),identicalNonzeroMainOutcomes=[r['adapter'] for r in a if r['exit']]);(s/'PAIRED-REPLAY.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
