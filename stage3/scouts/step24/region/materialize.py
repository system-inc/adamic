#!/usr/bin/env python3
"""Reuse the acceptance driver's exact header materialization, outside the repository."""
import hashlib,json,sys
from pathlib import Path
repository=Path(__file__).resolve().parents[4]
sys.path.insert(0,str(repository/'stage3/drivers/tsc'))
from corpus import materialize
root=repository/'stage3/drivers/tsc';out=Path(sys.argv[1]);out.mkdir(parents=True)
selection=json.loads((root/'selection.json').read_text())
rows=[]
for case in selection['cases']+[{'id':'tiny','options':{'strict':True,'target':'es2020'}}]:
 identifier=case['id'];folder=out/identifier;folder.mkdir();source=root/('tiny' if identifier=='tiny' else 'corpus/'+identifier)
 programs=sorted(source.glob('*.a')) if identifier=='tiny' else [root/case['path']]
 files=[]
 for program in programs:
  raw=program.read_bytes()
  if identifier!='tiny' and hashlib.sha256(raw).hexdigest()!=case['source_sha256']:raise RuntimeError('Input hash changed: '+identifier)
  if identifier=='tiny' and raw.startswith(b'// a-check:'):raw=raw.split(b'\n',1)[1]
  content,options=materialize(raw)
  if identifier!='tiny' and options!=case['options']:raise RuntimeError('Headers changed')
  name=program.with_suffix('.ts').name if identifier=='tiny' else program.name
  (folder/name).write_text(content);files.append(name)
 config={'compilerOptions':{'types':[],'skipDefaultLibCheck':True,'noErrorTruncation':True,'ignoreDeprecations':'6.0','noEmit':True,**case['options']},'files':files}
 (folder/'tsconfig.json').write_text(json.dumps(config)+'\n')
 rows.append({'id':identifier,'config':str(folder/'tsconfig.json'),'golden':str(source)})
(out/'manifest.json').write_text(json.dumps(rows,indent=2)+'\n')
assert len(rows)==301
print('Materialized 301 acceptance projects without editing their corpus or goldens.')
