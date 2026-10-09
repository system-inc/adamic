import json,os,subprocess,tempfile
from pathlib import Path
root=Path('/workspace/adamic');out=Path('/workspace/scratch/hidden-representation-mutants/tnode-clock');out.mkdir(exist_ok=True)
e=root/'internal/lower/expression.go'; c=root/'internal/lower/generic_constraint_storage.go';original=e.read_text();storage=c.read_text()
mutants=[
('default-node',e,original.replace('case ast.KindIdentifier:\n','case ast.KindIdentifier:\n\t\tif l.genericDepth > 0 && node.Text() == "node" { return ir.NumberConstant{}, nil }\n',1),'./internal/oracle','^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^hidden_boundary_generic_tnode[.]a$','stdout'),
('generic-value',e,original.replace('return nil, l.notYet(node, "a generic function as a value")','return ir.Undefined{Of: ir.Closure}, nil'),'./internal/lower','^TestHiddenTNodeGenericValueRemainsNotYet$','want generic value refusal'),
('source-object-header',c,storage.replace('return ir.Union, true','return ir.Object, true'),'./internal/lower','^TestRepresentationClockSourceCheckedTypes$','source: representation'),
('source-no-tagged-storage',c,storage.replace('return ir.Union, true','return 0, false'),'./internal/lower','^TestRepresentationClockSourceCheckedTypes$','source: representation'),
]
rows=[]
for name,source,changed,pkg,selector,marker in mutants:
 assert changed != source.read_text()
 with tempfile.TemporaryDirectory(prefix='hidden-clock-mutant-') as tmp:
  tmp=Path(tmp);replacement=tmp/source.name;replacement.write_text(changed)
  overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  cmd=['go','test','-overlay',str(overlay),pkg,'-run',selector,'-count=1','-timeout','90s','-v']
  log=out/(name+'.log')
  with log.open('w') as f:r=subprocess.run(cmd,cwd=root,env=os.environ,stdout=f,stderr=subprocess.STDOUT)
  text=log.read_text();caught=r.returncode!=0 and '--- FAIL:' in text and marker in text and '[build failed]' not in text
  rows.append(dict(mutant=name,caught=caught,exit=r.returncode,command=cmd));print(name,caught,flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
