import json,os,subprocess,time
from pathlib import Path
p=Path(__file__).resolve().parent;rows=json.loads((p/'rows.json').read_text())+['TestMathAndToFixedMatchJavaScript','TestDecodeCacheLockDeadline']
pattern='^('+'|'.join(rows)+')$';(p/'scope.json').write_text(json.dumps({'rows':rows,'bounded':True,'pattern':pattern,'unknown':'Every omitted package row'},indent=2))
for id in ['control']+[m['id'] for m in json.loads((p/'plan.json').read_text())]:
 env=dict(os.environ,ADAMIC_MUTANT='' if id=='control' else id,ADAMIC_NATIVE_SPLIT='u051-'+id)
 subprocess.run(['python3',str(p/'run.py'),id,'-run',pattern],env=env,check=True)
 r=json.loads((p/(id+'.json')).read_text())
 if id=='control' and r['exit']!=0:raise SystemExit('RED CONTROL')
 if any('panic:' in x for v in r['outputs'].values() for x in v):
  for row in json.loads((p/'rows.json').read_text()):
   subprocess.run(['python3',str(p/'run.py'),id+'-'+row,'-run','^'+row+'$'],env=env,check=True)
