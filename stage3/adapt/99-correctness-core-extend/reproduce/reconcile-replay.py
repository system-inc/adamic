import json,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');rows=json.loads((s/'paired-replay.json').read_text());expected=sorted(p.name for p in (repo/'stage3/adapt').iterdir() if p.is_dir());assert [r['adapter'] for r in rows]==expected
for r in rows:
 name=r['adapter'];a,b=r['sides']['before'],r['sides']['after'];assert a['exit']==b['exit'];assert b['changed']==[];assert a['changed']==(['src/compiler/core.ts'] if name=='99-correctness-core-extend' else [])
 if name!='99-correctness-core-extend':
  texts=[(s/('replay-'+side+'-'+name+'.log')).read_text().replace(str(s/('replay-'+side)),'<tree>') for side in ['before','after']];assert texts[0]==texts[1],name
def manifest(root):return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
assert manifest(s/'replay-before')==manifest(s/'replay-after')==manifest(s/'after')
(s/'replay-reconciled.json').write_text(json.dumps({'pass':True,'adapters':len(rows),'matchingExistingRefusals':sum(r['sides']['before']['exit']!=0 for r in rows),'onlyDifferentAdapter':'99-correctness-core-extend','finalSourceFiles':len(manifest(s/'after'))},indent=2)+'\n')
print('all replay outcomes independently reconciled')
